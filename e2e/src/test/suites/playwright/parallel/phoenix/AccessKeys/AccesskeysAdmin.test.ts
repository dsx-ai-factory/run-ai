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
    serviceAccountUtils,
    tags,
    TokenObject,
    usersManager
} from '../../../../../../modules/helpers';
import { IncomingMessage } from 'http';
import {
    AccessKeyPostResponse,
    UserPostResponse
} from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { expect } from 'chai';
import { test } from '@playwright/test';

test.describe(
    `Access Keys Administration tags:${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.identityManager] },
    function () {
        let editorUser: { response: IncomingMessage; body: UserPostResponse };
        let editorUserToken: TokenObject;
        let accessKeyName: string;
        let createdAccessKey: { response: IncomingMessage; body: AccessKeyPostResponse };

        test.beforeEach(async function () {
            accessKeyName = objectGenerator.generateShortNameByPrefix('accesskey');
            editorUser = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix('editoruser'),
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
            createdAccessKey = await accessKeyUtils.createAccessKey(editorUserToken.accessToken, {
                name: accessKeyName
            });
        });

        test.afterEach(async function () {
            try {
                await authUtils.deleteUser(
                    editorUser.body.id,
                    flowRootState.users.system_administrator.accessToken
                );
                await serviceAccountUtils.deleteAllServiceAccounts(
                    flowRootState.users.system_administrator.accessToken,
                    [accessKeyName, 'tenantapp', 'test-app']
                );
            } catch (e) {
                logger.error('Failed to delete service account' + e);
            }
        });

        test('should validate admin can get access keys by clientID', async function () {
            const getGetAppsByAdmin = await accessKeyUtils.getAccessKeysAdministration(
                flowRootState.users.system_administrator.accessToken,
                createdAccessKey.body.clientId
            );
            expect(getGetAppsByAdmin.body.accessKeys, 'admin get access keys by clientId should return an array').to.be.an('array');
            expect(getGetAppsByAdmin.body.accessKeys[0].name, 'first result should match the created access key name').to.equal(accessKeyName);
            expect(getGetAppsByAdmin.response.statusCode, 'admin get access keys by clientId should return 200 OK').to.equal(200);
        });

        test('should validate admin can get access keys by createdBy', async function () {
            const getGetAppsByAdmin = await accessKeyUtils.getAccessKeysAdministration(
                flowRootState.users.system_administrator.accessToken,
                undefined,
                editorUser.body.username
            );
            expect(getGetAppsByAdmin.body.accessKeys, 'admin get access keys by createdBy should return an array').to.be.an('array');
            expect(getGetAppsByAdmin.body.accessKeys[0].name, 'first result should match the created access key name').to.equal(accessKeyName);
            expect(getGetAppsByAdmin.response.statusCode, 'admin get access keys by createdBy should return 200 OK').to.equal(200);
        });

        test('should validate admin gets 400 when using both createdBy and clientID', async function () {
            try {
                await accessKeyUtils.getAccessKeysAdministration(
                    flowRootState.users.system_administrator.accessToken,
                    createdAccessKey.body.clientId,
                    editorUser.body.username
                );
                expect.fail('Should not reach this line');
            } catch (error) {
                expect(error.response.statusCode, 'using both createdBy and clientId should return 400 Bad Request').to.equal(400);
            }
        });

        test('should validate tenant application cannot be fetched by access keys administration API', async function () {
            try {
                const tenantApp = await serviceAccountUtils.createServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    { name: objectGenerator.generateShortNameByPrefix('tenantapp') }
                );
                const tenantServiceAccount = await serviceAccountUtils.getServiceAccountById(
                    flowRootState.users.system_administrator.accessToken,
                    tenantApp.body.id
                );
                await accessKeyUtils.getAccessKeysAdministration(
                    flowRootState.users.system_administrator.accessToken,
                    tenantServiceAccount.body.clientId
                );

                expect.fail('Should not reach this line');
            } catch (error) {
                expect(error.response.statusCode, 'fetching tenant application via access keys admin API should return 404 Not Found').to.equal(404);
            }
        });

        test('should validate tenant application cannot be deleted by user access key delete API', async function () {
            try {
                const tenantApp = await serviceAccountUtils.createServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    { name: objectGenerator.generateShortNameByPrefix('tenantapp') }
                );
                await accessKeyUtils.deleteAccessKey(
                    flowRootState.users.system_administrator.accessToken,
                    tenantApp.body.id
                );
                expect.fail('Should not reach this line');
            } catch (error) {
                expect(error.response.statusCode, 'deleting tenant application via user access key delete API should return 404 Not Found').to.equal(404);
            }
        });

        test('should validate admin can delete access keys of other users', async function () {
            const appDeleted = await accessKeyUtils.deleteAccessKeyAdministration(
                flowRootState.users.system_administrator.accessToken,
                createdAccessKey.body.id
            );
            expect(appDeleted.response.statusCode, 'admin delete access key of another user should return 204 No Content').to.equal(204);
        });

        test('should validate admin cannot delete tenant app by access keys administration delete API', async function () {
            const tenantApp = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: objectGenerator.generateShortNameByPrefix('tenantapp') }
            );
            try {
                await accessKeyUtils.deleteAccessKeyAdministration(
                    flowRootState.users.system_administrator.accessToken,
                    tenantApp.body.id
                );
                expect.fail('Should not reach this line');
            } catch (error) {
                expect(error.response.statusCode, 'deleting tenant app via admin access keys delete API should return 404 Not Found').to.equal(404);
            }
        });
    }
);
