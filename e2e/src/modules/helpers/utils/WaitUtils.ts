/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    CustomResourceDefinition,
    flowRootState,
    logger,
    operations,
    PodStatus,
    runaiHealthChecker,
    SchedulingUtils,
    Utils,
    UtilsError,
    waitUtils,
    WorkloadConditionType,
    workloadUtils,
    versionUtils,
    workloadsUtilsV2,
    InferenceWorkload,
    clusterUtils,
    nodeUtils,
    metricUtils,
    telemetryUtils,
    aiApplicationUtils
} from '..';
import { kubeCli, runaiCliV2 } from '../../cli';
import { cliV2WorkloadType, execFlags } from '../../cli/tools/RunaiCliV2.constants';
import { assetsApi, backendClusterApi, workloadsService } from '../../api';

import {
    KubernetesObject,
    V1Deployment,
    V1Ingress,
    V1Job,
    V1PersistentVolumeClaim,
    V1PersistentVolumeClaimList,
    V1Pod,
    V1Service,
    V1ServiceAccount
} from '@kubernetes/client-node';
import { crds } from '../models/customResourceDefinition';
import { WorkloadStatus } from '../models/baseWorkload';
import { DisplayedJob } from '../../api/apis/backendClusterApi';
import { Workspace, WorkspaceListResponse, WorkloadTemplate, Training } from '../../api/apis/assetsApi';
import { expect } from 'chai';
import { BaseV2Page } from '../../ui/runaiV2/pages/BaseV2Page';
import { isOCP } from './EnvironmentUtils';
import { Locator, Page } from 'playwright';
import { Workload } from '../../api/apis/workloadsApi/generated/model/workload';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { logClusterDiagnostics } from './testing/clusterDiagnosticsUtils';
import { Revision } from '../../api/apis/workloadsApi/generated/model/revision';
import { Pod as RevisionPod } from '../../api/apis/workloadsApi/generated/model/pod';
import { HistoryRecordType } from '../../api/apis/workloadsApi/generated/model/historyRecordType';
import { KnativeRevision, KnativeService } from '../models/knativeService';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { Event } from '../../api/apis/workloadsApi/generated/model/event';
import { WorkloadDetailed } from '../../api/apis/workloadsApi/generated/model/workloadDetailed';
import {
    GroupedMeasurementResponseValuesInner,
    MeasurementResponse,
    NodeMetricType,
    NodeTelemetryType
} from '../../api/apis/clusterApi';
import { WorkloadMetricType } from '../../api/apis/workloadsApi/generated/api';

export enum InferencePhase {
    Creating = <any>'Creating',
    Initializing = <any>'Initializing',
    Pending = <any>'Pending',
    Ready = <any>'Ready',
    Progressing = <any>'Progressing',
    Deleting = <any>'Deleting',
    Failed = <any>'Failed',
    Unknown = <any>'Unknown'
}

export type BatchJobState = 'running' | 'suspended' | 'completed' | 'failed';

export class WaitUtils {
    waitForPodTimeout = 180;

    async waitForCondition(
        f: () => Promise<boolean> | boolean,
        description: string,
        desiredResult = true,
        time = 120, // 2 min,
        delay = 1 // 1 second
    ): Promise<void> {
        let timeoutExpired = false;
        let timeoutError: string;
        const timeout = setTimeout(() => {
            timeoutExpired = true;
            timeoutError = `Timed out waiting for condition "${description}" for ${time} seconds`;
        }, time * 1000);

        while (!timeoutExpired) {
            if ((await f()) === desiredResult) {
                clearTimeout(timeout);
                return;
            }
            await Utils.delay(delay * 1000); // Prevent busy wait
        }
        //We exit the while only if the timeout expires
        throw new UtilsError(new Error(timeoutError));
    }

    async waitForWorkloadOwnershipProtectionReported(isEnable: boolean): Promise<void> {
        await this.waitForCondition(
            async () =>
                (await clusterUtils.getClusterReportedConfig())?.workloadOwnershipProtection === isEnable,
            `Wait for workloadOwnershipProtection to be ${isEnable ? 'enabled' : 'disabled'} and reported`,
            true,
            120
        );
    }

    // BLOCKS FOR THE WHOLE TIME - USE CAREFULLY
    async consistentlyCondition(
        f: () => Promise<boolean>,
        description: string,
        desiredResult = true,
        time = 30, // 0.5 min,
        delay = 10 // 10 seconds
    ): Promise<void> {
        time = time + 1;
        let timeoutExpired = false;
        const timeout = setTimeout(() => {
            timeoutExpired = true;
        }, time * 1000);

        while (!timeoutExpired) {
            if ((await f()) !== desiredResult) {
                clearTimeout(timeout);
                throw new UtilsError(
                    new Error(
                        `Condition "${description}" not met while consistently checking for ${time} seconds`
                    )
                );
            }
            await Utils.delay(delay * 1000); // Prevent busy wait
        }
    }

    async waitForObject(
        f: () => Promise<object>,
        description: string,
        time = 120, // 2 min
        sleepInterval = 1, // 1 second
        waitUndefined = false // when true - waits to object deleted etc.
    ): Promise<object> {
        let timeoutExpired = false;
        let timeoutError: string;
        const timeout = setTimeout(() => {
            timeoutExpired = true;
            timeoutError = `Timed out waiting for condition "${description}" for ${time} seconds`;
        }, time * 1000);

        while (!timeoutExpired) {
            const createdObjects = await f();
            if (createdObjects || waitUndefined) {
                clearTimeout(timeout);
                return createdObjects;
            }
            await Utils.delay(sleepInterval * 1000); // Prevent busy wait
        }
        //We exit the while only if the timeout expires
        throw new UtilsError(new Error(timeoutError));
    }

    async waitForCrd(
        crd: CustomResourceDefinition,
        name: string,
        namespace?: string,
        timeout = 40,
        desiredResult = true,
        sleepInterval = 1
    ): Promise<object> {
        let res: object;

        const isCrdCreated = async (): Promise<boolean> => {
            try {
                logger.debug(`wait for CRD ${crd.plural}`);
                res = await kubeCli.getCrd(crd, name, namespace);
                return true;
            } catch (e) {
                logger.debug(
                    `CRD ${crd.plural}.${crd.group}/${crd.version} name: ${name} not found or not exist yet`
                );
                return false;
            }
        };
        try {
            await this.waitForCondition(
                isCrdCreated,
                `waitForCrd crd ${crd.plural}.${crd.group}/${crd.version} name: ${name} to be ${desiredResult}`,
                desiredResult,
                timeout,
                sleepInterval
            );
        } catch (error) {
            const describeCrd = await kubeCli.getCrds(crd);
            logger.error(
                `waitForCrd crd ${crd.plural}.${crd.group}/${
                    crd.version
                } name: ${name} failed \n ****** Describe Crd ******:\n ${JSON.stringify(
                    describeCrd,
                    null,
                    2
                )}`
            );
            throw new UtilsError(error);
        }

        return res;
    }

    async waitForCrdProperty(
        crd: CustomResourceDefinition,
        name: string,
        namespace: string,
        propertyName: string,
        expectedValue?: string, // if undefined, just check if property exists
        subPropertyName?: string, // for nested properties
        timeout = 40
    ): Promise<void> {
        const doesCrdPropertyExist = async (): Promise<boolean> => {
            try {
                const crdObj = await kubeCli.getCrd(crd, name, namespace);
                if (!crdObj || !crdObj[propertyName]) {
                    return false;
                }
                if (subPropertyName && !crdObj[propertyName][subPropertyName]) {
                    return false;
                }
                if (!expectedValue) {
                    return true;
                }
                if (subPropertyName) {
                    return crdObj[propertyName][subPropertyName] === expectedValue;
                } else {
                    return crdObj[propertyName] === expectedValue;
                }
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            doesCrdPropertyExist,
            `waitForCrdProperty ${crd.plural}.${crd.group}/${crd.version} name: ${name} to have ${propertyName}=${expectedValue}`,
            true,
            timeout
        );
    }

    async waitAndGetCrd(
        crd: CustomResourceDefinition,
        name: string,
        namespace?: string,
        timeout = 40
    ): Promise<KubernetesObject> {
        await this.waitForCrd(crd, name, namespace, timeout);
        return await kubeCli.getCrd(crd, name, namespace);
    }

    async waitForDeployment(
        deploymentName: string,
        namespace: string,
        desiredResult: boolean
    ): Promise<void> {
        const doesDeploymentExist = async (): Promise<boolean> => {
            try {
                await kubeCli.getDeployment(deploymentName, namespace);
                logger.debug(`Deployment ${deploymentName} found`);
                return true;
            } catch (e) {
                logger.debug(`Deployment ${deploymentName} not found`);
                return false;
            }
        };

        await this.waitForCondition(
            doesDeploymentExist,
            `waitForDeployment ${namespace}/${deploymentName} to be ${desiredResult}`,
            desiredResult,
            40
        );
    }

    isDeploymentReadyFn(deploymentName: string, namespaceName: string): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            try {
                const deployment: V1Deployment = await operations.getDeployment(
                    namespaceName,
                    deploymentName
                );
                // This contains a bug
                return (
                    (deployment.status.conditions[0]['type'] === 'Progressing' &&
                        deployment.status.conditions[0]['status'] === 'True' &&
                        deployment.status.conditions[0]['reason'] === 'NewReplicaSetAvailable') ||
                    (deployment.status.conditions.length > 1 &&
                        deployment.status.conditions[1]['type'] === 'Progressing' &&
                        deployment.status.conditions[1]['status'] === 'True' &&
                        deployment.status.conditions[1]['reason'] === 'NewReplicaSetAvailable')
                );
            } catch (e) {
                // if the department does not exist
                return false;
            }
            // return deployment.status.availableReplicas === deployment.status.replicas;
        };
    }

    // BLOCKS FOR THE WHOLE TIME - USE CAREFULLY
    async consistentlyDeploymentReady(
        deploymentName: string,
        namespaceName: string,
        timeout = 30, // 0.5 minute
        desiredResult = true
    ): Promise<void> {
        const availableOrNotMsg = desiredResult ? '' : 'not ';
        await this.consistentlyCondition(
            this.isDeploymentReadyFn(deploymentName, namespaceName),
            `consistently checking for deployment ${deploymentName} in namespace ${namespaceName} to be ${availableOrNotMsg}available`,
            true,
            timeout
        );
    }

    async waitForDeploymentToBeAvailable(
        deploymentName: string,
        namespaceName: string,
        timeout = 120,
        desiredResult = true
    ): Promise<void> {
        const availableOrNotMsg = desiredResult ? '' : 'not ';
        await this.waitForCondition(
            this.isDeploymentReadyFn(deploymentName, namespaceName),
            `wait for deployment ${deploymentName} in namespace ${namespaceName} to be ${availableOrNotMsg}available`,
            desiredResult,
            timeout
        );
    }

    async waitForJobToBeFinished(
        jobName: string,
        namespaceName: string,
        workloadType: string,
        timeout = 120
    ): Promise<void> {
        await this.waitForCondition(
            this.isJobFinished(jobName, namespaceName, workloadType),
            `wait for job ${jobName} in namespace ${namespaceName} to be finished`,
            true,
            timeout
        );
    }

    async waitForJobToBeSuspend(
        jobName: string,
        namespaceName: string,
        workloadType: string,
        timeout = 120,
        shouldCheckPodsStatus = true
    ): Promise<void> {
        await this.waitForCondition(
            this.isJobSuspend(jobName, namespaceName, workloadType, shouldCheckPodsStatus),
            `wait for job ${jobName} in namespace ${namespaceName} to be suspend`,
            true,
            timeout
        );
    }

    isJobFinished(jobName: string, namespaceName: string, workloadType: string): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            const workload = await operations.getWorkload(jobName, namespaceName, workloadType);
            if (workload === null || workload['status'] === undefined) {
                return false;
            }
            return (
                workload['status']['completionTime'] !== undefined &&
                (workload['status']['runningPods'] === 0 || workload['status']['runningPods'] === undefined)
            );
        };
    }

    isJobSuspend(
        jobName: string,
        namespaceName: string,
        workloadType: string,
        shouldCheckPodsStatus = true
    ): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            const workload = await operations.getWorkload(jobName, namespaceName, workloadType);
            if (workload === null) {
                return false;
            }
            const IsSuspended = workloadUtils.isConditionTrue(workload, WorkloadConditionType.Suspended);
            if (shouldCheckPodsStatus) {
                return (
                    IsSuspended &&
                    (workload['status']['runningPods'] === 0 ||
                        workload['status']['runningPods'] === undefined)
                );
            } else {
                return IsSuspended;
            }
        };
    }

    isJobRunning(jobName: string, namespaceName: string, workloadType: string): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            const workload = await operations.getWorkload(jobName, namespaceName, workloadType);
            if (!workload) {
                return false;
            }
            if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
                return workloadUtils.isConditionTrue(workload, WorkloadConditionType.Initialized);
            } else {
                return workloadUtils.isConditionTrue(workload, WorkloadConditionType.Running);
            }
        };
    }

    isJobPending(jobName: string, namespaceName: string, workloadType: string): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            const workload = await operations.getWorkload(jobName, namespaceName, workloadType);
            if (workload === null) {
                return false;
            }
            return workloadUtils.isConditionTrue(workload, WorkloadConditionType.Pending);
        };
    }

    async waitForJobToBePending(
        jobName: string,
        namespaceName: string,
        workloadType: string,
        timeout = 120
    ): Promise<void> {
        await this.waitForCondition(
            this.isJobPending(jobName, namespaceName, workloadType),
            `wait for job ${jobName} in namespace ${namespaceName} to be pending`,
            true,
            timeout
        );
    }

    private static readonly batchJobStateChecks: Record<BatchJobState, (job: V1Job) => boolean> = {
        running: (job) => job.spec?.suspend !== true && (job.status?.ready ?? job.status?.active ?? 0) >= 1,
        suspended: (job) => workloadUtils.isConditionTrue(job, WorkloadConditionType.Suspended),
        // batch/v1 Job exposes a 'Complete' condition (not the RunAI 'Succeeded' type).
        completed: (job) =>
            (job.status?.conditions ?? []).some((c) => c.type === 'Complete' && c.status === 'True') ||
            (job.status?.succeeded ?? 0) >= (job.spec?.completions ?? 1),
        failed: (job) =>
            (job.status?.conditions ?? []).some((c) => c.type === 'Failed' && c.status === 'True')
    };

    isBatchJobInState(state: BatchJobState, name: string, namespaceName: string): () => Promise<boolean> {
        return async (): Promise<boolean> => {
            try {
                const job = await kubeCli.getBatchJob(name, namespaceName);
                return WaitUtils.batchJobStateChecks[state](job);
            } catch (err) {
                // The Job may not exist yet (e.g. right after submission) - keep polling.
                logger.debug(`isBatchJobInState(${state}): could not read job ${name}: ${err}`);
                return false;
            }
        };
    }

    async waitForBatchJobState(
        state: BatchJobState,
        name: string,
        namespaceName: string,
        timeout = 120
    ): Promise<void> {
        await this.waitForCondition(
            this.isBatchJobInState(state, name, namespaceName),
            `wait for batch job ${name} in namespace ${namespaceName} to be ${state}`,
            true,
            timeout
        );
    }

    async waitForBatchJobToBeSuspended(name: string, namespaceName: string, timeout = 120): Promise<void> {
        await this.waitForBatchJobState('suspended', name, namespaceName, timeout);
    }

    async waitForBatchJobToBeRunning(name: string, namespaceName: string, timeout = 120): Promise<void> {
        await this.waitForBatchJobState('running', name, namespaceName, timeout);
    }

    async waitForBatchJobToBeCompleted(name: string, namespaceName: string, timeout = 120): Promise<void> {
        await this.waitForBatchJobState('completed', name, namespaceName, timeout);
    }

    async waitForBatchJobToBeFailed(name: string, namespaceName: string, timeout = 120): Promise<void> {
        await this.waitForBatchJobState('failed', name, namespaceName, timeout);
    }

    // Given a workload phase, verifies the corresponding stable state on the underlying batch/v1 Job
    // that the generic WorkloadRunner CRD wraps (training/workspace). Transient phases
    // (Suspending/Resuming) are not persisted on the Job and are intentionally not asserted here -
    // they remain the caller's responsibility (UI/CLI/API), if checked at all.
    async waitForWrappedJobState(
        workloadName: string,
        namespaceName: string,
        workloadStatus: WorkloadStatus,
        timeout: number
    ): Promise<void> {
        switch (workloadStatus) {
            case WorkloadStatus.RUNNING:
                await this.waitForBatchJobToBeRunning(workloadName, namespaceName, timeout);
                break;
            case WorkloadStatus.STOPPED:
            case WorkloadStatus.SUSPENDED:
                await this.waitForBatchJobToBeSuspended(workloadName, namespaceName, timeout);
                break;
            case WorkloadStatus.COMPLETED:
                await this.waitForBatchJobToBeCompleted(workloadName, namespaceName, timeout);
                break;
            case WorkloadStatus.FAILED:
                await this.waitForBatchJobToBeFailed(workloadName, namespaceName, timeout);
                break;
            default:
                // Transient phases (Suspending/Resuming) are not asserted on the wrapped Job.
                break;
        }
    }

    async waitForWorkloadStatus(
        workload: BaseWorkload,
        expectedStatus: PodStatus,
        timeout = 320
    ): Promise<void> {
        let prevStatusesMessage: string;
        const isWorkloadInExpectedStatus = async (): Promise<boolean> => {
            const pods = await kubeCli.listPodsOfJob(workload.testingSpecificFields.namespace, workload.name);
            if (pods.length === 0) {
                return false;
            }
            const statusesMessage = `workload ${workload.name} pods statuses: ${JSON.stringify(
                pods.map((pod) => pod.status.phase)
            )}`;
            if (statusesMessage !== prevStatusesMessage) {
                // to avoid logging the same phase again and again
                logger.info(statusesMessage);
                prevStatusesMessage = statusesMessage;
            }
            return pods.every((pod) => pod.status.phase === expectedStatus);
        };

        try {
            await this.waitForCondition(
                isWorkloadInExpectedStatus,
                `workload ${workload.name} in namespace ${workload.testingSpecificFields.namespace} is in ${expectedStatus} state`,
                true,
                timeout
            );
        } catch (err) {
            const events = kubeCli.listEventsByNamespace(workload.testingSpecificFields.namespace);
            logger.info(`events in namespace ${workload.testingSpecificFields.namespace}:\n${events}`);
            logClusterDiagnostics(workload.testingSpecificFields.namespace, workload.name);
            throw err;
        }
    }

    async waitForInferenceWorkloadPhase(
        workload: InferenceWorkload,
        expectedPhase: InferencePhase,
        timeout = 320
    ): Promise<void> {
        await this.waitForCondition(
            () => workload.validatePhase(expectedPhase),
            `workload ${workload.name} in namespace ${workload.testingSpecificFields.namespace} is in ${expectedPhase} phase`,
            true,
            timeout
        );
    }

    async waitForWorkloadNotToExist(workload: BaseWorkload): Promise<void> {
        const podsNotExist = async (): Promise<boolean> => {
            try {
                const pods = await kubeCli.listPodsOfJob(
                    workload.testingSpecificFields.namespace,
                    workload.name
                );
                return pods.length === 0;
            } catch (e) {
                logger.debug(
                    `workload ${workload.name} in namespace ${workload.testingSpecificFields.namespace} no longer exist`
                );
                return false;
            }
        };

        await this.waitForCondition(
            podsNotExist,
            `workload ${workload.name} in namespace ${workload.testingSpecificFields.namespace} no longer exist`,
            true
        );
    }

    async waitForPods(
        jobName: string,
        namespace: string,
        instances = 1,
        timeout = this.waitForPodTimeout,
        expectedStatuses = [],
        isPrefixName = false,
        sleepInterval = 1
    ): Promise<V1Pod[]> {
        let podPhases = new Array(0);
        let prevPodsMessage: string;
        let prevPhasesMessage: string;
        const isPodConditionMet = async (): Promise<object> => {
            try {
                const pods = await kubeCli.listPodsOfJob(namespace, jobName, isPrefixName);
                const podsMessage = `pods for job ${jobName} in namespace ${namespace}: ${JSON.stringify(
                    pods.map((pod) => ({
                        name: pod.metadata.name,
                        status: pod.status.phase
                    }))
                )}`;
                if (podsMessage !== prevPodsMessage) {
                    logger.debug(podsMessage);
                    prevPodsMessage = podsMessage;
                }
                podPhases = [];
                pods.forEach((pod) => podPhases.push(pod.status.phase));
                const phasesMessage = `pod phases for job ${jobName} in namespace ${namespace}: ${JSON.stringify(podPhases)}`;
                if (phasesMessage !== prevPhasesMessage) {
                    logger.debug(phasesMessage);
                    prevPhasesMessage = phasesMessage;
                }
                if (pods.length >= instances) {
                    if (expectedStatuses.length) {
                        if (pods.every((pod) => expectedStatuses.includes(pod.status.phase))) {
                            return pods;
                        }
                        return null;
                    }
                    return pods;
                } else {
                    return undefined;
                }
            } catch (e) {
                logger.debug(`job ${jobName} in namespace ${namespace} not found`);
                return undefined;
            }
        };

        try {
            const podList = await this.waitForObject(
                isPodConditionMet,
                `wait for ${instances} pods to be created for job ${jobName} in namespace ${namespace}`,
                timeout,
                sleepInterval
            );
            return podList as V1Pod[];
        } catch (e) {
            logger.debug(
                `Pod phases before reached timeout in waitForPods for pod "${namespace}/${jobName}": ${JSON.stringify(
                    podPhases
                )}`
            );
            logClusterDiagnostics(namespace, jobName);
            e.message = `${e.message} - pod phases were: ${JSON.stringify(podPhases)}`;
            throw e;
        }
    }

    async waitForPodsDeletion(jobName: string, namespace: string, timeout = 70): Promise<boolean> {
        let arePodsDeleted;
        const arePodsDeletedFn = async (): Promise<boolean> => {
            try {
                const pods = await kubeCli.listPodsOfJob(namespace, jobName);
                if (pods.length === 0) {
                    logger.debug(`all pods for job ${jobName} in namespace ${namespace} deleted`);
                    arePodsDeleted = true;
                    return true;
                } else {
                    logger.debug(
                        `waiting for pods for job ${jobName} in namespace ${namespace} to be deleted`
                    );
                    arePodsDeleted = false;
                    return false;
                }
            } catch (e) {
                logger.debug(
                    `job ${jobName} in namespace ${namespace} not found or pods already deleted: ${e}`
                );
                return false;
            }
        };

        await this.waitForCondition(
            arePodsDeletedFn,
            `wait for pods to be deleted for job ${jobName} in namespace ${namespace}`,
            true,
            timeout
        );
        return arePodsDeleted;
    }

    async waitForPodToBeDeleted(podName: string, namespace: string, timeout = 60): Promise<void> {
        const isPodDeleted = async (): Promise<boolean> => {
            try {
                const allPodsInNamespace = await kubeCli.listPodsByNamepsace(namespace);
                const foundPod = allPodsInNamespace.find((pod) => pod.metadata.name === podName);
                return !foundPod;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            isPodDeleted,
            `wait for pod ${podName} to be deleted in namespace ${namespace}`,
            true,
            timeout
        );
    }

    async waitForPodPrefix(
        podNamePrefix: string,
        namespace: string,
        shouldExist = true,
        timeout = 120
    ): Promise<void> {
        const doesPodExist = async (): Promise<boolean> => {
            try {
                const allPodsInNamespace = await kubeCli.listPodsByNamepsace(namespace);
                const foundPod = allPodsInNamespace.find((pod) =>
                    pod.metadata.name.startsWith(podNamePrefix)
                );
                return !!foundPod;
            } catch (e) {
                return false;
            }
        };

        const description = `for pod with prefix ${podNamePrefix} in namespace ${namespace} to ${
            !shouldExist ? 'not ' : ''
        }exist`;

        await this.waitForCondition(
            doesPodExist,
            `waitForPodPrefix - wait ${description}`,
            shouldExist,
            timeout
        );

        logger.debug(`Successfully waited ${description}`);
    }

    async waitForPodsTerminationInNamespace(namespace: string, timeout: number): Promise<void> {
        const allPodsAreTerminated = async (): Promise<boolean> => {
            try {
                const pods = await kubeCli.listPodsByNamepsace(namespace);
                // all existing pods should have deletionTimestamp set. that means they are terminated but not yet done terminating (there's a grace period)
                if (pods.every((pod) => pod.metadata.deletionTimestamp)) {
                    return true;
                }
            } catch (e) {
                logger.error(`Failed to list pods in namespace ${namespace}: ${e}`);
            }
            return false;
        };

        await waitUtils.waitForCondition(
            allPodsAreTerminated,
            `wait for pods to be terminated for namespace ${namespace} for project cleanup.`,
            true,
            timeout
        );
    }

    async restartPodAndWait(podNamePrefix: string, namespace: string, timeout = 120): Promise<void> {
        const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, podNamePrefix);
        pods.forEach((pod) => {
            kubeCli.deletePod(pod.metadata.name, namespace);
            waitUtils.waitForPodToBeDeleted(pod.metadata.name, namespace, timeout);
        });

        // expect at least one new pod created and running
        await kubeCli.waitForPodToRun(namespace, podNamePrefix);
    }

    async getFullClaimName(claimNamePrefix: string, namespace): Promise<string> {
        logger.info(`wait for claim name with prefix: ${claimNamePrefix} in namespace: ${namespace}`);
        await this.waitForFullClaimName(claimNamePrefix, namespace);
        const pvcs = await kubeCli.listPVCsByNamespace(namespace);
        const pvc = pvcs.items.find((pvc) => pvc.metadata.name.startsWith(claimNamePrefix));
        const claimName = pvc?.metadata?.name;
        if (!claimName) {
            logger.error(
                `claim name with prefix: ${claimNamePrefix} was not found in namespace: ${namespace}`
            );
        } else {
            logger.info(`found claim name with prefix: ${claimNamePrefix} in namespace: ${namespace}`);
        }
        return claimName;
    }

    async waitForFullClaimName(claimNamePrefix: string, namespace, timeout = 90): Promise<void> {
        const doesClaimNameWithPrefixExist = async (): Promise<boolean> => {
            try {
                const pvcs = await kubeCli.listPVCsByNamespace(namespace);
                const pvc = pvcs.items.find((pvc) => pvc.metadata.name.startsWith(claimNamePrefix));
                return pvc.metadata.name !== undefined;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            doesClaimNameWithPrefixExist,
            `wait for claim name with prefix ${claimNamePrefix} to be created in namespace ${namespace}`,
            true,
            timeout
        );
    }

    async waitForPvcs(
        namespace: string,
        numberOfExpectedPVCs: number,
        timeout = 60,
        prefix?: string
    ): Promise<V1PersistentVolumeClaimList> {
        logger.debug(
            `wait For Pvcs, start with parameters: ${JSON.stringify({
                namespace,
                numberOfExpectedPVCs,
                timeout
            })}`
        );
        const isPVCsConditionMet = async (): Promise<V1PersistentVolumeClaimList> => {
            try {
                logger.debug('- isPVCsConditionMet start');
                const pvcList = await kubeCli.listPVCsByNamespace(namespace);
                logger.debug('- isPVCsConditionMet, pvcList:');
                pvcList.items.forEach((pvc) => {
                    logger.debug(
                        `--     pvc-metadata: ${JSON.stringify(pvc.metadata.name)}, pvc-phase: ${
                            pvc.status.phase
                        }`
                    );
                });
                logger.debug(
                    `- isPVCsConditionMet, numberOfExpectedPVCs: ${numberOfExpectedPVCs}, pvcList.items.length: ${pvcList.items.length}`
                );
                if (prefix) {
                    const filteredPVCs: V1PersistentVolumeClaim[] = [];
                    pvcList.items.forEach((pvc) => {
                        if (pvc.metadata.name.startsWith(prefix)) {
                            filteredPVCs.push(pvc);
                        }
                    });
                    pvcList.items = filteredPVCs;
                }
                if (numberOfExpectedPVCs === pvcList.items.length) {
                    logger.debug('- isPVCsConditionMet, succeeded - expected pvc number founded');
                    return pvcList;
                }
                logger.debug('- isPVCsConditionMet, failed for this attempt - expected pvc number not found');
                return undefined;
            } catch (e) {
                logger.debug(`failed to list pvcs in namespace ${namespace}: ${e}`);
                return undefined;
            }
        };

        const pvcList = await this.waitForObject(
            isPVCsConditionMet,
            `wait for pvcs to be created for namespace ${namespace}`,
            timeout
        );
        return pvcList as V1PersistentVolumeClaimList;
    }

    async waitForPodsByOwnerPrefix(
        ownerPrefix: string,
        namespace: string,
        instances = 1,
        timeout = 60
    ): Promise<V1Pod[]> {
        const isPodConditionMet = async (): Promise<object> => {
            try {
                const pods = await kubeCli.listPodsByOwnerPrefix(namespace, ownerPrefix);
                if (pods.length >= instances) {
                    logger.debug(
                        `found ${pods.length} pods for job ${ownerPrefix} in namespace ${namespace}`
                    );
                    return pods;
                } else {
                    logger.debug(
                        `waiting for ${instances} pods to be created for job ${ownerPrefix} in namespace ${namespace}`
                    );
                    return undefined;
                }
            } catch (e) {
                logger.debug(`job ${ownerPrefix} in namespace ${namespace} not found`);
                return undefined;
            }
        };

        const podList = await this.waitForObject(
            isPodConditionMet,
            `wait for ${instances} pods to be created for job ${ownerPrefix} in namespace ${namespace}`,
            timeout
        );
        return podList as V1Pod[];
    }

    async waitForService(serviceName: string, namespace: string, timeout = 70): Promise<V1Service> {
        const isSvcConditionMet = async (): Promise<object> => {
            try {
                return await kubeCli.getServiceByName(namespace, serviceName);
            } catch (e) {
                logger.debug(`service ${serviceName} in namespace ${namespace} not found`);
                return undefined;
            }
        };

        const service = await this.waitForObject(
            isSvcConditionMet,
            `wait for service ${serviceName} to be created in namespace ${namespace}`,
            timeout
        );
        return service as V1Service;
    }

    async waitForServiceAccount(name: string, namespace: string, timeout = 70): Promise<V1ServiceAccount> {
        const isConditionMet = async (): Promise<object> => {
            try {
                return await kubeCli.getServiceAccount(name, namespace);
            } catch (e) {
                logger.debug(`service account ${name} in namespace ${namespace} not found`);
                return undefined;
            }
        };

        const service = await this.waitForObject(
            isConditionMet,
            `wait for service account ${name} to be created in namespace ${namespace}`,
            timeout
        );
        return service as V1ServiceAccount;
    }

    async getPvc(name: string, namespace: string): Promise<V1PersistentVolumeClaim> {
        const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
            name,
            namespace
        });
        return pvc;
    }

    async waitForPvc(name: string, namespace: string, timeout = 90, shouldExist = true): Promise<void> {
        const doesPVCExist = async (): Promise<boolean> => {
            try {
                const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                    name,
                    namespace
                });
                return pvc !== undefined;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            doesPVCExist,
            `wait for Pvc ${name} to be created in namespace ${namespace}, shouldExist: ${shouldExist}`,
            shouldExist,
            timeout
        );
    }

    async waitForIngress(ingressName: string, namespace: string, timeout = 70): Promise<V1Ingress> {
        const isSvcConditionMet = async (): Promise<object> => {
            try {
                if (isOCP()) {
                    // There's no ingress in OCP. OpenShift uses Routes to manage Ingress
                    return kubeCli.getRouteByName(namespace, ingressName);
                } else {
                    return await kubeCli.getIngressByName(namespace, ingressName);
                }
            } catch (e) {
                logger.debug(`ingress ${ingressName} in namespace ${namespace} not found`);
                return undefined;
            }
        };

        const ingress = await this.waitForObject(
            isSvcConditionMet,
            `wait for ingress ${ingressName} to be created in namespace ${namespace}`,
            timeout
        );
        return ingress as V1Ingress;
    }

    async waitForJobToBeRunning(
        jobName: string,
        namespaceName: string,
        workloadType: string,
        timeout = 120
    ): Promise<void> {
        try {
            await this.waitForCondition(
                this.isJobRunning(jobName, namespaceName, workloadType),
                `wait for job ${jobName} in namespace ${namespaceName} to be running`,
                true,
                timeout
            );
        } catch (error) {
            logClusterDiagnostics(namespaceName, jobName);
            throw error;
        }
    }

    async waitForProjectRbac(namespace: string): Promise<void> {
        const roleBindingsCreated = async (): Promise<boolean> => {
            try {
                const roleBindings = await kubeCli.listRoleBindings(namespace);

                // Waiting for the following role bindings to be created:
                // - assets-sync                            (v2.9 and above)
                // - researcher-service-rw
                // - runai-admission-controller-project
                // - runai-binder-project                   (v2.13 and above)
                // - runai-project-controller-cluster-secret-per-project
                // - runai-project-controller-project
                // - runai-researcher
                // - runai-scheduler-rw
                // - status-updater                         (v2.13 and above)
                //
                // The following are no longer created after the controller retirement.
                // project-controller only renders them when the matching
                // RunaiConfig flag is enabled, and those flags now default to false:
                // - runai-job-controller-project           (spec.jobController.enabled)
                // - workload-controller-rw                 (spec.workloadController.enabled)
                const requiredRoleBindingsCount = 7;

                return roleBindings.length >= requiredRoleBindingsCount;
            } catch (e) {
                logger.error(`failed to get role bindings of project namespace ${namespace}: ${e}`);
                return false;
            }
        };

        await this.waitForCondition(
            roleBindingsCreated,
            `waitForRoleBindings of project namespace ${namespace} to be created`,
            true,
            70
        );
    }

    async waitForUserInProjectRoleBinding(
        namespace: string,
        researcherUserEmail: string,
        roleBindingName: string,
        timeout = 60
    ): Promise<void> {
        const roleBindingsCreated = async (): Promise<boolean> => {
            try {
                logger.debug(
                    `Waiting for ${researcherUserEmail} in ${roleBindingName} role binding in namespace ${namespace}`
                );
                const roleBinding = await kubeCli.getRoleBindingByName(namespace, roleBindingName);
                if (!roleBinding) {
                    return false;
                }
                return roleBinding.subjects.some(
                    (subject) => subject.kind === 'User' && subject.name === researcherUserEmail
                );
            } catch (e) {
                logger.error(
                    `failed to get the ${roleBindingName} role binding of project namespace ${namespace}: ${e}`
                );
                return false;
            }
        };

        await this.waitForCondition(
            roleBindingsCreated,
            `waitForUserInProjectRoleBinding of project namespace ${namespace} to be created`,
            true,
            timeout
        );
    }

    async waitForUserInClusterRoleBinding(
        userEmail: string,
        clusterRoleBindingName: string,
        timeout = 60
    ): Promise<void> {
        const clusterRoleBindingsCreated = async (): Promise<boolean> => {
            let clusterRoleBinding;
            try {
                logger.debug(`Waiting for ${userEmail} in ${clusterRoleBindingName} cluster role binding`);

                clusterRoleBinding = await kubeCli.getClusterRoleBinding(clusterRoleBindingName);

                if (!clusterRoleBinding) {
                    return false;
                }
                return clusterRoleBinding.subjects.some(
                    (subject) => subject.kind === 'User' && subject.name === userEmail
                );
            } catch (e) {
                logger.error(`failed to get the ${clusterRoleBindingName} cluster role binding: ${e}`);
                return false;
            }
        };

        await this.waitForCondition(
            clusterRoleBindingsCreated,
            'waitForUserInProjectClusterRoleBinding to be created',
            true,
            timeout
        );
    }

    async waitForPodGroups(
        jobName: string,
        namespace: string,
        instances = 1,
        timeout = 70
    ): Promise<KubernetesObject[]> {
        let podGroups: KubernetesObject[];
        const desiredResult = true;

        const allCrdsFound = async (): Promise<boolean> => {
            try {
                podGroups = await kubeCli.getCrds(crds.podGroupsCrd, namespace);
                podGroups = podGroups.filter((podGroup) => podGroup.metadata.name.includes(jobName));
                return podGroups.length >= instances;
            } catch (e) {
                logger.error(`${podGroups.length} not eql to ${instances}`, e);
                return false;
            }
        };

        await this.waitForCondition(
            allCrdsFound,
            `waitForCrd podGroups ${crds.podGroupsCrd.plural}.${crds.podGroupsCrd.group}/${crds.podGroupsCrd.version} jobName: ${jobName} to be ${desiredResult}`,
            desiredResult,
            timeout
        );

        return podGroups;
    }

    async waitForWorkload(
        crd: CustomResourceDefinition,
        jobName: string,
        namespace: string,
        timeout = 70
    ): Promise<KubernetesObject[]> {
        let workloads: KubernetesObject[];
        const desiredResult = true;

        const allCrdsFound = async (): Promise<boolean> => {
            try {
                workloads = await kubeCli.getCrds(crd, namespace);
                workloads = workloads.filter((workload) => workload.metadata.name.includes(jobName));
                return workloads.length > 0;
            } catch (e) {
                logger.error(`no ${crd.kind} found to match job ${jobName} in namespace ${namespace}`, e);
                return false;
            }
        };

        await this.waitForCondition(
            allCrdsFound,
            `waitForCrd ${crd.plural}.${crd.group}/${crd.version} jobName: ${jobName} to be ${desiredResult}`,
            desiredResult,
            timeout
        );

        return workloads;
    }

    async waitForTrainingWorkload(
        jobName: string,
        namespace: string,
        timeout = 70
    ): Promise<KubernetesObject[]> {
        return await this.waitForWorkload(crds.runaiTrainingWorkloadCrd, jobName, namespace, timeout);
    }

    async waitForInferenceWorkload(
        jobName: string,
        namespace: string,
        timeout = 70
    ): Promise<KubernetesObject[]> {
        return await this.waitForWorkload(crds.runaiInferenceWorkloadCrd, jobName, namespace, timeout);
    }

    async waitForRunaiJob(jobName: string, namespace: string, timeout = 70): Promise<KubernetesObject[]> {
        return await this.waitForWorkload(crds.runaiJobCrd, jobName, namespace, timeout);
    }

    async waitForDistributedWorkload(
        jobName: string,
        namespace: string,
        timeout = 70
    ): Promise<KubernetesObject[]> {
        return await this.waitForWorkload(crds.runaiDistributedWorkloadCrd, jobName, namespace, timeout);
    }

    async waitForPyTorchJob(jobName: string, namespace: string, timeout = 70): Promise<KubernetesObject[]> {
        return await this.waitForWorkload(crds.runaiPyTorchJobCrd, jobName, namespace, timeout);
    }

    async waitForWorkloadNonEmptyStatus(
        workloadName: string,
        namespace: string,
        crd: CustomResourceDefinition,
        timeout = 70
    ): Promise<boolean> {
        // for inference, to support both old and new controllers, we look for either status.creationStatus (old) or status.phase (new)
        const nonEmptyStatusFound = async (): Promise<boolean> => {
            try {
                const workload = await kubeCli.getWorkload(workloadName, namespace, crd);
                logger.debug(
                    `workload ${workloadName} namespace ${namespace} kind ${crd.kind} status: ${JSON.stringify(workload['status'])}`
                );
                return (
                    workload['status']['creationStatus'] !== undefined ||
                    workload['status']['phase'] !== undefined
                );
            } catch (e) {
                logger.error(`workload ${workloadName} has no status.creationStatus or status.phase`, e);
                return false;
            }
        };

        await this.waitForCondition(
            nonEmptyStatusFound,
            `waitForCrd ${crd.kind} workloadName: ${workloadName} namespace: ${namespace} to have either status.creationStatus set or status.phase`,
            true,
            timeout
        );

        return true;
    }

    async waitForContainersToBeReady(
        jobName: string,
        namespace: string,
        podName: string,
        numReadyContainers: number,
        timeout = 70
    ): Promise<boolean> {
        const numReadyContainersFound = async (): Promise<boolean> => {
            const allPodsInNamespace = await kubeCli.listPodsByNamepsace(namespace);
            const foundPod = allPodsInNamespace.find((pod) => pod.metadata.name.startsWith(podName));

            if (!foundPod) {
                return false;
            }

            const numReadyFound = foundPod?.status?.containerStatuses?.filter(
                (status) => status.ready
            ).length;
            return numReadyFound === numReadyContainers;
        };

        await this.waitForCondition(
            numReadyContainersFound,
            `waitForCrd pod ${podName} to have ${numReadyContainers} ready containers`,
            true,
            timeout
        );

        return true;
    }

    async waitForUnschedulableInNamespace(
        jobName: string,
        namespace: string,
        clusterTimeout = 240,
        description = '',
        failOnRunningPods = false,
        desiredResult = true,
        isPrefixName = false
    ): Promise<void> {
        if (description === '') {
            description = `Unschedulable event for job ${jobName} in namespace ${namespace}`;
            if (!desiredResult) {
                description = `No ${description}`;
            }
        }

        await waitUtils.waitForCondition(
            async (): Promise<boolean> => {
                const pods = await kubeCli.listPodsOfJob(namespace, jobName, isPrefixName);
                const foundUnschedulable =
                    pods.filter((pod) => SchedulingUtils.isPodUnschedulable(pod)).length === pods.length;

                if (failOnRunningPods) {
                    const runningPods = pods.filter((pod) => pod.status.phase === PodStatus.RUNNING);
                    expect(runningPods).length(
                        0,
                        `We are expecting the pods of the job ${jobName} to be unschedulable.` +
                            ` ${runningPods.length} pods are running.`
                    );
                }

                return foundUnschedulable;
            },
            description,
            desiredResult,
            clusterTimeout
        );
    }

    async waitForUnschedulable(
        jobName: string,
        projectName: string,
        clusterTimeout = 240,
        description = '',
        failOnRunningPods = false,
        desiredResult = true
    ): Promise<void> {
        await this.waitForUnschedulableInNamespace(
            jobName,
            Utils.namespaceFromProjectName(projectName),
            clusterTimeout,
            description,
            failOnRunningPods,
            desiredResult
        );
    }

    async waitForPodSchedulingDecisionTime(
        jobName: string,
        namespace: string,
        clusterTimeout = 240,
        description = '',
        desiredResult = true,
        isPrefixName = false
    ): Promise<Date> {
        if (description === '') {
            description = `PodScheduled set for job ${jobName} in namespace ${namespace}`;
            if (!desiredResult) {
                description = `No ${description}`;
            }
        }

        const isPodConditionMet = async (): Promise<object> => {
            try {
                const pods = await kubeCli.listPodsOfJob(namespace, jobName, isPrefixName);
                if (pods.length === 0) {
                    return undefined;
                }
                return SchedulingUtils.getPodSchedulingDecisionTime(pods[0]);
            } catch (e) {
                logger.debug(
                    `waitForPodSchedulingDecisionTime: job ${jobName} in namespace ${namespace} not found: ${e}`
                );
                return undefined;
            }
        };

        const podScheduledTime = await waitUtils.waitForObject(
            isPodConditionMet,
            description,
            clusterTimeout
        );

        return podScheduledTime as Date;
    }

    // TODO: add workloadType and change to waitForWorkloadCondition
    async waitForTrainingWorkloadCondition(
        jobName: string,
        namespace: string,
        conditionType: WorkloadConditionType,
        timeoutInSeconds = 70
    ): Promise<WorkloadConditionType> {
        let workloadCondition, trainWorkload;
        const workloadConditionFound = async (): Promise<boolean> => {
            trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
            workloadCondition = workloadUtils.getCondition(trainWorkload, conditionType);
            return workloadCondition !== null;
        };

        await this.waitForCondition(
            workloadConditionFound,
            `waitForTrainingWorkloadCondition, condition type: ${conditionType.toString()} `,
            true,
            timeoutInSeconds
        );

        return workloadCondition;
    }

    async waitForBackendJob(
        jobName: string,
        filter?: string,
        timeout = 60,
        delay = 1
    ): Promise<DisplayedJob> {
        let backendJob;
        const isBackendDeploymentFound = async (): Promise<boolean> => {
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            try {
                let allJobs;
                if (filter) {
                    allJobs = (
                        await backendClusterApi.jobsApi.listJobs(
                            flowRootState.cluster.uuid,
                            undefined,
                            filter
                        )
                    ).body;
                } else {
                    allJobs = (await backendClusterApi.jobsApi.listJobs(flowRootState.cluster.uuid)).body;
                }
                backendJob = allJobs.find((displayedJob) => displayedJob.jobName === jobName);
                return backendJob !== undefined;
            } catch (e) {
                logger.error(`Failed to get job ${jobName} from backend`, e);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendDeploymentFound,
            `waitForBackendJob jobName: ${jobName}`,
            true,
            timeout,
            delay
        );

        return backendJob;
    }

    async waitForWorkloadOwnershipProtectionWebhook(isEnabled: boolean, timeout = 120): Promise<void> {
        const legacyConfigurationName = 'runai-workload-ownership-protection-validation';
        const anyworkloadConfigurationName = 'runai-anyworkload-validation';
        const anyworkloadOwnershipWebhookName = 'workloadrunner-ownership.anyworkload-controller.runai.svc';

        const isOwnershipProtectionRegistered = async (): Promise<boolean> => {
            const configurations = (await kubeCli.listValidatingWebhooks()).items;
            // Clusters that still serve the v2alpha1 workload CRDs register a dedicated
            // configuration; newer ones only add an entry to the anyworkload configuration.
            const legacyRegistered = configurations.some(
                (configuration) => configuration.metadata?.name === legacyConfigurationName
            );
            const anyworkloadRegistered =
                configurations
                    .find((configuration) => configuration.metadata?.name === anyworkloadConfigurationName)
                    ?.webhooks?.some((webhook) => webhook.name === anyworkloadOwnershipWebhookName) ?? false;
            return legacyRegistered || anyworkloadRegistered;
        };

        await this.waitForCondition(
            isOwnershipProtectionRegistered,
            `waitForWorkloadOwnershipProtectionWebhook to be ${isEnabled ? 'registered' : 'unregistered'}`,
            isEnabled,
            timeout
        );
    }

    async waitForValidatingWebhook(webhookName: string, timeout = 60): Promise<void> {
        const validatingWebhookExistsFn = async (): Promise<boolean> => {
            const validatingWebhookConfigurationList = await kubeCli.listValidatingWebhooks();
            const foundWebhook = validatingWebhookConfigurationList.items.find(
                (obj) => obj.metadata.name === webhookName
            );
            if (foundWebhook) {
                return true;
            }
            return false;
        };

        await this.waitForCondition(
            validatingWebhookExistsFn,
            `waitForValidatingWebhook to have Validating Webhook: ${webhookName}`,
            true,
            timeout
        );
    }

    async waitForBackendJobStatus(
        jobName: string,
        expectedStatus: string,
        timeout = 60
    ): Promise<DisplayedJob> {
        let jobStatus: string;
        let backendJob: DisplayedJob;

        const isBackendJobWithCorrectStatus = async (): Promise<boolean> => {
            try {
                backendJob = await waitUtils.waitForBackendJob(jobName);
                jobStatus = backendJob.status;
                logger.info(`jobStatus: ${jobStatus}`);
                return jobStatus === expectedStatus;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendJobWithCorrectStatus,
            `jobName: ${jobName}, expected status: ${expectedStatus}, backend job: ${JSON.stringify(backendJob)}`,
            true,
            timeout
        );

        return backendJob;
    }

    async waitForBackendWorkloadPhase(
        workloadName: string,
        projectId: number | string,
        expectedStatus: Phase,
        timeout: number = 60,
        limit: any = undefined
    ): Promise<Workload> {
        let workload: Workload;
        let workloadPhase: Phase;
        let prevMessage: string;

        const isBackendWorkloadWithCorrectPhase = async (): Promise<boolean> => {
            try {
                const workloads = (
                    await workloadsUtilsV2.getWorkloadByName(workloadName, projectId.toString(), undefined)
                ).body.workloads;
                workload = workloads.find((w) => w.name === workloadName);
                if (!workload || workload.phase === undefined) {
                    return false;
                }
                let message = `workload phase: ${workload.phase}`;
                if (workload.phaseMessage) {
                    message += `, phase message: ${workload.phaseMessage.trim()}`;
                }
                if (message !== prevMessage) {
                    // to avoid logging the same phase again and again
                    logger.info(message);
                    prevMessage = message;
                }
                return workload.phase === expectedStatus;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        logger.debug(`waiting for workload ${workloadName} to become ${expectedStatus}`);
        try {
            await waitUtils.waitForCondition(
                isBackendWorkloadWithCorrectPhase,
                `workload: ${workloadName}, expected status: ${expectedStatus}`,
                true,
                timeout
            );
        } catch (err) {
            if (workload) {
                const namespace = 'runai-' + workload.projectName;
                const events = kubeCli.listEventsByNamespace(namespace);
                logger.info(`events in namespace ${namespace}:\n${events}`);
                logClusterDiagnostics(namespace, workloadName);
            }
            throw err;
        }

        return workload;
    }

    async waitForBackendWorkload(
        workloadName: string,
        projectId: string,
        timeout = 60,
        limit = undefined
    ): Promise<Workload> {
        let workload: Workload;

        const isBackendWorkloadAvailable = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)
                ) {
                    // Workloads API is not supported in versions below 2.16
                    return true;
                }
                const workloads = (
                    await workloadsUtilsV2.getWorkloadByName(workloadName, projectId.toString(), undefined)
                ).body.workloads;
                workload = workloads.find((w) => w.name === workloadName);
                return !!workload;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendWorkloadAvailable,
            `workload: ${workloadName} available in CP`,
            true,
            timeout
        );

        return workload;
    }

    async waitForBackendWorkloadByIdAndPhase(
        workloadId: string,
        phase: Phase,
        timeout = 60,
        runningPods?: number
    ): Promise<Workload> {
        let workload: WorkloadDetailed;
        let prevMessage: string;

        const isBackendWorkloadAvailable = async (): Promise<boolean> => {
            try {
                workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                let message = `workload phase: ${workload.phase}`;
                if (workload.phaseMessage) {
                    message += `, phase message: ${workload.phaseMessage.trim()}`;
                }
                if (message !== prevMessage) {
                    // to avoid logging the same phase again and again
                    logger.info(message);
                    prevMessage = message;
                }
                const phaseCorrect = !!workload && workload.phase === phase;

                let runningPodsCorrect = true;
                if (phaseCorrect && !!runningPods) {
                    logger.info(`runningPods: ${workload.runningPods}`);
                    runningPodsCorrect = !!workload && workload.runningPods === runningPods;
                }

                return phaseCorrect && runningPodsCorrect;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isBackendWorkloadAvailable,
                `workload: ${workloadId}, expected phase: ${phase}` +
                    (runningPods !== undefined ? `, expected runningPods: ${runningPods}` : ''),
                true,
                timeout,
                3 // slower than the 1s default to avoid stressing the apiserver/admission-controller under parallel load
            );
        } catch (err) {
            if (workload) {
                const namespace = 'runai-' + workload.projectName;
                const events = kubeCli.listEventsByNamespace(namespace);
                logger.info(`events in namespace ${namespace}:\n${events}`);
            }
            throw err;
        }

        return workload;
    }

    private async waitForBackendWorkloadByIdAndField<T>(
        workloadId: string,
        fieldName: string,
        matcher: (workload: WorkloadDetailed) => boolean,
        getCurrentValue: (workload: WorkloadDetailed) => T,
        expectedValue: T,
        timeout = 60
    ): Promise<Workload> {
        let workload: WorkloadDetailed;
        let prevValue: T;

        const isConditionMet = async (): Promise<boolean> => {
            try {
                workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                const currentValue = getCurrentValue(workload);
                if (currentValue !== prevValue) {
                    logger.info(`workload ${fieldName}: ${currentValue}`);
                    prevValue = currentValue;
                }
                return matcher(workload);
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isConditionMet,
                `workload: ${workloadId}, expected ${fieldName}: ${expectedValue}`,
                true,
                timeout
            );
        } catch (err) {
            if (workload) {
                const namespace = 'runai-' + workload.projectName;
                const events = kubeCli.listEventsByNamespace(namespace);
                logger.info(`events in namespace ${namespace}:\n${events}`);
            }
            throw err;
        }

        return workload;
    }

    async waitForBackendWorkloadByIdAndPriority(
        workloadId: string,
        priorityClassName: string,
        timeout = 60
    ): Promise<Workload> {
        return this.waitForBackendWorkloadByIdAndField(
            workloadId,
            'priorityClassName',
            (workload) => workload.priorityClassName === priorityClassName,
            (workload) => workload.priorityClassName,
            priorityClassName,
            timeout
        );
    }

    async waitForBackendWorkloadByIdAndCategory(
        workloadId: string,
        category: string,
        timeout = 60
    ): Promise<Workload> {
        return this.waitForBackendWorkloadByIdAndField(
            workloadId,
            'category',
            (workload) => workload.category === category,
            (workload) => workload.category,
            category,
            timeout
        );
    }

    async waitForWorkloadPhaseAndImage(
        workloadId: string,
        projectId: string,
        phase: Phase,
        expectedImage: string,
        timeout = 120,
        interval = 5
    ): Promise<Workload> {
        let workload: Workload;

        const isWorkloadReady = async (): Promise<boolean> => {
            try {
                const workloads = (await workloadsUtilsV2.getWorkloadsByProjectId(projectId)).body.workloads;
                workload = workloads.find((w: Workload) => w.id === workloadId);

                if (!workload || workload.phase !== phase) {
                    logger.debug(`Waiting: workload phase=${workload?.phase || 'not found'}`);
                    return false;
                }

                const actualImage = workload.images?.[0] || '';
                const hasExpectedImage =
                    actualImage === expectedImage || actualImage.startsWith(expectedImage + '@');

                if (!hasExpectedImage) {
                    logger.debug(`Waiting: image=${actualImage} (expected: ${expectedImage})`);
                }

                return hasExpectedImage;
            } catch (error) {
                logger.debug(`Error checking workload: ${error.message}`);
                return false;
            }
        };

        await this.waitForCondition(
            isWorkloadReady,
            `workload ${workloadId} phase=${phase} and image=${expectedImage}`,
            true,
            timeout,
            interval
        );

        return workload;
    }

    async waitForImagePullFailure(
        namespace: string,
        workloadName: string,
        timeout = 60,
        interval = 5
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                const relatedPods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);

                if (relatedPods.length === 0) {
                    logger.debug(
                        `No pods found for workloadName="${workloadName}" yet (namespace=${namespace}).`
                    );
                    return false;
                }

                return relatedPods.some((pod) =>
                    pod.status?.containerStatuses?.some((status) => {
                        const reason = status.state?.waiting?.reason;
                        if (reason === 'ImagePullBackOff' || reason === 'ErrImagePull') {
                            const message = status.state?.waiting?.message || '';
                            logger.info(
                                `Detected image pull failure in pod ${pod.metadata?.name}: ${reason} - ${message}`
                            );
                            return true;
                        }
                        return false;
                    })
                );
            },
            `pods of workload "${workloadName}" to show image pull failure`,
            true,
            timeout,
            interval
        );
    }

    async waitForBackendWorkloadPhaseInHistory(
        workloadId: string,
        workloadName: string,
        expectedPhase: Phase,
        timeout = 60
    ): Promise<void> {
        const isBackendWorkloadHadCorrectPhase = async (): Promise<boolean> => {
            try {
                const workloadStatusHistory = (
                    await workloadsUtilsV2.getWorkloadHistory(workloadId)
                ).body.records.filter((r) => r.meta.type === HistoryRecordType.PhaseUpdate);
                logger.info(`workloadStatusHistory: ${JSON.stringify(workloadStatusHistory)}`);
                return workloadStatusHistory.some((r) => r.spec.phaseUpdate.phase === expectedPhase);
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendWorkloadHadCorrectPhase,
            `workload: ${workloadName}, expected status: ${expectedPhase}`,
            true,
            timeout
        );
    }

    async waitForWorkspaceCreation(clusterId: string, workspaceName: string, timeout = 70): Promise<boolean> {
        const isWorkspaceCreatedConditionMet = async (): Promise<object> => {
            logger.debug(
                `Sending GET request from user: ${flowRootState.users.system_administrator.email}. With token: ${flowRootState.users.system_administrator.accessToken} `
            );
            try {
                assetsApi.workspaceApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const res: { body: WorkspaceListResponse } =
                    await assetsApi.workspaceApi.listWorkspaces(clusterId);
                const workspace: Workspace | undefined = res.body.entries.find(
                    (workspace: Workspace) => workspace.meta.name === workspaceName
                );
                if (workspace.job?.project && workspace.job?.status) {
                    logger.debug(`Workspace ${workspaceName} is created`);
                    return { found: true };
                }

                logger.debug(`Waiting for workspace ${workspaceName} to be created`);
                return undefined;
            } catch (e) {
                logger.debug(`waiting to workspace ${workspaceName} created had an error: ${e}`);
                return undefined;
            }
        };

        const isWorkspaceCreated = await this.waitForObject(
            isWorkspaceCreatedConditionMet,
            `wait for workspace ${workspaceName} to be created`,
            timeout
        );
        return !!isWorkspaceCreated;
    }

    async waitForWorkspaceRunning(clusterId: string, workspaceName: string, timeout = 70): Promise<boolean> {
        const isWorkspaceCreatedConditionMet = async (): Promise<object> => {
            try {
                assetsApi.workspaceApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const res: { body: WorkspaceListResponse } =
                    await assetsApi.workspaceApi.listWorkspaces(clusterId);
                const workspace: Workspace | undefined = res.body.entries.find(
                    (workspace: Workspace) => workspace.meta.name === workspaceName
                );
                if (workspace.job?.status === 'Active' && workspace.job?.jobUrl) {
                    logger.debug(`Workspace ${workspaceName} is running`);
                    return { found: true };
                }

                logger.debug(`Waiting for workspace ${workspaceName} to be running`);
                return undefined;
            } catch (e) {
                logger.debug(`wating to workspace ${workspaceName} running had an error: ${e}`);
                return undefined;
            }
        };

        const isWorkspaceCreated = await this.waitForObject(
            isWorkspaceCreatedConditionMet,
            `wait for workspace ${workspaceName} to be created`,
            timeout
        );
        return !!isWorkspaceCreated;
    }

    async waitForWorkspaceStatus(
        clusterId: string,
        workspaceName: string,
        status: string,
        timeout = 70
    ): Promise<void> {
        assetsApi.workspaceApi.accessToken = flowRootState.users.research_manager.accessToken;
        const isWorkspaceStatusMatch = async (): Promise<boolean> => {
            // Important: careful using this during UI tests, there are occasions where the status in the backend changed but the status in the UI wasnt and it may lead to issues
            const workspaces: Array<Workspace> = (await assetsApi.workspaceApi.listWorkspaces(clusterId)).body
                .entries;
            const workspace: Workspace | undefined = workspaces.find(
                (workspace: Workspace) => workspace.meta.name === workspaceName
            );
            return workspace?.job?.status === status;
        };

        await this.waitForCondition(
            isWorkspaceStatusMatch,
            `wait for workspace ${workspaceName} to be ${status}`,
            true,
            timeout
        );
    }

    async waitForWorkloadAdditionalField(
        workloadId: string,
        fieldName: string,
        fieldValue: string,
        timeout = 120
    ): Promise<WorkloadDetailed> {
        let workload: WorkloadDetailed;

        const isBackendWorkloadWithAdditionalFieldValue = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_16)
                ) {
                    // Workloads API is not supported in versions below 2.16
                    return true;
                }
                workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                return (
                    workload !== undefined &&
                    workload.additionalFields !== undefined &&
                    workload.additionalFields[fieldName] === fieldValue
                );
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendWorkloadWithAdditionalFieldValue,
            `workloadId: ${workloadId}, expected additional field/value: ${fieldName}/${fieldValue}`,
            true,
            timeout
        );

        return workload;
    }

    async waitUntilRunaiBackendIsReady(timeout = 200): Promise<void> {
        const isRunaiBackendReady = async (): Promise<boolean> => {
            const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai-backend');
            logger.info(`waiting pod status to be ready: ${podByNs}`);
            return podByNs;
        };
        await waitUtils.waitForCondition(isRunaiBackendReady, 'Runai backend is ready', true, timeout);
    }

    async waitForWorkspaceTemplateDeletion(name: string, projectId?: number, timeout = 70): Promise<boolean> {
        const isWorkspaceTemplateDeletedConditionMet = async (): Promise<object> => {
            try {
                assetsApi.templateApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const templates: Array<WorkloadTemplate> = (
                    await assetsApi.templateApi.listTemplates(flowRootState.cluster.uuid)
                ).body.entries;

                const count: number = templates.filter(
                    (template: WorkloadTemplate) => template.meta.name === name
                ).length;
                if (count === 0) {
                    logger.debug(`workspace template ${name} deleted`);
                    return { found: true };
                } else {
                    logger.debug(`waiting for workspace template ${name} to be deleted`);
                    return undefined;
                }
            } catch (e) {
                logger.debug(`wating to workspace template ${name} deleted had an error: ${e}`);
                return undefined;
            }
        };

        const isWorkspaceTemplateDeleted = await this.waitForObject(
            isWorkspaceTemplateDeletedConditionMet,
            `wait for workspace template ${name} to be deleted`,
            timeout
        );
        return !!isWorkspaceTemplateDeleted;
    }

    async waitForBackendRevisionPhase(
        revisionName: string,
        workloadId: string,
        expectedStatus: Phase,
        timeout = 60,
        limit = undefined
    ): Promise<Revision> {
        let revision: Revision;
        let revisionPhase: Phase;
        logger.debug('wait for revision ', revisionName);

        const isBackendRevisionWithCorrectPhase = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_23)
                ) {
                    // Revisions are not supported in versions below 2.23
                    return true;
                }
                const revisions = (
                    await workloadsService.revisionsApi.getInferenceWorkloadRevisions(workloadId)
                ).body.revisions;
                revision = revisions.find((r) => r.name === revisionName);
                if (!revision) {
                    return false;
                }
                if (revision.phase !== revisionPhase) {
                    // to avoid logging the same phase again and again
                    revisionPhase = revision.phase;
                    logger.info(`revisionPhase: ${revisionPhase}`);
                }
                return revisionPhase === expectedStatus;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendRevisionWithCorrectPhase,
            `revision: ${revisionName}, expected status: ${expectedStatus}`,
            true,
            timeout
        );

        return revision;
    }

    async waitForBackendRevision(
        revisionName: string,
        workloadId: string,
        timeout = 60,
        limit = undefined
    ): Promise<Revision> {
        let revision: Revision;

        const isBackendRevisionAvailable = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_23)
                ) {
                    // Revisions are not supported in versions below 2.23
                    return true;
                }
                const revisions = (
                    await workloadsService.revisionsApi.getInferenceWorkloadRevisions(workloadId)
                ).body.revisions;
                revision = revisions.find((r) => r.name === revisionName);
                return !!revision;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendRevisionAvailable,
            `revision: ${revisionName} available in CP`,
            true,
            timeout
        );

        return revision;
    }

    async waitForBackendRevisionPodsPhase(
        revisionId: string,
        expectedStatus: string,
        timeout = 60
    ): Promise<RevisionPod[]> {
        let pods: RevisionPod[];
        let podsPhaseCorrect: boolean = true;
        let podsPhases: string;

        const isBackendRevisionPodWithCorrectPhase = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_23)
                ) {
                    // Pods are not supported in versions below 2.23
                    return true;
                }
                pods = (await workloadsService.podsApi.getRevisionPods(revisionId)).body.pods;
                if (!pods) {
                    return false;
                }
                for (const pod of pods) {
                    if (pod.k8sPhase !== expectedStatus) {
                        podsPhaseCorrect = false;
                    }
                }

                if (!podsPhaseCorrect) {
                    podsPhases = pods.map((pod) => pod.k8sPhase).join(', ');
                    logger.info(`podsPhases: ${podsPhases}`);
                }
                return podsPhaseCorrect;
            } catch (e) {
                logger.error(`Got an error: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendRevisionPodWithCorrectPhase,
            `pod: ${podsPhases}, expected status: ${expectedStatus}`,
            true,
            timeout
        );

        return pods;
    }

    async waitForBackendRevisionEvents(
        revisionId: string,
        timeout = 60,
        minLength = 1,
        minLengthOfKindRevision = 0
    ): Promise<Array<Event>> {
        let events: Array<Event> = [];

        const hasBackendRevisionEvents = async (): Promise<boolean> => {
            try {
                if (
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_23)
                ) {
                    // Revision events are not supported in versions below 2.23
                    events = []; // ensure we always return an array
                    return true;
                }

                const resp = await workloadsService.eventsApi.getRevisionEvents(revisionId);
                events = resp?.body?.events ?? [];

                const meetsMinLength = events.length >= minLength;
                const shouldCheckRevisionKind = minLengthOfKindRevision > 0;

                if (meetsMinLength && shouldCheckRevisionKind) {
                    const revisionEventsCount = events.filter(
                        (e) => e?.involvedObject?.kind === 'Revision'
                    ).length;
                    return revisionEventsCount >= minLengthOfKindRevision;
                }

                return meetsMinLength;
            } catch (e: any) {
                logger.error(`Got an error: ${e?.message ?? e}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(hasBackendRevisionEvents, `revision: ${revisionId}`, true, timeout);

        return events;
    }

    // Waiting for element to get content that depends on an async operation.
    async waitForElementContent(page: BaseV2Page, selector: string, content: string): Promise<void> {
        const element = page.v2frame.locator(selector);
        await waitUtils.waitForCondition(
            async () => {
                const a = (await element.textContent()).trim();
                return a === content;
            },
            'wait for element to contain the text',
            true,
            3000
        );
    }

    async waitForElementToNotExist(page: BaseV2Page, selector: string, timeout?: number): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await page.v2frame.locator(selector).count()) === 0,
            '',
            true,
            timeout || 60,
            0.5
        );
    }

    async waitForElementToNotExistWithLocator(locator: Locator, timeout?: number): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await locator.count()) === 0,
            '',
            true,
            timeout || 60,
            0.5
        );
    }

    async waitForV1ElementToNotExist(page: Page, selector: string, timeout?: number): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await page.locator(selector).count()) === 0,
            '',
            true,
            timeout || 20,
            0.5
        );
    }

    async waitForElementToExist(page: BaseV2Page, selector: string, timeoutSeconds = 10): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await page.v2frame.locator(selector).count()) > 0,
            `Waiting for element ${selector} to exist`,
            true,
            timeoutSeconds,
            0.5
        );
    }

    async waitForElementToExistLocator(locator: Locator): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await locator.count()) > 0,
            `Waiting for element ${locator.toString()} to exist`,
            true,
            10,
            0.5
        );
    }

    async waitForElementToExistWithLocator(locator: Locator): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await locator.count()) > 0,
            `Waiting for element with locator: ${locator} to exist`,
            true,
            10,
            0.5
        );
    }

    async waitForV1ElementToExist(page: Page, selector: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => (await page.locator(selector).count()) > 0,
            `Waiting for element ${selector} to exist`,
            true,
            10,
            0.5
        );
    }

    async waitForSecret(secretName: string, namespace: string, timeout = 30): Promise<void> {
        const doesSecretExist = async (): Promise<boolean> => {
            try {
                const allSecretsInNamespace = await kubeCli.listSecretsByNamepsace(namespace);
                if (allSecretsInNamespace === undefined || allSecretsInNamespace.items === undefined) {
                    return false;
                }
                const foundSecret = allSecretsInNamespace.items.find((secret) => {
                    return secret.metadata.name === secretName;
                });
                return foundSecret !== undefined;
            } catch (e) {
                logger.info(`failed to get secrets in namespace: ${namespace}, error: ${e}`);
                return false;
            }
        };

        await this.waitForCondition(
            doesSecretExist,
            `wait for secret ${secretName} to be created in namespace ${namespace}`,
            true,
            timeout
        );
    }

    async waitForSecretToBeDeleted(secretName: string, namespace: string, timeout = 30): Promise<void> {
        const isSecretDeleted = async (): Promise<boolean> => {
            try {
                const allSecretsInNamespace = await kubeCli.listSecretsByNamepsace(namespace);
                if (allSecretsInNamespace === undefined || allSecretsInNamespace.items === undefined) {
                    return true;
                }
                const foundSecret = allSecretsInNamespace.items.find(
                    (secret) => secret.metadata.name === secretName
                );
                return !foundSecret;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            isSecretDeleted,
            `wait for secret ${secretName} to be deleted in namespace ${namespace}`,
            true,
            timeout
        );
    }

    async waitForWorkloadToBeDeleted(
        workloadName: string,
        timeout = 70,
        accessToken: string = flowRootState.users.research_manager.accessToken
    ): Promise<void> {
        workloadsService.workloadsApi.accessToken = accessToken;
        const isWorkloadExist = async (): Promise<boolean> => {
            const { body } = await workloadsService.workloadsApi.getWorkloads();
            const workload: Workload | undefined = body.workloads.find(
                (workload: Workload) => workload.name === workloadName
            );

            return !workload;
        };

        await this.waitForCondition(
            isWorkloadExist,
            `wait for workload ${workloadName} to be deleted`,
            true,
            timeout
        );
    }

    async waitForTrainingStatus(
        clusterId: string,
        trainingName: string,
        status: string,
        timeout = 70
    ): Promise<void> {
        assetsApi.trainingApi.accessToken = flowRootState.users.research_manager.accessToken;
        const isTrainingStatusMatch = async (): Promise<boolean> => {
            const trainings: Array<Training> = (await assetsApi.trainingApi.listTrainings(clusterId)).body
                .entries;
            const training: Training | undefined = trainings.find(
                (training: Training) => training.meta.name === trainingName
            );
            return training?.job?.status === status;
        };

        await this.waitForCondition(
            isTrainingStatusMatch,
            `wait for workspace ${trainingName} to be ${status}`,
            true,
            timeout
        );
    }

    async waitForKnativeService(
        name: string,
        namespace: string,
        requireUrl = true,
        timeout = 90
    ): Promise<KnativeService> {
        const knativeServiceExistsWithUrl = async (): Promise<object> => {
            try {
                const knativeService = await kubeCli.getKnativeService(name, namespace);
                if (requireUrl && knativeService?.status?.address?.url === undefined) {
                    return undefined;
                }
                return knativeService;
            } catch (e) {
                return undefined;
            }
        };

        const knativeService = await this.waitForObject(
            knativeServiceExistsWithUrl,
            `wait for KnativeService ${name} to be created in namespace ${namespace} and be assigned with a url`,
            timeout
        );

        return knativeService as KnativeService;
    }

    async waitForKnativeRevision(name: string, namespace: string, timeout = 90): Promise<KnativeRevision> {
        const knativeRevisionExists = async (): Promise<object> => {
            try {
                const knativeService = await kubeCli.getKnativeRevision(name, namespace);
                return knativeService;
            } catch (e) {
                return undefined;
            }
        };

        const knativeRevision = await this.waitForObject(
            knativeRevisionExists,
            `wait for KnativeRevision ${name} to be created in namespace ${namespace}`,
            timeout
        );

        return knativeRevision as KnativeRevision;
    }

    async sleep(ms): Promise<void> {
        return new Promise((resolve) => {
            setTimeout(resolve, ms);
        });
    }

    async waitForNoException(fn, description: string, time = 120, delay = 1): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    await fn();
                    return true;
                } catch (e) {
                    return false;
                }
            },
            description,
            true,
            time,
            delay
        );
    }

    async waitToLoadNewPageOrSaveChanges(): Promise<void> {
        await this.sleep(5_000);
    }

    async waitForInteractiveWorkloadToExist(
        workloadName: string,
        namespace: string,
        shouldExist = true,
        timeout = 70
    ): Promise<void> {
        const isInteractiveWorkloadExists = async (): Promise<boolean> => {
            try {
                const interactiveWorkload = await kubeCli.getInteractiveWorkload(workloadName, namespace);
                return !!interactiveWorkload;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            isInteractiveWorkloadExists,
            `wait for interactive workload ${workloadName} to be ${shouldExist ? 'created' : 'deleted'} in namespace ${namespace}`,
            shouldExist,
            timeout
        );
    }

    async waitForTrainingWorkloadToExist(
        workloadName: string,
        namespace: string,
        shouldExist = true,
        timeout = 70
    ): Promise<void> {
        const isTrainingWorkloadExists = async (): Promise<boolean> => {
            try {
                const trainingWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(workloadName, namespace);
                return !!trainingWorkload;
            } catch (e) {
                return false;
            }
        };

        await this.waitForCondition(
            isTrainingWorkloadExists,
            `wait for training workload ${workloadName} to be ${shouldExist ? 'created' : 'deleted'} in namespace ${namespace}`,
            shouldExist,
            timeout
        );
    }

    async waitForBatchJobToExist(
        workloadName: string,
        namespace: string,
        shouldExist = true,
        timeout = 70
    ): Promise<void> {
        const doesJobExist = async (): Promise<boolean> => {
            try {
                await kubeCli.getBatchJob(workloadName, namespace);
                return true;
            } catch {
                return false;
            }
        };

        await this.waitForCondition(
            doesJobExist,
            `wait for batch job ${workloadName} to be ${shouldExist ? 'created' : 'deleted'} in namespace ${namespace}`,
            shouldExist,
            timeout
        );
    }

    async waitForWorkloadPodNotExist(namespace: string, workloadName: string, timeout = 60): Promise<void> {
        await this.waitForWorkloadPodsNotExist(namespace, [workloadName], timeout);
    }

    async waitForWorkloadPodsNotExist(
        namespace: string,
        workloadNames: string[],
        timeout = 60
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const pods = await kubeCli.listPodsByNamepsace(namespace);
                return !pods.some((pod) =>
                    workloadNames.some((workloadName) => pod.metadata.name.includes(workloadName))
                );
            },
            `wait for pods of workloads ${workloadNames.join(', ')} to be deleted from namespace ${namespace}`,
            true,
            timeout
        );
    }

    async waitForRunaiExecNonEmptyOutput(
        workloadName: string,
        command: string,
        workloadType: cliV2WorkloadType = cliV2WorkloadType.training,
        timeout = 30,
        delay = 3
    ): Promise<string> {
        let stdout = '';
        await this.waitForCondition(
            () => {
                stdout = runaiCliV2.exec({ [execFlags.command]: command }, workloadName, workloadType).stdout;
                return stdout.trim() !== '';
            },
            `runai exec "${command}" in workload ${workloadName} to print output`,
            true,
            timeout,
            delay
        );
        return stdout;
    }

    async waitForElementToBeEnabled(
        page: BaseV2Page,
        selector: string,
        text: string,
        timeout = 30
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                const element = page.locator(selector).getByText(text, { exact: true });
                return (await element.isVisible()) && !(await element.getAttribute('disabled'));
            },
            `wait for element ${text} to be enabled`,
            true,
            timeout
        );
    }

    private positiveMeasurementValueExists(
        measurements: Array<MeasurementResponse>,
        nodeMetricType: NodeMetricType | WorkloadMetricType
    ): boolean {
        let metricValues: GroupedMeasurementResponseValuesInner[] = [];
        for (const measurement of measurements) {
            if (measurement.type === String(nodeMetricType)) {
                metricValues = measurement.values;
                if (metricValues && metricValues.some((v) => parseFloat(v.value) > 0)) {
                    return true;
                }
            }
        }
        return false;
    }

    async waitForPositiveNodeMetricsValue(nodeId: string, nodeMetricType: NodeMetricType): Promise<void> {
        const now = new Date();
        const start = new Date(now.getTime() - 10 * 60 * 1000);

        await this.waitForCondition(
            async () => {
                const end = new Date();
                const resp = await nodeUtils.getNodeMetrics(nodeId, [nodeMetricType], start, end);
                return this.positiveMeasurementValueExists(resp.body.measurements, nodeMetricType);
            },
            `expected ${nodeMetricType} to be defined`,
            true,
            300,
            5
        );
    }

    async waitForPositiveWorkloadMetricsValue(
        workloadId: string,
        workloadMetricType: WorkloadMetricType
    ): Promise<void> {
        const now = new Date();
        const start = new Date(now.getTime() - 10 * 60 * 1000);

        await this.waitForCondition(
            async () => {
                const end = new Date();
                const resp = await metricUtils.getWorkloadMetrics(
                    workloadId,
                    [workloadMetricType],
                    start,
                    end
                );
                return this.positiveMeasurementValueExists(resp.body.measurements, workloadMetricType);
            },
            `expected ${workloadMetricType} to be defined`,
            true,
            300,
            5
        );
    }

    async waitForIdleAllocatedGpuDevicesAppear(timeout: number = 600): Promise<void> {
        await this.waitForCondition(
            async () => {
                const allocatedIdleGpuDevices = await telemetryUtils.getNodeTelemetryFilteredByClusterId(
                    flowRootState.cluster.uuid,
                    NodeTelemetryType.IdleAllocatedGpus
                );
                const value = allocatedIdleGpuDevices?.values?.[0]?.value;
                return value !== undefined ? parseFloat(value) > 0 : false;
            },
            'waiting for idle allocated GPU devices to appear',
            true,
            timeout,
            5
        );
    }

    async waitForAiApplicationToAppear(appName: string, timeout = 60, interval = 3): Promise<string> {
        let appId = '';
        await this.waitForCondition(
            async () => {
                try {
                    const response = await workloadsUtilsV2.listApplications();
                    const appNames = response.body.applications?.map((a: any) => a.name) ?? [];
                    logger.info(`Listed AI applications: [${appNames.join(', ')}], looking for: ${appName}`);
                    const app = response.body.applications?.find((a: any) => a.name === appName);
                    if (app) {
                        appId = app.id;
                        return true;
                    }
                    return false;
                } catch {
                    return false;
                }
            },
            `Wait for AI application ${appName} to appear in API`,
            true,
            timeout,
            interval
        );
        return appId;
    }

    async waitForApplicationCRStatus(
        appName: string,
        namespace: string,
        expectedActionPhase: string,
        expectedHelmPhase?: string,
        timeout = 120,
        interval = 5,
        expectedActionType?: string
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                    const actionType = (cr as any)?.status?.action?.type;
                    const actionPhase = (cr as any)?.status?.action?.phase;
                    const helmPhase = (cr as any)?.status?.helmStatus?.helmPhase;
                    logger.info(
                        `Application CR ${appName}: actionType=${actionType}, actionPhase=${actionPhase}, helmPhase=${helmPhase}`
                    );

                    if (
                        expectedActionType &&
                        actionType?.toLowerCase() !== expectedActionType.toLowerCase()
                    ) {
                        return false;
                    }

                    const actionMatch = actionPhase?.toLowerCase() === expectedActionPhase.toLowerCase();

                    if (expectedHelmPhase) {
                        return actionMatch && helmPhase?.toLowerCase() === expectedHelmPhase.toLowerCase();
                    }
                    // If no expectedHelmPhase provided, expect helmPhase to not exist
                    return actionMatch && (helmPhase === undefined || helmPhase === null);
                } catch {
                    return false;
                }
            },
            `Wait for application CR ${appName} actionType=${expectedActionType ?? 'any'}, actionPhase=${expectedActionPhase}, helmPhase=${expectedHelmPhase ?? 'undefined'}`,
            true,
            timeout,
            interval
        );
    }

    async waitForApplicationCRK8sPhase(
        appName: string,
        namespace: string,
        expectedPhase: string,
        timeout = 60,
        interval = 5
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                    const k8sPhase = (cr as any)?.status?.k8sPhase;
                    logger.info(`Application CR ${appName}: k8sPhase=${k8sPhase}, expected=${expectedPhase}`);
                    return k8sPhase?.toLowerCase() === expectedPhase.toLowerCase();
                } catch {
                    return false;
                }
            },
            `Wait for application CR ${appName} k8sPhase=${expectedPhase}`,
            true,
            timeout,
            interval
        );
    }

    async waitForApplicationCrdToAppear(
        appName: string,
        namespace: string,
        timeout = 60,
        interval = 3
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                    logger.info(
                        `Application CR ${appName} found in namespace ${namespace}, status: ${JSON.stringify((cr as any)?.status?.action?.phase ?? 'no status')}`
                    );
                    return true;
                } catch {
                    logger.info(`Application CR ${appName} not yet found in namespace ${namespace}`);
                    return false;
                }
            },
            `Wait for application CR ${appName} in namespace ${namespace}`,
            true,
            timeout,
            interval
        );
    }

    async waitForAiApplicationPhase(
        applicationId: string,
        expectedPhase: string,
        timeout = 120,
        interval = 5
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    const resp = await aiApplicationUtils.getAiApplication(applicationId);
                    const phase = resp.body.phase;
                    logger.info(`AI application ${applicationId}: phase=${phase}, expected=${expectedPhase}`);
                    return phase?.toLowerCase() === expectedPhase.toLowerCase();
                } catch {
                    return false;
                }
            },
            `Wait for AI application ${applicationId} phase to become ${expectedPhase}`,
            true,
            timeout,
            interval
        );
    }

    async waitForAiApplicationToBeRemoved(
        applicationName: string,
        timeout = 60,
        interval = 5
    ): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    const listResp = await aiApplicationUtils.listAiApplications();
                    const appNames = listResp.body.applications.map((a: any) => a.name);
                    const stillExists = appNames.includes(applicationName);
                    logger.info(
                        `AI applications list: [${appNames.join(', ')}], ${applicationName} still exists: ${stillExists}`
                    );
                    return !stillExists;
                } catch {
                    return false;
                }
            },
            `Wait for ${applicationName} to be removed from API list`,
            true,
            timeout,
            interval
        );
    }

    async waitForNamespaceToExist(namespace: string, timeout = 90): Promise<void> {
        await this.waitForCondition(
            async () => {
                try {
                    await kubeCli.CoreV1Api.readNamespace({ name: namespace });
                    return true;
                } catch {
                    return false;
                }
            },
            `namespace '${namespace}' to exist in cluster`,
            true,
            timeout
        );
    }
}
