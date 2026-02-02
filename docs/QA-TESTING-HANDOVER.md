# AMSS — QA Testing Handover Package

> **System:** AMSS (Aircraft Maintenance Scheduling System)
> **Version:** UAT — February 2026
> **Prepared for:** QA tester or LLM-assisted testing

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Access & Credentials](#2-access--credentials)
3. [Environment Setup](#3-environment-setup)
4. [API Reference](#4-api-reference)
5. [Test Scenarios by Feature](#5-test-scenarios-by-feature)
6. [RBAC Test Matrix](#6-rbac-test-matrix)
7. [Non-Functional Testing](#7-non-functional-testing)
8. [Known Issues & Limitations](#8-known-issues--limitations)
9. [Bug Reporting](#9-bug-reporting)
10. [curl Cheatsheet](#10-curl-cheatsheet)

---

## 1. System Overview

### What AMSS Is

AMSS is a multi-tenant aircraft maintenance scheduling and compliance tracking system. It manages:

- **Fleet** — aircraft registry with status, flight hours, cycles
- **Maintenance tasks** — scheduling, assignment, state machine (scheduled → in_progress → completed/cancelled)
- **Parts inventory** — definitions, stock items, reservations for tasks
- **Compliance** — items linked to tasks, sign-off workflow, immutable after approval
- **Audit trail** — every create/update/delete logged, immutable, exportable
- **CSV import** — bulk aircraft, parts, and program ingestion
- **Webhooks** — event subscriptions with retry delivery
- **Reports** — fleet summary and compliance reports

### Architecture

```
Internet
    │
    ▼
┌─────────────────────────────────────────────┐
│         SafeLine WAF (ports 80/443)          │
│         51.79.85.92                          │
├───────────────────┬─────────────────────────┤
│                   │                          │
│  amss.leoulgirma.com    amss-api-uat.duckdns.org
│         │                       │            │
│    Host nginx:8080     K8s NodePort:30443    │
│         │                       │            │
│  /var/www/amss (SPA)     ingress-nginx       │
│                                 │            │
│                          amss-server (Go)    │
│                          amss-worker (Go)    │
│                                 │            │
│                     ┌───────────┴──────────┐ │
│                   PostgreSQL 16     Redis 7  │
└─────────────────────────────────────────────┘
```

### Live URLs

| Resource | URL |
|----------|-----|
| Frontend (SPA) | `https://amss.leoulgirma.com` |
| Backend API | `https://amss-api-uat.duckdns.org/api/v1` |
| Swagger UI | `https://amss-api-uat.duckdns.org/docs` |
| OpenAPI spec | `https://amss-api-uat.duckdns.org/openapi.yaml` |
| Health check | `https://amss-api-uat.duckdns.org/health` |
| Readiness check | `https://amss-api-uat.duckdns.org/ready` |
| SafeLine admin | `https://51.79.85.92:9443` |
| Marketing site | `https://amss.leoulgirma.com/marketing/` |

---

## 2. Access & Credentials

### Test Organization

| Field | Value |
|-------|-------|
| Organization name | Demo Airline |
| Organization ID | `4cb97629-c58a-415d-bf9c-b400bb5e3d84` |

### Test User Accounts

All accounts use password: **`ChangeMe123!`**

| Email | Role | Display Name | Permissions Summary |
|-------|------|-------------|---------------------|
| `admin@demo.local` | `admin` | Super Admin | Full access to everything |
| `tenant-admin@demo.local` | `tenant_admin` | Organization Admin | Full access within org (no cross-org) |
| `scheduler@demo.local` | `scheduler` | Maintenance Scheduler | Manage maintenance, assign tasks, view fleet/parts, order parts |
| `mechanic@demo.local` | `mechanic` | Mechanic | View-only + complete assigned tasks |
| `auditor@demo.local` | `auditor` | Auditor | View-only + export reports |

### How to Obtain a JWT Token

**Step 1 — Look up organizations for an email:**

```bash
curl -s https://amss-api-uat.duckdns.org/api/v1/auth/lookup \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@demo.local"}'
```

Response contains the `org_id`.

**Step 2 — Login:**

```bash
curl -s https://amss-api-uat.duckdns.org/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "4cb97629-c58a-415d-bf9c-b400bb5e3d84",
    "email": "admin@demo.local",
    "password": "ChangeMe123!"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

**Step 3 — Use the token:**

```bash
curl -s https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Authorization: Bearer <access_token>"
```

### Token Lifecycle

| Token | TTL | Refresh Method |
|-------|-----|----------------|
| Access token | 15 minutes | `POST /api/v1/auth/refresh` with `{"refresh_token": "..."}` |
| Refresh token | 7 days | Re-login required after expiry |

### SafeLine WAF Admin

| Field | Value |
|-------|-------|
| URL | `https://51.79.85.92:9443` |
| Username | `admin` |
| Password | `njSSmpQU` |

---

## 3. Environment Setup

### Option A: Test Against UAT (Recommended for QA)

No setup required. Use the live URLs in Section 2. The UAT environment is seeded with the Demo Airline organization and all 5 test users.

### Option B: Local Development Environment

**Prerequisites:** Docker, Docker Compose, Go 1.24, Node.js 20, npm

**Backend:**

```bash
cd /home/ubuntu/amss-backend

# Start PostgreSQL, Redis, Jaeger, Prometheus, server, worker
make dev-up

# Run database migrations + seed data
make migrate-up

# Verify
curl http://localhost:8080/health
```

Local backend runs at `http://localhost:8080`.

**Frontend:**

```bash
cd /home/ubuntu/amss-frontend

# Install dependencies
npm install

# Set API URL for local backend
echo "VITE_API_BASE_URL=http://localhost:8080/api/v1" > .env.local

# Start dev server
npm run dev
```

Local frontend runs at `http://localhost:5173`.

### Loading Extended Test Data

The base seed (from migrations) creates 1 aircraft, 1 task, 1 program, 1 compliance item. For richer testing:

```bash
# Connect to local Postgres and run extended seed
psql "postgres://amss:amss@localhost:5455/amss?sslmode=disable" \
  -f scripts/seed_test_data.sql
```

This adds:
- 6 aircraft (N100AM–N600AM) across operational/maintenance/grounded
- 6 part definitions (Brake Assembly, Engine Fan Blade, Hydraulic Pump, Avionics Display Unit, Tire Assembly, APU Starter Motor)
- 8 part items with various statuses and expiry dates
- 5 maintenance tasks in different states (scheduled, in_progress, completed)
- 7 compliance items (some signed off, some pending)
- 5 maintenance programs (A-Check, B-Check, C-Check, Landing Gear Overhaul)

### Resetting the Environment

```bash
cd /home/ubuntu/amss-backend

# Destroy everything and start fresh
make dev-down
make dev-up
make migrate-up

# Optionally add extended test data
psql "postgres://amss:amss@localhost:5455/amss?sslmode=disable" \
  -f scripts/seed_test_data.sql
```

### Running Existing Tests

```bash
# Backend unit tests
cd /home/ubuntu/amss-backend && go test ./...

# Backend integration tests (requires running Postgres + Redis)
cd /home/ubuntu/amss-backend && AMSS_INTEGRATION=1 go test ./internal/infra/postgres ./internal/infra/redis

# Frontend unit tests
cd /home/ubuntu/amss-frontend && npm run test

# Frontend tests with coverage
cd /home/ubuntu/amss-frontend && npm run test:coverage

# Frontend lint + type check
cd /home/ubuntu/amss-frontend && npm run lint && npm run typecheck
```

---

## 4. API Reference

Base URL: `https://amss-api-uat.duckdns.org/api/v1`

### Authentication (No Token Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/lookup` | Find organizations by email. Body: `{"email": "..."}` |
| POST | `/auth/login` | Authenticate. Body: `{"org_id": "...", "email": "...", "password": "..."}` |
| POST | `/auth/refresh` | Refresh access token. Body: `{"refresh_token": "..."}` |
| POST | `/auth/logout` | Revoke refresh token. Body: `{"refresh_token": "..."}` |

### Protected Endpoints (Bearer Token Required)

#### User Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/me` | Get current authenticated user |

#### Aircraft
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/aircraft` | List aircraft. Query: `?limit=&offset=&status=` |
| GET | `/aircraft/{id}` | Get single aircraft |
| POST | `/aircraft` | Create aircraft |
| PATCH | `/aircraft/{id}` | Update aircraft |
| DELETE | `/aircraft/{id}` | Soft-delete aircraft |

#### Maintenance Tasks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/maintenance-tasks` | List tasks. Query: `?limit=&offset=&state=&aircraft_id=` |
| GET | `/maintenance-tasks/{id}` | Get task with reservations and compliance |
| POST | `/maintenance-tasks` | Create task |
| PATCH | `/maintenance-tasks/{id}` | Update task fields |
| PATCH | `/maintenance-tasks/{id}/state` | Transition state. Body: `{"state": "in_progress"}` |
| DELETE | `/maintenance-tasks/{id}` | Soft-delete task |

#### Maintenance Programs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/maintenance-programs` | List programs |
| GET | `/maintenance-programs/{id}` | Get program |
| POST | `/maintenance-programs` | Create program |
| PATCH | `/maintenance-programs/{id}` | Update program |
| DELETE | `/maintenance-programs/{id}` | Soft-delete program |

#### Part Definitions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/part-definitions` | List part definitions |
| POST | `/part-definitions` | Create part definition |
| PATCH | `/part-definitions/{id}` | Update part definition |
| DELETE | `/part-definitions/{id}` | Soft-delete part definition |

#### Part Items (Inventory)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/part-items` | List part items. Query: `?status=&part_definition_id=` |
| POST | `/part-items` | Create part item |
| PATCH | `/part-items/{id}` | Update part item |
| DELETE | `/part-items/{id}` | Soft-delete part item |

#### Part Reservations
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/part-reservations` | Reserve a part for a task |
| PATCH | `/part-reservations/{id}/state` | Update reservation state (`used` or `released`) |

#### Compliance
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/compliance-items` | List compliance items. Query: `?task_id=&result=` |
| POST | `/compliance-items` | Create compliance item |
| PATCH | `/compliance-items/{id}` | Update compliance item (before sign-off only) |
| PATCH | `/compliance-items/{id}/sign-off` | Sign off — makes item immutable |

#### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | List users |
| GET | `/users/{id}` | Get user |
| POST | `/users` | Create user. Body must include `role`, `email`, `password` |
| PATCH | `/users/{id}` | Update user |
| DELETE | `/users/{id}` | Soft-delete user |

#### Organizations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/organizations` | List organizations |
| GET | `/organizations/{id}` | Get organization |
| POST | `/organizations` | Create organization |
| PATCH | `/organizations/{id}` | Update organization |

#### Audit Logs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/audit-logs` | List audit logs. Query: `?action=&entity_type=&user_id=&limit=&offset=` |
| GET | `/audit-logs/export` | Export logs as CSV or JSON |

#### Reports
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/reports/summary` | Fleet summary report |
| GET | `/reports/compliance` | Compliance report |

#### CSV Import
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/imports/csv` | Upload CSV (multipart form, field: `file`, param: `type=aircraft|parts|programs`) |
| GET | `/imports/{id}` | Get import job status |
| GET | `/imports/{id}/rows` | Get row-level results |

#### Webhooks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/webhooks` | List webhook subscriptions |
| POST | `/webhooks` | Create webhook. Body: `{"url": "...", "events": ["task.created", ...]}` |
| DELETE | `/webhooks/{id}` | Delete webhook |
| POST | `/webhooks/{id}/test` | Send test event to webhook |

### Infrastructure (No Token Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Returns 200 if server is up |
| GET | `/ready` | Returns 200 if DB + Redis are connected |
| GET | `/metrics` | Prometheus metrics (if enabled) |
| GET | `/openapi.yaml` | OpenAPI specification |
| GET | `/docs` | Swagger UI |
| GET | `/ws?org_id=&user_id=` | WebSocket for real-time events |

### Error Response Format

All errors return:
```json
{
  "code": "<error_code>",
  "message": "Human-readable description",
  "details": "Optional additional info"
}
```

Error codes: `auth`, `forbidden`, `validation`, `conflict`, `not_found`, `rate_limited`, `internal`, `unavailable`

### Rate Limits

| Scope | Limit | Header |
|-------|-------|--------|
| Authenticated requests (per org) | 100 req/min | `X-RateLimit-Remaining`, `X-RateLimit-Reset` |
| Login endpoint (per org) | 10 req/min | `Retry-After` (on 429) |

### Idempotency

POST requests accept an `Idempotency-Key` header. If the same key + endpoint + body hash is seen within 24 hours, the cached response is returned. If the same key is sent with a different body, a `409 Conflict` is returned.

### Request Validation

The backend uses `DisallowUnknownFields()` — sending unrecognized JSON fields will return a `400 validation` error.

---

## 5. Test Scenarios by Feature

Each scenario includes the steps, expected result, and priority. Use the test accounts from Section 2.

### 5.1 Authentication & Authorization — Priority: CRITICAL

#### TC-AUTH-01: Email lookup flow
1. `POST /auth/lookup` with `{"email": "admin@demo.local"}`
2. **Expected:** 200, response contains org list with `"Demo Airline"` and org_id `4cb97629-c58a-415d-bf9c-b400bb5e3d84`

#### TC-AUTH-02: Successful login
1. `POST /auth/login` with valid org_id, email, password
2. **Expected:** 200, response contains `access_token`, `refresh_token`, `expires_in: 900`

#### TC-AUTH-03: Invalid password
1. `POST /auth/login` with correct email but wrong password
2. **Expected:** 401, error code `auth`

#### TC-AUTH-04: Non-existent email
1. `POST /auth/lookup` with `{"email": "nobody@nowhere.com"}`
2. **Expected:** 200 with empty org list, or 404

#### TC-AUTH-05: Token refresh
1. Login, wait or use the refresh token immediately
2. `POST /auth/refresh` with `{"refresh_token": "<token>"}`
3. **Expected:** 200 with new `access_token`

#### TC-AUTH-06: Expired token
1. Use an expired access token to call `GET /aircraft`
2. **Expected:** 401

#### TC-AUTH-07: Logout / token revocation
1. Login, then `POST /auth/logout` with the refresh token
2. Try `POST /auth/refresh` with the revoked token
3. **Expected:** Refresh fails with 401

#### TC-AUTH-08: Missing Authorization header
1. Call `GET /aircraft` without Authorization header
2. **Expected:** 401

#### TC-AUTH-09: Login rate limit
1. Send 11 rapid login requests
2. **Expected:** 10 succeed, 11th returns 429 with `Retry-After` header

#### TC-AUTH-10: Get current user
1. Login as each role, call `GET /auth/me`
2. **Expected:** 200, response includes correct `role`, `email`, `org_id`

### 5.2 Aircraft Management — Priority: HIGH

#### TC-AIR-01: List aircraft
1. Login as admin, `GET /aircraft`
2. **Expected:** 200, returns array of aircraft in Demo Airline org

#### TC-AIR-02: Create aircraft
1. Login as admin
2. `POST /aircraft` with `{"tail_number": "N700AM", "model": "Cessna 172", "status": "operational", "capacity_slots": 1}`
3. **Expected:** 201, aircraft created with UUID

#### TC-AIR-03: Duplicate tail number
1. `POST /aircraft` with `{"tail_number": "N100AM", ...}` (already exists)
2. **Expected:** 409 conflict

#### TC-AIR-04: Update aircraft
1. `PATCH /aircraft/{id}` with `{"status": "maintenance"}`
2. **Expected:** 200, status changed

#### TC-AIR-05: Delete aircraft
1. `DELETE /aircraft/{id}`
2. `GET /aircraft/{id}`
3. **Expected:** Delete returns 200/204. Subsequent GET returns 404 (soft-deleted).

#### TC-AIR-06: Invalid status value
1. `POST /aircraft` with `{"status": "flying"}`
2. **Expected:** 400 validation error. Valid statuses: `operational`, `maintenance`, `grounded`

#### TC-AIR-07: next_due must be after last_maintenance
1. `POST /aircraft` with `last_maintenance` in the future and `next_due` in the past
2. **Expected:** 400 validation error

#### TC-AIR-08: Mechanic cannot create aircraft
1. Login as `mechanic@demo.local`, `POST /aircraft`
2. **Expected:** 403 forbidden

### 5.3 Maintenance Tasks — Priority: CRITICAL

#### TC-TASK-01: Create task
1. Login as scheduler
2. `POST /maintenance-tasks` with valid aircraft_id, type, start_time, end_time
3. **Expected:** 201, task created in `scheduled` state

#### TC-TASK-02: State transition: scheduled → in_progress
1. `PATCH /maintenance-tasks/{id}/state` with `{"state": "in_progress"}`
2. **Expected:** 200, state changed

#### TC-TASK-03: State transition: in_progress → completed
1. `PATCH /maintenance-tasks/{id}/state` with `{"state": "completed"}`
2. **Expected:** 200, state changed

#### TC-TASK-04: Invalid state transition: scheduled → completed
1. `PATCH /maintenance-tasks/{id}/state` with `{"state": "completed"}` on a `scheduled` task
2. **Expected:** 400 validation error (must go through in_progress first, or verify if direct transition is allowed)

#### TC-TASK-05: State transition: scheduled → cancelled
1. `PATCH /maintenance-tasks/{id}/state` with `{"state": "cancelled"}`
2. **Expected:** 200 (cancellation should be allowed from scheduled)

#### TC-TASK-06: Overlapping maintenance windows
1. Create task for aircraft N100AM from 2026-02-10T00:00:00Z to 2026-02-15T00:00:00Z
2. Create another task for same aircraft from 2026-02-12T00:00:00Z to 2026-02-17T00:00:00Z
3. **Expected:** Second task rejected with conflict (overlapping `active_window` constraint)

#### TC-TASK-07: Assign mechanic
1. Create task, then `PATCH /maintenance-tasks/{id}` with `{"assigned_mechanic_id": "<mechanic_user_id>"}`
2. **Expected:** 200, mechanic assigned

#### TC-TASK-08: end_time must be after start_time
1. `POST /maintenance-tasks` with `end_time` before `start_time`
2. **Expected:** 400 validation error

#### TC-TASK-09: Auditor cannot create tasks
1. Login as `auditor@demo.local`, `POST /maintenance-tasks`
2. **Expected:** 403 forbidden

#### TC-TASK-10: Kanban board (frontend)
1. Login to frontend, navigate to `/kanban`
2. Drag a task card from "Scheduled" column to "In Progress"
3. **Expected:** Card moves, API call fires `PATCH /maintenance-tasks/{id}/state`, task state updates

### 5.4 Parts Inventory — Priority: HIGH

#### TC-PART-01: Create part definition
1. `POST /part-definitions` with `{"name": "Test Part", "category": "Avionics"}`
2. **Expected:** 201

#### TC-PART-02: Duplicate part definition name
1. Create same part name twice within same org
2. **Expected:** 409 conflict

#### TC-PART-03: Create part item
1. `POST /part-items` with valid `part_definition_id`, `serial_number`, `status: "in_stock"`, `expiry_date` in the future
2. **Expected:** 201

#### TC-PART-04: Duplicate serial number
1. Create two part items with the same `serial_number`
2. **Expected:** 409 conflict

#### TC-PART-05: Expired part rejection
1. `POST /part-items` with `expiry_date` in the past
2. **Expected:** 400 validation (constraint: expiry_date > now())

#### TC-PART-06: Reserve part for task
1. Create a part item (in_stock), create a task
2. `POST /part-reservations` with `{"task_id": "...", "part_item_id": "..."}`
3. **Expected:** 201, reservation created

#### TC-PART-07: Double reservation prevention
1. Reserve the same part item for two different tasks
2. **Expected:** Second reservation fails (only one active reservation per part item)

#### TC-PART-08: Release reservation
1. `PATCH /part-reservations/{id}/state` with `{"state": "released"}`
2. **Expected:** 200, part available again

#### TC-PART-09: Use reservation
1. `PATCH /part-reservations/{id}/state` with `{"state": "used"}`
2. **Expected:** 200

### 5.5 Compliance — Priority: CRITICAL

#### TC-COMP-01: Create compliance item
1. `POST /compliance-items` with `{"task_id": "...", "description": "Safety inspection", "result": "pending"}`
2. **Expected:** 201

#### TC-COMP-02: Update before sign-off
1. `PATCH /compliance-items/{id}` with `{"result": "pass"}`
2. **Expected:** 200

#### TC-COMP-03: Sign off
1. `PATCH /compliance-items/{id}/sign-off`
2. **Expected:** 200, `sign_off_user_id` and `sign_off_time` populated

#### TC-COMP-04: Immutable after sign-off
1. After sign-off, try `PATCH /compliance-items/{id}` with `{"result": "fail"}`
2. **Expected:** 400 or 403 — compliance item is immutable after sign-off (enforced by database trigger)

#### TC-COMP-05: Compliance report
1. `GET /reports/compliance`
2. **Expected:** 200 with compliance summary data

### 5.6 Audit Logs — Priority: HIGH

#### TC-AUDIT-01: Automatic audit trail
1. Create an aircraft, then `GET /audit-logs?entity_type=aircraft`
2. **Expected:** Audit entry with `action: "create"`, correct `user_id`, `entity_id`, timestamp

#### TC-AUDIT-02: Audit on update
1. Update an aircraft, check audit logs
2. **Expected:** Entry with `action: "update"` and `details` containing the changes

#### TC-AUDIT-03: Audit on delete
1. Delete an aircraft, check audit logs
2. **Expected:** Entry with `action: "delete"`

#### TC-AUDIT-04: Audit immutability
1. Attempt to modify or delete an audit log entry via direct DB access
2. **Expected:** Blocked by `audit_logs_immutable` trigger

#### TC-AUDIT-05: Export audit logs
1. `GET /audit-logs/export` (check Accept header or query param for format)
2. **Expected:** CSV or JSON download of audit entries

#### TC-AUDIT-06: Audit log filtering
1. `GET /audit-logs?action=create&entity_type=aircraft&limit=5`
2. **Expected:** Only matching entries returned, paginated

### 5.7 CSV Import — Priority: MEDIUM

#### TC-IMP-01: Import aircraft CSV
1. `POST /imports/csv` with type=aircraft and a valid CSV file
2. Poll `GET /imports/{id}` until status is `completed`
3. **Expected:** New aircraft created, import status shows row counts

#### TC-IMP-02: Import with invalid data
1. Upload CSV with missing required fields or invalid values
2. **Expected:** Import status `failed` or individual rows marked `invalid` with error details

#### TC-IMP-03: Import row details
1. After import, `GET /imports/{id}/rows`
2. **Expected:** Each row shows status (`valid`, `invalid`, `applied`) and any errors

### 5.8 Webhooks — Priority: MEDIUM

#### TC-WH-01: Create webhook
1. `POST /webhooks` with `{"url": "https://webhook.site/...", "events": ["task.created"]}`
2. **Expected:** 201

#### TC-WH-02: Test webhook
1. `POST /webhooks/{id}/test`
2. **Expected:** 200, test event delivered to URL

#### TC-WH-03: Webhook fires on event
1. Create a webhook for `task.created`
2. Create a new maintenance task
3. **Expected:** Webhook URL receives POST with task data and HMAC signature header

#### TC-WH-04: Delete webhook
1. `DELETE /webhooks/{id}`
2. **Expected:** 200/204

### 5.9 Reports — Priority: MEDIUM

#### TC-RPT-01: Fleet summary
1. `GET /reports/summary`
2. **Expected:** 200, includes aircraft counts by status, task counts by state, part inventory stats

#### TC-RPT-02: Compliance report
1. `GET /reports/compliance`
2. **Expected:** 200, includes compliance item counts by result (pass/fail/pending)

### 5.10 Multi-Tenancy — Priority: CRITICAL

#### TC-MT-01: Data isolation
1. Create a second organization via `POST /organizations`
2. Create a user in the new org
3. Login as the new user, `GET /aircraft`
4. **Expected:** Returns empty list (no access to Demo Airline's aircraft)

#### TC-MT-02: Cannot access other org's resources
1. Login as admin in Demo Airline
2. Note an aircraft ID
3. Login as user in the new org
4. `GET /aircraft/{demo_airline_aircraft_id}`
5. **Expected:** 404 (org-scoped query filters it out)

#### TC-MT-03: Org-scoped unique constraints
1. Create aircraft with tail "N100AM" in Demo Airline (already exists) — should fail
2. Create aircraft with tail "N100AM" in the new org — should succeed
3. **Expected:** Unique constraints are per-org, not global

### 5.11 Users — Priority: HIGH

#### TC-USER-01: Create user
1. `POST /users` with `{"email": "test@demo.local", "password": "Test1234!", "role": "mechanic"}`
2. **Expected:** 201

#### TC-USER-02: Duplicate email in same org
1. Create user with `email: "admin@demo.local"` (already exists)
2. **Expected:** 409 conflict

#### TC-USER-03: List users
1. `GET /users`
2. **Expected:** Returns users in current org only

#### TC-USER-04: Update user role
1. `PATCH /users/{id}` with `{"role": "scheduler"}`
2. **Expected:** 200, role changed

#### TC-USER-05: Delete user
1. `DELETE /users/{id}`
2. **Expected:** 200/204, user soft-deleted

#### TC-USER-06: Mechanic cannot manage users
1. Login as mechanic, `POST /users`
2. **Expected:** 403 forbidden

### 5.12 Frontend-Specific — Priority: HIGH

#### TC-FE-01: Login flow (UI)
1. Go to `https://amss.leoulgirma.com/login`
2. Enter email `admin@demo.local`, click continue
3. Select "Demo Airline" from org list
4. Enter password `ChangeMe123!`, click login
5. **Expected:** Redirected to dashboard with fleet stats, charts, recent activity

#### TC-FE-02: Demo mode
1. On login page, click "Demo Mode" button
2. **Expected:** Enters app with mock data, all pages functional with sample data

#### TC-FE-03: Navigation
1. After login, click each sidebar link: Dashboard, Fleet, Maintenance, Calendar, Kanban, Team, Parts, Compliance, Reports, Audit, Notifications, Settings
2. **Expected:** Each page loads without errors

#### TC-FE-04: Role-based UI
1. Login as `mechanic@demo.local`
2. **Expected:** No "Add Aircraft" button on Fleet page, no "Add User" button on Team page, can see "Complete Task" on assigned tasks

#### TC-FE-05: Responsive design
1. Test at viewport widths: 375px (mobile), 768px (tablet), 1024px (desktop), 1440px (wide)
2. **Expected:** Layout adjusts, mobile navigation works, tables become scrollable or card-based

#### TC-FE-06: Session expiry
1. Login, wait 15+ minutes (or manually clear the access token from localStorage)
2. Perform an action
3. **Expected:** Token refresh happens automatically. If refresh token also expired, redirected to login.

#### TC-FE-07: WebSocket real-time updates
1. Login in two browser windows
2. Create a task in window 1
3. **Expected:** Task list in window 2 updates automatically (if WebSocket is connected)

---

## 6. RBAC Test Matrix

Test each role against each operation. Login as the role, attempt the operation.

**Legend:** ✅ = Allowed, ❌ = 403 Forbidden

| Operation | admin | tenant_admin | scheduler | mechanic | auditor |
|-----------|-------|-------------|-----------|----------|---------|
| **Dashboard** | | | | | |
| View dashboard | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Fleet** | | | | | |
| List aircraft | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create aircraft | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update aircraft | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete aircraft | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Maintenance** | | | | | |
| List tasks | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create task | ✅ | ✅ | ✅ | ❌ | ❌ |
| Update task | ✅ | ✅ | ✅ | ❌ | ❌ |
| Transition task state | ✅ | ✅ | ✅ | ✅ | ❌ |
| Assign task to mechanic | ✅ | ✅ | ✅ | ❌ | ❌ |
| Delete task | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Parts** | | | | | |
| List parts | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create part definition | ✅ | ✅ | ❌ | ❌ | ❌ |
| Create part item | ✅ | ✅ | ❌ | ❌ | ❌ |
| Order/reserve parts | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Users** | | | | | |
| List users | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create user | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update user | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete user | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Compliance** | | | | | |
| List compliance items | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create compliance item | ✅ | ✅ | ✅ | ❌ | ❌ |
| Sign off compliance | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Reports** | | | | | |
| View reports | ✅ | ✅ | ✅ | ❌ | ✅ |
| Export reports | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Audit** | | | | | |
| View audit logs | ✅ | ✅ | ✅ | ✅ | ✅ |
| Export audit logs | ✅ | ✅ | ✅ | ❌ | ✅ |
| **Organization** | | | | | |
| Create organization | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update organization | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Settings** | | | | | |
| View settings | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage settings | ✅ | ✅ | ❌ | ❌ | ❌ |

> **Note:** The `admin` role always bypasses role checks in the backend (`Role.IsAdmin()` returns true). The RBAC matrix above reflects frontend permission definitions in `src/lib/rbac.ts`. Backend enforcement may differ for some operations — verify both layers.

---

## 7. Non-Functional Testing

### 7.1 Security — Priority: CRITICAL

#### TC-SEC-01: SQL injection via WAF
```bash
curl -s "https://amss-api-uat.duckdns.org/api/v1/aircraft?id=1'+OR+1=1--" \
  -H "Authorization: Bearer <token>"
```
**Expected:** 403 from SafeLine WAF (blocked before reaching backend)

#### TC-SEC-02: XSS via WAF
```bash
curl -s "https://amss-api-uat.duckdns.org/api/v1/aircraft" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"tail_number": "<script>alert(1)</script>"}'
```
**Expected:** 403 from SafeLine WAF

#### TC-SEC-03: Path traversal via WAF
```bash
curl -s "https://amss-api-uat.duckdns.org/../../etc/passwd"
```
**Expected:** 403 from SafeLine WAF

#### TC-SEC-04: Unknown JSON fields rejected
```bash
curl -s https://amss-api-uat.duckdns.org/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"org_id": "...", "email": "admin@demo.local", "password": "ChangeMe123!", "extra_field": "test"}'
```
**Expected:** 400 validation error (DisallowUnknownFields)

#### TC-SEC-05: CORS policy
```bash
curl -s -X OPTIONS https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Origin: https://evil.com" \
  -H "Access-Control-Request-Method: GET" -I
```
**Expected:** No `Access-Control-Allow-Origin` header for unauthorized origins.

```bash
curl -s -X OPTIONS https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Origin: https://amss.leoulgirma.com" \
  -H "Access-Control-Request-Method: GET" -I
```
**Expected:** `Access-Control-Allow-Origin: https://amss.leoulgirma.com`

#### TC-SEC-06: Rate limiting
1. Send 101 requests in 60 seconds to `GET /aircraft`
2. **Expected:** Request 101 returns 429 with `Retry-After` and `X-RateLimit-Remaining: 0`

#### TC-SEC-07: JWT token structure
1. Decode a valid access token (base64)
2. **Expected:** Contains `sub` (user_id), `org_id`, `role`, `token_type: "access"`, `exp` (15 min from issue)

### 7.2 Performance — Priority: MEDIUM

#### TC-PERF-01: API response times
1. Measure response time for: `/health`, `/aircraft`, `/maintenance-tasks`, `/audit-logs`
2. **Target:** < 200ms for simple GETs, < 500ms for filtered/paginated queries

#### TC-PERF-02: Concurrent requests
1. Send 50 concurrent `GET /aircraft` requests
2. **Expected:** All return 200, no timeouts, response time < 1s

#### TC-PERF-03: Rate limiter accuracy
1. Send exactly 100 requests within a 60-second window
2. **Expected:** All 100 succeed. Request 101 gets 429.

### 7.3 Frontend Quality — Priority: MEDIUM

#### TC-FQ-01: PWA installable
1. Open `https://amss.leoulgirma.com` in Chrome
2. **Expected:** Browser shows "Install" option in address bar

#### TC-FQ-02: Offline mode
1. Install as PWA, then go offline
2. **Expected:** Previously loaded pages show cached data, new API calls fail gracefully

#### TC-FQ-03: Cross-browser
1. Test on Chrome, Firefox, Safari (latest versions)
2. **Expected:** Layout, navigation, and core interactions work on all three

#### TC-FQ-04: Console errors
1. Open DevTools, navigate through all pages
2. **Expected:** No JavaScript errors in console during normal operation

### 7.4 Infrastructure — Priority: HIGH

#### TC-INFRA-01: Health endpoint
```bash
curl -s https://amss-api-uat.duckdns.org/health
```
**Expected:** 200

#### TC-INFRA-02: Readiness endpoint
```bash
curl -s https://amss-api-uat.duckdns.org/ready
```
**Expected:** 200 (confirms PostgreSQL and Redis are reachable)

#### TC-INFRA-03: Metrics endpoint
```bash
curl -s https://amss-api-uat.duckdns.org/metrics
```
**Expected:** Prometheus-formatted metrics with HTTP request counters

#### TC-INFRA-04: SSL certificate validity
```bash
echo | openssl s_client -connect amss-api-uat.duckdns.org:443 -servername amss-api-uat.duckdns.org 2>/dev/null | openssl x509 -noout -dates
```
**Expected:** Certificate not expired, covers the correct domain

---

## 8. Known Issues & Limitations

### SafeLine WAF Config Regeneration
When the SafeLine management service regenerates tengine configs (e.g., after adding a site through the web UI), it overwrites custom SSL directives and WAF detection blocks (`@safeline`, `@safelinex`). These must be manually re-added to `/etc/nginx/sites-enabled/IF_backend_3` and `IF_backend_4` inside the tengine container. See `SAFELINE-WAF-FIX-DOCUMENTATION.md` for details.

### Redis Graceful Degradation
If Redis is down, the API still functions but:
- Rate limiting is disabled (all requests pass through)
- Distributed caching is unavailable
- WebSocket pub/sub may not work across pods

### Soft Delete Caveats
All entities use soft delete (`deleted_at` timestamp). Soft-deleted records:
- Are excluded from normal queries
- Still occupy unique constraint space (e.g., a deleted aircraft's tail number cannot be reused while soft-deleted)
- Unique constraints use `WHERE deleted_at IS NULL` — but verify this for edge cases

### Maintenance Task Overlap Constraint
The `maintenance_tasks_no_overlap` exclusion constraint uses GiST indexing on `active_window` (a generated column from `start_time` and `end_time`). This prevents overlapping maintenance windows for the same aircraft. Only applies to non-deleted, non-cancelled tasks.

### Frontend Demo Mode
Demo mode uses local Redux state with mock data. It does not test any backend API functionality. Always test with real API login for actual QA.

### Organization Policies
Each org has configurable rate limits and retention settings in the `org_policies` table. The defaults are:
- Rate limit: 100 req/min
- Login rate limit: 10 req/min
- Webhook attempts: 10 max
- Audit log retention: 365 days

### Deferred Items (from senior review)
- Pod resource limits are set but not production-tuned
- Kubernetes secrets are stored as base64 (not encrypted at rest with external KMS)
- No horizontal pod autoscaler configured yet

---

## 9. Bug Reporting

### Severity Definitions

| Severity | Definition | AMSS Examples |
|----------|-----------|---------------|
| **Critical** | System down, data loss, security breach | Auth bypass, data leak between tenants, compliance sign-off fails |
| **High** | Core feature broken, no workaround | Cannot create tasks, aircraft CRUD fails, audit log not recording |
| **Medium** | Feature degraded, workaround exists | Kanban drag-drop glitchy, report data slightly off, slow response |
| **Low** | Cosmetic, minor UX issues | UI alignment, typo, tooltip missing |

### Bug Report Template

```
Title: [Short description]
Severity: Critical | High | Medium | Low
Environment: UAT / Local
Browser (if frontend): Chrome 122 / Firefox 135 / Safari 18

Steps to Reproduce:
1. Login as [role]
2. Navigate to [page/endpoint]
3. Perform [action]
4. Observe [result]

Expected Result: [What should happen]
Actual Result: [What actually happened]

Evidence:
- Screenshot/video: [attached]
- curl command: [if API]
- Response body: [if API]
- Console errors: [if frontend]

Additional Context:
- Request ID (X-Request-ID header): [if available]
- Timestamp: [when it occurred]
```

### LLM-Friendly Bug Report Format

If testing via LLM, output each bug as JSON:

```json
{
  "id": "BUG-001",
  "title": "Mechanic role can create aircraft via API",
  "severity": "high",
  "category": "authorization",
  "feature": "fleet",
  "environment": "UAT",
  "steps": [
    "POST /auth/login as mechanic@demo.local",
    "POST /aircraft with valid body",
    "Observe: 201 Created instead of 403 Forbidden"
  ],
  "expected": "403 Forbidden - mechanic role should not create aircraft",
  "actual": "201 Created - aircraft was created successfully",
  "endpoint": "POST /api/v1/aircraft",
  "request_id": "abc-123-def",
  "notes": "Backend RBAC may not enforce fleet management restrictions for mechanic role"
}
```

---

## 10. curl Cheatsheet

All commands use the UAT API. Replace `$TOKEN` with a valid access token.

### Get Token

```bash
# Step 1: Lookup
curl -s https://amss-api-uat.duckdns.org/api/v1/auth/lookup \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@demo.local"}'

# Step 2: Login
TOKEN=$(curl -s https://amss-api-uat.duckdns.org/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "4cb97629-c58a-415d-bf9c-b400bb5e3d84",
    "email": "admin@demo.local",
    "password": "ChangeMe123!"
  }' | jq -r '.access_token')

echo $TOKEN
```

### Aircraft

```bash
# List
curl -s https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Authorization: Bearer $TOKEN" | jq

# Create
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tail_number": "N999QA", "model": "QA Test Aircraft", "status": "operational", "capacity_slots": 2}'

# Update
curl -s -X PATCH https://amss-api-uat.duckdns.org/api/v1/aircraft/<id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "maintenance"}'

# Delete
curl -s -X DELETE https://amss-api-uat.duckdns.org/api/v1/aircraft/<id> \
  -H "Authorization: Bearer $TOKEN"
```

### Maintenance Tasks

```bash
# List
curl -s https://amss-api-uat.duckdns.org/api/v1/maintenance-tasks \
  -H "Authorization: Bearer $TOKEN" | jq

# Create
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/maintenance-tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "aircraft_id": "<aircraft_uuid>",
    "type": "inspection",
    "start_time": "2026-02-10T08:00:00Z",
    "end_time": "2026-02-10T14:00:00Z",
    "notes": "QA test task"
  }'

# Transition state
curl -s -X PATCH https://amss-api-uat.duckdns.org/api/v1/maintenance-tasks/<id>/state \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"state": "in_progress"}'
```

### Parts

```bash
# List definitions
curl -s https://amss-api-uat.duckdns.org/api/v1/part-definitions \
  -H "Authorization: Bearer $TOKEN" | jq

# List items
curl -s https://amss-api-uat.duckdns.org/api/v1/part-items \
  -H "Authorization: Bearer $TOKEN" | jq

# Reserve part
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/part-reservations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"task_id": "<task_uuid>", "part_item_id": "<part_item_uuid>"}'
```

### Compliance

```bash
# List
curl -s https://amss-api-uat.duckdns.org/api/v1/compliance-items \
  -H "Authorization: Bearer $TOKEN" | jq

# Sign off
curl -s -X PATCH https://amss-api-uat.duckdns.org/api/v1/compliance-items/<id>/sign-off \
  -H "Authorization: Bearer $TOKEN"
```

### Audit & Reports

```bash
# Audit logs
curl -s "https://amss-api-uat.duckdns.org/api/v1/audit-logs?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq

# Export audit logs
curl -s "https://amss-api-uat.duckdns.org/api/v1/audit-logs/export" \
  -H "Authorization: Bearer $TOKEN" -o audit-export.csv

# Fleet summary
curl -s https://amss-api-uat.duckdns.org/api/v1/reports/summary \
  -H "Authorization: Bearer $TOKEN" | jq

# Compliance report
curl -s https://amss-api-uat.duckdns.org/api/v1/reports/compliance \
  -H "Authorization: Bearer $TOKEN" | jq
```

### Webhooks

```bash
# Create webhook (use webhook.site for testing)
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://webhook.site/<your-uuid>", "events": ["task.created", "task.updated"]}'

# Test webhook
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/webhooks/<id>/test \
  -H "Authorization: Bearer $TOKEN"
```

### Health & Infrastructure

```bash
curl -s https://amss-api-uat.duckdns.org/health
curl -s https://amss-api-uat.duckdns.org/ready
curl -s https://amss-api-uat.duckdns.org/metrics | head -20
```

### Idempotent Request

```bash
curl -s -X POST https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-unique-key-123" \
  -d '{"tail_number": "N888QA", "model": "Idempotent Test", "status": "operational", "capacity_slots": 1}'

# Sending the same request again with same key returns cached response
# Sending with same key but different body returns 409 Conflict
```

---

## Appendix A: Database Schema Quick Reference

### Core Tables

| Table | Key Fields | Notes |
|-------|-----------|-------|
| `organizations` | id, name | Multi-tenant root |
| `users` | id, org_id, email, role, password_hash | Unique email per org |
| `aircraft` | id, org_id, tail_number, model, status, flight_hours_total, cycles_total | Unique tail per org |
| `maintenance_programs` | id, org_id, aircraft_id (optional), name, interval_type, interval_value | Fleet-wide if aircraft_id is NULL |
| `maintenance_tasks` | id, org_id, aircraft_id, program_id, type, state, start_time, end_time, assigned_mechanic_id | Overlap constraint on active_window |
| `part_definitions` | id, org_id, name, category | Unique name per org |
| `part_items` | id, org_id, part_definition_id, serial_number, status, expiry_date | Unique serial per org |
| `part_reservations` | id, org_id, task_id, part_item_id, state, quantity | One active reservation per part item |
| `compliance_items` | id, org_id, task_id, description, result, sign_off_user_id, sign_off_time | Immutable after sign-off |
| `audit_logs` | id, org_id, entity_type, entity_id, action, user_id, details (JSONB) | Immutable (trigger-enforced) |
| `webhooks` | id, org_id, url, events[], secret | HMAC-signed deliveries |
| `webhook_deliveries` | id, webhook_id, event_id, status, attempt_count | Retry with backoff |
| `imports` | id, org_id, type, status, file_name, created_by, summary | Async processing |
| `import_rows` | id, import_id, row_number, raw (JSONB), status, errors (JSONB) | Per-row validation |
| `outbox_events` | id, org_id, event_type, payload, dedupe_key | Transactional outbox pattern |
| `idempotency_keys` | id, org_id, key, endpoint, request_hash, response_body, expires_at | 24-hour TTL |
| `org_policies` | org_id, retention_interval, api_rate_limit_per_min, max_webhook_attempts | Per-org config |
| `refresh_tokens` | id, org_id, user_id, token_hash, expires_at, revoked_at, rotated_from | Token rotation support |

### Enums

| Enum | Values |
|------|--------|
| `user_role` | admin, tenant_admin, scheduler, mechanic, auditor |
| `aircraft_status` | operational, maintenance, grounded |
| `maintenance_task_type` | inspection, repair, overhaul |
| `maintenance_task_state` | scheduled, in_progress, completed, cancelled |
| `part_item_status` | in_stock, used, disposed |
| `part_reservation_state` | reserved, used, released |
| `compliance_result` | pass, fail, pending |
| `audit_action` | create, update, delete, state_change |
| `import_type` | aircraft, parts, programs |
| `import_status` | pending, validating, applying, completed, failed |
| `import_row_status` | pending, valid, invalid, applied |
| `webhook_delivery_status` | pending, delivered, failed |

### Key Constraints

- `maintenance_tasks_no_overlap` — GiST exclusion on (org_id, aircraft_id, active_window) prevents overlapping maintenance windows for the same aircraft
- `audit_logs_immutable` — trigger rejects UPDATE/DELETE on audit_logs
- `compliance_items_immutable` — trigger rejects UPDATE after sign_off_time is set
- `part_reservations_active_idx` — only one active (non-released) reservation per part_item
- All `_uniq` indexes exclude soft-deleted records (`WHERE deleted_at IS NULL`)

---

## Appendix B: WebSocket Events

Connect: `wss://amss-api-uat.duckdns.org/ws?org_id=<org_id>&user_id=<user_id>`

| Event | Trigger | Payload |
|-------|---------|---------|
| `task:created` | New task created | Task object |
| `task:updated` | Task fields changed | Task object |
| `task:deleted` | Task soft-deleted | `{id: "..."}` |
| `task:status_changed` | Task state transition | `{id: "...", state: "..."}` |
| `aircraft:status_changed` | Aircraft status changed | Aircraft object |
| `notification` | System notification | Notification object |
| `part:low_stock` | Part inventory low | Part definition + count |
| `user:online` | User connected | `{user_id: "..."}` |
| `user:offline` | User disconnected | `{user_id: "..."}` |

---

## Appendix C: Seeded Test Data Reference

After running base migrations (`make migrate-up`):

| Entity | Count | Details |
|--------|-------|---------|
| Organizations | 1 | Demo Airline |
| Users | 5 | One per role |
| Aircraft | 1 | N100AM, Boeing 737-800, operational |
| Programs | 1 | A-check, calendar, 90 days |
| Tasks | 1 | Inspection, scheduled, assigned to mechanic |
| Compliance | 1 | Safety checklist, pending |

After running extended seed (`scripts/seed_test_data.sql`):

| Entity | Count | Details |
|--------|-------|---------|
| Aircraft | 6 | N100AM–N600AM (3 operational, 1 maintenance, 1 grounded, 1 original) |
| Part Definitions | 6 | Brake Assembly, Engine Fan Blade, Hydraulic Pump, Avionics Display, Tire Assembly, APU Starter |
| Part Items | 8 | Various statuses (in_stock, used, disposed) with expiry dates |
| Tasks | 5 | Across all states (scheduled, in_progress, completed) |
| Compliance Items | 7+ | Mix of pass, pending, with and without sign-off |
| Programs | 5+ | A-Check (hours), B-Check (hours), C-Check (calendar), Landing Gear (cycles) |
