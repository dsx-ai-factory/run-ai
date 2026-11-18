/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadsManagerService } from '../../api';
import { NimServiceSpec } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import { NimServiceResponse } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceResponse';
import { NimServiceCreateRequest } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceCreateRequest';
import { WorkloadV2CreateRequest } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2CreateRequest';
import { WorkloadV2MetadataCreateParams } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2MetadataCreateParams';
import { WorkloadV2UpdateRequest } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2UpdateRequest';
import { WorkloadV2 } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2';
import { WorkloadV2GetResponse } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2GetResponse';
import { WorkloadV2SyncStatus } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2SyncStatus';
import { WorkloadV2UpdateConfigurationPvc } from '../../api/apis/workloadsManagerApi/generated/model/workloadV2UpdateConfigurationPvc';
import { Interceptor } from '../../api/apis/workloadsManagerApi/generated/model/models';
import { HttpError } from '../../api/apis/workloadsManagerApi/generated/api/httpError';
import http from 'http';
import { logger, crds, waitUtils, BaseWorkload } from '../index';
import { CustomResourceDefinition } from '../models/customResourceDefinition';
import { flowRootState } from '../config/config';
import { expect } from '@playwright/test';
import { kubeCli } from '../../cli';
import {
    DynamoComponentConfig,
    DynamoGraphDeployment,
    DynamoGraphDeploymentCrd,
    DynamoGraphDeploymentV1Beta1,
    DynamoDefaultImage,
    DynamoTopologyConstraint,
    NodepoolConfig,
    ServiceConfig
} from '../models/externalWorkload';
// Type-only (erased at build) so it adds no runtime import edge to the existing WorkloadsManagerUtils <-> index cycle.
import type { WorkloadV2PvcEntry } from './storagePluginNewPvcsHelpers';

const DEFAULT_MANIFEST_IMAGE = 'busybox';
const DEFAULT_CONTAINER_NAME = 'app';
const DEFAULT_PRIORITY = 'medium';
const DEFAULT_CATEGORY = 'Build';

export class WorkloadsManagerUtils {
    set accessToken(accessToken: string | (() => string)) {
        workloadsManagerService.workloadsV2Api.accessToken = accessToken;
        workloadsManagerService.nimServicesApi.accessToken = accessToken;
    }

    async createWorkloadV2(
        name: string,
        projectId: string,
        manifest?: { [key: string]: any },
        metadata?: Partial<WorkloadV2MetadataCreateParams>,
        manifestBase64Encoded?: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadV2 }> {
        const workloadMetadata: WorkloadV2MetadataCreateParams = {
            name,
            projectId: String(projectId),
            ...metadata
        };

        const request: WorkloadV2CreateRequest = {
            metadata: workloadMetadata,
            ...(manifest !== undefined && { manifest }),
            ...(manifestBase64Encoded !== undefined && { manifestBase64Encoded })
        };

        try {
            workloadsManagerService.workloadsV2Api.accessToken = accessToken;
            const result = await workloadsManagerService.workloadsV2Api.createWorkloadV2(request);
            logger.info(`Successfully created workload v2: ${name}`);
            return result;
        } catch (error) {
            logger.error(`Failed to create workload v2: name="${name}", projectId="${projectId}"`, {
                error: error.message,
                stack: error.stack,
                statusCode: error.response?.statusCode,
                responseBody: error.response?.body
            });
            throw error;
        }
    }

    // Create a V2 workload declaring managed PVCs via metadata.configuration.pvcs, routed through the
    // typed createWorkloadV2. Returns the created workload id, or undefined when the response omits it —
    // the caller asserts on it, keeping this util assertion-free.
    async createWorkloadV2WithPvcs(
        name: string,
        projectId: string,
        pvcs: WorkloadV2PvcEntry[],
        manifest: { [key: string]: any },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<string | undefined> {
        // Only attach configuration.pvcs when the workload actually declares managed PVCs;
        // an empty/absent configuration is the truthful "declares no PVCs" submission.
        const metadata: Partial<WorkloadV2MetadataCreateParams> | undefined =
            pvcs.length > 0
                ? { configuration: { pvcs: pvcs as unknown as WorkloadV2UpdateConfigurationPvc[] } }
                : undefined;
        const { body } = await this.createWorkloadV2(
            name,
            projectId,
            manifest,
            metadata,
            undefined,
            accessToken
        );
        const id = body.metadata?.id;
        logger.info(`Created workload v2 with pvcs: name="${name}", id="${id}"`);
        return id;
    }

    // Issue the create with a raw (possibly malformed) body so the SERVER's additionalProperties:false /
    // DisallowUnknownFields rejection is exercised; returns {status, data} without throwing on a 4xx.
    // A non-HttpError (network/serialization failure, client bug) is re-thrown so the caller's test fails
    // honestly instead of it masquerading as a 4xx — error assertion stays with the caller.
    async tryCreateWorkloadV2(
        metadata: Record<string, unknown>,
        manifest: Record<string, unknown>,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ status: number; data: unknown }> {
        const api = workloadsManagerService.workloadsV2Api;
        api.accessToken = accessToken;
        // The typed serializer would strip the malformed/unknown fields under test, so override the
        // serialized body with the raw payload via a one-shot interceptor on the generated create call.
        const sendRawBody: Interceptor = (options) => {
            (options as { data?: unknown }).data = { metadata, manifest };
        };
        api.addInterceptor(sendRawBody);
        try {
            const result = await api.createWorkloadV2({
                metadata: {
                    name: String(metadata.name ?? ''),
                    projectId: String(metadata.projectId ?? '')
                }
            });
            return { status: result.response.statusCode ?? 0, data: result.body };
        } catch (error) {
            if (!(error instanceof HttpError)) {
                throw error;
            }
            return { status: error.statusCode ?? 0, data: error.body };
        } finally {
            const registered = (api as unknown as { interceptors: Interceptor[] }).interceptors;
            const idx = registered.indexOf(sendRawBody);
            if (idx >= 0) {
                registered.splice(idx, 1);
            }
        }
    }

    async getWorkloadV2ById(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadV2GetResponse }> {
        try {
            workloadsManagerService.workloadsV2Api.accessToken = accessToken;
            const result = await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
            logger.debug(`Successfully retrieved workload v2: ${workloadId}`);
            return result;
        } catch (error) {
            logger.error(`Failed to get workload v2 by ID: workloadId="${workloadId}"`, {
                error: error.message,
                stack: error.stack,
                statusCode: error.response?.statusCode,
                responseBody: error.response?.body
            });
            throw error;
        }
    }

    /** Poll until the workload's sync reaches a terminal state (Synced or Failed) and return the last
     * GET body without asserting which — for negative cases that EXPECT Failed. */
    async waitForWorkloadV2SyncTerminal(
        workloadId: string,
        timeoutSec = 120,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadV2GetResponse> {
        let last: WorkloadV2GetResponse | undefined;
        await waitUtils.waitForCondition(
            async () => {
                last = (await this.getWorkloadV2ById(workloadId, accessToken)).body;
                const status = last.syncInfo?.status;
                return status === WorkloadV2SyncStatus.Synced || status === WorkloadV2SyncStatus.Failed;
            },
            `workload ${workloadId} sync to reach a terminal state (Synced/Failed)`,
            true,
            timeoutSec
        );
        return last!;
    }

    /** Poll until the control-plane reports the workload synced to the cluster (syncInfo.status=Synced),
     * failing fast on Failed. The V2 "applied" signal, not pod readiness. */
    async waitForWorkloadV2Synced(
        workloadId: string,
        timeoutSec = 120,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<WorkloadV2GetResponse> {
        const last = await this.waitForWorkloadV2SyncTerminal(workloadId, timeoutSec, accessToken);
        if (last.syncInfo?.status !== WorkloadV2SyncStatus.Synced) {
            throw new Error(
                `workload ${workloadId} did not sync to the cluster successfully ` +
                    `(syncInfo.status: ${last.syncInfo?.status ?? 'n/a'}, message: ${last.syncInfo?.message ?? 'n/a'})`
            );
        }
        return last;
    }

    async deleteWorkloadV2(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: WorkloadV2 }> {
        workloadsManagerService.workloadsV2Api.accessToken = accessToken;
        return workloadsManagerService.workloadsV2Api.deleteWorkloadV2(workloadId);
    }

    async updateWorkloadV2Metadata(
        workloadId: string,
        manifest: { [key: string]: any },
        metadata?: { priority?: string; category?: string },
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        try {
            const updateRequest: WorkloadV2UpdateRequest = {
                metadata: metadata,
                manifest: manifest
            };

            workloadsManagerService.workloadsV2Api.accessToken = accessToken;
            const result = await workloadsManagerService.workloadsV2Api.updateWorkloadV2(
                workloadId,
                updateRequest
            );
            logger.debug(`Successfully updated workload v2 metadata: ${workloadId}`, metadata);
            return result;
        } catch (error) {
            logger.error(`Failed to update workload v2 metadata: workloadId="${workloadId}"`, {
                error: error.message,
                stack: error.stack,
                statusCode: error.response?.statusCode,
                responseBody: error.response?.body,
                requestedMetadata: metadata
            });
            throw error;
        }
    }

    async createNimService(
        name: string,
        projectId: string,
        spec: NimServiceSpec
    ): Promise<{ response: http.IncomingMessage; body: NimServiceResponse }> {
        const request: NimServiceCreateRequest = {
            metadata: {
                name: name,
                projectId: projectId
            },
            spec: spec
        };

        return workloadsManagerService.nimServicesApi.createNimService(request);
    }

    async getNimServiceById(
        workloadId: string
    ): Promise<{ response: http.IncomingMessage; body: NimServiceResponse }> {
        return workloadsManagerService.nimServicesApi.getNimServiceById(workloadId);
    }

    createStandardManifest(name: string, replicas: number = 1): { [key: string]: any } {
        return this.createDeploymentManifest(
            name,
            replicas,
            DEFAULT_MANIFEST_IMAGE,
            DEFAULT_CONTAINER_NAME,
            undefined,
            ['sh', '-c', 'while true; do echo hello; sleep 5; done']
        );
    }

    createDeploymentManifest(
        name: string,
        replicas: number,
        image: string,
        containerName: string = 'main',
        containerPort?: number,
        command?: string[]
    ): { [key: string]: any } {
        const container: any = {
            name: containerName,
            image: image
        };

        if (containerPort) {
            container.ports = [
                {
                    containerPort: containerPort
                }
            ];
        }

        if (command) {
            container.command = command;
        }

        return {
            apiVersion: 'apps/v1',
            kind: 'Deployment',
            metadata: {
                name: name
            },
            spec: {
                replicas: replicas,
                selector: {
                    matchLabels: {
                        app: name
                    }
                },
                template: {
                    metadata: {
                        labels: {
                            app: name
                        }
                    },
                    spec: {
                        containers: [container]
                    }
                }
            }
        };
    }

    createNginxDeploymentManifest(name: string, replicas: number, imageTag: string): { [key: string]: any } {
        return this.createDeploymentManifest(name, replicas, `nginx:${imageTag}`, 'nginx', 80);
    }

    createNIMServiceManifest(
        name: string,
        image: string,
        replicas: number,
        ngcAuthSecret: string
    ): { [key: string]: any } {
        return {
            apiVersion: 'apps.nvidia.com/v1alpha1',
            kind: 'NIMService',
            metadata: {
                name: name
            },
            spec: {
                image: image,
                replicas: replicas,
                ngcAuthSecret: ngcAuthSecret,
                servingPort: {
                    port: 8000
                },
                modelStore: {
                    pvc: {
                        existingPvc: false,
                        claimInfo: {
                            size: '10M',
                            accessModes: {
                                readWriteOnce: true
                            }
                        }
                    }
                },
                compute: {
                    cpuCoreRequest: 0.1,
                    cpuMemoryRequest: '10M'
                }
            }
        };
    }

    async createDeploymentWorkloadV2(options: {
        name: string;
        projectId: string;
        replicas: number;
        image?: string;
        containerName?: string;
        containerPort?: number;
        priority?: string;
        category?: string;
        useStandardManifest?: boolean;
    }): Promise<string> {
        let manifest: any;

        if (options.useStandardManifest) {
            manifest = this.createStandardManifest(options.name, options.replicas);
        } else {
            manifest = this.createDeploymentManifest(
                options.name,
                options.replicas,
                options.image || DEFAULT_MANIFEST_IMAGE,
                options.containerName || DEFAULT_CONTAINER_NAME,
                options.containerPort
            );
        }

        const metadata: WorkloadV2MetadataCreateParams = {
            name: options.name,
            projectId: options.projectId,
            priority: options.priority || DEFAULT_PRIORITY,
            category: options.category || DEFAULT_CATEGORY
        };

        const createRequest: WorkloadV2CreateRequest = {
            metadata: metadata,
            manifest: manifest
        };

        try {
            const createResponse =
                await workloadsManagerService.workloadsV2Api.createWorkloadV2(createRequest);
            const workloadId = createResponse.body.metadata?.id;

            if (createResponse.response.statusCode !== 202) {
                const errorMsg = `Expected status code 202, got ${createResponse.response.statusCode}`;
                logger.error(errorMsg, {
                    workloadId,
                    statusCode: createResponse.response.statusCode,
                    responseBody: createResponse.body
                });
                const error: any = new Error(errorMsg);
                error.workloadId = workloadId;
                error.statusCode = createResponse.response.statusCode;
                throw error;
            }

            if (!workloadId) {
                throw new Error('Workload ID not found in response');
            }

            const imageInfo = options.useStandardManifest
                ? `${DEFAULT_MANIFEST_IMAGE} (standard)`
                : options.image || DEFAULT_MANIFEST_IMAGE;
            logger.info(
                `Created Deployment workload v2: name="${options.name}", id="${workloadId}", ` +
                    `priority="${metadata.priority}", category="${metadata.category}", ` +
                    `replicas=${options.replicas}, image="${imageInfo}"`
            );

            return workloadId;
        } catch (error) {
            const workloadId = error.workloadId || error.response?.body?.metadata?.id;
            logger.error(`Failed to create Deployment workload v2: name="${options.name}"`, {
                error: error.message,
                stack: error.stack,
                statusCode: error.response?.statusCode || error.statusCode,
                responseBody: error.response?.body,
                workloadId: workloadId || 'N/A'
            });
            throw error;
        }
    }

    validateWorkloadV2UpdateResponse(
        responseBody: any,
        expectedWorkloadId: string,
        expectedWorkloadName: string,
        expectedProjectId: string
    ): void {
        expect(responseBody.metadata, 'Response body does not contain metadata').toBeDefined();
        expect(responseBody.metadata.id, 'Metadata ID mismatch').toBe(expectedWorkloadId);
        expect(responseBody.metadata.name, 'Metadata name mismatch').toBe(expectedWorkloadName);
        expect(responseBody.metadata.projectId, 'Metadata projectId mismatch').toBe(expectedProjectId);
        expect(responseBody.metadata.createdAt, 'Metadata does not contain createdAt').toBeDefined();
        expect(responseBody.metadata.updatedAt, 'Metadata does not contain updatedAt').toBeDefined();
        expect(responseBody.metadata.priority, 'Metadata does not contain priority').toBeDefined();
        expect(responseBody.metadata.category, 'Metadata does not contain category').toBeDefined();

        expect(
            responseBody.workloadManifest,
            'Response body does not contain workloadManifest'
        ).toBeDefined();
        expect(
            responseBody.workloadManifest.apiVersion,
            'WorkloadManifest does not contain apiVersion'
        ).toBeDefined();
        expect(responseBody.workloadManifest.kind, 'WorkloadManifest does not contain kind').toBeDefined();
        expect(
            responseBody.workloadManifest.metadata,
            'WorkloadManifest does not contain metadata'
        ).toBeDefined();
        expect(responseBody.workloadManifest.spec, 'WorkloadManifest does not contain spec').toBeDefined();

        expect(responseBody.rawManifest, 'Response body does not contain rawManifest').toBeDefined();
        expect(responseBody.rawManifest.apiVersion, 'RawManifest does not contain apiVersion').toBeDefined();
        expect(responseBody.rawManifest.kind, 'RawManifest does not contain kind').toBeDefined();

        expect(responseBody.desiredPhase, 'Response body does not contain desiredPhase').toBeDefined();
        logger.info('Validated Workload V2 UPDATE response structure');
    }

    validateDeploymentManifestStructure(workload: any): void {
        expect(workload.workloadManifest, 'workloadManifest should be defined').toBeDefined();
        expect(workload.workloadManifest.spec, 'workloadManifest.spec should be defined').toBeDefined();
        expect(workload.workloadManifest.spec.template, 'spec.template should be defined').toBeDefined();
        expect(workload.workloadManifest.spec.template.spec, 'template.spec should be defined').toBeDefined();
        expect(
            workload.workloadManifest.spec.template.spec.containers,
            'containers should be defined'
        ).toBeDefined();
        expect(
            workload.workloadManifest.spec.template.spec.containers.length,
            'containers should not be empty'
        ).toBeGreaterThan(0);
    }

    async submitAndWaitForDGD(
        dgdName: string,
        namespace: string,
        projectName: string,
        projectId: string,
        replicas: number,
        host: string,
        workloads: BaseWorkload[],
        nodepools?: NodepoolConfig[],
        services?: { [serviceName: string]: ServiceConfig[] },
        annotations?: { [key: string]: string },
        labels?: { [key: string]: string },
        runtimeVersionOverride?: string
    ): Promise<string> {
        const dgdWorkload = new DynamoGraphDeployment(
            dgdName,
            namespace,
            projectName,
            replicas,
            host,
            DynamoDefaultImage,
            undefined, // cpuLimit
            undefined, // cpuRequest
            undefined, // memoryLimit
            undefined, // memoryRequest
            undefined, // gpuLimit
            undefined, // gpuRequest
            nodepools,
            services,
            annotations,
            labels,
            runtimeVersionOverride
        );

        workloads.push(dgdWorkload);
        logger.info(`Submitting DGD: ${dgdName} in namespace: ${namespace}`);

        const manifest = dgdWorkload.getBody();

        const result = await this.createWorkloadV2(dgdName, projectId, manifest);

        const workloadId = result.body.metadata.id;
        logger.info(`DGD submitted with workload ID: ${workloadId}`);

        await this.waitForDgdSuccessfulState(dgdName, namespace, crds.dynamoGraphDeploymentCrd);
        return workloadId;
    }

    // v1beta1 DGDs carry a components list and native pod templates instead of the v1alpha1
    // services map, so they get their own submit path rather than more optional parameters.
    async submitAndWaitForDGDV1Beta1(options: {
        dgdName: string;
        namespace: string;
        projectId: string;
        workloads: BaseWorkload[];
        components: DynamoComponentConfig[];
        topologyConstraint?: DynamoTopologyConstraint;
    }): Promise<string> {
        const { dgdName, namespace, projectId, workloads, components, topologyConstraint } = options;

        const dgdWorkload = new DynamoGraphDeploymentV1Beta1(
            dgdName,
            namespace,
            components,
            topologyConstraint,
            DynamoDefaultImage
        );

        workloads.push(dgdWorkload);
        logger.info(`Submitting v1beta1 DGD: ${dgdName} in namespace: ${namespace}`);

        const result = await this.createWorkloadV2(dgdName, projectId, dgdWorkload.getBody());

        const workloadId = result.body.metadata.id;
        logger.info(`DGD submitted with workload ID: ${workloadId}`);

        await this.waitForDgdSuccessfulState(dgdName, namespace, crds.dynamoGraphDeploymentV1Beta1Crd);
        return workloadId;
    }

    private async waitForDgdSuccessfulState(
        dgdName: string,
        namespace: string,
        crd: CustomResourceDefinition
    ): Promise<void> {
        const expectedState = 'successful';
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const dgd = await kubeCli.getCrd<DynamoGraphDeploymentCrd>(crd, dgdName, namespace);
                    const currentState = dgd?.status?.state;
                    logger.info(`[DGD State Check] ${dgdName}: ${currentState} (expected: ${expectedState})`);

                    if (currentState === 'failed' || currentState === 'error') {
                        throw new Error(`DGD entered failure state: ${currentState}`);
                    }

                    return currentState === expectedState;
                } catch (error) {
                    logger.debug(`Error checking DGD state: ${error.message}`);
                    return false;
                }
            },
            `DGD ${dgdName} to reach ${expectedState} state`,
            true,
            360,
            10
        );

        logger.info(`DGD ${dgdName} reached ${expectedState} state`);
    }
}
