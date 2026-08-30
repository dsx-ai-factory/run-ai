DROP TABLE IF EXISTS workload_types;
CREATE TABLE IF NOT EXISTS workload_types (
    id              UUID NOT NULL PRIMARY KEY,
    name            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by      VARCHAR,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      VARCHAR
);
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_name ON workload_types(name);

INSERT INTO workload_types (id, name)
VALUES (gen_random_uuid(),'Deployment'),
       (gen_random_uuid(), 'SeldonDeployment'),
       (gen_random_uuid(), 'VirtualMachineInstance'),
       (gen_random_uuid(), 'TFJob'),
       (gen_random_uuid(), 'PyTorchJob'),
       (gen_random_uuid(), 'XGBoostJob'),
       (gen_random_uuid(), 'MPIJob'),
       (gen_random_uuid(), 'Notebook'),
       (gen_random_uuid(), 'Job'),
       (gen_random_uuid(), 'StatefulSet'),
       (gen_random_uuid(), 'ReplicaSet'),
       (gen_random_uuid(), 'Pod'),
       (gen_random_uuid(), 'AmlJob'),
       (gen_random_uuid(), 'Service'),
       (gen_random_uuid(), 'CronJob'),
       (gen_random_uuid(), 'DevWorkspace'),
       (gen_random_uuid(), 'RayCluster'),
       (gen_random_uuid(), 'RayService'),
       (gen_random_uuid(), 'RayJob'),
       (gen_random_uuid(), 'TaskRun'),
       (gen_random_uuid(), 'PipelineRun'),
       (gen_random_uuid(), 'Workflow'),
       (gen_random_uuid(), 'ScheduledWorkflow'),
       (gen_random_uuid(), 'JAXJob');

DROP VIEW IF EXISTS external_workloads_type_config_view;
DROP INDEX IF EXISTS idx_external_workloads_type_config_tenant_type;
DROP TABLE IF EXISTS external_workloads_type_config;
CREATE TABLE external_workloads_type_config (
    id                  UUID NOT NULL PRIMARY KEY,
    tenant_id           VARCHAR,
    scope_type          VARCHAR NOT NULL,
    workload_type_id    UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by          VARCHAR,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          VARCHAR,
    category_id         UUID,
    priority_id         UUID,
    FOREIGN KEY (workload_type_id) REFERENCES workload_types(id),
    FOREIGN KEY (category_id) REFERENCES workloads_categories(id),
    FOREIGN KEY (priority_id) REFERENCES priorities(id)
);

INSERT INTO external_workloads_type_config(id, scope_type, workload_type_id, category_id, priority_id)
SELECT gen_random_uuid(),
       'system',
       workload_types.id,
       CASE
           WHEN workload_types.name IN
                ('Deployment','SeldonDeployment','StatefulSet','ReplicaSet','Pod','Service','CronJob','RayService','PipelineRun','Workflow','ScheduledWorkflow')
               THEN (SELECT id FROM workloads_categories where name='Deploy')
           WHEN workload_types.name IN
                ('TaskRun','DevWorkspace','Job','Notebook')
               THEN (SELECT id FROM workloads_categories where name='Build')
           ELSE (SELECT id FROM workloads_categories where name='Train')
           END,
       CASE
           WHEN workload_types.name IN
                ('Deployment','SeldonDeployment','StatefulSet','ReplicaSet','Pod','Service','CronJob','RayService','PipelineRun','Workflow','ScheduledWorkflow')
               THEN (SELECT id FROM priorities where name='very-high')
           WHEN workload_types.name IN
                ('TaskRun','DevWorkspace','Job','Notebook')
               THEN (SELECT id FROM priorities where name='high')
           ELSE (SELECT id FROM priorities where name='low')
           END
FROM workload_types;

CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.created_at,
    workload_types.created_by,
    workload_types.updated_at,
    workload_types.updated_by,
    COALESCE(category_tenant.id, category_system.id) AS category_id,
    COALESCE(priority_tenant.id, priority_system.id) AS priority_id,
    COALESCE(category_tenant.name, category_system.name) AS category_name,
    COALESCE(priority_tenant.name, priority_system.name) AS priority_name,
    COALESCE(config_tenant.scope_type, config_system.scope_type) AS scope_type,
    COALESCE(config_tenant.tenant_id, config_system.tenant_id) AS tenant_id
FROM workload_types
         JOIN external_workloads_type_config config_tenant
              ON config_tenant.workload_type_id = workload_types.id
         JOIN external_workloads_type_config config_system
              ON config_system.workload_type_id = workload_types.id AND config_system.scope_type = 'system'
         LEFT JOIN workloads_categories category_tenant
             ON category_tenant.id = config_tenant.category_id
         LEFT JOIN priorities priority_tenant
             ON priority_tenant.id = config_tenant.priority_id
         LEFT JOIN workloads_categories category_system
             ON category_system.id = config_system.category_id
         LEFT JOIN priorities priority_system
             ON priority_system.id = config_system.priority_id;