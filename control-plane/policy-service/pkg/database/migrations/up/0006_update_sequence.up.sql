-- we used this sequence to mark every update on the table with increasing number
-- which will be used as the timestamp for the sync.
CREATE SEQUENCE IF NOT EXISTS effective_policies_update_count;

ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS update_count BIGINT;

CREATE OR REPLACE FUNCTION effective_policies_update_count()
RETURNS TRIGGER AS $$
BEGIN
    NEW.update_count := nextval('effective_policies_update_count');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_update_count BEFORE INSERT OR UPDATE ON effective_policies
    FOR EACH ROW
        EXECUTE PROCEDURE effective_policies_update_count();
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;

-- forced update allows us to forcefully mark this record as updated. it will be used mainly for
-- signalling indirect changes in the policy (imposed assets, for example) 
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS forced_update timestamptz;

-- use the forced update field to give increasing sequence number to all the existing records. 
-- if we have 5 records in the db, they will be timestamped with numbers 1 to 5
UPDATE effective_policies SET forced_update = NOW();

-- we extract records sorted by update count, so better set an index on this field
CREATE INDEX IF NOT EXISTS effective_policies_by_update ON effective_policies (update_count);
