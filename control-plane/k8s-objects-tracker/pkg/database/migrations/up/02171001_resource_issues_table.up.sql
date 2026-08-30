CREATE OR REPLACE FUNCTION delete_pvc_issues()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM resource_issues WHERE type = 'pvc' AND resource_id = OLD.id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION delete_secret_issues()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM resource_issues WHERE type = 'secret' AND resource_id = OLD.id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;



CREATE OR REPLACE FUNCTION delete_configmap_issues()
    RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM resource_issues WHERE type = 'config-map' AND resource_id = OLD.id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;




CREATE TABLE IF NOT EXISTS k8s_objects_tracker.resource_issues (
                                                           id UUID NOT NULL PRIMARY KEY,
                                                           resource_id UUID NOT NULL,
                                                           tenant_id INT NOT NULL,
                                                           cluster_id UUID NOT NULL,
                                                           scope_type TEXT NOT NULL,
                                                           scope_id TEXT NOT NULL,
                                                           type TEXT NOT NULL,
                                                           issue TEXT NOT NULL
    );

CREATE UNIQUE INDEX IF NOT EXISTS resource_issues_idx ON k8s_objects_tracker.resource_issues(tenant_id, cluster_id, resource_id, scope_type, scope_id);
CREATE INDEX IF NOT EXISTS resource_issues_scope_idx ON k8s_objects_tracker.resource_issues(tenant_id, scope_type, scope_id);

-- Creating triggers for automatic deletion of issues related to a resource being deleted
-- Trigger for PVCs deletion
DO $$
BEGIN
    BEGIN
        CREATE TRIGGER delete_pvc_issues_trigger
            AFTER DELETE ON k8s_objects_tracker.pvcs
            FOR EACH ROW
            EXECUTE FUNCTION delete_pvc_issues();
    EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;

-- Trigger for secrets deletion
DO $$
BEGIN
    BEGIN
        CREATE TRIGGER delete_secret_issues_trigger
            AFTER DELETE ON k8s_objects_tracker.secrets
            FOR EACH ROW
            EXECUTE FUNCTION delete_secret_issues();
    EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;

-- Trigger for configmaps deletion
DO $$
BEGIN
    BEGIN
        CREATE TRIGGER delete_configmap_issues_trigger
            AFTER DELETE ON k8s_objects_tracker.configmaps
            FOR EACH ROW
            EXECUTE FUNCTION delete_configmap_issues();
    EXCEPTION
            WHEN duplicate_object THEN NULL;
    END;
END $$;