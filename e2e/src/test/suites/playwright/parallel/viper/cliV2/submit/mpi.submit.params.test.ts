/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import {
    config,
    crds,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { V1Pod } from '@kubernetes/client-node';
import { test } from '@playwright/test';

const getPodGpuLimit = (pod: V1Pod): number =>
    (pod.spec?.containers ?? []).reduce(
        (total, container) =>
            total + parseInt((container.resources?.limits?.['nvidia.com/gpu'] as string) ?? '0'),
        0
    );

test.describe(`Submit MPI training with CLI v2: ${tags.tool.api} ${tags.teams.viper}`, function () {
    shouldRunTests(!isOCP())(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let jobName: string;

        test.beforeEach(async function () {
            projectName = objectGenerator.generateNameByPrefix('project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 8
            });
            jobName = objectGenerator.generateNameByPrefix('job');
        });

        test.afterEach(async function () {
            try {
                runaiCliV2.delete({ '-p': projectName } as any, jobName, cliV2WorkloadType.mpi);
            } catch (error) {
                logger.warn(`Failed to delete MPI workload ${jobName}: ${error}`);
            }
            await operations.deleteProjectAndNamespace(project.id);
        });

        test('should submit MPI job through the active route and validate its data', async function () {
            const submitFlags = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.command]: '-- sleep infinity',
                [SubmitFlagsMerged.gpuRequest]: '1',
                [SubmitFlagsMerged.workers]: '2'
            } as Partial<typeof SubmitFlagsMerged>;
            runaiCliV2.submit(submitFlags, jobName, cliV2WorkloadType.mpi);

            await workloadUtils.waitForDistributedSubmissionRoot(jobName, namespace);
            await waitUtils.waitForCrd(crds.mpiJobCrd, jobName, namespace);
            const pods = await waitUtils.waitForPods(jobName, namespace, 3, undefined, undefined, true);
            const masterPod = pods.find(
                (pod) =>
                    pod.metadata.labels['mpi-job-role'] === 'launcher' ||
                    pod.metadata.labels['training.kubeflow.org/job-role'] === 'master' ||
                    pod.metadata.labels['training.kubeflow.org/job-role'] === 'launcher'
            );
            expect(masterPod).not.to.be.undefined;
            expect(getPodGpuLimit(masterPod)).to.equal(0);
            expect(masterPod.spec.schedulerName).to.equal('runai-scheduler');

            const workerPods = pods.filter(
                (pod) =>
                    pod.metadata.labels['mpi-job-role'] === 'worker' ||
                    pod.metadata.labels['training.kubeflow.org/replica-type'] === 'worker' ||
                    pod.metadata.labels['training.kubeflow.org/job-role'] === 'worker'
            );
            expect(workerPods.length).to.equal(2);
            for (const pod of workerPods) {
                expect(getPodGpuLimit(pod)).to.equal(1);
                expect(pod.spec.schedulerName).to.equal('runai-scheduler');
            }
            expect(pods.reduce((total, pod) => total + getPodGpuLimit(pod), 0)).to.equal(2);
        });
    });
});
