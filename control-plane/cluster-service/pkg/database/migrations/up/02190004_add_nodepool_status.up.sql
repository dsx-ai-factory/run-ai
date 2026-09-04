-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_label_key TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_label_value TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_over_provisioning_ratio INTEGER DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_gpu_placement_strategy TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_cpu_placement_strategy TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_nodes VARCHAR ARRAY DEFAULT NULL;