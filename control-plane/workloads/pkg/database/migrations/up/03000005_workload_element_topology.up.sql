-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Requested constraints only; actual placement is computed at query time. NULL for
-- frameworks that do not declare topology per component.
ALTER TABLE workload_elements ADD COLUMN IF NOT EXISTS topology JSONB;
