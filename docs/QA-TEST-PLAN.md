# AMSS QA Test Plan - Bottom-Up Approach

## Current State Summary
- **API**: Live at `https://amss-api-uat.duckdns.org` (health/ready both 200)
- **Frontend**: Live at `https://amss.leoulgirma.com`
- **Database**: 19 tables, seed data present (1 org, 5 users, 6 aircraft, 6 programs, 5 tasks, 6 part defs, 8 part items)
- **Auth**: Working after password reset (seed hashes were stale)
- **Bug already found**: `tenant-admin@demo.local` was stored with role `admin` instead of `tenant_admin` (fixed)

---

## Phase 1: Infrastructure & Health (Smallest)
Quick smoke tests - just verify the plumbing works.

1. `GET /health` returns 200
2. `GET /ready` returns 200 (DB + Redis connectivity)
3. `GET /openapi.yaml` returns valid spec
4. `GET /docs` returns Swagger UI
5. `GET /metrics` returns Prometheus metrics (if enabled)
6. CORS headers present on preflight OPTIONS requests
7. `X-Request-Id` header returned on every response
8. Unknown routes return proper 404 (not stack trace)
9. WebSocket `/ws` endpoint accepts upgrade

---

## Phase 2: Authentication & Authorization
Must work before we can test anything else.

### 2a. Auth Flow
1. `POST /auth/lookup` - valid email returns org list
2. `POST /auth/lookup` - unknown email returns empty/error
3. `POST /auth/login` - valid creds return access + refresh tokens
4. `POST /auth/login` - wrong password returns 401
5. `POST /auth/login` - deleted user returns 401
6. `POST /auth/login` - rate limiting kicks in after repeated failures
7. `POST /auth/refresh` - valid refresh token returns new access token
8. `POST /auth/refresh` - expired/invalid refresh token returns 401
9. `POST /auth/logout` - invalidates refresh token
10. `GET /auth/me` - returns current user info
11. `GET /auth/me` - expired access token returns 401
12. JWT claims contain correct org_id, role, sub, exp

### 2b. RBAC (test per role: admin, tenant_admin, scheduler, mechanic, auditor)
Login as each role and attempt operations. Verify:
- admin: full CRUD on everything
- tenant_admin: manage users and org settings
- scheduler: create/manage tasks, read aircraft
- mechanic: update assigned tasks, manage compliance/parts
- auditor: read-only on everything, access audit logs

---

## Phase 3: CRUD Operations (Per Resource, Simple to Complex)
For each resource, test the standard operations:

### 3a. Organizations (simplest - single entity)
1. `GET /organizations` - list (should return Demo Airline)
2. `GET /organizations/{id}` - get by ID
3. `POST /organizations` - create new org
4. `PATCH /organizations/{id}` - update name
5. Invalid UUID in path returns 400
6. Non-existent ID returns 404

### 3b. Users
1. `GET /users` - list all users in org
2. `GET /users/{id}` - get specific user
3. `POST /users` - create user (all 5 roles)
4. `PATCH /users/{id}` - update email/role
5. `DELETE /users/{id}` - soft delete
6. Duplicate email in same org returns conflict
7. Invalid role value returns validation error
8. Password validation rules enforced on create

### 3c. Aircraft
1. `GET /aircraft` - list all (should return 6)
2. `GET /aircraft/{id}` - get with full details
3. `POST /aircraft` - create new aircraft
4. `PATCH /aircraft/{id}` - update model, status, etc.
5. `DELETE /aircraft/{id}` - soft delete
6. Duplicate tail_number in same org returns conflict
7. Status enum validation (operational, maintenance, grounded)
8. capacity_slots constraint validation

### 3d. Maintenance Programs
1. `GET /maintenance-programs` - list all
2. `GET /maintenance-programs/{id}` - get by ID
3. `POST /maintenance-programs` - create linked to aircraft
4. `PATCH /maintenance-programs/{id}` - update interval
5. `DELETE /maintenance-programs/{id}` - soft delete
6. Must reference valid aircraft_id
7. interval_type validation (calendar, flight_hours, cycles)

### 3e. Part Definitions (catalog)
1. `GET /part-definitions` - list catalog
2. `POST /part-definitions` - create definition
3. `PATCH /part-definitions/{id}` - update
4. `DELETE /part-definitions/{id}` - soft delete
5. Validation on required fields

### 3f. Part Items (inventory)
1. `GET /part-items` - list inventory
2. `POST /part-items` - add item to inventory
3. `PATCH /part-items/{id}` - update status/quantity
4. `DELETE /part-items/{id}` - soft delete
5. Must reference valid part_definition

---

## Phase 4: Business Logic (State Machines & Relationships)
Complex workflows that involve multiple resources.

### 4a. Maintenance Task State Machine
1. `POST /maintenance-tasks` - create task (starts in `scheduled` state)
2. `PATCH /maintenance-tasks/{id}/state` - transition scheduled -> in_progress
3. `PATCH /maintenance-tasks/{id}/state` - transition in_progress -> completed
4. Invalid state transitions return error (e.g., scheduled -> completed directly)
5. Task must reference valid aircraft and program
6. assigned_mechanic_id must be a valid mechanic user
7. capacity_slots: can't schedule more tasks than aircraft allows
8. Time range validation (end_time > start_time)

### 4b. Part Reservations
1. `POST /part-reservations` - reserve part for a task
2. `PATCH /part-reservations/{id}/state` - state transitions
3. Can't reserve more than available quantity
4. Distributed lock prevents race conditions
5. Reservation tied to valid task and part_item

### 4c. Compliance Items
1. `GET /compliance-items` - list
2. `POST /compliance-items` - create linked to task
3. `PATCH /compliance-items/{id}` - update description/result
4. `PATCH /compliance-items/{id}/sign-off` - sign off (only authorized users)
5. Must reference valid task

### 4d. CSV Import Pipeline
1. `POST /imports/csv` - upload CSV file
2. `GET /imports/{id}` - check import status
3. `GET /imports/{id}/rows` - see parsed rows
4. Invalid CSV format handling
5. Worker processes import asynchronously

### 4e. Webhooks
1. `POST /webhooks` - register webhook URL
2. `GET /webhooks` - list registered webhooks
3. `DELETE /webhooks/{id}` - remove webhook
4. `POST /webhooks/{id}/test` - fire test event
5. HTTPS required in production mode (not UAT)

### 4f. Reports
1. `GET /reports/summary` - dashboard summary
2. `GET /reports/compliance` - compliance report
3. Verify data accuracy against known DB state

### 4g. Audit Logs
1. `GET /audit-logs` - list logs
2. `GET /audit-logs/export` - export CSV/JSON
3. Verify actions create audit entries (create user, update task, etc.)

---

## Phase 5: API Quality & Edge Cases
Cross-cutting concerns tested across multiple endpoints.

1. **Pagination**: Verify `limit`, `offset`, `cursor` params work on list endpoints
2. **Filtering/Sorting**: Test query params for list endpoints
3. **Idempotency**: Send same request with `Idempotency-Key` header twice, verify no duplicates
4. **Rate Limiting**: Hit limits, verify 429 response with correct headers (X-RateLimit-Remaining, Retry-After)
5. **Validation**: Malformed JSON, missing required fields, wrong types
6. **Content-Type**: Wrong content type returns proper error
7. **Org isolation**: User from Org A can't see Org B's data
8. **Soft delete**: Deleted records don't appear in list endpoints
9. **Timestamps**: created_at/updated_at properly set and updated

---

## Phase 6: Security Testing
1. **No auth**: Protected endpoints return 401 without token
2. **Expired token**: Returns 401, not 500
3. **Tampered JWT**: Modified payload returns 401
4. **SQL injection**: Malicious strings in query params and body
5. **XSS in stored fields**: Script tags in names/descriptions are escaped
6. **CORS**: Only allowed origins get CORS headers
7. **Path traversal**: `../../etc/passwd` in file-related endpoints
8. **Mass assignment**: Extra fields in JSON body are ignored (DisallowUnknownFields)
9. **WAF (SafeLine)**: Verify WAF blocks obvious attack patterns

---

## Phase 7: Frontend Smoke Tests
Quick manual/scripted checks on the web UI.

1. Login page loads, form works
2. Dashboard loads after auth
3. Aircraft list displays correctly
4. Create/edit/delete flows for each resource
5. Task state transitions via UI
6. Responsive layout (mobile/desktop)
7. Error states display properly
8. Logout works

---

## Execution Order

We start from Phase 1 and work through sequentially. Each phase builds confidence for the next. Within each phase, I'll write and execute `curl`-based test scripts, record PASS/FAIL, and document any bugs found.

**Output**: A test results report with:
- Total tests run / passed / failed
- Bug list with severity, steps to reproduce, and evidence
- Recommendations
