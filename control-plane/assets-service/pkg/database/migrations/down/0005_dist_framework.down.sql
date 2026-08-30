UPDATE assets_pvc SET ephemeral = false WHERE ephemeral IS NULL;
ALTER TABLE assets_pvc ALTER COLUMN ephemeral SET NOT NULL;