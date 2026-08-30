CREATE OR REPLACE FUNCTION asset_name_validation()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.name IN (select name FROM assets_metadata WHERE NEW.kind = assets_metadata.kind AND
        ((NEW.tenant_id is NULL and assets_metadata.tenant_id is NULL) or NEW.tenant_id = assets_metadata.tenant_id) AND NEW.scope != assets_metadata.scope)
    THEN
        RAISE EXCEPTION 'non-% % asset named % already exists', NEW.scope, NEW.kind, NEW.name;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS validate_asset_name ON assets_metadata;
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
