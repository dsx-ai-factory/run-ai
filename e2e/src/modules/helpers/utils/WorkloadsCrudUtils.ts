/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadsCrudService } from '../../api';
import http from 'http';
import { Training } from '../../api/apis/workloadsApi/generated/model/training';
import { TrainingSpecSpec } from '../../api/apis/workloadsApi/generated/model/trainingSpecSpec';
import { TrainingCreationRequest } from '../../api/apis/workloadsApi/generated/model/trainingCreationRequest';
import { Workspace } from '../../api/apis/workloadsApi/generated/model/workspace';
import { WorkspaceSpecSpec } from '../../api/apis/workloadsApi/generated/model/workspaceSpecSpec';
import { WorkspaceCreationRequest } from '../../api/apis/workloadsApi/generated/model/workspaceCreationRequest';
import { Inference } from '../../api/apis/workloadsApi/generated/model/inference';
import { InferenceSpecSpec } from '../../api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { InferenceCreationRequest } from '../../api/apis/workloadsApi/generated/model/inferenceCreationRequest';
import { DistributedWorkload } from '../../api/apis/workloadsApi/generated/model/distributedWorkload';
import { DistributedSpecSpec } from '../../api/apis/workloadsApi/generated/model/distributedSpecSpec';
import { DistributedCreationRequest } from '../../api/apis/workloadsApi/generated/model/distributedCreationRequest';
import { flowRootState } from '../config/config';
import { InferenceUpdateRequest } from '../../api/apis/workloadsApi/generated/model/inferenceUpdateRequest';
import { InferenceUpdateSpecSpec } from '../../api/apis/workloadsApi/generated/model/inferenceUpdateSpecSpec';
import { HttpResponse, MasterSpec } from '../../api/apis/workloadsApi/generated/api';
import {
    DistributedInference,
    DistributedInferenceCreationRequest,
    DistributedInferenceSpecSpec
} from '../../api/apis/workloadsApi/generated/api';
import { waitUtils, workloadsUtilsV2, logger } from '..';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { UpdateRequest } from '../../api/apis/workloadsApi/generated/model/updateRequest';
import { kubeCli } from '../../cli';

export enum WorkloadType {
    WORKSPACE = 'workspace',
    TRAINING = 'training',
    INFERENCE = 'inference',
    DISTRIBUTED = 'distributed',
    DISTRIBUTED_INFERENCE = 'distributed-inference'
}

export class WorkloadsCrudApiUtils {
    basePath: string;

    set accessToken(accessToken: string | (() => string)) {
        workloadsCrudService.workspacesApi.accessToken = accessToken;
        workloadsCrudService.trainingsApi.accessToken = accessToken;
        workloadsCrudService.inferencesApi.accessToken = accessToken;
        workloadsCrudService.distributedInferencesApi.accessToken = accessToken;
        workloadsCrudService.distributedApi.accessToken = accessToken;
    }

    async createTraining(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: TrainingSpecSpec,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: Training }> {
        const request: TrainingCreationRequest = {
            name: name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId: projectId,
            clusterId: clusterId,
            spec: spec
        };
        workloadsCrudService.trainingsApi.accessToken = accessToken;
        return workloadsCrudService.trainingsApi.createTraining(request);
    }

    async createWorkspace(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: WorkspaceSpecSpec,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: Workspace }> {
        const request: WorkspaceCreationRequest = {
            name: name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId: projectId,
            clusterId: clusterId,
            spec: spec
        };
        workloadsCrudService.workspacesApi.accessToken = accessToken;
        return workloadsCrudService.workspacesApi.createWorkspace(request);
    }

    async createInference(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: InferenceSpecSpec,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: Inference }> {
        const request: InferenceCreationRequest = {
            name: name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId: projectId,
            clusterId: clusterId,
            spec: spec
        };
        workloadsCrudService.inferencesApi.accessToken = accessToken;
        return workloadsCrudService.inferencesApi.createInference(request);
    }

    // Anyworkload submissions are persisted as WorkloadRunner CRDs and synced back under a different
    // id, so the workspaces/inferences read model 404s on the submission id - wait on the pods instead.
    async createWorkspaceAndWaitForPodsRunning(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: WorkspaceSpecSpec,
        namespace: string,
        accessToken?,
        timeoutSeconds = 180
    ): Promise<Workspace> {
        const { body } = await this.createWorkspace(
            name,
            addSuffixToName,
            clusterId,
            projectId,
            spec,
            accessToken
        );
        await kubeCli.waitForJobPodsRunning(namespace, name, timeoutSeconds);
        return body;
    }

    async createDistributedInference(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: DistributedInferenceSpecSpec,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: DistributedInference }> {
        const request: DistributedInferenceCreationRequest = {
            name: name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId: projectId,
            clusterId: clusterId,
            spec: spec
        };
        workloadsCrudService.distributedInferencesApi.accessToken = accessToken;
        return workloadsCrudService.distributedInferencesApi.createDistributedInference(request);
    }

    async createDistributed(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: DistributedSpecSpec,
        masterSpecSameAsWorker = false,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: DistributedWorkload }> {
        const request: DistributedCreationRequest = {
            name: name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId: projectId,
            clusterId: clusterId,
            spec: spec,
            masterSpecSameAsWorker: masterSpecSameAsWorker
        };

        workloadsCrudService.distributedApi.accessToken = accessToken;
        return workloadsCrudService.distributedApi.createDistributed(request);
    }

    async createDistributedV2(
        name: string,
        addSuffixToName: boolean,
        clusterId: string,
        projectId: string,
        spec: DistributedSpecSpec,
        masterSpec?: MasterSpec
    ): Promise<{ response: http.IncomingMessage; body: DistributedWorkload }> {
        const masterSpecSameAsWorker = masterSpec === undefined;

        const request: DistributedCreationRequest = {
            name,
            useGivenNameAsPrefix: addSuffixToName,
            projectId,
            clusterId,
            spec,
            ...(masterSpecSameAsWorker
                ? { masterSpecSameAsWorker: true }
                : { masterSpec, masterSpecSameAsWorker: false })
        };

        return workloadsCrudService.distributedApi.createDistributed(request);
    }

    async getTraining(workloadId: string): Promise<{ response: http.IncomingMessage; body: Training }> {
        return workloadsCrudService.trainingsApi.getTraining(workloadId);
    }

    async getDistributed(
        workloadId: string
    ): Promise<{ response: http.IncomingMessage; body: DistributedWorkload }> {
        return workloadsCrudService.distributedApi.getDistributed(workloadId);
    }

    async getWorkspace(workloadId: string): Promise<{ response: http.IncomingMessage; body: Workspace }> {
        return workloadsCrudService.workspacesApi.getWorkspace(workloadId);
    }

    async getInference(workloadId: string): Promise<{ response: http.IncomingMessage; body: Inference }> {
        return workloadsCrudService.inferencesApi.getInference(workloadId);
    }

    async getDistributedInference(
        workloadId: string
    ): Promise<{ response: http.IncomingMessage; body: DistributedInference }> {
        return workloadsCrudService.distributedInferencesApi.getDistributedInference(workloadId);
    }

    async getDistributedInferenceByName(
        workloadName: string,
        projectId: number | string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeout = 120
    ): Promise<DistributedInference> {
        this.accessToken = accessToken;

        const workload = await waitUtils.waitForObject(
            async (): Promise<DistributedInference> => {
                let workloadId = '';
                try {
                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, projectId);
                } catch {
                    return undefined;
                }
                if (!workloadId) {
                    return undefined;
                }
                return (await this.getDistributedInference(workloadId)).body;
            },
            `wait for workload ${workloadName} to be in control plane`,
            timeout
        );

        return workload as DistributedInference;
    }

    async getInferenceByName(
        workloadName: string,
        projectId: number | string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeout = 120
    ): Promise<Inference> {
        this.accessToken = accessToken;

        const workload = await waitUtils.waitForObject(
            async (): Promise<Inference> => {
                let workloadId = '';
                try {
                    workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, projectId);
                } catch {
                    return undefined;
                }
                if (!workloadId) {
                    return undefined;
                }
                return (await this.getInference(workloadId)).body;
            },
            `wait for workload ${workloadName} to be in control plane`,
            timeout
        );

        return workload as Inference;
    }

    async deleteTraining(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.trainingsApi.deleteTraining(workloadId);
    }

    async deleteWorkspace(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.workspacesApi.deleteWorkspace(workloadId);
    }

    async deleteInference(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.inferencesApi.deleteInference(workloadId);
    }

    async deleteDistributedInference(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.distributedInferencesApi.deleteDistributedInference(workloadId);
    }

    async deleteDistributed(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.distributedApi.deleteDistributed(workloadId);
    }

    async deleteWorkload(
        workloadType: WorkloadType,
        workloadId: string
    ): Promise<{ response: http.IncomingMessage }> {
        switch (workloadType) {
            case WorkloadType.TRAINING:
                return await this.deleteTraining(workloadId);
            case WorkloadType.WORKSPACE:
                return await this.deleteWorkspace(workloadId);
            case WorkloadType.DISTRIBUTED:
                return await this.deleteDistributed(workloadId);
            case WorkloadType.INFERENCE:
                return await this.deleteInference(workloadId);
            case WorkloadType.DISTRIBUTED_INFERENCE:
                return await this.deleteDistributedInference(workloadId);
            default:
                throw new Error(`not recognized workload type: ${workloadType}`);
        }
    }

    async suspendTraining(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.trainingsApi.suspendTraining(workloadId);
    }

    async suspendWorkspace(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.workspacesApi.suspendWorkspace(workloadId);
    }

    async suspendDistributedTraining(
        workloadId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: HttpResponse }> {
        workloadsCrudService.distributedApi.accessToken = accessToken;
        return workloadsCrudService.distributedApi.suspendDistributed(workloadId);
    }

    async resumeTraining(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.trainingsApi.resumeTraining(workloadId);
    }

    async resumeWorkspace(workloadId: string): Promise<{ response: http.IncomingMessage }> {
        return workloadsCrudService.workspacesApi.resumeWorkspace(workloadId);
    }

    async resumeDistributedTraining(
        workloadId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: HttpResponse }> {
        workloadsCrudService.distributedApi.accessToken = accessToken;
        return workloadsCrudService.distributedApi.resumeDistributed(workloadId);
    }

    async waitForWorkloadPhase(
        workloadId: string,
        expectedPhase: Phase,
        workloadType: WorkloadType = WorkloadType.WORKSPACE,
        invalidPhase?: Phase,
        timeout = 60
    ): Promise<void> {
        let lastPhase: Phase | undefined = undefined;
        let lastResponse: any = undefined;

        // Select the appropriate getter method based on workload type
        const getWorkloadMethod = {
            [WorkloadType.WORKSPACE]: this.getWorkspace.bind(this),
            [WorkloadType.TRAINING]: this.getTraining.bind(this),
            [WorkloadType.INFERENCE]: this.getInference.bind(this),
            [WorkloadType.DISTRIBUTED]: this.getDistributed.bind(this),
            [WorkloadType.DISTRIBUTED_INFERENCE]: this.getDistributedInference.bind(this)
        }[workloadType];

        try {
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        const res = await getWorkloadMethod(workloadId);
                        lastResponse = res;
                        const currentPhase = res.body.actualPhase;
                        lastPhase = currentPhase;
                        logger.debug(
                            `[waitForWorkloadPhase] workload=${workloadId}, type=${workloadType}, expected=${expectedPhase}, current=${currentPhase}, result=${currentPhase === expectedPhase}`
                        );
                        if (invalidPhase && currentPhase === invalidPhase) {
                            logger.error(`[waitForWorkloadPhase] Invalid phase detected: ${currentPhase}`);
                            throw new Error(`Invalid phase detected: ${currentPhase}`);
                        }
                        return currentPhase === expectedPhase;
                    } catch (err) {
                        logger.debug(`[waitForWorkloadPhase] get${workloadType} failed: ${err.message}`);
                        if (err.message && err.message.includes('workload not found')) {
                            return false;
                        }
                        throw err;
                    }
                },
                `wait for ${workloadType} to be in ${expectedPhase} state`,
                true,
                timeout,
                1
            );
        } catch (err) {
            logger.error(
                `[waitForWorkloadPhase] Timeout! workload=${workloadId}, type=${workloadType}, expected=${expectedPhase}, lastPhase=${lastPhase}, lastResponse=${JSON.stringify(lastResponse)}`
            );
            throw new Error(`[waitForWorkloadPhase] Phase missing:[${expectedPhase}], ${err.message}`);
        }
    }

    async updateInferenceSpec(
        workloadId: string,
        spec: InferenceUpdateSpecSpec
    ): Promise<{ response: http.IncomingMessage; body: Inference }> {
        const request: InferenceUpdateRequest = {
            spec: spec
        };

        return workloadsCrudService.inferencesApi.updateInferenceSpec(workloadId, request);
    }

    async updateDistributedInferenceSpec(
        workloadId: string,
        updateRequest: UpdateRequest
    ): Promise<{ response: http.IncomingMessage; body: DistributedInference }> {
        return workloadsCrudService.distributedInferencesApi.updateDistributedInferenceSpec(
            workloadId,
            updateRequest
        );
    }
}
