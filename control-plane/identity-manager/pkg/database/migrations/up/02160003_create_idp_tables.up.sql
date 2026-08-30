CREATE TABLE IF NOT EXISTS idps (
    id                  UUID NOT NULL PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ,
    created_by          TEXT,
    updated_by          TEXT,
    tenant_id           TEXT NOT NULL,
    alias               TEXT NOT NULL,
    resource_name       TEXT NOT NULL,
    resource_content    TEXT NOT NULL,
    resource_type    TEXT NOT NULL  
);

CREATE INDEX IF NOT EXISTS tenant_id_alias_idx ON idps (tenant_id, alias);