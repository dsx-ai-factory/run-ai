/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsManagerUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    createNimService,
    createNgcAuthSecret,
    NGC_AUTH_SECRET_NAME,
    NIM_AUTOSCALING_TEST_TIMEOUT_MS
} from '../../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { NimServiceSpec } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import { ImagePullPolicy } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/model/imagePullPolicy';
import { isNimOperatorInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

supportedFrom(SupportedVersions.V2_24)(function () {
    test.describe(
        `nimservice api autoscaling tests, tags: ${tags.runOn.parallelCI} ${tags.resource.mimService} ${tags.teams.viper}`,
        { tag: [tags.resource.mimService] },
        function () {
            let project: Project;
            let workloadName: string;

            /**
             * Create a NIM Service spec for autoscaling tests with customizable autoscaling config
             */
            function createNimServiceSpecForAutoscaling(autoscalingConfig: {
                minReplicas: number;
                maxReplicas: number;
                metricThreshold: number;
                scaleWindowSeconds?: number;
            }): NimServiceSpec {
                return {
                    ngcAuthSecret: NGC_AUTH_SECRET_NAME,
                    image: config.quickStartImage.nim,
                    imagePullPolicy: ImagePullPolicy.IfNotPresent,
                    servingPort: {
                        port: 8000,
                        exposeExternally: false
                    },
                    modelStore: {
                        pvc: {
                            existingPvc: false,
                            claimInfo: {
                                size: '10M',
                                accessModes: {
                                    readWriteOnce: true
                                }
                            }
                        }
                    },
                    environmentVariables: [
                        {
                            name: 'NIM_REQUEST_LATENCY',
                            value: '2'
                        }
                    ],
                    autoscaling: {
                        metric: 'num_requests_running',
                        ...autoscalingConfig
                    },
                    replicas: null
                };
            }

            test.beforeAll(async function () {
                // Check if nim-operator is installed
                if (!(await isNimOperatorInstalled())) {
                    logger.warn('k8s-nim-operator is not installed on the cluster, skipping tests');
                    test.skip();
                }

                // Create project with CPU-only resources (GPU quota = 0)
                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('nim-autoscaling'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 0, limit: 0 } }]
                });
                logger.info(`Created project: ${project.name} (ID: ${project.id})`);

                // Create NGC auth secret
                await createNgcAuthSecret(project.status.namespace);
                logger.info(`Created NGC auth secret in namespace: ${project.status.namespace}`);
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');

                // Set authentication token for API calls
                workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}; error ${error}`
                    );
                }
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(`Failed to delete workloads of project ${project.name}; error: ${error}`);
                }
            });

            test('Validate HPA Creation Without Scale Window Configuration', async function () {
                test.setTimeout(NIM_AUTOSCALING_TEST_TIMEOUT_MS);

                // Test validates that creating a NIM Service without specifying scaleWindowSeconds results in correct HPA resource creation with basic configuration (behavior field is not set)

                const nimServiceSpec = createNimServiceSpecForAutoscaling({
                    minReplicas: 1,
                    maxReplicas: 4,
                    metricThreshold: 5
                    // scaleWindowSeconds not specified - HPA will use Kubernetes default behavior
                });

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                logger.info(`Created NIM Service: ${workloadName} without scaleWindowSeconds configuration`);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 90, ['Running'], true);
                logger.info('NIM Service Running with 1 replica');

                // Validate HPA resource creation and configuration
                logger.info('=== VALIDATING HPA RESOURCE ===');
                const hpa = await kubeCli.getHPA(workloadName, project.status.namespace);

                expect(hpa, 'HPA should be created').toBeDefined();
                logger.info('HPA resource exists');

                expect(hpa.spec.minReplicas, 'HPA minReplicas should match configuration').toBe(1);
                logger.info('HPA minReplicas = 1');

                expect(hpa.spec.maxReplicas, 'HPA maxReplicas should match configuration').toBe(4);
                logger.info('HPA maxReplicas = 4');

                expect(hpa.spec.metrics, 'HPA should have metrics configured').toBeDefined();
                expect(hpa.spec.metrics.length, 'HPA should have exactly one metric').toBe(1);
                logger.info('HPA has 1 metric configured');

                expect(hpa.spec.metrics[0].type, 'HPA metric type should be Pods').toBe('Pods');
                logger.info('HPA metric type = Pods');

                expect(
                    hpa.spec.metrics[0].pods.metric.name,
                    'HPA metric name should be num_requests_running'
                ).toBe('num_requests_running');
                logger.info('HPA metric name = num_requests_running');

                expect(
                    hpa.spec.metrics[0].pods.target.averageValue,
                    'HPA metric threshold should match configuration'
                ).toBe('5');
                logger.info('HPA metric threshold = 5');

                // Verify that behavior field is not set when scaleWindowSeconds is not specified
                expect(
                    hpa.spec.behavior,
                    'HPA behavior field should not be set when scaleWindowSeconds is not specified'
                ).toBeUndefined();
                logger.info('HPA behavior field not set (will use Kubernetes default behavior)');

                logger.info('All basic HPA validations passed');
            });

            test('Validate HPA With Explicit Scale Window Configuration', async function () {
                test.setTimeout(NIM_AUTOSCALING_TEST_TIMEOUT_MS);

                // Test validates that explicitly specifying scaleWindowSeconds (120s) creates HPA with behavior field and correct stabilizationWindow value

                const scaleWindowSeconds = 120; // Explicitly specify custom scale window
                const nimServiceSpec = createNimServiceSpecForAutoscaling({
                    minReplicas: 1,
                    maxReplicas: 4,
                    metricThreshold: 5,
                    scaleWindowSeconds
                });

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                logger.info(`Created NIM Service: ${workloadName} with scaleWindowSeconds=120`);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 90, ['Running'], true);
                logger.info('NIM Service Running with 1 replica');

                // Validate HPA scale window configuration
                logger.info('=== VALIDATING HPA WITH EXPLICIT SCALE WINDOW ===');
                const hpa = await kubeCli.getHPA(workloadName, project.status.namespace);

                expect(hpa, 'HPA should be created').toBeDefined();
                logger.info('HPA resource exists');

                expect(
                    hpa.spec.behavior,
                    'HPA behavior field should be configured when scaleWindowSeconds is specified'
                ).toBeDefined();
                logger.info('HPA behavior field exists');

                expect(
                    hpa.spec.behavior.scaleDown,
                    'HPA scaleDown behavior should be configured'
                ).toBeDefined();
                logger.info('HPA scaleDown behavior exists');

                expect(
                    hpa.spec.behavior.scaleDown.stabilizationWindowSeconds,
                    'HPA scaleDown stabilization window should match specified value (120s)'
                ).toBe(scaleWindowSeconds);
                logger.info(`HPA scaleDown stabilizationWindow correctly set to ${scaleWindowSeconds}s`);

                logger.info('HPA with explicit scale window configuration validated');
            });

            test('Validate HPA is Not Created When Autoscaling is Disabled', async function () {
                test.setTimeout(NIM_AUTOSCALING_TEST_TIMEOUT_MS);

                // Test validates that HPA is not created when autoscaling is not configured

                // Reuse existing spec creation function and disable autoscaling
                const nimServiceSpec = createNimServiceSpecForAutoscaling({
                    minReplicas: 1,
                    maxReplicas: 4,
                    metricThreshold: 5
                });
                // Remove autoscaling and use fixed replicas instead
                delete nimServiceSpec.autoscaling;
                nimServiceSpec.replicas = 1;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                logger.info(`Created NIM Service: ${workloadName} without autoscaling configuration`);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 90, ['Running'], true);
                logger.info('NIM Service Running with 1 replica');

                // Validate HPA is NOT created
                logger.info('=== VALIDATING HPA IS NOT CREATED ===');
                let hpaExists = true;
                try {
                    await kubeCli.getHPA(workloadName, project.status.namespace);
                } catch (error) {
                    // Expected: HPA should not exist
                    hpaExists = false;
                    logger.info('HPA does not exist (as expected)');
                }

                expect(hpaExists, 'HPA should not be created when autoscaling is disabled').toBe(false);
                logger.info('Validated that HPA is not created without autoscaling configuration');
            });
        }
    );
});
