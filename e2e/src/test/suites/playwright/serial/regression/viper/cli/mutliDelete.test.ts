/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import {
    cliV2workloadNoun,
    cliV2WorkloadType,
    workloadTypesNoInference,
    listFlags,
    SubmitFlagsMerged,
    DeleteFlagsMerged,
    cliLoginAsSystemAdmin
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { test } from '@playwright/test';

test.describe(`Validate workload multi delete   ${tags.teams.viper}`, function () {
    let jobNameSuffix: string;
    let project: Project;

    async function deleteProject(projectObj: Project): Promise<void> {
        try {
            await operations.deleteProjectAndNamespace(projectObj.id);
        } catch (error) {
            logger.error(`Failed to delete project ${projectObj.name}`);
        }
    }

    function getJobName(type: cliV2WorkloadType, num: number): string {
        return `${type}-${jobNameSuffix}${num}`;
    }

    function isWorkloadDeleted(type: cliV2WorkloadType): boolean {
        const listFlag = {
            [listFlags.json]: ''
        } as Partial<typeof listFlags>;
        const listResult = runaiCliV2.list(type, listFlag).stdout;
        return !listResult.includes(getJobName(type, 1)) && !listResult.includes(getJobName(type, 2));
    }

    async function generateWorkloads(): Promise<void> {
        const submitFlag = {
            [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
        } as Partial<typeof SubmitFlagsMerged>;
        Object.values(workloadTypesNoInference).forEach((type) => {
            runaiCliV2.submit(submitFlag, getJobName(type, 1), type as cliV2WorkloadType);
            runaiCliV2.submit(submitFlag, getJobName(type, 2), type as cliV2WorkloadType);
        });
        for (const type of Object.values(workloadTypesNoInference)) {
            await waitUtils.waitForBackendWorkloadPhase(getJobName(type, 1), project.id, Phase.Completed);
            await waitUtils.waitForBackendWorkloadPhase(getJobName(type, 2), project.id, Phase.Completed);
        }
    }

    test.beforeAll(async function () {
        cliLoginAsSystemAdmin();
        project = await projectResourcesUtils.createProjectAndNamespace('muti-delete');
        runaiCliV2.projectSet(project.name);
        jobNameSuffix = objectGenerator.generateShortNameByPrefix('delete', 5);
        await generateWorkloads();
    });

    test.afterAll(async function () {
        await deleteProject(project);
    });

    test.describe('validate multi workload generate', function () {
        const deleteFlag = {
            [DeleteFlagsMerged.byPassDialog]: ''
        } as Partial<typeof DeleteFlagsMerged>;
        Object.values(workloadTypesNoInference).forEach((type) => {
            test(`validate ${type} multi workload delete`, async function () {
                runaiCliV2.delete(
                    deleteFlag,
                    `${getJobName(type, 1)} ${getJobName(type, 2)}`,
                    cliV2workloadNoun
                );
                await waitUtils.waitForCondition(
                    () => {
                        return isWorkloadDeleted(type);
                    },
                    'wait for workload to be deleted',
                    true,
                    30
                );
            });
        });
    });
});
