-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS allow_limits BOOLEAN NOT NULL DEFAULT true;
UPDATE nodepools SET allow_limits = allow_quota;
