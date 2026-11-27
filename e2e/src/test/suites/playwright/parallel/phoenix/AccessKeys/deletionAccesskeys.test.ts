/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    accessKeyUtils,
    authorizationUtils,
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    TokenObject,
    usersManager
} from '../../../../../../modules/helpers';
import {
    AccessKeyPostResponse,
    UserPostResponse
} from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';

test.describe(
    `Access Keys Delete tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.identityManager] },
    function () {
        let accessKeyName: string;
        let createdAccessKey: { response: IncomingMessage; body: AccessKeyPostResponse };
        let editorUser: { response: IncomingMessage; body: UserPostResponse };
        let editorUserToken: TokenObject;

        test.beforeEach(async function () {
            editorUser = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix('editoruserdelete'),
                false
            );
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.body.username,
                flowRootState.users.editor.role.id
            );
            editorUserToken = await authUtils.authenticateUser(
                editorUser.body.username,
                editorUser.body.tempPassword
            );
            accessKeyName = objectGenerator.generateShortNameByPrefix('accesskey');
            createdAccessKey = await accessKeyUtils.createAccessKey(editorUserToken.accessToken, {
                name: accessKeyName
            });
        });

        test.afterEach(async function () {
            try {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    editorUser.body.id
                );
            } catch {
                logger.debug('User might already be deleted in the test');
            }
        });

        test('should delete user and validate his access key deleted', async function () {
            const validateAccessKeyExists = await accessKeyUtils.getAccessKeyById(
                editorUserToken.accessToken,
                createdAccessKey.body.id
            );
            expect(validateAccessKeyExists.body.id, 'access key id should match the created id before user deletion').to.be.equal(createdAccessKey.body.id);
            expect(validateAccessKeyExists.response.statusCode, 'get access key by id should return 200 OK before user deletion').to.be.equal(StatusCodes.OK);
            const validateAccessKeyExistWithAdminUser = await accessKeyUtils.getAccessKeysAdministration(
                flowRootState.users.system_administrator.accessToken,
                createdAccessKey.body.clientId
            );
            expect(validateAccessKeyExistWithAdminUser.body.accessKeys[0].id, 'admin should see access key id before user deletion').to.be.equal(
                createdAccessKey.body.id
            );
            expect(validateAccessKeyExistWithAdminUser.response.statusCode, 'admin get access keys should return 200 OK before user deletion').to.be.equal(StatusCodes.OK);

            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.body.id
            );

            try {
                await accessKeyUtils.getAccessKeyById(
                    flowRootState.users.system_administrator.accessToken,
                    createdAccessKey.body.id
                );
                expect.fail('Access key should be deleted after user deletion');
            } catch (error) {
                expect(error.response.statusCode, 'access key should return 404 Not Found after its owner is deleted').to.be.equal(StatusCodes.NOT_FOUND);
            }
        });
    }
);
