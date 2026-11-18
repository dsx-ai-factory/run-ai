/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { identityManagerApi } from '../../api';
import {
    RegenerateServiceAccountSecretResponse,
    ServiceAccount,
    ServiceAccountCreationRequest,
    ServiceAccountPatchRequest,
    ServiceAccountPostResponse
} from '../../api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { Utils } from './utils';
import { logger } from '../index';

export class ServiceAccountUtils {
    async getServiceAccounts(
        accessToken: string
    ): Promise<{ response: IncomingMessage; body: Array<ServiceAccount> }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const serviceAccounts = await identityManagerApi.serviceAccountsApi.getServiceAccounts();
        return serviceAccounts;
    }

    async createServiceAccount(
        accessToken: string,
        serviceAccountCreationRequest: ServiceAccountCreationRequest
    ): Promise<{ response: IncomingMessage; body: ServiceAccountPostResponse }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const serviceAccount = await identityManagerApi.serviceAccountsApi.createServiceAccount(
            serviceAccountCreationRequest
        );
        await Utils.delay(1000);
        return serviceAccount;
    }

    async deleteServiceAccount(
        accessToken: string,
        serviceAccountId: string
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const deleteServiceAccount =
            await identityManagerApi.serviceAccountsApi.deleteServiceAccountById(serviceAccountId);
        return deleteServiceAccount;
    }

    async deleteAllServiceAccounts(accessToken: string, namePrefix: string[]): Promise<number> {
        const serviceAccounts = await this.getServiceAccounts(accessToken);
        for (const sa of serviceAccounts.body) {
            if (namePrefix.some((prefix) => sa.name.startsWith(prefix))) {
                await this.deleteServiceAccount(accessToken, sa.id);
            }
        }
        logger.info(`Deleted ${serviceAccounts.body.length} service accounts`);
        logger.info(
            `Deleted service accounts: ${serviceAccounts.body.map((sa) => sa.name).join(', ')}`
        );
        return 200;
    }

    async getServiceAccountById(
        accessToken: string,
        serviceAccountId: string
    ): Promise<{ response: IncomingMessage; body: ServiceAccount }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const serviceAccount =
            await identityManagerApi.serviceAccountsApi.getServiceAccountById(serviceAccountId);
        return serviceAccount;
    }

    async getInstallerServiceAccount(accessToken: string): Promise<ServiceAccountPostResponse> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const installer = await identityManagerApi.serviceAccountsApi.getInstallerServiceAccount();
        return installer.body;
    }

    async regenerateServiceAccountSecret(
        accessToken: string,
        serviceAccountId: string
    ): Promise<{ response: IncomingMessage; body: RegenerateServiceAccountSecretResponse }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const serviceAccount =
            await identityManagerApi.serviceAccountsApi.regenerateServiceAccountSecret(serviceAccountId);
        return serviceAccount;
    }

    async updateServiceAccountById(
        accessToken: string,
        serviceAccountId: string,
        serviceAccountPatchRequest?: ServiceAccountPatchRequest
    ): Promise<{
        response: IncomingMessage;
        body?: any;
    }> {
        identityManagerApi.serviceAccountsApi.accessToken = accessToken;
        const serviceAccount = await identityManagerApi.serviceAccountsApi.updateServiceAccountById(
            serviceAccountId,
            serviceAccountPatchRequest
        );
        return serviceAccount;
    }
}
