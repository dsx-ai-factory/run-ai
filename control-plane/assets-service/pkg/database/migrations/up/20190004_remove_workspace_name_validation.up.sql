----------------------------------------------------------------------------------
-- Remove workload name validation logic
----------------------------------------------------------------------------------

-- Drop the trigger if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'workloads'::regclass AND tgname = 'validate_workload_name') THEN
        DROP TRIGGER validate_workload_name ON workloads;
    END IF;
END$$;

-- Drop the validation function if it exists
DROP FUNCTION IF EXISTS workload_name_validation();