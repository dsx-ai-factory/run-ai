/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    flowRootState,
    config,
    workloadsUtilsV2,
    metricUtils,
    settings,
    gpuTypes
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { StatusCodes } from 'http-status-codes';
import { PodMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`Advanced metrics Tests(GPU), tags: ${tags.resource.metrics} ${tags.teams.loong}`, () => {
    test.setTimeout(10 * 60 * 1000);
    shouldRunTests(settings.gpuType === gpuTypes.real)(function () {
        supportedFrom(SupportedVersions.V2_23)(function () {
            let projectName: string;
            let namespace: string;
            let project: Project;
            let workloadName: string;
            let workloadId: string;
            let podId: string;
            let allMetricsValid = false;

            const numberOfSamples = 5;

            test.beforeAll(async function () {
                // patch 'runaiconfig' to enable advanced metrics
                const result = await kubeCli.patchRunaiConfig([
                    {
                        op: 'add',
                        path: '/spec/prometheus',
                        value: {
                            config: {
                                advancedMetricsEnabled: true
                            }
                        }
                    }
                ]);
                expect(result).toBeTruthy();

                projectName = objectGenerator.generateShortNameByPrefix('proj-advanced-metrics', 5);
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
            });

            test.afterAll('deleting project and namespace', async () => {
                // patch 'runaiconfig' to disable advanced metrics
                try {
                    await kubeCli.patchRunaiConfig([
                        {
                            op: 'remove',
                            path: '/spec/prometheus'
                        }
                    ]);
                } catch (error) {
                    logger.error('failed to disable advanced metrics', error);
                }

                try {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                } catch (error) {
                    logger.error(`failed to delete workload ${workloadName}`, error);
                }

                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                    throw error;
                }
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateShortNameByPrefix('gpu-training', 5);
            });

            test('T4062438: check advanced metrics data', async function () {
                // create a GPU training using the image 'config.quickStartImage.training' defined in runai cnofig file
                // the image downloading time varies from 2-5 minutes.
                await test.step('create a GPU training workload', async () => {
                    workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                    const response = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.training,
                            compute: {
                                cpuCoreRequest: 0.2,
                                cpuMemoryRequest: '200M',
                                gpuDevicesRequest: 2
                            }
                        }
                    );

                    expect(response.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                    workloadId = response.body.workloadId;
                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        300
                    );

                    podId = (await workloadsUtilsV2.getWorkloadPods(workloadId)).body.pods[0].id;
                });

                await test.step('verify the advanced metrics', async () => {
                    const startTime: Date = new Date();
                    let endTime: Date = new Date();
                    const timeout = 5 * 60 * 1000;
                    const startTimeMillis = startTime.getTime();

                    const advancedMetrics = [
                        {
                            name: PodMetricType.GpuMemoryBandwidthUtilizationPerGpu,
                            matcher: (value: number): void => {
                                expect(value).toBeGreaterThan(0);
                                expect(value).toBeLessThanOrEqual(100);
                            }
                        },
                        {
                            name: PodMetricType.GpuPcieTransmittedBandwidthPerGpu,
                            matcher: (value: number): void => expect(value).toBeGreaterThan(0)
                        },
                        {
                            name: PodMetricType.GpuPcieReceivedBandwidthPerGpu,
                            matcher: (value: number): void => expect(value).toBeGreaterThan(0)
                        },
                        {
                            name: PodMetricType.GpuGraphicsEngineActivityPerGpu,
                            matcher: (value: number): void => {
                                expect(value).toBeGreaterThanOrEqual(0);
                                expect(value).toBeLessThanOrEqual(100);
                            }
                        },
                        {
                            name: PodMetricType.GpuTensorActivityPerGpu,
                            matcher: (value: number): void => expect(value).toBeGreaterThanOrEqual(0)
                        }
                    ];

                    // Helper function to check metrics
                    async function checkMetrics(): Promise<boolean> {
                        for (const metrics of advancedMetrics) {
                            const metricsResponse = (
                                await metricUtils.getPodMetrics(
                                    workloadId,
                                    podId,
                                    [metrics.name],
                                    startTime,
                                    endTime,
                                    numberOfSamples
                                )
                            ).body;
                            const measurements = metricsResponse.measurements;

                            for (const measurement of measurements) {
                                if (!measurement.values || measurement.values.length === 0) {
                                    return false;
                                }
                                measurement.values.forEach((value) =>
                                    metrics.matcher(parseFloat(value.value))
                                );
                            }
                        }
                        return true;
                    }

                    // Loop until all metrics are valid or timeout
                    while (endTime.getTime() - startTimeMillis < timeout) {
                        allMetricsValid = await checkMetrics();
                        if (allMetricsValid) {
                            break;
                        }

                        // Wait for 10 seconds before the next check
                        await new Promise((resolve) => setTimeout(resolve, 10000));
                        endTime = new Date();
                    }

                    expect(allMetricsValid).toBeTruthy();
                });
            });
        });
    });
});
