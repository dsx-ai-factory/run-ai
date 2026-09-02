-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_pods_network_topology_label_values;

ALTER TABLE pods DROP COLUMN IF EXISTS network_topology_label_values;
