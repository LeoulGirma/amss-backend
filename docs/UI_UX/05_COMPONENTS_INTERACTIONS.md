# AMSS UI/UX Specification - Components & Interactions

[Back to Index](00_INDEX.md) | [Previous: Screen Specifications](04_SCREENS.md)

---

# Part 7: Component Library

This section specifies reusable UI components for the AMSS application.

---

## 7.1 Layout Components

### 7.1.1 AppShell

**Purpose:** Main application wrapper providing consistent layout structure.

**Structure:**
```
+----------------------------------------------------------+
| Header (64px)                                            |
+----------+-----------------------------------------------+
|          |                                               |
| Sidebar  | Content Area                                  |
| (240px)  |                                               |
|          |                                               |
|          |                                               |
|          |                                               |
|          |                                               |
+----------+-----------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| sidebarCollapsed | boolean | false | Collapsed state of sidebar |
| onSidebarToggle | function | - | Callback when sidebar toggled |
| showOfflineIndicator | boolean | false | Show offline status |

**Responsive Behavior:**
- Desktop (>1024px): Full sidebar, fixed header
- Tablet (768-1024px): Collapsible sidebar, hamburger menu
- Mobile (<768px): Hidden sidebar, bottom navigation (not primary target)

**CSS Classes:**
```css
.app-shell {
  display: grid;
  grid-template-columns: auto 1fr;
  grid-template-rows: 64px 1fr;
  min-height: 100vh;
}

.app-shell--sidebar-collapsed {
  grid-template-columns: 64px 1fr;
}
```

---

### 7.1.2 Sidebar Navigation

**Purpose:** Primary navigation for the application.

**Structure:**
```
+------------------+
| [Logo]  AMSS [X] |
+------------------+
| Dashboard        |
| Aircraft         |
| Tasks            |
| Programs         |
| Parts            |
| Compliance       |
+------------------+
| Reports          |
| Settings         |
+------------------+
| [User Avatar]    |
| John Doe         |
| [Logout]         |
+------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| collapsed | boolean | false | Show icons only |
| activeItem | string | - | Currently active nav item |
| userRole | Role | - | Current user role for filtering |
| onCollapse | function | - | Collapse toggle callback |

**Navigation Items by Role:**

| Item | Admin | Tenant Admin | Scheduler | Mechanic | Auditor |
|------|-------|--------------|-----------|----------|---------|
| Dashboard | Yes | Yes | Yes | Yes | Yes |
| Aircraft | Yes | Yes | Yes | Yes | Yes |
| Tasks | Yes | Yes | Yes | Yes | Yes |
| Programs | Yes | Yes | Yes | No | Yes |
| Parts | Yes | Yes | Yes | Yes | Yes |
| Compliance | Yes | Yes | Yes | No | Yes |
| Reports | Yes | Yes | Yes | No | Yes |
| Users | Yes | Yes | No | No | No |
| Organizations | Yes | No | No | No | No |
| Audit Logs | Yes | Yes | No | No | Yes |
| Settings | Yes | Yes | Yes | Yes | Yes |

**Collapsed State:**
- Show icons only (24px)
- Tooltip on hover with full label
- Width: 64px

**Expanded State:**
- Icons with labels
- Width: 240px
- Smooth transition: 200ms ease-out

---

### 7.1.3 Header

**Purpose:** Top navigation bar with organization context and user actions.

**Structure:**
```
+----------------------------------------------------------+
| [=] AMSS    [Org: SkyFlight v]    [Bell] [?] [Avatar v]  |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| organizationName | string | - | Current organization |
| organizations | array | - | For admin org switcher |
| user | object | - | Current user info |
| notificationCount | number | 0 | Unread notification count |
| onOrgChange | function | - | Org switch callback (admin only) |
| onMenuToggle | function | - | Sidebar toggle callback |

**Elements:**
1. **Menu Toggle** (hamburger icon) - Toggle sidebar
2. **Logo** - Navigate to dashboard
3. **Organization Switcher** (admin only) - Switch between organizations
4. **Search** (optional) - Global search
5. **Notifications** - Bell icon with badge
6. **Help** - Link to documentation
7. **User Menu** - Profile, settings, logout

**User Menu Dropdown:**
```
+------------------------+
| John Doe               |
| john@example.com       |
| Role: Mechanic         |
+------------------------+
| Profile Settings       |
| Change Password        |
+------------------------+
| Logout                 |
+------------------------+
```

---

### 7.1.4 PageHeader

**Purpose:** Consistent page title area with breadcrumbs and actions.

**Structure:**
```
+----------------------------------------------------------+
| Home / Aircraft / N12345                                 |
+----------------------------------------------------------+
| N12345 - Cessna 172S                    [Edit] [Actions] |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| title | string | - | Page title |
| subtitle | string | - | Optional subtitle |
| breadcrumbs | array | - | Navigation path |
| actions | ReactNode | - | Action buttons |

**Breadcrumbs Format:**
```javascript
breadcrumbs={[
  { label: 'Home', href: '/' },
  { label: 'Aircraft', href: '/aircraft' },
  { label: 'N12345' }  // Current page (no href)
]}
```

---

### 7.1.5 ContentCard

**Purpose:** Standard container for content sections.

**Structure:**
```
+----------------------------------------------------------+
| Card Title                               [Card Actions]  |
+----------------------------------------------------------+
|                                                          |
| Card content goes here                                   |
|                                                          |
+----------------------------------------------------------+
| Card Footer (optional)                                   |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| title | string | - | Card header title |
| subtitle | string | - | Optional subtitle |
| actions | ReactNode | - | Header action buttons |
| footer | ReactNode | - | Footer content |
| padding | string | 'md' | Content padding (none, sm, md, lg) |
| loading | boolean | false | Show skeleton loader |

**CSS:**
```css
.content-card {
  background: var(--surface-color);
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.content-card__header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.content-card__body {
  padding: 20px;
}
```

---

## 7.2 Data Display Components

### 7.2.1 DataTable

**Purpose:** Sortable, filterable, paginated table for data lists.

**Structure:**
```
+----------------------------------------------------------+
| [Search]                              [Columns] [Export] |
+----------------------------------------------------------+
| [ ] | Col A ^    | Col B      | Col C      | Actions    |
|-----|------------|------------|------------|------------|
| [ ] | Value 1    | Value 2    | Value 3    | [...]      |
| [ ] | Value 4    | Value 5    | Value 6    | [...]      |
| [x] | Value 7    | Value 8    | Value 9    | [...]      |
+----------------------------------------------------------+
| Selected: 1                Showing 1-20 of 100  [< 1 2 >]|
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| columns | ColumnDef[] | - | Column definitions |
| data | array | - | Row data |
| loading | boolean | false | Show loading state |
| selectable | boolean | false | Enable row selection |
| selectedRows | array | [] | Currently selected rows |
| onSelectionChange | function | - | Selection callback |
| sortable | boolean | true | Enable column sorting |
| sortBy | object | - | Current sort state |
| onSortChange | function | - | Sort callback |
| pagination | object | - | Pagination state |
| onPageChange | function | - | Page change callback |
| onRowClick | function | - | Row click callback |
| emptyState | ReactNode | - | Empty state component |
| rowActions | function | - | Row action menu renderer |

**Column Definition:**
```typescript
interface ColumnDef {
  id: string;
  header: string;
  accessor: string | ((row) => any);
  sortable?: boolean;
  width?: string;
  align?: 'left' | 'center' | 'right';
  render?: (value, row) => ReactNode;
}
```

**Example Usage:**
```jsx
<DataTable
  columns={[
    { id: 'tail', header: 'Tail #', accessor: 'tailNumber', sortable: true },
    { id: 'model', header: 'Model', accessor: 'model' },
    { id: 'status', header: 'Status', accessor: 'status',
      render: (value) => <StatusBadge status={value} /> },
    { id: 'hours', header: 'Hours', accessor: 'flightHoursTotal', align: 'right' }
  ]}
  data={aircraft}
  selectable
  pagination={{ page: 1, perPage: 20, total: 100 }}
/>
```

---

### 7.2.2 StatusBadge

**Purpose:** Color-coded status indicator for various entity states.

**Variants:**

**Aircraft Status:**
| Status | Color | Icon |
|--------|-------|------|
| operational | Green (#22C55E) | Checkmark |
| maintenance | Amber (#F59E0B) | Wrench |
| grounded | Red (#EF4444) | X-Circle |

**Task Status:**
| Status | Color | Icon |
|--------|-------|------|
| scheduled | Blue (#3B82F6) | Calendar |
| in_progress | Amber (#F59E0B) | Play |
| completed | Green (#22C55E) | Check |
| cancelled | Gray (#6B7280) | X |
| overdue | Red (#EF4444) | Alert |

**Part Status:**
| Status | Color | Icon |
|--------|-------|------|
| in_stock | Green (#22C55E) | Check |
| reserved | Amber (#F59E0B) | Lock |
| used | Gray (#6B7280) | Used |
| disposed | Red (#EF4444) | Trash |

**Compliance Status:**
| Status | Color | Icon |
|--------|-------|------|
| pass | Green (#22C55E) | Check |
| fail | Red (#EF4444) | X |
| pending | Gray (#6B7280) | Clock |

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| status | string | - | Status value |
| type | string | - | Entity type (aircraft, task, part, compliance) |
| size | string | 'md' | Badge size (sm, md, lg) |
| showIcon | boolean | true | Show status icon |
| showLabel | boolean | true | Show status text |

**CSS:**
```css
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 9999px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge--operational {
  background: #DCFCE7;
  color: #166534;
}

.status-badge--maintenance {
  background: #FEF3C7;
  color: #92400E;
}

.status-badge--grounded {
  background: #FEE2E2;
  color: #991B1B;
}
```

---

### 7.2.3 KPICard

**Purpose:** Display key performance indicators with trend information.

**Structure:**
```
+---------------------------+
| Aircraft Operational      |
| [Plane Icon]              |
|                           |
| 12                        |
| +2 from last week         |
| [===== Sparkline =====]   |
+---------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| title | string | - | Metric name |
| value | number/string | - | Current value |
| previousValue | number | - | For trend calculation |
| trend | 'up' / 'down' / 'flat' | - | Override trend direction |
| trendLabel | string | - | Trend description |
| icon | ReactNode | - | Metric icon |
| sparklineData | number[] | - | Data for mini chart |
| loading | boolean | false | Show skeleton |
| onClick | function | - | Card click handler |

**Trend Colors:**
- Up (positive): Green
- Down (negative): Red
- Flat: Gray

---

### 7.2.4 Timeline

**Purpose:** Chronological display of events (audit logs, task history).

**Structure:**
```
+----------------------------------------------------------+
| Dec 27, 2025                                             |
|  |                                                       |
|  o-- 14:35  Task completed by John Doe                   |
|  |         100-Hour Inspection - N12345                  |
|  |                                                       |
|  o-- 10:15  Task started by John Doe                     |
|  |         Parts reserved: Oil Filter (SN-12345)         |
|  |                                                       |
| Dec 26, 2025                                             |
|  |                                                       |
|  o-- 16:00  Task assigned to John Doe                    |
|  |         by Jane Smith (Scheduler)                     |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| events | TimelineEvent[] | - | Event list |
| groupByDate | boolean | true | Group events by date |
| maxItems | number | 50 | Maximum items to show |
| onLoadMore | function | - | Infinite scroll callback |

**TimelineEvent:**
```typescript
interface TimelineEvent {
  id: string;
  timestamp: Date;
  title: string;
  description?: string;
  icon?: string;
  color?: string;
  user?: { name: string; avatar?: string };
}
```

---

### 7.2.5 AircraftCard

**Purpose:** Compact aircraft summary for grid/list views.

**Structure:**
```
+--------------------------------+
| [Plane Icon]         [Status] |
| N12345                        |
| Cessna 172S                   |
|                               |
| Hours: 2,450.5  Cycles: 1,800 |
|                               |
| Next Due: Jan 15 (5 days)     |
| 100-Hour Inspection           |
+--------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| aircraft | Aircraft | - | Aircraft data |
| showNextDue | boolean | true | Show next maintenance |
| onClick | function | - | Click handler |
| selected | boolean | false | Selection state |

---

### 7.2.6 TaskCard

**Purpose:** Task summary for dashboards and lists.

**Structure:**
```
+------------------------------------------+
| [Wrench] 100-Hour Inspection    [Status] |
| Aircraft: N12345                         |
|                                          |
| Due: Dec 28, 2025 (1 day)               |
| Assigned: John Doe                       |
|                                          |
| Program: 100-Hour Inspection             |
| Parts: 3 required, 2 reserved            |
+------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| task | MaintenanceTask | - | Task data |
| showAircraft | boolean | true | Show aircraft info |
| showAssignment | boolean | true | Show mechanic |
| showParts | boolean | false | Show parts status |
| actions | ReactNode | - | Quick action buttons |
| onClick | function | - | Click handler |

**Visual States:**
- Normal: Default styling
- Overdue: Red left border, warning icon
- In Progress: Amber left border
- Due Today/Tomorrow: Yellow highlight

---

## 7.3 Form Components

### 7.3.1 TextInput

**Purpose:** Single-line text input with validation.

**Structure:**
```
Label *
+------------------------------------------+
| Placeholder text                         |
+------------------------------------------+
Helper text or error message
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| value | string | - | Input value |
| onChange | function | - | Change handler |
| placeholder | string | - | Placeholder text |
| helperText | string | - | Helper/hint text |
| error | string | - | Error message |
| required | boolean | false | Required indicator |
| disabled | boolean | false | Disabled state |
| maxLength | number | - | Character limit |
| type | string | 'text' | Input type |
| leftIcon | ReactNode | - | Icon before input |
| rightIcon | ReactNode | - | Icon after input |

**States:**
- Default: Gray border
- Focus: Primary color border, subtle shadow
- Error: Red border, error message shown
- Disabled: Grayed out, not interactive

**i18n:**
- Labels and placeholders support translation keys
- Error messages from validation library

---

### 7.3.2 NumberInput

**Purpose:** Numeric input with increment/decrement controls.

**Structure:**
```
Flight Hours *
+------------------------------------------+
| [-]    2450.5                       [+] |
+------------------------------------------+
Enter total flight hours
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| value | number | - | Input value |
| onChange | function | - | Change handler |
| min | number | - | Minimum value |
| max | number | - | Maximum value |
| step | number | 1 | Increment step |
| precision | number | 0 | Decimal places |
| showControls | boolean | true | Show +/- buttons |
| error | string | - | Error message |

---

### 7.3.3 Select

**Purpose:** Dropdown selection with search and filtering.

**Structure:**
```
Aircraft *
+------------------------------------------+
| Select Aircraft                      [v] |
+------------------------------------------+
    +--------------------------------------+
    | Search...                            |
    +--------------------------------------+
    | N12345 - Cessna 172S                 |
    | N67890 - Piper PA-28                 |
    | N99999 - Beech A36                   |
    +--------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| value | any | - | Selected value |
| onChange | function | - | Change handler |
| options | Option[] | - | Available options |
| placeholder | string | 'Select...' | Placeholder |
| searchable | boolean | true | Enable search |
| clearable | boolean | false | Show clear button |
| loading | boolean | false | Loading state |
| error | string | - | Error message |

**Option Format:**
```typescript
interface Option {
  value: string | number;
  label: string;
  icon?: ReactNode;
  disabled?: boolean;
  group?: string;  // For grouped options
}
```

---

### 7.3.4 MultiSelect

**Purpose:** Multiple item selection with chips display.

**Structure:**
```
Events to Subscribe *
+------------------------------------------+
| [task.completed x] [aircraft.* x]   [v] |
+------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| value | any[] | [] | Selected values |
| onChange | function | - | Change handler |
| options | Option[] | - | Available options |
| maxSelections | number | - | Limit selections |
| searchable | boolean | true | Enable search |

---

### 7.3.5 DatePicker

**Purpose:** Date selection with calendar popup.

**Structure:**
```
Start Date *
+------------------------------------------+
| [Calendar] Dec 28, 2025             [x] |
+------------------------------------------+
    +--------------------------------------+
    |     < December 2025 >                |
    | Su Mo Tu We Th Fr Sa                 |
    |        1  2  3  4  5  6              |
    |  7  8  9 10 11 12 13                 |
    | 14 15 16 17 18 19 20                 |
    | 21 22 23 24 25 26 27                 |
    | [28] 29 30 31                        |
    +--------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| value | Date | - | Selected date |
| onChange | function | - | Change handler |
| minDate | Date | - | Minimum selectable |
| maxDate | Date | - | Maximum selectable |
| disabledDates | Date[] | - | Disabled dates |
| format | string | 'MMM D, YYYY' | Display format |
| clearable | boolean | true | Show clear button |

**i18n:**
- Month/day names localized
- Date format follows locale settings

---

### 7.3.6 DateTimePicker

**Purpose:** Combined date and time selection.

**Structure:**
```
Start Time *
+------------------------------------------+
| Dec 28, 2025 at 08:00 AM            [x] |
+------------------------------------------+
```

**Additional Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| timeFormat | '12h' / '24h' | '12h' | Time format |
| minuteStep | number | 15 | Minute increment |

---

### 7.3.7 DateRangePicker

**Purpose:** Select date range for filtering and reports.

**Structure:**
```
Date Range
+------------------------------------------+
| Dec 1, 2025 - Dec 31, 2025          [x] |
+------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| startDate | Date | - | Range start |
| endDate | Date | - | Range end |
| onChange | function | - | Change handler |
| presets | Preset[] | - | Quick select presets |

**Presets:**
- Today
- Yesterday
- Last 7 days
- Last 30 days
- This month
- Last month
- This year
- Custom range

---

### 7.3.8 FileUpload

**Purpose:** File upload with drag-and-drop support.

**Structure:**
```
Upload CSV File
+------------------------------------------+
|                                          |
|    [Cloud Upload Icon]                   |
|                                          |
|    Drag and drop your file here          |
|    or click to browse                    |
|                                          |
|    Accepted: .csv (max 10MB)             |
|                                          |
+------------------------------------------+
Uploaded: fleet_data.csv (2.4 MB) [x]
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| accept | string[] | - | Accepted file types |
| maxSize | number | 10MB | Max file size |
| multiple | boolean | false | Allow multiple |
| onUpload | function | - | Upload handler |
| onRemove | function | - | Remove handler |

**Use Cases:**
- CSV import for bulk data
- Photo upload for task documentation
- Document attachment

---

### 7.3.9 Checkbox, RadioGroup, Switch

**Checkbox:**
```
[ ] Remember this device
[x] I agree to the terms
```

**RadioGroup:**
```
Task Type *
( ) Inspection
(*) Repair
( ) Overhaul
```

**Switch:**
```
Active  [====O]  (toggle on)
Active  [O====]  (toggle off)
```

**Props (common):**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| label | string | - | Field label |
| checked/value | boolean | false | Current state |
| onChange | function | - | Change handler |
| disabled | boolean | false | Disabled state |

---

### 7.3.10 FormSection

**Purpose:** Group related form fields with heading.

**Structure:**
```
+----------------------------------------------------------+
| Section Title                                            |
| Optional description text                                |
+----------------------------------------------------------+
| [Form fields...]                                         |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| title | string | - | Section heading |
| description | string | - | Helper text |
| children | ReactNode | - | Form fields |
| collapsible | boolean | false | Allow collapse |

---

### 7.3.11 FormActions

**Purpose:** Consistent form action button layout.

**Structure:**
```
+----------------------------------------------------------+
|                              [Cancel]  [Save]  [Submit]  |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| onCancel | function | - | Cancel handler |
| onSave | function | - | Save draft handler |
| onSubmit | function | - | Submit handler |
| submitLabel | string | 'Submit' | Submit button text |
| loading | boolean | false | Loading state |
| disabled | boolean | false | Disable all |

---

## 7.4 Feedback Components

### 7.4.1 Toast/Snackbar

**Purpose:** Temporary notification messages.

**Structure:**
```
+--------------------------------------------------+
| [Icon] Message text                        [x]   |
+--------------------------------------------------+
```

**Variants:**
| Variant | Color | Icon | Duration |
|---------|-------|------|----------|
| success | Green | Checkmark | 3s |
| error | Red | X-Circle | 5s (or persistent) |
| warning | Amber | Alert | 4s |
| info | Blue | Info | 3s |

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| message | string | - | Notification text |
| variant | string | 'info' | Toast type |
| duration | number | 3000 | Auto-dismiss (ms) |
| action | object | - | Action button |
| onClose | function | - | Close handler |

**Position:** Bottom-right (desktop), bottom-center (mobile)

**Stacking:** Max 3 visible, older toasts dismissed

---

### 7.4.2 Modal/Dialog

**Purpose:** Overlay dialogs for confirmations, forms, and details.

**Structure:**
```
+----------------------------------------------+
| Modal Title                          [X]     |
+----------------------------------------------+
|                                              |
| Modal content goes here.                     |
|                                              |
| Can include forms, information,              |
| or confirmation messages.                    |
|                                              |
+----------------------------------------------+
|                     [Cancel]  [Confirm]      |
+----------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| open | boolean | false | Visibility state |
| onClose | function | - | Close handler |
| title | string | - | Modal header |
| size | string | 'md' | Width (sm, md, lg, xl) |
| closeOnOverlay | boolean | true | Close on backdrop click |
| closeOnEscape | boolean | true | Close on Escape key |
| footer | ReactNode | - | Footer buttons |

**Sizes:**
- sm: 400px
- md: 560px
- lg: 720px
- xl: 960px
- full: 100% viewport

**Variants:**
- **Confirmation:** Simple yes/no decision
- **Form:** Contains form fields
- **Detail:** Read-only information display
- **Destructive:** Red styling for dangerous actions

---

### 7.4.3 LoadingSpinner

**Purpose:** Indicate loading/processing state.

**Variants:**
| Variant | Use Case |
|---------|----------|
| Spinner | Button loading, inline loading |
| Circular Progress | Determinate progress |
| Skeleton | Content placeholder |

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| size | string | 'md' | Spinner size |
| color | string | 'primary' | Spinner color |
| label | string | - | Accessible label |

---

### 7.4.4 SkeletonLoader

**Purpose:** Placeholder during content loading.

**Structure:**
```
+----------------------------------------------------------+
| [====                                          ] Title   |
| [====================                          ] Subtitle|
+----------------------------------------------------------+
| [======================================]                 |
| [==========================]                             |
| [================================]                       |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| variant | string | 'text' | Shape (text, rect, circle) |
| width | string | '100%' | Element width |
| height | string | '1rem' | Element height |
| animation | string | 'pulse' | Animation type |

---

### 7.4.5 EmptyState

**Purpose:** Display when no data is available.

**Structure:**
```
+----------------------------------------------------------+
|                                                          |
|              [Illustration/Icon]                         |
|                                                          |
|              No Aircraft Found                           |
|                                                          |
|    Add your first aircraft to start tracking             |
|    maintenance schedules.                                |
|                                                          |
|              [+ Add Aircraft]                            |
|                                                          |
+----------------------------------------------------------+
```

**Variants:**
| Variant | Use Case | Action |
|---------|----------|--------|
| No Data | List is empty | Primary action button |
| No Results | Filter returns nothing | Clear filters link |
| Error | Failed to load | Retry button |
| Offline | No connection | Sync when online |

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| icon | ReactNode | - | Illustration |
| title | string | - | Main message |
| description | string | - | Helper text |
| action | object | - | Action button |

---

### 7.4.6 ProgressBar / ProgressCircle

**Purpose:** Show completion progress.

**ProgressBar Structure:**
```
Compliance Progress
[========        ] 60%
3 of 5 items complete
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| value | number | 0 | Progress (0-100) |
| max | number | 100 | Maximum value |
| showLabel | boolean | true | Show percentage |
| color | string | 'primary' | Bar color |
| size | string | 'md' | Bar thickness |

---

## 7.5 Navigation Components

### 7.5.1 Tabs

**Purpose:** Switch between content sections.

**Structure:**
```
[Overview] [Tasks] [History] [Settings]
+----------------------------------------------------------+
| Tab content displayed here                               |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| tabs | Tab[] | - | Tab definitions |
| activeTab | string | - | Currently active |
| onChange | function | - | Tab change handler |
| variant | string | 'underline' | Style variant |

**Variants:**
- **Underline:** Tabs with underline indicator (default)
- **Pill:** Tabs with background pill
- **Segment:** Segmented control style

---

### 7.5.2 Breadcrumbs

**Purpose:** Show navigation hierarchy.

**Structure:**
```
Home > Aircraft > N12345 > Tasks
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| items | BreadcrumbItem[] | - | Path items |
| separator | string | '>' | Separator character |
| maxItems | number | 4 | Max visible items |

---

### 7.5.3 Pagination

**Purpose:** Navigate through paginated data.

**Structure:**
```
Showing 21-40 of 100     [< Prev] [1] [2] [3] ... [5] [Next >]
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| page | number | 1 | Current page |
| perPage | number | 20 | Items per page |
| total | number | - | Total items |
| onChange | function | - | Page change handler |
| showInfo | boolean | true | Show item count |

---

### 7.5.4 StepIndicator

**Purpose:** Show progress through multi-step wizard.

**Structure:**
```
(1)---------(2)---------(3)---------(4)
Upload     Validate     Review     Complete
   [Done]     [Active]     [ ]        [ ]
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| steps | Step[] | - | Step definitions |
| currentStep | number | 1 | Active step |
| variant | string | 'horizontal' | Layout direction |

---

## 7.6 Data Visualization Components

### 7.6.1 PieChart

**Purpose:** Show distribution/proportions.

**Use Cases:**
- Aircraft status distribution
- Task type breakdown
- Cost category breakdown

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| data | DataPoint[] | - | Chart data |
| size | number | 200 | Chart diameter |
| donut | boolean | false | Donut style |
| showLegend | boolean | true | Display legend |
| showLabels | boolean | true | Show slice labels |

---

### 7.6.2 BarChart

**Purpose:** Compare values across categories.

**Use Cases:**
- Tasks by program
- Costs by aircraft
- Compliance by month

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| data | DataPoint[] | - | Chart data |
| orientation | string | 'vertical' | Bar direction |
| stacked | boolean | false | Stack multiple series |
| showGrid | boolean | true | Display grid lines |
| showTooltip | boolean | true | Hover tooltips |

---

### 7.6.3 LineChart

**Purpose:** Show trends over time.

**Use Cases:**
- Compliance rate trend
- Cost trends over time
- Flight hours accumulated

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| data | Series[] | - | Chart data series |
| xAxis | AxisConfig | - | X-axis configuration |
| yAxis | AxisConfig | - | Y-axis configuration |
| showArea | boolean | false | Fill area under line |
| showDots | boolean | true | Show data points |

---

### 7.6.4 GaugeChart

**Purpose:** Show single value against target.

**Use Cases:**
- Compliance rate percentage
- Capacity utilization
- Budget consumption

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| value | number | - | Current value |
| min | number | 0 | Minimum value |
| max | number | 100 | Maximum value |
| thresholds | Threshold[] | - | Color zones |
| showValue | boolean | true | Display value |

---

## 7.7 Aviation-Specific Components

### 7.7.1 ComplianceChecklist

**Purpose:** Display and interact with compliance items.

**Structure:**
```
+----------------------------------------------------------+
| Compliance Items                          Progress: 3/5   |
+----------------------------------------------------------+
| [x] Torque verified to spec               [Signed]       |
|     Signed by John Doe at 10:15 AM                       |
+----------------------------------------------------------+
| [x] Oil level checked                     [Signed]       |
|     Signed by John Doe at 10:20 AM                       |
+----------------------------------------------------------+
| [ ] Safety wire installed                 [Sign Off]     |
|     Result: [Select v]                                   |
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| items | ComplianceItem[] | - | Checklist items |
| taskState | TaskState | - | Parent task state |
| canSign | boolean | - | User can sign off |
| onSignOff | function | - | Sign-off handler |

---

### 7.7.2 PartReservationList

**Purpose:** Manage part reservations within a task.

**Structure:**
```
+----------------------------------------------------------+
| Reserved Parts (2)                        [+ Reserve]    |
+----------------------------------------------------------+
| Part Name          | Serial #   | Status  | Actions      |
|-------------------|------------|---------|--------------|
| Oil Filter        | SN-12345   | Reserved| [Use][Return]|
| Engine Oil (6 qt) | N/A        | Reserved| [Use][Return]|
+----------------------------------------------------------+
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| reservations | PartReservation[] | - | Part list |
| taskState | TaskState | - | Parent task state |
| onReserve | function | - | Add reservation |
| onUse | function | - | Mark as used |
| onReturn | function | - | Return to inventory |

---

### 7.7.3 TaskStateIndicator

**Purpose:** Visual representation of task state machine.

**Structure:**
```
[Scheduled] -----> [In Progress] -----> [Completed]
     |                  |
     +------------------+-----> [Cancelled]

Current: [In Progress]
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| currentState | TaskState | - | Current state |
| showAllStates | boolean | true | Show full machine |
| interactive | boolean | false | Allow transitions |
| onTransition | function | - | Transition handler |

---

### 7.7.4 AircraftStatusIndicator

**Purpose:** Visual aircraft status with contextual icon.

**Structure:**
```
[Airplane Icon] Operational
  or
[Wrench Icon] Maintenance
  or
[Warning Icon] Grounded
```

**Props:**
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| status | AircraftStatus | - | Current status |
| size | string | 'md' | Indicator size |
| showLabel | boolean | true | Show status text |
| animated | boolean | false | Pulse animation for grounded |

**Icons:**
- `operational`: Airplane icon (green)
- `maintenance`: Wrench icon (amber)
- `grounded`: Warning triangle (red, pulsing)

---

# Part 8: Interaction Patterns

This section defines consistent interaction behaviors across the AMSS application.

---

## 8.1 Form Validation Patterns

### 8.1.1 Inline Validation

**Trigger:** On field blur (when user leaves field)

**Behavior:**
1. User enters value in field
2. User moves to next field (blur event)
3. Validation runs immediately
4. If invalid: Show error message below field, highlight field in red
5. If valid: Remove any existing error, optional success indicator

**Visual Feedback:**
```
Email Address *
+------------------------------------------+
| invalid-email                            |  <- Red border
+------------------------------------------+
Please enter a valid email address          <- Red text

Email Address *
+------------------------------------------+
| user@example.com                     [v] |  <- Green checkmark (optional)
+------------------------------------------+
```

**Validation Timing:**
| Event | Validation | Use Case |
|-------|------------|----------|
| On blur | Run validation | Default for all fields |
| On change (debounced) | Run after 300ms | For async validation (uniqueness) |
| On submit | Validate all fields | Final validation before API call |

---

### 8.1.2 Submit Validation

**Trigger:** Form submit button clicked

**Behavior:**
1. User clicks Submit/Save button
2. Validate all fields simultaneously
3. If any invalid:
   - Show all error messages
   - Scroll to first error field
   - Focus first error field
   - Show toast: "Please fix the errors below"
4. If all valid:
   - Disable form fields and submit button
   - Show loading spinner in button
   - Make API call

**Error Summary (for complex forms):**
```
+----------------------------------------------------------+
| [!] Please fix the following errors:                     |
| - Email address is required                              |
| - Password must be at least 8 characters                 |
| - End time must be after start time                      |
+----------------------------------------------------------+
```

---

### 8.1.3 Server-Side Validation

**Trigger:** API returns validation error (400 Bad Request)

**Behavior:**
1. Parse error response
2. Map server field names to form fields
3. Display errors on appropriate fields
4. Show toast with summary message

**Error Response Format:**
```json
{
  "error": "validation",
  "details": [
    { "field": "email", "message": "Email already exists" },
    { "field": "tail_number", "message": "Tail number must be unique" }
  ]
}
```

---

## 8.2 State Transition Flows

### 8.2.1 Task State Transitions

**State Machine:**
```
                    +-- Cancel --+
                    |            v
[Scheduled] --> [In Progress] --> [Completed]
     |              |
     +-- Cancel ----+
            |
            v
      [Cancelled]
```

**Transition: Scheduled -> In Progress (Start Task)**

1. User clicks "Start Task" button
2. System checks pre-conditions:
   - Task is assigned to current user (or user has override permission)
   - Aircraft status is `grounded`
   - Current time is within 5 minutes of start time
3. If pre-conditions fail:
   - Show error modal with specific issue
   - Provide guidance on how to resolve
4. If pre-conditions pass:
   - Show confirmation dialog:
     ```
     Start this task?

     This will:
     - Reserve required parts
     - Record start time in audit log
     - Change task status to "In Progress"

     [Cancel] [Start Task]
     ```
5. On confirm:
   - Optimistic UI: Immediately update task status
   - Make API call
   - If success: Show success toast
   - If failure: Revert UI, show error toast

**Transition: In Progress -> Completed (Complete Task)**

1. User clicks "Complete Task" button
2. System checks pre-conditions:
   - All part reservations are used or returned
   - All compliance items are signed off
   - Completion notes are entered
3. If pre-conditions fail:
   - Show modal with checklist of incomplete items:
     ```
     Cannot complete task. Please resolve:

     [ ] 2 parts still reserved (not used/returned)
     [ ] 1 compliance item pending sign-off
     [x] Completion notes entered

     [Go to Parts] [Go to Compliance] [Cancel]
     ```
4. If pre-conditions pass:
   - Show confirmation dialog:
     ```
     Complete this task?

     This action cannot be undone. The following will be recorded:
     - Completion timestamp
     - Parts used
     - Compliance sign-offs

     [Cancel] [Complete Task]
     ```
5. On confirm:
   - Make API call (no optimistic update for completion)
   - Show loading state
   - If success: Redirect to task list, show success toast
   - If failure: Show error, allow retry

---

### 8.2.2 Aircraft Status Transitions

**State Machine:**
```
[Operational] <--> [Maintenance] <--> [Grounded]
       ^                                  |
       +----------------------------------+
```

**Ground Aircraft Flow:**

1. User selects "Ground Aircraft" from actions menu
2. Show confirmation:
   ```
   Ground aircraft N12345?

   This will:
   - Mark aircraft as not airworthy
   - Notify operations team (if webhooks configured)
   - Allow maintenance tasks to be started

   Reason (required):
   [________________________________]

   [Cancel] [Ground Aircraft]
   ```
3. Require reason for grounding
4. On confirm: Update status, create audit log

**Return to Service Flow:**

1. User selects "Return to Service"
2. System checks: All maintenance tasks completed
3. If incomplete tasks exist:
   ```
   Cannot return to service.

   The following tasks are still incomplete:
   - 100-Hour Inspection (In Progress)
   - Oil Change (Scheduled for today)

   [View Tasks] [Cancel]
   ```
4. If all clear, confirm and update

---

## 8.3 Optimistic UI Updates

### 8.3.1 When to Use Optimistic Updates

**Good Candidates:**
- Toggle switches (active/inactive)
- Status badge updates for non-critical changes
- Adding items to lists
- Form field auto-save

**Avoid Optimistic Updates:**
- Destructive actions (delete, cancel)
- Financial transactions
- Compliance sign-offs
- Task completion (regulatory implications)

### 8.3.2 Implementation Pattern

```typescript
async function updateAircraftStatus(id, newStatus) {
  // 1. Store previous state
  const previousStatus = aircraft.status;

  // 2. Optimistically update UI
  setAircraft({ ...aircraft, status: newStatus });

  try {
    // 3. Make API call
    await api.updateAircraft(id, { status: newStatus });

    // 4. Success: Show subtle confirmation
    toast.success('Status updated');
  } catch (error) {
    // 5. Failure: Revert to previous state
    setAircraft({ ...aircraft, status: previousStatus });

    // 6. Show error message
    toast.error('Failed to update status. Please try again.');
  }
}
```

---

## 8.4 Real-Time Updates Strategy

### 8.4.1 Update Methods

**Polling (Primary):**
- Dashboard data: Poll every 30 seconds
- Task list: Poll every 60 seconds
- Active task detail: Poll every 15 seconds

**WebSocket (Future Enhancement):**
- Real-time notifications
- Collaborative editing alerts
- Live dashboard updates

### 8.4.2 Polling Implementation

```typescript
// Dashboard polling
useEffect(() => {
  const pollInterval = setInterval(async () => {
    // Only poll if tab is visible
    if (document.visibilityState === 'visible') {
      const data = await fetchDashboardData();
      setDashboardData(data);
    }
  }, 30000);

  return () => clearInterval(pollInterval);
}, []);
```

### 8.4.3 Stale Data Indicators

When data may be stale (offline, polling paused):
```
+----------------------------------------------------------+
| [!] Data may be outdated. Last updated 5 minutes ago.    |
|     [Refresh Now]                                        |
+----------------------------------------------------------+
```

---

## 8.5 Drag-and-Drop (Task Scheduling)

### 8.5.1 Calendar View Drag-and-Drop

**Interaction:**
1. User hovers over task bar on calendar
2. Cursor changes to grab cursor
3. User clicks and holds
4. Task bar becomes semi-transparent
5. User drags to new date/time slot
6. Drop zones highlight when draggable enters
7. On drop:
   - Show confirmation if crossing day boundaries
   - Update start/end times
   - Recalculate duration if changed

**Constraints:**
- Cannot drag completed or cancelled tasks
- Cannot drag past today's date
- Snap to 15-minute intervals
- Show conflict warning if overlapping

### 8.5.2 Touch Support (Tablet)

**Long Press to Drag:**
1. User long-presses task (500ms)
2. Haptic feedback (if supported)
3. Task enters drag mode
4. User drags to new position
5. Release to drop

**Alternative:** Use dedicated "Reschedule" button for touch devices

---

## 8.6 Keyboard Navigation

### 8.6.1 Tab Order

**General Principles:**
- Logical top-to-bottom, left-to-right order
- Skip decorative elements
- Include all interactive elements
- Trap focus in modals

**Form Tab Order:**
```
1. First form field
2. Second form field
3. ...
4. Cancel button
5. Submit button
```

### 8.6.2 Keyboard Shortcuts

**Global Shortcuts:**

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + K` | Open global search |
| `Ctrl/Cmd + N` | Create new (context-dependent) |
| `Escape` | Close modal/dropdown |
| `?` | Show keyboard shortcuts help |

**Data Table Shortcuts:**

| Shortcut | Action |
|----------|--------|
| `Arrow Up/Down` | Navigate rows |
| `Enter` | Open selected row |
| `Space` | Toggle row selection |
| `Ctrl/Cmd + A` | Select all |

**Form Shortcuts:**

| Shortcut | Action |
|----------|--------|
| `Enter` | Submit form (when in last field) |
| `Escape` | Cancel/close form |
| `Tab` | Next field |
| `Shift + Tab` | Previous field |

### 8.6.3 Focus Indicators

**Visible Focus:**
- All interactive elements must have visible focus indicator
- Focus ring: 2px solid primary color with 2px offset
- High contrast mode: 3px black outline

```css
:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}
```

---

## 8.7 Touch Gestures for Tablet

### 8.7.1 Supported Gestures

| Gesture | Action |
|---------|--------|
| Tap | Select/activate |
| Long press | Context menu / drag start |
| Swipe left (on list item) | Show quick actions |
| Swipe right (on list item) | Mark as complete (where applicable) |
| Pinch | Zoom (calendar view) |
| Two-finger pan | Scroll large views |
| Pull down | Refresh |

### 8.7.2 Swipe Actions

**Task List Item Swipe:**
```
<-- Swipe Left                    Swipe Right -->
[Archive] [Delete]    Task Item    [Complete] [Start]
```

**Implementation:**
- Reveal distance: 80px per action
- Snap back if not fully swiped
- Haptic feedback on action commit

### 8.7.3 Touch Targets

**Minimum Size:** 44x44 pixels (per Apple HIG / Material Design)

**Spacing:** Minimum 8px between touch targets

---

## 8.8 Undo/Redo Patterns

### 8.8.1 Immediate Undo (Toast Action)

**Use Case:** Quick reversible actions

**Flow:**
1. User performs action (e.g., archive task)
2. Show toast with undo option:
   ```
   +--------------------------------------------------+
   | Task archived                          [Undo]    |
   +--------------------------------------------------+
   ```
3. Toast visible for 5 seconds
4. If "Undo" clicked: Revert action
5. If toast dismissed: Action becomes permanent

### 8.8.2 Form Undo

**Use Case:** Accidental form field changes

**Implementation:**
- Store form state on focus
- On `Ctrl/Cmd + Z`: Revert to previous state
- Track multiple undo levels (up to 10)

### 8.8.3 No Undo Available

**Scenarios Where Undo Not Possible:**
- Task completion (regulatory requirement)
- Compliance sign-off (immutable)
- Delete after grace period
- Published reports

**User Notification:**
```
+----------------------------------------------+
| [!] This action cannot be undone.            |
|                                              |
| Once completed, this task will be            |
| permanently recorded in the audit log.       |
|                                              |
|              [Cancel]  [I Understand]        |
+----------------------------------------------+
```

---

## 8.9 Bulk Actions

### 8.9.1 Selection Pattern

**Select Multiple:**
1. Enable selection mode (checkbox column or button)
2. Click rows to select/deselect
3. Shift+Click for range selection
4. Ctrl/Cmd+Click for multi-select

**Select All:**
- Header checkbox selects all visible rows
- "Select all X items" link for selecting across pages

### 8.9.2 Bulk Action Bar

**Appearance:** Fixed bar at bottom when items selected

```
+----------------------------------------------------------+
| 5 items selected            [Assign] [Cancel] [Delete]   |
+----------------------------------------------------------+
```

**Available Bulk Actions by Entity:**

| Entity | Bulk Actions |
|--------|--------------|
| Aircraft | Change status, Export, Delete |
| Tasks | Assign mechanic, Cancel, Export |
| Parts | Update quantity, Export |
| Users | Change role, Deactivate |

### 8.9.3 Bulk Action Confirmation

```
+----------------------------------------------+
| Assign 5 tasks?                      [X]     |
+----------------------------------------------+
| You are about to assign 5 tasks to:          |
|                                              |
| [Select Mechanic v]                          |
|                                              |
| Tasks to assign:                             |
| - 100-Hour Inspection - N12345              |
| - Oil Change - N67890                        |
| - Annual Inspection - N99999                 |
| + 2 more...                                  |
|                                              |
|              [Cancel]  [Assign All]          |
+----------------------------------------------+
```

---

## 8.10 Search and Filter Patterns

### 8.10.1 Global Search

**Trigger:** `Ctrl/Cmd + K` or click search icon

**Modal Search:**
```
+----------------------------------------------+
| Search AMSS                          [X]     |
+----------------------------------------------+
| [Search...]                                  |
+----------------------------------------------+
| Recent Searches                              |
| - N12345                                     |
| - 100-hour inspection                        |
|                                              |
| Quick Actions                                |
| - Create Task                                |
| - Add Aircraft                               |
+----------------------------------------------+
```

**Search Results:**
- Group by entity type (Aircraft, Tasks, Programs, Parts)
- Show max 5 per group
- Highlight matching text
- Keyboard navigation (arrow keys + Enter)

### 8.10.2 List Filtering

**Filter Bar Pattern:**
```
| Search: [____________]  Status: [All v]  Aircraft: [All v]  |
| Date: [Last 7 days v]  Assigned: [All v]      [Clear All]   |
```

**Active Filter Pills:**
```
| Filters: [Status: Overdue x] [Aircraft: N12345 x] [Clear All] |
```

**Filter Behavior:**
- Filters apply immediately (no "Apply" button needed)
- URL reflects filter state (shareable)
- Persist filters in session (navigating away and back)
- "Clear All" resets to defaults

### 8.10.3 Advanced Filtering

**Expandable Advanced Filters:**
```
[+ Show Advanced Filters]

+----------------------------------------------------------+
| Date Range: [Dec 1] to [Dec 31]                          |
| Created By: [Select user v]                              |
| Hours Range: [0] to [5000]                               |
| Include Cancelled: [ ]                                    |
+----------------------------------------------------------+
```

---

## 8.11 Infinite Scroll vs Pagination

### 8.11.1 Decision Matrix

| Criteria | Infinite Scroll | Pagination |
|----------|-----------------|------------|
| Data size | < 500 items | Any size |
| Need to bookmark position | No | Yes |
| User needs specific page | No | Yes |
| Mobile/tablet | Yes | Less ideal |
| Audit trail review | No | Yes |

### 8.11.2 Pagination (Default)

**Use For:**
- Aircraft list
- Task list
- User list
- Audit logs
- Reports

**Implementation:**
- Server-side pagination
- Default 20 items per page
- User can change: 10, 20, 50, 100
- Show total count
- URL includes page number

### 8.11.3 Infinite Scroll

**Use For:**
- Timeline/activity feed
- Comment threads
- Notification list

**Implementation:**
- Load next batch when 80% scrolled
- Show loading spinner at bottom
- "Load More" button as fallback
- Virtualize long lists (render only visible items)

---

## 8.12 Offline Support Patterns

### 8.12.1 Offline Detection

**Indicators:**
1. Network status in header:
   ```
   [Offline Icon] You are offline. Changes will sync when connected.
   ```
2. Toast on connection change:
   ```
   [Warning] You are now offline. Some features may be limited.
   ```
3. Form submission warning:
   ```
   [!] You're offline. Your changes will be saved and submitted when you reconnect.
   ```

### 8.12.2 Offline Data Access

**Cached Data:**
- Current user profile
- Recent aircraft list
- Assigned tasks (for mechanics)
- Recent audit logs

**Unavailable Offline:**
- Creating new entities
- Compliance sign-offs (requires verification)
- Report generation
- Bulk operations

### 8.12.3 Offline Queue

**Queue Pattern:**
1. User performs action offline
2. Action queued locally (IndexedDB)
3. Show pending indicator:
   ```
   Task started (pending sync) [Clock icon]
   ```
4. When online: Process queue in order
5. If conflict: Show resolution dialog

### 8.12.4 Sync Indicators

**Sync Status:**
```
[Synced] All changes saved              <- Green checkmark
[Syncing] Saving changes...             <- Spinner
[Pending] 3 changes waiting to sync     <- Warning icon
[Error] Failed to sync. [Retry]         <- Red error icon
```

---

## 8.13 Error Handling Patterns

### 8.13.1 Error Categories

| Category | User Message | Action |
|----------|--------------|--------|
| Network Error | "Unable to connect. Please check your internet." | Retry button |
| Server Error (500) | "Something went wrong. Please try again." | Retry button + support link |
| Validation Error (400) | Specific field errors | Fix and resubmit |
| Auth Error (401) | "Your session has expired." | Redirect to login |
| Permission Error (403) | "You don't have permission for this action." | Contact admin |
| Not Found (404) | "This item was not found." | Back button |
| Conflict (409) | Specific conflict message | Resolution options |

### 8.13.2 Error Recovery Patterns

**Retry with Backoff:**
```typescript
async function fetchWithRetry(url, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fetch(url);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await delay(Math.pow(2, i) * 1000); // Exponential backoff
    }
  }
}
```

**Graceful Degradation:**
- If dashboard API fails, show cached data with warning
- If single widget fails, show error state for that widget only
- If non-critical feature fails, hide feature with logged error

### 8.13.3 Error Boundaries

**Component Error Boundary:**
```
+----------------------------------------------------------+
| [!] This section couldn't load                           |
|                                                          |
| An error occurred while loading this content.            |
|                                                          |
| [Try Again]  [Report Issue]                              |
+----------------------------------------------------------+
```

---

## 8.14 Loading State Patterns

### 8.14.1 Loading Indicators by Context

| Context | Indicator | Duration |
|---------|-----------|----------|
| Page load | Full-page skeleton | Any |
| Section load | Section skeleton | Any |
| Button action | Button spinner | < 5s |
| Inline update | Inline spinner | < 2s |
| Background sync | Subtle indicator | Any |

### 8.14.2 Skeleton Loading

**Use:** Initial page/section load

**Implementation:**
- Match approximate layout of real content
- Animate with subtle pulse
- Transition smoothly to real content

### 8.14.3 Progress Indicators

**Use:** Operations with known duration

**Examples:**
- File upload: Progress bar with percentage
- CSV import: Step indicator + progress
- Report generation: Progress bar + ETA

---

## 8.15 Accessibility Patterns

### 8.15.1 ARIA Labels

**All Interactive Elements:**
```html
<button aria-label="Add new aircraft">
  <PlusIcon />
</button>

<input
  aria-label="Search aircraft"
  aria-describedby="search-help"
/>
<span id="search-help">Search by tail number or model</span>
```

### 8.15.2 Screen Reader Announcements

**Live Regions:**
```html
<div aria-live="polite" aria-atomic="true">
  <!-- Dynamic content announced to screen readers -->
  {statusMessage}
</div>
```

**Announce:**
- Form validation errors
- Toast messages
- Status changes
- Loading complete

### 8.15.3 Color Contrast

**Minimum Ratios:**
- Normal text: 4.5:1
- Large text (18pt+): 3:1
- UI components: 3:1

**Never Use Color Alone:**
- Pair colors with icons/text
- Use patterns for charts
- Underline links

---

This concludes Parts 6-8 of the AMSS UI/UX Specification Document. These specifications provide the foundation for building a consistent, accessible, and user-friendly aviation maintenance scheduling system.

---

**Document Control:**
- **Version:** 1.0
- **Created:** December 27, 2025
- **Author:** UI/UX Specification Team
- **Review Cycle:** Quarterly

**Related Documents:**
- Parts 1-5: Design Principles, Navigation, Responsive Design, Typography, Colors
- Parts 9-12: Accessibility, Performance, Testing, Implementation Guidelines

---

[Next: Quality Standards](06_QUALITY_STANDARDS.md)
