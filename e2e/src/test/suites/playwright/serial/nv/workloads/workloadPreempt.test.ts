/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    objectGenerator,
    logger,
    flowRootState,
    operations,
    Utils,
    projectResourcesUtils,
    workloadsCrudUtils,
    config
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';

test.describe(`worload preempt tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let namespace: string;
        let project: Project;
        let projectName: string;

        test.beforeAll(async () => {
            const projectName = objectGenerator.generateNameByPrefix('project');
            const namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test('T4039406: workload preempt', async () => {
            // Constants for hardcoded strings
            const sleepInfinityCommand = 'sleep infinity';
            const lowPriorityClass = 'low';
            const highPriorityClass = 'high';

            let lowPriorityTraining: any;
            let highPriorityTraining: any;

            await test.step('Create lower priority training', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    objectGenerator.generateNameByPrefix('low-priority'),
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: sleepInfinityCommand,
                        compute: { gpuDevicesRequest: 1 },
                        priorityClass: lowPriorityClass
                    }
                );
                lowPriorityTraining = training;
                await workloadsCrudUtils.waitForWorkloadPhase(
                    lowPriorityTraining.workloadId,
                    Phase.Running,
                    WorkloadType.TRAINING,
                    undefined,
                    120
                );
            });

            await test.step('Create multiple higher priority training workloads', async () => {
                const { body: training } = await workloadsCrudUtils.createTraining(
                    objectGenerator.generateNameByPrefix('high-priority'),
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: sleepInfinityCommand,
                        compute: { gpuDevicesRequest: 1 },
                        priorityClass: highPriorityClass
                    }
                );
                highPriorityTraining = training;

                await workloadsCrudUtils.waitForWorkloadPhase(
                    highPriorityTraining.workloadId,
                    Phase.Running,
                    WorkloadType.TRAINING,
                    undefined,
                    120
                );
            });

            await test.step('Check preemption behavior', async () => {
                // wait for low priority workload to be pending
                await workloadsCrudUtils.waitForWorkloadPhase(
                    lowPriorityTraining.workloadId,
                    Phase.Pending,
                    WorkloadType.TRAINING,
                    undefined,
                    120
                );
            });

            await test.step('delete all workloads', async () => {
                await Promise.all([
                    workloadsCrudUtils.deleteTraining(lowPriorityTraining.workloadId),
                    workloadsCrudUtils.deleteTraining(highPriorityTraining.workloadId)
                ]);
            });
        });
    });
});
