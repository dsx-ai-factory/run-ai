CREATE OR REPLACE FUNCTION update_pods_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.phase <> OLD.phase THEN
        NEW.phase_updated_at = NOW();
    ELSE
        NEW.phase_updated_at = OLD.phase_updated_at;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_pods_phases_update_time
            BEFORE INSERT OR UPDATE
            ON pods
            FOR EACH ROW
        EXECUTE FUNCTION update_pods_phases_update_time();
    EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;