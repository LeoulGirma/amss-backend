# AMSS Quick Reference Guide

**Version:** 1.0
**Last Updated:** 2024-01-15
**Applies To:** AMSS v1.x
**Audience:** All AMSS users
**Review Date:** Quarterly

---

## Document Purpose

This Quick Reference Guide provides rapid lookup information for all AMSS users:

- **Part I: Glossary** - Aviation and AMSS terminology
- **Part II: Common Tasks Cheat Sheet** - Quick task guides by role
- **Part III: Troubleshooting Guide** - Error codes and common issues
- **Part IV: FAQs** - Frequently asked questions by role

**Navigation:**
- Use `Ctrl+F` (Windows/Linux) or `Cmd+F` (Mac) to search for specific terms
- Click section links in the table of contents
- Reference your role-specific user guide for detailed procedures

---

## Table of Contents

### Part I: Glossary
- [Aviation Terminology](#aviation-terminology)
- [AMSS System Terms](#amss-system-terms)
- [Maintenance Concepts](#maintenance-concepts)
- [Regulatory Terms](#regulatory-terms)

### Part II: Common Tasks Cheat Sheet
- [System Administrator Tasks](#system-administrator-tasks)
- [Fleet Manager Tasks](#fleet-manager-tasks)
- [Maintenance Planner Tasks](#maintenance-planner-tasks)
- [Mechanic/Technician Tasks](#mechanictechnician-tasks)
- [Compliance Officer Tasks](#compliance-officer-tasks)

### Part III: Troubleshooting Guide
- [Error Code Reference](#error-code-reference)
- [Common Issues by Symptom](#common-issues-by-symptom)
- [Access & Permission Issues](#access--permission-issues)
- [Data Integrity Issues](#data-integrity-issues)

### Part IV: FAQs
- [System Administrator FAQs](#system-administrator-faqs)
- [Fleet Manager FAQs](#fleet-manager-faqs)
- [Maintenance Planner FAQs](#maintenance-planner-faqs)
- [Mechanic/Technician FAQs](#mechanictechnician-faqs)
- [Compliance Officer FAQs](#compliance-officer-faqs)
- [General System FAQs](#general-system-faqs)

---

# Part I: Glossary

## Aviation Terminology

### A

**A&P Mechanic**
Airframe and Powerplant mechanic - FAA-certified technician authorized to perform aircraft maintenance.

**AD (Airworthiness Directive)**
Mandatory inspection or modification issued by aviation authority (FAA/EASA) for safety-critical issues.

**ATA Chapter**
Air Transport Association specification system dividing aircraft into chapters (e.g., ATA 32 = Landing Gear).

### B

**Base Maintenance**
Heavy maintenance performed in a hangar, typically requiring aircraft grounding for days/weeks.

### C

**Calendar Time**
Time-based maintenance interval (e.g., "every 12 months") regardless of usage.

**Check (A/B/C/D)**
Progressive levels of scheduled maintenance inspections:
- **A-Check:** Minor inspection (50-100 flight hours)
- **B-Check:** Intermediate (300-600 flight hours)
- **C-Check:** Major structural (18-24 months, 3,000-6,000 hours)
- **D-Check:** Heavy maintenance (6-10 years, complete disassembly)

**Compliance**
Adherence to regulatory requirements and manufacturer specifications for airworthiness.

**Cycle**
One takeoff and landing sequence. Used for measuring fatigue on pressurized structures.

### D

**Deferred Maintenance**
Non-critical maintenance postponed under regulatory guidance (e.g., MEL - Minimum Equipment List).

### E

**EASA (European Union Aviation Safety Agency)**
European aviation regulatory authority (equivalent to FAA for EU).

### F

**FAA (Federal Aviation Administration)**
United States aviation regulatory authority.

**Flight Hours**
Total operating time of aircraft measured in hours. Primary metric for maintenance scheduling.

### H

**Hard Time**
Component must be replaced at specified interval regardless of condition.

### I

**Inspection**
Visual or technical examination of aircraft/components to verify airworthiness.

### L

**Line Maintenance**
Routine maintenance performed at airport gate/ramp between flights (minor inspections, servicing).

### M

**MEL (Minimum Equipment List)**
FAA-approved list of equipment that may be inoperative under specific conditions.

### O

**On-Condition**
Component monitored and replaced only when inspection shows degradation.

**Overhaul**
Complete disassembly, inspection, repair, and reassembly of component to like-new condition.

### P

**Part 91**
14 CFR Part 91 - General operating and flight rules (private/corporate operations).

**Part 135**
14 CFR Part 135 - Commuter and on-demand operations (air taxi, charter).

**Part 121**
14 CFR Part 121 - Air carrier operations (scheduled airlines).

**Part 145**
EASA Part-145 - Maintenance organization approvals.

**Preventive Maintenance**
Scheduled maintenance performed to prevent failures before they occur.

### R

**Repair**
Restoration of aircraft/component to airworthy condition after damage or wear.

**Return to Service (RTS)**
Authorization by mechanic/inspector that aircraft is airworthy after maintenance.

### S

**SB (Service Bulletin)**
Manufacturer recommendation for inspection, modification, or improvement (optional unless made mandatory by AD).

**Snag**
Defect or malfunction reported by pilot or discovered during inspection.

### T

**TBO (Time Between Overhaul)**
Manufacturer-recommended interval for component overhaul.

**Tail Number / Registration**
Unique aircraft identifier (e.g., N12345 for US-registered aircraft).

**TSN (Time Since New)**
Total operating time since aircraft/component was manufactured.

**TSO (Time Since Overhaul)**
Operating time accumulated since last overhaul.

### U

**Unscheduled Maintenance**
Corrective maintenance performed due to unexpected failures or defects.

---

## AMSS System Terms

### A

**Actor**
The user performing an action in AMSS. Includes role and organization context.

**Admin**
System administrator role with god-mode access across all organizations. Can create organizations and manage system-wide settings.

**Auditor**
Compliance officer role focused on regulatory reporting and audit trails. Read-only access to most data.

**Audit Log**
Immutable record of all system actions (create, update, delete, state changes) for compliance tracking.

### C

**Compliance Item**
Regulatory checklist item that must be verified and signed off before task completion.

**Compliance Rate**
Percentage of maintenance tasks completed on time without regulatory violations.

**Compliance Sign-Off**
Digital signature by authorized user confirming regulatory requirement has been met.

### M

**Mechanic**
Technician role responsible for executing maintenance tasks. Can start tasks, reserve parts, complete work.

### O

**Organization (Org)**
Top-level tenant in multi-tenant system. All data (users, aircraft, tasks) belongs to one organization. Organizations cannot see each other's data.

**Org ID**
UUID uniquely identifying an organization. All database queries filtered by org_id for multi-tenant isolation.

### P

**Part Reservation**
Temporary allocation of inventory part to specific maintenance task. Prevents double-booking of parts.

**Program**
Maintenance program defining scheduled inspections/tasks for aircraft (e.g., "100-Hour Inspection Program").

**Program Generator**
Background worker that automatically creates maintenance tasks based on program rules every 6 hours.

### R

**Role**
User permission level in AMSS. Five roles: `admin`, `tenant_admin`, `scheduler`, `mechanic`, `auditor`.

**Row-Level Security (RLS)**
PostgreSQL security feature ensuring users only see data from their organization.

### S

**Scheduler**
Maintenance planner role responsible for creating programs, scheduling tasks, and assigning work to mechanics.

### T

**Task State**
Current lifecycle status of maintenance task:
- `scheduled` - Planned but not started
- `in_progress` - Mechanic actively working
- `completed` - Finished and signed off
- `cancelled` - Abandoned (requires scheduler/admin)

**Task Type**
Category of maintenance work:
- `inspection` - Visual/technical examination
- `repair` - Fix specific defect
- `overhaul` - Complete rebuild

**Tenant Admin**
Fleet manager role with organization-wide visibility. Can view dashboards, configure webhooks, manage integrations.

### W

**Webhook**
HTTP callback to external system when AMSS events occur (e.g., task completed, aircraft grounded).

**Webhook Delivery**
Attempt to send webhook event to external URL. Tracked with status (`pending`, `delivered`, `failed`) and retry logic.

---

## Maintenance Concepts

### Lifecycle States

**Aircraft Status:**
- `operational` - Available for flight operations
- `maintenance` - Scheduled work in progress
- `grounded` - Not airworthy, cannot fly

**Task Lifecycle:**
```
scheduled → in_progress → completed
    ↓
cancelled (from scheduled or in_progress only)
```

**Part Reservation Status:**
- `reserved` - Part allocated to task
- `used` - Part consumed during task
- `returned` - Part reservation released

### Time Tracking

**Flight Hours vs. Cycles:**
- **Flight Hours:** Continuous time measure (e.g., 2,345.6 hours)
- **Cycles:** Discrete count of takeoff/landing events (e.g., 1,250 cycles)
- Both used for maintenance scheduling

**Due Date Calculation:**
- **Calendar-based:** Fixed date (e.g., "Annual Inspection due Dec 31, 2024")
- **Usage-based:** Hours/cycles threshold (e.g., "100-Hour Inspection due at 2,500 hours")
- **Whichever Comes First:** Most programs use both (e.g., "12 months OR 100 hours")

### Compliance Concepts

**Regulatory Chain:**
```
Regulation (14 CFR Part 43)
    ↓
Manufacturer Specification (Cessna Service Manual)
    ↓
Maintenance Program (100-Hour Inspection)
    ↓
Task Execution (with compliance sign-off)
    ↓
Audit Trail (immutable log)
```

**Required Documentation:**
- Task completion notes
- Parts used (serial numbers for critical components)
- Mechanic signature (digital in AMSS)
- Compliance sign-offs for regulatory items
- Return to service authorization

---

## Regulatory Terms

### 14 CFR (Code of Federal Regulations)

**Part 43 - Maintenance, Preventive Maintenance, Rebuilding, and Alteration**
- Defines who can perform maintenance
- Specifies required records
- Sets standards for return to service

**Part 91.417 - Maintenance Records**
- Requires permanent record of:
  - Total time in service
  - Maintenance/inspection status
  - Current inspection status
  - Airworthiness Directives compliance
- Records retained for minimum periods (1 year for routine, lifetime for major)

**Part 91.409 - Inspections**
- Annual inspection (every 12 months) for Part 91 aircraft
- 100-hour inspection for aircraft used for hire
- Progressive inspection programs

### EASA Regulations

**Part-M - Continuing Airworthiness**
- Maintenance program approval
- Airworthiness review certificates
- Record keeping requirements

**Part-145 - Approved Maintenance Organizations**
- Quality system requirements
- Personnel qualifications
- Facility standards

### FAA Forms

**Form 337 - Major Repair and Alteration**
- Required for major modifications
- Filed with FAA Aircraft Registration Branch

**Form 8130-3 - Airworthiness Approval Tag**
- Certifies parts/components meet airworthiness standards

---

# Part II: Common Tasks Cheat Sheet

## System Administrator Tasks

### Create New Organization

**API Endpoint:** `POST /api/v1/organizations`

**Quick Steps:**
1. Navigate to Organizations → Create New
2. Enter organization name
3. System generates org_id automatically
4. Organization created with isolated database access

**Required Permission:** `admin` role only

**Example Request:**
```bash
POST /api/v1/organizations
Authorization: Bearer <admin-token>

{
  "name": "SkyFlight Services LLC"
}
```

**Expected Response:**
```json
{
  "id": "org-uuid-abc123",
  "name": "SkyFlight Services LLC",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### Create User and Assign Role

**API Endpoint:** `POST /api/v1/users`

**Quick Steps:**
1. Navigate to Users → Create User
2. Fill in: name, email, role
3. Select organization (admin can select any; tenant_admin limited to own org)
4. System sends invitation email with temp password

**Role Options:**
- `admin` - System administrator (admin-only creation)
- `tenant_admin` - Fleet manager
- `scheduler` - Maintenance planner
- `mechanic` - Technician
- `auditor` - Compliance officer

**Example Request:**
```bash
POST /api/v1/users
Authorization: Bearer <admin-token>

{
  "name": "Jane Smith",
  "email": "jane.smith@skyflight.com",
  "role": "mechanic",
  "org_id": "org-uuid-abc123"
}
```

---

### Monitor System Health

**API Endpoint:** `GET /api/v1/admin/health`

**Quick Dashboard Checks:**
- API Server uptime (target: 99.9%)
- Database connection pool (alert if >80% used)
- Redis cache latency (target: <5ms)
- Background worker last run time (alert if >7 hours)

**Critical Thresholds:**
- Database disk usage >85% → immediate action
- Query latency p95 >100ms → investigate slow queries
- Webhook failure rate >5% → check external integrations

---

## Fleet Manager Tasks

### View Fleet-Wide Status

**API Endpoint:** `GET /api/v1/reports/fleet-status`

**Quick Dashboard Metrics:**
```
FLEET STATUS OVERVIEW
┌─────────────────────────────────────┐
│ Total Aircraft:           15        │
│ Operational:              12 (80%)  │
│ In Maintenance:            2 (13%)  │
│ Grounded:                  1 (7%)   │
├─────────────────────────────────────┤
│ Overdue Tasks:             3        │
│ Due This Week:             8        │
│ Compliance Rate:          94.2%     │
└─────────────────────────────────────┘
```

**Required Permission:** `tenant_admin` or `admin`

---

### Import Fleet Data via CSV

**API Endpoint:** `POST /api/v1/bulk/aircraft/import`

**CSV Format:**
```csv
registration,type,serial_number,manufacturer,model,year_manufactured,current_hours,current_cycles,status
N12345,Cessna 172S,172S12345,Cessna,172S,2018,245.5,450,operational
```

**Quick Steps:**
1. Prepare CSV file with required columns
2. Navigate to Bulk Operations → Import Aircraft
3. Upload CSV file
4. Review validation report
5. Confirm import (rollback if errors)

**Validation Rules:**
- Registration unique within organization
- Hours/cycles must be positive numbers
- Status must be: operational, maintenance, or grounded

---

### Configure Webhook

**API Endpoint:** `POST /api/v1/webhooks`

**Quick Steps:**
1. Navigate to Integrations → Webhooks → Create
2. Enter target URL (must be HTTPS)
3. Select events to subscribe:
   - `task.completed`
   - `task.state_changed`
   - `aircraft.grounded`
   - `compliance.overdue`
4. Generate webhook secret (for HMAC signature verification)
5. Save configuration

**Example Request:**
```bash
POST /api/v1/webhooks
Authorization: Bearer <tenant-admin-token>

{
  "url": "https://external-system.com/webhooks/amss",
  "events": ["task.completed", "aircraft.grounded"],
  "secret": "whsec_abcdef123456"
}
```

---

## Maintenance Planner Tasks

### Create Maintenance Program

**API Endpoint:** `POST /api/v1/programs`

**Quick Steps:**
1. Navigate to Programs → Create New
2. Enter program name (e.g., "100-Hour Inspection")
3. Define trigger conditions:
   - Hours interval (e.g., 100.0)
   - Cycles interval (e.g., optional)
   - Calendar interval (e.g., optional)
4. Set task duration (hours)
5. Add compliance items if required
6. Save program

**Example Request:**
```bash
POST /api/v1/programs
Authorization: Bearer <scheduler-token>

{
  "name": "100-Hour Inspection",
  "hours_interval": 100.0,
  "duration_hours": 8,
  "description": "FAA Part 91.409(b) required inspection"
}
```

**Auto-Generation:** Program Generator worker creates tasks every 6 hours based on aircraft hours/cycles.

---

### Assign Task to Mechanic

**API Endpoint:** `PATCH /api/v1/tasks/{task_id}`

**Quick Steps:**
1. Navigate to Tasks → Upcoming Tasks
2. Select task in `scheduled` state
3. Click "Assign Mechanic"
4. Choose mechanic from dropdown
5. Mechanic receives notification

**Constraints:**
- Task must be in `scheduled` state
- Mechanic must have `mechanic` role
- Mechanic must belong to same organization

**Example Request:**
```bash
PATCH /api/v1/tasks/task-uuid-5678
Authorization: Bearer <scheduler-token>

{
  "assigned_mechanic_id": "user-uuid-mechanic-42"
}
```

---

### Review Overdue Tasks

**API Endpoint:** `GET /api/v1/tasks?status=overdue`

**Quick Filter Options:**
- Overdue by: 1-7 days, 8-30 days, >30 days
- Aircraft: Filter by registration
- Task type: inspection, repair, overhaul
- Sort by: due date (oldest first)

**Action Items for Overdue Tasks:**
1. **Immediate Action (1-7 days overdue):** Assign mechanic, schedule ASAP
2. **Escalation (8-30 days):** Ground aircraft if safety-critical
3. **Regulatory Violation (>30 days):** Notify compliance officer, investigate root cause

---

## Mechanic/Technician Tasks

### View Assigned Tasks

**API Endpoint:** `GET /api/v1/tasks?assigned_to=me`

**Quick Dashboard View:**
```
MY ASSIGNED TASKS
┌──────────────────────────────────────────────────────┐
│ Today (Jan 15, 2024)                                 │
├──────────────────────────────────────────────────────┤
│ [scheduled]   08:00-16:00   N12345   100-Hr Insp    │
│ [in_progress] 14:00-18:00   N67890   Oil Change     │
├──────────────────────────────────────────────────────┤
│ Upcoming (Next 7 Days)                               │
├──────────────────────────────────────────────────────┤
│ [scheduled]   Jan 17   N54321   Annual Inspection   │
└──────────────────────────────────────────────────────┘
```

**Required Permission:** `mechanic` or `scheduler` role

---

### Start Maintenance Task

**API Endpoint:** `PATCH /api/v1/tasks/{task_id}`

**Quick Steps:**
1. Navigate to My Tasks → Scheduled Tasks
2. Select task assigned to you
3. Click "Start Task" (available 5 min before scheduled start)
4. Verify aircraft status = `grounded`
5. Task transitions: `scheduled` → `in_progress`

**Pre-Conditions:**
- Task in `scheduled` state
- Task assigned to you
- Aircraft is `grounded`
- Current time ≥ start_time - 5 minutes

**Example Request:**
```bash
PATCH /api/v1/tasks/task-uuid-5678
Authorization: Bearer <mechanic-token>

{
  "state": "in_progress"
}
```

---

### Reserve Parts for Task

**API Endpoint:** `POST /api/v1/part-reservations`

**Quick Steps:**
1. From active task view, click "Reserve Parts"
2. Search inventory by part number or description
3. Enter quantity needed
4. Confirm reservation
5. Part status: `available` → `reserved`

**Inventory Check:**
```
SEARCH RESULTS: "oil filter"
┌─────────────────────────────────────────────────────┐
│ Part Number: CH48110-1A                             │
│ Description: Oil Filter - Lycoming O-360            │
│ Available: 8 units                                  │
│ Location: Shelf B-12                                │
│ Cost: $24.50/unit                                   │
└─────────────────────────────────────────────────────┘
```

**Example Request:**
```bash
POST /api/v1/part-reservations
Authorization: Bearer <mechanic-token>

{
  "task_id": "task-uuid-5678",
  "part_id": "part-uuid-ch48110",
  "quantity": 1
}
```

---

### Complete Task

**API Endpoint:** `PATCH /api/v1/tasks/{task_id}`

**Pre-Completion Checklist:**
- [ ] All part reservations used or returned
- [ ] All compliance items signed off
- [ ] Work notes documented
- [ ] Current time ≥ end_time (or early completion approved)

**Quick Steps:**
1. Navigate to active task
2. Verify all compliance sign-offs complete
3. Enter completion notes (required)
4. Click "Complete Task"
5. Task transitions: `in_progress` → `completed`
6. Aircraft status can be changed to `operational`

**Example Request:**
```bash
PATCH /api/v1/tasks/task-uuid-5678
Authorization: Bearer <mechanic-token>

{
  "state": "completed",
  "notes": "100-hour inspection completed. Oil changed (5 qts Aeroshell 15W-50). No discrepancies found. Aircraft returned to service per 14 CFR Part 43."
}
```

---

### Return Unused Parts

**API Endpoint:** `PATCH /api/v1/part-reservations/{reservation_id}`

**Quick Steps:**
1. Navigate to active task → Reserved Parts
2. Select unused reservation
3. Click "Return to Inventory"
4. Reservation status: `reserved` → `returned`
5. Inventory quantity updated

**Example Request:**
```bash
PATCH /api/v1/part-reservations/reservation-uuid-789
Authorization: Bearer <mechanic-token>

{
  "status": "returned"
}
```

---

## Compliance Officer Tasks

### Generate Compliance Report

**API Endpoint:** `GET /api/v1/reports/compliance`

**Quick Steps:**
1. Navigate to Reports → Compliance Report
2. Set date range (e.g., last 12 months for annual audit)
3. Select aircraft (or "All Aircraft")
4. Click "Generate Report"
5. Review compliance rate, overdue items, violations
6. Export as PDF for FAA inspector

**Report Sections:**
- Overall compliance rate (%)
- Tasks by status (completed on time, overdue, cancelled)
- Compliance items by result (pass, fail, pending)
- Audit log summary
- Aircraft-specific breakdown

---

### Review Audit Logs

**API Endpoint:** `GET /api/v1/audit-logs`

**Quick Filters:**
- **Entity Type:** tasks, aircraft, parts, programs, users
- **Action:** create, update, delete, state_change
- **User:** Filter by specific mechanic/scheduler
- **Date Range:** Last 7 days, 30 days, 90 days, custom

**Sample Audit Log Entry:**
```json
{
  "id": "audit-uuid-123",
  "timestamp": "2024-01-15T14:35:22Z",
  "user_id": "user-uuid-mechanic-42",
  "user_name": "John Doe",
  "entity_type": "task",
  "entity_id": "task-uuid-5678",
  "action": "state_change",
  "details": {
    "old_state": "in_progress",
    "new_state": "completed",
    "aircraft": "N12345"
  },
  "ip_address": "192.168.1.50",
  "user_agent": "Mozilla/5.0..."
}
```

---

### Verify Task Completion Signatures

**API Endpoint:** `GET /api/v1/tasks/{task_id}`

**Verification Checklist:**
- [ ] Task in `completed` state
- [ ] Assigned mechanic matches completion user
- [ ] Completion timestamp within task scheduled window
- [ ] All compliance items have sign-offs
- [ ] Notes field populated (required per 14 CFR Part 43)
- [ ] Parts used have serial numbers (for critical components)

**Red Flags:**
- Different user completed task than assigned mechanic
- Completion timestamp outside task window without approval
- Missing compliance sign-offs
- Blank notes field

---

# Part III: Troubleshooting Guide

## Error Code Reference

### HTTP Status Codes

**400 Bad Request**
**Meaning:** Invalid input data (validation error)
**Common Causes:**
- Missing required fields
- Invalid data format (e.g., non-numeric hours)
- End time before start time
- Invalid enum value (e.g., unknown task state)

**Example Error Response:**
```json
{
  "error": "validation: end_time must be after start_time"
}
```

**Recovery:**
1. Review error message for specific validation failure
2. Correct input data
3. Retry request

---

**401 Unauthorized**
**Meaning:** Missing or invalid authentication token
**Common Causes:**
- JWT token expired (tokens valid 24 hours)
- Token not included in Authorization header
- Token revoked (user password changed)

**Example Error Response:**
```json
{
  "error": "unauthorized: invalid or expired token"
}
```

**Recovery:**
1. Log in again to get fresh token
2. Update Authorization header: `Bearer <new-token>`
3. Retry request

---

**403 Forbidden**
**Meaning:** User lacks permission for requested action
**Common Causes:**
- Mechanic trying to create program (scheduler-only)
- Non-admin trying to create organization
- Mechanic trying to start task assigned to different mechanic
- User trying to access different organization's data

**Example Error Response:**
```json
{
  "error": "forbidden: insufficient permissions for this action"
}
```

**Recovery:**
1. Verify your role has required permission (see Role Permission Matrix below)
2. Contact system administrator if role assignment incorrect
3. Request task reassignment if attempting to work on unassigned task

**Role Permission Matrix:**
| Action | admin | tenant_admin | scheduler | mechanic | auditor |
|--------|-------|--------------|-----------|----------|---------|
| Create organization | ✅ | ❌ | ❌ | ❌ | ❌ |
| Create user | ✅ | ✅ (own org) | ❌ | ❌ | ❌ |
| Create program | ✅ | ✅ | ✅ | ❌ | ❌ |
| Assign task | ✅ | ✅ | ✅ | ❌ | ❌ |
| Start task | ✅ | ✅ | ✅ | ✅ (if assigned) | ❌ |
| Complete task | ✅ | ✅ | ✅ | ✅ (if assigned) | ❌ |
| View audit logs | ✅ | ✅ | ✅ | ❌ | ✅ |
| Generate compliance report | ✅ | ✅ | ✅ | ❌ | ✅ |

---

**404 Not Found**
**Meaning:** Requested resource does not exist
**Common Causes:**
- Incorrect UUID in API path
- Resource deleted
- Typo in endpoint URL
- Attempting to access different organization's resource

**Example Error Response:**
```json
{
  "error": "not found: task not found"
}
```

**Recovery:**
1. Verify UUID is correct
2. Check resource wasn't deleted (query audit logs)
3. Verify endpoint URL spelling
4. Confirm resource belongs to your organization

---

**409 Conflict**
**Meaning:** Request conflicts with current system state
**Common Causes:**
- Invalid state transition (e.g., completed → in_progress)
- Aircraft not grounded when starting task
- Starting task before start_time - 5 minutes
- Completing task before end_time without early completion approval
- Part reservations not closed before completing task
- Compliance items not signed off before completing task

**Example Error Response:**
```json
{
  "error": "conflict: aircraft must be grounded"
}
```

**Recovery by Scenario:**

| Error Message | Root Cause | Recovery Steps |
|---------------|------------|----------------|
| `task must be scheduled` | Trying to start task already in progress | Refresh task status; if duplicate request, no action needed |
| `task must be in progress` | Trying to complete task not started | Start task first using state: "in_progress" |
| `aircraft must be grounded` | Aircraft status not set to grounded | Update aircraft status to "grounded" before starting task |
| `too early to start task` | Current time < start_time - 5min | Wait until scheduled start time or contact scheduler to reschedule |
| `task cannot be completed early` | Current time < end_time | Wait until end_time OR contact scheduler for early completion approval |
| `part reservations must be used or released` | Open part reservations exist | Mark reservations as "used" or "returned" |
| `required parts must be used` | Program defines required parts not used | Use required parts or contact scheduler to modify program |
| `compliance items must be signed off` | Missing compliance sign-offs | Complete all compliance checklist items |
| `notes are required` | Completing task without notes | Add completion notes per 14 CFR Part 43 requirements |
| `completed tasks cannot be cancelled` | Trying to cancel finished task | Contact admin if task truly needs reversal (rare, requires audit justification) |

---

**500 Internal Server Error**
**Meaning:** Unexpected server-side error
**Common Causes:**
- Database connection failure
- Bug in server code
- Out of memory condition
- Redis cache unavailable

**Example Error Response:**
```json
{
  "error": "internal server error",
  "request_id": "req-uuid-abc123"
}
```

**Recovery:**
1. Note request_id for troubleshooting
2. Wait 1-2 minutes and retry
3. If persistent, contact system administrator
4. Check system health dashboard (admins only)

---

## Common Issues by Symptom

### "I can't log in"

**Symptom:** Login fails with error
**Diagnostic Questions:**
1. What error message do you see?
2. Are you using correct email address?
3. Did you recently change your password?
4. Is this your first login (invitation email)?

**Solutions by Error:**

| Error Message | Cause | Solution |
|---------------|-------|----------|
| "Invalid credentials" | Wrong password | Use "Forgot Password" link; check email for reset |
| "Account locked" | Too many failed attempts | Contact system administrator to unlock |
| "User not found" | Email typo or not created | Verify email spelling; contact admin to confirm account exists |
| "Email not verified" | New account not activated | Check email for verification link; resend if needed |

---

### "I can't see my tasks"

**Symptom:** Task list is empty when tasks should exist
**Diagnostic Questions:**
1. What is your role? (mechanic sees only assigned tasks)
2. Are tasks scheduled for future dates?
3. What filters are applied?

**Solutions:**

**For Mechanics:**
- Tasks only appear if explicitly assigned to you by scheduler
- Check "All Tasks" view to confirm task exists but assigned to someone else
- Contact scheduler to request assignment

**For Schedulers:**
- Verify organization filter (top-right dropdown)
- Check date range filter (default: next 30 days)
- Confirm tasks were created (check Programs → Auto-Generated Tasks)

**For All Roles:**
- Clear all filters and try again
- Hard refresh browser (Ctrl+Shift+R)
- Log out and log back in

---

### "I can't start my task"

**Symptom:** "Start Task" button disabled or returns error
**Diagnostic Checklist:**

**Pre-Condition Verification:**
```
[ ] Task is in "scheduled" state (not already in_progress)
[ ] Task is assigned to ME (not another mechanic)
[ ] Aircraft status is "grounded" (not operational/maintenance)
[ ] Current time is within 5 minutes of start_time
[ ] I have "mechanic" role permission
```

**Common Failures:**

| Symptom | Cause | Solution |
|---------|-------|----------|
| Button grayed out | Task already started | Refresh page; if task shows in_progress, continue work |
| Error: "task must be scheduled" | Task in wrong state | Check task status; contact scheduler if cancelled/completed incorrectly |
| Error: "aircraft must be grounded" | Aircraft status incorrect | Navigate to aircraft details, change status to "grounded" |
| Error: "too early to start task" | Before start_time - 5min | Wait until scheduled time OR contact scheduler to adjust start time |
| Error: "forbidden" | Task assigned to different mechanic | Contact scheduler to reassign task OR ask assigned mechanic to start |

---

### "I can't complete my task"

**Symptom:** "Complete Task" button disabled or returns error
**Diagnostic Checklist:**

**Completion Requirements:**
```
[ ] Task is in "in_progress" state (not scheduled)
[ ] All part reservations closed (used or returned)
[ ] All compliance items signed off (no "pending" items)
[ ] Completion notes entered (required field)
[ ] Current time >= end_time (or early completion approved)
```

**Step-by-Step Debugging:**

**Step 1: Check Part Reservations**
1. Navigate to task → "Reserved Parts" tab
2. Verify each reservation status:
   - ✅ `used` - Part consumed during task
   - ✅ `returned` - Part returned to inventory
   - ❌ `reserved` - BLOCKER: Mark as used or returned

**Step 2: Check Compliance Items**
1. Navigate to task → "Compliance" tab
2. Verify each item result:
   - ✅ `pass` - Requirement met
   - ❌ `pending` - BLOCKER: Complete sign-off

**Step 3: Check Notes**
1. Scroll to "Completion Notes" field
2. Verify notes are NOT empty
3. Notes should include:
   - Work performed summary
   - Parts used (with serial numbers for critical items)
   - Discrepancies found (if any)
   - Return to service statement

**Step 4: Check Timing**
1. Compare current time to task end_time
2. If before end_time:
   - Contact scheduler for early completion approval
   - Scheduler updates task with `allow_early_completion: true`

---

### "Parts reservation failed"

**Symptom:** Cannot reserve part for task
**Diagnostic Questions:**
1. How many units are available in inventory?
2. Are there existing reservations on this part?
3. Is the part assigned to correct organization?

**Common Failures:**

| Error Message | Cause | Solution |
|---------------|-------|----------|
| "Insufficient quantity" | Requested qty > available | Reduce quantity OR wait for stock replenishment |
| "Part not found" | Wrong part_id or deleted | Search inventory again; verify part number |
| "Task not in progress" | Trying to reserve for scheduled task | Start task first, then reserve parts |
| "Duplicate reservation" | Already reserved this part | Check task → Reserved Parts; cancel if duplicate |

**Inventory Availability Calculation:**
```
Available Quantity = Total Stock - Reserved Quantity - Used Quantity

Example:
Total Stock: 10 units
Reserved: 3 units (other tasks)
Used: 2 units (completed tasks)
Available: 5 units ← Maximum you can reserve
```

---

### "Webhook not delivering"

**Symptom:** External system not receiving webhook events
**Diagnostic Checklist:**

**Webhook Configuration:**
```
[ ] URL is HTTPS (HTTP not allowed)
[ ] URL is publicly accessible (no localhost)
[ ] Event types subscribed to correct events
[ ] Secret configured for HMAC signature
[ ] External endpoint returns 2xx status code
```

**Step 1: Check Webhook Delivery Status**
1. Navigate to Integrations → Webhooks
2. Select webhook configuration
3. View "Recent Deliveries" tab
4. Check delivery status:

| Status | Meaning | Next Steps |
|--------|---------|------------|
| `delivered` | Success (2xx response) | No action needed |
| `failed` | Error response (4xx/5xx) | Check last_response_code and last_response_body |
| `pending` | Retry in progress | Wait for retry (exponential backoff: 1min, 5min, 15min, 1hr, 6hr) |

**Step 2: Verify Endpoint Health**
Test webhook endpoint manually:
```bash
curl -X POST https://external-system.com/webhooks/amss \
  -H "Content-Type: application/json" \
  -H "X-AMSS-Signature: test" \
  -d '{"event":"test","timestamp":"2024-01-15T10:00:00Z"}'
```

Expected response: HTTP 200-299

**Step 3: Check HMAC Signature Verification**
External system must verify signature:

```javascript
// Example: Node.js signature verification
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  const digest = 'sha256=' + hmac.update(payload).digest('hex');
  return crypto.timingSafeEqual(
    Buffer.from(signature),
    Buffer.from(digest)
  );
}

// In webhook handler:
const signature = req.headers['x-amss-signature'];
const isValid = verifyWebhook(req.body, signature, 'whsec_abcdef123456');
if (!isValid) {
  return res.status(401).send('Invalid signature');
}
```

**Step 4: Review Retry Logic**
AMSS retries failed deliveries:
- Attempt 1: Immediate
- Attempt 2: +1 minute
- Attempt 3: +5 minutes
- Attempt 4: +15 minutes
- Attempt 5: +1 hour
- Attempt 6: +6 hours
- After 6 failures: Marked as permanently failed

**Recovery:** Fix endpoint issues and use "Retry Failed Deliveries" button in webhook settings.

---

## Access & Permission Issues

### "Forbidden: insufficient permissions"

**Symptom:** Action blocked with 403 error
**Resolution Workflow:**

**Step 1: Verify Current Role**
1. Navigate to profile (top-right icon)
2. Check "Role" field
3. Compare to required role for action (see Role Permission Matrix above)

**Step 2: Identify Required Permission**
Common permission mismatches:

| Action | Required Role | Common Wrong Role |
|--------|---------------|-------------------|
| Create maintenance program | scheduler, tenant_admin, admin | mechanic |
| Assign task to mechanic | scheduler, tenant_admin, admin | mechanic |
| Start task | mechanic (if assigned), scheduler, admin | mechanic (not assigned) |
| View audit logs | auditor, tenant_admin, admin | mechanic, scheduler |
| Create organization | admin | tenant_admin |
| Configure webhooks | tenant_admin, admin | scheduler |

**Step 3: Request Role Change (if needed)**
1. Contact system administrator
2. Provide justification for role change
3. Admin updates role via Users → Edit User

**Step 4: Workaround for Task Assignment**
If mechanic trying to start task assigned to colleague:
1. Contact scheduler to reassign task
2. OR ask assigned mechanic to start task, then you assist

---

### "Cannot access different organization's data"

**Symptom:** Getting 404 errors when resource exists
**Root Cause:** Multi-tenant isolation prevents cross-org data access

**Example Scenario:**
- User belongs to "SkyFlight Services" (org-uuid-123)
- Trying to access task from "AeroMaintain LLC" (org-uuid-456)
- AMSS returns 404 (not 403) to prevent information leakage

**Resolution:**
1. Verify you're querying correct organization's data
2. Check organization filter in top-right dropdown
3. If you legitimately need access:
   - Contact admin to create second user account in other org
   - OR admin can reassign your user to different org (loses access to current org)

**For Admins Only:**
- Admin role can switch organization context
- Use organization selector in admin dashboard
- Be cautious: changes affect selected org only

---

## Data Integrity Issues

### "Aircraft hours/cycles decreasing"

**Symptom:** Aircraft counters show lower values than before
**Root Cause:** Manual counter update with incorrect value

**Investigation:**
1. Navigate to aircraft details
2. Check "History" tab for recent updates
3. Review audit log for counter changes:
```json
{
  "action": "update",
  "entity_type": "aircraft",
  "details": {
    "old_hours": 2450.5,
    "new_hours": 2350.0  ← Decreased by 100.5
  },
  "user_id": "user-uuid-planner-12"
}
```

**Recovery:**
1. Contact user who made update
2. Verify correct current hours/cycles from aircraft Hobbs meter or tech log
3. Update aircraft counters with correct values
4. Document correction in notes

**Prevention:**
- Always double-check values before submitting
- Use decimal notation (e.g., 2450.5, not 2450)
- Reference aircraft tech log as source of truth

---

### "Task auto-generated with wrong due date"

**Symptom:** Program Generator created task with unexpected due date
**Root Cause:** Program interval misconfiguration or aircraft counter issue

**Investigation Checklist:**

**Step 1: Verify Program Configuration**
1. Navigate to Programs → Select program
2. Check intervals:
   - `hours_interval` (e.g., 100.0)
   - `cycles_interval` (e.g., optional)
   - `calendar_interval_days` (e.g., 365 for annual)

**Step 2: Check Aircraft Counters**
1. Navigate to aircraft details
2. Verify current counters:
   - `current_hours` (e.g., 2450.5)
   - `current_cycles` (e.g., 3200)

**Step 3: Calculate Expected Due Date**
```
Example: 100-Hour Inspection Program
- Program: hours_interval = 100.0
- Aircraft current_hours = 2450.5
- Last task completed at: 2350.5 hours

Expected next task:
- Due at: 2350.5 + 100.0 = 2450.5 hours (DUE NOW)
- If current_hours = 2460.0 → Task would be OVERDUE by 9.5 hours
```

**Step 4: Verify Last Completed Task**
1. Navigate to Tasks → Filter by aircraft + program
2. Find most recent completed task
3. Check completion timestamp and aircraft hours at completion

**Common Issues:**

| Symptom | Cause | Solution |
|---------|-------|----------|
| Task due immediately after program creation | Program interval = aircraft current hours | Adjust program first_due_hours to future value |
| Task not generated at all | Aircraft hours haven't reached next interval | Update aircraft hours or wait for more flight time |
| Task due date far in future | Program interval too large (e.g., 10,000 hours instead of 100) | Edit program to correct interval |
| Multiple tasks generated | Program misconfigured with very small interval (e.g., 1 hour) | Delete duplicate tasks, fix program interval |

---

### "Compliance rate calculation seems wrong"

**Symptom:** Dashboard shows compliance rate that doesn't match expectations
**Calculation Formula:**
```
Compliance Rate = (Completed On Time) / (Total Scheduled Tasks) × 100%

Completed On Time = Tasks completed where:
  - completion_timestamp <= task.end_time
  - OR early_completion_approved = true

Excludes:
  - Cancelled tasks
  - Tasks still in scheduled/in_progress state
```

**Example Calculation:**
```
Organization: SkyFlight Services
Date Range: Last 12 months (Jan 1 - Dec 31, 2024)

Total Scheduled Tasks: 150
Completed On Time: 141
Completed Late: 5
Cancelled: 3
Still In Progress: 1

Compliance Rate = 141 / (141 + 5) × 100% = 96.6%

Note: In-progress and cancelled tasks excluded from denominator
```

**Verification Steps:**
1. Navigate to Reports → Compliance Report
2. Set same date range as dashboard
3. Export detailed task list
4. Manually count tasks by category:
   - On time: completion_timestamp <= end_time
   - Late: completion_timestamp > end_time AND !early_completion_approved
5. Compare to dashboard calculation

**Common Discrepancies:**

| Issue | Cause | Resolution |
|-------|-------|-----------|
| Rate higher than manual count | Cancelled tasks excluded | Verify cancelled tasks had valid reason (not gaming metrics) |
| Rate lower than expected | Early completions without approval flag | Update tasks with early_completion_approved if justified |
| Rate changes when date range changes | Different tasks in scope | Compliance rate varies by period; focus on trend, not absolute value |

---

# Part IV: FAQs

## System Administrator FAQs

### Q: How do I create a new organization?

**A:** Only users with `admin` role can create organizations.

**API Call:**
```bash
POST /api/v1/organizations
Authorization: Bearer <admin-token>

{
  "name": "NewAirline LLC"
}
```

**Web UI:**
1. Navigate to Organizations → Create New
2. Enter organization name
3. System auto-generates org_id and sets up isolated database access

**Post-Creation:**
- Create at least one `tenant_admin` user for the organization
- Tenant admin can then create additional users (schedulers, mechanics, auditors)

---

### Q: How do I reset a user's password?

**A:** Admins and tenant_admins can reset passwords for users in their organization.

**API Call:**
```bash
POST /api/v1/users/{user_id}/reset-password
Authorization: Bearer <admin-token>
```

**Web UI:**
1. Navigate to Users → Select user
2. Click "Reset Password"
3. System sends email with temporary password
4. User must change password on next login

**Security Note:** Resetting password invalidates all existing JWT tokens for that user.

---

### Q: How do I monitor system health?

**A:** Admin-only dashboard shows real-time system metrics.

**Key Metrics to Monitor:**

**API Server:**
- Uptime (target: 99.9%)
- Average response time (target: <50ms)
- Request rate (requests/second)
- Error rate (target: <0.1%)

**Database:**
- Connection pool usage (alert if >80%)
- Query latency p50/p95 (target: <20ms / <100ms)
- Disk usage (alert if >85%)
- Slow queries (alert if >1 second)

**Redis Cache:**
- Hit rate (target: >90%)
- Latency (target: <5ms)
- Memory usage (alert if >75%)

**Background Workers:**
- Program Generator: Last run timestamp (alert if >7 hours)
- Webhook Dispatcher: Success rate (alert if <95%)

**Access:** `GET /api/v1/admin/health`

---

### Q: How do I backup and restore data?

**A:** AMSS uses PostgreSQL for all persistent data.

**Backup Strategy:**

**Automated Daily Backups:**
```bash
# Cron job (runs at 2 AM daily)
0 2 * * * pg_dump -U amss -d amss | gzip > /backups/amss-$(date +\%Y\%m\%d).sql.gz
```

**Manual Backup:**
```bash
pg_dump -U amss -d amss -F c -f /backups/amss-manual-$(date +\%Y\%m\%d).backup
```

**Restore from Backup:**
```bash
# Stop API server first to prevent writes during restore
sudo systemctl stop amss-api

# Restore database
pg_restore -U amss -d amss -c /backups/amss-20240115.backup

# Restart API server
sudo systemctl start amss-api
```

**Testing Backups:**
- Restore to test database monthly
- Verify data integrity (record counts, audit log continuity)
- Test user login and basic workflows

**Retention Policy:**
- Daily backups: Keep 30 days
- Weekly backups: Keep 12 weeks
- Monthly backups: Keep 12 months
- Annual backups: Keep 7 years (FAA Part 91.417 compliance)

---

### Q: What do I do if the Program Generator stops working?

**A:** Program Generator is critical for auto-creating maintenance tasks.

**Symptoms:**
- No new tasks created for >7 hours
- Background worker status shows "Last run: 8 hours ago"
- System health dashboard shows worker alert

**Diagnostic Steps:**

**Step 1: Check Worker Process**
```bash
# View worker logs
sudo journalctl -u amss-worker -n 100

# Look for errors like:
# - Database connection failed
# - Redis connection failed
# - Panic / crash
```

**Step 2: Check Worker Status**
```bash
# Verify worker is running
sudo systemctl status amss-worker

# If stopped, restart:
sudo systemctl start amss-worker
```

**Step 3: Manual Trigger (Emergency)**
If worker won't start, trigger program generation via API:
```bash
POST /api/v1/admin/trigger-program-generator
Authorization: Bearer <admin-token>
```

**Step 4: Review Program Configurations**
If worker runs but no tasks generated:
- Check if programs exist: `GET /api/v1/programs`
- Verify programs linked to aircraft: `GET /api/v1/programs/{id}/assignments`
- Confirm aircraft hours/cycles updated recently

**Prevention:**
- Monitor worker last_run_timestamp
- Set up alerting for >7 hour delays
- Review worker logs daily for warnings

---

## Fleet Manager FAQs

### Q: How do I import a fleet of aircraft via CSV?

**A:** Bulk import saves time for large fleets.

**CSV Format:**
```csv
registration,type,serial_number,manufacturer,model,year_manufactured,current_hours,current_cycles,status
N12345,Cessna 172S,172S12345,Cessna,172S,2018,245.5,450,operational
N67890,Piper PA-28,PA28-12346,Piper,PA-28,2019,312.8,520,operational
```

**Field Requirements:**
- `registration` - Unique tail number (required)
- `type` - Aircraft type designation (required)
- `serial_number` - Manufacturer serial number (required)
- `manufacturer` - Aircraft manufacturer (required)
- `model` - Aircraft model (required)
- `year_manufactured` - Year built, 4 digits (required)
- `current_hours` - Decimal, ≥0 (required)
- `current_cycles` - Integer, ≥0 (required)
- `status` - operational, maintenance, or grounded (required)

**Import Process:**
1. Prepare CSV file
2. Navigate to Bulk Operations → Import Aircraft
3. Upload file
4. Review validation report (errors highlighted)
5. Fix errors if any, re-upload
6. Confirm import (creates aircraft records)

**Validation Rules:**
- Registration must be unique within organization
- Hours and cycles must be non-negative
- Status must match enum values exactly (case-sensitive)
- All required fields must be present

**Rollback:** If errors discovered after import, contact admin for database rollback within 24 hours.

---

### Q: What events can I subscribe to via webhooks?

**A:** AMSS sends real-time notifications for key events.

**Available Events:**

**Task Events:**
- `task.created` - New task auto-generated or manually created
- `task.state_changed` - Task transitioned (scheduled → in_progress → completed)
- `task.completed` - Task finished (subset of state_changed)
- `task.assigned` - Mechanic assigned to task
- `task.overdue` - Task passed due date without completion

**Aircraft Events:**
- `aircraft.grounded` - Aircraft status changed to grounded
- `aircraft.returned_to_service` - Aircraft status changed to operational
- `aircraft.hours_updated` - Flight hours/cycles updated
- `aircraft.maintenance_due` - Approaching maintenance threshold

**Parts Events:**
- `part.reserved` - Part allocated to task
- `part.used` - Part consumed during task
- `part.low_stock` - Inventory below minimum threshold

**Compliance Events:**
- `compliance.overdue` - Compliance item past due
- `compliance.signed_off` - Compliance item completed
- `compliance.violation` - Task completed late or without required sign-offs

**Webhook Configuration:**
```bash
POST /api/v1/webhooks
Authorization: Bearer <tenant-admin-token>

{
  "url": "https://external-system.com/webhooks/amss",
  "events": [
    "task.completed",
    "aircraft.grounded",
    "compliance.overdue"
  ],
  "secret": "whsec_abcdef123456"
}
```

---

### Q: How do I calculate total maintenance costs?

**A:** Fleet Manager dashboard shows cost breakdown.

**Cost Components:**

**Parts Cost:**
```
Total Parts Cost = Σ (quantity_used × unit_price)

Example:
- Oil Filter × 1 @ $24.50 = $24.50
- Engine Oil × 5 qt @ $12.00/qt = $60.00
- Spark Plug × 4 @ $18.50 = $74.00
Total Parts: $158.50
```

**Labor Cost:**
```
Total Labor Cost = task_duration_hours × hourly_labor_rate

Example:
- Task Duration: 8 hours
- Labor Rate: $85/hour
Total Labor: $680.00
```

**Overhead:**
```
Overhead = (Parts + Labor) × overhead_rate

Example:
- Parts + Labor: $838.50
- Overhead Rate: 15%
Total Overhead: $125.78
```

**Total Task Cost:**
```
Total = Parts + Labor + Overhead
Total = $158.50 + $680.00 + $125.78 = $964.28
```

**Fleet-Wide Report:**
Navigate to Reports → Maintenance Costs
- Filter by date range (e.g., last 12 months)
- Group by: Aircraft, Task Type, Month
- Export to CSV for accounting integration

---

## Maintenance Planner FAQs

### Q: How often does the Program Generator run?

**A:** Every 6 hours automatically.

**Schedule:**
- 00:00 UTC
- 06:00 UTC
- 12:00 UTC
- 18:00 UTC

**What It Does:**
1. Queries all active maintenance programs
2. Checks aircraft current hours/cycles
3. Calculates next due date for each program
4. Creates new tasks if:
   - Due date within next 30 days
   - No existing task for this program/aircraft combination
5. Updates task due dates if aircraft counters changed

**Manual Trigger:**
Admins can trigger manually via: `POST /api/v1/admin/trigger-program-generator`

**Why 6 hours?**
- Balances freshness with server load
- Most programs due on weeks/months timescale
- Aircraft hours updated daily (not real-time)

---

### Q: What's the difference between calendar-based and usage-based programs?

**A:** Different trigger mechanisms for different maintenance types.

**Calendar-Based Programs:**
- Triggered by elapsed time (days/months)
- Examples: Annual Inspection (365 days), 30-Day Check
- Use `calendar_interval_days` field
- Due date calculated from last completion date

```json
{
  "name": "Annual Inspection",
  "calendar_interval_days": 365,
  "hours_interval": null,
  "cycles_interval": null
}
```

**Usage-Based Programs:**
- Triggered by flight hours or cycles
- Examples: 100-Hour Inspection, 500-Cycle Landing Gear Inspection
- Use `hours_interval` or `cycles_interval` fields
- Due date calculated from aircraft counters

```json
{
  "name": "100-Hour Inspection",
  "hours_interval": 100.0,
  "calendar_interval_days": null
}
```

**Whichever Comes First (Most Common):**
- Uses BOTH calendar and usage intervals
- Task due when EITHER threshold reached
- Examples: "12 months OR 100 hours, whichever comes first"

```json
{
  "name": "Progressive Inspection",
  "calendar_interval_days": 365,
  "hours_interval": 100.0
}
```

**Program Generator Logic:**
```
For whichever-comes-first programs:
  calendar_due_date = last_completion + calendar_interval_days
  usage_due_hours = last_completion_hours + hours_interval

  actual_due_date = MIN(calendar_due_date, projected_date_for_usage_due_hours)
```

---

### Q: Can I defer maintenance to a later date?

**A:** Yes, with proper authorization and documentation.

**Scenarios:**

**Scenario 1: Non-Critical Maintenance**
- Deferable under MEL (Minimum Equipment List)
- Example: Inoperative cabin reading light
- Process:
  1. Verify MEL allows deferral
  2. Update task notes with MEL reference
  3. Reschedule task within MEL-allowed timeframe
  4. Document in aircraft tech log

**Scenario 2: Scheduled Inspection Postponement**
- Example: 100-Hour Inspection due at 2450 hours, but aircraft needed for urgent charter
- FAA allows 10% overage (up to 2500 hours)
- Process:
  1. Contact compliance officer for approval
  2. Document business justification
  3. Update task end_time
  4. Must complete before hard limit (2500 hours)

**Scenario 3: Calendar-Based Deferral**
- Example: Annual Inspection due Dec 31, but hangar space unavailable
- FAA allows grace period (typically 30 days) with FSDO approval
- Process:
  1. Contact local FSDO (Flight Standards District Office)
  2. Request extension with justification
  3. Document approval in AMSS task notes
  4. Update task due date
  5. Aircraft may be grounded until completed

**What You CANNOT Defer:**
- Airworthiness Directives (ADs) - mandatory, no exceptions
- Safety-critical items
- Items preventing return to service

**AMSS Workflow:**
```bash
PATCH /api/v1/tasks/{task_id}
Authorization: Bearer <scheduler-token>

{
  "end_time": "2024-12-31T23:59:59Z",  # New due date
  "notes": "Deferred per MEL 32-41-01. Max extension: 10 flight days. FSDO approval ref: ABC123."
}
```

---

### Q: How do I assign tasks to mechanics fairly?

**A:** Consider workload balancing and skill matching.

**Workload Balancing Metrics:**

**Option 1: Task Count**
```
Navigate to Reports → Mechanic Workload
View by: Next 7 Days

MECHANIC WORKLOAD (Jan 15-22, 2024)
┌─────────────────────────────────────────┐
│ John Doe        │ 5 tasks │ 32 hours   │
│ Jane Smith      │ 3 tasks │ 24 hours   │
│ Bob Johnson     │ 7 tasks │ 40 hours   │ ⚠️ Overloaded
│ Alice Williams  │ 2 tasks │ 16 hours   │ ✅ Available
└─────────────────────────────────────────┘
```

**Option 2: Task Hours**
- Sum of estimated duration for assigned tasks
- More accurate for complex inspections vs simple checks
- Target: 40 hours/week per mechanic

**Option 3: Skill Matching**
- Match task requirements to mechanic certifications
- Examples:
  - Avionics work → Mechanic with avionics certification
  - Engine overhaul → A&P with powerplant rating
  - Structural repair → Mechanic with sheet metal experience

**Assignment Workflow:**
1. View upcoming tasks (next 30 days)
2. Check mechanic workload dashboard
3. Identify under-utilized mechanic or skill match
4. Select task → Assign Mechanic → Choose from dropdown
5. Mechanic receives notification

**Fair Distribution Tips:**
- Rotate desirable vs undesirable tasks
- Consider mechanic preferences (some prefer engines, others avionics)
- Account for training opportunities (pair junior with senior on complex tasks)
- Monitor overtime (flag if exceeding 40 hours/week consistently)

---

## Mechanic/Technician FAQs

### Q: What do I do if I need a part that's not in inventory?

**A:** AMSS supports emergency part orders.

**Workflow:**

**Step 1: Verify Part Not Available**
1. Search inventory by part number
2. Check "Available" quantity (not just "Total Stock")
3. Verify no alternate/equivalent parts

**Step 2: Emergency Order**
1. From task view → "Request Part Order"
2. Enter part number, description, quantity
3. Add vendor (if known) and urgency level:
   - Standard (3-5 business days)
   - Expedited (next day)
   - AOG (Aircraft On Ground - same day)

**Step 3: Notify Scheduler**
- Task remains in `in_progress` state
- Scheduler receives notification of part order request
- Scheduler contacts vendor for order

**Step 4: Temporary Work**
While waiting for part:
- Complete non-blocked portions of task
- Document part delay in task notes
- Update expected completion time
- Consider MEL deferral if applicable

**Step 5: Part Arrival**
- Scheduler updates inventory with received part
- Reserve part for your task
- Complete remaining work

**Example Scenario:**
```
Task: 100-Hour Inspection on N12345
Issue: Oil filter out of stock

Notes:
"100-hour inspection 80% complete. Awaiting oil filter CH48110-1A
(ordered AOG from Aircraft Spruce, ETA 4 PM today). All other
inspections complete, no discrepancies found. Will install filter
and complete task upon delivery."
```

---

### Q: Can I start a task early if I have time?

**A:** Yes, within limits.

**Rules:**

**Scheduled Start Time:**
- Task has `start_time` (e.g., Jan 15, 08:00)
- You can start up to 5 minutes early (07:55)
- Earlier than that: Contact scheduler to adjust start_time

**Why 5-Minute Buffer?**
- Accounts for clock synchronization
- Allows setup time before official start
- Prevents gaming of time tracking

**Starting Significantly Early:**
If you want to start hours/days early:

**Scenario 1: Task scheduled for tomorrow, but you're free today**
- Contact scheduler: "I have availability today to start N12345 inspection early"
- Scheduler updates start_time and end_time
- You can then start task

**Scenario 2: Aircraft arrived early for maintenance**
- Common for transient aircraft
- Scheduler adjusts task schedule
- Verify aircraft is `grounded` status

**Cannot Start Early If:**
- Aircraft still in flight operations (status = operational)
- Required parts not yet arrived
- Hangar space not available
- Task assigned to different mechanic (contact scheduler for reassignment)

---

### Q: What if I can't finish a task by the scheduled end time?

**A:** Communicate early and document reasons.

**Step 1: Notify Scheduler ASAP**
Don't wait until end_time to report delays. Notify as soon as you identify blocker:
- "Discovered corrosion requiring additional repair (estimated +4 hours)"
- "Waiting for emergency part order (delayed until tomorrow)"
- "Task more complex than estimated (need +2 hours)"

**Step 2: Scheduler Adjusts End Time**
```bash
PATCH /api/v1/tasks/{task_id}
Authorization: Bearer <scheduler-token>

{
  "end_time": "2024-01-16T16:00:00Z",  # Extended by 1 day
  "notes": "Extended due to corrosion repair on wing spar. Additional 6 hours approved."
}
```

**Step 3: Document in Task Notes**
Add detailed notes explaining delay:
- What was discovered
- Additional work required
- Parts/materials needed
- Revised estimated completion

**Step 4: Complete Task**
Once work finished, complete normally. System allows completion after original end_time if scheduler adjusted.

**Example Notes:**
```
"100-hour inspection revealed corrosion on right main landing gear
strut (photo attached). Corrosion exceeds MEL limits, requires
repair per Cessna Service Manual Section 32-11-01. Ordered
replacement part (P/N 2413100-4, ETA tomorrow 10 AM). Task
extended by scheduler, will complete 1/16/24."
```

**Impact on Compliance:**
- Task completion timestamp compared to ADJUSTED end_time (not original)
- No compliance violation if scheduler approved extension
- Audit trail shows both original and adjusted timestamps

---

### Q: Do I need to sign off compliance items manually?

**A:** Yes, digital signature required for regulatory compliance.

**What Are Compliance Items?**
Regulatory checklist items mandated by:
- FAA regulations (14 CFR Part 43, 91, etc.)
- EASA regulations (Part-M, Part-145)
- Manufacturer service bulletins
- Airworthiness Directives (ADs)

**Examples:**
- "Verify AD 2023-05-12 compliance (wing spar inspection)"
- "Check fuel system for leaks per Part 43 Appendix D"
- "Torque propeller bolts to 350 in-lbs per maintenance manual"

**Sign-Off Process:**

**Step 1: Review Compliance Item**
- Read description and regulatory reference
- Perform required inspection/measurement
- Verify result meets specifications

**Step 2: Record Result**
- Pass: Item meets requirements
- Fail: Item does not meet requirements (blocks task completion)
- N/A: Not applicable (e.g., AD doesn't apply to this serial number)

**Step 3: Digital Signature**
- Click "Sign Off" button
- Enter password to confirm identity
- System records:
  - Your user ID (mechanic)
  - Timestamp
  - Result (pass/fail)
  - Any notes you added

**Step 4: Task Completion Blocked Until All Items Signed**
You cannot complete task if any compliance items still `pending`.

**Regulatory Significance:**
- Your digital signature = legal signature per 14 CFR Part 43.9
- Equivalent to handwritten signature in paper logbooks
- Admissible for FAA inspections
- Audit trail immutable (cannot change after sign-off)

**Example Compliance Item:**
```
Description: Verify AD 2023-12-05 compliance - inspect wing attach bolts for cracks
Regulatory Reference: FAA AD 2023-12-05
Method: Visual inspection with 10x magnification per AD instructions
Result: ✅ PASS - No cracks detected
Signed By: John Doe (Mechanic #1234567)
Signed At: 2024-01-15 14:35:22 UTC
Notes: All 8 wing attach bolts inspected. No cracks, corrosion, or deformation observed.
```

---

## Compliance Officer FAQs

### Q: How long are audit logs retained?

**A:** Permanent retention for regulatory compliance.

**Retention Policy:**

**Audit Logs:**
- Retained indefinitely (no automatic deletion)
- Stored in PostgreSQL with daily backups
- Immutable (cannot be edited or deleted)

**Why Permanent?**
- FAA Part 91.417 requires maintenance records retention:
  - Major repairs: Lifetime of aircraft
  - Routine maintenance: 1 year after superseded
  - Annual inspections: 1 year
- AMSS stores ALL logs to ensure compliance
- Audit logs provide chain of custody for all actions

**What's Logged:**
Every action creating, updating, or deleting data:
- User ID (who performed action)
- Timestamp (when)
- Entity type and ID (what was affected)
- Action type (create, update, delete, state_change)
- Details (old vs new values for updates)
- IP address and user agent (where/how)
- Request ID (for correlating related actions)

**Access:**
- Auditors: Full read access to audit logs
- Admins: Full read access
- Other roles: No access (unless granted by admin)

**Example Log Entry:**
```json
{
  "id": "audit-uuid-123456",
  "timestamp": "2024-01-15T14:35:22Z",
  "user_id": "user-uuid-mechanic-42",
  "user_name": "John Doe",
  "user_role": "mechanic",
  "entity_type": "task",
  "entity_id": "task-uuid-5678",
  "action": "state_change",
  "old_state": "in_progress",
  "new_state": "completed",
  "details": {
    "aircraft": "N12345",
    "program": "100-Hour Inspection",
    "completion_notes": "Inspection complete, no discrepancies..."
  },
  "ip_address": "192.168.1.50",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) ...",
  "request_id": "req-uuid-abc789"
}
```

**Exporting Audit Logs:**
For FAA inspections or legal discovery:
```bash
GET /api/v1/audit-logs?start_date=2023-01-01&end_date=2023-12-31
Authorization: Bearer <auditor-token>
```
Export as CSV or JSON for external analysis.

---

### Q: What's included in a compliance report?

**A:** Comprehensive regulatory documentation for FAA/EASA inspections.

**Report Sections:**

**1. Executive Summary**
```
COMPLIANCE REPORT
Organization: SkyFlight Services LLC
Report Period: Jan 1, 2024 - Dec 31, 2024
Generated: Jan 15, 2025
Generated By: Jane Smith (Auditor)

OVERALL COMPLIANCE RATE: 96.8%
Total Scheduled Tasks: 487
Completed On Time: 471
Completed Late: 12
Cancelled: 4
Currently Overdue: 3 ⚠️
```

**2. Aircraft Breakdown**
```
BY AIRCRAFT
┌────────────────────────────────────────────────────┐
│ N12345 │ Cessna 172S  │ 98.2% │ 54/55 on time     │
│ N67890 │ Piper PA-28  │ 100%  │ 38/38 on time     │
│ N54321 │ Beech A36    │ 91.4% │ 32/35 on time     │
│ N11111 │ Cessna 182T  │ 94.7% │ 36/38 on time     │
└────────────────────────────────────────────────────┘
```

**3. Task Type Breakdown**
```
BY TASK TYPE
┌────────────────────────────────────────────────────┐
│ Inspection   │ 298 tasks │ 97.3% on time         │
│ Repair       │ 142 tasks │ 95.8% on time         │
│ Overhaul     │  47 tasks │ 97.9% on time         │
└────────────────────────────────────────────────────┘
```

**4. Overdue Tasks Detail**
```
OVERDUE TASKS (3)
┌────────────────────────────────────────────────────┐
│ N54321 │ Annual Inspection    │ Due: Dec 31, 2024 │
│        │ Status: Scheduled    │ Overdue: 15 days  │
│        │ Action: Grounded until completion        │
├────────────────────────────────────────────────────┤
│ N12345 │ 100-Hour Inspection  │ Due: Jan 10, 2025 │
│        │ Status: In Progress  │ Overdue: 5 days   │
│        │ Action: Completion expected Jan 17       │
└────────────────────────────────────────────────────┘
```

**5. Compliance Items Summary**
```
COMPLIANCE ITEMS
Total Items: 1,247
Passed: 1,235 (99.0%)
Failed: 8 (0.6%)
Pending: 4 (0.3%)

Failed Items Detail:
- AD 2023-05-12 wing spar inspection: 3 aircraft required corrective action
- Fuel leak test Part 43 Appendix D: 2 aircraft required seal replacement
[All failures corrected and re-inspected per audit trail]
```

**6. Regulatory Violations**
```
VIOLATIONS (Late Completions Without Approval)
┌────────────────────────────────────────────────────┐
│ N54321 │ Progressive Inspection │ Due: Mar 15, 2024│
│        │ Completed: Mar 18, 2024 │ Late: 3 days     │
│        │ Root Cause: Parts delay │ Corrective: AOG  │
│        │                         │ parts process    │
└────────────────────────────────────────────────────┘
```

**7. Audit Log Summary**
```
AUDIT TRAIL STATISTICS
Total Log Entries: 8,456
- Task state changes: 1,461
- Part reservations: 2,134
- Compliance sign-offs: 1,247
- Aircraft updates: 487
- User actions: 3,127

All actions traceable to individual users with timestamps.
No log integrity issues detected.
```

**Export Formats:**
- PDF: For printing, FAA inspector review
- CSV: For spreadsheet analysis
- JSON: For integration with external compliance systems

---

### Q: How do I investigate a compliance violation?

**A:** Systematic audit trail analysis.

**Scenario:** Task completed 5 days late without approval extension.

**Investigation Workflow:**

**Step 1: Identify Violation**
```bash
GET /api/v1/reports/compliance?violations_only=true
```

Returns tasks where:
- `completion_timestamp > end_time`
- `early_completion_approved = false`

**Step 2: Gather Task Details**
```bash
GET /api/v1/tasks/{task_id}
```

Key information:
- Aircraft registration
- Task type and program
- Scheduled start/end times
- Actual start/completion times
- Assigned mechanic
- Completion notes

**Step 3: Review Audit Trail**
```bash
GET /api/v1/audit-logs?entity_type=task&entity_id={task_id}
```

Chronological sequence:
1. Task created (auto-generated or manual)
2. Task assigned to mechanic
3. Task started (state: scheduled → in_progress)
4. Part reservations created
5. Compliance items signed off
6. Task completed (state: in_progress → completed)

**Step 4: Interview Involved Parties**
- **Mechanic:** Why late completion?
- **Scheduler:** Was extension requested/denied?
- **Parts Department:** Were required parts available?

**Step 5: Root Cause Analysis**
Common causes:
- Parts delay (vendor issue)
- Unexpected additional work discovered (corrosion, damage)
- Mechanic unavailability (sick leave, overload)
- Scheduler error (unrealistic time estimate)
- Communication breakdown (extension requested but not processed)

**Step 6: Corrective Action**
- **Immediate:** Document findings in task notes
- **Short-term:** If justified (e.g., safety issue found), retroactively approve extension
- **Long-term:** Process improvements
  - Parts: Improve inventory management, AOG procedures
  - Scheduling: Adjust time estimates for complex tasks
  - Communication: Implement automated alerts for approaching deadlines

**Step 7: Document for Regulator**
Prepare written summary for FAA/EASA inspection:
```
VIOLATION INVESTIGATION REPORT

Task: 100-Hour Inspection, Aircraft N12345
Scheduled: Jan 10-11, 2024 (16 hours)
Completed: Jan 16, 2024
Late By: 5 days

Root Cause:
During inspection, mechanic discovered wing spar corrosion exceeding
MEL limits. Required emergency part order (P/N 2413100-4). Part
delayed due to vendor backorder. Aircraft remained grounded throughout.

Corrective Action:
- Part installed Jan 16, corrosion repaired per Service Manual 57-20-01
- Return to service inspection completed, no further issues
- Implemented AOG parts agreement with alternate vendor
- Updated inventory minimum stock levels for critical parts

Regulatory Compliance:
- Aircraft not operated during maintenance (remained grounded)
- Corrosion addressed before return to service
- No flight safety risk
- Maintenance documented per 14 CFR Part 43.9

Approved By: Chief Inspector (signature)
Date: Jan 17, 2024
```

---

## General System FAQs

### Q: What browsers are supported?

**A:** Modern browsers with JavaScript enabled.

**Fully Supported:**
- Chrome 90+ (recommended)
- Firefox 88+
- Safari 14+
- Edge 90+

**Not Supported:**
- Internet Explorer (any version)
- Browsers with JavaScript disabled
- Very old browser versions (>3 years old)

**Recommended Setup:**
- Chrome or Firefox for best performance
- Enable cookies (required for authentication)
- Allow localStorage (for UI preferences)
- Screen resolution: 1280x720 minimum (1920x1080 recommended)

---

### Q: How do I change my password?

**A:** User profile settings.

**Web UI:**
1. Click profile icon (top-right corner)
2. Select "Account Settings"
3. Click "Change Password"
4. Enter current password
5. Enter new password (twice for confirmation)
6. Click "Update Password"

**Password Requirements:**
- Minimum 12 characters
- Must include: uppercase, lowercase, number, special character
- Cannot reuse last 5 passwords
- Cannot contain your name or email

**Security Note:** Changing password invalidates all existing login sessions (JWT tokens). You'll need to log in again on all devices.

---

### Q: What if I forget my password?

**A:** Use password reset flow.

**Web UI:**
1. Click "Forgot Password?" on login page
2. Enter your email address
3. Check email for reset link (valid 1 hour)
4. Click link, enter new password
5. Log in with new password

**Email Not Received?**
- Check spam folder
- Verify email address spelling
- Wait 5 minutes (email delivery delay)
- Contact system administrator if still not received

**Security Features:**
- Reset link expires after 1 hour
- Link single-use only (invalidated after password change)
- Old password immediately invalidated
- All existing sessions logged out

---

### Q: Can I use AMSS on mobile devices?

**A:** Yes, responsive web design supports mobile.

**Supported Mobile Browsers:**
- iOS Safari 14+ (iPhone/iPad)
- Android Chrome 90+
- Android Firefox 88+

**Mobile-Optimized Features:**
- Responsive layout (adapts to screen size)
- Touch-friendly buttons (larger tap targets)
- Simplified navigation for small screens
- Quick actions (swipe gestures on some views)

**Limitations on Mobile:**
- CSV import/export better on desktop
- Complex reports easier to view on larger screens
- Bulk operations recommended on desktop

**Recommended Use Cases:**
- **Mobile:** Check task assignments, start/complete tasks, reserve parts, quick status updates
- **Desktop:** Program creation, scheduling, reporting, bulk operations, system administration

---

### Q: Is my data secure?

**A:** Multiple layers of security protect your data.

**Security Features:**

**1. Encryption:**
- **In Transit:** HTTPS/TLS 1.3 for all API calls
- **At Rest:** Database encryption (AES-256)
- **Passwords:** Bcrypt hashing (cannot be decrypted)

**2. Authentication:**
- JWT tokens with 24-hour expiration
- Token refresh on activity
- Automatic logout after 30 min inactivity

**3. Authorization:**
- Role-based access control (5 roles)
- Multi-tenant isolation (org-level)
- Row-level security in database

**4. Audit Trail:**
- All actions logged immutably
- IP address and user agent tracking
- Regulatory compliance documentation

**5. Infrastructure:**
- Regular security updates
- Daily backups (encrypted)
- Firewall protection
- DDoS mitigation

**6. Compliance:**
- SOC 2 Type II (in progress)
- GDPR compliant (EU data residency available)
- FAA Part 91.417 record retention

**Data Isolation:**
Organizations cannot see each other's data:
```sql
-- All queries automatically filtered by org_id
SELECT * FROM tasks WHERE org_id = <your-org-id>

-- Database enforces via row-level security (RLS)
-- Prevents accidental cross-org data access
```

---

### Q: How do I export my data?

**A:** Multiple export options available.

**Export Formats:**

**1. CSV Export (Individual Reports)**
- Navigate to desired report (Tasks, Aircraft, Parts, etc.)
- Click "Export CSV" button
- File downloads with current filters applied

**2. Bulk Export (All Data)**
```bash
GET /api/v1/bulk/export
Authorization: Bearer <tenant-admin-token>

Returns ZIP file containing:
- aircraft.csv
- tasks.csv
- programs.csv
- parts.csv
- part_reservations.csv
- users.csv
- compliance_items.csv
- audit_logs.csv
```

**3. API Access (Custom Integration)**
- Use API endpoints to query data programmatically
- Supports pagination for large datasets
- JSON format for easy parsing

**4. Database Backup (Admin Only)**
```bash
pg_dump -U amss -d amss -F c -f amss-export.backup
```

**Data Portability:**
- CSV format compatible with Excel, Google Sheets
- JSON format for software integration
- Database backup for full migration

**Retention After Export:**
- Data remains in AMSS (export doesn't delete)
- Maintain audit trail per regulatory requirements
- Contact admin for complete data deletion (GDPR right to erasure)

---

## End of Quick Reference Guide

**For Detailed Procedures:** See role-specific user guides:
- [00_ORIENTATION.md](00_ORIENTATION.md) - Getting started
- [01_SYSTEM_ADMINISTRATOR.md](01_SYSTEM_ADMINISTRATOR.md) - Admin guide
- [02_MAINTENANCE_PLANNER.md](02_MAINTENANCE_PLANNER.md) - Scheduler guide
- [03_MECHANIC_TECHNICIAN.md](03_MECHANIC_TECHNICIAN.md) - Mechanic guide
- [04_COMPLIANCE_OFFICER.md](04_COMPLIANCE_OFFICER.md) - Auditor guide
- [05_FLEET_MANAGER.md](05_FLEET_MANAGER.md) - Fleet manager guide

**For Developers:**
- [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md) - Architecture and onboarding
- [API_GUIDE.md](../API_GUIDE.md) - API reference and workflows
- [FAILURE_MODES.md](../FAILURE_MODES.md) - Error handling and recovery

**Support:**
- Email: support@amss.example.com
- Documentation Issues: [GitHub Issues](https://github.com/your-org/amss-backend/issues)

---

**Document Version:** 1.0
**Last Updated:** 2024-01-15
**Next Review:** 2024-04-15 (Quarterly)
