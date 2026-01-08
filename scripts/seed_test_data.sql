-- AMSS Test Data Seed Script
-- This script seeds comprehensive test data for testing permissions and features

-- Use the existing org_id
DO $$
DECLARE
    org_id UUID := '4cb97629-c58a-415d-bf9c-b400bb5e3d84';

    -- Aircraft IDs
    aircraft_1 UUID := 'e9b6d5d9-bea6-4fd7-8e70-f7c8ee6a91e1'; -- Existing
    aircraft_2 UUID := 'a1b2c3d4-e5f6-4789-abcd-ef0123456789';
    aircraft_3 UUID := 'b2c3d4e5-f6a7-4890-bcde-f01234567890';
    aircraft_4 UUID := 'c3d4e5f6-a7b8-4901-cdef-012345678901';
    aircraft_5 UUID := 'd4e5f6a7-b8c9-4012-defa-123456789012';
    aircraft_6 UUID := 'e5f6a7b8-c9d0-4123-efab-234567890123';

    -- Part Definition IDs
    part_def_1 UUID := 'f6a7b8c9-d0e1-4234-fabc-345678901234';
    part_def_2 UUID := 'a7b8c9d0-e1f2-4345-abcd-456789012345';
    part_def_3 UUID := 'b8c9d0e1-f2a3-4456-bcde-567890123456';
    part_def_4 UUID := 'c9d0e1f2-a3b4-4567-cdef-678901234567';
    part_def_5 UUID := 'd0e1f2a3-b4c5-4678-defa-789012345678';
    part_def_6 UUID := 'e1f2a3b4-c5d6-4789-efab-890123456789';

    -- Part Item IDs
    part_item_1 UUID := 'f2a3b4c5-d6e7-4890-fabc-901234567890';
    part_item_2 UUID := 'a3b4c5d6-e7f8-4901-abcd-012345678901';
    part_item_3 UUID := 'b4c5d6e7-f8a9-4012-bcde-123456789012';
    part_item_4 UUID := 'c5d6e7f8-a9b0-4123-cdef-234567890123';
    part_item_5 UUID := 'd6e7f8a9-b0c1-4234-defa-345678901234';
    part_item_6 UUID := 'e7f8a9b0-c1d2-4345-efab-456789012345';
    part_item_7 UUID := 'f8a9b0c1-d2e3-4456-fabc-567890123456';
    part_item_8 UUID := 'a9b0c1d2-e3f4-4567-abcd-678901234567';

    -- Task IDs
    task_1 UUID := '0ceea263-adc6-4414-b3bc-7db277f5809c'; -- Existing
    task_2 UUID := 'b0c1d2e3-f4a5-4678-bcde-789012345678';
    task_3 UUID := 'c1d2e3f4-a5b6-4789-cdef-890123456789';
    task_4 UUID := 'd2e3f4a5-b6c7-4890-defa-901234567890';
    task_5 UUID := 'e3f4a5b6-c7d8-4901-efab-012345678901';

    -- User IDs (get from existing users)
    mechanic_user_id UUID;
    scheduler_user_id UUID;

BEGIN
    -- Get user IDs
    SELECT id INTO mechanic_user_id FROM users WHERE email = 'mechanic@demo.local' AND deleted_at IS NULL;
    SELECT id INTO scheduler_user_id FROM users WHERE email = 'scheduler@demo.local' AND deleted_at IS NULL;

    -- =====================================================
    -- AIRCRAFT
    -- =====================================================

    -- Insert additional aircraft (skip if exists)
    INSERT INTO aircraft (id, org_id, tail_number, model, status, capacity_slots, flight_hours_total, cycles_total, last_maintenance, next_due, created_at, updated_at)
    VALUES
        (aircraft_2, org_id, 'N200AM', 'Airbus A320', 'operational', 4, 18200, 12000, NOW() - INTERVAL '60 days', NOW() + INTERVAL '30 days', NOW(), NOW()),
        (aircraft_3, org_id, 'N300AM', 'Embraer E175', 'maintenance', 2, 12800, 9500, NOW() - INTERVAL '90 days', NULL, NOW(), NOW()),
        (aircraft_4, org_id, 'N400AM', 'Boeing 777-300', 'operational', 6, 42100, 8500, NOW() - INTERVAL '30 days', NOW() + INTERVAL '180 days', NOW(), NOW()),
        (aircraft_5, org_id, 'N500AM', 'Airbus A321', 'operational', 4, 15600, 11000, NOW() - INTERVAL '7 days', NOW() + INTERVAL '60 days', NOW(), NOW()),
        (aircraft_6, org_id, 'N600AM', 'Boeing 737 MAX 8', 'grounded', 4, 8900, 5200, NOW() - INTERVAL '45 days', NULL, NOW(), NOW())
    ON CONFLICT (id) DO NOTHING;

    -- =====================================================
    -- PART DEFINITIONS
    -- =====================================================

    INSERT INTO part_definitions (id, org_id, name, category, created_at, updated_at)
    VALUES
        (part_def_1, org_id, 'Brake Assembly', 'Landing Gear', NOW(), NOW()),
        (part_def_2, org_id, 'Engine Fan Blade', 'Engine', NOW(), NOW()),
        (part_def_3, org_id, 'Hydraulic Pump', 'Hydraulics', NOW(), NOW()),
        (part_def_4, org_id, 'Avionics Display Unit', 'Avionics', NOW(), NOW()),
        (part_def_5, org_id, 'Tire Assembly Main', 'Landing Gear', NOW(), NOW()),
        (part_def_6, org_id, 'APU Starter Motor', 'APU', NOW(), NOW())
    ON CONFLICT (id) DO NOTHING;

    -- =====================================================
    -- PART ITEMS
    -- =====================================================

    INSERT INTO part_items (id, org_id, part_definition_id, serial_number, status, expiry_date, created_at, updated_at)
    VALUES
        (part_item_1, org_id, part_def_1, 'BRK-2024-001', 'in_stock', NOW() + INTERVAL '2 years', NOW(), NOW()),
        (part_item_2, org_id, part_def_1, 'BRK-2024-002', 'in_stock', NOW() + INTERVAL '2 years', NOW(), NOW()),
        (part_item_3, org_id, part_def_2, 'EFB-2024-001', 'in_stock', NOW() + INTERVAL '5 years', NOW(), NOW()),
        (part_item_4, org_id, part_def_3, 'HYD-2024-001', 'in_stock', NOW() + INTERVAL '3 years', NOW(), NOW()),
        (part_item_5, org_id, part_def_3, 'HYD-2024-002', 'used', NULL, NOW(), NOW()),
        (part_item_6, org_id, part_def_4, 'ADU-2024-001', 'in_stock', NOW() + INTERVAL '10 years', NOW(), NOW()),
        (part_item_7, org_id, part_def_5, 'TIR-2024-001', 'in_stock', NOW() + INTERVAL '1 year', NOW(), NOW()),
        (part_item_8, org_id, part_def_6, 'APU-2024-001', 'disposed', NULL, NOW(), NOW())
    ON CONFLICT (id) DO NOTHING;

    -- =====================================================
    -- MAINTENANCE TASKS
    -- =====================================================

    INSERT INTO maintenance_tasks (id, org_id, aircraft_id, type, state, start_time, end_time, assigned_mechanic_id, notes, created_at, updated_at)
    VALUES
        (task_2, org_id, aircraft_2, 'inspection', 'in_progress', NOW() - INTERVAL '1 day', NOW() + INTERVAL '2 days', mechanic_user_id, 'Engine borescope inspection on both engines', NOW(), NOW()),
        (task_3, org_id, aircraft_3, 'overhaul', 'scheduled', NOW() + INTERVAL '3 days', NOW() + INTERVAL '10 days', NULL, 'Complete overhaul of main landing gear assembly', NOW(), NOW()),
        (task_4, org_id, aircraft_4, 'repair', 'completed', NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days', mechanic_user_id, 'Repair auxiliary power unit following fault indication', NOW(), NOW()),
        (task_5, org_id, aircraft_5, 'inspection', 'scheduled', NOW() + INTERVAL '7 days', NOW() + INTERVAL '9 days', NULL, 'Routine A-Check maintenance per schedule', NOW(), NOW())
    ON CONFLICT (id) DO NOTHING;

    -- =====================================================
    -- COMPLIANCE ITEMS
    -- =====================================================

    INSERT INTO compliance_items (id, org_id, task_id, description, result, sign_off_user_id, sign_off_time, created_at, updated_at)
    VALUES
        (gen_random_uuid(), org_id, task_1, 'Pre-flight inspection checklist completed', 'pass', mechanic_user_id, NOW() - INTERVAL '1 day', NOW(), NOW()),
        (gen_random_uuid(), org_id, task_1, 'Engine oil levels verified', 'pass', mechanic_user_id, NOW() - INTERVAL '1 day', NOW(), NOW()),
        (gen_random_uuid(), org_id, task_2, 'Left engine borescope - no defects found', 'pass', NULL, NULL, NOW(), NOW()),
        (gen_random_uuid(), org_id, task_2, 'Right engine borescope - minor wear noted', 'pending', NULL, NULL, NOW(), NOW()),
        (gen_random_uuid(), org_id, task_4, 'APU fault code cleared', 'pass', mechanic_user_id, NOW() - INTERVAL '3 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, task_4, 'APU ground test successful', 'pass', mechanic_user_id, NOW() - INTERVAL '3 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, task_4, 'Documentation updated in maintenance log', 'pass', mechanic_user_id, NOW() - INTERVAL '3 days', NOW(), NOW())
    ON CONFLICT DO NOTHING;

    -- =====================================================
    -- MAINTENANCE PROGRAMS
    -- =====================================================

    INSERT INTO maintenance_programs (id, org_id, aircraft_id, name, interval_type, interval_value, last_performed, created_at, updated_at)
    VALUES
        (gen_random_uuid(), org_id, aircraft_1, 'A-Check', 'flight_hours', 500, NOW() - INTERVAL '30 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, aircraft_1, 'B-Check', 'flight_hours', 2000, NOW() - INTERVAL '90 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, aircraft_2, 'A-Check', 'flight_hours', 500, NOW() - INTERVAL '45 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, aircraft_4, 'C-Check', 'calendar', 365, NOW() - INTERVAL '180 days', NOW(), NOW()),
        (gen_random_uuid(), org_id, NULL, 'Landing Gear Overhaul', 'cycles', 10000, NULL, NOW(), NOW())
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Test data seeded successfully!';
END $$;
