drop view if exists events_of_workload_view;
drop view if exists timeout_events_of_workload_view;
create view timeout_events_of_workload_view AS (
     SELECT DISTINCT ON (e.id)
         wc.workload_id as workload_id,
         w.project_name,
         w.type as workload_type,
         w.submitted_by,
         e.id as event_id,
         e.created_at,
         e.cluster_id,
         e.tenant_id,
         e.type,
         e.message,
         e.reason,
         e.source,
         e.involved_object_uid,
         e.involved_object_kind,
         e.involved_object_name,
         e.involved_object_namespace
     FROM
         workloads_children as wc
             LEFT JOIN events as e
                ON wc.child_id = e.involved_object_uid
             LEFT JOIN workloads as w
                ON wc.workload_id = w.id
     WHERE reason = 'WorkloadTimeoutApproaching'
         );