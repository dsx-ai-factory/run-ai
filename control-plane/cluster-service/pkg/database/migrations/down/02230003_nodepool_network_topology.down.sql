DROP INDEX IF EXISTS idx_network_topology_name;

ALTER TABLE nodepools DROP COLUMN network_topology_name;
ALTER TABLE nodepools DROP COLUMN network_topology_id;
