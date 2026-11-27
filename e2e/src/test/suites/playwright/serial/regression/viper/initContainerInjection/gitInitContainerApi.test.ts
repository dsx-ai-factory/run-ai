/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';

import { kubeCli } from '../../../../../../../modules/cli';
import { logger, objectGenerator } from '../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

import {
    assertGitSyncInitContainerHealthy,
    InitContainerSuiteContext,
    setupInitContainerSuite,
    submitWorkspaceAndWaitRunning,
    teardownInitContainerSuite
} from './initContainerHelpers';

const GIT_REPOSITORY = 'https://github.com/octocat/Hello-World';
const GIT_BRANCH = 'master';
const GIT_MOUNT_PATH = '/mnt/data';

test.describe('Anyworkload init container injection - git data source via API', () => {
    supportedFrom(SupportedVersions.V2_26)(() => {
        let suite: InitContainerSuiteContext;
        let namespace: string;

        test.beforeAll(async () => {
            suite = await setupInitContainerSuite('e2e-init-ctr-git');
            ({ namespace } = suite);
        });

        test.afterAll(() => teardownInitContainerSuite(suite));

        test('should inject git-sync init container and clone the repo into the pod', async () => {
            const workspaceName = objectGenerator.generateShortNameByPrefix('init-ctr-git-ws');

            await test.step('submit workspace with a git data source via API', async () => {
                await submitWorkspaceAndWaitRunning(namespace, suite.project.id.toString(), workspaceName, {
                    storage: {
                        git: [{ repository: GIT_REPOSITORY, branch: GIT_BRANCH, path: GIT_MOUNT_PATH }]
                    }
                });
            });

            await test.step('verify git-sync init container is present and successful', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    await assertGitSyncInitContainerHealthy(pod);
                }
            });

            await test.step('verify cloned repo content is reachable in the pod', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    const { stdout } = await kubeCli.execInPod(pod.metadata?.name as string, namespace, [
                        'ls',
                        `${GIT_MOUNT_PATH}/Hello-World`
                    ]);
                    logger.info(`ls ${GIT_MOUNT_PATH}/Hello-World on ${pod.metadata?.name}: ${stdout}`);
                    expect(stdout).toContain('README');
                }
            });
        });
    });
});
