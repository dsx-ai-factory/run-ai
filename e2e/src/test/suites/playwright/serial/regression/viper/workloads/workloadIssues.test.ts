/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    describeFlags,
    workloadTypesNoInference,
    SubmitFlagsMerged,
    cliLoginAsSystemAdmin
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    versionUtils,
    waitUtils
} from '../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { test, expect, TestInfo } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Validate workload status   ${tags.teams.viper}`, { tag: [tags.resource.trainingWorkload] }, function () {
    const jobNameSuffix = objectGenerator.generateShortNameByPrefix('issues', 5);
    let project: Project;
    const failCommand = '-- cat nonExistCommand';

    function getSubmissionType(workloadType: cliV2WorkloadType): AnyWorkloadSubmissionType {
        if (workloadType === cliV2WorkloadType.workspace) {
            return AnyWorkloadSubmissionType.Workspace;
        }
        if (workloadType === cliV2WorkloadType.training) {
            return AnyWorkloadSubmissionType.Training;
        }
        return AnyWorkloadSubmissionType.Distributed;
    }

    function isClusterVersion3OrNewer(): boolean {
        return versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V3_0);
    }

    async function skipRedirectedIssueGap(
        type: cliV2WorkloadType,
        testInfo: TestInfo
    ): Promise<void> {
        if (
            !isClusterVersion3OrNewer() &&
            [cliV2WorkloadType.workspace, cliV2WorkloadType.training].includes(type) &&
            (await isAnyWorkloadSubmissionEnabled(getSubmissionType(type)))
        ) {
            skipPlaywrightTest(
                test,
                testInfo,
                'Failed redirected workloads do not report runtime issues'
            );
        }
    }

    async function deleteProject(projectObj: Project): Promise<void> {
        try {
            await operations.deleteProjectAndNamespace(projectObj.id);
        } catch (error) {
            logger.error(`Failed to delete project ${projectObj.name}`);
        }
    }

    async function generateWorkloads(): Promise<void> {
        const submitFlag = {
            [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
            [SubmitFlagsMerged.backoffLimit]: 1,
            [SubmitFlagsMerged.command]: failCommand
        } as Partial<typeof SubmitFlagsMerged>;

        Object.values(workloadTypesNoInference).forEach((type) => {
            const jobName = `${type}-${jobNameSuffix}`;
            runaiCliV2.submit(submitFlag, jobName, type as cliV2WorkloadType);
        });
        for (const type of Object.values(workloadTypesNoInference)) {
            const jobName = `${type}-${jobNameSuffix}`;
            await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Failed);
        }
    }

    test.beforeAll(async function () {
        cliLoginAsSystemAdmin();
        project = await projectResourcesUtils.createProjectAndNamespace('wl-issues');
        runaiCliV2.projectSet(project.name);
        await generateWorkloads();
    });

    test.afterAll(async function () {
        await deleteProject(project);
    });

    test.describe('validate failed workloads issues', function () {
        const describeFlag = {
            [describeFlags.output]: 'json'
        } as Partial<typeof describeFlags>;
        Object.values(workloadTypesNoInference).forEach((type) => {
            test(`validate ${type} issues using cli describe command`, async function ({}, testInfo: TestInfo) {
                await skipRedirectedIssueGap(type, testInfo);
                const jobName = `${type}-${jobNameSuffix}`;
                const workloadDescribe = JSON.parse(runaiCliV2.describe(describeFlag, jobName, type));
                expect(workloadDescribe['issues'][0]['type']).toBe('Compute');
            });
        });
    });
});
