-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

-- we used this sequence to mark every update on the table with increasing number
-- which will be used as the timestamp for the sync.
CREATE SEQUENCE IF NOT EXISTS crud_workloads_sequence;

CREATE TABLE IF NOT EXISTS crud_workloads (
        "id" uuid NOT NULL,

        "created_by" text        NOT NULL,
        "created_at" timestamptz NOT NULL,
        "updated_by" text        NOT NULL,
        "updated_at" timestamptz NOT NULL,
        "deleted_by" text,
        "deleted_at" timestamptz,

        "update_count" int8 NOT NULL,

        "tenant_id" int4 NOT NULL,
        "cluster_id" uuid NOT NULL,
        "project_id" text NOT NULL,

        "workload_type" text NOT NULL,

        "requested_name" text NOT NULL,
        "name" text NOT NULL,
        "namespace" text NOT NULL,

        "spec" jsonb NOT NULL,
        "master_spec" jsonb,

        "desired_phase" text NOT NULL,

        PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS crud_workloads_for_sync ON crud_workloads (tenant_id, cluster_id, update_count);

CREATE OR REPLACE FUNCTION crud_workloads_inc_update_count()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.update_count := nextval('crud_workloads_sequence');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
    BEGIN
        BEGIN
            CREATE TRIGGER crud_workloads_update_count_trigger BEFORE INSERT OR UPDATE ON crud_workloads
                FOR EACH ROW
            EXECUTE PROCEDURE crud_workloads_inc_update_count();
        EXCEPTION
            WHEN duplicate_object THEN NULL;
        END;
    END $$;
