/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { bashCli } from '../../cli';
import { logger } from '../logger/logger';
import { waitUtils } from '..';

export class HelmReleaseUtils {
    helmReleaseExists(releaseName: string, namespace: string): boolean {
        const res = bashCli.execWithOutput(`helm list -n ${namespace}`, false);
        return res.stdout.includes(releaseName);
    }

    private async execWithRetry(
        command: string,
        description: string,
        preRetryCheck?: () => boolean,
        maxAttempts = 3,
        sleepMs = 3000
    ): Promise<string> {
        let lastStdout = '';
        for (let attempt = 1; attempt <= maxAttempts; attempt++) {
            if (attempt > 1 && preRetryCheck?.()) {
                logger.info(`${description}: pre-retry check passed, skipping retry`);
                return lastStdout;
            }
            const result = bashCli.execWithOutput(command, false);
            lastStdout = result.stdout;
            logger.info(
                `${description} attempt ${attempt}: exit=${result.status}, stdout: ${result.stdout}, stderr: ${result.stderr}`
            );
            if (result.status === 0) {
                logger.info(`${description} succeeded`);
                return result.stdout;
            }
            if (attempt === maxAttempts) {
                throw new Error(
                    `${description} failed after ${maxAttempts} attempts, stderr: ${result.stderr}`
                );
            }
            await waitUtils.sleep(sleepMs);
        }
    }

    async helmInstall(releaseName: string, namespace: string, chartUrl: string, timeout: number = 60): Promise<string> {
        const command = `helm install ${releaseName} ${chartUrl} -n ${namespace}`;
        const stdout = await this.execWithRetry(command, `Helm install ${releaseName} in ${namespace}`, () =>
            this.helmReleaseExists(releaseName, namespace)
        );
        await waitUtils.waitForCondition(
            () => this.helmReleaseExists(releaseName, namespace),
            `Wait for ${releaseName} to be installed in ${namespace}`,
            true,
            timeout,
            5
        );
        return stdout;
    }

    async helmUpgrade(releaseName: string, namespace: string, chartUrl: string): Promise<string> {
        const command = `helm upgrade ${releaseName} ${chartUrl} -n ${namespace}`;
        return await this.execWithRetry(command, `Helm upgrade ${releaseName} in ${namespace}`);
    }

    /**
     * Attempt helm upgrade expecting failure. Does not retry.
     * Returns the stderr output for verification.
     * @param extraArgs - additional helm flags (e.g., '--set image.tag=fake --atomic --timeout 5s')
     */
    helmUpgradeExpectFailure(
        releaseName: string,
        namespace: string,
        chartUrl: string,
        extraArgs?: string
    ): string {
        const command = `helm upgrade ${releaseName} ${chartUrl} -n ${namespace}${extraArgs ? ' ' + extraArgs : ''}`;
        logger.info(`Helm upgrade (expect failure): ${command}`);
        const result = bashCli.execWithOutput(command, false);
        logger.info(
            `Helm upgrade result: exit=${result.status}, stdout: ${result.stdout}, stderr: ${result.stderr}`
        );
        if (result.status === 0) {
            throw new Error(`Expected helm upgrade ${releaseName} to fail but it succeeded`);
        }
        return result.stderr;
    }

    async helmUninstall(releaseName: string, namespace: string, timeout: number = 60): Promise<void> {
        if (!this.helmReleaseExists(releaseName, namespace)) {
            logger.info(`Release ${releaseName} not found in ${namespace}, skipping uninstall`);
            return;
        }
        const command = `helm uninstall ${releaseName} -n ${namespace}`;
        await this.execWithRetry(
            command,
            `Helm uninstall ${releaseName} from ${namespace}`,
            () => !this.helmReleaseExists(releaseName, namespace)
        );
        await waitUtils.waitForCondition(
            () => !this.helmReleaseExists(releaseName, namespace),
            `Wait for ${releaseName} to be uninstalled from ${namespace}`,
            true,
            timeout,
            5
        );
    }

    helmGetValues(releaseName: string, namespace: string): Record<string, any> {
        const q = (s: string): string => {
            const quote = '\'';
            return quote + s.replace(/'/g, quote + '\\' + quote + quote) + quote;
        };
        const result = bashCli.execWithOutput(`helm get values ${q(releaseName)} -n ${q(namespace)} -o json`, false);
        if (result.status !== 0) {
            throw new Error(`helm get values ${releaseName} failed: ${result.stderr}`);
        }
        try {
            return JSON.parse(result.stdout);
        } catch (e) {
            const msg = e instanceof Error ? e.message : String(e);
            throw new Error(`helm get values ${releaseName}: JSON.parse failed: ${msg}; stderr: ${result.stderr}`);
        }
    }
}

export interface SecretRef {
    secretName: string;
    key: string;
}

export function resolveHelmSecretRef(
    valuesSection: Record<string, any> | undefined,
    defaultSecretName: string,
    defaultKey: string,
    secretKeyField = 'adminPasswordKey'
): SecretRef {
    if (valuesSection?.existingSecret) {
        return {
            secretName: valuesSection.existingSecret,
            key: valuesSection.secretKeys?.[secretKeyField] ?? defaultKey,
        };
    }
    return { secretName: defaultSecretName, key: defaultKey };
}
