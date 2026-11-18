/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';

import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    config,
    datavolumeUtils,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    projectResourcesUtils,
    storageUtils,
    tags,
    waitUtils,
    orgUnitUtils
} from '../../../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { submitWorkload, validateWorkload, deleteWorkload } from '../../submitTestUtils';

async function bindPvcForDataVolume(pvcName: string, namespace: string): Promise<void> {
    const podName = objectGenerator.generateNameByPrefix('pvc-consumer');
    await kubeCli.CoreV1Api.createNamespacedPod({
        namespace,
        body: {
            metadata: { name: podName, namespace },
            spec: {
                restartPolicy: 'Never',
                containers: [
                    {
                        name: 'consumer',
                        image: config.quickStartImage.sleepInfinity,
                        volumeMounts: [{ name: 'source-pvc', mountPath: '/mnt/source' }]
                    }
                ],
                volumes: [
                    {
                        name: 'source-pvc',
                        persistentVolumeClaim: { claimName: pvcName }
                    }
                ]
            }
        }
    });

    try {
        await storageUtils.waitForPvcToBeBound(pvcName, namespace);
    } finally {
        try {
            await kubeCli.deletePod(podName, namespace);
        } catch (error) {
            logger.warn(`Failed to delete PVC consumer pod ${namespace}/${podName}: ${error}`);
        }
    }
}

test.describe(`CLI V2 Submit with DataVolume Flag - Positive Tests, ${tags.teams.viper},`, { tag: [tags.services.datavolume, tags.runOn.parallelCI] }, () => {
    let project1: Project;
    let workloadName: string;

    let pvcProjectName: string;
    let pvcProjectId: string;

    let dvProjectId: string;
    let dvProjectName: string;

    test.beforeAll(async () => {
        // Native controllers receive the DataVolume ID and resolve its PVC asynchronously in the
        // cluster. Redirected submission renders the PVC claim name in the control plane, so it
        // requires the DataVolume to be Ready before the workload can be composed.
        const redirectedSubmissionEnabled = (
            await Promise.all([
                isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training),
                isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)
            ])
        ).some(Boolean);

        project1 = await projectResourcesUtils.createProjectAndNamespace('cli-dv-pos');

        pvcProjectName = objectGenerator.generateNameByPrefix('pvc-proj');

        const pvcProjectRequest: PVCCreationRequest = {
            meta: {
                name: pvcProjectName,
                scope: Scope.Project,
                projectId: project1.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: '/mnt/dv-project',
                claimName: pvcProjectName,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: { readWriteOnce: true },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        const pvcProjectAsset = await dataSourcesUtils.createPvcAsset(pvcProjectRequest);
        pvcProjectId = pvcProjectAsset.meta.id;
        await waitUtils.waitForPvc(pvcProjectName, project1.namespace);
        if (redirectedSubmissionEnabled) {
            // Test storage classes may use WaitForFirstConsumer. Mount the origin PVC once so it
            // binds and the DataVolume controller can publish the resolved claim name in status.
            await bindPvcForDataVolume(pvcProjectName, project1.namespace);
        }

        dvProjectName = objectGenerator.generateNameByPrefix('dv-proj');
        const dvProject = await datavolumeUtils.createDatavolume({
            name: dvProjectName,
            description: 'Project-scoped datavolume for CLI test',
            originPvcName: pvcProjectName,
            projectId: String(project1.id)
        });
        dvProjectId = dvProject.id;

        if (redirectedSubmissionEnabled) {
            await datavolumeUtils.waitForApiDatavolumeToBeReady(dvProjectId);
        } else {
            await datavolumeUtils.waitForApiDatavolumeToExist(dvProjectId);
        }
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-dv');
    });

    test.afterAll(async () => {
        try {
            if (dvProjectId) {
                await datavolumeUtils.deleteDatavolume(dvProjectId);
                await datavolumeUtils.waitForApiDatavolumeToBeDeleted(dvProjectId);
            }
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('Datavolume cleanup failed:', error);
            }
        }

        try {
            if (pvcProjectId) await dataSourcesUtils.deletePvcAsset(pvcProjectId);
        } catch (error: any) {
            if (error?.statusCode !== 404) {
                console.error('PVC cleanup failed:', error);
            }
        }

        try {
            await orgUnitUtils.deleteProjectAndNamespace(project1.id.toString(), project1.name);
        } catch (error) {
            console.error('Project cleanup failed:', error);
        }
    });

    for (const wlType of [cliV2WorkloadType.training, cliV2WorkloadType.workspace]) {
        test(`submit ${wlType} workload with project-scoped datavolume`, async () => {
            submitWorkload(
                workloadName,
                {
                    [SubmitFlagsMerged.datavolume]: [`name=${dvProjectName},mountpath=/mnt/data`],
                    [SubmitFlagsMerged.project]: project1.name,
                    [SubmitFlagsMerged.image]: 'ubuntu',
                    [SubmitFlagsMerged.command]: '-- sleep infinity'
                } as Partial<typeof SubmitFlagsMerged>,
                wlType
            );

            await validateWorkload(
                workloadName,
                project1.id,
                {
                    fields: { image: 'ubuntu' },
                    storage: {
                        dataVolume: [{ id: dvProjectId }]
                    }
                },
                wlType
            );

            deleteWorkload(workloadName, project1.name, wlType);
        });
    }

    test('submit training workload with datavolume and pvc datasource combined', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.datavolume]: [`name=${dvProjectName},mountpath=/mnt/data`],
                [SubmitFlagsMerged.datasource]: [`type=pvc,name=${pvcProjectName}`],
                [SubmitFlagsMerged.project]: project1.name,
                [SubmitFlagsMerged.image]: 'ubuntu',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project1.id,
            {
                fields: { image: 'ubuntu' },
                storage: {
                    dataVolume: [{ id: dvProjectId }],
                    pvc: [
                        {
                            claimName: pvcProjectName,
                            path: '/mnt/dv-project',
                            existingPvc: true,
                            readOnly: false
                        }
                    ]
                }
            },
            cliV2WorkloadType.training
        );

        deleteWorkload(workloadName, project1.name, cliV2WorkloadType.training);
    });
});
