/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils
} from '../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI List Credential - Negative Tests, tags: ${tags.resource.credential} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        const ERROR_PROJECT_NOT_FOUND = /Failed to list credential assets.*not found|project.*not found/i;
        const ERROR_INVALID_TYPE = /invalid type|invalid credential type/i;
        const ERROR_CONFLICTING_OUTPUT_FORMATS =
            /if any flags in the group.*are set none of the others can be/i;
        const ERROR_INVALID_COLUMN = /invalid column|unknown column/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-cred-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        test('T5750401: List credentials with invalid project name', function () {
            test.step('Attempt to list with non-existent project', function () {
                logger.info('Attempting to list credentials with invalid project name');
                const nonExistentProject = 'non-existent-project-xyz';

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list -p ${nonExistentProject}`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_PROJECT_NOT_FOUND);
                logger.info('Verified: Command failed with appropriate error for invalid project');
            });
        });

        // Negative credential listing coverage remains disabled pending revalidation.
        test.skip('T5750402: List credentials with invalid type', function () {
            test.step('Attempt to list with invalid type', function () {
                logger.info('Attempting to list credentials with invalid type');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list --type invalid-type`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_INVALID_TYPE);
                logger.info('Verified: Command failed with error for invalid credential type');
            });
        });

        test('T5750403: List credentials with conflicting output formats', function () {
            test.step('Attempt to list with multiple output formats', function () {
                logger.info('Attempting to list with conflicting output formats (--json --yaml)');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list --json --yaml`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_CONFLICTING_OUTPUT_FORMATS);
                logger.info('Verified: Command failed with error - conflicting output format flags');
            });
        });

        test('T5750404: List credentials with invalid column name', function () {
            test.step('Attempt to list with invalid column name', function () {
                logger.info('Attempting to list with invalid column name');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} credential list --columns invalidColumnName`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_INVALID_COLUMN);
                logger.info('Verified: Command failed with error for invalid column name');
            });
        });
    });
});
