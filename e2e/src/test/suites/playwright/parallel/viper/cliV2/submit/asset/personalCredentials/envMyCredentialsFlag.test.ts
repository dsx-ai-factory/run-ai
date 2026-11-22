/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { kubeCli, runaiCliV2 } from '../../../../../../../../../modules/cli';
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

const ENV_MY_CREDENTIALS_TYPES: {
    type: MyCredentialType;
    envVarName: string;
    secretKey?: string;
    createItems: () => string[];
}[] = [
    {
        type: MyCredentialType.genericSecret,
        envVarName: 'PERSONAL_SECRET_VAR',
        secretKey: 'TOKEN',
        createItems: (): string[] => ['key=TOKEN,value=my-secret-value']
    },
    {
        type: MyCredentialType.dockerRegistry,
        envVarName: 'DR_ENV_VAR',
        createItems: (): string[] => ['user=user@example.com,password=dummy-registry-password,url=https://my.url.com']
    },
    {
        type: MyCredentialType.ngcApiKey,
        envVarName: 'NGC_KEY',
        createItems: (): string[] => ['ngc-api-key=test-ngc-key-value']
    }
];

const ENV_CRED_PREFIX: Record<MyCredentialType, string> = {
    [MyCredentialType.genericSecret]: 'm-gs',
    [MyCredentialType.dockerRegistry]: 'm-dr',
    [MyCredentialType.ngcApiKey]: 'm-ng'
};

test.describe(`CLI V2 Submit --env-my-credentials, ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_25)(() => {
        let project: Project;
        let workloadName: string;
        const credNames: Record<string, string> = {};

        test.beforeAll(async () => {
            const sysAdmin = flowRootState.users.system_administrator!;
            runaiCliV2.loginAsUser(sysAdmin.email, sysAdmin.password);

            project = await projectResourcesUtils.createProjectAndNamespace('cli-my-cred');

            for (const cred of ENV_MY_CREDENTIALS_TYPES) {
                const name = objectGenerator.generateShortNameByPrefix(ENV_CRED_PREFIX[cred.type], 4);
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
            for (const cred of ENV_MY_CREDENTIALS_TYPES) {
                runaiCliV2.myCredentialDelete(credNames[cred.type], cred.type, false);
            }
        });

        test.beforeEach(() => {
            workloadName = objectGenerator.generateShortNameByPrefix('wl');
        });

        test('submit with --env-my-credentials all personal credentials types', async () => {
            const envMyCredValues = ENV_MY_CREDENTIALS_TYPES.map((cred) => {
                const credName = credNames[cred.type];
                const parts = [`type=${cred.type},name=${cred.envVarName},credential-name=${credName}`];
                if (cred.secretKey) parts.push(`key=${cred.secretKey}`);
                return parts.join(',');
            });

            await test.step('Submit workload with --env-my-credentials', () => {
                submitWorkload(
                    workloadName,
                    {
                        [SubmitFlagsMerged.project]: project.name,
                        [SubmitFlagsMerged.image]: 'ubuntu',
                        [SubmitFlagsMerged.command]: '-- sleep infinity',
                        [SubmitFlagsMerged.envMyCredentials]: envMyCredValues
                    } as Partial<typeof SubmitFlagsMerged>,
                    cliV2WorkloadType.training
                );
            });

            await test.step('Validate credential env vars in pod spec', async () => {
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 120);

                await waitUtils.waitForCondition(
                    async () => (await kubeCli.listPodsOfJob(project.namespace, workloadName)).length >= 1,
                    `pods for ${workloadName}`,
                    true,
                    300
                );
                const pods = await kubeCli.listPodsOfJob(project.namespace, workloadName);
                expect(pods.length).toBeGreaterThanOrEqual(1);
                const podEnv = pods[0].spec?.containers?.[0]?.env ?? [];

                for (const cred of ENV_MY_CREDENTIALS_TYPES) {
                    const credName = credNames[cred.type];
                    const envEntry = podEnv.find((e) => e.name === cred.envVarName);
                    expect(
                        envEntry,
                        `Expected pod env var "${cred.envVarName}" for credential "${credName}". ` +
                            `Actual env names: ${podEnv.map((e) => e.name).join(', ')}`
                    ).toBeDefined();
                    expect(
                        envEntry!.valueFrom?.secretKeyRef?.name?.toLowerCase(),
                        `Expected secretKeyRef for "${cred.envVarName}" to reference credential "${credName}"`
                    ).toContain(credName.toLowerCase());
                }
            });

            await test.step('Delete workload', () => {
                deleteWorkload(workloadName, project.name, cliV2WorkloadType.training);
            });
        });
    });
});
