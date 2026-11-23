/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi } from '../../../../../../modules/api';
import {
    AssetKind,
    AssetSyncStatusEnum,
    PVCAsset,
    PVCCreationRequest,
    Scope,
    WorkspaceCreationRequestV2
} from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    assetsApiV2Utils,
    computeResourceUtils,
    dataSourcesUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    // tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { description, tag } from 'allure-js-commons';
import { test } from '@playwright/test';

const pvcScopes: Scope[] = [Scope.Project, Scope.Cluster];

pvcScopes.forEach((pvcScope) => {
    let project: Project;

    function getPvcAssetPayload(name: string, scope: Scope): PVCCreationRequest {
        return {
            meta: {
                name,
                scope,
                clusterId: scope === Scope.Cluster ? flowRootState.cluster.uuid : undefined,
                departmentId: scope === Scope.Department ? project.departmentId + '' : undefined,
                projectId: scope === Scope.Project ? project.id : undefined
            },
            spec: {
                path: '/home',
                readOnly: false,
                claimName: name,
                claimInfo: {
                    size: '1M',
                    storageClass: 'standard',
                    accessModes: {
                        readWriteOnce: true
                    }
                }
            }
        };
    }

    async function getWorkloadCreationPayloadWithPvc(
        workloadName: string,
        pvcId: string
    ): Promise<WorkspaceCreationRequestV2> {
        return {
            meta: {
                name: workloadName,
                useGivenNameAsPrefix: true,
                projectId: project.id.toString(),
                clusterId: flowRootState.cluster.uuid
            },
            assets: {
                environment: await environmentUtils.getId('tensorboard'),
                compute: await computeResourceUtils.getId('cpu-only'),
                datasources: [
                    {
                        id: pvcId,
                        kind: AssetKind.Pvc
                    }
                ]
            }
        };
    }

    test.describe(`2 Workloads using the same ${pvcScope} scoped PVC | ${tags.teams.viper}`, () => {
        test.beforeAll('creating project', async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApi.pVCApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (e) {
                logger.error('Failed to cleanup project', e);
            }
        });

        function addInfo(): void {
            tag(`${tags.teams.viper}`);
            description(
                'Flow:\n' +
                    `1. Submit two workload using the same ${pvcScope} scope PVC.\n` +
                    '2. Check:\n' +
                    '* workloads phase is "running"\n'
            );
        }

        const pvcName = objectGenerator.generateNameByPrefix(`${pvcScope}-pvc`);
        let pvcId: string;
        const workloadName1 = 'workload1';
        let workloadId1: string;
        const workloadName2 = 'workload2';
        let workloadId2: string;

        test('Create 2 workloads using the same PVC', async function () {
            await test.step('Create PVC', async () => {
                addInfo();
                const pvc = await dataSourcesUtils.createPvcAsset(getPvcAssetPayload(pvcName, pvcScope));
                pvcId = pvc.meta.id;
                await waitUtils.waitForPvc(pvcName, project.namespace);

                await waitUtils.waitForCondition(
                    async () => {
                        const readyPvc: PVCAsset = (
                            await assetsApi.pVCApi.getPvcAssetById(
                                pvcId,
                                undefined,
                                undefined,
                                undefined,
                                true
                            )
                        ).body;
                        return readyPvc.status.status === AssetSyncStatusEnum.NoIssuesFound;
                    },
                    'waiting for usedBy field to be undefined',
                    true
                );
                logger.info(`PVC ${pvcName} created successfully`);
            });

            await test.step('Create workloads', async () => {
                addInfo();
                let payload = await getWorkloadCreationPayloadWithPvc(workloadName1, pvcId);
                workloadId1 = (await assetsApiV2Utils.createWorkspace(payload)).body.meta.workloadId;

                payload = await getWorkloadCreationPayloadWithPvc(workloadName2, pvcId);
                workloadId2 = (await assetsApiV2Utils.createWorkspace(payload)).body.meta.workloadId;

                logger.info(`Workloads ${workloadName1} and ${workloadName2} created successfully`);
            });

            await test.step('Validate both workspaces are running', async () => {
                addInfo();
                await waitUtils.waitForCondition(
                    async () => {
                        const workloads = (await workloadsUtilsV2.getAllWorkloads()).body.workloads;
                        const workloadsStatuses = workloads
                            .filter((w) => [workloadName1, workloadName2].includes(w.name))
                            .map((workload) => workload.phase);
                        return workloadsStatuses.every((status) => status === Phase.Running);
                    },
                    'wait for workloads to be running',
                    true,
                    90,
                    2
                );
                logger.info('Both workloads are running');
            });

            await test.step('Delete workloads', async () => {
                addInfo();
                await Promise.all([
                    workloadsCrudUtils.deleteWorkspace(workloadId1),
                    workloadsCrudUtils.deleteWorkspace(workloadId2)
                ]);
                logger.info('Both workloads are deleted');
            });
        });
    });
});
