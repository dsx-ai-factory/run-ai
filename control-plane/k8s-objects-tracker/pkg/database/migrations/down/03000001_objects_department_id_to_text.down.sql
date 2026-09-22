-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Revert department_id back to UUID.
--
-- Only rows that carry a department_id (an org-unit id) can hold a value that fails the cast
-- back to UUID -- these are the non-cluster-scoped rows, which did not exist in the
-- pre-migration (UUID) schema (the objects table only tracked cluster-scoped objects, whose
-- department_id is empty). Delete them first; the remaining rows have an empty/NULL
-- department_id that maps cleanly to NULL.
--
-- Compare via ::text so the expression is valid whether the column is currently TEXT or
-- already UUID (keeps the migration safe to run more than once).
DELETE FROM k8s_objects_tracker.objects WHERE department_id::text <> '';

ALTER TABLE IF EXISTS k8s_objects_tracker.objects
    ALTER COLUMN department_id DROP DEFAULT,
    ALTER COLUMN department_id TYPE UUID USING NULLIF(department_id::text, '')::uuid;
