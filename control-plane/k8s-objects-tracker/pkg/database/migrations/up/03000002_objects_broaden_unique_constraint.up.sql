-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The original constraint (tenant_id, type, cluster_id, name, project_id) was
-- designed when the objects table only held cluster-scoped kinds (StorageClass,
-- DeviceClass). It cannot represent two rows that share (tenant, type, cluster,
-- name) but live in different scopes - e.g. a cluster-scoped ResourceClaimTemplate
-- named "foo" and a department-scoped one also named "foo" would both have an
-- empty project_id and collide.
--
-- Extend the key with scope_type and department_id so each (type, name) can
-- coexist across cluster / department / project scopes. Existing rows are
-- unaffected: all current rows are cluster-scoped (project_id='', department_id='',
-- scope_type='cluster'), so the broader key still uniquely identifies them.

ALTER TABLE k8s_objects_tracker.objects
    DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_name_project;

ALTER TABLE k8s_objects_tracker.objects
    ADD CONSTRAINT unique_tenant_type_cluster_scope_name_project_department
    UNIQUE (tenant_id, type, cluster_id, scope_type, name, project_id, department_id);
