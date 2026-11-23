/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
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
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`CLI Submit with Datavolume - Negative Tests, tags: ${tags.resource.datavolume} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        const ERROR_DV_NOT_FOUND = /datavolume '.*' not found/;
        const ERROR_INVALID_DV_FORMAT =
            /invalid datavolume format '.*': required format is name=NAME,mountpath=PATH/;

        let projectName: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-dv-neg', 5);
            const namespace = Utils.namespaceFromProjectName(projectName);

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
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) await operations.deleteProjectAndNamespace(project.id);
            });
        });

        test('T5750427: Submit with non-existent datavolume name', function () {
            test.step('Attempt submit with non-existent datavolume', function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('ws-bad-dv', 5);

                expect(() => {
                    runaiCliV2.submit(
                        {
                            [SubmitFlagsMerged.datavolume]: [
                                'name=non-existent-datavolume,mountpath=/mnt/data'
                            ],
                            [SubmitFlagsMerged.project]: projectName,
                            [SubmitFlagsMerged.image]: 'ubuntu:22.04',
                            [SubmitFlagsMerged.command]: '-- sleep infinity'
                        } as Partial<typeof SubmitFlagsMerged>,
                        workloadName,
                        cliV2WorkloadType.workspace
                    );
                }).toThrow(ERROR_DV_NOT_FOUND);
                logger.debug('Verified: Command failed with datavolume not found error');
            });
        });

        test('T5750428: Submit with invalid datavolume format (missing mountpath)', function () {
            test.step('Attempt submit with only name (no mountpath)', function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('ws-bad-fmt', 5);

                expect(() => {
                    runaiCliV2.submit(
                        {
                            [SubmitFlagsMerged.datavolume]: ['name=some-dv-name'],
                            [SubmitFlagsMerged.project]: projectName,
                            [SubmitFlagsMerged.image]: 'ubuntu:22.04',
                            [SubmitFlagsMerged.command]: '-- sleep infinity'
                        } as Partial<typeof SubmitFlagsMerged>,
                        workloadName,
                        cliV2WorkloadType.workspace
                    );
                }).toThrow(ERROR_INVALID_DV_FORMAT);
                logger.debug('Verified: Command failed with invalid format error (missing mountpath)');
            });
        });

        test('T5750429: Submit with invalid datavolume format (missing name)', function () {
            test.step('Attempt submit with only mountpath (no name)', function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('ws-bad-fmt2', 5);

                expect(() => {
                    runaiCliV2.submit(
                        {
                            [SubmitFlagsMerged.datavolume]: ['mountpath=/mnt/data'],
                            [SubmitFlagsMerged.project]: projectName,
                            [SubmitFlagsMerged.image]: 'ubuntu:22.04',
                            [SubmitFlagsMerged.command]: '-- sleep infinity'
                        } as Partial<typeof SubmitFlagsMerged>,
                        workloadName,
                        cliV2WorkloadType.workspace
                    );
                }).toThrow(ERROR_INVALID_DV_FORMAT);
                logger.debug('Verified: Command failed with invalid format error (missing name)');
            });
        });
    });
});
