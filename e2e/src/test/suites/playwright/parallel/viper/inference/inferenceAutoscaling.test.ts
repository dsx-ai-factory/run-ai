/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    inferenceUtils,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { expect } from 'chai';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { toInteger } from 'lodash-es';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { test } from '@playwright/test';
import { generateHttpTraffic } from './inferenceUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

test.describe(
    `inference workload autoscaling tests: ${tags.resource.inferenceWorkload} ${tags.tool.kubeCli} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        async function verifyNumberOfReplicas(
            workloadName: string,
            namespace: string,
            numReplicas: number,
            expectedStatuses: string[] = []
        ): Promise<void> {
            await waitUtils.waitForPods(workloadName, namespace, numReplicas, 30, expectedStatuses, true);
            logger.info(`inference workload has ${numReplicas} pods`);
        }

        async function verifyExactNumberOfReplicas(
            workloadName: string,
            namespace: string,
            numReplicas: number,
            timeout = 120,
            expectedStatuses: string[] = []
        ): Promise<void> {
            await waitUtils.waitForCondition(
                async () => {
                    const pods = await kubeCli.listPodsOfJob(namespace, workloadName, true);
                    return (
                        pods.length === numReplicas &&
                        (expectedStatuses.length === 0 ||
                            pods.every((pod) => expectedStatuses.includes(pod.status.phase)))
                    );
                },
                `workload ${workloadName} has exactly ${numReplicas} pods`,
                true,
                timeout,
                2
            );
            logger.info(`inference workload has exactly ${numReplicas} pods`);
        }

        async function sendInferenceRequestFromPod(namespace: string, url: string): Promise<void> {
            const clientName = objectGenerator.generateShortNameByPrefix('inference-client');
            await kubeCli.CoreV1Api.createNamespacedPod({
                namespace,
                body: {
                    apiVersion: 'v1',
                    kind: 'Pod',
                    metadata: { name: clientName, namespace },
                    spec: {
                        restartPolicy: 'Never',
                        containers: [
                            {
                                name: 'client',
                                image: config.quickStartImage.networkTools,
                                command: ['/bin/sh', '-c', 'sleep infinity']
                            }
                        ]
                    }
                }
            });

            try {
                await kubeCli.waitForPodToRun(namespace, clientName);
                const { stderr } = await kubeCli.execInPod(clientName, namespace, [
                    'curl',
                    '--fail',
                    '--silent',
                    '--show-error',
                    url
                ]);
                expect(stderr).to.equal('');
            } finally {
                await kubeCli.deletePod(clientName, namespace);
            }
        }

        let project: Project;
        let workloadName: string;

        test.beforeAll(async function () {
            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('project'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            workloadName = objectGenerator.generateNameByPrefix('workload');
        });

        test.afterAll(async function () {
            if (!project) {
                return;
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch {
                logger.error(
                    `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        test.afterEach(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(
                    String(project.id),
                    120,
                    flowRootState.users.system_administrator.accessToken,
                    false
                );
                await workloadUtils.clearProjectWorkloadsAndPods(project.name, 120);
            } catch {
                logger.error(
                    `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        // Shared configuration for concurrency autoscaling tests
        const concurrencyAutoscalingConfig = {
            metric: 'concurrency',
            concurrencyTarget: 10,
            minScale: 1,
            maxScale: 3,
            durationSeconds: 60 * 1.2,
            targetPercentage: 0.7 * 0.9, // 0.7 is knative default target percentage, 0.9 is in order not to pass the threshold and get too many replicas
            qps: 10
        };

        shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with autoscaling metric concurrency: scale up by one', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                test.setTimeout(180 * 1000);

                // scenario:
                // - creating an inference workload with autoscaling metric concurrency with target of 10
                // - checking that initial number of pods is the min scale
                // - generating traffic to increase the number of pods by 1

                const {
                    metric,
                    concurrencyTarget,
                    minScale,
                    maxScale,
                    durationSeconds,
                    targetPercentage,
                    qps
                } = concurrencyAutoscalingConfig;
                const concurrency = toInteger(concurrencyTarget * targetPercentage);

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceAutoScale,
                    minScale: minScale,
                    maxScale: maxScale,
                    metric: metric,
                    target: concurrencyTarget
                });

                await waitUtils.waitForWorkloadNonEmptyStatus(
                    workloadName,
                    project.status.namespace,
                    crds.runaiInferenceWorkloadCrd
                );

                const inferenceWorkload = await kubeCli.getInferenceWorkload(
                    workloadName,
                    project.status.namespace
                );
                expect(inferenceWorkload).to.not.be.undefined;
                expect(inferenceWorkload['spec']['minScale']['value']).to.equal(minScale);
                expect(inferenceWorkload['spec']['maxScale']['value']).to.equal(maxScale);
                expect(inferenceWorkload['spec']['metric']['value']).to.equal(metric);
                expect(inferenceWorkload['spec']['target']['value']).to.equal(concurrencyTarget);

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace
                );
                expect(knativeService).to.not.be.undefined;

                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale, ['Running']);

                let url = knativeService['status']['address']['url'];
                url = url + '?sleep=100'; // the sleep parameter is important in order to maintain the required concurrency

                logger.debug(`generating traffic to increase number of replicas to ${minScale + 1}`);
                await generateHttpTraffic({
                    url,
                    namespace: project.status.namespace,
                    durationSeconds,
                    concurrentRequests: (minScale + 1) * concurrency,
                    qps
                });
                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale + 1);
            })
        );

        supportedFrom(SupportedVersions.V2_20)(function () {
            // initialScale is supported from v2.20
            test('creates a workload with autoscaling metric concurrency: scale down by one', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                // scenario:
                // - creating an inference workload with initialScale=2 (minScale+1)
                // - checking that initial number of pods is 2
                // - waiting for the number of pods to decrease to the min scale (1)

                const { metric, concurrencyTarget, minScale, maxScale, durationSeconds } =
                    concurrencyAutoscalingConfig;
                const initialScale = minScale + 1;

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceAutoScale,
                    minScale: minScale,
                    maxScale: maxScale,
                    initialScale: initialScale,
                    metric: metric,
                    target: concurrencyTarget
                });

                await waitUtils.waitForWorkloadNonEmptyStatus(
                    workloadName,
                    project.status.namespace,
                    crds.runaiInferenceWorkloadCrd
                );

                const inferenceWorkload = await kubeCli.getInferenceWorkload(
                    workloadName,
                    project.status.namespace
                );
                expect(inferenceWorkload).to.not.be.undefined;
                expect(inferenceWorkload['spec']['initialScale']['value']).to.equal(initialScale);

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace
                );
                expect(knativeService).to.not.be.undefined;

                await verifyNumberOfReplicas(workloadName, project.status.namespace, initialScale, [
                    'Running'
                ]);

                logger.info(
                    `sleeping for ${durationSeconds} seconds to decrease number of replicas to ${minScale}`
                );
                await waitUtils.sleep(durationSeconds * 1000);
                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale);
            });
        });

        shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with autoscaling metric concurrency: scale up to max', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                test.setTimeout(180 * 1000);

                // scenario:
                // - creating an inference workload with autoscaling metric concurrency with target of 10
                // - checking that initial number of pods is the min scale
                // - generating traffic for more than maxScale replicas to verify maxScale is respected

                const {
                    metric,
                    concurrencyTarget,
                    minScale,
                    maxScale,
                    durationSeconds,
                    targetPercentage,
                    qps
                } = concurrencyAutoscalingConfig;
                const concurrency = toInteger(concurrencyTarget * targetPercentage);
                const targetReplicas = maxScale + 2; // request traffic for more than maxScale to verify the limit

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceAutoScale,
                    minScale: minScale,
                    maxScale: maxScale,
                    metric: metric,
                    target: concurrencyTarget
                });

                await waitUtils.waitForWorkloadNonEmptyStatus(
                    workloadName,
                    project.status.namespace,
                    crds.runaiInferenceWorkloadCrd
                );

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace
                );
                expect(knativeService).to.not.be.undefined;

                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale, ['Running']);

                let url = knativeService['status']['address']['url'];
                url = url + '?sleep=100'; // sleep here is important in order to maintain the required concurrency

                logger.debug(
                    `generating traffic for ${targetReplicas} replicas to verify maxScale (${maxScale}) is respected`
                );
                await generateHttpTraffic({
                    url,
                    namespace: project.status.namespace,
                    durationSeconds,
                    concurrentRequests: targetReplicas * concurrency,
                    qps
                });
                await verifyNumberOfReplicas(workloadName, project.status.namespace, maxScale);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with autoscaling metric throughput: scale up by one', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                test.setTimeout(180 * 1000);

                // scenario:
                // - creating an inference workload with autoscaling metric throughput with target of 100
                // - checking that initial number of pods is the min scale
                // - generating traffic to increase the number of pods by 1

                const metric = 'throughput';
                const throughputTarget = 100;
                const minScale = 2;
                const maxScale = 4;
                const durationSeconds = 60 * 1.2;
                const targetPercentage = 0.7 * 0.9; // 0.7 is knative default target percentage, 0.9 is in order not to pass the threshold and get too many replicas
                const qps = toInteger(throughputTarget * targetPercentage);

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceAutoScale,
                    minScale: minScale,
                    maxScale: maxScale,
                    metric: metric,
                    target: throughputTarget
                });

                await waitUtils.waitForWorkloadNonEmptyStatus(
                    workloadName,
                    project.status.namespace,
                    crds.runaiInferenceWorkloadCrd
                );

                const inferenceWorkload = await kubeCli.getInferenceWorkload(
                    workloadName,
                    project.status.namespace
                );
                expect(inferenceWorkload).to.not.be.undefined;

                expect(inferenceWorkload['spec']['minScale']['value']).to.equal(minScale);
                expect(inferenceWorkload['spec']['maxScale']['value']).to.equal(maxScale);
                expect(inferenceWorkload['spec']['metric']['value']).to.equal(metric);
                expect(inferenceWorkload['spec']['target']['value']).to.equal(throughputTarget);

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace
                );
                expect(knativeService).to.not.be.undefined;

                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale, ['Running']);

                const url = knativeService['status']['address']['url'];
                logger.debug(`generating traffic to increase number of replicas to ${minScale + 1}`);
                await generateHttpTraffic({
                    url,
                    namespace: project.status.namespace,
                    durationSeconds,
                    concurrentRequests: minScale + 1,
                    qps
                });
                await verifyNumberOfReplicas(workloadName, project.status.namespace, minScale + 1);
            })
        );

        supportedFrom(SupportedVersions.V3_0)(function () {
            test('create a workload with autoscaling scale-to-zero', async function () {
                test.setTimeout(420 * 1000);
                const durationSeconds = 60 * 1.5;

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                const [, createdWorkloadName] = await inferenceUtils.createInference(
                    workloadName,
                    {
                        image: config.quickStartImage.inferenceAutoScale,
                        servingPort: { container: 8080 },
                        autoscaling: {
                            minReplicas: 0,
                            maxReplicas: 1,
                            initialReplicas: 1,
                            metric: 'throughput',
                            metricThreshold: 100
                        }
                    },
                    String(project.id)
                );
                workloadName = createdWorkloadName;

                await workloadUtils.waitForWorkloadRunnerPhase(
                    workloadName,
                    project.status.namespace,
                    Phase.Running,
                    180
                );

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace,
                    true,
                    180
                );
                expect(knativeService).to.not.be.undefined;

                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 1, 120, [
                    'Running'
                ]);

                const podGroups = await waitUtils.waitForPodGroups(
                    workloadName,
                    project.status.namespace,
                    1,
                    90
                );
                expect(podGroups.some((podGroup) => podGroup['spec']?.['minMember'] === 0)).to.equal(true);

                logger.info(`sleeping for ${durationSeconds} seconds to decrease number of replicas to 0`);
                await waitUtils.sleep(durationSeconds * 1000);
                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 0, 120);

                const url = knativeService['status']['address']['url'];
                logger.debug('generating traffic to increase number of replicas to 1');
                await sendInferenceRequestFromPod(project.status.namespace, url);
                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 1, 120, [
                    'Running'
                ]);
            });

            test('create a workload with autoscaling scale-to-zero with scaleToZeroRetentionSeconds 60', async function () {
                test.setTimeout(480 * 1000);
                const durationSeconds = 60 * 1.5;

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                const [, createdWorkloadName] = await inferenceUtils.createInference(
                    workloadName,
                    {
                        image: config.quickStartImage.inferenceAutoScale,
                        servingPort: { container: 8080 },
                        autoscaling: {
                            minReplicas: 0,
                            maxReplicas: 1,
                            initialReplicas: 1,
                            metric: 'throughput',
                            metricThreshold: 100,
                            scaleToZeroRetentionSeconds: 60
                        }
                    },
                    String(project.id)
                );
                workloadName = createdWorkloadName;

                await workloadUtils.waitForWorkloadRunnerPhase(
                    workloadName,
                    project.status.namespace,
                    Phase.Running,
                    180
                );

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace,
                    true,
                    180
                );
                expect(knativeService).to.not.be.undefined;

                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 1, 120, [
                    'Running'
                ]);

                logger.info(
                    `sleeping for ${durationSeconds} seconds to trigger the scale-to-zero retention period`
                );
                await waitUtils.sleep(durationSeconds * 1000);

                logger.debug('verifying that the number of replicas is still 1');
                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 1, 15);
                logger.debug('number of replicas is still 1');

                logger.info(
                    `sleeping for ${durationSeconds} seconds to wait the scale-to-zero retention period to end and start deleting the pod`
                );
                await waitUtils.sleep(durationSeconds * 1000);

                logger.debug('verifying that the number of replicas is 0');
                await verifyExactNumberOfReplicas(workloadName, project.status.namespace, 0, 120);
                logger.debug('number of replicas is 0');
            });
        });

        supportedFrom(SupportedVersions.V2_20)(function () {
            test('create a workload with advanced autoscaling fields', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                // not testing the actual autoscaling, just verifying that the fields are accepted by knative

                const metric = 'concurrency';
                const target = 100;
                const targetUtilizationPercentage = '55.5';
                const concurrencyHardLimit = 1;
                const minScale = 0;
                const maxScale = 3;
                const initialScale = 1;
                const activationScale = 2;
                const scaleDownDelaySeconds = 60;
                const initializationTimeoutSeconds = 120;

                logger.info(
                    `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
                );
                const workload = await workloadUtils.createInferenceWorkload(
                    project.status.namespace,
                    workloadName,
                    {
                        image: config.quickStartImage.inferenceAutoScale,
                        minScale: minScale,
                        maxScale: maxScale,
                        initialScale: initialScale,
                        activationScale: activationScale,
                        metric: metric,
                        target: target,
                        targetUtilizationPercentage: targetUtilizationPercentage,
                        concurrencyHardLimit: concurrencyHardLimit,
                        scaleDownDelaySeconds: scaleDownDelaySeconds,
                        initializationTimeoutSeconds: initializationTimeoutSeconds
                    }
                );
                await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

                const inferenceWorkload = await kubeCli.getInferenceWorkload(
                    workloadName,
                    project.status.namespace
                );
                expect(inferenceWorkload).to.not.be.undefined;

                const knativeService = await waitUtils.waitForKnativeService(
                    workloadName,
                    project.status.namespace,
                    false
                );
                expect(knativeService).to.not.be.undefined;

                logger.debug('knative service created, verifying its spec');

                expect(knativeService['spec']).to.not.be.undefined;
                expect(knativeService['spec']['template']).to.not.be.undefined;
                expect(knativeService['spec']['template']['metadata']).to.not.be.undefined;
                expect(knativeService['spec']['template']['metadata']['annotations']).to.not.be.undefined;
                expect(knativeService['spec']['template']['spec']).to.not.be.undefined;

                const annotations = knativeService['spec']['template']['metadata']['annotations'];
                const spec = knativeService['spec']['template']['spec'];

                expect(annotations['autoscaling.knative.dev/min-scale']).to.equal(minScale.toString());
                expect(annotations['autoscaling.knative.dev/max-scale']).to.equal(maxScale.toString());
                expect(annotations['autoscaling.knative.dev/initial-scale']).to.equal(
                    initialScale.toString()
                );
                expect(annotations['autoscaling.knative.dev/activation-scale']).to.equal(
                    activationScale.toString()
                );
                expect(annotations['autoscaling.knative.dev/metric']).to.equal(metric.toString());
                expect(annotations['autoscaling.knative.dev/target']).to.equal(target.toString());
                expect(annotations['autoscaling.knative.dev/target-utilization-percentage']).to.equal(
                    targetUtilizationPercentage.toString()
                );
                expect(annotations['autoscaling.knative.dev/scale-down-delay']).to.equal(
                    scaleDownDelaySeconds.toString() + 's'
                );
                expect(annotations['serving.knative.dev/progress-deadline']).to.equal(
                    initializationTimeoutSeconds.toString() + 's'
                );
                expect(spec['containerConcurrency']).to.equal(concurrencyHardLimit);
            });
        });
    }
);
