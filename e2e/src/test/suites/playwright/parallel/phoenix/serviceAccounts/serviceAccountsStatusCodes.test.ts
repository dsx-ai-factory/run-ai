/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    flowRootState,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import {
    ServiceAccountPostResponse
} from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

test.describe(`Service Accounts Status Codes tags: ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.identityManager] }, function () {
    let serviceAccountName: string;
    let createdServiceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };

    test.beforeEach(async function () {
        serviceAccountName = objectGenerator.generateShortNameByPrefix(`serviceaccount-${Date.now()}`);
        createdServiceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: serviceAccountName }
        );
    });

    test.afterEach(async function () {
        try {
            await serviceAccountUtils.deleteServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                createdServiceAccount.body.id
            );
        } catch (e: any) {
            if (e?.response?.statusCode !== StatusCodes.NOT_FOUND) {
                throw e;
            }
        }
    });

    test('should validate user can not fetch service account by id', async function () {
        const error: any = await serviceAccountUtils.getServiceAccountById(
            flowRootState.users.viewer.accessToken,
            createdServiceAccount.body.id
        ).then(
            () => expect.fail('get service account by id with another user should fail'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'get service account by id with viewer user should return 403 Forbidden').to.be.equal(StatusCodes.FORBIDDEN);
    });

    test('should return unauthorized (401) when getting all service accounts with an invalid token', async function () {
        const error: any = await serviceAccountUtils.getServiceAccounts('invalidToken').then(
            () => expect.fail('get service accounts with invalid token should fail'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'get all service accounts with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });

    test('should return unauthorized (401) when creating a service account with an invalid token', async function () {
        const error: any = await serviceAccountUtils.createServiceAccount('invalidToken', {
            name: objectGenerator.generateNameByPrefix('statusserviceaccount')
        }).then(
            () => expect.fail('should not create a service account with invalid token'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'creating service account with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });

    test('should return unauthorized (401) when on get service account by id with an invalid token', async function () {
        const error: any = await serviceAccountUtils.getServiceAccountById('invalidToken', createdServiceAccount.body.id).then(
            () => expect.fail('should not get a service account by id with invalid token'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'get service account by id with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });

    test('should return unauthorized (401) when regenerating service account secret with an invalid token', async function () {
        const error: any = await serviceAccountUtils.regenerateServiceAccountSecret(
            'invalidToken',
            createdServiceAccount.body.id
        ).then(
            () => expect.fail('regenerate service account secret with invalid token should fail'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'regenerate secret with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });

    test('should return unauthorized when regenerating service account secret with an invalid token by id', async function () {
        const error: any = await serviceAccountUtils.regenerateServiceAccountSecret(
            'invalidToken',
            createdServiceAccount.body.id
        ).then(
            () => expect.fail('regenerate service account secret with invalid token by id should fail'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'regenerate secret by id with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });

    test('should return unauthorized (401) when updating service account by id with an invalid token', async function () {
        const error: any = await serviceAccountUtils.updateServiceAccountById(
            'invalidToken',
            createdServiceAccount.body.id,
            { enabled: false }
        ).then(
            () => expect.fail('should not update service account by id with invalid token'),
            (e: any) => e
        );
        expect(error.response.statusCode, 'update service account with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
    });
});
