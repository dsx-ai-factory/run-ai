CREATE TABLE IF NOT EXISTS metric_stores (
    id UUID NOT NULL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    tenant_id INT NOT NULL,
    datasource_id TEXT NOT NULL,
    metric_store_details JSONB NOT NULL
);

-- only one active row per tenant
CREATE UNIQUE INDEX IF NOT EXISTS uniq_metric_stores_tenant_active
    ON metric_stores (tenant_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_metric_stores_tenant_id
    ON metric_stores (tenant_id) WHERE deleted_at IS NULL;
