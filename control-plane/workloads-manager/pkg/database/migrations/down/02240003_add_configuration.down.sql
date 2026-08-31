-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Remove configuration column from workloads_queue table

ALTER TABLE workloads_queue 
DROP COLUMN IF EXISTS configuration;
