/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { tags } from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { listFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const ERROR_PROJECT_NOT_FOUND = /applications list failed\. project name: .*, not found/;

test.describe(`CLI V2 AI Application List Negative Tests, ${tags.teams.viper},`, () => {
    test('application list with an unknown project fails', () => {
        expect(() => {
            runaiCliV2.list(
                'application',
                { [listFlags.project]: 'non-existent-project-xyz' } as Partial<typeof listFlags>,
                false
            );
        }).toThrow(ERROR_PROJECT_NOT_FOUND);
    });
});
