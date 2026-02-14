# AMSS Strategic Improvement Plan

> Addressing Critical Aviation Industry Problems Through Platform Enhancement

**Date**: February 2026
**System**: Aircraft Maintenance Scheduling System (AMSS)
**Current Stack**: Go backend, React/TypeScript frontend, PostgreSQL, Redis, WebSocket, K3s

---

## Table of Contents

1. [Skill & Certification Matching on Task Assignment](#1-skill--certification-matching-on-task-assignment)
2. [Dynamic Rescheduling When Parts Are Unavailable](#2-dynamic-rescheduling-when-parts-are-unavailable)
3. [Multi-Regulatory Compliance Templates (FAA/EASA/ECAA)](#3-multi-regulatory-compliance-templates-faaeasaecaa)
4. [Real-Time Maintenance Dashboard via WebSocket](#4-real-time-maintenance-dashboard-via-websocket)

---

## 1. Skill & Certification Matching on Task Assignment

### Industry Context

The aviation maintenance industry faces a critical workforce shortage (projected 626,000 new technicians needed globally by 2042 per Boeing). Simultaneously, regulatory frameworks mandate strict qualification verification:

- **FAA 14 CFR Part 145.151**: Requires repair stations to maintain personnel rosters with certificate type, ratings, and employment history, updated within 5 business days of any change
- **EASA Part 145.A.35**: Certifying staff must hold Part-66 AML with applicable type ratings and demonstrate 6 months of actual relevant maintenance experience in any consecutive 2-year period (recency rule)
- **ICAO Annex 6**: Requires maintenance release signatures from personnel licensed per Annex 1

Leading MRO systems (AMOS, Ramco, IFS Maintenix) all enforce skill-to-task matching as a core feature, with Maintenix implementing labor rows per task with skill code requirements and inspector/performer segregation.

### Current State in AMSS

- `users` table has `role` (admin, tenant_admin, scheduler, mechanic, auditor) but **no certification or skill tracking**
- `maintenance_tasks` has `assigned_mechanic_id` but **no validation that the mechanic is qualified**
- No type rating, certification expiry, or recency tracking exists
- Compliance items exist per-task but are generic descriptions, not tied to regulatory frameworks

### Proposed Changes

#### 1.1 New Database Schema (Migration 00008)

```sql
-- Certification types recognized by the system
CREATE TYPE certification_authority AS ENUM ('faa', 'easa', 'ecaa', 'icao', 'other');

CREATE TABLE certification_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,              -- e.g., 'FAA_AP', 'EASA_B1', 'EASA_B2', 'FAA_IA'
  name text NOT NULL,                     -- e.g., 'Airframe & Powerplant'
  authority certification_authority NOT NULL,
  has_expiry boolean NOT NULL DEFAULT false,
  recency_required_months int,            -- e.g., 6 for EASA
  recency_period_months int,              -- e.g., 24 for EASA (6-in-24 rule)
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Aircraft type definitions (for type ratings)
CREATE TABLE aircraft_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  icao_code text NOT NULL UNIQUE,         -- e.g., 'B738', 'A320', 'B77W'
  manufacturer text NOT NULL,
  model text NOT NULL,
  series text,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Mechanic certifications
CREATE TABLE employee_certifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  cert_type_id uuid NOT NULL REFERENCES certification_types(id),
  certificate_number text NOT NULL,
  issue_date date NOT NULL,
  expiry_date date,                       -- NULL for non-expiring (e.g., FAA A&P)
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'expired', 'suspended', 'revoked')),
  verified_by uuid,
  verified_at timestamptz,
  document_url text,                      -- Link to scanned certificate
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  UNIQUE (org_id, user_id, cert_type_id)
);

-- Employee type ratings (per certification)
CREATE TABLE employee_type_ratings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  aircraft_type_id uuid NOT NULL REFERENCES aircraft_types(id),
  cert_id uuid NOT NULL REFERENCES employee_certifications(id),
  endorsement_date date NOT NULL,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'lapsed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id)
);

-- Special skills (NDT, welding, composites, etc.)
CREATE TYPE skill_category AS ENUM (
  'structural', 'avionics', 'engine', 'ndt', 'general'
);

CREATE TABLE skill_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,              -- e.g., 'NDT_UT', 'WELDING_TIG'
  name text NOT NULL,
  category skill_category NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE employee_skills (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  skill_type_id uuid NOT NULL REFERENCES skill_types(id),
  proficiency_level int NOT NULL DEFAULT 1 CHECK (proficiency_level BETWEEN 1 AND 5),
  qualification_date date NOT NULL,
  expiry_date date,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  UNIQUE (org_id, user_id, skill_type_id)
);

-- Recency tracking (hours logged per type per mechanic)
CREATE TABLE employee_recency_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  aircraft_type_id uuid NOT NULL REFERENCES aircraft_types(id),
  task_id uuid NOT NULL,
  work_date date NOT NULL,
  hours_logged numeric(6,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id)
);

-- Task skill requirements
CREATE TABLE task_skill_requirements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_type maintenance_task_type NOT NULL,
  aircraft_type_id uuid REFERENCES aircraft_types(id),  -- NULL = any aircraft
  cert_type_id uuid REFERENCES certification_types(id),
  skill_type_id uuid REFERENCES skill_types(id),
  min_proficiency_level int DEFAULT 1,
  is_certifying_role boolean NOT NULL DEFAULT false,
  is_inspection_role boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Link aircraft table to aircraft_types
ALTER TABLE aircraft ADD COLUMN aircraft_type_id uuid REFERENCES aircraft_types(id);
```

#### 1.2 Backend Changes

| File | Change |
|------|--------|
| `internal/domain/certification.go` | New domain types: `CertificationType`, `EmployeeCertification`, `EmployeeTypeRating`, `SkillType`, `EmployeeSkill`, `TaskSkillRequirement` |
| `internal/domain/task.go` | Add qualification check to `CanTransition()` for `in_progress` state |
| `internal/app/services/task_service.go` | Before assigning mechanic: query certifications, type ratings, recency; reject if unqualified |
| `internal/app/services/certification_service.go` | New service: CRUD for certifications, type ratings, skills; expiry checking; recency computation |
| `internal/infra/postgres/certification_repo.go` | New repository for certification queries |
| `internal/api/rest/handlers/certifications.go` | New REST handlers: `GET /users/{id}/certifications`, `POST /users/{id}/certifications`, `GET /tasks/{id}/qualified-mechanics` |
| `internal/api/rest/router.go` | Add certification routes |

#### 1.3 Qualification Validation Flow

```
Task Assignment Request
  |
  v
[1] Check mechanic has active certification for task type
  |
  v
[2] Check mechanic has type rating for aircraft's type
  |
  v
[3] Check certification not expired (expiry_date > now)
  |
  v
[4] Check recency: SUM(hours_logged) in last 24 months >= 6 months equivalent
  |
  v
[5] Check special skills if task requires them (proficiency >= min_level)
  |
  v
[6] At sign-off: re-validate + capture qualification snapshot in audit log
```

#### 1.4 Frontend Changes

- **Mechanic Profile Page**: Tab showing certifications, type ratings, skills with expiry dates
- **Task Assignment Dialog**: Filter mechanic dropdown to only show qualified mechanics; show warning badges for expiring certifications
- **New API Endpoint**: `GET /tasks/{id}/qualified-mechanics` returns ranked list of eligible mechanics
- **Dashboard Widget**: "Certifications Expiring Soon" alert card

#### 1.5 Effort & Priority

| Metric | Value |
|--------|-------|
| **Priority** | HIGH - Required for regulatory compliance |
| **Backend Effort** | ~2 weeks (schema + service + API) |
| **Frontend Effort** | ~1 week (profile page + assignment dialog) |
| **Risk** | Medium - Requires seeding certification types for FAA/EASA/ECAA |
| **Impact** | Prevents assigning unqualified mechanics; audit-ready personnel records |

---

## 2. Dynamic Rescheduling When Parts Are Unavailable

### Industry Context

Parts supply chain disruptions are a top industry challenge. AOG (Aircraft On Ground) events cost airlines $150,000+/hour. Leading MRO systems handle this through:

- **Constraint-based scheduling**: AMOS and Maintenix use constraint programming (CP) models treating maintenance as a Resource-Constrained Project Scheduling Problem (RCPSP)
- **Cascade impact analysis**: When a part becomes unavailable, the system identifies all dependent tasks and suggests rescheduling options
- **Aurora (Stottler Henke)**: Uses AI-driven constraint satisfaction with bottleneck avoidance to schedule around resource constraints, used by Boeing and US Navy
- **ePlane AI Schedule AI**: Cloud-native MRO scheduling using OR-Tools CP-SAT or Gurobi for real-time optimization
- **MRO-PRO**: Real-time line maintenance planning with traffic-light indicators for resource coverage gaps

### Current State in AMSS

- `part_reservations` table links parts to tasks with states (reserved, used, released)
- Task completion requires `AllReservationsClosed` and `RequiredPartsUsed` checks
- **No dependency tracking between tasks** (no task-depends-on-task relationships)
- **No automatic rescheduling** when a reserved part becomes unavailable
- **No priority/urgency field** on tasks (all tasks treated equally)
- **No cascade notification** when schedule changes affect downstream work

### Proposed Changes

#### 2.1 New Database Schema (Migration 00009)

```sql
-- Task dependencies (prerequisite relationships)
CREATE TABLE task_dependencies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  depends_on_task_id uuid NOT NULL,
  dependency_type text NOT NULL DEFAULT 'finish_to_start'
    CHECK (dependency_type IN ('finish_to_start', 'start_to_start', 'finish_to_finish')),
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  FOREIGN KEY (org_id, depends_on_task_id) REFERENCES maintenance_tasks(org_id, id),
  UNIQUE (org_id, task_id, depends_on_task_id),
  CHECK (task_id != depends_on_task_id)
);

-- Task priority levels
CREATE TYPE task_priority AS ENUM ('routine', 'urgent', 'aog', 'critical');

ALTER TABLE maintenance_tasks ADD COLUMN priority task_priority NOT NULL DEFAULT 'routine';
ALTER TABLE maintenance_tasks ADD COLUMN reschedule_count int NOT NULL DEFAULT 0;
ALTER TABLE maintenance_tasks ADD COLUMN reschedule_reason text;
ALTER TABLE maintenance_tasks ADD COLUMN original_start_time timestamptz;
ALTER TABLE maintenance_tasks ADD COLUMN original_end_time timestamptz;

-- Part availability tracking (stock level thresholds)
ALTER TABLE part_definitions ADD COLUMN min_stock_level int NOT NULL DEFAULT 0;
ALTER TABLE part_definitions ADD COLUMN reorder_point int NOT NULL DEFAULT 0;
ALTER TABLE part_definitions ADD COLUMN lead_time_days int;  -- Expected delivery time

-- Schedule change events (for notification and audit)
CREATE TABLE schedule_change_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  change_type text NOT NULL
    CHECK (change_type IN ('rescheduled', 'cancelled', 'priority_changed', 'mechanic_reassigned')),
  reason text NOT NULL,                   -- 'part_unavailable', 'mechanic_unavailable', 'aog_preemption', 'manual'
  old_start_time timestamptz,
  new_start_time timestamptz,
  old_end_time timestamptz,
  new_end_time timestamptz,
  triggered_by uuid,                      -- user who initiated
  affected_task_ids uuid[],               -- downstream tasks also rescheduled
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id)
);
```

#### 2.2 Rescheduling Engine Design

```
Part Becomes Unavailable (stock = 0 or reservation fails)
  |
  v
[1] IDENTIFY affected tasks
    - Query: tasks with reservation for this part + state IN (scheduled, in_progress)
  |
  v
[2] ASSESS impact per task
    - Can task proceed without this specific part? (check if alternative part exists)
    - What is the expected restock date (lead_time_days)?
    - What is the task priority? (AOG tasks get priority handling)
  |
  v
[3] FIND downstream dependencies
    - Traverse task_dependencies graph: all tasks depending on affected tasks
    - Compute cascade depth and total impacted work hours
  |
  v
[4] GENERATE rescheduling options
    Option A: Delay task to expected restock date + propagate to dependents
    Option B: Substitute with compatible part from inventory
    Option C: Split task - complete non-part-dependent work, defer rest
    Option D: Cancel and reschedule entire chain
  |
  v
[5] NOTIFY affected parties
    - WebSocket: TASK_RESCHEDULED event to all connected org users
    - In-app: notification to assigned mechanic + scheduler
    - Schedule change event logged for audit trail
  |
  v
[6] SCHEDULER REVIEW
    - Scheduler sees rescheduling suggestions in dashboard
    - Can accept, modify, or override suggestions
    - System applies changes atomically (all-or-nothing for dependency chains)
```

#### 2.3 Backend Changes

| File | Change |
|------|--------|
| `internal/domain/task.go` | Add `Priority`, `RescheduleCount`, `OriginalStartTime` fields; add `TaskDependency` struct |
| `internal/app/services/scheduling_service.go` | New service: dependency graph traversal, cascade rescheduling, option generation |
| `internal/app/services/parts_service.go` | Add stock level monitoring; trigger rescheduling check on part status change |
| `internal/api/rest/handlers/scheduling.go` | New endpoints: `POST /tasks/{id}/reschedule`, `GET /tasks/{id}/dependencies`, `GET /scheduling/conflicts` |
| `internal/infra/ws/hub.go` | Add `TASK_RESCHEDULED`, `SCHEDULE_CONFLICT` event types |

#### 2.4 Frontend Changes

- **Schedule Conflict Banner**: Prominent banner on maintenance page when conflicts exist
- **Dependency Visualization**: Simple dependency arrows on calendar/kanban views
- **Rescheduling Dialog**: Shows impact analysis (affected tasks, new dates) before confirmation
- **Priority Badges**: Color-coded priority on task cards (routine=gray, urgent=yellow, AOG=red)
- **Parts Availability Indicator**: On task detail, show part availability status with expected restock

#### 2.5 Effort & Priority

| Metric | Value |
|--------|-------|
| **Priority** | HIGH - Directly impacts fleet availability |
| **Backend Effort** | ~2.5 weeks (schema + dependency engine + scheduling service) |
| **Frontend Effort** | ~1.5 weeks (conflict UI + rescheduling dialog + dependency view) |
| **Risk** | Medium-High - Dependency graph traversal needs cycle detection; rescheduling atomicity |
| **Impact** | Reduces AOG time by enabling proactive rescheduling; improves planner productivity |

---

## 3. Multi-Regulatory Compliance Templates (FAA/EASA/ECAA)

### Industry Context

Airlines operating internationally must comply with multiple regulatory frameworks simultaneously. Ethiopian Airlines MRO, for example, holds FAA Part 145, EASA Part 145, and ECAA approvals, each with distinct requirements:

**Key Regulatory Differences**:

| Area | FAA (14 CFR) | EASA (Part-M/145) | ECAA |
|------|-------------|-------------------|------|
| Record retention | 1 year (Part 91.417) | 3 years (145.A.55) | Aligns with ICAO (5 years) |
| Return to service | FAA Form 337 (major repairs) | EASA Form 1 (CRS) | ECAA CRS form |
| Personnel records | Updated within 5 business days | Updated per 145.A.35 | Per ICAO Annex 1 |
| AD tracking | FAA AD system | EASA AD system | Both + ECAA supplements |
| Digital records | AC 120-78A (accepted) | Regulation on e-records | Case-by-case |
| Maintenance program | MSG-3 based | MSG-3 + operator's AMP | ICAO/EASA aligned |

**Ethiopian CAA (ECAA)** is largely aligned with ICAO standards and has been harmonizing with EASA. Ethiopian Airlines MRO is approved for B737, B757, B767, B777, B787, Q400, and MD11 under FAA, with B737NG, B757, B767, B777 under EASA.

Leading MRO systems handle this through:
- **OASES**: Ingests ADs/SBs from official sources, cross-references with fleet, auto-generates work packages with materials documentation
- **AMOS**: Single master data set serving all regulatory requirements; configurable per authority
- **Smart 145**: Built-in Part 145 compliance workflows with configurable regulatory templates

### Current State in AMSS

- `compliance_items` table is generic: `description`, `result` (pass/fail/pending), `sign_off_user_id`
- Frontend has mock compliance categories (AD, SB, inspection, certification, training) but these aren't backed by the API
- **No regulatory authority association** on compliance items
- **No AD/SB tracking** tied to specific aircraft or fleet-wide applicability
- **No configurable compliance templates** per authority
- **No work order template generation** from regulatory requirements
- **No record retention policy** differentiated by authority

### Proposed Changes

#### 3.1 New Database Schema (Migration 00010)

```sql
-- Regulatory authorities
CREATE TABLE regulatory_authorities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,              -- 'FAA', 'EASA', 'ECAA', 'TCCA'
  name text NOT NULL,
  country text NOT NULL,
  record_retention_years int NOT NULL,    -- 1 for FAA, 3 for EASA, 5 for ECAA
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Organization regulatory registrations
CREATE TABLE org_regulatory_registrations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  registration_number text NOT NULL,      -- e.g., FAA Part 145 cert number
  scope text,                             -- What work is approved
  effective_date date NOT NULL,
  expiry_date date,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'expired', 'suspended')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, authority_id)
);

-- Compliance directive types
CREATE TYPE directive_type AS ENUM ('ad', 'sb', 'eo', 'tcds', 'stc');
CREATE TYPE directive_applicability AS ENUM ('mandatory', 'recommended', 'optional');

-- Airworthiness Directives & Service Bulletins
CREATE TABLE compliance_directives (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  directive_type directive_type NOT NULL,
  reference_number text NOT NULL,         -- e.g., 'FAA AD 2024-15-08'
  title text NOT NULL,
  description text,
  applicability directive_applicability NOT NULL,
  affected_aircraft_types uuid[],         -- References aircraft_types
  effective_date date NOT NULL,
  compliance_deadline date,               -- NULL for recurring/ongoing
  recurrence_interval text,               -- e.g., 'every 500 FH', 'every 12 months'
  superseded_by uuid REFERENCES compliance_directives(id),
  source_url text,                        -- Link to official document
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Aircraft-specific directive compliance status
CREATE TABLE aircraft_directive_compliance (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  aircraft_id uuid NOT NULL,
  directive_id uuid NOT NULL REFERENCES compliance_directives(id),
  status text NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'in_progress', 'compliant', 'not_applicable', 'overdue')),
  compliance_date timestamptz,
  next_due_date timestamptz,              -- For recurring directives
  task_id uuid,                           -- Link to maintenance task that addresses this
  signed_off_by uuid,
  signed_off_at timestamptz,
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, aircraft_id) REFERENCES aircraft(org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  UNIQUE (org_id, aircraft_id, directive_id)
);

-- Compliance document templates (per authority)
CREATE TABLE compliance_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  template_code text NOT NULL,            -- e.g., 'FAA_337', 'EASA_FORM1', 'CRS'
  name text NOT NULL,
  description text,
  required_fields jsonb NOT NULL,         -- Schema definition for form fields
  template_content text,                  -- HTML/markdown template
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (authority_id, template_code)
);

-- Enhance existing compliance_items with authority link
ALTER TABLE compliance_items ADD COLUMN authority_id uuid REFERENCES regulatory_authorities(id);
ALTER TABLE compliance_items ADD COLUMN directive_id uuid REFERENCES compliance_directives(id);
ALTER TABLE compliance_items ADD COLUMN category text;  -- 'ad', 'sb', 'inspection', etc.
```

#### 3.2 Backend Changes

| File | Change |
|------|--------|
| `internal/domain/compliance.go` | Expand with `RegulatoryAuthority`, `ComplianceDirective`, `AircraftDirectiveCompliance`, `ComplianceTemplate` types |
| `internal/app/services/compliance_service.go` | Extend: directive tracking per aircraft, due date computation, fleet-wide applicability scanning |
| `internal/app/services/directive_service.go` | New: CRUD for directives, import from external sources, check fleet applicability |
| `internal/api/rest/handlers/directives.go` | New: `GET /directives`, `POST /directives`, `GET /aircraft/{id}/compliance-status`, `POST /aircraft/{id}/directives/{id}/comply` |
| `internal/api/rest/handlers/compliance.go` | Extend existing handlers with authority filtering and directive linking |

#### 3.3 Compliance Workflow

```
New AD/SB Published by Authority
  |
  v
[1] CREATE directive in compliance_directives
    - Specify authority, type, affected aircraft types, deadline
  |
  v
[2] SCAN fleet for affected aircraft
    - Match aircraft.aircraft_type_id against directive.affected_aircraft_types
    - Create aircraft_directive_compliance record per affected aircraft
  |
  v
[3] AUTO-GENERATE compliance items per aircraft
    - Link to directive
    - Set due dates based on directive deadline + aircraft utilization
  |
  v
[4] OPTIONALLY create maintenance tasks
    - Scheduler reviews generated compliance items
    - Creates tasks with linked compliance items
  |
  v
[5] TRACK completion
    - Mechanic completes task + signs off compliance item
    - System updates aircraft_directive_compliance status
    - For recurring directives: compute next_due_date
  |
  v
[6] REPORT generation
    - Per-authority compliance reports using authority-specific templates
    - Record retention enforced per authority's requirement
```

#### 3.4 Frontend Changes

- **Compliance Dashboard Overhaul**: Replace mock data with real AD/SB tracking
- **Authority Filter**: Filter compliance view by regulatory authority (FAA/EASA/ECAA)
- **Directive Management Page**: New page for creating/importing ADs, SBs, EOs
- **Fleet Applicability View**: Matrix showing which aircraft are affected by which directives
- **Compliance Report Export**: Generate authority-specific compliance reports (PDF)
- **Directive Alert Widget**: Dashboard widget showing upcoming compliance deadlines

#### 3.5 Seed Data

Pre-populate `regulatory_authorities`:
```
FAA  | Federal Aviation Administration | US  | 1 year retention
EASA | European Union Aviation Safety Agency | EU | 3 year retention
ECAA | Ethiopian Civil Aviation Authority | ET | 5 year retention
ICAO | Int'l Civil Aviation Organization | INT | 5 year retention
```

Pre-populate `compliance_templates`:
- FAA Form 337 (Major Repair/Alteration)
- FAA Form 8610-2 (Airman Certificate)
- EASA Form 1 (Authorized Release Certificate)
- Certificate of Release to Service (CRS) - generic
- Maintenance Release (Return to Service)

#### 3.6 Effort & Priority

| Metric | Value |
|--------|-------|
| **Priority** | HIGH - Core differentiator for Ethiopian/African market |
| **Backend Effort** | ~3 weeks (schema + directive service + compliance workflow + templates) |
| **Frontend Effort** | ~2 weeks (directive management + compliance overhaul + reports) |
| **Risk** | Medium - Regulatory rule encoding requires domain expertise validation |
| **Impact** | Enables multi-authority operations; makes AMSS viable for Ethiopian Airlines MRO and similar operators |

---

## 4. Real-Time Maintenance Dashboard via WebSocket

### Industry Context

Modern MRO control centers require real-time operational visibility. Industry leaders provide:

- **AMOS**: Configurable dashboard framework with role-based widget selection, fleet tracker, maintenance control
- **Ramco**: Role-based analytics for 5 user levels (executive, director, manager, mechanic, admin), real-time fleet tracking, predictive analytics
- **IFS Maintenix**: Production Planning & Control (PP&C) with critical path analysis, shift-level drill-down, real-time work updates
- **Veryon**: Mobile-first real-time compliance tools with eLogbook integration

**Key KPIs tracked in real-time**:
- Fleet Availability Rate (target: >95%)
- AOG Count and Duration
- Turnaround Time (TAT) per check type
- Mechanic Utilization Rate (target: 75-85%)
- Task Completion Rate (target: >95%)
- Parts Fill Rate (target: >97%)
- Certification Expiry Countdown

### Current State in AMSS

- WebSocket infrastructure exists (`internal/infra/ws/hub.go`) with org-scoped broadcasting
- Frontend has `WebSocketManager` singleton with event types: `task:created/updated/deleted/status_changed`, `aircraft:status_changed`, `part:low_stock`, `notification`, `user:online/offline`
- Dashboard shows: total aircraft, pending/overdue/completed tasks, on-time rate, avg completion time, fleet utilization, weekly bar chart, maintenance-by-type pie chart
- **Missing**: AOG-specific tracking, TAT metrics, mechanic utilization, parts fill rate, real-time alerts, role-based views, Gantt/timeline visualization

### Proposed Changes

#### 4.1 New WebSocket Event Types

```go
// Add to ws/hub.go or a constants file
const (
    EventDashboardMetrics      = "dashboard:metrics_snapshot"
    EventAOGDeclared           = "aircraft:aog_declared"
    EventAOGResolved           = "aircraft:aog_resolved"
    EventTATWarning            = "task:tat_warning"
    EventTaskOverdue           = "task:overdue"
    EventCertExpiring          = "compliance:cert_expiring"
    EventScheduleConflict      = "schedule:conflict"
    EventTaskRescheduled       = "task:rescheduled"
    EventMechanicAssignment    = "resource:assignment_changed"
)
```

#### 4.2 Server-Side Metrics Aggregation

New background worker job (runs every 30 seconds):

```go
type DashboardMetricsSnapshot struct {
    FleetAvailability    float64            `json:"fleet_availability"`
    AOGCount             int                `json:"aog_count"`
    AOGAircraft          []AOGAircraftInfo  `json:"aog_aircraft"`
    ActiveTasks          TaskCounts         `json:"active_tasks"`
    OverdueTasks         int                `json:"overdue_tasks"`
    CompletedToday       int                `json:"completed_today"`
    MechanicUtilization  float64            `json:"mechanic_utilization"`
    PartsFillRate        float64            `json:"parts_fill_rate"`
    OnTimeRate           float64            `json:"on_time_rate"`
    AvgTAT               map[string]float64 `json:"avg_tat"`  // per task type
    CertsExpiringIn30d   int                `json:"certs_expiring_30d"`
    Timestamp            string             `json:"timestamp"`
}
```

The worker queries PostgreSQL, computes the snapshot, and broadcasts via WebSocket to all connected dashboard clients. This avoids each client computing metrics independently.

#### 4.3 Backend Changes

| File | Change |
|------|--------|
| `cmd/worker/main.go` | Add dashboard metrics cron job (every 30s) |
| `internal/app/services/metrics_service.go` | New service: compute all dashboard KPIs from database |
| `internal/app/services/alert_service.go` | New service: threshold monitoring, alert creation, escalation |
| `internal/infra/ws/hub.go` | Add new event type constants; add targeted broadcast (to specific roles) |
| `internal/domain/alert.go` | New: `Alert` struct with level (info/warning/critical), acknowledgement, escalation |

#### 4.4 Alert System Schema

```sql
CREATE TYPE alert_level AS ENUM ('info', 'warning', 'critical');

CREATE TABLE alerts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  level alert_level NOT NULL,
  category text NOT NULL,                 -- 'parts', 'compliance', 'task', 'aircraft', 'resource'
  title text NOT NULL,
  description text,
  entity_type text NOT NULL,
  entity_id uuid NOT NULL,
  threshold_value numeric,
  current_value numeric,
  acknowledged boolean NOT NULL DEFAULT false,
  acknowledged_by uuid,
  acknowledged_at timestamptz,
  resolved boolean NOT NULL DEFAULT false,
  resolved_at timestamptz,
  escalation_level int NOT NULL DEFAULT 0,
  auto_escalate_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, acknowledged_by) REFERENCES users(org_id, id)
);
```

#### 4.5 Frontend Changes

##### New Dashboard Widgets

| Widget | Role | Description |
|--------|------|-------------|
| **AOG Counter** | All | Red pulsing badge with AOG aircraft count and details |
| **Fleet Status Board** | Director/Planner | Grid of aircraft tiles with status colors, per-aircraft maintenance status |
| **Mechanic Utilization** | Planner/Director | Bar chart showing utilization per mechanic for current shift |
| **Parts Fill Rate** | Planner | Gauge showing % of part requests fulfilled from stock |
| **TAT Tracker** | Director | Average turnaround time per check type vs. target |
| **Alert Feed** | All | Real-time alert stream with acknowledge/dismiss actions |
| **Certification Countdown** | Compliance | List of certifications expiring within 30/60/90 days |

##### Role-Based Dashboard Routing

```typescript
// Different widget configurations per role
const dashboardConfig: Record<UserRole, WidgetConfig[]> = {
  admin: [...allWidgets],
  tenant_admin: [...allWidgets],
  scheduler: [fleetBoard, ganttTimeline, partsAvailability, mechanicCalendar, alertFeed],
  mechanic: [myTasks, currentAircraft, partRequests, signOffQueue, alertFeed],
  auditor: [complianceStatus, auditTrail, certCountdown, directiveTracker],
}
```

##### Real-Time State Management

Extend existing WebSocket integration with:
- RTK Query cache invalidation on WebSocket events (bridge pattern)
- Batched update buffer for high-frequency events (200ms flush interval)
- Selective subscription based on visible dashboard panels (IntersectionObserver)

```typescript
// Bridge WebSocket events to RTK Query cache
wsManager.subscribe(WS_EVENTS.TASK_STATUS_CHANGED, () => {
  store.dispatch(api.util.invalidateTags(['Tasks']))
})
wsManager.subscribe(WS_EVENTS.AIRCRAFT_STATUS_CHANGED, () => {
  store.dispatch(api.util.invalidateTags(['Aircraft']))
})
wsManager.subscribe('dashboard:metrics_snapshot', (data) => {
  store.dispatch(setDashboardMetrics(data))
})
```

#### 4.6 Effort & Priority

| Metric | Value |
|--------|-------|
| **Priority** | MEDIUM-HIGH - Enhances existing infrastructure |
| **Backend Effort** | ~2 weeks (metrics worker + alert system + new WS events) |
| **Frontend Effort** | ~2.5 weeks (new widgets + role routing + alert feed + real-time bridge) |
| **Risk** | Low - Builds on existing WebSocket infrastructure |
| **Impact** | Transforms dashboard from static to real-time; enables MRO control center use case |

---

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)

| Week | Deliverable | Dependencies |
|------|-------------|--------------|
| 1 | Database migrations 00008-00010 (all schema changes) | None |
| 2 | Certification domain + service + API endpoints | Migration 00008 |
| 3 | Directive domain + service + API endpoints | Migration 00010 |
| 4 | Dashboard metrics worker + new WS events | None |

### Phase 2: Core Features (Weeks 5-8)

| Week | Deliverable | Dependencies |
|------|-------------|--------------|
| 5 | Task assignment qualification validation | Phase 1 Week 2 |
| 6 | Task dependencies + rescheduling engine | Migration 00009 |
| 7 | Alert system (backend) + compliance workflow | Phase 1 Week 3 |
| 8 | Frontend: Mechanic profile + certification pages | Phase 1 Week 2 |

### Phase 3: Integration (Weeks 9-12)

| Week | Deliverable | Dependencies |
|------|-------------|--------------|
| 9 | Frontend: Compliance dashboard overhaul + directive management | Phase 2 Week 7 |
| 10 | Frontend: Real-time dashboard widgets + role-based routing | Phase 2 Week 4 |
| 11 | Frontend: Rescheduling dialog + dependency visualization | Phase 2 Week 6 |
| 12 | Integration testing + UAT + deployment | All above |

### Total Estimated Effort

| Area | Effort |
|------|--------|
| Backend | ~9.5 weeks |
| Frontend | ~7 weeks |
| Testing & QA | ~2 weeks |
| **Total** | **~12 weeks** (with parallelization) |

---

## Architecture Decision Records

### ADR-1: Certification Validation at Multiple Gates

**Decision**: Validate mechanic qualifications at assignment time AND at sign-off time.
**Rationale**: A certification may expire between assignment and completion. The sign-off validation is the legally binding check per FAA/EASA.
**Trade-off**: Slightly more complex; a task may fail at completion if cert expired during work. Mitigation: alert mechanic when cert approaches expiry.

### ADR-2: Server-Side Dashboard Aggregation

**Decision**: Compute dashboard KPIs server-side in a background worker, push via WebSocket.
**Rationale**: Avoids N clients each computing fleet availability from raw data. Ensures consistency across concurrent viewers. Reduces database load.
**Trade-off**: 30-second staleness. Acceptable for operational dashboards; individual entity changes still pushed in real-time.

### ADR-3: Dependency Graph Without Constraint Solver

**Decision**: Implement task dependencies with graph traversal and cascade notifications, but NOT a full constraint-satisfaction solver.
**Rationale**: A CP-SAT solver (OR-Tools, Gurobi) would be ideal but adds significant complexity and a Python/C++ dependency to a Go stack. The current scale (10-50 aircraft per org) doesn't justify it. Start with graph-based cascade analysis; if demand grows, introduce a solver microservice.
**Trade-off**: No automatic optimal rescheduling; scheduler reviews suggestions manually. Acceptable for current scale.

### ADR-4: Multi-Authority via Configuration, Not Code Branches

**Decision**: Regulatory differences encoded as data (in `regulatory_authorities` and `compliance_templates` tables), not as code branches.
**Rationale**: Adding a new authority (e.g., TCCA for Canada) should require database inserts, not code changes. Template-driven compliance report generation keeps the system extensible.
**Trade-off**: Some complex authority-specific rules may not fit a template model. Mitigation: `required_fields` JSONB allows flexible form schemas.

---

## Success Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Qualification validation on assignment | 0% | 100% | All task assignments checked |
| Certification expiry tracking | None | 100% of mechanics | Tracked with 90/60/30 day alerts |
| Average reschedule response time | Manual/days | <1 hour | From part-unavailable to new schedule |
| Regulatory compliance coverage | FAA only (partial) | FAA + EASA + ECAA | All directives tracked per aircraft |
| Dashboard refresh latency | Manual refresh | <30 seconds | Real-time via WebSocket |
| AOG visibility | Not tracked | Real-time | Dedicated AOG counter with alerts |

---

## Sources

- [14 CFR Part 145 Subpart D - Personnel](https://www.ecfr.gov/current/title-14/chapter-I/subchapter-H/part-145/subpart-D)
- [EASA Part-66 License Categories](https://www.easa.europa.eu/en/the-agency/faqs/part-66)
- [EASA Part-145 Requirements](https://www.easa.europa.eu/en/the-agency/faqs/part-145)
- [Ethiopian Airlines MRO Certifications](https://corporate.ethiopianairlines.com/mro/about-us/certification)
- [Swiss-AS AMOS Modules](https://www.swiss-as.com/amos-mro/modules)
- [Ramco Aviation MRO Solution](https://www.ramco.com/products/aviation-software/maintenance-repair-and-overhaul/)
- [IFS Maintenix PP&C](https://www.ifs.com/assets/enterprise-asset-management/ifs-maintenix-production-planning-control)
- [Stottler Henke Aurora MRO Scheduling](https://stottlerhenke.com/products/aurora/mro/)
- [ePlane AI Schedule AI](https://www.eplaneai.com/blog/schedule-ai-real-time-optimization-of-mro-scheduling)
- [OASES MRO Software](https://www.oases.aero/)
- [Boeing Airplane Health Management](https://services.boeing.com/maintenance-engineering/maintenance-optimization/airplane-health-management-ahm)
- [CP in Aircraft Maintenance Scheduling (ScienceDirect)](https://www.sciencedirect.com/science/article/abs/pii/S0969699724000024)
- [Aviation MRO AI Workflow Optimization (GJETA 2025)](http://gjeta.com/sites/default/files/GJETA-2025-0122.pdf)
- [Ramco: AI in Aviation MRO](https://www.ramco.com/blog/aviation/ai-in-aviation-mro)
- [FAA AC 120-78A Digital Records](https://www.faa.gov/regulations_policies/advisory_circulars)
- [EASA Part 145.A.55 Record Keeping](https://sofemaonline.com/about/blog/concerning-maintenance-records-i-a-w-easa-part-145-a-55)
