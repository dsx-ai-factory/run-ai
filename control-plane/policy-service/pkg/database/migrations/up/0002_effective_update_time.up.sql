ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL default NOW();

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_timestamp BEFORE UPDATE ON effective_policies
        FOR EACH ROW
        EXECUTE PROCEDURE trigger_set_timestamp();
    EXCEPTION
        WHEN duplicate_object THEN
           NULL;
    END;
END $$;

