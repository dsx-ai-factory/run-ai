/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { flowRootState, serviceAccountUtils, tags } from '../../../../../../modules/helpers';

test.describe(`New-file selection validation ${tags.teams.phoenix}`, () => {
    test('runs a newly added service-account API test from the PR commit', async () => {
        const response = await serviceAccountUtils.getServiceAccounts(
            flowRootState.users.system_administrator.accessToken
        );
        expect(response).toBeDefined();
    });
});
