-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_cpu_quota;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_quota;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_ranks;
ALTER TABLE nodepools DROP COLUMN IF EXISTS allow_adjustable_weights;
