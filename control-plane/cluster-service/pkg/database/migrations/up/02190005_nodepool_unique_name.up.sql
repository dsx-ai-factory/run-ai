-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE UNIQUE INDEX IF NOT EXISTS np_unique_name ON nodepools (cluster_id, name) WHERE deleted_at IS NULL;
