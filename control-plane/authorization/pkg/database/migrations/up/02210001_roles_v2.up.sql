ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS scope_type text not null default 'system';
ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS scope_id text default null;
ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS scope_name text default null;
ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS custom boolean not null default false;
ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS enabled boolean not null default true;
ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS cluster_role_predefined_id bigint default null;

CREATE INDEX IF NOT EXISTS roles_scope_idx ON roles(scope_type, scope_id);

create table if not exists disabled_roles (
                  role_id       bigint not null
                      constraint fk_disabled_roles
                          references roles
                          on delete cascade,
                  scope_type text   not null,
                  scope_id text   not null
);
CREATE INDEX IF NOT EXISTS disabled_roles_role_idx ON disabled_roles(role_id);
CREATE INDEX IF NOT EXISTS disabled_roles_scope_idx ON disabled_roles(scope_type, scope_id);
