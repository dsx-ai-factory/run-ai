ALTER TABLE reconciler_versions
ADD COLUMN IF NOT EXISTS user_applications VARCHAR(10) NOT NULL DEFAULT 'v1';