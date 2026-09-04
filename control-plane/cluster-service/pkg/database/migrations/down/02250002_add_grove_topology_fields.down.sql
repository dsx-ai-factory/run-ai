-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Drop the view first since it depends on the columns via nt.*
DROP VIEW IF EXISTS network_topology_view;

ALTER TABLE network_topologies
    DROP COLUMN IF EXISTS grove_name,
    DROP COLUMN IF EXISTS grove_domains,
    DROP COLUMN IF EXISTS grove_conditions;

-- Recreate the view without the dropped columns
CREATE VIEW network_topology_view AS
SELECT
    nt.*,
    COALESCE(
        (
            SELECT jsonb_agg(np.name ORDER BY np.name)
            FROM nodepools np
            WHERE np.deleted_at IS NULL
              AND np.network_topology_id = nt.id::text
        ),
        '[]'::jsonb
    ) AS nodepools
FROM network_topologies nt
WHERE nt.deleted_at IS NULL;
