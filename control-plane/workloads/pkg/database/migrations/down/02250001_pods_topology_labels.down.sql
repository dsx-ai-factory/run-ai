DROP INDEX IF EXISTS idx_pods_network_topology_label_values;

ALTER TABLE pods DROP COLUMN IF EXISTS network_topology_label_values;
