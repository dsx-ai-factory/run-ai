-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Revert department_id back to UUID.
--
-- Rows that are not cluster-scoped may carry a numeric department_id (an org-unit id),
-- which cannot be cast to UUID. Such rows did not exist in the pre-migration (UUID) schema
-- -- the objects table only tracked cluster-scoped objects -- so drop them before reverting
-- the column type. The remaining cluster-scoped rows have an empty department_id that maps
-- cleanly to NULL.
DELETE FROM k8s_objects_tracker.objects WHERE scope_type <> 'cluster';

-- Cast through ::text so the expression is valid whether the column is currently TEXT or
-- already UUID (keeps the migration safe to run more than once).
ALTER TABLE IF EXISTS k8s_objects_tracker.objects
    ALTER COLUMN department_id DROP DEFAULT,
    ALTER COLUMN department_id TYPE UUID USING NULLIF(department_id::text, '')::uuid;
