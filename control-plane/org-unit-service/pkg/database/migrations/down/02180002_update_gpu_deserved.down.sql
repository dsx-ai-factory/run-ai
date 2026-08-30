ALTER TABLE resources
    ALTER COLUMN gpu_deserved DROP NOT NULL,
    ALTER COLUMN gpu_deserved DROP DEFAULT;
