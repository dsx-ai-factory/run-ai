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
    User,
    usersManager
} from '../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { assert } from 'chai';

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

import { AccessKeyPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/accessKeyPostResponse';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';

import { test, expect } from '@playwright/test';

test.describe(`CLI V2 login tests  ${tags.teams.viper}`, { tag: [tags.services.runaiCli] }, function () {
    const loginParams = [
        {
            username: 'system_administrator',
            user: (): User => flowRootState.users.system_administrator
        },
        {
            username: 'l1_researcher',
            user: (): User => flowRootState.users.l1_researcher
        },
        {
            username: 'research_manager',
            user: (): User => flowRootState.users.research_manager
        }
    ];
    const successfulAuthMessage = 'Authentication was successful\n';
    let browserEngine: BrowserEngine;
    let videoName: string;
    test.describe('Login tests with playwright', function () {
        async function startBrowser(): Promise<void> {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
        }

        test.afterEach(async function () {
            try {
                await browserEngine.stopAndArchive(this, 'workspaces-' + videoName);
            } catch (e) {
                logger.error(e);
            }
        });

        loginParams.forEach(function (user) {
            test(`test login via browser as ${user.username}`, async function () {
                await startBrowser();
                const result = await runaiCliV2.loginBrowser(browserEngine, user.user());
                expect(result.stdout).toContain(successfulAuthMessage);
            });
        });

        loginParams.forEach(function (user) {
            test(`test login via sso as ${user.username}`, async function () {
                await startBrowser();
                const result = await runaiCliV2.loginSSOv2(browserEngine, user.user());
                expect(result.stdout).toContain(successfulAuthMessage);
            });
        });
    });
    loginParams.forEach(function (user) {
        test(`login as ${user.username}`, function () {
            logger.debug(`Testing login for user ${user.username}`);
            const stdout = runaiCliV2.loginAsUser(user.user().email, user.user().password).stdout;
            expect(stdout).toContain(successfulAuthMessage);
        });
    });
    test.describe('Access key cli tests', function () {
        let editorApp: AccessKeyPostResponse;
        let noPermissionApp: AccessKeyPostResponse;
        let editorUser: UserPostResponse;
        let noPermissionUser: UserPostResponse;
        const clusterListNoAuthMsg =
            'Cluster list failed. You have insufficient permissions. Ensure you have the required role/permissions to run this command';

        async function createEditorApp(): Promise<void> {
            editorUser = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix(objectGenerator.generateShortNameByPrefix('editor', 3)),
                    false
                )
            ).body;
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.idToken,
                editorUser.username,
                flowRootState.users.editor.role.id
            );
            const editorUserToken = await authUtils.authenticateUser(
                editorUser.username,
                editorUser.tempPassword
            );
            editorApp = (
                await accessKeyUtils.createAccessKey(
                    editorUserToken.accessToken,
                    { name: objectGenerator.generateShortNameByPrefix('editor') }
                )
            ).body;
        }

        async function createAppWithoutPermissions(): Promise<void> {
            noPermissionUser = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix(
                        objectGenerator.generateShortNameByPrefix('noPermission', 3)
                    ),
                    false
                )
            ).body;
            const noPermissionUserToken = await authUtils.authenticateUser(
                noPermissionUser.username,
                noPermissionUser.tempPassword
            );
            noPermissionApp = (
                await accessKeyUtils.createAccessKey(
                    noPermissionUserToken.accessToken,
                    { name: objectGenerator.generateShortNameByPrefix('no-permission') }
                )
            ).body;
        }

        test.afterAll(async function () {
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.id
            );
        });

        test('validate editor user app', async function () {
            await createEditorApp();
            const stdout = runaiCliV2.loginWithAccessKey(editorApp.clientId, editorApp.clientSecret).stdout;
            expect(stdout).toContain(successfulAuthMessage);
            const listResult = runaiCliV2.list('cluster');
            expect(listResult.stdout).toContain(flowRootState.cluster.name);
        });
        test('validate no permission user app', async function () {
            await createAppWithoutPermissions();
            const stdout = runaiCliV2.loginWithAccessKey(noPermissionApp.clientId, noPermissionApp.clientSecret).stdout;
            expect(stdout).toContain(successfulAuthMessage);
            assert.throw(() => runaiCliV2.list('cluster'), clusterListNoAuthMsg);
        });
    });
});
