-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


CREATE INDEX IF NOT EXISTS workloads_by_name ON workloads (tenant_id, name);
CREATE INDEX IF NOT EXISTS workloads_by_type ON workloads (tenant_id, workload_type);