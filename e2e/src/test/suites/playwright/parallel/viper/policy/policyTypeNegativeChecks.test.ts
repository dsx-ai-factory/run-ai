/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { policyApi as policyApiHandler } from '../../../../../../modules/api';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    config,
    crds,
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
import { test, TestInfo } from '@playwright/test';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(`policy type - negative checks: ${tags.resource.policy} ${tags.tool.api} ${tags.teams.viper}`, function () {
    let projectName: string;
    let namespace: string;
    let project: Project;
    let jobName: string;
    let policyName: string;

    async function _createPolicy(kind: WorkloadType, defaults: object): Promise<void> {
        await policyUtils.createPolicy(
            kind === WorkloadType.Interactive,
            policyName,
            project.namespace,
            project.id,
            kind,
            defaults
        );
    }

    async function _deletePolicy(kind: WorkloadType): Promise<void> {
        if (kind === WorkloadType.Interactive) {
            await policyApiHandler.policyApi.deleteWorkspacePolicy(
                'project',
                undefined,
                project.id.toString()
            );
        } else {
            await policyApiHandler.policyApi.deleteTrainingPolicy(
                'project',
                undefined,
                project.id.toString()
            );
        }
    }

    async function _waitForWorkloadRunning(
        workloadKind: string,
        submissionType: AnyWorkloadSubmissionType
    ): Promise<void> {
        if (await isAnyWorkloadSubmissionEnabled(submissionType)) {
            await waitUtils.waitForBatchJobToExist(jobName, namespace);
            await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);
        } else {
            await waitUtils.waitForJobToBeRunning(jobName, namespace, workloadKind);
        }
    }

    async function _getCpuRequest(
        crd: typeof crds.runaiInteractiveWorkloadCrd,
        submissionType: AnyWorkloadSubmissionType
    ): Promise<number | string | undefined> {
        if (await isAnyWorkloadSubmissionEnabled(submissionType)) {
            const job = await kubeCli.getBatchJob(jobName, namespace);
            return job.spec?.template.spec?.containers[0].resources?.requests?.cpu;
        }
        return workloadUtils.getCpusCount(jobName, namespace, crd);
    }

    test.beforeAll(async function () {
        policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;

        // Enable control plane policies so they are applied to workloads
        await policyUtils.enableClusterPolicySync();
    });

    test.beforeEach(async function ({}, testInfo: TestInfo) {
        jobName = objectGenerator.generateShortNameByPrefix('job-for-policy');
        project = await projectResourcesUtils.createProjectAndNamespace('project-for-policy');
        projectName = project.name;
        namespace = project.namespace;
        policyName = `project-${project.id}`;
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
    test('training policy has no impact on interactive workloads', async function () {
        await _createPolicy(WorkloadType.Training, { compute: { cpuCoreRequest: 1 } });

        const submitFlags = {
            [SubmitFlagsMerged.image]: config.quickStartImage.interactive,
            [SubmitFlagsMerged.project]: project.name
        } as Partial<typeof SubmitFlagsMerged>;
        runaiCliV2.submit(submitFlags, jobName, cliV2WorkloadType.workspace);

        await _waitForWorkloadRunning(
            crds.runaiInteractiveWorkloadCrd.kind,
            AnyWorkloadSubmissionType.Workspace
        );

        // ensure that the policy is not applied
        const cpusCount = await _getCpuRequest(
            crds.runaiInteractiveWorkloadCrd,
            AnyWorkloadSubmissionType.Workspace
        );
        expect(cpusCount).to.undefined;

        await _deletePolicy(WorkloadType.Training);
    });
    test('workspace policy has no impact on training workloads', async function () {
        await _createPolicy(WorkloadType.Interactive, { compute: { cpuCoreRequest: 1 } });

        // submit an interactive workload

        const submitFlags = {
            [SubmitFlagsMerged.image]: config.quickStartImage.training,
            [SubmitFlagsMerged.project]: project.name
        } as Partial<typeof SubmitFlagsMerged>;
        runaiCliV2.submit(submitFlags, jobName, cliV2WorkloadType.training);
        await _waitForWorkloadRunning(crds.runaiTrainingWorkloadCrd.kind, AnyWorkloadSubmissionType.Training);

        // ensure that the policy is not applied
        const cpusCount = await _getCpuRequest(
            crds.runaiTrainingWorkloadCrd,
            AnyWorkloadSubmissionType.Training
        );
        expect(cpusCount).to.undefined;

        await _deletePolicy(WorkloadType.Interactive);
    });
});
