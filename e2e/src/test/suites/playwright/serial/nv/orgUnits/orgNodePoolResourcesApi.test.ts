/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    logger,
    objectGenerator,
    tags,
    flowRootState,
    nodeUtils,
    nodePoolUtils,
    orgUnitUtils,
    operations,
    apiUtils,
    departmentUtils,
    crdUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { NodepoolCreateFields, NodepoolPhase } from '../../../../../../modules/api/apis/clusterApi';
import { kubeCli } from '../../../../../../modules/cli';
import { StatusCodes } from 'http-status-codes';
import {
    BatchPatchNodePoolResourcesRequest,
    DepartmentCreationRequest,
    ResourcesPerNodePool,
    Units
} from '../../../../../../modules/api/apis/orgUnitApi';

test.describe(`Node Pool Resources API Tests, tags: ${tags.tool.api} ${tags.resource.nodepools} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let gpuWorkerNodeNames: string[];
        let totalGpuNodeCnt: number;
        let nodeToLabel: string;
        let defaultNpId: string;
        let npId: string;
        let defaultDeptId: string;
        let deptId: string;

        const createdNodePools: { id: string; name: string }[] = [];
        const nodeLabelKey = 'test';
        const nodeLabelVal1 = 'proj-test-group1';
        const npName = objectGenerator.generateShortNameByPrefix('np1', 5);
        let deptIds: number[];
        let projectIds: string[];

        const gpuQuota = 2;
        const gpuLimit = 4;
        const cpuQuota = 4000;
        const cpuLimit = 6000;
        const memoryQuota = 8096;
        const memoryLimit = 10240;
        const newGpuQuota = 1;
        const newGpuLimit = 2;
        const newCpuQuota = 2000;
        const newCpuLimit = 4000;
        const newMemoryQuota = 4048;
        const newMemoryLimit = 8096;

        const createResourceConfig = (
            gpuVal: number,
            gpuLimitVal: number,
            cpuVal: number,
            cpuLimitVal: number,
            memoryVal: number,
            memoryLimitVal: number
        ): ResourcesPerNodePool => {
            return {
                gpu: {
                    deserved: gpuVal,
                    limit: gpuLimitVal,
                    overQuotaWeight: gpuVal
                },
                cpu: {
                    deserved: cpuVal,
                    limit: cpuLimitVal,
                    overQuotaWeight: cpuVal
                },
                memory: {
                    deserved: memoryVal,
                    limit: memoryLimitVal,
                    overQuotaWeight: memoryVal,
                    units: Units.Mb
                }
            };
        };

        const validateResourceConfig = (
            actualResources: ResourcesPerNodePool,
            expectedGpu: number,
            expectedGpuLimit: number,
            expectedCpu: number,
            expectedCpuLimit: number,
            expectedMemory: number,
            expectedMemoryLimit: number
        ): void => {
            expect(actualResources.gpu).toEqual({
                deserved: expectedGpu,
                limit: expectedGpuLimit,
                overQuotaWeight: expectedGpu
            });
            expect(actualResources.cpu).toEqual({
                deserved: expectedCpu,
                limit: expectedCpuLimit,
                overQuotaWeight: expectedCpu
            });
            expect(actualResources.memory).toEqual({
                deserved: expectedMemory,
                limit: expectedMemoryLimit,
                overQuotaWeight: expectedMemory,
                units: Units.Mb
            });
        };

        const createCrdResourceConfig = (
            gpuVal: number,
            gpuLimitVal: number,
            cpuVal: number,
            cpuLimitVal: number,
            memoryVal: number,
            memoryLimitVal: number
        ): {
            gpu: { quota: number; limit: number; overQuotaWeight: number };
            cpu: { quota: number; limit: number; overQuotaWeight: number };
            memory: { quota: number; limit: number; overQuotaWeight: number };
        } => {
            return {
                gpu: {
                    quota: gpuVal,
                    limit: gpuLimitVal,
                    overQuotaWeight: gpuVal
                },
                cpu: {
                    quota: cpuVal,
                    limit: cpuLimitVal,
                    overQuotaWeight: cpuVal
                },
                memory: {
                    quota: memoryVal,
                    limit: memoryLimitVal,
                    overQuotaWeight: memoryVal
                }
            };
        };

        test.beforeAll(async () => {
            await apiUtils.enableCpuQuotaFeature();
            await apiUtils.enableProjectOverQuotaFeature();

            defaultDeptId = (await orgUnitUtils.getDefaultDepartment()).id;
            defaultNpId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

            // get gpu node list
            gpuWorkerNodeNames = await nodeUtils.getGpuWorkerNodeNames();
            totalGpuNodeCnt = gpuWorkerNodeNames.length;

            expect(totalGpuNodeCnt).toBeGreaterThan(0);

            // label a gpu node for node pool
            nodeToLabel = gpuWorkerNodeNames[0];
            kubeCli.labelNode(nodeToLabel, nodeLabelKey, nodeLabelVal1);

            // create node pool 1 includes 1 gpu node
            const npReq1 = Object.assign(new NodepoolCreateFields(), {
                name: npName,
                labelKey: nodeLabelKey,
                labelValue: nodeLabelVal1,
                clusterId: flowRootState.cluster.uuid
            });

            await nodePoolUtils.createClusterNodePoolWithRequest(npReq1).then((rsp) => {
                createdNodePools.push({ id: rsp.body.id, name: npName });
                npId = rsp.body.id;
                return rsp;
            });

            await nodePoolUtils.waitForClusterNodePoolPhase(npId, [NodepoolPhase.Ready], 60);
        });

        test.afterAll('clean up', async () => {
            await apiUtils.disableCpuQuotaFeature();
            await apiUtils.disableProjectOverQuotaFeature();

            // Unlabel nodes
            try {
                if (gpuWorkerNodeNames && gpuWorkerNodeNames.length > 0) {
                    kubeCli.unLabelNode(nodeToLabel, nodeLabelKey);
                }
            } catch (e) {
                logger.error(`Failed to unlabel node ${nodeToLabel}`, e);
            }

            // Delete node pools
            await Promise.all(
                createdNodePools.map(async (np) => {
                    try {
                        await nodePoolUtils.deleteNodePoolByIdWithWait(np.id, np.name);
                    } catch (e) {
                        logger.error(`Failed to delete node pool ${np.id}`, e);
                    }
                })
            );
        });

        test.describe('Department nodepool resources API tests', () => {
            const deptName = objectGenerator.generateShortNameByPrefix('dept-res-api', 3);

            test.beforeAll('create a department for testing', async () => {
                deptIds = [];
                const deptCreationRequest: DepartmentCreationRequest = {
                    clusterId: flowRootState.cluster.uuid,
                    name: deptName,
                    description: 'test dept resource apis',
                    defaultNodePools: [npName, 'default'],
                    resources: [
                        {
                            nodePool: {
                                name: npName,
                                id: npId
                            },
                            ...createResourceConfig(
                                gpuQuota,
                                gpuLimit,
                                cpuQuota,
                                cpuLimit,
                                memoryQuota,
                                memoryLimit
                            )
                        },
                        {
                            nodePool: {
                                name: 'default',
                                id: defaultNpId
                            },
                            ...createResourceConfig(
                                gpuQuota,
                                gpuLimit,
                                cpuQuota,
                                cpuLimit,
                                memoryQuota,
                                memoryLimit
                            )
                        }
                    ]
                };

                const dept = await orgUnitUtils.createDepartment(deptCreationRequest);
                deptId = dept.id;
                deptIds.push(Number(deptId));
            });

            test.afterAll('delete created department', async () => {
                if (deptIds.length > 0) {
                    await Promise.all(
                        deptIds.map(async (deptId) => {
                            try {
                                await departmentUtils.deleteDepartment(deptId);
                            } catch (e) {
                                logger.error(`Failed to delete department ${deptId}`, e);
                            }
                        })
                    );
                }
            });

            test('T5689798: test get department node pool resources', async () => {
                await test.step(`Get node pool ${npName} resource and validate`, async () => {
                    const resp = await orgUnitUtils.getDepartmentNodePoolResources([
                        `clusterId==${flowRootState.cluster.uuid},orgUnitId==${deptId}`
                    ]);

                    expect(resp.response.statusCode).toBe(StatusCodes.OK);
                    logger.info(JSON.stringify(resp.body.nodePoolResources));

                    // in case there are nodepools created by other tests that are not deleted
                    expect(resp.body.nodePoolResources.length).toBeGreaterThanOrEqual(2);

                    const npResources = resp.body.nodePoolResources;
                    // retrieve nodePoolName
                    const nodePoolNames = npResources.map((item: any) => item.nodePoolName);
                    expect(nodePoolNames).toEqual(expect.arrayContaining(['default', npName]));

                    for (const npObj of npResources) {
                        if (npObj.nodePoolName === 'default' || npObj.nodePoolName === npName) {
                            validateResourceConfig(
                                npObj.resources,
                                gpuQuota,
                                gpuLimit,
                                cpuQuota,
                                cpuLimit,
                                memoryQuota,
                                memoryLimit
                            );
                        }
                    }
                });
            });

            test('T5726183: test get department node pool resources weight sum', async () => {
                await test.step(`Get node pool ${deptName} weight sum and validate`, async () => {
                    const resp = await orgUnitUtils.sumDepartmentNodePoolResources(npId, [
                        `clusterId==${flowRootState.cluster.uuid},orgUnitId==${deptId}`
                    ]);

                    expect(resp.response.statusCode).toBe(StatusCodes.OK);
                    expect(resp.body.weightsSum).toEqual(gpuQuota);
                    expect(resp.body.deservedSum).toEqual(gpuQuota);
                });
            });

            test('T5689799: test batch patch department node pool resources', async () => {
                await test.step(`Batch patch department node pool ${npName} resource and validate the response`, async () => {
                    const newResources = createResourceConfig(
                        newGpuQuota,
                        newGpuLimit,
                        newCpuQuota,
                        newCpuLimit,
                        newMemoryQuota,
                        newMemoryLimit
                    );

                    const batchRequest: BatchPatchNodePoolResourcesRequest = {
                        patchRequests: [
                            {
                                orgUnitId: defaultDeptId,
                                nodePoolId: npId,
                                resources: newResources
                            },
                            {
                                orgUnitId: deptId,
                                nodePoolId: npId,
                                resources: newResources
                            }
                        ]
                    };
                    const resp = await orgUnitUtils.batchPatchDepartmentNodePoolResources(batchRequest);

                    expect(resp.response.statusCode).toBe(StatusCodes.OK);
                    expect(resp.body.totalSuccessful).toBe(2);

                    for (const deptObj of resp.body.succeeded) {
                        validateResourceConfig(
                            deptObj.resources,
                            newGpuQuota,
                            newGpuLimit,
                            newCpuQuota,
                            newCpuLimit,
                            newMemoryQuota,
                            newMemoryLimit
                        );
                    }
                });

                const newCrdResources = createCrdResourceConfig(
                    newGpuQuota,
                    newGpuLimit,
                    newCpuQuota,
                    newCpuLimit,
                    newMemoryQuota,
                    newMemoryLimit
                );

                await test.step(`Validate the default department nodepool ${npName} queue resource`, async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(
                        Number(defaultDeptId),
                        newCrdResources,
                        npName
                    );
                });

                await test.step(`Validate the ${deptName} department nodepool ${npName} queue resource`, async () => {
                    await crdUtils.validateDepartmentQueueCrdResources(
                        Number(deptId),
                        newCrdResources,
                        npName
                    );
                });
            });
        });

        test.describe('Project nodepool resources API tests', () => {
            const proj1Name = objectGenerator.generateShortNameByPrefix('proj1-res-api', 3);
            const proj2Name = objectGenerator.generateShortNameByPrefix('proj2-res-api', 3);
            let proj1Id: string;
            let proj2Id: string;
            let projects: Array<{ name: string; id: string }>;

            test.beforeAll('create projects for testing', async () => {
                projectIds = [];

                const projectResourceConfig = [
                    {
                        nodePool: {
                            name: npName,
                            id: npId
                        },
                        ...createResourceConfig(
                            gpuQuota,
                            gpuLimit,
                            cpuQuota,
                            cpuLimit,
                            memoryQuota,
                            memoryLimit
                        )
                    }
                ];

                // Create 2 projects
                const [project1, project2] = await Promise.all([
                    orgUnitUtils.createProjectAndValidate({
                        name: proj1Name,
                        clusterId: flowRootState.cluster.uuid,
                        parentId: defaultDeptId,
                        resources: projectResourceConfig
                    }),
                    orgUnitUtils.createProjectAndValidate({
                        name: proj2Name,
                        clusterId: flowRootState.cluster.uuid,
                        parentId: defaultDeptId,
                        resources: projectResourceConfig
                    })
                ]);

                proj1Id = project1.id;
                proj2Id = project2.id;
                projectIds.push(proj1Id, proj2Id);

                projects = [
                    { name: proj1Name, id: proj1Id },
                    { name: proj2Name, id: proj2Id }
                ];
            });

            test.afterAll('delete created projects', async () => {
                if (projectIds.length > 0) {
                    await Promise.all(
                        projectIds.map(async (projectId) => {
                            try {
                                await operations.deleteProjectAndNamespace(Number(projectId));
                            } catch (e) {
                                logger.error(`Failed to delete project ${projectId}`, e);
                            }
                        })
                    );
                }
            });

            test('T5689796: test get project node pool resources', async () => {
                for (const project of projects) {
                    await test.step(`Get project ${project.name} nodepool ${npName} resource and validate`, async () => {
                        const resp = await orgUnitUtils.getProjectNodePoolResources([
                            `clusterId==${flowRootState.cluster.uuid},orgUnitId==${project.id},nodePoolId==${npId}`
                        ]);

                        expect(resp.response.statusCode).toBe(StatusCodes.OK);
                        expect(resp.body.nodePoolResources.length).toBe(1);

                        for (const npObj of resp.body.nodePoolResources) {
                            validateResourceConfig(
                                npObj.resources,
                                gpuQuota,
                                gpuLimit,
                                cpuQuota,
                                cpuLimit,
                                memoryQuota,
                                memoryLimit
                            );
                        }
                    });
                }
            });

            test('T5726184: test get project node pool resources weight sum', async () => {
                for (const project of projects) {
                    await test.step(`Get node pool ${npName} weight sum and validate`, async () => {
                        const resp = await orgUnitUtils.sumProjectNodePoolResources(npId, [
                            `clusterId==${flowRootState.cluster.uuid},orgUnitId==${project.id}`
                        ]);

                        expect(resp.response.statusCode).toBe(StatusCodes.OK);
                        expect(resp.body.weightsSum).toEqual(gpuQuota);
                    });
                }
            });

            test('T5689797: test batch patch project node pool resources', async () => {
                await test.step(`Batch patch project node pool ${npName} resource and validate the response`, async () => {
                    const newResources = createResourceConfig(
                        newGpuQuota,
                        newGpuLimit,
                        newCpuQuota,
                        newCpuLimit,
                        newMemoryQuota,
                        newMemoryLimit
                    );

                    const batchRequest: BatchPatchNodePoolResourcesRequest = {
                        patchRequests: [
                            {
                                orgUnitId: proj1Id,
                                nodePoolId: npId,
                                resources: newResources
                            },
                            {
                                orgUnitId: proj2Id,
                                nodePoolId: npId,
                                resources: newResources
                            }
                        ]
                    };
                    const resp = await orgUnitUtils.batchPatchProjectNodePoolResources(batchRequest);

                    expect(resp.response.statusCode).toBe(StatusCodes.OK);
                    expect(resp.body.totalSuccessful).toBe(2);

                    for (const projObj of resp.body.succeeded) {
                        validateResourceConfig(
                            projObj.resources,
                            newGpuQuota,
                            newGpuLimit,
                            newCpuQuota,
                            newCpuLimit,
                            newMemoryQuota,
                            newMemoryLimit
                        );
                    }
                });

                const newCrdResources = createCrdResourceConfig(
                    newGpuQuota,
                    newGpuLimit,
                    newCpuQuota,
                    newCpuLimit,
                    newMemoryQuota,
                    newMemoryLimit
                );

                const projectTestCases = [{ name: proj1Name }, { name: proj2Name }];

                for (const project of projectTestCases) {
                    await test.step(`Validate the ${project.name} project nodepool ${npName} queue resource`, async () => {
                        await crdUtils.validateProjectQueueCrdResources(
                            project.name,
                            newCrdResources,
                            npName
                        );
                    });
                }
            });
        });
    });
});
