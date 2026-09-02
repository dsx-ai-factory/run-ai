-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS cluster_configs (
                                               id UUID PRIMARY KEY,
                                               created_at TIMESTAMPTZ DEFAULT now(),
                                               updated_at TIMESTAMPTZ DEFAULT now(),
                                               deleted_at TIMESTAMPTZ,
                                               tenant_id INTEGER NOT NULL,
                                               cluster_id UUID NOT NULL,
                                               config JSONB NOT NULL DEFAULT '{}' :: JSONB,
                                               status JSONB NOT NULL DEFAULT '{}' :: JSONB,

                                               FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE CASCADE,
                                               UNIQUE (cluster_id, tenant_id)
);

INSERT INTO
    cluster_configs (id, cluster_id, tenant_id)
SELECT
    c.id AS id,
    c.id AS cluster_id,
    c.tenant_id AS tenant_id
FROM
    clusters c
ON CONFLICT DO NOTHING;