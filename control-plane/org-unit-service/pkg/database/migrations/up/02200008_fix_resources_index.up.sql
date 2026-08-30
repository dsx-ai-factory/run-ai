DROP INDEX IF EXISTS pool_to_org_unit_data;

WITH ranked_resources AS (
    SELECT
        id,
        node_pool_id,
        org_unit_data_id,
        updated_at,
        deleted_at,
        ROW_NUMBER() OVER (
            PARTITION BY node_pool_id, org_unit_data_id
            ORDER BY updated_at DESC, id DESC  -- To break ties if updated_at is the same
            ) AS rn
    FROM resources
    WHERE deleted_at IS NULL
)
UPDATE resources AS B
SET deleted_at = NOW(), deleted_by='runai_duplication_migration'
    FROM ranked_resources AS A
WHERE B.id = A.id
  AND A.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS pool_to_org_unit_data ON resources (org_unit_data_id,node_pool_id) WHERE deleted_at IS NULL;
