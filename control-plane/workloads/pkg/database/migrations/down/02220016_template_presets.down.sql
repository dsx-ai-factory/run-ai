-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


-- Return the index to its original structure. the up script reversed the order ot the tenant id and scope type
DROP INDEX IF EXISTS crud_templates_by_name;
CREATE UNIQUE INDEX IF NOT EXISTS crud_templates_by_name ON workloads.crud_templates USING btree (tenant_id, scope_type, scope_id, name);

DROP TABLE IF EXISTS crud_templates_presets;
