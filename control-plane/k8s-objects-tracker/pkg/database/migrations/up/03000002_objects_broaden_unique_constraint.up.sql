-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- The original single constraint `unique_tenant_type_cluster_name_project` was
-- correct for the cluster-scoped kinds the objects table originally held
-- (StorageClass, DeviceClass) but is too narrow for ResourceClaimTemplate,
-- which legitimately needs the same (type, name) across cluster / department /
-- project scopes.
ALTER TABLE objects
    DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_name_project;

-- Narrow uniqueness for the cluster-scoped kinds. These rows have no project
-- scope - the service code never sets project_id for StorageClass or
-- DeviceClass, and the column defaults to '' - so (tenant, type, cluster,
-- name) is the honest key. The original constraint carried a vestigial
-- project_id in its key that never discriminated anything; dropping it here
-- is a no-op for existing data (every such row has project_id = '') and
-- removes the dead column from the index.
CREATE UNIQUE INDEX IF NOT EXISTS unique_cluster_scoped_objects
    ON objects (tenant_id, type, cluster_id, name)
    WHERE type IN ('StorageClass', 'DeviceClass');

-- Broader uniqueness for ResourceClaimTemplate. Includes scope_type and
-- department_id so a cluster-scoped RCT "foo" and a department-scoped RCT
-- "foo" in the same cluster can coexist without colliding.
CREATE UNIQUE INDEX IF NOT EXISTS unique_resource_claim_templates
    ON objects (tenant_id, cluster_id, scope_type, name, project_id, department_id)
    WHERE type = 'ResourceClaimTemplate';
