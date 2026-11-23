/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { policyApi as policyApiHandler } from '../../../../../../modules/api';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    config,
    crds,
    departmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    policyUtils,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';
import { ScopeType } from '../../../../../../modules/api/apis/policyApi';
import { test, TestInfo } from '@playwright/test';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const WorkloadTypeParams = [
    {
        policyType: 'training',
        image: config.quickStartImage.training,
        jobKind: crds.runaiTrainingWorkloadCrd.kind,
        crd: crds.runaiTrainingWorkloadCrd,
        kind: WorkloadType.Training,
        isInteractive: false
    },
    {
        policyType: 'interactive',
        image: config.quickStartImage.interactive,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind,
        crd: crds.runaiInteractiveWorkloadCrd,
        kind: WorkloadType.Interactive,
        isInteractive: true
    }
];

WorkloadTypeParams.forEach(function (params) {
    test.describe(`${params.policyType} policy scope - negative checks: ${tags.resource.policy} ${tags.tool.api} ${tags.teams.viper}`, function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let jobName: string;
        let policyName: string;
        let scope: 'project' | 'department';
        let departmentIdForPolicy: string;
        let projectIdForPolicy: number;
        let anyWorkloadSubmissionEnabled: boolean;

        async function _createPolicy(
            defaults: object,
            rules?: object,
            scope = ScopeType.Project,
            id = project.id
        ): Promise<void> {
            await policyUtils.createPolicy(
                params.isInteractive,
                policyName,
                scope === ScopeType.Project ? project.namespace : 'runai',
                id,
                params.kind,
                defaults,
                rules,
                scope
            );
        }

        async function _deletePolicy(scope: 'project' | 'department'): Promise<void> {
            if (params.isInteractive) {
                await policyApiHandler.policyApi.deleteWorkspacePolicy(
                    scope,
                    departmentIdForPolicy,
                    projectIdForPolicy?.toString()
                );
            } else {
                await policyApiHandler.policyApi.deleteTrainingPolicy(
                    scope,
                    departmentIdForPolicy,
                    projectIdForPolicy?.toString()
                );
            }
        }

        async function _waitForWorkloadRunning(
            workloadName: string,
            workloadNamespace: string
        ): Promise<void> {
            if (anyWorkloadSubmissionEnabled) {
                await waitUtils.waitForBatchJobToExist(workloadName, workloadNamespace);
                await waitUtils.waitForPods(workloadName, workloadNamespace, 1, 120, ['Running']);
            } else {
                await waitUtils.waitForJobToBeRunning(workloadName, workloadNamespace, params.jobKind);
            }
        }

        async function _getCpuRequest(workloadNamespace: string): Promise<number | string | undefined> {
            if (anyWorkloadSubmissionEnabled) {
                const job = await kubeCli.getBatchJob(jobName, workloadNamespace);
                return job.spec?.template.spec?.containers[0].resources?.requests?.cpu;
            }
            return workloadUtils.getCpusCount(jobName, workloadNamespace, params.crd);
        }

        test.beforeAll(async function () {
            policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                params.isInteractive
                    ? AnyWorkloadSubmissionType.Workspace
                    : AnyWorkloadSubmissionType.Training
            );

            // Enable control plane policies so they are applied to workloads
            await policyUtils.enableClusterPolicySync();
        });

        test.beforeEach(async function ({}, testInfo: TestInfo) {
            jobName = objectGenerator.generateShortNameByPrefix('job-for-policy');
            project = await projectResourcesUtils.createProjectAndNamespace('project-for-policy');
            projectName = project.name;
            namespace = project.namespace;
            departmentIdForPolicy = undefined;
            projectIdForPolicy = undefined;
        });

        test.afterEach(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });
        test.describe('project scoped policy', function () {
            test.afterEach(async () => {
                await _deletePolicy(scope);
            });
            test.beforeEach(function () {
                scope = 'project';
                projectIdForPolicy = project.id;
                policyName = `${scope}-${projectIdForPolicy}`;
            });
            test('cross-project policy isolation', async function () {
                await _createPolicy({ compute: { cpuCoreRequest: 1 } });

                const secondProject = await projectResourcesUtils.createProjectAndNamespace('second-project');
                // submit a job  from a different project
                const wlType = params.isInteractive
                    ? cliV2WorkloadType.workspace
                    : cliV2WorkloadType.training;
                const submitFlags = {
                    [SubmitFlagsMerged.image]: params.image,
                    [SubmitFlagsMerged.project]: secondProject.name
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(submitFlags, jobName, wlType);
                await _waitForWorkloadRunning(jobName, secondProject.namespace);

                // ensure that the policy is not applied
                const cpusCount = await _getCpuRequest(secondProject.namespace);
                expect(cpusCount).to.undefined;

                await operations.deleteProjectAndNamespace(secondProject.id); //cleanup
            });
        });

        test.describe('department scoped policy', function () {
            let projectInDifferentDepartment: Project;
            let departmentId: number;
            test.beforeEach(async () => {
                const secondDepartmentName = objectGenerator.generateShortNameByPrefix('department-for-e2e');
                departmentId = Number(
                    (await departmentUtils.createBackendDepartmentWithName(secondDepartmentName)).body.id
                );
                await departmentUtils.waitForDepartmentCrd(secondDepartmentName);
                projectInDifferentDepartment = await projectResourcesUtils.createProjectAndNamespace(
                    'p-b',
                    departmentId
                );
                scope = 'department';
                policyName = `${scope}-${departmentId}`;
                departmentIdForPolicy = departmentId.toString();
            });
            test.afterEach(async () => {
                await _deletePolicy(scope);
                await operations.deleteProjectAndNamespace(projectInDifferentDepartment.id);
            });

            test('cross-department policy isolation', async function () {
                const cpuCoreRequest = 0.1;
                await _createPolicy(
                    { compute: { cpuCoreRequest: cpuCoreRequest } },
                    undefined,
                    ScopeType.Department,
                    departmentId
                );
                // submit a job from a different department

                const wlType = params.isInteractive
                    ? cliV2WorkloadType.workspace
                    : cliV2WorkloadType.training;
                const submitFlags = {
                    [SubmitFlagsMerged.image]: params.image,
                    [SubmitFlagsMerged.project]: projectName
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(submitFlags, jobName, wlType);
                await _waitForWorkloadRunning(jobName, namespace);

                // ensure that the policy is not applied
                const cpusCount = await _getCpuRequest(namespace);
                expect(cpusCount).to.undefined;
            });
        });
    });
});
