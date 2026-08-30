---- Redo of 0013_add_assets_parameters.up.sql in idempotent in case it didn't run
ALTER TABLE assets_environment ADD COLUMN if not exists run_as_non_root boolean;
ALTER TABLE assets_environment ADD COLUMN if not exists seccomp_profile_type text;