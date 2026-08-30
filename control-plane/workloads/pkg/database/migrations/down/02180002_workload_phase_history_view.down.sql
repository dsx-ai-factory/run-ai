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

CREATE OR REPLACE FUNCTION update_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.phase <> OLD.phase THEN
        NEW.phase_updated_at = NOW();
insert into workloads_phase_history (created_at, tenant_id, cluster_id, workload_id, phase, phase_message) values (NEW.phase_updated_at, NEW.tenant_id, NEW.cluster_id, NEW.id, NEW.phase, NEW.phase_message);
ELSE
        NEW.phase_updated_at = OLD.phase_updated_at;
END IF;
    IF OLD is null OR NEW.k8s_phase <> OLD.k8s_phase THEN
        NEW.k8s_phase_updated_at = NOW();
ELSE
        NEW.k8s_phase_updated_at = OLD.k8s_phase_updated_at;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE workloads_phase_history DROP COLUMN IF EXISTS old_phase;
