/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags } from '../../../../../../modules/helpers';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { EventHistoryPage } from '../../../../../../modules/ui/runaiV2/pages/EventHistoryPage';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { test } from '@playwright/test';
import { expect } from '@playwright/test';

test.describe(`Filtering Functionality Validation by UI, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, () => {
    
    let videoName: string;
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let cluster: ClusterPage;
    let eventHistoryPage: EventHistoryPage;
    let applicationsPage: ApplicationsPage;
    let users: UsersV2;

    test.beforeEach('start browser', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        cluster = new ClusterPage(browserEngine.currentPage);
        eventHistoryPage = new EventHistoryPage(browserEngine.currentPage);
        applicationsPage = new ApplicationsPage(browserEngine.currentPage);
        users = new UsersV2(browserEngine.currentPage);
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should filter by cluster status and validate other pages without errors', async function () {
        videoName = 'filter-pagination-cluster';
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        await cluster.goToPage();
        const filter = await cluster.ClusterFilterSelect('Status', 'Disconnected');
        expect(filter, 'cluster status filter chip should show Status: Disconnected').toEqual('Status:\nDisconnected');
        await users.navigateToUsers();
        const validationOfUserPageNoAlertOrErrors = await users.validateUserPage();
        expect(validationOfUserPageNoAlertOrErrors, 'users page should show no alerts or errors after applying cluster filter').toEqual(true);
    });

    test('should validate audit service access key filtering ui', async function () {
        videoName = 'audit-service-access-key-filtering-by-audit-type';
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        await eventHistoryPage.goToPage();
        const filterChipText = await eventHistoryPage.filterByStatus('Succeeded');
        expect(filterChipText, 'audit event filter chip should contain Status label').toContain('Status:');
        expect(filterChipText, 'audit event filter chip should contain Succeeded value').toContain('Succeeded');
        const isFilterValid = await eventHistoryPage.verifyFilterChip('Status:', 'Succeeded');
        expect(isFilterValid, 'Status: Succeeded filter chip should be verified as valid in the UI').toBe(true);
        await applicationsPage.navigateToServiceAccountsPage();
        await applicationsPage.page.waitForTimeout(2000);
        const serviceAccountFilterText = await applicationsPage.filterByServiceAccount('test');
        expect(serviceAccountFilterText, 'service account filter chip should contain Service account label').toContain('Service account:');
        expect(serviceAccountFilterText, 'service account filter chip should contain the filtered value').toContain('test');
        const isServiceAccountFilterValid = await applicationsPage.verifyServiceAccountFilterChip(
            'Service account:',
            'test'
        );
        expect(isServiceAccountFilterValid, 'Service account: test filter chip should be verified as valid in the UI').toBe(true);
    });
});
