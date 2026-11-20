/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../../cli';
import { logger, waitUtils } from '..';
import { expect } from '@playwright/test';
import { V1Pod } from '@kubernetes/client-node';

export class NgcCredentialVerificationUtils {
    async waitForPodAndVerifyImagePullMessage(
        namespace: string,
        workloadPrefix: string,
        workloadName: string,
        expectedPodCount: number = 1
    ): Promise<string> {
        let pods: V1Pod[] = [];
        await waitUtils.waitForCondition(
            async () => {
                pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadPrefix);
                return pods.length >= expectedPodCount;
            },
            `Waiting for ${expectedPodCount} pod(s) to be created in namespace ${namespace}`,
            true,
            60
        );
        expect(pods.length).toBe(expectedPodCount);
        const podName = pods[0].metadata.name;
        logger.info(`Pod created: ${podName}, status: ${pods[0].status?.phase}`);

        let imagePullEvents = [];
        await waitUtils.waitForCondition(
            async () => {
                const allEvents = await kubeCli.CoreV1Api.listNamespacedEvent({ namespace });
                const podEvents = allEvents.items.filter((event) => event.involvedObject?.name === podName);

                imagePullEvents = podEvents.filter(
                    (event) =>
                        event.message?.includes('already present on machine') ||
                        event.message?.includes('Successfully pulled image')
                );

                if (imagePullEvents.length === 0) {
                    logger.debug(`Checking for image pull events... (${podEvents.length} total pod events)`);
                }

                return imagePullEvents.length > 0;
            },
            `Waiting for image pull events for pod ${podName}`,
            true,
            120,
            5
        );
        expect(imagePullEvents.length).toBeGreaterThan(0);
        logger.info(
            `Verified: Found ${imagePullEvents.length} image pull event(s) for workload ${workloadName}`
        );
        return podName;
    }

    async verifyNgcApiKeyEnvInPod(
        namespace: string,
        workloadPrefix: string,
        contextLabel?: string
    ): Promise<{ secretName: string; secretKey: string }> {
        const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadPrefix);
        expect(pods.length).toBeGreaterThan(0);

        const userContainer = pods[0].spec.containers.find((c) => c.name === 'user-container');
        expect(userContainer).toBeDefined();

        const ngcApiKeyEnv = userContainer.env?.find((e) => e.name === 'NGC_API_KEY');
        expect(ngcApiKeyEnv).toBeDefined();
        expect(ngcApiKeyEnv.valueFrom?.secretKeyRef).toBeDefined();

        const secretName = ngcApiKeyEnv.valueFrom.secretKeyRef.name;
        const secretKey = ngcApiKeyEnv.valueFrom.secretKeyRef.key;
        const prefix = contextLabel ? `${contextLabel} - ` : '';
        logger.info(
            `${prefix}NGC_API_KEY environment variable found, secret: ${secretName}, key: ${secretKey}`
        );

        return { secretName, secretKey };
    }

    async verifyNgcApiKeySecret(
        namespace: string,
        secretName: string,
        secretKey: string,
        expectedNgcApiKey: string,
        contextLabel?: string
    ): Promise<void> {
        const secret = await kubeCli.CoreV1Api.readNamespacedSecret({
            name: secretName,
            namespace
        });
        expect(secret.data).toBeDefined();
        expect(secret.data[secretKey]).toBeDefined();

        const decodedSecret = Buffer.from(secret.data[secretKey], 'base64').toString('utf-8');
        expect(decodedSecret.trim()).toEqual(expectedNgcApiKey);
        const prefix = contextLabel ? `${contextLabel} - ` : '';
        logger.info(`${prefix}Secret value matches NGC API key - verification successful!`);
    }
}
