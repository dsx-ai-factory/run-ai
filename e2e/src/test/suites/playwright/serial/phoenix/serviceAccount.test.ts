/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/api';
import { authUtils, config, logger, serviceAccountUtils, settings, tags } from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';

test.describe(`Phoenix tests service accounts SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, () => {
    let token: TokenResponse;

    test.beforeAll(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        const baseUrl = config.tenantUrl.replace('https://', '');
        const tenantUrl = `https://console.${baseUrl}`;

        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );
        logger.debug('Before all Phoenix hook successful ended');
    });

    test('should get service account default', async function () {
        await serviceAccountUtils.getServiceAccounts(token.accessToken).then((serviceAccounts) => {
            expect(serviceAccounts, 'service accounts response should not be null').to.not.be.null;
            expect(serviceAccounts.body.find((serviceAccount) => serviceAccount.name === 'default'), 'a service account named "default" should exist').to.not.be.null;
        });
    });

    test('should regenerate service account secret', async function () {
        await serviceAccountUtils.getServiceAccounts(token.accessToken).then(async (serviceAccounts) => {
            const serviceAccount = serviceAccounts.body.find((serviceAccount) => serviceAccount.name === 'default');
            const secret = await serviceAccountUtils.regenerateServiceAccountSecret(token.accessToken, serviceAccount.id);
            expect(secret, 'regenerated service account secret response should not be null').to.not.be.null;
            expect(secret.body.clientSecret, 'regenerated service account should have a non-null clientSecret').to.not.be.null;
        });
    });
});
