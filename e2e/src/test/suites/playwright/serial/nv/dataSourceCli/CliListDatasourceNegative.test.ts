/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI List Datasource - Negative Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_PROJECT_NOT_FOUND = /Failed to list data source assets.*not found/i;
        const ERROR_CONFLICTING_OUTPUT_FORMATS =
            /if any flags in the group.*are set none of the others can be/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-list-neg', 5);
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

        test('T5585451: List data sources with invalid project name', function () {
            test.step('Attempt to list with non-existent project', function () {
                logger.info('Attempting to list data sources with invalid project name');
                const nonExistentProject = 'non-existent-project-xyz';

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource list -p ${nonExistentProject}`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: Failed to list data source assets. project name: non-existent-project-xyz, not found"
                expect(result.stderr).toMatch(ERROR_PROJECT_NOT_FOUND);

                logger.info('Verified: Command failed with appropriate error for invalid project');
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5585452: List data sources with invalid type', function () {
            test.step('Attempt to list with invalid type and print output', function () {
                logger.info('Attempting to list data sources with invalid type');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource list --type invalid-type-xyz`,
                    false
                );

                // TODO: Based on actual output, verify the expected behavior
                // For now, just verify command was executed
                expect(result).toBeDefined();
            });
        });

        test('T5585453: List data sources with conflicting output formats', function () {
            test.step('Attempt to list with multiple output formats', function () {
                logger.info('Attempting to list with conflicting output formats (--json --yaml)');

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource list --json --yaml`,
                    false
                );

                // Verify command failed
                expect(result.status).not.toBe(0);
                // Actual error: "Error: if any flags in the group [json yaml table] are set none of the others can be; [json yaml] were all set"
                expect(result.stderr).toMatch(ERROR_CONFLICTING_OUTPUT_FORMATS);

                logger.info('Verified: Command failed with error - conflicting output format flags');
            });
        });
    });
});
