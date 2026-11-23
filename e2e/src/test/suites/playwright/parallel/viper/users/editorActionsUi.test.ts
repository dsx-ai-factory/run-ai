/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils
} from '../../../../../../modules/helpers';
import { isOCP } from '../../../../../../modules/helpers/utils/EnvironmentUtils';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

test.describe(`Editor Actions, tags: ${tags.tool.ui} ${tags.resource.users} ${tags.teams.viper} `, () => {
    shouldRunTests(!isOCP())(function () {
        let projectName: string;
        let namespace: string;
        let videoName: string;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let environment: EnvironmentPage;
        let dataSource: DataSourcePage;
        let computeResource: ComputeResourcePage;
        let project: Project;

        test.beforeAll(async function () {
            if (isOCP()) {
                logger.info('OCP cluster, skipping test');
                this.skip();
            }

            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateNameByPrefix('project-e2e-for-editor-actions');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            logger.info(`created project ${projectName}`);
        });
        test.afterAll(async function () {
            if (!isOCP()) {
                await operations.deleteProjectAndNamespace(project.id);
            }
        });
        test.beforeEach('start browser', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.editor.email,
                flowRootState.users.editor.password
            );
        });
        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });
        test('can create workload', async function () {
            videoName = 'editor-cannot-create-workload';
            workload = new WorkloadPage(browserEngine.currentPage);
            await workload.navigateToWorkloads();
            await workload.validateNewWorkloadButtonExist();
        });
        test('can create environments', async function () {
            videoName = 'editor-can-create-environment';
            environment = new EnvironmentPage(browserEngine.currentPage);
            await environment.goToPage();
            await environment.validateNewEnvironmentButtonExist();
        });
        test('can create compute resources', async function () {
            videoName = 'editor-can-create-compute-resources';
            computeResource = new ComputeResourcePage(browserEngine.currentPage);
            await computeResource.goToPage();
            await computeResource.validateNewComputeResourceExist();
        });
        test('can create data sources', async function () {
            videoName = 'editor-can-create-data-source';
            dataSource = new DataSourcePage(browserEngine.currentPage);
            await dataSource.goToPage();
            await dataSource.validateNewDataSourceButtonExist();
        });
    });
});
