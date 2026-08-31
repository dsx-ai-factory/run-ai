-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


ALTER TABLE effective_policies DROP COLUMN IF EXISTS "default_from_order";
ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "default_from_order" jsonb;
