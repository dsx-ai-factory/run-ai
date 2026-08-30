
CREATE TABLE IF NOT EXISTS assets_secret_volume (
    asset_id UUID NOT NULL PRIMARY KEY,
    credential_asset_id UUID NOT NULL,
    mount_path TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS assets_secret_volume_pkey ON assets_secret_volume (asset_id);
