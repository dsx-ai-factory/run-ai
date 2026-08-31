-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP INDEX IF EXISTS idx_network_topology_name;

ALTER TABLE nodepools DROP COLUMN network_topology_name;
ALTER TABLE nodepools DROP COLUMN network_topology_id;
