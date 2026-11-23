/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    assetsApiV2Utils,
    computeResourceUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { expect } from 'chai';
import { WorkspaceCreationRequestV2 } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(`Create a workload with terminationGracePeriodSeconds ${tags.teams.viper}`, function () {
    let project: Project;
    let projectName: string;
    let namespace: string;
    let workspaceName: string;
    let anyWorkloadSubmissionEnabled: boolean;

    test.beforeAll('create project and namespace', async function () {
        projectName = objectGenerator.generateNameByPrefix('p-workspace');
        workspaceName = objectGenerator.generateNameByPrefix('workspace');
        namespace = Utils.namespaceFromProjectName(projectName);

        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 1,
            maxAllowedGpus: 1
        });

        assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
        anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
            AnyWorkloadSubmissionType.Workspace
        );
    });

    test.afterAll('delete project and namespace', async function () {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });

    test('check terminationGracePeriodSeconds propagates to the workload', async function () {
        const environmentId = await environmentUtils.getId('jupyter-lab');
        const computeId = await computeResourceUtils.getId('cpu-only');

        const terminationGracePeriodSeconds = 60;

        const workspaceCreationRequest: WorkspaceCreationRequestV2 = {
            meta: {
                name: workspaceName,
                projectId: project.id.toString(),
                clusterId: flowRootState.cluster.uuid
            },
            assets: {
                environment: environmentId,
                compute: computeId
            },
            specificRunParams: {
                terminationGracePeriodSeconds
            }
        };

        await assetsApiV2Utils.createWorkspace(workspaceCreationRequest);

        if (anyWorkloadSubmissionEnabled) {
            await waitUtils.waitForBatchJobToExist(workspaceName, namespace);
            const job = await kubeCli.getBatchJob(workspaceName, namespace);
            expect(job.spec?.template.spec?.terminationGracePeriodSeconds).to.equal(
                terminationGracePeriodSeconds
            );
        } else {
            await waitUtils.waitForInteractiveWorkloadToExist(workspaceName, namespace);
            const interactiveWorkload: any = await kubeCli.getInteractiveWorkload(workspaceName, namespace);
            expect(interactiveWorkload.spec.terminationGracePeriodSeconds.value).to.equal(
                terminationGracePeriodSeconds
            );
        }
    });
});
