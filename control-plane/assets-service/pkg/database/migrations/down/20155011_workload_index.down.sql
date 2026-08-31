-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

---- Redo of 0011_workload_index.down.sql in idempotent in case it didn't run
DROP INDEX IF EXISTS workloads_by_name;
DROP INDEX IF EXISTS workloads_by_type;