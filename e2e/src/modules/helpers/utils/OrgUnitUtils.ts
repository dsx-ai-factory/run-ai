/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BatchPatchDepartmentResourcesResponse,
    BatchPatchNodePoolResourcesRequest,
    BatchPatchProjectResourcesResponse,
    Department,
    DepartmentCreationRequest,
    DepartmentForList,
    DepartmentNodePoolResourcesFilterSortFields,
    DepartmentProjectChild,
    DepartmentSumResponse,
    DepartmentUpdateRequest,
    GetDepartmentNodePoolResources200Response,
    GetProjectNodePoolResources200Response,
    GetProjectsRanking200Response,
    MetricsResponse,
    NodeType,
    NodeTypeCreateRequest,
    OrgUnitMetricType,
    Project,
    ProjectCreationRequest,
    ProjectNodePoolResourcesFilterSortFields,
    ProjectPhase,
    ProjectUpdateRequest,
    RankingMetricType,
    RankingOrderType,
    Resources,
    ResourcesCpu,
    ResourcesGpu,
    ResourcesMemory,
    ResourcesNodePool,
    ResourcesNullable,
    SumResponse,
    Units
} from '../../api/apis/orgUnitApi';
import { orgUnitService } from '../../api';
import { flowRootState } from '../config/config';
import { Utils } from './utils';
import { logger } from '../logger/logger';
import { crds, departmentUtils, projectUtils, versionUtils, waitUtils } from '../index';
import http from 'http';
import { SoftAssert } from '../SoftAssert';
import { StatusCodes } from 'http-status-codes';
import { kubeCli } from '../../cli';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { expect } from 'chai';
import { NodePool } from '../../api/apis/backendClusterApi';

const DEFAULT_DEPARTMENT_GPU_NUMBER = 10000;

// Mirrors services.MinOverQuotaWeight in the control-plane org-unit-service.
// When adjustable weights are off and deserved == 0, the API and queue CRD
// return this floor instead of 0 so the scheduler can still fairshare.
export const MIN_OVER_QUOTA_WEIGHT = 0.01;

export const expectedOQWhenAwOff = (deserved: number): number =>
    deserved === 0 ? MIN_OVER_QUOTA_WEIGHT : deserved;

export class OrgUnitUtils {
    ///////////////// Resources constructors ///////
    createResources(
        cpuQuotaFeatureOn: boolean,
        overQuotaWeightFeatureOn: boolean,
        nodePool: ResourcesNodePool,
        overQuotaWeight: number,
        deservedGpu: number,
        deservedCpu: number,
        deservedMemory: number
    ): Resources {
        const projectResources = new Resources();
        projectResources.nodePool = nodePool;
        projectResources.gpu = {
            deserved: deservedGpu,
            limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedGpu,
            overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : undefined
        };
        if (cpuQuotaFeatureOn) {
            projectResources.cpu = {
                deserved: deservedCpu,
                limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedCpu,
                overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : undefined
            };
            projectResources.memory = {
                deserved: deservedMemory,
                limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedMemory,
                overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : undefined,
                units: Units.Mib
            };
        }
        return projectResources;
    }

    getExpectedResources(
        cpuQuotaFeatureOn: boolean,
        overQuotaWeightFeatureOn: boolean,
        nodePool: ResourcesNodePool,
        overQuotaWeight: number,
        deservedGpu: number,
        deservedCpu: number,
        deservedMemory: number
    ): Resources {
        let resources = {
            nodePool: nodePool,
            gpu: {
                deserved: deservedGpu,
                limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedGpu,
                overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : expectedOQWhenAwOff(deservedGpu)
            }
        };
        if (cpuQuotaFeatureOn) {
            resources = Object.assign(resources, {
                cpu: {
                    deserved: deservedCpu,
                    limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedCpu,
                    overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : expectedOQWhenAwOff(deservedCpu)
                },
                memory: {
                    deserved: deservedMemory,
                    limit: !overQuotaWeightFeatureOn && overQuotaWeight === 1 ? -1 : deservedMemory,
                    overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : expectedOQWhenAwOff(deservedMemory),
                    units: Units.Mib
                }
            });
        }

        return resources;
    }

    createResourcesGPUOnly(
        nodePool: ResourcesNodePool,
        overQuotaWeightFeatureOn: boolean,
        deservedGpu: number,
        overQuotaAllowed = true,
        overQuotaWeight = 2
    ): Resources {
        const projectResources = new Resources();
        projectResources.nodePool = nodePool;
        projectResources.gpu = {
            deserved: deservedGpu,
            limit: overQuotaAllowed ? -1 : deservedGpu,
            overQuotaWeight: overQuotaWeightFeatureOn ? overQuotaWeight : undefined
        };
        return projectResources;
    }

    createResourcesWithOverQuotaAllowed(
        nodePool: ResourcesNodePool,
        overQuotaWeightFeatureOn: boolean,
        deservedGpu: number
    ): Resources {
        return this.createResourcesGPUOnly(nodePool, overQuotaWeightFeatureOn, deservedGpu);
    }

    createResourcesWithOverQuotaNotAllowed(
        nodePool: ResourcesNodePool,
        overQuotaWeightFeatureOn: boolean,
        deservedGpu: number
    ): Resources {
        return this.createResourcesGPUOnly(nodePool, overQuotaWeightFeatureOn, deservedGpu, false);
    }

    ///////////////// API Wrappers /////////////////
    async getProject(projectId: string, accessToken?: string): Promise<Project> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        const res = await orgUnitService.projectApi.getProject(projectId);
        return { ...res.body } as Project;
    }

    async createProject(
        projectCreationRequest: ProjectCreationRequest,
        accessToken?: string
    ): Promise<Project> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.projectApi.createProject(projectCreationRequest)).body;
    }

    async updateProject(
        projectId: string,
        projectUpdateRequest: ProjectUpdateRequest,
        accessToken?: string
    ): Promise<Project> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.projectApi.updateProject(projectId, projectUpdateRequest)).body;
    }

    async patchProjectResources(
        projectId: string,
        resources: Array<Resources>,
        accessToken?: string
    ): Promise<Array<Resources>> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.projectApi.patchProjectResources(projectId, resources)).body;
    }

    async updateProjectResources(
        projectId: string,
        resources: Array<Resources>,
        accessToken?: string
    ): Promise<Array<Resources>> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.projectApi.updateProjectResources(projectId, resources)).body;
    }

    async deleteProject(projectId: string, accessToken?: string): Promise<void> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        await orgUnitService.projectApi.deleteProject(projectId);
    }

    async deleteProjectAndNamespace(
        projectId: string,
        projectName: string,
        accessToken?: string
    ): Promise<void> {
        await this.deleteProject(projectId, accessToken);
        const relatedNamespaceName = Utils.namespaceFromProjectName(projectName);
        try {
            await kubeCli.CoreV1Api.deleteNamespace({ name: relatedNamespaceName });
        } catch (e) {
            // namespace deletion is not critical, so we only log the error if it's not a 404
            if (e.body?.code !== 404) {
                logger.error(
                    `Failed deleting namespace ${relatedNamespaceName}, error: ${JSON.stringify(e)}`
                );
            }
        }
        logger.info(`Project ${projectName} and namespace ${relatedNamespaceName} deleted`);
    }

    async deleteAllProjectsInCluster(clusterId?: string, accessToken?: string): Promise<void> {
        const projects = await this.getProjectsByClusterId(clusterId, accessToken);
        for (const project of projects) {
            await this.deleteProject(project.id, accessToken);
            logger.info(`Deleted project ${project.name} (${project.id})`);
        }
    }

    async getProjectNameById(projectId: string, accessToken?: string): Promise<string> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        const res = await orgUnitService.projectApi.getProject(projectId);
        return res.body.name;
    }

    async getProjectsByClusterId(clusterId?: string, accessToken?: string): Promise<Array<Project>> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        const clusterIdToFilterBy = clusterId ?? flowRootState.cluster.uuid;
        const resBody = (await orgUnitService.projectApi.getProjects([`clusterId==${clusterIdToFilterBy}`]))
            .body;
        if (!resBody || !resBody.projects) {
            return [];
        }
        return resBody.projects;
    }

    async getDepartments(accessToken?: string): Promise<Array<DepartmentForList>> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        const resBody = (
            await orgUnitService.departmentApi.getDepartments(['clusterId==' + flowRootState.cluster.uuid])
        ).body;
        if (!resBody || !resBody.departments) {
            return [];
        }
        return resBody.departments;
    }

    async getDefaultDepartment(accessToken?: string): Promise<Department> {
        const departments = await this.getDepartments(accessToken);
        const defaultDepartment = departments.find((dep) => dep.name === 'default');
        if (!defaultDepartment) {
            throw new Error('Default department not found');
        }
        return defaultDepartment as Department;
    }

    async validateAndSetDefaultDepartmentGpu(accessToken?: string): Promise<void> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        const defaultDepartment = (await this.getDepartments()).find((dep) => dep.name === 'default');
        if (!defaultDepartment) {
            return;
        }
        const resourcesForUpdate: Resources[] = [];
        defaultDepartment.resources.forEach((npResources) => {
            if (!npResources.gpu.deserved || npResources.gpu.deserved === -1) {
                npResources.gpu.deserved = DEFAULT_DEPARTMENT_GPU_NUMBER;
                npResources.gpu.overQuotaWeight = DEFAULT_DEPARTMENT_GPU_NUMBER;
                resourcesForUpdate.push(npResources);
            }
        });
        if (resourcesForUpdate.length > 0) {
            const updateResourcesResponse = await orgUnitService.departmentApi.patchDepartmentResources(
                defaultDepartment.id,
                resourcesForUpdate
            );
            expect(updateResourcesResponse.response.statusCode).equal(StatusCodes.OK);
        }
    }

    async createDepartment(
        departmentCreationRequest: DepartmentCreationRequest,
        accessToken?: string
    ): Promise<Department> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        await this.validateAndSetDefaultDepartmentGpu(accessToken);
        return (await orgUnitService.departmentApi.createDepartment(departmentCreationRequest)).body;
    }

    getResourcesFromV1(v1Request: any, defaultNodePool: NodePool = null): ResourcesNullable[] {
        const resources: ResourcesNullable[] = [];
        const nodePoolResources = v1Request.nodePoolsResources || v1Request.nodePoolResources;
        if (nodePoolResources) {
            nodePoolResources.forEach((v1Resource) => {
                const v2Resources: ResourcesNullable = {
                    nodePool: {
                        id: v1Resource?.nodePool?.id?.toString() || defaultNodePool?.id,
                        name: v1Resource?.nodePool?.name?.toString() || defaultNodePool?.name
                    },
                    gpu: {
                        deserved: v1Resource?.gpu?.deserved,
                        limit: v1Resource?.gpu?.maxAllowed,
                        overQuotaWeight: v1Resource?.gpu?.overQuotaWeight
                    }
                };

                if (v1Resource?.cpu) {
                    v2Resources.cpu = {
                        deserved: v1Resource.cpu.deserved,
                        limit: v1Resource.cpu.maxAllowed,
                        overQuotaWeight: v1Resource.cpu.overQuotaWeight
                    };
                }

                if (v1Resource?.memory) {
                    v2Resources.memory = {
                        deserved: v1Resource.memory.deserved,
                        limit: v1Resource.memory.maxAllowed,
                        overQuotaWeight: v1Resource.memory.overQuotaWeight,
                        units: Units.Mib
                    };
                }

                resources.push(v2Resources);
            });
        } else if (v1Request.resources) {
            const v2Resources: ResourcesNullable = {
                nodePool: {
                    id: v1Request.resources['nodePool']?.id?.toString() || defaultNodePool?.id,
                    name: v1Request.resources['nodePool']?.name?.toString() || defaultNodePool?.name
                },
                gpu: {
                    deserved: v1Request.resources['gpu'].deserved,
                    limit: v1Request.resources['gpu'].maxAllowed,
                    overQuotaWeight: v1Request.resources['gpu'].overQuotaWeight
                }
            };

            if (v1Request.resources['cpu']) {
                v2Resources.cpu = {
                    deserved: v1Request.resources['cpu'].deserved,
                    limit: v1Request.resources['cpu'].maxAllowed,
                    overQuotaWeight: v1Request.resources['cpu'].overQuotaWeight
                };
            }

            if (v1Request.resources['memory']) {
                v2Resources.memory = {
                    deserved: v1Request.resources['memory'].deserved,
                    limit: v1Request.resources['memory'].maxAllowed,
                    overQuotaWeight: v1Request.resources['memory'].overQuotaWeight,
                    units: Units.Mib
                };
            }

            resources.push(v2Resources);
        } else {
            const v2Resources: ResourcesNullable = {
                nodePool: {
                    id: defaultNodePool?.id?.toString(),
                    name: defaultNodePool?.name
                },
                gpu: {
                    deserved: v1Request.deservedGpus,
                    limit: v1Request.maxAllowedGpus,
                    overQuotaWeight: v1Request.deservedGpus
                }
            };
            resources.push(v2Resources);
        }
        return resources;
    }

    getDepartmentCreateRequestFromV1(
        v1Request: any,
        clusterId: string,
        defaultNodePool: NodePool
    ): DepartmentCreationRequest {
        const resources = this.getResourcesFromV1(v1Request, defaultNodePool);
        return {
            clusterId: clusterId,
            name: v1Request.name,
            resources: resources
        };
    }

    getDepartmentUpdateRequestFromV1(v1Request: any, defaultNodePool: NodePool): DepartmentUpdateRequest {
        const resources = this.getResourcesFromV1(v1Request, defaultNodePool);
        return {
            resources: resources
        };
    }

    async deleteDepartment(departmentId: string, accessToken?: string): Promise<void> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        await orgUnitService.departmentApi.deleteDepartment(departmentId);
    }

    async getDepartment(departmentId: string, accessToken?: string): Promise<Department> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        const res = await orgUnitService.departmentApi.getDepartment(departmentId);
        return { ...res.body } as Department;
    }

    async updateDepartment(
        departmentId: string,
        departmentUpdateRequest: DepartmentUpdateRequest,
        accessToken?: string
    ): Promise<Department> {
        await this.validateAndSetDefaultDepartmentGpu();
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.departmentApi.updateDepartment(departmentId, departmentUpdateRequest))
            .body;
    }

    async patchDepartmentResources(
        departmentId: string,
        resources: Array<Resources>,
        accessToken?: string
    ): Promise<Array<Resources>> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.departmentApi.patchDepartmentResources(departmentId, resources)).body;
    }

    async updateDepartmentResources(
        departmentId: string,
        resources: Array<Resources>,
        accessToken?: string
    ): Promise<Array<Resources>> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.departmentApi.updateDepartmentResources(departmentId, resources)).body;
    }

    async createNodeType(
        nodeTypeCreateRequest: NodeTypeCreateRequest,
        accessToken?: string
    ): Promise<NodeType> {
        orgUnitService.nodeTypeApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.nodeTypeApi.createNodeType(nodeTypeCreateRequest)).body;
    }

    async deleteNodeType(nodeTypeId: string, accessToken?: string): Promise<http.IncomingMessage> {
        orgUnitService.nodeTypeApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (await orgUnitService.nodeTypeApi.deleteNodeType(nodeTypeId)).response;
    }

    async getProjectMetrics(
        projectId: string,
        metricTypes: Array<OrgUnitMetricType>,
        start: Date,
        end: Date,
        samples: number,
        accessToken?: string
    ): Promise<MetricsResponse> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        return (
            await orgUnitService.projectApi.getProjectMetrics(
                projectId,
                [metricTypes.join(',') as unknown as OrgUnitMetricType],
                start,
                end,
                samples
            )
        ).body;
    }

    async getDepartmentMetrics(
        departmentId: string,
        metricTypes: Array<OrgUnitMetricType>,
        start: Date,
        end: Date,
        samples: number,
        accessToken?: string
    ): Promise<MetricsResponse> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        // metricType - a workaround to overcome a bug in openapi-generator
        return (
            await orgUnitService.departmentApi.getDepartmentMetrics(
                departmentId,
                [metricTypes.join(',') as unknown as OrgUnitMetricType],
                start,
                end,
                samples
            )
        ).body;
    }

    ///////////////// Department API Wrappers with Full Response /////////////////

    /**
     * Creates a department and returns the full response including status code
     */
    async createDepartmentApi(
        departmentCreationRequest: DepartmentCreationRequest,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: Department }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.createDepartment(departmentCreationRequest);
    }

    /**
     * Gets a department by ID and returns the full response including status code
     */
    async getDepartmentApi(
        departmentId: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: Department }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.getDepartment(departmentId);
    }

    /**
     * Updates department resources and returns the full response including status code
     */
    async updateDepartmentResourcesApi(
        departmentId: string,
        resources: Array<Resources>,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: Array<Resources> }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.updateDepartmentResources(departmentId, resources);
    }

    /**
     * Gets department metrics and returns the full response including status code
     */
    async getDepartmentMetricsApi(
        departmentId: string,
        metricTypes: Array<OrgUnitMetricType>,
        start: Date,
        end: Date,
        samples?: number,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: MetricsResponse }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;

        // metricType - a workaround to overcome a bug in openapi-generator
        return await orgUnitService.departmentApi.getDepartmentMetrics(
            departmentId,
            [metricTypes.join(',') as unknown as OrgUnitMetricType],
            start,
            end,
            samples
        );
    }

    ///////////////// Utils Wrappers /////////////////
    async createProjectAndValidate(
        projectCreationRequest: ProjectCreationRequest,
        accessToken?: string
    ): Promise<Project> {
        if (!projectCreationRequest.parentId) {
            orgUnitService.departmentApi.accessToken =
                accessToken ?? flowRootState.users.system_administrator.accessToken;
            projectCreationRequest.parentId = (await this.getDepartments()).find(
                (dep) => dep.name === 'default'
            ).id;
        }

        const namespace = projectCreationRequest.requestedNamespace
            ? projectCreationRequest.requestedNamespace
            : Utils.namespaceFromProjectName(projectCreationRequest.name);

        logger.debug(
            `Creating project V2 ${
                projectCreationRequest.name
            } in namespace ${namespace}, request: ${JSON.stringify(projectCreationRequest)}`
        );

        let project = await this.createProject(projectCreationRequest, accessToken);

        project = await this.waitForProjectFromBackendToBeReady(project.name, project.id);

        logger.info(`Project ${project.name} and namespace ${namespace} created using org unit service api`);
        return project;
    }

    async waitForProjectFromBackendNonEmptyStatus(projectName: string, projectId: string): Promise<Project> {
        let project: Project;

        const isProjectsStatusNonEmpty = async (): Promise<boolean> => {
            try {
                project = await this.getProject(projectId);
                const phase = project.status.phase;
                return (
                    phase !== ProjectPhase.Creating &&
                    phase !== ProjectPhase.Unknown &&
                    phase !== ProjectPhase.NotReady &&
                    phase !== null &&
                    phase !== undefined
                );
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectName}/${projectId} from backend v2; reason: ${JSON.stringify(
                        e
                    )}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsStatusNonEmpty,
            `waitForProjectFromBackendNonEmptyStatus v2 project name: ${projectName}`,
            true,
            90
        );
        return project;
    }

    async waitForProjectFromBackendToBeReady(
        projectName: string,
        projectId: string,
        timeout = 40
    ): Promise<Project> {
        let project: Project;

        const isProjectsStatusReady = async (): Promise<boolean> => {
            try {
                project = await this.getProject(projectId);
                return project.status.phase === ProjectPhase.Ready;
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectName}/${projectId} from backend v2; reason: ${JSON.stringify(
                        e
                    )}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isProjectsStatusReady,
                `waitForProjectFromBackendToBeReady v2 project name: ${projectName}`,
                true,
                timeout
            );
        } catch (err) {
            logger.error(
                `waitForProjectFromBackendToBeReady timed out for ${projectName}/${projectId}, last seen phase=${project?.status?.phase}`
            );
            throw err;
        }
        return project;
    }

    async waitForProjectFromBackendWithCorrectNamespace(
        projectName: string,
        projectId: string,
        namespaceName: string
    ): Promise<void> {
        const isProjectsNamespaceCorrect = async (): Promise<boolean> => {
            try {
                const project = await this.getProject(projectId);
                return project.status.namespace === namespaceName;
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectName}/${projectId} from backend v2; reason: ${JSON.stringify(
                        e
                    )}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsNamespaceCorrect,
            `waitForProjectFromBackendWithCorrectNamespace v2 project name: ${projectName}`,
            true,
            30
        );
    }

    async waitForProjectFromBackendReadyAllocated(projectId: string, allocatedGpu: number): Promise<Project> {
        let project;
        const isProjectsReady = async (): Promise<boolean> => {
            try {
                project = await this.getProject(projectId);
                return (
                    project.status.phase === ProjectPhase.Ready &&
                    project.status.quotaStatus.allocated.gpu === allocatedGpu
                );
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectId} from backend v2; reason: ${JSON.stringify(e)}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsReady,
            `waitForProjectFromBackendReady v2 project: ${projectId}`,
            true,
            60
        );
        return project;
    }

    async waitForProjectFromBackendReadyAndResources(
        projectId: string,
        resourcesNum: number
    ): Promise<Project> {
        let project;
        const isProjectsReady = async (): Promise<boolean> => {
            try {
                project = await this.getProject(projectId);
                return (
                    project.status.phase === ProjectPhase.Ready && project.resources.length === resourcesNum
                );
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectId} from backend v2; reason: ${JSON.stringify(e)}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsReady,
            `waitForProjectFromBackendReadyAndResources v2 project: ${projectId}`,
            true,
            60
        );
        return project;
    }

    async waitForDepartmentFromBackendReadyAndResources(
        departmentId: string,
        resourcesNum: number
    ): Promise<Department> {
        let department;
        const isDepartmentReady = async (): Promise<boolean> => {
            try {
                department = await this.getDepartment(departmentId);
                return department.resources.length === resourcesNum;
            } catch (e) {
                logger.error(
                    `Failed getting department ${departmentId} from backend v2; reason: ${JSON.stringify(e)}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isDepartmentReady,
            `waitForDepartmentFromBackendReadyAndResources v2 department: ${departmentId}`,
            true,
            60
        );
        return department;
    }

    validateProjectResponse(
        response: { response?: http.IncomingMessage; body?: Project },
        projectName: string,
        resources: Resources = {
            gpu: { deserved: 0 },
            cpu: null,
            memory: null,
            rank: null
        },
        requestedNamespace = '',
        numberOfNodePools = 1,
        departmentName = 'default',
        interactiveJobTimeLimitSeconds = null,
        interactiveJobMaxIdleDurationSeconds = null,
        interactiveJobPreemptIdleDurationSeconds = null,
        trainingJobTimeLimitSeconds = null,
        trainingJobMaxIdleDurationSeconds = null
    ): void {
        const softAssert = new SoftAssert();
        if (response.response) {
            softAssert.assert(
                response.response.statusCode.valueOf() === StatusCodes.OK,
                `Expected status code to be ${StatusCodes.OK}`
            );
        }

        if (response.body) {
            softAssert.assert(
                response.body.name === projectName,
                `Expected project name to be ${projectName}, got instead ${response.body.name}`
            );
            softAssert.assert(response.body.id !== undefined, 'Expected project id to not be defined');
            softAssert.assert(
                response.body.createdAt !== undefined && response.body.createdAt !== null,
                `Expected createdAt to be present, got instead ${response.body.createdAt}`
            );
            softAssert.assert(
                response.body.parent?.name === departmentName,
                `Expected parent department name to be ${departmentName}, got instead ${response.body.parent?.name}`
            );
            softAssert.assert(
                typeof response.body.parentId === 'string',
                `Expected parentId to be a string, got instead ${typeof response.body.parentId}`
            );
            softAssert.assert(
                response.body.clusterId === flowRootState.cluster.uuid,
                `Expected clusterId to be ${flowRootState.cluster.uuid}, got instead ${response.body.clusterId}`
            );
            softAssert.assert(
                response.body.requestedNamespace === requestedNamespace,
                `Expected requestedNamespace to be ${requestedNamespace}, got instead ${response.body.requestedNamespace}`
            );
            softAssert.assertDeepNonStrictEqual<ResourcesGpu>(
                response.body.resources[0].gpu,
                resources.gpu,
                'Expected GPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<ResourcesCpu>(
                response.body.resources[0].cpu,
                resources.cpu,
                'Expected CPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<ResourcesMemory>(
                response.body.resources[0].memory,
                resources.memory,
                'Expected memory resources to match'
            );
            softAssert.assert(
                response.body.resources[0].rank === resources.rank,
                `Expected rank resources to be ${resources.rank}, got instead ${response.body.resources[0].rank}`
            );
            softAssert.assert(
                response.body.resources.length === numberOfNodePools,
                `Expected number of node pools to be ${numberOfNodePools}, got instead ${response.body.resources.length}`
            );
            softAssert.assert(
                response.body.resources[0].nodePool.name === 'default',
                `Expected first node pool name to be default, got instead ${response.body.resources[0].nodePool.name}`
            );
            softAssert.assert(
                response.body.resources[0].nodePool.id === resources.nodePool.id,
                `Expected first node pool id to be ${resources.nodePool.id}, got instead ${response.body.resources[0].nodePool.id}`
            );
            softAssert.assert(
                Array.isArray(response.body.effective.defaultNodePools) &&
                    response.body.effective.defaultNodePools.includes('default'),
                'Expected effective defaultNodePools to include default'
            );
            softAssert.assert(
                !response.body.defaultNodePools ||
                    (Array.isArray(response.body.defaultNodePools) &&
                        response.body.defaultNodePools.length === 0),
                'Expected defaultNodePools to be empty'
            );

            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.schedulingRules?.interactiveJobTimeLimitSeconds,
                interactiveJobTimeLimitSeconds,
                `Expected interactiveJobTimeLimitSeconds to be ${interactiveJobTimeLimitSeconds}, got instead ${response.body.schedulingRules?.interactiveJobTimeLimitSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.schedulingRules?.interactiveJobMaxIdleDurationSeconds,
                interactiveJobMaxIdleDurationSeconds,
                `Expected interactiveJobMaxIdleDurationSeconds to be ${interactiveJobMaxIdleDurationSeconds}, got instead ${response.body.schedulingRules?.interactiveJobMaxIdleDurationSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.schedulingRules?.interactiveJobPreemptIdleDurationSeconds,
                interactiveJobPreemptIdleDurationSeconds,
                `Expected interactiveJobPreemptIdleDurationSeconds to be ${interactiveJobPreemptIdleDurationSeconds}, got instead ${response.body.schedulingRules?.interactiveJobPreemptIdleDurationSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.schedulingRules?.trainingJobTimeLimitSeconds,
                trainingJobTimeLimitSeconds,
                `Expected trainingJobTimeLimitSeconds to be ${trainingJobTimeLimitSeconds}, got instead ${response.body.schedulingRules?.trainingJobTimeLimitSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.schedulingRules?.trainingJobMaxIdleDurationSeconds,
                trainingJobMaxIdleDurationSeconds,
                `Expected trainingJobMaxIdleDurationSeconds to be ${trainingJobMaxIdleDurationSeconds}, got instead ${response.body.schedulingRules?.trainingJobMaxIdleDurationSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.effective.schedulingRules?.interactiveJobTimeLimitSeconds,
                interactiveJobTimeLimitSeconds,
                `Expected effective interactiveJobTimeLimitSeconds to be ${interactiveJobTimeLimitSeconds}, got instead ${response.body.schedulingRules?.interactiveJobTimeLimitSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.effective.schedulingRules?.interactiveJobMaxIdleDurationSeconds,
                interactiveJobMaxIdleDurationSeconds,
                `Expected effective interactiveJobMaxIdleDurationSeconds to be ${interactiveJobMaxIdleDurationSeconds}, got instead ${response.body.schedulingRules?.interactiveJobMaxIdleDurationSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.effective.schedulingRules?.interactiveJobPreemptIdleDurationSeconds,
                interactiveJobPreemptIdleDurationSeconds,
                `Expected effective interactiveJobPreemptIdleDurationSeconds to be ${interactiveJobPreemptIdleDurationSeconds}, got instead ${response.body.schedulingRules?.interactiveJobPreemptIdleDurationSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.effective.schedulingRules?.trainingJobTimeLimitSeconds,
                trainingJobTimeLimitSeconds,
                `Expected effective trainingJobTimeLimitSeconds to be ${trainingJobTimeLimitSeconds}, got instead ${response.body.schedulingRules?.trainingJobTimeLimitSeconds}`
            );
            softAssert.assertNonStrictEqual(
                // eslint-disable-next-line eqeqeq
                response.body.effective.schedulingRules?.trainingJobMaxIdleDurationSeconds,
                trainingJobMaxIdleDurationSeconds,
                `Expected effective trainingJobMaxIdleDurationSeconds to be ${trainingJobMaxIdleDurationSeconds}, got instead ${response.body.schedulingRules?.trainingJobMaxIdleDurationSeconds}`
            );
        }

        softAssert.throwErrorIfFailed();
    }

    validateResources(
        numberOfNodePools: number,
        responseResources: Array<Resources>,
        expectedResources: Resources = {
            gpu: { deserved: 0 },
            cpu: null,
            memory: null,
            rank: null
        }
    ): void {
        const softAssert = new SoftAssert();

        softAssert.assertDeepNonStrictEqual<ResourcesGpu>(
            responseResources[0].gpu,
            expectedResources.gpu,
            'Expected GPU resources to match'
        );
        softAssert.assertDeepNonStrictEqual<ResourcesCpu>(
            responseResources[0].cpu,
            expectedResources.cpu,
            'Expected CPU resources to match'
        );
        softAssert.assertDeepNonStrictEqual<ResourcesMemory>(
            responseResources[0].memory,
            expectedResources.memory,
            'Expected memory resources to match'
        );
        softAssert.assertDeepNonStrictEqual<string>(
            responseResources[0].rank,
            expectedResources.rank,
            'Expected rank resources to match'
        );
        softAssert.assert(
            responseResources.length === numberOfNodePools,
            `Expected number of node pools to be ${numberOfNodePools}, got instead ${responseResources.length}`
        );
        softAssert.assert(
            responseResources[0].nodePool.name === 'default',
            `Expected first node pool name to be default, got instead ${responseResources[0].nodePool.name}`
        );
        softAssert.assert(
            responseResources[0].nodePool.id === expectedResources.nodePool.id,
            `Expected first node pool id to be ${expectedResources.nodePool.id}, got instead ${responseResources[0].nodePool.id}`
        );

        softAssert.throwErrorIfFailed();
    }

    validateDepartmentResponse(
        depResponse: Department,
        depName: string,
        resources: Resources = {
            nodePool: { name: 'default', id: '1' },
            gpu: { deserved: 0 },
            cpu: null,
            memory: null,
            rank: null
        },
        numberOfNodePools = 1,
        projectsChildren?: Array<DepartmentProjectChild>
    ): void {
        const softAssert = new SoftAssert();

        softAssert.assert(
            depResponse.name === depName,
            `Expected department name to be ${depName}, got instead ${depResponse.name}`
        );
        softAssert.assert(depResponse.id !== undefined, 'Expected department id to not be defined');
        softAssert.assert(
            depResponse.createdAt !== undefined && depResponse.createdAt !== null,
            `Expected department createdAt to be present, got instead ${depResponse.createdAt}`
        );

        softAssert.assert(
            depResponse.clusterId === flowRootState.cluster.uuid,
            `Expected department clusterId to be ${flowRootState.cluster.uuid}, got instead ${depResponse.clusterId}`
        );
        softAssert.assertDeepNonStrictEqual<ResourcesGpu>(
            depResponse.resources[0].gpu,
            resources.gpu,
            'Expected department GPU resources to match'
        );
        softAssert.assertDeepNonStrictEqual<ResourcesCpu>(
            depResponse.resources[0].cpu,
            resources.cpu,
            'Expected department CPU resources to match'
        );
        softAssert.assertDeepNonStrictEqual<ResourcesMemory>(
            depResponse.resources[0].memory,
            resources.memory,
            'Expected department memory resources to match'
        );
        softAssert.assert(
            depResponse.resources.length === numberOfNodePools,
            `Expected department number of node pools to be ${numberOfNodePools}, got instead ${depResponse.resources.length}`
        );
        softAssert.assert(
            depResponse.resources[0].rank === resources.rank,
            `Expected department rank of node pools to be ${resources.rank}, got instead ${depResponse.resources[0].rank}`
        );

        const resDefaultNpResources = depResponse.resources.find((res) => res.nodePool?.name === 'default');
        softAssert.assert(
            resDefaultNpResources !== undefined,
            `Expected department resources to have a node pool with name default, but resources are: ${JSON.stringify(depResponse.resources)}`
        );
        softAssert.assert(
            resDefaultNpResources.nodePool?.id === resources.nodePool?.id,
            `Expected department first node pool id to be ${resources.nodePool?.id}, got instead ${resDefaultNpResources.nodePool?.id}`
        );

        if (projectsChildren) {
            softAssert.assert(
                depResponse.children?.length === projectsChildren.length,
                `Expected department number of project children to be ${projectsChildren.length}, got instead ${depResponse.children?.length}`
            );
            // projectsChildren and depResponse.children should have same elements, but not necessarily in the same order
            for (const projectChild of projectsChildren) {
                softAssert.assert(
                    depResponse.children?.find(
                        (child) => child.id === projectChild.id && child.name === projectChild.name
                    ) !== undefined,
                    `Expected department children to include project child with id ${projectChild.id} and name ${projectChild.name}, but got instead: ${JSON.stringify(depResponse.children)}`
                );
            }
        }
        softAssert.throwErrorIfFailed();
    }

    async validateProjectQueueCrd(
        queueName: string,
        project: Project,
        timeout = 30,
        desiredResult = true,
        resourcesIndex = 0
    ): Promise<void> {
        await projectUtils.validateQueueCrdOfProject(
            queueName,
            project.name,
            project.resources[resourcesIndex].gpu.deserved,
            project.resources[resourcesIndex].gpu.limit ?? -1,
            project.resources[resourcesIndex].gpu.overQuotaWeight ??
                (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)
                    ? 1
                    : 0),
            project.resources[resourcesIndex].cpu?.deserved,
            project.resources[resourcesIndex].cpu?.limit,
            project.resources[resourcesIndex].cpu?.overQuotaWeight,
            project.resources[resourcesIndex].memory?.deserved,
            project.resources[resourcesIndex].memory?.limit,
            project.resources[resourcesIndex].memory?.overQuotaWeight,
            project.resources[resourcesIndex].rank,
            timeout,
            desiredResult
        );
    }

    async validateDepartmentQueueCrd(
        queueName: string,
        department: Department,
        gpuQuota = undefined,
        timeout = 60,
        desiredResult = true
    ): Promise<void> {
        await departmentUtils.validateQueueCrdOfDepartment(
            queueName,
            department.id,
            department.name,
            department.resources[0].gpu.deserved,
            department.resources[0].gpu.limit,
            department.resources[0].gpu.overQuotaWeight,
            department.resources[0].cpu?.deserved,
            department.resources[0].cpu?.limit,
            department.resources[0].cpu?.overQuotaWeight,
            department.resources[0].memory?.deserved,
            department.resources[0].memory?.limit,
            department.resources[0].memory?.overQuotaWeight,
            gpuQuota,
            projectUtils.mapPriorityToNumber(department.resources[0].rank),
            timeout,
            desiredResult
        );
    }

    async waitForQueueToHaveGpuQuota(
        queueName: string,
        gpu: number,
        desiredResult = true,
        timeout = 30
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const queueCrd = await waitUtils.waitForCrd(crds.queueCrd, queueName, '');
                return (
                    queueCrd['spec']['resources']['gpu']['quota'] &&
                    Number(queueCrd['spec']['resources']['gpu']['quota']) === gpu
                );
            },
            `Waiting for queue crd ${queueName} to be updated with gpu=${gpu}`,
            desiredResult,
            timeout
        );
    }

    async waitForDepartmentInOrgUnitService(
        departmentId: string,
        timeout = 30,
        desiredResult = true
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const departments = await this.getDepartments();
                return departments.some((dep) => dep.id === departmentId);
            },
            `Waiting for department with id ${departmentId} to be present in org unit service`,
            desiredResult,
            timeout
        );
    }

    async waitForDepartmentByNameInOrgUnitService(
        departmentName: string,
        timeout = 30,
        desiredResult = true
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const departments = await this.getDepartments();
                return departments.some((dep) => dep.name === departmentName);
            },
            `Waiting for department with name ${departmentName} to be present in org unit service`,
            desiredResult,
            timeout
        );
    }

    copyProject(project: Project): Project {
        const projCopy = Object.assign({}, JSON.parse(JSON.stringify(project)) as Project) as Project;
        projCopy.createdAt = project.createdAt;
        return projCopy;
    }

    async getProjectsRanking(
        metricType: RankingMetricType = RankingMetricType.AvgGpuAllocation24h,
        order: RankingOrderType = RankingOrderType.Top,
        nodepoolName?: string
    ): Promise<{ response: http.IncomingMessage; body: GetProjectsRanking200Response }> {
        orgUnitService.projectApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const clusterId = flowRootState.cluster.uuid;

        return await orgUnitService.projectApi.getProjectsRanking(metricType, order, clusterId, nodepoolName);
    }

    async getDepartmentNodePoolResources(
        filterBy?: Array<string>,
        sortBy?: DepartmentNodePoolResourcesFilterSortFields,
        sortOrder?: 'asc' | 'desc',
        offset?: number,
        limit?: number,
        search?: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: GetDepartmentNodePoolResources200Response }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.getDepartmentNodePoolResources(
            filterBy,
            sortBy,
            sortOrder,
            offset,
            limit,
            search
        );
    }

    async sumDepartmentNodePoolResources(
        nodepoolId: string,
        filterBy?: Array<string>,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: DepartmentSumResponse }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.sumDepartmentNodePoolResources(nodepoolId, filterBy);
    }

    async batchPatchDepartmentNodePoolResources(
        batchPatchNodePoolResourcesRequest: BatchPatchNodePoolResourcesRequest,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: BatchPatchDepartmentResourcesResponse }> {
        orgUnitService.departmentApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.departmentApi.batchPatchDepartmentNodePoolResources(
            batchPatchNodePoolResourcesRequest
        );
    }

    async getProjectNodePoolResources(
        filterBy?: Array<string>,
        sortBy?: ProjectNodePoolResourcesFilterSortFields,
        sortOrder?: 'asc' | 'desc',
        offset?: number,
        limit?: number,
        search?: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: GetProjectNodePoolResources200Response }> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.projectApi.getProjectNodePoolResources(
            filterBy,
            sortBy,
            sortOrder,
            offset,
            limit,
            search
        );
    }

    async sumProjectNodePoolResources(
        nodepoolId: string,
        filterBy?: Array<string>,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: SumResponse }> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.projectApi.sumProjectNodePoolResources(nodepoolId, filterBy);
    }

    async batchPatchProjectNodePoolResources(
        batchPatchNodePoolResourcesRequest: BatchPatchNodePoolResourcesRequest,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: BatchPatchProjectResourcesResponse }> {
        orgUnitService.projectApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await orgUnitService.projectApi.batchPatchProjectNodePoolResources(
            batchPatchNodePoolResourcesRequest
        );
    }
}
