-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_cluster;
DROP CONSTRAINT IF EXISTS unique_node_name_per_cluster;
DROP TABLE IF EXISTS node;