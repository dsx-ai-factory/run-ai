/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    nodeUtils,
    tags,
    waitUtils,
    TrainingWorkload
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { verifyNodeExclusion } from '../../../../../../modules/helpers/Validations';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`Exclude Nodes CLI Tests, tags: ${tags.tool.kubeCli} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let workloadName: string;
        let workerNodeNames: string[];
        const alpineImage = 'alpine:latest';

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-cli-for-node-exclusion');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            workerNodeNames = await nodeUtils.getWorkerNodeNames();
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateShortNameByPrefix('exclude-node-workload', 5);
        });

        test('T5167077: submit a workload with node affinity directly to cluster through kubectl', async function () {
            const excludedNodeNames = Utils.getRandomElements(workerNodeNames, 1);
            const trainingWorkload = new TrainingWorkload(alpineImage, namespace, workloadName);
            const trainingCrd = trainingWorkload.getCrd();
            const trainingBody = trainingWorkload.getBody();

            trainingBody['spec']['nodeAffinityRequired'] = {
                nodeSelectorTerms: [
                    {
                        matchExpressions: [
                            {
                                key: 'kubernetes.io/hostname',
                                operator: 'NotIn',
                                values: excludedNodeNames
                            }
                        ]
                    }
                ]
            };
            await test.step('Submit a training workload with node affinity directly to cluster', async function () {
                await kubeCli.createCrd(trainingCrd, trainingBody, namespace);
                await waitUtils.waitForCrd(trainingCrd, workloadName, namespace, 60);
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        project.id,
                        Phase.Completed,
                        60
                    );
            });

            await test.step('Verify workload is not running on excluded node(s)', async () => {
                await verifyNodeExclusion(excludedNodeNames, namespace, workloadName);
            });

            await test.step('Delete workload', async () => {
                await kubeCli.deleteCrd(trainingCrd, workloadName, namespace);
            });
        });
    });
});
});
