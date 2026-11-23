/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    clusterService,
    config,
    flowRootState,
    logger,
    tags
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';

test.describe(`Cluster Versions Installer UI, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.services.clusterService}`,
    { tag: [tags.services.clusterService] }, function () {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let clusterPage: ClusterPage;

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        clusterPage = new ClusterPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
    });

    test.afterAll(async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'clusterVersionsInstaller');
    });

    test('should check if version is supported on cluster installer', async function () {
        const supportedVersions = Object.keys(
            await clusterService.getClusterVersions(flowRootState.users.system_administrator.accessToken)
        );
        const installerVersions = await clusterPage.getClusterVersions();
        logger.debug(`Supported versions: ${supportedVersions}`);
        logger.debug(`Installer versions: ${installerVersions}`);
        expect(installerVersions).toEqual(expect.arrayContaining(supportedVersions));
    });
});
