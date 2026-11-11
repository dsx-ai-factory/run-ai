/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    environmentUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    templateUtils
} from '../../../../../../../../modules/helpers';

import { test, TestInfo } from '@playwright/test';
import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { submitWorkload, validateWorkload, deleteWorkload } from '../submitTestUtils';
import { skipPlaywrightTest } from '../../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { EnvironmentCreationRequest, Scope } from '../../../../../../../../modules/api/apis/assetsApi';
import { ScopeType, WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`CLI V2 Submit with Template and Environment Flags, ${tags.teams.viper}`, { tag: [tags.services.assets, tags.services.workloads, tags.runOn.parallelCI] }, () => {
    let project: Project;
    let templateId: string;
    let templateName: string;
    let environmentName: string;
    let workloadName: string;

    // Environment asset values (different from template)
    let environmentId: string;
    const envAssetImage = 'alpine';
    const envAssetEnvVar = { name: 'ENV_ASSET_VAR', value: 'from_env_asset' };

    // Template inline values (different from asset)
    const templateImage = 'ubuntu';
    const templateEnvVar = { name: 'TEMPLATE_VAR', value: 'template_value' };

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('p');

        // Create environment asset with specific values (different from template)
        environmentName = objectGenerator.generateNameByPrefix('env');
        const envRequest: EnvironmentCreationRequest = {
            meta: {
                name: environmentName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid,
                workloadSupportedTypes: {
                    workspace: true,
                    training: true
                }
            },
            spec: {
                image: envAssetImage,
                command: 'sleep infinity',
                environmentVariables: [envAssetEnvVar]
            }
        };
        const envAsset = await environmentUtils.create(envRequest);
        environmentId = envAsset.meta.id;

        // Create template with inline values (different from asset)
        templateName = objectGenerator.generateNameByPrefix('tmpl');
        const templateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: project.id.toString(),
                templateName: templateName,
                description: 'Template for environment tests'
            },
            {
                image: templateImage,
                command: 'sleep infinity',
                environmentVariables: [templateEnvVar]
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
            if (environmentId) {
                await environmentUtils.delete(environmentId);
            }
            await operations.deleteProjectAndNamespace(project.id);
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    // ==========================================
    // Asset values win
    // ==========================================

    test(`--environment asset image overrides template image, ${tags.runOn.parallelCI}`, async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: envAssetImage,
                    environmentVariables: {
                        ENV_ASSET_VAR: 'from_env_asset',
                        TEMPLATE_VAR: 'template_value'
                    }
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

    test('--image flag overrides --environment asset image', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'nginx',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: 'nginx'
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('--env-var flag overrides --environment asset env var', async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.envVar]: ['ENV_ASSET_VAR=overridden_value']
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: envAssetImage,
                    environmentVariables: {
                        ENV_ASSET_VAR: 'overridden_value',
                        TEMPLATE_VAR: 'template_value'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('--env-var flag overrides template env var', async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.envVar]: ['TEMPLATE_VAR=overridden_template']
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: envAssetImage,
                    environmentVariables: {
                        ENV_ASSET_VAR: 'from_env_asset',
                        TEMPLATE_VAR: 'overridden_template'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('-image + --env-var flags override both asset values', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'nginx',
                [SubmitFlagsMerged.envVar]: ['ENV_ASSET_VAR=also_overridden'],
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: 'nginx',
                    environmentVariables: {
                        TEMPLATE_VAR: 'template_value',
                        ENV_ASSET_VAR: 'also_overridden'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
    // ==========================================
    // Add environment variable
    // ==========================================

    test(`--env-var adds new env var, ${tags.runOn.parallelCI}`, async ({}, testInfo: TestInfo) => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.environment]: environmentName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.envVar]: ['BRAND_NEW=brand_new']
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: envAssetImage,
                    environmentVariables: {
                        TEMPLATE_VAR: 'template_value',
                        ENV_ASSET_VAR: 'from_env_asset',
                        BRAND_NEW: 'brand_new'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
