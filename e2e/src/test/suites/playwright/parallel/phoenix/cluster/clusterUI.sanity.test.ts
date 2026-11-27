/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    clusterUtils,
    flowRootState,
    logger,
    objectGenerator,
    settings,
    tags
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';

test.describe(`Cluster UI, tags: ${tags.tool.ui} ${tags.resource.cluster} ${tags.teams.phoenix} ${tags.services.clusterService}`,
    { tag: [tags.services.clusterService] }, () => {
    // TODO: Restore the 30-second timeout once cluster connectivity checks complete reliably within it.
    test.use({ actionTimeout: 90 * 1000 });

    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let clusterPage: ClusterPage;
    let videoName: string;
    let clusterName: string;

    test.beforeAll(() => {
        browserEngine = new BrowserEngine();
    });

    test.beforeEach(async () => {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        clusterPage = new ClusterPage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    });

    test.afterEach(async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'cluster-' + videoName);
    });

    test('Should check that the cluster page exists', async () => {
        videoName = 'cluster-page-exists';
        await clusterPage.goToPage();
        const title = await clusterPage.getPageTitle();
        expect(clusterPage.url()).toContain('cluster');
        expect(title).toContain('Clusters');
        logger.debug('cluster page exists');
    });

    test.describe('Cluster creation', () => {
        async function fillFormSaas(): Promise<void> {
            await clusterPage.fillClusterName(clusterName);
            await clusterPage.selectClusterFirstVersion();
            await clusterPage.fillClusterUrl('https://test.com');
        }

        async function fillFormSelfHosted(): Promise<void> {
            await clusterPage.fillClusterName(clusterName);
            await clusterPage.selectClusterFirstVersion();
        }

        test.afterEach(async ({}, testInfo) => {
            if (clusterName) {
                try {
                    const clusters = await clusterUtils.list();
                    const cluster = clusters.find((c) => c.name === clusterName);
                    if (cluster) {
                        await clusterUtils.deleteCluster(
                            flowRootState.users.system_administrator.accessToken,
                            cluster.uuid
                        );
                        logger.info(`Cleanup: Deleted cluster ${clusterName} via API`);
                    }
                } catch {
                    logger.debug(`Cleanup: Cluster ${clusterName} not found or already deleted`);
                }
            }
        });

        test('should create cluster', async () => {
            videoName = 'create-cluster';
            clusterName = objectGenerator.generateShortNameByPrefix('cluster-for-e2e');
            // Moving to cluster creation page
            await clusterPage.goToPage();
            await clusterPage.clickOnAddNewClusterBtn();

            // Filling the form fields
            settings.isSelfHosted ? await fillFormSelfHosted() : await fillFormSaas();

            // Continue to the next step and leave the wizard
            await clusterPage.clickOnContinueButton();
            // Exiting the wizard and going back to the grid
            await clusterPage.clickOnDoneBtn();

            logger.info(`Cluster - ${clusterName} was created successfully`);

            // Selecting the cluster in the table
            const clusterNameCell = clusterPage.findByText('td', clusterName);
            await clusterNameCell.click();

            const firstRowStatusColumnText = await clusterPage.v2frame
                .locator('tr:nth-child(1) > td[aid="status"]')
                .innerText();
            expect(firstRowStatusColumnText.trim().toLowerCase()).toBe('waiting to connect');

            // Deleting the cluster
            await clusterPage.clickOnDeleteClusterBtn();
            await clusterPage.clickOnConfirmDeleteBtn();
            await clusterUtils.waitForClusterToBeDeleted(clusterName);
            logger.info(`Cluster - ${clusterName} was deleted successfully`);
        });
    });
});
