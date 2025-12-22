-- +goose Up
CREATE TYPE user_role AS ENUM (
  'admin',
  'tenant_admin',
  'scheduler',
  'mechanic',
  'auditor'
);

CREATE TYPE aircraft_status AS ENUM (
  'operational',
  'maintenance',
  'grounded'
);

CREATE TYPE maintenance_task_type AS ENUM (
  'inspection',
  'repair',
  'overhaul'
);

CREATE TYPE maintenance_task_state AS ENUM (
  'scheduled',
  'in_progress',
  'completed',
  'cancelled'
);

CREATE TYPE part_item_status AS ENUM (
  'in_stock',
  'used',
  'disposed'
);

CREATE TYPE part_reservation_state AS ENUM (
  'reserved',
  'used',
  'released'
);

CREATE TYPE maintenance_program_interval_type AS ENUM (
  'flight_hours',
  'cycles',
  'calendar'
);

CREATE TYPE compliance_result AS ENUM (
  'pass',
  'fail',
  'pending'
);

CREATE TYPE audit_action AS ENUM (
  'create',
  'update',
  'delete',
  'state_change'
);

CREATE TYPE import_type AS ENUM (
  'aircraft',
  'parts',
  'programs'
);

CREATE TYPE import_status AS ENUM (
  'pending',
  'validating',
  'applying',
  'completed',
  'failed'
);

CREATE TYPE import_row_status AS ENUM (
  'pending',
  'valid',
  'invalid',
  'applied'
);

CREATE TYPE webhook_delivery_status AS ENUM (
  'pending',
  'delivered',
  'failed'
);

-- +goose Down
DROP TYPE IF EXISTS webhook_delivery_status;
DROP TYPE IF EXISTS import_row_status;
DROP TYPE IF EXISTS import_status;
DROP TYPE IF EXISTS import_type;
DROP TYPE IF EXISTS audit_action;
DROP TYPE IF EXISTS compliance_result;
DROP TYPE IF EXISTS maintenance_program_interval_type;
DROP TYPE IF EXISTS part_reservation_state;
DROP TYPE IF EXISTS part_item_status;
DROP TYPE IF EXISTS maintenance_task_state;
DROP TYPE IF EXISTS maintenance_task_type;
DROP TYPE IF EXISTS aircraft_status;
DROP TYPE IF EXISTS user_role;
