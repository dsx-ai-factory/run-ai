/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    accessRulesClusterUtils,
    authorizationUtils,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    usersManager
} from '../../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';
import { expect as expectPlaywright } from '@playwright/test';
import { expect as expectChai } from 'chai';
import { IncomingMessage } from 'http';

test.describe(
    `Users Create Update Delete API, tags: ${tags.resource.users} ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.identityManager, tags.runOn.parallelCI] },
    function () {
        let user: { response: IncomingMessage; body: UserPostResponse };

        test.beforeEach('create user', async function () {
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix(`usertests-${Date.now()}`),
                false
            );
        });

        test.afterEach('delete created user', async function () {
            try {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.id
                );
            } catch (e) {
                logger.error(
                    `Failed to delete user or user is deleted on one of the tests ${JSON.stringify(e.message)}`
                );
            }
        });

        test('should create user by admin successfully', function () {
            expectChai(user.body.username).to.contains('usertests');
            expectChai(user.body.id).to.not.be.undefined;
            expectChai(user.body.tempPassword).to.not.be.null;
        });

        test('should update user by admin successfully', async function () {
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.username,
                flowRootState.users.editor.role.id
            );
            const userUpdated = await accessRulesClusterUtils.getAccessRulesByUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.username
            );
            expectPlaywright(userUpdated).toEqual(1);
        });

        test('should deleted user by admin successfully', async function () {
            const deletedUser = await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
            expectPlaywright(deletedUser.response.statusCode).toEqual(StatusCodes.NO_CONTENT);
        });
    }
);
