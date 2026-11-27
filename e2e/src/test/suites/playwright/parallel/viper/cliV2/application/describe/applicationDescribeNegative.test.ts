/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    logger,
    orgUnitUtils,
    projectResourcesUtils,
    tags
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { describeFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const ERROR_APP_NOT_FOUND = /application describe failed\. application ".*" not found in project ".*"/;

test.describe(`CLI V2 AI Application Describe Negative Tests, ${tags.teams.viper},`, () => {
    let project: Project;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-desc-neg');
    });

    test.afterAll(async () => {
        try {
            await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
        } catch (error) {
            logger.error('Project cleanup failed:', error);
        }
    });

    test('describe of a missing application fails with a not-found error', () => {
        expect(() => {
            runaiCliV2.applicationDescribe('no-such-application', {
                [describeFlags.project]: project.name
            });
        }).toThrow(ERROR_APP_NOT_FOUND);
    });
});
