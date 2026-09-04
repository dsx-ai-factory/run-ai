-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS access_rules ADD COLUMN IF NOT EXISTS phase TEXT;
ALTER TABLE IF EXISTS access_rules ADD COLUMN IF NOT EXISTS phase_message TEXT;
ALTER TABLE IF EXISTS access_rules ADD COLUMN IF NOT EXISTS phase_by_cluster JSONB;