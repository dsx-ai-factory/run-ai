-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE node
    ADD COLUMN IF NOT EXISTS network_topology_label_values TEXT[];
ALTER TABLE node
    ADD COLUMN IF NOT EXISTS network_topology_labels JSONB;
