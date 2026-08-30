ALTER TABLE workload_types
    ADD COLUMN IF NOT EXISTS type VARCHAR;

INSERT INTO workload_types (id, name, type)
VALUES (gen_random_uuid(), 'TrainingWorkload', 'native'),
       (gen_random_uuid(), 'DistributedWorkload', 'native'),
       (gen_random_uuid(), 'InteractiveWorkload', 'native'),
       (gen_random_uuid(), 'DistributedInferenceWorkload', 'native'),
       (gen_random_uuid(), 'InferenceWorkload', 'native');

UPDATE workload_types
SET type = 'external'
WHERE name IN (
               'Deployment',
               'SeldonDeployment',
               'VirtualMachineInstance',
               'TFJob',
               'PyTorchJob',
               'XGBoostJob',
               'MPIJob',
               'Notebook',
               'Job',
               'StatefulSet',
               'ReplicaSet',
               'Pod',
               'AmlJob',
               'Service',
               'CronJob',
               'DevWorkspace',
               'RayCluster',
               'RayService',
               'RayJob',
               'TaskRun',
               'PipelineRun',
               'Workflow',
               'ScheduledWorkflow',
               'JAXJob'
    );

CREATE UNIQUE INDEX IF NOT EXISTS external_workload_type_config_unique_per_tenant ON external_workloads_type_config (workload_type_id, tenant_id);