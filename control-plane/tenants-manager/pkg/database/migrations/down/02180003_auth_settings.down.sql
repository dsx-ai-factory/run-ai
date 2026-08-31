-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

create table if not exists auth_settings (
                                             id serial primary key,
                                             tenant_id bigint not null,
                                             created_at timestamp not null,
                                             updated_at timestamp not null,
                                             deleted_at timestamp,
                                             client_id text not null,
                                             client_secret text not null
);

create index if not exists idx_auth_setting_deleted_at on auth_settings (deleted_at);
create index if not exists idx_auth_setting_tenant_id on auth_settings (tenant_id);
create unique index if not exists auth_setting_tenant_id_key on auth_settings (tenant_id);
--
