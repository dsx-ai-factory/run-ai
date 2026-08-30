-- Replace sync views with optimized SQL functions
--
-- The project_sync_view and department_sync_view virtual views aggregate the
-- entire resources table across ALL tenants before filtering by tenant_id/cluster_id.
-- These functions filter by tenant/cluster FIRST, then only aggregate resources
-- for the matching org_unit_data_ids.
--
-- The views are dropped and replaced with custom composite types that match
-- the Go structs exactly (no unused quota columns).

DROP FUNCTION IF EXISTS get_sync_projects(TEXT, UUID);
DROP FUNCTION IF EXISTS get_sync_departments(TEXT, UUID);
DROP TYPE IF EXISTS project_sync_type CASCADE;
DROP TYPE IF EXISTS department_sync_type CASCADE;
-- Drop old views
DROP VIEW IF EXISTS project_sync_view CASCADE;
DROP VIEW IF EXISTS department_sync_view CASCADE;

-- Create return types matching the Go structs exactly:
--   ProjectSyncView = IntBaseModel + project fields + OrgUnitDataSyncView + calculated_updated_at
--   DepartmentSyncView = IntBaseModel + department fields + resources_json + calculated_updated_at
CREATE TYPE project_sync_type AS (
    -- IntBaseModel
    id                                    BIGINT,
    created_at                            TIMESTAMPTZ,
    updated_at                            TIMESTAMPTZ,
    deleted_at                            TIMESTAMPTZ,
    created_by                            TEXT,
    updated_by                            TEXT,
    deleted_by                            TEXT,
    -- Project fields
    tenant_id                             TEXT,
    cluster_id                            UUID,
    name                                  TEXT,
    requested_namespace                   TEXT,
    enforce_runai_scheduler               BOOLEAN,
    parent_id                             BIGINT,
    parent_name                           TEXT,
    deletion_request                      TIMESTAMPTZ,
    deletion_type                         TEXT,
    force_delete                          BOOLEAN,
    org_unit_data_id                      UUID,
    -- ProjectSyncView only uses status_last_updated_time from ProjectStatus
    status_last_updated_time              TIMESTAMPTZ,
    -- OrgUnitDataSyncView fields
    interactive_job_time_limit            INT,
    interactive_job_max_idle_duration     INT,
    interactive_job_preempt_idle_duration INT,
    training_job_time_limit               INT,
    training_job_max_idle_duration        INT,
    default_node_pool_list                TEXT[],
    resources_json                        JSON,
    training_node_types_view              JSON,
    workspace_node_types_view             JSON,
    -- Calculated
    calculated_updated_at                 TIMESTAMPTZ
);

CREATE TYPE department_sync_type AS (
    -- IntBaseModel
    id                                    BIGINT,
    created_at                            TIMESTAMPTZ,
    updated_at                            TIMESTAMPTZ,
    deleted_at                            TIMESTAMPTZ,
    created_by                            TEXT,
    updated_by                            TEXT,
    deleted_by                            TEXT,
    -- Department fields
    tenant_id                             TEXT,
    cluster_id                            UUID,
    name                                  TEXT,
    parent_id                             BIGINT,
    parent_name                           TEXT,
    org_unit_data_id                      UUID,
    status_last_updated_time              TIMESTAMPTZ,
    -- Sync fields
    resources_json                        JSON,
    -- Calculated
    calculated_updated_at                 TIMESTAMPTZ
);


CREATE OR REPLACE FUNCTION get_sync_projects(p_tenant_id TEXT, p_cluster_id UUID)
RETURNS SETOF project_sync_type AS $$
WITH matching_projects AS (
    SELECT project.org_unit_data_id AS oud_id,
           org_unit_data.training_node_types_effective_origin,
           org_unit_data.workspace_node_types_effective_origin
    FROM project
    INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
    WHERE project.tenant_id = p_tenant_id AND project.cluster_id = p_cluster_id
),
resources_combined AS (
    SELECT org_unit_data_id,
           MAX(resources.updated_at) FILTER (WHERE resources.deleted_at IS NULL) AS resources_updated_at,
           json_agg(resources.* ORDER BY resources.id) FILTER (WHERE resources.deleted_at IS NULL) AS resources_json,
           MAX(resources.deleted_at) AS resources_deleted_at
    FROM resources
    WHERE org_unit_data_id IN (SELECT oud_id FROM matching_projects)
    GROUP BY org_unit_data_id
),
train_node_types_effective AS (
    SELECT json_agg(node_type.*) AS training_node_types_view_effective,
           org_unit_train_node_type.org_unit_data_id,
           MAX(org_unit_train_node_type.created_at) AS train_node_type_created_at
    FROM org_unit_train_node_type
    LEFT JOIN node_type ON node_type.id = org_unit_train_node_type.node_type_id
    WHERE org_unit_train_node_type.org_unit_data_id IN (
        SELECT COALESCE(training_node_types_effective_origin, oud_id) FROM matching_projects
    )
    GROUP BY org_unit_train_node_type.org_unit_data_id
),
workspace_node_types_effective AS (
    SELECT json_agg(node_type.*) AS workspace_node_types_view_effective,
           org_unit_workspace_node_type.org_unit_data_id,
           MAX(org_unit_workspace_node_type.created_at) AS workspace_node_type_created_at
    FROM org_unit_workspace_node_type
    LEFT JOIN node_type ON node_type.id = org_unit_workspace_node_type.node_type_id
    WHERE org_unit_workspace_node_type.org_unit_data_id IN (
        SELECT COALESCE(workspace_node_types_effective_origin, oud_id) FROM matching_projects
    )
    GROUP BY org_unit_workspace_node_type.org_unit_data_id
)
SELECT project.id,
       project.created_at,
       project.updated_at,
       project.deleted_at,
       project.created_by,
       project.updated_by,
       project.deleted_by,
       project.tenant_id,
       project.cluster_id,
       project.name,
       project.requested_namespace,
       project.enforce_runai_scheduler,
       project.parent_id,
       project.parent_name,
       project.deletion_request,
       project.deletion_type,
       project.force_delete,
       project.org_unit_data_id,
       project.status_last_updated_time,
       interactive_job_time_limit_effective AS interactive_job_time_limit,
       interactive_job_max_idle_duration_effective AS interactive_job_max_idle_duration,
       interactive_job_preempt_idle_duration_effective AS interactive_job_preempt_idle_duration,
       training_job_time_limit_effective AS training_job_time_limit,
       training_job_max_idle_duration_effective AS training_job_max_idle_duration,
       default_node_pool_list_effective AS default_node_pool_list,
       rc.resources_json,
       tnte.training_node_types_view_effective AS training_node_types_view,
       wnte.workspace_node_types_view_effective AS workspace_node_types_view,
       GREATEST(
               COALESCE(project.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(rc.resources_updated_at, '1970-01-01')
           , COALESCE(rc.resources_deleted_at, '1970-01-01')
           , COALESCE(tnte.train_node_type_created_at, '1970-01-01')
           , COALESCE(wnte.workspace_node_type_created_at, '1970-01-01')
       ) AS calculated_updated_at
FROM project
INNER JOIN org_unit_data ON project.org_unit_data_id = org_unit_data.id
LEFT JOIN resources_combined rc ON rc.org_unit_data_id = org_unit_data.id
LEFT JOIN train_node_types_effective tnte ON tnte.org_unit_data_id = COALESCE(org_unit_data.training_node_types_effective_origin, org_unit_data.id)
LEFT JOIN workspace_node_types_effective wnte ON wnte.org_unit_data_id = COALESCE(org_unit_data.workspace_node_types_effective_origin, org_unit_data.id)
WHERE project.tenant_id = p_tenant_id AND project.cluster_id = p_cluster_id;
$$ LANGUAGE SQL STABLE;


CREATE OR REPLACE FUNCTION get_sync_departments(p_tenant_id TEXT, p_cluster_id UUID)
RETURNS SETOF department_sync_type AS $$
WITH matching_departments AS (
    SELECT department.org_unit_data_id AS oud_id
    FROM department
    WHERE department.tenant_id = p_tenant_id AND department.cluster_id = p_cluster_id
),
resources_combined AS (
    SELECT org_unit_data_id,
           MAX(resources.updated_at) FILTER (WHERE resources.deleted_at IS NULL) AS resources_updated_at,
           json_agg(resources.* ORDER BY resources.id) FILTER (WHERE resources.deleted_at IS NULL) AS resources_json,
           MAX(resources.deleted_at) AS resources_deleted_at
    FROM resources
    WHERE org_unit_data_id IN (SELECT oud_id FROM matching_departments)
    GROUP BY org_unit_data_id
)
SELECT department.id,
       department.created_at,
       department.updated_at,
       department.deleted_at,
       department.created_by,
       department.updated_by,
       department.deleted_by,
       department.tenant_id,
       department.cluster_id,
       department.name,
       department.parent_id,
       NULL::TEXT AS parent_name,
       department.org_unit_data_id,
       department.status_last_updated_time,
       rc.resources_json,
       GREATEST(
               COALESCE(department.updated_at, '1970-01-01')
           , COALESCE(org_unit_data.updated_at, '1970-01-01')
           , COALESCE(rc.resources_deleted_at, '1970-01-01')
           , COALESCE(rc.resources_updated_at, '1970-01-01')
       ) AS calculated_updated_at
FROM department
INNER JOIN org_unit_data ON department.org_unit_data_id = org_unit_data.id
LEFT JOIN resources_combined rc ON rc.org_unit_data_id = org_unit_data.id
WHERE department.tenant_id = p_tenant_id AND department.cluster_id = p_cluster_id;
$$ LANGUAGE SQL STABLE;
