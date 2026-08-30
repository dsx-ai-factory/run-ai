ALTER TABLE nodepools
    ADD COLUMN IF NOT EXISTS gpu_network_acceleration_detected BOOLEAN NOT NULL DEFAULT FALSE;