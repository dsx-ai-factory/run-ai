-- before returning uniqueness to the index, prevent duplicate policies
-- by adding suffix to them
WITH
    duplicates AS (
        SELECT
            id,
            tenant_id,
            workload_type,
            name,
            ROW_NUMBER() OVER (PARTITION BY tenant_id, workload_type, name) AS seq_num
        FROM
            policies
    ),
    to_update AS (
        SELECT
            id,
            CONCAT(name, '-', seq_num) AS new_name
        FROM
            duplicates
        WHERE
            seq_num > 1 -- only rows that would cause duplicates
    )
UPDATE policies p
SET
    name = to_update.new_name
FROM
	to_update
WHERE
    p.id = to_update.id;

DROP INDEX IF EXISTS by_name;
CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies(tenant_id, workload_type, name);
