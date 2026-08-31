-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

CREATE SEQUENCE IF NOT EXISTS id_sequence START WITH 4500000;

CREATE TABLE IF NOT EXISTS node_type (
    id                                      BIGINT DEFAULT nextval('id_sequence') PRIMARY KEY,
    created_at                              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                              TIMESTAMPTZ,
    deleted_at                              TIMESTAMPTZ,
    created_by                              TEXT,
    updated_by                              TEXT,
    deleted_by                              TEXT,
    cluster_id                              UUID,
    tenant_id                               TEXT NOT NULL,
    name                                    TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS node_type_unique ON node_type (tenant_id, cluster_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS org_unit_data (
    id                                      UUID NOT NULL PRIMARY KEY,
    created_at                              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                              TIMESTAMPTZ,
    deleted_at                              TIMESTAMPTZ,
    created_by                              TEXT,
    updated_by                              TEXT,
    deleted_by                              TEXT,
    interactive_job_time_limit              INTEGER,
    interactive_job_max_idle_duration       INTEGER,
    interactive_job_preempt_idle_duration   INTEGER,
    training_job_time_limit                 INTEGER,
    training_job_max_idle_duration          INTEGER,
    default_node_pool_list                  VARCHAR ARRAY DEFAULT NULL,
    node_type_defined                       BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS org_unit_data_deleted_idx ON org_unit_data(deleted_at);

CREATE TABLE IF NOT EXISTS org_unit_train_node_type (
    org_unit_data_id                        UUID NOT NULL,
    node_type_id                            BIGINT NOT NULL,
    CONSTRAINT fk_node_type FOREIGN KEY(node_type_id) REFERENCES node_type(id),
    CONSTRAINT fk_org_unit_data FOREIGN KEY(org_unit_data_id) REFERENCES org_unit_data(id)
);

CREATE TABLE IF NOT EXISTS org_unit_workspace_node_type (
    org_unit_data_id                        UUID NOT NULL,
    node_type_id                            BIGINT NOT NULL,
    CONSTRAINT fk_node_type FOREIGN KEY(node_type_id) REFERENCES node_type(id),
    CONSTRAINT fk_org_unit_data FOREIGN KEY(org_unit_data_id) REFERENCES org_unit_data(id)
);

CREATE TABLE IF NOT EXISTS department (
    id                  BIGINT DEFAULT nextval('id_sequence') PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    created_by          TEXT,
    updated_by          TEXT,
    deleted_by          TEXT,
    tenant_id           TEXT NOT NULL,
    cluster_id          UUID,
    name                TEXT NOT NULL,
    parent_id           BIGINT,
    org_unit_data_id    UUID,
    CONSTRAINT fk_parent FOREIGN KEY(parent_id) REFERENCES department(id),
    CONSTRAINT fk_org_unit_data FOREIGN KEY(org_unit_data_id) REFERENCES org_unit_data(id)
);
CREATE INDEX IF NOT EXISTS department_name_idx ON department(name);
CREATE INDEX IF NOT EXISTS department_deleted_idx ON department(deleted_at);
CREATE INDEX IF NOT EXISTS department_cluster_idx ON department(tenant_id, cluster_id);
CREATE UNIQUE INDEX IF NOT EXISTS department_name_unique ON department (tenant_id, cluster_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS project (
    id                  BIGINT DEFAULT nextval('id_sequence') PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ,
    deleted_at          TIMESTAMPTZ,
    created_by          TEXT,
    updated_by          TEXT,
    deleted_by          TEXT,
    tenant_id           TEXT NOT NULL,
    cluster_id          UUID,
    name                TEXT NOT NULL,
    namespace           TEXT,
    status              JSONB,
    phase               TEXT,
    parent_id           BIGINT,
    org_unit_data_id    UUID,
    CONSTRAINT fk_parent FOREIGN KEY(parent_id) REFERENCES department(id),
    CONSTRAINT fk_org_unit_data FOREIGN KEY(org_unit_data_id) REFERENCES org_unit_data(id)
);
CREATE INDEX IF NOT EXISTS project_name_idx ON project(name);
CREATE INDEX IF NOT EXISTS project_deleted_idx ON project(deleted_at);
CREATE INDEX IF NOT EXISTS project_cluster_idx ON project(tenant_id, cluster_id);
CREATE UNIQUE INDEX IF NOT EXISTS project_name_unique ON project (tenant_id, cluster_id, name) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS resources (
    id                          UUID NOT NULL PRIMARY KEY,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ,
    deleted_at                  TIMESTAMPTZ,
    created_by                  TEXT,
    updated_by                  TEXT,
    deleted_by                  TEXT,
    node_pool_id                TEXT NOT NULL,
    node_pool_name              TEXT NOT NULL,
    org_unit_data_id            UUID NOT NULL,
    gpu_deserved                DECIMAL(10,2),
    gpu_limit                   DECIMAL(10,2),
    gpu_over_quota_weight       SMALLINT,
    cpu_deserved                INT,
    cpu_limit                   INT,
    cpu_over_quota_weight       SMALLINT,
    memory_deserved             INT,
    memory_limit                INT,
    memory_over_quota_weight    SMALLINT,
    memory_units                TEXT,
    CONSTRAINT fk_org_unit_data FOREIGN KEY(org_unit_data_id) REFERENCES org_unit_data(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS pool_to_org_unit_data ON resources(deleted_at,org_unit_data_id,node_pool_id);
