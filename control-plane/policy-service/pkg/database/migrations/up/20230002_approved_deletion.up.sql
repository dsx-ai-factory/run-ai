-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- list of clusters for which deletion of policies has been approved
CREATE TABLE IF NOT EXISTS approved_deletions (
    tenant_id integer NOT NULL,
    cluster_id UUID NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS by_cluster ON approved_deletions (tenant_id, cluster_id);

-- we no longer need the sync status information. we still use the table
-- in the transition from sync to non-sync, but the info is not needed
DELETE FROM sync_status;