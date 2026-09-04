-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE policies DROP COLUMN IF EXISTS "imposed_assets";

ALTER TABLE effective_policies DROP COLUMN IF EXISTS "imposed_assets";

