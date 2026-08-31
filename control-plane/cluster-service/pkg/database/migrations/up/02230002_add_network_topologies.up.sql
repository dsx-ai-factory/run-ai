-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE TABLE IF NOT EXISTS network_topologies
(
    id         UUID PRIMARY KEY,
    name       VARCHAR       NOT NULL,
    levels     VARCHAR ARRAY NOT NULL,
    source     VARCHAR       NOT NULL,
    cluster_id UUID          NOT NULL,
    tenant_id  INTEGER       NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT now(),
    created_by VARCHAR,
    updated_by VARCHAR,
    phase      VARCHAR,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT nt_cluster_id_fk FOREIGN KEY (cluster_id) REFERENCES clusters (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_nt_tenant_cluster_id ON network_topologies (tenant_id, cluster_id);
CREATE INDEX IF NOT EXISTS idx_nt_source ON network_topologies (source);
CREATE UNIQUE INDEX IF NOT EXISTS nt_unique_name_per_tenant_cluster ON network_topologies (tenant_id, cluster_id, name) WHERE deleted_at IS NULL;
