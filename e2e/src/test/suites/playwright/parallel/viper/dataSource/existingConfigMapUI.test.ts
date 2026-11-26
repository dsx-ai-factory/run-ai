/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags
} from '../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { V1ObjectMeta } from '@kubernetes/client-node/';
import { kubeCli } from '../../../../../../modules/cli';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { expect, test } from '@playwright/test';

test.describe(
    `Existing ConfigMap UI, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`,
    { tag: [tags.services.k8sObjectsTracker, tags.services.assets, tags.runOn.parallelCI] },
    function () {
        let project: Project;
        let projectB: Project;
        let browserEngine: BrowserEngine;
        let videoName: string;
        let auth: AuthPage;
        let dataSource: DataSourcePage;
        let configMapMeta: V1ObjectMeta;
        let nameOfNewConfigMap: string;

        async function deleteProjectAndNamespace(projectToDelete: Project): Promise<void> {
            try {
                await operations.deleteProjectAndNamespace(projectToDelete.id);
                logger.info(
                    `deleted project ${projectToDelete.name} and namespace ${projectToDelete.namespace}`
                );
            } catch (error) {
                logger.error(
                    `failed to delete project ${projectToDelete.name} and namespace ${projectToDelete.namespace}`
                );
                throw error;
            }
        }

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            [project, projectB] = await Promise.all([
                projectResourcesUtils.createProjectAndNamespace('p-e2e-for-cred'),
                projectResourcesUtils.createProjectAndNamespace('p-b')
            ]);

            await projectUtils.waitForRoleBindingForUserInProject(
                project.namespace,
                flowRootState.users.system_administrator.email,
                projectUtils.newRbacRunaiSystemAdministratorRoleBindingName
            );
        });
        test.afterAll(async () => {
            await Promise.all([project, projectB].map(deleteProjectAndNamespace));
        });
        test.beforeEach(async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            nameOfNewConfigMap = objectGenerator.generateNameByPrefix('config-map');
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await kubeCli.deleteConfigMap(configMapMeta);
            } finally {
                await browserEngine.stopAndArchive(testInfo, videoName);
            }
        });

        test.describe('config-map with project namespace', function () {
            test('create and delete project scoped config-map', async function () {
                videoName = 'project-scoped-config-map';
                configMapMeta = await dataSourcesUtils.createConfigMapResource('cm-a', project.namespace);
                const configMapsInOtherProject = await dataSourcesUtils.getConfigMapInfo(
                    configMapMeta.name,
                    Scope.Project.toString(),
                    projectB.id.toString()
                );
                expect(
                    configMapsInOtherProject,
                    `ConfigMap ${configMapMeta.name} should not be visible in project ${projectB.id}`
                ).toEqual([]);
                await dataSource.goToPage();
                await dataSource.createConfigMapAsset(
                    nameOfNewConfigMap,
                    Scope.Project,
                    project.name,
                    configMapMeta.name,
                    '/var/path'
                );
                await dataSource.deleteDataSource(nameOfNewConfigMap);
            });
        });
    }
);
