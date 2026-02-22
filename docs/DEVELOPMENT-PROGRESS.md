# AMSS Development Progress Tracker

> Strategic Improvements Implementation — Started February 2026

---

## Phase 1: Foundation (Weeks 1-4)

### Week 1 — Database Migrations

| # | Task | Status | Notes |
|---|------|--------|-------|
| 1.1 | Migration 00008: Certification & skill tables | ✅ DONE | certification_types, aircraft_types, employee_certifications, employee_type_ratings, skill_types, employee_skills, employee_recency_log, task_skill_requirements |
| 1.2 | Migration 00009: Task dependencies & scheduling | ✅ DONE | task_dependencies, task_priority enum, schedule_change_events, part_definitions stock columns |
| 1.3 | Migration 00010: Regulatory compliance tables | ✅ DONE | regulatory_authorities, org_regulatory_registrations, compliance_directives, aircraft_directive_compliance, compliance_templates |
| 1.4 | Migration 00011: Alert system tables | ✅ DONE | alert_level enum, alerts table |
| 1.5 | Migration 00012: Seed data | ✅ DONE | Regulatory authorities (FAA/EASA/ECAA/ICAO), certification types, compliance templates, aircraft types, skill types |
| 1.6 | Run migrations in UAT | ✅ DONE | All 12 migrations applied cleanly (made idempotent for UAT) |

### Week 2 — Certification Domain & Service

| # | Task | Status | Notes |
|---|------|--------|-------|
| 2.1 | Domain types: `certification.go` | ✅ DONE | CertificationType, EmployeeCertification, EmployeeTypeRating, SkillType, EmployeeSkill, TaskSkillRequirement, QualificationCheckResult |
| 2.2 | Repository: `certification_repository.go` | ✅ DONE | CRUD + query by user, expiring certs, recency check, qualified mechanics lookup |
| 2.3 | Service: `certification_service.go` | ✅ DONE | CRUD, expiry checking, recency computation, qualification check, qualified mechanic lookup |
| 2.4 | REST handlers: `certifications.go` | ✅ DONE | GET/POST /users/{id}/certifications, GET /qualified-mechanics, GET /certification-types, GET /users/{id}/qualification-check |
| 2.5 | Router integration | ✅ DONE | All certification routes wired with RBAC |
| 2.6 | Tests: certification service | ⬜ TODO | Unit + integration tests |

### Week 3 — Directive Domain & Service

| # | Task | Status | Notes |
|---|------|--------|-------|
| 3.1 | Domain types: `directive.go` | ✅ DONE | RegulatoryAuthority, ComplianceDirective, AircraftDirectiveCompliance, ComplianceTemplate, OrgRegulatoryRegistration |
| 3.2 | Repository: `directive_repository.go` | ✅ DONE | CRUD + fleet applicability scan + per-aircraft compliance status + templates |
| 3.3 | Service: `directive_service.go` | ✅ DONE | Directive CRUD, fleet scanning, compliance tracking |
| 3.4 | REST handlers: `directives.go` | ✅ DONE | GET/POST /directives, GET /aircraft/{id}/compliance-status, POST /directives/{id}/scan-fleet, GET /compliance-templates/{id} |
| 3.5 | Router integration | ✅ DONE | All directive routes wired |
| 3.6 | Tests: directive service | ⬜ TODO | Unit + integration tests |

### Week 4 — Dashboard Metrics Worker & WebSocket Events

| # | Task | Status | Notes |
|---|------|--------|-------|
| 4.1 | Domain: `alert.go` | ✅ DONE | Alert struct with level, category, acknowledgement, escalation |
| 4.2 | Service: `metrics_service.go` | ✅ DONE | Domain type, repository (optimized SQL), service, handler, route wired |
| 4.3 | Service: `alert_service.go` | ✅ DONE | Alert CRUD, acknowledge, resolve, count unresolved |
| 4.4 | Repository: `alert_repository.go` | ✅ DONE | Create, list, acknowledge, resolve, count unresolved |
| 4.5 | REST handlers: `alerts.go` | ✅ DONE | GET /alerts, POST /alerts/{id}/acknowledge, POST /alerts/{id}/resolve |
| 4.6 | Worker: dashboard metrics cron job | ✅ DONE | MetricsBroadcaster job, 30-second interval, broadcasts to connected orgs |
| 4.7 | WebSocket: new event types | ✅ DONE | 15 event types: dashboard, aircraft, task, compliance, scheduling, alerts, parts |
| 4.8 | WebSocket: role-based broadcast | ✅ DONE | BroadcastToRoles() + ConnectedOrgIDs() + Client.Role field |
| 4.9 | Tests: metrics + alerts | ⬜ TODO | Unit + integration tests |

---

## Phase 2: Core Features (Weeks 5-8)

### Week 5 — Task Assignment Qualification Validation

| # | Task | Status | Notes |
|---|------|--------|-------|
| 5.1 | Update `task.go` domain | ✅ DONE | Qualification check in CanTransition() context already exists |
| 5.2 | Update `task_service.go` | ✅ DONE | Added Certs dependency, validateMechanicQualification() method |
| 5.3 | Qualification validation flow | ✅ DONE | 6-step check at Create, Update (mechanic change), and Completion (sign-off) |
| 5.4 | Tests: qualification validation | ⬜ TODO | Edge cases: expired cert, missing type rating, insufficient recency |

### Week 6 — Task Dependencies & Rescheduling Engine

| # | Task | Status | Notes |
|---|------|--------|-------|
| 6.1 | Domain: `scheduling.go` | ✅ DONE | TaskDependency, TaskPriority, ScheduleChangeEvent, RescheduleOption |
| 6.2 | Repository: `scheduling_repository.go` | ✅ DONE | TaskDependencyRepository + ScheduleChangeRepository |
| 6.3 | Service: `scheduling_service.go` | ✅ DONE | Dependency CRUD, cycle detection, cascade rescheduling, conflict detection |
| 6.4 | REST handlers: `scheduling.go` | ✅ DONE | POST /tasks/{id}/reschedule, GET/POST /tasks/{id}/dependencies, GET /scheduling/conflicts |
| 6.5 | Router integration | ✅ DONE | All scheduling routes wired |
| 6.6 | Update `parts_service.go` | ✅ DONE | Stock level monitoring on part use, auto-alert on low/out of stock, updated PartDefinition domain with stock fields |
| 6.7 | Tests: scheduling engine | ⬜ TODO | Cycle detection, cascade propagation, atomicity |

### Week 7 — Alert System & Compliance Workflow

| # | Task | Status | Notes |
|---|------|--------|-------|
| 7.1 | Alert service integration | ✅ DONE | AlertTrigger job: cert expiry (30/60/90d), task overdue, directive overdue. Low-stock alerts via PartReservationService |
| 7.2 | Compliance workflow | ✅ DONE | ScanFleet sets next_due from deadline, UpdateCompliance computes next_due for recurring directives (d/m/y intervals) |
| 7.3 | REST handlers: `alerts.go` | ✅ DONE | GET /alerts, POST /alerts/{id}/acknowledge, POST /alerts/{id}/resolve |
| 7.4 | Tests: alerts + compliance workflow | ⬜ TODO | Threshold triggers, escalation, recurring directive next-due |

### Week 8 — Frontend: Mechanic Profile & Certification Pages

| # | Task | Status | Notes |
|---|------|--------|-------|
| 8.1 | API integration: certification endpoints | ⬜ TODO | RTK Query hooks for cert CRUD |
| 8.2 | Mechanic profile: certifications tab | ⬜ TODO | List certs, type ratings, skills with expiry badges |
| 8.3 | Task assignment dialog update | ⬜ TODO | Filter to qualified mechanics, show warning badges |
| 8.4 | Dashboard widget: expiring certifications | ⬜ TODO | 30/60/90 day countdown alerts |

---

## Phase 3: Integration (Weeks 9-12)

### Week 9 — Frontend: Compliance Dashboard Overhaul

| # | Task | Status | Notes |
|---|------|--------|-------|
| 9.1 | Directive management page | ⬜ TODO | Create/import ADs, SBs, EOs |
| 9.2 | Fleet applicability matrix | ⬜ TODO | Which aircraft affected by which directives |
| 9.3 | Authority filter on compliance views | ⬜ TODO | FAA/EASA/ECAA filter tabs |
| 9.4 | Compliance report export | ⬜ TODO | Authority-specific PDF reports |

### Week 10 — Frontend: Real-Time Dashboard Widgets

| # | Task | Status | Notes |
|---|------|--------|-------|
| 10.1 | AOG counter widget | ⬜ TODO | Red pulsing badge with details |
| 10.2 | Fleet status board | ⬜ TODO | Aircraft tile grid with status colors |
| 10.3 | Mechanic utilization chart | ⬜ TODO | Per-mechanic bar chart for current shift |
| 10.4 | Parts fill rate gauge | ⬜ TODO | % of part requests fulfilled |
| 10.5 | Alert feed widget | ⬜ TODO | Real-time stream with acknowledge/dismiss |
| 10.6 | Role-based dashboard routing | ⬜ TODO | Different widget configs per role |
| 10.7 | WebSocket → RTK Query bridge | ⬜ TODO | Cache invalidation on WS events, batched updates |

### Week 11 — Frontend: Rescheduling & Dependencies

| # | Task | Status | Notes |
|---|------|--------|-------|
| 11.1 | Schedule conflict banner | ⬜ TODO | Prominent banner when conflicts exist |
| 11.2 | Dependency visualization | ⬜ TODO | Dependency arrows on calendar/kanban |
| 11.3 | Rescheduling dialog | ⬜ TODO | Impact analysis + confirmation |
| 11.4 | Priority badges on task cards | ⬜ TODO | Color-coded: routine/urgent/AOG/critical |
| 11.5 | Parts availability indicator | ⬜ TODO | Per-task part status with restock dates |

### Week 12 — Integration Testing & Deployment

| # | Task | Status | Notes |
|---|------|--------|-------|
| 12.1 | End-to-end integration tests | ⬜ TODO | Full workflow: directive → task → assign → complete → compliance |
| 12.2 | UAT testing | ⬜ TODO | All roles, all workflows |
| 12.3 | Performance testing | ⬜ TODO | WebSocket load, dashboard query performance |
| 12.4 | Production deployment | ⬜ TODO | Migrations, Docker build, K3s deploy |
| 12.5 | Documentation update | ⬜ TODO | API docs, user guides |

---

## Summary

| Phase | Scope | Progress |
|-------|-------|----------|
| Phase 1: Foundation | Migrations + Domain + Services | 23/25 |
| Phase 2: Core Features | Validation + Scheduling + Alerts | 12/18 |
| Phase 3: Integration | Frontend + Testing + Deploy | 0/17 |
| **Total** | | **35/60** |

---

## Files Created/Modified

### New Files
- `migrations/00008_certification_skill_tables.sql`
- `migrations/00009_task_dependencies_scheduling.sql`
- `migrations/00010_regulatory_compliance_tables.sql`
- `migrations/00011_alert_system.sql`
- `migrations/00012_seed_reference_data.sql`
- `internal/domain/certification.go`
- `internal/domain/directive.go`
- `internal/domain/alert.go`
- `internal/domain/scheduling.go`
- `internal/infra/postgres/certification_repository.go`
- `internal/infra/postgres/directive_repository.go`
- `internal/infra/postgres/alert_repository.go`
- `internal/infra/postgres/scheduling_repository.go`
- `internal/app/services/certification_service.go`
- `internal/app/services/directive_service.go`
- `internal/app/services/alert_service.go`
- `internal/app/services/scheduling_service.go`
- `internal/api/rest/handlers/certifications.go`
- `internal/api/rest/handlers/directives.go`
- `internal/api/rest/handlers/alerts.go`
- `internal/api/rest/handlers/scheduling.go`

### New Files (Phase 2 batch)
- `internal/domain/metrics.go` — DashboardMetrics snapshot struct
- `internal/infra/postgres/metrics_repository.go` — Optimized aggregate SQL for all KPIs
- `internal/app/services/metrics_service.go` — Dashboard metrics service
- `internal/api/rest/handlers/metrics.go` — GET /dashboard/metrics
- `internal/infra/ws/events.go` — 15 WebSocket event type constants
- `internal/jobs/metrics_broadcaster.go` — 30s cron, broadcasts metrics to connected orgs
- `internal/jobs/alert_trigger.go` — 15m cron, checks cert expiry/task overdue/directive overdue

### Modified Files
- `internal/domain/aircraft.go` — Added `AircraftTypeID` field
- `internal/domain/parts.go` — Added MinStockLevel, ReorderPoint, LeadTimeDays to PartDefinition
- `internal/app/ports/repositories.go` — Added 7 new repository interfaces + filter types (including MetricsRepository)
- `internal/api/rest/middleware/services.go` — Added Certifications, Directives, Alerts, Scheduling, Metrics to ServiceRegistry
- `internal/api/rest/router.go` — Wired all new services + routes + shared repos
- `internal/app/services/task_service.go` — Added Certs dependency, validateMechanicQualification() at create/update/complete
- `internal/app/services/part_reservation_service.go` — Added stock monitoring, auto-alert on low/out of stock
- `internal/app/services/directive_service.go` — Enhanced compliance workflow with next_due computation for recurring directives
- `internal/infra/postgres/part_definition_repository.go` — Updated queries for stock level fields
- `internal/infra/ws/hub.go` — Added Client.Role, BroadcastToRoles(), ConnectedOrgIDs()
- `cmd/server/main.go` — Added MetricsBroadcaster, Certs dependency on TaskService
- `cmd/worker/main.go` — Added AlertTrigger job, Certs dependency on TaskService
- `migrations/00005_seed_data.sql` — Made idempotent (skip if Demo Airline exists)
- `migrations/00006_refresh_tokens.sql` — Made idempotent (IF NOT EXISTS)
- `migrations/00007_add_import_file_path.sql` — Made idempotent (DO $$ EXCEPTION)
- `migrations/00008-00012` — Made all idempotent (IF NOT EXISTS, ON CONFLICT DO NOTHING)
