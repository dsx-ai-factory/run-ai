-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

ALTER TABLE applications ADD COLUMN IF NOT EXISTS actual_phase VARCHAR;
ALTER TABLE applications ADD COLUMN IF NOT EXISTS message VARCHAR;

-- Dedicated sync table for control-plane-initiated changes.
-- Only writes to this table (create / delete / update spec) bump update_count
-- via the trigger below, so agent status reports to the applications table
-- never cause the agent to re-pull its own changes.
CREATE SEQUENCE IF NOT EXISTS application_crud_sync_sequence;

CREATE TABLE IF NOT EXISTS application_crud_sync (
    application_id UUID PRIMARY KEY,
    tenant_id      VARCHAR NOT NULL,
    cluster_id     UUID    NOT NULL,
    spec           JSONB,
    desired_phase  VARCHAR,
    update_count   INT8        NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS application_crud_sync_for_sync
    ON application_crud_sync (tenant_id, cluster_id, update_count);

CREATE OR REPLACE FUNCTION application_crud_sync_inc_update_count()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.update_count := nextval('application_crud_sync_sequence');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
    BEGIN
        BEGIN
            CREATE TRIGGER application_crud_sync_update_count_trigger
                BEFORE INSERT OR UPDATE ON application_crud_sync
                FOR EACH ROW
            EXECUTE PROCEDURE application_crud_sync_inc_update_count();
        EXCEPTION
            WHEN duplicate_object THEN NULL;
        END;
    END $$;

DROP VIEW IF EXISTS applications_view;
CREATE VIEW applications_view AS (SELECT DISTINCT ON (a.id)
                                      a.*,
                                      cs.spec,
                                      cs.desired_phase,
                                      cs.update_count,
                                      sum(p.requested_gpu_limit) as requested_gpu_limit,
                                      sum(p.requested_gpu_request) as requested_gpu_request,
                                      array_agg(p.requested_gpu_memory_limit) FILTER (WHERE p.requested_gpu_memory_limit IS NOT NULL) as requested_gpu_memory_limit,
                                      array_agg(p.requested_gpu_memory_request) FILTER (WHERE p.requested_gpu_memory_request IS NOT NULL) as requested_gpu_memory_request,
                                      sum(p.requested_cpu_limit) as requested_cpu_limit,
                                      sum(p.requested_cpu_request) as requested_cpu_request,
                                      array_agg(p.requested_cpu_memory_limit) FILTER (WHERE p.requested_cpu_memory_limit IS NOT NULL) as requested_cpu_memory_limit,
                                      array_agg(p.requested_cpu_memory_request) FILTER (WHERE p.requested_cpu_memory_request IS NOT NULL) as requested_cpu_memory_request,
                                      sum(p.allocated_gpu) as allocated_gpu,
                                      array_agg(p.allocated_gpu_memory) FILTER (WHERE p.allocated_gpu_memory IS NOT NULL) as allocated_gpu_memory,
                                      sum(p.allocated_cpu) as allocated_cpu,
                                      array_agg(p.allocated_cpu_memory) FILTER (WHERE p.allocated_cpu_memory IS NOT NULL) as allocated_cpu_memory
                                 FROM applications as a
                                          LEFT JOIN application_crud_sync as cs
                                                    ON a.id = cs.application_id
                                          LEFT JOIN workloads as w
                                                    ON a.id = w.application_id and w.deleted_at is null
                                          LEFT JOIN pods as p
                                                    ON p.workload_id = w.id and p.deleted_at IS NULL
                                 GROUP BY a.id, cs.spec, cs.desired_phase, cs.update_count
                                     );
