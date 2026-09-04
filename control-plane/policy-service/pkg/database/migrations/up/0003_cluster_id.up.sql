-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies ADD COLUMN IF NOT EXISTS cluster_id UUID NULL;

CREATE INDEX IF NOT EXISTS by_tenant_policies     ON policies (tenant_id) WHERE scope = 'tenant';
CREATE INDEX IF NOT EXISTS by_department_policies ON policies (tenant_id, department_id) WHERE scope = 'department';
CREATE INDEX IF NOT EXISTS by_project_policies    ON policies (tenant_id, project_id) WHERE scope = 'project';

