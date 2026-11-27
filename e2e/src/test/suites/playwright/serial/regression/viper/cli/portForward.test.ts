/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    portForwardFlags,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { bashCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { assertExpectedValue } from '../../../../../../../conditionPolling/assertUsingPolling';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    `CLI Port Forward test, tags:   ${tags.teams.viper}`,
    { tag: [tags.services.clusterApi] },
    function () {
        let project: Project;
        let trainingName: string;
        test.beforeAll(async () => {
            cliLoginAsSystemAdmin();
            project = await projectResourcesUtils.createProjectAndNamespace('wl-issues');
            runaiCliV2.projectSet(project.name);
        });
        test.afterAll(async () => {
            if (trainingName) {
                await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                    trainingName,
                    project.id,
                    WorkloadType.TRAINING
                );
            }
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`Failed to delete project ${project.name}`);
            }
        });

        test('port forward for 30 secs', async ({}, testInfo) => {
            if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training)) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'CLI port-forward cannot locate WLR-backed training workloads'
                );
            }
            const submitFlag = {
                [SubmitFlagsMerged.image]: settings.isArm
                    ? 'nginx'
                    : config.quickStartImage.inferenceHelloWorld
            } as Partial<typeof SubmitFlagsMerged>;
            trainingName = objectGenerator.generateNameByPrefix('test-port-forward');
            const localPort = '30003';
            const podListenPort = '8080';
            runaiCliV2.submit(submitFlag, trainingName, cliV2WorkloadType.training);
            await waitUtils.waitForBackendWorkloadPhase(trainingName, project.id, Phase.Running);
            const portforwardFlag = {
                [portForwardFlags.port]: `${localPort}:${podListenPort}`
            } as Partial<typeof portForwardFlags>;
            const process = (
                await runaiCliV2.portForward(portforwardFlag, trainingName, cliV2WorkloadType.training)
            )[1];
            for (let i = 0; i < 10; i++) {
                assertExpectedValue(bashCli, bashCli.exec, 'Hello World!\n', [`curl 127.0.0.1:${localPort}`]);
                await waitUtils.sleep(3000);
            }
            process.kill();
        });
    }
);
