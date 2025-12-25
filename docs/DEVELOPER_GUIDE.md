# AMSS Developer Onboarding Guide

**Welcome to the AMSS (Aircraft Maintenance Scheduling System) development team!**

This guide will help you understand the system architecture, how all the pieces fit together, and how to contribute effectively.

---

## Table of Contents

1. [System Overview & Problem Domain](#1-system-overview--problem-domain)
2. [Architecture Deep Dive](#2-architecture-deep-dive)
3. [User Journey Flows](#3-user-journey-flows)
4. [Feature Interconnections Map](#4-feature-interconnections-map)
5. [Database Schema Overview](#5-database-schema-overview)
6. [Development Workflows](#6-development-workflows)
7. [Code Organization & Patterns](#7-code-organization--patterns)
8. [Common Development Tasks](#8-common-development-tasks)
9. [Key Technical Decisions](#9-key-technical-decisions)
10. [Contribution Guidelines](#10-contribution-guidelines)

---

## 1. System Overview & Problem Domain

### What Problems Does AMSS Solve?

**The Aircraft Maintenance Challenge:**

Aircraft maintenance is one of the most regulated and complex operational challenges in aviation. Organizations face:

1. **Compliance Complexity**: FAA/EASA regulations require strict adherence to maintenance schedules based on flight hours, cycles (takeoffs/landings), and calendar time.

2. **Scheduling Conflicts**: Maintenance tasks auto-generate based on aircraft usage, creating dynamic scheduling needs that conflict with flight operations.

3. **Audit Requirements**: Every maintenance action must be traceable with complete audit trails showing who did what and when.

4. **Parts Coordination**: Technicians need parts reserved before task execution, requiring real-time inventory visibility.

5. **Multi-System Integration**: Maintenance systems need to notify external systems (flight scheduling, inventory, billing) via webhooks.

6. **Bulk Operations**: Operators manage fleets of aircraft, requiring CSV import capabilities for initial setup and periodic updates.

### Who Uses AMSS?

**Primary Users:**

- **Maintenance Planners**: Create maintenance programs, schedule tasks, monitor compliance
- **Technicians**: Execute tasks, reserve parts, log work performed
- **Compliance Officers**: Review audit trails, generate reports, ensure regulatory adherence
- **Fleet Operators**: Manage multiple aircraft, bulk operations, system integration
- **Automation Systems**: External systems consuming webhooks and APIs

### Business Value Proposition

AMSS provides:
- ✅ **Automated Compliance**: Tasks auto-generate based on aircraft usage, eliminating manual tracking
- ✅ **Complete Audit Trails**: Every action logged with timestamps, user IDs, and change history
- ✅ **Reliable Integrations**: Outbox pattern ensures webhook delivery even during failures
- ✅ **Safe Operations**: Idempotency prevents duplicate tasks during retries
- ✅ **Scalability**: Separate server/worker architecture handles thousands of aircraft

---

## 2. Architecture Deep Dive

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │   ingress-nginx       │ (TLS termination via cert-manager)
         │   ports: 80, 443      │
         └───────────┬───────────┘
                     │
                     ▼
         ┌───────────────────────┐
         │  amss-server          │ (REST + gRPC APIs)
         │  Kubernetes Service   │
         │  ClusterIP: 8080      │
         └───────────┬───────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
         ▼                       ▼
┌────────────────┐      ┌────────────────┐
│  amss-server   │      │  amss-worker   │
│  Deployment    │      │  Deployment    │
│  (1 replica)   │      │  (1 replica)   │
└────────┬───────┘      └────────┬───────┘
         │                       │
         │       ┌───────────────┴───────────────┐
         │       │                               │
         ▼       ▼                               ▼
    ┌─────────────────┐                  ┌──────────────┐
    │   PostgreSQL    │                  │    Redis     │
    │   (Docker)      │                  │  (Docker)    │
    │   port: 5455    │                  │  port: 6379  │
    │   Firewall: ✓   │                  │  Firewall: ✓ │
    └─────────────────┘                  └──────────────┘
```

### Component Responsibilities

#### amss-server (REST + gRPC API)

**Purpose**: Synchronous API for client interactions

**Responsibilities:**
- HTTP REST endpoints (`/api/v1/*`)
- gRPC services (`:9090`)
- Request authentication (JWT validation)
- Request authorization (RBAC checks)
- Rate limiting (per organization)
- Idempotency handling (deduplication via keys)
- OpenAPI documentation (`/docs`, `/openapi.yaml`)
- Metrics (`/metrics`) and health checks (`/health`, `/ready`)

**Does NOT:**
- Process long-running operations
- Send webhooks directly
- Process CSV imports

#### amss-worker (Background Jobs)

**Purpose**: Asynchronous job processing

**Jobs:**
1. **Outbox Publisher**: Polls `outbox` table, publishes events to Redis streams
2. **Webhook Dispatcher**: Consumes Redis stream, sends HTTP POSTs to webhook URLs with retries
3. **Import Processor**: Reads CSV files, validates, inserts to database
4. **Program Generator**: Generates maintenance tasks based on aircraft hours/cycles/calendar
5. **Retention Cleaner** (planned): Deletes old audit logs based on retention policy

**Why Separate Worker?**
- Failures in webhook delivery don't affect API availability
- Long-running imports don't block API requests
- Independently scalable (can run multiple workers)
- Retry logic isolated from request handlers

#### PostgreSQL (System of Record)

**Purpose**: Primary database for all persistent data

**Key Features Used:**
- ACID transactions (critical for audit trails)
- Foreign key constraints (data integrity)
- Partial indexes (performance optimization)
- Row-level locks (concurrency control)
- JSONB columns (flexible metadata storage)

#### Redis (Cache & Queue)

**Purpose**: Rate limiting and async job queues

**Usage:**
1. **Rate Limiting**: Token bucket counters per org (e.g., `ratelimit:org:<uuid>`)
2. **Redis Streams**: Job queue for webhook delivery
3. **Idempotency Cache**: Track request IDs to prevent duplicate processing

**Why Not Use as Primary Store?**
- Not ACID-compliant (can lose data on crash)
- No complex queries (joins, aggregations)
- Redis is for speed, Postgres for durability

---

### Data Flow: Complete Request Lifecycle

#### Example: Create Maintenance Task (Synchronous)

```
1. Client → HTTP POST /api/v1/tasks
   Headers: Authorization: Bearer <jwt>, Idempotency-Key: <uuid>

2. ingress-nginx → Routes to amss-server ClusterIP:8080

3. Middleware Chain (amss-server):
   a. RequestID: Generate unique request ID
   b. Logging: Log request start
   c. Auth: Validate JWT signature, extract user/org claims
   d. RBAC: Check if user has "tasks:write" permission
   e. RateLimit: Check org rate limit in Redis (100 req/min)
   f. Idempotency: Check if Idempotency-Key already processed (Redis)

4. Handler:
   a. Parse JSON body → domain.Task struct
   b. Validate fields (required, formats, business rules)

5. Service Layer:
   a. Begin Postgres transaction
   b. Check if aircraft exists (FK validation)
   c. Check if maintenance program exists
   d. Insert task into `tasks` table
   e. Insert audit log entry
   f. Insert outbox event (task.created)
   g. Store idempotency key → result mapping in Redis
   h. Commit transaction

6. Response → HTTP 201 Created with task JSON

7. Async (background):
   a. Outbox Publisher (worker) polls `outbox` table every 100ms
   b. Finds new task.created event
   c. Publishes to Redis stream: "events:tasks"
   d. Marks outbox entry as published

   e. Webhook Dispatcher (worker) consumes Redis stream
   f. Looks up webhook subscriptions for org
   g. Sends HTTP POST to webhook URL
   h. Retries on failure (exponential backoff)
   i. Logs delivery status
```

#### Example: Bulk Import Aircraft (Asynchronous)

```
1. Client → HTTP POST /api/v1/imports/aircraft
   Body: multipart/form-data with CSV file

2. Handler:
   a. Save CSV to disk: /data/imports/<uuid>.csv
   b. Insert import job to `import_jobs` table (status: pending)
   c. Insert outbox event (import.queued)
   d. Return HTTP 202 Accepted with job ID

3. Outbox Publisher → Publishes import.queued to Redis stream

4. Import Processor (worker):
   a. Consumes import.queued event
   b. Reads CSV from disk
   c. Validates headers, row format
   d. Begin transaction
   e. For each row:
      - Parse aircraft data
      - Validate business rules
      - Insert or update `aircraft` table
      - Log success/failure
   f. Commit transaction
   g. Update import_jobs status → completed (or failed)
   h. Insert outbox event (import.completed)

5. Webhook Dispatcher → Sends import.completed webhook to client
```

---

## 3. User Journey Flows

### Journey 1: Admin Onboarding a New Organization

```
┌─────────────────────────────────────────────────────────────┐
│ 1. System Admin Creates Organization                        │
│    POST /api/v1/organizations                               │
│    → org_id: "acme-aviation"                                │
│    → Database: Insert into organizations table              │
│    → Audit: Log organization.created event                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Admin Creates First User                                 │
│    POST /api/v1/users                                       │
│    → email: "admin@acme.aero", role: "admin"               │
│    → Password hashed with bcrypt                            │
│    → Database: Insert into users table                      │
│    → Email: Send welcome email (if configured)              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. User Logs In                                             │
│    POST /api/v1/auth/login                                  │
│    → Validates email + password (bcrypt compare)            │
│    → Generates JWT access token (15min TTL)                 │
│    → Generates refresh token (7 days TTL)                   │
│    → Returns both tokens                                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Admin Invites Team Members                               │
│    POST /api/v1/users (multiple times)                      │
│    → role: "planner", "technician", "viewer"                │
│    → Each user gets unique credentials                      │
│    → RBAC permissions automatically assigned by role        │
└─────────────────────────────────────────────────────────────┘
```

**Key Touchpoints:**
- Organizations table
- Users table
- Auth service (JWT generation)
- RBAC middleware (permission checks)
- Audit logs

---

### Journey 2: Planner Sets Up Maintenance Program

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Planner Adds Aircraft to System                          │
│    POST /api/v1/aircraft                                    │
│    {                                                         │
│      "registration": "N12345",                              │
│      "serial_number": "SN-9876",                            │
│      "current_hours": 1500.5,                               │
│      "current_cycles": 800                                  │
│    }                                                         │
│    → Database: Insert into aircraft table                   │
│    → Webhook: aircraft.created event sent                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Planner Creates Maintenance Program                      │
│    POST /api/v1/programs                                    │
│    {                                                         │
│      "name": "100-Hour Inspection",                         │
│      "interval_type": "hours",                              │
│      "interval_value": 100,                                 │
│      "description": "Routine 100-hour check per FAR 91.409" │
│    }                                                         │
│    → Database: Insert into programs table                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. System Auto-Generates First Task                         │
│    Program Generator Job (worker):                          │
│    - Runs every 5 minutes                                   │
│    - Queries aircraft with linked programs                  │
│    - Calculates: next_due = current_hours + interval        │
│    - If near threshold, generates task                      │
│                                                              │
│    Generated Task:                                          │
│    {                                                         │
│      "program_id": "<uuid>",                                │
│      "aircraft_id": "<uuid>",                               │
│      "due_at_hours": 1600.0,                                │
│      "status": "scheduled"                                  │
│    }                                                         │
│    → Database: Insert into tasks table                      │
│    → Webhook: task.created event sent                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Aircraft Flies (Hours Accumulate)                        │
│    PUT /api/v1/aircraft/:id                                 │
│    { "current_hours": 1595.0 }                              │
│    → Task now shows "overdue" (1595 < 1600 - threshold)     │
│    → Email/webhook alert sent                               │
└─────────────────────────────────────────────────────────────┘
```

**Key Touchpoints:**
- Aircraft table
- Programs table
- Tasks table (auto-generated)
- Program Generator job (worker)
- Webhook dispatcher (notifications)

---

### Journey 3: Technician Executes Maintenance Task

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Technician Views Assigned Tasks                          │
│    GET /api/v1/tasks?assigned_to=<user_id>&status=scheduled│
│    → Returns list of pending tasks                          │
│    → Shows: aircraft, program, due date, parts needed       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Technician Reserves Required Parts                       │
│    POST /api/v1/tasks/:id/reserve-parts                     │
│    {                                                         │
│      "part_reservations": [                                 │
│        { "part_id": "<uuid>", "quantity": 2 }               │
│      ]                                                       │
│    }                                                         │
│    → Transaction:                                           │
│      a. Check inventory availability                        │
│      b. Create part_reservations records                    │
│      c. Decrement inventory.available_quantity              │
│      d. Update task status → "in_progress"                  │
│    → If insufficient inventory: HTTP 409 Conflict           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Technician Performs Work                                 │
│    (Manual work outside system)                             │
│    - Inspects aircraft                                      │
│    - Replaces parts                                         │
│    - Tests systems                                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Technician Logs Completion                               │
│    PUT /api/v1/tasks/:id/complete                           │
│    {                                                         │
│      "completion_notes": "Replaced oil filter, spark plugs",│
│      "completed_at_hours": 1605.2,                          │
│      "completed_at_cycles": 815                             │
│    }                                                         │
│    → Transaction:                                           │
│      a. Update task status → "completed"                    │
│      b. Set completed_at timestamp                          │
│      c. Set completed_by_user_id                            │
│      d. Mark part_reservations as consumed                  │
│      e. Insert audit log entry                              │
│      f. Insert outbox event (task.completed)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. System Auto-Generates Next Task                          │
│    Program Generator (triggered by task.completed):         │
│    - Calculates next due: 1605.2 + 100 = 1705.2 hours      │
│    - Creates new task with due_at_hours: 1705.2            │
│    → Continuous maintenance cycle established               │
└─────────────────────────────────────────────────────────────┘
```

**Key Touchpoints:**
- Tasks table (status transitions)
- Parts inventory (reservations, availability)
- Part reservations table
- Audit logs (complete trail)
- Outbox + webhooks (external notifications)

---

### Journey 4: External System Receives Webhook

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Organization Configures Webhook                          │
│    POST /api/v1/webhooks                                    │
│    {                                                         │
│      "url": "https://billing.acme.aero/webhooks/amss",      │
│      "events": ["task.completed", "part.consumed"],         │
│      "secret": "webhook-signing-secret"                     │
│    }                                                         │
│    → Database: Insert into webhooks table                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Event Occurs (e.g., Task Completed)                      │
│    → Outbox entry created with event payload                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Outbox Publisher (Worker)                                │
│    - Polls outbox table every 100ms                         │
│    - Finds unpublished events                               │
│    - Publishes to Redis stream: "events:tasks"              │
│    - Marks outbox entry published_at timestamp              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Webhook Dispatcher (Worker)                              │
│    - Consumes from Redis stream                             │
│    - Looks up webhook subscriptions for this event type     │
│    - For each webhook:                                      │
│      a. Builds HTTP POST payload                            │
│      b. Signs payload with HMAC-SHA256 (webhook secret)     │
│      c. Sends POST request                                  │
│      d. Records delivery attempt                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. External System Receives Webhook                         │
│    POST https://billing.acme.aero/webhooks/amss             │
│    Headers:                                                  │
│      X-AMSS-Signature: sha256=<hmac>                        │
│      X-AMSS-Event: task.completed                           │
│    Body:                                                     │
│    {                                                         │
│      "event": "task.completed",                             │
│      "data": { "task_id": "<uuid>", ... }                   │
│    }                                                         │
│    → External system validates signature                    │
│    → Processes event (e.g., generates invoice)              │
│    → Returns HTTP 200 OK                                    │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Delivery Confirmed                                       │
│    → Webhook dispatcher marks delivery successful           │
│    → No retries needed                                      │
│                                                              │
│    If Failure (e.g., HTTP 500):                             │
│    → Retry with exponential backoff                         │
│    → Attempt 1: immediate                                   │
│    → Attempt 2: 1 minute later                              │
│    → Attempt 3: 5 minutes later                             │
│    → Attempt 4: 15 minutes later                            │
│    → After 10 failures: mark as failed, alert admin         │
└─────────────────────────────────────────────────────────────┘
```

**Key Touchpoints:**
- Webhooks table (subscriptions)
- Outbox table (event staging)
- Redis streams (event queue)
- Webhook dispatcher (delivery + retries)
- HMAC signing (security)

---

## 4. Feature Interconnections Map

This map shows how every major feature connects to others:

```
┌─────────────────────────────────────────────────────────────┐
│                      ORGANIZATIONS                           │
│  (Multi-tenancy: all data scoped to org_id)                 │
└────────────────────────┬────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    ┌────────┐     ┌─────────┐    ┌──────────┐
    │ USERS  │     │ AIRCRAFT│    │ PROGRAMS │
    │ & RBAC │     └────┬────┘    └────┬─────┘
    └────┬───┘          │              │
         │              │              │
         │              └──────┬───────┘
         │                     │
         │                     ▼
         │              ┌────────────┐
         │              │   TASKS    │ (Auto-generated)
         │              └──────┬─────┘
         │                     │
         │              ┌──────┴─────────────┐
         │              │                    │
         │              ▼                    ▼
         │        ┌──────────┐        ┌──────────┐
         │        │  PARTS   │◄───────│   PART   │
         │        │INVENTORY │        │RESERVES  │
         │        └──────────┘        └──────────┘
         │              │                    │
         ▼              │                    │
    ┌──────────┐        │                    │
    │  AUTH    │        │                    │
    │  (JWT)   │        │                    │
    └────┬─────┘        │                    │
         │              │                    │
         │              ▼                    │
         │        ┌──────────────┐           │
         └───────►│  AUDIT LOGS  │◄──────────┘
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │    OUTBOX    │ (Transactional)
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │ REDIS STREAM │
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │   WEBHOOKS   │ (External integration)
                  └──────────────┘
                         │
                         ▼
                  ┌──────────────┐
                  │   REPORTS    │ (Read aggregations)
                  └──────────────┘
```

### Detailed Feature Connections

#### Organizations → Everything
- **Foreign Key**: Every table has `org_id`
- **Query Filter**: All queries filtered by org to ensure multi-tenancy
- **Rate Limiting**: Per-org rate limits (100 req/min per org)

#### Users & RBAC → Auth → All Write Operations
- **Login Flow**: User credentials → JWT access token
- **Authorization**: Every write operation checks RBAC permissions
- **Roles**:
  - `admin`: Full access
  - `planner`: Can create programs, aircraft, tasks
  - `technician`: Can execute tasks, reserve parts
  - `viewer`: Read-only access

#### Aircraft → Programs → Tasks
- **Relationship**: Many-to-Many (aircraft can have multiple programs)
- **Trigger**: When aircraft hours/cycles update → check if new tasks needed
- **Auto-Generation**: Program Generator job creates tasks based on intervals

#### Tasks → Part Reservations → Parts Inventory
- **Workflow**:
  1. Task created → planner identifies required parts
  2. Technician reserves parts → inventory decremented
  3. Task completed → reservations marked consumed
- **Inventory Check**: Prevents task execution if parts unavailable

#### All Write Operations → Audit Logs
- **Who**: `user_id`
- **What**: `action` (created, updated, deleted)
- **When**: `created_at` timestamp
- **Where**: `resource_type`, `resource_id`
- **Why**: Compliance and debugging

#### Write Operations → Outbox → Redis → Webhooks
- **Pattern**: Transactional Outbox
- **Flow**:
  1. Business operation + outbox insert in same transaction
  2. Outbox publisher polls → Redis stream
  3. Webhook dispatcher consumes → HTTP POST
- **Guarantee**: At-least-once delivery

#### All Data → Reports
- **Read Models**: Aggregations for dashboards
- **Examples**:
  - Compliance summary (% of tasks on-time)
  - Inventory levels (parts below threshold)
  - Technician workload (tasks per user)

---

## 5. Database Schema Overview

### Core Tables

#### organizations
```sql
CREATE TABLE organizations (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
```

**Relationships:**
- Parent of all other tables via `org_id` foreign key

---

#### users
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL, -- admin, planner, technician, viewer
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(org_id, email)
);

CREATE INDEX idx_users_org_email ON users(org_id, email);
```

**Relationships:**
- Belongs to organization
- Creates audit logs (created_by)
- Assigned to tasks

---

#### aircraft
```sql
CREATE TABLE aircraft (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  registration VARCHAR(50) NOT NULL,
  serial_number VARCHAR(100),
  current_hours DECIMAL(10,2) NOT NULL DEFAULT 0,
  current_cycles INT NOT NULL DEFAULT 0,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(org_id, registration)
);

CREATE INDEX idx_aircraft_org ON aircraft(org_id);
```

**Relationships:**
- Belongs to organization
- Has many programs (via `aircraft_programs`)
- Has many tasks

---

#### programs
```sql
CREATE TABLE programs (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  interval_type VARCHAR(50) NOT NULL, -- hours, cycles, days
  interval_value INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_programs_org ON programs(org_id);
```

**Relationships:**
- Belongs to organization
- Linked to aircraft (via `aircraft_programs`)
- Generates tasks

---

#### aircraft_programs (Join Table)
```sql
CREATE TABLE aircraft_programs (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  aircraft_id UUID NOT NULL REFERENCES aircraft(id),
  program_id UUID NOT NULL REFERENCES programs(id),
  last_completed_hours DECIMAL(10,2),
  last_completed_cycles INT,
  last_completed_date DATE,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(aircraft_id, program_id)
);

CREATE INDEX idx_aircraft_programs_aircraft ON aircraft_programs(aircraft_id);
CREATE INDEX idx_aircraft_programs_program ON aircraft_programs(program_id);
```

---

#### tasks
```sql
CREATE TABLE tasks (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  aircraft_id UUID NOT NULL REFERENCES aircraft(id),
  program_id UUID REFERENCES programs(id),
  status VARCHAR(50) NOT NULL, -- scheduled, in_progress, completed, cancelled
  due_at_hours DECIMAL(10,2),
  due_at_cycles INT,
  due_at_date DATE,
  completed_at TIMESTAMPTZ,
  completed_by UUID REFERENCES users(id),
  completed_at_hours DECIMAL(10,2),
  completed_at_cycles INT,
  assigned_to UUID REFERENCES users(id),
  completion_notes TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_tasks_org_status ON tasks(org_id, status);
CREATE INDEX idx_tasks_aircraft ON tasks(aircraft_id);
CREATE INDEX idx_tasks_assigned ON tasks(assigned_to) WHERE assigned_to IS NOT NULL;
```

**Relationships:**
- Belongs to organization, aircraft, program
- Assigned to user (technician)
- Completed by user
- Has many part reservations

---

#### parts
```sql
CREATE TABLE parts (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  part_number VARCHAR(100) NOT NULL,
  description TEXT,
  unit_price DECIMAL(10,2),
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(org_id, part_number)
);
```

---

#### inventory
```sql
CREATE TABLE inventory (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  part_id UUID NOT NULL REFERENCES parts(id),
  location VARCHAR(255),
  quantity_on_hand INT NOT NULL DEFAULT 0,
  quantity_reserved INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_inventory_part ON inventory(part_id);
```

**Computed Field:**
- `available_quantity = quantity_on_hand - quantity_reserved`

---

#### part_reservations
```sql
CREATE TABLE part_reservations (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  task_id UUID NOT NULL REFERENCES tasks(id),
  inventory_id UUID NOT NULL REFERENCES inventory(id),
  quantity INT NOT NULL,
  status VARCHAR(50) NOT NULL, -- reserved, consumed, released
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_part_reservations_task ON part_reservations(task_id);
```

---

#### outbox (Transactional Outbox Pattern)
```sql
CREATE TABLE outbox (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  event_type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_outbox_unpublished ON outbox(created_at)
  WHERE published_at IS NULL;
```

**Purpose**: Ensures events are published exactly once, even if worker crashes

**Flow**:
1. Business operation + outbox insert in same transaction
2. Worker polls for `published_at IS NULL`
3. Publishes to Redis stream
4. Updates `published_at`

---

#### webhooks
```sql
CREATE TABLE webhooks (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  url VARCHAR(500) NOT NULL,
  events TEXT[] NOT NULL, -- ['task.completed', 'part.consumed']
  secret VARCHAR(255) NOT NULL, -- HMAC signing key
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_webhooks_org_active ON webhooks(org_id, active);
```

---

#### webhook_deliveries
```sql
CREATE TABLE webhook_deliveries (
  id UUID PRIMARY KEY,
  webhook_id UUID NOT NULL REFERENCES webhooks(id),
  event_type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(50) NOT NULL, -- pending, success, failed
  attempts INT NOT NULL DEFAULT 0,
  last_attempt_at TIMESTAMPTZ,
  response_status INT,
  response_body TEXT,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status)
  WHERE status = 'pending';
```

---

#### audit_logs
```sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY,
  org_id UUID NOT NULL REFERENCES organizations(id),
  user_id UUID REFERENCES users(id),
  resource_type VARCHAR(100) NOT NULL,
  resource_id UUID NOT NULL,
  action VARCHAR(50) NOT NULL, -- created, updated, deleted
  changes JSONB, -- before/after values
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_logs_org_resource ON audit_logs(org_id, resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
```

---

### Entity Relationship Diagram

```
organizations
    │
    ├──► users
    │      └──► tasks (assigned_to, completed_by)
    │
    ├──► aircraft
    │      │
    │      ├──► aircraft_programs ◄──── programs
    │      │
    │      └──► tasks ◄──── programs
    │             │
    │             └──► part_reservations ◄──── inventory ◄──── parts
    │
    ├──► outbox ──► Redis Stream ──► webhooks ──► webhook_deliveries
    │
    └──► audit_logs (logs all operations)
```

---

## 6. Development Workflows

### Adding a New REST Endpoint

**Example: Add endpoint to update aircraft maintenance notes**

#### Step 1: Define the route

**File:** `internal/api/rest/routes.go`

```go
func (s *Server) registerRoutes() {
    // ... existing routes ...

    aircraftGroup := s.router.Group("/api/v1/aircraft")
    aircraftGroup.Use(s.authMiddleware, s.rbacMiddleware("aircraft:write"))
    {
        aircraftGroup.PUT("/:id/notes", s.handleUpdateAircraftNotes)
    }
}
```

#### Step 2: Create the handler

**File:** `internal/api/rest/aircraft_handlers.go`

```go
type UpdateAircraftNotesRequest struct {
    Notes string `json:"notes" validate:"required,max=5000"`
}

func (s *Server) handleUpdateAircraftNotes(c *gin.Context) {
    ctx := c.Request.Context()
    aircraftID := c.Param("id")

    // Parse request
    var req UpdateAircraftNotesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    // Validate
    if err := s.validator.Struct(req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Get org from context (set by auth middleware)
    orgID := c.GetString("org_id")
    userID := c.GetString("user_id")

    // Call service
    aircraft, err := s.aircraftService.UpdateNotes(ctx, app.UpdateNotesInput{
        OrgID:      orgID,
        AircraftID: aircraftID,
        Notes:      req.Notes,
        UserID:     userID,
    })
    if err != nil {
        if errors.Is(err, app.ErrNotFound) {
            c.JSON(404, gin.H{"error": "aircraft not found"})
            return
        }
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }

    c.JSON(200, aircraft)
}
```

#### Step 3: Implement service logic

**File:** `internal/app/aircraft_service.go`

```go
type UpdateNotesInput struct {
    OrgID      string
    AircraftID string
    Notes      string
    UserID     string
}

func (s *AircraftService) UpdateNotes(ctx context.Context, input UpdateNotesInput) (*domain.Aircraft, error) {
    // Begin transaction
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Get existing aircraft
    aircraft, err := s.repo.GetByID(ctx, tx, input.OrgID, input.AircraftID)
    if err != nil {
        return nil, err
    }

    // Update notes
    oldNotes := aircraft.Notes
    aircraft.Notes = input.Notes
    aircraft.UpdatedAt = time.Now()

    // Save
    if err := s.repo.Update(ctx, tx, aircraft); err != nil {
        return nil, err
    }

    // Audit log
    changes := map[string]interface{}{
        "before": map[string]string{"notes": oldNotes},
        "after":  map[string]string{"notes": input.Notes},
    }
    if err := s.auditRepo.Log(ctx, tx, domain.AuditLog{
        OrgID:        input.OrgID,
        UserID:       input.UserID,
        ResourceType: "aircraft",
        ResourceID:   aircraft.ID,
        Action:       "updated",
        Changes:      changes,
    }); err != nil {
        return nil, err
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return aircraft, nil
}
```

#### Step 4: Add OpenAPI documentation

**File:** `internal/api/rest/aircraft_handlers.go`

```go
// handleUpdateAircraftNotes godoc
// @Summary      Update aircraft maintenance notes
// @Description  Updates the maintenance notes field for an aircraft
// @Tags         aircraft
// @Accept       json
// @Produce      json
// @Param        id   path      string                        true  "Aircraft ID"
// @Param        body body      UpdateAircraftNotesRequest    true  "Notes"
// @Success      200  {object}  domain.Aircraft
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/aircraft/{id}/notes [put]
func (s *Server) handleUpdateAircraftNotes(c *gin.Context) {
    // ... implementation above ...
}
```

#### Step 5: Write tests

**File:** `internal/api/rest/aircraft_handlers_test.go`

```go
func TestUpdateAircraftNotes(t *testing.T) {
    // Setup
    server := setupTestServer(t)
    defer server.Cleanup()

    // Create test aircraft
    aircraft := createTestAircraft(t, server)

    // Update notes
    req := UpdateAircraftNotesRequest{
        Notes: "Needs engine inspection after 200 hours",
    }

    resp := server.PUT(fmt.Sprintf("/api/v1/aircraft/%s/notes", aircraft.ID), req)
    assert.Equal(t, 200, resp.StatusCode)

    // Verify update
    var updated domain.Aircraft
    json.Unmarshal(resp.Body, &updated)
    assert.Equal(t, req.Notes, updated.Notes)

    // Verify audit log created
    logs := server.GetAuditLogs(aircraft.ID)
    assert.Len(t, logs, 1)
    assert.Equal(t, "updated", logs[0].Action)
}
```

#### Step 6: Run and verify

```bash
# Run tests
go test ./internal/api/rest -run TestUpdateAircraftNotes

# Start server
go run ./cmd/server

# Test manually
curl -X PUT http://localhost:8080/api/v1/aircraft/<uuid>/notes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"notes":"Engine service due"}'

# Check OpenAPI docs
open http://localhost:8080/docs
```

---

### Adding a New Background Job

**Example: Add job to send email reminders for overdue tasks**

#### Step 1: Create the job

**File:** `internal/jobs/task_reminder_job.go`

```go
package jobs

import (
    "context"
    "time"
)

type TaskReminderJob struct {
    db        *sql.DB
    emailSvc  EmailService
    logger    *slog.Logger
}

func NewTaskReminderJob(db *sql.DB, emailSvc EmailService, logger *slog.Logger) *TaskReminderJob {
    return &TaskReminderJob{
        db:       db,
        emailSvc: emailSvc,
        logger:   logger,
    }
}

func (j *TaskReminderJob) Run(ctx context.Context) error {
    j.logger.Info("task_reminder_job_started")

    // Query overdue tasks
    tasks, err := j.getOverdueTasks(ctx)
    if err != nil {
        return err
    }

    // Send email for each task
    for _, task := range tasks {
        if err := j.sendReminder(ctx, task); err != nil {
            j.logger.Error("failed to send reminder", "task_id", task.ID, "error", err)
            continue
        }

        // Mark reminder sent
        if err := j.markReminderSent(ctx, task.ID); err != nil {
            j.logger.Error("failed to mark reminder sent", "task_id", task.ID, "error", err)
        }
    }

    j.logger.Info("task_reminder_job_completed", "reminders_sent", len(tasks))
    return nil
}

func (j *TaskReminderJob) getOverdueTasks(ctx context.Context) ([]domain.Task, error) {
    // Query tasks where:
    // - status = 'scheduled'
    // - due_at < now
    // - reminder_sent_at IS NULL
    query := `
        SELECT id, aircraft_id, assigned_to, due_at_hours, due_at_date
        FROM tasks
        WHERE status = 'scheduled'
          AND (due_at_date < CURRENT_DATE OR due_at_hours < (
              SELECT current_hours FROM aircraft WHERE id = tasks.aircraft_id
          ))
          AND reminder_sent_at IS NULL
    `
    // ... execute query and scan results
}

func (j *TaskReminderJob) sendReminder(ctx context.Context, task domain.Task) error {
    // Send email via email service
    return j.emailSvc.Send(ctx, EmailMessage{
        To:      task.AssignedTo.Email,
        Subject: "Maintenance Task Overdue",
        Body:    fmt.Sprintf("Task %s is overdue...", task.ID),
    })
}

func (j *TaskReminderJob) markReminderSent(ctx context.Context, taskID string) error {
    _, err := j.db.ExecContext(ctx,
        "UPDATE tasks SET reminder_sent_at = NOW() WHERE id = $1",
        taskID,
    )
    return err
}
```

#### Step 2: Register job in worker

**File:** `cmd/worker/main.go`

```go
func main() {
    // ... existing setup ...

    // Create task reminder job
    taskReminderJob := jobs.NewTaskReminderJob(db, emailService, logger)

    // Schedule to run every hour
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                if err := taskReminderJob.Run(context.Background()); err != nil {
                    logger.Error("task_reminder_job_failed", "error", err)
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    // ... rest of worker setup ...
}
```

#### Step 3: Add database migration

**File:** `migrations/20250124_add_reminder_sent_at.sql`

```sql
-- +goose Up
ALTER TABLE tasks ADD COLUMN reminder_sent_at TIMESTAMPTZ;

CREATE INDEX idx_tasks_reminder_pending
    ON tasks(due_at_date)
    WHERE status = 'scheduled' AND reminder_sent_at IS NULL;

-- +goose Down
DROP INDEX idx_tasks_reminder_pending;
ALTER TABLE tasks DROP COLUMN reminder_sent_at;
```

#### Step 4: Test the job

**File:** `internal/jobs/task_reminder_job_test.go`

```go
func TestTaskReminderJob(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    emailSvc := &mockEmailService{}
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    job := NewTaskReminderJob(db, emailSvc, logger)

    // Create overdue task
    task := createOverdueTask(t, db)

    // Run job
    err := job.Run(context.Background())
    assert.NoError(t, err)

    // Verify email sent
    assert.Len(t, emailSvc.SentEmails, 1)
    assert.Equal(t, task.AssignedTo.Email, emailSvc.SentEmails[0].To)

    // Verify reminder marked sent
    var reminderSentAt time.Time
    db.QueryRow("SELECT reminder_sent_at FROM tasks WHERE id = $1", task.ID).Scan(&reminderSentAt)
    assert.False(t, reminderSentAt.IsZero())
}
```

---

### Adding a Database Migration

**Example: Add support for aircraft photos**

#### Step 1: Create migration file

```bash
# Generate migration
goose -dir migrations create add_aircraft_photos sql
```

**File:** `migrations/20250124120000_add_aircraft_photos.sql`

```sql
-- +goose Up
CREATE TABLE aircraft_photos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  aircraft_id UUID NOT NULL REFERENCES aircraft(id) ON DELETE CASCADE,
  url VARCHAR(500) NOT NULL,
  caption TEXT,
  uploaded_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT fk_aircraft_photos_org
    FOREIGN KEY (org_id) REFERENCES organizations(id),
  CONSTRAINT fk_aircraft_photos_aircraft
    FOREIGN KEY (aircraft_id) REFERENCES aircraft(id)
);

CREATE INDEX idx_aircraft_photos_aircraft ON aircraft_photos(aircraft_id);
CREATE INDEX idx_aircraft_photos_org ON aircraft_photos(org_id);

-- +goose Down
DROP TABLE aircraft_photos;
```

#### Step 2: Run migration

```bash
# Apply migration
goose -dir migrations postgres "postgres://amss:amss@localhost:5455/amss?sslmode=disable" up

# Verify
psql postgres://amss:amss@localhost:5455/amss -c "\d aircraft_photos"
```

#### Step 3: Add to deployment

**File:** `deploy/helm/amss/templates/migration-job.yaml` (if using Helm hook)

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "amss.fullname" . }}-migrations-{{ .Release.Revision }}
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-weight": "-1"
    "helm.sh/hook-delete-policy": before-hook-creation
spec:
  template:
    spec:
      containers:
      - name: migrations
        image: "{{ .Values.image.server.repository }}:{{ .Values.image.server.tag }}"
        command: ["goose", "-dir", "migrations", "postgres", "$(DB_URL)", "up"]
        env:
        - name: DB_URL
          valueFrom:
            secretKeyRef:
              name: {{ include "amss.fullname" . }}-secret
              key: DB_URL
      restartPolicy: Never
```

---

## 7. Code Organization & Patterns

### Directory Structure

```
amss-backend/
├── cmd/                          # Entrypoints
│   ├── server/
│   │   └── main.go              # HTTP/gRPC server
│   └── worker/
│       └── main.go              # Background jobs
│
├── internal/                     # Private application code
│   ├── api/                     # API layer (thin)
│   │   ├── rest/                # REST handlers
│   │   │   ├── server.go
│   │   │   ├── routes.go
│   │   │   ├── middleware.go
│   │   │   ├── auth_handlers.go
│   │   │   ├── aircraft_handlers.go
│   │   │   └── ...
│   │   └── grpc/                # gRPC handlers
│   │       └── server.go
│   │
│   ├── app/                     # Application services (thick)
│   │   ├── ports.go             # Interface definitions
│   │   ├── auth_service.go      # Business logic
│   │   ├── aircraft_service.go
│   │   ├── task_service.go
│   │   └── ...
│   │
│   ├── domain/                  # Domain models
│   │   ├── organization.go
│   │   ├── user.go
│   │   ├── aircraft.go
│   │   ├── task.go
│   │   └── ...
│   │
│   ├── infra/                   # Infrastructure implementations
│   │   ├── postgres/            # Database repositories
│   │   │   ├── aircraft_repo.go
│   │   │   ├── task_repo.go
│   │   │   └── ...
│   │   └── redis/               # Redis implementations
│   │       ├── rate_limiter.go
│   │       └── stream.go
│   │
│   └── jobs/                    # Background workers
│       ├── outbox_publisher.go
│       ├── webhook_dispatcher.go
│       ├── import_processor.go
│       └── program_generator.go
│
├── api/                         # Public API definitions
│   └── proto/
│       └── amss.proto           # gRPC service definition
│
├── migrations/                  # Database migrations
│   ├── 001_initial_schema.sql
│   └── ...
│
├── deploy/                      # Deployment configs
│   └── helm/
│       └── amss/
│
├── docs/                        # Documentation
│   ├── architecture-diagrams.md
│   ├── DEVELOPER_GUIDE.md
│   └── API_GUIDE.md
│
└── scripts/                     # Utility scripts
    └── smoke_test.ps1
```

---

### Layered Architecture Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                         API Layer                            │
│  (HTTP/gRPC handlers - thin, no business logic)             │
│  - Parse requests                                            │
│  - Call service layer                                        │
│  - Format responses                                          │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│  (Business logic - thick, orchestrates operations)          │
│  - Validate business rules                                   │
│  - Coordinate multiple repos                                 │
│  - Manage transactions                                       │
│  - Emit events                                               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Repository Layer                         │
│  (Data access - thin, CRUD operations only)                 │
│  - SQL queries                                               │
│  - No business logic                                         │
│  - Return domain models                                      │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                        Database                              │
│  (Postgres - system of record)                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Principle**: **Dependency flows inward** (API → Service → Repository → Database)

**Example Flow:**
1. **Handler** receives HTTP request
2. **Handler** validates request format (JSON structure)
3. **Handler** calls **Service** method
4. **Service** validates business rules (e.g., "can't complete task without parts")
5. **Service** calls **Repository** methods (get, update, insert)
6. **Repository** executes SQL queries
7. **Service** commits transaction
8. **Handler** formats response

---

### Dependency Injection Pattern

**Why**: Testability, flexibility, decoupling

**Example:** Aircraft Service

```go
// Define interfaces (ports)
type AircraftRepository interface {
    GetByID(ctx context.Context, tx *sql.Tx, orgID, aircraftID string) (*domain.Aircraft, error)
    Update(ctx context.Context, tx *sql.Tx, aircraft *domain.Aircraft) error
}

type AuditRepository interface {
    Log(ctx context.Context, tx *sql.Tx, log domain.AuditLog) error
}

// Service depends on interfaces, not concrete implementations
type AircraftService struct {
    db        *sql.DB
    repo      AircraftRepository
    auditRepo AuditRepository
    logger    *slog.Logger
}

// Constructor injects dependencies
func NewAircraftService(
    db *sql.DB,
    repo AircraftRepository,
    auditRepo AuditRepository,
    logger *slog.Logger,
) *AircraftService {
    return &AircraftService{
        db:        db,
        repo:      repo,
        auditRepo: auditRepo,
        logger:    logger,
    }
}
```

**Benefits:**
- **Testing**: Inject mock repos for unit tests
- **Flexibility**: Swap Postgres for MySQL without changing service code
- **Decoupling**: Service doesn't know about SQL

---

### Error Handling Patterns

#### Sentinel Errors (Domain Errors)

**File:** `internal/app/errors.go`

```go
var (
    ErrNotFound          = errors.New("resource not found")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrConflict          = errors.New("resource conflict")
    ErrValidation        = errors.New("validation failed")
    ErrInsufficientParts = errors.New("insufficient parts in inventory")
)
```

**Usage in Service:**

```go
func (s *TaskService) Complete(ctx context.Context, input CompleteTaskInput) error {
    // Check if task exists
    task, err := s.repo.GetByID(ctx, tx, input.TaskID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return ErrNotFound // Return domain error
        }
        return fmt.Errorf("failed to get task: %w", err) // Wrap infrastructure error
    }

    // Check parts availability
    if !s.hasRequiredParts(ctx, task) {
        return ErrInsufficientParts // Return domain error
    }

    // ... rest of logic
}
```

**Usage in Handler:**

```go
func (s *Server) handleCompleteTask(c *gin.Context) {
    err := s.taskService.Complete(ctx, input)

    if err != nil {
        // Map domain errors to HTTP status codes
        switch {
        case errors.Is(err, app.ErrNotFound):
            c.JSON(404, gin.H{"error": "task not found"})
        case errors.Is(err, app.ErrInsufficientParts):
            c.JSON(409, gin.H{"error": "insufficient parts in inventory"})
        default:
            c.JSON(500, gin.H{"error": "internal server error"})
            s.logger.Error("complete_task_failed", "error", err)
        }
        return
    }

    c.JSON(200, gin.H{"status": "completed"})
}
```

---

## 8. Common Development Tasks

### Running the System Locally

#### Option 1: Docker Compose (Recommended for quick start)

```bash
# Start Postgres + Redis
docker-compose up -d

# Run migrations
make migrate-up

# Seed demo data
make seed

# Start server (Terminal 1)
go run ./cmd/server

# Start worker (Terminal 2)
go run ./cmd/worker

# Access API
curl http://localhost:8080/health  # → "ok"
curl http://localhost:8080/docs    # → Swagger UI
```

#### Option 2: Local Go (Requires Postgres/Redis already running)

```powershell
# Set environment variables
$env:DB_URL = "postgres://amss:amss@localhost:5455/amss?sslmode=disable"
$env:REDIS_ADDR = "localhost:6379"
$env:JWT_PRIVATE_KEY_PEM = $(cat jwt-private.pem)
$env:JWT_PUBLIC_KEY_PEM = $(cat jwt-public.pem)
$env:APP_ENV = "development"

# Run server
go run ./cmd/server

# In another terminal, run worker
go run ./cmd/worker
```

---

### Debugging with VS Code

**File:** `.vscode/launch.json`

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "env": {
        "DB_URL": "postgres://amss:amss@localhost:5455/amss?sslmode=disable",
        "REDIS_ADDR": "localhost:6379",
        "JWT_PRIVATE_KEY_PEM": "<paste key here>",
        "JWT_PUBLIC_KEY_PEM": "<paste key here>",
        "APP_ENV": "development",
        "LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Launch Worker",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/worker",
      "env": {
        "DB_URL": "postgres://amss:amss@localhost:5455/amss?sslmode=disable",
        "REDIS_ADDR": "localhost:6379",
        "APP_ENV": "development"
      }
    }
  ]
}
```

**Set Breakpoints:**
1. Open file (e.g., `internal/app/task_service.go`)
2. Click left margin to add breakpoint
3. Press F5 to start debugging
4. Make API call to trigger breakpoint

---

### Testing Strategies

#### Unit Tests (Fast, isolated)

**Test Services with Mock Repositories:**

```go
func TestTaskService_Complete(t *testing.T) {
    // Setup mocks
    mockRepo := &mockTaskRepository{}
    mockAuditRepo := &mockAuditRepository{}
    logger := slog.Default()

    service := app.NewTaskService(nil, mockRepo, mockAuditRepo, logger)

    // Configure mock behavior
    mockRepo.On("GetByID", mock.Anything, "task-123").Return(&domain.Task{
        ID:     "task-123",
        Status: "in_progress",
    }, nil)

    mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
    mockAuditRepo.On("Log", mock.Anything, mock.Anything).Return(nil)

    // Execute
    err := service.Complete(context.Background(), app.CompleteTaskInput{
        TaskID: "task-123",
    })

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

#### Integration Tests (Medium speed, real database)

**Test Repositories with Test Containers:**

```go
func TestPostgresTaskRepository(t *testing.T) {
    if os.Getenv("AMSS_INTEGRATION") != "1" {
        t.Skip("Skipping integration test")
    }

    // Start Postgres container
    ctx := context.Background()
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:16",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_USER":     "test",
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "test",
            },
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Get connection string
    host, _ := postgres.Host(ctx)
    port, _ := postgres.MappedPort(ctx, "5432")
    dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())

    // Connect and run migrations
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    runMigrations(t, db)

    // Test repository
    repo := postgres.NewTaskRepository(db)

    task := &domain.Task{
        ID:     uuid.NewString(),
        OrgID:  "org-123",
        Status: "scheduled",
    }

    err = repo.Insert(ctx, nil, task)
    assert.NoError(t, err)

    fetched, err := repo.GetByID(ctx, nil, "org-123", task.ID)
    assert.NoError(t, err)
    assert.Equal(t, task.ID, fetched.ID)
}
```

#### End-to-End Tests (Slow, full system)

**Smoke Test Script:** `scripts/smoke_test.ps1`

```powershell
# Start services
docker-compose up -d

# Wait for readiness
while ((curl http://localhost:8080/ready).StatusCode -ne 200) {
    Start-Sleep 1
}

# Test login
$loginResp = Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/v1/auth/login" -Body (@{
    org_id = "demo-org"
    email = "admin@demo.local"
    password = "ChangeMe123!"
} | ConvertTo-Json) -ContentType "application/json"

$token = $loginResp.access_token

# Test create aircraft
$aircraft = Invoke-RestMethod -Method POST -Uri "http://localhost:8080/api/v1/aircraft" `
    -Headers @{ Authorization = "Bearer $token" } `
    -Body (@{
        registration = "N12345"
        serial_number = "SN-9876"
    } | ConvertTo-Json) -ContentType "application/json"

Write-Host "✅ Smoke test passed!"
```

---

### Troubleshooting Common Errors

#### "connection refused" on Postgres/Redis

**Symptom:**
```
panic: dial tcp [::1]:5455: connect: connection refused
```

**Solution:**
```bash
# Check if containers are running
docker ps

# If not, start them
docker-compose up -d

# Verify ports
sudo ss -tlnp | grep -E "5455|6379"
```

---

#### "missing required env vars: DB_URL, JWT_PRIVATE_KEY_PEM"

**Symptom:**
```
panic: missing required env vars: DB_URL, JWT_PRIVATE_KEY_PEM, JWT_PUBLIC_KEY_PEM
```

**Solution:**
```bash
# Generate JWT keys if missing
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem

# Set environment variables
export DB_URL="postgres://amss:amss@localhost:5455/amss?sslmode=disable"
export JWT_PRIVATE_KEY_PEM=$(cat jwt-private.pem)
export JWT_PUBLIC_KEY_PEM=$(cat jwt-public.pem)
```

---

#### "401 Unauthorized" on API requests

**Symptom:**
```json
{"error": "unauthorized"}
```

**Possible Causes:**
1. **Missing token**: Add `Authorization: Bearer <token>` header
2. **Expired token**: Access tokens expire after 15 minutes, use refresh token
3. **Invalid signature**: JWT keys don't match between login and validation

**Solution:**
```bash
# Get fresh token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"org_id":"<uuid>","email":"admin@demo.local","password":"ChangeMe123!"}'

# Use token in subsequent requests
curl http://localhost:8080/api/v1/aircraft \
  -H "Authorization: Bearer <access_token>"
```

---

#### "rate limit exceeded"

**Symptom:**
```json
{"error": "rate limit exceeded, try again later"}
```

**Cause:** Organization exceeded 100 requests per minute

**Solution:**
1. **Development**: Increase rate limit in code or disable middleware
2. **Production**: Wait 1 minute or contact admin to adjust org-specific limit

---

#### Webhook delivery failures

**Symptom:** Webhooks not being received by external system

**Debugging Steps:**

```bash
# 1. Check if outbox events are being created
psql $DB_URL -c "SELECT * FROM outbox WHERE published_at IS NULL LIMIT 5;"

# 2. Check if worker is running
docker logs amss-worker --tail=50

# 3. Check webhook delivery status
psql $DB_URL -c "SELECT * FROM webhook_deliveries WHERE status = 'failed' LIMIT 5;"

# 4. Check webhook configuration
psql $DB_URL -c "SELECT * FROM webhooks WHERE active = true;"

# 5. Test webhook URL manually
curl -X POST https://your-webhook-url.com/endpoint \
  -H "Content-Type: application/json" \
  -d '{"test":"event"}'
```

---

## 9. Key Technical Decisions

### Why Postgres over NoSQL?

**Decision:** Use PostgreSQL as the primary database

**Reasoning:**
1. **ACID Transactions**: Maintenance operations require atomic multi-table updates (task + parts + audit log)
2. **Relational Data**: Aircraft → Programs → Tasks → Parts are inherently relational
3. **Audit Compliance**: Regulatory requirements demand transactional integrity
4. **Complex Queries**: Reports need joins, aggregations, and filtering across multiple tables
5. **Proven Reliability**: 30+ years of production use in critical systems

**Trade-offs:**
- ❌ Slower for write-heavy workloads (vs NoSQL)
- ❌ Vertical scaling required (vs horizontal NoSQL)
- ✅ Strong consistency guarantees
- ✅ Rich query capabilities
- ✅ Mature ecosystem

---

### Why Redis for Rate Limiting?

**Decision:** Use Redis for rate limiting and job queues

**Reasoning:**
1. **Speed**: In-memory operations (<1ms latency)
2. **Atomic Operations**: INCR command is atomic (no race conditions)
3. **TTL Support**: Keys auto-expire (no manual cleanup)
4. **Redis Streams**: Efficient pub/sub for webhook queue

**Why Not Postgres for Rate Limiting?**
- Rate limiting requires sub-millisecond response time
- Postgres writes are slower (disk I/O, WAL logging)
- Would add unnecessary load to primary database

**Trade-offs:**
- ❌ Data can be lost on Redis crash (but rate limits can be recalculated)
- ✅ Extremely fast
- ✅ Low resource usage

---

### Why Outbox Pattern for Webhooks?

**Problem:** How to ensure webhook delivery even if worker crashes?

**Naive Approach (Don't Use):**
```go
func CreateTask(task Task) error {
    // Insert task into database
    db.Insert(task)

    // Send webhook directly
    http.Post(webhookURL, taskJSON) // ❌ What if this fails? No retry!

    return nil
}
```

**Issues:**
- If webhook fails, event is lost forever
- If server crashes after DB insert but before webhook send, event is lost
- No retry mechanism
- No audit trail

**Outbox Pattern (Recommended):**

```go
func CreateTask(task Task) error {
    tx, _ := db.Begin()

    // Insert task
    tx.Insert(task)

    // Insert outbox event (same transaction!)
    tx.Insert(OutboxEvent{
        EventType: "task.created",
        Payload:   taskJSON,
    })

    tx.Commit() // Both inserts succeed or both fail

    // Worker polls outbox asynchronously
    return nil
}
```

**Benefits:**
- ✅ **At-least-once delivery**: Event persisted in database
- ✅ **Crash recovery**: Worker polls outbox on restart
- ✅ **Retry logic**: Worker handles retries with exponential backoff
- ✅ **Audit trail**: All deliveries logged in `webhook_deliveries`

**Implementation Flow:**
1. API server: Insert business data + outbox event in single transaction
2. Outbox publisher (worker): Polls `outbox` table every 100ms
3. Outbox publisher: Publishes to Redis stream
4. Webhook dispatcher (worker): Consumes Redis stream, sends HTTP POST
5. Webhook dispatcher: Retries on failure (1min, 5min, 15min, ...)

---

### Why Separate Server and Worker?

**Decision:** Split into `amss-server` (API) and `amss-worker` (jobs)

**Reasoning:**

**Failure Isolation:**
- If webhook dispatcher crashes, API stays up
- If CSV import consumes all CPU, API remains responsive

**Independent Scaling:**
- Run 5 server replicas (high request volume)
- Run 2 worker replicas (moderate background job volume)

**Resource Optimization:**
- Server: High memory, low CPU (caching, connections)
- Worker: High CPU, low memory (data processing)

**Deployment Flexibility:**
- Deploy server updates without stopping background jobs
- Deploy worker updates without API downtime

**Alternative (Monolith - Not Chosen):**
```
Single process runs API + background jobs
❌ Webhook retry loop blocks API threads
❌ CSV import spikes CPU, slows API response
❌ Can't scale independently
```

---

### Why Idempotency Keys?

**Problem:** Network retries can cause duplicate operations

**Example Scenario:**
```
1. Client sends: POST /tasks (create task)
2. Server creates task in DB
3. Server tries to send response
4. Network fails (timeout)
5. Client retries: POST /tasks (same request)
6. ❌ Server creates duplicate task!
```

**Solution: Idempotency Keys**

```bash
curl -X POST /api/v1/tasks \
  -H "Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{"aircraft_id":"abc","program_id":"xyz"}'

# If retried with same key:
# → Server returns cached result (201 + task JSON)
# → No duplicate task created
```

**Implementation:**

```go
func (s *Server) handleCreateTask(c *gin.Context) {
    idempotencyKey := c.GetHeader("Idempotency-Key")

    if idempotencyKey != "" {
        // Check if already processed
        cached, err := s.redis.Get(ctx, "idempotency:"+idempotencyKey).Result()
        if err == nil {
            // Return cached response
            c.Data(200, "application/json", []byte(cached))
            return
        }
    }

    // Process request
    task, err := s.taskService.Create(ctx, input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Cache result for 24 hours
    if idempotencyKey != "" {
        responseJSON, _ := json.Marshal(task)
        s.redis.Set(ctx, "idempotency:"+idempotencyKey, responseJSON, 24*time.Hour)
    }

    c.JSON(201, task)
}
```

**Benefits:**
- ✅ Safe retries (network failures, timeouts)
- ✅ Prevents duplicate charges, double bookings, etc.
- ✅ Industry standard (Stripe, Shopify, AWS use this)

---

### Why JWT over Session Cookies?

**Decision:** Use JWT (JSON Web Tokens) for authentication

**Reasoning:**

**Stateless:**
- Server doesn't need to store session data
- No Redis/DB lookup on every request
- Scales horizontally (any server can validate token)

**Mobile-Friendly:**
- Easy to store in mobile apps (not tied to cookies)
- Works with native apps, SPAs, and server-rendered apps

**Standard:**
- Industry standard (OAuth 2.0, OpenID Connect)
- Libraries available in every language

**Implementation:**
```
Login:
  email + password → JWT access token (15min) + refresh token (7 days)

Subsequent Requests:
  Authorization: Bearer <access_token>

Token Validation:
  1. Verify signature (RSA public key)
  2. Check expiration (exp claim)
  3. Extract user/org (custom claims)
  4. No database lookup needed!
```

**Trade-offs:**
- ❌ Can't revoke tokens before expiration (mitigated by short TTL)
- ❌ Tokens can be large (200-500 bytes)
- ✅ Extremely fast validation (no DB)
- ✅ Stateless scaling

**Alternative (Session Cookies - Not Chosen):**
```
Login → Session ID → Store in Redis → Set cookie
Every Request → Read cookie → Lookup session in Redis
❌ Redis becomes single point of failure
❌ Slower (network round-trip to Redis)
❌ Doesn't work well with mobile apps
```

---

## 10. Contribution Guidelines

### Code Style

**Go Conventions:**
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (automatic in most IDEs)
- Use `golangci-lint` for linting

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run
```

**Naming:**
- **Variables**: `camelCase` (e.g., `aircraftID`, `taskService`)
- **Functions**: `PascalCase` for exported (e.g., `CreateTask`), `camelCase` for private
- **Constants**: `PascalCase` (e.g., `MaxRetries`)
- **Acronyms**: Keep uppercase (e.g., `ID`, `URL`, `HTTP`)

**Example:**
```go
// Good
func (s *TaskService) GetByID(ctx context.Context, taskID string) (*domain.Task, error)

// Bad
func (s *TaskService) GetById(ctx context.Context, task_id string) (*domain.Task, error)
```

---

### PR Process

**1. Create Feature Branch**
```bash
git checkout -b feature/add-aircraft-photos
```

**2. Make Changes**
- Write code
- Add tests
- Update documentation

**3. Run Tests Locally**
```bash
# Unit tests
go test ./...

# Integration tests
AMSS_INTEGRATION=1 go test ./internal/infra/postgres

# Linting
golangci-lint run
```

**4. Commit with Conventional Commits**
```bash
git commit -m "feat(aircraft): add photo upload support"
git commit -m "fix(auth): validate token expiration correctly"
git commit -m "docs: update deployment guide with new env vars"
```

**Format:**
```
<type>(<scope>): <description>

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- test: Adding tests
- refactor: Code refactoring
- perf: Performance improvement
- chore: Maintenance (dependency updates, etc.)
```

**5. Push and Create PR**
```bash
git push origin feature/add-aircraft-photos
```

**6. PR Description Template**
```markdown
## Summary
Brief description of what this PR does.

## Changes
- Added `aircraft_photos` table
- Created upload endpoint `POST /api/v1/aircraft/:id/photos`
- Added S3 integration for photo storage

## Testing
- [ ] Unit tests added
- [ ] Integration tests added
- [ ] Manually tested with Postman

## Screenshots (if UI changes)
![Screenshot](url)

## Related Issues
Closes #123
```

---

### Code Review Checklist

**For Reviewers:**

**Functionality:**
- [ ] Does the code solve the problem?
- [ ] Are edge cases handled?
- [ ] Is error handling correct?

**Tests:**
- [ ] Unit tests cover new code?
- [ ] Integration tests if touching database?
- [ ] Tests actually test the right thing?

**Security:**
- [ ] No SQL injection vulnerabilities?
- [ ] User input validated?
- [ ] Authorization checks present?
- [ ] Secrets not committed?

**Performance:**
- [ ] Database queries optimized (indexes, LIMIT)?
- [ ] N+1 query problems avoided?
- [ ] Caching used where appropriate?

**Code Quality:**
- [ ] Code is readable?
- [ ] Functions are small and focused?
- [ ] No duplicated code?
- [ ] Proper error messages?

**Documentation:**
- [ ] OpenAPI spec updated?
- [ ] README updated if needed?
- [ ] Code comments explain "why", not "what"?

---

### Documentation Requirements

**When Adding New API Endpoint:**
1. Add OpenAPI comments (godoc format)
2. Update `docs/API_GUIDE.md` with usage example
3. Add to Postman collection (if exists)

**When Adding New Background Job:**
1. Document job purpose in code comment
2. Document schedule/frequency
3. Add monitoring/alerting if critical

**When Adding New Database Migration:**
1. Add `-- +goose Up` and `-- +goose Down` sections
2. Test rollback locally
3. Document any manual steps required

**When Adding New Environment Variable:**
1. Add to `README.md` environment variables section
2. Add to `.env.example`
3. Add to Helm chart `values.yaml`
4. Add to CI/CD secrets documentation

---

## Conclusion

You now have a complete understanding of the AMSS system:

✅ **Problem Domain**: Aircraft maintenance compliance, audit trails, parts coordination
✅ **Architecture**: Server/worker separation, Postgres/Redis, outbox pattern
✅ **User Journeys**: Admin onboarding, maintenance scheduling, task execution, webhooks
✅ **Feature Interconnections**: How organizations → users → aircraft → programs → tasks → parts → webhooks all connect
✅ **Database Schema**: Tables, relationships, indexes
✅ **Development Workflows**: Adding endpoints, jobs, migrations
✅ **Code Organization**: Layered architecture, dependency injection
✅ **Common Tasks**: Local setup, debugging, testing, troubleshooting
✅ **Technical Decisions**: Why Postgres, Redis, outbox pattern, JWT, idempotency
✅ **Contribution Process**: Code style, PR process, review checklist

**Next Steps:**
1. Set up local development environment
2. Run the smoke test to verify everything works
3. Pick a small feature to implement (e.g., add a new field to aircraft)
4. Read existing code in `internal/app/` to see patterns in action
5. Ask questions in team chat or create GitHub discussions

**Welcome to the team!** 🚀
