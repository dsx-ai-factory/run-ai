/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    environmentUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    workloadsUtilsV2
} from '../../../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { EnvironmentCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';
import { submitWorkload, validateWorkload, deleteWorkload } from '../../submitTestUtils';

test.describe(`CLI V2 Submit with Environment Asset, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, () => {
    let project: Project;
    let environmentAssetId: string;
    let environmentAssetName: string;
    let workloadName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-env-asset');
        environmentAssetName = objectGenerator.generateShortNameByPrefix('env');

        const envCreationRequest: EnvironmentCreationRequest = {
            meta: {
                name: environmentAssetName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid,
                workloadSupportedTypes: {
                    workspace: true,
                    training: true
                }
            },
            spec: {
                image: 'ubuntu',
                command: 'sleep infinity',
                environmentVariables: [
                    { name: 'ENV_ASSET_VAR', value: 'from_asset' },
                    { name: 'ENV_ASSET_VAR2', value: 'from_asset' }
                ]
            }
        };

        const envAsset = await environmentUtils.create(envCreationRequest);
        environmentAssetId = envAsset.meta.id;
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            if (environmentAssetId) {
                await environmentUtils.delete(environmentAssetId);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateShortNameByPrefix('wl');
    });

    for (const wlType of [cliV2WorkloadType.training, cliV2WorkloadType.workspace]) {
        test(`submit ${wlType} workload with environment asset, and additional env var`, async () => {
            test.setTimeout(210 * 1000);

            submitWorkload(
                workloadName,
                {
                    [SubmitFlagsMerged.environment]: environmentAssetName,
                    [SubmitFlagsMerged.project]: project.name,
                    [SubmitFlagsMerged.envVar]: ['ENV_NEW=new']
                } as Partial<typeof SubmitFlagsMerged>,
                wlType
            );

            await validateWorkload(
                workloadName,
                project.id,
                {
                    fields: {
                        image: 'ubuntu',
                        environmentVariables: {
                            ENV_ASSET_VAR: 'from_asset',
                            ENV_ASSET_VAR2: 'from_asset',
                            ENV_NEW: 'new'
                        }
                    }
                },
                wlType,
                true
            );

            deleteWorkload(workloadName, project.name, wlType);
        });
    }

    test('override environment asset - with different image, and environment variable values', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.environment]: environmentAssetName,
                [SubmitFlagsMerged.envVar]: ['ENV_ASSET_VAR=new_value'],
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'alpine',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: 'alpine',
                    environmentVariables: {
                        ENV_ASSET_VAR: 'new_value',
                        ENV_ASSET_VAR2: 'from_asset'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('backward compatibility - key=value format with --environment still works', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'ubuntu',
                [SubmitFlagsMerged.environment]: 'MY_VAR=my_value',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        // Wait for workload to appear in control plane before deletion
        await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
            workloadName,
            project.id.toString()
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
