/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { expect } from 'chai';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { inferenceUtils } from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { test } from '@playwright/test';

test.describe(
    `inference workload storage tests: ${tags.resource.inferenceWorkload} ${tags.tool.kubeCli} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        function verifyContainerStatus(pod: any): void {
            // two containers are expected, and both should be in state running. then, it means that the original command
            // succeeded, i.e. the data source path is valid (check that the pod is in status running is not sufficient)
            expect(pod.status.containerStatuses.length).to.equal(2);
            expect(pod.status.containerStatuses[0].state.running).not.to.be.undefined;
            expect(pod.status.containerStatuses[1].state.running).not.to.be.undefined;
        }
        let project: Project;

        // this is the command that knative's hello world image is using
        // we run it after our custom commands in order to get the server up and running
        const HELLO_WORLD_SERVER_COMMAND = '/server';

        test.beforeAll(async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('proj'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(function () {
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
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

        shouldRunNativeWorkloadCrdTests(() =>
        test('creates a workload with large shm', async function () {
            const workloadName = objectGenerator.generateNameByPrefix('wl');
            const workload = await workloadUtils.createInferenceWorkload(
                project.status.namespace,
                workloadName,
                {
                    command: `/bin/sh -c "echo hello > /dev/shm/hello && df /dev/shm && ${HELLO_WORLD_SERVER_COMMAND}"`,
                    largeShm: true
                }
            );
            await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

            const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 90, [
                'Running'
            ]);
            expect(pods.length).to.equal(1);
            })
        );

        supportedFrom(SupportedVersions.V2_19)(function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with nfs', async () => {
                const workloadName = objectGenerator.generateNameByPrefix('wl');
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceHelloWorld,
                    gpus: 0.5,
                    minScale: 1,
                    maxScale: 2,
                    nfs: {
                        nfs1: {
                            sourcePath: '/source',
                            targetPath: '/path',
                            readOnly: true,
                            nfsServer: 'hostname'
                        }
                    }
                });
                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace);
                expect(pods.length).to.equal(1);
                const pod = pods[0];

                const hasExpectedVolume = pod.spec.volumes.some(
                    (volume) => volume?.nfs?.path === '/source' && volume?.nfs?.server === 'hostname'
                );

                const hasExpectedVolumeMount = pod.spec.containers[0].volumeMounts.some(
                    (volumeMount) =>
                        volumeMount.name === 'nfs-volume-0' &&
                        volumeMount.mountPath === '/path' &&
                        volumeMount.readOnly === true
                );

                expect(hasExpectedVolume).to.true;
                expect(hasExpectedVolumeMount).to.true;
                })
            );
        });

        supportedFrom(SupportedVersions.V2_19)(function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with hostPath', async () => {
                const workloadName = objectGenerator.generateNameByPrefix('wl');
                await workloadUtils.createInferenceWorkload(project.status.namespace, workloadName, {
                    image: config.quickStartImage.inferenceHelloWorld,
                    gpus: 0.5,
                    minScale: 1,
                    maxScale: 2,
                    volumes: {
                        hostpath1: {
                            sourcePath: '/source',
                            targetPath: '/path'
                        }
                    }
                });
                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace);
                expect(pods.length).to.equal(1);
                const pod = pods[0];

                const hasExpectedVolume = pod.spec.volumes.some(
                    (volume) => volume?.hostPath?.path === '/source'
                );

                const hasExpectedVolumeMount = pod.spec.containers[0].volumeMounts.some(
                    (volumeMount) =>
                        volumeMount.name === 'host-path-volume-0' && volumeMount.mountPath === '/path'
                );

                expect(hasExpectedVolume).to.true;
                expect(hasExpectedVolumeMount).to.true;
                })
            );
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test('creates a workload with existing pvc', async function () {
            const workloadName = objectGenerator.generateNameByPrefix('wl');
            const pvcName = objectGenerator.generateShortNameByPrefix(`pvc-inf-new-${workloadName}`);
            const pvc = await dataSourcesUtils.createClusterPvc(pvcName, project.status.namespace);

            const workload = await workloadUtils.createInferenceWorkload(
                project.status.namespace,
                workloadName,
                {
                    command: `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`,
                    pvcs: {
                        pvc1: {
                            existingPvc: true,
                            claimName: pvc.metadata.name,
                            path: '/path'
                        }
                    }
                }
            );
            await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

            const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                'Running'
            ]);
            expect(pods.length).to.equal(1);

            verifyContainerStatus(pods[0]);
            })
        );

        supportedFrom(SupportedVersions.V2_16)(function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with configmap', async function () {
                const workloadName = objectGenerator.generateNameByPrefix('wl');
                const cmName = objectGenerator.generateShortNameByPrefix(`cm-inf-${workloadName}`);
                    const cm = await dataSourcesUtils.createClusterConfigMap(
                        cmName,
                        project.status.namespace
                    );
                const workload = await workloadUtils.createInferenceWorkload(
                    project.status.namespace,
                    workloadName,
                    {
                        command: `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`,
                        configMapVolumes: {
                            cm1: {
                                name: cm.metadata.name,
                                mountPath: '/path'
                            }
                        }
                    }
                );
                await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                    'Running'
                ]);
                expect(pods.length).to.equal(1);

                verifyContainerStatus(pods[0]);
                })
            );
        });

        supportedFrom(SupportedVersions.V2_17)(function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with git', async function () {
                const workloadName = objectGenerator.generateNameByPrefix('wl');
                const workload = await workloadUtils.createInferenceWorkload(
                    project.status.namespace,
                    workloadName,
                    {
                        command: `/bin/sh -c "ls /path/Hello-World && ${HELLO_WORLD_SERVER_COMMAND}"`,
                        gitSync: {
                            'git-1': {
                                repository: 'https://github.com/octocat/Hello-World',
                                branch: 'master',
                                directory: '/path'
                            }
                        }
                    }
                );
                await waitUtils.waitForInferenceWorkload(workload.name, project.status.namespace);

                const pods = await waitUtils.waitForPods(workloadName, project.status.namespace, 1, 60, [
                    'Running'
                ]);
                expect(pods.length).to.equal(1);

                verifyContainerStatus(pods[0]);
                })
            );
        });

        supportedFrom(SupportedVersions.V2_20)(function () {
            test('creates a workload with a workload volume - non-existing pvc', async function () {
                const requestedWorkloadName = objectGenerator.generateNameByPrefix('wl');
                const claimName = requestedWorkloadName + '-claim';
                const workloadSpec: InferenceSpecSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    servingPort: {
                        container: 8080
                    },
                    storage: {
                        pvc: [
                            {
                                name: 'vol1',
                                claimName: claimName,
                                path: '/data',
                                existingPvc: false,
                                claimInfo: {
                                    // didn't specify storage class, so it will the default one
                                    size: '1Gi',
                                    accessModes: { readWriteOnce: true }
                                }
                            }
                        ]
                    }
                };

                // create the workload and wait for it to be in phase Running
                const [workloadId, workloadName] = await inferenceUtils.createInference(
                    requestedWorkloadName,
                    workloadSpec,
                    String(project.id)
                );

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                logger.info(`workload ${workloadName} is in phase Running ${workloadId}`);

                // check that the pvc was created
                await waitUtils.getPvc(claimName, project.status.namespace);
                logger.info(`pvc ${claimName} was created`);

                // delete the workload and check that the pvc is deleted
                await workloadsCrudUtils.deleteInference(workloadId);
                await waitUtils.waitForPvc(claimName, project.status.namespace, 120, false);
                logger.info(`pvc ${claimName} was deleted`);
            });
        });

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('creates a workload with emptydir', async function () {
                const requestedWorkloadName = objectGenerator.generateNameByPrefix('wl');
                const workloadSpec: InferenceSpecSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    command: `/bin/sh -c "ls /path && ${HELLO_WORLD_SERVER_COMMAND}"`,
                    servingPort: {
                        container: 8080
                    },
                    storage: {
                        emptyDirVolume: [
                            {
                                path: '/path',
                                medium: 'Memory',
                                sizeLimit: '1Gi'
                            }
                        ]
                    }
                };

                // create the workload
                const [workloadId] = await inferenceUtils.createInference(
                    requestedWorkloadName,
                    workloadSpec,
                    String(project.id)
                );

                // wait for the workload to be in phase Running
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
            });
        });
    }
);
