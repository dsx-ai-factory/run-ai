/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    nodeUtils,
    tags,
    nodePoolUtils,
    flowRootState,
    workloadsCrudUtils,
    waitUtils,
    orgUnitUtils,
    workspaceUtils,
    inferenceUtils,
    trainingUtils,
    distributedUtils,
    policyUtils
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Project as OrgProject } from '../../../../../../modules/api/apis/orgUnitApi/generated/model/project';
import { bashCli } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { NodePool } from '../../../../../../modules/api/apis/backendClusterApi';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { verifyNodeExclusion } from '../../../../../../modules/helpers/Validations';
import { DistributedFramework, MatchExpressionOperator } from '../../../../../../modules/api/apis/policyApi';

test.describe(`Exclude Nodes API Tests, tags: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let workloadName: string;
        let workloadId: string;
        let workerNodeNames: string[];
        const nodePoolCreated: NodePool[] = [];

        const alpineImage = 'alpine:latest';
        const sleepCmd = 'sleep infinity';
        const nodeLabelKey = objectGenerator.generateShortNameByPrefix('test-ng', 5);
        const nodeLabelVal1 = 'group1';
        const nodeLabelVal2 = 'group2';

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-exclude-nodes', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            workerNodeNames = await nodeUtils.getWorkerNodeNames();
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateShortNameByPrefix('exclude-node-workload', 5);
        });

        test('T5167336: test workload API to create training workload with 1 node excluded', async function () {
            const excludedNodeNames = Utils.getRandomElements(workerNodeNames, 1);

            await test.step('Create training workload with 1 node excluded', async function () {
                const response = await trainingUtils.createTrainingV2(
                    project.id.toString(),
                    workloadName,
                    {
                        image: alpineImage,
                        command: sleepCmd
                    },
                    excludedNodeNames
                );

                expect(response.response.statusCode).toEqual(202);
                workloadId = response.body.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(workloadId)).body.actualPhase === Phase.Running,
                    `Wait for workload ${response.body.name} to be Running`,
                    true,
                    60
                );
            });

            await test.step('Verify workload is not running on excluded node', async function () {
                await verifyNodeExclusion(excludedNodeNames, namespace, workloadName);
            });

            await test.step('Delete workload', async function () {
                await workloadsCrudUtils.deleteTraining(workloadId);
            });
        });

        test('T5186250: test workload API to create distributed workload with 1 node excluded - worker & master with same config', async function () {
            const excludedNodeNames = Utils.getRandomElements(workerNodeNames, 1);

            await test.step('Create distributed workload with 1 node excluded', async function () {
                const response = await distributedUtils.createDistributedTraining(
                    project.id.toString(),
                    workloadName,
                    undefined,
                    undefined,
                    excludedNodeNames
                );

                expect(response.response.statusCode).toEqual(202);
                workloadId = response.body.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${response.body.name} to be Running`,
                    true,
                    60
                );
            });

            await test.step('Verify workload is not running on excluded node', async function () {
                await verifyNodeExclusion(excludedNodeNames, namespace, workloadName);
            });

            await test.step('Delete workload', async function () {
                await workloadsCrudUtils.deleteDistributed(workloadId);
            });
        });

        test('T5186251: test workload API to create distributed workload with 1 node excluded - worker & master with different config', async function () {
            const workerExcludedNodeNames = workerNodeNames.slice(0, 1);
            const masterExcludedNodeNames = workerNodeNames.slice(1, 2);

            await test.step('Create distributed workload with 1 node excluded', async function () {
                const response = await distributedUtils.createDistributedTraining(
                    project.id.toString(),
                    workloadName,
                    undefined,
                    {
                        distributedFramework: DistributedFramework.Mpi,
                        image: config.quickStartImage.ubuntu,
                        command: sleepCmd,
                        compute: {
                            gpuDevicesRequest: 0,
                            cpuCoreRequest: 0.2,
                            cpuMemoryRequest: '200M'
                        }
                    },
                    workerExcludedNodeNames,
                    masterExcludedNodeNames
                );

                expect(response.response.statusCode).toEqual(202);
                workloadId = response.body.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${response.body.name} to be Running`,
                    true,
                    60
                );
            });

            await test.step('Verify that the worker and master pods are not running on their respective excluded nodes', async function () {
                const configs = [
                    {
                        excludeNodeNames: workerExcludedNodeNames,
                        workloadName: `${workloadName}-worker`
                    },
                    {
                        excludeNodeNames: masterExcludedNodeNames,
                        workloadName: `${workloadName}-launcher`
                    }
                ];

                await Promise.all(
                    configs.map((config) =>
                        verifyNodeExclusion(config.excludeNodeNames, namespace, config.workloadName)
                    )
                );
            });

            await test.step('Delete workload', async function () {
                await workloadsCrudUtils.deleteDistributed(workloadId);
            });
        });

        test('T5167338: test workload API to create inference workload with 2 nodes excluded', async function () {
            const excludedNodeNames = Utils.getRandomElements(workerNodeNames, 2);
            const inferenceSpec = {
                image: config.quickStartImage.inferenceHelloWorld,
                servingPort: { container: 8033, protocol: ServingPortProtocol.Http }
            };
            inferenceSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                MatchExpressionOperator.NotIn,
                excludedNodeNames
            ).nodeAffinityRequired;
            await test.step('Create inference workload with 1 node excluded', async function () {
                [workloadId, workloadName] = await inferenceUtils.createInference(
                    workloadName,
                    inferenceSpec,
                    project.id.toString()
                );

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getInference(workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workloadName} to be Running`,
                    true,
                    60
                );
            });

            await test.step('Verify workload is not running on excluded node', async function () {
                await verifyNodeExclusion(excludedNodeNames, namespace, workloadName);
            });

            await test.step('Delete workload', async function () {
                await workloadsCrudUtils.deleteInference(workloadId);
            });
        });

        test.describe('Test workload API to create workspace workload with combined terms and node pools', () => {
            let npProject: OrgProject;
            let nodePoolNames: string[] = [];
            let nodeGroup1: string[];
            let nodeGroup2: string[];
            let excludedNodeNames: string[];

            const npProjectName = objectGenerator.generateShortNameByPrefix('proj-np', 5);
            const npNamespace = Utils.namespaceFromProjectName(npProjectName);

            test.beforeAll(async function () {
                excludedNodeNames = Utils.getRandomElements(workerNodeNames, 1);
                const mid = Math.floor(workerNodeNames.length / 2);
                nodeGroup1 = workerNodeNames.slice(0, mid);
                nodeGroup2 = workerNodeNames.slice(mid);

                // label nodes
                const nodeGroupsWithLabels: [string[], string][] = [
                    [nodeGroup1, nodeLabelVal1],
                    [nodeGroup2, nodeLabelVal2]
                ];
                nodeGroupsWithLabels.forEach(([group, labelVal]) => {
                    bashCli.exec(
                        `kubectl label node ${group.join(' ')} ${nodeLabelKey}=${labelVal} --overwrite`
                    );
                });

                // // create 2 node pools
                const poolConfigs = [{ group: nodeLabelVal1 }, { group: nodeLabelVal2 }];
                const nodePools = await Promise.all(
                    poolConfigs.map((cfg) =>
                        nodePoolUtils.createNodePool(
                            objectGenerator.generateShortNameByPrefix('np', 5),
                            nodeLabelKey,
                            cfg.group
                        )
                    )
                );
                nodePoolCreated.push(...nodePools);
                expect(nodePools.length).toEqual(2);

                nodePoolNames = nodePools.map((np) => np.name);

                const res = nodePools.map((np) => ({
                    nodePool: {
                        name: np.name,
                        id: np.id.toString()
                    },
                    gpu: {
                        deserved: 0,
                        limit: -1
                    }
                }));

                res.push({
                    nodePool: {
                        name: 'default',
                        id: '100000'
                    },
                    gpu: {
                        deserved: 0,
                        limit: -1
                    }
                });

                npProject = await orgUnitUtils.createProjectAndValidate({
                    clusterId: flowRootState.cluster.uuid,
                    resources: res,
                    name: npProjectName,
                    defaultNodePools: nodePoolNames
                });
            });

            test.afterAll('clean up test data', async function () {
                try {
                    await operations.deleteProjectAndNamespace(parseInt(npProject.id));
                    logger.info(`deleted project ${npProjectName} and namespace ${npNamespace}`);
                } catch (error) {
                    logger.error(
                        `failed to delete project ${npProjectName} and namespace ${npNamespace}`,
                        error
                    );
                }

                await Promise.all(
                    nodePoolCreated.map(async (nodePool) => {
                        try {
                            await nodePoolUtils.deleteNodePool(nodePool);
                        } catch (error) {
                            logger.error(error);
                        }
                    })
                );

                [nodeGroup1, nodeGroup2].forEach((group) => {
                    try {
                        bashCli.exec(`kubectl label node ${group.join(' ')} ${nodeLabelKey}-`);
                    } catch (error) {
                        logger.error(error);
                    }
                });
            });

            test('T5167337: test workload API to create workspace workload with and exclusion terms', async function () {
                await test.step('Create workspace workload with exclusion terms', async function () {
                    const response = await workspaceUtils.createWorkspace(
                        npProject.id.toString(),
                        workloadName,
                        {
                            image: alpineImage,
                            command: sleepCmd
                        },
                        excludedNodeNames
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;

                    await waitUtils.waitForCondition(
                        async () =>
                            (await workloadsCrudUtils.getWorkspace(workloadId)).body.actualPhase ===
                            Phase.Running,
                        `Wait for workload ${response.body.name} to be Running`,
                        true,
                        60
                    );
                });

                await test.step('Verify workload is not running on excluded node', async function () {
                    await verifyNodeExclusion(excludedNodeNames, npNamespace, workloadName);
                });

                await test.step('Delete workload', async function () {
                    await workloadsCrudUtils.deleteWorkspace(workloadId);
                });
            });
        });
    });
});
