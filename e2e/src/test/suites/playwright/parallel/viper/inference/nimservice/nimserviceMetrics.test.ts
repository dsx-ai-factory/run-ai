/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsManagerUtils,
    workloadsUtilsV2,
    metricUtils
} from '../../../../../../../modules/helpers';
import {
    createNimService,
    createNgcAuthSecret,
    createDefaultNimServiceSpec,
    NIM_HEAVY_TEST_TIMEOUT_MS,
    NIM_METRICS_DISTRIBUTED_TEST_TIMEOUT_MS
} from '../../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import {
    isNimOperatorInstalled,
    isLwsInstalled
} from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { NimServiceSpec } from '../../../../../../../modules/api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import { workloadsCrudService, workloadsService } from '../../../../../../../modules/api';
import { generateHttpTraffic, getInternalServingUrl } from '../inferenceUtils';
import { WorkloadMetricType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadMetricType';
import { PodMetricType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/podMetricType';
import { MetricsCompositeResponse } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/metricsCompositeResponse';
import { MeasurementResponse } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/measurementResponse';
import { HistogramSeries } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/histogramSeries';
import { test } from '@playwright/test';

import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`NIM Service Metrics tests, tags: ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.k8sFlavWorkloadsKit}`, function () {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let project: Project;
        let workloadName: string;
        let nimServiceSpec: NimServiceSpec;

        // NIM metrics to validate
        const nimMetrics: WorkloadMetricType[] = [
            WorkloadMetricType.NimNumRequestsRunning,
            WorkloadMetricType.NimRequestSuccessTotal
        ];

        const podMetrics: PodMetricType[] = [
            PodMetricType.NimNumRequestsRunning,
            PodMetricType.NimRequestSuccessTotal
        ];

        test.beforeAll(async function () {
            test.setTimeout(120 * 1000);

            if (!(await isNimOperatorInstalled())) {
                logger.warn('k8s-nim-operator is not installed on the cluster, skipping test');
                test.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('project'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
            await createNgcAuthSecret(project.status.namespace);

            workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsCrudService.inferencesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            workloadsService.workloadsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsService.podsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateNameByPrefix('workload');
            nimServiceSpec = createDefaultNimServiceSpec();
            nimServiceSpec.servingPort.exposeExternally = false;
            nimServiceSpec.replicas = 1;
        });

        test.afterAll(async function () {
            test.setTimeout(120 * 1000);
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
            test.setTimeout(120 * 1000);
            if (project === undefined) {
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
            } catch (error) {
                logger.error(`Failed to delete workloads of project ${project.name}; error: ${error}`);
            }
        });

        // Helper to parse metrics response
        function parseMetricsFromResponse(response: MetricsCompositeResponse): {
            measurements: MeasurementResponse[];
            histogram: HistogramSeries[];
            allMetrics: (MeasurementResponse | HistogramSeries)[];
            metricTypes: string[];
        } {
            const measurements = (response.measurements || []) as MeasurementResponse[];
            const histogram = (response.histogram || []) as HistogramSeries[];
            const allMetrics = [...measurements, ...histogram];
            const metricTypes = allMetrics.map((m) => m.type);
            return { measurements, histogram, allMetrics, metricTypes };
        }

        // Helper to verify metrics presence
        function verifyAllMetricsPresent(
            response: MetricsCompositeResponse,
            requestedMetrics: Array<WorkloadMetricType | PodMetricType>
        ): void {
            const { metricTypes: responseMetricTypes } = parseMetricsFromResponse(response);

            logger.info(`Requested metrics: ${JSON.stringify(requestedMetrics)}`);
            logger.info(`Response metrics: ${JSON.stringify(responseMetricTypes)}`);

            requestedMetrics.forEach((requestedMetric) => {
                expect(responseMetricTypes).to.include(
                    requestedMetric,
                    `Metric ${requestedMetric} should be present in response`
                );
            });
        }

        // Retry mechanism for fetching metrics
        async function fetchMetricsWithWait(
            fetcher: (start: Date, end: Date) => Promise<MetricsCompositeResponse>,
            timeoutMs: number = 300000,
            intervalSeconds: number = 5
        ): Promise<MetricsCompositeResponse> {
            let finalMetricsResponse: MetricsCompositeResponse | null = null;
            let attemptCount = 0;

            await waitUtils.waitForCondition(
                async () => {
                    attemptCount++;
                    logger.info(`Attempt ${attemptCount}: Getting NIM metrics`);

                    const { start, end } = metricUtils.getRecommendedStartEndForMetrics();
                    const response = await fetcher(start, end);

                    const { allMetrics } = parseMetricsFromResponse(response);
                    // Check if we have data points
                    const hasData = allMetrics.some((m: any) => m.values && m.values.length > 0);

                    if (hasData) {
                        finalMetricsResponse = response;
                        logger.info(`Attempt ${attemptCount} - Found metrics data!`);
                    } else {
                        logger.info(`Attempt ${attemptCount} - No data yet, waiting...`);
                    }

                    return hasData;
                },
                'Wait for NIM metrics data',
                true,
                timeoutMs / 1000,
                intervalSeconds
            );

            if (!finalMetricsResponse) {
                throw new Error('Failed to fetch metrics with data within timeout');
            }
            return finalMetricsResponse;
        }

        shouldRunNativeWorkloadCrdTests(() =>
        test('creates a nimservice, generates traffic, and validates metrics', async function () {
            test.setTimeout(NIM_HEAVY_TEST_TIMEOUT_MS);

            logger.info(`Creating NIM service: ${workloadName}`);
            const createResponse = await createNimService(workloadName, project, nimServiceSpec);
            const workloadId = createResponse.metadata.id;

            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 300);
            logger.info(`NIM service ${workloadId} is Running`);

            const internalUrl = await getInternalServingUrl(workloadId);
            expect(internalUrl).to.contain('cluster.local');
            logger.info(`Got serving URL: ${internalUrl}`);

            logger.info('Generating traffic to inference workload to produce metrics');
            await generateHttpTraffic({
                url: internalUrl.startsWith('http') ? internalUrl : `http://${internalUrl}`,
                namespace: project.status.namespace,
                durationSeconds: 15,
                concurrentRequests: 5,
                qps: 5
            });

            logger.info(`Fetching workload metrics for ${workloadId}`);
            const workloadMetricsResponse = await fetchMetricsWithWait(async (start: Date, end: Date) => {
                const { body } = await metricUtils.getWorkloadMetrics(workloadId, nimMetrics, start, end);
                return body;
            });
            verifyAllMetricsPresent(workloadMetricsResponse, nimMetrics);

            const {
                body: { pods }
            } = await workloadsUtilsV2.getWorkloadPods(workloadId);
            expect(pods).to.not.be.empty;
            const podId = pods[0].id;

            logger.info(`Fetching pod metrics for ${podId}`);
            const podMetricsResponse = await fetchMetricsWithWait(async (start: Date, end: Date) => {
                    const { body } = await metricUtils.getPodMetrics(
                        workloadId,
                        podId,
                        podMetrics,
                        start,
                        end
                    );
                return body;
            });
            verifyAllMetricsPresent(podMetricsResponse, podMetrics);
            })
        );

        test.describe('Distributed NIM Metrics', function () {
            let distributedWorkloadId: string;
            let distributedWorkloadName: string;
            let distributedNimSpec: NimServiceSpec;

            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.info('LWS is not installed, skipping distributed test');
                    test.skip();
                }
            });

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a distributed nimservice and validates metrics', async function () {
                test.setTimeout(NIM_METRICS_DISTRIBUTED_TEST_TIMEOUT_MS);

                distributedWorkloadName = objectGenerator.generateNameByPrefix('dist-nim-metrics');
                distributedNimSpec = createDefaultNimServiceSpec();
                distributedNimSpec.servingPort.exposeExternally = false;
                distributedNimSpec.replicas = 1;
                distributedNimSpec.multiNode = {
                    workers: 1
                };

                logger.info(`Creating distributed NIM workload: ${distributedWorkloadName}`);
                    const result = await createNimService(
                        distributedWorkloadName,
                        project,
                        distributedNimSpec
                    );
                distributedWorkloadId = result.metadata.id;

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        distributedWorkloadId,
                        Phase.Running,
                        300
                    );
                logger.info(`Distributed NIM service ${distributedWorkloadId} is Running`);

                const internalUrl = await getInternalServingUrl(distributedWorkloadId);
                logger.info(`Got serving URL: ${internalUrl}`);

                logger.info('Generating traffic to distributed inference workload');
                await generateHttpTraffic({
                    url: internalUrl.startsWith('http') ? internalUrl : `http://${internalUrl}`,
                    namespace: project.status.namespace,
                    durationSeconds: 15,
                    concurrentRequests: 5,
                    qps: 5
                });

                logger.info(`Fetching workload metrics for ${distributedWorkloadId} (Aggregate view)`);
                    const workloadMetricsResponse = await fetchMetricsWithWait(
                        async (start: Date, end: Date) => {
                    const { body } = await metricUtils.getWorkloadMetrics(
                        distributedWorkloadId,
                        nimMetrics,
                        start,
                        end
                    );
                    return body;
                        }
                    );
                verifyAllMetricsPresent(workloadMetricsResponse, nimMetrics);

                const {
                    body: { pods }
                } = await workloadsUtilsV2.getWorkloadPods(distributedWorkloadId);
                const leaderPod = pods.find((p) => p.task === 'leader');
                expect(leaderPod, 'Leader pod not found').to.not.be.undefined;
                const leaderPodId = leaderPod!.id;

                logger.info(`Fetching leader pod metrics for ${leaderPodId} (Leader-Specific view)`);
                const podMetricsResponse = await fetchMetricsWithWait(async (start: Date, end: Date) => {
                    const { body } = await metricUtils.getPodMetrics(
                        distributedWorkloadId,
                        leaderPodId,
                        podMetrics,
                        start,
                        end
                    );
                    return body;
                });
                verifyAllMetricsPresent(podMetricsResponse, podMetrics);
                })
            );
        });
    });
});
