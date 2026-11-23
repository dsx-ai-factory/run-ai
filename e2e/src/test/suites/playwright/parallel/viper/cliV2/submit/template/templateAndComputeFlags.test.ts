/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, TestInfo } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { submitWorkload, validateWorkload, deleteWorkload } from '../submitTestUtils';
import { skipPlaywrightTest } from '../../../../../../../../modules/helpers/utils/SkippedTestUtils';
import {
    computeResourceUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    templateUtils
} from '../../../../../../../../modules/helpers';
import { ComputeCreationRequest, Scope } from '../../../../../../../../modules/api/apis/assetsApi';
import { ScopeType, WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`CLI V2 Submit with Template and Compute Flags, ${tags.teams.viper}`, { tag: [tags.services.assets, tags.services.workloads, tags.runOn.parallelCI] }, () => {
    let project: Project;
    let templateId: string;
    let templateName: string;
    let computeName: string;
    let workloadName: string;

    // Compute asset values (different from template)
    let computeId: string;
    const computeAssetCpuRequest = 0.3;
    const computeAssetGpuDevicesRequest = 0;

    // Template inline compute values (different from asset)
    const templateCpuRequest = 0.1;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('p');

        // Create compute asset with specific values
        computeName = objectGenerator.generateShortNameByPrefix('compute');
        const computeRequest: ComputeCreationRequest = {
            meta: {
                name: computeName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                cpuCoreRequest: computeAssetCpuRequest,
                gpuDevicesRequest: computeAssetGpuDevicesRequest
            }
        };
        const computeAsset = await computeResourceUtils.create(computeRequest);
        computeId = computeAsset.meta.id;

        // Create template with inline compute values (different from asset)
        templateName = objectGenerator.generateNameByPrefix('tmpl');
        const templateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: project.id.toString(),
                templateName: templateName,
                description: 'Template for compute tests'
            },
            {
                image: 'ubuntu',
                command: 'sleep infinity',
                compute: {
                    cpuCoreRequest: templateCpuRequest
                }
            }
        );
        templateId = templateResponse.body.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateShortNameByPrefix('wl');
    });

    test.afterAll(async () => {
        try {
            if (templateId) {
                await templateUtils.deleteWorkloadTemplateV2(templateId, WorkloadType.Training);
            }
            if (computeId) {
                await computeResourceUtils.delete(computeId);
            }
            await operations.deleteProjectAndNamespace(project.id);
        } catch (error) {
            logger.error('Cleanup failed:', error);
        }
    });

    // ==========================================
    // Asset values win (--compute flag, no override)
    // ==========================================

    test('--compute asset cpuCoreRequest overrides template cpuCoreRequest', async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.compute]: computeName,
                [SubmitFlagsMerged.project]: project.name
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    cpuCoreRequest: computeAssetCpuRequest,
                    gpuDevicesRequest: computeAssetGpuDevicesRequest
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    // ==========================================
    // Override flags win
    // ==========================================

    test(`--cpu-core-request + -g flags override --compute asset values, ${tags.runOn.parallelCI}`, async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.compute]: computeName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.cpuCoreRequest]: '0.5',
                [SubmitFlagsMerged.gpuRequest]: computeAssetGpuDevicesRequest.toString()
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    cpuCoreRequest: 0.5,
                    gpuDevicesRequest: computeAssetGpuDevicesRequest
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
