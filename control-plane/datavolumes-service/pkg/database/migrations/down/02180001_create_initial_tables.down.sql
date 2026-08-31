-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop view if exists datavolumes_view;

drop index if exists idx_tenant_shared_scopes_scope;
drop index if exists idx_tenant_datavolume_shared_scopes;
drop table if exists shared_scopes;

drop index if exists idx_datavolumes_department_project;
drop index if exists idx_datavolumes_tenant_cluster;
drop table if exists datavolumes;