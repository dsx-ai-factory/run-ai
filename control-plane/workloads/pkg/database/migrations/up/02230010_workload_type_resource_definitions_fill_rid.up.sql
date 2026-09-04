-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

DROP VIEW IF EXISTS external_workloads_type_config_view;

DROP INDEX IF EXISTS workload_type_resource_definitions_version;
ALTER TABLE IF EXISTS workload_type_resource_definitions RENAME TO workload_type_resource_interfaces;
ALTER TABLE workload_type_resource_interfaces RENAME COLUMN resource_definition TO resource_interface;
CREATE UNIQUE INDEX IF NOT EXISTS workload_type_resource_interfaces_version ON workload_type_resource_interfaces(workload_type_id,version,scope_type,tenant_id);

-- Backfill "resource_interface" for existing rows in workload_type_resource_interfaces
WITH mapping AS (
    SELECT name, "group", version
    FROM (
             VALUES
                 ('Deployment', 'apps', 'v1'),
                 ('SeldonDeployment', 'machinelearning.seldon.io', 'v1'),
                 ('VirtualMachineInstance', 'kubevirt.io', 'v1'),
                 ('TFJob', 'kubeflow.org', 'v1'),
                 ('PyTorchJob', 'kubeflow.org', 'v1'),
                 ('XGBoostJob', 'kubeflow.org', 'v1'),
                 ('JAXJob', 'kubeflow.org', 'v1'),
                 ('MPIJob', 'kubeflow.org', 'v1'),
                 ('Notebook', 'kubeflow.org', 'v1beta1'),
                 ('Job', 'batch', 'v1'),
                 ('StatefulSet', 'apps', 'v1'),
                 ('ReplicaSet', 'apps', 'v1'),
                 ('Pod', '', 'v1'),
                 ('AmlJob', 'amlarc.azureml.com', 'v1alpha1'),
                 ('Service', 'serving.knative.dev', 'v1'),
                 ('CronJob', 'batch', 'v1'),
                 ('DevWorkspace', 'workspace.devfile.io', 'v1alpha2'),
                 ('RayCluster', 'ray.io', 'v1'),
                 ('RayService', 'ray.io', 'v1'),
                 ('RayJob', 'ray.io', 'v1'),
                 ('TaskRun', 'tekton.dev', 'v1'),
                 ('PipelineRun', 'tekton.dev', 'v1'),
                 ('Workflow', 'argoproj.io', 'v1alpha1'),
                 ('ScheduledWorkflow', 'kubeflow.org', 'v1alpha1'),

                 ('NIMCache', 'apps.nvidia.com', 'v1alpha1'),
                 ('NIMService', 'apps.nvidia.com', 'v1alpha1'),
                 ('SPOTRequest', 'egx.nvidia.io', 'v1')
         ) AS u(name, "group", version)
)
UPDATE workload_type_resource_interfaces r
SET resource_interface = jsonb_build_object(
    'spec', jsonb_build_object(
        'structureDefinition', jsonb_build_object(
            'rootComponent', jsonb_build_object(
                'kind', jsonb_build_object(
                    'group', m."group",
                    'kind', m.name,
                    'version', m.version
                )
            )
        )
    )
)
FROM mapping m, workload_types
WHERE workload_types.id = r.workload_type_id
  AND workload_types.name = m.name
  AND r.version = m.version;

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
                 FROM workload_type_resource_interfaces rd
                 WHERE rd.workload_type_id = workload_types.id
                   AND rd.scope_type = config.scope_type
                   AND ((rd.tenant_id IS NULL AND config.tenant_id IS NULL) OR rd.tenant_id = config.tenant_id)
             ), '[]'::json) AS resource_interfaces,
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