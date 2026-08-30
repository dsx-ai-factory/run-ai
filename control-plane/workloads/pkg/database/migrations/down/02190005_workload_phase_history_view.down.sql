CREATE OR REPLACE FUNCTION update_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.phase <> OLD.phase THEN
        NEW.phase_updated_at = NOW();
       -- this is the new part, when inserting I am adding the old_phase column
insert into workloads_phase_history (created_at, tenant_id, cluster_id, workload_id, phase, phase_message, old_phase) values (NEW.phase_updated_at, NEW.tenant_id, NEW.cluster_id, NEW.id, NEW.phase, NEW.phase_message, OLD.phase);
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