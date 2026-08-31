-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE workloads ADD COLUMN IF NOT EXISTS dist_specific_env jsonb;
ALTER TABLE workloads_assets ADD COLUMN IF NOT EXISTS dist_replica_type text;