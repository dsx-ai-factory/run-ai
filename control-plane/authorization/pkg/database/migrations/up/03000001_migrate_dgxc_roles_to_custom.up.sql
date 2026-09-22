-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The DGXC role mechanism (roles.role_type) is being removed.
--
-- Unreferenced DGXC roles are deleted. The rest are handed over: renamed to
-- ' - superseded' and marked custom. No rows are added and no access rule is touched -
-- each role keeps its id, so everything pointing at it stays valid.

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
        -- or named as a custom role's predefined cluster role
        OR EXISTS (SELECT 1
                   FROM roles c
                   JOIN roles r ON r.id = c.cluster_role_predefined_id
                   WHERE c.custom = true
                     AND r.role_type = 'DGXC'
                     AND r.custom = false
                     AND r.scope_type = 'system')
    INTO in_use;

    IF in_use THEN
        -- Point each role at whatever its bindings resolve to today, so the rename
        -- stays cosmetic: the agent prefers PredefinedRole.Name over the role's own
        -- name, and no ClusterRole ships under the renamed form.
        --
        --   "Cloud operator-DGXC"                    -> "Cloud operator"
        --   "Department administrator-DGXC"          -> "Department administrator"
        --   "Department administrator-DGXC (Legacy)" -> "Department administrator (Legacy)"
        --   "Application administrator"              -> "Editor"   ) RoleNamesMapping in
        --   "Cloud support"                          -> "Viewer"   ) the cluster, which
        --                                                            the rename would miss
        --
        -- The subquery matches on the pre-update snapshot, i.e. the name before the
        -- rename. It sits in SET, not FROM: a FROM clause is an inner join, and a role
        -- whose target is missing would not be renamed at all. COALESCE leaves a
        -- pointer that is already set alone.
        UPDATE roles d
        SET name = replace(d.name, '-DGXC', '') || ' - superseded',
            custom = true,
            cluster_role_predefined_id = COALESCE(
                d.cluster_role_predefined_id,
                (SELECT t.id
                 FROM roles t
                 WHERE t.scope_type = 'system'
                   AND t.role_type <> 'DGXC'
                   AND t.id <> d.id
                   AND t.name = CASE d.name
                                    WHEN 'Application administrator' THEN 'Editor'
                                    WHEN 'Cloud support'             THEN 'Viewer'
                                    ELSE replace(d.name, '-DGXC', '')
                                END)
            ),
            updated_at = NOW()
        WHERE d.role_type = 'DGXC'
          AND d.custom = false
          AND d.scope_type = 'system';

        -- Flatten any chain the rename just created. A custom role could already name
        -- one of these as its predefined cluster role; getPredefinedRoleForSync
        -- resolves a single hop, so it would hand the agent the ' - superseded' name and
        -- the binding would find no ClusterRole. Point it at the same target the
        -- renamed role now uses.
        UPDATE roles c
        SET cluster_role_predefined_id = d.cluster_role_predefined_id,
            updated_at = NOW()
        FROM roles d
        WHERE c.cluster_role_predefined_id = d.id
          AND d.role_type = 'DGXC'
          AND d.custom = true
          AND d.scope_type = 'system'
          AND d.name LIKE '% - superseded'
          AND d.cluster_role_predefined_id IS NOT NULL;
    ELSE
        -- role_to_permission_set has no ON DELETE clause and would block the delete.
        -- The permission sets themselves are shared and are left alone.
        DELETE FROM role_to_permission_set
        WHERE role_id IN (
            SELECT id FROM roles
            WHERE role_type = 'DGXC' AND custom = false AND scope_type = 'system'
        );

        -- Cascades: permissions, disabled_roles. Nothing references these - that is
        -- what put us on this branch.
        DELETE FROM roles
        WHERE role_type = 'DGXC' AND custom = false AND scope_type = 'system';
    END IF;
END $$;
