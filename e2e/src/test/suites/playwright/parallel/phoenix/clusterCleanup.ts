/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from '@playwright/test';
import {
    authUtils,
    clusterUtils,
    config,
    flowRootState,
    logger
} from '../../../../../modules/helpers';
import { ClusterPage } from '../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { clusterApi, backendClusterApi } from '../../../../../modules/api';

/** UI may show an empty cell or an em dash (—) for missing values. */
export const expectEmptyOrEmDashCell = (value: string): void => {
    const t = value.trim();
    expect(t === '' || t === '—').toBe(true);
};

export const getClusterUuidByName = async (clusterName: string): Promise<string | null> => {
    try {
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        clusterApi.clustersApi.accessToken = adminTokens.idToken;
        const clusters = await clusterApi.clustersApi.getClusters();
        const cluster = clusters.body.find((c) => c.name === clusterName);
        return cluster ? cluster.uuid : null;
    } catch (error) {
        logger.error(`Failed to find cluster UUID for ${clusterName}:`, error);
        return null;
    }
};

/** Try UI deletion first, then cluster service API, then backend API as fallbacks. */
export const enhancedClusterDeletion = async (
    clusterPage: ClusterPage,
    clusterName: string
): Promise<string[]> => {
    const errors: string[] = [];
    let uiDeletionSucceeded = false;

    try {
        logger.info(`Attempting UI deletion for cluster: ${clusterName}`);
        await clusterPage.deleteClusterByUI(clusterName);
        uiDeletionSucceeded = true;
        logger.info(`UI deletion succeeded for cluster: ${clusterName}`);
    } catch (uiError) {
        const uiErrorMsg = `UI deletion failed for ${clusterName}: ${uiError}`;
        logger.warn(uiErrorMsg);
        errors.push(uiErrorMsg);
    }

    if (uiDeletionSucceeded) {
        return errors;
    }

    const clusterUuid = await getClusterUuidByName(clusterName);
    if (!clusterUuid) {
        const noUuidMsg = `Could not find UUID for cluster ${clusterName}, skipping API deletion`;
        logger.error(noUuidMsg);
        errors.push(noUuidMsg);
        return errors;
    }

    try {
        logger.info(
            `Attempting cluster service API deletion for cluster: ${clusterName} (UUID: ${clusterUuid})`
        );
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        await clusterUtils.deleteCluster(adminTokens.idToken, clusterUuid);
        logger.info(`Cluster service API deletion succeeded for cluster: ${clusterName}`);
        return errors;
    } catch (clusterApiError) {
        const clusterApiErrorMsg = `Cluster service API deletion failed for ${clusterName}: ${clusterApiError}`;
        logger.warn(clusterApiErrorMsg);
        errors.push(clusterApiErrorMsg);
    }

    try {
        logger.info(
            `Attempting backend API deletion for cluster: ${clusterName} (UUID: ${clusterUuid})`
        );
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        backendClusterApi.clusterApi.accessToken = adminTokens.idToken;
        await backendClusterApi.clusterApi.v1K8sClustersUuidDelete(clusterUuid);
        logger.info(`Backend API deletion succeeded for cluster: ${clusterName}`);
    } catch (backendApiError) {
        const backendApiErrorMsg = `Backend API deletion failed for ${clusterName}: ${backendApiError}`;
        logger.error(backendApiErrorMsg);
        errors.push(backendApiErrorMsg);
    }

    return errors;
};

export const globalCleanupConnectivityClusters = async (): Promise<void> => {
    try {
        logger.info('Running global cleanup for connectivity-e2e clusters');
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        clusterApi.clustersApi.accessToken = adminTokens.idToken;
        const clusters = await clusterApi.clustersApi.getClusters();

        const connectivityClusters = clusters.body.filter(
            (cluster) => cluster.name && cluster.name.includes('connectivity-e2e')
        );

        if (connectivityClusters.length === 0) {
            logger.info('No connectivity test clusters found for cleanup');
            return;
        }

        logger.info(`Found ${connectivityClusters.length} connectivity test clusters to clean up`);
        for (const cluster of connectivityClusters) {
            try {
                await clusterUtils.deleteCluster(adminTokens.idToken, cluster.uuid);
                logger.info(`Global cleanup: deleted cluster ${cluster.name} (UUID: ${cluster.uuid})`);
            } catch (error) {
                logger.warn(`Global cleanup: failed to delete cluster ${cluster.name}: ${error}`);
            }
        }
    } catch (error) {
        logger.error('Global cleanup failed:', error);
    }
};

/** Try cluster service API deletion first, then backend API as fallback. */
export const enhancedApiClusterDeletion = async (clusterUuid: string): Promise<string[]> => {
    const errors: string[] = [];
    const adminToken = flowRootState.users.system_administrator.accessToken;

    try {
        logger.info(`Attempting cluster service API deletion for cluster UUID: ${clusterUuid}`);
        await clusterUtils.deleteCluster(adminToken, clusterUuid);
        logger.info(`Cluster service API deletion succeeded for UUID: ${clusterUuid}`);
        return errors;
    } catch (clusterApiError) {
        const clusterApiErrorMsg = `Cluster service API deletion failed for ${clusterUuid}: ${clusterApiError}`;
        logger.warn(clusterApiErrorMsg);
        errors.push(clusterApiErrorMsg);
    }

    try {
        logger.info(`Attempting backend API deletion fallback for cluster UUID: ${clusterUuid}`);
        backendClusterApi.clusterApi.accessToken = adminToken;
        await backendClusterApi.clusterApi.v1K8sClustersUuidDelete(clusterUuid);
        logger.info(`Backend API deletion succeeded for UUID: ${clusterUuid}`);
    } catch (backendApiError) {
        const backendApiErrorMsg = `Backend API deletion failed for ${clusterUuid}: ${backendApiError}`;
        logger.error(backendApiErrorMsg);
        errors.push(backendApiErrorMsg);
    }

    return errors;
};

export const globalCleanupTestClusters = async (): Promise<void> => {
    try {
        logger.info('Running global cleanup for test clusters');
        const adminToken = flowRootState.users.system_administrator.accessToken;
        clusterApi.clustersApi.accessToken = adminToken;
        const clusters = await clusterApi.clustersApi.getClusters();

        const testClusters = clusters.body.filter(
            (cluster) => cluster.name && cluster.name.includes('test-cluster')
        );

        if (testClusters.length === 0) {
            logger.info('No test clusters found for cleanup');
            return;
        }

        logger.info(`Found ${testClusters.length} test clusters to clean up`);
        for (const cluster of testClusters) {
            try {
                await clusterUtils.deleteCluster(adminToken, cluster.uuid);
                logger.info(`Global cleanup: deleted cluster ${cluster.name} (UUID: ${cluster.uuid})`);
            } catch (error) {
                logger.warn(`Global cleanup: failed to delete cluster ${cluster.name}: ${error}`);
            }
        }
    } catch (error) {
        logger.error('Global cleanup failed:', error);
    }
};
