-- delete the information we added to the sync status as part of transition from sync to non-sync
-- in order to allow the old sync mechanism to utilize it again
DELETE FROM sync_status;

DROP TABLE IF EXISTS approved_deletions;
