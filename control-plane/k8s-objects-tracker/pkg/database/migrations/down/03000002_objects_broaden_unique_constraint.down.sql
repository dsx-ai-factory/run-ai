-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Reverse the per-type split: drop the two partial unique indexes and restore
-- the single constraint that was in effect before 03000002.

DROP INDEX IF EXISTS unique_resource_claim_templates;
DROP INDEX IF EXISTS unique_cluster_scoped_objects;

ALTER TABLE objects
    ADD CONSTRAINT unique_tenant_type_cluster_name_project
    UNIQUE (tenant_id, type, cluster_id, name, project_id);
