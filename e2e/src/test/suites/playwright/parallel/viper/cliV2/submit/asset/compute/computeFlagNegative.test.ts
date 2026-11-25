/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { assert } from 'chai';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    computeResourceUtils,
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
import { ComputeCreationRequest, Scope } from '../../../../../../../../../modules/api/apis/assetsApi';

test.describe(`CLI V2 Submit with Compute Asset Negative Tests, ${tags.teams.viper}, ${tags.runOn.parallelCI}`, () => {
    let project: Project;
    let otherProject: Project;
    let otherProjectComputeId: string;
    let otherProjectComputeName: string;
    let workloadName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-compute-neg');

        // Create another project with a compute asset scoped to it
        otherProject = await projectResourcesUtils.createProjectAndNamespace('cli-compute-other');
        otherProjectComputeName = objectGenerator.generateNameByPrefix('compute-other');

        const computeRequest: ComputeCreationRequest = {
            meta: {
                name: otherProjectComputeName,
                scope: Scope.Project,
                projectId: otherProject.id,
                clusterId: flowRootState.cluster.uuid
            },
            spec: {
                cpuCoreRequest: 0.1
            }
        };

        const computeAsset = await computeResourceUtils.create(computeRequest);
        otherProjectComputeId = computeAsset.meta.id;
    });

    test.beforeEach(() => {
        workloadName = objectGenerator.generateNameByPrefix('wl-neg');
    });

    test.afterAll(async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            await operations.deleteProjectAndNamespace(otherProject.id);
            if (otherProjectComputeId) {
                await computeResourceUtils.delete(otherProjectComputeId);
            }
        } catch (error) {
            console.error('Cleanup failed:', error);
        }
    });

    test('fail when compute asset does not exist', () => {
        const nonExistentCompute = 'non-existent-compute-12345';

        const submitFlags = {
            [SubmitFlagsMerged.compute]: nonExistentCompute,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            /compute.*not found/i
        );
    });

    test('fail when compute asset is scoped to a different project', () => {
        const submitFlags = {
            [SubmitFlagsMerged.compute]: otherProjectComputeName,
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        assert.throw(
            () => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training),
            /compute.*not found/i
        );
    });
});
