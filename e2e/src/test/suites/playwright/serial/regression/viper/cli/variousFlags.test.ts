/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, objectGenerator, projectResourcesUtils, tags } from '../../../../../../../modules/helpers';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    listFlags,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test, expect } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`CLI flags test  ${tags.teams.viper}`, { tag: [tags.services.runaiCli] }, function () {
    let project: Project;
    let jobName: string;
    const expectedString = 'These shell commands are defined internally';
    test.beforeAll(async function () {
        cliLoginAsSystemAdmin();
        project = await projectResourcesUtils.createProjectAndNamespace('attach');
        runaiCliV2.projectSet(project.name);
        jobName = objectGenerator.generateNameByPrefix('attach');
    });
    test.describe('submit flags', function () {
        test('validate submit with attach flag', async ({}, testInfo) => {
            if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training)) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'CLI attach cannot locate WLR-backed training workloads'
                );
            }
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.attach]: ''
            } as Partial<typeof SubmitFlagsMerged>;
            const output = await runaiCliV2.submitAttach(
                submitFlag,
                jobName,
                cliV2WorkloadType.training,
                'help'
            );
            expect(output).toContain(expectedString);
        });
    });
    test.describe('list flags', function () {
        let secondProject: Project;
        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            secondProject = await projectResourcesUtils.createProjectAndNamespace('attach');
        });
        test('validate project list limit', () => {
            const listFlagsParams = {
                [listFlags.json]: '',
                [listFlags.limit]: 10
            } as Partial<typeof listFlags>;
            let listResult = JSON.parse(runaiCliV2.list('project', listFlagsParams).stdout);
            expect(listResult.projects.length).toBeGreaterThan(2);
            listFlagsParams[listFlags.limit] = 1;
            listResult = JSON.parse(runaiCliV2.list('project', listFlagsParams).stdout);
            expect(listResult.projects.length).toBe(1);
        });

        test('validate project list offset', () => {
            const listFlagsParams = {
                [listFlags.json]: '',
                [listFlags.limit]: 1
            } as Partial<typeof listFlags>;
            const firstProject = JSON.parse(runaiCliV2.list('project', listFlagsParams).stdout);
            listFlagsParams[listFlags.offset] = 1;
            const secondProject = JSON.parse(runaiCliV2.list('project', listFlagsParams).stdout);
            expect(firstProject.projects[0].id, 'Project ids should be different').not.toBe(
                secondProject.projects[0].id
            );
        });
    });
});
