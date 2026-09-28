-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- department_id was mistakenly defined as UUID. Org-unit ids (project and department) are
-- numeric identifiers and are stored as TEXT across the platform -- see project_id in this
-- same table. Align department_id with project_id: a TEXT column holding the numeric id.
--
-- The objects table currently holds only cluster-scoped StorageClasses, which have no
-- department. Those rows carry the zero UUID (00000000-0000-0000-0000-000000000000),
-- written by the Go model's uuid.UUID zero value -- so instead of casting that junk value
-- to a literal string, USING '' normalizes every existing row to an empty string.
ALTER TABLE IF EXISTS k8s_objects_tracker.objects
    ALTER COLUMN department_id TYPE TEXT USING '',
    ALTER COLUMN department_id SET DEFAULT '';
