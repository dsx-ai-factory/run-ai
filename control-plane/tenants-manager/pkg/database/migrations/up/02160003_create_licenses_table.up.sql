-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create table if not exists licenses (
    id serial primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    tenant_id bigint not null,
    license text not null
);

create index if not exists idx_licenses_deleted_at on licenses (deleted_at);
create index if not exists idx_licenses_tenant_id on licenses (tenant_id);
create unique index if not exists licenses_tenant_id_key on licenses (tenant_id);




