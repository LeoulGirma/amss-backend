# SafeLine WAF Fix Documentation

**System:** SafeLine WAF v9.2.0
**Server:** amss.leoulgirma.com / amss-api-uat.duckdns.org
**Last Updated:** February 2026

---

## Problem Description

The SafeLine WAF was not detecting or blocking any requests. The detector statistics showed `req_num_total: 0`, indicating that no requests were being forwarded from the tengine (nginx) module to the detector for inspection.

### Symptoms

- All requests passed through without WAF inspection
- Detector stats at `/stat` endpoint showed zero requests
- SQL injection and XSS attacks were not being blocked
- The t1k nginx module appeared to be configured but not functioning

---

## Root Cause Analysis

After extensive investigation, **two root causes** were identified:

### 1. Missing Internal Named Locations

The site configuration file (`/data/safeline/resources/nginx/sites-enabled/IF_backend_3`) was missing the internal named locations that define where the t1k module should forward requests for inspection.

The `t1k_intercept @safeline;` directive tells nginx to intercept requests and send them to the `@safeline` named location. However, this location was not defined in the site config, so the t1k module had no target to forward requests to.

**Missing locations:**
- `@safeline` - forwards requests to detector via `t1k_pass`
- `@safelinex` - forwards responses to detector via `tx_pass`
- `@safeline_chaos` - for chaos engineering features
- `@safeline_wr` - for waiting room features

### 2. Unreachable Backend Server

The backend upstream was configured as `127.0.0.1:8080`, which is localhost inside the tengine container. This address is not reachable from within the Docker container because it refers to the container's own loopback interface, not the host machine.

---

## Fixes Applied

### Fix 1: Add t1k_intercept Directives

Added the intercept directives at both server and location level in the site config:

```nginx
server {
    listen 0.0.0.0:80;
    listen 0.0.0.0:443;
    server_name amss.leoulgirma.com;
    t1k_intercept @safeline;      # Added
    tx_intercept @safelinex;      # Added
    ...

    location ^~ / {
        proxy_pass http://backend_3;
        ...
        t1k_intercept @safeline;  # Added
        tx_intercept @safelinex;  # Added
        ...
    }
}
```

### Fix 2: Add Internal Named Locations

Added the missing internal locations to the site config:

```nginx
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
```

### Fix 3: Update Backend Upstream

Changed the backend upstream from localhost to the Docker gateway IP:

**Before:**
```nginx
upstream backend_3 {
    server 127.0.0.1:8080;
    ...
}
```

**After:**
```nginx
upstream backend_3 {
    server 172.22.222.1:8080;
    ...
}
```

The IP `172.22.222.1` is the gateway of the SafeLine Docker network (`safeline-ce`), which routes to the host machine.

### Fix 4: Update Host Nginx to Listen on All Interfaces

The host nginx was only listening on `127.0.0.1:8080`. Changed it to listen on all interfaces:

**File:** `/etc/nginx/sites-available/amss.leoulgirma.com`

```nginx
server {
    listen 0.0.0.0:8080;  # Changed from 127.0.0.1:8080
    server_name amss.leoulgirma.com;
    ...
}
```

Then restarted nginx:
```bash
sudo systemctl restart nginx
```

---

## Files Modified

| File | Change |
|------|--------|
| `/data/safeline/resources/nginx/sites-enabled/IF_backend_3` | Added t1k_intercept directives and @safeline locations |
| `/data/safeline/resources/nginx/safeline_tcp.conf` | Updated upstream configuration |
| `/etc/nginx/sites-available/amss.leoulgirma.com` | Changed listen address to 0.0.0.0:8080 |
| Database `mgt_website` table | Updated upstreams to use 172.22.222.1:8080 |

---

## Verification

After applying the fixes, the WAF was tested with various attack patterns:

| Test | Expected | Result |
|------|----------|--------|
| Normal request | HTTP 200 | ✅ PASS |
| SQL Injection (`?id=1' OR '1'='1`) | HTTP 403 | ✅ BLOCKED |
| XSS Attack (`?q=<script>alert(1)</script>`) | HTTP 403 | ✅ BLOCKED |
| Path Traversal (`?file=../../../etc/passwd`) | HTTP 403 | ✅ BLOCKED |
| Command Injection (`?cmd=;cat /etc/passwd`) | HTTP 403 | ✅ BLOCKED |

---

## Prevention

To prevent this issue from recurring:

1. **Backup the site config** before any SafeLine management operations:
   ```bash
   cp /data/safeline/resources/nginx/sites-enabled/IF_backend_3 \
      /data/safeline/resources/nginx/sites-enabled/IF_backend_3.backup
   ```

2. **Monitor for config regeneration** - The SafeLine management service may regenerate configs and remove custom additions. If the WAF stops working, check if the `@safeline*` locations are still present.

3. **Verify after updates** - After upgrading SafeLine, verify the WAF is still functional by testing with a known attack pattern.

---

## Useful Commands

### Check detector stats
```bash
sudo docker exec safeline-mgt curl -s http://safeline-detector:8001/stat | jq '{total: .req_num_total, blocked: .req_num_blocked, attack: .req_num_attack}'
```

### Test WAF is blocking attacks
```bash
curl -s -o /dev/null -w '%{http_code}\n' -H "Host: amss.leoulgirma.com" \
  "http://localhost:80/?id=1'%20OR%20'1'='1"
# Should return 403
```

### Reload tengine nginx
```bash
sudo docker exec safeline-tengine nginx -t && sudo docker exec safeline-tengine nginx -s reload
```

### Check site config has required locations
```bash
sudo docker exec safeline-tengine grep -c "location @safeline" /etc/nginx/sites-enabled/IF_backend_3
# Should return 4 (for @safeline, @safelinex, @safeline_chaos, @safeline_wr)
```

---

## Architecture Overview

```
                    Internet
                        │
                        ▼
              ┌─────────────────┐
              │   SafeLine      │
              │   Tengine       │ Port 80/443
              │   (nginx+t1k)   │
              └────────┬────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
         ▼             ▼             ▼
    ┌─────────┐  ┌──────────┐  ┌─────────┐
    │@safeline│  │@safelinex│  │ proxy   │
    │  t1k    │  │   tx     │  │  pass   │
    └────┬────┘  └────┬─────┘  └────┬────┘
         │            │             │
         ▼            ▼             │
    ┌─────────────────────┐        │
    │  SafeLine Detector  │        │
    │  (port 8000)        │        │
    └─────────────────────┘        │
                                   ▼
                          ┌───────────────┐
                          │ Backend Server│
                          │ (172.22.222.1 │
                          │  :8080)       │
                          └───────────────┘
```

The t1k module intercepts requests, sends them to the detector for analysis, and based on the response either allows the request to proceed to the backend or returns a 403 Forbidden response.

---

## Incident 2: SSL Certificate Loss After Tengine Restart (January 29, 2026)

### Problem

After restarting the `safeline-tengine` container (which had been stopped because `safeline-detector` was down for 7 days), the SSL configuration was lost. HTTPS requests to `amss.leoulgirma.com` failed with SSL errors.

### Root Cause

SafeLine's management service regenerated the tengine site config (`IF_backend_3`) on container startup but did not include the SSL directives, even though the certificate was present in the database (`mgt_ssl_cert` table).

The regenerated config had:
```nginx
listen 0.0.0.0:443;  # Missing 'ssl' keyword
# No ssl_certificate or ssl_certificate_key directives
```

### Fix

Manually added SSL directives inside the tengine container:

```bash
sudo docker exec safeline-tengine sed -i \
  's|listen 0.0.0.0:443;|listen 0.0.0.0:443 ssl;\nssl_certificate /etc/nginx/certs/amss.crt;\nssl_certificate_key /etc/nginx/certs/amss.key;|' \
  /etc/nginx/sites-enabled/IF_backend_3

sudo docker exec safeline-tengine nginx -t && \
  sudo docker exec safeline-tengine nginx -s reload
```

### Prevention

After any SafeLine container restart, verify SSL directives are present:
```bash
sudo docker exec safeline-tengine grep -A2 "listen.*443" /etc/nginx/sites-enabled/IF_backend_3
# Should show: listen 0.0.0.0:443 ssl;
#              ssl_certificate /etc/nginx/certs/amss.crt;
#              ssl_certificate_key /etc/nginx/certs/amss.key;
```

---

## Incident 3: API Domain Unreachable / Port Conflict (February 1, 2026)

### Problem

The AMSS frontend login page showed "Failed to look up email" because the backend API at `amss-api-uat.duckdns.org` was unreachable. SafeLine returned a "Website Not Found" error page for requests to that domain.

### Root Cause

Two issues combined:

1. **Port conflict:** When `safeline-tengine` was restarted (Incident 2), it grabbed ports 80/443 via Docker port mapping. The Kubernetes ingress-nginx controller, which previously owned these ports via `hostPort`, could no longer bind to them.

2. **Missing SafeLine site:** `amss-api-uat.duckdns.org` was never configured as a site in SafeLine. Previously, ingress-nginx handled this domain directly on ports 80/443. With SafeLine now owning those ports, the domain had no route.

### Fix

Added `amss-api-uat.duckdns.org` as a second site in SafeLine, proxying to the ingress-nginx HTTPS NodePort:

1. **Extracted TLS cert** from Kubernetes secret `amss-uat-tls`:
   ```bash
   kubectl get secret amss-uat-tls -n amss-uat -o jsonpath='{.data.tls\.crt}' | base64 -d > api-tls.crt
   kubectl get secret amss-uat-tls -n amss-uat -o jsonpath='{.data.tls\.key}' | base64 -d > api-tls.key
   ```

2. **Added cert to SafeLine database** (`mgt_ssl_cert` table, id=2)

3. **Added website to SafeLine database** (`mgt_website` table, id=4):
   - `server_names`: `["amss-api-uat.duckdns.org"]`
   - `upstreams`: `["https://172.22.222.1:30443"]` (K8s ingress-nginx HTTPS NodePort)
   - `cert_id`: 2

4. **Configured tengine site** (`IF_backend_4`):
   - SSL termination with the extracted Let's Encrypt cert
   - `proxy_pass https://backend_4` to the ingress-nginx HTTPS NodePort
   - WebSocket support (Upgrade/Connection headers via `proxy_params`)
   - Long proxy timeouts (3600s) for WebSocket connections
   - WAF detection blocks (`@safeline`, `@safelinex`)

5. **Reloaded tengine:**
   ```bash
   sudo docker exec safeline-tengine nginx -t && \
     sudo docker exec safeline-tengine nginx -s reload
   ```

### New Traffic Flow

```
Internet → SafeLine (port 443, SSL termination)
  ├── amss.leoulgirma.com → host nginx (172.22.222.1:8080) → static files
  └── amss-api-uat.duckdns.org → ingress-nginx NodePort (172.22.222.1:30443) → amss-server pod
```

### Verification

```bash
# API health
curl -sk https://amss-api-uat.duckdns.org/health
# Expected: ok

# Auth lookup
curl -sk -X POST -H "Content-Type: application/json" \
  -d '{"email":"tenant-admin@demo.local"}' \
  https://amss-api-uat.duckdns.org/api/v1/auth/lookup
# Expected: {"organizations":[...]}

# CORS preflight
curl -sk -X OPTIONS \
  -H "Origin: https://amss.leoulgirma.com" \
  -H "Access-Control-Request-Method: POST" \
  https://amss-api-uat.duckdns.org/api/v1/auth/lookup
# Expected: Access-Control-Allow-Origin: https://amss.leoulgirma.com
```

---

## Current Architecture (as of February 2026)

```
                              Internet
                                 │
                    ┌────────────┴────────────┐
                    │                         │
          amss.leoulgirma.com     amss-api-uat.duckdns.org
                    │                         │
                    ▼                         ▼
          ┌───────────────────────────────────────────┐
          │           SafeLine WAF (Tengine)           │
          │         Ports 80/443 (Docker)              │
          │  t1k module → safeline-detector (8000)     │
          └──────────┬────────────────────┬────────────┘
                     │                    │
      IF_backend_3   │                    │  IF_backend_4
                     ▼                    ▼
          ┌──────────────────┐  ┌─────────────────────────┐
          │  Host Nginx      │  │  K8s ingress-nginx      │
          │  (port 8080)     │  │  (NodePort 30443/HTTPS) │
          └────────┬─────────┘  └────────────┬────────────┘
                   │                         │
                   ▼                         ▼
          ┌──────────────────┐  ┌─────────────────────────┐
          │  /var/www/amss/  │  │  amss-server Pod        │
          │  ├── React SPA   │  │  (Go API, port 8080)    │
          │  └── marketing/  │  └────────────┬────────────┘
          │      (Astro)     │               │
          └──────────────────┘               ▼
                                ┌─────────────────────────┐
                                │    PostgreSQL + Redis    │
                                │    (Docker on host)      │
                                └─────────────────────────┘
```

### SafeLine Sites

| Site | Config File | Domain | Upstream | SSL Cert |
|------|------------|--------|----------|----------|
| 3 | IF_backend_3 | amss.leoulgirma.com | http://172.22.222.1:8080 | /etc/nginx/certs/amss.crt (cert_id=1) |
| 4 | IF_backend_4 | amss-api-uat.duckdns.org | https://172.22.222.1:30443 | /etc/nginx/certs/api.crt (cert_id=2) |
