ALTER TABLE network_topologies
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE network_topologies
    ADD COLUMN IF NOT EXISTS updated_by VARCHAR;