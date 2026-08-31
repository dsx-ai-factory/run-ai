-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

drop view if exists workloads_view;
drop table if exists events;
drop table if exists containers;
drop index if exists idx_pods_workload;
drop index if exists idx_pods_project;
drop index if exists idx_pods_cluster;
drop index if exists idx_pods_tenant;
drop table if exists pods;
drop table if exists workloads_children;
drop table if exists workloads_status_history;
drop index if exists idx_workloads_project;
drop index if exists idx_workloads_cluster;
drop index if exists idx_workloads_tenant;
drop table if exists workloads;