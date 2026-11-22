/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { authV2Utils, flowRootState, tags } from '../../../../../../modules/helpers';
import { test } from '@playwright/test';

test.describe(`Login Screen Validation, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api}`, () => {
    test('should validate retrieval of the login screen URL', async () => {
        const response = await authV2Utils.getLoginScreenURL(
            flowRootState.users.system_administrator.accessToken,
            'test',
            'test'
        );
        expect(response, 'login screen URL endpoint should return 200').to.equal(200);
    });
});
