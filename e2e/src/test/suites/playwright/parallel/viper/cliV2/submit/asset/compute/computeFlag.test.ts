/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';

import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    computeResourceUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags
} from '../../../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { ComputeCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';
import { submitWorkload, validateWorkload, deleteWorkload } from '../../submitTestUtils';

test.describe(`CLI V2 Submit with Compute Asset, ${tags.teams.viper}`, () => {
    let project: Project;
    let computeAssetId: string;
    let computeAssetName: string;
    let workloadName: string;
    let computeCreationRequest: ComputeCreationRequest;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-compute');
        computeAssetName = objectGenerator.generateNameByPrefix('compute');

        computeCreationRequest = {
            meta: {
                name: computeAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                cpuCoreRequest: 0.1
            }
        };

        const computeAsset = await computeResourceUtils.create(computeCreationRequest);
        computeAssetId = computeAsset.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl');
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            if (computeAssetId) {
                await computeResourceUtils.delete(computeAssetId);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    for (const wlType of [cliV2WorkloadType.training, cliV2WorkloadType.workspace]) {
        test(`submit ${wlType} workload with compute asset`, async () => {
            test.setTimeout(210 * 1000);

            submitWorkload(
                workloadName,
                {
                    [SubmitFlagsMerged.compute]: computeAssetName,
                    [SubmitFlagsMerged.project]: project.name,
                    [SubmitFlagsMerged.image]: 'ubuntu',
                    [SubmitFlagsMerged.command]: '-- sleep infinity'
                } as Partial<typeof SubmitFlagsMerged>,
                wlType
            );

            await validateWorkload(
                workloadName,
                project.id,
                {
                    fields: { cpuCoreRequest: computeCreationRequest.spec.cpuCoreRequest }
                },
                wlType,
                true
            );

            deleteWorkload(workloadName, project.name, wlType);
        });
    }

    test(`override compute asset with --cpu-core-request flag, ${tags.runOn.parallelCI}`, async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.compute]: computeAssetName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'ubuntu',
                [SubmitFlagsMerged.gpuRequest]: '1',
                [SubmitFlagsMerged.cpuCoreRequest]: '0.2',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: { cpuCoreRequest: 0.2, gpuDevicesRequest: 1 }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
