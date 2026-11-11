/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    assetsApiV2Utils,
    config,
    crds,
    flowRootState,
    InteractiveWorkload,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect, test } from '@playwright/test';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

const WorkloadParams = [
    {
        wlType: 'training',
        image: config.quickStartImage.training,
        jobKind: crds.runaiTrainingWorkloadCrd.kind,
        wlCreationObject: (image: string, namespace: string, name: string): TrainingWorkload =>
            new TrainingWorkload(image, namespace, name)
    },
    {
        wlType: 'interactive',
        image: config.quickStartImage.interactive,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind,
        wlCreationObject: (image: string, namespace: string, name: string): InteractiveWorkload =>
            new InteractiveWorkload(image, namespace, name)
    }
];

shouldRunNativeWorkloadCrdTests(() => {
WorkloadParams.forEach(function (params) {
    test.describe(`Multi-fraction ${params.wlType} workload: ${tags.teams.viper} ${tags.resource.jobs}`, function () {
        let jobName: string;
        let project: Project;

        test.beforeEach(async () => {
            jobName = objectGenerator.generateNameByPrefix('w');
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterEach(async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                    logger.error(
                        `Failed to delete project ${project.name} and namespace ${project.namespace}`
                    );
            }
        });

        test.describe('gpuDevices', function () {
            test(`Create ${params.wlType} workload with valid GPU fraction < 1`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '0.4';
                workload.gpuDevices = 2;
                await operations.submitWorkloads([workload], PodStatus.RUNNING);

                    const podAnnotations = (await waitUtils.waitForPods(jobName, project.namespace))[0]
                        ?.metadata?.annotations;
                expect(podAnnotations['gpu-fraction']).toBe('0.4');
                expect(podAnnotations['gpu-fraction-num-devices']).toBe('2');
            });

            test(`Create ${params.wlType} workload with fractional GPU, and validate that the gpuDevices default value is 1`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '0.4';
                await operations.submitWorkloads([workload], PodStatus.RUNNING);

                    const podAnnotations = (await waitUtils.waitForPods(jobName, project.namespace))[0]
                        ?.metadata?.annotations;
                expect(podAnnotations['gpu-fraction']).toBe('0.4');
                expect(podAnnotations['gpu-fraction-num-devices']).toBe('1');
            });

            test(`Creation of ${params.wlType} fails when the value of 'gpuDevices' is 0, 'gpu' should not be specified"`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '0.4';
                workload.gpuDevices = 0;

                await expect(
                    operations.submitWorkloads([workload], PodStatus.RUNNING)
                ).rejects.toThrowError();
            });
        });

        test.describe('gpuMemory', function () {
            test(`Create ${params.wlType} workload with valid gpuMemory`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpuMemory = '1G';
                workload.gpuMemoryLimit = '2G';
                await operations.submitWorkloads([workload], PodStatus.RUNNING);

                const createdWorkload = await operations.getWorkload(
                    jobName,
                    project.namespace,
                    params.jobKind
                );
                expect(createdWorkload['spec'].gpuMemory.value).toBe('1G');
                expect(createdWorkload['spec'].gpuMemoryLimit.value).toBe('2G');
            });
            //Isn't relate to multi-frac, but since we don't have this test i'm adding it anyway
            test(`Submission of ${params.wlType} workload with gpuMemory > gpuMemoryLimit should fail`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpuMemory = '2G';
                workload.gpuMemoryLimit = '1G';
                await expect(
                    operations.submitWorkloads([workload], PodStatus.RUNNING)
                ).rejects.toThrowError();
            });

            test(`Submission of ${params.wlType} workload should fail if gpuLimit provided with non-fractional value`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '1';
                workload.gpuLimit = '0.5';
                await expect(
                    operations.submitWorkloads([workload], PodStatus.RUNNING)
                ).rejects.toThrowError();
            });

            test(`Submission of ${params.wlType} workload with both gpu and gpuMemory should fail`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '0.5';
                workload.gpuMemory = '1G';
                await expect(
                    operations.submitWorkloads([workload], PodStatus.RUNNING)
                ).rejects.toThrowError();
            });
        });

        test.describe('gpuLimit', function () {
            test(`Submission of ${params.wlType} workload with gpuLimit > 1 should fail`, async function () {
                const workload = params.wlCreationObject(params.image, project.namespace, jobName);
                workload.gpu = '0.4';
                workload.gpuLimit = '2';
                await expect(
                    operations.submitWorkloads([workload], PodStatus.RUNNING)
                ).rejects.toThrowError();
            });
        });
    });
});
});
