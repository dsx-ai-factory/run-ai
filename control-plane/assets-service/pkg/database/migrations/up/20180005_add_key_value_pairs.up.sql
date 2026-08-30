ALTER TABLE assets_generic_secret ADD COLUMN if not exists key_value_pairs jsonb;
ALTER TABLE assets_generic_secret ADD COLUMN if not exists key_id int4;