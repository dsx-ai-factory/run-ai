ALTER TABLE network_topologies
    DROP COLUMN IF EXISTS updated_at;
ALTER TABLE network_topologies
    DROP COLUMN IF EXISTS updated_by;