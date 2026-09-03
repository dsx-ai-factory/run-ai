-- SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
-- SPDX-License-Identifier: Apache-2.0

INSERT INTO workload_types (id, name, "group", source, created_by)
VALUES (gen_random_uuid(), 'InferenceService', 'serving.kserve.io', 'custom', 'Run.ai')
    ON CONFLICT (name, "group") DO NOTHING;

INSERT INTO external_workloads_type_config(id, scope_type, workload_type_id, category_id, priority_id, created_by, updated_by)
SELECT gen_random_uuid(),
       'system',
       workload_types.id,
       (SELECT id FROM workloads_categories WHERE name = 'Deploy' AND scope_type = 'system' LIMIT 1),
       (SELECT id FROM priorities WHERE name = 'very-high' AND scope_type = 'system' LIMIT 1),
       'Run.ai',
       'Run.ai'
FROM workload_types
WHERE workload_types.name = 'InferenceService'
  AND NOT EXISTS (
    SELECT 1 FROM external_workloads_type_config
    WHERE workload_type_id = workload_types.id
    );

INSERT INTO workload_type_resource_interfaces (id, tenant_id, scope_type, workload_type_id, version, resource_interface, created_by, updated_at)
SELECT gen_random_uuid(),
       NULL,
       'system',
       workload_types.id,
       'v1beta1',
       jsonb_build_object(
            'spec', jsonb_build_object(
                'structureDefinition', jsonb_build_object(
                    'rootComponent', jsonb_build_object(
                        'name', 'inferenceservice',
                        'kind', jsonb_build_object(
                            'group', 'serving.kserve.io',
                            'kind', 'InferenceService',
                            'version', 'v1beta1'
                        ),
                        'specDefinition', jsonb_build_object(
                            'fragmentedPodSpecDefinition', jsonb_build_object(
                                'resourcesPath', '.spec.domain.resources',
                                'priorityClassNamePath', '.spec.priorityClassName',
                                'nodeAffinityPath', '.spec.affinity.nodeAffinity'
                            )
                        ),
                        'statusDefinition', jsonb_build_object(
                            'conditionsDefinition', jsonb_build_object(
                                'path', '.status.conditions',
                                'typeFieldName', 'type',
                                'statusFieldName', 'status'
                            ),
                            'statusMappings', jsonb_build_object(
                                'running', jsonb_build_array(
                                    jsonb_build_object(
                                        'byConditions', jsonb_build_array(
                                            jsonb_build_object('type', 'PredictorReady', 'status', 'True'),
                                            jsonb_build_object('type', 'RoutesReady', 'status', 'True'),
                                            jsonb_build_object('type', 'LatestDeploymentReady', 'status', 'True')
                                        )
                                    )
                                ),
                                'failed', jsonb_build_array(
                                    jsonb_build_object(
                                        'byConditions', jsonb_build_array(
                                            jsonb_build_object('type', 'PredictorReady', 'status', 'False'),
                                            jsonb_build_object('type', 'PredictorConfigurationReady', 'status', 'False'),
                                            jsonb_build_object('type', 'RoutesReady', 'status', 'False')
                                        )
                                    )
                                )
                            )
                        )
                    ),
                    'childComponents', jsonb_build_array(
                        jsonb_build_object(
                            'name', 'predictor',
                            'kind', jsonb_build_object(
                                'group', 'apps',
                                'version', 'v1',
                                'kind', 'Deployment'
                            ),
                            'ownerRef', 'inferenceservice',
                            'specDefinition', jsonb_build_object(
                                'podSpecPath', '.spec.predictor',
                                'metadataPath', '.spec.predictor',
                                'fragmentedPodSpecDefinition', jsonb_build_object(
                                    'containerPath', '.spec.predictor.model'
                                )
                            ),
                            'scaleDefinition', jsonb_build_object(
                                'minReplicasPath', '.spec.predictor.minReplicas',
                                'maxReplicasPath', '.spec.predictor.maxReplicas'
                            ),
                            'podSelector', jsonb_build_object(
                                'keyPath', '.metadata.labels["component"]',
                                'value', 'predictor'
                            )
                        ),
                        jsonb_build_object(
                            'name', 'transformer',
                            'kind', jsonb_build_object(
                                'group', 'apps',
                                'version', 'v1',
                                'kind', 'Deployment'
                            ),
                            'ownerRef', 'inferenceservice',
                            'specDefinition', jsonb_build_object(
                                'podSpecPath', '.spec.transformer',
                                'metadataPath', '.spec.transformer'
                            ),
                            'scaleDefinition', jsonb_build_object(
                                'minReplicasPath', '.spec.transformer.minReplicas',
                                'maxReplicasPath', '.spec.transformer.maxReplicas'
                            ),
                            'podSelector', jsonb_build_object(
                                'keyPath', '.metadata.labels["component"]',
                                'value', 'transformer'
                            )
                        )
                    )
                ),
                'optimizationInstructions', jsonb_build_object(
                    'gangScheduling', jsonb_build_object(
                        'podGroups', jsonb_build_array(
                            jsonb_build_object(
                                'name', 'service',
                                'members', jsonb_build_array(
                                    jsonb_build_object(
                                        'componentName', 'predictor',
                                        'groupByKeyPaths', jsonb_build_array(
                                            '.metadata.labels["serving.kserve.io/inferenceservice"]'
                                        )
                                    ),
                                    jsonb_build_object(
                                        'componentName', 'transformer',
                                        'groupByKeyPaths', jsonb_build_array(
                                            '.metadata.labels["serving.kserve.io/inferenceservice"]'
                                        )
                                    )
                                )
                            )
                        )
                    )
                )
            )
        ),
       'Run.ai',
       NOW()
FROM workload_types
WHERE workload_types.name = 'InferenceService'
  AND NOT EXISTS (
    SELECT 1 FROM workload_type_resource_interfaces
    WHERE workload_type_id = workload_types.id
      AND version = 'v1beta1'
);