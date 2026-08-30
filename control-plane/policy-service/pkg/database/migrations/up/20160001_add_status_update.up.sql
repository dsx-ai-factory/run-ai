
CREATE TABLE IF NOT EXISTS sync_status(
                                          id UUID NOT NULL PRIMARY KEY,

                                          tenant_id INT NOT NULL,
                                          cluster_id UUID NOT NULL,

                                          workload_type TEXT NOT NULL,
                                          scope_type TEXT NOT NULL,
                                          scope_id TEXT,

                                          sent_to_agent_apply BIGINT NOT NULL,
                                          sent_to_agent_delete BIGINT NOT NULL,
                                          processed_by_agent BIGINT,
                                          message TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS sync_status_by_policy_cluster ON sync_status(scope_type, scope_id, workload_type, cluster_id);

CREATE INDEX IF NOT EXISTS sync_status_by_cluster ON sync_status(cluster_id);

CREATE INDEX IF NOT EXISTS sync_status_by_tenant ON sync_status(tenant_id);
