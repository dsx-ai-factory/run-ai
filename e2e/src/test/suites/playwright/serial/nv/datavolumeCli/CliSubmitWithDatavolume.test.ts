/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    datavolumeUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    storageUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    cliV2TrainingType,
    DeleteFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { submitWorkload, validateWorkload } from '../../../parallel/viper/cliV2/submit/submitTestUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`CLI Submit with Datavolume - Positive Tests, tags: ${tags.resource.datavolume} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        let project: Project;
        let projectName: string;
        let pvcAssetName: string;
        let pvcAssetId: string;
        const pvcAssetIdsToClean: string[] = [];
        let dvId: string;
        let dvName: string;

        type WorkloadToClean = {
            name: string;
            projectName: string;
            workloadType: cliV2WorkloadType;
            isDistributed?: boolean;
        };
        let workloadsToClean: WorkloadToClean[] = [];
        let datavolumesToClean: string[] = [];

        test.afterEach(async function () {
            await test.step('Cleanup workloads and extra datavolumes from test', async function () {
                for (const w of workloadsToClean) {
                    try {
                        runaiCliV2.delete(
                            { [DeleteFlagsMerged.project]: w.projectName } as Partial<
                                typeof DeleteFlagsMerged
                            >,
                            w.name,
                            w.workloadType,
                            true,
                            w.isDistributed ?? false
                        );
                    } catch (e) {
                        logger.warn(`Cleanup workload ${w.name} failed: ${e}`);
                    }
                }
                for (const id of datavolumesToClean) {
                    try {
                        await datavolumeUtils.deleteDatavolume(id);
                        await datavolumeUtils.waitForApiDatavolumeToBeDeleted(id);
                    } catch (e) {
                        logger.warn(`Cleanup datavolume ${id} failed: ${e}`);
                    }
                }
                workloadsToClean = [];
                datavolumesToClean = [];
            });
        });

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-dv-submit', 5);
            const namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 4
                });
                logger.debug(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Create PVC asset and datavolume', async function () {
                pvcAssetName = objectGenerator.generateShortNameByPrefix('pvc-dv', 5);
                const pvcRequest: PVCCreationRequest = {
                    meta: {
                        name: pvcAssetName,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        path: '/mnt/dv',
                        claimName: pvcAssetName,
                        existingPvc: false,
                        claimInfo: {
                            size: '1Gi',
                            accessModes: { readWriteOnce: true },
                            storageClass: 'standard'
                        },
                        readOnly: false
                    }
                };
                const pvcAsset = await dataSourcesUtils.createPvcAsset(pvcRequest);
                pvcAssetId = pvcAsset.meta.id;
                pvcAssetIdsToClean.push(pvcAssetId);
                await waitUtils.waitForPvc(pvcAssetName, project.namespace);
                await storageUtils.ensurePvcBound(pvcAssetName, project.namespace);

                dvName = objectGenerator.generateShortNameByPrefix('dv-submit', 5);
                const dv = await datavolumeUtils.createDatavolume({
                    name: dvName,
                    description: 'Datavolume for submit tests',
                    originPvcName: pvcAssetName,
                    projectId: String(project.id)
                });
                dvId = dv.id;

                await datavolumeUtils.waitForApiDatavolumeToBeReady(dvId);
                logger.debug(`Created datavolume: ${dvName} (ID: ${dvId})`);
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup datavolume and PVC', async function () {
                try {
                    if (dvId) {
                        await datavolumeUtils.deleteDatavolume(dvId);
                        await datavolumeUtils.waitForApiDatavolumeToBeDeleted(dvId);
                    }
                } catch (e: any) {
                    if (e?.statusCode !== 404) logger.warn(`Datavolume cleanup: ${e}`);
                }
                for (const id of pvcAssetIdsToClean) {
                    try {
                        await dataSourcesUtils.deletePvcAsset(id);
                    } catch (e: any) {
                        if (e?.statusCode !== 404) logger.warn(`PVC asset cleanup ${id}: ${e}`);
                    }
                }
            });
            await test.step('Cleanup project', async function () {
                if (project?.id) await operations.deleteProjectAndNamespace(project.id);
            });
        });

        /** Config for single-datavolume submit tests. deleteWorkloadType is the CLI noun used for delete (e.g. pyTorch for "runai pytorch delete"). */
        type SingleDvCaseConfig = {
            caseId: string;
            title: string;
            namePrefix: string;
            mountPath: string;
            image?: string;
            validateWorkloadType: cliV2WorkloadType;
            deleteWorkloadType: cliV2WorkloadType;
            deleteIsDistributed?: boolean;
            useSubmitWorkload: boolean;
            trainingType?: cliV2TrainingType;
            isDistributed?: boolean;
            /** Omit command flag (e.g. inference distributed does not support -c) */
            omitCommand?: boolean;
            extraFlags?: Partial<typeof SubmitFlagsMerged>;
        };

        async function runSingleDatavolumeCase(config: SingleDvCaseConfig): Promise<void> {
            const workloadName = objectGenerator.generateShortNameByPrefix(config.namePrefix, 5);
            const image = config.image ?? 'ubuntu:22.04';
            const baseFlags = {
                [SubmitFlagsMerged.datavolume]: [`name=${dvName},mountpath=${config.mountPath}`],
                [SubmitFlagsMerged.project]: projectName,
                [SubmitFlagsMerged.image]: image,
                ...(config.omitCommand ? {} : { [SubmitFlagsMerged.command]: '-- sleep infinity' }),
                ...(config.extraFlags as Record<string, string>)
            } as Partial<typeof SubmitFlagsMerged>;

            if (config.useSubmitWorkload) {
                submitWorkload(workloadName, baseFlags, config.validateWorkloadType);
            } else {
                const result = runaiCliV2.submit(
                    baseFlags,
                    workloadName,
                    config.validateWorkloadType,
                    config.trainingType,
                    config.isDistributed
                );
                if (
                    config.validateWorkloadType === cliV2WorkloadType.training ||
                    (config.validateWorkloadType === cliV2WorkloadType.inference && config.isDistributed)
                ) {
                    expect(result).toMatch(/Creating (training|inference|distributed) /);
                    expect(result).toContain(workloadName);
                } else {
                    expect(result).toContain(`Creating ${config.validateWorkloadType} ${workloadName}`);
                }
            }

            const workloadId = await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
                workloadName,
                String(project.id)
            );
            await workloadsUtilsV2.waitForWorkloadImages(workloadId);

            const isDistributedInference =
                config.validateWorkloadType === cliV2WorkloadType.inference && config.isDistributed;
            await validateWorkload(
                workloadName,
                project.id,
                {
                    fields: { image },
                    ...(isDistributedInference
                        ? {}
                        : {
                              storage: { dataVolume: [{ id: dvId, mountPath: config.mountPath }] }
                          })
                },
                config.deleteWorkloadType
            );

            workloadsToClean.push({
                name: workloadName,
                projectName,
                workloadType: config.deleteWorkloadType,
                isDistributed: config.deleteIsDistributed
            });
        }

        const SINGLE_DV_CASES: SingleDvCaseConfig[] = [
            {
                caseId: 'T5750417',
                title: 'Submit workspace with single datavolume',
                namePrefix: 'ws-dv',
                mountPath: '/mnt/data',
                validateWorkloadType: cliV2WorkloadType.workspace,
                deleteWorkloadType: cliV2WorkloadType.workspace,
                useSubmitWorkload: true
            },
            {
                caseId: 'T5750418',
                title: 'Submit training pytorch with single datavolume',
                namePrefix: 'pt-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.pyTorch,
                useSubmitWorkload: false,
                trainingType: cliV2TrainingType.pyTorch,
                extraFlags: { [SubmitFlagsMerged.gpuRequest]: '1' } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750419',
                title: 'Submit inference standard with single datavolume',
                namePrefix: 'inf-dv',
                mountPath: '/mnt/model',
                validateWorkloadType: cliV2WorkloadType.inference,
                deleteWorkloadType: cliV2WorkloadType.inference,
                useSubmitWorkload: false,
                extraFlags: {
                    [SubmitFlagsMerged.gpuRequest]: '1',
                    [SubmitFlagsMerged.servingPort]: '8080'
                } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750420',
                title: 'Submit training tensorflow with datavolume',
                namePrefix: 'tf-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.tensorFlow,
                useSubmitWorkload: false,
                trainingType: cliV2TrainingType.tensorFlow,
                extraFlags: { [SubmitFlagsMerged.gpuRequest]: '1' } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750422',
                title: 'Submit training standard with single datavolume',
                namePrefix: 'std-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.training,
                useSubmitWorkload: true,
                extraFlags: { [SubmitFlagsMerged.gpuRequest]: '1' } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750421',
                title: 'Submit inference distributed with single datavolume',
                namePrefix: 'inf-dist-dv',
                mountPath: '/mnt/model',
                validateWorkloadType: cliV2WorkloadType.inference,
                deleteWorkloadType: cliV2WorkloadType.inference,
                deleteIsDistributed: true,
                useSubmitWorkload: false,
                isDistributed: true,
                omitCommand: true,
                extraFlags: {
                    [SubmitFlagsMerged.gpuRequest]: '1',
                    [SubmitFlagsMerged.servingPort]: '8080'
                } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750423',
                title: 'Submit training mpi with single datavolume',
                namePrefix: 'mpi-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.mpi,
                useSubmitWorkload: false,
                trainingType: cliV2TrainingType.mpi,
                extraFlags: {
                    [SubmitFlagsMerged.gpuRequest]: '1',
                    [SubmitFlagsMerged.workers]: '2'
                } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750424',
                title: 'Submit training jax with single datavolume',
                namePrefix: 'jax-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.jax,
                useSubmitWorkload: false,
                trainingType: cliV2TrainingType.jax,
                extraFlags: { [SubmitFlagsMerged.gpuRequest]: '1' } as Partial<typeof SubmitFlagsMerged>
            },
            {
                caseId: 'T5750425',
                title: 'Submit training xgboost with single datavolume',
                namePrefix: 'xgb-dv',
                mountPath: '/data',
                validateWorkloadType: cliV2WorkloadType.training,
                deleteWorkloadType: cliV2WorkloadType.xgBoost,
                useSubmitWorkload: false,
                trainingType: cliV2TrainingType.xgBoost,
                extraFlags: { [SubmitFlagsMerged.workers]: '2' } as Partial<typeof SubmitFlagsMerged>
            }
        ];

        for (const config of SINGLE_DV_CASES) {
            test(`${config.caseId}: ${config.title}`, async function () {
                await runSingleDatavolumeCase(config);
            });
        }

        test('T5750426: Submit workload with multiple datavolumes', async function () {
            let dv2Name: string;
            let dv2Id: string;
            const workloadName = objectGenerator.generateShortNameByPrefix('ws-multi-dv', 5);

            await test.step('Create second PVC and datavolume, then submit workload', async function () {
                const pvc2Name = objectGenerator.generateShortNameByPrefix('pvc-dv2', 5);
                const pvc2Request: PVCCreationRequest = {
                    meta: {
                        name: pvc2Name,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        path: '/mnt/dv2',
                        claimName: pvc2Name,
                        existingPvc: false,
                        claimInfo: {
                            size: '1Gi',
                            accessModes: { readWriteOnce: true },
                            storageClass: 'standard'
                        },
                        readOnly: false
                    }
                };
                const pvc2Asset = await dataSourcesUtils.createPvcAsset(pvc2Request);
                pvcAssetIdsToClean.push(pvc2Asset.meta.id);
                await waitUtils.waitForPvc(pvc2Name, project.namespace);
                await storageUtils.ensurePvcBound(pvc2Name, project.namespace);

                dv2Name = objectGenerator.generateShortNameByPrefix('dv-multi2', 5);
                const dv2 = await datavolumeUtils.createDatavolume({
                    name: dv2Name,
                    description: 'Second DV for multi mount',
                    originPvcName: pvc2Name,
                    projectId: String(project.id)
                });
                await datavolumeUtils.waitForApiDatavolumeToBeReady(dv2.id);
                dv2Id = dv2.id;
                datavolumesToClean.push(dv2.id);

                submitWorkload(
                    workloadName,
                    {
                        [SubmitFlagsMerged.datavolume]: [
                            `name=${dvName},mountpath=/mnt/vol1`,
                            `name=${dv2Name},mountpath=/mnt/vol2`
                        ],
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.image]: 'ubuntu:22.04',
                        [SubmitFlagsMerged.command]: '-- sleep infinity'
                    } as Partial<typeof SubmitFlagsMerged>,
                    cliV2WorkloadType.workspace
                );
            });

            await test.step('Validate workload with both datavolumes attached', async function () {
                await validateWorkload(
                    workloadName,
                    project.id,
                    {
                        fields: { image: 'ubuntu:22.04' },
                        storage: {
                            dataVolume: [
                                { id: dvId, mountPath: '/mnt/vol1' },
                                { id: dv2Id, mountPath: '/mnt/vol2' }
                            ]
                        }
                    },
                    cliV2WorkloadType.workspace
                );
            });

            await test.step('Track workload for cleanup', function () {
                workloadsToClean.push({
                    name: workloadName,
                    projectName,
                    workloadType: cliV2WorkloadType.workspace
                });
            });
        });
    });
});
