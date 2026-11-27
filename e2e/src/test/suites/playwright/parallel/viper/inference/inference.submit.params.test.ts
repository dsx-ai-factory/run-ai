/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    authUtils,
    config,
    flowRootState,
    inferenceUtils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { User as UserV2 } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { kubeCli } from '../../../../../../modules/cli';
import { execSync } from 'child_process';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

// more advanced tests are in advanceSuite/inference

test.describe(
    `Inference workloads over kubectl, tags: ${tags.runOn.parallelCI} ${tags.resource.inferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    () => {
        test.describe.configure({ retries: 1 });
        let project: Project;
        let namespace: string;
        let inferenceName: string;
        let projectName: string;
        let user: UserV2;

        test.beforeAll('creating project', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            projectName = objectGenerator.generateNameByPrefix('project');
            namespace = Utils.namespaceFromProjectName(projectName);
            user = await authUtils.getUserV2(
                config.controlPlaneAdminUsername,
                flowRootState.users.system_administrator.accessToken
            );
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 2,
                maxAllowedGpus: 2,
                permissions: { users: [user.username], groups: [] }
            });
            logger.info(`created project ${projectName}`);
        });

        test.beforeEach('inference name', function () {
            inferenceName = objectGenerator.generateNameByPrefix('inference');
        });
        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
            }
        });
        test(`creates workload in cluster and control-plane with 1 replicas ${tags.runOn.parallelCI}`, async () => {
            const spec: InferenceSpecSpec = {
                image: config.quickStartImage.inferenceHelloWorld,
                servingPort: { container: 8080 },
                autoscaling: {
                    minReplicas: 1,
                    maxReplicas: 1,
                    metric: 'throughput',
                    metricThreshold: 50
                }
            };
            const [workloadId] = await inferenceUtils.createInference(
                inferenceName,
                spec,
                String(project.id)
            );
            logger.debug(`created inference ${inferenceName} in cluster and control-plane`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            await waitUtils.waitForPods(inferenceName, namespace, 1, 50, ['Running'], true);
            logger.debug(`successfully created inference ${inferenceName} and Running`);
        });

        supportedFrom(SupportedVersions.V2_24)(function () {
            test(`creates workload with custom working directory ${tags.runOn.parallelCI}`, async () => {
                const customWorkingDir = '/home';
                const spec: InferenceSpecSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    workingDir: customWorkingDir,
                    servingPort: { container: 8080 },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1,
                        metric: 'throughput',
                        metricThreshold: 50
                    }
                };
                const [workloadId] = await inferenceUtils.createInference(
                    inferenceName,
                    spec,
                    String(project.id)
                );
                logger.debug(`created inference ${inferenceName} with workingDir ${customWorkingDir}`);
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                await waitUtils.waitForPods(inferenceName, namespace, 1, 50, ['Running'], true);
                logger.debug(`successfully created inference ${inferenceName} and Running`);

                // Verify the actual pod's working directory (user-container only, not queue-proxy)
                const pods = await kubeCli.listPodsOfJob(namespace, inferenceName, true);
                expect(pods.length).to.be.greaterThan(0, 'Expected at least one pod');
                const podName = pods[0].metadata.name;
                const { stdout } = await kubeCli.execInPod(podName, namespace, ['pwd']);
                logger.debug(`pod ${podName} working directory: ${stdout}`);
                expect(stdout).to.equal(
                    customWorkingDir,
                    `Expected workingDir to be ${customWorkingDir} but got ${stdout}`
                );
            });
        });
    }
);
