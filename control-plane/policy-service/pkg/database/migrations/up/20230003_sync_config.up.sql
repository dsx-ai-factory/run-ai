-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE approved_deletions ADD COLUMN IF NOT EXISTS avoid_deletions_of_obsolete_policies BOOL;
