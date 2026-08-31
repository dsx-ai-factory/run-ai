-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0


CREATE TABLE IF NOT EXISTS crud_templates (
              "id" uuid NOT NULL,

              "created_by" text        NOT NULL,
              "created_at" timestamptz NOT NULL,
              "updated_by" text        NOT NULL,
              "updated_at" timestamptz NOT NULL,

              "workload_type" text NOT NULL,
              "name" text NOT NULL,
              "description" text,

              "tenant_id" int4 NOT NULL,
              "scope_type" text NOT NULL,
              "scope_id" text NOT NULL,

              "cluster_id" uuid,
              "department_id" text,
              "project_id" text,

              "spec" jsonb NOT NULL,
              "master_same_as_worker" bool,
              "master_spec" jsonb,

              "combined_spec" jsonb NOT NULL,
              "combined_master_spec" jsonb NOT NULL,

              "runnable" bool NOT NULL,
              "runnable_issue" text,

              PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS crud_templates_by_name ON workloads.crud_templates USING btree (tenant_id, scope_type, scope_id, name);

CREATE INDEX IF NOT EXISTS crud_templates_by_runnable ON workloads.crud_templates USING btree (tenant_id, runnable, scope_type);
