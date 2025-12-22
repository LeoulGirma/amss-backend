# Architecture Diagrams

This file contains diagram source for the current architecture plus planned items (marked).

## Mermaid (Architecture)

```mermaid
flowchart LR
  %% Actors
  user[Web Clients / Admins / Automation] -->|HTTP REST| server
  user -->|gRPC| server
  ext[Webhook Consumers] <-->|HTTP POST| worker

  %% Server
  subgraph S[amss-server]
    server[REST + gRPC API]
    mw[Middleware: Auth - Idempotency - RateLimit - RequestID - Logging]
    handlers[Handlers]
    services[App Services]
    server --> mw --> handlers --> services
  end

  %% Worker
  subgraph W[amss-worker]
    worker[Jobs]
    outbox[Outbox Publisher]
    webhook[Webhook Dispatcher]
    imports[Import Processor]
    program[Program Generator]
    retention[Retention Cleaner (planned)]
    worker --> outbox
    worker --> webhook
    worker --> imports
    worker --> program
    worker -.-> retention
  end

  %% Data Stores
  db[(Postgres)]
  redis[(Redis)]
  otlp[(OTLP Collector)]
  prom[(Prometheus Scraper)]

  %% Server connections
  services --> db
  server --> redis

  %% Worker connections
  outbox --> db
  outbox --> redis
  webhook --> db
  webhook --> ext
  imports --> db
  imports --> redis
  program --> db
  retention -.-> db

  %% Observability
  server -.-> otlp
  worker -.-> otlp
  prom -->|/metrics| server

  %% Planned: HTTP/gRPC/DB OTel instrumentation + CORS + org-policy rate limit
  server -.-> planned[Planned: OTel HTTP/gRPC/DB - CORS - Org-policy rate limits]
```

## PlantUML (Architecture)

```plantuml
@startuml
skinparam componentStyle rectangle
skinparam wrapWidth 200
skinparam maxMessageSize 200

actor "Web Clients / Admins / Automation" as User
actor "Webhook Consumers" as WebhookConsumer

rectangle "amss-server" {
  component "REST + gRPC API" as Api
  component "Middleware:\nAuth / Idempotency / RateLimit / RequestID / Logging" as Mw
  component "Handlers" as Handlers
  component "App Services" as Services
  Api --> Mw --> Handlers --> Services
}

rectangle "amss-worker" {
  component "Outbox Publisher" as Outbox
  component "Webhook Dispatcher" as Dispatcher
  component "Import Processor" as Importer
  component "Program Generator" as ProgramGen
  component "Retention Cleaner (planned)" as Retention
}

database "Postgres" as PG
queue "Redis" as Redis
cloud "OTLP Collector" as OTLP
cloud "Prometheus" as Prom

User --> Api : HTTP REST / gRPC
Services --> PG
Api --> Redis

Outbox --> PG
Outbox --> Redis
Dispatcher --> PG
Dispatcher --> WebhookConsumer
Importer --> PG
Importer --> Redis
ProgramGen --> PG
Retention ..> PG

Api ..> OTLP : traces (planned HTTP/gRPC/DB)
"amss-worker" ..> OTLP : traces
Prom --> Api : /metrics

note right of Api
Planned: CORS middleware,
Org-policy rate limits,
OTel HTTP/gRPC/DB
end note
@enduml
```

## Mermaid (Data Model)

```mermaid
erDiagram
  ORGANIZATIONS ||--o{ USERS : has
  ORGANIZATIONS ||--o{ AIRCRAFT : owns
  ORGANIZATIONS ||--o{ MAINTENANCE_PROGRAMS : defines
  AIRCRAFT ||--o{ MAINTENANCE_TASKS : schedules
  MAINTENANCE_PROGRAMS ||--o{ MAINTENANCE_TASKS : generates
  USERS ||--o{ MAINTENANCE_TASKS : assigned
  ORGANIZATIONS ||--o{ PART_DEFINITIONS : catalogs
  PART_DEFINITIONS ||--o{ PART_ITEMS : contains
  MAINTENANCE_TASKS ||--o{ PART_RESERVATIONS : uses
  PART_ITEMS ||--o{ PART_RESERVATIONS : reserved
  MAINTENANCE_TASKS ||--o{ COMPLIANCE_ITEMS : requires
  USERS ||--o{ COMPLIANCE_ITEMS : signs
  ORGANIZATIONS ||--o{ AUDIT_LOGS : records
  USERS ||--o{ AUDIT_LOGS : actor
  ORGANIZATIONS ||--o{ OUTBOX_EVENTS : emits
  ORGANIZATIONS ||--o{ WEBHOOKS : owns
  WEBHOOKS ||--o{ WEBHOOK_DELIVERIES : deliveries
  OUTBOX_EVENTS ||--o{ WEBHOOK_DELIVERIES : triggers
  ORGANIZATIONS ||--o{ IMPORTS : submits
  USERS ||--o{ IMPORTS : created_by
  IMPORTS ||--o{ IMPORT_ROWS : rows
  ORGANIZATIONS ||--|| ORG_POLICIES : policy
  ORGANIZATIONS ||--o{ IDEMPOTENCY_KEYS : idempotency

  ORGANIZATIONS {
    uuid id
    text name
    timestamptz deleted_at
  }
  USERS {
    uuid id
    uuid org_id
    citext email
    user_role role
    timestamptz deleted_at
  }
  AIRCRAFT {
    uuid id
    uuid org_id
    text tail_number
    aircraft_status status
    timestamptz deleted_at
  }
  MAINTENANCE_PROGRAMS {
    uuid id
    uuid org_id
    uuid aircraft_id
    maintenance_program_interval_type interval_type
    int interval_value
    timestamptz deleted_at
  }
  MAINTENANCE_TASKS {
    uuid id
    uuid org_id
    uuid aircraft_id
    uuid program_id
    maintenance_task_state state
    timestamptz start_time
    timestamptz end_time
    uuid assigned_mechanic_id
    timestamptz deleted_at
  }
  PART_DEFINITIONS {
    uuid id
    uuid org_id
    text name
    text category
    timestamptz deleted_at
  }
  PART_ITEMS {
    uuid id
    uuid org_id
    uuid part_definition_id
    text serial_number
    part_item_status status
    timestamptz deleted_at
  }
  PART_RESERVATIONS {
    uuid id
    uuid org_id
    uuid task_id
    uuid part_item_id
    part_reservation_state state
  }
  COMPLIANCE_ITEMS {
    uuid id
    uuid org_id
    uuid task_id
    compliance_result result
    uuid sign_off_user_id
    timestamptz deleted_at
  }
  AUDIT_LOGS {
    uuid id
    uuid org_id
    text entity_type
    audit_action action
    uuid user_id
    timestamptz timestamp
  }
  OUTBOX_EVENTS {
    uuid id
    uuid org_id
    text event_type
    jsonb payload
    timestamptz processed_at
  }
  WEBHOOKS {
    uuid id
    uuid org_id
    text url
    text[] events
  }
  WEBHOOK_DELIVERIES {
    uuid id
    uuid org_id
    uuid webhook_id
    uuid event_id
    webhook_delivery_status status
    timestamptz next_attempt_at
  }
  IMPORTS {
    uuid id
    uuid org_id
    import_type type
    import_status status
    text file_name
    uuid created_by
  }
  IMPORT_ROWS {
    uuid id
    uuid org_id
    uuid import_id
    int row_number
    import_row_status status
  }
  ORG_POLICIES {
    uuid org_id
    interval retention_interval
    int api_rate_limit_per_min
    int api_key_rate_limit_per_min
  }
  IDEMPOTENCY_KEYS {
    uuid id
    uuid org_id
    text key
    text endpoint
    timestamptz expires_at
  }
```

## PlantUML (Data Model)

```plantuml
@startuml
hide methods
hide stereotypes
skinparam classAttributeIconSize 0

class organizations {
  id : uuid
  name : text
  deleted_at : timestamptz
}
class users {
  id : uuid
  org_id : uuid
  email : citext
  role : user_role
  deleted_at : timestamptz
}
class aircraft {
  id : uuid
  org_id : uuid
  tail_number : text
  status : aircraft_status
  deleted_at : timestamptz
}
class maintenance_programs {
  id : uuid
  org_id : uuid
  aircraft_id : uuid
  interval_type : maintenance_program_interval_type
  interval_value : int
  deleted_at : timestamptz
}
class maintenance_tasks {
  id : uuid
  org_id : uuid
  aircraft_id : uuid
  program_id : uuid
  state : maintenance_task_state
  start_time : timestamptz
  end_time : timestamptz
  assigned_mechanic_id : uuid
  deleted_at : timestamptz
}
class part_definitions {
  id : uuid
  org_id : uuid
  name : text
  category : text
  deleted_at : timestamptz
}
class part_items {
  id : uuid
  org_id : uuid
  part_definition_id : uuid
  serial_number : text
  status : part_item_status
  deleted_at : timestamptz
}
class part_reservations {
  id : uuid
  org_id : uuid
  task_id : uuid
  part_item_id : uuid
  state : part_reservation_state
}
class compliance_items {
  id : uuid
  org_id : uuid
  task_id : uuid
  result : compliance_result
  sign_off_user_id : uuid
  deleted_at : timestamptz
}
class audit_logs {
  id : uuid
  org_id : uuid
  entity_type : text
  action : audit_action
  user_id : uuid
  timestamp : timestamptz
}
class outbox_events {
  id : uuid
  org_id : uuid
  event_type : text
  payload : jsonb
  processed_at : timestamptz
}
class webhooks {
  id : uuid
  org_id : uuid
  url : text
  events : text[]
}
class webhook_deliveries {
  id : uuid
  org_id : uuid
  webhook_id : uuid
  event_id : uuid
  status : webhook_delivery_status
  next_attempt_at : timestamptz
}
class imports {
  id : uuid
  org_id : uuid
  type : import_type
  status : import_status
  file_name : text
  created_by : uuid
}
class import_rows {
  id : uuid
  org_id : uuid
  import_id : uuid
  row_number : int
  status : import_row_status
}
class org_policies {
  org_id : uuid
  retention_interval : interval
  api_rate_limit_per_min : int
  api_key_rate_limit_per_min : int
}
class idempotency_keys {
  id : uuid
  org_id : uuid
  key : text
  endpoint : text
  expires_at : timestamptz
}

organizations "1" -- "*" users
organizations "1" -- "*" aircraft
organizations "1" -- "*" maintenance_programs
aircraft "1" -- "*" maintenance_tasks
maintenance_programs "1" -- "*" maintenance_tasks
users "1" -- "*" maintenance_tasks
organizations "1" -- "*" part_definitions
part_definitions "1" -- "*" part_items
maintenance_tasks "1" -- "*" part_reservations
part_items "1" -- "*" part_reservations
maintenance_tasks "1" -- "*" compliance_items
users "1" -- "*" compliance_items
organizations "1" -- "*" audit_logs
users "1" -- "*" audit_logs
organizations "1" -- "*" outbox_events
organizations "1" -- "*" webhooks
webhooks "1" -- "*" webhook_deliveries
outbox_events "1" -- "*" webhook_deliveries
organizations "1" -- "*" imports
users "1" -- "*" imports
imports "1" -- "*" import_rows
organizations "1" -- "1" org_policies
organizations "1" -- "*" idempotency_keys
@enduml
```

## Mermaid (Error-Path Sequences)

### Auth: invalid or missing token

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant Auth as AuthMiddleware

  Client->>API: GET /protected
  API->>Auth: Verify JWT
  alt missing or invalid token
    Auth-->>API: unauthorized
    API-->>Client: 401 AUTH
  else ok
    Auth-->>API: principal
    API-->>Client: 200 OK
  end
```

### Idempotency conflict

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant Idem as IdempotencyStore(PG)

  Client->>API: POST /resource (Idempotency-Key)
  API->>Idem: Get(org, key, endpoint)
  alt existing key, different hash
    Idem-->>API: conflict
    API-->>Client: 409 IDEMPOTENCY_CONFLICT
  else no conflict
    API-->>Client: continue
  end
```

### Rate limit exceeded (org or API key)

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant RL as RateLimiter(Redis)

  Client->>API: POST /resource
  API->>RL: Allow(scope)
  alt limit exceeded
    RL-->>API: deny + reset
    API-->>Client: 429 RATE_LIMITED (Retry-After)
  else allowed
    RL-->>API: ok
    API-->>Client: 2xx
  end
```

### Webhook delivery failure -> retry/fail

```mermaid
sequenceDiagram
  autonumber
  participant Dispatch as WebhookDispatcher
  participant WRepo as WebhookRepo(PG)
  participant DRepo as DeliveryRepo(PG)
  participant OutboxR as OutboxRepo(PG)
  participant Webhook as Webhook Consumer

  Dispatch->>DRepo: ClaimPending
  Dispatch->>WRepo: Get webhook
  Dispatch->>OutboxR: Get event
  Dispatch->>Webhook: POST payload
  alt network error or non-2xx
    Dispatch->>DRepo: Schedule retry or Failed
  else success
    Dispatch->>DRepo: Mark Delivered
  end
```

## PlantUML (Error-Path Sequences)

### Auth: invalid or missing token

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "AuthMiddleware" as Auth

Client -> API: GET /protected
API -> Auth: Verify JWT
alt missing or invalid token
  Auth --> API: unauthorized
  API --> Client: 401 AUTH
else ok
  Auth --> API: principal
  API --> Client: 200 OK
end
@enduml
```

### Idempotency conflict

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "IdempotencyStore (PG)" as Idem

Client -> API: POST /resource (Idempotency-Key)
API -> Idem: Get(org, key, endpoint)
alt existing key, different hash
  Idem --> API: conflict
  API --> Client: 409 IDEMPOTENCY_CONFLICT
else no conflict
  API --> Client: continue
end
@enduml
```

### Rate limit exceeded (org or API key)

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "RateLimiter (Redis)" as RL

Client -> API: POST /resource
API -> RL: Allow(scope)
alt limit exceeded
  RL --> API: deny + reset
  API --> Client: 429 RATE_LIMITED (Retry-After)
else allowed
  RL --> API: ok
  API --> Client: 2xx
end
@enduml
```

### Webhook delivery failure -> retry/fail

```plantuml
@startuml
participant "WebhookDispatcher" as Dispatch
participant "WebhookRepo (PG)" as WRepo
participant "DeliveryRepo (PG)" as DRepo
participant "OutboxRepo (PG)" as OutboxR
participant "Webhook Consumer" as Webhook

Dispatch -> DRepo: ClaimPending
Dispatch -> WRepo: Get webhook
Dispatch -> OutboxR: Get event
Dispatch -> Webhook: POST payload
alt network error or non-2xx
  Dispatch -> DRepo: Schedule retry or Failed
else success
  Dispatch -> DRepo: Mark Delivered
end
@enduml
```

## Mermaid (Sequence Diagrams)

### Auth: login, refresh, logout

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant RL as RateLimiter(Redis)
  participant AuthH as AuthHandler
  participant AuthS as AuthService
  participant AuthR as AuthRepo(PG)
  participant RefreshR as RefreshTokenRepo(PG)
  participant JWT as JWT Signer

  Client->>API: POST /auth/login {email, password}
  API->>RL: Allow(login:ip, login:email)
  alt rate limited
    RL-->>API: deny
    API-->>Client: 429 RATE_LIMITED
  else allowed
    API->>AuthH: Login
    AuthH->>AuthS: Authenticate(email, password)
    AuthS->>AuthR: GetByEmail
    AuthR-->>AuthS: user + password_hash
    AuthS->>AuthS: verify password
    alt invalid credentials
      AuthS-->>AuthH: ErrUnauthorized
      AuthH-->>Client: 401 AUTH_INVALID
    else ok
      AuthS->>RefreshR: Create refresh token
      AuthS->>JWT: Sign access + refresh
      AuthH-->>Client: 200 {tokens}
    end
  end

  Client->>API: POST /auth/refresh {refresh_token}
  API->>AuthH: Refresh
  AuthH->>AuthS: RefreshTokens
  AuthS->>RefreshR: Validate token
  AuthS->>AuthR: GetByID
  AuthS->>JWT: Sign new access + refresh
  AuthH-->>Client: 200 {tokens}

  Client->>API: POST /auth/logout {refresh_token}
  API->>AuthH: Logout
  AuthH->>RefreshR: Revoke token
  AuthH-->>Client: 204 No Content
```

### Create task -> audit -> outbox -> webhook delivery

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant MW as Auth/Idempotency/RateLimit
  participant TaskH as TaskHandler
  participant TaskS as TaskService
  participant TaskR as TaskRepo(PG)
  participant AuditR as AuditRepo(PG)
  participant OutboxR as OutboxRepo(PG)
  participant Worker as OutboxPublisher
  participant Redis as Redis Stream
  participant WRepo as WebhookRepo(PG)
  participant DRepo as WebhookDeliveryRepo(PG)
  participant Dispatch as WebhookDispatcher
  participant Webhook as Webhook Consumer

  Client->>API: POST /maintenance-tasks
  API->>MW: Auth + Idempotency + RateLimit
  MW-->>API: ok
  API->>TaskH: CreateTask
  TaskH->>TaskS: Create
  TaskS->>TaskR: Insert task
  TaskS->>AuditR: Insert audit log
  TaskS->>OutboxR: Enqueue event
  TaskH-->>Client: 201 Created

  Worker->>OutboxR: LockPending
  OutboxR-->>Worker: events
  Worker->>Redis: XADD amss.events
  Worker->>WRepo: ListByEvent
  Worker->>DRepo: Create deliveries
  Worker->>OutboxR: MarkProcessed

  Dispatch->>DRepo: ClaimPending
  DRepo-->>Dispatch: deliveries
  Dispatch->>WRepo: Get webhook config
  Dispatch->>OutboxR: Get event payload
  Dispatch->>Webhook: POST signed payload
  alt success
    Dispatch->>DRepo: Mark Delivered
  else retry/fail
    Dispatch->>DRepo: Schedule retry or Failed
  end
```

### Import CSV -> validate -> apply -> summary

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant ImportH as ImportHandler
  participant ImportS as ImportService
  participant ImportR as ImportRepo(PG)
  participant RowR as ImportRowRepo(PG)
  participant Redis as Redis Stream
  participant Worker as ImportProcessor
  participant AircraftR as AircraftRepo(PG)
  participant PartR as PartsRepos(PG)
  participant ProgramR as ProgramRepo(PG)

  Client->>API: POST /imports/csv (file)
  API->>ImportH: CreateImport
  ImportH->>ImportS: Create + enqueue job
  ImportS->>ImportR: Insert import (pending)
  ImportS->>Redis: XADD import.jobs {import_id}
  ImportH-->>Client: 202 Accepted

  Worker->>Redis: XREADGROUP import.jobs
  Worker->>ImportR: GetByID
  Worker->>ImportR: UpdateStatus(validating)
  Worker->>RowR: Insert rows (valid/invalid)
  Worker->>ImportR: UpdateStatus(applying)
  Worker->>AircraftR: Apply aircraft rows
  Worker->>PartR: Apply part defs/items rows
  Worker->>ProgramR: Apply program rows
  Worker->>RowR: Update row status
  Worker->>ImportR: UpdateStatus(completed/failed + summary)
```

### Program task generation job

```mermaid
sequenceDiagram
  autonumber
  participant Worker as ProgramGenerator
  participant ProgramS as MaintenanceProgramService
  participant ProgramR as ProgramRepo(PG)
  participant TaskR as TaskRepo(PG)
  participant AuditR as AuditRepo(PG)
  participant OutboxR as OutboxRepo(PG)

  Worker->>ProgramS: GenerateDueTasks(actor, batch)
  ProgramS->>ProgramR: List due programs
  loop each due program
    ProgramS->>TaskR: Create maintenance task
    ProgramS->>AuditR: Insert audit log
    ProgramS->>OutboxR: Enqueue task.created
  end
```

### Middleware: idempotency + rate limit

```mermaid
sequenceDiagram
  autonumber
  participant Client
  participant API as REST API
  participant Auth as AuthMiddleware
  participant Idem as IdempotencyStore(PG)
  participant RL as RateLimiter(Redis)
  participant Handler

  Client->>API: POST /resource (Idempotency-Key)
  API->>Auth: Verify JWT
  Auth-->>API: principal
  API->>RL: Allow(org/category)
  alt rate limited
    RL-->>API: deny
    API-->>Client: 429 RATE_LIMITED
  else allowed
    API->>Idem: Get(org, key, endpoint)
    alt cached response exists
      Idem-->>API: stored response
      API-->>Client: replay response
    else no record
      API->>Idem: Create placeholder + request hash
      API->>Handler: Execute handler
      Handler-->>API: response
      API->>Idem: Update response
      API-->>Client: response
    end
  end
```

## PlantUML (Sequence Diagrams)

### Auth: login, refresh, logout

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "RateLimiter (Redis)" as RL
participant "AuthHandler" as AuthH
participant "AuthService" as AuthS
participant "AuthRepo (PG)" as AuthR
participant "RefreshTokenRepo (PG)" as RefreshR
participant "JWT Signer" as JWT

Client -> API: POST /auth/login
API -> RL: Allow(login:ip, login:email)
alt rate limited
  RL --> API: deny
  API --> Client: 429 RATE_LIMITED
else allowed
  API -> AuthH: Login
  AuthH -> AuthS: Authenticate
  AuthS -> AuthR: GetByEmail
  AuthR --> AuthS: user + password_hash
  AuthS -> AuthS: verify password
  alt invalid credentials
    AuthS --> AuthH: ErrUnauthorized
    AuthH --> Client: 401 AUTH_INVALID
  else ok
    AuthS -> RefreshR: Create refresh token
    AuthS -> JWT: Sign access + refresh
    AuthH --> Client: 200 {tokens}
  end
end

Client -> API: POST /auth/refresh
API -> AuthH: Refresh
AuthH -> AuthS: RefreshTokens
AuthS -> RefreshR: Validate token
AuthS -> AuthR: GetByID
AuthS -> JWT: Sign new tokens
AuthH --> Client: 200 {tokens}

Client -> API: POST /auth/logout
API -> AuthH: Logout
AuthH -> RefreshR: Revoke token
AuthH --> Client: 204 No Content
@enduml
```

### Create task -> audit -> outbox -> webhook delivery

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "Middleware" as MW
participant "TaskHandler" as TaskH
participant "TaskService" as TaskS
participant "TaskRepo (PG)" as TaskR
participant "AuditRepo (PG)" as AuditR
participant "OutboxRepo (PG)" as OutboxR
participant "OutboxPublisher" as Worker
queue "Redis Stream" as Redis
participant "WebhookRepo (PG)" as WRepo
participant "WebhookDeliveryRepo (PG)" as DRepo
participant "WebhookDispatcher" as Dispatch
participant "Webhook Consumer" as Webhook

Client -> API: POST /maintenance-tasks
API -> MW: Auth + Idempotency + RateLimit
MW --> API: ok
API -> TaskH: CreateTask
TaskH -> TaskS: Create
TaskS -> TaskR: Insert task
TaskS -> AuditR: Insert audit log
TaskS -> OutboxR: Enqueue event
TaskH --> Client: 201 Created

Worker -> OutboxR: LockPending
OutboxR --> Worker: events
Worker -> Redis: XADD amss.events
Worker -> WRepo: ListByEvent
Worker -> DRepo: Create deliveries
Worker -> OutboxR: MarkProcessed

Dispatch -> DRepo: ClaimPending
DRepo --> Dispatch: deliveries
Dispatch -> WRepo: Get webhook config
Dispatch -> OutboxR: Get event payload
Dispatch -> Webhook: POST signed payload
alt success
  Dispatch -> DRepo: Mark Delivered
else retry/fail
  Dispatch -> DRepo: Schedule retry or Failed
end
@enduml
```

### Import CSV -> validate -> apply -> summary

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "ImportHandler" as ImportH
participant "ImportService" as ImportS
participant "ImportRepo (PG)" as ImportR
participant "ImportRowRepo (PG)" as RowR
queue "Redis Stream" as Redis
participant "ImportProcessor" as Worker
participant "AircraftRepo (PG)" as AircraftR
participant "PartsRepos (PG)" as PartR
participant "ProgramRepo (PG)" as ProgramR

Client -> API: POST /imports/csv (file)
API -> ImportH: CreateImport
ImportH -> ImportS: Create + enqueue job
ImportS -> ImportR: Insert import (pending)
ImportS -> Redis: XADD import.jobs {import_id}
ImportH --> Client: 202 Accepted

Worker -> Redis: XREADGROUP import.jobs
Worker -> ImportR: GetByID
Worker -> ImportR: UpdateStatus(validating)
Worker -> RowR: Insert rows (valid/invalid)
Worker -> ImportR: UpdateStatus(applying)
Worker -> AircraftR: Apply aircraft rows
Worker -> PartR: Apply part defs/items rows
Worker -> ProgramR: Apply program rows
Worker -> RowR: Update row status
Worker -> ImportR: UpdateStatus(completed/failed + summary)
@enduml
```

### Program task generation job

```plantuml
@startuml
participant "ProgramGenerator" as Worker
participant "MaintenanceProgramService" as ProgramS
participant "ProgramRepo (PG)" as ProgramR
participant "TaskRepo (PG)" as TaskR
participant "AuditRepo (PG)" as AuditR
participant "OutboxRepo (PG)" as OutboxR

Worker -> ProgramS: GenerateDueTasks(actor, batch)
ProgramS -> ProgramR: List due programs
loop each due program
  ProgramS -> TaskR: Create maintenance task
  ProgramS -> AuditR: Insert audit log
  ProgramS -> OutboxR: Enqueue task.created
end
@enduml
```

### Middleware: idempotency + rate limit

```plantuml
@startuml
actor Client
participant "REST API" as API
participant "AuthMiddleware" as Auth
participant "IdempotencyStore (PG)" as Idem
participant "RateLimiter (Redis)" as RL
participant "Handler" as Handler

Client -> API: POST /resource (Idempotency-Key)
API -> Auth: Verify JWT
Auth --> API: principal
API -> RL: Allow(org/category)
alt rate limited
  RL --> API: deny
  API --> Client: 429 RATE_LIMITED
else allowed
  API -> Idem: Get(org, key, endpoint)
  alt cached response exists
    Idem --> API: stored response
    API --> Client: replay response
  else no record
    API -> Idem: Create placeholder + request hash
    API -> Handler: Execute handler
    Handler --> API: response
    API -> Idem: Update response
    API --> Client: response
  end
end
@enduml
```
