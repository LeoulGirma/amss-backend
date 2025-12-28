# AMSS UI/UX Specification - Design System

[Back to Index](00_INDEX.md) | [Previous: Vision & Requirements](01_VISION_REQUIREMENTS.md)

---

# Part 3: Design System Foundation

## 3.1 Color Palette

### 3.1.1 Primary Brand Colors

```css
:root {
  /* Primary - Aviation Blue */
  --color-primary-50: #EFF6FF;
  --color-primary-100: #DBEAFE;
  --color-primary-200: #BFDBFE;
  --color-primary-300: #93C5FD;
  --color-primary-400: #60A5FA;
  --color-primary-500: #3B82F6;  /* Primary */
  --color-primary-600: #2563EB;
  --color-primary-700: #1D4ED8;
  --color-primary-800: #1E40AF;
  --color-primary-900: #1E3A8A;

  /* Secondary - Slate */
  --color-secondary-50: #F8FAFC;
  --color-secondary-100: #F1F5F9;
  --color-secondary-200: #E2E8F0;
  --color-secondary-300: #CBD5E1;
  --color-secondary-400: #94A3B8;
  --color-secondary-500: #64748B;  /* Secondary */
  --color-secondary-600: #475569;
  --color-secondary-700: #334155;
  --color-secondary-800: #1E293B;
  --color-secondary-900: #0F172A;
}
```

### 3.1.2 Aircraft Status Colors (Semantic)

```css
:root {
  /* Operational - Green */
  --color-operational: #10B981;
  --color-operational-light: #D1FAE5;
  --color-operational-dark: #065F46;

  /* Maintenance - Amber */
  --color-maintenance: #F59E0B;
  --color-maintenance-light: #FEF3C7;
  --color-maintenance-dark: #92400E;

  /* Grounded - Red */
  --color-grounded: #EF4444;
  --color-grounded-light: #FEE2E2;
  --color-grounded-dark: #991B1B;
}
```

### 3.1.3 Task Status Colors

```css
:root {
  /* Scheduled - Blue */
  --color-scheduled: #3B82F6;
  --color-scheduled-light: #DBEAFE;
  --color-scheduled-dark: #1E40AF;

  /* In Progress - Amber */
  --color-in-progress: #F59E0B;
  --color-in-progress-light: #FEF3C7;
  --color-in-progress-dark: #92400E;

  /* Completed - Green */
  --color-completed: #10B981;
  --color-completed-light: #D1FAE5;
  --color-completed-dark: #065F46;

  /* Cancelled - Gray */
  --color-cancelled: #6B7280;
  --color-cancelled-light: #F3F4F6;
  --color-cancelled-dark: #374151;
}
```

### 3.1.4 Compliance Status Colors

```css
:root {
  /* Pass - Green */
  --color-compliance-pass: #10B981;
  --color-compliance-pass-bg: #D1FAE5;

  /* Fail - Red */
  --color-compliance-fail: #EF4444;
  --color-compliance-fail-bg: #FEE2E2;

  /* Pending - Amber */
  --color-compliance-pending: #F59E0B;
  --color-compliance-pending-bg: #FEF3C7;
}
```

### 3.1.5 Feedback Colors

```css
:root {
  /* Success */
  --color-success: #10B981;
  --color-success-light: #D1FAE5;

  /* Warning */
  --color-warning: #F59E0B;
  --color-warning-light: #FEF3C7;

  /* Error */
  --color-error: #EF4444;
  --color-error-light: #FEE2E2;

  /* Info */
  --color-info: #3B82F6;
  --color-info-light: #DBEAFE;
}
```

## 3.2 Typography

### 3.2.1 Font Stack

```css
:root {
  /* Primary font - UI elements */
  --font-primary: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;

  /* Monospace - Technical data, codes */
  --font-mono: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
```

### 3.2.2 Type Scale

| Token | Size | Line Height | Weight | Usage |
|-------|------|-------------|--------|-------|
| `--text-xs` | 12px | 16px | 400 | Labels, captions |
| `--text-sm` | 14px | 20px | 400 | Body small, table cells |
| `--text-base` | 16px | 24px | 400 | Body text |
| `--text-lg` | 18px | 28px | 500 | Subheadings |
| `--text-xl` | 20px | 28px | 600 | Section titles |
| `--text-2xl` | 24px | 32px | 600 | Page titles |
| `--text-3xl` | 30px | 36px | 700 | Hero text |
| `--text-4xl` | 36px | 40px | 700 | Dashboard KPIs |

### 3.2.3 Typography Components

```css
/* Heading styles */
.heading-page {
  font-size: var(--text-2xl);
  font-weight: 600;
  line-height: 32px;
  color: var(--color-secondary-900);
}

.heading-section {
  font-size: var(--text-xl);
  font-weight: 600;
  line-height: 28px;
  color: var(--color-secondary-800);
}

.heading-card {
  font-size: var(--text-lg);
  font-weight: 500;
  line-height: 28px;
  color: var(--color-secondary-800);
}

/* Body styles */
.body-default {
  font-size: var(--text-base);
  font-weight: 400;
  line-height: 24px;
  color: var(--color-secondary-700);
}

.body-small {
  font-size: var(--text-sm);
  font-weight: 400;
  line-height: 20px;
  color: var(--color-secondary-600);
}

/* Technical data */
.text-mono {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  letter-spacing: -0.02em;
}

/* Tail numbers, part numbers */
.text-identifier {
  font-family: var(--font-mono);
  font-size: var(--text-base);
  font-weight: 600;
  text-transform: uppercase;
}
```

## 3.3 Spacing System

### 3.3.1 Base Unit: 8px Grid

```css
:root {
  --space-0: 0;
  --space-1: 4px;   /* 0.5 unit - tight spacing */
  --space-2: 8px;   /* 1 unit - default tight */
  --space-3: 12px;  /* 1.5 units */
  --space-4: 16px;  /* 2 units - default spacing */
  --space-5: 20px;  /* 2.5 units */
  --space-6: 24px;  /* 3 units - section spacing */
  --space-8: 32px;  /* 4 units - large spacing */
  --space-10: 40px; /* 5 units */
  --space-12: 48px; /* 6 units - page margins */
  --space-16: 64px; /* 8 units */
  --space-20: 80px; /* 10 units */
  --space-24: 96px; /* 12 units */
}
```

### 3.3.2 Component Spacing Tokens

```css
:root {
  /* Padding */
  --padding-button: var(--space-2) var(--space-4);
  --padding-input: var(--space-2) var(--space-3);
  --padding-card: var(--space-4);
  --padding-card-lg: var(--space-6);
  --padding-modal: var(--space-6);
  --padding-page: var(--space-6) var(--space-8);

  /* Gaps */
  --gap-inline: var(--space-2);
  --gap-stack: var(--space-4);
  --gap-section: var(--space-8);
  --gap-grid: var(--space-4);
}
```

## 3.4 Elevation & Shadows

### 3.4.1 Shadow Scale

```css
:root {
  /* Elevation 0 - Flat */
  --shadow-none: none;

  /* Elevation 1 - Slight raise */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);

  /* Elevation 2 - Cards */
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1),
               0 2px 4px -1px rgba(0, 0, 0, 0.06);

  /* Elevation 3 - Dropdowns, popovers */
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1),
               0 4px 6px -2px rgba(0, 0, 0, 0.05);

  /* Elevation 4 - Modals */
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1),
               0 10px 10px -5px rgba(0, 0, 0, 0.04);

  /* Elevation 5 - Notifications, tooltips */
  --shadow-2xl: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}
```

### 3.4.2 Elevation Usage

| Level | Shadow | Use Case |
|-------|--------|----------|
| 0 | none | Inline elements, table rows |
| 1 | sm | Subtle hover states, input fields |
| 2 | md | Cards, tiles, list items |
| 3 | lg | Dropdowns, menus, popovers |
| 4 | xl | Modals, dialogs, sidebars |
| 5 | 2xl | Floating notifications, command palette |

## 3.5 Border Radius

```css
:root {
  --radius-none: 0;
  --radius-sm: 4px;    /* Buttons, inputs */
  --radius-md: 6px;    /* Cards, panels */
  --radius-lg: 8px;    /* Modals, large cards */
  --radius-xl: 12px;   /* Feature cards */
  --radius-2xl: 16px;  /* Hero sections */
  --radius-full: 9999px; /* Pills, avatars */
}
```

## 3.6 Aviation Icons

### 3.6.1 Icon Library

Based on Heroicons with custom aviation extensions:

| Icon | Name | Usage |
|------|------|-------|
| Airplane | `aircraft` | Aircraft list, fleet views |
| Wrench | `maintenance` | Maintenance tasks |
| Clipboard Check | `compliance` | Compliance items |
| Calendar | `schedule` | Scheduling views |
| Box | `parts` | Parts inventory |
| Clock | `timer` | Time-based displays |
| User Group | `team` | Mechanic assignment |
| Shield Check | `verified` | Compliance pass |
| Exclamation Triangle | `warning` | Alerts, overdue items |
| Document | `audit` | Audit logs |

### 3.6.2 Icon Sizes

```css
:root {
  --icon-xs: 12px;  /* Inline with small text */
  --icon-sm: 16px;  /* Inline with body text */
  --icon-md: 20px;  /* Buttons, list items */
  --icon-lg: 24px;  /* Navigation, headers */
  --icon-xl: 32px;  /* Feature icons */
  --icon-2xl: 48px; /* Empty states, hero icons */
}
```

## 3.7 Animation & Motion

### 3.7.1 Duration Scale

```css
:root {
  --duration-instant: 0ms;
  --duration-fast: 100ms;    /* Micro-interactions */
  --duration-normal: 200ms;  /* Standard transitions */
  --duration-slow: 300ms;    /* Complex animations */
  --duration-slower: 500ms;  /* Page transitions */
}
```

### 3.7.2 Easing Functions

```css
:root {
  --ease-linear: linear;
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);
  --ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55);
}
```

### 3.7.3 Standard Transitions

```css
/* Button hover */
.transition-button {
  transition: background-color var(--duration-fast) var(--ease-out),
              transform var(--duration-fast) var(--ease-out);
}

/* Card hover */
.transition-card {
  transition: box-shadow var(--duration-normal) var(--ease-out),
              transform var(--duration-normal) var(--ease-out);
}

/* Sidebar collapse */
.transition-sidebar {
  transition: width var(--duration-slow) var(--ease-in-out);
}

/* Modal appearance */
.transition-modal {
  transition: opacity var(--duration-normal) var(--ease-out),
              transform var(--duration-normal) var(--ease-out);
}

/* Status badge pulse */
@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.animate-pulse {
  animation: pulse 2s var(--ease-in-out) infinite;
}
```

### 3.7.4 Reduced Motion Support

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

## 3.8 Dark Mode

### 3.8.1 Dark Mode Color Tokens

```css
[data-theme="dark"] {
  /* Backgrounds */
  --color-bg-primary: #0F172A;
  --color-bg-secondary: #1E293B;
  --color-bg-tertiary: #334155;
  --color-bg-elevated: #1E293B;

  /* Text */
  --color-text-primary: #F8FAFC;
  --color-text-secondary: #CBD5E1;
  --color-text-tertiary: #94A3B8;
  --color-text-disabled: #64748B;

  /* Borders */
  --color-border-default: #334155;
  --color-border-subtle: #1E293B;
  --color-border-emphasis: #475569;

  /* Status colors (adjusted for dark backgrounds) */
  --color-operational: #34D399;
  --color-maintenance: #FBBF24;
  --color-grounded: #F87171;

  /* Shadows (reduced in dark mode) */
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.3);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.4);
}
```

### 3.8.2 Dark Mode Implementation

```css
/* System preference detection */
@media (prefers-color-scheme: dark) {
  :root:not([data-theme="light"]) {
    /* Apply dark theme tokens */
  }
}

/* Manual toggle */
[data-theme="dark"] {
  /* Apply dark theme tokens */
}

[data-theme="light"] {
  /* Apply light theme tokens */
}
```

---

[Next: Information Architecture](03_INFORMATION_ARCHITECTURE.md)
