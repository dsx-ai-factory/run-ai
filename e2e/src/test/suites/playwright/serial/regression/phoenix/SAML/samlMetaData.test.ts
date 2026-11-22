/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { settings, tags } from '../../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`SAML Metadata, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, function () {
            test('should fail inserts random text to metadata saml', async function () {
                try {
                    await Keycloak.configLoginWithSSOSAML('testrandomname');
                } catch (error) {
                    expect(error.message, 'inserting random text as SAML metadata should fail with a configuration error').to.contain('Failed to configure SSO');
                }
            });
        });
    }
);
