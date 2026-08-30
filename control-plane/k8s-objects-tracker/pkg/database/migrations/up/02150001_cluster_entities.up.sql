
CREATE TABLE IF NOT EXISTS pvcs (
                                    id UUID NOT NULL PRIMARY KEY,
                                    tenant_id INT NOT NULL,
                                    cluster_id UUID NOT NULL,
                                    scope_type TEXT NOT NULL,
                                    scope_id TEXT NOT NULL,
                                    pvc_name TEXT NOT NULL,
                                    UNIQUE (tenant_id, cluster_id, scope_type, scope_id, pvc_name)
);

CREATE INDEX IF NOT EXISTS pvcs_idx ON pvcs(tenant_id, cluster_id, scope_type, scope_id, pvc_name);

CREATE TABLE IF NOT EXISTS secrets (
                                       id UUID NOT NULL PRIMARY KEY,
                                        tenant_id INT NOT NULL,
                                        cluster_id UUID NOT NULL,
                                        scope_type TEXT NOT NULL,
                                        scope_id TEXT NOT NULL,
                                        secret_name TEXT NOT NULL,
                                        secret_type TEXT NOT NULL,
                                        UNIQUE (tenant_id, cluster_id, scope_type, scope_id, secret_name)
);

CREATE INDEX IF NOT EXISTS secrets_idx ON secrets(tenant_id, cluster_id, scope_type, scope_id, secret_name);
