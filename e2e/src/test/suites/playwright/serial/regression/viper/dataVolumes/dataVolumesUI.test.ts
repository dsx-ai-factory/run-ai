/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    apiUtils,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { DataVolumePage } from '../../../../../../../modules/ui/runaiV2/pages/DataVolumePage';
import { V1ObjectMeta } from '@kubernetes/client-node';
import { kubeCli } from '../../../../../../../modules/cli';
import { ScopeType } from '../../../../../../../modules/api/apis/backendApi';

test.describe(`Data volume UI, tags: ${tags.tool.ui} ${tags.resource.datavolume} ${tags.teams.viper}`, function () {
    let project: Project;
    let projectName: string;
    let project2: Project;
    let project2Name: string;

    let existingPvc: V1ObjectMeta;
    let videoName: string;

    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let dataVolumePage: DataVolumePage;

    test.beforeAll('initialize browser engine', async () => {
        browserEngine = new BrowserEngine();
        await apiUtils.enableDataVolumes();

        projectName = objectGenerator.generateShortNameByPrefix('project-for-data-volume');
        project2Name = objectGenerator.generateShortNameByPrefix('project-to-share');
        [project, project2, existingPvc] = await Promise.all([
            projectResourcesUtils.createProjectAndNamespace(projectName),
            projectResourcesUtils.createProjectAndNamespace(project2Name),
            dataSourcesUtils.createExistingPVC(true)
        ]);
        logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
    });

    test.afterAll('disable data volumes FF', async ({}, testInfo) => {
        await Promise.all([
            apiUtils.disableDataVolumes(),
            kubeCli.CoreV1Api.deleteNamespacedPersistentVolumeClaim({
                name: existingPvc.name,
                namespace: existingPvc.namespace
            }),
            operations.deleteProjectAndNamespace(project.id),
            operations.deleteProjectAndNamespace(project2.id)
        ]);
    });

    test.beforeEach('login', async () => {
        await browserEngine.start();
        dataVolumePage = new DataVolumePage(browserEngine.currentPage);

        authPage = new AuthPage(browserEngine.currentPage);
        await authPage.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );

        videoName = 'data-volume-ui-';
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('Create, edit and delete data volume', async () => {
        videoName += 'create-edit-and-delete-data-volume';
        const dataVolumeName = objectGenerator.generateNameByPrefix('data-volume');
        await dataVolumePage.goToPage();
        // Create data volume
        await dataVolumePage.clickOnNewDataVolumeBtn();
        await dataVolumePage.fillNameInput(dataVolumeName);
        await dataVolumePage.selectOriginProjectScope(projectName);
        await dataVolumePage.selectOriginPvc(existingPvc.name);
        await dataVolumePage.selectSharedScope(ScopeType.Project, project2Name);
        await dataVolumePage.clickOnSubmitDataVolumeBtn();

        // Edit data volume
        await dataVolumePage.selectDataVolumeInTableByName(dataVolumeName);
        await dataVolumePage.clickOnEditDataVolumeBtn();
        await dataVolumePage.fillDescriptionInput('Data volume description');
        await dataVolumePage.selectSharedScope(ScopeType.Cluster, flowRootState.cluster.name);
        await dataVolumePage.clickOnSubmitDataVolumeBtn();

        // Delete data volume
        await dataVolumePage.selectDataVolumeInTableByName(dataVolumeName);
        await dataVolumePage.clickOnDeleteDataVolumeBtn();
        await dataVolumePage.clickOnDeleteDataVolumeConfirmBtn();
        logger.info(`Data volume deleted successfully: ${dataVolumeName}`);
    });
});
