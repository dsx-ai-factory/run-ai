DROP TRIGGER IF EXISTS validate_asset_name ON assets_metadata;
DROP FUNCTION IF EXISTS asset_name_validation();

ALTER TABLE assets_metadata
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_by TEXT;

-- we used this sequence to mark every update on the table with increasing number
-- which will be used by sync mechanism
CREATE SEQUENCE IF NOT EXISTS assets_metadata_update_count;

ALTER TABLE assets_metadata ADD COLUMN IF NOT EXISTS update_count BIGINT;

CREATE OR REPLACE FUNCTION assets_metadata_update_count()
RETURNS TRIGGER AS $$
BEGIN
    NEW.update_count := nextval('assets_metadata_update_count');
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
BEGIN
CREATE TRIGGER set_update_count BEFORE INSERT OR UPDATE ON assets_metadata
                                                     FOR EACH ROW
                                                     EXECUTE PROCEDURE assets_metadata_update_count();
EXCEPTION
        WHEN duplicate_object THEN NULL;
END;
END $$;

-- forced update allows us to forcefully mark this record as updated and thus increase the update counter
ALTER TABLE assets_metadata ADD COLUMN IF NOT EXISTS forced_update timestamptz;

-- use the forced update field to give increasing sequence number to all the existing records.
-- if we have 5 records in the db, they will be timestamped with numbers 1 to 5
UPDATE assets_metadata SET forced_update = NOW();

-- we extract records sorted by update count, so better set an index on this field
CREATE INDEX IF NOT EXISTS assets_metadata_by_update ON assets_metadata (update_count);
