/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    allureUtils,
    authorizationUtils,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    Utils,
    waitUtils
} from '..';
import { PROPERTY_TO_ROLE_MAPPING } from '../config/roleMapping';
import { expectedOQWhenAwOff } from './OrgUnitUtils';
import { backendClusterApi, orgUnitService, projectsDepartmentsApi } from '../../api';
import {
    AssignedResources,
    JobsNodeAffinityInteractive,
    JobsNodeAffinityTrain,
    NodePoolAssignedResources,
    NodePoolAssignedResourcesCreate,
    Project,
    ProjectCreateRequest
} from '../../api/apis/projectsDepartmentsApi';
import { Project as ProjectOU, ProjectCreationRequest } from '../../api/apis/orgUnitApi';
import { NodePool } from '../../api/apis/backendClusterApi';
import { flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import http from 'http';
import { ScopeType } from '../../api/apis/authorizationApi/generated/model/scopeType';
import { SubjectType } from '../../api/apis/authorizationApi/generated/model/subjectType';
import { kubeCli } from '../../cli';
import { attachment } from 'allure-js-commons';

export class ProjectResourcesUtils {
    getExpectedQueueResourcesOverQuotaOff(
        cpuQuotaFeatureOn: boolean,
        overQuotaWeight: number,
        deservedGpu: number,
        deservedCpu: number,
        deservedMemory: number
    ): any {
        if (cpuQuotaFeatureOn) {
            return {
                gpu: {
                    quota: deservedGpu,
                    limit: overQuotaWeight === 1 ? -1 : deservedGpu,
                    overQuotaWeight: expectedOQWhenAwOff(deservedGpu)
                },
                cpu: {
                    quota: deservedCpu,
                    limit: overQuotaWeight === 1 ? -1 : deservedCpu,
                    overQuotaWeight: expectedOQWhenAwOff(deservedCpu)
                },
                memory: {
                    quota: deservedMemory,
                    limit: overQuotaWeight === 1 ? -1 : deservedMemory,
                    overQuotaWeight: expectedOQWhenAwOff(deservedMemory)
                }
            };
        } else {
            return {
                gpu: {
                    quota: deservedGpu,
                    limit: overQuotaWeight === 1 ? -1 : deservedGpu,
                    overQuotaWeight: expectedOQWhenAwOff(deservedGpu)
                },
                cpu: {
                    quota: -1,
                    limit: -1,
                    overQuotaWeight: 1
                },
                memory: {
                    quota: -1,
                    limit: -1,
                    overQuotaWeight: 1
                }
            };
        }
    }

    createResourcesWithOverQuotaFlagOn(
        cpuQuotaFeatureOn: boolean,
        overQuotaWeight: number,
        deservedGpu: number,
        deservedCpu: number,
        deservedMemory: number
    ): AssignedResources {
        if (cpuQuotaFeatureOn) {
            return {
                gpu: {
                    deserved: deservedGpu,
                    maxAllowed: deservedGpu,
                    overQuotaWeight: overQuotaWeight
                },
                cpu: {
                    deserved: deservedCpu,
                    maxAllowed: deservedCpu,
                    overQuotaWeight: overQuotaWeight
                },
                memory: {
                    deserved: deservedMemory,
                    maxAllowed: deservedMemory,
                    overQuotaWeight: overQuotaWeight
                }
            };
        } else {
            return {
                gpu: {
                    deserved: deservedGpu,
                    maxAllowed: deservedGpu,
                    overQuotaWeight: overQuotaWeight
                },
                cpu: {
                    deserved: null,
                    maxAllowed: null,
                    overQuotaWeight: null
                },
                memory: {
                    deserved: null,
                    maxAllowed: null,
                    overQuotaWeight: null
                }
            };
        }
    }

    getExpectedQueueResourcesWithOverQuotaOn(
        cpuQuotaFeatureOn: boolean,
        overQuotaWeight: number,
        deservedGpu: number,
        deservedCpu: number,
        deservedMemory: number
    ): any {
        if (cpuQuotaFeatureOn) {
            return {
                gpu: {
                    quota: deservedGpu,
                    limit: deservedGpu,
                    overQuotaWeight: overQuotaWeight
                },
                cpu: {
                    quota: deservedCpu,
                    limit: deservedCpu,
                    overQuotaWeight: overQuotaWeight
                },
                memory: {
                    quota: deservedMemory,
                    limit: deservedMemory,
                    overQuotaWeight: overQuotaWeight
                }
            };
        } else {
            return {
                gpu: {
                    quota: deservedGpu,
                    limit: deservedGpu,
                    overQuotaWeight: overQuotaWeight
                },
                cpu: {
                    quota: -1,
                    limit: -1,
                    overQuotaWeight: 1
                },
                memory: {
                    quota: -1,
                    limit: -1,
                    overQuotaWeight: 1
                }
            };
        }
    }

    createResourcesGPUOnly(
        deservedGpu: number,
        overQuotaAllowed = true,
        overQuotaWeight = 2
    ): NodePoolAssignedResources {
        const projectResources = new NodePoolAssignedResources();
        projectResources.gpu = {
            deserved: deservedGpu,
            maxAllowed: overQuotaAllowed ? -1 : deservedGpu,
            overQuotaWeight: undefined
        };
        return projectResources;
    }

    createResourcesWithOverQuotaAllowed(deservedGpu: number): NodePoolAssignedResources {
        return this.createResourcesGPUOnly(deservedGpu);
    }

    // DEPRECATED use orgUnitUtils instead
    async createProject(
        namespace: string,
        projectCreateRequest: Partial<ProjectCreateRequest>,
        accessToken?: string,
        runaiCreatedNamespace = true,
        excludePermissions = false,
        additionalResourcesForNonDefaultNodePools?: NodePoolAssignedResourcesCreate[],
        clusterUuid = flowRootState.cluster.uuid,
        namespaceWaitTimeoutSec = 90
    ): Promise<Project> {
        projectsDepartmentsApi.projectsApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.editor.accessToken;
        projectsDepartmentsApi.departmentsApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.editor.accessToken;
        orgUnitService.departmentApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.editor.accessToken;
        orgUnitService.projectApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.editor.accessToken;
        backendClusterApi.nodePoolsApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.system_administrator.accessToken;

        if (!projectCreateRequest.departmentId) {
            projectCreateRequest.departmentId = (
                await projectsDepartmentsApi.departmentsApi.getDepartments(clusterUuid)
            ).body.find((defaultDepartment) => defaultDepartment.name === 'default').id;
        }

        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
        if (!projectCreateRequest.nodePoolsResources || !projectCreateRequest.nodePoolsResources.length) {
            if (projectCreateRequest.resources) {
                projectCreateRequest.nodePoolsResources = [
                    {
                        cpu: projectCreateRequest.resources?.cpu,
                        gpu: projectCreateRequest.resources?.gpu,
                        memory: projectCreateRequest.resources?.memory,
                        nodePool: { id: defaultNodePool.id, name: defaultNodePool.name }
                    }
                ];
            } else if (projectCreateRequest.deservedGpus && projectCreateRequest.maxAllowedGpus) {
                projectCreateRequest.nodePoolsResources = [
                    {
                        gpu: {
                            deserved: projectCreateRequest.deservedGpus,
                            maxAllowed: projectCreateRequest.maxAllowedGpus
                        },
                        memory: {
                            deserved: null,
                            maxAllowed: null
                        },
                        cpu: {
                            deserved: null,
                            maxAllowed: null
                        },
                        nodePool: { id: defaultNodePool.id, name: defaultNodePool.name }
                    }
                ];
            }
            if (additionalResourcesForNonDefaultNodePools) {
                projectCreateRequest.nodePoolsResources = projectCreateRequest.nodePoolsResources.concat(
                    additionalResourcesForNonDefaultNodePools
                );
            }
        }

        logger.debug(
            `Creating project ${
                projectCreateRequest.name
            } in namespace ${namespace}, request: ${JSON.stringify(projectCreateRequest)}`
        );

        const projectActualCreateRequest = projectCreateRequest as ProjectCreateRequest;
        const project = (
            await orgUnitService.projectApi.createProject(
                this.projectCreateRequestV1ToV2(projectActualCreateRequest, defaultNodePool)
            )
        ).body;
        await orgUnitUtils.waitForProjectFromBackendToBeReady(project.name, String(project.id));

        let relatedNamespaceName = namespace;
        if (runaiCreatedNamespace) {
            relatedNamespaceName = Utils.namespaceFromProjectName(project.name);
        }
        const createdProject = this.projectV2ToV1(project);
        createdProject.namespace = relatedNamespaceName;
        await this.waitForProjectNamespace(createdProject, namespaceWaitTimeoutSec);
        logger.info(`Project ${project.name} and namespace ${relatedNamespaceName} created`);
        const permissions = projectCreateRequest.permissions;
        if (permissions?.users?.length > 0) {
            const roleId = flowRootState.users.l1_researcher.role.id;
            for (const usr of permissions.users) {
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    usr,
                    roleId,
                    SubjectType.User,
                    project.id,
                    ScopeType.Project
                );
            }
        }
        if (permissions?.applications?.length > 0) {
            for (const app of permissions.applications) {
                await authorizationUtils.createAccessRuleForApp(
                    app,
                    PROPERTY_TO_ROLE_MAPPING.l1_researcher,
                    ScopeType.Project,
                    project.id
                );
            }
        }
        if (permissions?.groups?.length > 0) {
            const roleId = flowRootState.users.l1_researcher.role.id;
            for (const group of permissions.groups) {
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    group,
                    roleId,
                    ScopeType.Project,
                    project.id
                );
            }
        }
        return createdProject;
    }

    projectV2ToV1(projectOu: ProjectOU, tenantId: number = null): Project {
        if (!projectOu) {
            return null;
        }
        const nodePoolResources = [];
        projectOu.resources.forEach((res) => {
            const v1Resources = {
                id: projectOu.id,
                nodePool: {
                    id: res.nodePool.id,
                    name: res.nodePool.name
                },
                gpu: {
                    deserved: res.gpu.deserved,
                    maxAllowed: res.gpu.limit,
                    overQuotaWeight: res.gpu.overQuotaWeight
                }
            };
            if (res.cpu) {
                v1Resources['cpu'] = {
                    deserved: res.cpu.deserved,
                    maxAllowed: res.cpu.limit,
                    overQuotaWeight: res.cpu.overQuotaWeight
                };
            }
            if (res.memory) {
                v1Resources['memory'] = {
                    deserved: res.memory.deserved,
                    maxAllowed: res.memory.limit,
                    overQuotaWeight: res.memory.overQuotaWeight
                };
            }
            nodePoolResources.push(v1Resources);
        });
        let nodeAffinity = null;
        if (projectOu.nodeTypes?.workspace?.length > 0) {
            nodeAffinity = {
                interactive: {
                    affinityType: 'selectedTypes',
                    selectedTypes: projectOu.nodeTypes.workspace
                }
            };
        }
        if (projectOu.nodeTypes?.training?.length > 0) {
            if (nodeAffinity) {
                nodeAffinity.train = {
                    affinityType: 'selectedTypes',
                    selectedTypes: projectOu.nodeTypes.workspace
                };
            } else {
                nodeAffinity = {
                    interactive: {
                        affinityType: 'selectedTypes',
                        selectedTypes: projectOu.nodeTypes.workspace
                    }
                };
            }
        }
        return {
            id: Number(projectOu.id),
            name: projectOu.name,
            namespace:
                projectOu.status.namespace ||
                projectOu.requestedNamespace ||
                Utils.namespaceFromProjectName(projectOu.name),
            nodePoolsResources: nodePoolResources,
            departmentId: Number(projectOu.parentId),
            tenantId: tenantId,
            clusterUuid: projectOu.clusterId,
            departmentName: projectOu.parent?.name,
            nodeAffinity: nodeAffinity,
            createdAt: projectOu.createdAt,
            status: projectOu.status,
            phase: projectOu.status?.phase?.toString(),
            defaultNodePools: projectOu.defaultNodePools,
            interactiveJobTimeLimitSecs: projectOu.schedulingRules?.interactiveJobTimeLimitSeconds,
            interactiveJobMaxIdleDurationSecs:
                projectOu.schedulingRules?.interactiveJobMaxIdleDurationSeconds,
            interactivePreemptibleJobMaxIdleDurationSecs:
                projectOu.schedulingRules?.interactiveJobPreemptIdleDurationSeconds,
            trainingJobTimeLimitSecs: projectOu.schedulingRules?.trainingJobTimeLimitSeconds,
            trainingJobMaxIdleDurationSecs: projectOu.schedulingRules?.trainingJobMaxIdleDurationSeconds
        };
    }

    projectCreateRequestV1ToV2(
        v1Request: ProjectCreateRequest,
        defaultNodePool: NodePool
    ): ProjectCreationRequest {
        const resources = orgUnitUtils.getResourcesFromV1(v1Request, defaultNodePool);
        let nodeTypes = null;
        if (v1Request.nodeAffinity) {
            nodeTypes = {};
            if (
                v1Request.nodeAffinity.train &&
                v1Request.nodeAffinity.train.affinityType ===
                    JobsNodeAffinityTrain.AffinityTypeEnum.OnlySelected &&
                v1Request.nodeAffinity.train.selectedTypes &&
                v1Request.nodeAffinity.train.selectedTypes.length > 0
            ) {
                nodeTypes.training = v1Request.nodeAffinity.train.selectedTypes.map((nt) => String(nt.id));
            }
            if (
                v1Request.nodeAffinity.interactive &&
                v1Request.nodeAffinity.interactive.affinityType ===
                    JobsNodeAffinityInteractive.AffinityTypeEnum.OnlySelected &&
                v1Request.nodeAffinity.interactive.selectedTypes &&
                v1Request.nodeAffinity.interactive.selectedTypes.length > 0
            ) {
                nodeTypes.workspace = v1Request.nodeAffinity.interactive.selectedTypes.map((nt) =>
                    String(nt.id)
                );
            }
        }

        const v2Request = {
            name: v1Request.name,
            parentId: v1Request.departmentId.toString(),
            clusterId: flowRootState.cluster.uuid,
            requestedNamespace: v1Request.namespace,
            resources: resources,
            defaultNodePools: v1Request.defaultNodePools
        } as ProjectCreationRequest;
        if (nodeTypes) {
            v2Request.nodeTypes = nodeTypes;
        }
        const schedulingRulesMap = [];
        if (v1Request.interactiveJobTimeLimitSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobTimeLimitSeconds',
                value: v1Request.interactiveJobTimeLimitSecs
            });
        }
        if (v1Request.interactiveJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobMaxIdleDurationSeconds',
                value: v1Request.interactiveJobMaxIdleDurationSecs
            });
        }
        if (v1Request.interactivePreemptibleJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobPreemptIdleDurationSeconds',
                value: v1Request.interactivePreemptibleJobMaxIdleDurationSecs
            });
        }
        if (v1Request.trainingJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'trainingJobMaxIdleDurationSeconds',
                value: v1Request.trainingJobMaxIdleDurationSecs
            });
        }
        if (v1Request.trainingJobTimeLimitSecs) {
            schedulingRulesMap.push({
                key: 'trainingJobTimeLimitSeconds',
                value: v1Request.trainingJobTimeLimitSecs
            });
        }
        if (schedulingRulesMap.length > 0) {
            v2Request.schedulingRules = {};
            schedulingRulesMap.forEach((rule) => {
                v2Request.schedulingRules[rule.key] = rule.value;
            });
        }
        return v2Request;
    }

    projectV1ToV2(v1Project: Project): ProjectOU {
        const resources = orgUnitUtils.getResourcesFromV1(v1Project);
        let nodeTypes = null;
        if (v1Project.nodeAffinity) {
            nodeTypes = {};
            if (
                v1Project.nodeAffinity.train &&
                v1Project.nodeAffinity.train.affinityType ===
                    JobsNodeAffinityTrain.AffinityTypeEnum.OnlySelected &&
                v1Project.nodeAffinity.train.selectedTypes &&
                v1Project.nodeAffinity.train.selectedTypes.length > 0
            ) {
                nodeTypes.training = v1Project.nodeAffinity.train.selectedTypes.map((nt) => String(nt.id));
            }
            if (
                v1Project.nodeAffinity.interactive &&
                v1Project.nodeAffinity.interactive.affinityType ===
                    JobsNodeAffinityInteractive.AffinityTypeEnum.OnlySelected &&
                v1Project.nodeAffinity.interactive.selectedTypes &&
                v1Project.nodeAffinity.interactive.selectedTypes.length > 0
            ) {
                nodeTypes.workspace = v1Project.nodeAffinity.interactive.selectedTypes.map((nt) =>
                    String(nt.id)
                );
            }
        }

        const v2Request = {
            id: v1Project.id.toString(),
            name: v1Project.name,
            parentId: v1Project.departmentId.toString(),
            clusterId: flowRootState.cluster.uuid,
            requestedNamespace: v1Project.namespace,
            schedulingRules: {
                interactiveJobTimeLimitSeconds: v1Project.interactiveJobTimeLimitSecs,
                interactiveJobMaxIdleDurationSeconds: v1Project.interactiveJobMaxIdleDurationSecs,
                interactiveJobPreemptIdleDurationSeconds:
                    v1Project.interactivePreemptibleJobMaxIdleDurationSecs,
                trainingJobMaxIdleDurationSeconds: v1Project.trainingJobMaxIdleDurationSecs,
                trainingJobTimeLimitSeconds: v1Project.trainingJobTimeLimitSecs
            },
            resources: resources,
            defaultNodePools: v1Project.defaultNodePools
        } as ProjectOU;
        if (nodeTypes) {
            v2Request.nodeTypes = nodeTypes;
        }
        const schedulingRulesMap = [];
        if (v1Project.interactiveJobTimeLimitSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobTimeLimitSeconds',
                value: v1Project.interactiveJobTimeLimitSecs
            });
        }
        if (v1Project.interactiveJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobMaxIdleDurationSeconds',
                value: v1Project.interactiveJobMaxIdleDurationSecs
            });
        }
        if (v1Project.interactivePreemptibleJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'interactiveJobPreemptIdleDurationSeconds',
                value: v1Project.interactivePreemptibleJobMaxIdleDurationSecs
            });
        }
        if (v1Project.trainingJobMaxIdleDurationSecs) {
            schedulingRulesMap.push({
                key: 'trainingJobMaxIdleDurationSeconds',
                value: v1Project.trainingJobMaxIdleDurationSecs
            });
        }
        if (v1Project.trainingJobTimeLimitSecs) {
            schedulingRulesMap.push({
                key: 'trainingJobTimeLimitSeconds',
                value: v1Project.trainingJobTimeLimitSecs
            });
        }
        if (schedulingRulesMap.length > 0) {
            v2Request.schedulingRules = {};
            schedulingRulesMap.forEach((rule) => {
                v2Request.schedulingRules[rule.key] = rule.value;
            });
        }
        return v2Request;
    }

    // DEPRECATED use orgUnitUtils instead
    async updateProject(project: Project): Promise<{ response: http.IncomingMessage; body: Project }> {
        // async updateProject(project: ProjectOU): Promise<{ response: http.IncomingMessage; body: ProjectOU }> {
        const projectV2 = this.projectV1ToV2(project);
        const result = await orgUnitService.projectApi.updateProject(projectV2.id, projectV2);
        return { response: result.response, body: this.projectV2ToV1(result.body) };
    }

    copyProject(project: Project): Project {
        const projCopy = Object.assign({}, JSON.parse(JSON.stringify(project)) as Project) as Project;
        projCopy.createdAt = project.createdAt;
        return projCopy;
    }

    /**
     * Creates a project and returns only after its Kubernetes Namespace exists.
     */
    async createProjectAndNamespace(
        name: string,
        departmentId?: number,
        deservedGpus = 1,
        maxAllowedGpus = 1,
        clusterUuid: string = flowRootState.cluster.uuid,
        randomSuffix = true,
        namespaceWaitTimeoutSec = 90
    ): Promise<Project> {
        const projectName = randomSuffix ? objectGenerator.generateShortNameByPrefix(name) : name;
        const projectCreationPayload = {
            name: projectName,
            deservedGpus: deservedGpus,
            maxAllowedGpus: maxAllowedGpus
        };
        if (departmentId) {
            projectCreationPayload['departmentId'] = departmentId;
        }
        const namespace = Utils.namespaceFromProjectName(projectName);
        const project = await this.createProject(
            namespace,
            projectCreationPayload,
            undefined,
            true,
            false,
            undefined,
            clusterUuid,
            namespaceWaitTimeoutSec
        );
        project.namespace = namespace;
        return project;
    }

    private async waitForProjectNamespace(project: Project, namespaceWaitTimeoutSec: number): Promise<void> {
        const startedAt = Date.now();
        logger.info(
            `[createProject] backend project ready: id=${project.id} name=${project.name}; ` +
                `waiting for namespace=${project.namespace} timeoutSec=${namespaceWaitTimeoutSec}`
        );
        try {
            await waitUtils.waitForNamespaceToExist(project.namespace, namespaceWaitTimeoutSec);
            logger.info(
                `[createProject] namespace materialized: name=${project.namespace} ` +
                    `elapsedMs=${Date.now() - startedAt}`
            );
        } catch (err) {
            logger.error(
                `[createProject] namespace did NOT materialize within ${namespaceWaitTimeoutSec}s: ` +
                    `project.id=${project.id} project.name=${project.name} ` +
                    `expectedNamespace=${project.namespace} ` +
                    `elapsedMs=${Date.now() - startedAt}`
            );
            await this.attachNamespaceWaitDiagnostics(project);
            throw err;
        }
    }

    private async attachNamespaceWaitDiagnostics(project: Project): Promise<void> {
        try {
            const fresh = (await orgUnitService.projectApi.getProject(String(project.id))).body;
            const payload = JSON.stringify(
                { id: fresh.id, name: fresh.name, status: fresh.status, clusterId: fresh.clusterId },
                null,
                2
            );
            await attachment('namespace-wait: backend project state', payload, 'application/json');
            logger.error(`[diag] backend project status: ${JSON.stringify(fresh.status ?? {})}`);
        } catch (e) {
            logger.error(`[diag] failed to re-fetch project: ${e}`);
        }

        try {
            const prefix = project.namespace.split('-').slice(0, 3).join('-');
            const list = await kubeCli.CoreV1Api.listNamespace({});
            const matching = (list.items ?? [])
                .filter((ns) => ns.metadata?.name?.startsWith(prefix))
                .map((ns) => ({
                    name: ns.metadata?.name,
                    phase: ns.status?.phase,
                    created: ns.metadata?.creationTimestamp
                }));
            await attachment(
                'namespace-wait: cluster namespaces matching prefix',
                JSON.stringify(matching, null, 2),
                'application/json'
            );
            logger.error(`[diag] cluster namespaces matching '${prefix}*': ${JSON.stringify(matching)}`);
        } catch (e) {
            logger.error(`[diag] failed to list namespaces: ${e}`);
        }

        try {
            const labeled = await kubeCli.CoreV1Api.listNamespace({
                labelSelector: `runai/queue=${project.name}`
            });
            const names = (labeled.items ?? []).map((ns) => ns.metadata?.name);
            await attachment(
                'namespace-wait: namespaces labeled runai/queue',
                JSON.stringify({ projectName: project.name, namespaces: names }, null, 2),
                'application/json'
            );
            logger.error(
                `[diag] namespaces with label runai/queue=${project.name}: ${JSON.stringify(names)}`
            );
        } catch (e) {
            logger.error(`[diag] failed to list labeled namespaces: ${e}`);
        }

        try {
            // Substring match, so this catches both the pre-3.0 `runai-project-controller`
            // pod and the KRM-packaged `project-controller` that replaces it.
            await allureUtils.attachSystemComponentLogs(
                'project-controller',
                'namespace-wait: project-controller logs',
                'runai',
                500
            );
        } catch (e) {
            logger.error(`[diag] failed to attach controller logs: ${e}`);
        }
    }

    updateResourceConfig(
        resource: any,
        targetNodePoolId: string,
        gpuConfig: { deserved: number; maxAllowed: number; overQuotaWeight: number }
    ): any {
        const isTargetNodePool = resource.nodePool.id === targetNodePoolId;

        if (isTargetNodePool) {
            const updatedResource: any = {
                nodePool: resource.nodePool,
                gpu: {
                    deserved: gpuConfig.deserved,
                    limit: gpuConfig.maxAllowed,
                    overQuotaWeight: gpuConfig.overQuotaWeight
                }
            };
            if (resource.rank) {
                updatedResource.rank = resource.rank;
            }
            return updatedResource;
        }

        const cleanedResource: any = {
            nodePool: resource.nodePool,
            gpu: resource.gpu
        };
        if (resource.rank) {
            cleanedResource.rank = resource.rank;
        }
        return cleanedResource;
    }

    async updateProjectGpuConfig(
        projectId: string,
        targetNodePoolId: string,
        gpuConfig: {
            deserved: number;
            maxAllowed: number;
            overQuotaWeight: number;
        }
    ): Promise<void> {
        const currentProject = await orgUnitUtils.getProject(projectId);
        const updatedResources = currentProject.resources.map((resource) =>
            this.updateResourceConfig(resource, targetNodePoolId, gpuConfig)
        );
        await orgUnitUtils.patchProjectResources(projectId, updatedResources);
        logger.info(
            `Updated project GPU config for node pool '${targetNodePoolId}': deserved=${gpuConfig.deserved}, maxAllowed=${gpuConfig.maxAllowed}, overQuotaWeight=${gpuConfig.overQuotaWeight}`
        );
    }
}
