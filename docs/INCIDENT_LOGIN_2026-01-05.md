# AMSS Login Failure Investigation (2026-01-05)

This document records the investigation and remediation steps taken to fix AMSS
login failures and frontend email lookup errors in UAT.

## Summary

- Login requests were returning HTTP 500 due to Redis write failures in the rate
  limiter.
- The Redis instance configured in UAT was a read-only replica with a downed
  master link.
- Added logging to capture Redis errors on login rate-limit checks.
- Implemented Redis Sentinel support and set up local Redis HA components.
- Fixed CORS preflight failures blocking frontend `/auth/lookup` calls.
- Added a frontend test for the lookup error path.

## Root Cause

- `REDIS_ADDR` pointed to `51.79.85.92:6379` which was a replica:
  - `role:slave`
  - `master_link_status:down`
  - `slave_read_only:1`
- The rate limiter uses `INCR`, which fails on read-only replicas:
  - `READONLY You can't write against a read only replica.`
- Frontend `auth/lookup` failed due to CORS preflight returning 405 with missing
  `CORS_ALLOWED_ORIGINS` and missing browser client hint headers.

## Backend Code Changes

- Login rate-limit error logging (Redis error, request_id, client IP, hashed ID):
  - `internal/api/rest/handlers/auth.go`
- Pass logger into AuthHandler:
  - `internal/api/rest/router.go`
- Redis Sentinel support in config and clients:
  - `internal/config/config.go`
  - `cmd/server/main.go`
  - `cmd/worker/main.go`
- CORS header allowlist expanded for browser client hints:
  - `internal/api/rest/router.go`

## Helm / Config Changes

- Added Sentinel config to configmap:
  - `deploy/helm/amss/templates/configmap.yaml`
  - `deploy/helm/amss/values.yaml`
  - `deploy/helm/values-uat.yaml`
- Set UAT CORS origins:
  - `deploy/helm/values-uat.yaml`
  - `corsAllowedOrigins: "https://amss.leoulgirma.com,https://amss-api-uat.duckdns.org,http://51.79.85.92:5173,http://localhost:5173"`

## Redis / HA (Host)

- Promoted local Redis to master:
  - `docker exec amss-redis redis-cli REPLICAOF NO ONE`
- Added replica:
  - container `amss-redis-replica-1` on port 6380
- Added three Sentinels (host network):
  - configs:
    - `/home/ubuntu/redis-sentinel/1/sentinel.conf`
    - `/home/ubuntu/redis-sentinel/2/sentinel.conf`
    - `/home/ubuntu/redis-sentinel/3/sentinel.conf`
  - containers:
    - `amss-redis-sentinel-1` (port 26379)
    - `amss-redis-sentinel-2` (port 26380)
    - `amss-redis-sentinel-3` (port 26381)

## Images / Deployments

- `leoulgirma/amss-server:rate-limit-log-20260105073408`
  - Adds login rate-limit error logging.
- `leoulgirma/amss-server:sentinel-20260105074641`
- `leoulgirma/amss-worker:sentinel-20260105074641`
  - Adds Redis Sentinel support to server and worker.
- `leoulgirma/amss-server:cors-20260105080601`
  - Adds CORS header allowlist updates.

UAT Helm upgrades were performed with:

```bash
helm upgrade amss deploy/helm/amss -n amss-uat -f deploy/helm/values-uat.yaml \
  --set image.server.tag=<tag> \
  --set image.worker.tag=<tag>
```

## Frontend Test Added

Added a test that validates the lookup error message when `/auth/lookup` fails:

- `amss-frontend/src/features/auth/login-page.test.tsx`

Run:

```bash
npm test -- --run amss-frontend/src/features/auth/login-page.test.tsx
```

## Verification

- Login succeeded with valid credentials and correct org_id.
- CORS preflight now returns 200 with allowed origins and headers.
- `/auth/lookup` returns 200 from the frontend origin.

## Operational Note

The Helm values for UAT do not include secrets. Any `helm upgrade` without
passing `secrets.*` will wipe `amss-secret`, causing CrashLoopBackOff. Secrets
were restored after upgrades during this incident.

