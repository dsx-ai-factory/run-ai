-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies ADD COLUMN IF NOT EXISTS gvk_workload BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS validation_issues JSONB;
