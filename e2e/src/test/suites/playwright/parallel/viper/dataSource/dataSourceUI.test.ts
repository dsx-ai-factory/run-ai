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
    tags
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { AssetKind, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { test } from '@playwright/test';

test.describe(
    `Data source UI, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.viper}`,
    { tag: [tags.services.frontend] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let dataSource: DataSourcePage;

        let videoName: string;
        let projectName: string;
        let project: Project;

        const paramsForCopAndEditTest = [
            {
                type: AssetKind.Pvc,
                createMethod: 'createPvc',
                input: 'fillPvcNameInput',
                dataSourceLowerCase: 'pvc'
            },
            {
                type: AssetKind.Git,
                createMethod: 'createGit',
                input: 'fillGitNameInput',
                dataSourceLowerCase: 'git'
            },
            {
                type: AssetKind.Nfs,
                createMethod: 'createNFS',
                input: 'fillNFSNameInput',
                dataSourceLowerCase: 'nfs'
            },
            {
                type: AssetKind.HostPath,
                createMethod: 'createHostPath',
                input: 'fillHostPathNameInput',
                dataSourceLowerCase: 'host-path'
            },
            {
                type: AssetKind.S3,
                createMethod: 'createS3DataSource',
                input: 'fillS3NameInput',
                dataSourceLowerCase: 's3'
            }
        ];

        const paramsForRenameTest = {
            [AssetKind.Pvc]: {
                type: AssetKind.Pvc,
                name: objectGenerator.generateShortNameByPrefix('pvc'),
                payloadFunc: dataSourcesUtils.getPvcCreationPayload.bind(dataSourcesUtils),
                createFunc: dataSourcesUtils.createPvcAsset
            },
            [AssetKind.Git]: {
                type: AssetKind.Git,
                name: objectGenerator.generateShortNameByPrefix('git'),
                payloadFunc: dataSourcesUtils.getGitCreationPayload.bind(dataSourcesUtils),
                createFunc: dataSourcesUtils.createGitAsset
            },
            [AssetKind.Nfs]: {
                type: AssetKind.Nfs,
                name: objectGenerator.generateShortNameByPrefix('nfs'),
                payloadFunc: dataSourcesUtils.getNfsCreationPayload.bind(dataSourcesUtils),
                createFunc: dataSourcesUtils.createNfsAsset
            },
            [AssetKind.HostPath]: {
                type: AssetKind.HostPath,
                name: objectGenerator.generateShortNameByPrefix('host-path'),
                payloadFunc: dataSourcesUtils.getHostPathCreationPayload.bind(dataSourcesUtils),
                createFunc: dataSourcesUtils.createHostPathAsset
            },
            [AssetKind.S3]: {
                type: AssetKind.S3,
                name: objectGenerator.generateShortNameByPrefix('s3'),
                payloadFunc: dataSourcesUtils.getS3CreationPayload.bind(dataSourcesUtils),
                createFunc: dataSourcesUtils.createS3Asset
            }
        };

        async function initBrowserAndLogin(): Promise<void> {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        }

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();
            project = await projectResourcesUtils.createProjectAndNamespace('project-for-e2e');
            projectName = project.name;
        });

        test.afterAll(async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName} and namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName} and namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach(async () => {
            await initBrowserAndLogin();
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'data-source-' + videoName);
        });

        paramsForCopAndEditTest.forEach(({ type, createMethod, input, dataSourceLowerCase }) => {
            test(`${type} - create, copy & edit ${tags.runOn.parallelCI}`, async () => {
                videoName = `copy-and-edit-${dataSourceLowerCase}`;
                const dataSourceName = objectGenerator.generateShortNameByPrefix(dataSourceLowerCase);
                const copiedDataSourceName = objectGenerator.generateShortNameByPrefix(
                    `copy-${dataSourceName}`
                );

                await dataSource[createMethod](dataSourceName, {
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });

                await dataSource.openCopyAndEditForm(dataSourceName);
                await dataSource[input](copiedDataSourceName);

                if (type === AssetKind.S3) {
                    await dataSource.selectS3CredentialsByName('None (for public buckets)');
                }

                await dataSource.waitForNameAlreadyExistsErrorToDisappear();
                await dataSource.clickOnSubmitBtn();

                await dataSource.deleteDataSource(copiedDataSourceName);
                await dataSource.deleteDataSource(dataSourceName);
            });
        });

        test('Should rename all the data sources', async () => {
            videoName = 'rename-all-data-sources';
            await dataSource.goToPage();

            const dataSourceCreationPrms = Object.values(paramsForRenameTest).map(
                ({ name, payloadFunc, createFunc }) =>
                    createFunc(
                        payloadFunc({
                            name,
                            scope: Scope.Project,
                            scopeId: project.id
                        })
                    )
            );

            await Promise.all(dataSourceCreationPrms);

            for (const { name } of Object.values(paramsForRenameTest)) {
                const newDataSourceName = `${name}-new-name`;
                await dataSource.goToPage();
                await dataSource.renameDataSource(name, newDataSourceName);
            }
        });
    }
);
