-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create table if not exists roles (
    created_at    timestamp with time zone,
    updated_at    timestamp with time zone,
    created_by    text,
    id            bigserial primary key,
    name          text not null unique,
    description   text not null,
    internal_only boolean default false
);

create table if not exists permissions (
    created_at    timestamp with time zone,
    updated_at    timestamp with time zone,
    role_id       bigint not null
        constraint fk_roles_permissions
            references roles
            on delete cascade,
    resource_type text   not null,
    "create"      boolean,
    read          boolean,
    update        boolean,
    delete        boolean,
    sync          boolean,
    primary key (role_id, resource_type)
);

create index if not exists idx_permissions_resource_type
    on permissions (resource_type);

create table if not exists access_rules (
    created_at   timestamp with time zone,
    updated_at   timestamp with time zone,
    created_by   text,
    id           bigserial
        primary key,
    subject_id   text   not null,
    subject_type text   not null,
    role_id      bigint not null
        constraint fk_roles_access_rules
            references roles
            on delete cascade,
    scope_id     text   not null,
    scope_type   text   not null,
    scope_name   text   not null,
    tenant_id    bigint,
    cluster_uuid text   not null,
    deleted_at   timestamp with time zone
);

create index if not exists idx_access_rules_deleted_at
    on access_rules (deleted_at);

create index if not exists idx_access_rules_cluster_uuid
    on access_rules (cluster_uuid);

create index if not exists idx_access_rules_tenant_id
    on access_rules (tenant_id);

create unique index if not exists idx_subject_role_resource
    on access_rules (subject_id, role_id, scope_id, scope_type)
    where (deleted_at IS NULL);

