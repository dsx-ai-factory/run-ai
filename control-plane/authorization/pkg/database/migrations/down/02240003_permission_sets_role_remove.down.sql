-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS permissions ADD COLUMN IF NOT EXISTS permission_set_id UUID;
ALTER TABLE IF EXISTS permissions ADD CONSTRAINT fk_permission_set_permissions FOREIGN KEY(permission_set_id) REFERENCES permission_sets(id);
