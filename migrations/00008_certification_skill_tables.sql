-- +goose Up

-- Certification authority enum
CREATE TYPE certification_authority AS ENUM ('faa', 'easa', 'ecaa', 'icao', 'other');

-- Certification types recognized by the system
CREATE TABLE certification_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  authority certification_authority NOT NULL,
  has_expiry boolean NOT NULL DEFAULT false,
  recency_required_months int,
  recency_period_months int,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Aircraft type definitions (for type ratings)
CREATE TABLE aircraft_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  icao_code text NOT NULL UNIQUE,
  manufacturer text NOT NULL,
  model text NOT NULL,
  series text,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Mechanic certifications
CREATE TABLE employee_certifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  cert_type_id uuid NOT NULL REFERENCES certification_types(id),
  certificate_number text NOT NULL,
  issue_date date NOT NULL,
  expiry_date date,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'expired', 'suspended', 'revoked')),
  verified_by uuid,
  verified_at timestamptz,
  document_url text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  UNIQUE (org_id, user_id, cert_type_id)
);

-- Employee type ratings (per certification)
CREATE TABLE employee_type_ratings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  aircraft_type_id uuid NOT NULL REFERENCES aircraft_types(id),
  cert_id uuid NOT NULL REFERENCES employee_certifications(id),
  endorsement_date date NOT NULL,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'lapsed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id)
);

-- Special skill categories
CREATE TYPE skill_category AS ENUM (
  'structural', 'avionics', 'engine', 'ndt', 'general'
);

-- Skill type definitions
CREATE TABLE skill_types (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  category skill_category NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Employee skills
CREATE TABLE employee_skills (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  skill_type_id uuid NOT NULL REFERENCES skill_types(id),
  proficiency_level int NOT NULL DEFAULT 1 CHECK (proficiency_level BETWEEN 1 AND 5),
  qualification_date date NOT NULL,
  expiry_date date,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  UNIQUE (org_id, user_id, skill_type_id)
);

-- Recency tracking (hours logged per type per mechanic)
CREATE TABLE employee_recency_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  aircraft_type_id uuid NOT NULL REFERENCES aircraft_types(id),
  task_id uuid NOT NULL,
  work_date date NOT NULL,
  hours_logged numeric(6,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  FOREIGN KEY (org_id, task_id) REFERENCES maintenance_tasks(org_id, id)
);

-- Task skill requirements
CREATE TABLE task_skill_requirements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  task_type maintenance_task_type NOT NULL,
  aircraft_type_id uuid REFERENCES aircraft_types(id),
  cert_type_id uuid REFERENCES certification_types(id),
  skill_type_id uuid REFERENCES skill_types(id),
  min_proficiency_level int DEFAULT 1,
  is_certifying_role boolean NOT NULL DEFAULT false,
  is_inspection_role boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Link aircraft table to aircraft_types
ALTER TABLE aircraft ADD COLUMN aircraft_type_id uuid REFERENCES aircraft_types(id);

-- Indexes
CREATE INDEX employee_certifications_user_idx ON employee_certifications (org_id, user_id) WHERE status = 'active';
CREATE INDEX employee_certifications_expiry_idx ON employee_certifications (expiry_date) WHERE status = 'active' AND expiry_date IS NOT NULL;
CREATE INDEX employee_type_ratings_user_idx ON employee_type_ratings (org_id, user_id) WHERE status = 'active';
CREATE INDEX employee_skills_user_idx ON employee_skills (org_id, user_id);
CREATE INDEX employee_recency_log_user_type_idx ON employee_recency_log (org_id, user_id, aircraft_type_id, work_date);
CREATE INDEX task_skill_requirements_lookup_idx ON task_skill_requirements (org_id, task_type);

-- +goose Down
DROP INDEX IF EXISTS task_skill_requirements_lookup_idx;
DROP INDEX IF EXISTS employee_recency_log_user_type_idx;
DROP INDEX IF EXISTS employee_skills_user_idx;
DROP INDEX IF EXISTS employee_type_ratings_user_idx;
DROP INDEX IF EXISTS employee_certifications_expiry_idx;
DROP INDEX IF EXISTS employee_certifications_user_idx;

ALTER TABLE aircraft DROP COLUMN IF EXISTS aircraft_type_id;

DROP TABLE IF EXISTS task_skill_requirements;
DROP TABLE IF EXISTS employee_recency_log;
DROP TABLE IF EXISTS employee_skills;
DROP TABLE IF EXISTS skill_types;
DROP TYPE IF EXISTS skill_category;
DROP TABLE IF EXISTS employee_type_ratings;
DROP TABLE IF EXISTS employee_certifications;
DROP TABLE IF EXISTS aircraft_types;
DROP TABLE IF EXISTS certification_types;
DROP TYPE IF EXISTS certification_authority;
