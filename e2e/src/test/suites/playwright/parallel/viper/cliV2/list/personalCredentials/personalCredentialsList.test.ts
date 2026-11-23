/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { runaiCliV2 } from '../../../../../../../../modules/cli';
import { flowRootState, logger, objectGenerator, tags } from '../../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { listFlags, MyCredentialType } from '../../../../../../../../modules/cli/tools/RunaiCliV2.constants';

const LIST_CRED_PREFIX: Record<MyCredentialType, string> = {
    [MyCredentialType.genericSecret]: 'lgs',
    [MyCredentialType.dockerRegistry]: 'ldr',
    [MyCredentialType.ngcApiKey]: 'lng'
};

const LIST_CREDENTIAL_TYPES: { type: MyCredentialType; createItems: () => string[] }[] = [
    { type: MyCredentialType.genericSecret, createItems: (): string[] => ['key=K,value=V'] },
    { type: MyCredentialType.dockerRegistry, createItems: (): string[] => ['user=user@example.com,password=dummy-registry-password,url=https://my.url.com'] },
    { type: MyCredentialType.ngcApiKey, createItems: (): string[] => ['ngc-api-key=test-key'] }
];

test.describe(`CLI V2 my-credential list, ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_25)(() => {
        const credNames: Partial<Record<MyCredentialType, string>> = {};

        test.beforeAll((): void => {
            const sysAdmin = flowRootState.users.system_administrator!;
            runaiCliV2.loginAsUser(sysAdmin.email, sysAdmin.password);

            for (const cred of LIST_CREDENTIAL_TYPES) {
                const name = objectGenerator.generateShortNameByPrefix(LIST_CRED_PREFIX[cred.type], 4);
                credNames[cred.type] = name;
                const output = runaiCliV2.myCredentialCreate(name, cred.type, cred.createItems());
                logger.debug(`my-credential create output (${cred.type}): ${output}`);
                expect(output).toContain(`Credential "${name}" created.`);
            }
        });

        test.afterAll((): void => {
            for (const cred of LIST_CREDENTIAL_TYPES) {
                const name = credNames[cred.type];
                if (name) {
                    runaiCliV2.myCredentialDelete(name, cred.type, false);
                }
            }
        });

        test('list returns all credentials', (): void => {
            const allCredNames = Object.values(credNames) as string[];
            const result = runaiCliV2.myCredentialList();
            logger.debug(`my-credential list stdout: ${String(result.stdout)}`);
            expect(result.status).toBe(0);
            const stdout = String(result.stdout);
            for (const name of allCredNames) {
                expect(stdout, `Expected credential ${name} in list output`).toContain(name);
            }
        });

        for (const cred of LIST_CREDENTIAL_TYPES) {
            const credType = cred.type;
            test(`list with --type ${credType}`, (): void => {
                const credName = credNames[credType];
                expect(credName).toBeDefined();
                const result = runaiCliV2.myCredentialList({
                    [listFlags.type]: credType
                } as Partial<typeof listFlags>);
                logger.debug(`my-credential list --type ${credType} stdout: ${String(result.stdout)}`);
                expect(result.status).toBe(0);
                expect(String(result.stdout)).toContain(credName);
            });
        }
    });
});
