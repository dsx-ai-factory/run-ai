-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS by_tenant_policies;
DROP INDEX IF EXISTS by_department_policies;
DROP INDEX IF EXISTS by_project_policies;

ALTER TABLE policies DROP COLUMN IF EXISTS cluster_id;


