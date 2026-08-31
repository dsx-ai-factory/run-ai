-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS disabled_roles ADD CONSTRAINT unique_scope_disabled_role UNIQUE (scope_id, scope_type, role_id);
