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
import {
    cliV2workloadNoun,
    DeleteFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const applicationNoun = 'application' as unknown as typeof cliV2workloadNoun;

const ERROR_APP_NOT_FOUND = /application delete failed\. application ".*" not found in project ".*"/;
const ERROR_PROJECT_NOT_FOUND = /application delete failed\. project name: .*, not found/;

test.describe(`CLI V2 AI Application Delete Negative Tests, ${tags.teams.viper},`, () => {
    let project: Project;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-app-del-neg');
    });

    test.afterAll(async () => {
        try {
            await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
        } catch (error) {
            logger.error('Project cleanup failed:', error);
        }
    });

    test('delete of a missing application fails with a not-found error', () => {
        expect(() => {
            runaiCliV2.delete(
                {
                    [DeleteFlagsMerged.project]: project.name,
                    [DeleteFlagsMerged.byPassDialog]: ''
                } as Partial<typeof DeleteFlagsMerged>,
                'no-such-application',
                applicationNoun
            );
        }).toThrow(ERROR_APP_NOT_FOUND);
    });

    test('delete with an unknown project fails', () => {
        expect(() => {
            runaiCliV2.delete(
                {
                    [DeleteFlagsMerged.project]: 'non-existent-project-xyz',
                    [DeleteFlagsMerged.byPassDialog]: ''
                } as Partial<typeof DeleteFlagsMerged>,
                'no-such-application',
                applicationNoun
            );
        }).toThrow(ERROR_PROJECT_NOT_FOUND);
    });
});
