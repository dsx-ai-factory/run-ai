/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { tags, objectGenerator, projectResourcesUtils, orgUnitUtils, logger } from '../../../../../../../../modules/helpers';
import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { applicationSubmitFlags } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const CHART_URL = 'https://charts.bitnami.com/bitnami/nginx-18.1.11.tgz';

const ERROR_PROJECT_NOT_FOUND = /failed to create application\. project name: .*, not found/;
const ERROR_CREDENTIAL_NOT_FOUND = /credential.*not found/i;

test.describe(`CLI V2 AI Application Submit Negative Tests, ${tags.teams.viper},`, () => {
    let project: Project;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-sub-neg');
    });

    test.afterAll(async () => {
        try {
            await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
        } catch (error) {
            logger.error('Project cleanup failed:', error);
        }
    });

    test('submit with an unknown project fails with a not-found error', () => {
        const appName = objectGenerator.generateNameByPrefix('app-submit-badproj');

        expect(() => {
            runaiCliV2.applicationSubmit(appName, {
                [applicationSubmitFlags.project]: 'non-existent-project-xyz',
                [applicationSubmitFlags.chartUrl]: CHART_URL
            });
        }).toThrow(ERROR_PROJECT_NOT_FOUND);
    });

    test('submit with a non-existent credential fails with a not-found error', () => {
        const appName = objectGenerator.generateNameByPrefix('app-submit-badcred');

        expect(() => {
            runaiCliV2.applicationSubmit(appName, {
                [applicationSubmitFlags.project]: project.name,
                [applicationSubmitFlags.chartUrl]: CHART_URL,
                [applicationSubmitFlags.credential]: 'non-existent-credential-xyz'
            });
        }).toThrow(ERROR_CREDENTIAL_NOT_FOUND);
    });
});
