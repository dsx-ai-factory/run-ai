/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    Utils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';

test.describe(`Create single node workloads (Training & Workspace), tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let projectName: string;
        let namespace: string;
        let project;
        let dataSourcePage: DataSourcePage;
        let dataSourceName: string;
        let videoName: string;
        const storageClass = 'standard';
        const accessMode = 'read-write by one node';
        const claimSize = '2';
        const claimUnit = 'GB';
        const volumeMode = 'Filesystem';
        const mountPath = '/home/local/data';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-datasource');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async () => {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            dataSourceName = objectGenerator.generateShortNameByPrefix('pvc-e2e-for-datasource');
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                await dataSourcePage.goToPage();
                await dataSourcePage.deleteDataSource(dataSourceName);
            } catch (e) {
                logger.error('Error deleting PVC:', dataSourceName, e);
            }
            await browserEngine.stopAndArchive(testInfo, 'dataSource-functional-testing');
        });

        test('T4060373: Creating new container with creating the existed name PVC', async () => {
            videoName = 'duplicate-pvc-name-validation';
            const expectedError = 'This name already exists';

            await test.step('Create the first PVC data source', async () => {
                await dataSourcePage.createPVC({
                    name: dataSourceName,
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    storageClass: storageClass,
                    accessMode: accessMode,
                    claimSize: claimSize,
                    claimUnit: claimUnit,
                    volumeMode: volumeMode,
                    containerPath: mountPath
                });
            });

            await test.step('Create another PVC data source with the same name', async () => {
                await dataSourcePage.goToPage();
                await dataSourcePage.page.pause();
                await dataSourcePage.clickOnAddNewDataSourceBtn();
                await dataSourcePage.clickOnAddNewPvcBtn();
                await dataSourcePage.assetsScopeSection.selectScope({
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });
                await dataSourcePage.fillPvcNameInput(dataSourceName);
            });

            await test.step('Verify the error message as expected', async () => {
                const errorFieldSelector = '.q-field--error.runai-name-validation';
                await expect(dataSourcePage.page.locator(errorFieldSelector)).toBeVisible();
                const fieldErrorSelector = '.runai-name-validation div[role="alert"]';
                const errorText = await dataSourcePage.page.locator(fieldErrorSelector).innerText();
                console.log('errorText:', errorText);
                expect(errorText).toContain(expectedError);
                await dataSourcePage.clickOnCancelBtn();
                await dataSourcePage.clickOnLeaveBtn();
            });
        });
    });
});
