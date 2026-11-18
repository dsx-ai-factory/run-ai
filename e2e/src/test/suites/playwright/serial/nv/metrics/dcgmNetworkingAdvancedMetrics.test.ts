/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    nodeUtils,
    config,
    waitUtils,
    workloadsUtilsV2,
    workloadsCrudUtils,
    flowRootState
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Phase, WorkloadMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { Node, NodeMetricType } from '../../../../../../modules/api/apis/clusterApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { workloadsCrudService } from '../../../../../../modules/api';

test.describe(`DCGM Netowrking Advanced Metrics API Tests, tags: ${tags.tool.api} ${tags.resource.metrics} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let nodes: Node[];
        let nodeName: string;
        let supported: boolean = false;
        let workloadId: string;
        const projectName = objectGenerator.generateShortNameByPrefix('proj-metrics', 5);
        const namespace = Utils.namespaceFromProjectName(projectName);
        const workloadName: string = objectGenerator.generateShortNameByPrefix('workload-metrics', 5);
        const gpuTypes = ['A100', 'GB200'];

        test.beforeAll(async () => {
            for (const gpuType of gpuTypes) {
                nodes = await nodeUtils.getGpuNodes(gpuType);
                if (nodes.length > 0) {
                    supported = true;
                    break;
                }
            }

            // if the GPU types are not found in the env, then skip the tests
            if (!supported) {
                test.skip();
            }

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 2
            });

            // create a workload for testing
            workloadsCrudService.trainingsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            const { body: training } = await workloadsCrudUtils.createTraining(
                workloadName,
                false,
                flowRootState.cluster.uuid,
                project.id.toString(),
                {
                    image: config.quickStartImage.metricsGpu,
                    compute: { gpuDevicesRequest: 1 }
                }
            );
            workloadId = training.workloadId;

            await workloadsCrudUtils.waitForWorkloadPhase(
                training.workloadId,
                Phase.Running,
                WorkloadType.TRAINING,
                undefined,
                120
            );

            const pods = await kubeCli.listPodsByNamepsace(namespace);
            // only get workload pods in current test
            const workloadPods = pods.filter((pod) => pod.metadata.name.includes(workloadName));

            nodeName = workloadPods[0].spec?.nodeName;
        });

        test.afterAll(async () => {
            try {
                await workloadsCrudUtils.deleteTraining(workloadId);
            } catch (error) {
                logger.error(`failed to delete workload ${workloadName}`);
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
            }
        });

        test('T5419579: test workload NVLINK_BANDWIDTH_TOTAL metrics', async () => {
            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
            await waitUtils.waitForPositiveWorkloadMetricsValue(
                workloadId,
                WorkloadMetricType.NvlinkBandwidthTotal
            );
        });

        test('T5419580: test node NVLINK_BANDWIDTH_TOTAL metrics', async () => {
            test.setTimeout(10 * 60 * 1000);
            const nodeId = nodes.find((node) => node.name === nodeName)?.id;
            await waitUtils.waitForPositiveNodeMetricsValue(nodeId, NodeMetricType.NvlinkBandwidthTotal);

            const now = new Date();
            const start = new Date(now.getTime() - 10 * 60 * 1000);
            const end = now;

            // in response, only 1 NvlinkBandwidthTotal object along with multiple NvlinkBandwidthTotalPerGpu objects(diff GPUs)
            const resp = await nodeUtils.getNodeMetrics(
                nodeId,
                [NodeMetricType.NvlinkBandwidthTotal, NodeMetricType.NvlinkBandwidthTotalPerGpu],
                start,
                end
            );

            const measurements = resp.body.measurements;

            // aggregate NvlinkBandwidthTotalPerGpu data with same timestamps
            const result: Record<string, Record<string, number>> = {};
            for (const m of measurements) {
                if (!result[m.type]) {
                    result[m.type] = {};
                }

                for (const v of m.values) {
                    const ts = String(v.timestamp);
                    const val = parseFloat(v.value);

                    result[m.type][ts] = (result[m.type][ts] ?? 0) + val;
                }
            }

            // use 'toEqual' to deep compare
            expect(result[NodeMetricType.NvlinkBandwidthTotal]).toEqual(
                result[NodeMetricType.NvlinkBandwidthTotalPerGpu]
            );
        });
    });
});
