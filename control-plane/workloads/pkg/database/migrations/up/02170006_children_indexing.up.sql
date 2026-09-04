-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS idx_workloads_children_type ON workloads_children (workload_id, child_type);