/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    projectResourcesUtils,
    logger,
    flowRootState,
    tags,
    objectGenerator,
    operations,
    workloadsCrudUtils,
    waitUtils,
    policyUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { TrainingSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/trainingSpecSpec';
import { StatusCodes } from 'http-status-codes';
import { workloadsCrudService, policyApi as policyApiHandler } from '../../../../../../modules/api';
import { kubeCli } from '../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType } from '../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';

test.describe(`Privileged container configuration (API) ${tags.resource.workloads} ${tags.operations.submit} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let policyName: string;

        test.beforeEach(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('p');
            workloadsCrudService.trainingsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            policyName = `project-${project.id}`;
        });

        test.afterEach(async () => {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test('should fail to submit privileged workload without policy', async ({}, testInfo: TestInfo) => {
            const workloadName = objectGenerator.generateNameByPrefix('w');
            const workloadSpec: TrainingSpecSpec = {
                image: 'ubuntu',
                command: 'sleep infinity',
                security: { privileged: true }
            };

            await expect(
                workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    workloadSpec
                )
            ).rejects.toMatchObject({
                body: {
                    complianceIssues: [
                        {
                            field: 'security.privileged',
                            details: 'the administrator prohibited changing the value of this field'
                        }
                    ]
                }
            });
        });

        test('should succeed to submit privileged workload when policy allows it', async ({}, testInfo: TestInfo) => {
            const workloadName = objectGenerator.generateNameByPrefix('w');
            const workloadSpec: TrainingSpecSpec = {
                image: 'ubuntu',
                command: 'sleep infinity',
                security: { privileged: true }
            };

            await test.step('Create workload policy with privileged.canEdit=true', async () => {
                await policyUtils.createPolicy(
                    false,
                    policyName,
                    ScopeType.Project.toString(),
                    project.id,
                    WorkloadType.Training,
                    undefined,
                    {
                        security: {
                            privileged: { canEdit: true }
                        }
                    }
                );
            });

            await test.step('Submit privileged training workload', async () => {
                const result = await workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    workloadSpec
                );
                expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);

                const pods = await waitUtils.waitForPods(result.body.name, project.namespace, 1, 60, [
                    'Running'
                ]);
                expect(pods.length).toBe(1);
                expect(pods[0].spec.containers[0].securityContext?.privileged).toBeTruthy();

                logger.info('Verifying privileged behavior by running "mount" inside the container');
                const mountCheck = await kubeCli.execInPod(pods[0].metadata.name, project.namespace, [
                    'sh',
                    '-c',
                    'mkdir -p /mnt/test && mount -t tmpfs tmpfs /mnt/test'
                ]);
                expect(mountCheck.stderr).toBe('');
            });

            await test.step('Cleanup workload policy', async () => {
                await policyUtils.deleteWorkloadPolicyV2(WorkloadType.Training, {
                    scope: ScopeType.Project.toString(),
                    scopeId: project.id.toString()
                });
            });
        });
    });
});
