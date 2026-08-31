-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- Delete orphaned nodes where the cluster no longer exists or has been soft-deleted.
-- One time deletion, due do an issue in older versions - that is now fixed -
-- where nodes were not properly deleted when their clusters were soft-deleted.
DELETE FROM node n
WHERE NOT EXISTS (
    SELECT 1
    FROM clusters c
    WHERE c.id = n.cluster_uuid::uuid
        AND c.deleted_at IS NULL
);
