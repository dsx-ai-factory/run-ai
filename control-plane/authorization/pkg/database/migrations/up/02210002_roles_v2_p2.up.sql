-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS tenant_id text;
CREATE INDEX IF NOT EXISTS roles_tenant_idx ON roles(tenant_id);

