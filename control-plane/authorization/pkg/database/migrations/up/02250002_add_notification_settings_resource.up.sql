-- Migration to add notification-settings permissions mirroring security-settings for custom roles
-- This migration should only run once. If notification-settings already exists, skip entirely.

DO $$
BEGIN
    -- Guard: if notification-settings already exists anywhere, this migration already ran
    IF EXISTS (
        SELECT 1 FROM role_permissions WHERE resource_type = 'notification-settings'
    ) OR EXISTS (
        SELECT 1 FROM permission_sets WHERE name IN ('Notification settings edit', 'Notification settings read')
    ) THEN
        RETURN;
    END IF;

    -- Step 0: Create the notification-settings permission sets
    INSERT INTO permission_sets (id, name, description, internal, created_at, updated_at)
    VALUES (gen_random_uuid(), 'Notification settings edit', 'Grants permissions to create, read, update, and delete notification settings.', false, NOW(), NOW());

    INSERT INTO permission_sets (id, name, description, internal, created_at, updated_at)
    VALUES (gen_random_uuid(), 'Notification settings read', 'Grants permissions to view notification settings.', false, NOW(), NOW());

    -- Step 1: Custom roles referencing static permission sets via role_to_permission_set
    -- If a custom role references "Security settings edit", also add "Notification settings edit"
    INSERT INTO role_to_permission_set (role_id, permission_set_id)
    SELECT rtps.role_id, ps_notif.id
    FROM role_to_permission_set rtps
    INNER JOIN permission_sets ps_sec ON ps_sec.id = rtps.permission_set_id
    INNER JOIN permission_sets ps_notif ON ps_notif.name = 'Notification settings edit'
    INNER JOIN roles r ON r.id = rtps.role_id
    WHERE ps_sec.name = 'Security settings edit'
      AND r.custom = true;

    -- If a custom role references "Security settings read", also add "Notification settings read"
    INSERT INTO role_to_permission_set (role_id, permission_set_id)
    SELECT rtps.role_id, ps_notif.id
    FROM role_to_permission_set rtps
    INNER JOIN permission_sets ps_sec ON ps_sec.id = rtps.permission_set_id
    INNER JOIN permission_sets ps_notif ON ps_notif.name = 'Notification settings read'
    INNER JOIN roles r ON r.id = rtps.role_id
    WHERE ps_sec.name = 'Security settings read'
      AND r.custom = true;

    -- Step 2: Legacy custom roles whose permissions were migrated to internal permission sets
    INSERT INTO role_permissions (id, permission_set_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
    SELECT
        gen_random_uuid() AS id,
        rp.permission_set_id,
        'notification-settings' AS resource_type,
        rp."create",
        rp.read,
        rp.update,
        rp.delete,
        rp.sync,
        NOW() AS created_at,
        NOW() AS updated_at
    FROM role_permissions rp
    INNER JOIN permission_sets ps ON ps.id = rp.permission_set_id
    WHERE rp.resource_type = 'security-settings'
      AND ps.internal = true;

END $$;
