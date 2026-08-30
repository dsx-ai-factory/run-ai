create table if not exists logos (
    id serial primary key,
    tenant_id bigint not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    logo_png bytea not null
);

create index if not exists idx_logos_deleted_at on logos (deleted_at);
create index if not exists idx_logos_tenant_id on logos (tenant_id);
create unique index if not exists logos_tenant_id_key on logos (tenant_id);