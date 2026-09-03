-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS access_rules DROP COLUMN IF EXISTS phase_by_cluster;
ALTER TABLE IF EXISTS access_rules DROP COLUMN IF EXISTS phase_message;
ALTER TABLE IF EXISTS access_rules DROP COLUMN IF EXISTS phase;