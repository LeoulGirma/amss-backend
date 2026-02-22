-- +goose Up

-- Task priority levels
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE task_priority AS ENUM ('routine', 'urgent', 'aog', 'critical');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd

-- Add priority and rescheduling fields to maintenance_tasks
-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks ADD COLUMN priority task_priority NOT NULL DEFAULT 'routine';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks ADD COLUMN reschedule_count int NOT NULL DEFAULT 0;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks ADD COLUMN reschedule_reason text;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks ADD COLUMN original_start_time timestamptz;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks ADD COLUMN original_end_time timestamptz;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- Task dependencies (prerequisite relationships)
CREATE TABLE IF NOT EXISTS task_dependencies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  depends_on_task_id uuid NOT NULL,
  dependency_type text NOT NULL DEFAULT 'finish_to_start'
    CHECK (dependency_type IN ('finish_to_start', 'start_to_start', 'finish_to_finish')),
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  FOREIGN KEY (org_id, depends_on_task_id) REFERENCES maintenance_tasks(org_id, id),
  UNIQUE (org_id, task_id, depends_on_task_id),
  CHECK (task_id != depends_on_task_id)
);

-- Part availability tracking (stock level thresholds)
-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE part_definitions ADD COLUMN min_stock_level int NOT NULL DEFAULT 0;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE part_definitions ADD COLUMN reorder_point int NOT NULL DEFAULT 0;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE part_definitions ADD COLUMN lead_time_days int;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- Schedule change events (for notification and audit)
CREATE TABLE IF NOT EXISTS schedule_change_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_id uuid NOT NULL,
  change_type text NOT NULL
    CHECK (change_type IN ('rescheduled', 'cancelled', 'priority_changed', 'mechanic_reassigned')),
  reason text NOT NULL,
  old_start_time timestamptz,
  new_start_time timestamptz,
  old_end_time timestamptz,
  new_end_time timestamptz,
  triggered_by uuid,
  affected_task_ids uuid[],
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS task_dependencies_task_idx ON task_dependencies (org_id, task_id);
CREATE INDEX IF NOT EXISTS task_dependencies_depends_on_idx ON task_dependencies (org_id, depends_on_task_id);
CREATE INDEX IF NOT EXISTS maintenance_tasks_priority_idx ON maintenance_tasks (org_id, priority) WHERE deleted_at IS NULL AND state IN ('scheduled', 'in_progress');
CREATE INDEX IF NOT EXISTS schedule_change_events_task_idx ON schedule_change_events (org_id, task_id);

-- +goose Down
DROP INDEX IF EXISTS schedule_change_events_task_idx;
DROP INDEX IF EXISTS maintenance_tasks_priority_idx;
DROP INDEX IF EXISTS task_dependencies_depends_on_idx;
DROP INDEX IF EXISTS task_dependencies_task_idx;

DROP TABLE IF EXISTS schedule_change_events;

ALTER TABLE part_definitions DROP COLUMN IF EXISTS lead_time_days;
ALTER TABLE part_definitions DROP COLUMN IF EXISTS reorder_point;
ALTER TABLE part_definitions DROP COLUMN IF EXISTS min_stock_level;

DROP TABLE IF EXISTS task_dependencies;

ALTER TABLE maintenance_tasks DROP COLUMN IF EXISTS original_end_time;
ALTER TABLE maintenance_tasks DROP COLUMN IF EXISTS original_start_time;
ALTER TABLE maintenance_tasks DROP COLUMN IF EXISTS reschedule_reason;
ALTER TABLE maintenance_tasks DROP COLUMN IF EXISTS reschedule_count;
ALTER TABLE maintenance_tasks DROP COLUMN IF EXISTS priority;

DROP TYPE IF EXISTS task_priority;
