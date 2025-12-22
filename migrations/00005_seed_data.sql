-- +goose Up
WITH org AS (
  INSERT INTO organizations (name)
  VALUES ('Demo Airline')
  RETURNING id
),
users AS (
  INSERT INTO users (org_id, email, role, password_hash)
  SELECT org.id, v.email, v.role::user_role, crypt('ChangeMe123!', gen_salt('bf'))
  FROM org
  CROSS JOIN (VALUES
    ('admin@demo.local', 'admin'),
    ('tenant-admin@demo.local', 'tenant_admin'),
    ('scheduler@demo.local', 'scheduler'),
    ('mechanic@demo.local', 'mechanic'),
    ('auditor@demo.local', 'auditor')
  ) AS v(email, role)
  RETURNING id, org_id, role
),
policy AS (
  INSERT INTO org_policies (
    org_id,
    retention_interval,
    max_webhook_attempts,
    webhook_replay_window_seconds,
    api_rate_limit_per_min,
    api_key_rate_limit_per_min
  )
  SELECT org.id, interval '365 days', 10, 300, 100, 10
  FROM org
  RETURNING org_id
),
aircraft AS (
  INSERT INTO aircraft (
    org_id,
    tail_number,
    model,
    last_maintenance,
    next_due,
    status,
    capacity_slots
  )
  SELECT
    org.id,
    'N100AM',
    'Boeing 737-800',
    now() - interval '30 days',
    now() + interval '60 days',
    'operational',
    1
  FROM org
  RETURNING id, org_id
),
program AS (
  INSERT INTO maintenance_programs (
    org_id,
    aircraft_id,
    name,
    interval_type,
    interval_value,
    last_performed
  )
  SELECT
    org.id,
    aircraft.id,
    'A-check',
    'calendar',
    90,
    now() - interval '90 days'
  FROM org, aircraft
  RETURNING id, org_id
),
task AS (
  INSERT INTO maintenance_tasks (
    org_id,
    aircraft_id,
    program_id,
    type,
    state,
    start_time,
    end_time,
    assigned_mechanic_id,
    notes
  )
  SELECT
    org.id,
    aircraft.id,
    program.id,
    'inspection',
    'scheduled',
    date_trunc('hour', now() + interval '24 hours'),
    date_trunc('hour', now() + interval '30 hours'),
    (SELECT id FROM users WHERE role = 'mechanic' LIMIT 1),
    'Seeded inspection task'
  FROM org, aircraft, program
  RETURNING id, org_id
)
INSERT INTO compliance_items (
  org_id,
  task_id,
  description,
  result
)
SELECT org.id, task.id, 'Safety checklist', 'pending'
FROM org, task;

-- +goose Down
WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM compliance_items WHERE org_id IN (SELECT id FROM org);

WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM maintenance_tasks WHERE org_id IN (SELECT id FROM org);

WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM maintenance_programs WHERE org_id IN (SELECT id FROM org);

WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM aircraft WHERE org_id IN (SELECT id FROM org);

WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM users WHERE org_id IN (SELECT id FROM org);

WITH org AS (
  SELECT id FROM organizations WHERE name = 'Demo Airline'
)
DELETE FROM org_policies WHERE org_id IN (SELECT id FROM org);

DELETE FROM organizations WHERE name = 'Demo Airline';
