ALTER TABLE datavolumes
    ALTER COLUMN project_id DROP NOT NULL;

ALTER TABLE datavolumes
    ADD COLUMN IF NOT EXISTS scope VARCHAR;

ALTER TABLE datavolumes
    ADD COLUMN IF NOT EXISTS internal BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE datavolumes
    ADD COLUMN IF NOT EXISTS department_name VARCHAR;

ALTER TABLE datavolumes
    DROP CONSTRAINT IF EXISTS datavolume_pvc_unique_constraint;

CREATE UNIQUE INDEX IF NOT EXISTS datavolume_pvc_project_scope_unique
    ON datavolumes (cluster_id, project_id, origin_pvc_name)
    WHERE scope IS DISTINCT FROM 'department'
        AND scope IS DISTINCT FROM 'cluster';

CREATE UNIQUE INDEX IF NOT EXISTS datavolume_pvc_cluster_scope_unique
    ON datavolumes (cluster_id, origin_pvc_name)
    WHERE scope = 'cluster';

CREATE UNIQUE INDEX IF NOT EXISTS datavolume_pvc_department_scope_unique
    ON datavolumes (cluster_id, department_id, origin_pvc_name)
    WHERE scope = 'department';

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
    d.internal,
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
