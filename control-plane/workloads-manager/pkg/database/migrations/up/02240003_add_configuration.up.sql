-- Add configuration column to workloads_queue table
-- This column stores workload configuration as a structured JSONB object
-- Including MNNVL configuration with mode and GPU acceleration label keys

ALTER TABLE workloads_queue
ADD COLUMN IF NOT EXISTS configuration JSONB DEFAULT NULL;

COMMENT ON COLUMN workloads_queue.configuration IS 'Workload configuration to apply on the workload';
