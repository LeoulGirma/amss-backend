-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS users_org_email_uniq ON users (org_id, email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS aircraft_org_tail_uniq ON aircraft (org_id, tail_number) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS part_definitions_org_name_uniq ON part_definitions (org_id, name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS part_items_org_serial_uniq ON part_items (org_id, serial_number) WHERE deleted_at IS NULL;

-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE maintenance_tasks
    ADD CONSTRAINT maintenance_tasks_no_overlap
    EXCLUDE USING gist (org_id WITH =, aircraft_id WITH =, active_window WITH &&);
EXCEPTION WHEN duplicate_table THEN NULL;
         WHEN others THEN NULL;
END $$;
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS part_reservations_active_idx
  ON part_reservations (org_id, part_item_id)
  WHERE state = 'reserved';

CREATE INDEX IF NOT EXISTS outbox_events_unprocessed_idx
  ON outbox_events (org_id, next_attempt_at, created_at)
  WHERE processed_at IS NULL;

CREATE INDEX IF NOT EXISTS outbox_events_processed_idx
  ON outbox_events (org_id, processed_at);

CREATE INDEX IF NOT EXISTS idempotency_keys_expires_idx ON idempotency_keys (expires_at);

CREATE INDEX IF NOT EXISTS webhook_deliveries_pending_idx
  ON webhook_deliveries (org_id, next_attempt_at)
  WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS import_rows_lookup_idx ON import_rows (org_id, import_id, row_number);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION reject_audit_logs_mutation() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'audit_logs are immutable';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS audit_logs_immutable ON audit_logs;
CREATE TRIGGER audit_logs_immutable
  BEFORE UPDATE OR DELETE ON audit_logs
  FOR EACH ROW EXECUTE FUNCTION reject_audit_logs_mutation();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION prevent_compliance_update_after_signoff() RETURNS trigger AS $$
BEGIN
  IF OLD.sign_off_time IS NOT NULL THEN
    RAISE EXCEPTION 'compliance_items are immutable after sign off';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS compliance_items_immutable ON compliance_items;
CREATE TRIGGER compliance_items_immutable
  BEFORE UPDATE ON compliance_items
  FOR EACH ROW EXECUTE FUNCTION prevent_compliance_update_after_signoff();

-- +goose Down
DROP TRIGGER IF EXISTS compliance_items_immutable ON compliance_items;
DROP FUNCTION IF EXISTS prevent_compliance_update_after_signoff();
DROP TRIGGER IF EXISTS audit_logs_immutable ON audit_logs;
DROP FUNCTION IF EXISTS reject_audit_logs_mutation();

DROP INDEX IF EXISTS import_rows_lookup_idx;
DROP INDEX IF EXISTS webhook_deliveries_pending_idx;
DROP INDEX IF EXISTS idempotency_keys_expires_idx;
DROP INDEX IF EXISTS outbox_events_processed_idx;
DROP INDEX IF EXISTS outbox_events_unprocessed_idx;
DROP INDEX IF EXISTS part_reservations_active_idx;

ALTER TABLE maintenance_tasks
  DROP CONSTRAINT IF EXISTS maintenance_tasks_no_overlap;

DROP INDEX IF EXISTS part_items_org_serial_uniq;
DROP INDEX IF EXISTS part_definitions_org_name_uniq;
DROP INDEX IF EXISTS aircraft_org_tail_uniq;
DROP INDEX IF EXISTS users_org_email_uniq;
