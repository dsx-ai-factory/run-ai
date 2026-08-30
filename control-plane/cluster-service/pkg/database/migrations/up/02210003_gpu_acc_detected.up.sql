ALTER TABLE nodepools DROP COLUMN IF EXISTS gpu_network_acceleration_detected;
ALTER TABLE nodepools
    ADD COLUMN IF NOT EXISTS status_gpu_network_acceleration_detected BOOLEAN NOT NULL DEFAULT FALSE;