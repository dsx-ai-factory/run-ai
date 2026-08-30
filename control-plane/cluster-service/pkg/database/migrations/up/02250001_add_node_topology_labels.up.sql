ALTER TABLE node
    ADD COLUMN IF NOT EXISTS network_topology_label_values TEXT[];
ALTER TABLE node
    ADD COLUMN IF NOT EXISTS network_topology_labels JSONB;
