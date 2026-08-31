-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE IF EXISTS org_unit_data DROP COLUMN IF EXISTS sync_failed;

DROP VIEW IF EXISTS project_sync_view;
DROP VIEW IF EXISTS department_sync_view;
