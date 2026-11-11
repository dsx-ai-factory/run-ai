/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { authorizationUtils, config, tags } from '../../../../../../modules/helpers';

import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';

test.describe(`Access rules testing, tags: ${tags.teams.loong} ${tags.tool.api} ${tags.resource.accessRules}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        test('T5179384: access rules filter testing', async () => {
            const filterTests = [
                {
                    name: 'test type filter',
                    filterBy: ['subjectType!=app,subjectType!=group'],
                    field: 'subjectType',
                    expected: 'user',
                    matcher: (value: string, expected: string): void => expect(value).toEqual(expected)
                },
                {
                    name: 'test subject filter',
                    filterBy: ['subjectId=@l1-researcher'],
                    field: 'subjectId',
                    expected: 'l1-researcher',
                    matcher: (value: string, expected: string): void => expect(value).toContain(expected)
                },
                {
                    name: 'test role filter',
                    filterBy: ['roleName=@compute resource'],
                    field: 'roleName',
                    expected: 'Compute resource administrator',
                    matcher: (value: string, expected: string): void => expect(value).toEqual(expected)
                },
                {
                    name: 'test scope filter',
                    filterBy: ['scopeName=@runai'],
                    field: 'scopeName',
                    expected: 'runai',
                    matcher: (value: string, expected: string): void => expect(value).toEqual(expected)
                },
                {
                    name: 'test status filter',
                    filterBy: ['phase!=Creating,phase!=Deleting,phase!=NotReady'],
                    field: 'status.phase',
                    expected: 'Ready',
                    matcher: (value: string, expected: string): void => expect(value).toEqual(expected)
                },
                {
                    name: 'test authorized by filter',
                    filterBy: [`createdBy=@${config.controlPlaneAdminUsername}`],
                    field: 'createdBy',
                    expected: config.controlPlaneAdminUsername,
                    matcher: (value: string, expected: string): void => expect(value).toEqual(expected)
                }
            ];

            for (const testStep of filterTests) {
                await test.step(testStep.name, async () => {
                    const response = await authorizationUtils.getAccessRules({ filterBy: testStep.filterBy });
                    response.accessRules.forEach((accessRule) => {
                        const value = testStep.field.split('.').reduce((o, k) => o[k], accessRule);
                        testStep.matcher(value, testStep.expected);
                    });
                });
            }
        });
    });
});
