-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies ADD COLUMN IF NOT EXISTS "imposed_assets" varchar array;

ALTER TABLE effective_policies ADD COLUMN IF NOT EXISTS "imposed_assets" varchar array;

