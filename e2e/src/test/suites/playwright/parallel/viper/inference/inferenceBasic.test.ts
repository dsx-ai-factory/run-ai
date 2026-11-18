/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { expect } from 'chai';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { test } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    `inference workload basic tests: ${tags.resource.inferenceWorkload} ${tags.tool.kubeCli} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        test.describe.configure({ retries: 1 });
        let project: Project;
        let workloadName: string;

        test.beforeAll(async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('project'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateNameByPrefix('workload');
        });

        test.afterAll(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        test.afterEach(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await workloadUtils.clearProjectWorkloadsAndPods(project.name, 120);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        test('creates a workload with one replica', async function ({}, testInfo) {
            skipPlaywrightTest(
                test,
                testInfo,
                'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
            );
            test.setTimeout(180 * 1000);

            await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                image: config.quickStartImage.inferenceHelloWorld
            });

            await waitUtils.waitForWorkloadNonEmptyStatus(
                workloadName,
                project.status.namespace,
                crds.runaiInferenceWorkloadCrd
            );

            const knativeService = await waitUtils.waitForKnativeService(
                workloadName,
                project.status.namespace
            );
            expect(knativeService).to.not.be.undefined;

            await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, ['Running'], true);
        });

        test('creates a workload with two replicas', async function ({}, testInfo) {
            skipPlaywrightTest(
                test,
                testInfo,
                'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
            );
            test.setTimeout(180 * 1000);

            await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                image: config.quickStartImage.inferenceHelloWorld,
                minScale: 2,
                maxScale: 2
            });

            await waitUtils.waitForWorkloadNonEmptyStatus(
                workloadName,
                project.status.namespace,
                crds.runaiInferenceWorkloadCrd
            );

            const knativeService = await waitUtils.waitForKnativeService(
                workloadName,
                project.status.namespace
            );
            expect(knativeService).to.not.be.undefined;

            await waitUtils.waitForPods(workloadName, project.status.namespace, 2, 60, ['Running'], true);
        });

        test('deletes the pods when the workload is deleted - container with SIGTERM handling', async function ({}, testInfo) {
            skipPlaywrightTest(
                test,
                testInfo,
                'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
            );
            // the helloworld image has a SIGTERM handler so pod termination is quick
            const workload = await workloadUtils.createInferenceWorkload(
                project.status.namespace,
                workloadName,
                {
                    image: config.quickStartImage.inferenceHelloWorld
                }
            );
            await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

            let pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                'Running'
            ]);
            expect(pods.length).to.equal(1);

            logger.debug('deleting the workload');
            await kubeCli.deleteWorkload(workload);

            pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 10, ['Running']);
            expect(pods.length).to.equal(1);

            await waitUtils.sleep(60 * 1000);

            pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 0, 10);
            expect(pods.length).to.equal(0);
        });
    }
);
