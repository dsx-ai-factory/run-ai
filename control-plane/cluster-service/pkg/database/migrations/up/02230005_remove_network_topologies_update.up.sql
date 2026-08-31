-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE network_topologies
    DROP COLUMN IF EXISTS updated_at;
ALTER TABLE network_topologies
    DROP COLUMN IF EXISTS updated_by;