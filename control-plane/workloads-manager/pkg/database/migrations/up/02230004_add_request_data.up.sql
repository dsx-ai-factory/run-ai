ALTER TABLE workloads_queue ADD COLUMN IF NOT EXISTS input_request_path TEXT;
ALTER TABLE workloads_queue ADD COLUMN IF NOT EXISTS input_request_body JSONB;
