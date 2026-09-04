-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS  k8s_objects_tracker.objects DROP CONSTRAINT IF EXISTS unique_tenant_type_cluster_name_project;
DROP TABLE IF EXISTS  k8s_objects_tracker.objects;
