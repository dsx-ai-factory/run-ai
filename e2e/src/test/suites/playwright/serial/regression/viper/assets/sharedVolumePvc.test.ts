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
    waitUtils
} from '../../../../../../../modules/helpers';

import { test } from '@playwright/test';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { assetsApi, k8sObjectTrackerApi } from '../../../../../../../modules/api';
import { Scope } from '../../../../../../../modules/api/apis/assetsApi';
import { StorageClassConfigurationCreationRequest } from '../../../../../../../modules/api/apis/k8sObjectsTrackerApi/generated/api';
import ScopeTypeEnum = StorageClassConfigurationCreationRequest.ScopeTypeEnum;
import {
    Platform,
    supportedFrom,
    supportedPlatforms,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Shared Volume PVC, tags:  ${tags.resource.datasource} ${tags.teams.viper}`, { tag: [tags.services.k8sObjectsTracker, tags.services.assets, tags.runOn.serialCI] }, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        supportedPlatforms([Platform.GCP])(function () {
            let browserEngine;
            let projectName;
            let project2Name;
            let project;
            let project2;

            let auth: AuthPage;
            let dataSource: DataSourcePage;

            test.beforeAll('initialize browser engine', async () => {
                browserEngine = new BrowserEngine();

                projectName = objectGenerator.generateShortNameByPrefix('project-1');
                project2Name = objectGenerator.generateShortNameByPrefix('project-2');
                [project, project2] = await Promise.all([
                    projectResourcesUtils.createProjectAndNamespace(projectName),
                    projectResourcesUtils.createProjectAndNamespace(project2Name)
                ]);
                logger.info(`created project ${project.name} ,namespace: "${project.namespace}"`);
            });

            test.afterAll('delete projects + cleanup', async ({}, testInfo) => {
                await Promise.all([
                    operations.deleteProjectAndNamespace(project.id),
                    operations.deleteProjectAndNamespace(project2.id)
                ]);
            });

            test.beforeEach(async () => {
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                dataSource = new DataSourcePage(browserEngine.currentPage);
                await auth.signIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
            });

            test.afterEach(async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'shared-volume-pvc');
            });

            test('set storage class to support data sharing', async () => {
                k8sObjectTrackerApi.storageClassConfigurationApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                assetsApi.pVCApi.accessToken = flowRootState.users.system_administrator.accessToken;

                const immediateStorageClassName = 'standard';

                const res =
                    await k8sObjectTrackerApi.storageClassConfigurationApi.listStorageClassConfiguration([
                        `name==${immediateStorageClassName}`,
                        `scopeId==${flowRootState.cluster.uuid}`
                    ]);
                if (res.body.storageClassConfigurations.length) {
                    // If the storage class configuration already exists we patch it
                    await k8sObjectTrackerApi.storageClassConfigurationApi.patchStorageClassConfigurationById(
                        res.body.storageClassConfigurations[0].id,
                        {
                            permissions: {
                                allowDataSharing: true
                            }
                        }
                    );
                } else {
                    // If the storage class configuration doesn't exist we create it
                    await k8sObjectTrackerApi.storageClassConfigurationApi.createStorageClassConfiguration({
                        name: immediateStorageClassName,
                        scopeType: ScopeTypeEnum.Cluster,
                        scopeId: flowRootState.cluster.uuid,
                        permissions: {
                            allowDataSharing: true
                        }
                    });
                }

                const pvcName = objectGenerator.generateShortNameByPrefix('shared-pvc');

                // create pvc
                await dataSource.goToPage();
                await dataSource.clickOnAddNewDataSourceBtn();
                await dataSource.clickOnAddNewPvcBtn();
                await dataSource.selectCurrentCluster();
                await dataSource.assetsScopeSection.selectScope({
                    scope: Scope.Cluster,
                    scopeEntityName: flowRootState.cluster.name
                });
                await dataSource.selectPvcVolumeSharingOption('Shared volume between all projects ');
                await dataSource.fillPvcNameInput(pvcName);
                await dataSource.selectOptionByExactValue(
                    dataSource.PvcStorageClassSelect,
                    immediateStorageClassName
                );
                await dataSource.selectPvcAccessMode('Read-write by one node');
                await dataSource.selectPvcVolumeMode('Filesystem');
                await dataSource.fillPvcContainerInput('/home/local/data');
                await dataSource.clickOnSubmitBtn();
                await dataSource.selectDataSourceInTableByName(pvcName);

                const pvcAsset = (await assetsApi.pVCApi.listPvcAssets()).body.entries.find(
                    (a) => a.meta.name === pvcName
                );
                const claimName = pvcAsset.spec.claimName;

                await Promise.all([
                    waitUtils.waitForPvc(claimName, 'runai'),
                    waitUtils.waitForPvc(claimName, project.namespace),
                    waitUtils.waitForPvc(claimName, project2.namespace)
                ]);
                logger.info('PVCs were created in all the expected namespaces');

                // delete pvc
                await dataSource.clickOnDeleteDataSourceBtn();
                await dataSource.clickOnConfirmDeleteBtn();

                await Promise.all([
                    waitUtils.waitForPvc(claimName, 'runai', undefined, false),
                    waitUtils.waitForPvc(claimName, project.namespace, undefined, false),
                    waitUtils.waitForPvc(claimName, project2.namespace, undefined, false)
                ]);
                logger.info('PVCs were deleted in all the expected namespaces');
            });
        });
    });
});
