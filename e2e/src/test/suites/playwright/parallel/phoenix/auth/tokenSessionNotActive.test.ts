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
    usersManager
} from '../../../../../../modules/helpers';
import { tags } from '../../../../../../modules/helpers/config/tags';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { test } from '@playwright/test';
import { ScopeType, SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { IncomingMessage } from 'node:http';
import { StatusCodes } from 'http-status-codes';

test.describe(
    `Token Session Not Active, tags: ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit}`,
    { tag: [tags.services.identityManager, tags.services.keycloak] },
    function () {
        let userName: string;
        let password: string;
        let user: { response: IncomingMessage; body: UserPostResponse };

        test.beforeEach(async function () {
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix(`sessionnotactive-${Date.now()}`),
                false
            );
            userName = user.body.username;
            password = user.body.tempPassword;
            await authorizationUtils.createAccessRuleWithDefaults(
                userName,
                SubjectType.User,
                flowRootState.users.system_administrator.role.id,
                flowRootState.cluster.tenantId.toString(),
                ScopeType.Tenant,
                flowRootState.users.system_administrator.accessToken
            );
        });

        test.afterEach(async function () {
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
        });

        test('should validate refresh token not active error (SESSION_NOT_ACTIVE)', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            logger.info('Token obtained with access and refresh tokens');
            expect(token.body.accessToken, 'initial authentication should return a non-null accessToken').to.not.be.null;
            expect(token.body.refreshToken, 'initial authentication should return a non-null refreshToken').to.not.be.null;
            const refreshedToken = await usersManager.refreshToken(token.body.refreshToken, 'cli');
            expect(refreshedToken.accessToken, 'token refresh should return a non-null accessToken before logout').to.not.be.null;
            expect(refreshedToken.refreshToken, 'token refresh should return a non-null refreshToken before logout').to.not.be.null;
            logger.info('Token refresh successful before logout - session is active');
            const logoutResponse = await usersManager.logoutUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
            expect(logoutResponse.response.statusCode, 'logout user API should return 200 OK').to.equal(StatusCodes.OK);
            logger.info(`User ${userName} session deactivated`);
            try {
                await usersManager.refreshToken(refreshedToken.refreshToken, 'cli');
                expect.fail('Expected token refresh to fail with inactive session');
            } catch (error: any) {
                expect(error.response.statusCode, 'refresh token after session logout should return 400 Bad Request (SESSION_NOT_ACTIVE)').to.equal(StatusCodes.BAD_REQUEST);
                expect(error.body.message, 'error message should indicate invalid input for inactive session').to.include('Invalid input');
            }
        });

        test('should validate maximum refresh token reuse exceeded (MAX_ALLOWED_REFRESH_TOKEN_REUSE)', async function () {
            const token = await authUtils.authenticateTokenApi(userName, password);
            logger.info('Token obtained with access and refresh tokens');
            expect(token.body.accessToken, 'initial authentication should return a non-null accessToken').to.not.be.null;
            expect(token.body.refreshToken, 'initial authentication should return a non-null refreshToken').to.not.be.null;
            const originalRefreshToken = token.body.refreshToken;
            const refreshedToken = await usersManager.refreshToken(originalRefreshToken, 'cli');
            expect(refreshedToken.accessToken, 'first refresh token use should return a non-null accessToken').to.not.be.null;
            expect(refreshedToken.refreshToken, 'first refresh token use should return a new non-null refreshToken').to.not.be.null;
            logger.info('First token refresh successful - original refresh token used once');
            try {
                await usersManager.refreshToken(originalRefreshToken, 'cli');
                expect.fail('Expected token refresh to fail with maximum reuse exceeded');
            } catch (error: any) {
                expect(error.message, 'reusing an already-rotated refresh token should fail (MAX_ALLOWED_REFRESH_TOKEN_REUSE)').to.equal('Expected token refresh to fail with maximum reuse exceeded');
            }
        });
    }
);
