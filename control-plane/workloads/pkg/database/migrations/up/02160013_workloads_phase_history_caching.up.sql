ALTER TABLE workloads_phase_history ADD COLUMN IF NOT EXISTS cluster_id uuid;
ALTER TABLE workloads_phase_history ADD COLUMN IF NOT EXISTS tenant_id bigint;

CREATE INDEX IF NOT EXISTS idx_phase_history_tenant_workload ON workloads_phase_history (tenant_id, workload_id);

CREATE OR REPLACE FUNCTION update_phases_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.phase <> OLD.phase THEN
        NEW.phase_updated_at = NOW();
       -- this is the new part, when inserting I would like to cache the cluster and tenant id's
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


-- Create a function to handle the cache for existing rows
CREATE OR REPLACE FUNCTION phase_history_setup_trigger_for_existing_rows()
RETURNS VOID AS $$
BEGIN
    -- Fetch workload_id, tenant_id, and cluster_id from workloads table
    UPDATE workloads_phase_history AS ph
    SET
        tenant_id = w.tenant_id,
        cluster_id = w.cluster_id
        FROM workloads AS w
        WHERE
            ph.workload_id = w.id
          AND (ph.tenant_id IS NULL OR ph.cluster_id IS NULL);

    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Execute the function to set up triggers for existing rows
SELECT phase_history_setup_trigger_for_existing_rows();