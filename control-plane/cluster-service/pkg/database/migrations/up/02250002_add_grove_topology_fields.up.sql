-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE network_topologies
    ADD COLUMN IF NOT EXISTS grove_name TEXT,
    ADD COLUMN IF NOT EXISTS grove_domains TEXT[],
    ADD COLUMN IF NOT EXISTS grove_conditions JSONB;

-- Recreate the view to pick up the new columns
DROP VIEW IF EXISTS network_topology_view;
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
