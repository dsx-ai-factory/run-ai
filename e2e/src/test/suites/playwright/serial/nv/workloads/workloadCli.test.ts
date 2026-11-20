/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    nodeUtils,
    tags,
    waitUtils,
    settings,
    gpuTypes,
    flowRootState
} from '../../../../../../modules/helpers';
import { V1Pod } from '@kubernetes/client-node';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    cliV2TrainingType
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    supportedFrom,
    SupportedVersions,
    supportedPlatforms,
    shouldRunTests,
    Platform
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { SoftAssert } from '../../../../../../modules/helpers/SoftAssert';

test.describe(`Workload CLI Tests, tags: ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let maxGpuCount: number;

        test.beforeAll(async function () {
            maxGpuCount = await nodeUtils.getMaxGpuCountPerNode();
            projectName = objectGenerator.generateShortNameByPrefix('project-cli-for-workload');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 16
            });
            runaiCliV2.loginAsUser(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            runaiCliV2.projectSet(project.name);
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

        function checkInjectionOnGCP(podsYaml: V1Pod[]): void {
            for (const podYaml of podsYaml) {
                expect(Object.keys(podYaml.metadata.annotations).includes('container.tcpx-daemon')).toBe(
                    false
                );
                for (const container of podYaml.spec.containers) {
                    let ncclNetPluginIsNone = false;
                    let ncclTunerPluginIsNone = false;
                    const env = container.env || [];
                    for (const envVar of env) {
                        if (envVar.name === 'NCCL_NET_PLUGIN' && envVar.value === 'NONE') {
                            ncclNetPluginIsNone = true;
                        }
                        if (envVar.name === 'NCCL_TUNER_PLUGIN' && envVar.value === 'NONE') {
                            ncclTunerPluginIsNone = true;
                        }

                        if (ncclNetPluginIsNone && ncclTunerPluginIsNone) {
                            break;
                        }
                    }
                    expect(ncclNetPluginIsNone).toBe(true);
                    expect(ncclTunerPluginIsNone).toBe(true);
                }
            }
        }

        function checkInjectionOnAWS(podsYaml: V1Pod[]): void {
            for (const podYaml of podsYaml) {
                for (const container of podYaml.spec.containers) {
                    let fiEfaUseDeviceRdmaExist = false;
                    const env = container.env || [];
                    for (const envVar of env) {
                        if (envVar.name === 'FI_EFA_USE_DEVICE_RDMA') {
                            fiEfaUseDeviceRdmaExist = true;
                            break;
                        }
                    }
                    expect(fiEfaUseDeviceRdmaExist).toBe(false);
                }
            }
        }

        supportedPlatforms([Platform.GCP, Platform.EKS])(function () {
            shouldRunTests(settings.gpuType === gpuTypes.real)(function () {
                test.describe('Test GPU workloads', function () {
                    test('T3738054: create single node GPU job of training type', async function () {
                        if (maxGpuCount === 0) {
                            test.skip();
                        }
                        const softAssert = new SoftAssert();

                        for (const gpuCnt of [1, maxGpuCount]) {
                            const gpuSubmitFlag = {
                                [SubmitFlagsMerged.image]: config.quickStartImage.training,
                                [SubmitFlagsMerged.gpuRequest]: gpuCnt
                            } as Partial<typeof SubmitFlagsMerged>;
                            const gpuTrainingJob = objectGenerator.generateShortNameByPrefix(
                                `training-job-${gpuCnt}-gpu`,
                                5
                            );

                            await test.step(`Create ${gpuCnt} GPU training workload`, async () => {
                                runaiCliV2.submit(
                                    gpuSubmitFlag,
                                    gpuTrainingJob,
                                    cliV2WorkloadType.training,
                                    cliV2TrainingType.standard
                                );
                                await waitUtils.waitForBackendWorkloadPhase(
                                    gpuTrainingJob,
                                    project.id,
                                    Phase.Running,
                                    180
                                );
                            });

                            const gpuPodsYaml = await kubeCli.listPodsByNamepsace(namespace);

                            try {
                                supportedPlatforms([Platform.GCP])(function () {
                                    test.step('Check injections on GCP', () => {
                                        checkInjectionOnGCP(gpuPodsYaml);
                                    });
                                });

                                supportedPlatforms([Platform.EKS])(function () {
                                    test.step('Check injections on AWS', () => {
                                        checkInjectionOnAWS(gpuPodsYaml);
                                    });
                                });
                            } catch (error) {
                                softAssert.assert(
                                    false,
                                    `failed to verify injection for ${gpuCnt} GPU training workload \n${error}`
                                );
                            }
                            test.step(`Delete ${gpuCnt} GPU training workload`, () => {
                                runaiCliV2.delete(undefined, gpuTrainingJob, cliV2WorkloadType.training);
                            });
                        }

                        softAssert.throwErrorIfFailed();
                    });

                    test('T3738056: create single node GPU job of interactive type', async function () {
                        if (maxGpuCount === 0) {
                            test.skip();
                        }
                        const softAssert = new SoftAssert();

                        for (const gpuCnt of [1, maxGpuCount]) {
                            const gpuSubmitFlag = {
                                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                                [SubmitFlagsMerged.gpuRequest]: gpuCnt,
                                [SubmitFlagsMerged.command]: '-- sleep 1200'
                            } as Partial<typeof SubmitFlagsMerged>;

                            const gpuWsJob = objectGenerator.generateShortNameByPrefix(
                                `workspace-job-${gpuCnt}-gpu`,
                                5
                            );

                            await test.step(`Create ${gpuCnt} GPU workspace workload`, async () => {
                                runaiCliV2.submit(gpuSubmitFlag, gpuWsJob, cliV2WorkloadType.workspace);
                                await waitUtils.waitForBackendWorkloadPhase(
                                    gpuWsJob,
                                    project.id,
                                    Phase.Running,
                                    180
                                );
                            });

                            const gpuPodsYaml = await kubeCli.listPodsByNamepsace(namespace);

                            try {
                                supportedPlatforms([Platform.GCP])(function () {
                                    test.step('Check injections on GCP', () => {
                                        checkInjectionOnGCP(gpuPodsYaml);
                                    });
                                });

                                supportedPlatforms([Platform.EKS])(function () {
                                    test.step('Check injections on AWS', () => {
                                        checkInjectionOnAWS(gpuPodsYaml);
                                    });
                                });
                            } catch (error) {
                                softAssert.assert(
                                    false,
                                    `failed to verify injection for ${gpuCnt} GPU workspace workload \n${error}`
                                );
                            }

                            test.step(`Delete ${gpuCnt} GPU workspace workload`, () => {
                                runaiCliV2.delete(undefined, gpuWsJob, cliV2WorkloadType.workspace);
                            });
                        }

                        softAssert.throwErrorIfFailed();
                    });

                    test('T3738057: create distribute GPU job of training type with 1 GPU per pod', async function () {
                        if (maxGpuCount === 0) {
                            test.skip();
                        }
                        const gpuSubmitFlag = {
                            [SubmitFlagsMerged.image]: config.quickStartImage.distributedMpi,
                            [SubmitFlagsMerged.gpuRequest]: 1,
                            [SubmitFlagsMerged.workers]: 2,
                            [SubmitFlagsMerged.envVar]: 'RUNAI_SLEEP_SECS=60'
                        } as Partial<typeof SubmitFlagsMerged>;

                        const gpuTrainingJob = objectGenerator.generateShortNameByPrefix(
                            'distributed-mpi-training-job-1-gpu',
                            5
                        );

                        await test.step('Create distributed MPI training workload', async () => {
                            runaiCliV2.submit(
                                gpuSubmitFlag,
                                gpuTrainingJob,
                                cliV2WorkloadType.training,
                                cliV2TrainingType.mpi
                            );
                            await waitUtils.waitForBackendWorkloadPhase(
                                gpuTrainingJob,
                                project.id,
                                Phase.Running,
                                240
                            );
                        });

                        const gpuPodsYaml = await kubeCli.listPodsByNamepsace(namespace);

                        supportedPlatforms([Platform.GCP])(function () {
                            test.step('Check injections on GCP', () => {
                                checkInjectionOnGCP(gpuPodsYaml);
                            });
                        });
                        supportedPlatforms([Platform.EKS])(function () {
                            test.step('Check injections on AWS', () => {
                                checkInjectionOnAWS(gpuPodsYaml);
                            });
                        });

                        test.step('Delete distributed MPI training workload', () => {
                            runaiCliV2.delete(undefined, gpuTrainingJob, cliV2WorkloadType.mpi);
                        });
                    });
                });
            });
        });
    });
});
