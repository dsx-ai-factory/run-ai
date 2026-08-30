CREATE TABLE IF NOT EXISTS mapping_locking (
    tenant_id   integer      NOT NULL PRIMARY KEY,
    locked_at   timestamptz  NOT NULL
);
