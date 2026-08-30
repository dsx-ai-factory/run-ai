-- Migration to add users-minimal permission set to all custom roles.
-- This migration should only run once. If users-minimal already exists, skip entirely.
DO $$
BEGIN
    -- Guard: if users-minimal already exists anywhere, this migration already ran
    IF EXISTS (
        SELECT 1 FROM permission_sets WHERE name = 'Users minimal read'
    ) THEN
        RETURN;
    END IF;

    -- Step 0: Create the users-minimal permission set
    INSERT INTO permission_sets (id, name, description, internal, created_at, updated_at)
    VALUES (gen_random_uuid(), 'Users minimal read', 'Grants permissions to view minimal users data.', false, NOW(), NOW());

    -- Step 1: Custom roles referencing static permission sets via role_to_permission_set
    -- All custom roles get the "Users minimal read" permission set
    INSERT INTO role_to_permission_set (role_id, permission_set_id)
    SELECT DISTINCT rtps.role_id, ps_um.id
    FROM role_to_permission_set rtps
    INNER JOIN permission_sets ps_um ON ps_um.name = 'Users minimal read'
    INNER JOIN roles r ON r.id = rtps.role_id
    WHERE r.custom = true
      AND EXISTS (
        SELECT 1 FROM role_to_permission_set rtps3
        INNER JOIN permission_sets ps3 ON ps3.id = rtps3.permission_set_id
        WHERE rtps3.role_id = rtps.role_id
          AND ps3.internal = false
      )
      AND NOT EXISTS (
        SELECT 1 FROM role_to_permission_set rtps2
        WHERE rtps2.role_id = rtps.role_id
          AND rtps2.permission_set_id = ps_um.id
      );

    -- Step 2: Legacy custom roles whose permissions were migrated to a single internal permission set.
    -- Scan permission_sets directly so we emit at most one users-minimal row per permission set;
    -- scanning role_permissions would multiply output rows by the number of existing resource_types
    -- on the same permission set, and gen_random_uuid() defeats SELECT DISTINCT.
    INSERT INTO role_permissions (id, permission_set_id, resource_type, "create", read, update, delete, sync, created_at, updated_at)
    SELECT
        gen_random_uuid() AS id,
        ps.id AS permission_set_id,
        'users-minimal' AS resource_type,
        false AS "create",
        true AS read,
        false AS update,
        false AS delete,
        false AS sync,
        NOW() AS created_at,
        NOW() AS updated_at
    FROM permission_sets ps
    WHERE ps.internal = true
      AND ps.name LIKE '%-internal-ps'
      AND EXISTS (
        SELECT 1 FROM role_permissions rp WHERE rp.permission_set_id = ps.id
      )
      AND EXISTS (
        SELECT 1
        FROM role_to_permission_set rtps
        INNER JOIN roles r ON r.id = rtps.role_id
        WHERE rtps.permission_set_id = ps.id
          AND r.custom = true
          AND (SELECT COUNT(*) FROM role_to_permission_set rtps2 WHERE rtps2.role_id = r.id) = 1
      )
      AND NOT EXISTS (
        SELECT 1 FROM role_permissions rp2
        WHERE rp2.permission_set_id = ps.id
          AND rp2.resource_type = 'users-minimal'
      );

END $$;