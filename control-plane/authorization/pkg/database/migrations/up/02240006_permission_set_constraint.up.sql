-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS role_to_permission_set ADD CONSTRAINT role_to_permission_sets_key UNIQUE (role_id, permission_set_id);