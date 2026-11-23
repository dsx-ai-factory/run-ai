/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    tags,
    waitUtils,
    workloadUtils,
    objectGenerator,
    Utils,
    projectResourcesUtils,
    logger,
    operations,
    crds,
    CustomResourceDefinition
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

interface frameworkParams {
    type: cliV2WorkloadType;
}

const xgboostParams: frameworkParams = {
    type: cliV2WorkloadType.xgBoost
};
const mpiParams: frameworkParams = {
    type: cliV2WorkloadType.mpi
};

const tfParams: frameworkParams = {
    type: cliV2WorkloadType.tensorFlow
};

const pytorchParams: frameworkParams = {
    type: cliV2WorkloadType.pyTorch
};

interface DistributedTestEnvironment {
    projectName: string;
    project: Project;
    namespace: string;
    workloadName: string;
}

export const paramsAllFrameworks: frameworkParams[] = [tfParams, xgboostParams, mpiParams, pytorchParams];

function getConcreteJobCrd(type: cliV2WorkloadType): CustomResourceDefinition {
    switch (type) {
        case cliV2WorkloadType.mpi:
            return crds.mpiJobCrd;
        case cliV2WorkloadType.pyTorch:
            return crds.pyTorchJobCrd;
        case cliV2WorkloadType.tensorFlow:
            return crds.tfJobCrd;
        case cliV2WorkloadType.xgBoost:
            return crds.xgboostJobCrd;
        default:
            throw new Error(`Unsupported distributed workload type: ${type}`);
    }
}

test.describe(
    `Distributed workload phase  ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.runOn.parallelCI, tags.resource.distributedTrainingWorkload] },
    function () {
        let testEnvironment: DistributedTestEnvironment;
        const sleepInfinity = 'sleep infinity';
        const failPod = '-- cat nonExistCommand';
        const image = config.quickStartImage.ubuntu;
        test.afterEach(async function () {
            try {
                if (testEnvironment?.project) {
                    await operations.deleteProjectAndNamespace(testEnvironment.project.id);
                }
            } catch (error) {
                logger.warn(`Failed to delete project: ${error}`);
            }
        });

        async function createDistributedWorkload(
            params: frameworkParams,
            workloadImage: string,
            workloadFields: { command?: string; gpu?: string } = {}
        ): Promise<DistributedTestEnvironment> {
            const projectName = objectGenerator.generateNameByPrefix('project');
            const namespace = Utils.namespaceFromProjectName(projectName);
            const project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            const workloadName = objectGenerator.generateNameByPrefix('job');
            const submitFlags = {
                [SubmitFlagsMerged.image]: workloadImage,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.workers]: '2'
            } as Partial<typeof SubmitFlagsMerged>;
            if (workloadFields.command) {
                submitFlags[SubmitFlagsMerged.command] = workloadFields.command.startsWith('--')
                    ? workloadFields.command
                    : `-- ${workloadFields.command}`;
            }
            if (workloadFields.gpu) {
                submitFlags[SubmitFlagsMerged.gpuRequest] = workloadFields.gpu;
            }

            runaiCliV2.submit(submitFlags, workloadName, params.type);
            await workloadUtils.waitForDistributedSubmissionRoot(workloadName, namespace);
            await waitUtils.waitForCrd(getConcreteJobCrd(params.type), workloadName, namespace);
            await waitUtils.waitForPods(workloadName, namespace);

            return { projectName, project, namespace, workloadName };
        }

        async function validateStatusForPeriod(seconds: number): Promise<void> {
            for (let i = 0; i < seconds; i++) {
                await waitUtils.waitForBackendWorkloadPhase(
                    testEnvironment.workloadName,
                    testEnvironment.project.id,
                    Phase.Initializing
                );
                await Utils.delay(1000);
            }
        }

        paramsAllFrameworks.forEach(function (params) {
            test(`completed ${params.type}`, async function () {
                testEnvironment = await createDistributedWorkload(params, image);
                await waitUtils.waitForBackendWorkloadPhase(
                    testEnvironment.workloadName,
                    testEnvironment.project.id,
                    Phase.Completed
                );
            });
            test(`initializing ${params.type}`, async function () {
                testEnvironment = await createDistributedWorkload(params, 'invalid-image');
                await validateStatusForPeriod(3);
            });
            test(`running ${params.type}`, async function () {
                testEnvironment = await createDistributedWorkload(params, image, { command: sleepInfinity });
                await waitUtils.waitForBackendWorkloadPhase(
                    testEnvironment.workloadName,
                    testEnvironment.project.id,
                    Phase.Running
                );
            });
            test(`pending ${params.type}`, async function () {
                testEnvironment = await createDistributedWorkload(params, image, { gpu: '100' });
                await waitUtils.waitForBackendWorkloadPhase(
                    testEnvironment.workloadName,
                    testEnvironment.project.id,
                    Phase.Pending
                );
            });
            test(`failed ${params.type}`, async function () {
                const projectName = objectGenerator.generateNameByPrefix('project');
                const namespace = Utils.namespaceFromProjectName(projectName);
                const project: Project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });
                const submitFlag = {
                    [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                    [SubmitFlagsMerged.backoffLimit]: 1,
                    [SubmitFlagsMerged.command]: failPod,
                    [SubmitFlagsMerged.project]: project.name
                } as Partial<typeof SubmitFlagsMerged>;
                const workloadName = objectGenerator.generateNameByPrefix('job');
                runaiCliV2.submit(submitFlag, workloadName, params.type);
                testEnvironment = { projectName, project, namespace, workloadName };
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Failed);
            });
        });
    }
);
