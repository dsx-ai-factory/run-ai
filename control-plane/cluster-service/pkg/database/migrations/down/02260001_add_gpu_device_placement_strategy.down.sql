-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools DROP COLUMN IF EXISTS gpu_device_placement_strategy;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_gpu_device_placement_strategy;
