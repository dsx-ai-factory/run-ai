/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    authUtils,
    clusterService,
    clusterUtils,
    flowRootState,
    settings,
    tags,
    logger
} from '../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { enhancedApiClusterDeletion, globalCleanupTestClusters } from '../clusterCleanup';

test.describe(`Cluster Service API, tags: ${tags.tool.api} ${tags.resource.cluster} ${tags.teams.phoenix} ${tags.services.clusterService}`,
    { tag: [tags.services.clusterService] }, () => {
    let createdClusters = [];

    test.beforeAll(async () => {
        await clusterUtils.deleteDisconnectedClusters();
    });

    test.afterAll(async () => {
        await clusterUtils.deleteDisconnectedClusters();
        // Global cleanup for any leftover test clusters
        await globalCleanupTestClusters();
    });

    test.afterEach(async () => {
        const cleanupErrors: string[] = [];

        // Enhanced cluster deletion with dual API strategy and error accumulation
        for (const cluster of createdClusters) {
            try {
                const deletionErrors = await enhancedApiClusterDeletion(cluster.uuid);
                cleanupErrors.push(...deletionErrors);
            } catch (error) {
                const errorMsg = `Unexpected error during cluster deletion for ${cluster.uuid}: ${error}`;
                logger.error(errorMsg);
                cleanupErrors.push(errorMsg);
            }
        }

        createdClusters = [];

        // Report accumulated errors if any occurred
        if (cleanupErrors.length > 0) {
            const combinedError = `Cleanup completed with ${cleanupErrors.length} error(s):\n${cleanupErrors.join('\n')}`;
            logger.error(combinedError);
            throw new Error(combinedError);
        }
    });

    test('should create a cluster', async () => {
        const clusterName = `test-cluster ${Date.now()}`;
        const cluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            clusterName,
            '',
            '1.0.0'
        );

        createdClusters.push(cluster);

        expect(cluster.name).toBe(clusterName);
        expect(cluster.domain).toBe('');
        expect(cluster.version).toBe('1.0.0');
    });

    test('should delete cluster', async () => {
        const deletedCluster = `test-cluster ${Date.now()}`;
        const cluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            deletedCluster,
            'test-domain',
            '1.0.0'
        );

        await clusterUtils.deleteCluster(flowRootState.users.system_administrator.accessToken, cluster.uuid);
    });

    test('should get cluster info', async () => {
        const cluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            `test-cluster ${Date.now()}`,
            'test-domain',
            'v2.100'
        );

        createdClusters.push(cluster);

        const clusterInfo = await clusterService.getClusterInfo(
            flowRootState.users.system_administrator.accessToken,
            flowRootState.cluster.uuid,
            cluster.version
        );
        expect(clusterInfo).toBeInstanceOf(Object);
        expect(
            clusterInfo.chartRepoURL.includes('https://runai') ||
                clusterInfo.chartRepoURL.includes('https://helm.ngc.nvidia.com/nvidia')
        ).toBe(true);
        expect(clusterInfo.clientSecret).not.toBe('');
        expect(clusterInfo.installationStr).toContain(cluster.version);
        expect(clusterInfo.repositoryName).toContain('runai');
    });

    test('should get versions list', async () => {
        const clusterVersions = await clusterService.getClusterVersions(
            flowRootState.users.system_administrator.accessToken
        );
        const numberDotNumberRegex = /^\d+\.\d+/;
        expect(clusterVersions).toBeInstanceOf(Object);
        for (const [key, value] of Object.entries(clusterVersions)) {
            expect(key).toMatch(numberDotNumberRegex);
            expect(value).toEqual(expect.any(String));
            expect(value).not.toBe('');
        }
    });

    test('should update cluster liveness', () => {
        shouldRunTests(settings.isSelfHosted)(async () => {
            const clusterBeforeUpdateLiveness = (
                await clusterService.getCluster(
                    flowRootState.users.system_administrator.accessToken,
                    flowRootState.cluster.uuid
                )
            ).lastLiveness;
            const clientSecret = await Keycloak.getClientSecretFromSpecificClient(
                'runai',
                `runai-internal.${flowRootState.cluster.uuid}.agent`
            );
            const token = await authUtils.getClusterServiceToken(
                `runai-internal.${flowRootState.cluster.uuid}.agent`,
                clientSecret
            );
            const req = await clusterService.updateClusterLiveness(token.idToken, flowRootState.cluster.uuid);
            const cluster = await clusterService.getCluster(
                flowRootState.users.system_administrator.accessToken,
                flowRootState.cluster.uuid
            );
            expect(cluster.lastLiveness).not.toEqual(clusterBeforeUpdateLiveness);
            expect(req).toBe(204);
        });
    });

    test('should update cluster info', () => {
        shouldRunTests(settings.isSelfHosted)(async () => {
            const clusterInfo = await clusterService.getCluster(
                flowRootState.users.system_administrator.accessToken,
                flowRootState.cluster.uuid
            );
            expect(clusterInfo.domain).toBe(flowRootState.cluster.domain);
            expect(clusterInfo.version).toBe(flowRootState.cluster.version);
            expect(clusterInfo.status).toBeDefined();
            expect(clusterInfo.status.conditions).toBeDefined();
            expect(clusterInfo.status.config).toBeDefined();
            expect(clusterInfo.status.operands).toBeDefined();
            expect(clusterInfo.status.platform).toBeDefined();
            expect(clusterInfo.status.state).toBeDefined();
            expect(clusterInfo.status.dependencies).toBeDefined();
            expect(clusterInfo.status.dependencies.optional).toBeDefined();
            expect(clusterInfo.status.dependencies.required).toBeDefined();
        });
    });

    test('should get clusters', async () => {
        // Get initial size of clusters list
        let clusters = await clusterService.getClusters(
            flowRootState.users.system_administrator.accessToken
        );
        const initialSize = clusters.length;

        // Create a new cluster to ensure there's at least one cluster
        const clusterName = `test-cluster ${Date.now()}`;
        const cluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            clusterName,
            '',
            '1.0.0'
        );

        createdClusters.push(cluster);

        await clusterUtils.waitForClusterToBeCreated(clusterName);

        clusters = await clusterService.getClusters(flowRootState.users.system_administrator.accessToken);
        expect(Array.isArray(clusters)).toBe(true);
        expect(clusters.length).toBe(initialSize + 1);

        // Assert that the created cluster is in the list
        let createdCluster = clusters.find((c) => c.name === clusterName);
        expect(createdCluster).toBeDefined();
        expect(createdCluster.name).toBe(clusterName);
        expect(createdCluster.version).toBe('1.0.0');

        // Create another cluster and verify the list contains both clusters
        const anotherClusterName = `test-cluster ${Date.now()}`;
        const anotherCluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            anotherClusterName,
            '',
            '2.0.0'
        );

        createdClusters.push(anotherCluster);

        await clusterUtils.waitForClusterToBeCreated(anotherClusterName);

        clusters = await clusterService.getClusters(flowRootState.users.system_administrator.accessToken);
        expect(Array.isArray(clusters)).toBe(true);
        expect(clusters.length).toBe(initialSize + 2);

        createdCluster = clusters.find((c) => c.name === clusterName);
        expect(createdCluster).toBeDefined();
        expect(createdCluster.name).toBe(clusterName);
        expect(createdCluster.version).toBe('1.0.0');

        const createdAnotherCluster = clusters.find((c) => c.name === anotherClusterName);
        expect(createdAnotherCluster).toBeDefined();
        expect(createdAnotherCluster.name).toBe(anotherClusterName);
        expect(createdAnotherCluster.version).toBe('2.0.0');
    });

    test('should update cluster', async () => {
        const clusterName = `test-cluster-update ${Date.now()}`;
        const cluster = await clusterUtils.createCluster(
            flowRootState.users.system_administrator.accessToken,
            clusterName,
            'test-domain',
            '1.0.0'
        );

        createdClusters.push(cluster);

        await clusterUtils.waitForClusterToBeCreated(clusterName);

        const updatedClusterName = `test-cluster-updated-name ${Date.now()}`;
        const updatedCluster = {
            ...cluster,
            name: updatedClusterName
        };

        const response = await clusterService.updateCluster(
            flowRootState.users.system_administrator.accessToken,
            cluster.uuid,
            updatedCluster
        );

        expect(response).toBe(204);

        const updatedClusterInfo = await clusterService.getCluster(
            flowRootState.users.system_administrator.accessToken,
            cluster.uuid
        );
        expect(updatedClusterInfo.name).toBe(updatedClusterName);
    });
});
