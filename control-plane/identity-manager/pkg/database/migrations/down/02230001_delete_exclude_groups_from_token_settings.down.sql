ALTER TABLE identity_manager.realms
ADD COLUMN IF NOT EXISTS exclude_groups_from_tokens BOOLEAN DEFAULT FALSE;

