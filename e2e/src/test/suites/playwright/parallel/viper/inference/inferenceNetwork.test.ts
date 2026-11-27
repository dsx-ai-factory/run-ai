/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { expect } from 'chai';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';
import { InferencePhase } from '../../../../../../modules/helpers/utils/WaitUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

supportedFrom(SupportedVersions.V2_17)(function () {
    test.describe(
        `inference workload network tests: ${tags.resource.inferenceWorkload} ${tags.tool.kubeCli} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
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
                if (!project) {
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

            test('creates a workload with valid readiness probe', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                logger.info('creating a workload with valid readiness probe');
                const workload = await workloadUtils.createInferenceWorkload(
                    project.status.namespace,
                    workloadName,
                    {
                        readinessProbe: {
                            httpGetPath: '/',
                            httpGetPort: 8080
                        }
                    }
                );
                logger.info('waiting for workload to be created');
                await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);
                logger.info('waiting for pods to be running');
                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                    'Running'
                ]);
                expect(pods.length).to.equal(1);

                // we expect both containers to become ready - user-container and queue-proxy
                logger.info('waiting for 2 containers to become ready');
                await waitUtils.waitForContainersToBeReady(
                    workloadName,
                    project.status.namespace,
                    pods[0].metadata.name,
                    2
                );
            });

            test('creates a workload with invalid readiness probe', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                logger.info('creating a workload with invalid readiness probe');
                const workload = await workloadUtils.createInferenceWorkload(
                    project.status.namespace,
                    workloadName,
                    {
                        readinessProbe: {
                            httpGetPort: 6060 // no response from this port
                        }
                    }
                );
                logger.info('waiting for workload to be created');
                await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);
                logger.info('waiting for pods to be running');
                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                    'Running'
                ]);
                expect(pods.length).to.equal(1);

                // we intentionally give the containers time, as if we check too early, the test might pass not due to the probe
                // but due to the fact that the container is not ready yet
                await waitUtils.sleep(5000);

                // we expect only user-container to become ready, but not queue-proxy
                logger.info('waiting for 1 to become ready');
                await waitUtils.waitForContainersToBeReady(
                    workloadName,
                    project.status.namespace,
                    pods[0].metadata.name,
                    1
                );
            });

            // TODO: Enable this test when Knative hostIpc/hostNetwork feature flags are enabled by default
            // This test modifies the cluster-wide Knative config-features ConfigMap which can affect
            // other tests running in parallel. To test locally, uncomment this block and ensure no
            // other inference tests are running simultaneously.
            //
            // supportedFrom(SupportedVersions.V2_25)(function () {
            //     const knativeNamespace = 'knative-serving';
            //     const configMapName = 'config-features';
            //     let originalHostIpc: string | undefined;
            //     let originalHostNetwork: string | undefined;
            //
            //     test.beforeAll(async function () {
            //         const cm = await kubeCli.getConfigMap(configMapName, knativeNamespace);
            //         originalHostIpc = cm.data?.['kubernetes.podspec-hostipc'];
            //         originalHostNetwork = cm.data?.['kubernetes.podspec-hostnetwork'];
            //         if (originalHostIpc !== 'enabled' || originalHostNetwork !== 'enabled') {
            //             cm.data = cm.data || {};
            //             cm.data['kubernetes.podspec-hostipc'] = 'enabled';
            //             cm.data['kubernetes.podspec-hostnetwork'] = 'enabled';
            //             await kubeCli.updateConfigMap(configMapName, knativeNamespace, cm);
            //         }
            //     });
            //
            //     test.afterAll(async function () {
            //         const cm = await kubeCli.getConfigMap(configMapName, knativeNamespace);
            //         cm.data = cm.data || {};
            //         if (originalHostIpc === undefined) {
            //             delete cm.data['kubernetes.podspec-hostipc'];
            //         } else {
            //             cm.data['kubernetes.podspec-hostipc'] = originalHostIpc;
            //         }
            //         if (originalHostNetwork === undefined) {
            //             delete cm.data['kubernetes.podspec-hostnetwork'];
            //         } else {
            //             cm.data['kubernetes.podspec-hostnetwork'] = originalHostNetwork;
            //         }
            //         await kubeCli.updateConfigMap(configMapName, knativeNamespace, cm);
            //     });
            //
            //     test('should submit inference with hostIpc and hostNetwork and verify on pods', async function () {
            //         workloadsCrudUtils.accessToken =
            //             flowRootState.users.system_administrator.accessToken;
            //         const workloadSpec: InferenceSpecSpec = {
            //             image: config.quickStartImage.inferenceHelloWorld,
            //             servingPort: { container: 8080 },
            //             security: { hostIpc: true, hostNetwork: true }
            //         };
            //         const result = await workloadsCrudUtils.createInference(
            //             workloadName, false, flowRootState.cluster.uuid,
            //             project.id.toString(), workloadSpec
            //         );
            //         expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            //         const pods = await waitUtils.waitForPods(
            //             result.body.name, project.status.namespace, 1, 90, ['Running'], true
            //         );
            //         expect(pods.length).to.equal(1);
            //         expect(pods[0].spec.hostIPC).to.equal(true);
            //         expect(pods[0].spec.hostNetwork).to.equal(true);
            //     });
            // });

            supportedFrom(SupportedVersions.V2_24)(function () {
                test('creates a workload with serving port auth and valid readiness probe', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                    );
                    logger.info('creating a workload with authorization and valid readiness probe');
                    const workload = await workloadUtils.createInferenceWorkload(
                        project.status.namespace,
                        workloadName,
                        {
                            servingPortAuthorizationType: 'authenticatedUsers',
                            readinessProbe: {
                                httpGetPath: '/',
                                httpGetPort: 8080
                            }
                        }
                    );

                    logger.info('waiting for workload to be in phase Ready');
                    await waitUtils.waitForInferenceWorkloadPhase(workload, InferencePhase.Ready, 90);
                });

                test('creates a workload with serving port auth and invalid readiness probe - should not become ready', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                    );
                    logger.info('creating a workload with authorization and invalid readiness probe');
                    const workload = await workloadUtils.createInferenceWorkload(
                        project.status.namespace,
                        workloadName,
                        {
                            servingPortAuthorizationType: 'authenticatedUsers',
                            readinessProbe: {
                                httpGetPort: 6060 // wrong port - hello-world image only responds on 8080
                            }
                        }
                    );

                    logger.info('waiting for workload to be in phase Initializing');
                    await waitUtils.waitForInferenceWorkloadPhase(workload, InferencePhase.Initializing, 90);

                    // wait 10 more seconds and verify the workload is still in phase Initializing
                    await waitUtils.sleep(10000);
                    logger.info(
                        'waiting for workload to be in phase Initializing after sleeping for 10 seconds'
                    );
                    await waitUtils.waitForInferenceWorkloadPhase(workload, InferencePhase.Initializing, 10);
                });
            });
        }
    );
});
