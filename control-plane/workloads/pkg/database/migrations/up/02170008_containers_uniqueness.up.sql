-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE indexname = 'idx_containers_unique_cluster_pod_name'
    ) THEN
        CREATE TABLE IF NOT EXISTS containers_unique AS
        SELECT DISTINCT ON (cluster_id, pod_id, name) *
        FROM containers;

        ALTER TABLE containers_unique
            ADD CONSTRAINT fk_containers_unique_pods
                FOREIGN KEY (pod_id)
                    REFERENCES pods(id);

        CREATE UNIQUE INDEX IF NOT EXISTS idx_containers_unique_cluster_pod_name ON containers_unique (cluster_id, pod_id, name);

        ALTER TABLE IF EXISTS containers RENAME TO containers_to_delete;
        ALTER TABLE IF EXISTS containers_unique RENAME TO containers;

        drop table if exists containers_to_delete;
END IF;
END $$;