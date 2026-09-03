-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE nodepools DROP COLUMN IF EXISTS status_node_level_scheduler_enabled;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_reserved_gpu_memory_for_swap_operations;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_cpu_swap_memory_size;
ALTER TABLE nodepools DROP COLUMN IF EXISTS status_swap_enabled;

ALTER TABLE nodepools DROP COLUMN IF EXISTS node_level_scheduler_enabled;
ALTER TABLE nodepools DROP COLUMN IF EXISTS reserved_gpu_memory_for_swap_operations;
ALTER TABLE nodepools DROP COLUMN IF EXISTS cpu_swap_memory_size;
ALTER TABLE nodepools DROP COLUMN IF EXISTS swap_enabled;