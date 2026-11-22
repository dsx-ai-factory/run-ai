/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';

import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    Utils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { TrainingSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/trainingSpecSpec';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/api';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';

import { kubeCli } from '../../../../../../../modules/cli';

import { enableAnyWorkloadSubmission, RestoreAnyWorkload } from '../anyWorkloadSubmission';

// Regression guard for workload submission builds a cluster storage config by
// fetching the cluster. The old code used the authorized external endpoint gated by
// cluster:read, so researchers without that permission got a 503 wrapping a 403. The fix
// switched to the internal unauthenticated endpoint, so both a researcher and an admin
// can submit. The cluster storage config is only built on the anyworkload path, so the suite
// turns the master switch on first; the test fails on the pre-fix binary and passes on the fixed one.
test.describe('Cluster storage config submission RBAC - API', () => {
    supportedFrom(SupportedVersions.V2_26)(() => {
        let project: Project;
        let researcherAccessRule: AccessRule | undefined;
        let restoreAnyWorkload: RestoreAnyWorkload | undefined;
        const submittedWorkloadIds: string[] = [];

        const trainingSpec: TrainingSpecSpec = {
            image: config.quickStartImage.training,
            command: 'sleep infinity',
            compute: {
                cpuCoreRequest: 0.1,
                cpuMemoryRequest: '100M',
                gpuDevicesRequest: 0
            }
        };

        const adminToken = (): string => flowRootState.users.system_administrator.accessToken;

        const submitters = [
            {
                name: 'regular L1 researcher',
                namePrefix: 'csc-researcher',
                token: (): string => flowRootState.users.l1_researcher.accessToken,
                successMessage:
                    'L1 researcher submission must succeed - pre-fix it returned 503 from the cluster:read endpoint'
            },
            {
                name: 'admin',
                namePrefix: 'csc-admin',
                token: adminToken,
                successMessage: 'admin submission must succeed'
            }
        ];

        test.beforeAll(async () => {
            restoreAnyWorkload = await enableAnyWorkloadSubmission();

            project = await projectResourcesUtils.createProjectAndNamespace('csc-rbac');

            const researcher = flowRootState.users.l1_researcher;
            researcherAccessRule = await authorizationUtils.createAccessRuleForUser(
                adminToken(),
                researcher.email,
                researcher.role.id,
                SubjectType.User,
                project.id.toString(),
                ScopeType.Project
            );
            logger.info(`granted L1 researcher access to project ${project.name} (${project.id})`);
        });

        test.afterAll(async () => {
            workloadsCrudUtils.accessToken = adminToken();
            for (const workloadId of submittedWorkloadIds) {
                try {
                    await workloadsCrudUtils.deleteTraining(workloadId);
                } catch (error) {
                    logger.error(`failed to delete training ${workloadId}: ${error}`);
                }
            }
            if (researcherAccessRule) {
                try {
                    await authorizationUtils.deleteAccessRule(researcherAccessRule, adminToken());
                } catch (error) {
                    logger.error(`failed to delete researcher access rule: ${error}`);
                }
            }
            if (project) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (error) {
                    logger.error(`failed to delete project ${project?.id}: ${error}`);
                }
            }
            if (restoreAnyWorkload) {
                await restoreAnyWorkload();
            }
        });

        for (const submitter of submitters) {
            test(`${submitter.name} user can submit a training workload`, async () => {
                const workloadName = objectGenerator.generateShortNameByPrefix(submitter.namePrefix);

                const result = await workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    trainingSpec,
                    submitter.token()
                );

                expect(result.response.statusCode, submitter.successMessage).toBe(StatusCodes.ACCEPTED);
                expect(result.body.name).toBe(workloadName);
                expect(result.body.projectId).toBe(project.id.toString());
                expect(result.body.workloadId).toBeTruthy();
                submittedWorkloadIds.push(result.body.workloadId);

                await test.step(`${submitter.name} workload reaches Running`, async () => {
                    // The anyworkload submission id is not resolvable via the trainings read model,
                    // so wait on the pod instead of polling the workload API by id.
                    await kubeCli.waitForJobPodsRunning(
                        Utils.namespaceFromProjectName(project.name),
                        workloadName
                    );
                });
            });
        }
    });
});
