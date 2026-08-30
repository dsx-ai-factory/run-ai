ALTER TABLE reconciler_versions
RENAME COLUMN user_applications TO access_keys;

ALTER TABLE reconciler_versions
RENAME COLUMN applications TO service_accounts;

ALTER TABLE reconciler_versions
RENAME COLUMN internal_applications TO internal_service_accounts;

