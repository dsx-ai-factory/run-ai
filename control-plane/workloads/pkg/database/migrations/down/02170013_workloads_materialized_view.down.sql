DROP INDEX IF EXISTS idx_view_workloads_department_name;
DROP INDEX IF EXISTS idx_view_workloads_department_id;
DROP INDEX IF EXISTS idx_view_workloads_project;
DROP INDEX IF EXISTS idx_view_workloads_cluster;
DROP INDEX IF EXISTS idx_view_workloads_tenant;
DROP INDEX IF EXISTS idx_view_workloads_submitted_by;
DROP INDEX IF EXISTS idx_view_workloads_deletion_time;
DROP INDEX IF EXISTS idx_view_workloads_creation_time;
DROP INDEX IF EXISTS idx_view_workloads_project_name;
DROP INDEX IF EXISTS idx_view_workloads_name;
DROP INDEX IF EXISTS idx_view_workloads_type;
DROP INDEX IF EXISTS idx_view_workloads_id;

drop MATERIALIZED view if exists workloads_view;