/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { authorizationUtils, tags } from '../../../../../../modules/helpers';
import {
    accessRulesFilterCases,
    getAccessRuleFieldValue
} from '../../../../../../modules/helpers/utils/AccessRulesFilterUtils';
import { test } from '@playwright/test';

test.describe(
    'Access rules testing',
    { tag: [tags.teams.loong, tags.tool.api, tags.resource.accessRules, tags.runOn.serialCI] },
    () => {
        test('T5179384: access rules filter testing', async () => {
            for (const testStep of accessRulesFilterCases) {
                await test.step(testStep.name, async () => {
                    const response = await authorizationUtils.getAccessRules({ filterBy: testStep.filterBy });
                    response.accessRules.forEach((accessRule) => {
                        testStep.matcher(
                            getAccessRuleFieldValue(accessRule, testStep.field),
                            testStep.expected
                        );
                    });
                });
            }
        });
    }
);
