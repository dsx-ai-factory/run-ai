-- Update the resource interface for InferenceService workload type
UPDATE workload_type_resource_interfaces
SET resource_interface = jsonb_build_object(
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
                                'componentTypeSelector', jsonb_build_object(
                                    'keyPath', '.metadata.labels["component"]',
                                    'value', 'predictor'
                                )
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
                                'componentTypeSelector', jsonb_build_object(
                                    'keyPath', '.metadata.labels["component"]',
                                    'value', 'transformer'
                                )
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
    updated_at = NOW()
WHERE workload_type_id IN (
    SELECT id FROM workload_types
    WHERE name = 'InferenceService' AND "group" = 'serving.kserve.io'
)
AND version = 'v1beta1';

