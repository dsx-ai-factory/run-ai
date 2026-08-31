-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS min_guaranteed_runtime TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_min_guaranteed_runtime TEXT DEFAULT NULL;
