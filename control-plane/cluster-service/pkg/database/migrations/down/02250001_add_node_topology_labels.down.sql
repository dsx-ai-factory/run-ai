-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE node
    DROP COLUMN IF EXISTS network_topology_label_values;
ALTER TABLE node
    DROP COLUMN IF EXISTS network_topology_labels;