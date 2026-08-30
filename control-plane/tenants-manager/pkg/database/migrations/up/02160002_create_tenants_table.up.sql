create table if not exists tenants (
    id serial         primary key,
    created_at        timestamp not null,
    updated_at        timestamp not null,
    deleted_at        timestamp,
    uuid              text not null,
    mimir_id          text,
    name              text not null,
    display_name      text not null,
    status            text not null,
    contract_type     text not null,
    eula_signing_user text
);


do $$
begin
    if not exists (select * from tenants where id > 1000) then
        perform setval('tenants_id_seq', 1000);
    end if;
end $$;

alter table tenants add column if not exists mimir_id text;

create index if not exists idx_tenants_deleted_at on tenants (deleted_at);
create index if not exists tenants_name_key on tenants (name);
create index if not exists tenants_uuid_key on tenants (uuid);

create unique index if not exists idx_tenants_uuid on tenants (uuid);
create unique index if not exists idx_tenants_mimir_id on tenants (mimir_id);
