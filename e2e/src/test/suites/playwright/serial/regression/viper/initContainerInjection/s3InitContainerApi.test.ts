/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';

import { kubeCli } from '../../../../../../../modules/cli';
import { logger, objectGenerator, waitUtils } from '../../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

import {
    assertInitContainerHealthy,
    assertOwnedByWorkloadRunner,
    findInitContainer,
    initContainerNames,
    InitContainerSuiteContext,
    setupInitContainerSuite,
    submitWorkspaceAndWaitRunning,
    teardownInitContainerSuite
} from './initContainerHelpers';

// goofys mounts via anonymous bucket HEAD + ListObjects, so the bucket must allow anonymous LIST,
// not just GetObject. noaa-goes16 (AWS Open Data, us-east-1) allows both.
const PUBLIC_BUCKET = 'noaa-goes16';
const S3_URL = 'https://s3.amazonaws.com';
const MOUNT_PATH = '/mnt/s3';

test.describe('Anyworkload init container injection - s3 data source via API', () => {
    supportedFrom(SupportedVersions.V2_26)(() => {
        let suite: InitContainerSuiteContext;
        let namespace: string;

        test.beforeAll(async () => {
            suite = await setupInitContainerSuite('e2e-init-ctr-s3');
            ({ namespace } = suite);
        });

        test.afterAll(() => teardownInitContainerSuite(suite));

        test('should inject s3-fs sidecar init container and mount the bucket', async () => {
            const workspaceName = objectGenerator.generateShortNameByPrefix('init-ctr-s3-ws');

            await test.step(`submit workspace with an s3 data source for bucket ${PUBLIC_BUCKET} via API`, async () => {
                await submitWorkspaceAndWaitRunning(namespace, suite.project.id.toString(), workspaceName, {
                    storage: {
                        s3: [{ bucket: PUBLIC_BUCKET, url: S3_URL, path: MOUNT_PATH }]
                    }
                });
            });

            await test.step('verify s3-fs sidecar init container is present and running', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    await assertOwnedByWorkloadRunner(pod);
                    const s3fs = findInitContainer(pod, (n) => n.startsWith(initContainerNames.s3FsPrefix));
                    if (!s3fs) {
                        throw new Error(
                            `Expected init container starting with "${initContainerNames.s3FsPrefix}" on pod ${pod.metadata?.name}.`
                        );
                    }
                    await assertInitContainerHealthy(pod, s3fs.name, 'sidecar');
                }
            });

            await test.step('verify the s3-fs sidecar established the bucket mount', async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, workspaceName);
                expect(pods.length).toBeGreaterThan(0);
                for (const pod of pods) {
                    // Assert on the mount entry rather than bucket contents: the composer's job is
                    // the mount, and the FUSE entry appears shortly after the sidecar reports Running.
                    let mounts = '';
                    await waitUtils.waitForCondition(
                        async () => {
                            ({ stdout: mounts } = await kubeCli.execInPod(
                                pod.metadata?.name as string,
                                namespace,
                                ['cat', '/proc/mounts']
                            ));
                            return mounts.includes(MOUNT_PATH);
                        },
                        `s3 fuse mount established at ${MOUNT_PATH} on ${pod.metadata?.name}`,
                        true,
                        60,
                        3
                    );
                    const mountLine = mounts.split('\n').find((line) => line.includes(MOUNT_PATH)) ?? '';
                    logger.info(`mount entry for ${MOUNT_PATH} on ${pod.metadata?.name}: ${mountLine}`);
                    expect(mountLine).toContain(MOUNT_PATH);
                    expect(mountLine.toLowerCase()).toContain('fuse');
                }
            });
        });
    });
});
