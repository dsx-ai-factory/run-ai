-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The DGXC role mechanism (roles.role_type) is being removed.
--
-- If nothing references the DGXC roles - no access rule is assigned to one, and no
-- custom role names one as its predefined cluster role - they are simply deleted.
--
-- Otherwise they are kept, but handed over: '-DGXC' is dropped from the name, a
-- '-legacy' suffix is added, and custom is set so the static role definitions no
-- longer own them. No rows are added and no access rule is touched - each role keeps
-- its id, so access rules and cluster_role_predefined_id stay valid.

DO $$
DECLARE
    in_use BOOLEAN;
BEGIN
    SELECT
        -- assigned to someone
        EXISTS (SELECT 1
                FROM access_rules a
                JOIN roles r ON r.id = a.role_id
                WHERE r.role_type = 'DGXC'
                  AND r.custom = false
                  AND r.scope_type = 'system')
        -- or serving as the predefined cluster role of a custom role. roles are hard
        -- deleted - the table has no deleted_at - so every row here is live.
        OR EXISTS (SELECT 1
                   FROM roles c
                   JOIN roles r ON r.id = c.cluster_role_predefined_id
                   WHERE c.custom = true
                     AND r.role_type = 'DGXC'
                     AND r.custom = false
                     AND r.scope_type = 'system')
    INTO in_use;

    IF in_use THEN
        -- "Cloud operator-DGXC"                    -> "Cloud operator-legacy"
        -- "Department administrator-DGXC (Legacy)" -> "Department administrator (Legacy)-legacy"
        -- "Application administrator"              -> "Application administrator-legacy"
        UPDATE roles
        SET name = replace(name, '-DGXC', '') || '-legacy',
            custom = true,
            updated_at = NOW()
        WHERE role_type = 'DGXC'
          AND custom = false
          AND scope_type = 'system';
    ELSE
        -- role_to_permission_set has no ON DELETE clause and would block the delete.
        -- Permission sets themselves are shared and are left alone.
        DELETE FROM role_to_permission_set
        WHERE role_id IN (
            SELECT id FROM roles
            WHERE role_type = 'DGXC' AND custom = false AND scope_type = 'system'
        );

        -- Cascades: permissions, disabled_roles. No access rule can reference these -
        -- that is what put us on this branch.
        DELETE FROM roles
        WHERE role_type = 'DGXC' AND custom = false AND scope_type = 'system';
    END IF;
END $$;
