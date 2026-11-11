/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { logger, waitUtils } from '..';
import { kubeCli } from '../../cli';
import { V1Deployment } from '@kubernetes/client-node';
import axios from 'axios';
import https from 'https';

const insecureAgent = new https.Agent({ rejectUnauthorized: false });

export class WorkloadAuthUtils {
    async getAuthProxyDeployments(namespace: string): Promise<V1Deployment[]> {
        const deployments = await kubeCli.AppsV1Api.listNamespacedDeployment({ namespace });
        return deployments.items.filter((d) => d.metadata?.name?.includes('auth-proxy'));
    }

    async verifyNoAuthProxyDeployment(namespace: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const deployments = await this.getAuthProxyDeployments(namespace);
                return deployments.length === 0;
            },
            `no auth-proxy deployment in namespace ${namespace}`,
            true,
            30,
            5
        );
        logger.info(`Verified no auth-proxy deployment in namespace ${namespace}`);
    }

    async verifyAuthProxyDeploymentExists(namespace: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const deployments = await this.getAuthProxyDeployments(namespace);
                logger.info(
                    `Auth-proxy deployments: ${deployments.map((d) => d.metadata?.name).join(', ') || 'none'}`
                );
                return deployments.length > 0;
            },
            `auth-proxy deployment to appear in namespace ${namespace}`,
            true,
            60,
            5
        );

        const deployments = await this.getAuthProxyDeployments(namespace);
        expect(deployments[0].spec?.replicas).to.equal(1);
        logger.info('Verified auth-proxy deployment exists with 1 replica');
    }

    async getAuthProxyContainerArgs(namespace: string): Promise<string[]> {
        const deployments = await this.getAuthProxyDeployments(namespace);
        expect(deployments.length).to.be.greaterThan(0);
        const container = deployments[0].spec?.template?.spec?.containers?.find(
            (c) => c.name === 'oauth2-proxy'
        );
        expect(container, 'oauth2-proxy container not found').to.not.be.undefined;
        return container?.args || [];
    }

    async verifyAuthProxyGroupArgs(namespace: string, expectedGroups: string[]): Promise<void> {
        const args = await this.getAuthProxyContainerArgs(namespace);
        for (const group of expectedGroups) {
            const hasGroup = args.some((arg) => arg === `--allowed-group=${group}`);
            expect(hasGroup, `Expected --allowed-group=${group} in auth-proxy args`).to.be.true;
        }
        logger.info(`Verified auth-proxy has --allowed-group for: ${expectedGroups.join(', ')}`);
    }

    async verifyAuthProxyConfig(namespace: string, wlName: string): Promise<void> {
        const args = await this.getAuthProxyContainerArgs(namespace);

        const expectedArgs = [
            '--proxy-prefix=/oauth2',
            '--reverse-proxy=true',
            '--http-address=0.0.0.0:4180',
            '--provider=keycloak-oidc',
            '--skip-jwt-bearer-tokens=true',
            '--code-challenge-method=S256'
        ];
        for (const expected of expectedArgs) {
            expect(
                args.some((a) => a.includes(expected)),
                `Missing arg: ${expected}`
            ).to.be.true;
        }

        const upstreamArg = args.find((a) => a.includes('--upstream='));
        expect(upstreamArg, 'Missing --upstream arg').to.not.be.undefined;
        expect(upstreamArg).to.include(`${wlName}.${namespace}.svc.cluster.local`);

        logger.info(`Verified auth-proxy configuration for workload ${wlName}`);
    }

    async verifyUsernameLoaderInitContainer(namespace: string): Promise<void> {
        const deployments = await this.getAuthProxyDeployments(namespace);
        expect(deployments.length).to.be.greaterThan(0);

        const initContainers = deployments[0].spec?.template?.spec?.initContainers || [];
        const initContainerNames = initContainers.map((c) => c.name);
        logger.info(`Auth-proxy init containers: ${JSON.stringify(initContainerNames)}`);

        const usernameLoader = initContainers.find((c) => c.name === 'runai-auth-user-loader');
        expect(
            usernameLoader,
            `runai-auth-user-loader init container not found. Available init containers: [${initContainerNames.join(', ')}]`
        ).to.not.be.undefined;
        logger.info('Verified runai-auth-user-loader init container exists');
    }

    async verifyAuthProxyServicePort(namespace: string): Promise<void> {
        const services = await kubeCli.CoreV1Api.listNamespacedService({ namespace });
        const authProxySvc = services.items.find((s) => s.metadata?.name?.includes('auth-proxy'));
        expect(authProxySvc, 'Auth proxy service not found').to.not.be.undefined;

        const port = authProxySvc?.spec?.ports?.[0];
        expect(port?.port, 'Auth proxy service port should be 80').to.equal(80);
        expect(port?.targetPort, 'Auth proxy target port should be 4180').to.satisfy(
            (tp: any) => tp === 4180 || tp === '4180'
        );
        logger.info('Verified auth proxy service port 80 -> targetPort 4180');
    }

    async verifyOwnerReferences(namespace: string): Promise<void> {
        const deployments = await this.getAuthProxyDeployments(namespace);
        expect(deployments.length).to.be.greaterThan(0);

        const ownerRefs = deployments[0].metadata?.ownerReferences || [];
        expect(ownerRefs.length, 'Auth proxy should have ownerReferences').to.be.greaterThan(0);
        logger.info(`Verified ownerReferences: ${ownerRefs.map((r) => `${r.kind}/${r.name}`).join(', ')}`);
    }

    async verifyExternalUrlHttpStatus(
        url: string,
        token: string,
        expectedStatus: number | number[]
    ): Promise<void> {
        const acceptedStatuses = Array.isArray(expectedStatus) ? expectedStatus : [expectedStatus];
        const maxRetries = 6;
        let lastError: Error | null = null;

        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                if (attempt > 1) {
                    await waitUtils.sleep(2000);
                }

                logger.info(
                    `Verifying URL status (attempt ${attempt}/${maxRetries}): expecting ${acceptedStatuses.join('/')}`
                );

                const headers: Record<string, string> = {};
                if (token) {
                    headers.Authorization = `Bearer ${token}`;
                }

                const response = await axios.get(url, {
                    headers,
                    timeout: 10000,
                    httpsAgent: insecureAgent,
                    validateStatus: () => true,
                    maxRedirects: 0
                });
                const actualStatus = response.status;

                if (actualStatus === 503 && attempt < maxRetries) {
                    logger.warn('Got 503, retrying...');
                    continue;
                }

                expect(acceptedStatuses).to.include(actualStatus);
                logger.info(`Verified HTTP ${actualStatus} as expected`);
                return;
            } catch (error) {
                lastError = error as Error;
                if (attempt < maxRetries) {
                    logger.warn(`Attempt ${attempt} failed: ${(error as Error).message}, retrying...`);
                    continue;
                }
            }
        }
        throw lastError || new Error(`Failed to verify URL status after ${maxRetries} attempts`);
    }

    async waitForWorkloadIngress(namespace: string, wlName: string): Promise<any> {
        let ingress: any;
        await waitUtils.waitForCondition(
            async () => {
                const ingresses = await kubeCli.networkingV1Api.listNamespacedIngress({ namespace });
                ingress = ingresses.items.find((i) => i.metadata?.name?.includes(wlName));
                return !!ingress;
            },
            `ingress for ${wlName} in namespace ${namespace}`,
            true,
            60,
            5
        );
        logger.info(`Found ingress: ${ingress.metadata.name}`);
        return ingress;
    }

    async verifyIngressBackendPointsTo(
        namespace: string,
        wlName: string,
        shouldPointToAuthProxy: boolean
    ): Promise<void> {
        const ingress = await this.waitForWorkloadIngress(namespace, wlName);
        const backend = ingress.spec?.rules?.[0]?.http?.paths?.[0]?.backend;
        const serviceName = backend?.service?.name || '';

        if (shouldPointToAuthProxy) {
            expect(serviceName).to.include('auth-proxy');
            logger.info(`Verified ingress backend points to auth-proxy: ${serviceName}`);
        } else {
            expect(serviceName).to.not.include('auth-proxy');
            logger.info(`Verified ingress backend points directly to workload: ${serviceName}`);
        }
    }

    async verifyAuthProxyDeploymentCleanedUp(namespace: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const deployments = await this.getAuthProxyDeployments(namespace);
                return deployments.length === 0;
            },
            `auth-proxy deployment to be cleaned up in ${namespace}`,
            true,
            60,
            5
        );
        logger.info(`Verified auth-proxy deployment cleaned up in namespace ${namespace}`);
    }

    async getWorkloadRunnerNetworking(namespace: string, wlName: string): Promise<any> {
        let networking: any;
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const runners = await kubeCli.CustomObjectsApi.listNamespacedCustomObject({
                        group: 'run.ai',
                        version: 'v1alpha1',
                        namespace,
                        plural: 'workloadrunners'
                    });
                    const items = (runners as any).items || [];
                    const runner = items.find((r: any) => r.metadata?.name?.includes(wlName));
                    if (runner?.spec?.instructions?.networking) {
                        networking = runner.spec.instructions.networking;
                        return true;
                    }
                    return false;
                } catch {
                    return false;
                }
            },
            `WorkloadRunner networking for ${wlName}`,
            true,
            60,
            5
        );
        return networking;
    }
}
