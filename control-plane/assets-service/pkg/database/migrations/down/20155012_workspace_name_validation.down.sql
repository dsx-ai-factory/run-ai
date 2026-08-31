-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

---- Redo of 0012_workspace_name_validation.down.sql in idempotent in case it didn't run
DROP TRIGGER IF EXISTS validate_workload_name ON workloads;
