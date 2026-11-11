/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';

import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`OIDC SSO local users, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, { tag: [tags.services.identityManager, tags.services.keycloak] }, function () {
            test('create local user in sso', async function () {
                const createdUser = await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('localUser'),
                    false
                );

                logger.info('check that the user exists in keycloak');
                const adminTokens = await authUtils.authenticateUser(
                    createdUser.body.username,
                    createdUser.body.tempPassword
                );
                expect(adminTokens, 'local user created with SSO enabled should be able to authenticate and receive a token').to.not.be.null;
            });
        });
    }
);
