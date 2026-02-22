-- +goose Up

-- Seed regulatory authorities
INSERT INTO regulatory_authorities (code, name, country, record_retention_years) VALUES
  ('FAA',  'Federal Aviation Administration',           'US',  1),
  ('EASA', 'European Union Aviation Safety Agency',     'EU',  3),
  ('ECAA', 'Ethiopian Civil Aviation Authority',        'ET',  5),
  ('ICAO', 'International Civil Aviation Organization', 'INT', 5)
ON CONFLICT (code) DO NOTHING;

-- Seed certification types
INSERT INTO certification_types (code, name, authority, has_expiry, recency_required_months, recency_period_months) VALUES
  -- FAA certifications
  ('FAA_AP',  'Airframe & Powerplant',                 'faa',  false, NULL, NULL),
  ('FAA_A',   'Airframe Mechanic',                     'faa',  false, NULL, NULL),
  ('FAA_P',   'Powerplant Mechanic',                   'faa',  false, NULL, NULL),
  ('FAA_IA',  'Inspection Authorization',              'faa',  true,  NULL, NULL),
  ('FAA_RII', 'Repairman (Inspection)',                'faa',  false, NULL, NULL),
  -- EASA certifications (Part-66)
  ('EASA_A',  'Part-66 Category A (Line Maintenance)', 'easa', true,  6, 24),
  ('EASA_B1', 'Part-66 Category B1 (Mechanical)',      'easa', true,  6, 24),
  ('EASA_B2', 'Part-66 Category B2 (Avionics)',        'easa', true,  6, 24),
  ('EASA_B3', 'Part-66 Category B3 (Piston Engines)',  'easa', true,  6, 24),
  ('EASA_C',  'Part-66 Category C (Base Maintenance)', 'easa', true,  6, 24),
  -- ECAA certifications
  ('ECAA_AMT', 'Aircraft Maintenance Technician',      'ecaa', true,  6, 24),
  ('ECAA_AVI', 'Avionics Technician',                  'ecaa', true,  6, 24)
ON CONFLICT (code) DO NOTHING;

-- Seed common aircraft types
INSERT INTO aircraft_types (icao_code, manufacturer, model, series) VALUES
  ('B738', 'Boeing',  '737-800',   'NG'),
  ('B739', 'Boeing',  '737-900',   'NG'),
  ('B37M', 'Boeing',  '737 MAX 8', 'MAX'),
  ('B38M', 'Boeing',  '737 MAX 8', 'MAX'),
  ('B752', 'Boeing',  '757-200',   NULL),
  ('B763', 'Boeing',  '767-300',   NULL),
  ('B77W', 'Boeing',  '777-300ER', NULL),
  ('B789', 'Boeing',  '787-9',     'Dreamliner'),
  ('B78X', 'Boeing',  '787-10',    'Dreamliner'),
  ('A320', 'Airbus',  'A320',      'CEO'),
  ('A20N', 'Airbus',  'A320neo',   'NEO'),
  ('A359', 'Airbus',  'A350-900',  'XWB'),
  ('A35K', 'Airbus',  'A350-1000', 'XWB'),
  ('DH8D', 'De Havilland', 'Dash 8-400', 'Q400')
ON CONFLICT (icao_code) DO NOTHING;

-- Seed skill types
INSERT INTO skill_types (code, name, category) VALUES
  -- NDT methods
  ('NDT_VT',  'Visual Testing (VT)',              'ndt'),
  ('NDT_PT',  'Penetrant Testing (PT)',           'ndt'),
  ('NDT_MT',  'Magnetic Particle Testing (MT)',   'ndt'),
  ('NDT_UT',  'Ultrasonic Testing (UT)',          'ndt'),
  ('NDT_ET',  'Eddy Current Testing (ET)',        'ndt'),
  ('NDT_RT',  'Radiographic Testing (RT)',        'ndt'),
  -- Structural
  ('WELD_TIG',   'TIG Welding',                   'structural'),
  ('WELD_MIG',   'MIG Welding',                   'structural'),
  ('COMPOSITE',  'Composite Repair',              'structural'),
  ('SHEET_METAL','Sheet Metal Repair',             'structural'),
  -- Avionics
  ('AVIONICS_NAV',  'Navigation Systems',          'avionics'),
  ('AVIONICS_COM',  'Communication Systems',       'avionics'),
  ('AVIONICS_FMS',  'Flight Management Systems',   'avionics'),
  ('AVIONICS_RADAR','Weather Radar Systems',       'avionics'),
  -- Engine
  ('ENGINE_TURBOFAN', 'Turbofan Engine Overhaul',  'engine'),
  ('ENGINE_APU',      'APU Maintenance',           'engine'),
  -- General
  ('FUEL_TANK',  'Fuel Tank Entry & Repair',       'general'),
  ('HYDRAULIC',  'Hydraulic Systems',              'general'),
  ('PNEUMATIC',  'Pneumatic Systems',              'general')
ON CONFLICT (code) DO NOTHING;

-- Seed compliance templates
INSERT INTO compliance_templates (authority_id, template_code, name, description, required_fields)
SELECT ra.id, v.template_code, v.name, v.description, v.required_fields::jsonb
FROM (VALUES
  ('FAA', 'FAA_337', 'FAA Form 337 - Major Repair/Alteration',
   'Required for major repairs and alterations per 14 CFR Part 43',
   '{"fields": ["aircraft_id", "registration_number", "repair_station_number", "description_of_work", "approval_basis", "data_used", "weight_balance_change", "inspector_signature", "date"]}'),

  ('FAA', 'FAA_8610', 'FAA Form 8610-2 - Airman Certificate/Rating',
   'Application for airman certificate or rating',
   '{"fields": ["applicant_name", "certificate_type", "rating_requested", "experience_log", "examiner_signature", "date"]}'),

  ('EASA', 'EASA_FORM1', 'EASA Form 1 - Authorized Release Certificate',
   'Certificate of Release to Service for components per Part-145',
   '{"fields": ["organization_name", "approval_reference", "part_number", "serial_number", "description", "work_performed", "remarks", "certifying_staff_signature", "date"]}'),

  ('EASA', 'EASA_CRS', 'Certificate of Release to Service',
   'EASA Part-145 maintenance release certificate',
   '{"fields": ["aircraft_registration", "aircraft_type", "work_order_number", "description_of_work", "limitations", "certifying_staff_name", "authorization_number", "signature", "date"]}'),

  ('ECAA', 'ECAA_CRS', 'ECAA Certificate of Release to Service',
   'Ethiopian CAA maintenance release certificate',
   '{"fields": ["aircraft_registration", "aircraft_type", "work_performed", "reference_documents", "technician_name", "license_number", "signature", "date"]}'),

  ('ICAO', 'MAINT_RELEASE', 'Maintenance Release (Generic)',
   'ICAO Annex 6 compliant maintenance release',
   '{"fields": ["aircraft_registration", "description_of_work", "reference_documents", "certifying_person", "license_number", "signature", "date"]}')
) AS v(authority_code, template_code, name, description, required_fields)
JOIN regulatory_authorities ra ON ra.code = v.authority_code
ON CONFLICT (authority_id, template_code) DO NOTHING;

-- +goose Down
DELETE FROM compliance_templates;
DELETE FROM skill_types;
DELETE FROM aircraft_types;
DELETE FROM certification_types;
DELETE FROM regulatory_authorities;
