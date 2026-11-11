/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    crds,
    logger,
    BaseWorkload,
    waitUtils,
    Utils,
    TrainingWorkload,
    InteractiveWorkload,
    InferenceWorkload,
    config,
    WorkloadConditionType,
    objectGenerator,
    projectResourcesUtils,
    operations,
    DistributedJobType,
    DistributedWorkload,
    CustomResourceDefinition,
    flowRootState,
    versionUtils,
    workloadsUtilsV2
} from '../index';
import { kubeCli, runaiCliV2 } from '../../cli';
import { Project } from '../../api/apis/projectsDepartmentsApi';
import {
    WorkloadCondition,
    WorkloadStatus,
    Pvc,
    Nfs,
    Volume,
    ConfigMap,
    EmptyDir,
    GitSync
} from '../models/baseWorkload';
import { policyApi, workloadsService } from '../../api';
import { Action } from '../../api/apis/workloadsApi/generated/model/action';
import { BatchResponse } from '../../api/apis/workloadsApi/generated/model/batchResponse';
import { supportedFrom, SupportedVersions } from './SupporetedEnvsUtils';
import { InferencePolicyChangeRequestV2, InferencePolicyV2 } from '../../api/apis/policyApi';
import { WorkloadPriority } from '../../api/apis/workloadsApi/generated/model/workloadPriority';
import { ClusterPriority } from '../../api/apis/workloadsApi/generated/model/clusterPriority';
import { WorkloadTypeConfig } from '../../api/apis/workloadsApi/generated/model/workloadTypeConfig';
import { WorkloadTypeUpdateFields } from '../../api/apis/workloadsApi/generated/model/workloadTypeUpdateFields';
import { WorkloadTypePreemptibility } from '../../api/apis/workloadsApi/generated/model/workloadTypePreemptibility';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import http from 'http';
import { cliV2WorkloadType, SubmitFlagsMerged } from '../../cli/tools/RunaiCliV2.constants';
import {
    AnyWorkloadSubmissionType,
    getWorkloadSubmissionRoute,
    WorkloadSubmissionRoute
} from './AnyWorkloadSubmissionUtils';
import { findWorkloadRunnerByWorkloadName, WorkloadRunnerResource } from './WorkloadRunnerUtils';

const workers = 2;
const numberOfGpus = 1;

export enum WorkloadType {
    Interactive,
    Train,
    Mpi,
    Pytorch,
    Tensor,
    Xgboost
}

const workloadTypeToImage: Record<number, string> = {};
workloadTypeToImage[WorkloadType.Interactive] = config.quickStartImage.interactive;
workloadTypeToImage[WorkloadType.Train] = config.quickStartImage.training;
workloadTypeToImage[WorkloadType.Pytorch] = config.quickStartImage.distributedPytorch;
workloadTypeToImage[WorkloadType.Mpi] = config.quickStartImage.distributedMpi;
workloadTypeToImage[WorkloadType.Tensor] = config.quickStartImage.distributedTf;

export enum EWorkloadStatusToPhase {
    'Succeeded' = 'Completed',
    'Running' = 'Running',
    'Pending' = 'Pending',
    'Failed' = 'Failed',
    'Error' = 'Failed',
    'Suspended' = 'Stopped'
}

const distributedWorkloadTypes: WorkloadType[] = [
    WorkloadType.Pytorch,
    WorkloadType.Tensor,
    WorkloadType.Mpi,
    WorkloadType.Xgboost
];

export class TestEnvironment {
    projectName: string;

    project: Project;

    namespace: string;

    workload: BaseWorkload;

    jobName: string;

    constructor(
        projectName: string,
        project: Project,
        namespace: string,
        workload: BaseWorkload,
        jobName: string
    ) {
        this.projectName = projectName;
        this.project = project;
        this.namespace = namespace;
        this.workload = workload;
        this.jobName = jobName;
    }
}

export interface InferenceWorkloadCreationOptions {
    image?: string;

    gpus?: number;

    command?: string;

    minScale?: number;

    maxScale?: number;

    initialScale?: number;

    activationScale?: number;

    scaleToZeroRetentionSeconds?: number;

    scaleDownDelaySeconds?: number;

    initializationTimeoutSeconds?: number;

    metric?: string;

    metricName?: string;

    target?: number;

    targetUtilizationPercentage?: string;

    concurrencyHardLimit?: number;

    mps?: boolean;

    largeShm?: boolean;

    pvcs?: { [key: string]: Pvc };

    nfs?: { [key: string]: Nfs };

    volumes?: { [key: string]: Volume };

    gitSync?: { [key: string]: GitSync };

    configMapVolumes?: { [key: string]: ConfigMap };

    emptyDirVolumes?: { [key: string]: EmptyDir };

    readinessProbe?: object;

    servingPortAuthorizationType?: string;

    servingPortAuthorizedUsers?: string;

    servingPortAuthorizedGroups?: string;

    workingDir?: string;
}

export interface SubmissionRoot {
    crd: CustomResourceDefinition;
    name: string;
    resource: Record<string, any>;
    route: WorkloadSubmissionRoute;
}

export class WorkloadUtils {
    async batchWorkloads(action: Action, workloadsIds: string[]): Promise<BatchResponse> {
        workloadsService.workloadsBatchApi.accessToken = flowRootState.users.system_administrator.accessToken;

        const response = await workloadsService.workloadsBatchApi.batchWorkloads({
            ids: workloadsIds,
            action
        });
        return response.body;
    }

    async deleteMultipleWorkloads(workloadsIds: string[]): Promise<BatchResponse> {
        return this.batchWorkloads(Action.Delete, workloadsIds);
    }

    async suspendMultipleWorkloads(workloadsIds: string[]): Promise<BatchResponse> {
        return this.batchWorkloads(Action.Suspend, workloadsIds);
    }

    async resumeMultipleWorkloads(workloadsIds: string[]): Promise<BatchResponse> {
        return this.batchWorkloads(Action.Resume, workloadsIds);
    }

    getDistributedWorkload(
        workloadType: number,
        namespace: string,
        jobName: string,
        image: string
    ): DistributedWorkload {
        let jobType: string;
        switch (workloadType) {
            case WorkloadType.Pytorch:
                jobType = DistributedJobType.PyTorchJobType;
                break;
            case WorkloadType.Tensor:
                jobType = DistributedJobType.TensorJobType;
                break;
            case WorkloadType.Mpi:
                jobType = DistributedJobType.MpiJobType;
                break;
            case WorkloadType.Xgboost:
                jobType = DistributedJobType.XGBoostJobType;
                break;
            default:
                throw new Error('Invalid workload type');
        }
        const distributedWorkload = new DistributedWorkload(image, namespace, jobName);
        distributedWorkload.workers = workers;
        distributedWorkload.jobType = jobType;
        return distributedWorkload;
    }

    async createWorkloadAndWaitUntilCreated(
        workloadType: number,
        namespace: string,
        jobName: string,
        jobImage?: string,
        workloadFields?: Record<string, string>,
        username?: string
    ): Promise<BaseWorkload> {
        let workload: BaseWorkload;
        const image = jobImage || workloadTypeToImage[workloadType];
        const isDistributed: boolean = distributedWorkloadTypes.includes(workloadType);
        const crd: CustomResourceDefinition = isDistributed
            ? crds.runaiDistributedWorkloadCrd
            : crds.runaiJobCrd;
        if (workloadType === WorkloadType.Train) {
            workload = new TrainingWorkload(image, namespace, jobName);
        } else if (workloadType === WorkloadType.Interactive) {
            workload = new InteractiveWorkload(image, namespace, jobName);
        } else if (isDistributed) {
            workload = this.getDistributedWorkload(workloadType, namespace, jobName, image);
        } else {
            throw new Error('Invalid workload type');
        }
        for (const key in workloadFields) {
            workload[key] = workloadFields[key];
        }
        workload.username = username;
        await kubeCli.createWorkload(workload);
        await waitUtils.waitForCrd(crd, jobName, namespace);
        return workload;
    }

    /**
     * Returns a TestEnvironment.
     * @summary create new project and new job & wait for crd.
     * @param workloadType what WorkloadType should be created.
     * @param jobImage image for the job
     * @param workloadFields dictionary that contain workload fields(as keys) and values.
     * */
    async createNamespaceProjectAndWorkload(
        workloadType: number,
        jobImage?: string,
        workloadFields?: Record<string, string>
    ): Promise<TestEnvironment> {
        const projectName = objectGenerator.generateNameByPrefix('project');
        const namespace = Utils.namespaceFromProjectName(projectName);
        const project: Project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: numberOfGpus,
            maxAllowedGpus: numberOfGpus
        });
        const jobName = objectGenerator.generateNameByPrefix('job');
        const workload = await this.createWorkloadAndWaitUntilCreated(
            workloadType,
            namespace,
            jobName,
            jobImage,
            workloadFields
        );
        return new TestEnvironment(projectName, project, namespace, workload, jobName);
    }

    async deleteEnv(
        workload: BaseWorkload,
        projectId: number,
        projectName: string,
        namespace: string
    ): Promise<void> {
        try {
            await operations.deleteWorkloads([workload]);
            await operations.deleteProjectAndNamespace(projectId);
        } catch {
            logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
        }
    }

    async deleteNamespaceProjectAndWorkload(testEnvironment: TestEnvironment): Promise<void> {
        await this.deleteEnv(
            testEnvironment.workload,
            testEnvironment.project.id,
            testEnvironment.projectName,
            testEnvironment.namespace
        );
    }

    async delete(testJob: BaseWorkload): Promise<void> {
        try {
            await kubeCli.deleteCrd(testJob.getCrd(), testJob.name, testJob.testingSpecificFields.namespace);
            logger.info(`Deleted workload for ${testJob.name}`);
        } catch (e) {
            logger.debug(`Failed to delete workload ${testJob.name}`);
            throw e;
        }
    }

    async deleteTraining(name: string, namespace: string): Promise<void> {
        try {
            await kubeCli.deleteCrd(crds.runaiTrainingWorkloadCrd, name, namespace);
            logger.info(`Deleted workload for ${name}`);
        } catch (e) {
            logger.debug(`Failed to delete workload ${name}`);
            throw e;
        }
    }

    async submit(testJob: BaseWorkload): Promise<void> {
        logger.info(`Submitting workload ${testJob.testingSpecificFields.namespace}/${testJob.name}`);
        await kubeCli.createCrd(testJob.getCrd(), testJob.getBody(), testJob.testingSpecificFields.namespace);
    }

    async waitForTrainingWorkloadCrd(
        workloadName: string,
        workloadNamespace: string
    ): Promise<TrainingWorkload> {
        logger.info(`Getting training workload ${workloadNamespace}/${workloadName}`);
        return (await waitUtils.waitForCrd(
            crds.runaiTrainingWorkloadCrd,
            workloadName,
            workloadNamespace
        )) as TrainingWorkload;
    }

    async waitForInteractiveWorkloadCrd(
        workloadName: string,
        workloadNamespace: string
    ): Promise<InteractiveWorkload> {
        logger.info(`Getting interactive workload ${workloadNamespace}/${workloadName}`);
        return (await waitUtils.waitForCrd(
            crds.runaiInteractiveWorkloadCrd,
            workloadName,
            workloadNamespace
        )) as InteractiveWorkload;
    }

    async waitForInferenceWorkloadCrd(
        workloadName: string,
        workloadNamespace: string
    ): Promise<InferenceWorkload> {
        logger.info(`Getting inference workload ${workloadNamespace}/${workloadName}`);
        return (await waitUtils.waitForCrd(
            crds.runaiInferenceWorkloadCrd,
            workloadName,
            workloadNamespace
        )) as InferenceWorkload;
    }

    async isWorkloadDeleted(workload: BaseWorkload, timeout: number): Promise<void> {
        const getWorkload = async (): Promise<boolean> => {
            try {
                await kubeCli.getCrd(
                    workload.getCrd(),
                    workload.name,
                    workload.testingSpecificFields.namespace
                );
            } catch (error: unknown) {
                const err = (error as Error).message;
                return err.includes('NotFound');
            }
            return false;
        };

        await waitUtils.waitForCondition(
            getWorkload,
            `wait for pods to be terminated for namespace ${workload.testingSpecificFields.namespace} for project cleanup.`,
            true,
            timeout
        );
    }

    async clearProjectWorkloadsAndPods(projectName: string, timeout: number): Promise<void> {
        const namespace = Utils.namespaceFromProjectName(projectName);

        await kubeCli.CustomObjectsApi.deleteCollectionNamespacedCustomObject({
            group: crds.runaiTrainingWorkloadCrd.group,
            version: crds.runaiTrainingWorkloadCrd.version,
            namespace,
            plural: crds.runaiTrainingWorkloadCrd.plural
        });
        await kubeCli.CustomObjectsApi.deleteCollectionNamespacedCustomObject({
            group: crds.runaiInteractiveWorkloadCrd.group,
            version: crds.runaiInteractiveWorkloadCrd.version,
            namespace,
            plural: crds.runaiInteractiveWorkloadCrd.plural
        });
        await kubeCli.CustomObjectsApi.deleteCollectionNamespacedCustomObject({
            group: crds.runaiInferenceWorkloadCrd.group,
            version: crds.runaiInferenceWorkloadCrd.version,
            namespace,
            plural: crds.runaiInferenceWorkloadCrd.plural
        });
        supportedFrom(SupportedVersions.V2_23)(async function () {
            await kubeCli.CustomObjectsApi.deleteCollectionNamespacedCustomObject({
                group: crds.runaiDistributedInferenceWorkloadCrd.group,
                version: crds.runaiDistributedInferenceWorkloadCrd.version,
                namespace,
                plural: crds.runaiDistributedInferenceWorkloadCrd.plural
            });
        });
        supportedFrom(SupportedVersions.V2_23)(async function () {
            try {
                await kubeCli.CustomObjectsApi.deleteCollectionNamespacedCustomObject({
                    group: crds.nimServiceCrd.group,
                    version: crds.nimServiceCrd.version,
                    namespace,
                    plural: crds.nimServiceCrd.plural
                });
            } catch {
                // Ignore any exceptions or failures gracefully, as NIMService CRD is optional
            }
        });
        await kubeCli.CoreV1Api.deleteCollectionNamespacedPod({
            namespace
        });

        await waitUtils.waitForPodsTerminationInNamespace(namespace, timeout);
    }

    async getNodeNameByJob(namespace: string, testJob: BaseWorkload): Promise<string> {
        const pods = await kubeCli.listPodsOfJob(namespace, testJob.name);
        if (pods.length === 0) {
            return '';
        }
        return pods[0].spec.nodeName;
    }

    async createInferenceWorkload(
        namespaceName: string,
        workloadName: string,
        options?: InferenceWorkloadCreationOptions
    ): Promise<InferenceWorkload> {
        const workload = new InferenceWorkload(namespaceName, workloadName, options);
        await kubeCli.createWorkload(workload);
        logger.debug(`Created inference workload ${workloadName} in namespace ${namespaceName}`);
        return workload;
    }

    isConditionTrue(workload: object, conditionType: WorkloadConditionType): boolean {
        if (workload['status'] === undefined || workload['status']['conditions'] === undefined) {
            return false;
        }
        for (const c of workload['status']['conditions']) {
            if (c['type'] === conditionType) {
                return c['status'] === 'True';
            }
        }
        return false;
    }

    getCondition(workload: object, conditionType: WorkloadConditionType): WorkloadCondition {
        if (workload['status'] === undefined || workload['status']['conditions'] === undefined) {
            return null;
        }
        for (const c of workload['status']['conditions']) {
            if (c['type'] === conditionType) {
                return WorkloadCondition.fromRawCondition(c);
            }
        }
        return null;
    }

    isConditionFalseOrUnset(workload: object, conditionType: WorkloadConditionType): boolean {
        if (workload['status'] === undefined) return false;
        for (const condition of workload['status']['conditions']) {
            if (condition['type'] === conditionType) {
                return condition['status'] === 'False';
            }
        }
        return true;
    }

    async createWorkload(
        wlType: cliV2WorkloadType,
        image: string,
        command: string,
        project: Project,
        expectedJobStatus?: WorkloadStatus,
        numGpus = '1',
        workloadName?: string,
        nodePools?: string[]
    ): Promise<string> {
        const jobName = workloadName ? workloadName : objectGenerator.generateNameByPrefix('cli-job');
        const submitFlag = {
            [SubmitFlagsMerged.image]: image,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.gpuRequest]: numGpus
        } as Partial<typeof SubmitFlagsMerged>;
        if (nodePools) {
            submitFlag[SubmitFlagsMerged.nodePools] = nodePools;
        }
        if (command) {
            submitFlag[SubmitFlagsMerged.command] = command;
        }
        runaiCliV2.submit(submitFlag, jobName, wlType);
        if (expectedJobStatus) {
            const hasOnlyWorkloads = versionUtils.isVersionSupported(
                flowRootState.controlPlaneVersion,
                SupportedVersions.V2_23
            );
            if (hasOnlyWorkloads) {
                await waitUtils.waitForBackendWorkloadPhase(
                    jobName,
                    project.id,
                    EWorkloadStatusToPhase[expectedJobStatus]
                );
            } else {
                await waitUtils.waitForBackendJobStatus(jobName, expectedJobStatus);
            }
        }
        return jobName;
    }

    async waitForWorkloadPhase(
        jobName: string,
        namespace: string,
        crd: CustomResourceDefinition,
        expectedPhase: string,
        timeout = 90
    ): Promise<void> {
        const getWorkloadPhase = async (): Promise<boolean> => {
            try {
                const describePayload = await kubeCli.getWorkload(jobName, namespace, crd);
                logger.info('status: ' + describePayload['status']['phase']);
                return describePayload['status']['phase'] === expectedPhase;
            } catch (error: unknown) {
                throw (error as Error).message;
            }
        };

        await waitUtils.waitForCondition(
            getWorkloadPhase,
            `wait for workload phase to be equal to ${expectedPhase}`,
            true,
            timeout
        );
    }

    async waitForWorkloadBackendPhase(
        workloadName: string,
        projectName: string,
        expectedPhase: Phase,
        timeout = 90
    ): Promise<void> {
        const getWorkloadPhase = async (): Promise<boolean> => {
            try {
                const workload = await workloadsUtilsV2.getWorkloadByNameAndProjectName(
                    workloadName,
                    projectName
                );
                return workload && workload.phase === expectedPhase;
            } catch (error: unknown) {
                throw (error as Error).message;
            }
        };

        await waitUtils.waitForCondition(
            getWorkloadPhase,
            `wait for workload phase to be equal to ${expectedPhase}`,
            true,
            timeout
        );
    }

    async waitForWorkloadClusterResource(
        workloadName: string,
        namespace: string,
        workloadType: AnyWorkloadSubmissionType,
        legacyCrd: CustomResourceDefinition
    ): Promise<void> {
        const route = await getWorkloadSubmissionRoute(workloadType);
        if (route === WorkloadSubmissionRoute.WorkloadRunner) {
            await this.getWorkloadRunnerByWorkloadName(workloadName, namespace);
        } else {
            await waitUtils.waitForCrd(legacyCrd, workloadName, namespace);
        }
    }

    async getWorkloadRunnerByWorkloadName(
        workloadName: string,
        namespace: string,
        timeout = 60
    ): Promise<WorkloadRunnerResource> {
        let workloadRunner: WorkloadRunnerResource;
        await waitUtils.waitForCondition(
            async () => {
                const runners = (await kubeCli.getCrds(
                    crds.workloadRunnerCrd,
                    namespace
                )) as WorkloadRunnerResource[];
                workloadRunner = findWorkloadRunnerByWorkloadName(runners, workloadName);
                return workloadRunner !== undefined;
            },
            `Wait for WorkloadRunner matching workload ${workloadName} in namespace ${namespace}`,
            true,
            timeout,
            2
        );
        return workloadRunner;
    }

    async waitForWorkloadRunnerPhase(
        workloadName: string,
        namespace: string,
        expectedPhase: Phase,
        timeout = 60
    ): Promise<WorkloadRunnerResource> {
        let workloadRunner: WorkloadRunnerResource;
        await waitUtils.waitForCondition(
            async () => {
                const runners = (await kubeCli.getCrds(
                    crds.workloadRunnerCrd,
                    namespace
                )) as WorkloadRunnerResource[];
                workloadRunner = findWorkloadRunnerByWorkloadName(runners, workloadName);
                return workloadRunner?.status?.phase === expectedPhase;
            },
            `Wait for WorkloadRunner ${workloadName} phase to be ${expectedPhase}`,
            true,
            timeout,
            2
        );
        return workloadRunner;
    }

    async waitForDistributedSubmissionRoot(
        workloadName: string,
        namespace: string,
        timeout = 60
    ): Promise<SubmissionRoot> {
        const workloadRunner = await this.getWorkloadRunnerByWorkloadName(workloadName, namespace, timeout);
        if (!workloadRunner?.metadata?.name) {
            throw new Error(`No distributed submission root found for workload ${workloadName}`);
        }
        return {
            crd: crds.workloadRunnerCrd,
            name: workloadRunner.metadata.name,
            resource: workloadRunner,
            route: WorkloadSubmissionRoute.WorkloadRunner
        };
    }

    getPriorityClassFromWorkloadRunner(value: unknown): string | undefined {
        if (!value || typeof value !== 'object') return undefined;

        const priorityFields = ['priorityClassName', 'priorityClass'];
        for (const field of priorityFields) {
            const fieldValue = (value as Record<string, any>)[field];
            if (typeof fieldValue === 'string') return fieldValue;
            if (fieldValue && typeof fieldValue === 'object') {
                if (typeof fieldValue.value === 'string') return fieldValue.value;
                const nestedPriority = this.getPriorityClassFromWorkloadRunner(fieldValue);
                if (nestedPriority) return nestedPriority;
            }
        }

        for (const nestedValue of Object.values(value)) {
            const nestedPriority = this.getPriorityClassFromWorkloadRunner(nestedValue);
            if (nestedPriority) return nestedPriority;
        }
        return undefined;
    }

    async getEnvVars(jobName: string, namespace: string, crd: CustomResourceDefinition): Promise<void> {
        try {
            const payload = await kubeCli.getWorkload(jobName, namespace, crd);
            return payload?.['spec']?.['environment']?.['items'];
        } catch (error: unknown) {
            throw (error as Error).message;
        }
    }

    async getCpusCount(jobName: string, namespace: string, crd: CustomResourceDefinition): Promise<number> {
        try {
            const describePayload = await kubeCli.getWorkload(jobName, namespace, crd);
            if (describePayload?.['spec']?.['cpu']?.['value']) {
                return Number(describePayload['spec']['cpu']['value']);
            }
            return undefined;
        } catch (error: unknown) {
            throw (error as Error).message;
        }
    }

    async getCpuLimitCount(
        jobName: string,
        namespace: string,
        crd: CustomResourceDefinition
    ): Promise<number> {
        try {
            const describePayload = await kubeCli.getWorkload(jobName, namespace, crd);
            return Number(describePayload['spec']['cpuLimit']['value']);
        } catch (error: unknown) {
            throw (error as Error).message;
        }
    }

    async getPvc(jobName: string, namespace: string, crd: CustomResourceDefinition): Promise<number> {
        try {
            const describePayload = await kubeCli.getWorkload(jobName, namespace, crd);
            return Number(describePayload['spec']['cpuLimit']['value']);
        } catch (error: unknown) {
            throw (error as Error).message;
        }
    }

    async createInferencePolicy(
        policyRequest: InferencePolicyChangeRequestV2,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<InferencePolicyV2> {
        try {
            policyApi.policyApi.accessToken = accessToken;
            const res = await policyApi.policyApi.overwriteInferencePolicyV2(false, false, policyRequest);
            logger.debug('Policy created successfully!');
            return res.body;
        } catch (error) {
            logger.error(`Error creating inference policy: ${error}`);
            throw error;
        }
    }

    async deletePolicy(
        projectId: number | string,
        scope: 'tenant' | 'cluster' | 'department' | 'project' = 'project',
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            policyApi.policyApi.accessToken = accessToken;
            await policyApi.policyApi.deleteInferencePolicy(scope, undefined, projectId.toString());
            logger.debug('Policy deleted successfully!');
        } catch (error) {
            logger.error(`Error deleting inference policy: ${error}`);
            throw error;
        }
    }

    async getWorkloadTypes(
        filterBy?: Array<string>,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<WorkloadTypeConfig>> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.listWorkloadTypes(false, filterBy);
            logger.debug('Workload types fetched successfully!');
            return response.body.types;
        } catch (error) {
            logger.error(`Error fetching workload types: ${error}`);
            throw error;
        }
    }

    async getWorkloadTypeConfigBy(
        workloadTypeName: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadTypeConfig> {
        const workloadTypes = await this.getWorkloadTypes([`name==${workloadTypeName}`], accessToken);
        return workloadTypes[0];
    }

    async getWorkloadTypeById(
        workloadTypeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadTypeConfig> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.getWorkloadType(workloadTypeId);
            logger.debug(`Workload type details for ${workloadTypeId} fetched successfully!`);
            return response.body;
        } catch (error) {
            logger.error(`Error fetching workload type details: ${error}`);
            throw error;
        }
    }

    async updateWorkloadTypeById(
        workloadTypeId: string,
        updateFields: WorkloadTypeUpdateFields,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadTypeConfig }> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.updateWorkloadType(
                workloadTypeId,
                updateFields
            );
            logger.debug(`Workload type ${workloadTypeId} updated successfully!`);
            return response;
        } catch (error) {
            logger.error(`Error updating workload type ${workloadTypeId}: ${error}`);
            throw error;
        }
    }

    async getWorkloadPriorities(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadPriority[]> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.getWorkloadPriorities();
            logger.debug('Workload priorities fetched successfully!');
            return response.body.priorities;
        } catch (error) {
            logger.error(`Error fetching workload priorities: ${error}`);
            throw error;
        }
    }

    async getWorkloadClusterPriorities(
        clusterId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<ClusterPriority[]> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response =
                await workloadsService.workloadPropertiesApi.getWorkloadClusterPriorities(clusterId);
            logger.debug(`Workload cluster priorities for cluster ${clusterId} fetched successfully!`);
            return response.body.priorities;
        } catch (error) {
            logger.error(`Error fetching workload cluster priorities: ${error}`);
            throw error;
        }
    }

    async getWorkloadCategories(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<any[]> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.listCategories();
            logger.debug('Workload categories fetched successfully!');
            return response.body.categories;
        } catch (error) {
            logger.error(`Error fetching workload categories: ${error}`);
            throw error;
        }
    }

    async createWorkloadType(
        name: string,
        categoryId: string,
        priorityId: string,
        group: string = 'batch',
        version: string = 'v1',
        customResourceInterfaces?: any[],
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<any> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;

            // Use custom resourceInterfaces if provided, otherwise use default
            const defaultResourceInterfaces = [
                {
                    spec: {
                        structureDefinition: {
                            rootComponent: {
                                name: name, // component name is required
                                kind: {
                                    group,
                                    version,
                                    kind: name // kind should match name exactly
                                },
                                statusDefinition: {
                                    phasePath: 'status.phase',
                                    conditionsDefinition: {
                                        path: 'status.conditions'
                                    }
                                }
                            }
                        },
                        statusDetails: {
                            phasePath: 'status.phase',
                            conditionsDefinition: {
                                path: 'status.conditions'
                            },
                            statusMappings: {
                                running: {
                                    byConditions: [
                                        {
                                            type: 'Ready',
                                            status: 'True'
                                        }
                                    ]
                                }
                            }
                        }
                    }
                }
            ];

            const createFields: any = {
                categoryId,
                priorityId,
                name,
                group,
                resourceInterfaces: customResourceInterfaces || defaultResourceInterfaces
            };

            logger.debug(`Creating workload type with fields: ${JSON.stringify(createFields, null, 2)}`);
            const response = await workloadsService.workloadPropertiesApi.createWorkloadType(createFields);
            logger.debug(`Workload type '${name}' created successfully!`);
            return response;
        } catch (error) {
            logger.error(`Error creating workload type '${name}': ${error}`);
            throw error;
        }
    }

    async deleteWorkloadType(
        workloadTypeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<any> {
        try {
            workloadsService.workloadPropertiesApi.accessToken = accessToken;
            const response = await workloadsService.workloadPropertiesApi.deleteWorkloadType(workloadTypeId);
            logger.debug(`Workload type '${workloadTypeId}' deleted successfully!`);
            return response;
        } catch (error) {
            logger.error(`Error deleting workload type '${workloadTypeId}': ${error}`);
            throw error;
        }
    }

    /**
     * Updates a workload type configuration by name
     *
     * @param workloadTypeName - The name of the workload type (e.g., 'PyTorch', 'DynamoGraphDeployment')
     * @param priorityName - The name of the priority (e.g., 'high', 'low', 'very-high')
     * @param categoryName - The name of the category (e.g., 'Train', 'Deploy')
     * @param preemptibility - The preemptibility
     * @param accessToken - Optional access token (defaults to system admin token)
     * @returns Promise<void>
     * @throws Error if workload type, priority, or category not found
     */
    async updateWorkloadTypeConfiguration(
        workloadTypeName: string,
        priorityName: string,
        categoryName: string,
        preemptibility: WorkloadTypePreemptibility,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadTypeConfig> {
        try {
            logger.info(`Looking up workload type: ${workloadTypeName}`);
            const workloadTypes = await this.getWorkloadTypes(undefined, accessToken);
            const workloadType = workloadTypes.find((wt) => wt.name === workloadTypeName);

            if (!workloadType) {
                throw new Error(
                    `Workload type '${workloadTypeName}' not found. Available types: ${workloadTypes.map((wt) => wt.name).join(', ')}`
                );
            }

            logger.info(`Looking up priority: ${priorityName}`);
            const priorities = await this.getWorkloadPriorities(accessToken);
            const priority = priorities.find((p) => p.name === priorityName);

            if (!priority) {
                throw new Error(
                    `Priority '${priorityName}' not found. Available priorities: ${priorities.map((p) => p.name).join(', ')}`
                );
            }

            logger.info(`Looking up category: ${categoryName}`);
            const categories = await this.getWorkloadCategories(accessToken);
            const category = categories.find((c) => c.name === categoryName);

            if (!category) {
                throw new Error(
                    `Category '${categoryName}' not found. Available categories: ${categories.map((c) => c.name).join(', ')}`
                );
            }

            const updateFields: WorkloadTypeUpdateFields = {
                priorityId: priority.id,
                categoryId: category.id,
                preemptibility: preemptibility
            };

            const response = await this.updateWorkloadTypeById(workloadType.id, updateFields, accessToken);
            return response.body;
        } catch (error) {
            logger.error(`Failed to update workload type configuration: ${error.message}`);
            throw error;
        }
    }

    createPyTorchJobManifest(
        name: string,
        namespace: string,
        image: string,
        command: string,
        workerReplicas = 2,
        gpuCount = 2
    ): any {
        return {
            apiVersion: 'kubeflow.org/v1',
            kind: 'PyTorchJob',
            metadata: { name, namespace },
            spec: {
                pytorchReplicaSpecs: {
                    Master: {
                        replicas: 1,
                        restartPolicy: 'OnFailure',
                        template: {
                            spec: {
                                containers: [
                                    {
                                        name: 'pytorch',
                                        image,
                                        command: ['sh', '-c', command],
                                        resources: {
                                            limits: {
                                                'nvidia.com/gpu': gpuCount
                                            },
                                            requests: {
                                                'nvidia.com/gpu': gpuCount
                                            }
                                        }
                                    }
                                ]
                            }
                        }
                    },
                    Worker: {
                        replicas: workerReplicas,
                        restartPolicy: 'OnFailure',
                        template: {
                            spec: {
                                containers: [
                                    {
                                        name: 'pytorch',
                                        image,
                                        command: ['sh', '-c', command],
                                        resources: {
                                            limits: {
                                                'nvidia.com/gpu': gpuCount
                                            },
                                            requests: {
                                                'nvidia.com/gpu': gpuCount
                                            }
                                        }
                                    }
                                ]
                            }
                        }
                    }
                }
            }
        };
    }
}
