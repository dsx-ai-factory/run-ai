/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    authUtils,
    usersManager
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';

test.describe(`Access Rules UI, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.accessRules}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let trainingPage: TrainingPage;
        let loginUIPage: LoginUi;

        let videoName: string;
        let projectName1: string;
        let projectName2: string;
        let project1: Project;
        let project2: Project;
        let trainingName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName1 = objectGenerator.generateShortNameByPrefix('project1-e2e-for-access-rule');
            projectName2 = objectGenerator.generateShortNameByPrefix('project2-e2e-for-access-rule');

            [project1, project2] = await Promise.all([
                projectResourcesUtils.createProject(`runai-${projectName1}`, {
                    name: projectName1,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                }),
                projectResourcesUtils.createProject(`runai-${projectName2}`, {
                    name: projectName2,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                })
            ]);
        });

        test.afterAll('deleting project and namespace', async () => {
            for (const project of [project1, project2]) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${project.id}`);
                } catch (error) {
                    logger.error(`failed to delete project ${project.id}`);
                    throw error;
                }
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            loginUIPage = new LoginUi(browserEngine.currentPage);
            trainingName = objectGenerator.generateShortNameByPrefix('training');
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.describe('administrator user assigns multiple access rules to a user', function () {
            test('T3795506: assign multiple L1 researcher role with different project scopes', async function () {
                test.setTimeout(240 * 1000);
                videoName = 'admin-assign-multi-access-rules';
                const cardsLocator =
                    '//div[@aid="project-section"]//div[@class="project-card workload-card-content"]';

                let user: UserPostResponse;
                const email = objectGenerator.generateEmailByPrefixWithRandom('ar-user', 5);

                const adminTokens = await authUtils.authenticateUser(
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword
                );

                await test.step('create a user', async () => {
                    user = (await usersManager.createLocalUser(adminTokens.idToken, email)).body;
                });

                await test.step('bind the user with L1 researcher role with project1 scope', async () => {
                    await authorizationUtils.bindUserWithRoles(
                        user,
                        [PROPERTY_TO_ROLE_MAPPING.l1_researcher],
                        {
                            scopeId: String(project1.id),
                            scopeType: ScopeType.Project
                        },
                        adminTokens.idToken
                    );
                });

                await test.step('logout and login with created user', async () => {
                    await authPage.signIn(user.username, user.tempPassword);
                });

                await test.step(`user can only see project ${projectName1} when creating workload`, async () => {
                    await trainingPage.goToPage();
                    await trainingPage.openTrainingForm();
                    await trainingPage.page.waitForSelector(cardsLocator, {
                        state: 'visible',
                        timeout: 10000
                    });
                    const projectCardList = trainingPage.locator(cardsLocator);

                    expect(await projectCardList.count()).toEqual(1);
                    expect(await projectCardList.innerText()).toContain(projectName1);
                    await trainingPage.closeWorkloadForm();
                });

                await test.step('user can create training sucessfully', async () => {
                    await trainingPage.createFlexibleTraining(projectName1, trainingName);
                    await trainingPage.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                });

                await test.step('delete training', async () => {
                    await trainingPage.deleteWorkload(trainingName);
                });

                await test.step('bind the user with a 2nd L1 researcher role with project2 scope', async () => {
                    await authorizationUtils.bindUserWithRoles(
                        user,
                        [PROPERTY_TO_ROLE_MAPPING.l1_researcher],
                        {
                            scopeId: String(project2.id),
                            scopeType: ScopeType.Project
                        },
                        adminTokens.idToken
                    );
                });

                await test.step('logout and login with created user', async () => {
                    await loginUIPage.logoutToLoginPage();
                    await authPage.signIn(user.username, user.tempPassword);
                });

                await test.step(`user can see project ${projectName1} and project ${projectName2} when creating workload`, async () => {
                    await trainingPage.goToPage();
                    await trainingPage.openTrainingForm();
                    const projectCardList = trainingPage.locator(cardsLocator);
                    await expect(projectCardList).toHaveCount(2, { timeout: 10000 });

                    const texts = [
                        await projectCardList.nth(0).innerText(),
                        await projectCardList.nth(1).innerText()
                    ];

                    expect(texts.some((text) => text.includes(projectName1))).toBeTruthy();
                    expect(texts.some((text) => text.includes(projectName2))).toBeTruthy();

                    await trainingPage.workloadFormSelectProject(projectName2);
                    await trainingPage.fillTrainingName(trainingName);
                    await trainingPage.closeWorkloadForm();
                });

                await test.step('user can create training sucessfully', async () => {
                    await trainingPage.createFlexibleTraining(projectName2, trainingName);
                    await trainingPage.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                });

                await test.step('delete training', async () => {
                    await trainingPage.deleteWorkload(trainingName);
                });

                await test.step('delete the user', async () => {
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        user.id
                    );
                });
            });
        });
    });
});
