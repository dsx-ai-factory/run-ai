create table if not exists metric_stores (
    id serial primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    tenant_id bigint not null unique,
    metric_store jsonb not null
);

create unique index if not exists idx_metric_stores_tenant_id on metric_stores (tenant_id);