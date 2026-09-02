-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- we do not want this index to be unique because it prevents two clusters to have
-- the same policy name for no reason. it may look ugly in the policies grid, but
-- it has no impact on functionality so its up to the user if it is ok for him to
-- have multiple names or not.
DROP INDEX IF EXISTS by_name;
CREATE INDEX IF NOT EXISTS by_name ON policies(tenant_id, workload_type, name);
