-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_parent_id_not_deleted
    ON project(parent_id)
    WHERE deleted_at IS NULL;
