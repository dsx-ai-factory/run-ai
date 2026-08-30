-- Remove configuration column from workloads_queue table

ALTER TABLE workloads_queue 
DROP COLUMN IF EXISTS configuration;
