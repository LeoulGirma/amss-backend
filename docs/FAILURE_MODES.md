# AMSS Failure Modes & Recovery

**Last Updated:** December 24, 2025
**Environment:** Production & UAT
**Audience:** DevOps, SREs, On-Call Engineers

This document describes what happens when AMSS components fail, how to detect failures, and recovery procedures.

---

## Table of Contents

1. [Database Failures](#1-database-failures)
2. [Application Failures](#2-application-failures)
3. [Infrastructure Failures](#3-infrastructure-failures)
4. [Network & Connectivity Failures](#4-network--connectivity-failures)
5. [Data Integrity Issues](#5-data-integrity-issues)
6. [Backup & Recovery Procedures](#6-backup--recovery-procedures)
7. [Disaster Recovery](#7-disaster-recovery)
8. [Monitoring & Alerting](#8-monitoring--alerting)

---

## 1. Database Failures

### PostgreSQL Crash

**Failure Scenario:** PostgreSQL container crashes or becomes unresponsive

**Impact:**
- **Severity:** 🔴 **CRITICAL**
- API returns `503 Service Unavailable` on all write operations
- `/ready` health check fails
- `/health` endpoint still returns `200` (doesn't check DB)
- All ongoing requests fail with database connection errors
- Worker jobs fail (cannot read/write data)

**Detection:**
```bash
# Health check fails
curl https://amss-api-uat.duckdns.org/ready
# Response: 503

# Check container status
sudo docker ps | grep amss-postgres
# If not running: container crashed

# Check logs
sudo docker logs amss-postgres --tail=100
```

**Common Causes:**
1. Out of memory (Postgres killed by OOM killer)
2. Disk full (cannot write WAL logs)
3. Corrupted data files
4. Manual stop/restart

**Automatic Recovery:**
- Docker restart policy: `--restart unless-stopped`
- Container automatically restarts on crash
- **RTO (Recovery Time Objective):** 10-30 seconds

**Manual Recovery:**
```bash
# 1. Check container status
sudo docker ps -a | grep amss-postgres

# 2. Check logs for error
sudo docker logs amss-postgres --tail=50

# 3. Restart container
sudo docker restart amss-postgres

# 4. Verify recovery
sudo docker logs amss-postgres --tail=20 -f
# Wait for: "database system is ready to accept connections"

# 5. Test connectivity
psql "postgres://amss:amss@localhost:5455/amss" -c "SELECT 1;"

# 6. Verify API
curl https://amss-api-uat.duckdns.org/ready
# Should return: "ready"
```

**Data Loss Risk:**
- ✅ **None** (if crash) - PostgreSQL WAL ensures durability
- ⚠️ **Partial** (if disk corruption) - Last few seconds of writes may be lost

**Prevention:**
```bash
# Monitor disk usage
df -h /var/lib/docker

# Monitor memory
docker stats amss-postgres

# Set memory limits (recommended)
sudo docker run -d \
  --name amss-postgres \
  --memory="2g" \
  --memory-swap="2g" \
  -e POSTGRES_USER=amss \
  -e POSTGRES_PASSWORD=amss \
  -e POSTGRES_DB=amss \
  -p 0.0.0.0:5455:5432 \
  --restart unless-stopped \
  postgres:16
```

---

### PostgreSQL Data Corruption

**Failure Scenario:** Database files become corrupted

**Impact:**
- Postgres fails to start
- Error: `could not open file`, `invalid page header`
- Cannot read/write data

**Detection:**
```bash
# Container restarts in loop
sudo docker logs amss-postgres
# Shows: PANIC: could not open file "base/..."
```

**Recovery:**
```bash
# 1. Stop container
sudo docker stop amss-postgres

# 2. Restore from backup (see Backup Procedures section)
# Extract backup
cd /backups
tar -xzf postgres-backup-2025-01-24.tar.gz

# 3. Restore data directory
sudo docker run --rm \
  -v /backups/postgres-data:/backup \
  -v amss-postgres-data:/var/lib/postgresql/data \
  postgres:16 \
  bash -c "rm -rf /var/lib/postgresql/data/* && cp -r /backup/* /var/lib/postgresql/data/"

# 4. Start Postgres
sudo docker start amss-postgres

# 5. Verify
psql "postgres://amss:amss@localhost:5455/amss" -c "\dt"
```

**Data Loss:**
- ⚠️ Depends on backup age
- **RPO (Recovery Point Objective):** 24 hours (with daily backups)

---

### Redis Crash

**Failure Scenario:** Redis container crashes or becomes unresponsive

**Impact:**
- **Severity:** 🟡 **MEDIUM** (graceful degradation)
- Rate limiting **disabled** (all requests accepted)
- Idempotency checks **disabled** (risk of duplicate operations)
- Webhook queue **paused** (events accumulate in database outbox)
- API continues serving requests (no downtime!)

**Why It's Not Critical:**
- AMSS designed to degrade gracefully without Redis
- Postgres is primary data store
- Redis only used for:
  - Rate limiting (safety feature, not required)
  - Job queues (can rebuild from outbox)
  - Idempotency cache (short TTL, acceptable loss)

**Detection:**
```bash
# Check container
sudo docker ps | grep amss-redis

# Check API logs
kubectl -n amss-uat logs deployment/amss-server --tail=50 | grep -i redis
# Shows: "redis connection failed"

# Verify degraded mode
curl https://amss-api-uat.duckdns.org/health
# Still returns 200 (API continues)
```

**Automatic Recovery:**
- Docker restart policy: `--restart unless-stopped`
- **RTO:** 5-10 seconds

**Manual Recovery:**
```bash
# 1. Restart container
sudo docker restart amss-redis

# 2. Verify
redis-cli -h localhost -p 6379 PING
# Response: PONG

# 3. Check worker logs (webhooks should resume)
kubectl -n amss-uat logs deployment/amss-worker --tail=20 -f
```

**Data Loss:**
- ✅ **Acceptable** - All critical data in Postgres
- Rate limit counters reset (users get fresh quota)
- Idempotency cache lost (24hr TTL, low risk)
- Webhook queue rebuilt from outbox table

**Prevention:**
- Not critical to prevent (graceful degradation)
- Optional: Enable Redis persistence (RDB snapshots)

---

### Database Connection Pool Exhaustion

**Failure Scenario:** All database connections consumed, new requests blocked

**Impact:**
- Requests timeout waiting for connection
- API returns `500 Internal Server Error`
- Logs show: `connection pool exhausted`

**Causes:**
- Long-running queries
- Connection leaks (not released)
- Traffic spike exceeding pool size

**Detection:**
```bash
# Check Postgres connections
psql "postgres://amss:amss@localhost:5455/amss" -c "
  SELECT count(*), state
  FROM pg_stat_activity
  WHERE datname = 'amss'
  GROUP BY state;
"

# Look for high 'active' count (>50 is suspicious)
```

**Recovery:**
```bash
# 1. Kill long-running queries
psql "postgres://amss:amss@localhost:5455/amss" -c "
  SELECT pg_terminate_backend(pid)
  FROM pg_stat_activity
  WHERE datname = 'amss'
    AND state = 'active'
    AND query_start < NOW() - INTERVAL '5 minutes';
"

# 2. Restart AMSS pods (releases connections)
kubectl -n amss-uat rollout restart deployment/amss-server
kubectl -n amss-uat rollout restart deployment/amss-worker

# 3. Monitor recovery
kubectl -n amss-uat get pods -w
```

**Prevention:**
```go
// Recommended connection pool settings
db.SetMaxOpenConns(25)      // Max connections
db.SetMaxIdleConns(5)       // Idle connections
db.SetConnMaxLifetime(5*time.Minute)  // Recycle connections
```

---

## 2. Application Failures

### amss-server Pod Crash

**Failure Scenario:** Server pod crashes (panic, OOM, segfault)

**Impact:**
- **Severity:** 🔴 **CRITICAL** (if only 1 replica)
- API becomes unavailable
- Client requests timeout or return 502/503
- No new requests processed

**Detection:**
```bash
# Check pod status
kubectl -n amss-uat get pods | grep amss-server
# Shows: CrashLoopBackOff or Error

# Check crash reason
kubectl -n amss-uat describe pod amss-server-xxxxx
# Look at: Last State > Terminated > Reason/Exit Code

# View logs before crash
kubectl -n amss-uat logs amss-server-xxxxx --previous
```

**Common Causes:**
| Cause | Exit Code | Log Pattern |
|-------|-----------|-------------|
| Panic | 2 | `panic: ...` |
| OOM (Out of Memory) | 137 | `OOMKilled` |
| Missing env var | 1 | `missing required env vars` |
| Database unreachable | 1 | `connection refused` |

**Automatic Recovery:**
- Kubernetes restart policy: `Always`
- Pod automatically recreated
- **RTO:** 10-30 seconds (depending on startup time)

**Manual Recovery:**
```bash
# 1. View crash logs
kubectl -n amss-uat logs deployment/amss-server --previous --tail=100

# 2. If config issue, fix and redeploy
helm upgrade amss deploy/helm/amss \
  -n amss-uat \
  -f deploy/helm/values-uat.yaml

# 3. If stuck in crash loop, delete pod (forces recreation)
kubectl -n amss-uat delete pod amss-server-xxxxx

# 4. Monitor recovery
kubectl -n amss-uat get pods -w
kubectl -n amss-uat logs deployment/amss-server -f
```

**Data Loss:** ✅ None (stateless application)

**Prevention:**
```yaml
# Add resource limits (prevents OOM on node)
resources:
  limits:
    memory: 512Mi
  requests:
    memory: 128Mi

# Add liveness probe (auto-restart on hang)
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 30
```

---

### amss-worker Pod Crash

**Failure Scenario:** Worker pod crashes during job processing

**Impact:**
- **Severity:** 🟡 **MEDIUM**
- Webhooks delayed (not lost)
- CSV imports paused
- Program task generation paused
- **No API downtime** (worker is background process)

**Why It's Not Critical:**
- Worker jobs are idempotent (safe to retry)
- Outbox pattern ensures at-least-once delivery
- Events queued in database (durable)

**Detection:**
```bash
# Check pod status
kubectl -n amss-uat get pods | grep amss-worker

# Check webhook deliveries (should be recent)
psql "postgres://amss:amss@localhost:5455/amss" -c "
  SELECT event_type, status, last_attempt_at
  FROM webhook_deliveries
  ORDER BY created_at DESC
  LIMIT 10;
"
# If last_attempt_at is old (>5 min), worker is down
```

**Automatic Recovery:**
- Kubernetes restarts pod
- Worker resumes from last checkpoint
- **RTO:** 10-30 seconds

**Manual Recovery:**
```bash
# 1. Check crash reason
kubectl -n amss-uat logs deployment/amss-worker --previous

# 2. Restart if stuck
kubectl -n amss-uat rollout restart deployment/amss-worker

# 3. Verify jobs resume
kubectl -n amss-uat logs deployment/amss-worker -f | grep -i "job\|outbox\|webhook"
```

**Data Loss:** ✅ None (jobs are retried)

---

### Memory Leak / Resource Exhaustion

**Failure Scenario:** Application slowly consumes all memory

**Impact:**
- Gradual performance degradation
- Eventually: OOM kill by Kubernetes

**Detection:**
```bash
# Monitor memory usage over time
kubectl -n amss-uat top pods

# Check for memory growth trend
kubectl -n amss-uat get pods -o jsonpath='{.items[*].status.containerStatuses[0].restartCount}'
# High restart count = repeated OOM kills
```

**Recovery:**
```bash
# 1. Immediate: Restart pod
kubectl -n amss-uat rollout restart deployment/amss-server

# 2. Long-term: Add memory limits
# Edit values.yaml
resources:
  limits:
    memory: 512Mi  # Pod killed if exceeds this
```

**Prevention:**
- Add memory limits (forces hard cap)
- Implement memory profiling (Go: pprof)
- Regular restarts (via CronJob) if leak unfixable

---

### Stuck Background Job

**Failure Scenario:** Worker job stuck in infinite loop or waiting indefinitely

**Impact:**
- Webhook delivery stops for that event
- Import job never completes
- Other jobs continue (parallel processing)

**Detection:**
```bash
# Check for old pending deliveries
psql "postgres://amss:amss@localhost:5455/amss" -c "
  SELECT id, event_type, attempts, last_attempt_at, created_at
  FROM webhook_deliveries
  WHERE status = 'pending'
    AND created_at < NOW() - INTERVAL '1 hour';
"

# Check worker logs for repeated errors
kubectl -n amss-uat logs deployment/amss-worker | grep -i "error\|retry"
```

**Recovery:**
```bash
# 1. Restart worker (breaks stuck loop)
kubectl -n amss-uat rollout restart deployment/amss-worker

# 2. Manually retry failed delivery
psql "postgres://amss:amss@localhost:5455/amss" -c "
  UPDATE webhook_deliveries
  SET status = 'pending', attempts = 0
  WHERE id = '<stuck-delivery-id>';
"

# 3. If webhook URL is permanently down, disable it
psql "postgres://amss:amss@localhost:5455/amss" -c "
  UPDATE webhooks
  SET active = false
  WHERE id = '<webhook-id>';
"
```

---

## 3. Infrastructure Failures

### Kubernetes Node Failure

**Failure Scenario:** VPS host crashes or becomes unreachable

**Impact:**
- **Severity:** 🔴 **CRITICAL**
- All pods on node become unavailable
- Complete AMSS outage

**Detection:**
```bash
# Node status
kubectl get nodes
# Shows: NotReady or Unknown

# Check system logs
sudo journalctl -u k3s -n 100
```

**Recovery:**
```bash
# 1. If VPS is responsive, restart k3s
sudo systemctl restart k3s

# 2. If VPS is down, contact hosting provider
# OVH/Linode/DigitalOcean support ticket

# 3. When node recovers, verify pods restart
kubectl -n amss-uat get pods

# 4. If pods stuck, delete them (force reschedule)
kubectl -n amss-uat delete pods --all
```

**Prevention:**
- **Multi-node cluster** (recommended for production)
- Regular backups (can restore to new VPS)
- Monitoring alerts for node health

---

### Disk Full

**Failure Scenario:** Node runs out of disk space

**Impact:**
- New pod creation fails
- Container logs cannot be written
- Database writes fail

**Detection:**
```bash
# Check disk usage
df -h

# Check Docker overlay usage
sudo du -sh /var/lib/docker/overlay2/*

# Check pod eviction events
kubectl get events --all-namespaces | grep -i evict
```

**Recovery:**
```bash
# 1. Clean Docker images
sudo docker image prune -a -f

# 2. Clean old logs
sudo journalctl --vacuum-size=100M

# 3. Clean unused volumes
sudo docker volume prune -f

# 4. If critical, expand disk or add volume
```

**Prevention:**
```bash
# Set up disk monitoring alert at 80% usage
# Add log rotation for application logs
# Regular cleanup cron job
```

---

### cert-manager Failure

**Failure Scenario:** cert-manager pod crashes, certificate renewal fails

**Impact:**
- **Immediate:** None (existing cert still valid)
- **Future:** Certificate expires in 3 months, HTTPS breaks

**Detection:**
```bash
# Check cert-manager pods
kubectl -n cert-manager get pods

# Check certificate status
kubectl -n amss-uat get certificate
# Look for: Ready=False

# Check certificate expiry
kubectl -n amss-uat get certificate amss-uat-tls -o jsonpath='{.status.notAfter}'
```

**Recovery:**
```bash
# 1. Restart cert-manager
kubectl -n cert-manager rollout restart deployment/cert-manager

# 2. Manually trigger renewal
kubectl -n amss-uat delete certificaterequest --all
kubectl -n amss-uat delete secret amss-uat-tls

# Certificate will be auto-recreated
```

**Data Loss:** ✅ None (certificate can be re-issued)

---

### ingress-nginx Crash

**Failure Scenario:** Ingress controller pod crashes

**Impact:**
- **Severity:** 🔴 **CRITICAL**
- HTTPS traffic cannot reach backend
- API becomes unreachable from internet

**Detection:**
```bash
# Check ingress controller
kubectl -n ingress-nginx get pods

# Test from outside
curl https://amss-api-uat.duckdns.org/health
# Returns: Connection refused or timeout
```

**Recovery:**
```bash
# 1. Restart ingress controller
kubectl -n ingress-nginx rollout restart deployment/ingress-nginx-controller

# 2. Verify
kubectl -n ingress-nginx get pods
curl https://amss-api-uat.duckdns.org/health
```

**Automatic Recovery:** Kubernetes restarts pod automatically

---

## 4. Network & Connectivity Failures

### DNS Resolution Failure

**Failure Scenario:** Domain name doesn't resolve to VPS IP

**Impact:**
- Users cannot reach API via domain name
- Direct IP access still works

**Detection:**
```bash
# Test DNS
nslookup amss-api-uat.duckdns.org
# Should return: 51.79.85.92

# Test via IP (works if only DNS issue)
curl https://51.79.85.92/health -H "Host: amss-api-uat.duckdns.org" -k
```

**Recovery:**
```bash
# Update DNS record at DuckDNS
curl "https://www.duckdns.org/update?domains=amss-api-uat&token=<token>&ip=51.79.85.92"

# Wait for propagation (usually 5 minutes)
# Verify
nslookup amss-api-uat.duckdns.org
```

---

### Firewall Blocking Database Access

**Failure Scenario:** iptables rules accidentally removed or misconfigured

**Impact:**
- Kubernetes pods cannot reach Postgres/Redis
- API returns 503 errors

**Detection:**
```bash
# Check iptables rules
sudo iptables -L INPUT -n | grep -E "5455|6379"

# Should show:
# ACCEPT tcp -- 10.42.0.0/24 anywhere tcp dpt:5455
# ACCEPT tcp -- 127.0.0.1 anywhere tcp dpt:5455
# DROP tcp -- anywhere anywhere tcp dpt:5455

# Test from pod
kubectl -n amss-uat exec -it amss-server-xxxxx -- sh
nc -zv 51.79.85.92 5455
# Should connect
```

**Recovery:**
```bash
# 1. Restore iptables rules
sudo iptables -I INPUT -p tcp --dport 5455 -s 127.0.0.1 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 5455 -s 10.42.0.0/24 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 5455 -j DROP

sudo iptables -I INPUT -p tcp --dport 6379 -s 127.0.0.1 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 6379 -s 10.42.0.0/24 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 6379 -j DROP

# 2. Save rules
sudo iptables-save | sudo tee /etc/iptables/rules.v4

# 3. Test connectivity
kubectl -n amss-uat logs deployment/amss-server --tail=10
# Should no longer show connection errors
```

---

### External Webhook Endpoint Down

**Failure Scenario:** Customer's webhook URL is unreachable

**Impact:**
- Webhook deliveries fail
- Retries continue (up to 10 attempts over 28 hours)
- Eventually marked as failed

**Detection:**
```bash
# Check failed deliveries
psql "postgres://amss:amss@localhost:5455/amss" -c "
  SELECT w.url, wd.event_type, wd.attempts, wd.response_status
  FROM webhook_deliveries wd
  JOIN webhooks w ON w.id = wd.webhook_id
  WHERE wd.status = 'failed'
  ORDER BY wd.created_at DESC
  LIMIT 10;
"
```

**Recovery:**
```bash
# 1. Contact customer to fix their endpoint

# 2. Manually retry after fix
psql "postgres://amss:amss@localhost:5455/amss" -c "
  UPDATE webhook_deliveries
  SET status = 'pending', attempts = 0
  WHERE id = '<delivery-id>';
"

# 3. Or disable webhook temporarily
psql "postgres://amss:amss@localhost:5455/amss" -c "
  UPDATE webhooks
  SET active = false
  WHERE url = '<failing-url>';
"
```

---

## 5. Data Integrity Issues

### Duplicate Task Creation

**Failure Scenario:** Task created twice due to idempotency failure

**Impact:**
- Duplicate maintenance tasks in system
- Technician may perform work twice
- Inaccurate compliance reporting

**Causes:**
- Client retry without idempotency key
- Idempotency cache expired (Redis restart)
- Race condition (rare)

**Detection:**
```sql
-- Find duplicate tasks
SELECT aircraft_id, program_id, due_at_hours, COUNT(*)
FROM tasks
WHERE status = 'scheduled'
GROUP BY aircraft_id, program_id, due_at_hours
HAVING COUNT(*) > 1;
```

**Recovery:**
```sql
-- Manual deduplication (keep oldest, delete newer)
WITH duplicates AS (
  SELECT id, ROW_NUMBER() OVER (
    PARTITION BY aircraft_id, program_id, due_at_hours
    ORDER BY created_at ASC
  ) AS rn
  FROM tasks
  WHERE status = 'scheduled'
)
DELETE FROM tasks
WHERE id IN (SELECT id FROM duplicates WHERE rn > 1);
```

**Prevention:**
- Always use idempotency keys for writes
- Client-side request deduplication

---

### Orphaned Part Reservations

**Failure Scenario:** Task deleted but part reservations remain

**Impact:**
- Parts marked as reserved but not actually used
- Inventory appears lower than reality

**Detection:**
```sql
-- Find orphaned reservations
SELECT pr.id, pr.quantity, i.part_number
FROM part_reservations pr
JOIN inventory i ON i.id = pr.inventory_id
LEFT JOIN tasks t ON t.id = pr.task_id
WHERE t.id IS NULL
  AND pr.status = 'reserved';
```

**Recovery:**
```sql
-- Release orphaned reservations
UPDATE part_reservations
SET status = 'released'
WHERE task_id NOT IN (SELECT id FROM tasks)
  AND status = 'reserved';

-- Recalculate inventory
UPDATE inventory
SET quantity_reserved = (
  SELECT COALESCE(SUM(quantity), 0)
  FROM part_reservations
  WHERE inventory_id = inventory.id
    AND status = 'reserved'
);
```

---

### Audit Log Gaps

**Failure Scenario:** Audit log entry not created during operation

**Impact:**
- Missing compliance trail
- Cannot trace who performed action

**Detection:**
```sql
-- Check for operations without audit logs
SELECT t.id, t.status, t.updated_at
FROM tasks t
WHERE t.status = 'completed'
  AND NOT EXISTS (
    SELECT 1 FROM audit_logs
    WHERE resource_type = 'task'
      AND resource_id = t.id
      AND action = 'completed'
  )
LIMIT 10;
```

**Recovery:**
```sql
-- Cannot recreate past audit logs
-- Document gap and implement prevention
```

**Prevention:**
- Audit logging in same transaction as operation
- Regular audit log completeness checks

---

## 6. Backup & Recovery Procedures

### Daily Database Backup

**Automated Backup Script:**

**File:** `/home/ubuntu/scripts/backup-postgres.sh`
```bash
#!/bin/bash
set -e

BACKUP_DIR="/backups/postgres"
DATE=$(date +%Y-%m-%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/postgres-backup-$DATE.sql.gz"

# Create backup directory
mkdir -p $BACKUP_DIR

# Dump database
sudo docker exec amss-postgres pg_dump -U amss amss | gzip > $BACKUP_FILE

# Keep last 30 days only
find $BACKUP_DIR -name "postgres-backup-*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"

# Verify backup integrity
gunzip -t $BACKUP_FILE && echo "✅ Backup verified"
```

**Schedule with cron:**
```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /home/ubuntu/scripts/backup-postgres.sh >> /var/log/postgres-backup.log 2>&1
```

---

### Restore from Backup

**Procedure:**
```bash
# 1. Stop AMSS (prevent writes during restore)
kubectl -n amss-uat scale deployment/amss-server --replicas=0
kubectl -n amss-uat scale deployment/amss-worker --replicas=0

# 2. Drop existing database
sudo docker exec -i amss-postgres psql -U amss -c "DROP DATABASE IF EXISTS amss;"
sudo docker exec -i amss-postgres psql -U amss -c "CREATE DATABASE amss;"

# 3. Restore backup
gunzip -c /backups/postgres/postgres-backup-2025-01-24-020000.sql.gz | \
  sudo docker exec -i amss-postgres psql -U amss -d amss

# 4. Verify restore
sudo docker exec amss-postgres psql -U amss -d amss -c "\dt"

# 5. Restart AMSS
kubectl -n amss-uat scale deployment/amss-server --replicas=1
kubectl -n amss-uat scale deployment/amss-worker --replicas=1

# 6. Test application
curl https://amss-api-uat.duckdns.org/health
```

**Estimated Downtime:** 5-15 minutes (depending on database size)

---

### Point-in-Time Recovery

**Scenario:** Restore to state before bad data was written

**Requirements:** WAL archiving enabled (not currently configured)

**To Enable WAL Archiving:**
```bash
# 1. Configure Postgres
sudo docker exec amss-postgres psql -U amss -c "
  ALTER SYSTEM SET wal_level = 'replica';
  ALTER SYSTEM SET archive_mode = 'on';
  ALTER SYSTEM SET archive_command = 'cp %p /backups/wal/%f';
"

# 2. Restart Postgres
sudo docker restart amss-postgres

# 3. Create WAL archive directory
sudo mkdir -p /backups/wal
sudo chmod 777 /backups/wal
```

**Restore to Point in Time:**
```bash
# 1. Restore base backup
# ... (same as above)

# 2. Apply WAL files up to desired time
# This requires pg_restore with --target-time
# Complex procedure - consult Postgres documentation
```

**Note:** Point-in-time recovery adds complexity. Only enable if required for compliance.

---

### Backup Verification

**Monthly Test Restore:**
```bash
#!/bin/bash
# File: /home/ubuntu/scripts/test-restore.sh

# 1. Create test database
sudo docker exec amss-postgres psql -U amss -c "CREATE DATABASE amss_test;"

# 2. Restore latest backup to test DB
LATEST_BACKUP=$(ls -t /backups/postgres/postgres-backup-*.sql.gz | head -1)
gunzip -c $LATEST_BACKUP | \
  sudo docker exec -i amss-postgres psql -U amss -d amss_test

# 3. Verify data
COUNT=$(sudo docker exec amss-postgres psql -U amss -d amss_test -t -c "SELECT COUNT(*) FROM tasks;")

if [ $COUNT -gt 0 ]; then
  echo "✅ Backup verified: $COUNT tasks found"
else
  echo "❌ Backup verification failed: No data"
  exit 1
fi

# 4. Cleanup
sudo docker exec amss-postgres psql -U amss -c "DROP DATABASE amss_test;"
```

**Schedule monthly:**
```bash
# First day of month at 3 AM
0 3 1 * * /home/ubuntu/scripts/test-restore.sh >> /var/log/backup-test.log 2>&1
```

---

### Offsite Backup

**Upload to S3/Cloud Storage:**
```bash
#!/bin/bash
# File: /home/ubuntu/scripts/offsite-backup.sh

# Install AWS CLI first: sudo apt-get install awscli

BACKUP_FILE=$(ls -t /backups/postgres/postgres-backup-*.sql.gz | head -1)
S3_BUCKET="s3://amss-backups-production"

# Upload to S3
aws s3 cp $BACKUP_FILE $S3_BUCKET/ \
  --storage-class GLACIER \
  --metadata "created=$(date)"

echo "✅ Offsite backup complete: $BACKUP_FILE → $S3_BUCKET"
```

**Schedule weekly:**
```bash
# Every Sunday at 4 AM
0 4 * * 0 /home/ubuntu/scripts/offsite-backup.sh >> /var/log/offsite-backup.log 2>&1
```

---

## 7. Disaster Recovery

### Complete VPS Loss

**Scenario:** VPS destroyed, need to rebuild from scratch

**Recovery Steps:**

#### 1. Provision New VPS
```bash
# Same specs as original:
# - Ubuntu 24.10
# - 8GB RAM
# - 80GB disk
# - Same region (for low latency)
```

#### 2. Install k3s
```bash
curl -sfL https://get.k3s.io | sh -

# Verify
kubectl get nodes
```

#### 3. Install Dependencies
```bash
# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Docker
sudo apt-get update
sudo apt-get install -y docker.io
```

#### 4. Restore PostgreSQL
```bash
# Create Postgres container
sudo docker run -d \
  --name amss-postgres \
  -e POSTGRES_USER=amss \
  -e POSTGRES_PASSWORD=amss \
  -e POSTGRES_DB=amss \
  -p 0.0.0.0:5455:5432 \
  --restart unless-stopped \
  postgres:16

# Wait for startup
sleep 10

# Restore from S3 backup
aws s3 cp s3://amss-backups-production/postgres-backup-latest.sql.gz /tmp/
gunzip -c /tmp/postgres-backup-latest.sql.gz | \
  sudo docker exec -i amss-postgres psql -U amss -d amss

# Verify
sudo docker exec amss-postgres psql -U amss -d amss -c "\dt"
```

#### 5. Restore Redis
```bash
sudo docker run -d \
  --name amss-redis \
  -p 0.0.0.0:6379:6379 \
  --restart unless-stopped \
  redis:7
```

#### 6. Apply Firewall Rules
```bash
# PostgreSQL
sudo iptables -I INPUT -p tcp --dport 5455 -s 127.0.0.1 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 5455 -s 10.42.0.0/24 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 5455 -j DROP

# Redis
sudo iptables -I INPUT -p tcp --dport 6379 -s 127.0.0.1 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 6379 -s 10.42.0.0/24 -j ACCEPT
sudo iptables -I INPUT -p tcp --dport 6379 -j DROP

# Save
sudo mkdir -p /etc/iptables
sudo iptables-save | sudo tee /etc/iptables/rules.v4
```

#### 7. Deploy cert-manager
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.crds.yaml

helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.13.3

# Create ClusterIssuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@amss-api-uat.duckdns.org
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

#### 8. Deploy ingress-nginx
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace

# Patch for hostPort
kubectl -n ingress-nginx patch deployment ingress-nginx-controller --type='json' -p='[
  {"op":"add","path":"/spec/template/spec/containers/0/ports/0/hostPort","value":80},
  {"op":"add","path":"/spec/template/spec/containers/0/ports/1/hostPort","value":443}
]'
```

#### 9. Deploy AMSS
```bash
# Clone repository
git clone <AMSS_REPO_URL> /home/ubuntu/amss-backend
cd /home/ubuntu/amss-backend

# Create namespace
kubectl create namespace amss-uat

# Create secrets
kubectl -n amss-uat create secret docker-registry amss-registry \
  --docker-username="$DOCKER_USER" \
  --docker-password="$DOCKER_TOKEN" \
  --docker-email="devnull@example.com"

kubectl -n amss-uat create secret generic amss-secret \
  --from-file=jwtPrivateKeyPem=jwt-private.pem \
  --from-file=jwtPublicKeyPem=jwt-public.pem \
  --from-literal=DB_URL="postgres://amss:amss@51.79.85.92:5455/amss?sslmode=disable"

# Deploy
helm upgrade --install amss deploy/helm/amss \
  -n amss-uat \
  -f deploy/helm/values-uat.yaml

# Wait for rollout
kubectl -n amss-uat rollout status deployment/amss-server
```

#### 10. Update DNS
```bash
# Get new VPS IP
curl ifconfig.me

# Update DuckDNS
curl "https://www.duckdns.org/update?domains=amss-api-uat&token=<token>&ip=<new-ip>"

# Wait 5 minutes for DNS propagation
```

#### 11. Verify
```bash
curl https://amss-api-uat.duckdns.org/health
# Should return: "ok"

curl https://amss-api-uat.duckdns.org/ready
# Should return: "ready"
```

**Total Recovery Time:** 2-4 hours (depending on backup size)

---

### Data Center Outage

**Scenario:** Entire OVH data center goes down

**Impact:** Complete service outage until recovery

**Mitigation:**
- **Multi-region deployment** (recommended for critical systems)
- Deploy AMSS in 2 regions (e.g., OVH Canada + AWS US-East)
- Use DNS failover (Route 53, Cloudflare)
- Cross-region database replication

**Not currently implemented** (single-region deployment)

---

## 8. Monitoring & Alerting

### Critical Metrics to Monitor

#### Application Health
```bash
# HTTP endpoint monitoring (every 60s)
curl https://amss-api-uat.duckdns.org/health
# Alert if: Not "ok" for 2 consecutive checks

curl https://amss-api-uat.duckdns.org/ready
# Alert if: Not "ready" for 2 consecutive checks
```

#### Database Health
```bash
# Postgres connection count
psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname='amss';"
# Alert if: > 50 connections

# Database size
psql -c "SELECT pg_size_pretty(pg_database_size('amss'));"
# Alert if: > 80% of available disk

# Slow queries
psql -c "SELECT query, query_start FROM pg_stat_activity WHERE state='active' AND query_start < NOW() - INTERVAL '5 minutes';"
# Alert if: Any query > 5 minutes
```

#### System Resources
```bash
# Disk usage
df -h / | awk 'NR==2 {print $5}'
# Alert if: > 80%

# Memory usage
free -h | awk 'NR==2 {print $3/$2 * 100}'
# Alert if: > 85%

# Pod restarts
kubectl -n amss-uat get pods -o jsonpath='{.items[*].status.containerStatuses[*].restartCount}'
# Alert if: Increases (pod crashing)
```

#### Application Metrics
```bash
# Webhook delivery failures
psql -c "SELECT COUNT(*) FROM webhook_deliveries WHERE status='failed' AND created_at > NOW() - INTERVAL '1 hour';"
# Alert if: > 10 in last hour

# Overdue tasks
psql -c "SELECT COUNT(*) FROM tasks WHERE status='scheduled' AND due_at_date < CURRENT_DATE;"
# Alert if: > 5 overdue tasks
```

---

### Recommended Monitoring Stack

#### Option 1: Prometheus + Grafana (Kubernetes-native)

**Install Prometheus:**
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

**AMSS already exposes `/metrics` endpoint** (when `PROMETHEUS_ENABLED=true`)

**Grafana Dashboard:**
- HTTP request rate
- Response time (p50, p95, p99)
- Error rate
- Database query duration
- Pod CPU/memory usage

---

#### Option 2: External Monitoring (Simpler)

**UptimeRobot (Free tier):**
- Monitor `/health` endpoint every 5 minutes
- Email alert on downtime

**Better Uptime:**
- Monitor API availability
- Incident management
- Status page for users

**New Relic / Datadog:**
- Full APM (Application Performance Monitoring)
- Automatic error tracking
- Log aggregation

---

### Alert Escalation Policy

| Severity | Response Time | Escalation |
|----------|---------------|------------|
| 🔴 **CRITICAL** (API down) | Immediate | Page on-call engineer |
| 🟡 **WARNING** (Database slow) | 15 minutes | Slack notification |
| 🟢 **INFO** (High traffic) | Next business day | Email summary |

**On-Call Rotation:**
- Use PagerDuty or OpsGenie
- 24/7 coverage for production
- Escalate to senior engineer after 30min

---

### Incident Response Checklist

When alert fires:

```
[ ] 1. Acknowledge alert (stop repeat notifications)
[ ] 2. Check system status (kubectl get pods, docker ps)
[ ] 3. Review recent changes (helm history, git log)
[ ] 4. Check logs (kubectl logs, docker logs)
[ ] 5. Attempt automatic recovery (restart pod/container)
[ ] 6. Escalate if needed (call senior engineer)
[ ] 7. Document incident (what, when, why, resolution)
[ ] 8. Post-mortem (how to prevent in future)
```

---

## Summary

### Recovery Time Objectives (RTO)

| Failure | RTO | Automatic Recovery |
|---------|-----|--------------------|
| Postgres crash | 10-30s | ✅ Yes (Docker restart) |
| Redis crash | 5-10s | ✅ Yes (Docker restart) |
| Server pod crash | 10-30s | ✅ Yes (Kubernetes restart) |
| Worker pod crash | 10-30s | ✅ Yes (Kubernetes restart) |
| Node failure | Hours | ❌ Manual (or multi-node cluster) |
| Complete VPS loss | 2-4 hours | ❌ Manual rebuild |

### Recovery Point Objectives (RPO)

| Data | RPO | Backup Frequency |
|------|-----|------------------|
| Database | 24 hours | Daily backups |
| Redis cache | Acceptable loss | No backup (ephemeral) |
| Audit logs | 0 (never lost) | In Postgres (backed up) |

### Prevention > Recovery

**Best Practices Implemented:**
- ✅ Automatic restarts (Docker, Kubernetes)
- ✅ Database backups (daily)
- ✅ Firewall protection (iptables)
- ✅ Health checks (`/health`, `/ready`)

**Recommended Additions:**
- ⚠️ Multi-node cluster (high availability)
- ⚠️ Automated monitoring (Prometheus)
- ⚠️ Point-in-time recovery (WAL archiving)
- ⚠️ Offsite backups (S3/Glacier)
- ⚠️ Multi-region deployment (disaster recovery)

---

**This document should be reviewed quarterly and updated after each incident.**

*Last Reviewed: December 24, 2025*
