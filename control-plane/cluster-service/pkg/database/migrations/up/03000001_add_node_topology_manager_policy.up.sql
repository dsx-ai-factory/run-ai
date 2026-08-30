ALTER TABLE node
    ADD COLUMN IF NOT EXISTS topology_manager_policy TEXT;
