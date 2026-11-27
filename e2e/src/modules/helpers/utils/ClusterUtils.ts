/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import compareVersions from 'semver-compare';
import { Cluster } from '../../api/apis/backendClusterApi';
import { authUtils, clusterService, clusterUtils, config, flowRootState, logger, waitUtils } from '..';
import { backendClusterApi, clusterApi } from '../../api';
import {
    ClusterDependencyStatus,
    ClusterDisplayedStatus,
    ClusterReportedConfig,
    ClustersMinimalData,
    DisplayedCluster,
    GetClustersMinimalData200Response
} from '../../api/apis/clusterApi';
import { bashCli } from '../../cli';
import http from 'http';

export class ClusterUtils {
    async list(accessToken = flowRootState.users.system_administrator.accessToken): Promise<Array<Cluster>> {
        backendClusterApi.clusterApi.accessToken = accessToken;
        const a = await backendClusterApi.clusterApi.v1K8sClustersGet();
        return a.body;
    }

    async getDisconnectedClusters(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<Cluster>> {
        backendClusterApi.clusterApi.accessToken = accessToken;
        const response = (await backendClusterApi.clusterApi.v1K8sClustersGet()).response;

        return response['body'].filter((cluster: any) => cluster.status?.state === 'Disconnected');
    }

    waitForClusterToBeDeleted(clusterName: string): Promise<void> {
        return waitUtils.waitForCondition(
            async () => {
                const clustersNames = (await this.list()).map((cluster) => cluster.name);
                return !clustersNames.includes(clusterName);
            },
            'Waiting for cluster to be deleted',
            true,
            10
        );
    }

    waitForClusterToBeCreated(clusterName: string): Promise<void> {
        return waitUtils.waitForCondition(
            async () => {
                const clustersNames = (await this.list()).map((cluster) => cluster.name);
                return clustersNames.includes(clusterName);
            },
            'Waiting for cluster to be created',
            true,
            10
        );
    }

    async deleteDisconnectedClusters(): Promise<void> {
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        backendClusterApi.clusterApi.accessToken = adminTokens.idToken;
        const disconnectedClusters = await clusterUtils.getDisconnectedClusters();
        for (const cluster of disconnectedClusters) {
            try {
                await backendClusterApi.clusterApi.v1K8sClustersUuidDelete(cluster.uuid);
                logger.info(`Successfully deleted cluster: ${cluster.name}`);
            } catch (e) {
                logger.info(
                    `Couldn't delete disconnected cluster with UUID: ${cluster.uuid}, Got error: ${e}`
                );
            }
        }
    }

    async isMoreThanOneClusterInEnvironment(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<boolean> {
        clusterApi.clustersApi.accessToken = accessToken;
        const clusters = (await clusterApi.clustersApi.getClusters()).body;
        return clusters.length > 1;
    }

    async createCluster(
        accessToken: string,
        name: string,
        domain: string,
        version: string
    ): Promise<DisplayedCluster> {
        clusterApi.clustersApi.accessToken = accessToken;
        const cluster = await clusterApi.clustersApi.createCluster({
            name: name,
            domain: domain,
            version: version
        });
        return cluster.body;
    }

    async deleteCluster(accessToken: string, clusterUuid: string): Promise<void> {
        clusterApi.clustersApi.accessToken = accessToken;
        await clusterApi.clustersApi.deleteCluster(clusterUuid);
    }

    async getClusterStatus(
        accessToken: string,
        clusterUuid: string
    ): Promise<ClusterDisplayedStatus['state']> {
        clusterApi.clustersApi.accessToken = accessToken;
        const cluster = await clusterApi.clustersApi.getClusters();
        return cluster.body.find((cluster) => cluster.uuid === clusterUuid).status.state;
    }

    async getClusterReportedConfig(
        accessToken?: string,
        clusterUuid?: string
    ): Promise<ClusterReportedConfig | null> {
        clusterApi.clustersApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        clusterUuid = clusterUuid ?? flowRootState.cluster.uuid;

        const cluster = await clusterApi.clustersApi.getClusters();
        return cluster.body.find((cluster) => cluster.uuid === clusterUuid).status.config;
    }

    async getRequiredDependency(
        name: string,
        accessToken?: string,
        clusterUuid?: string
    ): Promise<ClusterDependencyStatus | undefined> {
        clusterApi.clustersApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        clusterUuid = clusterUuid ?? flowRootState.cluster.uuid;

        const clusters = await clusterApi.clustersApi.getClusters();
        const cluster = clusters.body.find((candidate) => candidate.uuid === clusterUuid);
        return cluster?.status?.dependencies?.required?.[name];
    }

    // Whether the cluster as a whole supports CUDA fractions. When this is false,
    // cluster-sync suppresses per-node fractioning conditions so that nodes do not
    // each repeat a cluster-wide reason - so every node-level case depends on it.
    // Mirrors the fail-open default cluster-sync applies (ptr.Deref(Available, true)):
    // an absent component means supported, so only an explicit false suppresses.
    async isGpuFractioningSupported(): Promise<boolean> {
        const gpuStack = await clusterUtils.getRequiredDependency('gpu-stack');
        return gpuStack?.components?.['gpu-fractioning']?.available !== false;
    }

    getClusterPrefix(clusterName: string): string {
        return clusterName.replace(/-\d+$/, '');
    }

    /**
     * Returns the newest cluster version the control plane offers for installation.
     * cluster-service returns the latest patch per minor ({ '2.26': '2.26.50', ... }), already filtered by the
     * control plane's supported range, so the highest value is always installable on this environment.
     */
    async getLatestClusterVersion(accessToken: string): Promise<string> {
        const clusterVersions: Record<string, string> = await clusterService.getClusterVersions(accessToken);
        const versions = Object.values(clusterVersions ?? {});
        if (versions.length === 0) {
            throw new Error('Cluster service returned no installable cluster versions');
        }
        const latest = versions.sort(compareVersions).at(-1);
        logger.info(`Latest installable cluster version: ${latest}`);
        return latest;
    }

    async installCluster(
        accessToken: string,
        clusterName: string,
        domain: string,
        version: string,
        controlPlaneUrl: string
    ): Promise<DisplayedCluster> {
        const cluster = await this.createCluster(accessToken, clusterName, `${domain}:6443`, version);
        logger.info(`Cluster created: ${cluster.name}`);

        // New realms no longer have the shared runai-internal.installer client, so the installer authenticates
        // with the per-cluster bootstrap token instead of a client secret.
        const installationInfo = await clusterService.getClusterInfo(accessToken, cluster.uuid, version);
        if (!installationInfo.bootstrapToken) {
            throw new Error(`No bootstrap token returned for cluster ${cluster.name}`);
        }
        logger.info(`Installer bootstrap token retrieved for cluster: ${cluster.name}`);

        const helmRepoAddCommand =
            'helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update';
        bashCli.exec(helmRepoAddCommand);
        logger.info('Helm repo added successfully');

        const helmRepoUpdateCommand = 'helm repo update';
        bashCli.exec(helmRepoUpdateCommand);
        logger.info('Helm repo updated successfully');

        // Clean up any existing release
        try {
            bashCli.exec('helm uninstall runai-cluster -n runai');
            logger.info('Removed existing runai-cluster release');
        } catch (e) {
            logger.debug(`No existing runai-cluster release to remove: ${e}`);
        }

        const helmInstallCommand =
            'helm install runai-cluster runai/runai-cluster -n runai ' +
            `--set controlPlane.url=${controlPlaneUrl} ` +
            `--set controlPlane.bootstrapToken=${installationInfo.bootstrapToken} ` +
            `--set cluster.uid=${cluster.uuid} ` +
            `--set cluster.url=${cluster.domain} ` +
            `--version=${version} --create-namespace`;

        logger.info(`Installing cluster with command: ${helmInstallCommand}`);
        const result = await bashCli.asyncExec(helmInstallCommand);
        logger.info(`Helm install result: ${result}`);
        logger.info('Cluster installed successfully with helm');
        return cluster;
    }

    ///////////////// Cluster API Wrappers with Full Response /////////////////

    /**
     * Gets clusters minimal data and returns the full response including status code
     */
    async getClustersMinimalDataApi(
        filterBy?: Array<string>,
        offset?: number,
        limit?: number,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: GetClustersMinimalData200Response }> {
        clusterApi.clustersApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await clusterApi.clustersApi.getClustersMinimalData(filterBy, offset, limit);
    }

    /**
     * Gets cluster minimal data by ID and returns the full response including status code
     */
    async getClusterMinimalDataByIdApi(
        clusterUuid: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: ClustersMinimalData }> {
        clusterApi.clustersApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await clusterApi.clustersApi.getClusterMinimalDataById(clusterUuid);
    }
}
