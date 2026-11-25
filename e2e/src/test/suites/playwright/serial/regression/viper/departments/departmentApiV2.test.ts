/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    tags,
    versionUtils
} from '../../../../../../../modules/helpers';
import {
    Department,
    DepartmentUpdateRequest,
    Project,
    Resources,
    Units
} from '../../../../../../../modules/api/apis/orgUnitApi';
import { test, expect } from '@playwright/test';
import { Department as DepartmentV1 } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { projectsDepartmentsApi } from '../../../../../../../modules/api';
import { StatusCodes } from 'http-status-codes';
import { SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Departments API V2, tags:  ${tags.resource.department} ${tags.teams.viper}`, () => {
    let defaultNodePoolId: string;
    const projectCreationRequest = {
        resources: [
            {
                nodePool: { id: defaultNodePoolId, name: 'default' },
                gpu: {
                    deserved: 1,
                    limit: 1
                }
            }
        ],
        name: objectGenerator.generateNameByPrefix('project-v2'),
        clusterId: '',
        parentId: ''
    };

    test.beforeAll(async () => {
        await nodePoolUtils.deleteAllNodePools();
        logger.info('Enable departments feature set to true successfully');
        defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);
        projectCreationRequest.resources[0].nodePool.id = defaultNodePoolId;
        projectCreationRequest.clusterId = flowRootState.cluster.uuid;
    });

    test.describe('Create department with project + validate department info', () => {
        let department: Department;
        let project: Project;
        test.afterAll(async () => {
            try {
                if (project) await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (e) {
                logger.error(`Failed to delete project ${project.id}, error: ${JSON.stringify(e)}`);
            }
            if (department) await orgUnitUtils.deleteDepartment(department.id);
        });

        test('Create department through org unit service', async () => {
            const departmentName = objectGenerator.generateNameByPrefix('department-v2');
            const rankName = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_20
            )
                ? 'High'
                : undefined;
            department = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 10 },
                        rank: rankName
                    }
                ]
            });

            orgUnitUtils.validateDepartmentResponse(department, departmentName, {
                nodePool: { id: defaultNodePoolId, name: 'default' },
                gpu: { deserved: 10 },
                rank: rankName || 'MediumLow'
            });
        });

        test('Update department rank', async () => {
            const rankName = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_20
            )
                ? 'High'
                : undefined;
            const updatedDepRank = {
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 12 },
                        rank: rankName
                    }
                ]
            } as DepartmentUpdateRequest;
            department = await orgUnitUtils.updateDepartment(department.id, updatedDepRank);

            orgUnitUtils.validateDepartmentResponse(department, department.name, {
                nodePool: { id: defaultNodePoolId, name: 'default' },
                gpu: { deserved: 12 },
                rank: rankName || 'MediumLow'
            });
        });

        test('Create project under department', async () => {
            projectCreationRequest.parentId = department.id;
            project = await orgUnitUtils.createProject(projectCreationRequest);
        });

        test('Validate department and project in backend', async () => {
            const departmentResponse = await orgUnitUtils.getDepartment(department.id);
            orgUnitUtils.validateDepartmentResponse(
                departmentResponse,
                department.name,
                department.resources[0],
                1,
                [project]
            );

            const projectResponse = await orgUnitUtils.getProject(project.id);
            orgUnitUtils.validateProjectResponse(
                { body: projectResponse },
                project.name,
                project.resources[0],
                '',
                1,
                department.name
            );
        });

        test('Try to delete department with project child', async () => {
            try {
                await orgUnitUtils.deleteDepartment(department.id);
                test.fail(true, 'Expect to fail deleting department with project child');
            } catch (e) {
                expect(e.statusCode).toBe(400);
                expect(e.response.body.message).toBe('delete department request validation failed');
                expect(e.response.body.details).toBe(
                    `department ${department.id} cannot be deleted. the department has children`
                );
            }
        });

        test('Delete project and validate its not in the department info', async () => {
            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            const departmentResponse = await orgUnitUtils.getDepartment(department.id);
            orgUnitUtils.validateDepartmentResponse(
                departmentResponse,
                department.name,
                department.resources[0],
                1,
                []
            );
            project = null;
        });
    });

    test.describe('Test departments API V2 + V1 combination', () => {
        let departmentFromOrgUnit: Department;
        let departmentFromBackend: DepartmentV1;
        let project: Project;

        test.afterAll(async () => {
            try {
                if (project) await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (e) {
                logger.error(`Failed to delete project ${project.id}, error: ${e}`);
            }
            try {
                if (departmentFromOrgUnit) await orgUnitUtils.deleteDepartment(departmentFromOrgUnit.id);
            } catch (e) {
                logger.error(`Failed to delete department ${departmentFromOrgUnit.id}, error: ${e}`);
            }
            try {
                if (departmentFromBackend)
                    await departmentUtils.deleteDepartment(Number(departmentFromBackend.id));
            } catch (e) {
                logger.error(`Failed to delete department ${departmentFromBackend.id}, error: ${e}`);
            }
        });

        test('Create department through org unit service', async () => {
            const departmentName = objectGenerator.generateNameByPrefix('department-v2-org-unit');
            departmentFromOrgUnit = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        nodePool: { id: defaultNodePoolId, name: 'default' },
                        gpu: { deserved: 1.5 }
                    }
                ]
            });

            await orgUnitUtils.waitForDepartmentInOrgUnitService(departmentFromOrgUnit.id);
        });

        test('Get department from backend and validate', async () => {
            const departmentsResponse = await departmentUtils.getDepartments();
            const departmentResponse = departmentsResponse.body.find(
                (d) => String(d.id) === String(departmentFromOrgUnit.id)
            );
            departmentUtils.validateDepartmentResponse(
                { body: departmentResponse },
                1.5,
                -1,
                {
                    deserved: 1.5,
                    maxAllowed: -1,
                    overQuotaWeight: 1.5
                },
                { deserved: null, maxAllowed: null, overQuotaWeight: null },
                {
                    deserved: null,
                    maxAllowed: null,
                    overQuotaWeight: null
                },
                1,
                {
                    deserved: 1.5,
                    maxAllowed: -1,
                    overQuotaWeight: 1.5
                }
            );
        });

        test('Create department with project child through backend', async () => {
            const departmentName = objectGenerator.generateNameByPrefix('department-v2-backend');
            const departmentFromBackend = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        gpu: {
                            deserved: 2.5,
                            limit: -1
                        }
                    }
                ]
            });
            await orgUnitUtils.waitForDepartmentInOrgUnitService(String(departmentFromBackend.id));
            projectCreationRequest.parentId = String(departmentFromBackend.id);
            project = await orgUnitUtils.createProject(projectCreationRequest);
        });
    });

    test.describe('Limit over/under subscription', () => {
        let testProjects: Project[];
        let testDepartment: Department;
        const departmentDeservedGpus = 10;
        const departmentDeservedCpus = 100;
        const departmentDeservedMemory = 1000;
        const numOfProjects = 2;
        let defaultNodePool: NodePool;

        test.beforeAll(async () => {
            defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

            await apiUtils.enableLimitOverQuotaSubscription();
            logger.info('Limit over quota subscription set to true successfully');
            await apiUtils.enableCpuQuotaFeature();
            logger.info('Cpu quota feature set to true successfully');

            // Department behavior below has an unresolved failure; revalidate before restoring the assertion.
            // Change gpu resources to departmentDeservedGpus after fixing the bug
            // Change the request to 1 gpu after fixing the bug
            const depResources = {
                nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                gpu: {
                    deserved: departmentDeservedGpus + 1,
                    limit: departmentDeservedGpus + 1,
                    overQuotaWeight: departmentDeservedGpus + 1
                },
                cpu: {
                    deserved: departmentDeservedCpus
                },
                memory: {
                    deserved: departmentDeservedMemory,
                    units: Units.Mib
                }
            } as Resources;
            const projResources = {
                nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                gpu: {
                    deserved: departmentDeservedGpus / numOfProjects
                },
                cpu: {
                    deserved: departmentDeservedCpus / numOfProjects
                },
                memory: {
                    deserved: departmentDeservedMemory / numOfProjects,
                    units: Units.Mib
                }
            } as Resources;

            const departmentName = objectGenerator.generateNameByPrefix('department-v2-limit');

            testDepartment = await orgUnitUtils.createDepartment({
                name: departmentName,
                clusterId: flowRootState.cluster.uuid,
                resources: [depResources]
            });
            expect(testDepartment).toBeTruthy();

            depResources.rank = 'MediumLow';

            orgUnitUtils.validateDepartmentResponse(testDepartment, departmentName, depResources);

            const promises = [];
            for (let i = 0; i < numOfProjects; i++) {
                const projectName = objectGenerator.generateNameByPrefix('project-v2-limit');

                logger.debug(
                    `Creating project ${projectName} in department ${testDepartment.name}/${testDepartment.id}`
                );

                promises.push(
                    orgUnitUtils.createProject({
                        name: projectName,
                        resources: [projResources],
                        clusterId: flowRootState.cluster.uuid,
                        parentId: testDepartment.id
                    })
                );
            }
            testProjects = await Promise.all(promises);
        });

        test.afterAll(async () => {
            projectsDepartmentsApi.projectsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            projectsDepartmentsApi.departmentsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            try {
                logger.debug('Deleting created projects through backend');
                const promises = [];
                for (const proj of testProjects.values()) {
                    logger.debug(`Deleting created project ${proj.name}`);
                    promises.push(orgUnitUtils.deleteProject(proj.id));
                }
                await Promise.all(promises);
            } catch (e) {
                logger.error(`Failed to delete projects. Error: ${JSON.stringify(e)}`);
            }

            try {
                logger.debug(`Deleting created department ${testDepartment.name} through backend`);
                await departmentUtils.deleteDepartment(Number(testDepartment.id));
            } catch (e) {
                logger.error(
                    `Failed to delete department ${testDepartment.name}. Error: ${JSON.stringify(e)}`
                );
            }

            await apiUtils.disableCpuQuotaFeature();
            await apiUtils.disableLimitOverQuotaSubscription();
        });

        test('Update project causing over subscription - should fail', async () => {
            logger.debug('Starting test: Update project causing over subscription - should fail');
            const proj0 = await orgUnitUtils.getProject(String(testProjects[0].id));

            const updatedProjGpu = orgUnitUtils.copyProject(proj0);
            updatedProjGpu.resources[0].gpu.deserved = updatedProjGpu.resources[0].gpu.deserved + 2;
            updatedProjGpu.resources[0].gpu.overQuotaWeight = updatedProjGpu.resources[0].gpu.deserved;
            try {
                await orgUnitUtils.updateProject(proj0.id, { resources: updatedProjGpu.resources });
                test.fail(true, 'Should fail to update project because of requested GPUs is too high');
            } catch (e) {
                if (e.message.includes('Should fail to update project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved GPUs')) {
                    throw new Error(
                        `Should fail to update project because of requested GPUs is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }

            const updatedProjCpu = orgUnitUtils.copyProject(proj0);
            updatedProjCpu.resources[0].cpu.deserved = updatedProjCpu.resources[0].cpu.deserved + 10;
            updatedProjCpu.resources[0].cpu.overQuotaWeight = updatedProjCpu.resources[0].cpu.deserved;
            try {
                await orgUnitUtils.updateProject(proj0.id, { resources: updatedProjCpu.resources });
                test.fail(true, 'Should fail to update project because of requested CPUs is too high');
            } catch (e) {
                if (e.message.includes('Should fail to update project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved CPUs')) {
                    throw new Error(
                        `Should fail to update project because of requested CPUs is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }

            const updatedProjMemory = orgUnitUtils.copyProject(proj0);
            updatedProjMemory.resources[0].memory.deserved =
                updatedProjMemory.resources[0].memory.deserved + 100;
            updatedProjMemory.resources[0].memory.overQuotaWeight =
                updatedProjMemory.resources[0].memory.deserved;
            try {
                await orgUnitUtils.updateProject(proj0.id, { resources: updatedProjMemory.resources });
                test.fail(true, 'Should fail to update project because of requested memory is too high');
            } catch (e) {
                if (e.message.includes('Should fail to update project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved memory')) {
                    throw new Error(
                        `Should fail to update project because of requested memory is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }
        });

        test('Create project causing over subscription - should fail', async () => {
            logger.debug('Starting test: Create project causing over subscription - should fail');
            const projectName = objectGenerator.generateNameByPrefix('p');

            const projResources = {
                nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                gpu: {
                    deserved: 2
                },
                cpu: {
                    deserved: 0
                },
                memory: {
                    deserved: 0,
                    units: Units.Mib
                }
            } as Resources;

            try {
                await orgUnitUtils.createProject({
                    name: projectName,
                    resources: [projResources],
                    clusterId: flowRootState.cluster.uuid,
                    parentId: testDepartment.id
                });
                test.fail(true, 'Should fail to create project because of requested GPUs is too high');
            } catch (e) {
                if (e.message.includes('Should fail to create project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot set the deserved GPUs')) {
                    throw new Error(
                        `Should fail to create project because of requested GPUs is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }

            projResources.gpu.deserved = 1;
            projResources.cpu.deserved = 10;
            try {
                await orgUnitUtils.createProject({
                    name: projectName,
                    resources: [projResources],
                    clusterId: flowRootState.cluster.uuid,
                    parentId: testDepartment.id
                });
                test.fail(true, 'Should fail to create project because of requested CPUs is too high');
            } catch (e) {
                if (e.message.includes('Should fail to create project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot set the deserved CPUs')) {
                    throw new Error(
                        `Should fail to create project because of requested CPUs is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }

            projResources.cpu.deserved = 0;
            projResources.memory.deserved = 10;
            try {
                await orgUnitUtils.createProject({
                    name: projectName,
                    resources: [projResources],
                    clusterId: flowRootState.cluster.uuid,
                    parentId: testDepartment.id
                });
                test.fail(true, 'Should fail to create project because of requested memory is too high');
            } catch (e) {
                if (e.message.includes('Should fail to create project')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot set the deserved memory')) {
                    throw new Error(
                        `Should fail to create project because of requested memory is too high, e: ${JSON.stringify(e)}`
                    );
                }
            }
        });

        test('Update department causing under subscription - should fail', async () => {
            logger.debug('Starting test: Update project causing over subscription - should fail');
            const dep0 = await orgUnitUtils.getDepartment(testDepartment.id);

            const updatedDepGpu = {
                resources: [
                    {
                        nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                        gpu: { deserved: testDepartment.resources[0].gpu.deserved - 2 }
                    }
                ]
            } as DepartmentUpdateRequest;

            try {
                await orgUnitUtils.updateDepartment(dep0.id, updatedDepGpu);
                test.fail(true, 'Should fail to update department because of requested GPUs is too low');
            } catch (e) {
                if (e.message.includes('Should fail to update department')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved GPUs')) {
                    throw new Error(
                        `Should fail to update department because of requested GPUs is too low, e: ${JSON.stringify(e)}`
                    );
                }
            }

            const updatedDepCpu = {
                resources: [
                    {
                        nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                        gpu: { deserved: testDepartment.resources[0].gpu.deserved },
                        cpu: { deserved: testDepartment.resources[0].cpu.deserved - 1 }
                    }
                ]
            } as DepartmentUpdateRequest;

            try {
                await orgUnitUtils.updateDepartment(dep0.id, updatedDepCpu);
                test.fail(true, 'Should fail to update department because of requested CPUs is too low');
            } catch (e) {
                if (e.message.includes('Should fail to update department')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved CPUs')) {
                    throw new Error(
                        `Should fail to update department because of requested CPUs is too low, e: ${JSON.stringify(e)}`
                    );
                }
            }

            const updatedDepMemory = {
                resources: [
                    {
                        nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name },
                        gpu: { deserved: testDepartment.resources[0].gpu.deserved },
                        memory: {
                            deserved: testDepartment.resources[0].memory.deserved - 1,
                            units: Units.Mib
                        }
                    }
                ]
            } as DepartmentUpdateRequest;

            try {
                await orgUnitUtils.updateDepartment(dep0.id, updatedDepMemory);
                test.fail(true, 'Should fail to update department because of requested memory is too low');
            } catch (e) {
                if (e.message.includes('Should fail to update department')) {
                    throw e;
                }
                expect(e.statusCode).toBe(StatusCodes.BAD_REQUEST);
                if (!e.message.includes('Cannot modify the deserved memory')) {
                    throw new Error(
                        `Should fail to update department because of requested memory is too low, e: ${JSON.stringify(e)}`
                    );
                }
            }
        });
    });
});
