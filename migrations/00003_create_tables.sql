-- +goose Up
CREATE TABLE organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  email citext NOT NULL,
  role user_role NOT NULL,
  password_hash text NOT NULL,
  last_login timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id)
);

CREATE TABLE aircraft (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  tail_number text NOT NULL,
  model text NOT NULL,
  last_maintenance timestamptz,
  next_due timestamptz,
  status aircraft_status NOT NULL,
  capacity_slots int NOT NULL CHECK (capacity_slots > 0),
  flight_hours_total int NOT NULL DEFAULT 0,
  cycles_total int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id),
  CHECK (next_due IS NULL OR last_maintenance IS NULL OR next_due > last_maintenance)
);

CREATE TABLE maintenance_programs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  aircraft_id uuid,
  name text NOT NULL,
  interval_type maintenance_program_interval_type NOT NULL,
  interval_value int NOT NULL CHECK (interval_value > 0),
  last_performed timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, aircraft_id) REFERENCES aircraft(org_id, id)
);

CREATE TABLE maintenance_tasks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  aircraft_id uuid NOT NULL,
  program_id uuid,
  type maintenance_task_type NOT NULL,
  state maintenance_task_state NOT NULL,
  start_time timestamptz NOT NULL,
  end_time timestamptz NOT NULL,
  assigned_mechanic_id uuid,
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  active_window tstzrange GENERATED ALWAYS AS (
    CASE
      WHEN state IN ('scheduled', 'in_progress') AND deleted_at IS NULL
      THEN tstzrange(start_time, end_time, '[)')
      ELSE NULL
    END
  ) STORED,
  UNIQUE (org_id, id),
  CHECK (end_time > start_time),
  FOREIGN KEY (org_id, aircraft_id) REFERENCES aircraft(org_id, id),
  FOREIGN KEY (org_id, program_id) REFERENCES maintenance_programs(org_id, id),
  FOREIGN KEY (org_id, assigned_mechanic_id) REFERENCES users(org_id, id)
);

CREATE TABLE part_definitions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  name text NOT NULL,
  category text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id)
);

CREATE TABLE part_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  part_definition_id uuid NOT NULL,
  serial_number text NOT NULL,
  status part_item_status NOT NULL,
  expiry_date timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id),
  CHECK (expiry_date IS NULL OR expiry_date > now()),
  FOREIGN KEY (org_id, part_definition_id) REFERENCES part_definitions(org_id, id)
);

CREATE TABLE part_reservations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  part_item_id uuid NOT NULL,
  state part_reservation_state NOT NULL,
  quantity int NOT NULL DEFAULT 1 CHECK (quantity > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  FOREIGN KEY (org_id, part_item_id) REFERENCES part_items(org_id, id)
);

CREATE TABLE compliance_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  description text NOT NULL,
  result compliance_result NOT NULL,
  sign_off_user_id uuid,
  sign_off_time timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  FOREIGN KEY (org_id, sign_off_user_id) REFERENCES users(org_id, id)
);

CREATE TABLE audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  entity_type text NOT NULL,
  entity_id uuid NOT NULL,
  action audit_action NOT NULL,
  user_id uuid NOT NULL,
  request_id uuid NOT NULL,
  ip_address text,
  user_agent text,
  entity_version int NOT NULL DEFAULT 0,
  timestamp timestamptz NOT NULL DEFAULT now(),
  details jsonb,
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id)
);

CREATE TABLE outbox_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  dedupe_key text NOT NULL,
  attempt_count int NOT NULL DEFAULT 0,
  last_error text,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  locked_at timestamptz,
  locked_by text,
  created_at timestamptz NOT NULL DEFAULT now(),
  processed_at timestamptz,
  UNIQUE (org_id, id),
  UNIQUE (org_id, dedupe_key)
);

CREATE TABLE idempotency_keys (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  key text NOT NULL,
  endpoint text NOT NULL,
  request_hash text NOT NULL,
  response_body jsonb,
  status_code int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  UNIQUE (org_id, id),
  UNIQUE (org_id, key, endpoint)
);

CREATE TABLE webhooks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  url text NOT NULL,
  events text[] NOT NULL,
  secret text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, id)
);

CREATE TABLE webhook_deliveries (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  webhook_id uuid NOT NULL,
  event_id uuid NOT NULL,
  attempt_count int NOT NULL DEFAULT 0,
  last_error text,
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  status webhook_delivery_status NOT NULL,
  last_response_code int,
  last_response_body text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, webhook_id) REFERENCES webhooks(org_id, id),
  FOREIGN KEY (org_id, event_id) REFERENCES outbox_events(org_id, id)
);

CREATE TABLE imports (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  type import_type NOT NULL,
  status import_status NOT NULL,
  file_name text NOT NULL,
  created_by uuid NOT NULL,
  summary jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, created_by) REFERENCES users(org_id, id)
);

CREATE TABLE import_rows (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  import_id uuid NOT NULL,
  row_number int NOT NULL,
  raw jsonb NOT NULL,
  status import_row_status NOT NULL,
  errors jsonb,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, id),
  FOREIGN KEY (org_id, import_id) REFERENCES imports(org_id, id)
);

CREATE TABLE org_policies (
  org_id uuid PRIMARY KEY REFERENCES organizations(id),
  retention_interval interval NOT NULL,
  max_webhook_attempts int NOT NULL DEFAULT 10,
  webhook_replay_window_seconds int NOT NULL DEFAULT 300,
  api_rate_limit_per_min int NOT NULL DEFAULT 100,
  api_key_rate_limit_per_min int NOT NULL DEFAULT 10,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS org_policies;
DROP TABLE IF EXISTS import_rows;
DROP TABLE IF EXISTS imports;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS compliance_items;
DROP TABLE IF EXISTS part_reservations;
DROP TABLE IF EXISTS part_items;
DROP TABLE IF EXISTS part_definitions;
DROP TABLE IF EXISTS maintenance_tasks;
DROP TABLE IF EXISTS maintenance_programs;
DROP TABLE IF EXISTS aircraft;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;
