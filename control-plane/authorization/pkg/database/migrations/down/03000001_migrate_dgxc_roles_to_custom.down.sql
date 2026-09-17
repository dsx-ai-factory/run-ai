-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Reverses the rename branch of 03000001: drops the '-legacy' suffix, puts '-DGXC'
-- back where it was, and clears custom.
--
-- The mapping is spelled out rather than derived, because stripping '-legacy' does not
-- say where '-DGXC' belonged - "Cloud operator-legacy" came from "Cloud operator-DGXC"
-- but "Application administrator-legacy" came from "Application administrator", and
-- in "Department administrator-DGXC (Legacy)" the marker sat mid-name.
--
-- The delete branch cannot be reversed: those roles are gone and have to come back
-- from the static role definitions.

UPDATE roles r
SET name = m.old_name,
    custom = false,
    updated_at = NOW()
FROM (VALUES
    ('Application administrator-legacy',             'Application administrator'),
    ('Cloud support-legacy',                         'Cloud support'),
    ('Cloud operator-legacy',                        'Cloud operator-DGXC'),
    ('Department administrator-legacy',              'Department administrator-DGXC'),
    ('Department administrator (Legacy)-legacy',     'Department administrator-DGXC (Legacy)')
) AS m(new_name, old_name)
WHERE r.name = m.new_name
  AND r.scope_type = 'system'
  AND r.custom = true
  AND r.created_by = 'Run:ai';
