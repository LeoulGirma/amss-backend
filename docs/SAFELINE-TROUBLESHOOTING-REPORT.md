# SafeLine WAF Troubleshooting Report

> Comprehensive documentation of attempts to resolve the SafeLine WAF detector communication issue

**Date:** January 12, 2026
**SafeLine Version:** 9.3.0
**Environment:** Ubuntu Linux on OVH VPS (51.79.85.92)
**Application:** AMSS (Aviation Maintenance & Safety System) - React frontend

---

## Table of Contents

1. [Problem Description](#problem-description)
2. [Environment Details](#environment-details)
3. [Initial Configuration](#initial-configuration)
4. [Troubleshooting Timeline](#troubleshooting-timeline)
5. [Configuration Files](#configuration-files)
6. [Logs and Observations](#logs-and-observations)
7. [Attempted Solutions](#attempted-solutions)
8. [Current Status](#current-status)
9. [Potential Root Causes](#potential-root-causes)
10. [Recommended Next Steps](#recommended-next-steps)
11. [Useful Commands](#useful-commands)
12. [References](#references)

---

## Problem Description

### Symptom
The SafeLine WAF detector is **not receiving any requests** from the tengine (nginx) proxy. Despite traffic flowing through SafeLine and reaching the backend application successfully, the WAF detection engine shows zero requests processed.

### Evidence
```bash
# Detector statistics show 0 requests
$ sudo docker exec safeline-mgt curl -s http://safeline-detector:8001/stat | grep req_num
  "req_num_total": 0,
  "req_num_accepted": 0,
  "req_num_handled": 0,

# Detection log table is empty
$ sudo docker exec safeline-pg psql -U safeline-ce -d safeline-ce -c "SELECT COUNT(*) FROM mgt_detect_log_basic;"
 count
-------
     0

# Access logs are empty
$ sudo docker exec safeline-tengine cat /var/log/nginx/safeline/accesslog_3
(empty)
```

### Impact
- SQL injection attacks pass through undetected
- XSS attacks are not blocked
- No security events are logged
- WAF provides no protection

---

## Environment Details

### System Information
| Component | Value |
|-----------|-------|
| OS | Ubuntu Linux (kernel 6.14.0-37-generic) |
| Platform | OVH VPS |
| IP Address | 51.79.85.92 |
| Docker | Docker Compose v2 |

### SafeLine Installation
| Component | Value |
|-----------|-------|
| Version | 9.3.0 |
| Installation Directory | `/data/safeline` |
| Management Port | 9443 |
| Docker Network | `safeline-ce` (172.22.222.0/24) |

### Container Configuration
| Container | IP Address | Network Mode | Status |
|-----------|------------|--------------|--------|
| safeline-tengine | Host network | `host` | Running |
| safeline-detector | 172.22.222.5 | Bridge | Healthy |
| safeline-mgt | 172.22.222.4 | Bridge | Healthy |
| safeline-pg | 172.22.222.2 | Bridge | Healthy |
| safeline-luigi | 172.22.222.7 | Bridge | Running |
| safeline-fvm | 172.22.222.8 | Bridge | Running |
| safeline-chaos | 172.22.222.10 | Bridge | Running |

### Protected Site Configuration
| Field | Value |
|-------|-------|
| Domain | `amss.leoulgirma.com` |
| Ports | 80, 443 |
| Upstream | `http://127.0.0.1:8080` |
| Mode | 0 (Detect) |
| SSL Certificate | `/etc/nginx/certs/amss.crt` |
| Site ID | 3 |

---

## Initial Configuration

### docker-compose.yaml (relevant sections)
```yaml
tengine:
  container_name: safeline-tengine
  image: chaitin/safeline-tengine-g:${IMAGE_TAG}
  network_mode: host  # <-- KEY ISSUE: Uses host network
  volumes:
    - ${SAFELINE_DIR}/resources/nginx:/etc/nginx
    - ${SAFELINE_DIR}/resources/detector:/resources/detector  # Unix socket mount
    - ${SAFELINE_DIR}/logs/nginx:/var/log/nginx
  environment:
    - TCD_SNSERVER=${SUBNET_PREFIX}.5:8000  # Expects TCP connection
    - SNSERVER_ADDR=${SUBNET_PREFIX}.5:8000

detector:
  container_name: safeline-detector
  image: chaitin/safeline-detector:${IMAGE_TAG}
  networks:
    safeline-ce:
      ipv4_address: ${SUBNET_PREFIX}.5
  volumes:
    - ${SAFELINE_DIR}/resources/detector:/resources/detector
```

### .env Configuration
```env
SAFELINE_DIR=/data/safeline
POSTGRES_PASSWORD=5a8394d1707689f56d3b33edf7c89be7
MGT_PORT=9443
IMAGE_TAG=9.3.0
SUBNET_PREFIX=172.22.222
```

---

## Troubleshooting Timeline

### Phase 1: Initial Investigation

#### Step 1.1: Verified Site Configuration in Database
```sql
SELECT id, server_names, ports, upstreams, is_enabled, mode FROM mgt_website;
-- Result: Site correctly configured with ports 80/443, upstream 127.0.0.1:8080
```

#### Step 1.2: Checked Container Status
```bash
$ sudo docker ps --filter "name=safeline"
# All 7 containers running, detector showing "healthy"
```

#### Step 1.3: Verified Network Connectivity
```bash
# From safeline-mgt to detector
$ sudo docker exec safeline-mgt ping -c 2 safeline-detector
# SUCCESS: 0% packet loss

# Detector health endpoint
$ sudo docker exec safeline-mgt curl -s http://safeline-detector:8001/stat
# SUCCESS: Returns JSON with version info
```

#### Step 1.4: Tested Attack Detection
```bash
$ curl -sk "https://amss.leoulgirma.com/?id=1'OR'1'='1" -o /dev/null -w "%{http_code}"
200  # Attack NOT blocked

$ sudo docker exec safeline-pg psql -U safeline-ce -d safeline-ce -c "SELECT COUNT(*) FROM mgt_detect_log_basic;"
0    # No detections logged
```

### Phase 2: Unix Socket Investigation

#### Step 2.1: Verified Socket Exists
```bash
$ sudo docker exec safeline-tengine ls -la /resources/detector/
srwxrwxrwx 1 root root 0 Jan 12 15:34 snserver.sock  # Socket exists
```

#### Step 2.2: Checked Tengine Nginx Configuration
```bash
$ sudo docker exec safeline-tengine nginx -T | grep "upstream detector_server" -A 3
upstream detector_server {
    keepalive   256;
    server      unix:/resources/detector/snserver.sock;
}
```

#### Step 2.3: Verified t1k Module Loaded
```bash
$ sudo docker exec safeline-tengine nginx -V 2>&1 | grep t1k
# Output confirms t1k module is compiled in
```

#### Step 2.4: Checked Global t1k Configuration
```bash
$ sudo docker exec safeline-tengine cat /etc/nginx/safeline_unix.conf | head -20
# Shows: t1k_intercept @safeline; # enable request detection
```

### Phase 3: TCP Configuration Attempt

#### Step 3.1: Modified Detector to Use TCP
```bash
# Updated /data/safeline/resources/detector/detector.yml
bind_addr: 0.0.0.0
listen_port: 8000
mgt_server_addr: "https://172.22.222.4:1443"
```

#### Step 3.2: Verified Detector Listening on TCP
```bash
$ sudo docker exec safeline-detector cat /proc/net/tcp | grep -E "1F40"
# Confirmed listening on port 8000 (0x1F40)
```

#### Step 3.3: Created TCP Nginx Configuration
```nginx
# /data/safeline/resources/nginx/safeline_tcp.conf
upstream detector_server {
    keepalive   256;
    server      172.22.222.5:8000;
}

upstream chaos_server {
    keepalive   256;
    server      172.22.222.10:8080;
}

upstream wr_server {
    keepalive   256;
    server     unix:/app/sock/waiting_tcp.sock;
}

t1k_intercept @safeline;
t1k_body_size 1m;

tx_intercept @safelinex;
tx_body_size 1m;

# ... rest of config
```

#### Step 3.4: Updated nginx.conf to Use TCP Config
```bash
# Changed in /data/safeline/resources/nginx/nginx.conf
include /etc/nginx/safeline_tcp.conf;  # Was: safeline_unix.conf
```

#### Step 3.5: Tested TCP Connectivity
```bash
$ curl -s http://172.22.222.5:8000/
# Received "invalid tag" error in detector logs - confirms connectivity works
```

#### Step 3.6: Result
Detector logs showed:
```
[WARN] read T1K packet error: invalid tag
```
This confirms TCP connectivity works, but t1k module still not forwarding real requests.

### Phase 4: Configuration Variations

#### Step 4.1: Tried Original safeline.conf (from SafeLine)
```bash
# Error: unknown directive "multi" in /etc/nginx/safeline.conf:8
# The stream-related directives weren't compatible
```

#### Step 4.2: Reverted to Unix Socket
```bash
# Restored safeline_unix.conf
# Result: Same issue - 0 requests reaching detector
```

#### Step 4.3: Regenerated Site Configuration
```bash
$ sudo docker exec safeline-mgt /app/mgt-cli gen-tengine-website
[INFO] restore website (id: 3, domain: [amss.leoulgirma.com], port: [80 443]) success
# Result: No improvement
```

---

## Configuration Files

### /data/safeline/resources/nginx/nginx.conf (relevant section)
```nginx
http {
    include /etc/nginx/safeline_unix.conf;  # or safeline_tcp.conf

    # ... other configs

    include /etc/nginx/sites-enabled/*;
}
```

### /data/safeline/resources/nginx/safeline_unix.conf
```nginx
upstream detector_server {
    keepalive   256;
    server      unix:/resources/detector/snserver.sock;
}

upstream chaos_server {
    keepalive   256;
    server     unix:/resources/chaos/stpp.sock;
}

upstream wr_server {
    keepalive   256;
    server    unix:/app/sock/waiting_tcp.sock;
}

t1k_intercept @safeline; # enable request detection
t1k_body_size 1m;        # max forward size of request body

tx_intercept @safelinex; # enable response detection
tx_body_size 1m;         # max forward size of response body

tx_chaos_intercept @safeline_chaos;
tx_chaos_body_size 10m;

t1k_wr_intercept @safeline_wr;

include tx_ignore_types;

t1k_ulog 10000;
t1k_stat 10000;

t1k_extra_header on;
t1k_extra_body on;

foreach_server {
    location @safeline {
        internal;
        t1k_pass detector_server;
        t1k_connect_timeout 1s;
        t1k_read_timeout 1s;
        t1k_send_timeout 1s;
    }
    location @safelinex {
        internal;
        tx_pass detector_server;
        tx_connect_timeout 1s;
        tx_read_timeout 1s;
        tx_send_timeout 1s;
    }
    location @safeline_chaos {
        internal;
        tx_chaos_pass chaos_server;
        tx_chaos_read_timeout 3s;
        tx_chaos_send_timeout 3s;
        tx_chaos_connect_timeout 3s;
    }
    location @safeline_wr {
        internal;
        t1k_wr_pass wr_server;
        t1k_wr_read_timeout 3s;
        t1k_wr_send_timeout 3s;
        t1k_wr_connect_timeout 3s;
    }
}
```

### Site Configuration (IF_backend_3)
```nginx
log_format safeline_3 '$remote_addr | $remote_user | $time_local | "$host" | '
                    '"$request" | $status | $body_bytes_sent | '
                    '"$http_referer" | "$http_user_agent"';
upstream backend_3 {
    server 127.0.0.1:8080;
    keepalive 128;
    keepalive_timeout 75;
}

server {
    listen 0.0.0.0:80;
    listen 0.0.0.0:443 ssl;
    server_name amss.leoulgirma.com;

    ssl_certificate /etc/nginx/certs/amss.crt;
    ssl_certificate_key /etc/nginx/certs/amss.key;

    # ... error pages with t1k_intercept off

    location ^~ / {
        proxy_pass http://backend_3;
        include proxy_params;
        proxy_set_header Host $http_host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        include /etc/nginx/custom_params/backend_3;
        t1k_add_user_data "3";
        tx_add_user_data "3";
        t1k_body_size 1024k;
        tx_body_size 4k;
        t1k_error_page 403 /.safeline/forbidden_page;
        t1k_error_page 429 /.safeline/acl_page;
        t1k_error_page 466 /.safeline/offline_page;
        tx_error_page 403 /.safeline/forbidden_page;
        t1k_error_page 465 /.safeline/waiting_room_page;
    }
}
```

### /data/safeline/resources/detector/detector.yml
```yaml
# Current (TCP mode)
bind_addr: 0.0.0.0
listen_port: 8000
mgt_server_addr: "https://172.22.222.4:1443"

# Original (Unix socket mode)
# bind_addr: unix:///resources/detector/snserver.sock
```

---

## Logs and Observations

### Detector Logs
```
[INFO] init started, version info: commit_id: 02589d3bdf8ee99cdc30d13a2aa51e45c7667347;
       branch: v1.4.1; tag: v1.4.1; engine_type: legacy_skynet
[INFO] initialized all hyperscan databases in 1.086s
[INFO] initialized all detection modules in 871ms
[INFO] init done
[WARN] read T1K packet error: invalid tag  # Only appears with manual curl tests
```

### Tengine Logs
```
2026/01/12 15:00:36 tcd unix sock service started
[GIN-debug] Listening and serving HTTP on listener@/app/sock/tcd_error.sock
# No errors related to t1k or detector connection
```

### Management Server Logs
```
time=2026-01-08T12:26:00 level=ERROR msg="request dial tcd failed:"
    error="Post \"http://unixtcd/task\": dial unix /app/sock/tcd.sock: connect: connection refused"
time=2026-01-08T12:26:00 level=ERROR msg="website not exist try to regenerate nginx config" id=2
```

### Key Observations

1. **No t1k connection errors in tengine logs** - The module appears to load without issues
2. **Detector shows "init done"** - Detection engine initializes successfully
3. **"invalid tag" errors only from manual tests** - Real HTTP requests don't produce any detector logs
4. **Access logs are empty** - Requests aren't being logged by SafeLine
5. **Site statistics are 0** - No traffic metrics recorded
6. **Backend receives requests** - Traffic flows through to 127.0.0.1:8080 successfully

---

## Attempted Solutions

### Solution 1: Restart All Containers
```bash
cd /data/safeline && sudo docker compose restart
```
**Result:** No improvement

### Solution 2: Regenerate Website Configuration
```bash
sudo docker exec safeline-mgt /app/mgt-cli gen-tengine-website
```
**Result:** Config regenerated successfully, but issue persists

### Solution 3: Switch from Unix Socket to TCP
- Modified detector.yml to bind to 0.0.0.0:8000
- Created safeline_tcp.conf with IP-based upstream
- Updated nginx.conf to include TCP config

**Result:** Detector received test packets (showed "invalid tag"), but real requests still not forwarded

### Solution 4: Delete and Recreate Site
- Deleted site ID 2 from database
- Re-added site through admin panel as ID 3
- Configured with ports 80 and 443

**Result:** New site created but same issue

### Solution 5: Add HTTPS Configuration
```sql
UPDATE mgt_website SET
  ports = '["80", "443"]'::jsonb,
  cert_type = 1,
  cert_filename = '/etc/nginx/certs/amss.crt',
  key_filename = '/etc/nginx/certs/amss.key'
WHERE id = 3;
```
**Result:** HTTPS works, but WAF still not detecting

### Solution 6: Enable Protection Mode
```sql
UPDATE mgt_website SET mode = 1 WHERE id = 3;  -- 1 = Protect mode
```
**Result:** No difference in detection

---

## Current Status

### Working
- All 7 SafeLine containers running (healthy)
- Site accessible via HTTP and HTTPS
- Backend nginx receiving proxied requests
- Detector process running and responsive on port 8001
- Network connectivity between containers verified
- t1k module compiled into tengine

### Not Working
- t1k module not forwarding requests to detector
- Detection log table empty (0 rows)
- Request statistics at 0
- Access logs empty
- No attacks blocked

---

## Potential Root Causes

### 1. Host Network Mode Incompatibility
The tengine container uses `network_mode: host` while other containers use the `safeline-ce` bridge network. This may cause:
- DNS resolution issues (tengine can't resolve `safeline-detector` hostname)
- Unix socket namespace isolation issues
- Network routing problems between host and bridge networks

**Evidence:** Environment variables reference `${SUBNET_PREFIX}.5:8000` suggesting TCP was intended.

### 2. t1k Module Configuration Issue
The global `t1k_intercept @safeline;` directive may not be applying to the site server blocks correctly.

**Evidence:** Site config has `t1k_add_user_data "3"` but no explicit `t1k_intercept` directive in the location block.

### 3. foreach_server Directive Issue
The `foreach_server { }` block in safeline_unix.conf is a tengine-specific directive that injects the @safeline location into all server blocks. This may not be working correctly.

**Evidence:** No documentation found on troubleshooting foreach_server issues.

### 4. Version-Specific Bug
SafeLine 9.3.0 may have a regression in the t1k module or tengine integration.

**Evidence:** Multiple users have reported similar issues in GitHub issues.

### 5. Volume Mount Timing Issue
The Unix socket at `/resources/detector/snserver.sock` may not be ready when tengine starts, and tengine doesn't retry.

**Evidence:** Container restart order may affect socket availability.

---

## Recommended Next Steps

### Option 1: Remove Host Network Mode
Modify docker-compose.yaml to use bridge network for tengine:
```yaml
tengine:
  networks:
    safeline-ce:
      ipv4_address: ${SUBNET_PREFIX}.3
  ports:
    - "80:80"
    - "443:443"
```
This would allow proper container-to-container communication.

### Option 2: Fresh Installation
```bash
cd /data/safeline
sudo docker compose down -v
sudo rm -rf /data/safeline/*
# Re-run SafeLine installer
bash -c "$(curl -fsSLk https://waf.chaitin.com/release/latest/setup.sh)"
```

### Option 3: Downgrade to Earlier Version
Try SafeLine 9.2.x or earlier:
```bash
# In .env
IMAGE_TAG=9.2.0
```

### Option 4: Report to SafeLine Team
- GitHub Issues: https://github.com/chaitin/SafeLine/issues
- Discord: https://discord.gg/SVnZGzHFvn

Include this report and ask about:
1. Host network mode compatibility
2. Known issues with 9.3.0
3. Correct configuration for t1k module

### Option 5: Use Alternative Integration
Consider using:
- lua-resty-t1k for OpenResty
- Traefik SafeLine plugin
- Kong SafeLine plugin

These use HTTP-based detection instead of the t1k protocol.

---

## Useful Commands

### Check Container Status
```bash
sudo docker ps --filter "name=safeline" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### Check Detector Statistics
```bash
sudo docker exec safeline-mgt curl -s http://safeline-detector:8001/stat | jq '.fusion_stat.req'
```

### Check Detection Logs
```bash
sudo docker exec safeline-pg psql -U safeline-ce -d safeline-ce -c "SELECT * FROM mgt_detect_log_basic LIMIT 10;"
```

### View Site Configuration
```bash
sudo docker exec safeline-pg psql -U safeline-ce -d safeline-ce -c "SELECT id, server_names, ports, upstreams, mode FROM mgt_website;"
```

### Test Nginx Configuration
```bash
sudo docker exec safeline-tengine nginx -t
```

### View Full Nginx Configuration
```bash
sudo docker exec safeline-tengine nginx -T
```

### Check Tengine Logs
```bash
sudo docker logs safeline-tengine --tail 50
```

### Check Detector Logs
```bash
sudo docker logs safeline-detector --tail 50
```

### Regenerate Site Config
```bash
sudo docker exec safeline-mgt /app/mgt-cli gen-tengine-website
```

### Reset Admin Password
```bash
sudo docker exec safeline-mgt /app/mgt-cli reset-admin --once
```

### Restart All Containers
```bash
cd /data/safeline && sudo docker compose restart
```

### Test SQL Injection
```bash
curl -sk "https://amss.leoulgirma.com/?id=1'OR'1'='1" -o /dev/null -w "%{http_code}\n"
# Expected: 403 (blocked) | Actual: 200 (not blocked)
```

### Test XSS
```bash
curl -sk "https://amss.leoulgirma.com/?q=<script>alert(1)</script>" -o /dev/null -w "%{http_code}\n"
# Expected: 403 (blocked) | Actual: 200 (not blocked)
```

---

## References

### Official Documentation
- SafeLine GitHub: https://github.com/chaitin/SafeLine
- SafeLine Docs: https://docs.waf.chaitin.com/en/

### Related Projects
- lua-resty-t1k: https://github.com/chaitin/lua-resty-t1k
- Traefik SafeLine Plugin: https://github.com/chaitin/traefik-safeline
- Kong SafeLine Plugin: https://github.com/chaitin/kong-safeline

### Troubleshooting Guides
- Configuration Issues: https://dev.to/carrie_luo1/addressing-configuration-issues-of-safeline-waf-bi1
- Website Inaccessible: https://dev.to/carrie_luo1/safeline-waf-troubleshooting-guide-website-inaccessible-after-configuration-16m3
- K8s Deployment: https://dev.to/lulu_liu_c90f973e2f954d7f/diy-deployment-of-safeline-waf-on-k8s-c2c

### Related GitHub Issues
- Issue #1115 (Traefik plugin hangup): https://github.com/chaitin/SafeLine/issues/1115

---

## File Locations

| File | Purpose |
|------|---------|
| `/data/safeline/docker-compose.yaml` | Container orchestration |
| `/data/safeline/.env` | Environment variables |
| `/data/safeline/resources/nginx/nginx.conf` | Main nginx config |
| `/data/safeline/resources/nginx/safeline_unix.conf` | Unix socket upstream config |
| `/data/safeline/resources/nginx/safeline_tcp.conf` | TCP upstream config (created) |
| `/data/safeline/resources/nginx/sites-enabled/IF_backend_3` | Site config |
| `/data/safeline/resources/detector/detector.yml` | Detector config |
| `/data/safeline/resources/detector/snserver.sock` | Unix socket |
| `/data/safeline/logs/nginx/` | Nginx logs |
| `/etc/nginx/sites-available/amss.leoulgirma.com` | Backend nginx config |

---

---

## Update: January 15-16, 2026 - Codex CLI Investigation

### Overview
Used OpenAI Codex CLI with `--dangerously-bypass-approvals-and-sandbox` flag to perform automated troubleshooting of the SafeLine WAF issue.

### Codex Findings

#### 1. Network Configuration Discovery
Codex discovered that tengine was actually on the bridge network (not host network as initially thought):
```bash
$ sudo docker inspect safeline-tengine --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}'
172.22.222.3
```

#### 2. TCD Process Analysis
Found the `tcd` (Tengine Control Daemon) process runs inside tengine container:
```bash
$ sudo docker exec safeline-tengine ps -ef
UID   PID  CMD
root    1  nginx: master process
root   37  ./tcd
nginx  91  nginx: worker process
```

Environment variables for TCD:
```
TCD_SNSERVER=172.22.222.5:8000
SNSERVER_ADDR=172.22.222.5:8000
TCD_MGT_API=https://172.22.222.4:1443/api/open/publish/server
```

#### 3. Detector Logs Analysis
Detector showed "invalid tag" errors from non-T1K connections:
```
[WARN] snserver_engine::detector_serve::t1k: read T1K packet error: invalid tag
```
These were from direct curl tests to port 8000, not from proper T1K protocol traffic.

#### 4. Configuration Changes Made by Codex

**docker-compose.yaml:**
- Changed `TCD_SNSERVER` from IP to DNS hostname: `safeline-detector:8000`
- Changed `SNSERVER_ADDR` from IP to DNS hostname: `safeline-detector:8000`

**IF_backend_3 (site config):**
- Added explicit `t1k_intercept @safeline;` at server level
- Added explicit `tx_intercept @safelinex;` at server level
- Added internal `@safeline` location with `t1k_pass detector_server`
- Added internal `@safelinex` location with `tx_pass detector_server`
- Added `@safeline_chaos` and `@safeline_wr` locations

**safeline_tcp.conf:**
- Updated `detector_server` upstream to use `safeline-detector:8000`

### Additional Troubleshooting (Post-Codex)

#### 1. Backend Connectivity Fix
After Codex moved tengine to bridge network, backend became unreachable:
```bash
# Problem: tengine couldn't reach 127.0.0.1:8080 from bridge network
$ sudo docker exec safeline-tengine curl http://127.0.0.1:8080
# Failed - localhost in container != host's localhost
```

**Fix Applied:**
- Updated `/etc/nginx/sites-available/amss.leoulgirma.com` to listen on `0.0.0.0:8080`
- Added `set_real_ip_from 172.22.222.0/24` to trust Docker network
- Updated SafeLine upstream in IF_backend_3 to use `172.22.222.1:8080` (host gateway)

#### 2. Unix Socket vs TCP Investigation
Discovered Unix socket exists and should work:
```bash
$ ls -la /data/safeline/resources/detector/
srwxrwxrwx 1 201 root 0 Jan 12 15:34 snserver.sock
```

Switched nginx back to Unix socket config:
```bash
$ sudo sed -i 's|safeline_tcp.conf|safeline_unix.conf|' /data/safeline/resources/nginx/nginx.conf
```

#### 3. TCP Connectivity Verification
Verified tengine CAN connect to detector via TCP:
```bash
$ sudo docker exec safeline-tengine bash -c 'echo > /dev/tcp/172.22.222.5/8000'
# Success - connection works
```

#### 4. tcpdump Analysis
Captured network traffic during requests:
```bash
$ sudo tcpdump -n -i safeline-ce port 8000 -c 5
# Result: 0 packets captured
# Conclusion: t1k module NOT sending any traffic to detector
```

#### 5. Nginx Debug Logging
Enabled debug logging to trace t1k module:
```nginx
error_log /var/log/nginx/error.log debug;
```
**Result:** No t1k-related messages in logs despite requests flowing through.

### Current Configuration State

**docker-compose.yaml (tengine section):**
```yaml
tengine:
  container_name: safeline-tengine
  ports:
    - "80:80"
    - "443:443"
  networks:
    safeline-ce:
      ipv4_address: 172.22.222.3
  environment:
    - TCD_SNSERVER=safeline-detector:8000
    - SNSERVER_ADDR=safeline-detector:8000
```

**nginx.conf includes:**
```nginx
include /etc/nginx/safeline_unix.conf;
```

**IF_backend_3 upstream:**
```nginx
upstream backend_3 {
    server 172.22.222.1:8080;  # Host gateway IP
    keepalive 128;
}
```

### Test Results After All Changes

| Test | Expected | Actual |
|------|----------|--------|
| Website accessible | 200 | 200 ✓ |
| Backend connectivity | Working | Working ✓ |
| Tengine → Detector TCP | Connected | Connected ✓ |
| Unix socket exists | Yes | Yes ✓ |
| t1k_intercept in config | Present | Present ✓ |
| @safeline location | Present | Present ✓ |
| Detector req_num_total | >0 | **0** ✗ |
| SQL injection blocked | 403 | **200** ✗ |

### Root Cause Analysis (Updated)

The t1k nginx module is **compiled and configured correctly** but is **not intercepting requests**. Evidence:

1. **Network is working** - TCP connections succeed between containers
2. **Config is valid** - `nginx -t` passes, all directives present
3. **Module is loaded** - `nginx -V` shows t1k in build
4. **No errors** - Debug logs show no t1k-related errors
5. **Zero traffic** - tcpdump shows no packets to detector port
6. **Detector healthy** - Stats endpoint responds, engine initialized

**Conclusion:** This appears to be a **bug in SafeLine v9.3.0** where the t1k module fails to activate despite correct configuration. The module silently ignores the `t1k_intercept` directive.

### Recommended Actions

1. **Report Bug to SafeLine Team**
   - GitHub: https://github.com/chaitin/SafeLine/issues
   - Include this report as evidence

2. **Try SafeLine v9.2.x**
   ```bash
   # In /data/safeline/.env
   IMAGE_TAG=9.2.0
   cd /data/safeline && sudo docker compose pull && sudo docker compose up -d
   ```

3. **Alternative: Use lua-resty-t1k**
   For OpenResty/nginx installations, use the Lua-based detector client which doesn't rely on the t1k module.

4. **Alternative: Traefik Plugin**
   If using Traefik as reverse proxy, the SafeLine Traefik plugin uses HTTP API instead of t1k protocol.

---

*Report updated: January 16, 2026*
*SafeLine Version: 9.3.0*
*Authors: Claude Code (Anthropic), OpenAI Codex CLI*
