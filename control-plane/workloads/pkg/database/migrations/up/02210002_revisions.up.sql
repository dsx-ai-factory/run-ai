-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create table if not exists revisions (
    created_at   timestamp with time zone not null default now(),
    deleted_at    timestamp with time zone,
    id uuid   primary key,
    name   varchar not null,
    type   varchar not null,
    workload_id   uuid not null,
    tenant_id   bigint not null,
    cluster_id   uuid not null,
    department_id   bigint not null,
    project_id   bigint not null,
    requested_gpu_limit   NUMERIC(1000, 2),
    requested_gpu_request   NUMERIC(1000, 2),
    requested_gpu_memory_limit   varchar,
    requested_gpu_memory_request   varchar,
    requested_cpu_limit   NUMERIC(1000, 2),
    requested_cpu_request   NUMERIC(1000, 2),
    requested_cpu_memory_limit   varchar,
    requested_cpu_memory_request   varchar,
    requested_mig_device   varchar array,
    phase   varchar not null,
    phase_message   varchar not null default '',
    phase_updated_at   timestamp with time zone not null,
    requested_pods_number   bigint,
    requested_pods_min   bigint,
    requested_pods_max   bigint,
    requested_pods_parallelism   bigint,
    requested_pods_completions   bigint,
    requested_nodepools   varchar array,
    images   varchar array,
    command   varchar,
    arguments   varchar,
    environment_variables   jsonb default null,
    conditions   jsonb default null,
    additional_fields   jsonb not null default '{}'::jsonb
);

create index if not exists idx_revisions_workload ON revisions (workload_id);

create table if not exists revisions_children (
    revision_id   uuid,
    foreign key (revision_id) REFERENCES revisions(id) ON DELETE CASCADE,
    child_type   text,
    child_id   uuid not null,
    CONSTRAINT revision_children_unique_constraint UNIQUE (revision_id, child_id)
);
