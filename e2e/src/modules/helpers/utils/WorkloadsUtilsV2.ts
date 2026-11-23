/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadsService } from '../../api';
import { flowRootState } from '../config/config';
import http from 'http';
import { GetWorkloads200Response } from '../../api/apis/workloadsApi/generated/model/getWorkloads200Response';
import { Workload } from '../../api/apis/workloadsApi/generated/model/workload';
import {
    BaseWorkload,
    CustomResourceDefinition,
    logger,
    PodStatus,
    versionUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../index';
import { kubeCli } from '../../cli';
import { ListPods200Response } from '../../api/apis/workloadsApi/generated/model/listPods200Response';
import { GetWorkloadEvents200Response } from '../../api/apis/workloadsApi/generated/model/getWorkloadEvents200Response';
import { expect } from 'chai';
import { WorkloadRequestResources } from '../../api/apis/workloadsApi/generated/model/workloadRequestResources';
import { Project } from '../../api/apis/projectsDepartmentsApi';
import { GetWorkloadHistory200Response } from '../../api/apis/workloadsApi/generated/model/getWorkloadHistory200Response';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { WorkloadDetailed } from '../../api/apis/workloadsApi/generated/model/workloadDetailed';
import { PhaseReason } from '../../api/apis/workloadsApi/generated/model/phaseReason';
import { CountWorkloads200Response } from '../../api/apis/workloadsApi/generated/model/countWorkloads200Response';
import type { WorkloadTemplate } from '../../api/apis/assetsApi';
import { AIApplication } from '../../api/apis/workloadsApi/generated/model/aIApplication';
import { WorkloadElementsResponse } from '../../api/apis/workloadsApi/generated/model/workloadElementsResponse';
import { CountWorkloadElements200Response } from '../../api/apis/workloadsApi/generated/model/countWorkloadElements200Response';
import { GetWorkloadEndpoints200Response } from '../../api/apis/workloadsApi/generated/model/getWorkloadEndpoints200Response';
import { AIApplicationVerbosity } from '../../api/apis/workloadsApi/generated/model/aIApplicationVerbosity';

export enum WorkloadType {
    WORKSPACE = 'workspace',
    INFERENCE = 'inference',
    DISTRIBUTED_INFERENCE = 'distributedinference',
    TRAINING = 'training',
    DISTRIBUTED = 'distributed'
}

// The API rejects limit > 500, and its default of 50 silently truncates a busy cluster.
const MAX_WORKLOADS_PAGE_SIZE = 500;

const WORKLOAD_STATUS_POLL_DELAY_SECONDS = 2;

export class WorkloadsUtilsV2 {
    async getWorkloadByWorkloadId(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadDetailed }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkload(workloadId);
    }

    async waitForWorkloadImages(workloadId: string, timeout = 60): Promise<WorkloadDetailed> {
        let workload: WorkloadDetailed;
        await waitUtils.waitForCondition(
            async () => {
                workload = (await this.getWorkloadByWorkloadId(workloadId)).body;
                return (workload?.images?.length ?? 0) > 0;
            },
            `wait for workload ${workloadId} images to be reported`,
            true,
            timeout,
            5
        );
        return workload;
    }

    async waitForWorkloadsPhase(workloadsIds: string[], expectedPhase: string, timeout = 90): Promise<void> {
        const getAllWorkloadPhasesMatch = async (): Promise<boolean> => {
            try {
                const doesPhasesMatchExpectedPhase = await Promise.all(
                    workloadsIds.map(async (id) => {
                        const phase = (await this.getWorkloadByWorkloadId(id)).body.phase;
                        return phase.toString() === expectedPhase;
                    })
                );
                return doesPhasesMatchExpectedPhase.every(Boolean);
            } catch (error: unknown) {
                logger.error(`Error while fetching workload status: ${(error as Error).message}`);
                return false;
            }
        };

        logger.info(`Waiting for workloads phase to be: ${expectedPhase}`);
        await waitUtils.waitForCondition(
            getAllWorkloadPhasesMatch,
            `Wait for all workloads to reach phase: ${expectedPhase}`,
            true,
            timeout,
            3
        );
    }

    async getWorkloadPods(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: ListPods200Response }> {
        workloadsService.podsApi.accessToken = accessToken;
        return await workloadsService.podsApi.getWorkloadPods(workloadId);
    }

    async getWorkloadEvents(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloadEvents200Response }> {
        workloadsService.eventsApi.accessToken = accessToken;
        return await workloadsService.eventsApi.getWorkloadEvents(workloadId);
    }

    async getWorkloadHistory(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloadHistory200Response }> {
        workloadsService.eventsApi.accessToken = accessToken;
        return await workloadsService.eventsApi.getWorkloadHistory(workloadId);
    }

    async getWorkloadIdByName(workloadName: string, projectId: number | string): Promise<string> {
        const projectIdString = projectId.toString();
        const allWorkloads = (await this.getWorkloadByName(workloadName, projectIdString)).body.workloads;
        const workload: Workload | undefined = allWorkloads.find((w) => w.name === workloadName);
        if (!workload) {
            throw new Error(`Workload ${workloadName} not found in project ${projectIdString}`);
        }
        return workload.id;
    }

    async getWorkloadByName(
        workloadName: string,
        projectId: number | string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        limit = 200
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            undefined,
            undefined,
            limit,
            undefined,
            undefined,
            ['projectId==' + projectId, 'name==' + workloadName]
        );
    }

    // '=@' (contains), not the '==' used by getWorkloadByName: callers pass partial names.
    async getWorkloadsByNameContainsAndProjectId(
        workloadName: string,
        projectId: number | string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        limit = 200
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            undefined,
            undefined,
            limit,
            undefined,
            undefined,
            ['projectId==' + projectId, 'name=@' + workloadName]
        );
    }

    async getAllWorkloads(
        accessToken = flowRootState.users.system_administrator.accessToken,
        filters: string[] = []
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            filters
        );
    }

    async getWorkloadsByProjectId(
        projectId: string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        limit = undefined
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            undefined,
            undefined,
            limit,
            undefined,
            undefined,
            ['projectId==' + projectId]
        );
    }

    async getWorkloadsByClusterId(
        clusterId: string,
        accessToken = flowRootState.users.system_administrator.accessToken,
        limit = undefined
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            undefined,
            undefined,
            limit,
            undefined,
            undefined,
            ['clusterId==' + clusterId]
        );
    }

    async countWorkloadsByClusterId(
        clusterId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: CountWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.countWorkloads(undefined, ['clusterId==' + clusterId]);
    }

    async getAllDeletedWorkloads(
        projectId,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloads(
            true,
            undefined,
            undefined,
            undefined,
            undefined,
            ['projectId==' + projectId.toString()]
        );
    }

    async waitForDeletedWorkload(
        workloadId,
        projectId,
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeout = 30
    ): Promise<Workload> {
        const getWorkload = async (): Promise<Workload> => {
            try {
                const workload = (
                    await workloadsUtilsV2.getAllDeletedWorkloads(projectId, accessToken)
                ).body.workloads.find((w) => w.id === workloadId);
                return workload;
            } catch (error) {
                return undefined;
            }
        };

        const workload = await waitUtils.waitForObject(
            getWorkload,
            `wait for workload ${workloadId} to be in control plane`,
            timeout
        );

        return workload as Workload;
    }

    async getWorkloadByNameAndProjectName(
        workloadName: string,
        projectName: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Workload> {
        const workloads = (
            await this.getAllWorkloads(accessToken, [`name==${workloadName}`, `projectName==${projectName}`])
        ).body.workloads;
        return workloads.length > 0 ? workloads[0] : undefined;
    }

    async getWorkloadByNameContainAndProjectNameEq(
        workloadName: string,
        projectName: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Workload> {
        const workloads = (
            await this.getAllWorkloads(accessToken, [`projectName==${projectName},name=@${workloadName}`])
        ).body.workloads;
        return workloads.length > 0 ? workloads[0] : undefined;
    }

    async getWorkloadPodsByNameAndProjectName(
        workloadName: string,
        projectName: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: ListPods200Response }> {
        const workload = await this.getWorkloadByNameAndProjectName(workloadName, projectName, accessToken);
        return await this.getWorkloadPods(workload.id, accessToken);
    }

    async waitForWorkloadPodsByNameAndProjectName(
        workloadName: string,
        projectName: string,
        timeout = 30
    ): Promise<void> {
        const getWorkloadPods = async (): Promise<boolean> => {
            try {
                return (
                    (await this.getWorkloadPodsByNameAndProjectName(workloadName, projectName))?.body?.pods
                        ?.length > 0
                );
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            getWorkloadPods,
            `wait for workload ${workloadName} pods to be in control plane`,
            true,
            timeout
        );
    }

    async waitForBackendWorkloadByNameAndProjectName(
        workloadName: string,
        projectName: string,
        timeoutInSeconds = 60,
        delay = 2,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Workload> {
        let backendWorkload;
        const isBackWorkloadFound = async (): Promise<boolean> => {
            try {
                backendWorkload = await this.getWorkloadByNameAndProjectName(
                    workloadName,
                    projectName,
                    accessToken
                );
                return backendWorkload !== undefined;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackWorkloadFound,
            `waitForBackendWorkloadByNameAndProjectName workloadName: ${workloadName}`,
            true,
            timeoutInSeconds,
            delay
        );

        return backendWorkload;
    }

    async waitForBackendWorkloadPodsByNameAndProject(
        workloadName: string,
        project: Project,
        timeoutInSeconds = 60,
        delay = 2,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        let backendWorkloadPods;
        let podsProjectId;
        const projectId = project.id.toString();
        const projectName = project.name;

        const isBackWorkloadPodsFound = async (): Promise<boolean> => {
            try {
                backendWorkloadPods = (
                    await this.getWorkloadPodsByNameAndProjectName(workloadName, projectName, accessToken)
                ).body.pods;

                for (const pod of backendWorkloadPods) {
                    podsProjectId = pod.projectId;
                    if (podsProjectId !== projectId) return false;
                }
                return true;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackWorkloadPodsFound,
            `waitForBackendWorkloadPodsByNameAndProject workloadName: ${workloadName}`,
            true,
            timeoutInSeconds,
            delay
        );
    }

    async waitForBackendWorkloadAndPodsWithCorrectProjectAndNamespace(
        workloadName: string,
        expectedProject: Project,
        expectedNamespace: string,
        timeout = 60
    ): Promise<void> {
        if (!versionUtils.isVersionSupported(flowRootState.cluster.version, '2.16')) {
            logger.debug('Skipping waitForBackendWorkloadAndPodsWithCorrectProjectAndNamespace');
            return;
        }
        const promisesToAwait = [];
        promisesToAwait.push(
            this.waitForBackendWorkloadByNameAndProjectName(workloadName, expectedProject.name, timeout),
            this.waitForBackendWorkloadPodsByNameAndProject(workloadName, expectedProject, timeout)
        );
        await Promise.all(promisesToAwait);
    }

    async getWorkloadIdByWorkloadNameAndNamespace(
        workloadType: CustomResourceDefinition,
        workloadName: string,
        namespace: string
    ): Promise<string> {
        const workloadCrd = await kubeCli.getCrd(workloadType, workloadName, namespace);
        const condition =
            workloadCrd['metadata']['annotations'] === undefined ||
            workloadCrd['metadata']['annotations']['workloadId'] === undefined;
        const workloadId = condition
            ? workloadCrd['metadata']['uid']
            : workloadCrd['metadata']['annotations']['workloadId'];

        logger.info(`WorkloadId of ${workloadName} is ${workloadId}`);
        return workloadId;
    }

    async getWorkloadEndpoints(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloadEndpoints200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloadEndpoints(workloadId);
    }

    /**
     * @summary Get workload by workload id and validate optional parameters
     * @param workloadId - workload id
     * @param options - optional parameters
     * @param options.expectedWorkloadRequestedResources - expected workload requested resources
     * @param options.expectedPhase - expected workload phase
     * @param options.expectedRequestedPodsMin - expected requested pods min (for inference workload for example)
     * @param options.expectedRequestedPodsMax - expected requested pods max (for inference workload for example)
     * @param options.expectedType - expected workload type
     */

    async getWorkloadAndValidate(
        workloadId: string,
        options: {
            expectedWorkloadRequestedResources?: WorkloadRequestResources;
            expectedPhase?: Phase;
            expectedRequestedPodsMin?: number;
            expectedRequestedPodsMax?: number;
            expectedType?: string;
            expectedSource?: string;
            expectedRequestedPods?: number;
        }
    ): Promise<{ response: http.IncomingMessage; body: Workload }> {
        const {
            expectedWorkloadRequestedResources,
            expectedPhase,
            expectedRequestedPodsMin,
            expectedRequestedPodsMax,
            expectedType,
            expectedSource,
            expectedRequestedPods
        } = options;
        logger.info(`Validating workload ${workloadId}`);
        const workload = await this.getWorkloadByWorkloadId(workloadId);
        expect(workload.response.statusCode).to.equal(200);
        if (expectedWorkloadRequestedResources) {
            for (const [key] of Object.entries(expectedWorkloadRequestedResources)) {
                if (key !== undefined) {
                    // if not a nested object
                    if (typeof expectedWorkloadRequestedResources[key] !== 'object') {
                        expect(workload.body.workloadRequestedResources[key]).to.equal(
                            expectedWorkloadRequestedResources[key]
                        );
                    } else {
                        // if nested object
                        for (const [nestedKey] of Object.entries(expectedWorkloadRequestedResources[key])) {
                            if (expectedWorkloadRequestedResources[key] !== undefined) {
                                expect(workload.body.workloadRequestedResources[key][nestedKey]).to.equal(
                                    expectedWorkloadRequestedResources[key][nestedKey]
                                );
                            }
                        }
                    }
                }
            }
        }
        if (expectedPhase) {
            expect(workload.body.phase).to.equal(expectedPhase);
        }
        if (expectedRequestedPodsMin) {
            expect(workload.body.requestedPods.min).to.equal(expectedRequestedPodsMin);
        }
        if (expectedRequestedPodsMax) {
            expect(workload.body.requestedPods.max).to.equal(expectedRequestedPodsMax);
        }
        if (expectedType) {
            expect(workload.body.type).to.equal(expectedType);
        }
        if (expectedSource) {
            expect(workload.body.source).to.equal(expectedSource);
        }
        if (expectedRequestedPods) {
            expect(workload.body.requestedPods.number).to.equal(expectedRequestedPods);
        }
        return workload;
    }

    async waitForWorkloadInControlPlane(workloadId: string, timeout = 30): Promise<void> {
        const getWorkload = async (): Promise<boolean> => {
            try {
                const workload = (await this.getWorkloadByWorkloadId(workloadId)).body;
                return workload.id === workloadId;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            getWorkload,
            `wait for workload ${workloadId} to be in control plane`,
            true,
            timeout
        );
    }

    async waitForWorkloadWithEndpoints(workloadId: string, timeout = 45): Promise<void> {
        const getWorkload = async (): Promise<boolean> => {
            try {
                const workload = (await this.getWorkloadByWorkloadId(workloadId)).body;
                return workload.id === workloadId && !!workload.endpoints;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            getWorkload,
            `wait for workload ${workloadId} to be in control plane and have endpoints`,
            true,
            timeout
        );
    }

    async waitForWorkloadWithEndpointUrl(
        workloadId: string,
        expectedUrl: string,
        timeout = 45
    ): Promise<void> {
        const hasMatchingEndpoint = async (): Promise<boolean> => {
            try {
                const workload = (await this.getWorkloadByWorkloadId(workloadId)).body;
                return (
                    workload.id === workloadId &&
                    workload.endpoints?.some(
                        (ep: { url: string }) => ep.url.replace(/\/$/, '') === expectedUrl
                    )
                );
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            hasMatchingEndpoint,
            `wait for workload ${workloadId} to have endpoint with url ${expectedUrl}`,
            true,
            timeout
        );
    }

    async waitForWorkloadIdByWorkloadNameAndProjectId(
        workloadName: string,
        projectId: string,
        timeout = 50
    ): Promise<string> {
        let workloadId: string;
        const getWorkloadId = async (): Promise<boolean> => {
            try {
                // Keep the id this lookup already resolved. Re-reading it after the wait means a
                // second, independent list call that can still miss the workload and throw, even
                // though we just proved it is there.
                workloadId = await this.getWorkloadIdByName(workloadName, projectId);
                return workloadId !== undefined;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            getWorkloadId,
            `wait for workload ${workloadName} to be in control plane`,
            true,
            timeout
        );
        return workloadId;
    }

    async waitForOneWorkloadStatus(
        workloads: BaseWorkload[],
        status: PodStatus,
        timeout = 30
    ): Promise<void> {
        const getWorkload = async (): Promise<boolean> => {
            try {
                for (const workload of workloads) {
                    if (await workload.validateStatus(status)) {
                        return true;
                    }
                }
                return false;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            getWorkload,
            `wait for one workload to be in status ${status}`,
            true,
            timeout
        );
    }

    async getNumberOfWorkloadsInPhase(
        phase: Phase,
        clusterId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<number> {
        const allWorkloads = (await this.getWorkloadsByClusterId(clusterId, accessToken)).body.workloads;
        return allWorkloads.filter((w) => w.phase === phase).length;
    }

    async waitForApiWorkloadWithPhase(
        workloadId: string,
        expectedPhase: Phase,
        workloadName: string,
        desiredResult = true,
        timeout = 60,
        delay = 1,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadDetailed> {
        let apiWorkload: WorkloadDetailed;

        const doesWorkloadHaveExpectedPhase = async (): Promise<boolean> => {
            try {
                apiWorkload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId, accessToken)).body;
                logger.debug(`Workload ${workloadName} has phase ${apiWorkload.phase}`);
                return apiWorkload.phase === expectedPhase;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesWorkloadHaveExpectedPhase,
            `wait for workload ${workloadName} to have phase ${expectedPhase}`,
            desiredResult,
            timeout,
            delay
        );

        return apiWorkload;
    }

    async waitForApiWorkloadWithPhaseReason(
        workloadId: string,
        expectedPhaseReason: PhaseReason | null,
        workloadName: string,
        desiredResult = true,
        timeout = 60,
        delay = 1,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadDetailed> {
        let apiWorkload: WorkloadDetailed;

        const doesWorkloadHaveExpectedPhaseReason = async (): Promise<boolean> => {
            try {
                apiWorkload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId, accessToken)).body;
                logger.debug(`Workload ${workloadName} has phase reason ${apiWorkload.phaseReason}`);
                if (expectedPhaseReason) {
                    return apiWorkload.phaseReason === expectedPhaseReason;
                }
                return apiWorkload.phaseReason === undefined;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesWorkloadHaveExpectedPhaseReason,
            `wait for workload ${workloadName} to have phase reason ${expectedPhaseReason}`,
            desiredResult,
            timeout,
            delay
        );

        return apiWorkload;
    }

    async deleteAllWorkloads(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        try {
            workloadsService.workloadsApi.accessToken = accessToken;
            const allWorkloads = (await workloadsService.workloadsApi.getWorkloads(false)).body.workloads;

            await Promise.all(
                allWorkloads.map((workload) =>
                    this.deleteWorkload(workload.id, workload.type?.toLowerCase() as WorkloadType)
                )
            );
            if (allWorkloads.length > 0) {
                logger.info('deleted all workloads successfully');
            }
        } catch (error) {
            logger.warn(`Failed to delete workloads: ${error}`);
        }
    }

    /**
     * Delete all workloads in a project via the batch API.
     *
     * @param projectId         project to clear
     * @param timeout           how long to wait for physical deletion to complete (seconds)
     * @param accessToken       caller token; defaults to system-admin
     * @param waitForDeletion   when `true` (default), polls until every workload is gone (up to
     *                          `timeout`). Set to `false` for cleanup-time use cases that don't
     *                          assert on deletion — the batch API call is still issued and
     *                          checked, but the function returns as soon as the API accepts the
     *                          request, saving up to `timeout` seconds per test.
     */
    async deleteWorkloadsByProjectId(
        projectId: string,
        timeout = 120,
        accessToken = flowRootState.users.system_administrator.accessToken,
        waitForDeletion = true
    ): Promise<void> {
        try {
            workloadsService.workloadsApi.accessToken = accessToken;
            const allWorkloads = (await this.getWorkloadsByProjectId(projectId, accessToken)).body.workloads;

            if (allWorkloads.length === 0) {
                logger.info(`No workloads found in project ${projectId} to delete`);
                return;
            }

            // Initiate deletion of all workloads
            await workloadUtils.deleteMultipleWorkloads(allWorkloads.map((workload) => workload.id));

            if (!waitForDeletion) {
                logger.info(
                    `initiated deletion of ${allWorkloads.length} workload(s) in project ${projectId} (no wait)`
                );
                return;
            }

            // Wait for all workloads to be actually deleted
            const areAllWorkloadsDeleted = async (): Promise<boolean> => {
                try {
                    const remainingWorkloads = (await this.getWorkloadsByProjectId(projectId, accessToken))
                        .body.workloads;
                    return remainingWorkloads.length === 0;
                } catch (error) {
                    logger.debug(`Error checking workloads during deletion wait: ${error}`);
                    return false;
                }
            };

            await waitUtils.waitForCondition(
                areAllWorkloadsDeleted,
                `wait for all workloads in project ${projectId} to be deleted`,
                true,
                timeout,
                2
            );

            logger.info(`deleted all workloads of project ${projectId} successfully`);
        } catch (error) {
            logger.warn(`Failed to delete workloads of project ${projectId}: ${error}`);
            throw error;
        }
    }

    async deleteWorkload(
        workloadId: string,
        workloadType: WorkloadType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage }> {
        workloadsCrudUtils.accessToken = accessToken;
        switch (workloadType) {
            case WorkloadType.WORKSPACE:
                return await workloadsCrudUtils.deleteWorkspace(workloadId);
            case WorkloadType.INFERENCE:
                return await workloadsCrudUtils.deleteInference(workloadId);
            case WorkloadType.DISTRIBUTED_INFERENCE:
                return await workloadsCrudUtils.deleteDistributedInference(workloadId);
            case WorkloadType.TRAINING:
                return await workloadsCrudUtils.deleteTraining(workloadId);
            case WorkloadType.DISTRIBUTED:
                return await workloadsCrudUtils.deleteDistributed(workloadId);
        }
    }

    async deleteWorkloadByNameAndProjectId(
        workloadName: string,
        projectId: string | number,
        workloadType: WorkloadType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage }> {
        const workloadId = await this.waitForWorkloadIdByWorkloadNameAndProjectId(
            workloadName,
            projectId.toString()
        );
        const deleteRes = await this.deleteWorkload(workloadId, workloadType, accessToken);
        logger.info(`Deleted workload ${workloadName} with id ${workloadId}`);
        return deleteRes;
    }

    /**
     *
     * @description Wait for the first workload that including the given name to reach a specific status
     * This is a temp solution to support deleting inference workload for old clusters (example 2.13 version since the workload name is not the same as the inference name).
     *
     * @param workloadName the sub/full name of the workload to wait for
     * @param status the status to wait for
     * @param timeout the wait time in seconds
     * @param projectId scopes the lookup server-side; prefer it, as the unscoped
     *        cluster-wide list can truncate and hide the workload entirely.
     *
     */
    async waitForStatusInWorkload(
        workloadName: string,
        status: Phase,
        timeout = 70,
        projectId?: string | number
    ): Promise<void> {
        workloadsService.workloadsApi.accessToken = flowRootState.users.system_administrator.accessToken;

        const fetchCandidates = async (): Promise<Workload[]> => {
            const { body } = projectId
                ? await workloadsUtilsV2.getWorkloadsByNameContainsAndProjectId(workloadName, projectId)
                : await workloadsUtilsV2.getWorkloadsByClusterId(
                      flowRootState.cluster.uuid,
                      flowRootState.users.system_administrator.accessToken,
                      MAX_WORKLOADS_PAGE_SIZE
                  );
            return body.workloads ?? [];
        };

        const isWorkloadExist = async (): Promise<boolean> => {
            let workloads: Workload[];
            try {
                workloads = await fetchCandidates();
            } catch (error) {
                logger.debug(`failed to list workloads while waiting for ${workloadName}: ${error}`);
                return false;
            }

            const workload: Workload | undefined = workloads.find((workload: Workload) =>
                workload.name?.includes(workloadName)
            );

            logger.debug(
                workload
                    ? `workload status: ${workload.phase}`
                    : `could not find workload contains the name ${workloadName} (${workloads.length} workload(s) returned)`
            );

            return workload?.phase === status;
        };

        try {
            await waitUtils.waitForCondition(
                isWorkloadExist,
                `wait for workload including the name ${workloadName} to be ${status}`,
                true,
                timeout,
                WORKLOAD_STATUS_POLL_DELAY_SECONDS
            );
        } catch (error) {
            // On failure, fetch the workload and log its phase and phaseMessage if available
            try {
                const workloads = await fetchCandidates();
                const workload: Workload | undefined = workloads.find((workload: Workload) =>
                    workload.name?.includes(workloadName)
                );
                if (workload) {
                    logger.error(
                        `Workload '${workload.name}' did not reach phase '${status}'. Current phase: '${workload.phase}'. Phase message: '${workload.phaseMessage ?? 'N/A'}'`
                    );
                } else {
                    logger.error(
                        `Workload with name containing '${workloadName}' was not found when logging failure. ` +
                            `Query was ${projectId ? `scoped to project ${projectId}` : 'cluster-wide'} and returned ` +
                            `${workloads.length} workload(s): [${workloads.map((w) => w.name).join(', ')}]`
                    );
                }
            } catch (fetchError) {
                logger.error(
                    `Failed to fetch workload for logging after waitForStatusInWorkload failure: ${fetchError}`
                );
            }
            throw error;
        }
    }

    async listApplications(
        options: {
            accessToken?: string;
            projectId?: string;
            verbosity?: AIApplicationVerbosity;
        } = {}
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        const {
            accessToken = flowRootState.users.system_administrator.accessToken,
            projectId,
            verbosity
        } = options;

        workloadsService.aiApplicationsApi.accessToken = accessToken;
        return await workloadsService.aiApplicationsApi.listApplications(
            verbosity ? verbosity : undefined,
            undefined,
            undefined,
            undefined,
            undefined,
            projectId ? ['projectId==' + projectId] : undefined
        );
    }

    async getApplication(
        applicationId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: AIApplication }> {
        workloadsService.aiApplicationsApi.accessToken = accessToken;
        return await workloadsService.aiApplicationsApi.getApplication(applicationId);
    }

    async waitForIdleWorkloadToAppear(
        timeout: number = 25 * 60
    ): Promise<{ response: http.IncomingMessage; body: GetWorkloads200Response }> {
        let workloads: { response: http.IncomingMessage; body: GetWorkloads200Response };

        const idleWorkloadAppear = async (): Promise<boolean> => {
            try {
                workloads = await this.getAllWorkloads(flowRootState.users.system_administrator.accessToken, [
                    'idleAllocatedGpus>=1',
                    `clusterId==${flowRootState.cluster.uuid}`
                ]);

                return workloads.body.workloads.length >= 1;
            } catch (error) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            idleWorkloadAppear,
            'wait for at least 1 workload to become idle',
            true,
            timeout,
            5
        );

        return workloads;
    }

    async getWorkloadElements(
        workloadId: string,
        offset?: number,
        limit?: number,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadElementsResponse }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.getWorkloadElements(workloadId, offset, limit);
    }

    async countWorkloadElements(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: CountWorkloadElements200Response }> {
        workloadsService.workloadsApi.accessToken = accessToken;
        return await workloadsService.workloadsApi.countWorkloadElements(workloadId);
    }

    async hasExpectedElementCount(workloadId: string, expectedCount: number): Promise<boolean> {
        const res = await this.getWorkloadElements(workloadId);
        const elements = res.body.elements;
        return Array.isArray(elements) && elements.length === expectedCount;
    }

    // Elements are written by cluster-sync asynchronously, so callers poll on this
    // before asserting the topology constraints of a specific component.
    async elementsHaveTopology(workloadId: string, instanceKeys: string[]): Promise<boolean> {
        const elements = (await this.getWorkloadElements(workloadId)).body.elements || [];
        return instanceKeys.every((instanceKey) =>
            elements.some((el) => el.instanceKey === instanceKey && !!el.topology?.requiredPlacement)
        );
    }

    async elementIdsMatch(workloadId: string, expectedIds: string[]): Promise<boolean> {
        const normalize = (ids: string[]): string[] => [...ids].sort();
        const res = await this.getWorkloadElements(workloadId);
        const apiIds = res.body.elements?.map((el) => el.id) || [];
        return (
            apiIds.length > 0 && JSON.stringify(normalize(apiIds)) === JSON.stringify(normalize(expectedIds))
        );
    }

    async allPodsHaveElements(workloadId: string, expectedPodCount: number): Promise<boolean> {
        const res = await this.getWorkloadPods(workloadId);
        const pods = res.body.pods || [];
        if (pods.length < expectedPodCount) return false;
        return pods.every((pod: any) => pod.elements?.length > 0);
    }
}
