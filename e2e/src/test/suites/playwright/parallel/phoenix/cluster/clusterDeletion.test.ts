/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    clusterService,
    clusterUtils,
    flowRootState,
    logger,
    objectGenerator,
    projectResourcesUtils,
    settings,
    tags,
    workloadUtils,
    TrainingWorkload,
    waitUtils,
    nodePoolUtils,
    orgUnitUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { BaseWorkload } from '../../../../../../modules/helpers/models/baseWorkload';
import { WorkloadsUtilsV2 } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { ClusterState } from '../../../../../../modules/api/apis/clusterApi';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
const workloadsUtilsV2 = new WorkloadsUtilsV2();

shouldRunNativeWorkloadCrdTests(() => {
    test.describe(
        `Cluster Deletion Tests, tags:${tags.runOn.clusterDeletion} ${tags.teams.phoenix} ${tags.services.clusterService}`,
        { tag: [tags.services.clusterService] },
        () => {
    test.describe.configure({ retries: 0, timeout: 360_000 });
    test.describe('cluster deletion workflow', () => {
        test.skip(!settings.isClusterDeletionWorkflow, 'Requires a cluster deletion environment');
        let testCluster: any;
        let testProject: Project;
        let testWorkload: BaseWorkload;
        let projectName: string;
        let namespaceName: string;
        let workloadName: string;
        let nodePool: any;

        test.beforeAll(async () => {
            // Clean up any existing disconnected clusters
            await clusterUtils.deleteDisconnectedClusters();

            // Get the existing connected cluster
            const clusters = await clusterService.getClusters(
                flowRootState.users.system_administrator.accessToken
            );
                    const connectedClusters = clusters.filter(
                        (c) => c.status.state === ClusterState.Connected
                    );
            if (connectedClusters.length === 0) {
                throw new Error('No connected clusters found in the environment');
            }
            testCluster = connectedClusters[0];

            // Create a node pool in the cluster
            const nodePoolName = objectGenerator.generateNameByPrefix('nodepool');
            nodePool = await nodePoolUtils.createNodePool(
                nodePoolName,
                'run.ai/type',
                'test',
                flowRootState.users.system_administrator.accessToken
            );

            // Create a project in the cluster
            projectName = objectGenerator.generateNameByPrefix('project');
            namespaceName = `runai-${projectName}`;
            testProject = await projectResourcesUtils.createProject(
                namespaceName,
                {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                },
                undefined,
                true,
                false,
                undefined,
                testCluster.uuid
            );

            // Create a workload in the project
            workloadName = objectGenerator.generateNameByPrefix('workload');
            const workload = new TrainingWorkload('ubuntu', namespaceName, workloadName);
            workload.command = 'sleep infinity';
            await workloadUtils.submit(workload);
            testWorkload = workload;

            // Wait for workload to be created
            await workloadUtils.waitForTrainingWorkloadCrd(workloadName, namespaceName);

            // Delete the cluster
            await clusterUtils.deleteCluster(
                flowRootState.users.system_administrator.accessToken,
                testCluster.uuid
            );
        });

        test.afterAll(async () => {
            // Clean up any remaining resources
            try {
                await clusterUtils.deleteDisconnectedClusters();
            } catch (error) {
                logger.error(`Failed to clean up clusters: ${error}`);
            }
        });

        test('should verify cluster, project, and workload are deleted', async () => {
            // Verify all resources are deleted in parallel
            await Promise.all([
                // Verify cluster is deleted
                waitUtils.waitForCondition(
                    async () => {
                        const clusters = await clusterService.getClusters(
                            flowRootState.users.system_administrator.accessToken
                        );
                        logger.info(`Clusters: ${JSON.stringify(clusters)}`);
                        return !clusters.some((c) => c.uuid === testCluster.uuid);
                    },
                    'Waiting for cluster to be deleted',
                    true,
                    300,
                    10
                ),

                // Wait for workload to be deleted
                waitUtils.waitForCondition(
                    async () => {
                        const workloads = await workloadsUtilsV2.getWorkloadsByProjectId(
                            testProject.id.toString()
                        );
                        logger.info(`Workloads: ${JSON.stringify(workloads.body.workloads)}`);
                        return !workloads.body.workloads.some((w) => w.name === workloadName);
                    },
                    'Waiting for workload to be deleted',
                    true,
                    300,
                    10
                ),

                // Wait for project to be deleted
                waitUtils.waitForCondition(
                    async () => {
                        const projects = await orgUnitUtils.getProjectsByClusterId(testCluster.uuid);
                                logger.info(
                                    `Projects for cluster ${testCluster.uuid}: ${JSON.stringify(projects)}`
                                );
                        return !projects.some((p) => p.name === projectName);
                    },
                    'Waiting for project to be deleted',
                    true,
                    300,
                    10
                )
            ]);
        });
    });
        }
    );
});
