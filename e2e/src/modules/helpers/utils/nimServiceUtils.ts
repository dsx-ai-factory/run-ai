/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, workloadsManagerUtils, waitUtils, config, inferenceUtils, workloadsUtilsV2 } from '..';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { NimServiceResponse } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceResponse';
import { NimServiceUpdateRequest } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceUpdateRequest';
import { workloadsManagerService } from '../../api';
import { kubeCli } from '../../cli';
import { Project } from '../../api/apis/orgUnitApi';
import { NimServiceSpec } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import { NimServiceServingPortAuthorizationType } from '../../api/apis/workloadsManagerApi/generated/model/nimServiceServingPortAuthorizationType';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';

/* Timeouts aligned with NIM pod probes: 240s wait fits startupProbe (initialDelaySeconds 30, periodSeconds 10) + scheduling/image pull; 300s for distributed (CLUSTER_START_TIMEOUT); test timeout must exceed sum of waits in each test. */
export const NIM_TEST_TIMEOUT_MS = 300 * 1000; // single create+wait(240) + assertions
export const NIM_HEAVY_TEST_TIMEOUT_MS = 480 * 1000; // single create + wait(300) + traffic + metrics
export const NIM_AUTOSCALING_TEST_TIMEOUT_MS = 420 * 1000; // create + wait(240) + waitForPods(90)
export const NIM_METRICS_DISTRIBUTED_TEST_TIMEOUT_MS = 360 * 1000; // create + wait(300) distributed
export const NIM_UPDATE_TEST_TIMEOUT_MS = 540 * 1000; // create+wait(240) + update+wait(240) + buffer
export const NIM_UPDATE_SYNC_TEST_TIMEOUT_MS = 720 * 1000; // create+wait(240) + update+wait(240) + waitForCondition(120-180)

export const NGC_AUTH_SECRET_NAME = 'ngc-secret';

export function createDefaultNimServiceSpec(): NimServiceSpec {
    return {
        ngcAuthSecret: NGC_AUTH_SECRET_NAME,
        image: config.quickStartImage.nim,
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
        },
        replicas: 1
    };
}

export async function createNimService(
    workloadName,
    project: Project,
    nimServiceSpec
): Promise<NimServiceResponse> {
    logger.debug(`creating nimservice ${workloadName} in namespace ${project.status.namespace}`);
    const result = await workloadsManagerUtils.createNimService(workloadName, project.id, nimServiceSpec);

    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

    return result.body;
}

export async function getNimServiceById(workloadId: string): Promise<NimServiceResponse> {
    logger.debug(`getting nimservice ${workloadId}`);
    const result = await workloadsManagerUtils.getNimServiceById(workloadId);
    expect(result.response.statusCode).to.equal(StatusCodes.OK);
    return result.body;
}

export async function updateNimServiceSpec(
    workloadId: string,
    specPatch: NimServiceUpdateRequest['spec'],
    expectedStatusCode: number | number[] = StatusCodes.ACCEPTED,
    fieldsToRemovePatch: NimServiceUpdateRequest['fieldsToRemove'] = undefined
): Promise<{ response: any; body: NimServiceResponse }> {
    const updateRequest: NimServiceUpdateRequest = { spec: specPatch, fieldsToRemove: fieldsToRemovePatch };
    logger.info(
        `Patching NIM service ${workloadId} spec: ${JSON.stringify(
            updateRequest.spec
        )} (expecting status=${JSON.stringify(expectedStatusCode)})`
    );
    try {
        const updateResponse = await workloadsManagerService.nimServicesApi.updateNimServiceSpec(
            workloadId,
            updateRequest
        );
        const actualStatusCode = updateResponse.response.statusCode;
        if (Array.isArray(expectedStatusCode)) {
            expect(expectedStatusCode, 'Unexpected PATCH response statusCode').to.include(actualStatusCode);
        } else {
            expect(actualStatusCode, 'Unexpected PATCH response statusCode').to.equal(expectedStatusCode);
        }
        logger.info(
            `PATCH accepted for ${workloadId}: status=${actualStatusCode}, response.spec=${JSON.stringify(
                updateResponse.body?.spec
            )}`
        );
        return updateResponse;
    } catch (error) {
        const statusCode = error.response?.statusCode || error.statusCode;
        const errorBody = error.response?.body || error.body || {};
        logger.error(
            `PATCH failed for ${workloadId}: status=${statusCode}, request.spec=${JSON.stringify(
                updateRequest.spec
            )}, response=${JSON.stringify(errorBody)}`
        );
        throw error;
    }
}

export async function createNgcAuthSecret(namespace: string): Promise<void> {
    const secretName = NGC_AUTH_SECRET_NAME;
    const secretData = {
        // Dummy fixture: base64-encoded 'key\n', not an NGC credential.
        NGC_API_KEY: 'a2V5Cg=='
    };

    await kubeCli.createSecret(secretName, namespace, secretData);
    logger.info(`Created secret ${secretName} in namespace ${namespace}`);
    await kubeCli.getSecret(secretName, namespace);
    logger.info(`Verified secret ${secretName} exists in namespace ${namespace}`);
}

export async function createAndWaitForNimService(
    workloadName: string,
    project: Project,
    spec: NimServiceSpec,
    phase: Phase = Phase.Running,
    timeoutSeconds = 240
): Promise<string> {
    const createResponse = await createNimService(workloadName, project, spec);
    const workloadId = createResponse.metadata.id;

    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, phase, timeoutSeconds);
    logger.info(`NIM service ${workloadId} created and reached ${phase} phase`);

    return workloadId;
}

export function verifyErrorResponse(
    error: any,
    expectedStatusCode: number | number[],
    messageKeywords: string[],
    errorContext: string
): void {
    const statusCode = error.response?.statusCode || error.statusCode;
    if (Array.isArray(expectedStatusCode)) {
        expect(expectedStatusCode).to.include(statusCode);
    } else {
        expect(statusCode).to.equal(expectedStatusCode);
    }
    logger.info(`Verified HTTP status code: ${statusCode}`);

    const errorBody = error.response?.body || error.body || {};
    const errorMessage = errorBody.message || errorBody.error || '';
    const errorDetails = errorBody.details || '';
    const fullErrorText = `${errorMessage} ${errorDetails}`.toLowerCase();

    expect(fullErrorText).to.satisfy(
        (msg: string) => messageKeywords.some((keyword) => msg.includes(keyword.toLowerCase())),
        `Error message should indicate ${errorContext}`
    );
    logger.info(
        `Verified error message indicates ${errorContext}: ${errorMessage}${errorDetails ? ', details: ' + errorDetails : ''}`
    );
}

export async function executeWithTokenSwitch<T>(
    apiService: any,
    newToken: string,
    operation: () => Promise<T>
): Promise<T> {
    const originalToken = apiService.accessToken;
    try {
        apiService.accessToken = newToken;
        return await operation();
    } finally {
        apiService.accessToken = originalToken;
    }
}

export async function verifyNimServiceFieldUnchanged(
    workloadId: string,
    fieldPath: string,
    expectedValue: any
): Promise<void> {
    const getResponse = await getNimServiceById(workloadId);
    const fieldParts = fieldPath.split('.');
    let currentValue: any = getResponse;
    for (const part of fieldParts) {
        currentValue = currentValue[part];
    }
    expect(currentValue).to.equal(expectedValue);
    logger.info(`Verified ${fieldPath} unchanged: ${expectedValue}`);
}

export function validateNimServiceUpdateResponse(
    responseBody: NimServiceResponse,
    expectedWorkloadId: string,
    expectedWorkloadName: string,
    expectedProjectId: string
): void {
    expect(responseBody.metadata, 'Response body does not contain metadata').to.not.be.undefined;
    expect(responseBody.metadata.id, 'Metadata ID mismatch').to.equal(expectedWorkloadId);
    expect(responseBody.metadata.name, 'Metadata name mismatch').to.equal(expectedWorkloadName);
    expect(String(responseBody.metadata.projectId), 'Metadata projectId mismatch').to.equal(
        expectedProjectId
    );
    expect(responseBody.metadata.createdAt, 'Metadata does not contain createdAt').to.not.be.undefined;
    expect(String(responseBody.metadata.createdAt).length, 'createdAt should not be empty').to.be.greaterThan(
        0
    );
    expect(responseBody.metadata.updatedAt, 'Metadata does not contain updatedAt').to.not.be.undefined;
    expect(String(responseBody.metadata.updatedAt).length, 'updatedAt should not be empty').to.be.greaterThan(
        0
    );

    const createdAt = new Date(responseBody.metadata.createdAt);
    const updatedAt = new Date(responseBody.metadata.updatedAt);
    expect(updatedAt.getTime(), 'updatedAt should be after createdAt').to.be.greaterThan(createdAt.getTime());

    expect(responseBody.spec, 'Response body does not contain spec').to.not.be.undefined;

    logger.info('Validated NIM service UPDATE response structure');
}

export function buildNimServiceAuthSpec(overrides?: {
    authorizationType?: string;
    authorizedUsers?: string[];
    authorizedGroups?: string[];
    exposeExternally?: boolean;
}): NimServiceSpec {
    const spec = createDefaultNimServiceSpec();
    spec.servingPort.exposeExternally = overrides?.exposeExternally ?? true;
    delete spec.servingPort.grpcPort;

    if (overrides?.authorizationType !== undefined) {
        spec.servingPort.authorizationType =
            overrides.authorizationType as unknown as NimServiceServingPortAuthorizationType;
    }
    if (overrides?.authorizedUsers !== undefined) {
        spec.servingPort.authorizedUsers = overrides.authorizedUsers;
    }
    if (overrides?.authorizedGroups !== undefined) {
        spec.servingPort.authorizedGroups = overrides.authorizedGroups;
    }
    return spec;
}

export async function createRunningNimService(options: {
    name: string;
    project: Project;
    authorizationType?: string;
    authorizedUsers?: string[];
    authorizedGroups?: string[];
    exposeExternally?: boolean;
    timeoutSeconds?: number;
    skipExternalUrl?: boolean;
}): Promise<{ workloadId: string; externalUrl: string }> {
    const spec = buildNimServiceAuthSpec({
        authorizationType: options.authorizationType,
        authorizedUsers: options.authorizedUsers,
        authorizedGroups: options.authorizedGroups,
        exposeExternally: options.exposeExternally
    });

    const workloadId = await createAndWaitForNimService(
        options.name,
        options.project,
        spec,
        Phase.Running,
        options.timeoutSeconds ?? 180
    );

    if (options.skipExternalUrl) {
        return { workloadId, externalUrl: '' };
    }

    let externalUrl = '';
    await waitUtils.waitForCondition(
        async () => {
            const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
            const urls = workload.body.urls || [];
            externalUrl = urls.find((u: string) => !u.includes('svc.cluster.local')) || '';
            logger.info(
                `Available URLs: ${JSON.stringify(urls)}, selected external URL: ${externalUrl || 'none yet'}`
            );
            return externalUrl !== '';
        },
        `external URL for workload ${workloadId}`,
        true,
        60,
        5
    );
    expect(externalUrl).to.match(/^https?:\/\//);

    return { workloadId, externalUrl };
}

export async function createNimServiceExpectingError(
    name: string,
    project: Project,
    spec: NimServiceSpec,
    expectedStatus: number | number[],
    messageKeywords: string[],
    errorContext: string
): Promise<void> {
    let result: any;
    try {
        result = await workloadsManagerUtils.createNimService(name, project.id, spec);
    } catch (error) {
        verifyErrorResponse(error, expectedStatus, messageKeywords, errorContext);
        return;
    }

    const statusCode = result.response.statusCode;
    const expectedArr = Array.isArray(expectedStatus) ? expectedStatus : [expectedStatus];
    expect(
        expectedArr,
        `Expected API to reject with ${expectedArr.join('/')}, but got ${statusCode} (${errorContext})`
    ).to.include(statusCode);
}
