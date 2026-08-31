-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM datavolumes WHERE project_id IS NULL) THEN
        -- should we raise exception?
        -- RAISE EXCEPTION 'Rollback failed: NULL project_id values exist';
        UPDATE datavolumes
        SET project_id = 'default-project-id'
        WHERE project_id IS NULL;
    END IF;

    ALTER TABLE datavolumes
        ALTER COLUMN project_id SET NOT NULL;
END$$;


ALTER TABLE datavolumes
    ALTER COLUMN project_id SET NOT NULL;

ALTER TABLE datavolumes
    DROP COLUMN IF EXISTS scope;

ALTER TABLE datavolumes
    DROP COLUMN IF EXISTS internal;

ALTER TABLE datavolumes
    DROP COLUMN IF EXISTS department_name;

DROP INDEX IF EXISTS datavolume_pvc_project_scope_unique;

DROP INDEX IF EXISTS datavolume_pvc_cluster_scope_unique;

DROP INDEX IF EXISTS datavolume_pvc_department_scope_unique;

ALTER TABLE datavolumes
    ADD CONSTRAINT datavolume_pvc_unique_constraint
        UNIQUE (cluster_id, project_id, origin_pvc_name);


DROP VIEW if exists datavolumes_view;
CREATE VIEW datavolumes_view AS (SELECT DISTINCT ON (d.id, s.scope_id, s.scope_type)
    d.id,
    d.created_at,
    d.updated_at,
    d.created_by,
    d.updated_by,
    d.name,
    d.description,
    d.project_id,
    d.origin_pvc_name,
    d.project_name,
    d.department_id,
    d.tenant_id,
    d.cluster_id,
    d.phase,
    d.phase_message,
    d.conditions,
    d.datavolume_pvc_name,
    d.datavolume_pv_name,
    d.should_delete_original_volume,
    d.search_vector,
    s.scope_id,
    s.scope_type,
    s.scope_name,
    s.department_id as shared_scope_department_id,
    s.project_id    as shared_scope_project_id,
    s.created_at    as shared_scope_created_at,
    s.created_by    as shared_scope_created_by

    FROM datavolumes as d
        LEFT JOIN shared_scopes as s ON d.id = s.datavolume_id
);
