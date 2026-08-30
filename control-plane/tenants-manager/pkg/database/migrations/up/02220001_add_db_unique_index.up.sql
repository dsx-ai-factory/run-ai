DROP INDEX IF EXISTS tenants_name_key;
create unique index if not exists tenants_name_key on tenants (name);