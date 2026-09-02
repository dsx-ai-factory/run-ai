-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- as we no longer support soft-delete following this rollback, need to
-- physically remove anything that was marked for deletion                     
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'assets'
      AND table_name = 'assets_metadata'
      AND column_name = 'deleted_at'
  ) THEN
DELETE FROM assets_metadata WHERE deleted_at IS NOT NULL;
END IF;
END $$;

ALTER TABLE assets_metadata
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS deleted_by;

DROP TRIGGER IF EXISTS set_update_count ON assets_metadata;

DROP INDEX IF EXISTS assets_metadata_by_update;

ALTER TABLE assets_metadata DROP COLUMN IF EXISTS update_count;
ALTER TABLE assets_metadata DROP COLUMN IF EXISTS forced_update;

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