-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE identity_manager.realms
ADD COLUMN IF NOT EXISTS exclude_groups_from_tokens BOOLEAN DEFAULT FALSE;

