/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    crds,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    projectUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import {
    Department,
    NodeType,
    Project,
    ProjectCreationRequest
} from '../../../../../../../modules/api/apis/orgUnitApi';
import {
    AssignedResources,
    JobsNodeAffinityTrain,
    Project as OldProject,
    ProjectCreateRequest
} from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import {
    getProjectJobLimitSecs,
    getProjectNodeAffinity
} from '../../../../../../../modules/helpers/utils/OrgUnitCrdFields';
import { StatusCodes } from 'http-status-codes';
import { orgUnitService, projectsDepartmentsApi } from '../../../../../../../modules/api';
import AffinityTypeEnum = JobsNodeAffinityTrain.AffinityTypeEnum;
import { HttpError } from '../../../../../../../modules/api/apis/projectsDepartmentsApi/generated/api/apis';
import { test, expect } from '@playwright/test';

const crdTimeout = 60;

test.describe(`Projects advanced API Effective, tags:  ${tags.resource.projects} ${tags.tool.api} ${tags.teams.viper}`, () => {
    const defaultNodePoolName = 'default';
    let defaultNodePoolId: string;
    let testProject: Project;
    let projectName: string;
    let createdNp: NodePool;
    let department: Department;
    let nodeType1: NodeType;
    let nodeType2: NodeType;
    const deservedGpus = 1;

    async function createNodePool(): Promise<void> {
        const npName = objectGenerator.generateNameByPrefix('np');
        const createdNpResponse = await nodePoolUtils.createBackendNodePool(npName, 'key-A', 'value-B');
        expect(createdNpResponse.response.statusCode).toBe(StatusCodes.CREATED);
        createdNp = createdNpResponse.body;
    }

    async function createNodeType(): Promise<NodeType> {
        const name = objectGenerator.generateNameByPrefix('nt');
        return await orgUnitUtils.createNodeType({ name: name, clusterId: flowRootState.cluster.uuid });
    }

    test.beforeAll(async () => {
        projectsDepartmentsApi.projectsApi.accessToken =
            flowRootState.users.system_administrator.accessToken;
        orgUnitService.projectApi.accessToken = flowRootState.users.system_administrator.accessToken;

        await apiUtils.disableProjectOverQuotaFeature();
        await apiUtils.disableCpuQuotaFeature();

        await nodePoolUtils.deleteAllNodePools();
        defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
        await createNodePool();
        nodeType1 = await createNodeType();
        nodeType2 = await createNodeType();
    });

    test.afterAll(async () => {
        if (createdNp) {
            try {
                await nodePoolUtils.deleteNodePool(createdNp);
                createdNp = null;
            } catch {
                logger.error('Failed to delete node pool');
            }
        }

        createdNp = null;
        if (nodeType1) {
            await orgUnitUtils.deleteNodeType(nodeType1.id);
        }
        if (nodeType2) {
            await orgUnitUtils.deleteNodeType(nodeType2.id);
        }
    });

    test.describe('department without effective fields', () => {
        const projectIds = [];

        test.beforeAll(async () => {
            const departmentName = objectGenerator.generateNameByPrefix('department');
            department = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 10 }
                    }
                ]
            });
            await orgUnitUtils.waitForDepartmentInOrgUnitService(department.id);
        });

        test.afterAll(async () => {
            const promises = projectIds.map((projectId) => orgUnitUtils.deleteProject(projectId));
            await Promise.all(promises);

            if (department) {
                try {
                    await orgUnitUtils.deleteDepartment(department.id);
                } catch (e) {
                    logger.error(`Failed to delete department ${department.id}, error: ${e}`);
                }
            }
        });

        function validateProject(getResponse: Project, withOverride: boolean): void {
            if (withOverride) {
                expect(getResponse.schedulingRules).toEqual({
                    trainingJobTimeLimitSeconds: undefined,
                    trainingJobMaxIdleDurationSeconds: undefined,
                    interactiveJobPreemptIdleDurationSeconds: undefined,
                    interactiveJobMaxIdleDurationSeconds: undefined,
                    interactiveJobTimeLimitSeconds: 2
                });
                expect(getResponse.effective.schedulingRules).toEqual({
                    trainingJobTimeLimitSeconds: undefined,
                    trainingJobMaxIdleDurationSeconds: undefined,
                    interactiveJobPreemptIdleDurationSeconds: undefined,
                    interactiveJobMaxIdleDurationSeconds: undefined,
                    interactiveJobTimeLimitSeconds: 2
                });
                expect(getResponse.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.nodeTypes.training.length).toBe(1);
                expect(getResponse.nodeTypes.training).toContain(nodeType1.id);
                expect(getResponse.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
                expect(getResponse.effective.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.effective.nodeTypes.training.length).toBe(1);
                expect(getResponse.effective.nodeTypes.training).toContain(nodeType1.id);
                expect(getResponse.effective.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
                expect(getResponse.defaultNodePools).toEqual([createdNp.name]);
                expect(getResponse.effective.defaultNodePools).toEqual([createdNp.name]);
            } else {
                expect(getResponse.schedulingRules).toBeUndefined();
                expect(getResponse.effective.schedulingRules).toBeUndefined();
                expect(getResponse.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.nodeTypes.training).toBeUndefined();
                expect(getResponse.effective.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.effective.nodeTypes.training).toBeUndefined();
                expect(getResponse.defaultNodePools).toBeUndefined();
                expect(getResponse.effective.defaultNodePools).toEqual(['default']);
            }
        }

        function validateOldProject(getOldResponse: OldProject, withOverride: boolean): void {
            if (withOverride) {
                expect(getOldResponse.interactiveJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.interactiveJobTimeLimitSecs).toBe(2);
                expect(getOldResponse.interactivePreemptibleJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobTimeLimitSecs).toBeUndefined();
                expect(getOldResponse.nodeAffinity).toEqual({
                    train: {
                        affinityType: 'only_selected',
                        selectedTypes: [
                            {
                                id: parseInt(nodeType1.id),
                                name: nodeType1.name
                            }
                        ]
                    },
                    interactive: {
                        affinityType: 'no_limit',
                        selectedTypes: []
                    }
                });
                expect(getOldResponse.interactiveNodeAffinity).toBeUndefined();
                expect(getOldResponse.defaultNodePools).toEqual([createdNp.name]);
            } else {
                expect(getOldResponse.interactiveJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.interactiveJobTimeLimitSecs).toBeUndefined();
                expect(getOldResponse.interactivePreemptibleJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobTimeLimitSecs).toBeUndefined();
                expect(getOldResponse.trainNodeAffinity).toBeUndefined();
                expect(getOldResponse.nodeAffinity).toEqual({
                    train: {
                        affinityType: 'no_limit',
                        selectedTypes: []
                    },
                    interactive: {
                        affinityType: 'no_limit',
                        selectedTypes: []
                    }
                });
                expect(getOldResponse.defaultNodePools).toEqual(['default']);
            }
        }

        function validateCrd(projectCrd: object, withOverride: boolean): void {
            if (withOverride) {
                expect(projectCrd).toBeDefined();
                expect(projectCrd['spec'].defaultNodePools).toEqual([createdNp.name]);
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityInteractive')).toBeUndefined();
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityTrain')).toEqual([nodeType1.name]);
                expect(getProjectJobLimitSecs(projectCrd, 'trainingJobTimeLimitSecs')).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'trainingJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactivePreemptibleJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(getProjectJobLimitSecs(projectCrd, 'interactiveJobTimeLimitSecs')).toBe(2);
            } else {
                expect(projectCrd).toBeDefined();
                expect(projectCrd['spec'].defaultNodePools).toEqual(['default']);
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityInteractive')).toBeUndefined();
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityTrain')).toBeUndefined();
                expect(getProjectJobLimitSecs(projectCrd, 'trainingJobTimeLimitSecs')).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'trainingJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactivePreemptibleJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobTimeLimitSecs')
                ).toBeUndefined();
            }
        }

        test('project without effective values', async () => {
            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            testProject = await orgUnitUtils.createProjectAndValidate({
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: {
                            deserved: deservedGpus,
                            limit: deservedGpus
                        }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: department.id
            });

            const projectId = String(testProject.id);
            projectIds.push(projectId);
            logger.info('validating org-unit project ' + projectName);
            const getResponse = await orgUnitUtils.getProject(projectId);
            validateProject(getResponse, false);

            logger.info('validating old api project ' + projectName);
            const getOldResponse = (
                await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
            ).body.find((pro) => pro.id.toString() === testProject.id);
            validateOldProject(getOldResponse, false);

            logger.info('validating project crd ' + projectName);
            const projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, '', crdTimeout);
            validateCrd(projectCrd, false);
        });

        test('project with effective values', async () => {
            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            testProject = await orgUnitUtils.createProjectAndValidate({
                schedulingRules: {
                    interactiveJobTimeLimitSeconds: 2
                },
                nodeTypes: {
                    training: [nodeType1.id]
                },
                defaultNodePools: [createdNp.name],
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: {
                            deserved: deservedGpus,
                            limit: deservedGpus
                        }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: department.id
            });

            const projectId = String(testProject.id);
            projectIds.push(projectId);
            logger.info('validating BE project ' + projectName);
            const getResponse = await orgUnitUtils.getProject(projectId);
            validateProject(getResponse, true);

            logger.info('validating old api project ' + projectName);
            const getOldResponse = (
                await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
            ).body.find((pro) => pro.id.toString() === testProject.id);
            validateOldProject(getOldResponse, true);

            logger.info('validating project crd ' + projectName);
            const projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, '', crdTimeout);
            validateCrd(projectCrd, true);
        });
    });

    test.describe('department with effective fields', () => {
        const projectIds = [];

        test.beforeAll(async () => {
            const departmentName = objectGenerator.generateNameByPrefix('department');
            department = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 10 }
                    }
                ],
                schedulingRules: {
                    trainingJobTimeLimitSeconds: 100
                },
                defaultNodePools: [createdNp.name, defaultNodePoolName],
                nodeTypes: {
                    training: [nodeType2.id, nodeType1.id]
                }
            });
            await orgUnitUtils.waitForDepartmentInOrgUnitService(department.id);
        });

        test.afterAll(async () => {
            const promises = projectIds.map((projectId) => orgUnitUtils.deleteProject(projectId));
            await Promise.all(promises);

            if (department) {
                try {
                    await orgUnitUtils.deleteDepartment(department.id);
                } catch (e) {
                    logger.error(`Failed to delete department ${department.id}, error: ${e}`);
                }
            }
        });

        function validateProject(
            getResponse: Project,
            withOverride: boolean,
            trainingJobMaxIdleDurationSeconds?: number
        ): void {
            if (withOverride) {
                if (trainingJobMaxIdleDurationSeconds) {
                    expect(getResponse.schedulingRules).toEqual({
                        trainingJobTimeLimitSeconds: undefined,
                        trainingJobMaxIdleDurationSeconds: trainingJobMaxIdleDurationSeconds,
                        interactiveJobPreemptIdleDurationSeconds: undefined,
                        interactiveJobMaxIdleDurationSeconds: undefined,
                        interactiveJobTimeLimitSeconds: 200
                    });
                    expect(getResponse.effective.schedulingRules).toEqual({
                        trainingJobTimeLimitSeconds: 100,
                        trainingJobMaxIdleDurationSeconds: trainingJobMaxIdleDurationSeconds,
                        interactiveJobPreemptIdleDurationSeconds: undefined,
                        interactiveJobMaxIdleDurationSeconds: undefined,
                        interactiveJobTimeLimitSeconds: 200
                    });
                } else {
                    expect(getResponse.schedulingRules).toEqual({
                        trainingJobTimeLimitSeconds: undefined,
                        trainingJobMaxIdleDurationSeconds: undefined,
                        interactiveJobPreemptIdleDurationSeconds: undefined,
                        interactiveJobMaxIdleDurationSeconds: undefined,
                        interactiveJobTimeLimitSeconds: 200
                    });
                    expect(getResponse.effective.schedulingRules).toEqual({
                        trainingJobTimeLimitSeconds: 100,
                        trainingJobMaxIdleDurationSeconds: undefined,
                        interactiveJobPreemptIdleDurationSeconds: undefined,
                        interactiveJobMaxIdleDurationSeconds: undefined,
                        interactiveJobTimeLimitSeconds: 200
                    });
                }
                expect(getResponse.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.nodeTypes.training.length).toBe(1);
                expect(getResponse.nodeTypes.training).toContain(nodeType1.id);
                expect(getResponse.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
                expect(getResponse.effective.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.effective.nodeTypes.training.length).toBe(1);
                expect(getResponse.effective.nodeTypes.training).toContain(nodeType1.id);
                expect(getResponse.effective.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
                expect(getResponse.defaultNodePools).toBeUndefined();
                expect(getResponse.effective.defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
            } else {
                expect(getResponse.schedulingRules).toBeUndefined();
                expect(getResponse.effective.schedulingRules).toEqual({
                    trainingJobTimeLimitSeconds: 100,
                    trainingJobMaxIdleDurationSeconds: undefined,
                    interactiveJobPreemptIdleDurationSeconds: undefined,
                    interactiveJobMaxIdleDurationSeconds: undefined,
                    interactiveJobTimeLimitSeconds: undefined
                });
                expect(getResponse.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.nodeTypes.training).toBeUndefined();
                expect(getResponse.effective.nodeTypes.workspace).toBeUndefined();
                expect(getResponse.effective.nodeTypes.training.length).toBe(2);
                expect(getResponse.effective.nodeTypes.training).toContain(nodeType1.id);
                expect(getResponse.effective.nodeTypes.training).toContain(nodeType2.id);
                expect(getResponse.effective.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
                expect(getResponse.effective.nodeTypes.names[nodeType2.id]).toBe(nodeType2.name);
                expect(getResponse.defaultNodePools).toBeUndefined();
                expect(getResponse.effective.defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
            }
        }

        function validateOldProject(getOldResponse: OldProject, withOverride: boolean): void {
            if (withOverride) {
                expect(getOldResponse.interactiveJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.interactiveJobTimeLimitSecs).toBe(200);
                expect(getOldResponse.interactivePreemptibleJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobTimeLimitSecs).toBe(100);
                expect(getOldResponse.nodeAffinity).toEqual({
                    train: {
                        affinityType: 'only_selected',
                        selectedTypes: [
                            {
                                id: parseInt(nodeType1.id),
                                name: nodeType1.name
                            }
                        ]
                    },
                    interactive: {
                        affinityType: 'no_limit',
                        selectedTypes: []
                    }
                });
                expect(getOldResponse.defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
            } else {
                expect(getOldResponse.interactiveJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.interactiveJobTimeLimitSecs).toBeUndefined();
                expect(getOldResponse.interactivePreemptibleJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobMaxIdleDurationSecs).toBeUndefined();
                expect(getOldResponse.trainingJobTimeLimitSecs).toBe(100);
                expect(getOldResponse.nodeAffinity.interactive).toEqual({
                    affinityType: 'no_limit',
                    selectedTypes: []
                });
                expect(getOldResponse.nodeAffinity.train.affinityType).toBe('only_selected');
                expect(getOldResponse.nodeAffinity.train.selectedTypes.length).toBe(2);
                expect(getOldResponse.nodeAffinity.train.selectedTypes).toContainEqual({
                    id: parseInt(nodeType1.id),
                    name: nodeType1.name
                });
                expect(getOldResponse.nodeAffinity.train.selectedTypes).toContainEqual({
                    id: parseInt(nodeType2.id),
                    name: nodeType2.name
                });
                expect(getOldResponse.defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
            }
        }

        function validateCrd(projectCrd: object, withOverride: boolean): void {
            if (withOverride) {
                expect(projectCrd).toBeDefined();
                expect(projectCrd['spec'].defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityInteractive')).toBeUndefined();
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityTrain')).toEqual([nodeType1.name]);
                expect(getProjectJobLimitSecs(projectCrd, 'trainingJobTimeLimitSecs')).toBe(100);
                expect(
                    getProjectJobLimitSecs(projectCrd, 'trainingJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactivePreemptibleJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(getProjectJobLimitSecs(projectCrd, 'interactiveJobTimeLimitSecs')).toBe(200);
            } else {
                expect(projectCrd).toBeDefined();
                expect(projectCrd['spec'].defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
                expect(getProjectNodeAffinity(projectCrd, 'nodeAffinityInteractive')).toBeUndefined();
                const nodeAffinityTrain = getProjectNodeAffinity(projectCrd, 'nodeAffinityTrain');
                expect(nodeAffinityTrain.length).toBe(2);
                expect(nodeAffinityTrain).toContain(nodeType1.name);
                expect(nodeAffinityTrain).toContain(nodeType2.name);
                expect(getProjectJobLimitSecs(projectCrd, 'trainingJobTimeLimitSecs')).toBe(100);
                expect(
                    getProjectJobLimitSecs(projectCrd, 'trainingJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactivePreemptibleJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobMaxIdleDurationSecs')
                ).toBeUndefined();
                expect(
                    getProjectJobLimitSecs(projectCrd, 'interactiveJobTimeLimitSecs')
                ).toBeUndefined();
            }
        }

        test('test department', async () => {
            logger.debug('validating department');
            const getResponse = await orgUnitUtils.getDepartment(department.id);
            expect(getResponse.name).toBe(department.name);
            expect(getResponse.schedulingRules).toEqual({
                trainingJobTimeLimitSeconds: 100,
                trainingJobMaxIdleDurationSeconds: undefined,
                interactiveJobPreemptIdleDurationSeconds: undefined,
                interactiveJobMaxIdleDurationSeconds: undefined,
                interactiveJobTimeLimitSeconds: undefined
            });
            expect(getResponse.nodeTypes.workspace).toBeUndefined();
            expect(getResponse.nodeTypes.training.length).toBe(2);
            expect(getResponse.nodeTypes.training).toContain(nodeType1.id);
            expect(getResponse.nodeTypes.training).toContain(nodeType2.id);
            expect(getResponse.nodeTypes.names[nodeType1.id]).toBe(nodeType1.name);
            expect(getResponse.nodeTypes.names[nodeType2.id]).toBe(nodeType2.name);
            expect(getResponse.defaultNodePools).toEqual([createdNp.name, defaultNodePoolName]);
        });

        test('project without effective values', async () => {
            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            testProject = await orgUnitUtils.createProjectAndValidate({
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: {
                            deserved: deservedGpus,
                            limit: deservedGpus
                        }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: department.id
            });

            const projectId = String(testProject.id);
            projectIds.push(projectId);
            logger.info('validating org-unit project ' + projectName);
            const getResponse = await orgUnitUtils.getProject(projectId);
            validateProject(getResponse, false);

            logger.info('validating old api project ' + projectName);
            const getOldResponse = (
                await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
            ).body.find((pro) => pro.id.toString() === testProject.id);
            validateOldProject(getOldResponse, false);

            logger.info('validating project crd ' + projectName);
            const projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, '', crdTimeout);
            validateCrd(projectCrd, false);
        });

        test('project with effective values', async () => {
            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            testProject = await orgUnitUtils.createProjectAndValidate({
                schedulingRules: {
                    interactiveJobTimeLimitSeconds: 200
                },
                nodeTypes: {
                    training: [nodeType1.id]
                },
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: {
                            deserved: deservedGpus,
                            limit: deservedGpus
                        }
                    }
                ],
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: department.id
            });

            const projectId = String(testProject.id);
            projectIds.push(projectId);
            logger.info('validating BE project ' + projectName);
            let getResponse = await orgUnitUtils.getProject(projectId);
            validateProject(getResponse, true);

            logger.info('validating old api project ' + projectName);
            const getOldResponse = (
                await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
            ).body.find((pro) => pro.id.toString() === testProject.id);
            validateOldProject(getOldResponse, true);

            logger.info('validating project crd ' + projectName);
            const projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, '', crdTimeout);
            validateCrd(projectCrd, true);

            getResponse.schedulingRules.trainingJobMaxIdleDurationSeconds = 1;
            await orgUnitUtils.updateProject(projectId, getResponse);

            logger.info('validating BE project ' + projectName);
            getResponse = await orgUnitUtils.getProject(projectId);
            validateProject(getResponse, true, 1);

            getResponse.schedulingRules.trainingJobTimeLimitSeconds = 2000;
            try {
                await orgUnitService.projectApi.updateProject(projectId, getResponse);
                throw new Error('cannot exceed parent time limit');
            } catch (e) {
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                // expect(e.message).to.include('failed to update project, override parent\'s values failed');
            }
        });

        test('post failure', async () => {
            projectName = objectGenerator.generateNameByPrefix('p');

            logger.debug(`Creating project ${projectName}`);
            try {
                await orgUnitService.projectApi.createProject({
                    schedulingRules: {
                        trainingJobTimeLimitSeconds: 2000
                    },
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: {
                                deserved: deservedGpus,
                                limit: deservedGpus
                            }
                        }
                    ],
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    parentId: department.id
                });
                throw new Error('cannot exceed parent time limit');
            } catch (e) {
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                expect(e.message).toContain(
                    'TrainingJobTimeLimit is 2000 and cannot be larger than parents duration 100'
                );
            }
        });
    });
});
