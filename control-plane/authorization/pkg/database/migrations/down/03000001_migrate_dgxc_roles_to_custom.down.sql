-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Reverses the rename branch of 03000001: drops the '-Unmanaged' suffix, puts '-DGXC'
-- back where it was, and clears custom.
--
-- The mapping is spelled out rather than derived, because stripping '-Unmanaged' does not
-- say where '-DGXC' belonged - "Cloud operator-Unmanaged" came from "Cloud operator-DGXC"
-- but "Application administrator-Unmanaged" came from "Application administrator", and
-- in "Department administrator-DGXC (Legacy)" the marker sat mid-name.
--
-- The delete branch cannot be reversed: those roles are gone and have to come back
-- from the static role definitions.

-- The static definitions all carry ClusterRolePredefinedId: nil, so the pointer the up
-- migration set is cleared rather than restored.
UPDATE roles r
SET name = m.old_name,
    custom = false,
    role_type = 'DGXC',
    cluster_role_predefined_id = NULL,
    updated_at = NOW()
FROM (VALUES
    ('Application administrator-Unmanaged',             'Application administrator'),
    ('Cloud support-Unmanaged',                         'Cloud support'),
    ('Cloud operator-Unmanaged',                        'Cloud operator-DGXC'),
    ('Department administrator-Unmanaged',              'Department administrator-DGXC'),
    ('Department administrator (Legacy)-Unmanaged',     'Department administrator-DGXC (Legacy)')
) AS m(new_name, old_name)
WHERE r.name = m.new_name
  AND r.scope_type = 'system'
  AND r.custom = true
  AND r.created_by = 'Run:ai';
