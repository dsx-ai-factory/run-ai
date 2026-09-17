-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Restores the column, but not its values: every role comes back as 'ALL'. Rolling
-- back further, into 03000001, therefore cannot rely on role_type to tell a DGXC role
-- from its twin - that migration's down identifies roles by name instead.

ALTER TABLE IF EXISTS roles ADD COLUMN IF NOT EXISTS role_type varchar default 'ALL';
