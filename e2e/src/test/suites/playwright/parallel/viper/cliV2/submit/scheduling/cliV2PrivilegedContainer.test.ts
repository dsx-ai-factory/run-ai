/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { Project } from '../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    flowRootState,
    objectGenerator,
    orgUnitUtils,
    policyUtils,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import {
    cliV2WorkloadType,
    DeleteFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Phase } from '../../../../../../../../modules/api/apis/assetsApi';
import { assert } from 'chai';
import { ScopeType } from '../../../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../../../modules/helpers/utils/policyUtils';
import { policyApi as policyApiHandler } from '../../../../../../../../modules/api';

test.describe(`CLI V2 Submit --privileged, ${tags.teams.viper}`, () => {
    let project: Project;
    const workloadName = objectGenerator.generateNameByPrefix('t');
    let policyName: string;

    test.beforeAll(async () => {
        project = await projectResourcesUtils.createProjectAndNamespace('cli-privileged');
        policyName = `project-${project.id}`;
    });

    test.afterAll(async () => {
        await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), policyName);
    });

    test('Submit privileged workspace blocked by system policy', () => {
        const submitFlags = {
            [SubmitFlagsMerged.project]: project.name,
            [SubmitFlagsMerged.image]: 'ubuntu',
            [SubmitFlagsMerged.privileged]: '',
            [SubmitFlagsMerged.command]: '-- sleep infinity'
        } as Partial<typeof SubmitFlagsMerged>;

        const errMsg =
            'Field: security.privileged, Details: the administrator prohibited changing the value of this field';
        assert.throw(() => runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.workspace), errMsg);
    });

    test('Set canEdit=true and submit privileged workspace', async () => {
        await test.step('Create workload policy with privileged.canEdit=true', async () => {
            test.setTimeout(210 * 1000);

            policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            await policyUtils.createPolicy(
                true,
                policyName,
                ScopeType.Project.toString(),
                project.id,
                WorkloadType.Interactive,
                undefined,
                {
                    security: {
                        privileged: { canEdit: true }
                    }
                }
            );
        });

        await test.step('Submit privileged workspace workload', () => {
            const submitFlags = {
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.image]: 'ubuntu',
                [SubmitFlagsMerged.privileged]: '',
                [SubmitFlagsMerged.command]: '-- sleep infinity'
            } as Partial<typeof SubmitFlagsMerged>;

            runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.workspace);
        });

        await test.step('Wait until workload is Running', async () => {
            await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running);
        });

        await test.step('Validate workload appears in workload list', () => {
            const workloadListOutput = runaiCliV2
                .list('workload', { '--project': project.name } as any)
                .stdout.toString();
            expect(workloadListOutput).toContain(workloadName);
        });

        await test.step('Cleanup workload and policy', async () => {
            runaiCliV2.delete(
                { [DeleteFlagsMerged.project]: project.name } as Partial<typeof DeleteFlagsMerged>,
                workloadName,
                cliV2WorkloadType.workspace
            );
            await policyUtils.deleteWorkloadPolicyV2(WorkloadType.Interactive, {
                scope: ScopeType.Project.toString(),
                scopeId: project.id.toString()
            });
        });
    });
});
