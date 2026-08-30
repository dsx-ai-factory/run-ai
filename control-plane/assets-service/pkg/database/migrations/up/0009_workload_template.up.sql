DROP TRIGGER IF EXISTS validate_asset_name ON assets_metadata;

update assets_metadata set kind = 'workload-template' where kind = 'workspace-template';

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER validate_asset_name
            BEFORE INSERT OR UPDATE
            ON assets_metadata
            FOR EACH ROW
        EXECUTE FUNCTION asset_name_validation();
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;