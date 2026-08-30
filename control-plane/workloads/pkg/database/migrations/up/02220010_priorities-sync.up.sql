DO $$
    BEGIN
        IF EXISTS
            ( SELECT 1
              FROM   information_schema.tables
              WHERE  table_schema = 'workloads'
                AND  table_name = 'priorities'
            )
        THEN
            UPDATE priorities
            SET source = 'Run.ai'
            WHERE source = 'Run:ai';
        END IF ;
    END
$$ ;

DROP INDEX IF EXISTS priorities_status_unique;
CREATE UNIQUE INDEX priorities_status_unique ON priority_statuses (priority_id, cluster_id);
