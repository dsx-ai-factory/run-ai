-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS project_org_unit_data_id_idx ON project (org_unit_data_id);
CREATE INDEX IF NOT EXISTS department_org_unit_data_id_idx ON department (org_unit_data_id);
CREATE INDEX IF NOT EXISTS train_org_unit_data_id_idx ON org_unit_train_node_type (org_unit_data_id);
CREATE INDEX IF NOT EXISTS workspace_org_unit_data_id_idx ON org_unit_workspace_node_type (org_unit_data_id);
