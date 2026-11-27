/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    listFlags,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

test.describe(`CLI project tests , tags: ${tags.runOn.serialCI}  ${tags.teams.viper}`, function () {
    test.describe('Submit 2 workloads to 2 different projects - validate default project', function () {
        let defaultProject: Project;
        let specificProject: Project;
        test.beforeAll(function () {
            cliLoginAsSystemAdmin();
        });
        test.beforeEach(async function () {
            defaultProject = await projectResourcesUtils.createProjectAndNamespace('default');
            specificProject = await projectResourcesUtils.createProjectAndNamespace('specific');
            runaiCliV2.projectSet(defaultProject.name);
        });

        test.afterEach(async function () {
            try {
                await operations.deleteProjectAndNamespace(defaultProject.id);
            } catch (error) {
                logger.error(`Failed to delete project ${defaultProject.name}`);
            }
        });
        test('submit job that is not in the default project', async function () {
            const specificProjectJob = objectGenerator.generateShortNameByPrefix('specific-project-job', 5);
            const defaultProjectJob = objectGenerator.generateShortNameByPrefix('default-project-job', 5);
            const flags = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.command]: '-- echo "hello"'
            } as Partial<typeof SubmitFlagsMerged>;

            runaiCliV2.submit(flags, defaultProjectJob, cliV2WorkloadType.workspace);
            flags[SubmitFlagsMerged.project] = specificProject.name;
            runaiCliV2.submit(flags, specificProjectJob, cliV2WorkloadType.workspace);

            await waitUtils.waitForBackendWorkloadPhase(
                defaultProjectJob,
                defaultProject.id,
                Phase.Completed
            );
            await waitUtils.waitForBackendWorkloadPhase(
                specificProjectJob,
                specificProject.id,
                Phase.Completed
            );

            // validate that the job is in the second project list
            const listFlagsParams = {
                [listFlags.project]: specificProject.name
            } as Partial<typeof listFlags>;
            const listSpecificProjectResult = runaiCliV2.list('workspace', listFlagsParams).stdout.toString();
            expect(
                listSpecificProjectResult.includes(specificProjectJob),
                `job ${specificProjectJob} should be in the list`
            ).toBe(true);
            expect(
                listSpecificProjectResult.includes(defaultProject.name),
                `default project ${defaultProject.name} should not be in the list with -p ${specificProject.name}`
            ).toBe(false);

            // validate that the job is not in the default project list
            const listDefaultProjectResult = runaiCliV2.list('workspace').stdout.toString();
            expect(
                listDefaultProjectResult.includes(specificProjectJob),
                `workload ${specificProjectJob} should not be in the list`
            ).toBe(false);
            expect(
                listDefaultProjectResult.includes(specificProject.name),
                `specific project ${specificProject.name} should not be in the list without -p flag`
            ).toBe(false);
            expect(
                listDefaultProjectResult.includes(defaultProject.name),
                `default project ${defaultProject.name} should be in the list without -p flag`
            ).toBe(true);
        });
    });
});
