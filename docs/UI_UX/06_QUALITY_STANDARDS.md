# AMSS UI/UX Specification - Quality Standards

[Back to Index](00_INDEX.md) | [Previous: Components & Interactions](05_COMPONENTS_INTERACTIONS.md)

---

## Part 9: Accessibility Requirements

### 9.1 WCAG 2.1 AA Compliance Overview

The Aviation Maintenance Scheduling System (AMSS) must meet Web Content Accessibility Guidelines (WCAG) 2.1 Level AA compliance. This is mandatory for ensuring the system is usable by personnel with disabilities and meets organizational accessibility requirements.

#### 9.1.1 Compliance Principles

| Principle | Description | AMSS Application |
|-----------|-------------|------------------|
| **Perceivable** | Information must be presentable in ways users can perceive | Alt text, captions, color contrast |
| **Operable** | Interface components must be operable | Keyboard navigation, timing adjustments |
| **Understandable** | Information and operation must be understandable | Clear labels, error prevention |
| **Robust** | Content must be robust enough for assistive technologies | Valid HTML, ARIA implementation |

#### 9.1.2 Success Criteria Mapping

```yaml
# Priority success criteria for AMSS
critical:
  - 1.1.1: Non-text Content (Level A)
  - 1.3.1: Info and Relationships (Level A)
  - 1.4.3: Contrast (Minimum) (Level AA)
  - 2.1.1: Keyboard (Level A)
  - 2.4.3: Focus Order (Level A)
  - 4.1.2: Name, Role, Value (Level A)

high:
  - 1.4.11: Non-text Contrast (Level AA)
  - 2.4.6: Headings and Labels (Level AA)
  - 2.4.7: Focus Visible (Level AA)
  - 3.3.1: Error Identification (Level A)
  - 3.3.2: Labels or Instructions (Level A)

medium:
  - 1.4.4: Resize Text (Level AA)
  - 2.4.1: Bypass Blocks (Level A)
  - 2.4.4: Link Purpose (Level A)
  - 3.2.3: Consistent Navigation (Level AA)
```

### 9.2 Screen Reader Support

#### 9.2.1 ARIA Labels and Roles

All interactive elements must have appropriate ARIA attributes:

```html
<!-- Navigation landmark -->
<nav aria-label="Main navigation" role="navigation">
  <ul role="menubar">
    <li role="none">
      <a role="menuitem" href="/dashboard" aria-current="page">
        Dashboard
      </a>
    </li>
    <li role="none">
      <a role="menuitem" href="/aircraft" aria-haspopup="true">
        Aircraft
      </a>
    </li>
  </ul>
</nav>

<!-- Aircraft status card -->
<article
  aria-labelledby="aircraft-title-n12345"
  aria-describedby="aircraft-status-n12345"
>
  <h3 id="aircraft-title-n12345">N12345</h3>
  <p id="aircraft-status-n12345">
    Status: <span class="status-operational">Operational</span>
  </p>
</article>

<!-- Action button with context -->
<button
  aria-label="Start maintenance task: 100-hour inspection for N12345"
  aria-describedby="task-requirements"
>
  Start Task
</button>
```

#### 9.2.2 Live Regions for Dynamic Content

```html
<!-- Status announcements (polite) -->
<div
  aria-live="polite"
  aria-atomic="true"
  class="sr-only"
  id="status-announcer"
>
  <!-- Dynamically updated: "Task saved successfully" -->
</div>

<!-- Critical alerts (assertive) -->
<div
  aria-live="assertive"
  aria-atomic="true"
  role="alert"
  id="alert-announcer"
>
  <!-- Dynamically updated: "Aircraft N12345 has been grounded" -->
</div>

<!-- Progress updates -->
<div
  aria-live="polite"
  aria-busy="true"
  role="status"
>
  Syncing offline changes: 3 of 7 complete
</div>
```

#### 9.2.3 Required ARIA Patterns by Component

| Component | Required ARIA | Example |
|-----------|---------------|---------|
| Modal Dialog | `role="dialog"`, `aria-modal="true"`, `aria-labelledby` | Confirmation dialogs |
| Dropdown Menu | `role="menu"`, `aria-expanded`, `aria-haspopup` | Aircraft actions menu |
| Tabs | `role="tablist"`, `role="tab"`, `role="tabpanel"` | Task detail tabs |
| Alert | `role="alert"`, `aria-live="assertive"` | Grounding warnings |
| Data Grid | `role="grid"`, `aria-rowcount`, `aria-colcount` | Aircraft list |
| Progress | `role="progressbar"`, `aria-valuenow`, `aria-valuemin/max` | Sync progress |
| Combobox | `role="combobox"`, `aria-autocomplete`, `aria-controls` | Aircraft search |

### 9.3 Keyboard Navigation

#### 9.3.1 Focus Management

```typescript
// Focus management utility
const FocusManager = {
  // Store last focused element before modal opens
  trapFocus(containerEl: HTMLElement) {
    const focusableElements = containerEl.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

    containerEl.addEventListener('keydown', (e) => {
      if (e.key === 'Tab') {
        if (e.shiftKey && document.activeElement === firstElement) {
          e.preventDefault();
          lastElement.focus();
        } else if (!e.shiftKey && document.activeElement === lastElement) {
          e.preventDefault();
          firstElement.focus();
        }
      }
    });

    firstElement.focus();
  },

  // Restore focus when modal closes
  restoreFocus(previousElement: HTMLElement) {
    previousElement?.focus();
  }
};
```

#### 9.3.2 Skip Links

```html
<!-- Skip link structure (first element in body) -->
<a href="#main-content" class="skip-link">
  Skip to main content
</a>
<a href="#main-navigation" class="skip-link">
  Skip to navigation
</a>
<a href="#search" class="skip-link">
  Skip to search
</a>

<style>
.skip-link {
  position: absolute;
  top: -40px;
  left: 0;
  background: var(--color-primary);
  color: white;
  padding: 8px 16px;
  z-index: 9999;
  transition: top 0.2s;
}

.skip-link:focus {
  top: 0;
}
</style>
```

#### 9.3.3 Keyboard Shortcuts

| Shortcut | Action | Context |
|----------|--------|---------|
| `?` | Show keyboard shortcuts help | Global |
| `g` then `d` | Go to Dashboard | Global |
| `g` then `a` | Go to Aircraft | Global |
| `g` then `t` | Go to Tasks | Global |
| `/` | Focus search | Global |
| `Escape` | Close modal/dropdown | Modal/Dropdown open |
| `Enter` | Submit form / Confirm action | Form/Dialog |
| `n` | New item (aircraft/task/etc.) | List views |
| `j` / `k` | Navigate list items down/up | List views |
| `x` | Select/deselect item | List views |
| `Ctrl+S` | Save current form | Edit views |

```typescript
// Keyboard shortcut implementation
const shortcuts: ShortcutConfig[] = [
  {
    key: '?',
    description: 'Show keyboard shortcuts',
    handler: () => showShortcutsModal(),
    global: true,
  },
  {
    keys: ['g', 'd'],
    description: 'Go to Dashboard',
    handler: () => navigate('/dashboard'),
    global: true,
  },
  {
    key: 'Escape',
    description: 'Close modal',
    handler: () => closeActiveModal(),
    context: 'modal',
  },
];
```

### 9.4 Color Contrast Requirements

#### 9.4.1 Contrast Ratios

| Element Type | Minimum Ratio | WCAG Level |
|--------------|---------------|------------|
| Normal text (< 18px) | 4.5:1 | AA |
| Large text (>= 18px or >= 14px bold) | 3:1 | AA |
| UI components and graphics | 3:1 | AA |
| Focus indicators | 3:1 | AA |

#### 9.4.2 Color Palette Contrast Matrix

```scss
// Verified contrast ratios against backgrounds
$colors-verified: (
  // Text on white background (#FFFFFF)
  'text-primary': #1F2937,      // 12.63:1 - Pass AA/AAA
  'text-secondary': #4B5563,    // 7.51:1  - Pass AA/AAA
  'text-muted': #6B7280,        // 5.03:1  - Pass AA

  // Status colors on white background
  'status-operational': #047857, // 5.94:1  - Pass AA
  'status-maintenance': #B45309, // 4.56:1  - Pass AA
  'status-grounded': #B91C1C,    // 5.92:1  - Pass AA

  // Interactive elements
  'link-default': #1D4ED8,       // 6.87:1  - Pass AA
  'link-hover': #1E40AF,         // 8.59:1  - Pass AA/AAA

  // On dark backgrounds (sidebar #1F2937)
  'text-on-dark': #F9FAFB,       // 12.63:1 - Pass AA/AAA
  'text-muted-on-dark': #D1D5DB, // 8.23:1  - Pass AA/AAA
);
```

#### 9.4.3 Non-Color Status Indicators

Status must never rely on color alone:

```html
<!-- Aircraft status with icon + text + color -->
<span class="status status-grounded">
  <svg aria-hidden="true" class="status-icon">
    <use href="#icon-grounded" />
  </svg>
  <span class="status-text">Grounded</span>
</span>

<!-- Task priority with pattern -->
<span class="priority priority-critical">
  <span class="priority-indicator" aria-hidden="true">!!!</span>
  <span class="priority-text">Critical</span>
</span>
```

### 9.5 Focus Indicators

#### 9.5.1 Focus Ring Styles

```scss
// Global focus styles
:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

// Custom focus ring for specific components
.btn:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.3);
}

// High contrast focus for dark backgrounds
.sidebar-link:focus-visible {
  outline: 2px solid #FFFFFF;
  outline-offset: 2px;
  background-color: rgba(255, 255, 255, 0.1);
}

// Focus ring colors
:root {
  --color-focus: #3B82F6;
  --color-focus-error: #EF4444;
  --color-focus-success: #10B981;
}

// Remove default outline only when custom styles applied
:focus:not(:focus-visible) {
  outline: none;
}
```

### 9.6 Form Accessibility

#### 9.6.1 Label and Input Association

```html
<!-- Explicit label association -->
<div class="form-field">
  <label for="aircraft-tail-number" class="form-label">
    Tail Number
    <span class="required-indicator" aria-hidden="true">*</span>
  </label>
  <input
    type="text"
    id="aircraft-tail-number"
    name="tailNumber"
    aria-required="true"
    aria-describedby="tail-number-hint tail-number-error"
    pattern="^N[0-9]{1,5}[A-Z]{0,2}$"
  />
  <p id="tail-number-hint" class="form-hint">
    FAA format: N followed by up to 5 numbers and up to 2 letters
  </p>
  <p id="tail-number-error" class="form-error" role="alert" aria-live="polite">
    <!-- Populated when validation fails -->
  </p>
</div>

<!-- Fieldset for grouped inputs -->
<fieldset>
  <legend>Maintenance Schedule</legend>
  <div class="form-field">
    <label for="schedule-date">Scheduled Date</label>
    <input type="date" id="schedule-date" name="scheduledDate" />
  </div>
  <div class="form-field">
    <label for="schedule-time">Scheduled Time (UTC)</label>
    <input type="time" id="schedule-time" name="scheduledTime" />
  </div>
</fieldset>
```

#### 9.6.2 Error Announcements

```typescript
// Form validation with accessible error handling
const validateForm = (formData: FormData): ValidationResult => {
  const errors: FieldError[] = [];

  if (!formData.get('tailNumber')) {
    errors.push({
      field: 'tailNumber',
      message: 'Tail number is required',
    });
  }

  // Announce errors to screen readers
  if (errors.length > 0) {
    const announcer = document.getElementById('error-announcer');
    announcer.textContent = `Form has ${errors.length} error${errors.length > 1 ? 's' : ''}. ${errors[0].message}`;

    // Focus first error field
    const firstErrorField = document.getElementById(errors[0].field);
    firstErrorField?.focus();
  }

  return { valid: errors.length === 0, errors };
};
```

### 9.7 Modal Accessibility

```html
<!-- Accessible modal structure -->
<div
  role="dialog"
  aria-modal="true"
  aria-labelledby="modal-title"
  aria-describedby="modal-description"
  class="modal"
>
  <div class="modal-content">
    <header class="modal-header">
      <h2 id="modal-title">Confirm Aircraft Grounding</h2>
      <button
        type="button"
        aria-label="Close dialog"
        class="modal-close"
      >
        <svg aria-hidden="true"><!-- X icon --></svg>
      </button>
    </header>

    <div class="modal-body">
      <p id="modal-description">
        Are you sure you want to ground aircraft N12345? This action will
        cancel all scheduled flights and require maintenance clearance.
      </p>
    </div>

    <footer class="modal-footer">
      <button type="button" class="btn btn-secondary">
        Cancel
      </button>
      <button type="button" class="btn btn-danger">
        Ground Aircraft
      </button>
    </footer>
  </div>
</div>
```

### 9.8 Data Table Accessibility

```html
<!-- Accessible data table for aircraft list -->
<table
  role="grid"
  aria-label="Aircraft inventory"
  aria-describedby="table-summary"
  aria-rowcount="156"
>
  <caption id="table-summary" class="sr-only">
    Aircraft inventory showing tail number, type, status, and last maintenance date.
    Sortable by clicking column headers.
  </caption>

  <thead>
    <tr>
      <th scope="col" aria-sort="ascending">
        <button aria-label="Sort by tail number, currently ascending">
          Tail Number
          <svg aria-hidden="true"><!-- Sort icon --></svg>
        </button>
      </th>
      <th scope="col" aria-sort="none">
        <button aria-label="Sort by aircraft type">
          Type
        </button>
      </th>
      <th scope="col" aria-sort="none">Status</th>
      <th scope="col" aria-sort="none">Last Maintenance</th>
      <th scope="col">
        <span class="sr-only">Actions</span>
      </th>
    </tr>
  </thead>

  <tbody>
    <tr aria-rowindex="1">
      <th scope="row">N12345</th>
      <td>Cessna 172</td>
      <td>
        <span class="status status-operational">
          <svg aria-hidden="true"><!-- Icon --></svg>
          Operational
        </span>
      </td>
      <td>
        <time datetime="2024-01-15T14:30:00Z">
          Jan 15, 2024 14:30 UTC
        </time>
      </td>
      <td>
        <button aria-label="Actions for aircraft N12345" aria-haspopup="menu">
          <svg aria-hidden="true"><!-- More icon --></svg>
        </button>
      </td>
    </tr>
  </tbody>
</table>
```

### 9.9 Touch Target Sizes

```scss
// Minimum touch target: 44x44px
.touch-target {
  min-width: 44px;
  min-height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

// Button sizing
.btn {
  min-height: 44px;
  padding: 10px 16px;
}

.btn-icon {
  min-width: 44px;
  min-height: 44px;
  padding: 10px;
}

// Form inputs
.form-input,
.form-select {
  min-height: 44px;
  padding: 10px 12px;
}

// Checkbox/Radio touch area
.form-checkbox,
.form-radio {
  position: relative;
}

.form-checkbox input,
.form-radio input {
  position: absolute;
  width: 44px;
  height: 44px;
  opacity: 0;
  cursor: pointer;
}

// List item touch targets
.list-item {
  min-height: 48px;
  padding: 12px 16px;
}

// Spacing between touch targets
.action-group {
  gap: 8px; // Minimum 8px between adjacent targets
}
```

### 9.10 Motion and Animation

#### 9.10.1 Reduced Motion Support

```scss
// Default animations
.fade-in {
  animation: fadeIn 0.3s ease-out;
}

.slide-in {
  animation: slideIn 0.3s ease-out;
}

.spinner {
  animation: spin 1s linear infinite;
}

// Respect user preference
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }

  .fade-in {
    opacity: 1;
  }

  .slide-in {
    transform: none;
  }

  // Keep essential motion (loading indicators)
  .spinner {
    animation: none;
    border-style: dotted; // Static alternative
  }
}
```

#### 9.10.2 Animation Guidelines

| Animation Type | Duration | Reduced Motion Behavior |
|----------------|----------|------------------------|
| Micro-interactions | 100-200ms | Instant |
| Page transitions | 200-300ms | Instant |
| Modal open/close | 200ms | Instant |
| Loading spinners | Continuous | Static indicator |
| Progress bars | Continuous | Static progress |
| Alerts/toasts | 300ms in, 200ms out | Instant |

### 9.11 Testing Requirements and Tools

#### 9.11.1 Automated Testing

```typescript
// Jest + axe-core integration
import { axe, toHaveNoViolations } from 'jest-axe';

expect.extend(toHaveNoViolations);

describe('Aircraft List accessibility', () => {
  it('should have no accessibility violations', async () => {
    const { container } = render(<AircraftList />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  it('should have proper heading hierarchy', async () => {
    const results = await axe(container, {
      rules: {
        'heading-order': { enabled: true },
      },
    });
    expect(results).toHaveNoViolations();
  });
});
```

#### 9.11.2 Testing Checklist

| Test Type | Tool | Frequency |
|-----------|------|-----------|
| Automated scan | axe-core, Lighthouse | Every PR |
| Keyboard navigation | Manual | Every feature |
| Screen reader (Windows) | NVDA | Weekly |
| Screen reader (macOS) | VoiceOver | Weekly |
| Screen reader (mobile) | TalkBack/VoiceOver | Monthly |
| Color contrast | Colour Contrast Analyser | Every design change |
| Zoom testing | Browser zoom 200% | Every feature |
| High contrast mode | Windows High Contrast | Monthly |

#### 9.11.3 Screen Reader Testing Protocol

```markdown
## NVDA Testing Checklist

### Navigation
- [ ] Can navigate to main content with skip link
- [ ] Page title announced on load
- [ ] Landmarks are announced (main, nav, aside)
- [ ] Headings provide logical structure (h1-h6)

### Interactive Elements
- [ ] Buttons announce name and role
- [ ] Links announce destination
- [ ] Form fields announce label and state
- [ ] Error messages are announced

### Dynamic Content
- [ ] Live regions announce updates
- [ ] Modal focus is trapped
- [ ] Loading states are announced
- [ ] Status changes are communicated

### Tables
- [ ] Table caption/summary is announced
- [ ] Row/column headers are associated
- [ ] Can navigate cells with arrow keys
```

---

## Part 10: Offline Support Strategy

### 10.1 Service Worker Architecture

#### 10.1.1 Service Worker Registration

```typescript
// src/service-worker/register.ts
export async function registerServiceWorker(): Promise<void> {
  if ('serviceWorker' in navigator) {
    try {
      const registration = await navigator.serviceWorker.register(
        '/sw.js',
        { scope: '/' }
      );

      // Handle updates
      registration.addEventListener('updatefound', () => {
        const newWorker = registration.installing;
        newWorker?.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            // New version available
            notifyUserOfUpdate();
          }
        });
      });

      console.log('Service Worker registered:', registration.scope);
    } catch (error) {
      console.error('Service Worker registration failed:', error);
    }
  }
}
```

#### 10.1.2 Service Worker Structure

```typescript
// src/service-worker/sw.ts
import { precacheAndRoute } from 'workbox-precaching';
import { registerRoute } from 'workbox-routing';
import {
  CacheFirst,
  NetworkFirst,
  StaleWhileRevalidate,
} from 'workbox-strategies';
import { ExpirationPlugin } from 'workbox-expiration';
import { BackgroundSyncPlugin } from 'workbox-background-sync';

// Precache static assets (injected at build time)
precacheAndRoute(self.__WB_MANIFEST);

// Cache strategies by route
const strategies = {
  static: new CacheFirst({
    cacheName: 'static-assets-v1',
    plugins: [
      new ExpirationPlugin({
        maxEntries: 100,
        maxAgeSeconds: 30 * 24 * 60 * 60, // 30 days
      }),
    ],
  }),

  api: new NetworkFirst({
    cacheName: 'api-responses-v1',
    networkTimeoutSeconds: 10,
    plugins: [
      new ExpirationPlugin({
        maxEntries: 200,
        maxAgeSeconds: 24 * 60 * 60, // 24 hours
      }),
    ],
  }),

  images: new StaleWhileRevalidate({
    cacheName: 'images-v1',
    plugins: [
      new ExpirationPlugin({
        maxEntries: 50,
        maxAgeSeconds: 7 * 24 * 60 * 60, // 7 days
      }),
    ],
  }),
};
```

### 10.2 Cache Strategies by Resource Type

#### 10.2.1 Static Assets (Cache-First)

```typescript
// Static assets: JS, CSS, fonts
registerRoute(
  ({ request }) =>
    request.destination === 'script' ||
    request.destination === 'style' ||
    request.destination === 'font',
  strategies.static
);

// App shell HTML
registerRoute(
  ({ request }) => request.mode === 'navigate',
  new NetworkFirst({
    cacheName: 'pages-v1',
    networkTimeoutSeconds: 5,
    plugins: [
      new ExpirationPlugin({
        maxEntries: 20,
      }),
    ],
  })
);
```

#### 10.2.2 API Responses (Network-First with Cache Fallback)

```typescript
// Read-only API endpoints (aircraft list, task details)
registerRoute(
  ({ url }) => url.pathname.startsWith('/api/') &&
               ['GET'].includes(url.method),
  new NetworkFirst({
    cacheName: 'api-cache-v1',
    networkTimeoutSeconds: 10,
    plugins: [
      new ExpirationPlugin({
        maxEntries: 500,
        maxAgeSeconds: 60 * 60, // 1 hour
      }),
      {
        cacheWillUpdate: async ({ response }) => {
          // Only cache successful responses
          return response.status === 200 ? response : null;
        },
      },
    ],
  })
);

// Critical data endpoints (assigned tasks for mechanic)
registerRoute(
  ({ url }) => url.pathname.match(/\/api\/mechanics\/\d+\/tasks/),
  new NetworkFirst({
    cacheName: 'critical-data-v1',
    networkTimeoutSeconds: 5,
    plugins: [
      new ExpirationPlugin({
        maxEntries: 100,
        maxAgeSeconds: 4 * 60 * 60, // 4 hours
      }),
    ],
  })
);
```

#### 10.2.3 Critical Data (Background Sync)

```typescript
// Background sync for write operations
const bgSyncPlugin = new BackgroundSyncPlugin('offline-mutations', {
  maxRetentionTime: 24 * 60, // 24 hours in minutes
  onSync: async ({ queue }) => {
    let entry;
    while ((entry = await queue.shiftRequest())) {
      try {
        await fetch(entry.request.clone());
        notifyUser('Offline changes synced successfully');
      } catch (error) {
        await queue.unshiftRequest(entry);
        throw error;
      }
    }
  },
});

// Queue mutations when offline
registerRoute(
  ({ url, request }) =>
    url.pathname.startsWith('/api/') &&
    ['POST', 'PUT', 'PATCH', 'DELETE'].includes(request.method),
  new NetworkOnly({
    plugins: [bgSyncPlugin],
  }),
  'POST'
);
```

### 10.3 Offline-First Workflows

#### 10.3.1 Mechanic Task Execution

```typescript
// Offline task manager
class OfflineTaskManager {
  private db: IDBDatabase;

  async viewAssignedTasks(mechanicId: string): Promise<Task[]> {
    // Try network first
    try {
      const response = await fetch(`/api/mechanics/${mechanicId}/tasks`, {
        signal: AbortSignal.timeout(5000),
      });
      const tasks = await response.json();
      await this.cacheTasksLocally(tasks);
      return tasks;
    } catch {
      // Fall back to cached data
      return this.getCachedTasks(mechanicId);
    }
  }

  async startTask(taskId: string): Promise<void> {
    const update = {
      taskId,
      status: 'in_progress',
      startedAt: new Date().toISOString(),
      pendingSync: true,
    };

    // Update local state immediately
    await this.updateLocalTask(update);

    // Queue for sync
    await this.queueMutation({
      type: 'TASK_STATUS_UPDATE',
      payload: update,
      timestamp: Date.now(),
    });
  }

  async recordNotes(taskId: string, notes: string): Promise<void> {
    const noteEntry = {
      id: crypto.randomUUID(),
      taskId,
      content: notes,
      createdAt: new Date().toISOString(),
      pendingSync: true,
    };

    await this.saveLocalNote(noteEntry);
    await this.queueMutation({
      type: 'TASK_NOTE_ADD',
      payload: noteEntry,
      timestamp: Date.now(),
    });
  }
}
```

#### 10.3.2 Parts Reservation

```typescript
// Parts reservation with optimistic updates
async function reservePart(
  partId: string,
  taskId: string,
  quantity: number
): Promise<ReservationResult> {
  const reservation = {
    id: crypto.randomUUID(),
    partId,
    taskId,
    quantity,
    status: 'pending_sync',
    createdAt: new Date().toISOString(),
  };

  // Optimistic local update
  await offlineStore.createReservation(reservation);

  // Update UI immediately
  updatePartsInventoryUI(partId, -quantity);

  // Queue for background sync
  await syncQueue.add({
    type: 'PART_RESERVATION',
    payload: reservation,
    conflictStrategy: 'server_wins', // Server has authoritative stock count
  });

  return {
    success: true,
    reservation,
    pendingSync: !navigator.onLine,
  };
}
```

#### 10.3.3 Compliance Sign-offs

```typescript
// Compliance sign-off with digital signature
async function submitComplianceSignoff(
  taskId: string,
  checklistItems: ChecklistItem[],
  signature: SignatureData
): Promise<SignoffResult> {
  const signoff = {
    id: crypto.randomUUID(),
    taskId,
    checklistItems,
    signature: {
      data: signature.base64,
      timestamp: new Date().toISOString(),
      mechanicId: getCurrentUser().id,
      certificateNumber: getCurrentUser().certificateNumber,
    },
    status: 'pending_sync',
  };

  // Store locally with cryptographic hash
  const hash = await generateSignoffHash(signoff);
  signoff.integrityHash = hash;

  await offlineStore.saveSignoff(signoff);

  // Queue with high priority
  await syncQueue.add({
    type: 'COMPLIANCE_SIGNOFF',
    payload: signoff,
    priority: 'high',
    requiresConnectivity: true,
    maxRetries: 10,
  });

  return {
    success: true,
    signoff,
    warning: !navigator.onLine
      ? 'Sign-off saved locally. Will sync when connection is restored.'
      : null,
  };
}
```

### 10.4 IndexedDB Schema

```typescript
// IndexedDB schema definition
const DB_NAME = 'amss-offline-db';
const DB_VERSION = 1;

interface DBSchema {
  aircraft: {
    key: string;
    value: Aircraft;
    indexes: {
      'by-status': string;
      'by-tenant': string;
    };
  };
  tasks: {
    key: string;
    value: Task;
    indexes: {
      'by-aircraft': string;
      'by-mechanic': string;
      'by-status': string;
      'by-due-date': string;
    };
  };
  parts: {
    key: string;
    value: Part;
    indexes: {
      'by-part-number': string;
      'by-status': string;
    };
  };
  syncQueue: {
    key: string;
    value: SyncQueueItem;
    indexes: {
      'by-priority': number;
      'by-timestamp': number;
      'by-type': string;
    };
  };
  pendingSignoffs: {
    key: string;
    value: PendingSignoff;
    indexes: {
      'by-task': string;
    };
  };
  userPreferences: {
    key: string;
    value: UserPreference;
  };
}

// Database initialization
async function initDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);

    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;

      // Aircraft store
      const aircraftStore = db.createObjectStore('aircraft', { keyPath: 'id' });
      aircraftStore.createIndex('by-status', 'status');
      aircraftStore.createIndex('by-tenant', 'tenantId');

      // Tasks store
      const tasksStore = db.createObjectStore('tasks', { keyPath: 'id' });
      tasksStore.createIndex('by-aircraft', 'aircraftId');
      tasksStore.createIndex('by-mechanic', 'assignedMechanicId');
      tasksStore.createIndex('by-status', 'status');
      tasksStore.createIndex('by-due-date', 'dueDate');

      // Sync queue store
      const syncStore = db.createObjectStore('syncQueue', { keyPath: 'id' });
      syncStore.createIndex('by-priority', 'priority');
      syncStore.createIndex('by-timestamp', 'timestamp');
      syncStore.createIndex('by-type', 'type');
    };
  });
}
```

### 10.5 Sync Queue Management

```typescript
// Sync queue manager
class SyncQueueManager {
  private queue: SyncQueueItem[] = [];
  private isProcessing = false;

  async add(item: Omit<SyncQueueItem, 'id' | 'attempts'>): Promise<void> {
    const queueItem: SyncQueueItem = {
      ...item,
      id: crypto.randomUUID(),
      attempts: 0,
      createdAt: Date.now(),
    };

    await this.persistToIndexedDB(queueItem);
    this.updatePendingIndicator();

    if (navigator.onLine) {
      this.processQueue();
    }
  }

  async processQueue(): Promise<void> {
    if (this.isProcessing) return;
    this.isProcessing = true;

    const items = await this.getQueuedItems();
    const sorted = items.sort((a, b) => {
      // Priority first, then timestamp
      if (a.priority !== b.priority) {
        return this.priorityValue(b.priority) - this.priorityValue(a.priority);
      }
      return a.timestamp - b.timestamp;
    });

    for (const item of sorted) {
      try {
        await this.processItem(item);
        await this.removeFromQueue(item.id);
      } catch (error) {
        await this.handleSyncError(item, error);
      }
    }

    this.isProcessing = false;
    this.updatePendingIndicator();
  }

  private updatePendingIndicator(): void {
    const count = this.queue.length;
    const indicator = document.getElementById('pending-sync-indicator');

    if (indicator) {
      indicator.textContent = count > 0 ? `${count} pending` : '';
      indicator.hidden = count === 0;
    }
  }
}
```

### 10.6 Conflict Resolution UI

```typescript
// Conflict resolution component
interface ConflictResolutionProps {
  conflict: SyncConflict;
  onResolve: (resolution: 'local' | 'server' | 'merge') => void;
}

const ConflictResolutionModal: React.FC<ConflictResolutionProps> = ({
  conflict,
  onResolve,
}) => {
  return (
    <Modal
      title="Sync Conflict Detected"
      aria-describedby="conflict-description"
    >
      <div id="conflict-description" className="conflict-info">
        <p>
          Your offline changes conflict with updates made on the server.
          Please choose how to resolve this conflict.
        </p>

        <div className="conflict-comparison">
          <div className="conflict-version">
            <h4>Your Changes (Offline)</h4>
            <pre>{JSON.stringify(conflict.localData, null, 2)}</pre>
            <time>Modified: {formatDateTime(conflict.localTimestamp)}</time>
          </div>

          <div className="conflict-version">
            <h4>Server Version</h4>
            <pre>{JSON.stringify(conflict.serverData, null, 2)}</pre>
            <time>Modified: {formatDateTime(conflict.serverTimestamp)}</time>
            <p>Changed by: {conflict.serverModifiedBy}</p>
          </div>
        </div>
      </div>

      <div className="conflict-actions">
        <Button variant="secondary" onClick={() => onResolve('local')}>
          Keep My Changes
        </Button>
        <Button variant="secondary" onClick={() => onResolve('server')}>
          Use Server Version
        </Button>
        <Button variant="primary" onClick={() => onResolve('merge')}>
          Review & Merge
        </Button>
      </div>
    </Modal>
  );
};
```

### 10.7 Offline Status Indicators

```typescript
// Offline status banner component
const OfflineStatusBanner: React.FC = () => {
  const { isOnline, pendingChanges, lastSyncTime } = useOfflineStatus();

  if (isOnline && pendingChanges === 0) {
    return null;
  }

  return (
    <div
      className={`offline-banner ${isOnline ? 'syncing' : 'offline'}`}
      role="status"
      aria-live="polite"
    >
      {!isOnline ? (
        <>
          <OfflineIcon aria-hidden="true" />
          <span>You are offline. Changes will sync when reconnected.</span>
        </>
      ) : (
        <>
          <SyncIcon className="spinning" aria-hidden="true" />
          <span>Syncing {pendingChanges} pending changes...</span>
        </>
      )}

      {lastSyncTime && (
        <span className="last-sync">
          Last synced: {formatRelativeTime(lastSyncTime)}
        </span>
      )}
    </div>
  );
};

// Icon state variations
const getIconState = (isOnline: boolean, hasPending: boolean): IconState => {
  if (!isOnline) return 'offline'; // Gray with slash
  if (hasPending) return 'syncing'; // Animated
  return 'synced'; // Green check
};
```

### 10.8 Background Sync Triggers

```typescript
// Background sync registration
async function registerBackgroundSync(): Promise<void> {
  if ('serviceWorker' in navigator && 'sync' in ServiceWorkerRegistration.prototype) {
    const registration = await navigator.serviceWorker.ready;

    // Register periodic sync for critical data refresh
    if ('periodicSync' in registration) {
      try {
        await registration.periodicSync.register('refresh-critical-data', {
          minInterval: 15 * 60 * 1000, // 15 minutes
        });
      } catch {
        // Periodic sync not available, fall back to manual refresh
      }
    }

    // Listen for online status
    window.addEventListener('online', async () => {
      await registration.sync.register('sync-offline-changes');
    });
  }
}

// Service worker sync handler
self.addEventListener('sync', (event: SyncEvent) => {
  if (event.tag === 'sync-offline-changes') {
    event.waitUntil(syncOfflineChanges());
  }
});

self.addEventListener('periodicsync', (event: PeriodicSyncEvent) => {
  if (event.tag === 'refresh-critical-data') {
    event.waitUntil(refreshCriticalData());
  }
});
```

### 10.9 Storage Quota Management

```typescript
// Storage quota monitoring
class StorageManager {
  async checkQuota(): Promise<StorageQuota> {
    if ('storage' in navigator && 'estimate' in navigator.storage) {
      const estimate = await navigator.storage.estimate();
      return {
        used: estimate.usage || 0,
        available: estimate.quota || 0,
        percentUsed: ((estimate.usage || 0) / (estimate.quota || 1)) * 100,
      };
    }
    return { used: 0, available: 0, percentUsed: 0 };
  }

  async requestPersistence(): Promise<boolean> {
    if ('storage' in navigator && 'persist' in navigator.storage) {
      return navigator.storage.persist();
    }
    return false;
  }

  async cleanupOldData(): Promise<void> {
    const quota = await this.checkQuota();

    if (quota.percentUsed > 80) {
      // Remove old cached API responses
      const cache = await caches.open('api-cache-v1');
      const keys = await cache.keys();
      const oldEntries = keys.slice(0, Math.floor(keys.length * 0.3));
      await Promise.all(oldEntries.map((key) => cache.delete(key)));

      // Remove old sync queue items (already synced)
      await this.pruneOldSyncItems();
    }
  }
}
```

---

## Part 11: Internationalization (i18n)

### 11.1 Language Support Strategy

#### 11.1.1 Initial Language Support

| Language | Locale Code | Status | Priority |
|----------|-------------|--------|----------|
| English (US) | `en-US` | Primary | Launch |
| English (UK) | `en-GB` | Variant | Launch |
| Spanish | `es` | Secondary | Launch |
| French | `fr` | Planned | Phase 2 |
| German | `de` | Planned | Phase 2 |
| Portuguese (Brazil) | `pt-BR` | Planned | Phase 3 |
| Arabic | `ar` | Planned | Phase 3 (RTL) |

#### 11.1.2 Locale Detection and Selection

```typescript
// Locale detection priority
const detectLocale = (): string => {
  // 1. User preference (stored)
  const storedLocale = localStorage.getItem('amss-locale');
  if (storedLocale && supportedLocales.includes(storedLocale)) {
    return storedLocale;
  }

  // 2. URL parameter (for sharing links)
  const urlLocale = new URLSearchParams(window.location.search).get('lang');
  if (urlLocale && supportedLocales.includes(urlLocale)) {
    return urlLocale;
  }

  // 3. Browser language
  const browserLocale = navigator.language;
  const matchedLocale = findBestMatch(browserLocale, supportedLocales);
  if (matchedLocale) {
    return matchedLocale;
  }

  // 4. Default fallback
  return 'en-US';
};
```

### 11.2 Translation File Structure

```
/locales
  /en-US
    common.json          # Shared UI elements
    navigation.json      # Menu and navigation
    aircraft.json        # Aircraft management
    tasks.json           # Maintenance tasks
    parts.json           # Parts inventory
    compliance.json      # Compliance and auditing
    errors.json          # Error messages
    validation.json      # Form validation messages
  /es
    common.json
    navigation.json
    ...
```

#### 11.2.1 Translation File Format

```json
// /locales/en-US/aircraft.json
{
  "aircraft": {
    "title": "Aircraft",
    "list": {
      "title": "Aircraft Inventory",
      "empty": "No aircraft found",
      "searchPlaceholder": "Search by tail number or type..."
    },
    "status": {
      "operational": "Operational",
      "maintenance": "In Maintenance",
      "grounded": "Grounded"
    },
    "fields": {
      "tailNumber": "Tail Number",
      "aircraftType": "Aircraft Type",
      "totalFlightHours": "Total Flight Hours",
      "totalCycles": "Total Cycles",
      "lastMaintenance": "Last Maintenance"
    },
    "actions": {
      "add": "Add Aircraft",
      "edit": "Edit Aircraft",
      "ground": "Ground Aircraft",
      "returnToService": "Return to Service"
    },
    "confirmations": {
      "ground": {
        "title": "Confirm Aircraft Grounding",
        "message": "Are you sure you want to ground {{tailNumber}}? This will cancel all scheduled flights."
      }
    }
  }
}
```

```json
// /locales/es/aircraft.json
{
  "aircraft": {
    "title": "Aeronaves",
    "list": {
      "title": "Inventario de Aeronaves",
      "empty": "No se encontraron aeronaves",
      "searchPlaceholder": "Buscar por matrícula o tipo..."
    },
    "status": {
      "operational": "Operativa",
      "maintenance": "En Mantenimiento",
      "grounded": "Fuera de Servicio"
    }
  }
}
```

### 11.3 Date/Time Formatting

#### 11.3.1 Aviation Date/Time Standards

```typescript
// Date/time formatting utilities
const DateTimeFormatter = {
  // Aviation records MUST show UTC
  formatUTC(date: Date, locale: string): string {
    return new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      timeZone: 'UTC',
      timeZoneName: 'short',
    }).format(date);
    // Output: "Jan 15, 2024, 14:30 UTC"
  },

  // Local time for scheduling UI
  formatLocal(date: Date, locale: string, timeZone: string): string {
    return new Intl.DateTimeFormat(locale, {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      timeZone,
      timeZoneName: 'short',
    }).format(date);
  },

  // Dual display for critical records
  formatDual(date: Date, locale: string, localTimeZone: string): DualTime {
    return {
      utc: this.formatUTC(date, locale),
      local: this.formatLocal(date, locale, localTimeZone),
    };
  },

  // Relative time (for recent activity)
  formatRelative(date: Date, locale: string): string {
    const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
    const diff = Date.now() - date.getTime();
    const minutes = Math.floor(diff / 60000);

    if (minutes < 60) return rtf.format(-minutes, 'minute');
    if (minutes < 1440) return rtf.format(-Math.floor(minutes / 60), 'hour');
    return rtf.format(-Math.floor(minutes / 1440), 'day');
  },
};
```

### 11.4 Number Formatting

```typescript
// Number formatting for aviation-specific values
const NumberFormatter = {
  // Flight hours (always 1 decimal place)
  formatFlightHours(hours: number, locale: string): string {
    return new Intl.NumberFormat(locale, {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
    }).format(hours) + ' hrs';
    // en-US: "1,234.5 hrs"
    // es: "1.234,5 hrs"
  },

  // Cycles (whole numbers with grouping)
  formatCycles(cycles: number, locale: string): string {
    return new Intl.NumberFormat(locale, {
      maximumFractionDigits: 0,
    }).format(cycles);
    // en-US: "1,234"
    // es: "1.234"
  },

  // Part quantities
  formatQuantity(quantity: number, locale: string): string {
    return new Intl.NumberFormat(locale).format(quantity);
  },

  // Percentages
  formatPercent(value: number, locale: string): string {
    return new Intl.NumberFormat(locale, {
      style: 'percent',
      minimumFractionDigits: 0,
      maximumFractionDigits: 1,
    }).format(value);
  },
};
```

### 11.5 Currency Formatting

```typescript
// Currency formatting for maintenance costs
const CurrencyFormatter = {
  format(amount: number, currency: string, locale: string): string {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
    // en-US, USD: "$1,234.56"
    // es, EUR: "1.234,56 €"
  },

  // Compact format for large values
  formatCompact(amount: number, currency: string, locale: string): string {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency,
      notation: 'compact',
      compactDisplay: 'short',
    }).format(amount);
    // en-US: "$1.2K"
  },
};
```

### 11.6 Pluralization Rules

```typescript
// i18n configuration with pluralization
// Using react-i18next or similar
const i18nConfig = {
  interpolation: {
    escapeValue: false,
  },
  pluralSeparator: '_',
  contextSeparator: '_',
};

// Translation file with plurals
// /locales/en-US/tasks.json
{
  "tasks": {
    "count_zero": "No tasks",
    "count_one": "{{count}} task",
    "count_other": "{{count}} tasks",

    "dueIn_zero": "Due today",
    "dueIn_one": "Due in {{count}} day",
    "dueIn_other": "Due in {{count}} days"
  }
}

// /locales/es/tasks.json
{
  "tasks": {
    "count_zero": "Sin tareas",
    "count_one": "{{count}} tarea",
    "count_other": "{{count}} tareas"
  }
}

// Usage
t('tasks.count', { count: taskCount });
```

### 11.7 RTL Layout Considerations

```scss
// RTL-aware styling
[dir='rtl'] {
  // Flip horizontal margins/paddings
  .sidebar {
    right: 0;
    left: auto;
    border-left: 1px solid var(--border-color);
    border-right: none;
  }

  // Flip icons that indicate direction
  .icon-chevron-right {
    transform: scaleX(-1);
  }

  // Adjust text alignment
  .form-label {
    text-align: right;
  }

  // Flip flex direction where needed
  .breadcrumb {
    flex-direction: row-reverse;
  }
}

// Logical properties (modern approach)
.card {
  margin-inline-start: 16px;  // margin-left in LTR, margin-right in RTL
  margin-inline-end: 8px;
  padding-inline: 16px;
  border-inline-start: 4px solid var(--accent);
}
```

### 11.8 Aviation Terminology Consistency

```json
// /locales/en-US/glossary.json
{
  "glossary": {
    "terms": {
      "AD": {
        "short": "AD",
        "long": "Airworthiness Directive",
        "description": "Mandatory safety modification or inspection"
      },
      "MEL": {
        "short": "MEL",
        "long": "Minimum Equipment List",
        "description": "List of equipment that may be inoperative"
      },
      "TTSN": {
        "short": "TTSN",
        "long": "Total Time Since New",
        "description": "Accumulated operating time since manufacture"
      },
      "TSO": {
        "short": "TSO",
        "long": "Time Since Overhaul",
        "description": "Operating time since last overhaul"
      }
    }
  }
}
```

### 11.9 Translation Workflow

```yaml
# Translation workflow process
workflow:
  extraction:
    tool: i18next-parser
    source: src/**/*.{ts,tsx}
    output: locales/{{lng}}/{{ns}}.json

  translation:
    platform: Crowdin / Lokalise
    process:
      1. Extract new strings from code
      2. Upload to translation platform
      3. Translators work on translations
      4. Review by aviation terminology expert
      5. Download approved translations
      6. Merge into codebase

  quality:
    - Automated checks for missing translations
    - Context screenshots for translators
    - Glossary enforcement for aviation terms
    - Pseudo-localization testing

  updates:
    frequency: Weekly during active development
    hotfix: Same-day for critical user-facing strings
```

### 11.10 Fallback Language Handling

```typescript
// Fallback chain configuration
const i18nFallbacks: FallbackConfig = {
  'en-GB': ['en-US'],
  'es-MX': ['es', 'en-US'],
  'es-AR': ['es', 'en-US'],
  'pt-PT': ['pt-BR', 'en-US'],
  default: ['en-US'],
};

// Missing translation handler
const handleMissingTranslation = (
  key: string,
  locale: string
): string => {
  // Log missing translation for tracking
  console.warn(`Missing translation: ${key} for locale: ${locale}`);

  // In development, show key
  if (process.env.NODE_ENV === 'development') {
    return `[MISSING: ${key}]`;
  }

  // In production, fall back gracefully
  return key.split('.').pop() || key;
};
```

---

## Part 12: Error Handling & Edge Cases

### 12.1 Error Message Patterns

#### 12.1.1 Inline Field Errors

```typescript
// Field-level validation error component
interface FieldErrorProps {
  fieldId: string;
  error: string | null;
}

const FieldError: React.FC<FieldErrorProps> = ({ fieldId, error }) => {
  if (!error) return null;

  return (
    <p
      id={`${fieldId}-error`}
      className="field-error"
      role="alert"
      aria-live="polite"
    >
      <ErrorIcon aria-hidden="true" />
      {error}
    </p>
  );
};

// Usage in form
<div className="form-field">
  <label htmlFor="tail-number">Tail Number</label>
  <input
    id="tail-number"
    aria-invalid={!!errors.tailNumber}
    aria-describedby="tail-number-error tail-number-hint"
    {...register('tailNumber')}
  />
  <FieldError fieldId="tail-number" error={errors.tailNumber?.message} />
</div>
```

#### 12.1.2 Form-Level Errors

```typescript
// Form submission error banner
interface FormErrorBannerProps {
  errors: string[];
  onDismiss?: () => void;
}

const FormErrorBanner: React.FC<FormErrorBannerProps> = ({ errors, onDismiss }) => {
  if (errors.length === 0) return null;

  return (
    <div
      className="form-error-banner"
      role="alert"
      aria-live="assertive"
    >
      <div className="error-header">
        <ErrorIcon aria-hidden="true" />
        <strong>
          {errors.length === 1
            ? 'There was an error with your submission'
            : `There were ${errors.length} errors with your submission`}
        </strong>
        {onDismiss && (
          <button
            type="button"
            aria-label="Dismiss errors"
            onClick={onDismiss}
          >
            <CloseIcon aria-hidden="true" />
          </button>
        )}
      </div>
      <ul className="error-list">
        {errors.map((error, index) => (
          <li key={index}>{error}</li>
        ))}
      </ul>
    </div>
  );
};
```

#### 12.1.3 Toast Notifications

```typescript
// Toast notification system
type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastConfig {
  type: ToastType;
  message: string;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

const toastDurations: Record<ToastType, number> = {
  success: 4000,
  info: 5000,
  warning: 6000,
  error: 8000, // Errors stay longer
};

// Toast component
const Toast: React.FC<ToastConfig & { onClose: () => void }> = ({
  type,
  message,
  action,
  onClose,
}) => {
  return (
    <div
      className={`toast toast-${type}`}
      role={type === 'error' ? 'alert' : 'status'}
      aria-live={type === 'error' ? 'assertive' : 'polite'}
    >
      <ToastIcon type={type} aria-hidden="true" />
      <p className="toast-message">{message}</p>
      {action && (
        <button
          type="button"
          className="toast-action"
          onClick={action.onClick}
        >
          {action.label}
        </button>
      )}
      <button
        type="button"
        className="toast-close"
        aria-label="Dismiss notification"
        onClick={onClose}
      >
        <CloseIcon aria-hidden="true" />
      </button>
    </div>
  );
};
```

#### 12.1.4 Full-Page Errors

```typescript
// 404 Not Found page
const NotFoundPage: React.FC = () => {
  return (
    <main className="error-page">
      <div className="error-content">
        <h1>Page Not Found</h1>
        <p className="error-code">Error 404</p>
        <p className="error-description">
          The page you are looking for does not exist or has been moved.
        </p>
        <div className="error-actions">
          <Button as={Link} to="/dashboard" variant="primary">
            Go to Dashboard
          </Button>
          <Button variant="secondary" onClick={() => window.history.back()}>
            Go Back
          </Button>
        </div>
        <p className="error-help">
          If you believe this is an error, please{' '}
          <a href="/support">contact support</a>.
        </p>
      </div>
    </main>
  );
};

// 500 Server Error page
const ServerErrorPage: React.FC<{ errorId?: string }> = ({ errorId }) => {
  return (
    <main className="error-page">
      <div className="error-content">
        <h1>Something Went Wrong</h1>
        <p className="error-code">Error 500</p>
        <p className="error-description">
          We encountered an unexpected error. Our team has been notified.
        </p>
        {errorId && (
          <p className="error-reference">
            Reference ID: <code>{errorId}</code>
          </p>
        )}
        <div className="error-actions">
          <Button variant="primary" onClick={() => window.location.reload()}>
            Try Again
          </Button>
          <Button as={Link} to="/dashboard" variant="secondary">
            Go to Dashboard
          </Button>
        </div>
      </div>
    </main>
  );
};
```

#### 12.1.5 Modal Errors

```typescript
// Error modal for action failures
const ActionErrorModal: React.FC<{
  error: ActionError;
  onRetry?: () => void;
  onClose: () => void;
}> = ({ error, onRetry, onClose }) => {
  return (
    <Modal
      title="Action Failed"
      variant="error"
      onClose={onClose}
    >
      <div className="modal-body">
        <p>{error.userMessage}</p>
        {error.details && (
          <details className="error-details">
            <summary>Technical Details</summary>
            <pre>{error.details}</pre>
          </details>
        )}
      </div>
      <div className="modal-footer">
        <Button variant="secondary" onClick={onClose}>
          Close
        </Button>
        {onRetry && (
          <Button variant="primary" onClick={onRetry}>
            Try Again
          </Button>
        )}
      </div>
    </Modal>
  );
};
```

### 12.2 Error Message Content Guidelines

| Error Type | Do | Avoid |
|------------|-----|-------|
| Validation | "Tail number must start with 'N' followed by 1-5 digits" | "Invalid format" |
| Permission | "You need scheduler permissions to modify task assignments" | "Access denied" |
| Network | "Unable to connect. Check your internet connection and try again" | "Network error" |
| Server | "Something went wrong on our end. Try again in a few minutes" | "Error 500" |
| Not Found | "This aircraft record may have been deleted or moved" | "404 Not Found" |
| Conflict | "This record was modified by another user. Review changes and try again" | "Conflict error" |

### 12.3 Empty States

#### 12.3.1 No Data Yet (First-Time Use)

```typescript
// Empty state for new tenant with no aircraft
const EmptyAircraftList: React.FC = () => {
  const { hasPermission } = usePermissions();

  return (
    <div className="empty-state">
      <AircraftIllustration aria-hidden="true" />
      <h2>No Aircraft Yet</h2>
      <p>
        Get started by adding your first aircraft to the fleet.
        You can also import aircraft from a CSV file.
      </p>
      {hasPermission('aircraft:create') && (
        <div className="empty-state-actions">
          <Button as={Link} to="/aircraft/new" variant="primary">
            Add Aircraft
          </Button>
          <Button as={Link} to="/aircraft/import" variant="secondary">
            Import from CSV
          </Button>
        </div>
      )}
    </div>
  );
};
```

#### 12.3.2 No Search Results

```typescript
// Empty state for search with no results
const NoSearchResults: React.FC<{ query: string; onClear: () => void }> = ({
  query,
  onClear,
}) => {
  return (
    <div className="empty-state empty-state-search">
      <SearchIllustration aria-hidden="true" />
      <h2>No Results Found</h2>
      <p>
        No aircraft match "<strong>{query}</strong>".
        Try adjusting your search terms or filters.
      </p>
      <div className="empty-state-suggestions">
        <h3>Suggestions:</h3>
        <ul>
          <li>Check the spelling of the tail number</li>
          <li>Try searching by aircraft type instead</li>
          <li>Remove some filters to broaden your search</li>
        </ul>
      </div>
      <Button variant="secondary" onClick={onClear}>
        Clear Search
      </Button>
    </div>
  );
};
```

#### 12.3.3 No Permissions

```typescript
// Empty state for restricted access
const NoPermissionState: React.FC<{ resource: string }> = ({ resource }) => {
  return (
    <div className="empty-state empty-state-permission">
      <LockIllustration aria-hidden="true" />
      <h2>Access Restricted</h2>
      <p>
        You do not have permission to view {resource}.
        Contact your administrator if you need access.
      </p>
      <Button as={Link} to="/dashboard" variant="primary">
        Return to Dashboard
      </Button>
    </div>
  );
};
```

#### 12.3.4 Filtered to Empty

```typescript
// Empty state when filters exclude all results
const FilteredEmptyState: React.FC<{
  totalCount: number;
  onClearFilters: () => void;
}> = ({ totalCount, onClearFilters }) => {
  return (
    <div className="empty-state empty-state-filtered">
      <FilterIllustration aria-hidden="true" />
      <h2>No Matching Results</h2>
      <p>
        Your current filters exclude all {totalCount} items.
        Try adjusting or clearing your filters.
      </p>
      <Button variant="secondary" onClick={onClearFilters}>
        Clear All Filters
      </Button>
    </div>
  );
};
```

### 12.4 Loading States

#### 12.4.1 Initial Page Load (Skeleton)

```typescript
// Skeleton loader for aircraft list
const AircraftListSkeleton: React.FC = () => {
  return (
    <div className="aircraft-list-skeleton" aria-busy="true" aria-label="Loading aircraft list">
      {/* Header skeleton */}
      <div className="skeleton skeleton-header" />

      {/* Filter bar skeleton */}
      <div className="skeleton-filters">
        <div className="skeleton skeleton-input" />
        <div className="skeleton skeleton-button" />
      </div>

      {/* Table skeleton */}
      <div className="skeleton-table">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="skeleton-row">
            <div className="skeleton skeleton-cell" style={{ width: '15%' }} />
            <div className="skeleton skeleton-cell" style={{ width: '25%' }} />
            <div className="skeleton skeleton-cell" style={{ width: '15%' }} />
            <div className="skeleton skeleton-cell" style={{ width: '20%' }} />
            <div className="skeleton skeleton-cell" style={{ width: '10%' }} />
          </div>
        ))}
      </div>
    </div>
  );
};
```

#### 12.4.2 Action in Progress (Button Spinner)

```typescript
// Button with loading state
interface LoadingButtonProps extends ButtonProps {
  isLoading: boolean;
  loadingText?: string;
}

const LoadingButton: React.FC<LoadingButtonProps> = ({
  isLoading,
  loadingText,
  children,
  disabled,
  ...props
}) => {
  return (
    <button
      {...props}
      disabled={disabled || isLoading}
      aria-busy={isLoading}
    >
      {isLoading ? (
        <>
          <Spinner className="btn-spinner" aria-hidden="true" />
          <span>{loadingText || 'Loading...'}</span>
        </>
      ) : (
        children
      )}
    </button>
  );
};

// Usage
<LoadingButton
  isLoading={isSaving}
  loadingText="Saving..."
  onClick={handleSave}
>
  Save Changes
</LoadingButton>
```

#### 12.4.3 Background Refresh (Subtle Indicator)

```typescript
// Subtle refresh indicator in header
const RefreshIndicator: React.FC<{ isRefreshing: boolean }> = ({ isRefreshing }) => {
  if (!isRefreshing) return null;

  return (
    <div
      className="refresh-indicator"
      role="status"
      aria-label="Refreshing data"
    >
      <Spinner size="small" aria-hidden="true" />
      <span className="sr-only">Refreshing...</span>
    </div>
  );
};
```

### 12.5 Permission Denied States

```typescript
// Permission handling strategy
type PermissionStrategy = 'hide' | 'disable' | 'show-message';

const getPermissionStrategy = (
  action: string,
  context: string
): PermissionStrategy => {
  // Navigation items: hide if no permission
  if (context === 'navigation') return 'hide';

  // List actions: disable with tooltip
  if (context === 'list-action') return 'disable';

  // Page content: show access denied message
  if (context === 'page') return 'show-message';

  return 'hide';
};

// Disabled action with explanation
const DisabledAction: React.FC<{
  action: string;
  requiredPermission: string;
}> = ({ action, requiredPermission }) => {
  return (
    <Tooltip content={`You need ${requiredPermission} permission to ${action}`}>
      <button disabled className="btn btn-disabled" aria-disabled="true">
        {action}
      </button>
    </Tooltip>
  );
};
```

### 12.6 Session Timeout Handling

```typescript
// Session timeout manager
const SESSION_WARNING_THRESHOLD = 5 * 60 * 1000; // 5 minutes before expiry

const useSessionTimeout = () => {
  const [showWarning, setShowWarning] = useState(false);
  const [timeRemaining, setTimeRemaining] = useState<number | null>(null);

  useEffect(() => {
    const checkSession = () => {
      const expiresAt = getSessionExpiry();
      const remaining = expiresAt - Date.now();

      if (remaining <= 0) {
        // Session expired - redirect to login
        redirectToLogin('session_expired');
      } else if (remaining <= SESSION_WARNING_THRESHOLD) {
        // Show warning modal
        setShowWarning(true);
        setTimeRemaining(remaining);
      }
    };

    const interval = setInterval(checkSession, 30000); // Check every 30s
    return () => clearInterval(interval);
  }, []);

  const extendSession = async () => {
    await refreshToken();
    setShowWarning(false);
  };

  return { showWarning, timeRemaining, extendSession };
};

// Session warning modal
const SessionWarningModal: React.FC<{
  timeRemaining: number;
  onExtend: () => void;
  onLogout: () => void;
}> = ({ timeRemaining, onExtend, onLogout }) => {
  const minutes = Math.ceil(timeRemaining / 60000);

  return (
    <Modal title="Session Expiring" onClose={onExtend}>
      <div className="modal-body">
        <p>
          Your session will expire in {minutes} minute{minutes !== 1 ? 's' : ''}.
          Would you like to stay logged in?
        </p>
      </div>
      <div className="modal-footer">
        <Button variant="secondary" onClick={onLogout}>
          Log Out
        </Button>
        <Button variant="primary" onClick={onExtend}>
          Stay Logged In
        </Button>
      </div>
    </Modal>
  );
};
```

### 12.7 Network Error Handling

```typescript
// Retry logic with exponential backoff
const fetchWithRetry = async (
  url: string,
  options: RequestInit,
  maxRetries = 3
): Promise<Response> => {
  let lastError: Error;

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch(url, {
        ...options,
        signal: AbortSignal.timeout(10000),
      });

      if (response.ok) return response;

      // Don't retry client errors (4xx)
      if (response.status >= 400 && response.status < 500) {
        throw new ApiError(response.status, await response.json());
      }

      // Retry server errors (5xx)
      throw new NetworkError(`Server error: ${response.status}`);
    } catch (error) {
      lastError = error as Error;

      if (attempt < maxRetries - 1) {
        // Exponential backoff: 1s, 2s, 4s
        await sleep(Math.pow(2, attempt) * 1000);
      }
    }
  }

  throw lastError!;
};
```

### 12.8 Conflict Resolution

```typescript
// Optimistic locking conflict handler
interface ConflictState {
  hasConflict: boolean;
  localVersion: any;
  serverVersion: any;
  serverModifiedBy: string;
  serverModifiedAt: Date;
}

const ConflictResolutionUI: React.FC<{
  conflict: ConflictState;
  onResolve: (choice: 'local' | 'server' | 'merge') => void;
}> = ({ conflict, onResolve }) => {
  return (
    <Modal title="Edit Conflict" variant="warning">
      <div className="conflict-body">
        <Alert variant="warning">
          This record was modified by {conflict.serverModifiedBy} while you were editing.
        </Alert>

        <div className="conflict-versions">
          <div className="version-card">
            <h4>Your Changes</h4>
            <DiffViewer data={conflict.localVersion} />
          </div>
          <div className="version-card">
            <h4>Server Version</h4>
            <DiffViewer data={conflict.serverVersion} />
            <p className="version-meta">
              Modified {formatRelativeTime(conflict.serverModifiedAt)}
            </p>
          </div>
        </div>
      </div>

      <div className="modal-footer">
        <Button variant="secondary" onClick={() => onResolve('server')}>
          Discard My Changes
        </Button>
        <Button variant="secondary" onClick={() => onResolve('local')}>
          Overwrite Server
        </Button>
        <Button variant="primary" onClick={() => onResolve('merge')}>
          Review & Merge
        </Button>
      </div>
    </Modal>
  );
};
```

### 12.9 Rate Limiting Feedback

```typescript
// Rate limit handler
const handleRateLimitResponse = (response: Response): RateLimitInfo => {
  const retryAfter = response.headers.get('Retry-After');
  const remaining = response.headers.get('X-RateLimit-Remaining');
  const resetTime = response.headers.get('X-RateLimit-Reset');

  return {
    retryAfterSeconds: retryAfter ? parseInt(retryAfter, 10) : 60,
    remainingRequests: remaining ? parseInt(remaining, 10) : 0,
    resetTime: resetTime ? new Date(parseInt(resetTime, 10) * 1000) : null,
  };
};

// Rate limit toast
const showRateLimitToast = (info: RateLimitInfo) => {
  toast.warning(
    `Too many requests. Please wait ${info.retryAfterSeconds} seconds before trying again.`,
    { duration: info.retryAfterSeconds * 1000 }
  );
};
```

### 12.10 Concurrent Editing Indicators

```typescript
// Real-time presence indicator
const ConcurrentEditorsBanner: React.FC<{
  editors: User[];
  resourceType: string;
}> = ({ editors, resourceType }) => {
  if (editors.length === 0) return null;

  return (
    <div className="concurrent-editors-banner" role="status">
      <UsersIcon aria-hidden="true" />
      <span>
        {editors.length === 1 ? (
          <>
            <strong>{editors[0].name}</strong> is also editing this {resourceType}
          </>
        ) : (
          <>
            <strong>{editors.length} others</strong> are also editing this {resourceType}
          </>
        )}
      </span>
      <div className="editor-avatars">
        {editors.slice(0, 3).map((editor) => (
          <Avatar key={editor.id} user={editor} size="small" />
        ))}
        {editors.length > 3 && (
          <span className="avatar-overflow">+{editors.length - 3}</span>
        )}
      </div>
    </div>
  );
};
```

---


---

[Next: Aviation Standards & Roadmap](07_AVIATION_ROADMAP.md)
