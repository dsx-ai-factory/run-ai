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
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    listFlags,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';

test.describe(
    `CLI basic data leak tests , tags: ${tags.teams.viper}`,
    { tag: [tags.runOn.basicCI, tags.runOn.parallelCI] },
    function () {
        let firstProject: Project;
        let secondProject: Project;
        let trainingName: string;
        let workspaceName: string;

        async function generateWorkloads(): Promise<void> {
            trainingName = objectGenerator.generateNameByPrefix('dataleak-training');
            workspaceName = objectGenerator.generateNameByPrefix('dataleak-workspace');

            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.project]: firstProject.name
            } as Partial<typeof SubmitFlagsMerged>;
            runaiCliV2.submit(submitFlag, trainingName, cliV2WorkloadType.training);
            runaiCliV2.submit(submitFlag, workspaceName, cliV2WorkloadType.workspace);
            await waitUtils.waitForBackendWorkloadPhase(trainingName, firstProject.id, Phase.Completed);
            await waitUtils.waitForBackendWorkloadPhase(workspaceName, firstProject.id, Phase.Completed);
        }

        test.beforeAll(async () => {
            cliLoginAsSystemAdmin();
            firstProject = await projectResourcesUtils.createProjectAndNamespace('dataleak');
            secondProject = await projectResourcesUtils.createProjectAndNamespace('dataleak2');
            await generateWorkloads();
        });
        test.afterAll(async () => {
            await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                trainingName,
                firstProject.id,
                WorkloadType.TRAINING
            );
            await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                workspaceName,
                firstProject.id,
                WorkloadType.WORKSPACE
            );
            try {
                await operations.deleteProjectAndNamespace(firstProject.id);
                await operations.deleteProjectAndNamespace(secondProject.id);
            } catch (error) {
                logger.error(`Failed to delete projects ${firstProject.name} ${secondProject.name}`);
            }
        });

        function validateListValue(shouldBeZero = true, projectName: string): void {
            const listFlagsParams = {
                [listFlags.json]: '',
                [listFlags.project]: projectName
            } as Partial<typeof listFlags>;
            const listResult = runaiCliV2.list('workload', listFlagsParams).stdout;
            const expectedValue = shouldBeZero ? 0 : 2;
            const listResultLength = JSON.parse(listResult)['workloads'].length;
            expect(
                expectedValue === listResultLength,
                `expected to get ${expectedValue} in list workloads, got ${listResultLength}`
            ).toBe(true);
        }

        test('cli workload list leak tests', () => {
            validateListValue(false, firstProject.name);
            validateListValue(true, secondProject.name);
        });
        test('api workload list leak tests', async () => {
            const secondProjectWorkloads = await workloadsUtilsV2.getWorkloadByNameContainAndProjectNameEq(
                'aa||dataleak',
                secondProject.name
            );
            expect(
                secondProjectWorkloads,
                `get workloads with project ${secondProject.name} shouldnt return any workload`
            ).toBeUndefined();
        });
    }
);
