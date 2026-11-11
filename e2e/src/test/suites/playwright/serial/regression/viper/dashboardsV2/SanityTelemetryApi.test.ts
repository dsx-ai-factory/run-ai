/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    nodeUtils,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    tags,
    telemetryUtils,
    Utils,
    versionUtils,
    workloadsUtilsV2,
    waitUtils
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import { submitTrainWorkload } from '../../../../../../../modules/helpers/models/workloads';
import { WorkloadTelemetryType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadTelemetryType';
import { test, expect } from '@playwright/test';
import { NodeTelemetryType } from '../../../../../../../modules/api/apis/clusterApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Sanity Telemetry API, tags:  ${tags.resource.metrics} ${tags.teams.viper}`, () => {
    let projects, workloads, gpusPerWorkload, numOfGpus, ghostWorkloads, ghostWorkloadsGroupedByProject;
    const numOfProjects = 3;
    const numOfWorkloads = 3;

    test.beforeAll(async () => {
        // clean the env before starting the test
        await operations.forceCleanup();
        numOfGpus = await kubeCli.getNumOfGpus();
    });

    supportedFrom(SupportedVersions.V2_15)(() => {
        test.describe('Node telemetry', () => {
            test('Should get node telemetry - ReadyGpus', async () => {
                logger.debug('Get ReadyGpus node telemetry');
                const readyGpusTelemetryRes = await telemetryUtils.getNodeTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    NodeTelemetryType.ReadyGpus
                );
                const readyGpus = await nodeUtils.getAllReadyGpusCount();
                expect(readyGpusTelemetryRes).toBeTruthy();
                expect(readyGpusTelemetryRes.type).toBe(NodeTelemetryType.ReadyGpus);
                expect(readyGpusTelemetryRes.values.length).toBe(1);
                expect(readyGpusTelemetryRes.values[0].value).toBe(readyGpus.toString());
                // We assume that the number of ready gpus is equal to the number of gpus in the cluster
                expect(readyGpusTelemetryRes.values[0].value).toBe(numOfGpus.toString());
            });

            test('Should get node telemetry - TotalGpus & TotalGpuNodes', async () => {
                logger.debug('Get TotalGpus node telemetry');
                const totalGpusTelemetryRes = await telemetryUtils.getNodeTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    NodeTelemetryType.TotalGpus
                );
                expect(totalGpusTelemetryRes).toBeTruthy();
                expect(totalGpusTelemetryRes.type).toBe(NodeTelemetryType.TotalGpus);
                expect(totalGpusTelemetryRes.values.length).toBe(1);
                expect(totalGpusTelemetryRes.values[0].value).toBe(numOfGpus.toString());
                if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.15')) {
                    const totalGpus = await nodeUtils.getAllGpusCount();
                    expect(totalGpusTelemetryRes.values[0].value).toBe(totalGpus.toString());
                }

                logger.debug('Get TotalGpuNodes node telemetry');
                const totalGpuNodesTelemetryRes = await telemetryUtils.getNodeTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    NodeTelemetryType.TotalGpuNodes
                );
                expect(totalGpuNodesTelemetryRes).toBeTruthy();
                expect(totalGpuNodesTelemetryRes.type).toBe(NodeTelemetryType.TotalGpuNodes);
                const totalGpuNodesCount = await nodeUtils.getNodesWithGpusCount();
                expect(totalGpuNodesTelemetryRes.values.length).toBe(1);
                expect(totalGpuNodesTelemetryRes.values[0].value).toBe(totalGpuNodesCount.toString());
            });
        });
    });

    shouldRunNativeWorkloadCrdTests(() =>
        test.describe('Workload telemetry', () => {
            let runningWorkloadsInBackend;

            test.beforeAll(async () => {
                ghostWorkloads = await telemetryUtils.waitForWorkloadTelemetryStabilization(
                    flowRootState.cluster.uuid,
                    WorkloadTelemetryType.WorkloadsCount
                );
                logger.debug(`Ghost workloads baseline: ${ghostWorkloads.values[0]?.value}`);

                ghostWorkloadsGroupedByProject = await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                    WorkloadTelemetryType.WorkloadsCount,
                    ['ProjectId']
                );

                let projectName, namespace;
                const deservedGpus = Math.floor(numOfGpus / numOfProjects);
                const projectPromises = [];
                for (let i = 0; i < numOfProjects; i++) {
                    projectName = objectGenerator.generateNameByPrefix(`project-${i + 1}-for-telemetry`);
                    namespace = Utils.namespaceFromProjectName(projectName);
                    projectPromises.push(
                        projectResourcesUtils.createProject(namespace, {
                            name: projectName,
                            deservedGpus,
                            maxAllowedGpus: deservedGpus
                        })
                    );
                }
                projects = await Promise.all(projectPromises);
                gpusPerWorkload = Math.floor(deservedGpus / numOfWorkloads);

                const workloadsPromises = [];
                for (let i = 0; i < numOfProjects; i++) {
                    for (let j = 0; j < numOfWorkloads; j++) {
                        workloadsPromises.push(
                            submitTrainWorkload(
                                `workload-${i + 1}-${j + 1}-for-telemetry`,
                                projects[i].name,
                                config.quickStartImage.ubuntu,
                                'sleep infinity',
                                gpusPerWorkload,
                                PodStatus.PENDING
                            )
                        );
                    }
                }
                workloads = await Promise.all(workloadsPromises);
            });

            test.afterAll(async () => {
                try {
                    await operations.deleteWorkloads(workloads);
                } catch (error) {
                    logger.error('Failed to delete workloads:', error);
                }

                await Promise.all(
                    projects.map((project) => operations.deleteProjectAndNamespace(project.id))
                );
            });

            test.beforeAll(async () => {
                // wait for at least one workload to be running
                await workloadsUtilsV2.waitForOneWorkloadStatus(workloads, PodStatus.RUNNING);

                // Verify number of running workloads for minimum expected allocated gpus
                runningWorkloadsInBackend = await workloadsUtilsV2.getNumberOfWorkloadsInPhase(
                    Phase.Running,
                    flowRootState.cluster.uuid
                );

                logger.debug(`Number of running workloads: ${runningWorkloadsInBackend}`);
            });

            test('Should get workload telemetry, filtered by clusterId', async () => {
                const expectedWorkloadCount = numOfProjects * numOfWorkloads;
                const ghostCount = Number(ghostWorkloads.values[0]?.value ?? 0);

                // Poll until telemetry reflects the expected workload count (eventual consistency)
                logger.debug(
                    `Waiting for WorkloadsCount telemetry to reach ${expectedWorkloadCount} ` +
                        `(ghost baseline: ${ghostCount})`
                );
                await waitUtils.waitForCondition(
                    async () => {
                        const res = await telemetryUtils.getWorkloadTelemetryFilteredByClusterId(
                            flowRootState.cluster.uuid,
                            WorkloadTelemetryType.WorkloadsCount
                        );
                        const totalCount = Number(res.values[0]?.value ?? 0);
                        const diff = totalCount - ghostCount;
                        logger.debug(
                            `WorkloadsCount telemetry: total=${totalCount}, ghost=${ghostCount}, diff=${diff}, expected=${expectedWorkloadCount}`
                        );
                        return diff >= expectedWorkloadCount;
                    },
                    `workload telemetry count to reach expected ${expectedWorkloadCount}`,
                    true,
                    60,
                    5
                );

                // Final assertion after stabilization
                const workloadCountRes = await telemetryUtils.getWorkloadTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    WorkloadTelemetryType.WorkloadsCount
                );
                expect(workloadCountRes).toBeTruthy();
                expect(workloadCountRes.type).toBe(WorkloadTelemetryType.WorkloadsCount);
                expect(workloadCountRes.values.length).toBe(1);
                const actualDiff = Number(workloadCountRes.values[0].value) - ghostCount;
                expect(actualDiff).toBeGreaterThanOrEqual(expectedWorkloadCount);

                logger.debug('Get GpuAllocation workload telemetry, filtered by clusterId');
                await telemetryUtils.waitForMinWorkloadTelemetryValue(
                    WorkloadTelemetryType.GpuAllocation,
                    runningWorkloadsInBackend * gpusPerWorkload
                );
                const gpuAllocationRes = await telemetryUtils.getWorkloadTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    WorkloadTelemetryType.GpuAllocation
                );
                expect(gpuAllocationRes).toBeTruthy();
                expect(gpuAllocationRes.type).toBe(WorkloadTelemetryType.GpuAllocation);
                expect(gpuAllocationRes.values.length).toBe(1);
                expect(Number(gpuAllocationRes.values[0].value)).toBeGreaterThanOrEqual(
                    runningWorkloadsInBackend * gpusPerWorkload
                );
            });

            test('Should get workload telemetry, using groupBy', async () => {
                logger.debug('Get WorkloadsCount workload telemetry, using groupBy ProjectId');
                const workloadCountGroupByProjectIdRes =
                    await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.WorkloadsCount,
                        ['ProjectId']
                    );
                if (ghostWorkloadsGroupedByProject.values.length !== 0) {
                    workloadCountGroupByProjectIdRes.values = workloadCountGroupByProjectIdRes.values.filter(
                        (telemetryResponseValuesInner) =>
                            !ghostWorkloadsGroupedByProject.values.some((ghostWorkload) =>
                                telemetryResponseValuesInner?.groups?.some(
                                    (group) =>
                                        group.key === ghostWorkload.groups[0].key &&
                                        group.value === ghostWorkload.groups[0].value
                                )
                            )
                    );
                }
                expect(workloadCountGroupByProjectIdRes).toBeTruthy();
                expect(workloadCountGroupByProjectIdRes.type).toBe(WorkloadTelemetryType.WorkloadsCount);
                expect(workloadCountGroupByProjectIdRes.values.length).toBe(numOfProjects);
                for (const value of workloadCountGroupByProjectIdRes.values) {
                    expect(value.value).toBe(numOfWorkloads.toString());
                }

                logger.debug('Get GpuAllocation workload telemetry, using groupBy ProjectId');
                const gpuAllocationGroupByProjectIdRes =
                    await telemetryUtils.getWorkloadTelemetryUsingGroupBy(
                        WorkloadTelemetryType.GpuAllocation,
                        ['ProjectId']
                    );
                if (ghostWorkloadsGroupedByProject.values.length !== 0) {
                    gpuAllocationGroupByProjectIdRes.values = gpuAllocationGroupByProjectIdRes.values.filter(
                        (telemetryResponseValuesInner) =>
                            !ghostWorkloadsGroupedByProject.values.some((ghostWorkload) =>
                                telemetryResponseValuesInner?.groups?.some(
                                    (group) =>
                                        group.key === ghostWorkload.groups[0].key &&
                                        group.value === ghostWorkload.groups[0].value
                                )
                            )
                    );
                }
                expect(gpuAllocationGroupByProjectIdRes).toBeTruthy();
                expect(gpuAllocationGroupByProjectIdRes.type).toBe(WorkloadTelemetryType.GpuAllocation);
                expect(gpuAllocationGroupByProjectIdRes.values.length).toBe(numOfProjects);
                for (const value of gpuAllocationGroupByProjectIdRes.values) {
                    expect([0, gpusPerWorkload * 1, gpusPerWorkload * 2, gpusPerWorkload * 3]).toContain(
                        Number(value.value)
                    );
                }
            });
        })
    );
});
