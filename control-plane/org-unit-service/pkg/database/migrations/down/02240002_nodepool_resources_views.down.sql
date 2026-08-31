-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS project_node_pool_resources_view;
DROP VIEW IF EXISTS department_node_pool_resources_view;

DROP INDEX IF EXISTS resources_priority_idx;
DROP INDEX IF EXISTS resources_node_pool_name_idx;
DROP INDEX IF EXISTS resources_node_pool_id_idx;