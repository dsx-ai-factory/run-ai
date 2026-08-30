CREATE TABLE IF NOT EXISTS assets_ngc_api_key (
    asset_id UUID NOT NULL PRIMARY KEY,
    ngc_api_key TEXT NOT NULL,
    key_id INTEGER
); 