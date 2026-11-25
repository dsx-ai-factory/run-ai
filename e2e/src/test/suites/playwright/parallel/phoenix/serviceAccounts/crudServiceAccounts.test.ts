/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';

test.describe(`Service Accounts CRUD tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`, { tag: [tags.services.identityManager] }, function () {
    let serviceAccountName: string;
    let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };

    test.beforeEach(async function () {
        serviceAccountName = objectGenerator.generateShortNameByPrefix(`serviceacc-${Date.now()}`);
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: serviceAccountName }
        );
    });

    test.afterEach(async function () {
        try {
            await serviceAccountUtils.deleteServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                serviceAccount.body.id
            );
        } catch (e) {
            logger.error(
                `Failed to delete Service Account with id: ${serviceAccount.body.id} ${JSON.stringify(e)}`
            );
        }
    });

    test('should create service account', function () {
        expect(serviceAccount.body.name, 'created service account name should match the submitted name').to.equal(serviceAccountName);
        expect(serviceAccount.body.clientSecret, 'created service account should have a non-empty clientSecret').to.not.be.empty;
        expect(serviceAccount.body.id, 'created service account should have a non-empty id').to.not.be.empty;
        expect(serviceAccount.body.clientId, 'created service account should have a non-empty clientId').to.not.be.empty;
    });

    test('should delete service account', async function () {
        const deletedServiceAccount = await serviceAccountUtils.deleteServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            serviceAccount.body.id
        );
        expect(deletedServiceAccount.response.statusCode, 'delete service account should return 204 No Content').to.equal(StatusCodes.NO_CONTENT);
    });
});
