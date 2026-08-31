-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'containers_to_delete') THEN
        DROP INDEX IF EXISTS idx_containers_unique_cluster_pod_name;
        DROP TABLE IF EXISTS containers_unique;
        DROP TABLE IF EXISTS containers;
        ALTER TABLE containers_to_delete RENAME TO containers;
    ELSE
        DROP INDEX IF EXISTS idx_containers_unique_cluster_pod_name;
        DROP TABLE IF EXISTS containers_unique;
END IF;
END $$;
