/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    crds,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    projectUtils,
    versionUtils,
    waitUtils
} from '../index';
import http from 'http';
import { projectsDepartmentsApi, orgUnitService } from '../../api';
import {
    AssignedResources,
    NodePoolAssignedResourcesCreate,
    NodePoolAssignedResourcesCreateNodePool
} from '../../api/apis/projectsDepartmentsApi';
import { Department } from '../../api/apis/projectsDepartmentsApi';
import { Department as DepartmentOU } from '../../api/apis/orgUnitApi';
import { getOrgUnitUpdatedAt, hasOrgUnitSyncTimestamp, resolveOrgUnitCrds } from './OrgUnitCrdFields';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { kubeCli } from '../../cli';
import {
    AssignedResourcesCpu,
    AssignedResourcesGpu,
    AssignedResourcesMemory,
    DepartmentAccessControl,
    NodePoolAssignedResourcesCreateCpu,
    NodePoolAssignedResourcesCreateGpu,
    NodePoolAssignedResourcesCreateMemory
} from '../../api/apis/projectsDepartmentsApi';
import { validatePropertyIfNotNull } from './testing/helpersUtils';
import { SoftAssert } from '../SoftAssert';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { KubernetesObject } from '@kubernetes/client-node';

export class DepartmentUtils {
    async waitForDepartmentCrd(departmentName: string, timeout = 60, desiredResult = true): Promise<object> {
        await resolveOrgUnitCrds();
        return await waitUtils.waitForCrd(crds.departmentCrd, departmentName, '', timeout, desiredResult);
    }

    async validateDepartmentQueueCrd(
        queueName: string,
        department: Department,
        timeout = 60,
        desiredResult = true
    ): Promise<void> {
        await this.validateQueueCrdOfDepartment(
            queueName,
            department.id.toString(),
            department.name,
            department.resources.gpu.deserved,
            department.resources.gpu.maxAllowed,
            department.resources.gpu.overQuotaWeight,
            department.resources.cpu.deserved,
            department.resources.cpu.maxAllowed,
            department.resources.cpu.overQuotaWeight,
            department.resources.memory.deserved,
            department.resources.memory.maxAllowed,
            department.resources.memory.overQuotaWeight,
            undefined,
            projectUtils.mapPriorityToNumber(department.resources[0]?.rank) || undefined,
            timeout,
            desiredResult
        );
    }

    async validateQueueCrdOfDepartment(
        queueName: string,
        departmentId: string,
        departmentName: string,
        gpuDeserved: number,
        gpuLimit: number,
        gpuOverQuotaWeight: number,
        cpuDeserved: number = null,
        cpuLimit: number = null,
        cpuOverQuotaWeight: number = null,
        memoryDeserved: number = null,
        memoryLimit: number = null,
        memoryOverQuotaWeight: number = null,
        desiredQuota: number = undefined,
        priority: number = undefined,
        timeout = 60,
        desiredResult = true
    ): Promise<void> {
        if (!versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)) {
            return;
        }
        if (desiredQuota) {
            await orgUnitUtils.waitForQueueToHaveGpuQuota(queueName, desiredQuota);
        }
        const queueCrd = await waitUtils.waitForCrd(crds.queueCrd, queueName, '', timeout, desiredResult);
        expect(queueCrd).property('metadata');
        expect(queueCrd).property('spec');
        expect(queueCrd['metadata']).property('labels');
        expect(queueCrd['metadata']['labels']).property('runai/department-name').equal(departmentName);
        expect(queueCrd['spec']).property('displayName').equal(departmentName);
        expect(queueCrd['spec']).property('resources');

        validatePropertyIfNotNull(queueCrd['spec']['resources']['gpu'], gpuDeserved, 'quota', 0);

        validatePropertyIfNotNull(queueCrd['spec']['resources']['gpu'], gpuLimit, 'limit', -1);

        validatePropertyIfNotNull(
            queueCrd['spec']['resources']['gpu'],
            gpuOverQuotaWeight,
            'overQuotaWeight',
            versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)
                ? 1
                : gpuDeserved
        );

        validatePropertyIfNotNull(queueCrd['spec']['resources']['cpu'], cpuLimit, 'limit', -1);
        validatePropertyIfNotNull(queueCrd['spec']['resources']['cpu'], cpuDeserved, 'quota', -1);
        validatePropertyIfNotNull(
            queueCrd['spec']['resources']['cpu'],
            cpuOverQuotaWeight,
            'overQuotaWeight',
            1
        );

        validatePropertyIfNotNull(queueCrd['spec']['resources']['memory'], memoryLimit, 'limit', -1);
        validatePropertyIfNotNull(queueCrd['spec']['resources']['memory'], memoryDeserved, 'quota', -1);
        validatePropertyIfNotNull(
            queueCrd['spec']['resources']['memory'],
            memoryOverQuotaWeight,
            'overQuotaWeight',
            1
        );
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
            validatePropertyIfNotNull(queueCrd['spec'], priority, 'priority', 100);
        }
    }

    async waitForDepartmentCrdUpdated(
        departmentName: string,
        updatedAt: Date,
        timeout = 90
    ): Promise<Department> {
        let departmentCrd;
        const departmentUpdatedStr = updatedAt.toISOString().split('.')[0] + 'Z';
        const updatedDepartmentCrd = async (): Promise<boolean> => {
            departmentCrd = await this.waitForDepartmentCrd(departmentName);
            const crdUpdatedAt = getOrgUnitUpdatedAt(departmentCrd);
            logger.info(
                `department crd "${departmentName}" updatedAt: ${crdUpdatedAt}, waiting for: ${departmentUpdatedStr}`
            );
            if (crdUpdatedAt) {
                const crdUpdatedStr = crdUpdatedAt.split('.')[0] + 'Z';
                return hasOrgUnitSyncTimestamp(departmentCrd) && crdUpdatedStr >= departmentUpdatedStr;
            }
            logger.info(`department crd "${departmentName}" does not have updatedAt yet`);
            return false;
        };

        await waitUtils.waitForCondition(
            updatedDepartmentCrd,
            `wait for department crd "${departmentName}" to be updated at ${departmentUpdatedStr}`,
            true,
            timeout
        );
        return departmentCrd;
    }

    async createBackendDepartment(
        deservedGpus = 0,
        maxAllowedGpus = 0,
        resources?: AssignedResources,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body?: DepartmentOU }> {
        const departmentName = objectGenerator.generateNameByPrefix('d');
        return this.createBackendDepartmentWithName(
            departmentName,
            deservedGpus,
            maxAllowedGpus,
            resources,
            accessToken
        );
    }

    async createBackendDepartmentWithName(
        departmentName: string,
        deservedGpus = 0,
        maxAllowedGpus = 0,
        resources?: AssignedResources,
        accessToken = flowRootState.users.system_administrator.accessToken,
        additionalResourcesForNonDefaultNodePools?: NodePoolAssignedResourcesCreate[]
    ): Promise<{ response: http.IncomingMessage; body?: DepartmentOU }> {
        orgUnitService.departmentApi.accessToken = accessToken;
        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
        const nodePoolResources = [
            {
                cpu: resources?.cpu,
                gpu: resources
                    ? resources.gpu
                    : { deserved: deservedGpus, maxAllowed: maxAllowedGpus, overQuotaWeight: deservedGpus },
                memory: resources?.memory,
                nodePool: {
                    id: defaultNodePool.id,
                    name: defaultNodePool.name
                } as NodePoolAssignedResourcesCreateNodePool
            }
        ];
        if (additionalResourcesForNonDefaultNodePools) {
            for (const additionalResources of additionalResourcesForNonDefaultNodePools) {
                nodePoolResources.push({
                    cpu: additionalResources.cpu,
                    gpu: additionalResources.gpu,
                    memory: additionalResources.memory,
                    nodePool: additionalResources.nodePool
                });
            }
        }
        const createRequest = orgUnitUtils.getDepartmentCreateRequestFromV1(
            {
                name: departmentName,
                deservedGpus: deservedGpus,
                maxAllowedGpus: maxAllowedGpus,
                resources: resources,
                nodePoolsResources: nodePoolResources
            },
            flowRootState.cluster.uuid,
            defaultNodePool
        );
        const res = await orgUnitService.departmentApi.createDepartment(createRequest);
        return { response: res.response, body: res.body as DepartmentOU };
    }

    async createBackendDepartmentAndValidate(
        deservedGpus: number,
        maxAllowedGpus: number,
        resources?: AssignedResources
    ): Promise<{ departmentName: string; departmentId: number }> {
        const createResponse = await this.createBackendDepartment(deservedGpus, maxAllowedGpus, resources);
        expect(createResponse.response.statusCode.valueOf()).equal(StatusCodes.CREATED);
        expect(createResponse.body).property('id');

        return {
            departmentName: createResponse.body.name,
            departmentId: Number(createResponse.body.id)
        };
    }

    async updateDepartment(
        departmentId: number,
        departmentSpec: Department
    ): Promise<{ response: http.IncomingMessage; body: DepartmentOU }> {
        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
        return await orgUnitService.departmentApi.updateDepartment(
            departmentId.toString(),
            orgUnitUtils.getDepartmentUpdateRequestFromV1(departmentSpec, defaultNodePool)
        );
    }

    copyDepartment(department: Department): Department {
        const depCopy = Object.assign({}, JSON.parse(JSON.stringify(department)) as Department) as Department;
        depCopy.createdAt = department.createdAt;
        return depCopy;
    }

    async updateDepartmentAndValidate(
        departmentId: number,
        departmentSpec: Department
    ): Promise<DepartmentOU> {
        const res = await this.updateDepartment(departmentId, departmentSpec);
        expect(res.response.statusCode.valueOf()).equal(StatusCodes.OK);
        expect(res.body.id).equal(departmentId);
        return res.body;
    }

    async getBackendDepartment(departmentId: string): Promise<DepartmentOU> {
        return await orgUnitUtils.getDepartment(departmentId);
    }

    async getBackendDepartmentAndValidate(departmentId: string): Promise<DepartmentOU> {
        const getRes = await this.getBackendDepartment(departmentId);
        expect(getRes.id).equal(departmentId);
        return getRes;
    }

    async deleteDepartment(departmentId: number): Promise<void> {
        return await orgUnitUtils.deleteDepartment(departmentId.toString());
    }

    async getDepartments(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: Department[] }> {
        projectsDepartmentsApi.departmentsApi.accessToken = accessToken;
        return await projectsDepartmentsApi.departmentsApi.getDepartments(flowRootState.cluster.uuid);
    }

    async waitForDepartmentInBackend(departmentName: string): Promise<void> {
        const doesDepartmentExist = async (): Promise<boolean> => {
            try {
                const departments = await this.getDepartments();
                return departments.body.some((department) => department.name === departmentName);
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesDepartmentExist,
            `waitForDepartmentInBackend department name: ${departmentName}`,
            true,
            30
        );
    }

    async getDefaultDepartmentByCluster(
        clusterId: string = flowRootState.cluster.uuid,
        projectDepartmentAccessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Department> {
        projectsDepartmentsApi.departmentsApi.accessToken = projectDepartmentAccessToken;
        return (await projectsDepartmentsApi.departmentsApi.getDepartments(clusterId)).body.find(
            (defaultDepartment) => defaultDepartment.name === 'default'
        );
    }

    async deleteDepartmentCleanup(departmentId?: number): Promise<void> {
        if (!departmentId) {
            return;
        }

        try {
            return await this.deleteDepartment(departmentId);
        } catch (e) {
            logger.warn(`Could not cleanup, unable to delete department ${departmentId}, error: ${e}`);
        }
    }

    async waitForDepartmentCrdToBeDeleted(departmentName: string): Promise<void> {
        const doesDepartmentExist = async (): Promise<boolean> => {
            try {
                await kubeCli.getCrd(crds.departmentCrd, departmentName, undefined);
                return true;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesDepartmentExist,
            `waitForDepartmentCrdToBeDeleted department name: ${departmentName}`,
            false,
            40
        );
    }

    async setDepartmentAdmin(
        accessTokens: string,
        clusterId: string,
        departmentAdminEmail?: string
    ): Promise<DepartmentAccessControl> {
        projectsDepartmentsApi.departmentsApi.accessToken = accessTokens;
        const defaultDepartmentId = (
            await departmentUtils.getDefaultDepartmentByCluster(flowRootState.cluster.uuid)
        ).id;
        const response = await projectsDepartmentsApi.departmentsApi.updateDepartmentAdmins(
            clusterId,
            defaultDepartmentId,
            {
                departmentAdmins: [
                    flowRootState.users.department_admin.email || departmentAdminEmail
                ]
            }
        );
        logger.debug('Department admin set successfully.');
        return response.body;
    }

    async deleteDepartmentAndAllItsProjects(departmentId: number): Promise<void> {
        projectsDepartmentsApi.projectsApi.accessToken =
            flowRootState.users.system_administrator.accessToken;

        const department = await this.getBackendDepartmentAndValidate(departmentId.toString());
        const projectsByCluster = await orgUnitUtils.getProjectsByClusterId(department.clusterId);
        const projects = [];
        projectsByCluster.forEach((p) => {
            if (p.parentId === department.id) {
                projects.push(p);
            }
        });

        const promisesToAwait = [];
        for (const project of projects) {
            promisesToAwait.push(orgUnitUtils.deleteProject(project.id));
            promisesToAwait.push(projectUtils.waitForProjectCrdToBeDeleted(project.name));
        }
        await Promise.all(promisesToAwait);
        await this.deleteDepartment(departmentId);
    }

    async departmentsCleanup(): Promise<void> {
        const departments = await this.getDepartments();
        for (const department of departments.body) {
            if (department.name === 'default') {
                continue;
            }
            await this.deleteDepartmentCleanup(department.id);
        }
        logger.debug('Departments cleanup completed.');
    }

    validateDepartmentResponse(
        response: { response?: http.IncomingMessage; body?: Department },
        deservedGpus = 0,
        maxAllowedGpus = -1,
        resourcesGpu: AssignedResourcesGpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowedGpus,
            overQuotaWeight: deservedGpus
        },
        resourcesCpu: AssignedResourcesCpu = { deserved: null, maxAllowed: null, overQuotaWeight: null },
        resourcesMemory: AssignedResourcesMemory = {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null
        },
        numberOfNodePools = 1,
        resourcesNodePoolGpu: NodePoolAssignedResourcesCreateGpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowedGpus,
            overQuotaWeight: deservedGpus
        },
        resourcesNodePoolCpu: NodePoolAssignedResourcesCreateCpu = {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null
        },
        resourcesNodePoolMemory: NodePoolAssignedResourcesCreateMemory = {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null
        },
        projects: Array<string> = [],
        projectsDeservedGpus = null,
        departmentName?: string
    ): void {
        const softAssert = new SoftAssert();

        if (response.response) {
            softAssert.assert(
                response.response.statusCode.valueOf() === StatusCodes.OK,
                `Expected department status code to be ${StatusCodes.OK}`
            );
        }

        if (response.body) {
            softAssert.assert(response.body.id !== undefined, 'Expected department id to be defined');
            softAssert.assert(
                !departmentName || response.body.name === departmentName,
                `Expected department name to be ${departmentName}, got instead ${response.body.name}`
            );

            softAssert.assert(
                response.body.nodePoolsResources.length === numberOfNodePools,
                `Expected department number of node pools to be ${numberOfNodePools}, got instead ${response.body.nodePoolsResources.length}`
            );
            softAssert.assert(
                response.body.nodePoolsResources[0].nodePool.name === 'default',
                `Expected department first node pool name to be default, got instead ${response.body.nodePoolsResources[0].nodePool.name}`
            );

            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateGpu>(
                response.body.nodePoolsResources[0].gpu,
                resourcesNodePoolGpu,
                'Expected department node pool GPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateCpu>(
                response.body.nodePoolsResources[0].cpu,
                resourcesNodePoolCpu,
                'Expected department node pool CPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateMemory>(
                response.body.nodePoolsResources[0].memory,
                resourcesNodePoolMemory,
                'Expected department node pool memory resources to match'
            );

            softAssert.assert(
                response.body.createdAt !== undefined && response.body.createdAt !== null,
                `Expected department createdAt to be present (not null or undefined), instead got ${response.body.createdAt}`
            );

            try {
                expect(response.body.projects).to.deep.equal(
                    projects,
                    `Expected department projects: ${response.body.projects} to deep equal ${projects}`
                );
            } catch (error) {
                softAssert.assert(false, `Expected projects to match: ${error.message}`);
            }

            softAssert.assertNonStrictEqual(
                response.body.projectsDeservedGpus,
                projectsDeservedGpus,
                `Expected department projectsDeservedGpus to match, got: ${response.body.projectsDeservedGpus} instead of ${projectsDeservedGpus}`
            );
        }

        softAssert.throwErrorIfFailed();
    }

    validateDepartmentWithGpusResponse(
        response: { response?: http.IncomingMessage; body?: Department },
        deservedGpus = 0,
        maxAllowedGpus = -1,
        numberOfNodePools = 1,
        deservedCpu = null,
        deservedMemory = null,
        projects: Array<string> = [],
        projectsDeservedGpus = null,
        departmentName?: string,
        resourcesNodePoolCpuMaxAllowed = null,
        resourcesNodePoolMemoryMaxAllowed = null
    ): void {
        const resourcesGpu: AssignedResourcesGpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowedGpus,
            overQuotaWeight: deservedGpus
        };
        const resourcesCpu: AssignedResourcesCpu = {
            deserved: deservedCpu,
            maxAllowed: deservedCpu,
            overQuotaWeight: deservedCpu
        };
        const resourcesMemory: AssignedResourcesMemory = {
            deserved: deservedMemory,
            maxAllowed: deservedMemory,
            overQuotaWeight: deservedMemory
        };
        const resourcesNodePoolGpu: NodePoolAssignedResourcesCreateGpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowedGpus,
            overQuotaWeight: deservedGpus
        };
        const resourcesNodePoolCpu: NodePoolAssignedResourcesCreateCpu = {
            deserved: deservedCpu,
            maxAllowed: resourcesNodePoolCpuMaxAllowed,
            overQuotaWeight: deservedCpu
        };
        const resourcesNodePoolMemory: NodePoolAssignedResourcesCreateMemory = {
            deserved: deservedMemory,
            maxAllowed: resourcesNodePoolMemoryMaxAllowed,
            overQuotaWeight: deservedMemory
        };
        this.validateDepartmentResponse(
            response,
            deservedGpus,
            maxAllowedGpus,
            resourcesGpu,
            resourcesCpu,
            resourcesMemory,
            numberOfNodePools,
            resourcesNodePoolGpu,
            resourcesNodePoolCpu,
            resourcesNodePoolMemory,
            projects,
            projectsDeservedGpus,
            departmentName
        );
    }

    validateDepartmentAccessRulesResponse(
        response: { response: http.IncomingMessage; body: DepartmentAccessControl },
        departmentId: number,
        departmentAdmins: string[]
    ): void {
        expect(response.response.statusCode.valueOf()).equal(StatusCodes.CREATED);
        expect(response.body.departmentId).equal(departmentId);
        expect(response.body.departmentAdmins).to.have.members(departmentAdmins);
    }

    async getDepartmentId(deptName: string, timeout: number = 10): Promise<number> {
        let deptId: number;

        const idShown = async (): Promise<boolean> => {
            const deptCrd = (await this.waitForDepartmentCrd(deptName)) as KubernetesObject;
            if (deptCrd.metadata?.labels?.['runai/department-id'] !== undefined) {
                deptId = Number(deptCrd.metadata?.labels?.['runai/department-id']);
                return true;
            }

            return false;
        };

        await waitUtils.waitForCondition(idShown, 'wait for department id updated in the CRD', true, timeout);

        return deptId;
    }
}
