-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The original single constraint `unique_tenant_type_cluster_name_project` was
-- correct for the cluster-scoped kinds the objects table originally held
-- (StorageClass, DeviceClass) but is too narrow for ResourceClaimTemplate,
-- which legitimately needs the same (type, name) across cluster / department /
-- project scopes.
ALTER TABLE objects
    DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_name_project;

-- Narrow uniqueness for the cluster-scoped kinds. These never had multiple
-- scopes so (tenant, type, cluster, name, project) is the right identity.
-- Project_id is kept in the key because the original constraint had it and
-- removing it would subtly change behaviour for any legacy project-scoped row.
CREATE UNIQUE INDEX unique_cluster_scoped_objects
    ON objects (tenant_id, type, cluster_id, name, project_id)
    WHERE type IN ('StorageClass', 'DeviceClass');

-- Broader uniqueness for ResourceClaimTemplate. Includes scope_type and
-- department_id so a cluster-scoped RCT "foo" and a department-scoped RCT
-- "foo" in the same cluster can coexist without colliding.
CREATE UNIQUE INDEX unique_resource_claim_templates
    ON objects (tenant_id, cluster_id, scope_type, name, project_id, department_id)
    WHERE type = 'ResourceClaimTemplate';
