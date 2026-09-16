-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX IF NOT EXISTS idx_workloads_source ON workloads (source);
CREATE INDEX IF NOT EXISTS idx_view_workloads_source ON workloads_materialized_view (source);
CREATE INDEX IF NOT EXISTS idx_view_workloads_partial_source ON workloads_partial_materialized_view (source);
