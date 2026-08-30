ALTER TABLE assets_config_map ADD COLUMN IF NOT EXISTS default_mode text;
ALTER TABLE assets_secret_volume ADD COLUMN IF NOT EXISTS default_mode text;
