-- +goose Up

CREATE TYPE alert_level AS ENUM ('info', 'warning', 'critical');

CREATE TABLE alerts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  level alert_level NOT NULL,
  category text NOT NULL,
  title text NOT NULL,
  description text,
  entity_type text NOT NULL,
  entity_id uuid NOT NULL,
  threshold_value numeric,
  current_value numeric,
  acknowledged boolean NOT NULL DEFAULT false,
  acknowledged_by uuid,
  acknowledged_at timestamptz,
  resolved boolean NOT NULL DEFAULT false,
  resolved_at timestamptz,
  escalation_level int NOT NULL DEFAULT 0,
  auto_escalate_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, acknowledged_by) REFERENCES users(org_id, id)
);

-- Indexes
CREATE INDEX alerts_org_unresolved_idx ON alerts (org_id, created_at DESC) WHERE resolved = false;
CREATE INDEX alerts_org_level_idx ON alerts (org_id, level) WHERE resolved = false;
CREATE INDEX alerts_escalation_idx ON alerts (auto_escalate_at) WHERE resolved = false AND auto_escalate_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS alerts_escalation_idx;
DROP INDEX IF EXISTS alerts_org_level_idx;
DROP INDEX IF EXISTS alerts_org_unresolved_idx;
DROP TABLE IF EXISTS alerts;
DROP TYPE IF EXISTS alert_level;
