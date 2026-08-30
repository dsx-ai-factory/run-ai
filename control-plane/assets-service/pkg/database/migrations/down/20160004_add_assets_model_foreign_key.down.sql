-- Remove foreign key constraint for environment_asset_id
ALTER TABLE assets_model DROP CONSTRAINT IF EXISTS assets_model_environment_asset_fk;

-- Remove foreign key constraint for compute_asset_id
ALTER TABLE assets_model DROP CONSTRAINT IF EXISTS assets_model_compute_asset_fk;
