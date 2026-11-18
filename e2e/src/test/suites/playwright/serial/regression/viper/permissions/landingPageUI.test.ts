/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    settings,
    tags,
    usersManager,
    waitUtils
} from '../../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { navigateUtils } from '../../../../../../../modules/ui';

import {
    PROPERTY_TO_ROLE_MAPPING,
    RoleDisplayName,
    ROLE_NAME_TO_PROPERTY
} from '../../../../../../../modules/helpers/config/roleMapping';
import { test, expect } from '@playwright/test';
import { backendApi } from '../../../../../../../modules/api';
import { AuthUtils } from '../../../../../../../modules/helpers/utils/AuthUtils';
import { AuthEntityType } from '../../../../../../../modules/api/apis/backendApi';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';

test.describe(`Landing page UI tests, by all users: , ${tags.tool.ui}, ${tags.teams.viper}`, () => {
    let videoName: string;
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;

    const rolesByLandingPage: Record<string, RoleDisplayName[]> = {
        workloads: [
            PROPERTY_TO_ROLE_MAPPING.compute_resource_administrator,
            PROPERTY_TO_ROLE_MAPPING.credentials_administrator,
            PROPERTY_TO_ROLE_MAPPING.data_source_administrator,
            PROPERTY_TO_ROLE_MAPPING.department_viewer,
            PROPERTY_TO_ROLE_MAPPING.environment_administrator,
            PROPERTY_TO_ROLE_MAPPING.research_manager,
            PROPERTY_TO_ROLE_MAPPING.l1_researcher,
            PROPERTY_TO_ROLE_MAPPING.l2_researcher,
            PROPERTY_TO_ROLE_MAPPING.template_administrator
        ],
        overview: [
            PROPERTY_TO_ROLE_MAPPING.department_admin,
            PROPERTY_TO_ROLE_MAPPING.editor,
            PROPERTY_TO_ROLE_MAPPING.ml_engineer,
            PROPERTY_TO_ROLE_MAPPING.system_administrator,
            PROPERTY_TO_ROLE_MAPPING.viewer
        ]
    };

    const rolesWithUserFilterInWorkloadsLandingPage: RoleDisplayName[] = [
        PROPERTY_TO_ROLE_MAPPING.l1_researcher,
        PROPERTY_TO_ROLE_MAPPING.l2_researcher
    ];

    test.beforeAll(async () => {
        if (!settings.enableAuth) {
            // this test suite is only interesting if auth is enabled
            test.skip();
        }
        browserEngine = new BrowserEngine();

        // log all access rules for debugging
        logger.debug('All access rules:');
        const accessRules = await authorizationUtils.getAccessRules();
        logger.debug(
            accessRules.accessRules.map((rule) =>
                JSON.stringify({ subjectId: rule.subjectId, roleName: rule.roleName })
            )
        );
    });

    test.beforeEach(async () => {
        await browserEngine.start();
    });

    test.afterEach(async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test.describe('Workloads UI landing page', () => {
        rolesByLandingPage.workloads.forEach((role) => {
            test(`should login as ${role} user and land on Workloads UI`, async () => {
                videoName = `landing-page-${role.replace(/ /g, '-')}`;
                authPage = new AuthPage(browserEngine.page);
                await authPage.loginAsRole(role);
                logger.info(`Logged in using ${role} user, expected page title: Workloads}`);
                await navigateUtils.varifyPageTitle(authPage, 'Workloads');

                logger.debug(`Current URL: ${browserEngine.currentPage.url()}`);

                expect(browserEngine.currentPage.url()).toContain('/workloads');

                if (rolesWithUserFilterInWorkloadsLandingPage.includes(role)) {
                    const currentUrl = browserEngine.currentPage.url();
                    const filter = `advancedFilters=[{"term":"${flowRootState.users[ROLE_NAME_TO_PROPERTY[role]].email}","value":"==","label":"Created by","name":"submittedBy","type":"FreeText"}]`;
                    const encodedFilter = filter.replace(/"/g, '%22').replace(' ', '+');
                    expect(currentUrl).toContain(encodedFilter);

                    const filterSelector = `.runai-page-filter:has-text("equals ${flowRootState.users[ROLE_NAME_TO_PROPERTY[role]].email}")`;
                    try {
                        await waitUtils.waitForElementToExist(authPage, filterSelector);
                    } catch (error) {
                        logger.error(`Failed to find element with selector: ${filterSelector}`);
                        throw error;
                    }
                } else {
                    expect(browserEngine.currentPage.url()).not.toContain('advancedFilters');
                }
            });
        });
    });

    test.describe('Overview UI landing page', () => {
        rolesByLandingPage.overview.forEach((role) => {
            test(`should login as ${role} user and land on Overview UI`, async () => {
                videoName = `landing-page-${role}`;
                authPage = new AuthPage(browserEngine.page);
                await authPage.loginAsRole(role);
                logger.info(`Logged in using ${role} user, expected page title: Overview`);
                await navigateUtils.varifyPageTitle(authPage, 'Overview');

                expect(browserEngine.currentPage.url()).toContain('/dashboards');
            });
        });
    });

    test.describe('No roles landing page', () => {
        let user: UserPostResponse;
        test.beforeAll(async () => {
            backendApi.authenticationApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const userCreationRequest = AuthUtils.buildUserCreationRequestWithEntityType(
                objectGenerator.generateEmailByPrefix('no-permissions'),
                null,
                [],
                AuthEntityType.RegularUser
            );
            logger.debug(`Creating user without permissions: ${JSON.stringify(userCreationRequest)}`);
            user = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('no-permissions'),
                    false
                )
            ).body;
        });

        test.afterAll(async ({}, testInfo) => {
            await authUtils.deleteUser(user.id, flowRootState.users.system_administrator.accessToken);
        });

        test('should login as user without permissions and land on no-permissions UI', async () => {
            videoName = 'landing-page-no-permissions';
            authPage = new AuthPage(browserEngine.page);
            await authPage.signInAndVerifyNoRolesPage(user);
            logger.info('Logged in using user without permissions');
            expect(browserEngine.currentPage.url()).toContain('/no-permission');
        });
    });
});
