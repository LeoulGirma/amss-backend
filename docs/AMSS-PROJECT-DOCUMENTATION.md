# AMSS - Aircraft Maintenance Scheduling System

## Project Documentation

**Domain:** amss.leoulgirma.com
**API Domain:** amss-api-uat.duckdns.org
**Server IP:** 51.79.85.92
**Last Updated:** February 2026

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Frontend](#frontend)
4. [Marketing Website](#marketing-website)
5. [Backend API](#backend-api)
6. [Database Layer](#database-layer)
7. [Kubernetes (K3s)](#kubernetes-k3s)
8. [Web Application Firewall (SafeLine)](#web-application-firewall-safeline)
9. [Infrastructure](#infrastructure)
10. [Deployment](#deployment)
11. [Monitoring & Maintenance](#monitoring--maintenance)

---

## Project Overview

AMSS (Aircraft Maintenance Scheduling System) is a full-stack web application for managing and scheduling aircraft maintenance operations. The system features a React 19 frontend, a Go API backend running on Kubernetes, a static marketing website built with Astro, and is protected by a SafeLine Web Application Firewall (WAF).

### Key Features

- Aircraft fleet management and status monitoring
- Maintenance task scheduling (Kanban + calendar views)
- Compliance and audit trail tracking (AD/SB, digital sign-offs)
- Parts inventory management
- Team and role-based access control (5 roles)
- Real-time dashboard with charts and analytics
- Real-time notifications via WebSocket
- Progressive Web App (PWA) capabilities
- Dark/Light theme support
- Demo mode for evaluation
- WAF protection against OWASP Top 10 attacks

---

## Architecture

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

### Traffic Flow

| Domain | SafeLine Site | Upstream | Serves |
|--------|--------------|----------|--------|
| `amss.leoulgirma.com` | IF_backend_3 | `http://172.22.222.1:8080` (host nginx) | React SPA + Marketing site |
| `amss-api-uat.duckdns.org` | IF_backend_4 | `https://172.22.222.1:30443` (K8s NodePort) | Go API + WebSocket |

---

## Frontend

### Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| **React** | 19.1.0 | UI Framework |
| **TypeScript** | 5.8.3 | Type safety |
| **Vite** | 7.3.0 | Build tool and dev server |
| **Redux Toolkit / RTK Query** | | State management and API data fetching |
| **React Router** | v7 | Client-side routing |
| **shadcn/ui (Radix UI)** | | Accessible component primitives |
| **Tailwind CSS** | 4.1 | Utility-first styling |
| **Recharts** | | Data visualization |
| **Vite PWA Plugin** | | Progressive Web App (Workbox) |

### Source Code

**Location:** `/home/ubuntu/amss-frontend/`

### Deployed Location

```
/var/www/amss/
├── index.html              # React SPA entry point
├── favicon.svg             # App icon
├── manifest.webmanifest    # PWA manifest
├── sw.js                   # Service Worker (Workbox)
├── assets/                 # JS/CSS bundles (code-split)
│   ├── react-vendor-*.js
│   ├── router-*.js
│   ├── radix-*.js
│   ├── redux-*.js
│   ├── charts-*.js
│   ├── icons-*.js
│   ├── date-*.js
│   └── forms-*.js
└── marketing/              # Astro marketing site (see below)
    ├── index.html
    ├── features/
    └── pricing/
```

### Features

- **Dark/Light Theme:** System preference detection with localStorage persistence
- **PWA Support:** Installable app with offline capabilities, Workbox service worker
- **Code Splitting:** Vendor chunks (react, router, radix, redux, charts, icons, date, forms) for optimal caching
- **Responsive Design:** Mobile-first approach
- **Demo Mode:** Full access with sample data, no backend required
- **Service Worker Denylist:** `/marketing` path excluded from SPA service worker interception via `navigateFallbackDenylist`

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/amss.leoulgirma.com
server {
    listen 0.0.0.0:8080;
    server_name amss.leoulgirma.com;

    absolute_redirect off;

    # Trust SafeLine proxy for real client IPs
    set_real_ip_from 127.0.0.1;
    set_real_ip_from 172.22.222.0/24;
    real_ip_header X-Forwarded-For;
    real_ip_recursive on;

    root /var/www/amss;
    index index.html;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/json application/xml;

    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    # Marketing static files and SPA fallback
    # $uri/index.html comes BEFORE $uri/ to avoid 301 directory redirects
    location / {
        try_files $uri $uri/index.html $uri/ /index.html;
    }

    # Cache static assets for 1 year
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Don't cache index.html
    location = /index.html {
        expires -1;
        add_header Cache-Control "no-store, no-cache, must-revalidate";
    }
}
```

**Key `try_files` detail:** `$uri/index.html` is listed before `$uri/` to serve the Astro-generated `index.html` files at `/marketing/`, `/marketing/features/`, `/marketing/pricing/` directly without triggering nginx 301 directory redirects.

---

## Marketing Website

### Overview

A 3-page static marketing/landing site for the AMSS product, built with Astro and served under the `/marketing` path on `amss.leoulgirma.com`.

### Technology Stack

| Technology | Purpose |
|------------|---------|
| **Astro 5.x** | Static site generator (zero JS by default) |
| **Tailwind CSS 4.x** | Styling via `@theme` tokens (Vite plugin) |
| **TypeScript** | Content data files |

### Source Code

**Location:** `/home/ubuntu/amss-marketing/`

### Pages

| Page | URL | Description |
|------|-----|-------------|
| Landing | `/marketing/` | Hero, pain points, feature cards, metrics, testimonials, CTA |
| Features | `/marketing/features` | 6 feature sections with detail, compliance badges |
| Pricing | `/marketing/pricing` | 3-tier pricing, feature comparison matrix, FAQ |

### Design

- Dark/premium theme (aerospace noir) matching the main AMSS app
- CSS-only scroll animations (IntersectionObserver)
- No JavaScript frameworks shipped to browser
- Astro `base: '/marketing/'` for path-prefixed deployment

### Deployment

Built static files are copied to `/var/www/amss/marketing/` and served by the same nginx instance as the React SPA. The SPA's Workbox service worker has a `navigateFallbackDenylist: [/^\/marketing/]` to prevent intercepting marketing page navigation.

---

## Backend API

### Overview

The AMSS backend is a Go API server running in Kubernetes (k3s). It serves RESTful HTTP endpoints on port 8080 and gRPC on port 9090.

**Source Code:** `/home/ubuntu/amss-backend/`
**Image:** `leoulgirma/amss-server:cors-20260105080601`
**API Domain:** `amss-api-uat.duckdns.org`

### Components

| Component | Namespace | Replicas | Description |
|-----------|-----------|----------|-------------|
| amss-server | amss-uat | 1 | HTTP API (8080) + gRPC (9090) |
| amss-worker | amss-uat | 1 | Background job processor |

### Key API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/lookup` | No | Look up organizations by email |
| POST | `/api/v1/auth/login` | No | Authenticate and get tokens |
| POST | `/api/v1/auth/refresh` | No | Refresh access token |
| GET | `/api/v1/auth/me` | JWT | Get current user info |
| GET | `/health` | No | Health check |
| GET | `/ready` | No | Readiness (checks DB + Redis) |
| GET | `/metrics` | No | Prometheus metrics |

### Configuration

Managed via Kubernetes ConfigMap (`amss-config`) and Secret (`amss-secret`):

| Key | Value |
|-----|-------|
| `HTTP_ADDR` | `:8080` |
| `GRPC_ADDR` | `:9090` |
| `APP_ENV` | `uat` |
| `CORS_ALLOWED_ORIGINS` | `https://amss.leoulgirma.com, https://amss-api-uat.duckdns.org` |
| `REDIS_SENTINELS` | `51.79.85.92:26379,51.79.85.92:26380,51.79.85.92:26381` |

---

## Database Layer

### PostgreSQL

Primary relational database for application data.

| Setting | Value |
|---------|-------|
| **Container** | amss-postgres |
| **Image** | postgres:16 |
| **Port** | 5455 (external) → 5432 (internal) |
| **Storage** | Docker volume |

### Redis (High Availability Setup)

Session management and caching with master-replica-sentinel architecture.

#### Topology

```
┌─────────────────┐     ┌────────────────────┐
│  Redis Master   │────▶│  Redis Replica     │
│  (amss-redis)   │     │  (amss-redis-      │
│  Port: 6379     │     │   replica-1)       │
└────────┬────────┘     │  Port: 6380        │
         │              └────────────────────┘
         │
    ┌────┴────┐
    ▼    ▼    ▼
┌──────┬──────┬──────┐
│ S1   │ S2   │ S3   │  Sentinels
└──────┴──────┴──────┘  (Failover)
```

#### Containers

| Container | Role | Port |
|-----------|------|------|
| amss-redis | Master | 6379 |
| amss-redis-replica-1 | Replica | 6380 |
| amss-redis-sentinel-1 | Sentinel | - |
| amss-redis-sentinel-2 | Sentinel | - |
| amss-redis-sentinel-3 | Sentinel | - |

---

## Kubernetes (K3s)

A lightweight Kubernetes distribution is installed for container orchestration.

### Configuration

**Config File:** `/home/ubuntu/.kube/config`
**API Server:** https://127.0.0.1:6443

### Key Files

```
/home/ubuntu/
├── .kube/
│   └── config                      # Cluster config
├── amss-servicemonitor.yaml        # Prometheus ServiceMonitor
└── prometheus-values.yaml          # Prometheus Helm values
```

### Monitoring Stack

- **Prometheus:** Metrics collection
- **ServiceMonitor:** Custom metrics scraping

---

## Web Application Firewall (SafeLine)

### Overview

SafeLine WAF by Chaitin protects the application against OWASP Top 10 vulnerabilities including SQL injection, XSS, path traversal, and command injection.

### Version

**SafeLine:** v9.2.0

### Components

| Container | Purpose | Port |
|-----------|---------|------|
| safeline-tengine | Nginx + t1k WAF module | 80, 443 |
| safeline-detector | Attack detection engine | 8000 |
| safeline-mgt | Management API & Dashboard | 9443 |
| safeline-pg | PostgreSQL for WAF | - |
| safeline-luigi | Log processing | - |
| safeline-chaos | Chaos engineering | - |
| safeline-fvm | File verification | - |

### Dashboard Access

| | |
|---|---|
| **URL** | https://51.79.85.92:9443 |
| **Username** | admin |

### Network Configuration

**Docker Network:** safeline-ce (172.22.222.0/24)

| Container | IP Address |
|-----------|------------|
| Gateway (Host) | 172.22.222.1 |
| Tengine | 172.22.222.3 |
| MGT | 172.22.222.4 |
| Detector | 172.22.222.5 |

### Protection Features

| Attack Type | Status |
|-------------|--------|
| SQL Injection | ✅ Blocked (HTTP 403) |
| XSS (Cross-Site Scripting) | ✅ Blocked (HTTP 403) |
| Path Traversal | ✅ Blocked (HTTP 403) |
| Command Injection | ✅ Blocked (HTTP 403) |
| CSRF | ✅ Protected |
| Rate Limiting | ✅ Configurable |

### Sites Configured

| Site ID | Domain | Upstream | SSL Cert |
|---------|--------|----------|----------|
| IF_backend_3 | `amss.leoulgirma.com` | `http://172.22.222.1:8080` (host nginx) | Let's Encrypt (cert_id=1) |
| IF_backend_4 | `amss-api-uat.duckdns.org` | `https://172.22.222.1:30443` (K8s NodePort) | Let's Encrypt (cert_id=2) |

### Configuration Files

```
/data/safeline/
├── .env                            # Environment config
├── docker-compose.yaml             # Container definitions
├── resources/
│   ├── nginx/
│   │   ├── nginx.conf
│   │   ├── safeline_tcp.conf      # Detector upstream config
│   │   └── sites-enabled/
│   │       ├── IF_backend_3       # amss.leoulgirma.com site config
│   │       └── IF_backend_4       # amss-api-uat.duckdns.org site config
│   └── detector/
│       └── detector.yml           # Detector settings
└── logs/
    ├── nginx/                     # Access/error logs
    └── detector/                  # Detection logs
```

### Known Issue: Config Regeneration

SafeLine's management service may regenerate tengine site configs and remove custom additions (SSL directives, WAF detection blocks). After restarting SafeLine containers, verify:
1. `ssl` directive is present on `listen 443` lines
2. `ssl_certificate` / `ssl_certificate_key` directives point to correct cert files
3. `@safeline` and `@safelinex` location blocks exist in each site config

---

## Infrastructure

### Server Specifications

| Spec | Value |
|------|-------|
| **Provider** | OVH VPS |
| **Hostname** | vps-9a02be74 |
| **OS** | Ubuntu 25.04 |
| **CPU** | 4 cores |
| **RAM** | 7.6 GB |
| **Public IP** | 51.79.85.92 |

### Port Mapping

| Port | Service | Bound By |
|------|---------|----------|
| 22 | SSH | sshd |
| 80 | SafeLine WAF (HTTP) | docker-proxy (safeline-tengine) |
| 443 | SafeLine WAF (HTTPS) | docker-proxy (safeline-tengine) |
| 5455 | PostgreSQL (firewalled) | Docker |
| 6379 | Redis Master (firewalled) | Docker |
| 6380 | Redis Replica (firewalled) | Docker |
| 6443 | Kubernetes API | k3s |
| 8080 | Host Nginx (frontend + marketing) | nginx |
| 9443 | SafeLine Dashboard | docker-proxy (safeline-mgt) |
| 30080 | ingress-nginx HTTP NodePort | k3s |
| 30443 | ingress-nginx HTTPS NodePort | k3s |

### DNS Configuration

| Record | Type | Value |
|--------|------|-------|
| amss.leoulgirma.com | A | 51.79.85.92 |
| amss-api-uat.duckdns.org | A | 51.79.85.92 |

---

## Deployment

### Frontend (React SPA)

```bash
cd /home/ubuntu/amss-frontend
npm install && npm run build
sudo cp -r dist/* /var/www/amss/
sudo systemctl reload nginx
```

### Marketing Website (Astro)

```bash
cd /home/ubuntu/amss-marketing
npm install && npm run build
sudo cp -r dist/* /var/www/amss/marketing/
# No nginx reload needed (static files)
```

### Backend (Kubernetes/Helm)

```bash
# Build and push Docker image
cd /home/ubuntu/amss-backend
docker build -t leoulgirma/amss-server:<tag> .
docker push leoulgirma/amss-server:<tag>

# Deploy via Helm
helm upgrade amss ./charts/amss -n amss-uat \
  -f charts/amss/values-uat.yaml \
  --set server.image.tag=<tag>
```

### SafeLine WAF

```bash
# Start SafeLine
cd /data/safeline && docker compose up -d

# Stop SafeLine
cd /data/safeline && docker compose down

# View logs
docker logs safeline-tengine
docker logs safeline-detector

# Reload tengine config (after manual edits)
sudo docker exec safeline-tengine nginx -t && \
  sudo docker exec safeline-tengine nginx -s reload
```

---

## Monitoring & Maintenance

### Health Checks

```bash
# Check WAF is blocking attacks
curl -sk -o /dev/null -w '%{http_code}\n' \
  "https://amss.leoulgirma.com/?id=1'%20OR%20'1'='1"
# Expected: 403

# Check API health
curl -sk https://amss-api-uat.duckdns.org/health
# Expected: ok

# Check API readiness (verifies DB + Redis)
curl -sk https://amss-api-uat.duckdns.org/ready
# Expected: ready

# Check detector stats
docker exec safeline-mgt curl -s http://safeline-detector:8001/stat | jq

# Check Kubernetes pods
sudo KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl get pods -A

# Check Docker containers
docker ps

# Check host nginx status
sudo systemctl status nginx
```

### Log Locations

| Service | Log Location |
|---------|--------------|
| SafeLine Nginx | `/data/safeline/logs/nginx/` |
| SafeLine Detector | `/data/safeline/logs/detector/` |
| Host Nginx | `/var/log/nginx/` |
| K8s API Server | `sudo KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl logs -n amss-uat <pod>` |

### Backup Recommendations

1. **Database:** Daily PostgreSQL dumps
2. **Redis:** RDB snapshots enabled
3. **Configuration:** `/data/safeline/` and `/etc/nginx/`
4. **Application:** Git repositories (`amss-backend`, `amss-frontend`, `amss-marketing`)

### Common Issues & Solutions

| Issue | Solution |
|-------|----------|
| WAF blocking legitimate requests | Check SafeLine dashboard at `https://51.79.85.92:9443`, adjust rules |
| API "Website Not Found" from SafeLine | Verify `IF_backend_4` config has SSL directives; reload tengine |
| SafeLine lost SSL after restart | Re-add `ssl`, `ssl_certificate`, `ssl_certificate_key` to site configs |
| React Router 404 on marketing pages | Service worker is intercepting; verify `navigateFallbackDenylist` in `vite.config.ts` |
| kubectl: "invalid character '<'" | Use explicit kubeconfig: `sudo KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl ...` |
| 502 Bad Gateway on API | Check `amss-server` pod status, logs; verify NodePort 30443 is accessible |
| Frontend not updating | Clear browser cache/service worker; hard refresh |
| Redis connection failed | Check sentinel status, restart Redis cluster |

---

## State Machine Diagrams

### 1. WAF Request Processing State Machine

```
                              ┌─────────────────┐
                              │   NEW REQUEST   │
                              └────────┬────────┘
                                       │
                                       ▼
                              ┌─────────────────┐
                              │  t1k_intercept  │
                              │  (Capture Req)  │
                              └────────┬────────┘
                                       │
                                       ▼
                              ┌─────────────────┐
                              │  Forward to     │
                              │  @safeline      │
                              └────────┬────────┘
                                       │
                                       ▼
                              ┌─────────────────┐
                              │   Detector      │
                              │   Analysis      │
                              └────────┬────────┘
                                       │
                    ┌──────────────────┼──────────────────┐
                    │                  │                  │
                    ▼                  ▼                  ▼
           ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
           │    CLEAN     │   │   ATTACK     │   │  RATE LIMIT  │
           │  (Score: 0)  │   │ (Score: >70) │   │  (Too Fast)  │
           └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
                  │                  │                  │
                  ▼                  ▼                  ▼
           ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
           │  PASS (200)  │   │ BLOCK (403)  │   │  ACL (429)   │
           │  → Backend   │   │ → Error Page │   │ → Error Page │
           └──────────────┘   └──────────────┘   └──────────────┘
```

### 2. Redis Sentinel Failover State Machine

```
                    ┌─────────────────────────────────────┐
                    │           NORMAL STATE              │
                    │  Master: amss-redis (6379)          │
                    │  Replica: amss-redis-replica-1      │
                    └─────────────────┬───────────────────┘
                                      │
                                      │ Master becomes
                                      │ unreachable
                                      ▼
                    ┌─────────────────────────────────────┐
                    │         SUBJECTIVE DOWN             │
                    │  (Single Sentinel detects failure)  │
                    └─────────────────┬───────────────────┘
                                      │
                                      │ Quorum agrees
                                      │ (2+ Sentinels)
                                      ▼
                    ┌─────────────────────────────────────┐
                    │          OBJECTIVE DOWN             │
                    │  (Consensus: Master is down)        │
                    └─────────────────┬───────────────────┘
                                      │
                                      │ Leader election
                                      │ among Sentinels
                                      ▼
                    ┌─────────────────────────────────────┐
                    │         FAILOVER START              │
                    │  (Leader Sentinel coordinates)      │
                    └─────────────────┬───────────────────┘
                                      │
                                      │ Promote replica
                                      │ to master
                                      ▼
                    ┌─────────────────────────────────────┐
                    │        FAILOVER COMPLETE            │
                    │  New Master: amss-redis-replica-1   │
                    │  Old Master: becomes replica        │
                    └─────────────────┬───────────────────┘
                                      │
                                      │ Clients notified
                                      │ via Sentinel
                                      ▼
                    ┌─────────────────────────────────────┐
                    │         NEW NORMAL STATE            │
                    │  Master: amss-redis-replica-1       │
                    │  Replica: amss-redis (when online)  │
                    └─────────────────────────────────────┘
```

### 3. User Session Lifecycle

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│     ┌───────────┐                                                   │
│     │  VISITOR  │                                                   │
│     │(Anonymous)│                                                   │
│     └─────┬─────┘                                                   │
│           │                                                         │
│           │ Login Request                                           │
│           ▼                                                         │
│     ┌───────────┐     ┌───────────┐                                │
│     │ VALIDATE  │────▶│  INVALID  │──────┐                         │
│     │CREDENTIALS│fail │CREDENTIALS│      │                         │
│     └─────┬─────┘     └───────────┘      │                         │
│           │                              │                         │
│           │ success                      │ Max retries             │
│           ▼                              ▼                         │
│     ┌───────────┐                  ┌───────────┐                   │
│     │  CREATE   │                  │  LOCKED   │                   │
│     │  SESSION  │                  │  ACCOUNT  │                   │
│     └─────┬─────┘                  └───────────┘                   │
│           │                                                         │
│           │ Store in Redis                                          │
│           ▼                                                         │
│     ┌───────────┐                                                   │
│     │AUTHENTICATED│◀─────────────────────────────┐                 │
│     │  (Active)   │                              │                 │
│     └─────┬───────┘                              │                 │
│           │                                      │                 │
│     ┌─────┴─────────────┬────────────────┐      │                 │
│     │                   │                │      │                 │
│     ▼                   ▼                ▼      │                 │
│ ┌─────────┐      ┌───────────┐    ┌──────────┐ │                 │
│ │ LOGOUT  │      │  TIMEOUT  │    │  REFRESH │ │                 │
│ │(Manual) │      │ (Expired) │    │  TOKEN   │─┘                 │
│ └────┬────┘      └─────┬─────┘    └──────────┘                   │
│      │                 │                                          │
│      ▼                 ▼                                          │
│ ┌─────────────────────────┐                                       │
│ │    SESSION DESTROYED    │                                       │
│ │   (Redis key deleted)   │                                       │
│ └───────────┬─────────────┘                                       │
│             │                                                      │
│             ▼                                                      │
│       ┌───────────┐                                                │
│       │  VISITOR  │                                                │
│       │(Anonymous)│                                                │
│       └───────────┘                                                │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4. Application Deployment Pipeline

```
┌─────────────────────────────────────────────────────────────────────┐
│                     DEPLOYMENT STATE MACHINE                        │
│                                                                     │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐  │
│  │   CODE   │────▶│  BUILD   │────▶│   TEST   │────▶│  DEPLOY  │  │
│  │  COMMIT  │     │  (Vite)  │     │  (Jest)  │     │  (SCP)   │  │
│  └──────────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘  │
│                        │                │                │         │
│                        │ fail           │ fail           │ fail    │
│                        ▼                ▼                ▼         │
│                   ┌─────────────────────────────────────────┐      │
│                   │              ROLLBACK                   │      │
│                   │         (Restore previous)              │      │
│                   └─────────────────────────────────────────┘      │
│                                                                     │
│  Deployment Steps:                                                  │
│  ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐  │
│  │  Upload  │────▶│  Backup  │────▶│  Deploy  │────▶│  Verify  │  │
│  │  Files   │     │  Old Ver │     │  New Ver │     │  Health  │  │
│  └──────────┘     └──────────┘     └──────────┘     └────┬─────┘  │
│                                                          │         │
│                        ┌─────────────────────────────────┤         │
│                        │                                 │         │
│                        ▼                                 ▼         │
│                   ┌──────────┐                    ┌──────────┐     │
│                   │  FAILED  │                    │ SUCCESS  │     │
│                   │(Rollback)│                    │ (Live!)  │     │
│                   └──────────┘                    └──────────┘     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Quick Reference

### URLs

- **Application:** https://amss.leoulgirma.com
- **Marketing Site:** https://amss.leoulgirma.com/marketing/
- **API:** https://amss-api-uat.duckdns.org
- **SafeLine Dashboard:** https://51.79.85.92:9443
- **Kubernetes API:** https://127.0.0.1:6443

### Important Files

```
/var/www/amss/                         # Frontend build (SPA + marketing)
/home/ubuntu/amss-frontend/            # Frontend source code
/home/ubuntu/amss-backend/             # Backend source code
/home/ubuntu/amss-marketing/           # Marketing site source code
/data/safeline/                        # SafeLine WAF
/etc/nginx/sites-available/            # Host nginx configs
/etc/rancher/k3s/k3s.yaml             # Kubernetes kubeconfig
```

### Service Management

```bash
# Host Nginx (serves frontend + marketing)
sudo systemctl restart nginx

# K8s Backend
sudo KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl rollout restart deployment/amss-server -n amss-uat

# SafeLine WAF
cd /data/safeline && docker compose restart

# Redis
docker restart amss-redis amss-redis-replica-1
```

---

## Documentation History

| Date | Change |
|------|--------|
| Feb 2026 | Added amss-api-uat.duckdns.org as SafeLine site (IF_backend_4) to fix API routing |
| Jan 2026 | Built and deployed Astro marketing website at /marketing path |
| Jan 2026 | Added service worker navigateFallbackDenylist for /marketing |
| Jan 2026 | Fixed SafeLine SSL cert loss after tengine container restart |
| Jan 2026 | Fixed SafeLine WAF t1k module configuration |
| Jan 2026 | Added @safeline locations to site config |
| Jan 2026 | Connected frontend to real backend API (compliance, kanban, notifications) |
| Jan 2026 | Fixed demo mode login loop |
| Jan 2026 | Set up HTTPS via SafeLine WAF with Let's Encrypt certs |
| Dec 2025 | Initial SafeLine WAF deployment |
| Dec 2025 | Redis HA cluster setup |
| Dec 2025 | K3s Kubernetes installation |
| Dec 2025 | Initial AMSS deployment |
