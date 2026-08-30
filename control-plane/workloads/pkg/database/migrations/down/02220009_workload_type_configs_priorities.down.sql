DROP VIEW IF EXISTS external_workloads_type_config_view;
DROP TABLE IF EXISTS external_workloads_type_config;
DROP INDEX IF EXISTS workload_type_name;
DROP TABLE IF EXISTS workload_types;

CREATE TABLE IF NOT EXISTS external_workloads_type_config (
      id uuid NOT NULL PRIMARY KEY,
      tenant_id bigint NOT NULL DEFAULT -1,
      type varchar NOT NULL,
      category_id uuid NOT NULL,
      scope_type varchar NOT NULL DEFAULT 'tenant',
      updated_at timestamp with time zone DEFAULT now(),
      updated_by varchar,
      FOREIGN KEY (category_id) REFERENCES workloads_categories(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_external_workloads_type_config_tenant_type ON external_workloads_type_config (tenant_id, type);

-- Insert default workload type to category mappings
WITH category_ids AS (
    SELECT id, name
    FROM workloads_categories
    WHERE scope_type = 'system'
)
INSERT INTO external_workloads_type_config (id, tenant_id, type, category_id, scope_type, updated_at, updated_by)
SELECT
    gen_random_uuid(),
    -1,
    type,
    c.id,
    'system',
    now(),
    'system'
FROM (
         VALUES
             ('Deployment', 'Deploy'),
             ('SeldonDeployment', 'Deploy'),
             ('VirtualMachineInstance', 'Train'),
             ('TFJob', 'Train'),
             ('PyTorchJob', 'Train'),
             ('XGBoostJob', 'Train'),
             ('MPIJob', 'Train'),
             ('Notebook', 'Build'),
             ('Job', 'Build'),
             ('StatefulSet', 'Deploy'),
             ('ReplicaSet', 'Deploy'),
             ('Pod', 'Deploy'),
             ('AmlJob', 'Train'),
             ('Service', 'Deploy'),
             ('CronJob', 'Deploy'),
             ('DevWorkspace', 'Build'),
             ('RayCluster', 'Train'),
             ('RayService', 'Deploy'),
             ('RayJob', 'Train'),
             ('TaskRun', 'Build'),
             ('PipelineRun', 'Deploy'),
             ('Workflow', 'Deploy'),
             ('ScheduledWorkflow', 'Deploy')
     ) AS mappings(type, category_name)
         JOIN category_ids c ON c.name = mappings.category_name
ON CONFLICT (tenant_id, type)
    DO UPDATE SET
                  category_id = EXCLUDED.category_id,
                  updated_at = now(),
                  updated_by = 'system';

CREATE VIEW external_workloads_type_config_view
AS (
   SELECT
       t.*,
       c.name as category_name
   FROM
       external_workloads_type_config as t
           JOIN workloads_categories as c ON c.id = t.category_id
       );
