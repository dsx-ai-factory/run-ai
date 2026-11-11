/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { assert } from 'chai';
import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    templateUtils
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { ScopeType, WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`CLI V2 Submit with Template Negative Tests, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, { tag: [tags.services.assets, tags.services.workloads, tags.runOn.parallelCI] }, () => {
    let project: Project;
    let otherProject: Project;
    let otherProjectTemplateId: string;
    let otherProjectTemplateName: string;
    let trainingOnlyTemplateId: string;
    let trainingOnlyTemplateName: string;
    let workloadName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('p');

        // Create another project with a template scoped to it
        otherProject = await projectResourcesUtils.createProjectAndNamespace('p2');
        otherProjectTemplateName = objectGenerator.generateShortNameByPrefix('tmpl-other');

        const templateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: otherProject.id.toString(),
                templateName: otherProjectTemplateName,
                description: 'Template for scope isolation test'
            },
            {
                environmentVariables: [{ name: 'TEST_VAR', value: 'test_value' }]
            }
        );
        otherProjectTemplateId = templateResponse.body.meta.id;

        // Create training-only template in main project
        trainingOnlyTemplateName = objectGenerator.generateShortNameByPrefix('tmpl-training');
        const trainingOnlyTemplateResponse = await templateUtils.createWorkloadTemplateV2(
            {
                workloadType: WorkloadType.Training,
                scopeType: ScopeType.Project,
                scopeId: project.id.toString(),
                templateName: trainingOnlyTemplateName,
                description: 'Training-only template'
            },
            {
                environmentVariables: [{ name: 'TRAINING_VAR', value: 'training_value' }]
            }
        );
        trainingOnlyTemplateId = trainingOnlyTemplateResponse.body.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-neg');
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await operations.deleteProjectAndNamespace(otherProject.id);
            if (otherProjectTemplateId) {
                await templateUtils.deleteWorkloadTemplateV2(otherProjectTemplateId, WorkloadType.Training);
            }
            if (trainingOnlyTemplateId) {
                await templateUtils.deleteWorkloadTemplateV2(trainingOnlyTemplateId, WorkloadType.Training);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test('fail when template does not exist', () => {
        const nonExistentTemplate = 'non-existent-template';

        const submitFlags = {
            [SubmitFlagsMerged.template]: nonExistentTemplate,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `failed to submit. the template '${nonExistentTemplate}' does not support training workloads`
        );
    });

    test('fail when template is scoped to a different project', () => {
        const submitFlags = {
            [SubmitFlagsMerged.template]: otherProjectTemplateName,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `failed to submit. the template '${otherProjectTemplateName}' does not support training workloads`
        );
    });

    test('fail when template flag is empty string', () => {
        const rawFlags = `--template "" --project ${project.name} --image ubuntu -c -- sleep infinity`;

        assert.throw(
            () => runaiCliV2.submitRawFlags(workloadName, cliV2WorkloadType.training, rawFlags),
            'flag --template requires a non-empty value.'
        );
    });

    test('fail when using training template for workspace', () => {
        const submitFlags = {
            [SubmitFlagsMerged.template]: trainingOnlyTemplateName,
            [SubmitFlagsMerged.project]: project.name
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.workspace),
            `failed to submit. the template '${trainingOnlyTemplateName}' does not support workspace workloads`
        );
    });
});
