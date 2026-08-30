ALTER TABLE tenants ADD COLUMN IF NOT EXISTS alias TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_alias ON tenants (alias) WHERE alias IS NOT NULL;

-- Enforce cross-field uniqueness between tenant name and alias columns.
CREATE OR REPLACE FUNCTION tenant_name_alias_cross_uniqueness() RETURNS TRIGGER AS $$
BEGIN
    -- Check if the new/updated name conflicts with any existing alias
    IF EXISTS (
        SELECT 1 FROM tenants
        WHERE alias = NEW.name
          AND id != NEW.id
    ) THEN
        RAISE EXCEPTION 'tenant name "%" conflicts with an existing tenant alias', NEW.name
            USING ERRCODE = '23505';
    END IF;

    -- Check if the new/updated alias conflicts with any existing name
    IF NEW.alias IS NOT NULL THEN
        IF EXISTS (
            SELECT 1 FROM tenants
            WHERE name = NEW.alias
              AND id != NEW.id
        ) THEN
            RAISE EXCEPTION 'tenant alias "%" conflicts with an existing tenant name', NEW.alias
                USING ERRCODE = '23505';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_tenant_name_alias_cross_uniqueness
    BEFORE INSERT OR UPDATE OF name, alias ON tenants
    FOR EACH ROW EXECUTE PROCEDURE tenant_name_alias_cross_uniqueness();
