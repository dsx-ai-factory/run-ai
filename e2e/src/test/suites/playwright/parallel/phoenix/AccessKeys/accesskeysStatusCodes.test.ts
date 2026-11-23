/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    accessKeyUtils,
    flowRootState,
    objectGenerator,
    tags
} from '../../../../../../modules/helpers';
import {
    AccessKeyPostResponse
} from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

test.describe(`Access Keys Status Codes tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`, { tag: [tags.services.identityManager] }, function () {
    let accessKeyName: string;
    let createdAccessKey: { response: IncomingMessage; body: AccessKeyPostResponse };

    test.beforeEach(async function () {
        accessKeyName = objectGenerator.generateShortNameByPrefix(`accesskey-${Date.now()}`);
        createdAccessKey = await accessKeyUtils.createAccessKey(
            flowRootState.users.system_administrator.accessToken,
            { name: accessKeyName }
        );
    });

    test.afterEach(async function () {
        try {
            await accessKeyUtils.deleteAccessKey(
                flowRootState.users.system_administrator.accessToken,
                createdAccessKey.body.id
            );
        } catch (e: any) {
            if (e?.response?.statusCode !== StatusCodes.NOT_FOUND) {
                throw e;
            }
        }
    });

    test('should return unauthorized (401) when getting all access keys with an invalid token', async function () {
        try {
            await accessKeyUtils.getAccessKeys('invalidToken');
            expect.fail('get access keys with invlaid token should fail');
        } catch (e) {
            expect(e.response.statusCode, 'get all access keys with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
        }
    });

    test('should return unauthorized (401) when creating an access key with an invalid token', async function () {
        try {
            await accessKeyUtils.createAccessKey('invalidToken', {
                name: objectGenerator.generateNameByPrefix('statusaccesskey')
            });
            expect.fail('should not create an access key with invalid token');
        } catch (e) {
            expect(e.response.statusCode, 'creating access key with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
        }
    });

    test('should return unauthorized (401) when on get access key by id with an invalid token', async function () {
        try {
            await accessKeyUtils.getAccessKeyById('invalidToken', createdAccessKey.body.id);
            expect.fail('should not get an access key by id with invalid token');
        } catch (e) {
            expect(e.response.statusCode, 'get access key by id with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
        }
    });

    test('should return unauthorized (401) when regenerating access key secret with an invalid token', async function () {
        try {
            await accessKeyUtils.regenerateAccessKeySecret('invalidToken', createdAccessKey.body.id);
            expect.fail('regenerate access key secret with invalid token should fail');
        } catch (e) {
            expect(e.response.statusCode, 'regenerate access key secret with invalid token should return 401 Unauthorized').to.be.equal(StatusCodes.UNAUTHORIZED);
        }
    });
});
