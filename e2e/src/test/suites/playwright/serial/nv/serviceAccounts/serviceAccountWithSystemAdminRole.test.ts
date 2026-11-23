/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authUtils,
    flowRootState,
    objectGenerator,
    serviceAccountUtils,
    tags,
    tenantUtils,
    usersManager
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import {
    AccessRulesBatchCreateFields,
    ScopeType,
    SubjectType
} from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { User } from '../../../../../../modules/helpers/config/types';
import { StatusCodes } from 'http-status-codes';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../modules/helpers/config/roleMapping';

test.describe(`Service Account with system admin access role - role testing, tags: ${tags.teams.loong} ${tags.tool.api} ${tags.resource.applications}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let serviceAccountName: string;
        let serviceAccountToken: string;
        let user: User;
        let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };
        let videoName: string;
        let userCreated = false;
        let userDeleted = false;
        let browerClosed = false;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            serviceAccountName = objectGenerator.generateShortNameByPrefix('sa-system-admin', 5);
            serviceAccount = await serviceAccountUtils.createServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                { name: serviceAccountName }
            );

            const roles = await authorizationUtils.getRoles(
                flowRootState.users.system_administrator.accessToken
            );
            const sysAdminId = roles.find(
                (role) => role.name === PROPERTY_TO_ROLE_MAPPING.system_administrator
            )?.id;

            const tenant = await tenantUtils.getTenant(flowRootState.users.system_administrator.accessToken);

            const accessRulesBatchCreateFields: AccessRulesBatchCreateFields = {
                payload: {
                    roleId: sysAdminId,
                    scopes: [
                        {
                            scopeId: String(tenant.id),
                            scopeType: ScopeType.Tenant
                        }
                    ],
                    subjects: [
                        {
                            subjectId: serviceAccount.body.clientId,
                            subjectType: SubjectType.ServiceAccount
                        }
                    ]
                }
            };

            const response = await authorizationUtils.bindAccessRuleToLocalUser(accessRulesBatchCreateFields);
            expect(response.response.statusCode).toEqual(StatusCodes.OK);
        });

        test.afterAll('deleting project and namespace', async ({}, testInfo) => {
            if (!browerClosed) {
                await browserEngine.stopAndArchive(testInfo, videoName);
            }

            if (serviceAccount) {
                await serviceAccountUtils.deleteServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    serviceAccount.body.id
                );
            }

            if (userCreated && !userDeleted) {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.userId
                );
            }
        });

        test('T3794760: create and delete system admin local user by API using service account access token', async ({}, testInfo) => {
            await test.step('get service account access token', async () => {
                serviceAccountToken = (
                    await authUtils.authenticateAppAndGetToken(
                        serviceAccount.body.clientId,
                        serviceAccount.body.clientSecret
                    )
                ).accessToken;
            });

            await test.step('create a local user with system admin rule using the access token', async () => {
                const getRolesResp = await authorizationUtils.getRoles(serviceAccountToken);
                const systemAdminRole = getRolesResp.filter(
                    (role) => role.name === PROPERTY_TO_ROLE_MAPPING.system_administrator
                )[0];
                user = await usersManager.createUserWithRole(
                    serviceAccountToken,
                    systemAdminRole,
                    objectGenerator.generateShortNameByPrefix('sa-sysadmin', 5)
                );
                userCreated = true;
            });

            await test.step('login with the created user and should be able to saw all the menus', async () => {
                videoName = 'created-local-user-ui-check';
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);

                await authPage.signIn(user.email, user.password);

                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`)
                ).toBeVisible();
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)).toBeVisible();
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Organization}"]`)
                ).toBeVisible();
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).toBeVisible();
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).toBeVisible();

                await browserEngine.stopAndArchive(testInfo, videoName);
                browerClosed = true;
            });

            await test.step('delete the local user with service account access token', async () => {
                const resp = await usersManager.deleteLocalUser(serviceAccountToken, user.userId);
                expect(resp.response.statusCode).toEqual(StatusCodes.NO_CONTENT);
                userDeleted = true;
            });
        });
    });
});
