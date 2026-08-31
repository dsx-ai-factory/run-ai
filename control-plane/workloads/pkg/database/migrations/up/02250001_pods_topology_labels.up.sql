-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE pods ADD COLUMN IF NOT EXISTS network_topology_label_values text[];

CREATE INDEX IF NOT EXISTS idx_pods_network_topology_label_values ON pods USING btree(network_topology_label_values);
