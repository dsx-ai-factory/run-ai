-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP TABLE IF EXISTS pod_elements;

ALTER TABLE workloads DROP COLUMN IF EXISTS structure_elements_total_count;

DROP TABLE IF EXISTS workload_elements;
