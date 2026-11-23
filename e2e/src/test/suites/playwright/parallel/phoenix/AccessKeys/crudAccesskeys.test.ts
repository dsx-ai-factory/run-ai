/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    accessKeyUtils,
    flowRootState,
    logger,
    objectGenerator,
    tags
} from '../../../../../../modules/helpers';
import { AccessKeyPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';

test.describe(
    `Access Keys CRUD tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.identityManager] },
    function () {
        let accessKeyName: string;
        let accessKey: { response: IncomingMessage; body: AccessKeyPostResponse };

        test.beforeEach(async function () {
            accessKeyName = objectGenerator.generateShortNameByPrefix('accesskey');
            accessKey = await accessKeyUtils.createAccessKey(
                flowRootState.users.system_administrator.accessToken,
                { name: accessKeyName }
            );
        });

        test.afterEach(async function () {
            try {
                await accessKeyUtils.deleteAccessKey(
                    flowRootState.users.system_administrator.accessToken,
                    accessKey.body.id
                );
            } catch (e) {
                logger.error(
                    `Failed to delete Access Key with id: ${accessKey.body.id} ${JSON.stringify(e)}`
                );
            }
        });

        test('should create access key', function () {
            expect(accessKey.body.name, 'created access key name should match the submitted name').to.equal(accessKeyName);
            expect(accessKey.body.clientSecret, 'created access key should have a non-empty clientSecret').to.not.be.empty;
            expect(accessKey.body.id, 'created access key should have a non-empty id').to.not.be.empty;
            expect(accessKey.body.clientId, 'created access key should have a non-empty clientId').to.not.be.empty;
        });

        test('should delete access key', async function () {
            const deletedAccessKey = await accessKeyUtils.deleteAccessKey(
                flowRootState.users.system_administrator.accessToken,
                accessKey.body.id
            );
            expect(deletedAccessKey.response.statusCode, 'delete access key should return 204 No Content').to.equal(StatusCodes.NO_CONTENT);
        });
    }
);
