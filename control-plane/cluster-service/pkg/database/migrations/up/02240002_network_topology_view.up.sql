-- Create a view that lists each network topology and the nodepools using it
-- Nodepools are associated via nodepools.network_topology_id matching network_topologies.id
-- Both network topologies and nodepools are filtered to non-deleted entries

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

