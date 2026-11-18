/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, TestInfo } from '@playwright/test';
import {
    flowRootState,
    tags,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    computeResourceUtils,
    environmentUtils,
    assetsApiV2Utils,
    objectGenerator,
    waitUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { StatusCodes } from 'http-status-codes';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

let project: Project;
let anyWorkloadSubmissionEnabled: boolean;

test.describe(
    `Workspace V2 API, tags: ${tags.resource.workloads} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.services.workloads, tags.runOn.parallelCI] },
    () => {
        test.beforeAll(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                AnyWorkloadSubmissionType.Workspace
            );
        });

        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });

        async function getWorkloadCreationPayload(workloadName: string, environmentId: string): Promise<any> {
            return {
                meta: {
                    name: workloadName,
                    useGivenNameAsPrefix: true,
                    projectId: project.id.toString(),
                    clusterId: flowRootState.cluster.uuid
                },
                assets: {
                    environment: environmentId,
                    compute: await computeResourceUtils.getId('cpu-only')
                },
                specificRunParams: {
                    autoDeletionTimeAfterCompletionSeconds: 1
                }
            };
        }

        async function isWorkspaceNotFound(workloadId: string): Promise<boolean> {
            try {
                const response = (await workloadsCrudUtils.getWorkspace(workloadId)).response;
                return response.statusCode === StatusCodes.NOT_FOUND;
            } catch (error) {
                if (error.response?.statusCode === StatusCodes.NOT_FOUND) {
                    return true; // Workspace is not found (deleted)
                }
                throw error; // Re-throw other unexpected errors
            }
        }

        test('should delete workspace after workload completion with autoDeletionTimeAfterCompletionSeconds', async ({}, testInfo: TestInfo) => {
            if (anyWorkloadSubmissionEnabled) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'Redirected workspaces are not automatically deleted after completion'
                );
            }

            const workloadName = objectGenerator.generateNameByPrefix('workspace');
            const environment = await environmentUtils.createEnvironmentWithArbitraryFields(
                project.id,
                'sleep 1'
            );

            const result = await assetsApiV2Utils.createWorkspace(
                await getWorkloadCreationPayload(workloadName, environment.meta.id)
            );

            await waitUtils.waitForCondition(
                () => isWorkspaceNotFound(result.body.meta.workloadId),
                `Waiting for workspace ${workloadName} to be deleted`,
                true,
                20
            );

            await environmentUtils.delete(environment.meta.id);
        });
    }
);
