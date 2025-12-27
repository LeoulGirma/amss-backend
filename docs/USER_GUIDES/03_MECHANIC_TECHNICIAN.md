# AMSS Mechanic/Technician Guide

**Version:** 1.0
**Last Updated:** December 27, 2025
**Applies To:** AMSS v1.x
**Audience:** Licensed A&P Mechanics, Certified Technicians (`mechanic` role)
**Review Date:** March 27, 2026

---

## Table of Contents

### Part I: Orientation
1. [Your Role in Aircraft Safety](#part-i-orientation)
2. [Your Daily Workflow](#your-daily-workflow)
3. [Mental Model: Task Lifecycle](#mental-model-task-lifecycle)

### Part II: Daily Task Execution
4. [Scenario 1: View My Assigned Tasks](#scenario-1-view-my-assigned-tasks)
5. [Scenario 2: Start a Maintenance Task](#scenario-2-start-a-maintenance-task)
6. [Scenario 3: Reserve Required Parts](#scenario-3-reserve-required-parts)
7. [Scenario 4: Complete a Maintenance Task](#scenario-4-complete-a-maintenance-task)
8. [Scenario 5: Return Unused Parts to Inventory](#scenario-5-return-unused-parts-to-inventory)
9. [Scenario 6: Handle Task Errors](#scenario-6-handle-task-errors)

### Part III: Parts Management
10. [Working with Parts Inventory](#part-iii-parts-management)
11. [Parts Traceability and Audit Requirements](#parts-traceability-and-audit-requirements)

### Part IV: Compliance & Documentation
12. [Signing Off on Compliance Items](#part-iv-compliance--documentation)
13. [Adding Work Notes and Photos](#adding-work-notes-and-photos)
14. [Understanding Audit Trails](#understanding-audit-trails)

### Part V: Troubleshooting
15. [Common Errors and Solutions](#part-v-troubleshooting)
16. [When to Contact Your Supervisor](#when-to-contact-your-supervisor)

---

## Part I: Orientation

### Your Role in Aircraft Safety

As a licensed A&P mechanic or certified technician, you are the last line of defense in aircraft safety. The maintenance work you perform directly impacts airworthiness, passenger safety, and regulatory compliance.

**Your Responsibilities:**
- Execute assigned maintenance tasks according to manufacturer procedures and FAA/EASA regulations
- Ensure aircraft are safe to return to service
- Document all work performed with complete accuracy
- Sign off on compliance items verifying work meets regulatory standards
- Use the correct parts and verify proper installation
- Maintain complete audit trails for regulatory inspections

**What AMSS Does for You:**
- Provides a clear list of tasks assigned to you
- Ensures required parts are available before you start work
- Automatically reserves parts when you start a task (prevents double-booking)
- Creates a complete digital record of your work (who, what, when, which parts)
- Provides digital compliance checklists you sign off on
- Generates audit trails for FAA/EASA compliance

**What AMSS Does NOT Do:**
- ❌ AMSS does not replace manufacturer maintenance manuals or FAA-approved procedures
- ❌ AMSS does not make airworthiness determinations (you do)
- ❌ AMSS cannot sign off on your behalf (your digital signature is required)
- ❌ AMSS cannot override safety regulations

**WARNING:** You are legally responsible for the quality and safety of your work. AMSS is a tool to help you document and track maintenance, but it does not replace your professional judgment or regulatory responsibilities.

---

### Your Daily Workflow

A typical day as a mechanic using AMSS:

#### Morning (7:00 AM - 8:00 AM)

**1. Log In and Review Assigned Tasks**
- Open AMSS web interface
- Check "My Assigned Tasks" dashboard
- Review tasks due today (color-coded: green = on-time, yellow = due soon, red = overdue)
- Note which aircraft are scheduled for maintenance

**2. Coordinate with Planners**
- If you see tasks assigned to you that conflict (e.g., two aircraft simultaneously), notify your planner
- Check parts availability for your tasks
- Verify aircraft are available (not flying today)

#### Mid-Morning (8:00 AM - 12:00 PM)

**3. Start First Task**
- Select task from your list (e.g., "100-Hour Inspection - N12345")
- Review task details (what needs to be done, which parts are required)
- Verify aircraft is grounded (AMSS will check this automatically)
- Click "Start Task" → Status changes to "In Progress"
- AMSS automatically reserves required parts from inventory

**4. Perform Maintenance Work**
- Follow manufacturer maintenance manual procedures (not in AMSS - refer to printed manuals)
- Use the parts that AMSS reserved for you
- Take photos of completed work (for documentation)

#### Afternoon (12:00 PM - 4:00 PM)

**5. Sign Off on Compliance Items**
- As you complete each step, sign the digital compliance checklist in AMSS
- Examples: "Torque verified to spec", "AD compliance checked", "Test flight performed"
- Once signed, compliance items are immutable (regulatory requirement)

**6. Complete Task**
- Enter completion notes in AMSS (describe work performed)
- Upload photos (oil filter, torque wrench setting, etc.)
- Mark which parts were used vs returned to inventory
- Click "Complete Task" → Status changes to "Completed"
- AMSS creates audit log entry with your user ID and timestamp

#### End of Day (4:00 PM - 5:00 PM)

**7. Review and Hand Off**
- Check if any tasks remain "In Progress" (should finish before leaving or notify planner)
- Review tomorrow's assigned tasks
- Log out

---

### Mental Model: Task Lifecycle

Understanding the task lifecycle helps you use AMSS effectively:

```
┌──────────────────────────────────────────────────────────┐
│                   TASK LIFECYCLE                          │
└──────────────────────────────────────────────────────────┘

   SCHEDULED                          (Planner creates task or system auto-generates)
       │
       │  ✅ Mechanic clicks "Start Task"
       │  ✅ Aircraft is grounded
       │  ✅ Current time within scheduled window
       │  ✅ Parts available for reservation
       │
       ▼
   IN_PROGRESS                        (You are actively working on it)
       │
       │  ✅ All parts marked "used" or "returned"
       │  ✅ All compliance items signed off
       │  ✅ Completion notes entered
       │  ✅ Photos uploaded (if required)
       │
       ▼
   COMPLETED                          (Task done, audit trail created)
```

**Alternate Path: Cancellation**
```
   SCHEDULED ──────────┐
       │               │
       │               │  ⚠️ Planner or Admin cancels
       ▼               │
   IN_PROGRESS ────────┘
       │
       ▼
   CANCELLED
```

**Key Rules:**

1. **You can only start tasks assigned to you**
   - AMSS enforces this via role-based access control
   - If you try to start someone else's task, you'll get a "Forbidden" error

2. **Aircraft must be grounded to start maintenance**
   - Safety rule: You cannot work on an aircraft that's operational or flying
   - AMSS checks aircraft status before allowing you to start

3. **You cannot complete a task until ALL compliance items are signed off**
   - Regulatory requirement: Complete documentation is mandatory
   - AMSS will prevent completion if any compliance items are unsigned

4. **Once completed, tasks cannot be "un-completed"**
   - Audit integrity: Completed tasks are immutable
   - If you made a mistake, create a new task to correct it

---

## Part II: Daily Task Execution

### Scenario 1: View My Assigned Tasks

**GOAL:** See which maintenance tasks are assigned to you and prioritize your work.

**PRE-CONDITIONS:**
- [ ] You have a valid AMSS account with `mechanic` role
- [ ] You are logged into the AMSS web interface
- [ ] Planner has assigned tasks to you

**HAPPY PATH:**

1. **Navigate to Dashboard**
   - After logging in, you land on the Mechanic Dashboard
   - **System Response:** Dashboard displays "My Assigned Tasks" section

2. **Review Task List**
   - Tasks are sorted by due date (soonest first)
   - **System Response:** Each task shows:
     - Aircraft tail number (e.g., "N12345")
     - Task name (e.g., "100-Hour Inspection")
     - Due date/hours (e.g., "Due at 2500 flight hours" or "Due: Dec 28, 2025")
     - Status indicator:
       - ⚪ White/Scheduled: Not started yet
       - 🔵 Blue/In Progress: You're currently working on it
       - ✅ Green/Completed: Done
       - 🔴 Red: Overdue
     - Parts status: ✅ Reserved | ⚠️ Pending | ❌ Unavailable

3. **Filter Tasks (Optional)**
   - Click filter dropdown
   - Options: "Show only today's tasks", "Show only in-progress", "Show overdue"
   - **System Response:** Task list updates based on filter

4. **View Task Details**
   - Click on a task row
   - **System Response:** Task details page opens showing:
     - Full description of work required
     - Required parts list
     - Compliance checklist
     - Notes from planner
     - Photos/attachments

**VERIFICATION:**
✅ You can see all tasks assigned to you
✅ Tasks are sorted by priority/due date
✅ You can identify which tasks need immediate attention

**COMMON FAILURES:**

| Symptom | Cause | Recovery |
|---------|-------|----------|
| "No tasks shown" | No tasks assigned to you yet | Contact your planner to assign work |
| "Dashboard won't load" | Network connectivity issue | Refresh browser, check internet connection |
| "Tasks show for wrong aircraft" | Viewing another mechanic's dashboard | Verify you're logged in with correct account |

**RELATED SCENARIOS:**
- Next: [Scenario 2: Start a Maintenance Task](#scenario-2-start-a-maintenance-task)
- See also: [Scenario 6: Handle Task Errors](#scenario-6-handle-task-errors)

---

### Scenario 2: Start a Maintenance Task

**GOAL:** Begin work on an assigned maintenance task, which reserves parts and changes task status to "In Progress".

**PRE-CONDITIONS:**
- [ ] Task is assigned to you
- [ ] Task status is `scheduled` (not already in progress)
- [ ] Aircraft is `grounded` (not operational)
- [ ] Current time is within ±5 minutes of scheduled start time (or planner override)
- [ ] Required parts are available in inventory

**HAPPY PATH:**

1. **Select Task from Dashboard**
   - Click on the task you want to start (e.g., "Oil Change - N12345")
   - **System Response:** Task details page opens

2. **Review Task Requirements**
   - Read task description and required work
   - Check compliance checklist items
   - Review required parts list
   - **System Response:** All information displayed clearly

3. **Verify Aircraft Status**
   - Check "Aircraft Status" field
   - **System Response:** Should show "Grounded" (✅ green indicator)
   - If shows "Operational" or "Maintenance" (by another mechanic), see Troubleshooting

4. **Click "Start Task" Button**
   - Button located at top-right of task details page
   - **System Response:** System performs validation:
     - ✅ Checks you are the assigned mechanic
     - ✅ Checks aircraft is grounded
     - ✅ Checks current time is within allowed window
     - ✅ Checks parts availability

5. **Confirm Parts Reservation**
   - Popup appears: "This will reserve the following parts: [list]. Continue?"
   - Click "Confirm"
   - **System Response:**
     - Task status changes to `in_progress` (🔵 blue indicator)
     - Parts status changes to `reserved` (locked for this task)
     - Start timestamp recorded with your user ID
     - Audit log entry created

6. **Begin Physical Work**
   - Proceed to aircraft with reserved parts
   - Perform maintenance according to manufacturer manual
   - (AMSS does not guide the physical work - refer to maintenance manuals)

**VERIFICATION:**
✅ Task status shows "In Progress" with blue indicator
✅ Parts show "Reserved" status
✅ Start timestamp shows current time and your name
✅ No other mechanic can reserve the same parts

**COMMON FAILURES:**

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Error: "Aircraft must be grounded" | Aircraft status is `operational` | Coordinate with operations to ground aircraft, then retry |
| Error: "Parts not available" | Required parts out of stock | Contact planner to order parts or use substitute (with approval) |
| Error: "You are not assigned to this task" | Task assigned to another mechanic | Contact planner to reassign task to you |
| Error: "Too early to start this task" | More than 5 minutes before scheduled start | Wait until scheduled time or ask planner for override |
| Error: "Task already in progress" | You or another mechanic already started it | Check if you have duplicate browser tabs open; if not, contact planner |

**REGULATORY NOTE:** Starting a task creates an audit trail entry per 14 CFR Part 43 requirements. The timestamp and your user ID are permanently recorded.

**RELATED SCENARIOS:**
- Next: [Scenario 3: Reserve Required Parts](#scenario-3-reserve-required-parts) (automatic upon start)
- Next: [Scenario 4: Complete a Maintenance Task](#scenario-4-complete-a-maintenance-task)
- See also: [Troubleshooting - Aircraft Status Issues](#common-errors-and-solutions)

---

### Scenario 3: Reserve Required Parts

**GOAL:** Reserve parts from inventory for a maintenance task to prevent other mechanics from using them.

**NOTE:** Parts reservation usually happens **automatically** when you start a task. This scenario covers manual reservation or adding additional parts mid-task.

**PRE-CONDITIONS:**
- [ ] Task is `in_progress` (you've started it)
- [ ] Parts exist in inventory with `in_stock` status
- [ ] You are the assigned mechanic

**HAPPY PATH:**

1. **Navigate to Task Details Page**
   - Click on your in-progress task
   - **System Response:** Task details page opens

2. **Go to "Parts" Tab**
   - Click "Parts" tab in task details
   - **System Response:** Shows two sections:
     - **Already Reserved:** Parts auto-reserved when you started task
     - **Additional Parts:** Option to reserve more parts if needed

3. **Review Already Reserved Parts**
   - List shows: Part name, Serial number, Status: "Reserved"
   - **System Response:** These parts are locked for this task only

4. **Add Additional Parts (if needed)**
   - Click "+ Add Part"
   - **System Response:** Part search dialog opens

5. **Search for Part**
   - Enter part name or serial number (e.g., "Oil Filter")
   - Click "Search"
   - **System Response:** Shows matching parts in inventory with `in_stock` status

6. **Select Part and Reserve**
   - Click on the part item you want (e.g., "Oil Filter SN-12345")
   - Enter quantity (usually 1)
   - Click "Reserve Part"
   - **System Response:**
     - Part status changes to `reserved`
     - Part added to "Already Reserved" list
     - Distributed lock acquired (prevents double-booking)
     - Audit log entry created

7. **Verify Reservation**
   - Part now shows in "Already Reserved" section
   - **System Response:** Status indicator shows "🔒 Reserved for this task"

**VERIFICATION:**
✅ Part appears in "Already Reserved" list
✅ Part status is `reserved`
✅ Serial number matches the physical part you're using
✅ Other mechanics cannot reserve this part (locked)

**COMMON FAILURES:**

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Error: "Part not available" | Part already reserved by another task | Search for alternate part with different serial number, or wait for release |
| Error: "Insufficient quantity in stock" | Only 0 parts available | Contact planner to order parts |
| Error: "You do not have permission to reserve parts" | You're not assigned to this task | Verify you're working on the correct task |
| Error: "Lock acquisition failed" | Concurrent reservation attempt (rare) | Retry after 5 seconds; if persists, contact system admin |

**WARNING:** When you reserve a part, you are **committed to using it or returning it**. Do not reserve parts "just in case" - only reserve parts you will actually use.

**PARTS TRACEABILITY:**
Once a part is marked "used" (at task completion), it is permanently linked to:
- This task
- This aircraft
- You (the mechanic who used it)
- Timestamp

This traceability is required for FAA/EASA regulations and airworthiness directives (ADs).

**RELATED SCENARIOS:**
- Previous: [Scenario 2: Start a Maintenance Task](#scenario-2-start-a-maintenance-task) (auto-reserves parts)
- Next: [Scenario 4: Complete a Maintenance Task](#scenario-4-complete-a-maintenance-task) (marks parts as used)
- Alternative: [Scenario 5: Return Unused Parts](#scenario-5-return-unused-parts-to-inventory)

---

### Scenario 4: Complete a Maintenance Task

**GOAL:** Record that maintenance work is finished, mark parts as used, sign off on compliance, and return aircraft to service.

**PRE-CONDITIONS:**
- [ ] Task is `in_progress` (you've started it)
- [ ] Physical maintenance work is complete
- [ ] All compliance checklist items are signed off (`pass` or `fail` result, not `pending`)
- [ ] All reserved parts are marked `used` or `returned` (no parts left in `reserved` state)
- [ ] Completion notes are entered (describing work performed)
- [ ] Photos uploaded (if required by organization policy)

**HAPPY PATH:**

1. **Finish Physical Work**
   - Complete all maintenance according to manufacturer manual
   - Verify all work meets airworthiness standards
   - (This happens outside AMSS - at the aircraft)

2. **Navigate to Task Details Page**
   - Open the in-progress task in AMSS
   - **System Response:** Task details page shows status: "In Progress"

3. **Mark Parts as Used**
   - Click "Parts" tab
   - For each reserved part, click "Mark as Used"
   - **System Response:** Part status changes to `used`
   - (Alternatively, if you didn't use a part, click "Return to Inventory" - see Scenario 5)

4. **Sign Off on Compliance Items**
   - Click "Compliance" tab
   - **System Response:** Shows checklist of required compliance items
   - For each item:
     - Select result: `pass` or `fail`
     - If `pass`: Click "Sign Off" → Enter your digital signature (password confirmation)
     - If `fail`: Enter notes explaining failure → Create follow-up task (or contact planner)
   - **System Response:** Each signed item shows ✅ with your name and timestamp
   - **NOTE:** Once signed, compliance items are **immutable** (cannot be changed - regulatory requirement)

5. **Enter Completion Notes**
   - Click "Notes" tab
   - Enter work performed description
   - Example: "Replaced oil filter SN-12345, drained 6 quarts old oil, added 6 quarts fresh oil (AeroShell 15W-50), torqued drain plug to 35 ft-lbs, ran engine for 5 minutes, verified no leaks, oil pressure normal at 60 PSI."
   - **System Response:** Notes saved

6. **Upload Photos (if required)**
   - Click "Photos" tab
   - Click "+ Upload Photo"
   - Select photos from your device (e.g., photo of new oil filter, torque wrench setting)
   - **System Response:** Photos uploaded and attached to task

7. **Verify All Pre-Conditions Met**
   - System shows checklist:
     - ✅ All parts accounted for (used or returned)
     - ✅ All compliance items signed off
     - ✅ Completion notes entered
     - ✅ Photos uploaded (if required)
   - If any are ❌, complete them first

8. **Click "Complete Task" Button**
   - Button at top-right of task details page
   - **System Response:** System performs final validation:
     - ✅ All parts used or returned
     - ✅ All compliance items signed
     - ✅ Completion notes present
     - ✅ Not before scheduled end time (unless override)

9. **Confirm Completion**
   - Popup: "Completing this task will record it as finished and create an audit trail. Continue?"
   - Click "Confirm"
   - **System Response:**
     - Task status changes to `completed` (✅ green indicator)
     - Completion timestamp recorded with your user ID
     - Aircraft maintenance counters updated (if applicable)
     - Audit log entry created (immutable record)
     - Webhook notifications sent (if configured)
     - Next scheduled task auto-generated (if recurring program)

10. **Return Aircraft to Service**
    - Aircraft status may change to `operational` (depending on organization policy)
    - Physical maintenance logbook entry made (per FAA Part 43.9)
    - (AMSS does not replace paper logbook - both are required)

**VERIFICATION:**
✅ Task status shows "Completed" with green ✅ indicator
✅ Completion timestamp shows current time and your name
✅ All parts show "Used" status (or "Returned")
✅ All compliance items show signed ✅ with your name
✅ Audit log entry exists (viewable by compliance officers)

**COMMON FAILURES:**

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Error: "Cannot complete - parts not accounted for" | Some parts still in `reserved` status | Go to Parts tab, mark each part as "Used" or "Return to Inventory" |
| Error: "Cannot complete - compliance items unsigned" | Some compliance items still `pending` | Go to Compliance tab, sign off on all items (pass/fail) |
| Error: "Cannot complete - completion notes required" | No notes entered | Go to Notes tab, enter work description |
| Error: "Cannot complete - too early" | Before scheduled end time | Wait until end time or ask planner for early completion override |
| Error: "You do not have permission to complete this task" | You're not the assigned mechanic | Verify you're working on correct task; contact planner if incorrect assignment |

**CAUTION:** Once you click "Complete Task" and confirm, the task is **permanently completed**. You cannot undo this action. If you made a mistake:
1. Do not try to modify the completed task (you can't - it's immutable)
2. Create a new task to correct the mistake
3. Add notes to the new task referencing the original task

**REGULATORY NOTE:** Task completion creates an audit trail entry satisfying 14 CFR Part 43.9 (maintenance record entry requirements). Your digital signature is legally equivalent to a handwritten signature per FAA Order 8900.1.

**RELATED SCENARIOS:**
- Previous: [Scenario 3: Reserve Required Parts](#scenario-3-reserve-required-parts)
- Alternative: [Scenario 5: Return Unused Parts](#scenario-5-return-unused-parts-to-inventory)
- See also: [Part IV: Compliance & Documentation](#part-iv-compliance--documentation)

---

### Scenario 5: Return Unused Parts to Inventory

**GOAL:** Return parts that were reserved for a task but were not used, making them available for other mechanics.

**WHEN TO USE:**
- You reserved a part but discovered you didn't need it
- You ordered extra parts "just in case" but didn't use all of them
- You completed the task and have leftover parts

**PRE-CONDITIONS:**
- [ ] Task is `in_progress`
- [ ] Parts are in `reserved` status (reserved for your task)
- [ ] Physical part is returned to parts inventory location

**HAPPY PATH:**

1. **Navigate to Task Details Page**
   - Open your in-progress task
   - **System Response:** Task details page opens

2. **Go to "Parts" Tab**
   - Click "Parts" tab
   - **System Response:** Shows "Already Reserved" parts list

3. **Select Part to Return**
   - Find the part you want to return (e.g., "Oil Filter SN-12346")
   - Verify part is `reserved` status (not `used`)
   - **System Response:** Part shows "🔒 Reserved" indicator

4. **Click "Return to Inventory"**
   - Button next to the part row
   - **System Response:** Confirmation dialog appears

5. **Confirm Return**
   - Popup: "This will release the part reservation and return it to inventory. Other mechanics will be able to use it. Continue?"
   - Click "Confirm"
   - **System Response:**
     - Part reservation status changes to `released`
     - Part item status returns to `in_stock`
     - Part removed from your "Already Reserved" list
     - Distributed lock released (other mechanics can now reserve it)
     - Audit log entry created (who released it, when)

6. **Physically Return Part**
   - Take the physical part back to inventory storage
   - (AMSS does not track physical location - only digital status)

**VERIFICATION:**
✅ Part no longer appears in your "Already Reserved" list
✅ Part status is `in_stock` (viewable in Parts Inventory module)
✅ Other mechanics can now reserve this part
✅ Audit log shows you released the part

**COMMON FAILURES:**

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Error: "Cannot return - part already marked used" | You previously marked it as "Used" | Cannot undo - part is permanently used. If mistake, contact planner to correct inventory |
| Error: "You do not have permission to release this part" | Part reserved for different task | Verify you're on correct task |
| Button is grayed out | Part not reserved for this task | Check "Parts" tab - may already be returned or used |

**NOTE:** Returning a part does **not** affect the part's physical condition or expiry date. If the part is damaged or expired, notify your planner or parts manager separately.

**RELATED SCENARIOS:**
- Alternative: [Scenario 4: Complete a Maintenance Task](#scenario-4-complete-a-maintenance-task) (mark part as used)
- See also: [Part III: Parts Management](#part-iii-parts-management)

---

### Scenario 6: Handle Task Errors

**GOAL:** Troubleshoot and resolve common errors you encounter while working with tasks.

**COMMON ERROR 1: "Aircraft must be grounded to start this task"**

**What it means:** You're trying to start a task on an aircraft that's currently `operational` or already undergoing maintenance by another mechanic.

**Why it happens:**
- Aircraft is still marked "operational" in the system (even if physically parked)
- Another mechanic is already working on the aircraft (different task)
- Aircraft status wasn't updated after last flight

**How to fix:**
1. Check aircraft physical status (is it really parked and available?)
2. Coordinate with operations team to update aircraft status to `grounded`
3. If another mechanic is working on it, coordinate to avoid conflicts
4. Once aircraft status is `grounded`, retry starting your task

---

**COMMON ERROR 2: "Parts not available - insufficient quantity in stock"**

**What it means:** The parts required for this task are out of stock.

**Why it happens:**
- Parts were used on other tasks and not replenished
- Parts order delayed
- Incorrect inventory count

**How to fix:**
1. Contact your maintenance planner immediately
2. Planner options:
   - Order emergency parts (if available from supplier today)
   - Use alternate part (requires FAA approval for some parts)
   - Reschedule task to later date
3. Do not proceed with task if parts are unavailable - safety violation
4. Document the delay in task notes

---

**COMMON ERROR 3: "Cannot complete task - compliance items not signed"**

**What it means:** You're trying to complete a task but haven't signed off on all compliance checklist items.

**Why it happens:**
- You forgot to sign some compliance items
- Some items are still `pending` (not marked pass/fail)

**How to fix:**
1. Go to "Compliance" tab
2. Review all compliance items
3. For each item marked `pending`:
   - If work was done correctly: Mark `pass` and sign off
   - If work failed: Mark `fail`, add notes, and contact planner (create follow-up task)
4. Once all items signed, retry completing task

**WARNING:** Do not sign compliance items if the work wasn't actually done or doesn't meet standards. This is a safety and regulatory violation.

---

**COMMON ERROR 4: "You are not assigned to this task"**

**What it means:** You're trying to start or complete a task assigned to another mechanic.

**Why it happens:**
- Planner assigned task to wrong mechanic
- Task was reassigned and you didn't refresh your dashboard
- You're looking at another mechanic's task by mistake

**How to fix:**
1. Verify the task details (check assigned mechanic name)
2. If it should be assigned to you: Contact planner to reassign
3. If it's correctly assigned to someone else: Don't work on it (RBAC violation)
4. Refresh your dashboard to see your actual assigned tasks

---

**COMMON ERROR 5: "Lock acquisition failed - please retry"**

**What it means:** AMSS couldn't acquire a distributed lock on a part (rare concurrent reservation conflict).

**Why it happens:**
- Another mechanic tried to reserve the same part at the exact same moment
- Redis (lock service) temporarily unavailable

**How to fix:**
1. Wait 5 seconds
2. Retry the operation (click "Reserve Part" again)
3. If error persists after 3 retries: Contact system administrator
4. This is usually transient (fixes itself)

---

**WHEN TO CONTACT YOUR SUPERVISOR:**

Contact your maintenance planner or supervisor if:
- ❌ Parts are unavailable and task is urgent
- ❌ Compliance item fails (work doesn't meet standards)
- ❌ You discover additional work required (beyond original task scope)
- ❌ Aircraft has damage or airworthiness issue discovered during maintenance
- ❌ You're unsure if a part is the correct substitute
- ❌ Task deadline cannot be met

Contact your system administrator if:
- ❌ AMSS won't load (blank screen, error messages)
- ❌ Lock errors persist after multiple retries
- ❌ Your account is locked or password reset needed
- ❌ You see data that doesn't make sense (wrong aircraft, duplicate tasks)

**RELATED SCENARIOS:**
- See also: [Part V: Troubleshooting](#part-v-troubleshooting) (complete error reference)

---

## Part III: Parts Management

### Working with Parts Inventory

As a mechanic, you interact with parts inventory daily. Understanding how AMSS tracks parts ensures compliance with FAA/EASA traceability requirements.

#### The Two-Level Parts System

**Part Definition (Catalog Entry):**
- **What it is:** A template or catalog entry for a type of part
- **Example:** "Oil Filter - Lycoming IO-360"
- **Attributes:** Name, category, manufacturer, part number
- **Your interaction:** You search for parts by definition name

**Part Item (Physical Instance):**
- **What it is:** A specific physical part with a unique serial number
- **Example:** "Oil Filter SN-12345" (one instance of the definition above)
- **Attributes:** Serial number, status (`in_stock`, `reserved`, `used`, `disposed`), expiry date
- **Your interaction:** You reserve and use specific part items

#### Parts Status Flow

```
in_stock → reserved (when you start task) → used (when you complete task)
                ↓
            released (if you return unused part) → in_stock
```

#### Viewing Parts Inventory

**To check if a part is available:**

1. Navigate to "Parts Inventory" module (main menu)
2. Search by part name or serial number
3. **System Response:** Shows matching parts with:
   - Part definition name
   - Serial numbers of available items
   - Status (in_stock, reserved, used)
   - Location (if configured)
   - Expiry date (for time-limited parts)

**Status Indicators:**
- ✅ Green `in_stock`: Available for reservation
- 🔒 Yellow `reserved`: Locked for another task
- ❌ Red `used`: Already consumed
- ⏳ Orange: Expiring soon (within 30 days)
- 🚫 Gray `disposed`: Removed from service

---

### Parts Traceability and Audit Requirements

**Why Traceability Matters:**

Aviation regulations (14 CFR Part 43, EASA Part-M) require complete traceability of parts:
- **Which part** was installed (serial number)
- **On which aircraft** (tail number)
- **By whom** (mechanic name/certificate)
- **When** (timestamp)
- **For which task** (task description)

**How AMSS Ensures Traceability:**

When you complete a task and mark parts as "used":
1. Part item status changes to `used` (permanent, cannot be changed back)
2. Part is permanently linked to:
   - Task ID
   - Aircraft ID
   - Your user ID (traceable to your A&P certificate)
   - Completion timestamp
3. Audit log entry created (immutable)

**Viewing Part History:**

To see where a part was used (compliance officers need this for AD compliance):
1. Go to "Parts Inventory" → Search for part by serial number
2. Click on the part item
3. **System Response:** Shows complete history:
   - When it entered inventory (received date)
   - Who reserved it (mechanic name)
   - Which task it was used on (task ID, aircraft)
   - When it was marked "used" (completion timestamp)

**REGULATORY NOTE:** This traceability satisfies 14 CFR Part 43.9 requirements for maintenance records and enables tracking of airworthiness directives (ADs) affecting specific part serial numbers.

---

### Handling Time-Limited Parts (Expiry Dates)

Some aircraft parts have **shelf life limits** (sealants, composites, rubber components, batteries, etc.). AMSS tracks expiry dates to prevent use of expired parts.

#### How Expiry Dates Work

**Part Items with Expiry:**
- Each part item has an optional `expiry_date` field
- Example: "Sealant SN-98765" expires December 31, 2025

**System Behavior:**
- If current date > expiry date: Part cannot be reserved (system blocks it)
- If current date is within 30 days of expiry: Warning shown ("Expiring soon")

#### What You Should Do

**When Reserving Parts:**
1. Check expiry date in parts list (shown in yellow if expiring soon)
2. If part expired: Contact planner to order fresh stock
3. Do not use expired parts (airworthiness violation)

**When Receiving New Parts:**
- Your planner or parts manager enters expiry date when part arrives
- You don't need to update expiry dates (not your responsibility)

**WARNING:** Using an expired time-limited part is a serious safety and regulatory violation. AMSS prevents this by blocking reservation of expired parts, but always verify physical part expiry labels match the system.

---

### Substituting Parts (with Approval)

Sometimes the exact part you need is out of stock, and you must use a substitute.

#### When Substitution is Allowed

**Allowed (with documentation):**
- Manufacturer-approved alternate part numbers
- FAA-approved PMA (Parts Manufacturer Approval) parts
- Same part from different supplier (verify part number matches)

**NOT Allowed (without engineering approval):**
- "Similar" parts that aren't certified alternates
- Parts from non-approved manufacturers
- Used/salvaged parts without proper certification

#### How to Substitute a Part in AMSS

**Pre-Conditions:**
- [ ] You have written approval (email, work order, or engineering sign-off)
- [ ] Substitute part is FAA/EASA approved
- [ ] Substitute part is in AMSS inventory

**Steps:**
1. Contact your maintenance planner or supervisor
2. Provide: Aircraft, task, original part number, substitute part number
3. Planner updates task to specify substitute part
4. You reserve and use the substitute part as normal
5. **Add note to task:** "Used substitute part [number] per planner approval [date]"

**CAUTION:** Never substitute parts without approval. If in doubt, stop work and contact your supervisor.

---

## Part IV: Compliance & Documentation

### Signing Off on Compliance Items

Compliance items are the **digital checklist** that verifies your work meets regulatory and manufacturer standards. Signing off on compliance is legally equivalent to signing a paper logbook.

#### What is a Compliance Item?

**Definition:** A specific requirement that must be verified during maintenance.

**Examples:**
- "Torque on oil drain plug verified to 35 ft-lbs per maintenance manual"
- "AD 2024-12-05 compliance verified (wing spar inspection)"
- "Functional test performed - engine runs smooth, oil pressure 60 PSI"
- "Safety wire installed per AC 43.13-1B"

**Attributes:**
- **Description:** What must be checked
- **Result:** `pending` (not done yet), `pass` (meets standards), `fail` (doesn't meet standards)
- **Sign-Off:** User who signed, timestamp (immutable once signed)

#### How to Sign Off on Compliance Items

**Pre-Conditions:**
- [ ] You have physically performed the work described
- [ ] Work meets manufacturer specifications and FAA/EASA standards
- [ ] You are willing to certify this work (your A&P certificate reputation is on the line)

**Steps:**

1. **Navigate to Task Details → Compliance Tab**
   - Open your in-progress task
   - Click "Compliance" tab
   - **System Response:** Shows checklist of all compliance items

2. **Review Each Compliance Item**
   - Read the description carefully
   - Verify you actually performed this check

3. **Mark Result**
   - If work passes: Select `pass` from dropdown
   - If work fails: Select `fail` from dropdown
   - **System Response:** Result field updates

4. **Sign Off (if Pass)**
   - Click "Sign Off" button
   - **System Response:** Password confirmation dialog appears
   - Enter your AMSS password (digital signature)
   - Click "Confirm"
   - **System Response:**
     - Compliance item marked ✅ signed
     - Your name and timestamp recorded
     - Item becomes **immutable** (cannot be changed)

5. **Handle Failures (if Fail)**
   - Mark result as `fail`
   - Enter detailed notes explaining why it failed
   - Example: "Torque on drain plug only reached 30 ft-lbs, bolt threads damaged"
   - Contact your supervisor immediately
   - Do not complete the task until failure is resolved
   - **System Response:** Creates alert for planner

**VERIFICATION:**
✅ Compliance item shows ✅ with your name
✅ Timestamp shows when you signed
✅ Sign-off is permanent (grayed out, cannot be changed)

#### Rules for Compliance Sign-Offs

**You MUST sign off if:**
- ✅ You personally performed the work or witnessed it
- ✅ Work meets all specifications in the maintenance manual
- ✅ You would stake your A&P certificate on this work

**You MUST NOT sign off if:**
- ❌ Someone else did the work and you didn't verify it
- ❌ Work doesn't meet specifications (torque wrong, parts incorrect, etc.)
- ❌ You're unsure if it meets standards
- ❌ You're being pressured to sign something you didn't do

**WARNING:** Falsely signing off on compliance items is:
- A violation of 14 CFR Part 43.12 (falsification of maintenance records)
- Grounds for FAA certificate suspension or revocation
- Potentially criminal if it leads to an accident

**When in Doubt:** Mark it `fail`, add notes, and contact your supervisor. It's always better to delay a task than to sign off on substandard work.

---

### Adding Work Notes and Photos

Complete documentation is essential for regulatory compliance and troubleshooting future issues.

#### Work Notes Requirements

**What to Include:**
- **Work performed:** Describe what you actually did
  - Example: "Replaced oil filter SN-12345, drained 6 quarts old oil"
- **Parts used:** List parts with serial numbers
  - Example: "Installed new oil filter SN-12345, added 6 quarts AeroShell 15W-50"
- **Measurements:** Record torque values, clearances, pressures
  - Example: "Torqued drain plug to 35 ft-lbs, oil pressure 60 PSI after engine run"
- **Observations:** Note anything unusual
  - Example: "Old filter had minor metal particles (normal for 100-hour interval)"
- **Test results:** Functional tests performed
  - Example: "Engine run 5 minutes, no leaks observed, all parameters normal"

**How to Add Notes:**
1. Go to Task Details → "Notes" tab
2. Click "Add Note"
3. Enter your notes (be detailed and specific)
4. **System Response:** Notes saved with your user ID and timestamp

**Best Practices:**
- ✅ Be specific (not vague)
  - Good: "Torqued to 35 ft-lbs per MM Section 71-00-00"
  - Bad: "Did the work correctly"
- ✅ Use measurements and numbers
- ✅ Reference maintenance manual sections
- ✅ Spell out abbreviations the first time
- ❌ Don't editorialize ("This was a pain" - unprofessional)
- ❌ Don't leave notes blank

---

#### Adding Photos

Photos provide visual evidence of work performed (especially useful for compliance and warranty claims).

**What to Photograph:**
- Completed installations (new parts installed)
- Torque wrench settings (showing correct torque value)
- Removed parts (showing wear patterns, damage)
- Safety wire installations (proving proper technique)
- Inspection findings (cracks, corrosion, wear)
- Test equipment readings (oil pressure gauge, compression tester)

**How to Upload Photos:**
1. Take photos with your phone/camera during or after work
2. Go to Task Details → "Photos" tab
3. Click "+ Upload Photo"
4. Select photo from device
5. Add caption (e.g., "New oil filter SN-12345 installed and safety-wired")
6. **System Response:** Photo uploaded and attached to task

**Best Practices:**
- ✅ Take photos before and after (comparison)
- ✅ Include a reference for scale (if showing damage)
- ✅ Ensure photo is in focus and well-lit
- ✅ Caption each photo clearly
- ❌ Don't upload dozens of redundant photos (keep it relevant)
- ❌ Don't include people's faces (privacy)

**NOTE:** Photos are part of the permanent maintenance record. They may be reviewed by FAA inspectors, warranty claims, or accident investigators.

---

### Understanding Audit Trails

Every action you take in AMSS is logged in an **audit trail**. This is a regulatory requirement and protects you professionally.

#### What Gets Logged

**All of Your Actions:**
- Task started (timestamp, your user ID, aircraft, task ID)
- Parts reserved (which parts, when)
- Parts marked as used (serial numbers, when)
- Compliance items signed off (which items, pass/fail, timestamp)
- Task completed (when, notes added)
- Notes added (what you wrote, when)
- Photos uploaded (when)

**Additional Logged Data:**
- Your IP address (proves you were at work location)
- Request ID (for troubleshooting)
- User agent (which device/browser you used)

#### Why This Matters to You

**Protection for You:**
- If someone questions whether you did the work: Audit log proves it
- If someone claims you didn't sign off: Audit log shows your signature timestamp
- If aircraft has an issue later: Audit log proves you followed procedures

**Protection for the Organization:**
- Regulatory inspections: FAA can see complete work history
- Warranty claims: Prove maintenance was done per manufacturer specs
- Accident investigations: Complete timeline of who did what

#### Viewing Your Audit Trail

**To see your own work history:**
1. Go to "My Profile" → "My Audit Log"
2. **System Response:** Shows all your actions with timestamps

**NOTE:** You can only see **your own** audit logs. Compliance officers and admins can see everyone's logs (they have auditor role).

#### Audit Log Retention

**How long are audit logs kept?**
- Minimum: 7 years (configurable by organization)
- **They are never deleted**, even if you delete a task or leave the company
- Audit logs are **append-only** (cannot be modified or deleted - regulatory requirement)

**REGULATORY NOTE:** 14 CFR Part 91.417 requires maintenance records to be retained for specific periods. AMSS exceeds these requirements by retaining all audit logs permanently.

---

## Part V: Troubleshooting

### Complete Error Reference

Below is a comprehensive list of all error messages you might encounter, with explanations and solutions.

#### Authentication & Access Errors

**Error: "401 Unauthorized - Invalid credentials"**
- **Cause:** Wrong email or password
- **Fix:** Double-check email (case-sensitive), verify password, use "Forgot Password" link if needed

**Error: "401 Unauthorized - Token expired"**
- **Cause:** Your session expired (inactive for >15 minutes)
- **Fix:** Click "Login" and enter credentials again

**Error: "403 Forbidden - You do not have permission"**
- **Cause:** Your role doesn't allow this action (e.g., mechanic trying to create a program)
- **Fix:** Contact planner or admin; you may need a different role

**Error: "403 Forbidden - Not assigned to this task"**
- **Cause:** Task is assigned to another mechanic
- **Fix:** Ask planner to reassign task to you, or work on tasks assigned to you

---

#### Task Operation Errors

**Error: "Aircraft must be grounded to start this task"**
- **Cause:** Aircraft status is `operational` or `maintenance` (by another mechanic)
- **Fix:**
  1. Verify aircraft is physically parked and available
  2. Coordinate with operations to update aircraft status to `grounded`
  3. If another mechanic is working on it, coordinate to avoid conflicts
  4. Retry starting task once aircraft is `grounded`

**Error: "Too early to start this task"**
- **Cause:** Current time is more than 5 minutes before scheduled start time
- **Fix:** Wait until scheduled start time, or ask planner for early start override

**Error: "Task already in progress"**
- **Cause:** You or another mechanic already started this task
- **Fix:** Check if you have duplicate browser tabs open; close extras and refresh

**Error: "Cannot complete task - parts not accounted for"**
- **Cause:** Some reserved parts are still in `reserved` status (not marked `used` or `returned`)
- **Fix:**
  1. Go to "Parts" tab
  2. For each reserved part: Click "Mark as Used" or "Return to Inventory"
  3. Retry completing task

**Error: "Cannot complete task - compliance items not signed"**
- **Cause:** Some compliance items are still `pending` (not marked pass/fail)
- **Fix:**
  1. Go to "Compliance" tab
  2. For each pending item: Mark `pass` or `fail` and sign off
  3. If any fail, contact supervisor before completing
  4. Retry completing task

**Error: "Cannot complete task - completion notes required"**
- **Cause:** Organization policy requires completion notes
- **Fix:** Go to "Notes" tab, add detailed notes describing work performed

**Error: "Cannot complete task - too early"**
- **Cause:** Before scheduled end time
- **Fix:** Wait until scheduled end time, or ask planner for early completion override

---

#### Parts Errors

**Error: "Part not available - insufficient quantity in stock"**
- **Cause:** Part is out of stock or all items are reserved for other tasks
- **Fix:**
  1. Contact maintenance planner immediately
  2. Planner will order parts or find substitute
  3. Do not proceed with task if critical parts unavailable

**Error: "Lock acquisition failed - please retry"**
- **Cause:** Rare concurrent reservation conflict (another mechanic tried to reserve same part simultaneously)
- **Fix:**
  1. Wait 5 seconds
  2. Click "Reserve Part" again
  3. If persists after 3 retries: Contact system administrator

**Error: "Part has expired - cannot reserve"**
- **Cause:** Part item expiry date has passed
- **Fix:** Contact planner to order fresh stock; do not use expired parts

**Error: "Cannot return part - already marked as used"**
- **Cause:** You previously marked this part "used"
- **Fix:** Cannot undo; if this was a mistake, contact planner to correct inventory

---

#### Validation Errors

**Error: "Invalid input: [field] is required"**
- **Cause:** You left a required field blank
- **Fix:** Fill in the missing field and retry

**Error: "Invalid input: end_time must be after start_time"**
- **Cause:** Logical error in timestamps
- **Fix:** Check start/end times, correct the error

**Error: "Conflict: Duplicate task already exists"**
- **Cause:** A task already exists for this aircraft/program combination
- **Fix:** Check existing tasks; if duplicate, contact planner to resolve

---

#### System Errors

**Error: "500 Internal Server Error"**
- **Cause:** Unexpected server error
- **Fix:**
  1. Refresh the page (Ctrl+R or Cmd+R)
  2. If persists: Contact system administrator with error details

**Error: "503 Service Unavailable"**
- **Cause:** AMSS server is down or under maintenance
- **Fix:**
  1. Wait 5 minutes and retry
  2. If persists: Contact system administrator

**Error: "Network error - cannot connect to server"**
- **Cause:** Your internet connection is down or firewall blocking
- **Fix:**
  1. Check your Wi-Fi/ethernet connection
  2. Try accessing another website to verify internet works
  3. Contact IT if internet is working but AMSS won't load

---

### Performance Issues

#### Slow Page Loading

**Symptoms:**
- Pages take >10 seconds to load
- Spinning "loading" indicator stays for long time

**Causes & Fixes:**
1. **Slow internet connection**
   - Check your connection speed
   - Close bandwidth-heavy applications (video streaming)
   - Move closer to Wi-Fi router or use ethernet

2. **Too many browser tabs open**
   - Close unused tabs
   - Restart browser

3. **Browser cache full**
   - Clear browser cache:
     - Chrome: Ctrl+Shift+Del → Clear cache
     - Firefox: Ctrl+Shift+Del → Clear cache
     - Safari: Cmd+Option+E

4. **Server load (many users)**
   - Use AMSS during off-peak hours if possible
   - Contact system admin if persistent

---

#### Page Freezes or Crashes

**Symptoms:**
- Page stops responding
- Browser shows "Page Unresponsive" message

**Causes & Fixes:**
1. **Browser memory issue**
   - Close browser completely
   - Restart browser
   - Open AMSS in fresh tab

2. **Large file upload (photos)**
   - Reduce photo file sizes before uploading
   - Upload photos one at a time (not batch upload)

3. **Browser compatibility**
   - Use supported browsers: Chrome 90+, Firefox 88+, Safari 14+
   - Update browser to latest version

---

### When to Contact Support

#### Contact Your Maintenance Planner/Supervisor If:

- ❌ Parts are unavailable and task is urgent (safety issue)
- ❌ Compliance item fails (work doesn't meet standards)
- ❌ You discover additional work required (beyond original task scope)
- ❌ Aircraft has damage or airworthiness issue discovered during maintenance
- ❌ You're unsure if a part is the correct substitute
- ❌ Task deadline cannot be met due to external factors
- ❌ You need task reassignment or schedule change

#### Contact Your System Administrator If:

- ❌ AMSS won't load (blank screen, constant errors)
- ❌ Login fails repeatedly with correct password
- ❌ Lock errors persist after multiple retries
- ❌ Your account is locked ("Too many login attempts")
- ❌ You see data that doesn't make sense (wrong aircraft, duplicate tasks, missing tasks)
- ❌ System performance is extremely slow (pages timeout)
- ❌ Photos won't upload after multiple attempts

#### Provide This Information When Reporting Issues:

To help support resolve your issue quickly:
1. **Your name and role** (e.g., "John Doe, mechanic")
2. **Exact error message** (screenshot if possible)
3. **What you were trying to do** (e.g., "Complete oil change task for N12345")
4. **When it happened** (timestamp)
5. **Browser and device** (e.g., "Chrome on Windows laptop")
6. **Steps to reproduce** (what you clicked, in order)

**Example:**
"Hi, I'm trying to complete Task #12345 (Oil Change - N12345) but getting error 'Cannot complete - parts not accounted for'. I've marked all 3 parts as used in the Parts tab, but error persists. This happened at 10:30 AM on Dec 27. Using Chrome on my work laptop. Screenshot attached."

---

## Summary

You've completed the Mechanic/Technician User Guide. You should now understand:

✅ **Your role** in aircraft safety and regulatory compliance
✅ **Daily workflow** from viewing tasks to completing them
✅ **Task lifecycle** (scheduled → in_progress → completed)
✅ **Parts management** (reservation, use, traceability)
✅ **Compliance sign-offs** (how to sign, when to sign, legal implications)
✅ **Documentation** (notes, photos, audit trails)
✅ **Troubleshooting** (errors, performance issues, when to get help)

**Key Takeaways:**
- Always verify aircraft is grounded before starting maintenance
- Never sign off on work you didn't perform or verify
- All your actions are logged in an immutable audit trail
- When in doubt, contact your supervisor before proceeding

**Related Guides:**
- [Orientation Guide](00_ORIENTATION.md) - Basic concepts refresher
- [Maintenance Planner Guide](02_MAINTENANCE_PLANNER.md) - Understand how tasks are created
- [Quick Reference](06_QUICK_REFERENCE.md) - Cheat sheets and FAQs

**Regulatory References:**
- 14 CFR Part 43 - Maintenance, Preventive Maintenance, Rebuilding, and Alteration
- 14 CFR Part 91.417 - Maintenance Records
- FAA Advisory Circular 43.13-1B - Acceptable Methods, Techniques, and Practices
- EASA Part-M - Continuing Airworthiness Management

---

**Questions or Feedback?**
Contact your maintenance planner or system administrator.

**Document Version:** 1.0 (Complete)
**Last Updated:** December 27, 2025
**Next Review:** March 27, 2026
