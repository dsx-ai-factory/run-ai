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
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Datavolume - Negative Tests, tags: ${tags.resource.datavolume} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        const ERROR_PROJECT_NOT_FOUND = /Failed to list data volume assets\. project name: .*, not found/;
        const ERROR_CONFLICTING_OUTPUT_FORMATS =
            /if any flags in the group \[json yaml table\] are set none of the others can be/;
        const ERROR_INVALID_COLUMN = /invalid column name: /;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-dv-neg-list', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.debug(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.debug(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.debug(`Deleted project: ${projectName}`);
                }
            });
        });

        test('T5750414: List data volumes with invalid project name', function () {
            test.step('Attempt to list with non-existent project', function () {
                logger.debug('Attempting to list data volumes with invalid project name');
                const nonExistentProject = 'non-existent-project-xyz';

                expect(() => {
                    runaiCliV2.list('datavolume', { [listFlags.project]: nonExistentProject } as any, false);
                }).toThrow(ERROR_PROJECT_NOT_FOUND);
                logger.debug('Verified: Command failed with appropriate error for invalid project');
            });
        });

        test('T5750415: List data volumes with conflicting output formats', function () {
            test.step('Attempt to list with multiple output formats', function () {
                logger.debug('Attempting to list with conflicting output formats (--json --yaml)');

                expect(() => {
                    runaiCliV2.list(
                        'datavolume',
                        { [listFlags.json]: '', [listFlags.yaml]: '' } as any,
                        false
                    );
                }).toThrow(ERROR_CONFLICTING_OUTPUT_FORMATS);
                logger.debug('Verified: Command failed with error - conflicting output format flags');
            });
        });

        test('T5750416: List data volumes with invalid column name', function () {
            test.step('Attempt to list with invalid column name', function () {
                logger.debug('Attempting to list with invalid column name');

                expect(() => {
                    runaiCliV2.list('datavolume', { [listFlags.columns]: 'invalidColumnName' } as any, false);
                }).toThrow(ERROR_INVALID_COLUMN);
                logger.debug('Verified: Command failed with error for invalid column name');
            });
        });
    });
});
