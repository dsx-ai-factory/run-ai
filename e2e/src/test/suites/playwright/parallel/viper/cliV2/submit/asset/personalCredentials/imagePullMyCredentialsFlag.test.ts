/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, TestInfo } from '@playwright/test';
import { kubeCli, runaiCliV2 } from '../../../../../../../../../modules/cli';
import { skipPlaywrightTest } from '../../../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { Project } from '../../../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    credentialUtils,
    flowRootState,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    cliV2WorkloadType,
    MyCredentialType,
    SubmitFlagsMerged
} from '../../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

import { submitWorkload, deleteWorkload } from '../../submitTestUtils';
import { Phase } from '../../../../../../../../../modules/api/apis/assetsApi';

const IMAGE_PULL_CREDENTIAL_TYPES = [
    {
        type: MyCredentialType.dockerRegistry,
        createItems: (): string[] => ['user=user@example.com,password=dummy-registry-password,url=https://my.url.com']
    },
    {
        type: MyCredentialType.ngcApiKey,
        createItems: (): string[] => ['ngc-api-key=test-ngc-key-value']
    }
];

const IP_CRED_PREFIX: Record<MyCredentialType.dockerRegistry | MyCredentialType.ngcApiKey, string> = {
    [MyCredentialType.dockerRegistry]: 'ipd',
    [MyCredentialType.ngcApiKey]: 'ipn'
};

test.describe(`CLI V2 Submit --image-pull-my-credentials, ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_25)(() => {
        let project: Project;
        let workloadName: string;
        const credNames: Record<string, string> = {};

        test.beforeAll(async () => {
            const sysAdmin = flowRootState.users.system_administrator!;
            runaiCliV2.loginAsUser(sysAdmin.email, sysAdmin.password);

            project = await projectResourcesUtils.createProjectAndNamespace('cli-mycred-ip');
            workloadName = objectGenerator.generateShortNameByPrefix('wl');

            for (const cred of IMAGE_PULL_CREDENTIAL_TYPES) {
                const name = objectGenerator.generateShortNameByPrefix(IP_CRED_PREFIX[cred.type], 4);
                credNames[cred.type] = name;
                const output = runaiCliV2.myCredentialCreate(name, cred.type, cred.createItems());
                expect(output).toContain(`Credential "${name}" created.`);
            }

            for (const name of Object.values(credNames)) {
                await credentialUtils.waitForPersonalCredentialSecrets(name, sysAdmin.idToken!);
            }
        });

        test.afterAll(async () => {
            await operations.deleteProjectAndNamespace(project.id);
            for (const cred of IMAGE_PULL_CREDENTIAL_TYPES) {
                runaiCliV2.myCredentialDelete(credNames[cred.type], cred.type, false);
            }
        });

        test('submit with --image-pull-my-credentials all personal credentials types', async () => {
            const imagePullMyCredValues = IMAGE_PULL_CREDENTIAL_TYPES.map((cred) =>
                `type=${cred.type},name=${credNames[cred.type]}`
            );

            await test.step('Submit workload with --image-pull-my-credentials', () => {
                submitWorkload(
                    workloadName,
                    {
                        [SubmitFlagsMerged.project]: project.name,
                        [SubmitFlagsMerged.image]: 'ubuntu',
                        [SubmitFlagsMerged.command]: '-- sleep infinity',
                        [SubmitFlagsMerged.imagePullMyCredentials]: imagePullMyCredValues
                    } as Partial<typeof SubmitFlagsMerged>,
                    cliV2WorkloadType.training
                );
            });

            await test.step('Validate image pull secrets in pod spec', async () => {
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 120);
                await waitUtils.waitForCondition(
                    async () => (await kubeCli.listPodsOfJob(project.namespace, workloadName)).length >= 1,
                    `pods for ${workloadName}`,
                    true,
                    120
                );
                const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName);
                expect(pods.length).toBeGreaterThanOrEqual(1);
                const secretNames = (pods[0].spec?.imagePullSecrets ?? []).map((s) => s.name?.toLowerCase() ?? '');
                for (const cred of IMAGE_PULL_CREDENTIAL_TYPES) {
                    const credName = credNames[cred.type];
                    expect(
                        secretNames.some((sn) => sn.includes(credName.toLowerCase())),
                        `Expected pod imagePullSecret containing "${credName}". Actual: ${secretNames.join(', ')}`
                    ).toBe(true);
                }
            });

            await test.step('Delete workload', () => {
                deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
            });
        });
    });
});
