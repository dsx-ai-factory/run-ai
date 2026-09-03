-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE crud_workloads DROP COLUMN IF EXISTS department_id;

DROP INDEX IF EXISTS crud_workloads_by_project;
DROP INDEX IF EXISTS crud_workloads_by_department;
