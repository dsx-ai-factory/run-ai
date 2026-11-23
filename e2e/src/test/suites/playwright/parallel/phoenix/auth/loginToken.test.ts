/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    usersManager,
    Utils
} from '../../../../../../modules/helpers';
import { tags } from '../../../../../../modules/helpers/config/tags';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { test } from '@playwright/test';
import { IncomingMessage } from 'node:http';

test.describe(
    `Authentication API Token Validation, tags: ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit}`,
    { tag: [tags.services.identityManager, tags.services.keycloak, tags.runOn.parallelCI] },
    function () {
        let userName: string;
        let password: string;
        let user: { response: IncomingMessage; body: UserPostResponse };

        test.beforeAll(async function () {
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix(`tokentest-${Date.now()}`),
                false
            );
            userName = user.body.username;
            password = user.body.tempPassword;
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                userName,
                flowRootState.users.system_administrator.role.id
            );
        });

        test.afterAll(async function () {
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
        });

        test('should authenticates successfully with correct username and password', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            expect(token.body.accessToken, 'successful authentication should return a non-null accessToken').to.not.be.null;
            expect(token.body.refreshToken, 'successful authentication should return a non-null refreshToken').to.not.be.null;
            expect(token.body.idToken, 'successful authentication should return a non-null idToken').to.not.be.null;
        });

        test('should fails authentication with incorrect password', async function () {
            try {
                await authUtils.authenticateTokenApi(userName, 'wrongPassword');
                expect.fail('Expected to fail fail auth with wrong user password');
            } catch (error) {
                expect(error.statusCode, 'authentication with wrong password should return 401 Unauthorized').to.equal(401);
            }
        });

        test('should fails authentication with non-existent username', async function () {
            try {
                await authUtils.authenticateTokenApi('dfdfdfdsfsdfsdfdsfsdf', password);
                expect.fail('Expected to fail fail auth with wrong user password');
            } catch (error) {
                expect(error.statusCode, 'authentication with non-existent username should return 401 Unauthorized').to.equal(401);
            }
        });

        test('should handles multiple concurrent authentication requests successfully', async function () {
            const authPromises = [];
            for (let i = 0; i < 5; i++) {
                await Utils.delay(1000);
                authPromises.push(authUtils.authenticateTokenApi(userName, password));
                logger.debug(`Started auth request ${i}`);
            }

            const results = await Promise.all(authPromises);
            results.forEach((result) => {
                expect(result.body.accessToken, 'each concurrent auth response should have a non-null accessToken').to.not.be.null;
                expect(result.body.refreshToken, 'each concurrent auth response should have a non-null refreshToken').to.not.be.null;
                expect(result.body.idToken, 'each concurrent auth response should have a non-null idToken').to.not.be.null;
            });
        });

        test('should validate token expiration time', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            const expMin = authUtils.getTokenExpMin(token.body.accessToken);
            expect(expMin, 'token expiry time in minutes should be greater than 0').to.be.greaterThan(0);
            expect(authUtils.isJwtAboutToExpire(token.body.accessToken, 10), 'freshly issued token should not be about to expire within 10 minutes').to.be.false;
        });

        test('should validate token contains required claims', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            const expectedClaims = {
                sub: userName,
                email: user.body.username,
                scope: 'openid profile email'
            };
            expect(authUtils.validateAllParamsInToken(token.body.accessToken, expectedClaims), 'token should contain all required claims (sub, email, scope)').to.be.true;
        });

        test('should validate token contains groups claim', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            expect(authUtils.validateAllParamsInToken(token.body.accessToken, { groups: [] }), 'token should not have an empty groups claim').to.be.false;
        });
    }
);
