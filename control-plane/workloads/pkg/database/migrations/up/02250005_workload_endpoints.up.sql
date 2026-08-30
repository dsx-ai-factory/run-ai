-- Add endpoints column to workloads table
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS endpoints JSONB;

-- #################################################################
--    Update the workloads_view
-- #################################################################

DROP MATERIALIZED VIEW IF EXISTS workloads_materialized_view;
DROP VIEW IF EXISTS workloads_view;

CREATE VIEW workloads_view AS (SELECT DISTINCT ON (w.id)
    w.*,
    c.name as category,
    (case w.phase when 'Completed' then w.phase_updated_at else null end) as completed_at,
    -- preemptible is determined by preemptibility column if it exists, otherwise fallback to priority logic
    (CASE
        WHEN w.preemptibility = 'preemptible' THEN TRUE
        WHEN w.preemptibility = 'non-preemptible' THEN FALSE
        ELSE
            CASE
                WHEN w.priority IS NULL THEN NULL
                WHEN w.priority > 100 THEN FALSE
            ELSE TRUE
        END
    END) AS preemptible,
    (case w.phase when 'Pending' then coalesce(w.total_calculated_pending_time_seconds, 0) + EXTRACT(EPOCH FROM (NOW() - w.phase_updated_at))::BIGINT else w.total_calculated_pending_time_seconds end) as total_pending_time_seconds,
    -- this is the new part, calculating total_running_time_seconds if the phase is Running.
    (case w.phase when 'Running' then coalesce(w.total_calculated_running_time_seconds, 0) + EXTRACT(EPOCH FROM (NOW() - w.phase_updated_at))::BIGINT else w.total_calculated_running_time_seconds end) as total_running_time_seconds,

    count(case p.phase when 'Running' then 1 else null end) as number_of_running_pods,
    array_agg(DISTINCT p.current_nodepool) FILTER (WHERE current_nodepool IS NOT NULL) as current_nodepools,
    array_agg(DISTINCT p.node_name) FILTER (WHERE node_name IS NOT NULL) as nodes,
    case WHEN -1 = ANY(array_agg(p.requested_gpu_limit)) THEN -1
    ELSE sum(p.requested_gpu_limit) FILTER (WHERE p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled')) end as pods_requested_gpu_limit,
    case WHEN -1 = ANY(array_agg(p.requested_gpu_request)) THEN -1
    ELSE sum(p.requested_gpu_request) FILTER (WHERE p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled')) end as pods_requested_gpu_request,
    array_agg(p.requested_gpu_memory_limit) FILTER (WHERE p.requested_gpu_memory_limit IS NOT NULL and (p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled'))) as pods_requested_gpu_memory_limit,
    array_agg(p.requested_gpu_memory_request) FILTER (WHERE p.requested_gpu_memory_request IS NOT NULL and (p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled'))) as pods_requested_gpu_memory_request,
    case WHEN -1 = ANY(array_agg(p.requested_cpu_limit)) THEN -1
    ELSE sum(p.requested_cpu_limit) FILTER (WHERE p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled')) end as pods_requested_cpu_limit,
    case WHEN -1 = ANY(array_agg(p.requested_cpu_request)) THEN -1
    ELSE sum(p.requested_cpu_request) FILTER (WHERE p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled')) end as pods_requested_cpu_request,
    array_agg(p.requested_cpu_memory_limit) FILTER (WHERE p.requested_cpu_memory_limit IS NOT NULL and (p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled'))) as pods_requested_cpu_memory_limit,
    array_agg(p.requested_cpu_memory_request) FILTER (WHERE p.requested_cpu_memory_request IS NOT NULL and (p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled'))) as pods_requested_cpu_memory_request,
    array_agg(p.requested_mig_device) FILTER (WHERE p.requested_mig_device IS NOT NULL and (p.phase NOT IN ('Succeeded', 'Failed', 'Unknown', 'ContainerStatusUnknown', 'Error', 'ContainerCannotRun', 'CrashLoopBackOff', 'Completed', 'OOMKilled'))) as pods_requested_mig_device,
    sum(p.allocated_gpu) as allocated_gpu,
    array_agg(p.allocated_gpu_memory) FILTER (WHERE p.allocated_gpu_memory IS NOT NULL) as allocated_gpu_memory,
    sum(p.allocated_cpu) as allocated_cpu,
    array_agg(p.allocated_cpu_memory) FILTER (WHERE p.allocated_cpu_memory IS NOT NULL) as allocated_cpu_memory,
    array_agg(p.allocated_mig_device) FILTER (WHERE p.allocated_mig_device IS NOT NULL) as allocated_mig_devices
                              FROM
                                  workloads as w
                                  LEFT JOIN workloads_categories as c
                              ON c.id = w.category_id
                                  LEFT JOIN pods as p
                                  ON p.workload_id = w.id and p.deleted_at IS NULL
                              GROUP BY w.id, c.name
                                  );

create MATERIALIZED VIEW IF NOT EXISTS workloads_materialized_view AS (SELECT * from workloads_view);

CREATE UNIQUE INDEX IF NOT EXISTS idx_view_workloads_id ON workloads_materialized_view (id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_type ON workloads_materialized_view (type);
CREATE INDEX IF NOT EXISTS idx_view_workloads_name ON workloads_materialized_view (name);
CREATE INDEX IF NOT EXISTS idx_view_workloads_project_name ON workloads_materialized_view (project_name);
CREATE INDEX IF NOT EXISTS idx_view_workloads_creation_time ON workloads_materialized_view (created_at);
CREATE INDEX IF NOT EXISTS idx_view_workloads_deletion_time ON workloads_materialized_view (deleted_at) where deleted_at is not null;
CREATE INDEX IF NOT EXISTS idx_view_workloads_submitted_by ON workloads_materialized_view (submitted_by) where submitted_by <> '';
CREATE INDEX IF NOT EXISTS idx_view_workloads_tenant ON workloads_materialized_view (tenant_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_cluster ON workloads_materialized_view (cluster_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_project ON workloads_materialized_view (project_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_department_id ON workloads_materialized_view (department_id);
CREATE INDEX IF NOT EXISTS idx_view_workloads_department_name ON workloads_materialized_view (department_name);
CREATE INDEX IF NOT EXISTS idx_view_workloads_search_vector_idx ON workloads_materialized_view using gin(search_vector);
