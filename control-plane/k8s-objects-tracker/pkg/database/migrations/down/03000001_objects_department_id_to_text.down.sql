-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Revert department_id back to UUID. The objects table holds only cluster-scoped
-- StorageClasses (empty department), so map empty strings to NULL on the way back.
-- Cast through ::text so the expression is valid whether the column is currently TEXT
-- or already UUID (keeps the migration safe to run more than once).
ALTER TABLE IF EXISTS k8s_objects_tracker.objects
    ALTER COLUMN department_id DROP DEFAULT,
    ALTER COLUMN department_id TYPE UUID USING NULLIF(department_id::text, '')::uuid;
