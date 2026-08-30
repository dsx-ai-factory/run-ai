DELETE FROM identity_manager.realm_settings
WHERE setting_key = 'excludeGroupsFromToken';

ALTER TABLE identity_manager.realms
DROP COLUMN IF EXISTS exclude_groups_from_tokens; 