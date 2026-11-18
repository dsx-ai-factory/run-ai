/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';

import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    templateUtils
} from '../../../../../../../../modules/helpers';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { ScopeType, WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { submitWorkload, validateWorkload, deleteWorkload } from '../submitTestUtils';

test.describe(`CLI V2 Submit with Template, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, { tag: [tags.services.assets, tags.services.workloads, tags.runOn.parallelCI] }, () => {
    let project: Project;
    let templateId: string;
    let templateName: string;
    let workloadName: string;

    const templateCpuRequest = 0.1;
    const templateEnvVar = { name: 'TEMPLATE_VAR', value: 'template_value' };

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-template');
        templateName = objectGenerator.generateShortNameByPrefix('tmpl');

        const templateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: project.id.toString(),
                templateName: templateName,
                description: 'Test template for CLI e2e'
            },
            {
                image: 'ubuntu',
                command: 'sleep infinity',
                compute: {
                    cpuCoreRequest: templateCpuRequest
                },
                environmentVariables: [templateEnvVar]
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
            await operations.deleteProjectAndNamespace(project.id);
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test('submit training workload with template', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.project]: project.name
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: 'ubuntu',
                    cpuCoreRequest: templateCpuRequest,
                    environmentVariables: { TEMPLATE_VAR: 'template_value' }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('override template with image, cpuCoreRequest, and environment-variable flags', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'alpine',
                [SubmitFlagsMerged.cpuCoreRequest]: '0.5',
                [SubmitFlagsMerged.envVar]: ['TEMPLATE_VAR=overridden_value'],
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
                    cpuCoreRequest: 0.5,
                    environmentVariables: {
                        TEMPLATE_VAR: 'overridden_value'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });

    test('add environment variable to template', async () => {
        submitWorkload(
            workloadName,
            {
                [SubmitFlagsMerged.template]: templateName,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.envVar]: ['NEW_VAR=new_value']
            } as Partial<typeof SubmitFlagsMerged>,
            cliV2WorkloadType.training
        );

        await validateWorkload(
            workloadName,
            project.id,
            {
                fields: {
                    image: 'ubuntu',
                    cpuCoreRequest: templateCpuRequest,
                    environmentVariables: {
                        TEMPLATE_VAR: 'template_value',
                        NEW_VAR: 'new_value'
                    }
                }
            },
            cliV2WorkloadType.training,
            true
        );

        deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
    });
});
