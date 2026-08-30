-- The V2 asset-sync API has been removed. Its bookkeeping table is obsolete;
-- the current agent sync flow uses the separate sync_status table.
DROP TABLE IF EXISTS assets_sync_status;
