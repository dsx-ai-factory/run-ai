-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;

ALTER TABLE workload_types ADD COLUMN IF NOT EXISTS "group" VARCHAR;
DROP INDEX IF EXISTS workload_type_name;
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_name_group ON workload_types(name,"group");

ALTER TABLE workload_type_statuses ADD COLUMN IF NOT EXISTS conditions JSONB;
ALTER TABLE workload_type_statuses ADD COLUMN IF NOT EXISTS version VARCHAR;

-- holds resource_definition for a workload_type per version (and scope_type,tenant_id)
CREATE TABLE workload_type_resource_definitions (
    id                  UUID NOT NULL PRIMARY KEY,
    tenant_id           VARCHAR,
    scope_type          VARCHAR NOT NULL,
    workload_type_id    UUID,
    version             VARCHAR,
    resource_definition JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by          VARCHAR,

    FOREIGN KEY (workload_type_id) REFERENCES workload_types(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_resource_definitions_version ON workload_type_resource_definitions(workload_type_id,version,scope_type,tenant_id);

-- Backfill "group" for existing rows in external_workloads_type_config
WITH mapping AS (
  SELECT name, "group"
  FROM (
    VALUES
      ('Deployment', 'apps'),
      ('SeldonDeployment', 'machinelearning.seldon.io'),
      ('VirtualMachineInstance', 'kubevirt.io'),
      ('TFJob', 'kubeflow.org'),
      ('PyTorchJob', 'kubeflow.org'),
      ('XGBoostJob', 'kubeflow.org'),
      ('JAXJob', 'kubeflow.org'),
      ('MPIJob', 'kubeflow.org'),
      ('Notebook', 'kubeflow.org'),
      ('Job', 'batch'),
      ('StatefulSet', 'apps'),
      ('ReplicaSet', 'apps'),
      ('Pod', ''),
      ('AmlJob', 'amlarc.azureml.com'),
      ('Service', 'serving.knative.dev'),
      ('CronJob', 'batch'),
      ('DevWorkspace', 'workspace.devfile.io'),
      ('RayCluster', 'ray.io'),
      ('RayService', 'ray.io'),
      ('RayJob', 'ray.io'),
      ('TaskRun', 'tekton.dev'),
      ('PipelineRun', 'tekton.dev'),
      ('Workflow', 'argoproj.io'),
      ('ScheduledWorkflow', 'kubeflow.org'),

      ('TrainingWorkload', 'run.ai'),
      ('DistributedWorkload', 'run.ai'),
      ('InteractiveWorkload', 'run.ai'),
      ('InferenceWorkload', 'run.ai'),
      ('DistributedInferenceWorkload', 'run.ai'),

      ('NIMCache', 'apps.nvidia.com'),
      ('NIMService', 'apps.nvidia.com'),
      ('SPOTRequest', 'egx.nvidia.io')
  ) AS u(name, "group")
)
UPDATE workload_types t
SET "group" = m."group"
FROM mapping m
WHERE t.name = m.name;

-- Insert "version" in workload_type_resource_definitions
WITH mapping AS (
  SELECT name, version
  FROM (
    VALUES
      ('Deployment', 'v1'),
      ('SeldonDeployment', 'v1'),
      ('VirtualMachineInstance', 'v1'),
      ('TFJob', 'v1'),
      ('PyTorchJob', 'v1'),
      ('XGBoostJob', 'v1'),
      ('JAXJob', 'v1'),
      ('MPIJob', 'v1'),
      ('Notebook', 'v1beta1'),
      ('Job', 'v1'),
      ('StatefulSet', 'v1'),
      ('ReplicaSet', 'v1'),
      ('Pod', 'v1'),
      ('AmlJob', 'v1alpha1'),
      ('Service', 'v1'),
      ('CronJob', 'v1'),
      ('DevWorkspace', 'v1alpha2'),
      ('RayCluster', 'v1'),
      ('RayService', 'v1'),
      ('RayJob', 'v1'),
      ('TaskRun', 'v1'),
      ('PipelineRun', 'v1'),
      ('Workflow', 'v1alpha1'),
      ('ScheduledWorkflow', 'v1alpha1'),

      ('NIMCache', 'v1alpha1'),
      ('NIMService', 'v1alpha1'),
      ('SPOTRequest', 'v1')
  ) AS u(name, version)
)
INSERT INTO workload_type_resource_definitions (id, scope_type, workload_type_id, version, resource_definition, created_by)
SELECT gen_random_uuid(), 'system', t.id, m.version, '{}'::jsonb, 'Run.ai'
FROM workload_types t
JOIN mapping m ON t.name = m.name
ON CONFLICT (workload_type_id,version,scope_type,tenant_id) DO NOTHING;

CREATE VIEW external_workloads_type_config_view AS
SELECT
    workload_types.id,
    workload_types.name,
    workload_types.group,
    workload_types.source,
    config.created_at,
    config.created_by,
    config.updated_at,
    config.updated_by,
    COALESCE((
        SELECT array_to_json(array_agg(row_to_json(rd) ORDER BY rd.created_at))
        FROM workload_type_resource_definitions rd
        WHERE rd.workload_type_id = workload_types.id
          AND rd.scope_type = config.scope_type
          AND ((rd.tenant_id IS NULL AND config.tenant_id IS NULL) OR rd.tenant_id = config.tenant_id)
    ), '[]'::json) AS resource_definitions,
    COALESCE(category.id, system_category.id) AS category_id,
    COALESCE(priority.id, system_priority.id) AS priority_id,
    COALESCE(category.name, system_category.name) AS category_name,
    COALESCE(priority.name, system_priority.name) AS priority_name,
    config.scope_type,
    config.tenant_id
FROM workload_types
         INNER JOIN external_workloads_type_config config
                    ON config.workload_type_id = workload_types.id
         LEFT JOIN external_workloads_type_config config_system
                   ON config_system.workload_type_id = workload_types.id AND config_system.scope_type = 'system'
         
         LEFT JOIN workloads_categories category
                   ON category.id = config.category_id
         LEFT JOIN priorities priority
                   ON priority.id = config.priority_id
         LEFT JOIN workloads_categories system_category
                   ON system_category.id = config_system.category_id
         LEFT JOIN priorities system_priority
                   ON system_priority.id = config_system.priority_id;