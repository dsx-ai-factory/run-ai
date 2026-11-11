/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, crds, flowRootState, logger, Utils, versionUtils, waitUtils } from '../index';
import {
    AssignedResourcesCpu,
    AssignedResourcesGpu,
    AssignedResourcesMemory,
    NodePoolAssignedResourcesCreateCpu,
    NodePoolAssignedResourcesCreateGpu,
    NodePoolAssignedResourcesCreateMemory,
    Project
} from '../../api/apis/projectsDepartmentsApi';

import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { projectsDepartmentsApi } from '../../api';
import { kubeCli } from '../../cli';
import http from 'http';
import { validatePropertyIfNotNull } from './testing/helpersUtils';
import {
    getOrgUnitUpdatedAt,
    getProjectJobLimitSecs,
    getProjectParent,
    ProjectJobLimitField,
    projectJobLimitFields,
    resolveOrgUnitCrds
} from './OrgUnitCrdFields';
import { SoftAssert } from '../SoftAssert';
import { SupportedVersions } from './SupporetedEnvsUtils';

export enum EProjectPhase {
    Creating = 'Creating',
    Updating = 'Updating',
    Deleting = 'Deleting',
    Ready = 'Ready',
    NotReady = 'NotReady',
    Initializing = 'Initializing',
    Unknown = 'Unknown'
}

export class ProjectUtils {
    newRbacMinVersion = '2.14';

    newRbacRunaiResearcherRoleBindingName = 'runai-l1-researcher-ns';

    oldRunaiResearcherRoleBindingName = 'runai-researcher';

    newRbacRunaiResearchManagerRoleBindingName = 'runai-research-manager-ns';

    oldRbacRunaiResearchManagerRoleBindingName = 'runai-research-manager';

    newRbacRunaiSystemAdministratorRoleBindingName = 'runai-system-administrator-ns';

    mapPriorityToNumber(priorityName: string): number {
        const priorityMap = {
            Highest: 255,
            VeryHigh: 240,
            High: 210,
            MediumHigh: 180,
            Medium: 150,
            MediumLow: 100,
            Low: 50,
            VeryLow: 20,
            Lowest: 1
        };
        const priorityValue = priorityMap[priorityName];
        if (!priorityValue) {
            return 100;
        }
        return priorityValue;
    }

    async waitForProjectCrdNonEmptyStatus(projectName: string, desiredPhase = ''): Promise<object> {
        await resolveOrgUnitCrds();
        let project: object;
        let phase: string;

        const isProjectsStatusNonEmpty = async (): Promise<boolean> => {
            try {
                project = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, undefined, 60);
                const namespace = project['status'].namespace;
                if (desiredPhase) {
                    phase = project['status'].phase;
                    if (phase !== desiredPhase) {
                        return false;
                    }
                }
                return namespace !== '' && namespace !== null && namespace !== undefined;
            } catch (e) {
                logger.error(`Failed getting project crd ${projectName}; reason: ${JSON.stringify(e)}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsStatusNonEmpty,
            `waitForProjectCrdNonEmptyStatus project name: ${projectName}; project phase was: ${phase}`,
            true,
            90
        );
        return project;
    }

    async waitForProjectCrdUpdated(projectName: string, updatedAt: Date, timeout = 90): Promise<Project> {
        await resolveOrgUnitCrds();
        let projectCrd;
        const projectUpdatedStr = updatedAt.toISOString().split('.')[0] + 'Z';
        const updatedProjectCrd = async (): Promise<boolean> => {
            projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, undefined, 60);
            const crdUpdatedAt = getOrgUnitUpdatedAt(projectCrd);
            if (crdUpdatedAt) {
                const crdUpdatedStr = crdUpdatedAt.split('.')[0] + 'Z';
                return projectCrd['status'] && crdUpdatedStr >= projectUpdatedStr;
            }
            return false;
        };

        await waitUtils.waitForCondition(
            updatedProjectCrd,
            `wait for project crd "${projectName}" to be updated at ${projectUpdatedStr}`,
            true,
            timeout
        );
        return projectCrd;
    }

    async waitForProjectFromBackendNonEmptyStatus(
        projectName: string,
        projectId: string,
        phase?: EProjectPhase
    ): Promise<Project> {
        let project: Project;

        const isProjectsStatusNonEmpty = async (): Promise<boolean> => {
            try {
                project = (
                    await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
                ).body.find((p) => p.id.toString() === projectId);
                const namespace = project.status.namespace;
                const statusNotEmpty = namespace !== '' && namespace !== null && namespace !== undefined;
                if (!phase) {
                    return statusNotEmpty;
                }

                return statusNotEmpty && String(phase) === project.phase;
            } catch (e) {
                logger.error(
                    `Failed getting project ${projectName}/${projectId} from backend; reason: ${JSON.stringify(
                        e
                    )}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isProjectsStatusNonEmpty,
            `waitForProjectFromBackendNonEmptyStatus project name: ${projectName}`,
            true,
            90
        );
        return project;
    }

    async waitForProjectCrdToBeDeleted(projectName: string): Promise<void> {
        const doesProjectExist = async (): Promise<boolean> => {
            try {
                await kubeCli.getCrd(crds.runaiProjectCrd, projectName, undefined);
                return true;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesProjectExist,
            `waitForProjectCrdToBeDeleted project name: ${projectName}`,
            false,
            40
        );
    }

    getRunaiResearcherRoleBindingName(): string {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, this.newRbacMinVersion)) {
            return this.newRbacRunaiResearcherRoleBindingName;
        }
        return this.oldRunaiResearcherRoleBindingName;
    }

    getRunaiResearchManagerRoleBindingName(): string {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, this.newRbacMinVersion)) {
            return this.newRbacRunaiResearchManagerRoleBindingName;
        }
        return this.oldRbacRunaiResearchManagerRoleBindingName;
    }

    getSystemAdministratorRoleBindingName(): string {
        return this.newRbacRunaiSystemAdministratorRoleBindingName;
    }

    async waitForRoleBindingForUserInProject(
        namespace: string,
        userEmail: string,
        roleBindingName: string,
        isClusterRoleBindingInOldVersions = false,
        timeout = 60
    ): Promise<void> {
        // this function is used for waiting for role binding/cluster role binding after assigning user in project
        // in old versions, the role binding was cluster role binding for some roles
        // (researcher manager, system administrator, ml engineer, viewer)
        if (
            isClusterRoleBindingInOldVersions &&
            !versionUtils.isVersionSupported(flowRootState.cluster.version, this.newRbacMinVersion)
        ) {
            await waitUtils.waitForUserInClusterRoleBinding(userEmail, roleBindingName, timeout);
        } else {
            await waitUtils.waitForUserInProjectRoleBinding(namespace, userEmail, roleBindingName, timeout);
        }
    }

    async getProjectIdByProjectName(projectName: string, accessToken?: string): Promise<number> {
        projectsDepartmentsApi.projectsApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        const res = await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid);
        expect(res.response.statusCode).to.be.equal(StatusCodes.OK);
        const projects = res.body;
        const project = projects.find((project) => project.name === projectName);
        return project.id;
    }

    async getProjects(clusterUuid?, accessToken?): Promise<Project[]> {
        clusterUuid = clusterUuid || flowRootState.cluster.uuid;
        projectsDepartmentsApi.projectsApi.accessToken =
            accessToken || flowRootState.users.system_administrator.accessToken;
        return (await projectsDepartmentsApi.projectsApi.getProjects(clusterUuid)).body;
    }

    validateProjectResponse(
        response: { response?: http.IncomingMessage; body?: Project },
        projectName: string,
        deservedGpus = 0,
        maxAllowedGpus = 0,
        gpuOverQuotaWeight = 0,
        namespace = '',
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
        departmentName = 'default',
        interactiveJobTimeLimitSecs = null,
        interactiveJobMaxIdleDurationSecs = null,
        interactivePreemptibleJobMaxIdleDurationSecs = null,
        trainingJobTimeLimitSecs = null,
        trainingJobMaxIdleDurationSecs = null,
        nodeAffinity = {
            train: {
                affinityType: 'no_limit',
                selectedTypes: []
            },
            interactive: {
                affinityType: 'no_limit',
                selectedTypes: []
            }
        }
    ): void {
        const softAssert = new SoftAssert();
        const resourcesNodePoolGpu: NodePoolAssignedResourcesCreateGpu = {
            deserved: deservedGpus,
            maxAllowed: maxAllowedGpus,
            overQuotaWeight: gpuOverQuotaWeight
        };

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
            // The swapEnabled field is not consistently returned, so its assertion remains disabled.
            //softAssert.assert(response.body.swapEnabled !== undefined, 'Expected swapEnabled to be defined');
            softAssert.assert(
                response.body.departmentName === departmentName,
                `Expected department name to be ${departmentName}, got instead ${response.body.departmentName}`
            );
            softAssert.assert(
                typeof response.body.departmentId === 'number',
                `Expected departmentId to be a number, got instead ${typeof response.body.departmentId}`
            );
            softAssert.assert(
                Array.isArray(response.body.status.quotaStatuses),
                'Expected quotaStatuses to be an array'
            );

            softAssert.assert(
                Object.values(EProjectPhase).includes(response.body.phase as EProjectPhase),
                `Expected phase to be one of EProjectPhase, got instead ${response.body.phase}`
            );
            softAssert.assert(
                response.body.tenantId === flowRootState.cluster.tenantId,
                `Expected tenantId to be ${flowRootState.cluster.tenantId}, got instead ${response.body.tenantId}`
            );
            softAssert.assert(
                response.body.clusterUuid === flowRootState.cluster.uuid,
                `Expected clusterUuid to be ${flowRootState.cluster.uuid}, got instead ${response.body.clusterUuid}`
            );
            softAssert.assert(
                response.body.namespace === namespace,
                `Expected namespace to be ${namespace}, got instead ${response.body.namespace}`
            );

            softAssert.assert(
                response.body.nodePoolsResources.length === numberOfNodePools,
                `Expected number of node pools to be ${numberOfNodePools}, got instead ${response.body.nodePoolsResources.length}`
            );
            softAssert.assert(
                response.body.nodePoolsResources[0].nodePool.name === 'default',
                `Expected first node pool name to be default, got instead ${response.body.nodePoolsResources[0].nodePool.name}`
            );
            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateGpu>(
                response.body.nodePoolsResources[0].gpu,
                resourcesNodePoolGpu,
                'Expected node pool GPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateCpu>(
                response.body.nodePoolsResources[0].cpu,
                resourcesNodePoolCpu,
                'Expected node pool CPU resources to match'
            );
            softAssert.assertDeepNonStrictEqual<NodePoolAssignedResourcesCreateMemory>(
                response.body.nodePoolsResources[0].memory,
                resourcesNodePoolMemory,
                'Expected node pool memory resources to match'
            );

            softAssert.assert(
                Array.isArray(response.body.permissions.users) &&
                    response.body.permissions.users.includes(config.controlPlaneAdminUsername),
                `Expected users to include controlPlaneAdminUsername, got instead ${response.body.permissions.users}`
            );
            softAssert.assert(
                Array.isArray(response.body.permissions.groups),
                'Expected groups to be an array'
            );
            softAssert.assert(
                Array.isArray(response.body.defaultNodePools) &&
                    response.body.defaultNodePools.includes('default'),
                `Expected defaultNodePools to include default, got instead ${response.body.defaultNodePools}`
            );

            softAssert.assertNonStrictEqual(
                response.body.interactiveJobTimeLimitSecs,
                interactiveJobTimeLimitSecs,
                `Expected interactiveJobTimeLimitSecs to be ${interactiveJobTimeLimitSecs}, got instead ${response.body.interactiveJobTimeLimitSecs}`
            );
            softAssert.assertNonStrictEqual(
                response.body.interactiveJobMaxIdleDurationSecs,
                interactiveJobMaxIdleDurationSecs,
                `Expected interactiveJobMaxIdleDurationSecs to be ${interactiveJobMaxIdleDurationSecs}, got instead ${response.body.interactiveJobMaxIdleDurationSecs}`
            );
            softAssert.assertNonStrictEqual(
                response.body.interactivePreemptibleJobMaxIdleDurationSecs,
                interactivePreemptibleJobMaxIdleDurationSecs,
                `Expected interactivePreemptibleJobMaxIdleDurationSecs to be ${interactivePreemptibleJobMaxIdleDurationSecs}, got instead ${response.body.interactivePreemptibleJobMaxIdleDurationSecs}`
            );
            softAssert.assertNonStrictEqual(
                response.body.trainingJobTimeLimitSecs,
                trainingJobTimeLimitSecs,
                `Expected trainingJobTimeLimitSecs to be ${trainingJobTimeLimitSecs}, got instead ${response.body.trainingJobTimeLimitSecs}`
            );
            softAssert.assertNonStrictEqual(
                response.body.trainingJobMaxIdleDurationSecs,
                trainingJobMaxIdleDurationSecs,
                `Expected trainingJobMaxIdleDurationSecs to be ${trainingJobMaxIdleDurationSecs}, got instead ${response.body.trainingJobMaxIdleDurationSecs}`
            );

            try {
                expect(response.body.nodeAffinity).to.deep.equal(
                    nodeAffinity,
                    `Expected nodeAffinity to be ${JSON.stringify(nodeAffinity)}, got instead ${JSON.stringify(response.body.nodeAffinity)}`
                );
            } catch (error) {
                softAssert.assert(false, `Expected nodeAffinity to match: ${error.message}`);
            }
        }

        softAssert.throwErrorIfFailed();
    }

    async validateProjectCrd(
        projectName: string,
        departmentName = 'default',
        nodePoolName = 'default',
        timeout = 30,
        projectId?: number,
        namespace?: string,
        propertyToWait?: string,
        propertyValue?: string,
        subPropertyToWait: string = undefined
    ): Promise<void> {
        if (namespace === undefined) {
            namespace = Utils.namespaceFromProjectName(projectName);
        }
        const projectCrd = await waitUtils.waitForCrd(crds.runaiProjectCrd, projectName, '', timeout, true);
        if (propertyToWait) {
            await waitUtils.waitForCrdProperty(
                crds.runaiProjectCrd,
                projectName,
                '',
                propertyToWait,
                propertyValue,
                subPropertyToWait,
                timeout
            );
        }
        expect(projectCrd).to.be.ok;
        expect(projectCrd).property('metadata').property('name').to.equal(projectName);
        expect(projectCrd).property('metadata').property('creationTimestamp').to.be.ok;
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)) {
            expect(projectCrd)
                .property('metadata')
                .property('labels')
                .property('runai/project-id')
                .to.satisfy((id) => (projectId ? id === String(projectId) : id !== undefined));
        }
        expect(projectCrd).property('status').property('namespace').to.equal(namespace);
        expect(projectCrd).property('status').property('phase').equal('Ready');
        // The parent org unit is spec.department on the legacy CRD and spec.parent on the KAI one.
        expect(getProjectParent(projectCrd)).to.equal(departmentName);
        expect(projectCrd)
            .property('spec')
            .property('defaultNodePools')
            .to.be.an('array')
            .include(nodePoolName);
    }

    async validateProjectCrdSpec(
        projectName: string,
        timeout = 30,
        projectId?: number,
        namespace?: string,
        specPropertyToWait?: string,
        specPropertyValue?: unknown
    ): Promise<void> {
        if (namespace === undefined) {
            namespace = Utils.namespaceFromProjectName(projectName);
        }
        const crdFieldSet = async (): Promise<boolean> => {
            try {
                const crdObj = await kubeCli.getCrd(crds.runaiProjectCrd, projectName, '');
                if (!crdObj) {
                    return false;
                }
                // The job-limit fields moved from spec to run.ai/*-seconds annotations on the KAI
                // CRD; read them through the accessor so callers keep using the legacy field name.
                if (specPropertyToWait in projectJobLimitFields) {
                    return (
                        getProjectJobLimitSecs(crdObj, specPropertyToWait as ProjectJobLimitField) ===
                        specPropertyValue
                    );
                }
                if (!crdObj['spec']) {
                    return false;
                }

                return crdObj['spec'][specPropertyToWait] === specPropertyValue;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            crdFieldSet,
            `wait for project crd ${specPropertyToWait} to be ${specPropertyValue}`,
            true,
            timeout
        );
    }

    async validateProjectQueueCrd(
        queueName: string,
        project: Project,
        timeout = 30,
        desiredResult = true
    ): Promise<void> {
        await this.validateQueueCrdOfProject(
            queueName,
            project.name,
            project.resources.gpu.deserved,
            project.resources.gpu.maxAllowed,
            project.resources.gpu.overQuotaWeight,
            project.resources.cpu.deserved,
            project.resources.cpu.maxAllowed,
            project.resources.cpu.overQuotaWeight,
            project.resources.memory.deserved,
            project.resources.memory.maxAllowed,
            project.resources.memory.overQuotaWeight,
            null,
            timeout,
            desiredResult
        );
    }

    async validateQueueCrdOfProject(
        queueName: string,
        projectName: string,
        gpuDeserved: number,
        gpuLimit: number,
        gpuOverQuotaWeight: number,
        cpuDeserved: number = null,
        cpuLimit: number = null,
        cpuOverQuotaWeight: number = null,
        memoryDeserved: number = null,
        memoryLimit: number = null,
        memoryOverQuotaWeight: number = null,
        priorityName: string = null,
        timeout = 30,
        desiredResult = true
    ): Promise<void> {
        if (!versionUtils.isVersionSupported(flowRootState.cluster.version, '2.13')) {
            return;
        }

        const quotaName = versionUtils.isVersionSupported(
            flowRootState.cluster.version,
            SupportedVersions.V2_16
        )
            ? 'quota'
            : 'deserved';
        const queueCrd = await waitUtils.waitForCrd(crds.queueCrd, queueName, '', timeout, desiredResult);
        expect(queueCrd).property('metadata');
        expect(queueCrd).property('spec');
        expect(queueCrd['metadata']).property('labels');
        expect(queueCrd['metadata']['labels']).property('project').equal(projectName);
        expect(queueCrd['metadata']).property('name').equal(queueName).equal(projectName);
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)) {
            expect(queueCrd['spec']).property('displayName').equal(projectName);
        }
        expect(queueCrd['spec']).property('resources');
        expect(queueCrd['spec']['resources']).property('gpu').property(quotaName).equal(gpuDeserved);
        expect(queueCrd['spec']['resources']).property('gpu').property('limit').equal(gpuLimit);
        expect(queueCrd['spec']['resources'])
            .property('gpu')
            .property('overQuotaWeight')
            .equal(gpuOverQuotaWeight);
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
            const priorityValue = this.mapPriorityToNumber(priorityName);
            expect(queueCrd['spec']).property('priority').equal(priorityValue);
        }
        validatePropertyIfNotNull(queueCrd['spec']['resources']['cpu'], cpuDeserved, quotaName, -1);
        validatePropertyIfNotNull(queueCrd['spec']['resources']['cpu'], cpuLimit, 'limit', -1);
        validatePropertyIfNotNull(
            queueCrd['spec']['resources']['cpu'],
            cpuOverQuotaWeight,
            'overQuotaWeight',
            1
        );
        validatePropertyIfNotNull(queueCrd['spec']['resources']['memory'], memoryDeserved, quotaName, -1);
        validatePropertyIfNotNull(queueCrd['spec']['resources']['memory'], memoryLimit, 'limit', -1);
        validatePropertyIfNotNull(
            queueCrd['spec']['resources']['memory'],
            memoryOverQuotaWeight,
            'overQuotaWeight',
            1
        );
    }
}
