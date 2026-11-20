/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';

import { AuthorizationType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/authorizationType';
import { kubeCli } from '../../../../../../../modules/cli';
import { objectGenerator } from '../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

import {
    assertInitContainerHealthy,
    assertOwnedByWorkloadRunner,
    authenticatedEmailsFileFromProxyArgs,
    authVolumePath,
    decodeLoaderPayloads,
    findInitContainer,
    initContainerNames,
    InitContainerSuiteContext,
    proxyArgs,
    setupInitContainerSuite,
    submitWorkspaceAndWaitRunning,
    teardownInitContainerSuite
} from './initContainerHelpers';

const EXPOSED_PORT = 8888;

test.describe('Anyworkload init container injection - auth proxy with user loader via API', () => {
    supportedFrom(SupportedVersions.V2_26)(() => {
        let suite: InitContainerSuiteContext;
        let namespace: string;
        let authorizedEmail: string;

        test.beforeAll(async () => {
            suite = await setupInitContainerSuite('e2e-init-ctr-auth-loader');
            ({ namespace, authorizedEmail } = suite);
        });

        test.afterAll(() => teardownInitContainerSuite(suite));

        test('should inject oauth2 proxy plus user loader for specific-users access', async () => {
            const workspaceName = objectGenerator.generateShortNameByPrefix('init-ctr-auth-loader');

            await test.step('submit workspace with specific-users access via API', async () => {
                await submitWorkspaceAndWaitRunning(namespace, suite.project.id.toString(), workspaceName, {
                    exposedUrls: [
                        {
                            container: EXPOSED_PORT,
                            authorizationType: AuthorizationType.AuthorizedUsers,
                            authorizedUsers: [authorizedEmail]
                        }
                    ]
                });
            });

            await test.step('verify both proxy sidecar and user loader init container ran successfully', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    await assertOwnedByWorkloadRunner(pod);
                    await assertInitContainerHealthy(pod, initContainerNames.usernameLoader, 'oneShot');
                    await assertInitContainerHealthy(
                        pod,
                        `${initContainerNames.oauth2ProxyPrefix}0`,
                        'sidecar'
                    );
                }
            });

            await test.step('verify the loader writes the auth file the proxy enforces', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const authFile = authenticatedEmailsFileFromProxyArgs(pod, 0);
                    expect(authFile.startsWith(authVolumePath)).toBe(true);

                    const loader = findInitContainer(pod, (n) => n === initContainerNames.usernameLoader);
                    expect(loader).toBeDefined();
                    const command = loader?.command?.join(' ') ?? '';
                    expect(command).toContain(`base64 -d > ${authFile}`);
                }
            });

            await test.step('verify proxy args reference the authenticated-emails-file and not a wildcard domain', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const argsJoined = proxyArgs(pod, 0).join(' ');
                    expect(argsJoined).toContain('--authenticated-emails-file=');
                    expect(argsJoined).not.toContain('--email-domain=*');
                }
            });

            await test.step('verify the loader encodes the authorized user into the auth file', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const loader = findInitContainer(pod, (n) => n === initContainerNames.usernameLoader);
                    const command = loader?.command?.join(' ') ?? '';
                    const decoded = decodeLoaderPayloads(command);
                    expect(decoded).toContain(authorizedEmail);
                }
            });
        });
    });
});
