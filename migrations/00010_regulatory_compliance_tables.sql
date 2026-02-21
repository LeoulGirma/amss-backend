-- +goose Up

-- Regulatory authorities
CREATE TABLE regulatory_authorities (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  country text NOT NULL,
  record_retention_years int NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Organization regulatory registrations
CREATE TABLE org_regulatory_registrations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  registration_number text NOT NULL,
  scope text,
  effective_date date NOT NULL,
  expiry_date date,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'expired', 'suspended')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, authority_id)
);

-- Compliance directive types
CREATE TYPE directive_type AS ENUM ('ad', 'sb', 'eo', 'tcds', 'stc');
CREATE TYPE directive_applicability AS ENUM ('mandatory', 'recommended', 'optional');

-- Airworthiness Directives & Service Bulletins
CREATE TABLE compliance_directives (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  directive_type directive_type NOT NULL,
  reference_number text NOT NULL,
  title text NOT NULL,
  description text,
  applicability directive_applicability NOT NULL,
  affected_aircraft_types uuid[],
  effective_date date NOT NULL,
  compliance_deadline date,
  recurrence_interval text,
  superseded_by uuid REFERENCES compliance_directives(id),
  source_url text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Aircraft-specific directive compliance status
CREATE TABLE aircraft_directive_compliance (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  aircraft_id uuid NOT NULL,
  directive_id uuid NOT NULL REFERENCES compliance_directives(id),
  status text NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'in_progress', 'compliant', 'not_applicable', 'overdue')),
  compliance_date timestamptz,
  next_due_date timestamptz,
  task_id uuid,
  signed_off_by uuid,
  signed_off_at timestamptz,
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, aircraft_id) REFERENCES aircraft(org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id),
  UNIQUE (org_id, aircraft_id, directive_id)
);

-- Compliance document templates (per authority)
CREATE TABLE compliance_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  authority_id uuid NOT NULL REFERENCES regulatory_authorities(id),
  template_code text NOT NULL,
  name text NOT NULL,
  description text,
  required_fields jsonb NOT NULL,
  template_content text,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (authority_id, template_code)
);

-- Enhance existing compliance_items with authority link
ALTER TABLE compliance_items ADD COLUMN authority_id uuid REFERENCES regulatory_authorities(id);
ALTER TABLE compliance_items ADD COLUMN directive_id uuid REFERENCES compliance_directives(id);
ALTER TABLE compliance_items ADD COLUMN category text;

-- Indexes
CREATE INDEX compliance_directives_authority_idx ON compliance_directives (org_id, authority_id);
CREATE INDEX compliance_directives_type_idx ON compliance_directives (org_id, directive_type);
CREATE INDEX aircraft_directive_compliance_aircraft_idx ON aircraft_directive_compliance (org_id, aircraft_id);
CREATE INDEX aircraft_directive_compliance_status_idx ON aircraft_directive_compliance (org_id, status) WHERE status IN ('pending', 'overdue');
CREATE INDEX org_regulatory_registrations_org_idx ON org_regulatory_registrations (org_id);

-- +goose Down
DROP INDEX IF EXISTS org_regulatory_registrations_org_idx;
DROP INDEX IF EXISTS aircraft_directive_compliance_status_idx;
DROP INDEX IF EXISTS aircraft_directive_compliance_aircraft_idx;
DROP INDEX IF EXISTS compliance_directives_type_idx;
DROP INDEX IF EXISTS compliance_directives_authority_idx;

ALTER TABLE compliance_items DROP COLUMN IF EXISTS category;
ALTER TABLE compliance_items DROP COLUMN IF EXISTS directive_id;
ALTER TABLE compliance_items DROP COLUMN IF EXISTS authority_id;

DROP TABLE IF EXISTS compliance_templates;
DROP TABLE IF EXISTS aircraft_directive_compliance;
DROP TABLE IF EXISTS compliance_directives;
DROP TYPE IF EXISTS directive_applicability;
DROP TYPE IF EXISTS directive_type;
DROP TABLE IF EXISTS org_regulatory_registrations;
DROP TABLE IF EXISTS regulatory_authorities;
