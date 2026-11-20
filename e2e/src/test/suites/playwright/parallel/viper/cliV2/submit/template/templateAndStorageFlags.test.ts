/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    templateUtils,
    waitUtils
} from '../../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { PVCCreationRequest, Scope } from '../../../../../../../../modules/api/apis/assetsApi';
import { ScopeType, WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { submitWorkload, validateWorkload, deleteWorkload } from '../submitTestUtils';

test.describe(`CLI V2 Submit with Template and Storage Flags, ${tags.teams.viper}`, { tag: [tags.services.assets, tags.services.workloads, tags.runOn.parallelCI] }, () => {
    let project: Project;
    let templateId: string;
    let templateName: string;
    let workloadName: string;

    // PVC datasources for --datasource flag
    let pvc1Id: string;
    let pvc1Name: string;
    const pvc1MountPath = '/data/pvc1';

    let pvc2Id: string;
    let pvc2Name: string;
    const pvc2MountPath = '/data/pvc2';

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('p');

        // Create first PVC datasource
        pvc1Name = objectGenerator.generateNameByPrefix('pvc1');
        const pvc1Request: PVCCreationRequest = {
            meta: {
                name: pvc1Name,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: pvc1MountPath,
                claimName: pvc1Name,
                existingPvc: false,
                claimInfo: {
                    size: '1Gi',
                    accessModes: { readWriteOnce: true },
                    storageClass: 'standard'
                },
                readOnly: false
            }
        };
        const pvc1Asset = await dataSourcesUtils.createPvcAsset(pvc1Request);
        pvc1Id = pvc1Asset.meta.id;
        await waitUtils.waitForPvc(pvc1Name, project.namespace);

        // Create second PVC datasource
        pvc2Name = objectGenerator.generateNameByPrefix('pvc2');
        const pvc2Request: PVCCreationRequest = {
            meta: {
                name: pvc2Name,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                path: pvc2MountPath,
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
        pvc2Id = pvc2Asset.meta.id;
        await waitUtils.waitForPvc(pvc2Name, project.namespace);

        // Create template with no datasource (datasources added via --datasource flag)
        templateName = objectGenerator.generateNameByPrefix('tmpl');
        const templateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: project.id.toString(),
                templateName: templateName,
                description: 'Template for storage tests'
            },
            {
                image: 'ubuntu',
                command: 'sleep infinity'
            }
        );
        templateId = templateResponse.body.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl');
    });

    test.afterAll(async () => {
        try {
            if (templateId) {
                await templateUtils.deleteWorkloadTemplateV2(templateId, WorkloadType.Training);
            }
            if (pvc2Id) {
                await dataSourcesUtils.deletePvcAsset(pvc2Id);
            }
            if (pvc1Id) {
                await dataSourcesUtils.deletePvcAsset(pvc1Id);
            }
            await operations.deleteProjectAndNamespace(project.id);
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    // ==========================================
    // Add datasource via --datasource flag
    // ==========================================

    test('--datasource flag adds single datasource to template', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.datasource]: [`type=pvc,name=${pvc1Name}`]
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                storage: {
                    pvc: [
                        {
                            claimName: pvc1Name,
                            path: pvc1MountPath,
                            existingPvc: true,
                            readOnly: false
                        }
                    ]
                }
            },
            cliV2WorkloadType.training
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test(`--datasource flag adds multiple datasources to template, ${tags.runOn.parallelCI}`, async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.datasource]: [`type=pvc,name=${pvc1Name}`, `type=pvc,name=${pvc2Name}`]
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                storage: {
                    pvc: [
                        {
                            claimName: pvc1Name,
                            path: pvc1MountPath,
                            existingPvc: true,
                            readOnly: false
                        },
                        {
                            claimName: pvc2Name,
                            path: pvc2MountPath,
                            existingPvc: true,
                            readOnly: false
                        }
                    ]
                }
            },
            cliV2WorkloadType.training
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
