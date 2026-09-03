-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DELETE FROM identity_manager.realm_settings
WHERE setting_key = 'excludeGroupsFromToken';

ALTER TABLE identity_manager.realms
DROP COLUMN IF EXISTS exclude_groups_from_tokens; 