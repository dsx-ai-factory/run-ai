create table if not exists datavolumes (
    id            uuid primary key,
    created_at    timestamp with time zone not null default now(),
    updated_at    timestamp with time zone not null default now(),
    created_by    varchar not null default '',
    updated_by    varchar not null default '',
    name          varchar not null,
    description   varchar not null default '',
    project_id    varchar not null,
    origin_pvc_name varchar not null,
    project_name  varchar not null,
    department_id varchar not null,
    tenant_id     bigint not null,
    cluster_id    uuid not null,
    phase         varchar not null default 'Unknown',
    phase_message varchar not null default '',
    conditions    jsonb default null,
    datavolume_pvc_name varchar not null default '',
    datavolume_pv_name varchar not null default '',
    should_delete_original_volume boolean not null default false,
    CONSTRAINT datavolume_tenant_unique_constraint UNIQUE (tenant_id, name),
    CONSTRAINT datavolume_pvc_unique_constraint UNIQUE (cluster_id, project_id, origin_pvc_name)
);

CREATE INDEX IF NOT EXISTS idx_datavolumes_tenant_cluster ON datavolumes (tenant_id, cluster_id);
CREATE INDEX IF NOT EXISTS idx_datavolumes_department_project ON datavolumes (department_id, project_id);

create table if not exists shared_scopes (
    datavolume_id uuid,
    foreign key (datavolume_id) REFERENCES datavolumes(id) ON DELETE CASCADE,
    scope_id   varchar   not null,
    scope_type varchar   not null,
    CONSTRAINT datavolume_shared_scopes_unique_constraint UNIQUE (datavolume_id, scope_id, scope_type),
    scope_name   varchar   not null,
    tenant_id    bigint not null,
    created_at   timestamp with time zone not null default now(),
    created_by   varchar not null default '',

    project_id   varchar not null,
    project_name  varchar not null,
    department_id varchar not null
);


CREATE INDEX IF NOT EXISTS idx_tenant_datavolume_shared_scopes ON shared_scopes (tenant_id, datavolume_id);
CREATE INDEX IF NOT EXISTS idx_tenant_shared_scopes_scope ON shared_scopes (tenant_id, scope_id, scope_type);

drop view if exists datavolumes_view;
create view datavolumes_view AS (SELECT DISTINCT ON (d.id, s.scope_id, s.scope_type)
    d.id,
    d.created_at,
    d.updated_at,
    d.created_by,
    d.updated_by,
    d.name,
    d.description,
    d.project_id,
    d.origin_pvc_name,
    d.project_name,
    d.department_id,
    d.tenant_id,
    d.cluster_id,
    d.phase,
    d.phase_message,
    d.conditions,
    d.datavolume_pvc_name,
    d.datavolume_pv_name,
    d.should_delete_original_volume,
    s.scope_id,
    s.scope_type,
    s.scope_name,
    s.department_id as shared_scope_department_id,
    s.project_id    as shared_scope_project_id,
    s.project_name  as shared_scope_project_name,
    s.created_at    as shared_scope_created_at,
    s.created_by    as shared_scope_created_by

    FROM datavolumes as d
        LEFT JOIN shared_scopes as s ON d.id = s.datavolume_id
);