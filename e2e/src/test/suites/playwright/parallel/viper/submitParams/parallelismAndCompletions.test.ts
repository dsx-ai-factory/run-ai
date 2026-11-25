/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    computeResourceUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    trainingUtils,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { expect } from 'chai';
import { TrainingCreationRequest } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

test.describe(`Create a workload with parallelism and completions ${tags.teams.viper}`, function () {
    let project: Project;
    let projectName: string;
    let namespace: string;
    let trainingName: string;
    let anyWorkloadSubmissionEnabled: boolean;

    test.beforeAll('create project and namespace', async function () {
        projectName = objectGenerator.generateNameByPrefix('p-training');
        trainingName = objectGenerator.generateNameByPrefix('training');
        namespace = Utils.namespaceFromProjectName(projectName);

        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 1,
            maxAllowedGpus: 1
        });

        anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
            AnyWorkloadSubmissionType.Training
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

    test('check parallelism and completions propagates to the workload', async function () {
        const environmentId = await environmentUtils.getId('jupyter-lab');
        const computeId = await computeResourceUtils.getId('cpu-only');

        const completions = 3;
        const parallelism = 2;

        const trainingCreationRequest: TrainingCreationRequest = {
            name: trainingName,
            projectId: project.id.toString(),
            clusterId: flowRootState.cluster.uuid,
            assets: {
                environment: environmentId,
                compute: computeId
            },
            spec: {
                parallelism,
                completions
            }
        };

        await trainingUtils.createTraining(trainingCreationRequest);

        if (anyWorkloadSubmissionEnabled) {
            await waitUtils.waitForBatchJobToExist(trainingName, namespace);
            const job = await kubeCli.getBatchJob(trainingName, namespace);
            expect(job.spec?.completions).to.equal(completions);
            expect(job.spec?.parallelism).to.equal(parallelism);
        } else {
            await waitUtils.waitForTrainingWorkloadToExist(trainingName, namespace);
            const trainingWorkload: any = await kubeCli.getTrainWorkloadOfRunaiJob(trainingName, namespace);
            expect(trainingWorkload.spec.completions.value).to.equal(completions);
            expect(trainingWorkload.spec.parallelism.value).to.equal(parallelism);
        }
    });
});
