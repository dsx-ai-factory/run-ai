/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    metricUtils
} from '../../../../../../modules/helpers';
import { MetricsCompositeResponse } from '../../../../../../modules/api/apis/workloadsApi/generated/model/metricsCompositeResponse';
import { MeasurementResponse } from '../../../../../../modules/api/apis/workloadsApi/generated/model/measurementResponse';
import { HistogramSeries } from '../../../../../../modules/api/apis/workloadsApi/generated/model/histogramSeries';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { workloadsService } from '../../../../../../modules/api';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { isLwsInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadMetricType';
import { PodMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/podMetricType';
import { kubeCli } from '../../../../../../modules/cli';
import {
    createInference,
    createNimMetricsFixtureSpec,
    createDistributedInference
} from '../../../parallel/viper/inference/inferenceUtils.js';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { DistributedInferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

test.describe(`NIM metrics tests, tags: ${tags.resource.inferenceWorkload} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        test.setTimeout(300 * 1000);
        let project: Project;
        let singleReplicaWorkloadId: string;
        let singleReplicaWorkloadName: string;
        let singleReplicaWorkloadSpec: InferenceSpecSpec;
        let twoReplicaWorkloadId: string;
        let twoReplicaWorkloadName: string;
        let twoReplicaWorkloadSpec: InferenceSpecSpec;
        let distributedworkloadId: string;
        let distributedworkloadName: string;
        let distributedworkloadSpec: DistributedInferenceSpecSpec;

        const nimMetrics: WorkloadMetricType[] = [
            WorkloadMetricType.NimNumRequestsRunning,
            WorkloadMetricType.NimNumRequestsWaiting,
            WorkloadMetricType.NimNumRequestMax,
            WorkloadMetricType.NimRequestSuccessTotal,
            WorkloadMetricType.NimRequestFailureTotal,
            WorkloadMetricType.NimGpuCacheUsagePerc,
            WorkloadMetricType.NimTimeToFirstTokenSeconds,
            WorkloadMetricType.NimE2ERequestLatencySeconds,
            WorkloadMetricType.NimTimeToFirstTokenSecondsPercentiles,
            WorkloadMetricType.NimE2ERequestLatencySecondsPercentiles
        ];

        const podMetrics: PodMetricType[] = [
            PodMetricType.NimNumRequestsRunning,
            PodMetricType.NimNumRequestsWaiting,
            PodMetricType.NimNumRequestMax,
            PodMetricType.NimRequestSuccessTotal,
            PodMetricType.NimRequestFailureTotal,
            PodMetricType.NimGpuCacheUsagePerc,
            PodMetricType.NimTimeToFirstTokenSeconds,
            PodMetricType.NimE2ERequestLatencySeconds,
            PodMetricType.NimTimeToFirstTokenSecondsPercentiles,
            PodMetricType.NimE2ERequestLatencySecondsPercentiles
        ];

        const NIM_METRIC = {
            GPU_CACHE_USAGE_PERC: WorkloadMetricType.NimGpuCacheUsagePerc,
            REQUEST_SUCCESS_TOTAL: WorkloadMetricType.NimRequestSuccessTotal,
            TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES: WorkloadMetricType.NimTimeToFirstTokenSecondsPercentiles
        } as const;

        const NAME_PREFIX = {
            SINGLE_REPLICA: 'singlereplica',
            TWO_REPLICA: 'tworeplica'
        } as const;

        let nimFixtureSpec: Pick<InferenceSpecSpec, 'image' | 'command' | 'args' | 'compute'>;

        const LABEL = {
            NIM_WORKLOAD: 'run.ai/nim-workload'
        } as const;

        const WAIT_CONFIG = {
            WORKLOAD_TIMEOUT_MS: 300,
            WORKLOAD_RETRIES: 10,
            POD_TIMEOUT_SECONDS: 300,
            POD_RETRIES: 10
        } as const;

        test.beforeAll(async function () {
            nimFixtureSpec = createNimMetricsFixtureSpec();
            project = await projectResourcesUtils.createProjectAndNamespace('project', null, 8, 8);
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsService.workloadsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll(async function () {
            if (!project) return;
            try {
                await operations.deleteProjectNamespaceAndWorkloads(project, 120);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads and project ${project.name} namespace ${project.namespace} with error ${error}`
                );
            }
        });

        // Shared helper to fetch NIM metrics with waiting logic
        type Composite = MetricsCompositeResponse;

        function parseMetricsFromResponse(response: Composite): {
            measurements: MeasurementResponse[];
            histogram: HistogramSeries[];
            allMetrics: Array<MeasurementResponse | HistogramSeries>;
            metricTypes: string[];
        } {
            const measurements = (response.measurements || []) as MeasurementResponse[];
            const histogram = (response.histogram || []) as HistogramSeries[];
            const allMetrics = [...measurements, ...histogram];
            const metricTypes = allMetrics.map((m) => m.type);
            return { measurements, histogram, allMetrics, metricTypes };
        }

        async function fetchMetricsWithWait(
            fetcher: (start: Date, end: Date) => Promise<Composite>,
            timeoutMs: number,
            retries: number
        ): Promise<Composite> {
            let finalMetricsResponse: Composite | null = null;
            let attemptCount = 0;

            await waitUtils.waitForCondition(
                async () => {
                    attemptCount++;
                    logger.info(`Attempt ${attemptCount}: Getting NIM metrics`);

                    const { start, end } = metricUtils.getRecommendedStartEndForMetrics();
                    const response = await fetcher(start, end);

                    const { allMetrics, histogram } = parseMetricsFromResponse(response);
                    const hasData = allMetrics.some((m: any) => m.values && m.values.length > 0);
                    const hasPercentilesData = histogram.some(
                        (m) =>
                            String(m.type) === String(NIM_METRIC.TIME_TO_FIRST_TOKEN_SECONDS_PERCENTILES) &&
                            (m as any).values &&
                            (m as any).values.length > 0
                    );

                    logger.info(
                        `Attempt ${attemptCount} - Has data: ${hasData}, Has percentiles: ${hasPercentilesData}`
                    );

                    if (hasData && hasPercentilesData) {
                        finalMetricsResponse = response;
                        logger.info(`Attempt ${attemptCount} - Found metrics data with percentiles!`);
                    } else {
                        logger.info(`Attempt ${attemptCount} - No complete data yet, waiting...`);
                    }

                    return hasData && hasPercentilesData;
                },
                'Wait for NIM metrics data',
                true,
                timeoutMs,
                retries
            );

            return finalMetricsResponse;
        }

        // Helper function to get workload metrics
        async function getWorkloadMetrics(
            workloadId: string,
            nimMetrics: WorkloadMetricType[]
        ): Promise<MetricsCompositeResponse> {
            logger.info(`Fetching workload metrics for ${workloadId}`);
            return await fetchMetricsWithWait(
                async (start: Date, end: Date) => {
                    const { body } = await metricUtils.getWorkloadMetrics(workloadId, nimMetrics, start, end);
                    return body;
                },
                WAIT_CONFIG.WORKLOAD_TIMEOUT_MS,
                WAIT_CONFIG.WORKLOAD_RETRIES
            );
        }

        // Helper function to get pod metrics
        async function getPodMetrics(
            workloadId: string,
            podId: string,
            nimMetrics: PodMetricType[]
        ): Promise<MetricsCompositeResponse> {
            logger.info(`Fetching pod metrics for ${podId} (workload ${workloadId})`);
            return await fetchMetricsWithWait(
                async (start: Date, end: Date) => {
                    const { body } = await metricUtils.getPodMetrics(
                        workloadId,
                        podId,
                        nimMetrics,
                        start,
                        end
                    );
                    return body;
                },
                WAIT_CONFIG.POD_TIMEOUT_SECONDS,
                WAIT_CONFIG.POD_RETRIES
            );
        }

        // Helper function to verify all requested metrics are present in response
        function verifyAllMetricsPresent(
            response: MetricsCompositeResponse,
            requestedMetrics: Array<WorkloadMetricType | PodMetricType | string>
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

        // Helper function to get specific metric value from response
        function getMetricValue(
            response: MetricsCompositeResponse | null,
            metricType: string | WorkloadMetricType | PodMetricType
        ): any {
            if (!response) {
                logger.info(`Response is null/undefined for metric ${metricType}`);
                return null;
            }

            const { allMetrics } = parseMetricsFromResponse(response);
            const metric = allMetrics.find((m) => m.type === metricType) as any;
            if (!metric || !metric.values || metric.values.length === 0) {
                logger.info(`Metric ${metricType} not found or has no values`);
                return null;
            }

            const latestValue = metric.values[metric.values.length - 1];
            return (latestValue as any).data ?? (latestValue as any).value ?? null;
        }

        test.describe('NIM Metrics Validation Tests -- Non-distributed', function () {
            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }
                singleReplicaWorkloadName = objectGenerator.generateNameByPrefix(NAME_PREFIX.SINGLE_REPLICA);
                twoReplicaWorkloadName = objectGenerator.generateNameByPrefix(NAME_PREFIX.TWO_REPLICA);

                singleReplicaWorkloadSpec = {
                    ...nimFixtureSpec,
                    servingPort: {
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1,
                        metric: 'throughput',
                        metricThreshold: 100
                    },
                    labels: [
                        {
                            name: LABEL.NIM_WORKLOAD,
                            value: 'true'
                        }
                    ]
                };
                twoReplicaWorkloadSpec = {
                    ...nimFixtureSpec,
                    servingPort: {
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 2,
                        maxReplicas: 2,
                        metric: 'throughput',
                        metricThreshold: 100
                    },
                    labels: [
                        {
                            name: LABEL.NIM_WORKLOAD,
                            value: 'true'
                        }
                    ]
                };

                // Create two workloads with different replica counts
                logger.info('Creating workload with 1 replica...');
                [singleReplicaWorkloadId, singleReplicaWorkloadName] = await createInference(
                    singleReplicaWorkloadName,
                    singleReplicaWorkloadSpec,
                    project
                );
                logger.info(`singleReplicaWorkloadId: ${singleReplicaWorkloadId}`);
                logger.info(`singleReplicaWorkloadName: ${singleReplicaWorkloadName}`);

                logger.info('Creating workload with 2 replicas...');
                [twoReplicaWorkloadId, twoReplicaWorkloadName] = await createInference(
                    twoReplicaWorkloadName,
                    twoReplicaWorkloadSpec,
                    project
                );
                logger.info(`twoReplicaWorkloadId: ${twoReplicaWorkloadId}`);
                logger.info(`twoReplicaWorkloadName: ${twoReplicaWorkloadName}`);

                // Wait for both workloads to be in Running status
                await waitUtils.waitForBackendWorkloadPhase(
                    singleReplicaWorkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
                await waitUtils.waitForBackendWorkloadPhase(
                    twoReplicaWorkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
            });

            test('T5404273:Check the metrics names of runai inference workload by workload', async function () {
                await test.step('Verify NIM metrics for workload', async () => {
                    logger.info(`Testing metrics for workload: ${singleReplicaWorkloadId}`);
                    const finalMetricsResponse1 = await getWorkloadMetrics(
                        singleReplicaWorkloadId,
                        nimMetrics
                    );
                    verifyAllMetricsPresent(finalMetricsResponse1, nimMetrics);
                });
            });

            test('T5404274:Check the metrics names of runai inference workload by pod', async function () {
                await test.step('Verify NIM metrics per pod for workload', async () => {
                    logger.info('Testing pod metrics for workload', singleReplicaWorkloadId);
                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(singleReplicaWorkloadId);
                    expect(pods).to.not.be.empty;
                    const podMetricsResponse = await getPodMetrics(
                        singleReplicaWorkloadId,
                        pods[0].id,
                        podMetrics
                    );
                    verifyAllMetricsPresent(podMetricsResponse, podMetrics);
                });
            });

            test('T5404275: Check the metrics of runai inference workload(1 replica)', async function () {
                let workloadMetricsResponse: MetricsCompositeResponse | null;
                let podMetricsResponse: MetricsCompositeResponse | null;
                await test.step('Get NIM metrics per workload', async () => {
                    logger.info('Get NIM metrics for workload', singleReplicaWorkloadId);
                    workloadMetricsResponse = await getWorkloadMetrics(singleReplicaWorkloadId, nimMetrics);
                    logger.info(
                        `workloadMetricsResponse: ${JSON.stringify(workloadMetricsResponse, null, 2)}`
                    );
                    const isValid = metricUtils.validateNIMMetrics(workloadMetricsResponse);
                    expect(isValid).to.be.true;
                });

                await test.step('Get NIM metrics per pod', async () => {
                    logger.info('Get NIM metrics for pod', singleReplicaWorkloadId);
                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(singleReplicaWorkloadId);
                    expect(pods).to.not.be.empty;
                    podMetricsResponse = await getPodMetrics(singleReplicaWorkloadId, pods[0].id, podMetrics);
                    logger.info(`podMetricsResponse: ${JSON.stringify(podMetricsResponse, null, 2)}`);
                });

                await test.step('Compare the metrics of workload and pod', () => {
                    // NIM_GPU_CACHE_USAGE_PERC is the same (average of single pod = same as pod)
                    const workloadGpuCache = getMetricValue(
                        workloadMetricsResponse,
                        NIM_METRIC.GPU_CACHE_USAGE_PERC
                    );
                    const podGpuCache = getMetricValue(podMetricsResponse, NIM_METRIC.GPU_CACHE_USAGE_PERC);
                    logger.info(`GPU Cache - Workload: ${workloadGpuCache}, Pod: ${podGpuCache}`);
                    expect(parseFloat(workloadGpuCache)).to.be.closeTo(parseFloat(podGpuCache), 4.0);

                    // NIM_REQUEST_SUCCESS_TOTAL is the same (sum of single pod = same as pod)
                    const workloadRequestSuccess = getMetricValue(
                        workloadMetricsResponse,
                        NIM_METRIC.REQUEST_SUCCESS_TOTAL
                    );
                    const podRequestSuccess = getMetricValue(
                        podMetricsResponse,
                        NIM_METRIC.REQUEST_SUCCESS_TOTAL
                    );
                    logger.info(
                        `Request Success - Workload: ${workloadRequestSuccess}, Pod: ${podRequestSuccess}`
                    );
                    expect(workloadRequestSuccess).to.equal(podRequestSuccess);
                });
            });

            test('T5404276: Check the metrics of runai inference workload(2 replicas)', async function () {
                let workloadMetricsResponse: MetricsCompositeResponse | null;
                let podMetricsResponse1: MetricsCompositeResponse | null;
                let podMetricsResponse2: MetricsCompositeResponse | null;
                await test.step('Get NIM metrics per workload', async () => {
                    logger.info('Get NIM metrics for workload', twoReplicaWorkloadId);
                    workloadMetricsResponse = await getWorkloadMetrics(twoReplicaWorkloadId, nimMetrics);
                    logger.info(
                        `workloadMetricsResponse: ${JSON.stringify(workloadMetricsResponse, null, 2)}`
                    );
                    const isValid = metricUtils.validateNIMMetrics(workloadMetricsResponse);
                    expect(isValid).to.be.true;
                });

                await test.step('Get NIM metrics per pod', async () => {
                    logger.info('Get NIM metrics for pod', twoReplicaWorkloadId);
                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(twoReplicaWorkloadId);
                    expect(pods).to.not.be.empty;
                    podMetricsResponse1 = await getPodMetrics(twoReplicaWorkloadId, pods[0].id, podMetrics);
                    podMetricsResponse2 = await getPodMetrics(twoReplicaWorkloadId, pods[1].id, podMetrics);
                    logger.info(`podMetricsResponse1: ${JSON.stringify(podMetricsResponse1, null, 2)}`);
                    logger.info(`podMetricsResponse2: ${JSON.stringify(podMetricsResponse2, null, 2)}`);
                    verifyAllMetricsPresent(podMetricsResponse1, podMetrics);
                    verifyAllMetricsPresent(podMetricsResponse2, podMetrics);
                });
                await test.step('Compare the metrics of workload and pod', () => {
                    // NIM_GPU_CACHE_USAGE_PERC: average of pods (with tolerance for floating point precision)
                    const workloadGpuCache = parseFloat(
                        getMetricValue(workloadMetricsResponse, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                    );
                    const podGpuCache1 = parseFloat(
                        getMetricValue(podMetricsResponse1, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                    );
                    const podGpuCache2 = parseFloat(
                        getMetricValue(podMetricsResponse2, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                    );
                    const expectedAverage = (podGpuCache1 + podGpuCache2) / 2;
                    logger.info(`GPU Cache - Workload: ${workloadGpuCache}, Expected: ${expectedAverage}`);
                    expect(workloadGpuCache).to.be.closeTo(expectedAverage, 4.0);

                    // NIM_REQUEST_SUCCESS_TOTAL: sum of pods
                    const workloadRequestSuccess = parseInt(
                        getMetricValue(workloadMetricsResponse, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                    );
                    const podRequestSuccess1 = parseInt(
                        getMetricValue(podMetricsResponse1, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                    );
                    const podRequestSuccess2 = parseInt(
                        getMetricValue(podMetricsResponse2, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                    );
                    const expectedSum = podRequestSuccess1 + podRequestSuccess2;
                    logger.info(
                        `Request Success - Workload: ${workloadRequestSuccess}, Expected: ${expectedSum}`
                    );
                    expect(workloadRequestSuccess).to.equal(expectedSum);
                });
            });

            test('T5404280: Invalid API parameter', async function () {
                await test.step('get invalid metric name should return error', async () => {
                    const { start, end } = metricUtils.getRecommendedStartEndForMetrics();
                    try {
                        await workloadsService.workloadsApi.getWorkloadMetrics(
                            singleReplicaWorkloadId,
                            ['INVALID_METRIC_TYPE'] as any,
                            start,
                            end
                        );
                        throw new Error('Expected API call to fail');
                    } catch (error: any) {
                        expect(error.statusCode).to.equal(400);
                        expect(error.body.message).to.include('metricType: invalid value');
                    }
                });
            });

            test('T5404281: Get nim metrics for non-nim workload', async function () {
                let nonNimWorkloadId: string;
                await test.step('Create a regular workload', async () => {
                    const workloadName = objectGenerator.generateShortNameByPrefix('regular-workload');
                    const { body: workload } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        true,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity'
                        }
                    );
                    nonNimWorkloadId = workload.workloadId;
                    logger.info(`Created regular workload: ${workloadName}, ID: ${nonNimWorkloadId}`);
                });

                await test.step('Try to get NIM metrics for non-NIM workload', async () => {
                    const { start, end } = metricUtils.getRecommendedStartEndForMetrics();

                    try {
                        await workloadsService.workloadsApi.getWorkloadMetrics(
                            nonNimWorkloadId,
                            [
                                WorkloadMetricType.NimNumRequestsRunning,
                                WorkloadMetricType.NimNumRequestsWaiting
                            ],
                            start,
                            end
                        );
                        throw new Error('Expected API call to fail');
                    } catch (error: any) {
                        expect(error.statusCode).to.equal(400);
                        expect(error.body.message).to.include('workload is not identified as NVIDIA NIM');
                    }

                    await workloadsCrudUtils.deleteTraining(nonNimWorkloadId);
                });
            });
        });

        test.describe('NIM Metrics Validation Tests -- Distributed', function () {
            test.beforeAll(async function () {
                if (!isLwsInstalled()) {
                    logger.info('LWS is not installed, skipping test');
                    return;
                }
                // Generate workload names
                distributedworkloadName = objectGenerator.generateNameByPrefix('distinf');
                distributedworkloadSpec = {
                    workers: 2,
                    replicas: 2,
                    servingPort: {
                        port: 8000
                    },
                    leader: {
                        ...nimFixtureSpec,
                        labels: [
                            {
                                name: LABEL.NIM_WORKLOAD,
                                value: 'true'
                            }
                        ]
                    },
                    worker: {
                        ...nimFixtureSpec
                    }
                };
                // Create two workloads with different replica counts
                logger.info('Creating distributed workload...');
                [distributedworkloadId, distributedworkloadName] = await createDistributedInference(
                    distributedworkloadName,
                    distributedworkloadSpec,
                    project
                );
                logger.info(`workloadId1: ${distributedworkloadId}`);
                logger.info(`workloadName: ${distributedworkloadName}`);
                // Wait for both workloads to be in Ready status
                await waitUtils.waitForBackendWorkloadPhase(
                    distributedworkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
                logger.info(`Inference workload ${distributedworkloadName} is now in Ready status`);
            });

            test('T5404277:Check the metrics of runai inference workload by workload', async function () {
                await test.step('Verify NIM metrics for workload', async () => {
                    logger.info(`Testing metrics for workload: ${distributedworkloadId}`);
                    const finalMetricsResponse = await getWorkloadMetrics(distributedworkloadId, nimMetrics);
                    logger.info(`finalMetricsResponse: ${JSON.stringify(finalMetricsResponse, null, 2)}`);
                    verifyAllMetricsPresent(finalMetricsResponse, nimMetrics);
                    const isValid1 = metricUtils.validateNIMMetrics(finalMetricsResponse);
                    expect(isValid1).to.be.true;
                });
            });

            test('T5404278:Check the metrics of runai inference workload by pod', async function () {
                let podMetricsResponse1: any;
                let podMetricsResponse2: any;

                await test.step('Get NIM metrics for leader pod', async () => {
                    logger.info('Get NIM metrics for pod', distributedworkloadId);
                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(distributedworkloadId);
                    expect(pods).to.not.be.empty;
                    logger.info(
                        `pods: ${JSON.stringify(
                            pods.map((p: { id: string; name: string }) => ({ id: p.id, name: p.name })),
                            null,
                            2
                        )}`
                    );
                    // filter leader pods with a concise regex: end with -0 or -1 and the
                    // previous segment is not 0 or 1 (excludes workers like -0-1 / -1-1)
                    const targetPods = pods.filter((p: { name: string }) =>
                        /-(?![01]-)[^-]+-(0|1)$/.test(String(p.name))
                    );
                    logger.info(
                        `leader pods: ${JSON.stringify(
                            targetPods.map((p: any) => ({ id: p.id, name: p.name })),
                            null,
                            2
                        )}`
                    );

                    // ensure exactly two leader pods: -0 and -1
                    if (targetPods.length !== 2) {
                        throw new Error(
                            `Expected exactly 2 leader pods (-0 and -1), found ${targetPods.length}`
                        );
                    }
                    const first = targetPods[0];
                    const second = targetPods[1];

                    podMetricsResponse1 = await getPodMetrics(distributedworkloadId, first.id, podMetrics);
                    podMetricsResponse2 = await getPodMetrics(distributedworkloadId, second.id, podMetrics);
                    logger.info(`pod ${first.name} metrics: ${JSON.stringify(podMetricsResponse1, null, 2)}`);
                    logger.info(
                        `pod ${second.name} metrics: ${JSON.stringify(podMetricsResponse2, null, 2)}`
                    );

                    verifyAllMetricsPresent(podMetricsResponse1, podMetrics);
                    verifyAllMetricsPresent(podMetricsResponse2, podMetrics);

                    await test.step('Compare the metrics of workload and pod', async () => {
                        // Compare workload vs selected pods (referencing 2-replica checks)
                        const workloadMetricsResponse = await getWorkloadMetrics(
                            distributedworkloadId,
                            nimMetrics
                        );

                        // GPU cache average
                        const workloadGpuCache = parseFloat(
                            getMetricValue(workloadMetricsResponse, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                        );
                        const podGpuCache1 = parseFloat(
                            getMetricValue(podMetricsResponse1, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                        );
                        const podGpuCache2 = parseFloat(
                            getMetricValue(podMetricsResponse2, NIM_METRIC.GPU_CACHE_USAGE_PERC)
                        );
                        const podsGpuAvg = (podGpuCache1 + podGpuCache2) / 2;
                        logger.info(`GPU Cache - Workload: ${workloadGpuCache}, Expected avg: ${podsGpuAvg}`);
                        expect(workloadGpuCache).to.be.closeTo(podsGpuAvg, 4.0);

                        // Request success total sum
                        const workloadRequestSuccess = parseInt(
                            getMetricValue(workloadMetricsResponse, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                        );
                        const podRequestSuccess1 = parseInt(
                            getMetricValue(podMetricsResponse1, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                        );
                        const podRequestSuccess2 = parseInt(
                            getMetricValue(podMetricsResponse2, NIM_METRIC.REQUEST_SUCCESS_TOTAL)
                        );
                        const podsRequestSuccessSum = podRequestSuccess1 + podRequestSuccess2;
                        logger.info(
                            `Request Success - Workload: ${workloadRequestSuccess}, Expected sum: ${podsRequestSuccessSum}`
                        );
                        expect(workloadRequestSuccess).to.equal(podsRequestSuccessSum);
                    });
                });
            });
        });
    });
});
