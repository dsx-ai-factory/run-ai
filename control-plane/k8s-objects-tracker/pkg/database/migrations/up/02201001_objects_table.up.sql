CREATE TABLE IF NOT EXISTS k8s_objects_tracker.objects (
       -- Meta data columns
       id            UUID DEFAULT gen_random_uuid() PRIMARY KEY,
       type          TEXT NOT NULL,
       name          TEXT NOT NULL,
       scope_type    TEXT NOT NULL,
       project_id    TEXT DEFAULT '',
       department_id UUID NULL,
       cluster_id    UUID NOT NULL,
       tenant_id     INT NOT NULL,
       spec          jsonb NOT NULL,
       CONSTRAINT unique_tenant_type_cluster_name_project UNIQUE (tenant_id, type, cluster_id, name, project_id)
);
