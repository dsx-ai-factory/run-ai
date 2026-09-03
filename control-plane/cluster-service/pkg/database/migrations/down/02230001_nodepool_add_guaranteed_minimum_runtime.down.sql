-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools DROP COLUMN IF EXISTS status_min_guaranteed_runtime;
ALTER TABLE nodepools DROP COLUMN IF EXISTS min_guaranteed_runtime;
