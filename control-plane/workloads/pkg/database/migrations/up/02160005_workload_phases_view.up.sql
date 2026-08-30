drop view if exists workloads_phases_view;
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