/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    User,
    UserPostResponse
} from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    authUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    TokenObject,
    usersManager,
    usersUtils
} from '../../../../../../../modules/helpers';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { IncomingMessage } from 'http';

test.describe(`Users Password, tags: ${tags.resource.users} ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.identityManager, tags.services.notificationsService, tags.services.notificationsProxy] }, function () {
    let user: { response: IncomingMessage; body: UserPostResponse };
    let token: TokenObject;
    let getUser: User;

    test.beforeEach(async function () {
        user = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('userpassword'),
            false
        );
        token = await authUtils.authenticateUser(user.body.username, user.body.tempPassword);
        getUser = await usersManager.getUserById(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
    });

    test.afterEach(async function () {
        try {
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
        } catch (e) {
            logger.error(`Failed to delete user or user is deleted ${JSON.stringify(e.message)}`);
        }
    });

    test('should change user password', async function () {
        await usersUtils.changeUserPassword(
            token.idToken,
            user.body.tempPassword,
            config.controlPlaneAdminPassword
        );
        const validation = await usersManager.getUserById(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
        expect(validation.updatedAt, 'user updatedAt should change after password change').to.not.equal(getUser.updatedAt);

        // Wait for password change to propagate in SaaS environments
        await new Promise((resolve) => setTimeout(resolve, 1000));

        token = await authUtils.authenticateUser(user.body.username, config.controlPlaneAdminPassword);
        expect(token, 'authentication with new password should return a token').to.not.be.null;
    });

    test('should reset user password', async function () {
        const resetedPassword = await usersUtils.resetUserPassword(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
        expect(resetedPassword.body, 'password reset response body should not be undefined').to.not.be.undefined;
        expect(resetedPassword.body, 'password reset response body should not be null').to.not.be.null;
        const validation = await usersManager.getUserById(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
        expect(validation.updatedAt, 'user updatedAt should change after password reset').to.not.equal(getUser.updatedAt);
    });

    test('should fail to login with old password after reset password', async function () {
        await usersUtils.resetUserPassword(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );

        // Wait for password reset to propagate in SaaS environments
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
            await authUtils.authenticateUser(user.body.username, user.body.tempPassword);
        } catch (error) {
            expect(error.message, 'logging in with old password after reset should return 401').to.contain('HTTP request failed with the following statusCode 401');
        }
    });

    test('should fail to change password with wrong current password', async function () {
        try {
            await usersUtils.changeUserPassword(
                token.idToken,
                'wrongPassword',
                config.controlPlaneAdminPassword
            );
        } catch (error) {
            expect(error.message, 'changing password with wrong current password should return 401').to.contain('HTTP request failed with the following statusCode 401');
        }
    });

    test('should fail to change password with empty new password', async function () {
        try {
            await usersUtils.changeUserPassword(token.idToken, user.body.tempPassword, '');
        } catch (error) {
            expect(error.message, 'changing password with empty new password should return 400').to.contain('HTTP request failed with the following statusCode 400');
        }
    });

    test('should fail to change password with empty current password', async function () {
        try {
            await usersUtils.changeUserPassword(token.idToken, '', config.controlPlaneAdminPassword);
        } catch (error) {
            expect(error.message, 'changing password with empty current password should return 400').to.contain('HTTP request failed with the following statusCode 400');
        }
    });

    test('should fail to change password with empty current and new password', async function () {
        try {
            await usersUtils.changeUserPassword(token.idToken, '', '');
        } catch (error) {
            expect(error.statusCode, 'changing password with both passwords empty should return 400').to.eql(400);
        }
    });

    test('should fail to login with old password after change password', async function () {
        await usersUtils.changeUserPassword(
            token.idToken,
            user.body.tempPassword,
            config.controlPlaneAdminPassword
        );

        // Wait for password change to propagate in SaaS environments
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
            await authUtils.authenticateUser(user.body.username, user.body.tempPassword);
        } catch (error) {
            expect(error.message, 'logging in with old password after change should return 401').to.contain('HTTP request failed with the following statusCode 401');
        }
    });
});
