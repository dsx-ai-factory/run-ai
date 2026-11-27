/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    flowRootState,
    objectGenerator,
    settings,
    tags,
    logger
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    enhancedClusterDeletion,
    expectEmptyOrEmDashCell,
    globalCleanupConnectivityClusters
} from '../clusterCleanup';

// Only run if version and environment are supported
(supportedFrom as any)(SupportedVersions.V2_17)(() => {
    if (!settings.isSelfHosted) return;

    test.describe(`Cluster Connectivity UI, tags: ${tags.tool.ui} ${tags.resource.cluster} ${tags.teams.phoenix} ${tags.services.clusterService}`,
        { tag: [tags.services.clusterService] }, () => {
        // TODO: Restore the 30-second timeout once cluster connectivity checks complete reliably within it.
        test.use({ actionTimeout: 90 * 1000 });

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let clusterPage: ClusterPage;
        let clusterName: string;

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            clusterPage = new ClusterPage(browserEngine.currentPage);
            await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
            clusterName = objectGenerator.generateShortNameByPrefix('connectivity-e2e');
            await clusterPage.createClusterByUI(clusterName);
        });

        test.afterEach(async ({}, testInfo) => {
            const cleanupErrors: string[] = [];
            try {
                const deletionErrors = await enhancedClusterDeletion(clusterPage, clusterName);
                cleanupErrors.push(...deletionErrors);
            } catch (error) {
                const errorMsg = `Unexpected error during cluster deletion: ${error}`;
                logger.error(errorMsg);
                cleanupErrors.push(errorMsg);
            } finally {
                await browserEngine.stopAndArchive(testInfo, 'clusterConnectivity');
            }

            if (cleanupErrors.length > 0) {
                const combinedError = `Cleanup completed with ${cleanupErrors.length} error(s):\n${cleanupErrors.join('\n')}`;
                logger.error(combinedError);
                throw new Error(combinedError);
            }
        });

        // Global cleanup hook to handle any leftover connectivity test clusters
        test.afterAll(async () => {
            await globalCleanupConnectivityClusters();
        });

        test('Should check 2 clusters with different connectivity states', async () => {
            const clusterConnectivityData = await clusterPage.getClusterTableData();

            // First cluster (newly created)
            expectEmptyOrEmDashCell(clusterConnectivityData[0][0]);
            expect(clusterConnectivityData[0][1]).toContain('connectivity-e2e');
            expect(clusterConnectivityData[0][2]).toContain('-');
            expect(clusterConnectivityData[0][3]).toContain('-');
            expect(clusterConnectivityData[0][4].trim().toLowerCase()).toBe('waiting to connect');
            expect(clusterConnectivityData[0][5]).toBe('Never');
            expect(clusterConnectivityData[0][6]).toContain('2');
            expectEmptyOrEmDashCell(clusterConnectivityData[0][7]);
            expectEmptyOrEmDashCell(clusterConnectivityData[0][8]);

            // Second cluster (existing)
            expectEmptyOrEmDashCell(clusterConnectivityData[1][0]);
            expect(clusterConnectivityData[1][1]).toBe(flowRootState.cluster.name);
            expect(clusterConnectivityData[1][2]).not.toBe('');
            expect(clusterConnectivityData[1][3]).toContain('1');
            expect(clusterConnectivityData[1][4].trim().toLowerCase()).toContain('connected');
            expect(clusterConnectivityData[1][5].trim().toLowerCase()).toContain('now');
            expect(clusterConnectivityData[1][6]).toContain('2');
            expect(clusterConnectivityData[1][7]).toBe(config.tenantUrl);
            expect(clusterConnectivityData[1][8]).not.toBe('');
        });
    });
});
