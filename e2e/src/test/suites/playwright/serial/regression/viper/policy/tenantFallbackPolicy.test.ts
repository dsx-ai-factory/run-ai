/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { assetsApi, policyApi } from '../../../../../../../modules/api';
import { FallbackPolicyChangeRequest, ScopeType } from '../../../../../../../modules/api/apis/policyApi';
import {
    dataSourcesUtils,
    flowRootState,
    objectGenerator,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { DataSourcePage } from '../../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { PVCAsset, Scope } from '../../../../../../../modules/api/apis/assetsApi';

test.describe(
    `Should block the users from creating certain resources ${tags.teams.viper}`,
    { tag: [tags.services.policy, tags.runOn.serialCI] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let dataSource: DataSourcePage;
        let pvc: PVCAsset;

        let videoName: string;
        test.beforeAll(async () => {
            policyApi.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApi.pVCApi.accessToken = flowRootState.users.system_administrator.accessToken;
            pvc = await dataSourcesUtils.createPVC(
                objectGenerator.generateShortNameByPrefix('pvc'),
                Scope.Cluster,
                undefined
            );
            await waitUtils.waitForPvc(pvc.meta.name, 'runai', 90, true);
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterAll(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            await policyApi.policyApi.deleteFallbackPolicy('tenant');
            await dataSourcesUtils.deletePvcAsset(pvc.meta.id);
        });

        test('should hide certain data source types depending on fallback policy', async () => {
            videoName = 'tenant-fallback-policy';
            const fallbackPolicyChangeRequest: FallbackPolicyChangeRequest = {
                meta: {
                    scope: ScopeType.Tenant,
                    name: 'tenant-fallback'
                },
                policy: {
                    rules: {
                        storage: {
                            pvc: {
                                instances: {
                                    canAdd: false
                                }
                            }
                        }
                    },
                    defaults: {
                        storage: {
                            pvc: {
                                instances: [{ name: pvc.meta.name, path: '/mnt/pvc', existingPvc: true }]
                            }
                        }
                    }
                }
            };
            await policyApi.policyApi.updateFallbackPolicy(false, false, fallbackPolicyChangeRequest);

            await dataSource.goToPage();
            await dataSource.clickOnAddNewDataSourceBtn();
            const nfsBtn = dataSource.v2frame.locator('[aid="NFS-btn"]');
            const pvcBtn = dataSource.v2frame.locator('[aid="PVC-btn"]');
            expect(await nfsBtn.isVisible()).to.be.true;
            expect(await pvcBtn.isVisible()).to.be.false;
        });
    }
);
