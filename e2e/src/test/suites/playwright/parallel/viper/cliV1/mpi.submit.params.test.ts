/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    objectGenerator,
    logger,
    tags,
    waitUtils,
    config,
    Utils,
    projectResourcesUtils
} from '../../../../../../modules/helpers';
import { operations } from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { DistributedWorkload } from '../../../../../../modules/helpers';
import { isOCP } from '../../../../../../modules/helpers/utils/EnvironmentUtils';
import { submitMpiWorkload } from '../../../../../../modules/helpers/models/workloads';
import {
    shouldRunNativeWorkloadCrdTests,
    shouldRunTests
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`Submit mpi training job by CLI:  ${tags.tool.api} ${tags.teams.viper} `, function () {
    test.describe('runai job parameters - mpi', function () {
        shouldRunTests(!isOCP())(function () {
            let projectName: string, namespace: string;
            let project: Project;
            let workload: DistributedWorkload;
            let jobName: string;
            const defaultImagePath = config.quickStartImage.ubuntu;
            const sleepInfinity = 'sleep infinity';

            test.beforeEach(async function () {
                projectName = objectGenerator.generateNameByPrefix('project');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 8,
                    maxAllowedGpus: 8
                });

                jobName = objectGenerator.generateNameByPrefix('job');

                // CLI v1 resolves projects by listing namespaces with label runai/queue=<project>,
                // so the namespace must exist in the cluster before any CLI v1 command can reference the project.
                await waitUtils.waitForNamespaceToExist(namespace);
            });

            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName}`);
                } catch {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('should submit MPI job and validate its data', async function () {
                    workload = await submitMpiWorkload(
                        'timeout-mpi',
                        project,
                        sleepInfinity,
                        defaultImagePath
                    );
                const pods = await waitUtils.waitForPods(
                    workload.name,
                    namespace,
                    3,
                    undefined,
                    undefined,
                    true
                );
                const masterPod = pods.find(
                    (p) =>
                        p.metadata.labels['mpi-job-role'] === 'launcher' ||
                        p.metadata.labels['training.kubeflow.org/job-role'] === 'master' ||
                        p.metadata.labels['training.kubeflow.org/job-role'] === 'launcher'
                );
                expect(masterPod).not.to.be.undefined;
                logger.debug(`Master pod: ${JSON.stringify(masterPod, null, 2)}`);
                expect(masterPod.metadata.annotations).to.include.keys('totalGPUs');
                expect(masterPod.metadata.annotations['totalGPUs']).to.equal('2'); // 1 for each worker
                expect(masterPod.spec.schedulerName).to.equal('runai-scheduler');

                // need to validate that code in the keys.exist
                const workerPods = pods.filter(
                    (p) =>
                        p.metadata.labels['mpi-job-role'] === 'worker' ||
                        p.metadata.labels['training.kubeflow.org/replica-type'] === 'worker' ||
                        p.metadata.labels['training.kubeflow.org/job-role'] === 'worker'
                );
                logger.info(`found ${workerPods.length} worker pods`);
                expect(workerPods.length).to.equal(2);
                for (const pod of workerPods) {
                    logger.debug(
                        `Worker pod annotations: ${JSON.stringify(pod.metadata.annotations, null, 2)}`
                    );
                    logger.debug(`Worker pod spec: ${JSON.stringify(pod.spec, null, 2)}`);
                    expect(pod.spec.schedulerName).to.equal('runai-scheduler');
                }
            });
        });
    });
});
});
