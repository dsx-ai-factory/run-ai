/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    dataSourcesUtils,
    flowRootState,
    isOCP,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { bashCli, kubeCli } from '../../../../../../modules/cli';
import { workloadsCrudService } from '../../../../../../modules/api';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DistributedSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedSpecSpec';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { TrainingSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/trainingSpecSpec';
import { StatusCodes } from 'http-status-codes';

test.describe(`Submit training with multiple mount paths of the same pvc - sanity, tags: ${tags.resource.datasource} ${tags.teams.viper}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        shouldRunTests(settings.isSelfHosted && !settings.isAws)(function () {
            let project: Project;
            let workloadName: string;
            let trainingSpec: TrainingSpecSpec | DistributedSpecSpec;
            let mountPaths: string[];
            let existingPvc1Name: string;
            let existingPvc2Name: string;

            test.beforeAll(async () => {
                project = await projectResourcesUtils.createProjectAndNamespace('p');
                existingPvc1Name = objectGenerator.generateShortNameByPrefix('pvc1');
                existingPvc2Name = objectGenerator.generateShortNameByPrefix('pvc2');

                await dataSourcesUtils.createPVC(existingPvc1Name, Scope.Project, project);
                await dataSourcesUtils.createPVC(existingPvc2Name, Scope.Project, project);

                await waitUtils.waitForPvc(existingPvc1Name, project.namespace);
                await waitUtils.waitForPvc(existingPvc2Name, project.namespace);

                workloadsCrudService.trainingsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                mountPaths = ['/home/a', '/home/b', '/home/c', '/home/d'];

                trainingSpec = {
                    image: 'ubuntu',
                    command: 'sleep infinity',
                    storage: {
                        pvc: [
                            {
                                claimName: existingPvc1Name,
                                existingPvc: true,
                                path: mountPaths[0],
                                readOnly: true
                            },
                            {
                                claimName: existingPvc1Name,
                                existingPvc: true,
                                path: mountPaths[1],
                                readOnly: true
                            },
                            {
                                claimName: existingPvc2Name,
                                existingPvc: true,
                                path: mountPaths[2],
                                readOnly: true
                            },
                            {
                                name: 'pvc-new',
                                existingPvc: false,
                                path: mountPaths[3],
                                readOnly: false,
                                claimInfo: {
                                    size: '1G'
                                }
                            }
                        ]
                    }
                };
            });

            test.afterAll(async () => {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch {
                    logger.error(
                        `Failed to delete project ${project.name} and namespace ${project.namespace}`
                    );
                }

                const createdPvcs = (await kubeCli.listPVCsByNamespace(project.namespace)).items;
                for (const pvc of createdPvcs) {
                    try {
                        await kubeCli.deletePvc(pvc.metadata.name, project.namespace);
                    } catch (e) {
                        logger.warn(`Failed to delete PVC assets: ${e.message}`);
                    }
                }
            });

            function verifyMountPathsExistInPod(
                podName: string,
                namespace: string,
                mountPaths: string[]
            ): void {
                for (const mountPath of mountPaths) {
                    try {
                        bashCli.exec(`kubectl exec ${podName} -n ${namespace} -- ls ${mountPath}`);
                        logger.info(`ls ${mountPath} succeeded in pod ${podName}`);
                    } catch (error) {
                        throw new Error(
                            `Mount path ${mountPath} does not exist in pod ${podName} or cannot be listed. Error: ${error}`
                        );
                    }
                }
            }

            test('Create training and check mount paths exist in pod', () => {
                shouldRunTests(!isOCP())(async function () {
                    workloadName = objectGenerator.generateShortNameByPrefix('training');

                    const createResult = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        trainingSpec
                    );

                    expect(createResult.response.statusCode).toBe(StatusCodes.ACCEPTED);
                    await waitUtils.waitForPods(createResult.body.name, project.namespace, 1, 60, [
                        'Running'
                    ]);

                    const podName = `${createResult.body.name}-0-0`;
                    verifyMountPathsExistInPod(podName, project.namespace, mountPaths);

                    await workloadsCrudUtils.deleteTraining(createResult.body.workloadId);
                });
            });

            test('Create distributed training and check mount paths exist in pod', () => {
                shouldRunTests(!isOCP())(async function () {
                    workloadName = objectGenerator.generateShortNameByPrefix('distributed');
                    (trainingSpec as DistributedSpecSpec).distributedFramework = DistributedFramework.Mpi;
                    const createResult = await workloadsCrudUtils.createDistributed(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        trainingSpec,
                        true
                    );
                    expect(createResult.response.statusCode).toBe(StatusCodes.ACCEPTED);

                    const pods = await waitUtils.waitForPods(
                        createResult.body.name,
                        project.namespace,
                        2,
                        60,
                        ['Running'],
                        true
                    );
                    verifyMountPathsExistInPod(pods[0].metadata.name, project.namespace, mountPaths);
                    verifyMountPathsExistInPod(pods[1].metadata.name, project.namespace, mountPaths);

                    await workloadsCrudUtils.deleteDistributed(createResult.body.workloadId);
                });
            });
        });
    });
});
