# SafeLine WAF Setup and Management Guide

> Web Application Firewall for AMSS - Aviation Maintenance & Safety System

## Table of Contents
1. [Overview](#overview)
2. [Current Setup](#current-setup)
3. [Accessing SafeLine Admin Panel](#accessing-safeline-admin-panel)
4. [Adding Your Website](#adding-your-website)
5. [SSL/HTTPS Configuration](#sslhttps-configuration)
6. [WAF Rules and Protection](#waf-rules-and-protection)
7. [Management Commands](#management-commands)
8. [Troubleshooting](#troubleshooting)
9. [Uninstalling SafeLine](#uninstalling-safeline)

---

## Overview

SafeLine is a self-hosted Web Application Firewall (WAF) that protects your web applications from:
- SQL Injection attacks
- Cross-Site Scripting (XSS)
- Command Injection
- Path Traversal
- Bot attacks
- And many more OWASP Top 10 vulnerabilities

**Architecture:**
```
Internet → SafeLine WAF (port 80/443) → Nginx (port 8080) → Static Files
```

---

## Current Setup

### Installation Details
| Component | Value |
|-----------|-------|
| Installation Directory | `/data/safeline` |
| SafeLine Version | Latest |
| Management Port | 9443 (HTTPS) |
| Nginx Backend Port | 8080 (internal only) |
| Database | PostgreSQL (containerized) |

### Running Containers
```bash
# Check container status
sudo docker ps --filter "name=safeline"
```

Expected containers:
- `safeline-tengine` - Nginx-based traffic handler
- `safeline-mgt` - Management server
- `safeline-pg` - PostgreSQL database
- `safeline-detector` - Attack detection engine
- `safeline-fvm` - Feature vector module
- `safeline-luigi` - Task scheduler
- `safeline-chaos` - Dynamic protection

---

## Accessing SafeLine Admin Panel

### URL
```
https://amss.leoulgirma.com:9443
```

Or locally:
```
https://localhost:9443
```

### Default Credentials
| Field | Value |
|-------|-------|
| Username | `admin` |
| Password | `njSSmpQU` |

**IMPORTANT:** Change this password immediately after first login!

### Reset Admin Password
If you forget the password:
```bash
sudo docker exec safeline-mgt /app/mgt-cli reset-admin --once
```

---

## Adding Your Website

### Step 1: Login to SafeLine Admin Panel
1. Open `https://amss.leoulgirma.com:9443`
2. Login with admin credentials
3. Change your password if prompted

### Step 2: Add a New Site
1. Click **"Sites"** or **"Protected Sites"** in the sidebar
2. Click **"Add Site"** or **"+"** button
3. Fill in the configuration:

| Field | Value |
|-------|-------|
| Domain | `amss.leoulgirma.com` |
| Listen Port | `80` (HTTP) or `443` (HTTPS) |
| Upstream Server | `127.0.0.1:8080` |
| Protocol | HTTP |

4. Click **"Save"** or **"Add"**

### Step 3: Verify Configuration
After adding the site, SafeLine will:
1. Generate nginx configuration
2. Start listening on the specified port
3. Begin forwarding traffic to your backend

Test it:
```bash
curl -I http://amss.leoulgirma.com
```

---

## SSL/HTTPS Configuration

### Option 1: Let SafeLine Handle SSL (Recommended)
1. In SafeLine admin panel, edit your site
2. Set Listen Port to `443`
3. Enable HTTPS
4. Either:
   - Upload your SSL certificate and key
   - Or use SafeLine's built-in Let's Encrypt integration

### Option 2: Use Existing Certificates
If you have Let's Encrypt certificates:
```bash
# Copy certificates to SafeLine
sudo cp /etc/letsencrypt/live/amss.leoulgirma.com/fullchain.pem /data/safeline/resources/nginx/certs/
sudo cp /etc/letsencrypt/live/amss.leoulgirma.com/privkey.pem /data/safeline/resources/nginx/certs/
```

Then configure in the admin panel.

### Option 3: Get New Let's Encrypt Certificate via SafeLine
SafeLine can automatically obtain and renew Let's Encrypt certificates through its admin panel.

---

## WAF Rules and Protection

### Protection Modes
| Mode | Description |
|------|-------------|
| **Detect** | Log attacks but don't block (for testing) |
| **Protect** | Block detected attacks |

### Built-in Protections
- SQL Injection
- XSS (Cross-Site Scripting)
- Command Injection
- Path Traversal
- XXE (XML External Entity)
- SSRF (Server-Side Request Forgery)
- Malicious File Upload

### Configuring Rules
1. Go to **"Protection Settings"** or **"WAF Rules"**
2. Enable/disable specific rule categories
3. Set sensitivity levels
4. Add custom whitelist rules if needed

### Rate Limiting
1. Go to **"Rate Limiting"** or **"Traffic Control"**
2. Set requests per second/minute limits
3. Configure blocking duration

---

## Management Commands

### Start SafeLine
```bash
cd /data/safeline
sudo docker compose up -d
```

### Stop SafeLine
```bash
cd /data/safeline
sudo docker compose down
```

### Restart SafeLine
```bash
cd /data/safeline
sudo docker compose restart
```

### Restart Specific Container
```bash
sudo docker restart safeline-tengine
sudo docker restart safeline-mgt
```

### View Logs
```bash
# All container logs
sudo docker compose -f /data/safeline/compose.yaml logs -f

# Specific container logs
sudo docker logs -f safeline-tengine
sudo docker logs -f safeline-mgt
sudo docker logs -f safeline-detector

# Nginx access/error logs
sudo tail -f /data/safeline/logs/nginx/access.log
sudo tail -f /data/safeline/logs/nginx/error.log
```

### Check Container Status
```bash
sudo docker ps --filter "name=safeline" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### Update SafeLine
```bash
cd /data/safeline
sudo docker compose pull
sudo docker compose up -d
```

### Reset Admin Password
```bash
sudo docker exec safeline-mgt /app/mgt-cli reset-admin --once
```

### Clear Attack Statistics
```bash
sudo docker exec safeline-mgt /app/mgt-cli clear-statistics
```

### Clear Logs
```bash
sudo docker exec safeline-mgt /app/mgt-cli clear-log
```

---

## Troubleshooting

### Port 80 Already in Use
If port 80 is occupied:
```bash
# Check what's using port 80
sudo ss -tlnp | grep :80

# Stop the conflicting service
sudo systemctl stop nginx  # if nginx is on port 80
```

### SafeLine Not Starting
```bash
# Check container logs
sudo docker logs safeline-mgt
sudo docker logs safeline-tengine

# Restart all containers
cd /data/safeline
sudo docker compose down
sudo docker compose up -d
```

### Website Not Accessible
1. Verify site is added in SafeLine admin panel
2. Check tengine is listening:
   ```bash
   sudo ss -tlnp | grep :80
   ```
3. Check backend is accessible:
   ```bash
   curl -I http://127.0.0.1:8080
   ```
4. Check SafeLine logs for errors:
   ```bash
   sudo docker logs safeline-tengine
   ```

### High Memory Usage
```bash
# Check container resource usage
sudo docker stats --no-stream

# Restart containers if needed
cd /data/safeline
sudo docker compose restart
```

### Legitimate Requests Being Blocked
1. Check attack logs in SafeLine admin panel
2. Identify the rule causing false positives
3. Add a whitelist rule for the specific path/IP
4. Or adjust rule sensitivity

---

## Uninstalling SafeLine

### Stop and Remove Containers
```bash
cd /data/safeline
sudo docker compose down -v
```

### Remove Data (Optional)
```bash
sudo rm -rf /data/safeline
```

### Restore Original Nginx Config
```bash
# Update nginx to listen on port 80 again
sudo nano /etc/nginx/sites-available/amss.leoulgirma.com
# Change: listen 127.0.0.1:8080; → listen 80;

sudo nginx -t && sudo systemctl reload nginx
```

---

## File Locations

| File/Directory | Purpose |
|----------------|---------|
| `/data/safeline/` | Main SafeLine directory |
| `/data/safeline/compose.yaml` | Docker Compose configuration |
| `/data/safeline/.env` | Environment variables |
| `/data/safeline/resources/nginx/` | Nginx configuration |
| `/data/safeline/logs/nginx/` | Access and error logs |
| `/data/safeline/resources/postgres/` | Database data |

---

## Configuration Files

### .env File
Location: `/data/safeline/.env`
```env
SAFELINE_DIR=/data/safeline
IMAGE_TAG=latest
MGT_PORT=9443
POSTGRES_PASSWORD=5a8394d1707689f56d3b33edf7c89be7
SUBNET_PREFIX=172.22.222
IMAGE_PREFIX=chaitin
ARCH_SUFFIX=
RELEASE=
REGION=-g
MGT_PROXY=0
```

### Nginx Backend Configuration
Location: `/etc/nginx/sites-available/amss.leoulgirma.com`
```nginx
server {
    listen 127.0.0.1:8080;
    server_name amss.leoulgirma.com;
    root /var/www/amss;
    # ... rest of config
}
```

---

## Security Recommendations

1. **Change default password** immediately after installation
2. **Restrict management port** (9443) access to trusted IPs only
3. **Enable HTTPS** for your protected sites
4. **Review attack logs** regularly
5. **Keep SafeLine updated** for latest security patches
6. **Backup configuration** before making changes

### Restrict Management Access (Optional)
Add iptables rule to limit 9443 access:
```bash
# Allow only from specific IP
sudo iptables -A INPUT -p tcp --dport 9443 -s YOUR_IP -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 9443 -j DROP
```

---

## Quick Reference

| Action | Command |
|--------|---------|
| Start SafeLine | `cd /data/safeline && sudo docker compose up -d` |
| Stop SafeLine | `cd /data/safeline && sudo docker compose down` |
| Restart | `cd /data/safeline && sudo docker compose restart` |
| View logs | `sudo docker logs -f safeline-tengine` |
| Reset password | `sudo docker exec safeline-mgt /app/mgt-cli reset-admin --once` |
| Update | `cd /data/safeline && sudo docker compose pull && sudo docker compose up -d` |
| Check status | `sudo docker ps --filter "name=safeline"` |

---

## Known Issues

### Detector Not Receiving Requests (v9.3.0)

**Status:** Unresolved

There is a known issue where the SafeLine WAF detector does not receive requests from tengine, resulting in:
- `req_num_total: 0` in detector statistics
- Empty detection logs (`mgt_detect_log_basic` table)
- Attacks passing through unblocked

**Suspected Cause:** The `network_mode: host` configuration for tengine may be incompatible with the Unix socket or TCP communication to the detector container on the bridge network.

**Attempted Solutions:**
1. Unix socket communication (default) - Failed
2. TCP communication (172.22.222.5:8000) - Failed
3. Site reconfiguration - Failed
4. Container restarts - Failed

**Full Details:** See `/home/ubuntu/SAFELINE-TROUBLESHOOTING-REPORT.md`

**Workaround Options:**
1. Fresh reinstall without host network mode
2. Use alternative WAF integration (lua-resty-t1k, Traefik plugin)
3. Report to SafeLine team for fix

---

## Resources

- [SafeLine GitHub](https://github.com/chaitin/SafeLine)
- [Official Documentation](https://docs.waf.chaitin.com/en/)
- [Discord Community](https://discord.gg/SVnZGzHFvn)
- [Troubleshooting Report](/home/ubuntu/SAFELINE-TROUBLESHOOTING-REPORT.md)

---

*Last updated: January 12, 2026*
