/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    tags
} from '../../../../../../../modules/helpers';
import { idpNameOcp } from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { expect } from 'chai';
import { Platform, supportedPlatforms } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

supportedPlatforms([Platform.OPENSHIFT])(function () {
    test.describe(`Openshift SSO IDP tests, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, function () {
        test.afterAll(async function () {
            try {
                await idpUtils.deleteIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOcp
                );
            } catch (error) {
                logger.error(error);
            }
        });

        test('should create Openshift SSO idp and login for token', async function () {
            await Keycloak.setupSSOOCP();
            const token = await authUtils.authenticateSSOOCPUser(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword
            );
            expect(token, 'OpenShift SSO user should receive a non-null token after login').to.not.be.null;
        });
    });
});
