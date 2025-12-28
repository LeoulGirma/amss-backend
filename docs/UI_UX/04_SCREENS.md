# AMSS UI/UX Specification - Screen Specifications

[Back to Index](00_INDEX.md) | [Previous: Information Architecture](03_INFORMATION_ARCHITECTURE.md)

---

# Part 6: Core Screen Specifications

This section provides detailed specifications for each screen in the AMSS application, organized by functional category.

---

## 6.1 Authentication Screens

### 6.1.1 Login Screen

**URL Pattern:** `/login`

**Purpose:** Authenticate users and establish session credentials.

**User Story:** As a user, I want to securely log into the system so that I can access my organization's maintenance data.

**Layout Description:**
```
+----------------------------------------------------------+
|                     [AMSS Logo]                          |
|              Aircraft Maintenance Scheduling              |
+----------------------------------------------------------+
|                                                          |
|  +----------------------------------------------------+  |
|  |               Welcome Back                         |  |
|  |                                                    |  |
|  |  Email Address                                     |  |
|  |  +----------------------------------------------+  |  |
|  |  |  user@example.com                            |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  Password                                          |  |
|  |  +----------------------------------------------+  |  |
|  |  |  ************            [Show/Hide Toggle] |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  [ ] Remember this device                         |  |
|  |                                                    |  |
|  |  +----------------------------------------------+  |  |
|  |  |              Sign In                         |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  [Forgot Password?]                               |  |
|  +----------------------------------------------------+  |
|                                                          |
|  Language: [English v]     [Privacy] [Terms]            |
+----------------------------------------------------------+
```

**Data Displayed:**
- Organization logo (if custom branding enabled)
- Language selector (i18n support)

**API Endpoints:**
- `POST /api/v1/auth/login` - Submit credentials
- `GET /api/v1/auth/session` - Validate existing session

**Actions Available:**
| Action | All Users |
|--------|-----------|
| Submit login | Yes |
| Toggle password visibility | Yes |
| Navigate to password reset | Yes |
| Change language | Yes |

**Form Fields and Validation:**

| Field | Type | Required | Validation Rules |
|-------|------|----------|------------------|
| Email | TextInput | Yes | Valid email format, max 255 chars |
| Password | TextInput (password) | Yes | Min 8 chars, required |
| Remember Device | Checkbox | No | Boolean |

**State Variations:**

| State | Visual Changes |
|-------|----------------|
| Default | Empty form, Sign In button disabled |
| Valid Input | Sign In button enabled (primary color) |
| Loading | Sign In button shows spinner, form disabled |
| Error - Invalid Credentials | Toast error, password field cleared, focus on password |
| Error - Account Locked | Toast error with lockout duration, link to support |
| Error - Network | Toast error with retry option |
| Success | Redirect to dashboard |

**Offline Behavior:**
- Display offline indicator in header
- Show cached organization branding if available
- Queue login attempt for when connection restores
- Provide clear message: "You are offline. Login requires internet connection."

**i18n Keys:**
- `auth.login.title`
- `auth.login.email_label`
- `auth.login.password_label`
- `auth.login.remember_device`
- `auth.login.submit`
- `auth.login.forgot_password`
- `auth.errors.invalid_credentials`
- `auth.errors.account_locked`

---

### 6.1.2 Password Reset Screen

**URL Pattern:** `/password-reset`

**Purpose:** Allow users to request a password reset link.

**User Story:** As a user who forgot my password, I want to reset it via email so that I can regain access to my account.

**Layout Description:**
```
+----------------------------------------------------------+
|                     [AMSS Logo]                          |
+----------------------------------------------------------+
|                                                          |
|  +----------------------------------------------------+  |
|  |             Reset Your Password                    |  |
|  |                                                    |  |
|  |  Enter your email address and we'll send you      |  |
|  |  instructions to reset your password.              |  |
|  |                                                    |  |
|  |  Email Address                                     |  |
|  |  +----------------------------------------------+  |  |
|  |  |  user@example.com                            |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  +----------------------------------------------+  |  |
|  |  |           Send Reset Link                    |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  [< Back to Login]                                |  |
|  +----------------------------------------------------+  |
|                                                          |
+----------------------------------------------------------+
```

**API Endpoints:**
- `POST /api/v1/auth/password-reset-request`

**Form Fields:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Email | TextInput | Yes | Valid email format |

**State Variations:**
| State | Visual Changes |
|-------|----------------|
| Default | Empty email field |
| Submitting | Button shows spinner |
| Success | Show confirmation message, hide form |
| Error | Toast with error message |

---

### 6.1.3 First-Time Password Change Screen

**URL Pattern:** `/password-change/required`

**Purpose:** Force users with temporary passwords to set a new password.

**User Story:** As a new user with a temporary password, I want to set my own secure password so that my account is properly secured.

**Layout Description:**
```
+----------------------------------------------------------+
|                     [AMSS Logo]                          |
+----------------------------------------------------------+
|                                                          |
|  +----------------------------------------------------+  |
|  |           Set Your New Password                    |  |
|  |                                                    |  |
|  |  [!] Your password must be changed before          |  |
|  |      continuing.                                   |  |
|  |                                                    |  |
|  |  Current Password                                  |  |
|  |  +----------------------------------------------+  |  |
|  |  |  ************                                |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  New Password                                      |  |
|  |  +----------------------------------------------+  |  |
|  |  |  ************                                |  |  |
|  |  +----------------------------------------------+  |  |
|  |  Password Strength: [====    ] Good               |  |
|  |                                                    |  |
|  |  Confirm New Password                              |  |
|  |  +----------------------------------------------+  |  |
|  |  |  ************                                |  |  |
|  |  +----------------------------------------------+  |  |
|  |                                                    |  |
|  |  Password Requirements:                            |  |
|  |  [x] At least 8 characters                        |  |
|  |  [x] Contains uppercase letter                    |  |
|  |  [x] Contains lowercase letter                    |  |
|  |  [x] Contains number                              |  |
|  |  [ ] Contains special character                   |  |
|  |                                                    |  |
|  |  +----------------------------------------------+  |  |
|  |  |           Update Password                    |  |  |
|  |  +----------------------------------------------+  |  |
|  +----------------------------------------------------+  |
|                                                          |
+----------------------------------------------------------+
```

**Form Fields:**
| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Current Password | TextInput (password) | Yes | Must match stored password |
| New Password | TextInput (password) | Yes | Min 8 chars, complexity requirements |
| Confirm Password | TextInput (password) | Yes | Must match new password |

**API Endpoints:**
- `POST /api/v1/auth/password-change`

---

## 6.2 Aircraft Management Screens

### 6.2.1 Aircraft List Screen

**URL Pattern:** `/aircraft`

**Purpose:** Display all aircraft in the organization's fleet with filtering and search capabilities.

**User Story:** As a maintenance planner, I want to see all aircraft in my fleet so that I can quickly find and manage specific aircraft.

**Layout Description:**
```
+----------------------------------------------------------+
| [=] AMSS    Aircraft    Tasks    Programs    Parts   ... |
+----------------------------------------------------------+
| Fleet Management                              [+ Add Aircraft]
+----------------------------------------------------------+
| Search: [___________________] [Search]                    |
|                                                          |
| Filters: [Status: All v] [Model: All v] [Clear Filters]  |
+----------------------------------------------------------+
| Fleet Overview                                           |
| +----------+ +----------+ +----------+ +----------+      |
| | 15       | | 12       | | 2        | | 1        |      |
| | Total    | | Oper.    | | Maint.   | | Grounded |      |
| +----------+ +----------+ +----------+ +----------+      |
+----------------------------------------------------------+
| [ ] | Tail #   | Model       | Status      | Hours    | Cycles | Next Due        |
|-----|----------|-------------|-------------|----------|--------|-----------------|
| [ ] | N12345   | Cessna 172S | Operational | 2,450.5  | 1,800  | Jan 15 (5 days) |
| [ ] | N67890   | Piper PA-28 | Maintenance | 3,100.2  | 2,200  | In Progress     |
| [ ] | N99999   | Beech A36   | Grounded    | 1,500.0  | 1,100  | Overdue         |
+----------------------------------------------------------+
| Showing 1-15 of 15    [< Prev]  [1] [2] [3]  [Next >]    |
+----------------------------------------------------------+
```

**Data Displayed:**
- Summary KPI cards (total, operational, maintenance, grounded)
- Aircraft table with sortable columns

**API Endpoints:**
- `GET /api/v1/aircraft` - List aircraft with filters
- `GET /api/v1/aircraft/stats` - Summary statistics

**Query Parameters:**
```
?status=operational,maintenance,grounded
&model=Cessna 172S
&search=N123
&sort=tail_number:asc
&page=1
&per_page=20
```

**Actions by Role:**

| Action | Admin | Tenant Admin | Scheduler | Mechanic | Auditor |
|--------|-------|--------------|-----------|----------|---------|
| View list | Yes | Yes | Yes | Yes | Yes |
| Add aircraft | Yes | Yes | Yes | No | No |
| Edit aircraft | Yes | Yes | Yes | No | No |
| Delete aircraft | Yes | Yes | No | No | No |
| Export list | Yes | Yes | Yes | No | Yes |
| Bulk select | Yes | Yes | Yes | No | No |

**State Variations:**
| State | Visual Changes |
|-------|----------------|
| Loading | Skeleton loader for table |
| Empty | EmptyState: "No aircraft found" with Add button |
| No Results | EmptyState: "No aircraft match your filters" |
| Error | Error state with retry button |
| Offline | Show cached data with offline badge |

**Status Badge Colors:**
- `operational`: Green (#22C55E)
- `maintenance`: Yellow/Amber (#F59E0B)
- `grounded`: Red (#EF4444)

---

### 6.2.2 Aircraft Detail Screen

**URL Pattern:** `/aircraft/:id`

**Purpose:** Display comprehensive information about a single aircraft including maintenance history.

**User Story:** As a maintenance planner, I want to see all details about an aircraft so that I can make informed maintenance decisions.

**Layout Description:**
```
+----------------------------------------------------------+
| [<] Back to Fleet                                        |
+----------------------------------------------------------+
| N12345 - Cessna 172S                    [Edit] [Actions v]|
| Status: [Operational]                                     |
+----------------------------------------------------------+
|                                                          |
| +----------------------+  +-----------------------------+ |
| | Aircraft Details     |  | Maintenance Counters        | |
| |----------------------|  |-----------------------------| |
| | Tail Number: N12345  |  | Flight Hours: 2,450.5      | |
| | Model: Cessna 172S   |  | Cycles: 1,800              | |
| | Capacity Slots: 2    |  | Last Update: Dec 27, 2025  | |
| | Created: Jun 1, 2024 |  | [Update Hours/Cycles]      | |
| +----------------------+  +-----------------------------+ |
|                                                          |
| [Overview] [Tasks] [Maintenance History] [Programs]      |
+----------------------------------------------------------+
| Upcoming Tasks (3)                                       |
| +------------------------------------------------------+ |
| | 100-Hour Inspection          Due: Jan 15 (5 days)   | |
| | [Scheduled] Assigned: John Doe                       | |
| +------------------------------------------------------+ |
| | Oil Change                   Due: Jan 20 (10 days)  | |
| | [Scheduled] Unassigned                               | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Maintenance Programs (2)                                 |
| +------------------------------------------------------+ |
| | 100-Hour Inspection    Every 100 flight hours       | |
| | Annual Inspection      Every 365 days               | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

**Tabs:**
1. **Overview** - Summary cards and quick stats
2. **Tasks** - All tasks for this aircraft
3. **Maintenance History** - Completed tasks timeline
4. **Programs** - Linked maintenance programs

**API Endpoints:**
- `GET /api/v1/aircraft/:id` - Aircraft details
- `GET /api/v1/aircraft/:id/tasks` - Tasks for aircraft
- `GET /api/v1/aircraft/:id/programs` - Linked programs
- `PATCH /api/v1/aircraft/:id` - Update aircraft
- `DELETE /api/v1/aircraft/:id` - Soft delete

**Actions Dropdown:**
- Change Status
- Update Hours/Cycles
- Ground Aircraft
- Return to Service
- View Audit Log
- Delete Aircraft (admin only)

---

### 6.2.3 Aircraft Create/Edit Form

**URL Pattern:** `/aircraft/new` or `/aircraft/:id/edit`

**Purpose:** Create a new aircraft or edit existing aircraft details.

**User Story:** As a maintenance planner, I want to register a new aircraft so that I can begin tracking its maintenance.

**Layout Description:**
```
+----------------------------------------------------------+
| [<] Back                                                 |
+----------------------------------------------------------+
| Add New Aircraft                           [Cancel] [Save]|
+----------------------------------------------------------+
|                                                          |
| Basic Information                                        |
| +------------------------------------------------------+ |
| | Tail Number *                                        | |
| | +--------------------------------------------------+ | |
| | | N12345                                           | | |
| | +--------------------------------------------------+ | |
| | Must be unique. Example: N12345, G-ABCD            | |
| |                                                      | |
| | Model *                                              | |
| | +--------------------------------------------------+ | |
| | | Cessna 172S                                      | | |
| | +--------------------------------------------------+ | |
| |                                                      | |
| | Capacity Slots *                                     | |
| | +--------------------------------------------------+ | |
| | | 2                                                | | |
| | +--------------------------------------------------+ | |
| | Maximum concurrent maintenance tasks                 | |
| +------------------------------------------------------+ |
|                                                          |
| Maintenance Counters                                     |
| +------------------------------------------------------+ |
| | Current Flight Hours *        Current Cycles *       | |
| | +-----------------------+    +---------------------+ | |
| | | 2450.5                |    | 1800                | | |
| | +-----------------------+    +---------------------+ | |
| |                                                      | |
| | Last Maintenance Date                                | |
| | +--------------------------------------------------+ | |
| | | [Calendar Picker] Dec 15, 2025                   | | |
| | +--------------------------------------------------+ | |
| +------------------------------------------------------+ |
|                                                          |
| Status                                                   |
| +------------------------------------------------------+ |
| | ( ) Operational  ( ) Maintenance  ( ) Grounded      | |
| +------------------------------------------------------+ |
|                                                          |
+----------------------------------------------------------+
```

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Tail Number | TextInput | Yes | Unique, alphanumeric + hyphen, max 10 chars |
| Model | TextInput | Yes | Max 50 chars |
| Capacity Slots | NumberInput | Yes | Integer 1-10 |
| Flight Hours | NumberInput | Yes | Decimal >= 0, max 2 decimals |
| Cycles | NumberInput | Yes | Integer >= 0 |
| Last Maintenance | DatePicker | No | Cannot be future date |
| Status | RadioGroup | Yes | One of: operational, maintenance, grounded |

**Validation Rules:**
- Tail number uniqueness checked on blur (debounced API call)
- Flight hours and cycles cannot decrease (for edit mode)
- Status change may trigger confirmation dialog

**API Endpoints:**
- `POST /api/v1/aircraft` - Create
- `PATCH /api/v1/aircraft/:id` - Update

---

### 6.2.4 Update Hours/Cycles Modal

**URL Pattern:** Modal overlay on `/aircraft/:id`

**Purpose:** Quick update of aircraft flight hours and cycles after daily operations.

**Layout Description:**
```
+----------------------------------------------+
| Update Maintenance Counters          [X]     |
+----------------------------------------------+
| Aircraft: N12345 (Cessna 172S)               |
|                                              |
| Current Values:                              |
| Flight Hours: 2,450.5    Cycles: 1,800       |
|                                              |
| New Values:                                  |
| +--------------------+ +------------------+  |
| | Flight Hours *     | | Cycles *         |  |
| | [2459.0         ]  | | [1804         ]  |  |
| +--------------------+ +------------------+  |
| Change: +8.5 hours       Change: +4 cycles   |
|                                              |
| Update Notes                                 |
| +------------------------------------------+ |
| | Dec 27 operations: 8.5 hours, 4 cycles   | |
| +------------------------------------------+ |
|                                              |
| [!] Values can only increase. This action    |
|     will trigger task generation check.      |
|                                              |
|              [Cancel]  [Update]              |
+----------------------------------------------+
```

**Validation:**
- New hours >= current hours
- New cycles >= current cycles
- Reasonable increment warning (> 100 hours in one update)

---

## 6.3 Maintenance Tasks Screens

### 6.3.1 Task Calendar/Gantt View

**URL Pattern:** `/tasks/calendar`

**Purpose:** Visual timeline view of scheduled maintenance tasks.

**User Story:** As a maintenance planner, I want to see tasks on a timeline so that I can plan maintenance windows and identify conflicts.

**Layout Description:**
```
+----------------------------------------------------------+
| Task Scheduling                        [List View] [+ Task]|
+----------------------------------------------------------+
| [< Dec 2025]  [Today]  [Jan 2026 >]                      |
| View: [Day] [Week] [Month]    Filter: [All Aircraft v]   |
+----------------------------------------------------------+
|        | Mon 27 | Tue 28 | Wed 29 | Thu 30 | Fri 31 |    |
|--------|--------|--------|--------|--------|--------|    |
| N12345 |========|========|        |        |        |    |
|        |100-Hr Inspection |        |        |        |    |
|--------|--------|--------|--------|--------|--------|    |
| N67890 |        |        |==Oil===|        |        |    |
|        |        |        |Change  |        |        |    |
|--------|--------|--------|--------|--------|--------|    |
| N99999 |########|########|########|        |        |    |
|        |  OVERDUE: Annual Inspection       |        |    |
+----------------------------------------------------------+
| Legend: [==] Scheduled  [##] Overdue  [~~] In Progress   |
+----------------------------------------------------------+
```

**Color Coding:**
- Blue: Scheduled tasks
- Yellow/Amber: In Progress
- Green: Completed
- Red: Overdue
- Gray: Cancelled

**Interactions:**
- Click task bar to open detail modal
- Drag task bar to reschedule (scheduler/admin only)
- Hover for quick info tooltip
- Pinch to zoom on tablet

**API Endpoints:**
- `GET /api/v1/tasks/calendar?start_date=2025-12-27&end_date=2026-01-02`

---

### 6.3.2 Task List View

**URL Pattern:** `/tasks`

**Purpose:** Filterable, sortable list of all maintenance tasks.

**User Story:** As a mechanic, I want to see all tasks assigned to me so that I can plan my work day.

**Layout Description:**
```
+----------------------------------------------------------+
| Maintenance Tasks                               [+ Task]  |
+----------------------------------------------------------+
| View: [All Tasks] [My Tasks] [Overdue] [Due This Week]   |
+----------------------------------------------------------+
| Filters:                                                 |
| Status: [All v]  Aircraft: [All v]  Assigned: [All v]   |
| Date Range: [This Week v]                   [Clear All]  |
+----------------------------------------------------------+
| Task Summary                                             |
| +----------+ +----------+ +----------+ +----------+      |
| | 24       | | 3        | | 18       | | 3        |      |
| | Total    | | Overdue  | | Scheduled| | Progress |      |
| +----------+ +----------+ +----------+ +----------+      |
+----------------------------------------------------------+
| [Sort: Due Date v]                      Showing 1-20 of 45|
+----------------------------------------------------------+
| +------------------------------------------------------+ |
| | [!] 100-Hour Inspection - N12345                     | |
| | Due: Dec 25 (2 days overdue)   Status: [OVERDUE]     | |
| | Assigned: John Doe             Program: 100-Hr Insp  | |
| | [View Details]  [Start Task]                         | |
| +------------------------------------------------------+ |
| +------------------------------------------------------+ |
| | Oil Change - N67890                                  | |
| | Due: Dec 28 (1 day)            Status: [Scheduled]   | |
| | Assigned: Unassigned           Program: Oil Change   | |
| | [View Details]  [Assign]                             | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

**Filter Options:**
- Status: All, Scheduled, In Progress, Completed, Cancelled, Overdue
- Aircraft: Dropdown with all aircraft
- Assigned: All, Unassigned, Specific mechanic
- Program: All programs
- Date Range: Today, This Week, This Month, Custom

**API Endpoints:**
- `GET /api/v1/tasks`

**Query Parameters:**
```
?state=scheduled,in_progress
&aircraft_id=uuid
&assigned_mechanic_id=uuid
&program_id=uuid
&due_after=2025-12-27
&due_before=2026-01-03
&sort=due_date:asc
```

---

### 6.3.3 Task Detail Screen

**URL Pattern:** `/tasks/:id`

**Purpose:** Complete task information with parts, compliance, and workflow actions.

**User Story:** As a mechanic, I want to see all task details so that I can complete the work and document it properly.

**Layout Description:**
```
+----------------------------------------------------------+
| [<] Back to Tasks                                        |
+----------------------------------------------------------+
| 100-Hour Inspection - N12345                             |
| Status: [In Progress]                    [Actions v]      |
+----------------------------------------------------------+
| +-------------------------+  +---------------------------+|
| | Task Information        |  | Timeline                  ||
| |-------------------------|  |---------------------------||
| | Aircraft: N12345        |  | Dec 27 09:00 - Started    ||
| | Program: 100-Hr Insp    |  | Dec 27 08:45 - Assigned   ||
| | Type: Inspection        |  | Dec 26 14:00 - Created    ||
| | Scheduled: Dec 27-28    |  |                           ||
| | Assigned: John Doe      |  |                           ||
| +-------------------------+  +---------------------------+|
+----------------------------------------------------------+
| [Details] [Parts] [Compliance] [Notes] [History]         |
+----------------------------------------------------------+
| Parts (3 required, 2 reserved)                           |
| +------------------------------------------------------+ |
| | Part Name          | Status    | Qty | Serial #     | |
| |--------------------|-----------|-----|--------------|  |
| | Oil Filter         | Reserved  | 1   | SN-12345     | |
| | Engine Oil (qt)    | Reserved  | 6   | N/A          | |
| | Spark Plugs        | Pending   | 4   | -            | |
| |                                                      | |
| | [+ Reserve Part]                                     | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Compliance Items (5 items, 3 signed)                     |
| +------------------------------------------------------+ |
| | [x] Torque verified to spec             [Signed]     | |
| | [x] Oil level checked                   [Signed]     | |
| | [x] Filter inspected                    [Signed]     | |
| | [ ] Safety wire installed               [Sign Off]   | |
| | [ ] Test run completed                  [Sign Off]   | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Notes                                                    |
| +------------------------------------------------------+ |
| | [Add Note...]                                        | |
| |                                                      | |
| | John Doe - Dec 27, 10:30 AM                          | |
| | Started inspection. Oil filter shows normal wear.    | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

**Tabs:**
1. **Details** - Basic task information
2. **Parts** - Part reservations and usage
3. **Compliance** - Checklist with sign-off
4. **Notes** - Work notes and photos
5. **History** - Audit trail

**Actions by State and Role:**

| Action | State Required | Admin | Tenant Admin | Scheduler | Mechanic | Auditor |
|--------|----------------|-------|--------------|-----------|----------|---------|
| View | Any | Yes | Yes | Yes | Yes | Yes |
| Edit Details | Scheduled | Yes | Yes | Yes | No | No |
| Assign | Scheduled | Yes | Yes | Yes | No | No |
| Start | Scheduled | Yes | Yes | Yes | If assigned | No |
| Reserve Parts | In Progress | Yes | Yes | Yes | If assigned | No |
| Sign Compliance | In Progress | No | No | No | If assigned | No |
| Add Notes | In Progress | Yes | Yes | Yes | If assigned | No |
| Complete | In Progress | Yes | Yes | Yes | If assigned | No |
| Cancel | Sch/In Prog | Yes | Yes | Yes | No | No |

**State Transition Confirmation Dialogs:**

**Start Task:**
```
+----------------------------------------------+
| Start Task?                          [X]     |
+----------------------------------------------+
| You are about to start this task:            |
|                                              |
| 100-Hour Inspection - N12345                 |
|                                              |
| This will:                                   |
| - Change task status to "In Progress"        |
| - Reserve required parts                     |
| - Record start time in audit log             |
|                                              |
| Verify:                                      |
| [x] Aircraft N12345 is grounded              |
| [x] Required parts are available             |
|                                              |
|              [Cancel]  [Start Task]          |
+----------------------------------------------+
```

**Complete Task:**
```
+----------------------------------------------+
| Complete Task?                       [X]     |
+----------------------------------------------+
| Before completing, verify:                   |
|                                              |
| [x] All parts marked as used or returned     |
| [x] All compliance items signed off          |
| [x] Completion notes entered                 |
|                                              |
| This action cannot be undone.                |
|                                              |
|              [Cancel]  [Complete Task]       |
+----------------------------------------------+
```

---

### 6.3.4 Task Create/Edit Form

**URL Pattern:** `/tasks/new` or `/tasks/:id/edit`

**Purpose:** Create a new maintenance task or edit scheduled task details.

**Layout Description:**
```
+----------------------------------------------------------+
| Create Maintenance Task                   [Cancel] [Save] |
+----------------------------------------------------------+
|                                                          |
| Basic Information                                        |
| +------------------------------------------------------+ |
| | Aircraft *                                           | |
| | +--------------------------------------------------+ | |
| | | [Select Aircraft v]                              | | |
| | +--------------------------------------------------+ | |
| |                                                      | |
| | Task Type *                                          | |
| | ( ) Inspection  ( ) Repair  ( ) Overhaul            | |
| |                                                      | |
| | Program (Optional)                                   | |
| | +--------------------------------------------------+ | |
| | | [Select Program v]                               | | |
| | +--------------------------------------------------+ | |
| +------------------------------------------------------+ |
|                                                          |
| Schedule                                                 |
| +------------------------------------------------------+ |
| | Start Date/Time *          End Date/Time *           | |
| | +---------------------+    +---------------------+   | |
| | | Dec 28, 2025 08:00 |    | Dec 28, 2025 16:00 |   | |
| | +---------------------+    +---------------------+   | |
| |                                                      | |
| | Estimated Duration: 8 hours                          | |
| +------------------------------------------------------+ |
|                                                          |
| Assignment (Optional)                                    |
| +------------------------------------------------------+ |
| | Assign to Mechanic                                   | |
| | +--------------------------------------------------+ | |
| | | [Select Mechanic v]                              | | |
| | +--------------------------------------------------+ | |
| +------------------------------------------------------+ |
|                                                          |
| Notes                                                    |
| +------------------------------------------------------+ |
| | +--------------------------------------------------+ | |
| | |                                                  | | |
| | +--------------------------------------------------+ | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Aircraft | Select | Yes | Valid aircraft ID |
| Task Type | RadioGroup | Yes | inspection, repair, overhaul |
| Program | Select | No | Valid program ID |
| Start Time | DateTimePicker | Yes | Cannot be in past |
| End Time | DateTimePicker | Yes | Must be after start |
| Mechanic | Select | No | Valid user with mechanic role |
| Notes | TextArea | No | Max 2000 chars |

---

## 6.4 Maintenance Programs Screens

### 6.4.1 Programs List Screen

**URL Pattern:** `/programs`

**Purpose:** View and manage maintenance program definitions.

**User Story:** As a maintenance planner, I want to see all maintenance programs so that I can ensure aircraft are covered by appropriate inspection schedules.

**Layout Description:**
```
+----------------------------------------------------------+
| Maintenance Programs                         [+ Program]  |
+----------------------------------------------------------+
| Search: [___________________]                             |
| Filter: [Interval Type: All v]  [Status: Active v]       |
+----------------------------------------------------------+
| Program Name     | Interval        | Aircraft | Tasks    |
|--------------------|-----------------|----------|---------|
| 100-Hour Inspection| 100 flight hrs | 15       | 45      |
| Annual Inspection  | 365 days        | 15       | 15      |
| Oil Change         | 50 flight hrs   | 15       | 120     |
| Engine Overhaul    | 2000 flight hrs | 3        | 2       |
+----------------------------------------------------------+
```

**API Endpoints:**
- `GET /api/v1/programs`
- `POST /api/v1/programs`

---

### 6.4.2 Program Detail Screen

**URL Pattern:** `/programs/:id`

**Purpose:** View program configuration and linked aircraft.

**Layout Description:**
```
+----------------------------------------------------------+
| [<] Back to Programs                                     |
+----------------------------------------------------------+
| 100-Hour Inspection                         [Edit] [...]  |
+----------------------------------------------------------+
| Program Configuration                                    |
| +------------------------------------------------------+ |
| | Name: 100-Hour Inspection                            | |
| | Interval Type: Flight Hours                          | |
| | Interval Value: 100 hours                            | |
| | Created: Jun 1, 2024                                 | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| [Linked Aircraft] [Generated Tasks] [History]            |
+----------------------------------------------------------+
| Linked Aircraft (15)                                     |
| +------------------------------------------------------+ |
| | Tail #   | Model       | Current Hrs | Next Due     | |
| |----------|-------------|-------------|--------------|  |
| | N12345   | Cessna 172S | 2,450       | 2,500 (50h)  | |
| | N67890   | Cessna 172S | 3,100       | 3,200 (100h) | |
| +------------------------------------------------------+ |
| [+ Link Aircraft]  [- Remove Selected]                   |
+----------------------------------------------------------+
```

---

### 6.4.3 Program Create/Edit Form

**URL Pattern:** `/programs/new` or `/programs/:id/edit`

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Name | TextInput | Yes | Unique within org, max 100 chars |
| Description | TextArea | No | Max 500 chars |
| Interval Type | Select | Yes | flight_hours, cycles, calendar |
| Interval Value | NumberInput | Yes | Positive integer |
| Aircraft Selection | MultiSelect | No | Valid aircraft IDs |

**Interval Type Options:**
- **Flight Hours**: Task due every X flight hours
- **Cycles**: Task due every X takeoff/landing cycles
- **Calendar**: Task due every X days

---

## 6.5 Parts Inventory Screens

### 6.5.1 Parts Catalog List

**URL Pattern:** `/parts`

**Purpose:** Browse part definitions in the inventory catalog.

**User Story:** As a parts manager, I want to see all part types in inventory so that I can manage stock levels.

**Layout Description:**
```
+----------------------------------------------------------+
| Parts Inventory                           [+ Part Definition]
+----------------------------------------------------------+
| View: [Catalog] [Items] [Low Stock] [Reservations]       |
+----------------------------------------------------------+
| Search: [___________________]                             |
| Category: [All v]                                        |
+----------------------------------------------------------+
| Part Name                | Category    | In Stock | Reserved|
|--------------------------|-------------|----------|---------|
| Oil Filter - Lycoming    | Filters     | 8        | 2       |
| Spark Plug - Champion    | Ignition    | 24       | 4       |
| Engine Oil (quart)       | Fluids      | 48       | 12      |
| Air Filter               | Filters     | 5        | 0       |
+----------------------------------------------------------+
```

---

### 6.5.2 Part Items List

**URL Pattern:** `/parts/:definitionId/items`

**Purpose:** View individual part items with serial numbers and status.

**Layout Description:**
```
+----------------------------------------------------------+
| [<] Back to Parts                                        |
+----------------------------------------------------------+
| Oil Filter - Lycoming                     [+ Add Items]   |
| Category: Filters                                        |
+----------------------------------------------------------+
| Items (8 in stock, 2 reserved)                           |
+----------------------------------------------------------+
| Serial Number | Status    | Expiry     | Reserved For    |
|---------------|-----------|------------|-----------------|
| SN-12345      | In Stock  | Dec 2026   | -               |
| SN-12346      | Reserved  | Dec 2026   | Task #1234      |
| SN-12347      | In Stock  | Jan 2026   | -               |
| SN-12348      | Used      | -          | Task #1200      |
+----------------------------------------------------------+
```

**Status Badge Colors:**
- `in_stock`: Green
- `reserved`: Yellow/Amber
- `used`: Gray
- `disposed`: Red

---

### 6.5.3 Part Reservation Interface (Within Task)

**URL Pattern:** Modal on `/tasks/:id`

**Purpose:** Reserve parts from inventory for a maintenance task.

**Layout Description:**
```
+----------------------------------------------+
| Reserve Parts                        [X]     |
+----------------------------------------------+
| Task: 100-Hour Inspection - N12345           |
+----------------------------------------------+
| Search Parts: [oil filter___________]        |
+----------------------------------------------+
| Search Results                               |
| +------------------------------------------+ |
| | Oil Filter - Lycoming                    | |
| | Available: 6 items                       | |
| | +--------------------------------------+ | |
| | | SN-12345  | Dec 2026  | [Reserve]    | | |
| | | SN-12346  | Dec 2026  | [Reserve]    | | |
| | | SN-12347  | Jan 2026  | [Reserve]    | | |
| | +--------------------------------------+ | |
| +------------------------------------------+ |
+----------------------------------------------+
| Reserved for This Task                       |
| +------------------------------------------+ |
| | SN-12340 - Oil Filter    [Mark Used]     | |
| |                          [Return]        | |
| +------------------------------------------+ |
+----------------------------------------------+
|                              [Done]          |
+----------------------------------------------+
```

**Actions:**
- **Reserve**: Lock part for this task
- **Mark Used**: Convert reservation to usage (permanent)
- **Return**: Release reservation back to inventory

---

## 6.6 Compliance Management Screens

### 6.6.1 Compliance Checklist Interface (Within Task)

**URL Pattern:** Tab on `/tasks/:id`

**Purpose:** Display compliance items requiring sign-off before task completion.

**Layout Description:**
```
+----------------------------------------------------------+
| Compliance Items                                         |
+----------------------------------------------------------+
| All items must be signed off before completing task.     |
+----------------------------------------------------------+
| Progress: [========    ] 3 of 5 signed                   |
+----------------------------------------------------------+
|                                                          |
| +------------------------------------------------------+ |
| | [x] Torque on drain plug verified to 35 ft-lbs       | |
| |     Result: PASS                                      | |
| |     Signed: John Doe - Dec 27, 10:15 AM              | |
| +------------------------------------------------------+ |
|                                                          |
| +------------------------------------------------------+ |
| | [x] Oil level checked after engine run               | |
| |     Result: PASS                                      | |
| |     Signed: John Doe - Dec 27, 10:20 AM              | |
| +------------------------------------------------------+ |
|                                                          |
| +------------------------------------------------------+ |
| | [ ] Safety wire installed per AC 43.13-1B            | |
| |     Result: [Pending v]                              | |
| |     [Sign Off]                                        | |
| +------------------------------------------------------+ |
|                                                          |
| +------------------------------------------------------+ |
| | [ ] Functional test completed - engine parameters    | |
| |     Result: [Pending v]                              | |
| |     [Sign Off]                                        | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

---

### 6.6.2 Compliance Sign-Off Modal

**URL Pattern:** Modal overlay

**Purpose:** Confirm compliance item result and digital signature.

**Layout Description:**
```
+----------------------------------------------+
| Sign Off Compliance Item             [X]     |
+----------------------------------------------+
| Safety wire installed per AC 43.13-1B        |
|                                              |
| Result: *                                    |
| ( ) Pass - Requirement met                   |
| ( ) Fail - Requirement not met               |
|                                              |
| Notes (optional for pass, required for fail) |
| +------------------------------------------+ |
| |                                          | |
| +------------------------------------------+ |
|                                              |
| Your Digital Signature                       |
| By clicking "Sign Off", you certify that:    |
| - You have personally performed or verified  |
|   the work described above                   |
| - The work meets all regulatory standards    |
|                                              |
| Enter password to confirm:                   |
| +------------------------------------------+ |
| | ************                             | |
| +------------------------------------------+ |
|                                              |
| [!] This sign-off is permanent and cannot    |
|     be changed after submission.             |
|                                              |
|              [Cancel]  [Sign Off]            |
+----------------------------------------------+
```

**Validation:**
- Result must be selected (Pass or Fail)
- Notes required if result is Fail
- Password must match current user's password
- Sign-off is immutable after submission

---

### 6.6.3 Compliance Reports Screen

**URL Pattern:** `/reports/compliance`

**Purpose:** Generate compliance reports for regulatory audits.

**User Story:** As a compliance officer, I want to generate compliance reports so that I can provide documentation for FAA inspections.

**Layout Description:**
```
+----------------------------------------------------------+
| Compliance Reports                           [Generate]   |
+----------------------------------------------------------+
| Report Parameters                                        |
| +------------------------------------------------------+ |
| | Date Range *                                         | |
| | [Jan 1, 2025] to [Dec 31, 2025]                      | |
| |                                                      | |
| | Aircraft                                             | |
| | ( ) All Aircraft  ( ) Specific: [Select v]          | |
| |                                                      | |
| | Include:                                             | |
| | [x] Task completion summary                          | |
| | [x] Compliance item sign-offs                        | |
| | [x] Parts usage and traceability                     | |
| | [x] Audit log summary                                | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Preview                                                  |
| +------------------------------------------------------+ |
| | COMPLIANCE REPORT                                    | |
| | Organization: SkyFlight Services                     | |
| | Period: Jan 1, 2025 - Dec 31, 2025                   | |
| |                                                      | |
| | SUMMARY                                              | |
| | Compliance Rate: 96.8%                               | |
| | Total Tasks: 487                                     | |
| | Completed On Time: 471                               | |
| | Overdue: 12                                          | |
| | ...                                                  | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Export: [PDF] [CSV] [Print]                              |
+----------------------------------------------------------+
```

---

## 6.7 User Management Screens

### 6.7.1 User List Screen

**URL Pattern:** `/users`

**Purpose:** Manage user accounts within the organization.

**User Story:** As an admin, I want to see all users so that I can manage access and roles.

**Layout Description:**
```
+----------------------------------------------------------+
| User Management                               [+ Add User]|
+----------------------------------------------------------+
| Search: [___________________]                             |
| Filter: [Role: All v]  [Status: Active v]                |
+----------------------------------------------------------+
| Name             | Email                | Role      | Status|
|------------------|----------------------|-----------|-------|
| John Doe         | john@example.com     | Mechanic  | Active|
| Jane Smith       | jane@example.com     | Scheduler | Active|
| Bob Johnson      | bob@example.com      | Auditor   | Active|
+----------------------------------------------------------+
```

**Actions by Role:**

| Action | Admin | Tenant Admin | Others |
|--------|-------|--------------|--------|
| View all users | Yes | Yes (own org) | No |
| Create user | Yes | Yes | No |
| Edit user | Yes | Yes (non-admin) | No |
| Delete user | Yes | Yes (non-admin) | No |
| Reset password | Yes | Yes | No |
| Assign admin role | Yes | No | No |

---

### 6.7.2 User Create/Edit Form

**URL Pattern:** `/users/new` or `/users/:id/edit`

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Email | TextInput | Yes | Valid email, unique |
| Role | Select | Yes | admin, tenant_admin, scheduler, mechanic, auditor |
| Password | TextInput | Yes (create) | Min 8 chars, complexity |
| Require Password Change | Checkbox | No | Default: true for new users |

---

### 6.7.3 Password Reset Modal

**URL Pattern:** Modal overlay

**Layout Description:**
```
+----------------------------------------------+
| Reset Password                       [X]     |
+----------------------------------------------+
| User: john@example.com                       |
|                                              |
| Option 1: Generate Random Password           |
| [Generate and Send Email]                    |
|                                              |
| Option 2: Set Custom Password                |
| +------------------------------------------+ |
| | New Password: [____________]             | |
| +------------------------------------------+ |
| [x] Require password change on next login    |
| [Set Password]                               |
|                                              |
+----------------------------------------------+
```

---

## 6.8 Organization Management Screens (Admin Only)

### 6.8.1 Organization List Screen

**URL Pattern:** `/admin/organizations`

**Purpose:** System-wide organization (tenant) management.

**User Story:** As a system administrator, I want to see all organizations so that I can manage the multi-tenant system.

**Layout Description:**
```
+----------------------------------------------------------+
| Organization Management                  [+ Organization] |
+----------------------------------------------------------+
| System Stats                                             |
| +----------+ +----------+ +----------+ +----------+      |
| | 12       | | 11       | | 245      | | 487      |      |
| | Total    | | Active   | | Users    | | Aircraft |      |
| +----------+ +----------+ +----------+ +----------+      |
+----------------------------------------------------------+
| Org Name            | Users | Aircraft | Created    |     |
|---------------------|-------|----------|------------|-----|
| SkyFlight Services  | 45    | 15       | Jun 1, 2024| [...|
| AeroMaintain LLC    | 120   | 42       | Jan 15, 2024|[...|
| NewAirline LLC      | 0     | 0        | Dec 27, 2025|[...|
+----------------------------------------------------------+
```

---

### 6.8.2 Organization Create/Settings Screen

**URL Pattern:** `/admin/organizations/new` or `/admin/organizations/:id/settings`

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| Name | TextInput | Yes | Unique, max 100 chars |
| Timezone | Select | No | Valid IANA timezone |
| Audit Retention (years) | NumberInput | No | 1-10, default 7 |
| Feature Flags | Checkboxes | No | CSV import, webhooks, etc. |

---

## 6.9 Audit Logs Viewer

### 6.9.1 Audit Logs Screen

**URL Pattern:** `/audit-logs`

**Purpose:** View immutable audit trail of all system actions.

**User Story:** As a compliance officer, I want to review audit logs so that I can verify maintenance records for regulatory audits.

**Layout Description:**
```
+----------------------------------------------------------+
| Audit Logs                                    [Export]    |
+----------------------------------------------------------+
| Filters                                                  |
| Date Range: [Last 7 Days v]  Entity: [All v]            |
| Action: [All v]  User: [All v]              [Apply]      |
+----------------------------------------------------------+
| Showing 1-50 of 1,247                                    |
+----------------------------------------------------------+
| Timestamp       | User        | Action  | Entity    | Details|
|-----------------|-------------|---------|-----------|--------|
| Dec 27, 14:35   | John Doe    | Update  | Task      | [View] |
| Dec 27, 14:30   | John Doe    | Create  | Compliance| [View] |
| Dec 27, 10:15   | Jane Smith  | Update  | Aircraft  | [View] |
| Dec 27, 09:00   | System      | Create  | Task      | [View] |
+----------------------------------------------------------+
```

---

### 6.9.2 Audit Log Detail Modal

**URL Pattern:** Modal overlay

**Layout Description:**
```
+----------------------------------------------+
| Audit Log Entry                      [X]     |
+----------------------------------------------+
| ID: audit-uuid-123456                        |
| Timestamp: Dec 27, 2025 14:35:22 UTC         |
|                                              |
| User                                         |
| Name: John Doe                               |
| Role: Mechanic                               |
| ID: user-uuid-789                            |
|                                              |
| Action: state_change                         |
|                                              |
| Entity                                       |
| Type: Task                                   |
| ID: task-uuid-456                            |
|                                              |
| Changes                                      |
| +------------------------------------------+ |
| | Field      | Old Value  | New Value     | |
| |------------|------------|---------------|  |
| | state      | in_progress| completed     | |
| | updated_at | 14:30:00   | 14:35:22      | |
| +------------------------------------------+ |
|                                              |
| Request Details                              |
| IP: 192.168.1.50                             |
| User Agent: Chrome/120.0...                  |
| Request ID: req-uuid-abc                     |
|                                              |
+----------------------------------------------+
```

---

## 6.10 Report Generation Screens

### 6.10.1 Reports Dashboard

**URL Pattern:** `/reports`

**Purpose:** Central hub for generating various reports.

**Layout Description:**
```
+----------------------------------------------------------+
| Reports                                                  |
+----------------------------------------------------------+
| Quick Reports                                            |
| +------------------+ +------------------+ +---------------+|
| | Fleet Status     | | Compliance       | | Cost Analysis ||
| | [Generate]       | | [Generate]       | | [Generate]    ||
| +------------------+ +------------------+ +---------------+|
+----------------------------------------------------------+
| Scheduled Reports                                        |
| +------------------------------------------------------+ |
| | Weekly Fleet Status - Every Monday 8 AM              | |
| | Monthly Compliance - 1st of month                    | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

---

### 6.10.2 Fleet Status Report

**URL Pattern:** `/reports/fleet-status`

**Filters:**
- Date: As of date (default: today)
- Aircraft: All or specific selection
- Include: Aircraft details, upcoming tasks, overdue items

**Report Contents:**
- Fleet summary (operational, maintenance, grounded counts)
- Aircraft breakdown with status and next due
- Upcoming maintenance forecast (7/30 days)
- Overdue items requiring attention

---

### 6.10.3 Cost Analysis Report

**URL Pattern:** `/reports/costs`

**Filters:**
- Date Range: Custom period
- Group By: Aircraft, Program, Month
- Cost Types: Parts, Labor, Total

**Report Contents:**
- Total costs by category
- Cost breakdown charts (pie, bar)
- Cost trends over time (line chart)
- Aircraft-specific cost analysis

---

## 6.11 Webhook Configuration Screens

### 6.11.1 Webhook List Screen

**URL Pattern:** `/settings/webhooks`

**Purpose:** Configure outbound webhook integrations.

**User Story:** As a fleet manager, I want to configure webhooks so that external systems receive real-time updates.

**Layout Description:**
```
+----------------------------------------------------------+
| Webhook Configuration                      [+ Add Webhook]|
+----------------------------------------------------------+
| Endpoint URL              | Events      | Status | Actions|
|---------------------------|-------------|--------|--------|
| https://erp.example.com   | task.*      | Active | [...]  |
| https://notify.example.com| aircraft.*  | Active | [...]  |
+----------------------------------------------------------+
```

---

### 6.11.2 Webhook Create/Edit Form

**URL Pattern:** `/settings/webhooks/new` or `/settings/webhooks/:id/edit`

**Form Fields:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| URL | TextInput | Yes | Valid HTTPS URL |
| Events | MultiSelect | Yes | At least one event |
| Secret | TextInput | No | Auto-generated if blank |
| Active | Switch | No | Default: true |

**Event Options:**
- `task.created`
- `task.state_changed`
- `task.completed`
- `aircraft.grounded`
- `aircraft.hours_updated`
- `compliance.overdue`
- `part.low_stock`

---

### 6.11.3 Webhook Test and Delivery Status

**URL Pattern:** `/settings/webhooks/:id/deliveries`

**Layout Description:**
```
+----------------------------------------------------------+
| Webhook: https://erp.example.com                         |
+----------------------------------------------------------+
| [Test Webhook]                                           |
+----------------------------------------------------------+
| Recent Deliveries                                        |
| +------------------------------------------------------+ |
| | Dec 27, 14:35 | task.completed | Delivered  | 200    | |
| | Dec 27, 14:30 | task.started   | Delivered  | 200    | |
| | Dec 27, 10:15 | aircraft.update| Failed     | 503    | |
| |               |                | Retry: 3/5 |        | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

---

## 6.12 Settings Screens

### 6.12.1 Organization Settings

**URL Pattern:** `/settings/organization`

**Purpose:** Configure organization-level settings.

**Sections:**
- General (name, timezone, date format)
- Data Retention (audit log retention period)
- Feature Flags (enable/disable features)
- Branding (logo, colors - if applicable)

---

### 6.12.2 User Profile Settings

**URL Pattern:** `/settings/profile`

**Purpose:** Personal account settings.

**Layout Description:**
```
+----------------------------------------------------------+
| My Profile                                               |
+----------------------------------------------------------+
| Account Information                                      |
| +------------------------------------------------------+ |
| | Email: john.doe@example.com                          | |
| | Role: Mechanic                                       | |
| | Organization: SkyFlight Services                     | |
| | Member Since: Jun 1, 2024                            | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
| Security                                                 |
| [Change Password]                                        |
+----------------------------------------------------------+
| Preferences                                              |
| +------------------------------------------------------+ |
| | Language: [English (US) v]                           | |
| | Date Format: [MM/DD/YYYY v]                          | |
| | Time Format: [12-hour v]                             | |
| | Timezone: [America/New_York v]                       | |
| +------------------------------------------------------+ |
+----------------------------------------------------------+
```

---


---

[Next: Components & Interactions](05_COMPONENTS_INTERACTIONS.md)
