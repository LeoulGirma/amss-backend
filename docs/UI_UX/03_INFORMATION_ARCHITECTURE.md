# AMSS UI/UX Specification - Information Architecture

[Back to Index](00_INDEX.md) | [Previous: Design System](02_DESIGN_SYSTEM.md)

---

# Part 4: Information Architecture

## 4.1 Navigation Structure by Role

### 4.1.1 Admin Navigation

```
Primary Navigation (Sidebar)
├── Dashboard
├── Organizations
│   ├── List All
│   ├── Create New
│   └── [Org Detail]
├── Users
│   ├── All Users
│   ├── Invite User
│   └── Role Management
├── System
│   ├── Settings
│   ├── Integrations
│   ├── Webhooks
│   └── API Keys
├── Reports
│   ├── Usage Analytics
│   ├── Compliance Summary
│   └── Audit Logs
└── Support
    ├── Documentation
    └── System Status
```

### 4.1.2 Tenant Admin Navigation

```
Primary Navigation (Sidebar)
├── Dashboard
├── Fleet
│   ├── Aircraft List
│   ├── Add Aircraft
│   └── Fleet Analytics
├── Maintenance
│   ├── All Tasks
│   ├── Calendar View
│   └── Programs
├── Team
│   ├── Members
│   ├── Invite Member
│   └── Workload View
├── Parts
│   ├── Inventory
│   ├── Orders
│   └── Reservations
├── Compliance
│   ├── Overview
│   ├── Pending Items
│   └── Reports
├── Integrations
│   ├── Webhooks
│   └── API Access
└── Settings
    ├── Organization
    ├── Notifications
    └── Billing
```

### 4.1.3 Scheduler Navigation

```
Primary Navigation (Sidebar)
├── Dashboard
├── Schedule
│   ├── Calendar
│   ├── Timeline (Gantt)
│   └── Conflicts
├── Tasks
│   ├── All Tasks
│   ├── Create Task
│   ├── Pending Approval
│   └── Templates
├── Aircraft
│   ├── Fleet Status
│   ├── Availability
│   └── Maintenance Windows
├── Programs
│   ├── Active Programs
│   ├── Upcoming Due
│   └── Create Program
├── Mechanics
│   ├── Availability
│   ├── Workload
│   └── Assignments
└── Reports
    ├── Schedule Adherence
    └── Resource Utilization
```

### 4.1.4 Mechanic Navigation

```
Primary Navigation (Sidebar)
├── Dashboard
├── My Tasks
│   ├── Today
│   ├── Upcoming
│   ├── In Progress
│   └── Completed
├── Aircraft
│   └── Assigned Aircraft
├── Parts
│   ├── My Reservations
│   └── Request Parts
├── Compliance
│   ├── Pending Sign-offs
│   └── My Sign-offs
└── Profile
    ├── Certifications
    └── Work History
```

### 4.1.5 Auditor Navigation

```
Primary Navigation (Sidebar)
├── Dashboard
├── Compliance
│   ├── All Items
│   ├── Pending Review
│   ├── Signed Off
│   └── Failed Items
├── Audit Logs
│   ├── All Activity
│   ├── By User
│   ├── By Aircraft
│   └── By Task
├── Reports
│   ├── Compliance Reports
│   ├── Audit Trail Export
│   └── Regulatory Reports
└── Aircraft
    └── Compliance Status
```

## 4.2 Complete Sitemap

### 4.2.1 Public Pages

```
/login
/forgot-password
/reset-password/:token
/accept-invite/:token
```

### 4.2.2 Authenticated Pages

```
/dashboard

/aircraft
/aircraft/:id
/aircraft/:id/maintenance
/aircraft/:id/compliance
/aircraft/:id/parts
/aircraft/:id/history
/aircraft/new

/tasks
/tasks/:id
/tasks/:id/compliance
/tasks/:id/parts
/tasks/new
/tasks/calendar
/tasks/timeline

/programs
/programs/:id
/programs/:id/tasks
/programs/new

/parts
/parts/:id
/parts/inventory
/parts/reservations
/parts/reservations/:id
/parts/orders
/parts/new

/compliance
/compliance/:id
/compliance/pending
/compliance/reports

/audit
/audit/logs
/audit/logs/:id
/audit/reports
/audit/export

/team
/team/members
/team/members/:id
/team/invite
/team/workload

/organizations (admin only)
/organizations/:id
/organizations/new

/users (admin only)
/users/:id
/users/invite
/users/roles

/settings
/settings/profile
/settings/notifications
/settings/organization
/settings/integrations
/settings/webhooks
/settings/api-keys
/settings/billing

/reports
/reports/compliance
/reports/maintenance
/reports/utilization
/reports/costs

/help
/help/documentation
/help/support
```

## 4.3 URL Routing Patterns

### 4.3.1 RESTful URL Structure

```
Pattern: /:resource(/:id)(/:sub-resource)(/:sub-id)

Examples:
/aircraft                    # List all aircraft
/aircraft/new                # Create aircraft form
/aircraft/:id                # Aircraft detail
/aircraft/:id/edit           # Edit aircraft form
/aircraft/:id/maintenance    # Aircraft maintenance tasks
/aircraft/:id/compliance     # Aircraft compliance items

/tasks                       # List all tasks
/tasks/:id                   # Task detail
/tasks/:id/parts             # Task parts reservations
/tasks/:id/compliance        # Task compliance items
```

### 4.3.2 Query Parameter Conventions

```
Filtering:
?status=operational,maintenance
?type=inspection,repair
?assigned_to=:user_id
?aircraft_id=:aircraft_id

Sorting:
?sort=created_at
?sort=-updated_at    # Descending
?sort=status,name    # Multi-field

Pagination:
?page=1
?limit=25

Date Ranges:
?start_date=2024-01-01
?end_date=2024-12-31

Search:
?q=search+term
```

## 4.4 Breadcrumb Patterns

### 4.4.1 Breadcrumb Structure

```
Format: Home > Section > Subsection > Current Page

Examples:
Dashboard
Fleet > Aircraft > N12345
Maintenance > Tasks > Task #1234 > Compliance
Parts > Inventory > Part #ABC123
Settings > Integrations > Webhooks
```

### 4.4.2 Breadcrumb Component

```jsx
<Breadcrumb>
  <BreadcrumbItem href="/dashboard">
    <HomeIcon />
    <span>Dashboard</span>
  </BreadcrumbItem>
  <BreadcrumbSeparator />
  <BreadcrumbItem href="/aircraft">Fleet</BreadcrumbItem>
  <BreadcrumbSeparator />
  <BreadcrumbItem href="/aircraft/123">N12345</BreadcrumbItem>
  <BreadcrumbSeparator />
  <BreadcrumbItem current>Maintenance</BreadcrumbItem>
</Breadcrumb>
```

## 4.5 Global Search

### 4.5.1 Search Scope

| Entity | Searchable Fields | Result Display |
|--------|-------------------|----------------|
| Aircraft | Tail number, model, serial | Tail + model + status badge |
| Tasks | ID, notes, aircraft tail | Task type + aircraft + status |
| Parts | Part number, name, serial | Part number + name + quantity |
| Users | Name, email | Name + email + role badge |
| Programs | Name, description | Name + aircraft + interval |
| Compliance | Description, reference | Item + task + status |

### 4.5.2 Search UI (Command Palette)

```
Keyboard Shortcut: Cmd/Ctrl + K

┌─────────────────────────────────────────────────┐
│ Search aircraft, tasks, parts...                │
├─────────────────────────────────────────────────┤
│ Recent Searches                                 │
│   N12345                                        │
│   inspection tasks                              │
├─────────────────────────────────────────────────┤
│ Quick Actions                                   │
│   + Create Task                                 │
│   + Add Aircraft                                │
│   + Reserve Part                                │
├─────────────────────────────────────────────────┤
│ Navigation                                      │
│   > Dashboard                                   │
│   > Calendar                                    │
│   > Settings                                    │
└─────────────────────────────────────────────────┘
```

## 4.6 Navigation Components

### 4.6.1 Sidebar Navigation (Desktop)

```
Width: 240px (expanded), 64px (collapsed)
Position: Fixed left
Background: var(--color-bg-secondary)
Border: 1px solid var(--color-border-subtle)
```

### 4.6.2 Bottom Tab Bar (Mobile)

```
Position: Fixed bottom
Height: 64px
Items: 5 max (Home, Fleet, Tasks, Parts, Profile)
```

### 4.6.3 Header Bar

```
Height: 64px
Position: Fixed top (with sidebar offset on desktop)

Components:
- Menu toggle (mobile/tablet)
- Page title with breadcrumb
- Global search trigger
- Notifications bell with count badge
- User menu dropdown
```

---

# Part 5: Role-Based Dashboards

## 5.1 Admin Dashboard

### 5.1.1 Purpose
System-wide visibility across all organizations, user activity, and platform health.

### 5.1.2 KPI Cards

| Widget | Primary Metric | Secondary Metric |
|--------|---------------|------------------|
| Total Organizations | Count of active orgs | New this month |
| Total Users | Count of all users | New this month |
| Active Tasks | Tasks in progress | On-time percentage |
| System Health | Uptime percentage | Current status |

### 5.1.3 Key Widgets
- **Organization Activity Chart**: 30-day line chart of API calls, task completions
- **Recent Sign-ups**: 5 most recent organizations
- **System Alerts Panel**: Errors, warnings, auto-refresh every 60s
- **API Usage Chart**: Horizontal bar chart by organization
- **Organizations Table**: Name, Users, Aircraft, Tasks, Status

## 5.2 Tenant Admin Dashboard

### 5.2.1 Purpose
Comprehensive view of fleet operations, maintenance status, team workload, and cost metrics.

### 5.2.2 KPI Cards

| Widget | Primary Metric | Alert Threshold |
|--------|---------------|-----------------|
| Fleet Status | Operational count / Total | < 80% operational |
| Active Tasks | In-progress count | Any overdue |
| Compliance Rate | Pass percentage | < 95% |
| Monthly Costs | Total spend | > budget |

### 5.2.3 Key Widgets
- **Fleet Status Grid**: Aircraft cards with status badges
- **Upcoming Maintenance**: Grouped by Today, Tomorrow, This Week
- **Team Workload**: Horizontal progress bars per mechanic
- **Cost Breakdown**: Donut chart (Labor, Parts, External, Other)
- **Integration Status**: Webhook/integration health

## 5.3 Scheduler Dashboard

### 5.3.1 Purpose
Efficient task scheduling, resource availability monitoring, conflict resolution.

### 5.3.2 KPI Cards

| Widget | Primary Metric | Alert State |
|--------|---------------|-------------|
| Today's Tasks | Count scheduled today | Starting within 1 hour |
| Unassigned Tasks | Tasks without mechanic | Any unassigned |
| Schedule Conflicts | Overlapping tasks | Any conflicts |
| Mechanic Availability | Available / Total | < 50% available |

### 5.3.3 Key Widgets
- **Weekly Schedule Grid**: Gantt-lite view (aircraft rows x day columns)
- **Tasks Requiring Action**: Overdue, Starting Soon, Unassigned
- **Mechanic Availability Panel**: Name, shift hours, status
- **Programs Due Soon**: Flight hours, cycles, calendar days remaining
- **Quick Actions**: Create Task, Open Calendar, Assign Mechanics

## 5.4 Mechanic Dashboard

### 5.4.1 Purpose
Focused view of assigned work, parts availability, and compliance requirements.

### 5.4.2 Key Widgets
- **Current Task Card (Hero)**: Full width, task name, aircraft, location, progress bar
- **Compliance Items Panel**: Pending sign-offs with sign-off buttons
- **Reserved Parts Panel**: Parts list with locations/status
- **Upcoming Tasks**: Grouped by Today, Tomorrow
- **Recent Completions**: Last 5 completed tasks

### 5.4.3 Actions
- View Checklist (opens compliance items)
- View Parts (opens reservations)
- Complete Task (with confirmation modal)

## 5.5 Auditor Dashboard

### 5.5.1 Purpose
Comprehensive visibility into regulatory adherence, audit trails, and documentation status.

### 5.5.2 KPI Cards

| Widget | Primary Metric | Alert State |
|--------|---------------|-------------|
| Compliance Rate | Pass percentage | < 95% (amber), < 90% (red) |
| Pending Sign-offs | Count awaiting | Any urgent items |
| Overdue Items | Count past due | Any > 0 (red) |
| Audit Logs Today | Entry count | Anomaly detection |

### 5.5.3 Key Widgets
- **Compliance Overview Chart**: 12-month stacked bar + category breakdown
- **Items Requiring Review**: Overdue and pending sign-off lists
- **Recent Audit Activity**: Chronological activity stream
- **Aircraft Compliance Table**: Status, compliance %, next due, action
- **Quick Reports**: Pre-configured report templates
- **Regulatory References**: Links to FAA Part 43, 91.417, EASA Part-M/145

---

[Next: Screen Specifications](04_SCREENS.md)
