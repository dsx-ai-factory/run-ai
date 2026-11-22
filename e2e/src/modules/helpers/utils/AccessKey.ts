/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { identityManagerApi } from '../../api';
import {
    AccessKey,
    AccessKeyCreationRequest,
    AccessKeyPostResponse,
    AccessKeys,
    AccessKeysAdministration,
    RegenerateAccessKeySecret200Response
} from '../../api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { Utils } from './utils';

export class AccessKeysUtils {
    async getAccessKeys(accessToken: string): Promise<{ response: IncomingMessage; body: AccessKeys }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const accessKeys = await identityManagerApi.accessKeysApi.getAccessKeys();
        return accessKeys;
    }

    async createAccessKey(
        accessToken: string,
        accessKeyCreationRequest: AccessKeyCreationRequest
    ): Promise<{ response: IncomingMessage; body: AccessKeyPostResponse }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const accessKey = await identityManagerApi.accessKeysApi.createAccessKey(accessKeyCreationRequest);
        await Utils.delay(1000);
        return accessKey;
    }

    async deleteAccessKey(
        accessToken: string,
        accessKeyId: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const deleteAccessKey = await identityManagerApi.accessKeysApi.deleteAccessKeyById(accessKeyId);
        return deleteAccessKey;
    }

    async getAccessKeyById(
        accessToken: string,
        accessKeyId: string
    ): Promise<{ response: IncomingMessage; body: AccessKey }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const accessKey = await identityManagerApi.accessKeysApi.getAccessKeyById(accessKeyId);
        return accessKey;
    }

    async regenerateAccessKeySecret(
        accessToken: string,
        accessKeyId: string
    ): Promise<{ response: IncomingMessage; body: RegenerateAccessKeySecret200Response }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const accessKey = await identityManagerApi.accessKeysApi.regenerateAccessKeySecret(accessKeyId);
        return accessKey;
    }

    async getAccessKeysAdministration(
        accessToken: string,
        clientId?: string,
        createdBy?: string
    ): Promise<{ response: IncomingMessage; body: AccessKeysAdministration }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const accessKeysAdministration = await identityManagerApi.accessKeysApi.getAccessKeysAdministration(
            clientId,
            createdBy
        );
        return accessKeysAdministration;
    }

    async deleteAccessKeyAdministration(
        accessToken: string,
        accessKeyId: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.accessKeysApi.accessToken = accessToken;
        const deleteAccessKey =
            await identityManagerApi.accessKeysApi.deleteAccessKeyByIdAdministration(accessKeyId);
        return deleteAccessKey;
    }
}
