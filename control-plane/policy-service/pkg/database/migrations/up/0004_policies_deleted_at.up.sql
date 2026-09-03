-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS deleted_by text NULL;

DROP INDEX IF EXISTS by_name;
CREATE UNIQUE INDEX IF NOT EXISTS by_name ON policies (tenant_id, workload_type, name, deleted_at);
