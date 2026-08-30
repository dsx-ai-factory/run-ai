DROP INDEX IF EXISTS idx_events_tenant_cluster_obj;
DROP INDEX IF EXISTS idx_events_create_time;
DROP INDEX IF EXISTS idx_phase_create_time;
DROP INDEX IF EXISTS idx_phase_workload;

drop view if exists workloads_phases_view;
drop view if exists workloads_events_view;
create view workloads_events_view AS (
    SELECT DISTINCT ON (e.id)
        w.id as workload_id,
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
      workloads as w
    INNER JOIN  workloads_children  wc
        ON      wc.workload_id = w.id
    INNER JOIN  events  e
        ON      wc.child_id = e.involved_object_uid OR w.id = e.involved_object_uid
);
create view workloads_phases_view AS (
    SELECT
        w.id as workload_id,
        w.cluster_id,
        w.tenant_id,
        wph.created_at,
        wph.phase,
        wph.phase_message
    FROM
      workloads as w
    INNER JOIN  workloads_phase_history  wph
        ON      wph.workload_id = w.id
);