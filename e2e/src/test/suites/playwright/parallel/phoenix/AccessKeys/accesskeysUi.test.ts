/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    authorizationUtils,
    flowRootState,
    logger,
    usersManager,
    objectGenerator,
    tags
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { LocalUserSettings } from '../../../../../../modules/ui/runaiV2/pages/localUserSettings';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { IncomingMessage } from 'http';

test.describe(`Access Keys UI tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let localUserSettings: LocalUserSettings;
        let editorUser: { response: IncomingMessage; body: UserPostResponse };
        let videoName: string;

        test.beforeAll('init browserEngine', function () {
            browserEngine = new BrowserEngine();
        });

        test.beforeEach('init browsers and classes', async function () {
            editorUser = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix('editoruserui'),
                false
            );
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                editorUser.body.username,
                flowRootState.users.editor.role.id
            );
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            await auth.signIn(editorUser.body.username, editorUser.body.tempPassword);
            localUserSettings = new LocalUserSettings(browserEngine.currentPage);
        });

        test.afterEach('stop browserEngine', async ({}, testInfo) => {
            try {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    editorUser.body.id
                );
            } catch (e) {
                logger.error(`Failed to sign out: ${e}`);
            } finally {
                await browserEngine.stopAndArchive(testInfo, videoName);
            }
        });

        test('should create and delete access key by ui', async () => {
            videoName = 'create-delete-access-key';
            const accessKeyName = objectGenerator.generateShortNameByPrefix('accesskeyui');
            await localUserSettings.openLocalSettingsPage();
            const createdAccessKey = await localUserSettings.createAccessKeyUi(accessKeyName);
            expect(createdAccessKey, 'createAccessKeyUi should return the created access key name').to.not.be.undefined;
            expect(createdAccessKey, 'created access key name shown in UI should match the submitted name').to.eql(accessKeyName);
            logger.info(`created access key: ${createdAccessKey}`);
            await localUserSettings.deleteAccessKeyUi();
            const deleted = await localUserSettings.validateAccessKeyUiNoExist(accessKeyName);
            expect(deleted, 'access key should not exist in UI after deletion').to.be.false;
            logger.info(`access key deleted, access key exist: ${deleted ? 'true' : 'false'}`);
        });

        test('should regenerate secret access key by ui', async () => {
            videoName = 'regenerate-secret-access-key';
            const accessKeyName = objectGenerator.generateShortNameByPrefix('accesskeyui');
            await localUserSettings.openLocalSettingsPage();
            const createdAccessKey = await localUserSettings.createAccessKeyUi(accessKeyName);
            expect(createdAccessKey, 'createAccessKeyUi should return the created access key name').to.not.be.undefined;
            expect(createdAccessKey, 'created access key name shown in UI should match the submitted name').to.eql(accessKeyName);
            logger.info(`created access key: ${createdAccessKey}`);
            const validationRegenarate = await localUserSettings.regenerateClientSecret();
            expect(validationRegenarate, 'regenerateClientSecret UI action should return true on success').to.be.true;
            await localUserSettings.deleteAccessKeyUi();
            const deleted = await localUserSettings.validateAccessKeyUiNoExist(accessKeyName);
            expect(deleted, 'access key should not exist in UI after deletion following regeneration').to.be.false;
            logger.info(`access key deleted, access key exist: ${deleted ? 'true' : 'false'}`);
        });
    });
});
