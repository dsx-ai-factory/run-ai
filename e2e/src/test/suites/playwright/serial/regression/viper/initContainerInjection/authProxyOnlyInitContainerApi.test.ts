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
    authVolumeName,
    findInitContainer,
    initContainerNames,
    InitContainerSuiteContext,
    proxyArgs,
    setupInitContainerSuite,
    submitWorkspaceAndWaitRunning,
    teardownInitContainerSuite
} from './initContainerHelpers';

const EXPOSED_PORT = 8888;

test.describe('Anyworkload init container injection - auth proxy only via API', () => {
    supportedFrom(SupportedVersions.V2_26)(() => {
        let suite: InitContainerSuiteContext;
        let namespace: string;

        test.beforeAll(async () => {
            suite = await setupInitContainerSuite('e2e-init-ctr-auth-proxy');
            ({ namespace } = suite);
        });

        test.afterAll(() => teardownInitContainerSuite(suite));

        test('should inject oauth2 proxy sidecar without user loader for all-authenticated access', async () => {
            const workspaceName = objectGenerator.generateShortNameByPrefix('init-ctr-auth-proxy');

            await test.step('submit workspace with an authenticated-users exposed URL via API', async () => {
                await submitWorkspaceAndWaitRunning(namespace, suite.project.id.toString(), workspaceName, {
                    exposedUrls: [
                        {
                            container: EXPOSED_PORT,
                            authorizationType: AuthorizationType.AuthenticatedUsers
                        }
                    ]
                });
            });

            await test.step('verify oauth2 proxy sidecar is present and no user loader is injected', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    await assertOwnedByWorkloadRunner(pod);

                    await assertInitContainerHealthy(
                        pod,
                        `${initContainerNames.oauth2ProxyPrefix}0`,
                        'sidecar'
                    );

                    const loader = findInitContainer(pod, (n) => n === initContainerNames.usernameLoader);
                    expect(
                        loader,
                        `unexpected username loader container on pod ${pod.metadata?.name}`
                    ).toBeUndefined();
                }
            });

            await test.step('verify proxy args use email-domain wildcard and not an emails file', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const args = proxyArgs(pod, 0).join(' ');
                    expect(args).toContain('--email-domain=*');
                    expect(args).not.toContain('--authenticated-emails-file');
                    expect(args).not.toContain('--allowed-group=');
                }
            });

            await test.step('verify runai-auth empty-dir volume is mounted into the pod', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const authVolume = pod.spec?.volumes?.find((v) => v.name === authVolumeName);
                    expect(
                        authVolume?.emptyDir,
                        `expected emptyDir volume "${authVolumeName}" on pod ${pod.metadata?.name}`
                    ).toBeDefined();
                }
            });
        });
    });
});
