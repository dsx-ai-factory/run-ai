-- add new table assets_model with fields:
--  id, title, description, thumbnail, huggingface_id, compute_asset_id, environment_asset_id
create table if not exists assets_model (
    asset_id uuid primary key,
    title TEXT not null,
    description TEXT not null,
    thumbnail bytea not null,
    huggingface_id TEXT not null,
    compute_asset_id uuid not null,
    environment_asset_id uuid not null
);