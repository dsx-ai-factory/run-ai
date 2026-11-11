/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { assert } from 'chai';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    environmentUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags
} from '../../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { EnvironmentCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI V2 Submit with Environment Asset Negative Tests, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, () => {
    let project: Project;
    let otherProject: Project;
    let otherProjectEnvId: string;
    let otherProjectEnvName: string;
    let trainingOnlyEnvId: string;
    let trainingOnlyEnvName: string;
    let workloadName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('p1');

        // Create another project with an environment asset scoped to it
        otherProject = await projectResourcesUtils.createProjectAndNamespace('p2');
        otherProjectEnvName = objectGenerator.generateNameByPrefix('env-other');

        const envRequest: EnvironmentCreationRequest = {
            meta: {
                name: otherProjectEnvName,
                scope: Scope.Project,
                projectId: otherProject.id,
                clusterId: flowRootState.cluster.uuid,
                workloadSupportedTypes: {
                    workspace: true,
                    training: true
                }
            },
            spec: {
                image: 'ubuntu',
                command: 'sleep infinity'
            }
        };

        const envAsset = await environmentUtils.create(envRequest);
        otherProjectEnvId = envAsset.meta.id;

        // Create environment asset that only supports training
        trainingOnlyEnvName = objectGenerator.generateNameByPrefix('env-training');
        const trainingOnlyEnvRequest: EnvironmentCreationRequest = {
            meta: {
                name: trainingOnlyEnvName,
                scope: Scope.Project,
                projectId: project.id,
                clusterId: flowRootState.cluster.uuid,
                workloadSupportedTypes: {
                    workspace: false,
                    training: true
                }
            },
            spec: {
                image: 'ubuntu',
                command: 'sleep infinity'
            }
        };

        const trainingOnlyEnvAsset = await environmentUtils.create(trainingOnlyEnvRequest);
        trainingOnlyEnvId = trainingOnlyEnvAsset.meta.id;
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await operations.deleteProjectAndNamespace(otherProject.id);
            if (otherProjectEnvId) {
                await environmentUtils.delete(otherProjectEnvId);
            }
            if (trainingOnlyEnvId) {
                await environmentUtils.delete(trainingOnlyEnvId);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-neg');
    });

    test('fail when environment asset does not exist', () => {
        const nonExistentEnv = 'non-existent-env-12345';

        const submitFlags = {
            [SubmitFlagsMerged.environment]: nonExistentEnv,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `Error: failed to submit. the environment '${nonExistentEnv}' does not support training workloads`
        );
    });

    test('fail when environment asset is scoped to a different project', () => {
        const submitFlags = {
            [SubmitFlagsMerged.environment]: otherProjectEnvName,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            `Error: failed to submit. the environment '${otherProjectEnvName}' does not support training workloads`
        );
    });

    test('fail when using training-only environment asset for workspace', () => {
        const submitFlags = {
            [SubmitFlagsMerged.environment]: trainingOnlyEnvName,
            [SubmitFlagsMerged.project]: project.name
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.workspace),
            `Error: failed to submit. the environment '${trainingOnlyEnvName}' does not support workspace workloads`
        );
    });
});
