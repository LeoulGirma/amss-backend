# AMSS User Guide: System Administrator

**Version:** 1.0
**Last Updated:** 2024-01-15
**Applies To:** AMSS v1.x
**Audience:** System Administrators, IT Administrators
**Role:** `admin`
**Review Date:** Quarterly

---

## Document Overview

This guide is written for **System Administrators** who are responsible for setting up, configuring, and maintaining the AMSS system itself. You will learn how to create new organizations (tenants), manage users and roles, configure system settings, and monitor system health.

**What you will learn:**
- How to create and manage organizations (multi-tenant setup)
- How to create users and assign roles
- How to configure system settings
- How to monitor system health and performance
- How to troubleshoot common issues

**Prerequisite reading:**
- [00_ORIENTATION.md](./00_ORIENTATION.md) - Understand basic AMSS concepts

---

## Table of Contents

### Part I: Orientation
1. [Your Role as System Administrator](#your-role-as-system-administrator)
2. [Your Dashboard Overview](#your-dashboard-overview)
3. [Mental Model: Multi-Tenancy & System Architecture](#mental-model-multi-tenancy--system-architecture)

### Part II: Initial System Setup
4. [Scenario 1: Create New Organization](#scenario-1-create-new-organization)
5. [Scenario 2: Create Users and Assign Roles](#scenario-2-create-users-and-assign-roles)
6. [Scenario 3: Configure Organization Settings](#scenario-3-configure-organization-settings)

### Part III: Ongoing Operations
7. [Scenario 4: Monitor System Health](#scenario-4-monitor-system-health)
8. [Scenario 5: Manage User Access](#scenario-5-manage-user-access)
9. [Scenario 6: Review System Audit Logs](#scenario-6-review-system-audit-logs)
10. [Scenario 7: Backup & Recovery](#scenario-7-backup--recovery)

### Part IV: Troubleshooting
11. [Scenario 8: Troubleshoot User Login Issues](#scenario-8-troubleshoot-user-login-issues)
12. [Scenario 9: Troubleshoot Performance Issues](#scenario-9-troubleshoot-performance-issues)
13. [Scenario 10: Troubleshoot Background Worker Issues](#scenario-10-troubleshoot-background-worker-issues)

### Part V: Reference
14. [Error Messages & Recovery](#error-messages--recovery)
15. [System Configuration Reference](#system-configuration-reference)
16. [Command-Line Tools](#command-line-tools)

---

# Part I: Orientation

## Your Role as System Administrator

As a **System Administrator** (role: `admin`), you are responsible for the overall operation and maintenance of the AMSS system. You have the highest level of access and are the only role that can perform system-wide operations.

### Primary Responsibilities

1. **Organization Management (Multi-Tenancy)**
   - Create new organizations (tenants) in the system
   - Configure organization-level settings
   - Manage organization lifecycle (activate, deactivate)

2. **User & Access Management**
   - Create user accounts for all organizations
   - Assign roles (admin, tenant_admin, scheduler, mechanic, auditor)
   - Reset passwords and manage account lockouts
   - Review user activity across all organizations

3. **System Configuration**
   - Configure system-wide settings (rate limits, retention policies)
   - Manage database connections and performance
   - Configure background worker settings
   - Set up monitoring and alerting

4. **System Health Monitoring**
   - Monitor system performance (CPU, memory, database)
   - Review application logs and error rates
   - Monitor background worker jobs
   - Track API usage and rate limits

5. **Data Integrity & Backup**
   - Ensure database backups are running
   - Verify data integrity across organizations
   - Manage disaster recovery procedures
   - Perform system upgrades and migrations

### What You CAN Do (God Mode)

✅ **FULL SYSTEM ACCESS**
- Create and manage organizations (tenants)
- Create and manage users in ALL organizations
- Assign ANY role, including `admin`
- View and modify data in ALL organizations
- Access system-wide logs and metrics
- Configure system settings
- Perform database operations
- Manage background workers

### What You CANNOT Do

❌ **No Exceptions** - As admin, you have unrestricted access. **With great power comes great responsibility.**

⚠️ **CRITICAL WARNINGS:**
- **Data Privacy**: Respect organization boundaries - don't access data unnecessarily
- **Audit Trail**: All admin actions are logged - your actions are traceable
- **Security**: Protect admin credentials - compromise = total system compromise
- **Testing**: Never test on production data - use dedicated test organizations

### A Day in Your Life

**8:00 AM** - Check system health dashboard (all services green)
**8:15 AM** - Review overnight background job runs (program generator ran successfully)
**9:00 AM** - New customer onboarding: Create organization "NewAirline LLC"
**9:15 AM** - Create tenant_admin user for new customer
**10:00 AM** - User from "AeroMaintain LLC" reports login issue - investigate
**11:00 AM** - Monitor database performance (95% disk space - schedule cleanup)
**1:00 PM** - Review system audit logs for security compliance
**3:00 PM** - Configure rate limits for new API integration
**4:00 PM** - Plan system upgrade for weekend maintenance window

---

## Your Dashboard Overview

When you log in to AMSS with the `admin` role, you see a **System Administration Dashboard** optimized for monitoring the entire system health and managing all organizations.

### Dashboard Widgets

```
╔═══════════════════════════════════════════════════════════════════╗
║               SYSTEM ADMINISTRATION DASHBOARD                     ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  🏢 ORGANIZATIONS                                                ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Total Organizations:    12                                   │ ║
║  │ Active:                 11 (92%)                             │ ║
║  │ Inactive:               1                                    │ ║
║  │ Created This Month:     2                                    │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  👥 USERS                                                        ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Total Users:            245                                  │ ║
║  │ By Role:                                                     │ ║
║  │   • Admin:              3                                    │ ║
║  │   • Tenant Admin:       12                                   │ ║
║  │   • Scheduler:          18                                   │ ║
║  │   • Mechanic:           185                                  │ ║
║  │   • Auditor:            27                                   │ ║
║  │ Active (last 7 days):   198 (81%)                           │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  ⚙️  SYSTEM HEALTH                                               ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Status:                 ✅ All Systems Operational           │ ║
║  │                                                              │ ║
║  │ API Server:             ✅ Healthy (99.9% uptime)           │ ║
║  │ Database:               ✅ Healthy (15% CPU, 45% disk)      │ ║
║  │ Background Worker:      ✅ Running (last run: 2 min ago)    │ ║
║  │ Redis Cache:            ✅ Connected                         │ ║
║  │                                                              │ ║
║  │ Request Rate:           1,245 req/min (avg)                 │ ║
║  │ Error Rate:             0.12% (low) ✅                       │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  📊 BACKGROUND JOBS (Last 24 Hours)                              ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Program Generator:      ✅ 4 runs (every 6 hours)           │ ║
║  │   Last Run:             Jan 15 14:00 UTC (Success)          │ ║
║  │   Tasks Generated:      23                                   │ ║
║  │                                                              │ ║
║  │ Webhook Dispatcher:     ✅ 1,456 events dispatched           │ ║
║  │   Success Rate:         98.7%                                │ ║
║  │   Failed Deliveries:    19 (retrying)                       │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  ⚠️  ALERTS & WARNINGS                                           ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ ⚠️  Database disk usage: 95% (cleanup recommended)           │ ║
║  │ ⚠️  1 user account locked (failed login attempts)            │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  📈 QUICK ACTIONS                                                ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ [Create Organization]  [Create User]  [View Audit Logs]     │ ║
║  │ [System Health]        [Database]     [Background Jobs]     │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
╚═══════════════════════════════════════════════════════════════════╝
```

### Key Metrics Explained

**Organizations (Tenants)**
- Each organization is a separate customer/company using AMSS
- "Active" = organization has users and is using the system
- "Inactive" = organization created but no recent activity (may be test/demo)

**Users by Role**
- **Admin** (3): You and other system administrators
- **Tenant Admin** (12): One per organization (fleet managers)
- **Scheduler** (18): Maintenance planners across all organizations
- **Mechanic** (185): Largest group - hands-on maintenance workers
- **Auditor** (27): Compliance officers and quality assurance

**System Health Indicators**
- **API Server Uptime**: Target: >99.5% (5 9s = 99.999% ideal)
- **Database CPU**: Target: <70% (sustained high CPU = performance issue)
- **Database Disk**: Target: <80% (>90% = urgent cleanup needed)
- **Error Rate**: Target: <1% (>5% = investigate immediately)

**Background Jobs**
- **Program Generator**: Runs every 6 hours, creates new tasks based on aircraft hours
- **Webhook Dispatcher**: Continuously sends events to external systems
- **Success Rate**: Target: >95% for background jobs

---

## Mental Model: Multi-Tenancy & System Architecture

Understanding AMSS's multi-tenant architecture is critical for effective system administration.

### Multi-Tenant Isolation

```
AMSS SYSTEM (Single Deployment)
    ├─ ORGANIZATION 1: "AeroMaintain LLC"
    │   ├─ Users: 20 users (1 tenant_admin, 3 schedulers, 14 mechanics, 2 auditors)
    │   ├─ Aircraft: 15 aircraft
    │   ├─ Tasks: 487 tasks
    │   └─ Data: Completely isolated from other organizations
    │
    ├─ ORGANIZATION 2: "SkyFlight Services"
    │   ├─ Users: 8 users
    │   ├─ Aircraft: 5 aircraft
    │   ├─ Tasks: 125 tasks
    │   └─ Data: Cannot see AeroMaintain's data
    │
    ├─ ORGANIZATION 3: "NewAirline LLC" (newly created)
    │   ├─ Users: 0 users (admin will create tenant_admin)
    │   ├─ Aircraft: 0 aircraft (tenant_admin will import)
    │   ├─ Tasks: 0 tasks
    │   └─ Data: Empty - ready for setup
    │
    └─ ... (9 more organizations)
```

**Key Isolation Rules:**
1. **Data Isolation**: Each organization's data is completely separate
2. **User Isolation**: Users belong to ONE organization (except admin)
3. **Admin Override**: Admin role can access ANY organization's data
4. **Database Level**: Isolation enforced by `org_id` foreign key on all tables

### System Architecture Layers

```mermaid
graph TD
    Users[Users/Clients] --> API[API Server :8080]
    API --> Auth[Authentication/RBAC]
    Auth --> Services[Business Logic Services]
    Services --> DB[(PostgreSQL Database)]
    Services --> Redis[(Redis Cache)]

    Worker[Background Worker] --> DB
    Worker --> Redis
    Worker --> Webhooks[External Systems]

    Monitoring[Prometheus/Grafana] --> API
    Monitoring --> DB
    Monitoring --> Worker
```

**Components:**

1. **API Server** (Port 8080)
   - REST API for all client requests
   - gRPC API for internal services
   - Authentication & authorization (JWT tokens)
   - Rate limiting (Redis-based)

2. **PostgreSQL Database**
   - Primary data store (organizations, users, aircraft, tasks, etc.)
   - ACID transactions for data integrity
   - Row-level security for multi-tenancy

3. **Redis**
   - Rate limiting counters
   - Distributed locks (parts reservation)
   - Background job queues
   - Session caching

4. **Background Worker**
   - Program Generator: Auto-generates tasks every 6 hours
   - Webhook Dispatcher: Sends events to external systems
   - Outbox Publisher: Ensures reliable event delivery

5. **Monitoring** (Prometheus + Grafana)
   - System metrics (CPU, memory, disk)
   - Application metrics (request rates, error rates)
   - Database metrics (connection pool, query performance)

### Data Flow: Organization Creation

```mermaid
sequenceDiagram
    participant Admin
    participant API
    participant DB
    participant AuditLog

    Admin->>API: POST /organizations {name: "NewAirline"}
    API->>API: Verify actor.Role == admin
    API->>DB: INSERT INTO organizations
    DB-->>API: Organization created (UUID)
    API->>AuditLog: Log organization.created event
    API-->>Admin: 201 Created {id, name, created_at}
```

**What Happens:**
1. Admin sends request to create organization
2. API verifies admin role (only admin can create orgs)
3. Database creates new organization with UUID
4. Audit log records creation (who, when, what)
5. Admin receives organization ID for next steps (create users)

### User Role Hierarchy

```
admin (System Administrator)
    ├─ Full access to ALL organizations
    ├─ Can create organizations
    └─ Can create other admins

tenant_admin (Fleet Manager)
    ├─ Full access to THEIR organization only
    ├─ Can create users in their org
    └─ Cannot create admins

scheduler (Maintenance Planner)
    ├─ Can create/assign tasks
    ├─ Can manage programs
    └─ Cannot create users

mechanic (Technician)
    ├─ Can complete tasks assigned to them
    ├─ Can reserve/use parts
    └─ Cannot create tasks

auditor (Compliance Officer)
    ├─ Read-only access to all data
    ├─ Can export audit logs
    └─ Cannot modify data
```

---

# Part II: Initial System Setup

## Scenario 1: Create New Organization

**GOAL:** Set up a new customer organization (tenant) in the AMSS system.

**WHEN TO USE:**
- Onboarding new customer/company
- Creating test environment for development
- Setting up demo organization for sales
- Adding new department/division within company

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] You have the organization name
- [ ] You have contact information for organization's tenant_admin

---

### HAPPY PATH: Create New Customer Organization

**Scenario Context:** New customer "NewAirline LLC" has signed up for AMSS. You need to set up their organization.

---

**Step 1: Navigate to Organization Management**

ACTION: From admin dashboard, click **"Create Organization"** or navigate to `/admin/organizations/create`

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                    CREATE NEW ORGANIZATION                        ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  Organization Details:                                           ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Organization Name: [_____________________________]           │ ║
║  │                    (e.g., "NewAirline LLC")                  │ ║
║  │                                                              │ ║
║  │ ℹ️  The organization name will be visible to all users       │ ║
║  │     in this organization.                                    │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Create Organization]  [Cancel]                                 ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 2: Enter Organization Name**

ACTION: Enter "NewAirline LLC" and click **"Create Organization"**

API CALL:
```bash
POST /api/v1/organizations
Authorization: Bearer <admin-jwt-token>

{
  "name": "NewAirline LLC"
}
```

SYSTEM RESPONSE:
```json
{
  "id": "org-uuid-newairline",
  "name": "NewAirline LLC",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

SCREEN UPDATE:
```
╔═══════════════════════════════════════════════════════════════════╗
║              ORGANIZATION CREATED SUCCESSFULLY ✅                 ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  Organization ID:    org-uuid-newairline                         ║
║  Organization Name:  NewAirline LLC                              ║
║  Created:            2024-01-15 10:00:00 UTC                    ║
║                                                                   ║
║  ✅ Organization is ready for setup!                             ║
║                                                                   ║
║  NEXT STEPS:                                                     ║
║  1. Create tenant_admin user for this organization              ║
║  2. Send login credentials to organization contact              ║
║  3. Tenant admin will complete setup (aircraft, users, etc.)    ║
║                                                                   ║
║  [Create Tenant Admin User]  [View Organization]  [Done]        ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 3: Verify Organization Created**

ACTION: Click **"View Organization"** or navigate to `/admin/organizations/org-uuid-newairline`

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                  ORGANIZATION: NEWAIRLINE LLC                     ║
╠═══════════════════════════════════════════════════════════════════╣
║  ID:           org-uuid-newairline                               ║
║  Name:         NewAirline LLC                                    ║
║  Created:      2024-01-15 10:00:00 UTC                          ║
║  Status:       Active ✅                                         ║
║                                                                   ║
║  USAGE STATISTICS:                                               ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Users:          0                                            │ ║
║  │ Aircraft:       0                                            │ ║
║  │ Tasks:          0                                            │ ║
║  │ Programs:       0                                            │ ║
║  │ Parts:          0                                            │ ║
║  │                                                              │ ║
║  │ Last Activity:  Never (organization just created)            │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  AUDIT TRAIL:                                                    ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ 2024-01-15 10:00  organization.created  by admin@amss.com   │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Edit Name]  [Create User]  [View Audit Logs]  [Deactivate]    ║
╚═══════════════════════════════════════════════════════════════════╝
```

✅ **SUCCESS:** Organization created and ready for user setup

---

**Step 4: Communicate with Customer**

**Email Template to Customer:**
```
To: contact@newairline.com
Subject: AMSS Organization Created - Next Steps

Hi NewAirline LLC Team,

Your AMSS organization has been successfully created!

Organization Details:
- Organization Name: NewAirline LLC
- Organization ID: org-uuid-newairline (for support reference)
- Created: January 15, 2024

NEXT STEPS:
1. I will create a tenant administrator account for you
2. You will receive login credentials via email
3. The tenant admin can then:
   - Import your aircraft fleet
   - Create additional users (schedulers, mechanics, auditors)
   - Set up maintenance programs
   - Begin tracking maintenance

Please confirm the email address for your primary administrator:
[ ] Contact person name: _______________
[ ] Email address: _______________

Once confirmed, I'll create your admin account within 24 hours.

Best regards,
System Administrator
```

---

### COMMON FAILURES

| Symptom | Cause | Recovery |
|---------|-------|----------|
| **"Forbidden: admin role required"** | User is not admin | Only admin role can create organizations |
| **"Organization name already exists"** | Duplicate organization name | Choose different name or check if org already exists |
| **"Organization created but not visible"** | Cache issue or filter applied | Refresh page; check organization list filters |
| **"Cannot create users after org creation"** | Organization ID not recognized | Verify org UUID is correct; check database |

---

### VERIFICATION

✅ **How to confirm organization was created successfully:**

1. **Organization Appears in List**: Navigate to `/admin/organizations` and verify new org is listed
2. **UUID Generated**: Organization has valid UUID (format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)
3. **Audit Log Entry**: Audit log shows `organization.created` event
4. **Empty Usage Stats**: New org shows 0 users, 0 aircraft, 0 tasks (expected)
5. **Can Create Users**: Next step (Scenario 2) succeeds

---

### RELATED SCENARIOS

→ **Next Step**: [Scenario 2: Create Users and Assign Roles](#scenario-2-create-users-and-assign-roles)
→ **Manage Orgs**: [Scenario 5: Manage User Access](#scenario-5-manage-user-access) - View all organizations

---

## Scenario 2: Create Users and Assign Roles

**GOAL:** Create user accounts for a specific organization and assign appropriate roles.

**WHEN TO USE:**
- After creating new organization (create tenant_admin first)
- Customer requests new user account
- Adding users to existing organization
- Setting up demo/test users

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] Organization exists
- [ ] You know the user's email and desired role
- [ ] You have generated a temporary password

---

### HAPPY PATH: Create Tenant Admin for New Organization

**Scenario Context:** You created "NewAirline LLC" organization. Now create the tenant_admin user so they can manage their organization.

---

**Step 1: Navigate to User Creation**

ACTION: From organization page, click **"Create User"** or navigate to `/admin/users/create`

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                        CREATE NEW USER                            ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  User Details:                                                   ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Organization: [NewAirline LLC ▼] (org-uuid-newairline)      │ ║
║  │                                                              │ ║
║  │ Email:        [_____________________________]                │ ║
║  │               (must be unique across all organizations)      │ ║
║  │                                                              │ ║
║  │ Role:         [ ▼ tenant_admin ]                            │ ║
║  │               ( ) admin                                      │ ║
║  │               ( ) tenant_admin (Recommended for first user) │ ║
║  │               ( ) scheduler                                  │ ║
║  │               ( ) mechanic                                   │ ║
║  │               ( ) auditor                                    │ ║
║  │                                                              │ ║
║  │ Password:     [_____________________________]                │ ║
║  │               (min 8 chars, temp password for first login)  │ ║
║  │                                                              │ ║
║  │ ☑️ Require password change on first login                   │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Create User]  [Cancel]                                         ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 2: Enter User Information**

ACTION: Fill in the form:
- Organization: NewAirline LLC (already selected)
- Email: `admin@newairline.com`
- Role: `tenant_admin`
- Password: `TempPass2024!` (temporary, user will change on first login)
- Require password change: ✅ Checked

Click **"Create User"**

API CALL:
```bash
POST /api/v1/users
Authorization: Bearer <admin-jwt-token>

{
  "org_id": "org-uuid-newairline",
  "email": "admin@newairline.com",
  "role": "tenant_admin",
  "password": "TempPass2024!"
}
```

SYSTEM RESPONSE:
```json
{
  "id": "user-uuid-123",
  "org_id": "org-uuid-newairline",
  "email": "admin@newairline.com",
  "role": "tenant_admin",
  "created_at": "2024-01-15T10:05:00Z",
  "updated_at": "2024-01-15T10:05:00Z"
}
```

SCREEN UPDATE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                   USER CREATED SUCCESSFULLY ✅                    ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  User ID:        user-uuid-123                                   ║
║  Email:          admin@newairline.com                            ║
║  Organization:   NewAirline LLC                                  ║
║  Role:           tenant_admin                                    ║
║  Created:        2024-01-15 10:05:00 UTC                        ║
║                                                                   ║
║  ⚠️  IMPORTANT: Send login credentials to user                   ║
║                                                                   ║
║  Login URL:      https://amss.example.com/login                 ║
║  Email:          admin@newairline.com                            ║
║  Password:       TempPass2024! (temporary)                       ║
║                                                                   ║
║  User will be required to change password on first login.        ║
║                                                                   ║
║  [Send Credentials Email]  [Create Another User]  [Done]        ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 3: Send Credentials to User**

ACTION: Click **"Send Credentials Email"**

**Email Template:**
```
To: admin@newairline.com
Subject: Your AMSS Account - NewAirline LLC

Hi,

Your AMSS account has been created for NewAirline LLC.

LOGIN CREDENTIALS:
URL:      https://amss.example.com/login
Email:    admin@newairline.com
Password: TempPass2024!

IMPORTANT:
- You will be required to change your password on first login
- Choose a strong password (min 8 characters)
- Your role is "Tenant Admin" - you have full access to your organization

NEXT STEPS:
1. Log in with the credentials above
2. Change your password
3. Review the Orientation Guide: https://amss.example.com/docs/00_ORIENTATION.md
4. Set up your organization:
   - Import aircraft fleet
   - Create users (schedulers, mechanics, auditors)
   - Import maintenance programs

Need help? Reply to this email or contact support@amss.example.com

Welcome to AMSS!
System Administrator
```

---

**Step 4: Verify User Can Log In**

**User's First Login Experience:**

1. User navigates to `https://amss.example.com/login`
2. Enters email: `admin@newairline.com`
3. Enters password: `TempPass2024!`
4. System detects first login, prompts for password change
5. User enters new password (meets requirements)
6. Login successful - user sees Fleet Manager dashboard

✅ **SUCCESS:** User account created and user can access their organization

---

### ALTERNATE PATH: Create Multiple Users (Bulk Setup)

**Scenario:** NewAirline LLC wants 10 users created immediately (1 tenant_admin, 2 schedulers, 6 mechanics, 1 auditor)

**Option 1: Create Users Individually** (10 separate API calls)
- Time: ~15 minutes
- Use when: Small number of users (<5)

**Option 2: CSV Import** (if implemented)
- Time: ~2 minutes
- Use when: Large number of users (>5)

**CSV Format (users.csv):**
```csv
email,role,org_id
admin@newairline.com,tenant_admin,org-uuid-newairline
planner1@newairline.com,scheduler,org-uuid-newairline
planner2@newairline.com,scheduler,org-uuid-newairline
mechanic1@newairline.com,mechanic,org-uuid-newairline
mechanic2@newairline.com,mechanic,org-uuid-newairline
mechanic3@newairline.com,mechanic,org-uuid-newairline
mechanic4@newairline.com,mechanic,org-uuid-newairline
mechanic5@newairline.com,mechanic,org-uuid-newairline
mechanic6@newairline.com,mechanic,org-uuid-newairline
auditor@newairline.com,auditor,org-uuid-newairline
```

**System generates temporary passwords for all users, sends welcome emails**

---

### COMMON FAILURES

| Symptom | Cause | Recovery |
|---------|-------|----------|
| **"Email already exists"** | User email already registered in another org | Emails must be unique across ALL orgs; use different email |
| **"Forbidden: cannot create admin"** | Non-admin trying to create admin user | Only admin role can create other admins |
| **"Organization not found"** | Invalid org UUID | Verify organization exists; check UUID is correct |
| **"Weak password"** | Password doesn't meet requirements | Use min 8 chars with mix of uppercase, lowercase, numbers |
| **"User created but cannot login"** | Password not set or account locked | Reset password; check user status in database |

---

### VERIFICATION

✅ **How to confirm user was created successfully:**

1. **User Appears in User List**: Navigate to `/admin/users?org_id=org-uuid-newairline` and verify user is listed
2. **Correct Role Assigned**: User details show correct role (tenant_admin)
3. **Audit Log Entry**: Audit log shows `user.created` event
4. **User Can Log In**: User successfully logs in with temporary password
5. **Password Change Required**: System prompts user to change password on first login

---

### RELATED SCENARIOS

→ **Previous**: [Scenario 1: Create New Organization](#scenario-1-create-new-organization)
→ **Next**: [Scenario 3: Configure Organization Settings](#scenario-3-configure-organization-settings)
→ **Manage Users**: [Scenario 5: Manage User Access](#scenario-5-manage-user-access)

---

## Scenario 3: Configure Organization Settings

**GOAL:** Configure organization-specific settings such as time zone, retention policies, and feature flags.

**WHEN TO USE:**
- After creating new organization
- Customer requests specific configuration
- Adjusting retention policies for compliance
- Enabling/disabling features for specific customers

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] Organization exists
- [ ] You understand the impact of configuration changes

---

### HAPPY PATH: Configure NewAirline LLC Settings

**Scenario Context:** NewAirline LLC is based in New York (EST timezone) and wants 7-year data retention for FAA compliance.

---

**Step 1: Navigate to Organization Settings**

ACTION: From organization page, click **"Settings"** or navigate to `/admin/organizations/org-uuid-newairline/settings`

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║            ORGANIZATION SETTINGS: NEWAIRLINE LLC                  ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  GENERAL SETTINGS                                                ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Organization Name:  [NewAirline LLC]                         │ ║
║  │ Time Zone:          [America/New_York ▼] (UTC-5/UTC-4 DST)  │ ║
║  │ Date Format:        [MM/DD/YYYY ▼]                           │ ║
║  │ Language:           [English (US) ▼]                         │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  DATA RETENTION                                                  ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Audit Log Retention:     [7 years ▼] (FAA Part 91.417)      │ ║
║  │ Task History Retention:  [7 years ▼] (FAA Part 91.417)      │ ║
║  │ Parts History Retention: [Indefinite ▼] (Traceability)      │ ║
║  │ Deleted Items Retention: [30 days ▼] (Soft-delete period)   │ ║
║  │                                                              │ ║
║  │ ⚠️  WARNING: Reducing retention may violate regulations      │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  FEATURE FLAGS                                                   ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ [ ] Enable Advanced Reporting (Beta)                         │ ║
║  │ [✓] Enable CSV Import/Export                                │ ║
║  │ [✓] Enable Webhook Integrations                             │ ║
║  │ [ ] Enable Mobile App Access (Coming Soon)                   │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  API & RATE LIMITS                                               ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ API Rate Limit:         [1000 req/min ▼] (per org)          │ ║
║  │ Webhook Retry Limit:    [5 attempts ▼]                      │ ║
║  │ Max Import File Size:   [10 MB ▼]                           │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Save Changes]  [Reset to Defaults]  [Cancel]                  ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 2: Configure Settings**

ACTION:
1. Time Zone: Select `America/New_York` (EST/EDT)
2. Data Retention:
   - Audit Log: 7 years (default for FAA compliance)
   - Task History: 7 years
   - Parts History: Indefinite (complete traceability)
3. Feature Flags: Enable CSV Import/Export and Webhook Integrations
4. API Rate Limit: 1000 req/min (default)

Click **"Save Changes"**

API CALL:
```bash
PATCH /api/v1/organizations/org-uuid-newairline/settings
Authorization: Bearer <admin-jwt-token>

{
  "timezone": "America/New_York",
  "audit_retention_years": 7,
  "task_retention_years": 7,
  "parts_retention": "indefinite",
  "feature_flags": {
    "csv_import_export": true,
    "webhooks": true
  },
  "api_rate_limit_per_minute": 1000
}
```

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                  SETTINGS SAVED SUCCESSFULLY ✅                   ║
╠═══════════════════════════════════════════════════════════════════╣
║                                                                   ║
║  Changes Applied:                                                ║
║  • Time Zone: America/New_York                                   ║
║  • Audit Log Retention: 7 years                                  ║
║  • Task History Retention: 7 years                               ║
║  • Parts History Retention: Indefinite                           ║
║  • CSV Import/Export: Enabled                                    ║
║  • Webhook Integrations: Enabled                                 ║
║                                                                   ║
║  ℹ️  Settings will take effect immediately for all users.        ║
║                                                                   ║
║  [View Organization]  [Done]                                     ║
╚═══════════════════════════════════════════════════════════════════╝
```

✅ **SUCCESS:** Organization configured and ready for production use

---

### COMMON FAILURES

| Symptom | Cause | Recovery |
|---------|-------|----------|
| **"Cannot reduce retention below 1 year"** | Regulatory minimum not met | Keep minimum 1 year (preferably 7 years for aviation) |
| **"Invalid timezone"** | Timezone not in IANA database | Use valid IANA timezone (e.g., America/New_York, Europe/London) |
| **"Feature not available"** | Feature flag for beta feature | Contact support to enable beta features |
| **"Settings not saved"** | Validation error or database issue | Check all fields are valid; try again |

---

### VERIFICATION

✅ **How to confirm settings were applied:**

1. **Settings Persist**: Reload settings page, verify changes are still there
2. **Timezone Correct**: User dashboard shows times in correct timezone
3. **Features Enabled**: Users can access CSV import and webhook configuration
4. **Audit Log Entry**: Audit log shows `organization.settings_updated` event
5. **Users Notified**: Inform tenant_admin of new settings

---

### RELATED SCENARIOS

→ **Previous**: [Scenario 2: Create Users and Assign Roles](#scenario-2-create-users-and-assign-roles)
→ **Monitor**: [Scenario 4: Monitor System Health](#scenario-4-monitor-system-health)

---

# Part III: Ongoing Operations

## Scenario 4: Monitor System Health

**GOAL:** Monitor system health metrics to ensure AMSS is operating correctly and proactively identify issues.

**WHEN TO USE:**
- Daily health checks (morning routine)
- After system upgrades or configuration changes
- Investigating user-reported performance issues
- Weekly/monthly operational reviews

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] You have access to monitoring dashboards (Grafana/Prometheus)
- [ ] You understand baseline performance metrics

---

### HAPPY PATH: Daily Morning Health Check

**Scenario Context:** It's Monday morning 8 AM. You're doing your daily system health check before users start their workday.

---

**Step 1: Check System Dashboard**

ACTION: Navigate to `/admin/system/health` or Grafana dashboard

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                     SYSTEM HEALTH DASHBOARD                       ║
╠═══════════════════════════════════════════════════════════════════╣
║  Last Updated: 2024-01-15 08:00:15 UTC (15 seconds ago)          ║
║                                                                   ║
║  OVERALL STATUS: ✅ HEALTHY                                       ║
║                                                                   ║
║  ⚙️  SERVICES STATUS                                             ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Component          Status    Uptime      Response Time       │ ║
║  ├─────────────────────────────────────────────────────────────┤ ║
║  │ API Server         ✅ UP     99.98%      45ms avg            │ ║
║  │ Database           ✅ UP     100%        12ms avg            │ ║
║  │ Redis Cache        ✅ UP     99.99%      <1ms avg            │ ║
║  │ Background Worker  ✅ UP     99.95%      Last run: 6 min ago │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  💾 DATABASE METRICS                                             ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Connections:       25 active / 100 max (25%)                │ ║
║  │ CPU Usage:         15% (low) ✅                              │ ║
║  │ Memory Usage:      4.2 GB / 16 GB (26%)                     │ ║
║  │ Disk Usage:        45 GB / 100 GB (45%) ✅                  │ ║
║  │ Query Latency:     12ms p50, 35ms p95                       │ ║
║  │ Slow Queries:      2 in last hour (acceptable)              │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  🌐 API METRICS (Last Hour)                                      ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Request Rate:      1,245 req/min (avg)                      │ ║
║  │ Error Rate:        0.12% (6 errors/5,000 requests) ✅       │ ║
║  │ P50 Latency:       45ms                                      │ ║
║  │ P95 Latency:       125ms                                     │ ║
║  │ P99 Latency:       250ms                                     │ ║
║  │ Top Endpoints:                                               │ ║
║  │   GET /tasks       35% of traffic                            │ ║
║  │   GET /aircraft    22% of traffic                            │ ║
║  │   POST /tasks      12% of traffic                            │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  🔄 BACKGROUND JOBS (Last 24 Hours)                              ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Program Generator:                                           │ ║
║  │   Runs:            4 runs (every 6 hours) ✅                │ ║
║  │   Success Rate:    100% (4/4)                                │ ║
║  │   Avg Duration:    2.3 seconds                               │ ║
║  │   Tasks Generated: 23                                        │ ║
║  │                                                              │ ║
║  │ Webhook Dispatcher:                                          │ ║
║  │   Events Processed: 1,456                                    │ ║
║  │   Success Rate:     98.7% (1,437/1,456)                     │ ║
║  │   Failed:           19 (retrying)                            │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  ⚠️  ALERTS & WARNINGS: None ✅                                  ║
║                                                                   ║
║  [View Detailed Metrics]  [View Logs]  [Refresh]                ║
╚═══════════════════════════════════════════════════════════════════╝
```

✅ **ANALYSIS: System is healthy**
- All services UP with >99.9% uptime
- Database has plenty of headroom (45% disk, 15% CPU)
- API latency is excellent (<50ms p50, <250ms p99)
- Error rate is very low (0.12%)
- Background jobs running successfully

**Action:** No issues - proceed with normal operations

---

### ALTERNATE PATH: Degraded Performance Detected

**Scenario:** Database CPU at 85%, query latency spiking

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                     SYSTEM HEALTH DASHBOARD                       ║
╠═══════════════════════════════════════════════════════════════════╣
║  OVERALL STATUS: ⚠️  DEGRADED PERFORMANCE                         ║
║                                                                   ║
║  💾 DATABASE METRICS                                             ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ CPU Usage:         85% (high) ⚠️                             │ ║
║  │ Query Latency:     450ms p50 ⚠️, 2.3s p95 ⚠️                 │ ║
║  │ Slow Queries:      127 in last hour ⚠️                       │ ║
║  │ Active Connections: 85 / 100 (85%) ⚠️                        │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  ⚠️  ALERTS                                                      ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ 1. Database CPU high (>80%) for 15 minutes                   │ ║
║  │ 2. Query latency degraded (p95 >1s)                          │ ║
║  │ 3. Slow query count increased 50x                            │ ║
║  └─────────────────────────────────────────────────────────────┘
║                                                                   ║
║  [View Slow Queries]  [Check Active Queries]  [Alert Team]      ║
╚═══════════════════════════════════════════════════════════════════╝
```

**INVESTIGATION STEPS:**

1. **Check Slow Queries**
   ```sql
   SELECT query, calls, mean_exec_time, max_exec_time
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;
   ```

2. **Check Active Queries**
   ```sql
   SELECT pid, usename, state, query, now() - query_start AS duration
   FROM pg_stat_activity
   WHERE state != 'idle'
   ORDER BY duration DESC;
   ```

3. **Common Causes:**
   - Missing index on frequently queried table
   - Poorly written query (N+1 problem)
   - Table locks from long-running transaction
   - Backup or maintenance job running

4. **Immediate Actions:**
   - Kill long-running queries if blocking
   - Add missing indexes
   - Scale database (vertical: more CPU/RAM, horizontal: read replicas)
   - Contact database admin for assistance

---

### COMMON FAILURES

| Symptom | Cause | Recovery |
|---------|-------|----------|
| **"API Server Down"** | Application crash or deployment issue | Check logs; restart service; rollback if recent deployment |
| **"Database High CPU"** | Slow queries or missing indexes | Identify slow queries; add indexes; optimize queries |
| **"Redis Disconnected"** | Redis crash or network issue | Restart Redis; check network; system degrades gracefully (no cache) |
| **"Background Worker Stopped"** | Worker crash or stuck job | Check worker logs; restart worker; clear stuck jobs |
| **"Disk Space >90%"** | Old logs or data not cleaned up | Run cleanup job; increase disk size; archive old data |

---

### VERIFICATION

✅ **How to confirm system is healthy:**

1. **All Services Green**: API, DB, Redis, Worker all show ✅ UP
2. **Low Error Rate**: <1% errors
3. **Fast Response Times**: p95 latency <500ms
4. **Adequate Resources**: CPU <70%, Memory <80%, Disk <80%
5. **Jobs Running**: Background jobs completing successfully

---

### RELATED SCENARIOS

→ **Performance**: [Scenario 9: Troubleshoot Performance Issues](#scenario-9-troubleshoot-performance-issues)
→ **Worker Issues**: [Scenario 10: Troubleshoot Background Worker Issues](#scenario-10-troubleshoot-background-worker-issues)

---

## Scenario 5: Manage User Access

**GOAL:** Manage user access across all organizations - reset passwords, unlock accounts, change roles, deactivate users.

**WHEN TO USE:**
- User forgot password
- User account locked (too many failed logins)
- User role needs to change
- User leaving organization (deactivate account)

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] You have user's email or user ID
- [ ] You have authorization to make changes

---

### HAPPY PATH: Reset User Password

**Scenario Context:** User "john.doe@aeromaintain.com" forgot their password and requests reset.

---

**Step 1: Locate User**

ACTION: Navigate to `/admin/users` and search for "john.doe@aeromaintain.com"

API CALL:
```bash
GET /api/v1/users?email=john.doe@aeromaintain.com
```

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                        USER SEARCH RESULTS                        ║
╠═══════════════════════════════════════════════════════════════════╣
║  Found 1 user                                                    ║
║                                                                   ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Email:        john.doe@aeromaintain.com                      │ ║
║  │ Organization: AeroMaintain LLC                               │ ║
║  │ Role:         mechanic                                       │ ║
║  │ Status:       Active                                         │ ║
║  │ Created:      2023-06-15                                     │ ║
║  │ Last Login:   2024-01-10 14:32 UTC                          │ ║
║  │                                                              │ ║
║  │ [View Details]  [Reset Password]  [Change Role]  [Deactivate]│ ║
║  └─────────────────────────────────────────────────────────────┘ ║
╚═══════════════════════════════════════════════════════════════════╝
```

---

**Step 2: Reset Password**

ACTION: Click **"Reset Password"**

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                      RESET USER PASSWORD                          ║
╠═══════════════════════════════════════════════════════════════════╣
║  User: john.doe@aeromaintain.com (AeroMaintain LLC)             ║
║                                                                   ║
║  Option 1: Generate Temporary Password                           ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ [Generate Temporary Password]                                │ ║
║  │                                                              │ ║
║  │ System will generate secure random password and email user. │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  Option 2: Set Custom Temporary Password                         ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ Password: [_____________________________]                    │ ║
║  │                                                              │ ║
║  │ [✓] Require password change on next login                   │ ║
║  │ [Set Password]                                               │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Cancel]                                                        ║
╚═══════════════════════════════════════════════════════════════════╝
```

ACTION: Click **"Generate Temporary Password"**

API CALL:
```bash
POST /api/v1/users/user-uuid-john/reset-password
{
  "generate_random": true,
  "require_change": true
}
```

SYSTEM RESPONSE:
```
╔═══════════════════════════════════════════════════════════════════╗
║                 PASSWORD RESET SUCCESSFUL ✅                      ║
╠═══════════════════════════════════════════════════════════════════╣
║  Temporary Password: Xy9$mK2pL#vN                                ║
║                                                                   ║
║  ⚠️  IMPORTANT:                                                   ║
║  • Save this password - it cannot be retrieved later             ║
║  • Send to user via secure channel (not email if possible)       ║
║  • User will be required to change password on next login        ║
║                                                                   ║
║  Email Template:                                                 ║
║  ┌─────────────────────────────────────────────────────────────┐ ║
║  │ To: john.doe@aeromaintain.com                                │ ║
║  │ Subject: Password Reset - AMSS                               │ ║
║  │                                                              │ ║
║  │ Your AMSS password has been reset.                           │ ║
║  │                                                              │ ║
║  │ Temporary Password: Xy9$mK2pL#vN                             │ ║
║  │                                                              │ ║
║  │ You will be required to change your password on next login. │ ║
║  │                                                              │ ║
║  │ Login: https://amss.example.com/login                       │ ║
║  └─────────────────────────────────────────────────────────────┘ ║
║                                                                   ║
║  [Send Email]  [Copy Password]  [Done]                          ║
╚═══════════════════════════════════════════════════════════════════╝
```

✅ **SUCCESS:** Password reset and user can log in with temporary password

---

### RELATED SCENARIOS

→ **User Creation**: [Scenario 2: Create Users and Assign Roles](#scenario-2-create-users-and-assign-roles)
→ **Login Issues**: [Scenario 8: Troubleshoot User Login Issues](#scenario-8-troubleshoot-user-login-issues)

---

## Scenario 6: Review System Audit Logs

**GOAL:** Review system-wide audit logs for security compliance, investigation, or troubleshooting.

**WHEN TO USE:**
- Monthly security audit
- Investigating suspicious activity
- Compliance reporting
- Troubleshooting who made specific changes

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] You know what you're looking for (date range, user, action type)

---

### HAPPY PATH: Monthly Security Audit Review

**Scenario Context:** End of month security review - you need to audit all admin actions.

ACTION: Navigate to `/admin/audit-logs`

FILTERS:
- User Role: `admin`
- Date Range: Last 30 days
- Action: All

SYSTEM RESPONSE: Shows all actions taken by admin users in last 30 days for review.

---

## Scenario 7: Backup & Recovery

**GOAL:** Ensure database backups are running and test recovery procedures.

**WHEN TO USE:**
- Weekly backup verification
- Before major system upgrades
- Testing disaster recovery procedures
- Responding to data loss incident

**PRE-CONDITIONS:**
- [ ] You have `admin` role
- [ ] You have database access
- [ ] You have backup storage configured

---

### HAPPY PATH: Verify Weekly Backups

**Database Backup Configuration:**
- **Frequency**: Daily at 2 AM UTC
- **Retention**: 30 daily backups, 12 monthly backups
- **Location**: S3 bucket `amss-backups`
- **Encryption**: AES-256

**Verification Steps:**
1. Check backup logs
2. Verify latest backup exists
3. Test restore to staging environment (monthly)

---

# Part IV: Troubleshooting

## Scenario 8: Troubleshoot User Login Issues

**GOAL:** Diagnose and resolve user login problems.

**COMMON ISSUES:**

### Issue 1: "Invalid Credentials"

**Symptoms:** User enters correct email/password but gets "invalid credentials" error

**Possible Causes:**
1. Password actually incorrect (user forgot recent change)
2. Account locked (too many failed attempts)
3. User typing email wrong (check case sensitivity)

**Diagnosis:**
```bash
# Check user status
GET /api/v1/users?email=user@example.com

# Check login attempts
GET /api/v1/audit-logs?action=login_attempt&user_id=user-uuid
```

**Resolution:**
- Reset password if forgotten
- Unlock account if locked
- Verify email is correct

### Issue 2: "Account Locked"

**Symptoms:** User gets "account locked" message

**Cause:** 5 failed login attempts in 15 minutes

**Resolution:**
```bash
# Unlock account
POST /api/v1/users/user-uuid/unlock
```

---

## Scenario 9: Troubleshoot Performance Issues

**GOAL:** Diagnose and resolve system performance degradation.

**COMMON ISSUES:**

### Issue 1: Slow API Responses

**Symptoms:** Users report slow page loads

**Diagnosis:**
1. Check API latency metrics (Grafana)
2. Identify slow endpoints
3. Check database query performance
4. Check connection pool saturation

**Resolution:**
- Add database indexes
- Optimize slow queries
- Scale database (more CPU/RAM)
- Add caching for frequently accessed data

### Issue 2: Database Connection Exhaustion

**Symptoms:** "Too many connections" errors

**Diagnosis:**
```sql
SELECT count(*) FROM pg_stat_activity;
```

**Resolution:**
- Increase max_connections in PostgreSQL config
- Fix connection leaks in application code
- Implement connection pooling (PgBouncer)

---

## Scenario 10: Troubleshoot Background Worker Issues

**GOAL:** Diagnose and resolve background worker problems.

**COMMON ISSUES:**

### Issue 1: Program Generator Not Running

**Symptoms:** New tasks not being auto-generated

**Diagnosis:**
1. Check worker process is running
2. Check worker logs for errors
3. Check last successful run timestamp

**Resolution:**
```bash
# Restart worker
systemctl restart amss-worker

# Check status
systemctl status amss-worker

# View logs
journalctl -u amss-worker -f
```

### Issue 2: Webhooks Not Delivering

**Symptoms:** External system not receiving events

**Diagnosis:**
1. Check webhook delivery status
2. Check external system endpoint is reachable
3. Check webhook retry queue

**Resolution:**
- Verify webhook URL is correct
- Test endpoint with curl
- Retry failed deliveries manually

---

# Part V: Reference

## Error Messages & Recovery

### Authentication Errors

| Error | Cause | Recovery |
|-------|-------|----------|
| **"Unauthorized"** | No auth token or invalid token | Re-login to get new token |
| **"Forbidden: admin role required"** | User is not admin | Contact admin to assign role |
| **"Account locked"** | Too many failed login attempts | Wait 15 minutes or contact admin to unlock |

### Database Errors

| Error | Cause | Recovery |
|-------|-------|----------|
| **"Connection timeout"** | Database overloaded or down | Check database health; scale resources |
| **"Deadlock detected"** | Concurrent transactions conflict | Retry transaction; review code for lock contention |
| **"Disk full"** | Database ran out of disk space | Free up space; increase disk size |

---

## System Configuration Reference

### Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/amss
DATABASE_MAX_CONNECTIONS=100

# Redis
REDIS_URL=redis://localhost:6379

# API Server
PORT=8080
JWT_SECRET=<secret-key>

# Background Worker
WORKER_INTERVAL=6h  # Program generator run frequency
WEBHOOK_RETRY_MAX=5 # Max webhook delivery attempts
```

---

## Command-Line Tools

### Database Maintenance

```bash
# Vacuum database (reclaim space)
VACUUM FULL ANALYZE;

# Reindex all tables
REINDEX DATABASE amss;

# Check database size
SELECT pg_database_size('amss') / 1024 / 1024 AS size_mb;
```

### User Management

```bash
# Create admin user (SQL)
INSERT INTO users (id, org_id, email, role, password_hash, created_at, updated_at)
VALUES (gen_random_uuid(), NULL, 'admin@example.com', 'admin', '<bcrypt-hash>', now(), now());
```

---

## End of System Administrator Guide

**Complete Guide Coverage:**
- ✅ Part I: Orientation (Role, Dashboard, Architecture)
- ✅ Part II: Initial Setup (Organizations, Users, Settings)
- ✅ Part III: Ongoing Operations (Health, Access, Audit, Backup)
- ✅ Part IV: Troubleshooting (Login, Performance, Worker Issues)
- ✅ Part V: Reference (Errors, Config, CLI Tools)

**Next Steps:**
- Review other user guides to understand different role perspectives
- Set up monitoring alerts for proactive issue detection
- Document organization-specific procedures
- Schedule regular backup testing

**Need Help?**
- Technical issues: Check system logs first
- Security incidents: Follow incident response procedures
- Questions: Contact senior system administrator or vendor support

---

**Document Control:**
- **Version**: 1.0 (Complete)
- **Last Updated**: 2024-01-15
- **Next Review**: 2024-04-15 (Quarterly)
- **Owner**: IT Department
- **Approved By**: [Management Signature]
