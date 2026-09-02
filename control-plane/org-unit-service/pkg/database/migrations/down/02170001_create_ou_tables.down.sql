-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS pool_to_org_unit_data;
DROP TABLE IF EXISTS resources;
DROP INDEX IF EXISTS project_name_unique;
DROP INDEX IF EXISTS project_cluster_idx;
DROP INDEX IF EXISTS project_deleted_idx;
DROP INDEX IF EXISTS project_name_idx;
DROP TABLE IF EXISTS project;
DROP INDEX IF EXISTS department_name_unique;
DROP INDEX IF EXISTS department_cluster_idx;
DROP INDEX IF EXISTS department_deleted_idx;
DROP INDEX IF EXISTS department_name_idx;
DROP TABLE IF EXISTS department;
DROP TABLE IF EXISTS org_unit_workspace_node_type;
DROP TABLE IF EXISTS org_unit_train_node_type;
DROP INDEX IF EXISTS org_unit_data_deleted_idx;
DROP TABLE IF EXISTS org_unit_data;
DROP INDEX IF EXISTS node_type_unique;
DROP TABLE IF EXISTS node_type;
DROP SEQUENCE IF EXISTS id_sequence;