/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { datavolumesService } from '../../api';
import { DatavolumePhase } from '../../api/apis/datavolumesApi/generated/model/datavolumePhase';
import { crds, flowRootState, logger, waitUtils } from '..';
import { Datavolume } from '../../api/apis/datavolumesApi/generated/model/datavolume';
import { DatavolumeRequestType } from '../../api/apis/datavolumesApi/generated/model/datavolumeRequestType';
import { DatavolumeCreationFields } from '../../api/apis/datavolumesApi/generated/model/datavolumeCreationFields';
import { DatavolumePatchFields } from '../../api/apis/datavolumesApi/generated/model/datavolumePatchFields';
import { DatavolumeNoSharedScopes } from '../../api/apis/datavolumesApi/generated/model/datavolumeNoSharedScopes';
import { SharedScopesPatchRequest } from '../../api/apis/datavolumesApi/generated/model/sharedScopesPatchRequest';
import { SharedScope } from '../../api/apis/datavolumesApi/generated/model/sharedScope';
import { StatusCodes } from 'http-status-codes';

export class DatavolumeUtils {
    //////////////////// API Wrappers ////////////////////
    async getDatavolume(
        datavolumeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Datavolume> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (await datavolumesService.datavolumesApi.getDatavolume(datavolumeId)).body;
    }

    async getDatavolumes(
        requestType: DatavolumeRequestType,
        usableInProjectId?: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<DatavolumeNoSharedScopes>> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (await datavolumesService.datavolumesApi.getDatavolumes(requestType, usableInProjectId)).body
            .datavolumes;
    }

    async createDatavolume(
        datavolumeCreationFields: DatavolumeCreationFields,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Datavolume> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (await datavolumesService.datavolumesApi.createDatavolume(datavolumeCreationFields)).body;
    }

    async deleteDatavolume(
        datavolumeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        await datavolumesService.datavolumesApi.deleteDatavolume(datavolumeId);
    }

    async patchDatavolume(
        datavolumeId: string,
        datavolumePatchFields: DatavolumePatchFields,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<DatavolumeNoSharedScopes> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (await datavolumesService.datavolumesApi.patchDatavolume(datavolumeId, datavolumePatchFields))
            .body;
    }

    async patchDatavolumeSharedScopes(
        datavolumeId: string,
        sharedScopesPatchRequest: SharedScopesPatchRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<SharedScope>> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (
            await datavolumesService.datavolumesApi.patchDatavolumeSharedScopes(
                datavolumeId,
                sharedScopesPatchRequest
            )
        ).body;
    }

    async getDatavolumeSharedScopes(
        datavolumeId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<SharedScope>> {
        datavolumesService.datavolumesApi.accessToken = accessToken;
        return (await datavolumesService.datavolumesApi.getDatavolumeSharedScopes(datavolumeId)).body;
    }

    //////////////////// API Utils ////////////////////
    async deleteAllDatavolumes(): Promise<void> {
        const datavolumes = await this.getDatavolumes(DatavolumeRequestType.Originated);
        for (const datavolume of datavolumes) {
            await this.deleteDatavolume(datavolume.id);
        }
    }

    async validateApiDatavolumePhase(datavolumeId: string, expectedPhase: DatavolumePhase): Promise<boolean> {
        try {
            const datavolume = (await datavolumesService.datavolumesApi.getDatavolume(datavolumeId)).body;
            logger.debug(`datavolume ${datavolumeId} has phase ${datavolume.status.phase}`);
            return datavolume.status.phase === expectedPhase;
        } catch (e: any) {
            if (e?.statusCode === StatusCodes.NOT_FOUND) {
                logger.debug(`datavolume ${datavolumeId} not found yet, will retry`);
                return false;
            }
            throw e;
        }
    }

    async waitForApiDatavolumeToBeInPhase(
        datavolumeId: string,
        expectedPhase: DatavolumePhase
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => this.validateApiDatavolumePhase(datavolumeId, expectedPhase),
            `waiting for datavolume ${datavolumeId} to be in phase ${expectedPhase}`,
            true,
            120,
            5
        );
    }

    async waitForApiDatavolumeToBeReady(datavolumeId: string): Promise<void> {
        await this.waitForApiDatavolumeToBeInPhase(datavolumeId, DatavolumePhase.Ready);
    }

    async waitForApiDatavolumeToBeNotReady(datavolumeId: string): Promise<void> {
        await this.waitForApiDatavolumeToBeInPhase(datavolumeId, DatavolumePhase.NotReady);
    }

    async waitForApiDatavolumeToExist(datavolumeId: string): Promise<void> {
        const doesDvExist = async (): Promise<boolean> => {
            try {
                await datavolumesService.datavolumesApi.getDatavolume(datavolumeId);
                return true;
            } catch (e: any) {
                if (e?.statusCode === StatusCodes.NOT_FOUND) {
                    logger.debug(`datavolume ${datavolumeId} not found yet, will retry`);
                    return false;
                }
                throw e;
            }
        };

        await waitUtils.waitForCondition(
            doesDvExist,
            `waiting for datavolume ${datavolumeId} to exist`,
            true,
            60,
            5
        );
    }

    async waitForApiDatavolumeToBeDeleted(datavolumeId: string): Promise<void> {
        const doesDvExist = async (): Promise<boolean> => {
            try {
                await datavolumesService.datavolumesApi.getDatavolume(datavolumeId);
                return true;
            } catch (e) {
                if (e?.statusCode === StatusCodes.NOT_FOUND || e?.statusCode === StatusCodes.FORBIDDEN) {
                    return false;
                }
                logger.info(`GET datavolume ${datavolumeId} got error that is not NOT FOUND - ${e}`);
                return true;
            }
        };

        await waitUtils.waitForCondition(
            doesDvExist,
            `waiting for datavolume ${datavolumeId} to be deleted`,
            false,
            60,
            5
        );
    }

    //////////////////// CRD Utils ////////////////////
    async waitForDatavolumeCrd(datavolumeName: string, desiredResult = true): Promise<object> {
        return await waitUtils.waitForCrd(crds.datavolumeCrd, datavolumeName, undefined, 30, desiredResult);
    }

    async validateCrdDatavolumePhase(datavolumeName: string, expectedPhase: string): Promise<boolean> {
        const dv = await waitUtils.waitForCrd(crds.datavolumeCrd, datavolumeName);
        const phase = dv['status']?.phase;
        logger.debug(`datavolume crd ${datavolumeName} has phase ${phase}`);
        return phase === expectedPhase;
    }

    async waitForCrdDatavolumeToBeInPhase(datavolumeName: string, expectedPhase: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => this.validateCrdDatavolumePhase(datavolumeName, expectedPhase),
            `waiting for datavolume ${datavolumeName} to be in phase ${expectedPhase}`,
            true,
            120,
            5
        );
    }

    async waitForCrdDatavolumeToBeReady(datavolumeName: string): Promise<void> {
        await this.waitForCrdDatavolumeToBeInPhase(datavolumeName, 'Ready');
    }

    async waitForCrdDatavolumeToBeNotReady(datavolumeName: string): Promise<void> {
        await this.waitForCrdDatavolumeToBeInPhase(datavolumeName, 'NotReady');
    }
}
