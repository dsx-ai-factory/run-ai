-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS swap_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS cpu_swap_memory_size TEXT DEFAULT '100G';
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS reserved_gpu_memory_for_swap_operations TEXT DEFAULT '2G';
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS node_level_scheduler_enabled BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_swap_enabled BOOLEAN DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_cpu_swap_memory_size TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_reserved_gpu_memory_for_swap_operations TEXT DEFAULT NULL;
ALTER TABLE nodepools ADD COLUMN IF NOT EXISTS status_node_level_scheduler_enabled BOOLEAN DEFAULT NULL;

UPDATE nodepools
SET node_level_scheduler_enabled = CASE
    WHEN over_provisioning_ratio > 1 THEN TRUE
    ELSE FALSE
END,
status_node_level_scheduler_enabled = CASE
    WHEN status_over_provisioning_ratio IS NULL THEN NULL
    WHEN status_over_provisioning_ratio > 1 THEN TRUE
    ELSE FALSE
END
WHERE deleted_at IS NULL;
