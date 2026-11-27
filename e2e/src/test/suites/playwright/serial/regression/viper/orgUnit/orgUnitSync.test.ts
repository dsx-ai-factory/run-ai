/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectUtils,
    tags,
    waitUtils,
    Utils,
    authorizationUtils,
    usersManager,
    versionUtils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import { backendApi, clusterApi } from '../../../../../../../modules/api';
import { expectedOQWhenAwOff } from '../../../../../../../modules/helpers/utils/OrgUnitUtils';
import {
    Department,
    Project,
    ProjectPhase,
    Resources,
    QuotaStatusNodePool
} from '../../../../../../../modules/api/apis/orgUnitApi';
import { test, expect } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { WorkloadStatus } from '../../../../../../../modules/helpers/models/baseWorkload';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { Training } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/training';
import { KubernetesObject } from '@kubernetes/client-node/dist/types';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    getOrgUnitLastSyncTimestamp,
    setOrgUnitUpdatedAt
} from '../../../../../../../modules/helpers/utils/OrgUnitCrdFields';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Org unit sync tests, tags:  ${tags.resource.projects} ${tags.resource.department} ${tags.teams.viper}`,
    {
        tag: [
            tags.services.orgUnitService,
            tags.services.agent,
            tags.services.clusterSync,
            tags.services.projectController,
            tags.runOn.serialCI
        ]
    },
    () => {
        supportedFrom(SupportedVersions.V2_20)(() => {
            let createdDepartments: Department[] = [];
            let createdProjects: Project[] = [];
            let hasOnlyWorkloads: boolean;

            test.beforeAll(async () => {
                hasOnlyWorkloads = versionUtils.isVersionSupported(
                    flowRootState.cluster.version,
                    SupportedVersions.V2_23
                );
                clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
                backendApi.organizationalApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                await operations.forceCleanup();
                await nodePoolUtils.deleteAllClusterNodePools();
            });

            test.afterAll(async () => {
                await nodePoolUtils.deleteAllClusterNodePools();
                for (const project of createdProjects) {
                    await orgUnitUtils.deleteProject(project.id);
                }
                createdProjects = [];
                for (const department of createdDepartments) {
                    await orgUnitUtils.deleteDepartment(department.id);
                }
                createdDepartments = [];
            });

            async function createDepartment(
                defaultNodePoolId: string,
                nodePoolId: string = null,
                nodePoolName: string = null
            ): Promise<Department> {
                const departmentName = objectGenerator.generateNameByPrefix('d');

                const resources = [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 10, limit: -1, overQuotaWeight: 10 }
                    }
                ];
                if (nodePoolId) {
                    resources.push({
                        nodePool: { id: nodePoolId, name: nodePoolName },
                        gpu: { deserved: 20, limit: -1, overQuotaWeight: 20 }
                    });
                }
                return await orgUnitUtils.createDepartment({
                    name: departmentName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: resources
                });
            }

            async function createProjectWithNodePool(
                defaultNodePoolId: string,
                nodePoolId: string,
                nodePoolName: string
            ): Promise<Project> {
                const projectName = objectGenerator.generateNameByPrefix('d');

                return await orgUnitUtils.createProject({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 2, limit: -1, overQuotaWeight: 2 }
                        },
                        {
                            nodePool: { id: nodePoolId, name: nodePoolName },
                            gpu: { deserved: 1, limit: -1, overQuotaWeight: 1 }
                        }
                    ]
                });
            }

            async function createProject(defaultNodePoolId: string): Promise<Project> {
                const projectName = objectGenerator.generateNameByPrefix('d');

                return await orgUnitUtils.createProject({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 2, limit: -1, overQuotaWeight: 2 }
                        }
                    ]
                });
            }

            async function createProjectWithParent(
                defaultNodePoolId: string,
                departmentId: string
            ): Promise<Project> {
                const projectName = objectGenerator.generateNameByPrefix('p');

                return await orgUnitUtils.createProject({
                    name: projectName,
                    parentId: departmentId,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 2, limit: -1, overQuotaWeight: 2 }
                        }
                    ]
                });
            }

            test.describe('basic tests', () => {
                let nodePool;
                let defaultNodePoolId;

                function validateResourcesAndGetMap(
                    resources: Resources[],
                    hasNodePool: boolean
                ): Map<string, Resources> {
                    if (hasNodePool) {
                        expect(resources.length).toBe(2);
                    } else {
                        expect(resources.length).toBe(1);
                    }
                    const npToResourcesMap: Map<string, Resources> = resources.reduce(
                        (map, resources) => {
                            map[resources.nodePool.id] = resources;
                            return map;
                        },
                        {} as Map<string, Resources>
                    );
                    expect(npToResourcesMap[defaultNodePoolId]).toBeDefined();

                    return npToResourcesMap;
                }

                function validateResources(resources: Resources, deserved: number): void {
                    expect(resources.gpu.deserved).toBe(deserved);
                    expect(resources.gpu.limit).toBe(-1);
                    expect(resources.gpu.overQuotaWeight).toBe(expectedOQWhenAwOff(deserved));
                }

                async function validateQueue(queueName: string, deserved: number): Promise<void> {
                    const queueCrd = await waitUtils.waitForCrd(crds.queueCrd, queueName, '');
                    const queueGpuQuota = queueCrd['spec']['resources']['gpu'];
                    expect(queueGpuQuota['quota']).toBe(deserved);
                    expect(queueGpuQuota['limit']).toBe(-1);
                    expect(queueGpuQuota['overQuotaWeight']).toBe(expectedOQWhenAwOff(deserved));
                }

                async function validateDepartmentCrd(department: Department): Promise<void> {
                    const departmentCrd = await departmentUtils.waitForDepartmentCrdUpdated(
                        department.name,
                        department.updatedAt
                    );
                    expect(departmentCrd).toBeDefined();
                    expect(departmentCrd).toHaveProperty('metadata');
                    expect(departmentCrd).toHaveProperty('spec');
                    expect(departmentCrd['metadata']).toHaveProperty('name', department.name);
                    expect(getOrgUnitLastSyncTimestamp(departmentCrd)).toBeTruthy();
                }

                async function validateDepartment(
                    departmentId: string,
                    afterUpdate: boolean,
                    afterResourcesUpdate: boolean,
                    hasNodePool: boolean
                ): Promise<void> {
                    const department = await orgUnitUtils.waitForDepartmentFromBackendReadyAndResources(
                        departmentId,
                        hasNodePool ? 2 : 1
                    );
                    expect(department).toBeDefined();
                    if (!afterUpdate) {
                        expect(department.createdAt).toEqual(department.updatedAt);
                    } else {
                        expect(department.createdAt).not.toEqual(department.updatedAt);
                    }
                    const npToResourcesMap: Map<string, Resources> = validateResourcesAndGetMap(
                        department.resources,
                        hasNodePool
                    );
                    validateResources(npToResourcesMap[defaultNodePoolId] as Resources, 10);

                    if (hasNodePool) {
                        expect(npToResourcesMap[nodePool.id]).toBeDefined();
                        if (!afterResourcesUpdate) {
                            validateResources(npToResourcesMap[nodePool.id] as Resources, 20);
                        } else {
                            validateResources(npToResourcesMap[nodePool.id] as Resources, 12);
                        }
                    }

                    await validateDepartmentCrd(department);

                    await validateQueue('q-' + departmentId, 10);

                    if (hasNodePool) {
                        if (!afterResourcesUpdate) {
                            await validateQueue('q-' + departmentId + '-' + nodePool.name, 20);
                        } else {
                            await validateQueue('q-' + departmentId + '-' + nodePool.name, 12);
                        }
                    }
                }

                async function validateProjectCrd(project: Project): Promise<void> {
                    const projectCrd = await projectUtils.waitForProjectCrdUpdated(
                        project.name,
                        project.updatedAt
                    );
                    expect(projectCrd).toBeDefined();
                    expect(projectCrd).toHaveProperty('metadata');
                    expect(projectCrd).toHaveProperty('spec');
                    expect(projectCrd).toHaveProperty('status');
                    expect(projectCrd['metadata']).toHaveProperty('name', project.name);
                    expect(getOrgUnitLastSyncTimestamp(projectCrd)).toBeTruthy();
                }

                async function validateProject(
                    projectId: string,
                    afterUpdate: boolean,
                    afterResourcesUpdate: boolean,
                    hasNodePool: boolean
                ): Promise<void> {
                    const project = await orgUnitUtils.waitForProjectFromBackendReadyAndResources(
                        projectId,
                        hasNodePool ? 2 : 1
                    );
                    expect(project).toBeDefined();
                    if (!afterUpdate) {
                        expect(project.createdAt).toEqual(project.updatedAt);
                    } else {
                        expect(project.createdAt).not.toEqual(project.updatedAt);
                    }
                    expect(project.status.phase).toBe(ProjectPhase.Ready);
                    const npToResourcesMap: Map<string, Resources> = validateResourcesAndGetMap(
                        project.resources,
                        hasNodePool
                    );
                    validateResources(npToResourcesMap[defaultNodePoolId] as Resources, 2);

                    if (hasNodePool) {
                        expect(npToResourcesMap[nodePool.id]).toBeDefined();
                        if (!afterResourcesUpdate) {
                            validateResources(npToResourcesMap[nodePool.id] as Resources, 1);
                        } else {
                            validateResources(npToResourcesMap[nodePool.id] as Resources, 10);
                        }
                    }

                    await validateProjectCrd(project);

                    await validateQueue(project.name, 2);

                    if (hasNodePool) {
                        if (!afterResourcesUpdate) {
                            await validateQueue(project.name + '-' + nodePool.name, 1);
                        } else {
                            await validateQueue(project.name + '-' + nodePool.name, 10);
                        }
                    }
                }

                async function validateProjectStatus(projectId: string, allocated: number): Promise<void> {
                    const project = await orgUnitUtils.waitForProjectFromBackendReadyAllocated(
                        projectId,
                        allocated
                    );
                    expect(project).toBeDefined();
                    const npToAllocationMap: Map<string, QuotaStatusNodePool> =
                        project.status.nodePoolQuotaStatuses.reduce(
                            (map, status) => {
                                map[status.nodePoolName] = status;
                                return map;
                            },
                            {} as Map<string, QuotaStatusNodePool>
                        );

                    let nodePoolAllocation = npToAllocationMap['default'] as QuotaStatusNodePool;
                    expect(nodePoolAllocation).toBeDefined();
                    expect(nodePoolAllocation.allocated.gpu).toBe(allocated);
                    expect(nodePoolAllocation.requested.gpu).toBe(allocated);
                    expect(nodePoolAllocation.allocatedNonPreemptible.gpu).toBeUndefined();

                    nodePoolAllocation = npToAllocationMap[nodePool.name] as QuotaStatusNodePool;
                    expect(nodePoolAllocation).toBeDefined();
                    expect(nodePoolAllocation.allocated.gpu).toBeUndefined();
                    expect(nodePoolAllocation.requested.gpu).toBeUndefined();
                    expect(nodePoolAllocation.allocatedNonPreemptible.gpu).toBeUndefined();

                    const quotaStatus = project.status.quotaStatus;
                    expect(quotaStatus).toBeDefined();
                    expect(quotaStatus.allocated.gpu).toBe(allocated);
                    expect(quotaStatus.requested.gpu).toBe(allocated);
                    expect(quotaStatus.allocatedNonPreemptible.gpu).toBeUndefined();
                }

                async function createWorkload(project: Project): Promise<Training> {
                    const workloadName = objectGenerator.generateNameByPrefix('training');
                    const { body: trainWorkload } = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity',
                            nodePools: ['default'],
                            compute: { gpuDevicesRequest: 1 }
                        }
                    );
                    if (hasOnlyWorkloads) {
                        await waitUtils.waitForBackendWorkloadPhase(
                            trainWorkload.name,
                            parseInt(project.id, 10),
                            Phase.Running
                        );
                    } else {
                        await Promise.all([
                            waitUtils.waitForBackendJobStatus(trainWorkload.name, WorkloadStatus.RUNNING),
                            waitUtils.waitForBackendWorkloadPhase(
                                trainWorkload.name,
                                parseInt(project.id, 10),
                                Phase.Running
                            )
                        ]);
                    }
                    return trainWorkload;
                }

                test.beforeAll(async () => {
                    defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromCluster()).id);
                    logger.info('create node pool');
                    const nodePoolName = objectGenerator.generateNameByPrefix('n');
                    const nodePoolRes = await nodePoolUtils.createClusterNodePool(
                        nodePoolName,
                        nodePoolName + '-key',
                        nodePoolName + '-value'
                    );
                    expect(nodePoolRes.response.statusCode).toBe(StatusCodes.CREATED);
                    nodePool = nodePoolRes.body;
                    logger.info('create department');
                    const department = await createDepartment(defaultNodePoolId, nodePool.id, nodePoolName);
                    createdDepartments.push(department);
                    logger.info('create project');
                    const project = await createProjectWithNodePool(
                        defaultNodePoolId,
                        nodePool.id,
                        nodePoolName
                    );
                    createdProjects.push(project);
                    logger.info('validate node pool');
                });

                test('Basic test steps', async () => {
                    await test.step('basic department create / update', async function () {
                        logger.info('Department create / update test');
                        const departmentId = createdDepartments[0].id;

                        logger.info('validate department + queue');
                        await validateDepartment(departmentId, false, false, true);

                        logger.info('update department description');
                        let department = await orgUnitUtils.getDepartment(departmentId);
                        expect(department).toBeDefined();
                        department.description = 'new desc';
                        await orgUnitUtils.updateDepartment(departmentId, department);

                        logger.info('validate department + queue');
                        await validateDepartment(departmentId, true, false, true);

                        logger.info('update department');
                        department = await orgUnitUtils.getDepartment(departmentId);
                        expect(department).toBeDefined();
                        const resources = department.resources;
                        const npToResourcesMap: Map<string, Resources> = resources.reduce(
                            (map, resources) => {
                                map[resources.nodePool.id] = resources;
                                return map;
                            },
                            {} as Map<string, Resources>
                        );
                        const resourcesForUpdate = npToResourcesMap[nodePool.id];
                        resourcesForUpdate.gpu.deserved = 12;
                        resourcesForUpdate.gpu.overQuotaWeight = 12;
                        department.resources = [npToResourcesMap[defaultNodePoolId], resourcesForUpdate];
                        await orgUnitUtils.updateDepartment(departmentId, department);

                        logger.info('validate department + queue');
                        await validateDepartment(departmentId, true, true, true);
                    });

                    await test.step('basic project create / update', async function () {
                        logger.info('Project create / update test');
                        const projectId = createdProjects[0].id;

                        logger.info('validate project + queue');
                        await validateProject(projectId, false, false, true);

                        logger.info('update project description');
                        let project = await orgUnitUtils.getProject(projectId);
                        expect(project).toBeDefined();
                        project.description = 'proj desc';
                        await orgUnitUtils.updateProject(projectId, project);

                        logger.info('validate project + queue');
                        await validateProject(projectId, true, false, true);

                        logger.info('update project');
                        project = await orgUnitUtils.getProject(projectId);
                        expect(project).toBeDefined();
                        const resources = project.resources;
                        const npToResourcesMap: Map<string, Resources> = resources.reduce(
                            (map, resources) => {
                                map[resources.nodePool.id] = resources;
                                return map;
                            },
                            {} as Map<string, Resources>
                        );
                        const resourcesForUpdate = npToResourcesMap[nodePool.id];
                        resourcesForUpdate.gpu.deserved = 10;
                        resourcesForUpdate.gpu.overQuotaWeight = 10;
                        project.resources = [npToResourcesMap[defaultNodePoolId], resourcesForUpdate];
                        await orgUnitUtils.updateProject(projectId, project);

                        logger.info('validate project + queue');
                        await validateProject(projectId, true, true, true);
                    });

                    await test.step('project status change', async function () {
                        logger.info('Project status change test');
                        const project = createdProjects[0];

                        logger.info('validate project status');
                        await validateProjectStatus(project.id, undefined);

                        logger.info('submit workload with 1 gpus');
                        const trainWorkload = await createWorkload(project);

                        logger.info('validate project status');
                        await validateProjectStatus(project.id, 1);

                        logger.info('delete workload');
                        await workloadsCrudUtils.deleteTraining(trainWorkload.workloadId);

                        logger.info('validate project status');
                        await validateProjectStatus(project.id, undefined);
                    });

                    await test.step('delete node pool', async function () {
                        logger.info('test delete node pool');
                        const delRes = await clusterApi.nodepoolsApi.deleteNodepool(nodePool.id);
                        expect(delRes.response.statusCode).toEqual(StatusCodes.ACCEPTED);

                        logger.info('validate node pool deleted');
                        await nodePoolUtils.waitForDeletedNodePool(nodePool.id);

                        logger.info('validate department + queue');
                        await validateDepartment(createdDepartments[0].id, true, true, false);

                        logger.info('validate project + queue');
                        await validateProject(createdProjects[0].id, true, true, false);
                    });
                });
            });

            test.describe('access-rules change test', function () {
                let defaultNodePoolId: string;
                let user: UserPostResponse;
                let accessRule: AccessRule;

                test.beforeAll(async function () {
                    defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromCluster()).id);
                });

                test.afterAll(async function () {
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        user.id
                    );
                });

                test('create user after project', async function () {
                    logger.info('create project');
                    let project = await createProject(defaultNodePoolId);
                    createdProjects.push(project);

                    logger.info('validate project');
                    project = await orgUnitUtils.waitForProjectFromBackendReadyAndResources(project.id, 1);
                    await projectUtils.waitForProjectCrdUpdated(project.name, project.updatedAt);

                    //create user
                    logger.info('create researcher');
                    const userEmail = objectGenerator.generateEmailByPrefix(
                        `researcher-${Utils.generateShortUuid().slice(0, 4)}`
                    );
                    user = (
                        await usersManager.createLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            userEmail
                        )
                    ).body;
                    accessRule = await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        user.username,
                        flowRootState.users.l1_researcher.role.id
                    );

                    logger.info('delete access rule');
                    await authorizationUtils.deleteAccessRule(accessRule);
                });
            });

            test.describe('permissions tests', function () {
                let defaultNodePool;

                async function updateDepartmentCrd(
                    department: object,
                    name: string
                ): Promise<KubernetesObject> {
                    return await kubeCli.patchCrd(crds.departmentCrd, department, name);
                }

                async function updateProjectCrd(project: object, name: string): Promise<KubernetesObject> {
                    return await kubeCli.patchCrd(crds.runaiProjectCrd, project, name);
                }

                test.beforeAll(async function () {
                    defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromCluster();

                    logger.info('create department');
                    const department = await createDepartment(defaultNodePool.id);
                    createdDepartments.push(department);
                    logger.info('create project');
                    const project = await createProjectWithParent(defaultNodePool.id, department.id);
                    createdProjects.push(project);
                });

                test('update project crd', async function () {
                    const projectName = createdProjects[0].name;
                    const projectCrd = await projectUtils.waitForProjectCrdNonEmptyStatus(projectName);
                    // Any write is enough to prove the webhook rejects it; use whichever location
                    // actually carries updatedAt so the mutation is not silently pruned.
                    setOrgUnitUpdatedAt(projectCrd, '2025-01-01T12:00:00Z');
                    try {
                        await updateProjectCrd(projectCrd, projectName);
                        throw new Error('user should not have permissions for project crd update');
                    } catch (e) {
                        expect(e.message).toContain('Failure');
                        expect(e.message).toContain('operation not allowed');
                        expect(e.message).toContain('Forbidden');
                        expect(e.message).toContain('403');
                        expect(e.message).toContain('admission webhook');
                        expect(e.message).toContain('runai-project-controller.runai.svc');
                    }
                });

                test('update department crd', async function () {
                    const departmentName = createdDepartments[0].name;
                    const departmentCrd = await departmentUtils.waitForDepartmentCrd(departmentName);
                    setOrgUnitUpdatedAt(departmentCrd, '2025-01-01T12:00:00Z');
                    try {
                        await updateDepartmentCrd(departmentCrd, departmentName);
                        throw new Error('user should not have permissions for department crd update');
                    } catch (e) {
                        expect(e.message).toContain('Failure');
                        expect(e.message).toContain('operation not allowed');
                        expect(e.message).toContain('Forbidden');
                        expect(e.message).toContain('403');
                        expect(e.message).toContain('admission webhook');
                        expect(e.message).toContain('runai-project-controller.runai.svc');
                    }
                });
            });
        });
    }
);
