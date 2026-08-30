CREATE TABLE IF NOT EXISTS assets_name_locks(
    id UUID NOT NULL PRIMARY KEY,
    tenant_id INT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    created_by TEXT NOT NULL,
    created_at timestamp NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS asset_name_unique_lock ON assets_name_locks (tenant_id, name, kind);
