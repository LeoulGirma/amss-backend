# AMSS API & User Guide

**Complete guide for using the AMSS (Aircraft Maintenance Scheduling System) API**

This guide is for:
- API consumers integrating with AMSS
- Frontend developers building UIs
- System integrators setting up webhooks
- End users understanding the platform

---

## Table of Contents

1. [Getting Started](#1-getting-started)
2. [Authentication](#2-authentication)
3. [Core Workflows](#3-core-workflows)
4. [API Reference](#4-api-reference)
5. [Webhook Integration](#5-webhook-integration)
6. [Import/Export Guide](#6-importexport-guide)
7. [Troubleshooting & FAQ](#7-troubleshooting--faq)
8. [Rate Limiting & Best Practices](#8-rate-limiting--best-practices)

---

## 1. Getting Started

### Base URL

**UAT Environment:**
```
https://amss-api-uat.duckdns.org
```

**Production Environment:**
```
https://amss-api.yourdomain.com
```

### API Endpoints

- **REST API**: `https://amss-api-uat.duckdns.org/api/v1/*`
- **gRPC API**: `amss-api-uat.duckdns.org:443` (TLS required)
- **OpenAPI Spec**: `https://amss-api-uat.duckdns.org/openapi.yaml`
- **Swagger UI**: `https://amss-api-uat.duckdns.org/docs`

### Health Checks

```bash
# Check if API is operational
curl https://amss-api-uat.duckdns.org/health
# Response: "ok"

# Check if API can connect to database
curl https://amss-api-uat.duckdns.org/ready
# Response: "ready"
```

### Quick Start Example

```bash
# 1. Login
curl -X POST https://amss-api-uat.duckdns.org/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "your-org-uuid",
    "email": "admin@yourcompany.com",
    "password": "YourPassword123!"
  }'

# Response:
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}

# 2. Use access token in subsequent requests
curl https://amss-api-uat.duckdns.org/api/v1/aircraft \
  -H "Authorization: Bearer <access_token>"
```

---

## 2. Authentication

### Login Flow

AMSS uses **JWT (JSON Web Tokens)** for authentication.

#### Step 1: Get Access Token

```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@company.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLWlkIiwib3JnX2lkIjoib3JnLWlkIiwiZXhwIjoxNzA2MTIzNDU2fQ...",
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLWlkIiwidHlwZSI6InJlZnJlc2giLCJleHAiOjE3MDY3MjgyNTZ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Token Details:**
- **access_token**: Use for API requests (expires in 15 minutes)
- **refresh_token**: Use to get new access token (expires in 7 days)
- **expires_in**: Seconds until access token expires (900 = 15 minutes)

#### Step 2: Use Access Token

```bash
GET /api/v1/organizations
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

**All API requests** (except login) require the `Authorization` header:
```
Authorization: Bearer <access_token>
```

#### Step 3: Refresh Token When Expired

When you receive `401 Unauthorized` with `"error": "token expired"`:

```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

---

### User Roles & Permissions

AMSS implements **Role-Based Access Control (RBAC)**. Each user has one role:

| Role | Permissions | Typical Use Case |
|------|------------|------------------|
| **admin** | Full access to all resources | System administrators |
| **planner** | Create/update aircraft, programs, tasks | Maintenance planners |
| **technician** | Execute tasks, reserve parts, log work | Field technicians |
| **viewer** | Read-only access | Auditors, management |

**Permission Examples:**

```
Admin:
✅ Create users
✅ Create aircraft
✅ Create maintenance programs
✅ Execute tasks
✅ View reports

Planner:
❌ Create users
✅ Create aircraft
✅ Create maintenance programs
✅ Assign tasks
✅ View reports

Technician:
❌ Create users
❌ Create aircraft
❌ Create programs
✅ Execute assigned tasks
✅ Reserve parts
✅ Log completion

Viewer:
❌ Create/update anything
✅ View all resources
✅ View reports
```

---

## 3. Core Workflows

### Workflow 1: User Management

#### Create a User

```bash
POST /api/v1/users
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "technician@company.com",
  "password": "SecurePass123!",
  "role": "technician",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "id": "user-uuid",
  "org_id": "org-uuid",
  "email": "technician@company.com",
  "role": "technician",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2025-01-24T10:30:00Z"
}
```

**Password Requirements:**
- Minimum 8 characters
- At least one uppercase letter
- At least one number
- At least one special character

#### List Users

```bash
GET /api/v1/users
Authorization: Bearer <token>
```

**Response:**
```json
{
  "users": [
    {
      "id": "user-uuid",
      "email": "admin@company.com",
      "role": "admin",
      "created_at": "2025-01-20T09:00:00Z"
    }
  ],
  "total": 1
}
```

#### Update User Role

```bash
PUT /api/v1/users/{user_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "planner"
}
```

---

### Workflow 2: Aircraft Management

#### Add Aircraft

```bash
POST /api/v1/aircraft
Authorization: Bearer <token>
Content-Type: application/json

{
  "registration": "N12345",
  "serial_number": "SN-987654",
  "manufacturer": "Cessna",
  "model": "172S",
  "current_hours": 1500.5,
  "current_cycles": 800,
  "metadata": {
    "year": 2018,
    "color": "white",
    "home_base": "KJFK"
  }
}
```

**Response:**
```json
{
  "id": "aircraft-uuid",
  "org_id": "org-uuid",
  "registration": "N12345",
  "serial_number": "SN-987654",
  "manufacturer": "Cessna",
  "model": "172S",
  "current_hours": 1500.5,
  "current_cycles": 800,
  "metadata": {
    "year": 2018,
    "color": "white",
    "home_base": "KJFK"
  },
  "created_at": "2025-01-24T10:35:00Z",
  "updated_at": "2025-01-24T10:35:00Z"
}
```

#### Update Aircraft Hours/Cycles

```bash
PUT /api/v1/aircraft/{aircraft_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_hours": 1605.2,
  "current_cycles": 815
}
```

**Note**: Updating hours/cycles triggers the system to check if new maintenance tasks are due.

#### Get Aircraft Details

```bash
GET /api/v1/aircraft/{aircraft_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "aircraft-uuid",
  "registration": "N12345",
  "current_hours": 1605.2,
  "current_cycles": 815,
  "upcoming_tasks": [
    {
      "id": "task-uuid",
      "program_name": "100-Hour Inspection",
      "due_at_hours": 1700.0,
      "status": "scheduled"
    }
  ]
}
```

#### List All Aircraft

```bash
GET /api/v1/aircraft?page=1&limit=20
Authorization: Bearer <token>
```

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Results per page (max: 100, default: 20)
- `registration`: Filter by registration (partial match)
- `manufacturer`: Filter by manufacturer

**Response:**
```json
{
  "aircraft": [...],
  "total": 15,
  "page": 1,
  "limit": 20,
  "pages": 1
}
```

---

### Workflow 3: Maintenance Programs

#### Create Maintenance Program

```bash
POST /api/v1/programs
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "100-Hour Inspection",
  "description": "Routine 100-hour inspection per FAR 91.409(b)",
  "interval_type": "hours",
  "interval_value": 100,
  "requirements": [
    "Oil change",
    "Spark plug inspection",
    "Compression check",
    "Airframe inspection"
  ]
}
```

**Interval Types:**
- `hours`: Based on flight hours
- `cycles`: Based on takeoff/landing cycles
- `days`: Calendar-based (e.g., annual inspection)

**Response:**
```json
{
  "id": "program-uuid",
  "name": "100-Hour Inspection",
  "interval_type": "hours",
  "interval_value": 100,
  "created_at": "2025-01-24T10:40:00Z"
}
```

#### Link Program to Aircraft

```bash
POST /api/v1/aircraft/{aircraft_id}/programs
Authorization: Bearer <token>
Content-Type: application/json

{
  "program_id": "program-uuid",
  "last_completed_hours": 1500.0,
  "last_completed_date": "2025-01-01"
}
```

**Effect**: System auto-generates first task due at `1500 + 100 = 1600 hours`

---

### Workflow 4: Task Execution

#### View Assigned Tasks

```bash
GET /api/v1/tasks?assigned_to={user_id}&status=scheduled
Authorization: Bearer <token>
```

**Response:**
```json
{
  "tasks": [
    {
      "id": "task-uuid",
      "aircraft": {
        "id": "aircraft-uuid",
        "registration": "N12345",
        "current_hours": 1595.0
      },
      "program": {
        "id": "program-uuid",
        "name": "100-Hour Inspection"
      },
      "status": "scheduled",
      "due_at_hours": 1600.0,
      "is_overdue": false,
      "assigned_to": {
        "id": "user-uuid",
        "email": "technician@company.com"
      },
      "required_parts": [
        {
          "part_number": "OIL-FILTER-123",
          "quantity": 1
        }
      ]
    }
  ]
}
```

#### Reserve Parts for Task

```bash
POST /api/v1/tasks/{task_id}/reserve-parts
Authorization: Bearer <token>
Content-Type: application/json

{
  "part_reservations": [
    {
      "part_id": "part-uuid",
      "quantity": 1
    }
  ]
}
```

**Response:**
```json
{
  "task_id": "task-uuid",
  "status": "in_progress",
  "reservations": [
    {
      "id": "reservation-uuid",
      "part_id": "part-uuid",
      "part_number": "OIL-FILTER-123",
      "quantity": 1,
      "status": "reserved"
    }
  ]
}
```

**Error Handling:**
```json
// If insufficient inventory:
{
  "error": "insufficient inventory",
  "message": "Part OIL-FILTER-123 has only 0 available (1 requested)",
  "code": "INSUFFICIENT_INVENTORY"
}
```

#### Complete Task

```bash
PUT /api/v1/tasks/{task_id}/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "completed_at_hours": 1605.2,
  "completed_at_cycles": 815,
  "completion_notes": "Replaced oil filter, changed oil. Spark plugs checked - all within spec. No discrepancies noted.",
  "attachments": [
    {
      "type": "photo",
      "url": "https://storage.example.com/photo1.jpg",
      "description": "Oil filter condition"
    }
  ]
}
```

**Response:**
```json
{
  "id": "task-uuid",
  "status": "completed",
  "completed_at": "2025-01-24T14:30:00Z",
  "completed_by": {
    "id": "user-uuid",
    "email": "technician@company.com"
  },
  "completed_at_hours": 1605.2,
  "next_task": {
    "id": "next-task-uuid",
    "due_at_hours": 1705.2,
    "status": "scheduled"
  }
}
```

**Note**: Completing a task automatically:
1. Marks part reservations as consumed
2. Decrements inventory
3. Creates audit log entry
4. Triggers webhook (if configured)
5. Generates next task based on program interval

---

### Workflow 5: Parts Inventory

#### Add Part Definition

```bash
POST /api/v1/parts
Authorization: Bearer <token>
Content-Type: application/json

{
  "part_number": "OIL-FILTER-CH48110",
  "description": "Champion Oil Filter CH48110 for Lycoming engines",
  "manufacturer": "Champion Aerospace",
  "unit_price": 12.50,
  "min_stock_level": 5,
  "category": "filters"
}
```

#### Add Inventory Item

```bash
POST /api/v1/inventory
Authorization: Bearer <token>
Content-Type: application/json

{
  "part_id": "part-uuid",
  "location": "Hangar 3 - Shelf A2",
  "quantity_on_hand": 10
}
```

#### Check Inventory Availability

```bash
GET /api/v1/inventory?part_number=OIL-FILTER-CH48110
Authorization: Bearer <token>
```

**Response:**
```json
{
  "items": [
    {
      "id": "inventory-uuid",
      "part_id": "part-uuid",
      "part_number": "OIL-FILTER-CH48110",
      "location": "Hangar 3 - Shelf A2",
      "quantity_on_hand": 10,
      "quantity_reserved": 2,
      "quantity_available": 8,
      "min_stock_level": 5,
      "is_low_stock": false
    }
  ]
}
```

**Calculated Fields:**
- `quantity_available = quantity_on_hand - quantity_reserved`
- `is_low_stock = quantity_available < min_stock_level`

---

### Workflow 6: Reports & Compliance

#### Get Compliance Summary

```bash
GET /api/v1/reports/compliance
Authorization: Bearer <token>
```

**Response:**
```json
{
  "org_id": "org-uuid",
  "period": {
    "start": "2025-01-01",
    "end": "2025-01-31"
  },
  "summary": {
    "total_aircraft": 15,
    "total_tasks_due": 45,
    "tasks_completed_on_time": 40,
    "tasks_overdue": 5,
    "compliance_rate": 88.9
  },
  "by_aircraft": [
    {
      "aircraft_id": "aircraft-uuid",
      "registration": "N12345",
      "tasks_due": 3,
      "tasks_completed": 2,
      "tasks_overdue": 1
    }
  ]
}
```

#### Get Technician Workload Report

```bash
GET /api/v1/reports/workload?user_id={technician_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "user_id": "user-uuid",
  "email": "technician@company.com",
  "workload": {
    "tasks_assigned": 8,
    "tasks_in_progress": 2,
    "tasks_completed_this_month": 12,
    "avg_completion_time_hours": 2.5
  }
}
```

#### Get Inventory Status Report

```bash
GET /api/v1/reports/inventory?low_stock=true
Authorization: Bearer <token>
```

**Response:**
```json
{
  "low_stock_items": [
    {
      "part_number": "SPARK-PLUG-REM37BY",
      "description": "Champion Spark Plug REM37BY",
      "quantity_available": 2,
      "min_stock_level": 8,
      "reorder_needed": 6
    }
  ],
  "total_low_stock_items": 1
}
```

---

## 4. API Reference

### Common Request Headers

```
Authorization: Bearer <access_token>      # Required for all authenticated endpoints
Content-Type: application/json            # For POST/PUT requests
Idempotency-Key: <uuid>                   # Optional: Prevents duplicate operations
```

### Common Response Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Invalid request format or validation error |
| 401 | Unauthorized | Missing or invalid access token |
| 403 | Forbidden | User lacks permission for this operation |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource conflict (e.g., duplicate, idempotency) |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error (contact support) |

### Error Response Format

```json
{
  "error": "validation_failed",
  "message": "Invalid aircraft registration format",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "registration",
    "value": "12345",
    "constraint": "must start with letter"
  }
}
```

### Pagination

List endpoints support pagination:

```bash
GET /api/v1/aircraft?page=2&limit=50
```

**Response:**
```json
{
  "data": [...],
  "total": 150,
  "page": 2,
  "limit": 50,
  "pages": 3
}
```

### Filtering & Sorting

```bash
# Filter by status
GET /api/v1/tasks?status=scheduled

# Filter by date range
GET /api/v1/tasks?created_after=2025-01-01&created_before=2025-01-31

# Sort results
GET /api/v1/aircraft?sort=registration&order=asc

# Combine filters
GET /api/v1/tasks?assigned_to={user_id}&status=in_progress&sort=due_at_hours
```

### Idempotency

**Use idempotency keys for safe retries:**

```bash
POST /api/v1/tasks
Authorization: Bearer <token>
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{
  "aircraft_id": "aircraft-uuid",
  "program_id": "program-uuid"
}
```

**Benefits:**
- Network failures: Safe to retry
- Duplicate clicks: Won't create duplicate tasks
- Cached for 24 hours

**When to use:**
- All write operations (POST, PUT, DELETE)
- Payment-related operations
- Critical business operations

**When NOT needed:**
- Read operations (GET)
- Operations that are naturally idempotent (e.g., updating to same value)

---

## 5. Webhook Integration

### Overview

Webhooks allow AMSS to notify your systems when events occur.

**Delivery Guarantee**: At-least-once (may receive duplicates, use idempotency)

### Supported Events

| Event | Trigger | Payload |
|-------|---------|---------|
| `task.created` | New task generated | Task object |
| `task.assigned` | Task assigned to user | Task + user |
| `task.completed` | Task marked complete | Task + completion data |
| `part.consumed` | Part used in task | Part + quantity |
| `aircraft.created` | New aircraft added | Aircraft object |
| `inventory.low_stock` | Inventory below threshold | Part + current quantity |
| `import.completed` | CSV import finished | Import job + results |

### Setting Up Webhooks

#### Step 1: Create Webhook Subscription

```bash
POST /api/v1/webhooks
Authorization: Bearer <token>
Content-Type: application/json

{
  "url": "https://your-system.com/webhooks/amss",
  "events": ["task.completed", "part.consumed"],
  "secret": "your-webhook-secret-key-keep-this-safe",
  "description": "Notify billing system"
}
```

**Response:**
```json
{
  "id": "webhook-uuid",
  "url": "https://your-system.com/webhooks/amss",
  "events": ["task.completed", "part.consumed"],
  "active": true,
  "created_at": "2025-01-24T11:00:00Z"
}
```

**⚠️ Security Note**: Store the `secret` securely. You'll use it to validate webhook signatures.

---

#### Step 2: Implement Webhook Endpoint

**Your Server** (`https://your-system.com/webhooks/amss`):

```javascript
// Example: Node.js/Express
const express = require('express');
const crypto = require('crypto');

app.post('/webhooks/amss', express.json(), (req, res) => {
  // 1. Verify signature
  const signature = req.headers['x-amss-signature'];
  const payload = JSON.stringify(req.body);
  const secret = process.env.AMSS_WEBHOOK_SECRET;

  const expectedSignature = 'sha256=' +
    crypto.createHmac('sha256', secret)
      .update(payload)
      .digest('hex');

  if (signature !== expectedSignature) {
    console.error('Invalid webhook signature');
    return res.status(401).send('Unauthorized');
  }

  // 2. Process event
  const { event, data } = req.body;

  switch (event) {
    case 'task.completed':
      console.log(`Task ${data.id} completed by ${data.completed_by.email}`);
      // Create invoice, update billing system, etc.
      break;

    case 'part.consumed':
      console.log(`Part ${data.part_number} consumed (qty: ${data.quantity})`);
      // Update inventory system, trigger reorder, etc.
      break;
  }

  // 3. Return 200 OK to acknowledge receipt
  res.status(200).send('OK');
});
```

**Python/Flask Example:**

```python
import hashlib
import hmac
from flask import Flask, request

app = Flask(__name__)

@app.route('/webhooks/amss', methods=['POST'])
def handle_webhook():
    # 1. Verify signature
    signature = request.headers.get('X-AMSS-Signature')
    payload = request.get_data()
    secret = os.environ['AMSS_WEBHOOK_SECRET']

    expected_signature = 'sha256=' + hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()

    if signature != expected_signature:
        return 'Unauthorized', 401

    # 2. Process event
    event_data = request.get_json()
    event_type = event_data['event']
    data = event_data['data']

    if event_type == 'task.completed':
        print(f"Task {data['id']} completed")
        # Your business logic here

    # 3. Acknowledge
    return 'OK', 200
```

---

#### Step 3: Test Webhook

```bash
# Create a test task to trigger webhook
POST /api/v1/tasks/{task_id}/complete
Authorization: Bearer <token>
Content-Type: application/json

{
  "completed_at_hours": 1605.2,
  "completion_notes": "Test completion"
}
```

**Your server should receive:**

```json
POST https://your-system.com/webhooks/amss
Headers:
  X-AMSS-Signature: sha256=abc123...
  X-AMSS-Event: task.completed
  Content-Type: application/json

Body:
{
  "event": "task.completed",
  "timestamp": "2025-01-24T11:05:00Z",
  "data": {
    "id": "task-uuid",
    "aircraft_id": "aircraft-uuid",
    "program_id": "program-uuid",
    "status": "completed",
    "completed_at": "2025-01-24T11:05:00Z",
    "completed_by": {
      "id": "user-uuid",
      "email": "technician@company.com"
    },
    "completed_at_hours": 1605.2,
    "completion_notes": "Test completion"
  }
}
```

---

### Webhook Retry Behavior

**Retry Schedule** (if your server returns non-200 status):

| Attempt | Delay | Total Time |
|---------|-------|------------|
| 1 | Immediate | 0s |
| 2 | 1 minute | 1m |
| 3 | 5 minutes | 6m |
| 4 | 15 minutes | 21m |
| 5 | 1 hour | 1h 21m |
| 6 | 3 hours | 4h 21m |
| 7-10 | 6 hours each | Up to 28h |

**After 10 failures**: Webhook marked as failed, admin alerted

**To prevent retries**: Return `200 OK` even if you encounter errors processing the event (log errors internally)

---

### Webhook Security Best Practices

1. **Always Verify Signatures**
   ```javascript
   // Don't trust payload without signature verification
   if (!verifySignature(req.headers['x-amss-signature'], req.body)) {
     return res.status(401).send('Unauthorized');
   }
   ```

2. **Use HTTPS Only**
   - AMSS only sends webhooks to `https://` URLs
   - Self-signed certificates not supported

3. **Implement Idempotency**
   ```javascript
   // Track processed event IDs to handle duplicates
   const eventId = req.body.id;
   if (await isAlreadyProcessed(eventId)) {
     return res.status(200).send('Already processed');
   }
   await markAsProcessed(eventId);
   ```

4. **Timeout Quickly**
   - Acknowledge receipt within 5 seconds
   - Process asynchronously in background
   ```javascript
   app.post('/webhooks/amss', (req, res) => {
     // Queue for background processing
     queue.add('process-webhook', req.body);

     // Acknowledge immediately
     res.status(200).send('OK');
   });
   ```

---

### Monitoring Webhook Deliveries

```bash
# Get webhook delivery history
GET /api/v1/webhooks/{webhook_id}/deliveries?limit=50
Authorization: Bearer <token>
```

**Response:**
```json
{
  "deliveries": [
    {
      "id": "delivery-uuid",
      "event_type": "task.completed",
      "status": "success",
      "attempts": 1,
      "response_status": 200,
      "created_at": "2025-01-24T11:05:00Z",
      "last_attempt_at": "2025-01-24T11:05:01Z"
    },
    {
      "id": "delivery-uuid-2",
      "event_type": "part.consumed",
      "status": "failed",
      "attempts": 5,
      "response_status": 500,
      "response_body": "Internal Server Error",
      "created_at": "2025-01-24T10:00:00Z",
      "last_attempt_at": "2025-01-24T11:00:00Z"
    }
  ]
}
```

---

### Webhook Debugging

#### Test Webhook Locally (Development)

Use **ngrok** to expose local server:

```bash
# Start ngrok
ngrok http 3000

# Use ngrok URL for webhook
POST /api/v1/webhooks
{
  "url": "https://abc123.ngrok.io/webhooks/amss",
  "events": ["task.completed"]
}
```

#### View Webhook Logs

```bash
# Check delivery attempts
GET /api/v1/webhooks/{webhook_id}/deliveries?status=failed

# Retry failed delivery manually
POST /api/v1/webhooks/deliveries/{delivery_id}/retry
```

---

## 6. Import/Export Guide

### CSV Import

AMSS supports bulk import of:
- Aircraft
- Parts
- Maintenance Programs
- Users (admin only)

#### Import Aircraft via CSV

**Step 1: Prepare CSV File**

**File:** `aircraft_import.csv`
```csv
registration,serial_number,manufacturer,model,current_hours,current_cycles
N12345,SN-001,Cessna,172S,1500.5,800
N67890,SN-002,Piper,PA-28,2300.0,1200
N24680,SN-003,Beechcraft,A36,850.0,450
```

**CSV Format Requirements:**
- UTF-8 encoding
- Header row required
- Comma-separated
- Fields with commas must be quoted: `"Cessna, LLC"`

**Step 2: Upload CSV**

```bash
POST /api/v1/imports/aircraft
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: aircraft_import.csv
```

**Response:**
```json
{
  "import_id": "import-uuid",
  "status": "pending",
  "filename": "aircraft_import.csv",
  "total_rows": 3,
  "created_at": "2025-01-24T12:00:00Z"
}
```

**Step 3: Check Import Status**

```bash
GET /api/v1/imports/{import_id}
Authorization: Bearer <token>
```

**Response (In Progress):**
```json
{
  "id": "import-uuid",
  "status": "processing",
  "filename": "aircraft_import.csv",
  "total_rows": 3,
  "processed_rows": 2,
  "successful_rows": 2,
  "failed_rows": 0,
  "progress_percent": 66.7
}
```

**Response (Completed):**
```json
{
  "id": "import-uuid",
  "status": "completed",
  "filename": "aircraft_import.csv",
  "total_rows": 3,
  "processed_rows": 3,
  "successful_rows": 3,
  "failed_rows": 0,
  "progress_percent": 100,
  "results": {
    "created": 3,
    "updated": 0,
    "errors": []
  },
  "completed_at": "2025-01-24T12:01:30Z"
}
```

**Response (With Errors):**
```json
{
  "status": "completed",
  "total_rows": 3,
  "successful_rows": 2,
  "failed_rows": 1,
  "results": {
    "created": 2,
    "updated": 0,
    "errors": [
      {
        "row": 2,
        "registration": "N67890",
        "error": "duplicate registration",
        "message": "Aircraft with registration N67890 already exists"
      }
    ]
  }
}
```

---

#### Import Parts via CSV

**File:** `parts_import.csv`
```csv
part_number,description,manufacturer,unit_price,min_stock_level,category
OIL-FILTER-CH48110,Champion Oil Filter CH48110,Champion Aerospace,12.50,5,filters
SPARK-PLUG-REM37BY,Spark Plug REM37BY,Champion Aerospace,8.75,12,ignition
BRAKE-PAD-30-5,Brake Pad Set,McCauley,45.00,4,brakes
```

**Upload:**
```bash
POST /api/v1/imports/parts
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: parts_import.csv
```

---

### CSV Export

#### Export Aircraft List

```bash
GET /api/v1/exports/aircraft?format=csv
Authorization: Bearer <token>
```

**Response:**
```
Content-Type: text/csv
Content-Disposition: attachment; filename="aircraft_export_2025-01-24.csv"

registration,serial_number,manufacturer,model,current_hours,current_cycles,created_at
N12345,SN-001,Cessna,172S,1500.5,800,2025-01-20T10:00:00Z
N67890,SN-002,Piper,PA-28,2300.0,1200,2025-01-21T14:30:00Z
```

#### Export Tasks (with filters)

```bash
GET /api/v1/exports/tasks?status=completed&created_after=2025-01-01&format=csv
Authorization: Bearer <token>
```

#### Export Inventory Status

```bash
GET /api/v1/exports/inventory?low_stock=true&format=csv
Authorization: Bearer <token>
```

---

## 7. Troubleshooting & FAQ

### Common Errors

#### "401 Unauthorized"

**Possible Causes:**
1. Missing Authorization header
2. Expired access token
3. Invalid token signature

**Solutions:**
```bash
# Check if token is expired
curl https://jwt.io/ # Paste token, check 'exp' claim

# Get fresh token
POST /api/v1/auth/refresh
{
  "refresh_token": "<your_refresh_token>"
}

# Verify Authorization header format
Authorization: Bearer <token>  # ✅ Correct
Authorization: <token>          # ❌ Missing "Bearer"
```

---

#### "403 Forbidden"

**Cause:** User lacks required permissions

**Example:**
```json
{
  "error": "forbidden",
  "message": "User role 'viewer' cannot perform 'tasks:write'"
}
```

**Solution:**
- Contact admin to update user role
- Verify required permission in API docs

**Role Permissions:**
- `tasks:write` → planner, admin
- `tasks:execute` → technician, admin
- `users:write` → admin only

---

#### "409 Conflict - Idempotency key already used"

**Cause:** Retrying request with same idempotency key

**Response:**
```json
{
  "error": "idempotency_conflict",
  "message": "Request with this idempotency key already processed",
  "original_response": {
    "id": "task-uuid",
    "status": "completed"
  }
}
```

**Solution:**
- This is expected behavior (prevents duplicates)
- Use the `original_response` data
- Generate new idempotency key only for genuinely new requests

---

#### "429 Too Many Requests"

**Cause:** Rate limit exceeded

**Response:**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Organization has exceeded 100 requests per minute",
  "retry_after": 30
}
```

**Solution:**
```javascript
// Implement exponential backoff
async function makeRequest(url, options, retries = 3) {
  try {
    const response = await fetch(url, options);

    if (response.status === 429) {
      const retryAfter = response.headers.get('Retry-After') || 60;
      await sleep(retryAfter * 1000);
      return makeRequest(url, options, retries - 1);
    }

    return response;
  } catch (error) {
    if (retries > 0) {
      await sleep(2000);
      return makeRequest(url, options, retries - 1);
    }
    throw error;
  }
}
```

---

#### "500 Internal Server Error"

**Cause:** Unexpected server error

**What to do:**
1. **Retry**: May be transient issue
2. **Check status page**: [https://status.amss.example.com]
3. **Contact support**: Include request ID from response

**Response:**
```json
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred",
  "request_id": "req_abc123",
  "timestamp": "2025-01-24T12:00:00Z"
}
```

**When contacting support, provide:**
- Request ID
- Timestamp
- Request payload (remove sensitive data)
- Expected vs actual behavior

---

### FAQ

**Q: How long are access tokens valid?**
A: 15 minutes. Use refresh token to get new access token without re-login.

**Q: How long are refresh tokens valid?**
A: 7 days. After expiration, user must log in again.

**Q: Can I revoke a token before expiration?**
A: No. JWTs are stateless. Use short TTLs (15min) to limit exposure. Contact support to disable compromised accounts.

**Q: What happens if I update aircraft hours and a task becomes overdue?**
A: System automatically marks task as overdue and sends webhook notification (if configured).

**Q: How do I know which parts are needed for a task?**
A: `GET /api/v1/tasks/{id}` includes `required_parts` array with part numbers and quantities.

**Q: What happens if I try to reserve parts that aren't in stock?**
A: API returns `409 Conflict` with error message indicating insufficient inventory.

**Q: Can I update a completed task?**
A: No. Completed tasks are immutable for audit compliance. Create a new task if work needs to be redone.

**Q: How do I export all my data?**
A: Use export endpoints:
```bash
GET /api/v1/exports/aircraft?format=csv
GET /api/v1/exports/tasks?format=csv
GET /api/v1/exports/inventory?format=csv
```

**Q: Are webhooks guaranteed to be delivered?**
A: At-least-once delivery with up to 10 retry attempts over 28 hours. Implement idempotency to handle duplicates.

**Q: Can I filter tasks by aircraft registration instead of aircraft ID?**
A: Yes:
```bash
GET /api/v1/tasks?aircraft_registration=N12345
```

**Q: How do I generate JWT keys for testing?**
A:
```bash
# Generate RSA key pair
openssl genrsa -out jwt-private.pem 2048
openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem
```

---

## 8. Rate Limiting & Best Practices

### Rate Limits

**Per Organization:**
- 100 requests per minute
- 5,000 requests per hour
- 50,000 requests per day

**Per Endpoint:**
- Auth login: 10 requests per minute (per IP)
- Password reset: 5 requests per hour (per email)

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1706097660
```

---

### Best Practices

#### 1. Use Refresh Tokens

**❌ Don't:** Re-login every 15 minutes
```javascript
// Bad: Creates unnecessary load
setInterval(() => {
  login(email, password);
}, 15 * 60 * 1000);
```

**✅ Do:** Use refresh token
```javascript
// Good: Efficient token management
async function getAccessToken() {
  if (isTokenExpired(accessToken)) {
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken })
    });
    accessToken = response.data.access_token;
  }
  return accessToken;
}
```

---

#### 2. Implement Exponential Backoff

```javascript
async function retryWithBackoff(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;

      const delay = Math.min(1000 * Math.pow(2, i), 30000);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
}

// Usage
const aircraft = await retryWithBackoff(() =>
  fetch('/api/v1/aircraft/123', {
    headers: { Authorization: `Bearer ${token}` }
  })
);
```

---

#### 3. Use Pagination for Large Datasets

**❌ Don't:** Fetch all records
```javascript
// Bad: May timeout or hit memory limits
const allAircraft = await fetch('/api/v1/aircraft?limit=999999');
```

**✅ Do:** Paginate and stream
```javascript
// Good: Process in chunks
async function* fetchAllAircraft() {
  let page = 1;
  let hasMore = true;

  while (hasMore) {
    const response = await fetch(`/api/v1/aircraft?page=${page}&limit=100`);
    const data = await response.json();

    yield* data.aircraft;

    hasMore = page < data.pages;
    page++;
  }
}

// Usage
for await (const aircraft of fetchAllAircraft()) {
  console.log(aircraft.registration);
}
```

---

#### 4. Use Idempotency Keys for Writes

```javascript
import { v4 as uuidv4 } from 'uuid';

async function createTask(aircraftId, programId) {
  const idempotencyKey = uuidv4();

  return fetch('/api/v1/tasks', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Idempotency-Key': idempotencyKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ aircraft_id: aircraftId, program_id: programId })
  });
}
```

---

#### 5. Cache Reference Data

**Cache these infrequently-changing resources:**
- User roles and permissions
- Part definitions
- Maintenance program templates

```javascript
// Cache part definitions for 1 hour
const partsCache = new Map();

async function getPart(partId) {
  const cached = partsCache.get(partId);
  if (cached && Date.now() - cached.timestamp < 3600000) {
    return cached.data;
  }

  const part = await fetch(`/api/v1/parts/${partId}`);
  partsCache.set(partId, { data: part, timestamp: Date.now() });
  return part;
}
```

---

#### 6. Handle Errors Gracefully

```javascript
async function completeTask(taskId, data) {
  try {
    const response = await fetch(`/api/v1/tasks/${taskId}/complete`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    if (!response.ok) {
      const error = await response.json();

      switch (response.status) {
        case 401:
          // Token expired, refresh and retry
          await refreshToken();
          return completeTask(taskId, data);

        case 409:
          if (error.code === 'INSUFFICIENT_INVENTORY') {
            // Show user-friendly message
            alert(`Cannot complete: ${error.message}`);
            return;
          }
          break;

        case 429:
          // Rate limited, wait and retry
          const retryAfter = response.headers.get('Retry-After') || 60;
          await sleep(retryAfter * 1000);
          return completeTask(taskId, data);

        case 500:
          // Log error and notify support
          console.error(`Request ID: ${error.request_id}`, error);
          notifySupport(error);
          break;
      }

      throw new Error(error.message);
    }

    return await response.json();

  } catch (error) {
    console.error('Failed to complete task:', error);
    throw error;
  }
}
```

---

## Support & Resources

### Documentation

- **OpenAPI Spec**: [https://amss-api-uat.duckdns.org/openapi.yaml](https://amss-api-uat.duckdns.org/openapi.yaml)
- **Swagger UI**: [https://amss-api-uat.duckdns.org/docs](https://amss-api-uat.duckdns.org/docs)
- **Developer Guide**: [docs/DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md)
- **GitHub Repository**: [github.com/yourorg/amss-backend](https://github.com/yourorg/amss-backend)

### API Status

- **Status Page**: [https://status.amss.example.com](https://status.amss.example.com)
- **Health Check**: [https://amss-api-uat.duckdns.org/health](https://amss-api-uat.duckdns.org/health)

### Support Channels

- **Email**: support@amss.example.com
- **GitHub Issues**: [github.com/yourorg/amss-backend/issues](https://github.com/yourorg/amss-backend/issues)
- **Slack**: #amss-api-support

### API Client Libraries

- **JavaScript/TypeScript**: `npm install @amss/client`
- **Python**: `pip install amss-client`
- **Go**: `go get github.com/yourorg/amss-go-client`

---

**Happy Building!** 🛩️
