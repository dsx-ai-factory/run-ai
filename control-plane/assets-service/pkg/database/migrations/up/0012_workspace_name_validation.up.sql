----------------------------------------------------------------------------------
-- Workload name validator
----------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION workload_name_validation()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.name IN (select name FROM workloads WHERE NEW.tenant_id = workloads.tenant_id AND
            NEW.cluster_id = workloads.cluster_id AND NEW.project_id = workloads.project_id)
  THEN
      RAISE EXCEPTION 'workload named % already exists', NEW.name;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

----------------------------------------------------------------------------------
-- Apply Name Validation
----------------------------------------------------------------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'workloads'::regclass AND tgname = 'validate_workload_name') THEN
        CREATE TRIGGER validate_workload_name
            BEFORE INSERT OR UPDATE ON workloads
            FOR EACH ROW
            EXECUTE PROCEDURE workload_name_validation();
    END IF;
END$$;

