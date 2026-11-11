/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
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

test.describe(`Users Status Codes, tags: ${tags.resource.users} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.identityManager] }, function () {
    let user: { response: IncomingMessage; body: UserPostResponse };

    test.beforeEach(async function () {
        user = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix(`userstatuscodes-${Date.now()}`),
            false
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

    test('should fail delete user as UNAUTHORIZED', async function () {
        const error: any = await usersManager.deleteLocalUser(
            'UNAUTHORIZED',
            user.body.id
        ).then(
            () => expectChai.fail('delete user as unauthorized should fail'),
            (e: any) => e
        );
        expectPlaywright(error.statusCode).toEqual(StatusCodes.UNAUTHORIZED);
    });

    test('should fail create user as UNAUTHORIZED', async function () {
        const error: any = await usersManager.createLocalUser(
            'UNAUTHORIZED',
            objectGenerator.generateEmailByPrefix(`userstatuscodes-${Date.now()}`),
            false
        ).then(
            () => expectChai.fail('create user as unauthorized should fail'),
            (e: any) => e
        );
        expectPlaywright(error.statusCode).toEqual(StatusCodes.UNAUTHORIZED);
    });
});
