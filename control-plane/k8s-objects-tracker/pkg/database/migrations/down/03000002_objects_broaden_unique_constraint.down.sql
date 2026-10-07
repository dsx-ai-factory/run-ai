-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE k8s_objects_tracker.objects
    DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_scope_name_project_department;

ALTER TABLE k8s_objects_tracker.objects
    ADD CONSTRAINT unique_tenant_type_cluster_name_project
    UNIQUE (tenant_id, type, cluster_id, name, project_id);
