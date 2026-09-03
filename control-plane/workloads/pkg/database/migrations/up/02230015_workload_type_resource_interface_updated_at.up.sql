-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workload_type_resource_interfaces ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

UPDATE workload_type_resource_interfaces SET updated_at = created_at WHERE updated_at IS NULL;