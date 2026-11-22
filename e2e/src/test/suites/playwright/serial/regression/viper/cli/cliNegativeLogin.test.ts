/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { ServiceAccountPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/serviceAccountPostResponse';
import { IncomingMessage } from 'http';
import { test, expect } from '@playwright/test';

test.describe(`CLI V2  Negative login tests  ${tags.teams.viper}`, { tag: [tags.services.runaiCli] }, function () {
    const failedAuthMessage = 'Error: Authenticate failed. Authentication failed';
    const missArgMessage = 'Error: flag needs an argument';
    const wrongUsername = 'wrongUser';
    const wrongPassword = 'password';
    test.describe('Login tests with username', function () {
        test('wrong username and pass', () => {
            const cmdOutput = runaiCliV2.loginAsUser(wrongUsername, wrongPassword, false);
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
        test('valid username wrong pass', () => {
            const cmdOutput = runaiCliV2.loginAsUser(
                flowRootState.users.system_administrator.email,
                wrongPassword,
                false
            );
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
        test('valid pass wrong username', () => {
            const cmdOutput = runaiCliV2.loginAsUser(
                wrongUsername,
                flowRootState.users.system_administrator.password,
                false
            );
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
        test('valid username missing pass', () => {
            const cmdOutput = runaiCliV2.loginAsUser(
                flowRootState.users.system_administrator.email,
                '',
                false
            );
            expect(cmdOutput.stderr).toContain(missArgMessage);
            expect(cmdOutput.stdout).toBe('');
        });
    });
    test.describe('Login tests with access-key', function () {
        const failedAuthMessage = 'Error: authentication server issue. Authentication failed';
        let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };
        const serviceAccountName = objectGenerator.generateShortNameByPrefix('login-access-key');
        test.beforeAll(async function () {
            serviceAccount = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: serviceAccountName }
            );
        });
        test('wrong client-id valid secret', function () {
            const cmdOutput = runaiCliV2.loginWithAccessKey(
                wrongUsername,
                serviceAccount.body.clientSecret,
                false
            );
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
        test('wrong secret valid client-id', function () {
            const cmdOutput = runaiCliV2.loginWithAccessKey(
                serviceAccount.body.clientId,
                wrongPassword,
                false
            );
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
        test('wrong client-id and secret', function () {
            const cmdOutput = runaiCliV2.loginWithAccessKey(wrongUsername, wrongPassword, false);
            expect(cmdOutput.stderr).toContain(failedAuthMessage);
            expect(cmdOutput.stdout).toBe('');
        });
    });
    test.describe('Login tests with sso', function () {
        const ssoFailedAuthMessage = 'Authenticate failed. failed to exchange code for token';
        let browserEngine: BrowserEngine;
        let videoName: string;
        test.beforeEach(async function () {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
        });
        test.afterEach('stop browser', async function () {
            await browserEngine.stopAndArchive(this, 'policy-' + videoName);
        });

        test('enter wrong token', async function () {
            const cmdOutput = await runaiCliV2.loginSSO(browserEngine, 'system_administrator', wrongPassword);
            expect(cmdOutput.stderr).toContain(ssoFailedAuthMessage);
        });
    });
});
