ALTER TABLE resources
    ALTER COLUMN gpu_deserved SET NOT NULL,
    ALTER COLUMN gpu_deserved SET DEFAULT 0;
