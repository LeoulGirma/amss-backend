-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
  v_org_id     uuid;
  v_aircraft_id uuid;
  v_program_id  uuid;
  v_task_id     uuid;
  v_mechanic_id uuid;
BEGIN
  -- Skip entirely if Demo Airline already exists (previously seeded outside goose)
  SELECT id INTO v_org_id FROM organizations WHERE name = 'Demo Airline';
  IF v_org_id IS NOT NULL THEN
    RETURN;
  END IF;

  INSERT INTO organizations (name) VALUES ('Demo Airline') RETURNING id INTO v_org_id;

  INSERT INTO users (org_id, email, role, password_hash)
  SELECT v_org_id, v.email, v.role::user_role, crypt('ChangeMe123!', gen_salt('bf'))
  FROM (VALUES
    ('admin@demo.local', 'admin'),
    ('tenant-admin@demo.local', 'tenant_admin'),
    ('scheduler@demo.local', 'scheduler'),
    ('mechanic@demo.local', 'mechanic'),
    ('auditor@demo.local', 'auditor')
  ) AS v(email, role);

  INSERT INTO org_policies (
    org_id, retention_interval, max_webhook_attempts,
    webhook_replay_window_seconds, api_rate_limit_per_min, api_key_rate_limit_per_min
  ) VALUES (v_org_id, interval '365 days', 10, 300, 100, 10);

  INSERT INTO aircraft (
    org_id, tail_number, model, last_maintenance, next_due, status, capacity_slots
  ) VALUES (
    v_org_id, 'N100AM', 'Boeing 737-800',
    now() - interval '30 days', now() + interval '60 days',
    'operational', 1
  ) RETURNING id INTO v_aircraft_id;

  INSERT INTO maintenance_programs (
    org_id, aircraft_id, name, interval_type, interval_value, last_performed
  ) VALUES (
    v_org_id, v_aircraft_id, 'A-check', 'calendar', 90, now() - interval '90 days'
  ) RETURNING id INTO v_program_id;

  SELECT id INTO v_mechanic_id FROM users WHERE org_id = v_org_id AND role = 'mechanic' LIMIT 1;

  INSERT INTO maintenance_tasks (
    org_id, aircraft_id, program_id, type, state,
    start_time, end_time, assigned_mechanic_id, notes
  ) VALUES (
    v_org_id, v_aircraft_id, v_program_id, 'inspection', 'scheduled',
    date_trunc('hour', now() + interval '24 hours'),
    date_trunc('hour', now() + interval '30 hours'),
    v_mechanic_id, 'Seeded inspection task'
  ) RETURNING id INTO v_task_id;

  INSERT INTO compliance_items (org_id, task_id, description, result)
  VALUES (v_org_id, v_task_id, 'Safety checklist', 'pending');
END $$;
-- +goose StatementEnd

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
