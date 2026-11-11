/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    computeResourceUtils,
    crds,
    credentialUtils,
    dataSourcesUtils,
    environmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    orgUnitUtils,
    PodStatus,
    reportApiUtils,
    serviceAccountUtils,
    Utils,
    versionUtils,
    waitUtils,
    workloadsUtilsV2,
    workloadUtils,
    workspaceTemplateUtils
} from '.';
import { orgUnitService, projectsDepartmentsApi } from '../api';
import { bashCli, kubeCli } from '../cli';
import { Project } from '../api/apis/projectsDepartmentsApi';
import { CustomResourceDefinition, Scope } from './models/customResourceDefinition';
import { KubernetesObject, V1Deployment } from '@kubernetes/client-node';
import { expect } from 'chai';
import { SupportedVersions } from './utils/SupporetedEnvsUtils';
import { StatusCodes } from 'http-status-codes';

export class Operations {
    RunaiProjectControllerName = 'runai-project-controller';

    RunaiClusterSyncName = 'cluster-sync';

    RunaiSchedulerName = 'runai-scheduler';

    ExcludedNodepoolName = 'runai-excluded-nodes';

    async toggleRunaiOperand(operandName: string, shouldEnable: boolean): Promise<void> {
        const patch = {
            spec: {
                [operandName]: {
                    enabled: shouldEnable
                }
            }
        };

        try {
            const currentConfig = await kubeCli.getCrd(crds.runaiConfigCrd, 'runai', 'runai');
            const updatedConfig = this.deepMerge(currentConfig, patch);
            await kubeCli.patchCrd(crds.runaiConfigCrd, updatedConfig, 'runai', 'runai');
            logger.info(`runaiconfig patched: ${operandName}, enabled: ${shouldEnable}`);
        } catch (err) {
            logger.info(`Error patching runaiconfig: ${operandName}, err: ${err}`);
            throw err;
        }

        await waitUtils.waitForPodPrefix(operandName, 'runai', shouldEnable);
    }

    async deleteProjectAndNamespace(
        projectId: number,
        accessToken?: string,
        clusterUuid?: string
    ): Promise<void> {
        orgUnitService.projectApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.system_administrator.accessToken;
        const project = (await orgUnitService.projectApi.getProject(projectId.toString())).body;
        await orgUnitService.projectApi.deleteProject(projectId.toString());
        const relatedNamespaceName = project.status.namespace || Utils.namespaceFromProjectName(project.name);
        await kubeCli.CoreV1Api.deleteNamespace({ name: relatedNamespaceName });
        logger.info(`Project ${project.name} and namespace ${relatedNamespaceName} deleted`);
    }

    async deleteProjectNamespaceAndWorkloads(project: Project, timeout: number): Promise<void> {
        await workloadUtils.clearProjectWorkloadsAndPods(project.name, timeout);
        await this.deleteProjectAndNamespace(project.id);
    }

    async deleteDeployment(deploymentName: string, namespaceName: string, clusterId: string): Promise<void> {
        await kubeCli.AppsV1Api.deleteNamespacedDeployment({
            name: deploymentName,
            namespace: namespaceName
        });
    }

    private async removeAllFinalizersFromCRs(
        crd: CustomResourceDefinition,
        namespaces?: string[]
    ): Promise<void> {
        let crs: KubernetesObject[] = [];

        switch (crd.scope) {
            case Scope.CLUSTER: {
                const crsObjs = await kubeCli.getCrds(crd);
                crs = crs.concat(crsObjs);
                break;
            }
            case Scope.NAMESPACE: {
                if (!namespaces) {
                    // Get all namespaces
                    namespaces = (await kubeCli.CoreV1Api.listNamespace()).items.map(
                        (ns) => ns.metadata.name
                    );
                }

                await Promise.all(
                    namespaces.map(async (namespace) => {
                        crs = crs.concat(await kubeCli.getCrds(crd, namespace));
                    })
                );
                break;
            }
            default: {
                throw new Error(`Unsupported scope: ${crd.scope}`);
            }
        }

        if (crs.length === 0) {
            return;
        }

        logger.debug(`Removing finalizers from CRs: ${crs.map((cr) => `${cr.kind}/${cr.metadata.name}`)}`);

        for (const cr of crs) {
            // prettier-ignore
            const cmd =
                `kubectl patch ${crd.plural} ${cr.metadata?.name} ${cr.metadata?.namespace ? '-n ' + cr.metadata.namespace : ''} --type json -p '[{"op": "remove", "path": "/metadata/finalizers"}]'`;

            try {
                // Using bashCli instead of kubeCli because kubeCli patch fail patching resources that are marked for deletion for some reason
                bashCli.exec(cmd);
            } catch (err) {
                logger.debug(`Failed to remove finalizers from CR ${cr.kind}/${cr.metadata.name}: ${err}`);
            }
        }
    }

    private async deleteAllPodsInNamespace(namespace: string): Promise<void> {
        await kubeCli.CoreV1Api.deleteCollectionNamespacedPod({
            namespace,
            gracePeriodSeconds: 0
        });
    }

    // Forcefully deletes the following:
    // - Data sources
    // - Credentials
    // - Computed resources with prefix "compute"
    // - Environments with prefix "environment"
    // - Projects
    // - Templates
    // - Workloads
    // - Cluster scoped NodePools
    // - Consumption reports (NodePool, Project, Department)
    // - Applications with prefix: test-app / e2e-app
    // - RunAI Managed Namespaces
    async forceCleanup(clusterId = flowRootState.cluster.uuid): Promise<void> {
        projectsDepartmentsApi.projectsApi.accessToken = flowRootState.users.editor.accessToken;
        projectsDepartmentsApi.departmentsApi.accessToken = flowRootState.users.editor.accessToken;
        orgUnitService.departmentApi.accessToken = flowRootState.users.editor.accessToken;
        try {
            logger.debug('Deleting all project scoped assets from the control plane');
            await this.deleteAssets(clusterId);
        } catch (error) {
            logger.warn(`Failed to delete all project scoped assets: ${error}`);
        }

        try {
            logger.debug('Deleting all workloads from the control plane');
            await workloadsUtilsV2.deleteAllWorkloads();
        } catch (error) {
            logger.warn(`Failed to workloads: ${error}`);
        }

        try {
            logger.debug('Deleting all projects from the control plane');
            await this.deleteAllProjects(clusterId);
        } catch (error) {
            logger.warn(`Failed to delete all projects: ${error}`);
        }

        try {
            logger.debug('Deleting all departments from the control plane');
            await this.deleteNonDefaultDepartments(clusterId);
        } catch (error) {
            logger.warn(`Failed to delete all departments: ${error}`);
        }

        try {
            logger.debug('Deleting all runai nodepools from the control plane');
            await nodePoolUtils.deleteAllClusterNodePools(undefined, clusterId);
        } catch (error) {
            logger.warn(`Failed to delete all nodepools: ${error}`);
        }
        try {
            logger.debug('Deleting all projects namespaces');
            await this.deleteAllProjectsNamespaces();
        } catch (error) {
            logger.warn(`Failed to delete all projects namespaces: ${error}`);
        }

        try {
            await waitUtils.waitForCondition(async () => {
                const projectNamespacesNames = await this.listProjectsNamespaces();

                if (projectNamespacesNames.length === 0) {
                    return true;
                }

                // Workloads sometimes get stuck terminating. Remove forced deletion once normal cleanup reliably completes.
                await Promise.all([
                    this.removeAllFinalizersFromCRs(crds.runaiTrainingWorkloadCrd, projectNamespacesNames),
                    this.removeAllFinalizersFromCRs(crds.runaiInferenceWorkloadCrd, projectNamespacesNames),
                    this.removeAllFinalizersFromCRs(crds.runaiInteractiveWorkloadCrd, projectNamespacesNames),
                    this.deleteAllPodsInNamespaces(projectNamespacesNames)
                ]);
                return false;
            }, 'runai projects namespaces to be deleted');
        } catch (error) {
            logger.debug(`Failed to wait for waitForCondition: ${error}`);
        }
        try {
            await serviceAccountUtils.deleteAllServiceAccounts(
                flowRootState.users.system_administrator.accessToken,
                ['test-app', 'e2e-app', 'login-application', 'cli-v2', 'login-access-key']
            );
            logger.debug('All service accounts deleted');
        } catch (error) {
            logger.warn(`Failed to delete all service accounts: ${error}`);
        }
        for (const id of [
            flowRootState.reportIds.consumptionReportIdByNodePool,
            flowRootState.reportIds.consumptionReportIdByProject,
            flowRootState.reportIds.consumptionReportIdByDepartment
        ]) {
            if (id) {
                try {
                    await reportApiUtils.deleteReport(
                        id,
                        flowRootState.users.system_administrator.accessToken
                    );
                } catch (error) {
                    if (error.statusCode !== StatusCodes.NOT_FOUND) {
                        logger.warn(`Failed to delete report ${id}: ${error}`);
                    }
                }
            }
        }

        logger.debug('Force cleanup completed successfully');
    }

    private async deleteAllProjectsNamespaces(): Promise<void> {
        const projectNamespacesNames = await this.listProjectsNamespaces();

        for (const nsName of projectNamespacesNames) {
            await kubeCli.CoreV1Api.deleteNamespace({ name: nsName, gracePeriodSeconds: 0 });
            logger.debug(`Deleted project namespace: ${nsName}`);
        }
    }

    private async listProjectsNamespaces(): Promise<string[]> {
        const namespaces = await kubeCli.CoreV1Api.listNamespace();
        const runaiProjectNamespaces = namespaces.items.filter(
            (namespace) =>
                namespace.metadata && namespace.metadata.labels && namespace.metadata.labels['runai/queue']
        );
        return runaiProjectNamespaces.map((namespace) => namespace.metadata.name);
    }

    private async deleteNonDefaultDepartments(clusterId = flowRootState.cluster.uuid): Promise<void> {
        const departments = (await orgUnitService.departmentApi.getDepartments(['clusterId==' + clusterId]))
            .body.departments;

        logger.debug('Deleting all non-default departments');
        for (const department of departments) {
            if (department.name !== 'default') {
                await orgUnitUtils.deleteDepartment(department.id.toString());
            }
        }
    }

    async submitWorkloads(
        workloads: BaseWorkload[],
        expectedStatus?: PodStatus,
        timeout = 320
    ): Promise<void> {
        await Promise.all(
            workloads.map((workload) => {
                logger.debug(`Creating workload ${workload.name} with image : ${workload.image}`);
                return kubeCli.createCrd(
                    workload.getCrd(),
                    workload.getBody(),
                    workload.testingSpecificFields.namespace
                );
            })
        ).catch((err: unknown) => {
            expect.fail('failed to create workload' + err);
        });

        if (expectedStatus) {
            await Promise.all(
                workloads.map((workload) =>
                    waitUtils.waitForWorkloadStatus(workload, expectedStatus, timeout)
                )
            );
        }
    }

    async deleteWorkloads(workloads: BaseWorkload[]): Promise<void> {
        await Promise.all(
            workloads.map(async (workload) => {
                try {
                    await kubeCli.deleteCrd(
                        workload.getCrd(),
                        workload.name,
                        workload.testingSpecificFields.namespace
                    );
                    logger.debug('Deleted workload: ' + workload.name);
                } catch (e) {
                    logger.error(
                        `Failed to delete workload ${workload.name} in namespace ${workload.testingSpecificFields.namespace}: ${e}`
                    );
                    throw e;
                }
            })
        );

        await Promise.all(
            workloads.map(async (workload) => {
                try {
                    await waitUtils.waitForWorkloadNotToExist(workload);
                } catch (e) {
                    logger.error(
                        `workload ${workload.name} exist in namespace ${workload.testingSpecificFields.namespace}, error: ${e}`
                    );
                    throw e;
                }
            })
        );
    }

    async getDeployment(namespace: string, deploymentName: string): Promise<V1Deployment> {
        logger.debug(`Getting deployment ${deploymentName} in namespace ${namespace}`);
        return await kubeCli.AppsV1Api.readNamespacedDeployment({ name: deploymentName, namespace });
    }

    private async deleteAllPodsInNamespaces(namespaces: string[]): Promise<void> {
        await Promise.all(
            namespaces.map(async (namespace) => {
                await this.deleteAllPodsInNamespace(namespace);
            })
        );
    }

    async deleteAssets(
        clusterId = flowRootState.cluster.uuid,
        environmentPrefix = 'environment',
        computePrefix = 'compute'
    ): Promise<void> {
        await Promise.all([
            environmentUtils.deleteEnvironments(environmentPrefix),
            dataSourcesUtils.deleteAllDataSources(),
            computeResourceUtils.deleteComputeResources(computePrefix),
            credentialUtils.deleteAllCredentials(),
            workspaceTemplateUtils.deleteAllTemplates()
        ]);
    }

    // Doesn't wait for the projects to be deleted in the cluster
    async deleteAllProjects(clusterId = flowRootState.cluster.uuid): Promise<void> {
        projectsDepartmentsApi.projectsApi.accessToken =
            flowRootState.users.system_administrator.accessToken;
        orgUnitService.projectApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const projects = (await orgUnitService.projectApi.getProjects([`clusterId==${clusterId}`])).body
            .projects;

        for (const project of projects) {
            try {
                await orgUnitService.projectApi.deleteProject(project.id);
            } catch (error) {
                logger.warn(`Failed to delete project ${project.name}: ${error}`);
            }
        }
    }

    async updateRunaiConfig(patch: object): Promise<void> {
        try {
            const currentConfig = await kubeCli.getCrd(crds.runaiConfigCrd, 'runai', 'runai');
            const updatedConfig = this.deepMerge(currentConfig, patch);
            await kubeCli.patchCrd(crds.runaiConfigCrd, updatedConfig, 'runai', 'runai');
            logger.info(`runaiconfig patched: ${JSON.stringify(patch)}`);
        } catch (err) {
            logger.error(`Error patching runaiconfig: ${JSON.stringify(patch)}, err: ${err}`);
            throw err;
        }
    }

    private deepMerge(target: any, source: any): any {
        const output = Object.assign({}, target);
        if (this.isObject(target) && this.isObject(source)) {
            Object.keys(source).forEach((key) => {
                if (this.isObject(source[key])) {
                    if (!(key in target)) {
                        Object.assign(output, { [key]: source[key] });
                    } else {
                        output[key] = this.deepMerge(target[key], source[key]);
                    }
                } else {
                    Object.assign(output, { [key]: source[key] });
                }
            });
        }
        return output;
    }

    private isObject(item: any): boolean {
        return item && typeof item === 'object' && !Array.isArray(item);
    }

    async setSchedulerPlacementStrategy(strategy: string): Promise<void> {
        const patch = {
            spec: {
                'runai-scheduler': {
                    placementStrategy: strategy
                }
            }
        };

        logger.info(`Updating scheduler placement strategy to "${strategy}"`);
        return await this.updateRunaiConfig(patch);
    }

    async waitForProjectControllerNamespacesConfiguration(enabled: boolean): Promise<void> {
        const projControllerConfiguredProperly = async (): Promise<boolean> => {
            const projController = await kubeCli.getDeployment(this.RunaiProjectControllerName, 'runai');
            const containers = projController.spec.template.spec.containers;
            const projControllerContainer = containers.find(
                (containers) => containers.name === this.RunaiProjectControllerName
            );
            if (!projControllerContainer) {
                return false;
            }

            const args = projControllerContainer.args;
            const namespacesIndex = args.indexOf('--namespaces');
            if (enabled) {
                return namespacesIndex >= 0;
            }
            return namespacesIndex === -1;
        };

        await waitUtils.waitForCondition(
            projControllerConfiguredProperly,
            `waitForProjectControllerNamespacesConfiguration to have CreateNamespaces enabled: ${enabled}`,
            true,
            70
        );
    }

    async toggleNamespaceCreation(enable: boolean): Promise<void> {
        const patch = {
            spec: {
                'project-controller': {
                    createNamespaces: enable
                }
            }
        };

        logger.info(`Updating project-controller with createNamespaces=${enable}`);
        await this.updateRunaiConfig(patch);

        await this.waitForProjectControllerNamespacesConfiguration(enable);
        await waitUtils.waitForDeploymentToBeAvailable(this.RunaiProjectControllerName, 'runai');
    }

    async disableNamespaceCreation(): Promise<void> {
        await this.toggleNamespaceCreation(false);
    }

    async enableNamespaceCreation(): Promise<void> {
        await this.toggleNamespaceCreation(true);
    }

    async toggleWorkloadOwnershipProtection(enable: boolean): Promise<void> {
        const patch = {
            spec: {
                global: {
                    enableWorkloadOwnershipProtection: enable
                }
            }
        };

        logger.info(`Updating runaiconfig global with enableWorkloadOwnershipProtection=${enable}`);
        await this.updateRunaiConfig(patch);

        await waitUtils.waitForWorkloadOwnershipProtectionWebhook(enable);
    }

    async disableWorkloadOwnershipProtection(): Promise<void> {
        await this.toggleWorkloadOwnershipProtection(false);
    }

    async enableWorkloadOwnershipProtection(): Promise<void> {
        await this.toggleWorkloadOwnershipProtection(true);
    }

    async updateClusterSyncConfig(
        maxNumOfPodGroupsPerRequest: string,
        maxNumOfPodsPerRequest: string,
        enableClusterSync = true
    ): Promise<void> {
        const clusterSyncSpecName = versionUtils.isVersionSupported(
            flowRootState.cluster.version,
            SupportedVersions.V2_23
        )
            ? 'maxNumOfWorkloadsPerRequest'
            : versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)
              ? 'MaxNumOfPodGroupsPerRequest'
              : 'maxNumOfPodGroupsPerRequest';

        const maxNumOfPodsPerRequestName = !versionUtils.isVersionSupported(
            flowRootState.cluster.version,
            SupportedVersions.V2_16
        )
            ? 'MaxNumOfPodsPerRequest'
            : 'maxNumOfPodsPerRequest';

        const patch = {
            spec: {
                [this.RunaiClusterSyncName]: {
                    enabled: enableClusterSync,
                    [clusterSyncSpecName]: maxNumOfPodGroupsPerRequest
                        ? Number(maxNumOfPodGroupsPerRequest)
                        : null,
                    [maxNumOfPodsPerRequestName]: maxNumOfPodsPerRequest
                        ? Number(maxNumOfPodsPerRequest)
                        : null
                }
            }
        };

        logger.info('Updating runaiconfig cluster-sync');
        await this.updateRunaiConfig(patch);

        const envVarName = versionUtils.isVersionSupported(
            flowRootState.cluster.version,
            SupportedVersions.V2_23
        )
            ? 'MAX_NUM_OF_WORKLOADS_PER_REQUEST'
            : 'MAX_NUM_OF_POD_GROUPS_PER_REQUEST';

        if (maxNumOfPodGroupsPerRequest && enableClusterSync) {
            await this.waitForRunaiServiceEnvVarConfiguration(
                this.RunaiClusterSyncName,
                maxNumOfPodGroupsPerRequest,
                envVarName
            );
        }

        await waitUtils.waitForDeploymentToBeAvailable(
            this.RunaiClusterSyncName,
            'runai',
            120,
            enableClusterSync
        );
    }

    async reduceSchedulerResources(
        requestedCpu = '200m',
        requestedMemory = '256Mi',
        limitCpu = '1000m',
        limitMemory = '512Mi'
    ): Promise<void> {
        const schedulerDeploymentName = this.RunaiSchedulerName + '-default';

        const patch = {
            spec: {
                [this.RunaiSchedulerName]: {
                    resources: {
                        requests: {
                            cpu: requestedCpu,
                            memory: requestedMemory
                        },
                        limits: {
                            cpu: limitCpu,
                            memory: limitMemory
                        }
                    }
                }
            }
        };

        logger.info('Reducing runai-scheduler resources (updating runaiconfig)');
        await this.updateRunaiConfig(patch);

        await this.waitForRunaiServiceResourcesConfiguration(
            schedulerDeploymentName,
            requestedCpu,
            this.RunaiSchedulerName
        );

        await waitUtils.waitForDeploymentToBeAvailable(schedulerDeploymentName, 'runai', 120);
    }

    async revertSchedulerResourcesToDefault(): Promise<void> {
        const patch = {
            spec: {
                [this.RunaiSchedulerName]: null
            }
        };

        logger.info('Reverting runai-scheduler resources (updating runaiconfig)');
        await this.updateRunaiConfig(patch);
    }

    async waitForRunaiServiceEnvVarConfiguration(
        runaiServiceName: string,
        expectedEnvVarValue: string,
        envVarName: string
    ): Promise<void> {
        const runaiServiceEnvVarAsExpected = async (): Promise<boolean> => {
            try {
                const runaiServiceDeployment = await kubeCli.getDeployment(runaiServiceName, 'runai');
                const containers = runaiServiceDeployment.spec.template.spec.containers;
                const container = containers.find((containers) => containers.name.includes(runaiServiceName));
                const envs = container.env;
                const foundEnvVar = envs.find((env) => env.name === envVarName);
                return foundEnvVar && foundEnvVar.value === expectedEnvVarValue;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            runaiServiceEnvVarAsExpected,
            `waitForRunaiServiceEnvVarConfiguration (${runaiServiceName}) to have ${envVarName}=${expectedEnvVarValue}`,
            true,
            70
        );

        logger.debug(
            `Successfully waited for ${runaiServiceName} deployment to have env ${envVarName}=${expectedEnvVarValue}`
        );
    }

    async waitForRunaiServiceResourcesConfiguration(
        runaiDeploymentName: string,
        expectedRequestedCpu: string,
        runaiServiceName = ''
    ): Promise<void> {
        if (!runaiServiceName || runaiServiceName === '') {
            runaiServiceName = runaiDeploymentName;
        }

        const runaiServiceResourcesAsExpected = async (): Promise<boolean> => {
            try {
                const runaiServiceDeployment = await kubeCli.getDeployment(runaiDeploymentName, 'runai');
                const containers = runaiServiceDeployment.spec.template.spec.containers;
                const container = containers.find((containers) => containers.name.includes(runaiServiceName));
                const requestedCpu = container.resources.requests['cpu'];
                return requestedCpu === expectedRequestedCpu;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            runaiServiceResourcesAsExpected,
            `waitForRunaiServiceResourcesConfiguration (${runaiServiceName}) to have resources.requests.cpu=${expectedRequestedCpu}`,
            true,
            70
        );

        logger.debug(
            `Successfully waited for ${runaiServiceName} deployment to have resources.requests.cpu=${expectedRequestedCpu}`
        );
    }

    async getWorkload(name: string, namespace: string, workloadType: string): Promise<KubernetesObject> {
        if (workloadType === crds.runaiInteractiveWorkloadCrd.kind) {
            return await kubeCli.getInteractiveWorkload(name, namespace);
        } else if (workloadType === crds.runaiTrainingWorkloadCrd.kind) {
            return await kubeCli.getTrainWorkloadOfRunaiJob(name, namespace);
        } else if (workloadType === crds.runaiInferenceWorkloadCrd.kind) {
            return await kubeCli.getInferenceWorkload(name, namespace);
        }
        return null;
    }
}
