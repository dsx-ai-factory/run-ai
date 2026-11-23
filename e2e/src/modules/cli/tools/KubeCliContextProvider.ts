/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, logger, objectGenerator, settings, supportedPlatforms, config, tenantUtils } from '../../helpers';
import * as k8s from '@kubernetes/client-node';

export class KubeCliContextProvider {
    protected kubeClient: k8s.KubeConfig;

    protected cluster: k8s.Cluster;

    adminContext: k8s.Context;

    runaiContext: k8s.Context;

    constructor(kubeconfigPath: string, clusterName?: string) {
        logger.debug(`KubeCli: kubeconfigPath: ${kubeconfigPath}`);
        this.kubeClient = new k8s.KubeConfig();
        kubeconfigPath ? this.kubeClient.loadFromFile(kubeconfigPath) : this.kubeClient.loadFromDefault();
        this.cluster = this.getCluster(clusterName);
    }

    async initKubeContext(clusterName: string, initRunaiUser = true): Promise<void> {
        this.cluster = this.getCluster(clusterName);
        this.adminContext = this.getAdminContext(this.cluster.name);
        initRunaiUser && (await this.createRunaiContext(this.cluster.name));

        logger.debug(
            `Using cluster ${this.cluster.name} with contexts: admin -  ${JSON.stringify(
                this.adminContext
            )}, runai - ${JSON.stringify(this.runaiContext)}`
        );
        logger.debug(
            `KubeContext init, contexts: ${JSON.stringify(
                this.kubeClient.getContexts()
            )}, users: ${JSON.stringify(this.kubeClient.getUsers())}`
        );
    }

    protected async generateRunaiAuthenticatedUser(): Promise<k8s.User> {
        // plceholder for future implementation
        const tenantConfig = await tenantUtils.getTenantConfiguration(flowRootState.users.system_administrator.accessToken);
        const user: k8s.User = {
            name: objectGenerator.generateNameByPrefix('runai-auth'),
            authProvider: {
                name: 'oidc',
                config: {
                    'client-id': 'runai-cli',
                    airgapped: !tenantConfig.isMultiTenant,
                    'auth-flow': 'cli',
                    'idp-issuer-url': tenantConfig.authHost,
                    realm: tenantConfig.tenantName
                }
            }
        };
        logger.debug(`Created runai user: ${JSON.stringify(user)}`);
        this.kubeClient.addUser(user);
        return user;
    }

    protected async createRunaiContext(clusterName: string): Promise<void> {
        if (!settings.enableAuth) {
            this.runaiContext = this.getAdminContext(clusterName);
        }
        const contextName = objectGenerator.generateNameByPrefix('runai-context');
        let runaiUser: k8s.User;
        let runaiContext: k8s.Context;
        switch (settings.platform) {
            case supportedPlatforms.k8s:
                runaiUser = await this.createRunaiContextForK8s();
                runaiContext = {
                    name: contextName,
                    cluster: clusterName,
                    user: runaiUser.name
                };
                this.kubeClient.addContext(runaiContext);
                this.runaiContext = runaiContext;
                break;
            case supportedPlatforms.ocp:
                logger.debug('Skip ocp runai context creation, context creates dynamically');
                break;
        }
    }

    getCluster(clusterName?: string): k8s.Cluster {
        const clusters = this.kubeClient.getClusters();
        let cluster: k8s.Cluster;

        if (clusters.length === 0) {
            throw new Error('KubeContext error: No clusters found in kubeconfig');
        }
        if (clusters.length === 1) {
            cluster = clusters[0];
        } else {
            const contextClusterName = clusterName ? clusterName : config.clusterNamePrefix;
            cluster = clusters.find((cluster) => cluster.name.startsWith(contextClusterName));
        }
        if (!cluster) {
            throw new Error(`KubeContext error: Could not find cluster with name ${clusterName}`);
        }
        logger.debug(`Using kubeconfig cluster ${cluster.name}`);
        return cluster;
    }

    getCurrentContext(): string {
        return this.kubeClient.getCurrentContext();
    }

    getAdminContext(clusterName: string): k8s.Context {
        const contexts = this.kubeClient.getContexts();
        let adminContext: k8s.Context;
        const clusterContexts = contexts.filter((context) => context.cluster === clusterName);
        if (!clusterContexts) {
            throw new Error(`Could not find context for cluster ${clusterName}`);
        }
        // User with certificate authentication is assigned as admin.
        if (clusterContexts.length === 1) {
            adminContext = clusterContexts[0];
        } else {
            adminContext = clusterContexts.find((context) => this.kubeClient.getUser(context.user).certData);
        }
        logger.debug(`Found admin context ${JSON.stringify(adminContext)}`);
        if (!adminContext) {
            throw new Error(
                `Could not find admin context for cluster ${clusterName}, available contexts: ${JSON.stringify(
                    clusterContexts
                )}`
            );
        }
        return adminContext;
    }

    protected async createRunaiContextForK8s(): Promise<k8s.User> {
        return await this.generateRunaiAuthenticatedUser();
    }

    useAdminContext(): void {
        logger.debug('Using admin context');
        this.kubeClient.setCurrentContext(this.adminContext.name);
    }
}
