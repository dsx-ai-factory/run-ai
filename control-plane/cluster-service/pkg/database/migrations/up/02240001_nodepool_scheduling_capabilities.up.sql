-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS allow_adjustable_weights BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS allow_ranks BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS allow_quota BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS allow_cpu_quota BOOLEAN NOT NULL DEFAULT false;