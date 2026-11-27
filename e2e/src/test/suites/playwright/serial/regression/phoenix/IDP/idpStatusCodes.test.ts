/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { flowRootState, idpUtils, logger, settings, tags } from '../../../../../../../modules/helpers';

import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { IdpPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/idpPostResponse';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`IDP Status Codes tests, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let createdIdp: IdpPostResponse;
            test.beforeAll(async function () {
                try {
                    const discoveryURL = Keycloak.getDiscoveryDocumentURL('runai');
                    createdIdp = await idpUtils.createIdpOidc(
                        flowRootState.users.system_administrator.accessToken,
                        `e2e${Date.now()}`,
                        `clientId${Date.now()}`,
                        'test-client-secret',
                        discoveryURL
                    );
                } catch (error) {
                    throw new Error(`Failed to configure SSO, error: ${error}`);
                }
            });

            test.afterAll(async function () {
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        createdIdp.alias
                    );
                } catch (error) {
                    logger.error(`Failed to delete idp, or it just deleted error: ${error}`);
                }
            });

            test('should validate 200 status code for get idp', async () => {
                const idpResponse = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(idpResponse.response.statusCode, 'getting IDP should return 200').to.eql(200);
            });

            test('should validate 403 Insufficient permissions status code for editor get idp', async () => {
                const getEditorIdp = await idpUtils.getIdp(
                    flowRootState.users.editor.accessToken,
                    createdIdp.alias
                );
                expect(getEditorIdp.response.statusCode, 'editor getting IDP should return 200').to.eql(200);
            });

            test('should validate 404 non exist status code for get idp', async function () {
                try {
                    await idpUtils.getIdp(
                        flowRootState.users.system_administrator.accessToken,
                        createdIdp.alias + 'nonExist'
                    );
                    expect.fail('Expected 404 status code');
                } catch (error) {
                    expect(error.statusCode, 'getting a non-existent IDP should return 404').to.eql(404);
                }
            });

            test('should validate 200 status code for IDP update', async () => {
                const updatedClientId = `updatedClientId${Date.now()}`;
                const updatedSecret = `updatedSecret${Date.now()}`;
                const discoveryURL = Keycloak.getDiscoveryDocumentURL('runai');
                const updateIdp = await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias,
                    updatedClientId,
                    updatedSecret,
                    discoveryURL
                );

                expect(updateIdp.response.statusCode, 'updating IDP should return 200').to.eql(200);
            });

            test('should validate editor 403 status code for IDP update ', async () => {
                const updatedClientId = `updatedClientId${Date.now()}`;
                const updatedSecret = `updatedSecret${Date.now()}`;
                const discoveryURL = Keycloak.getDiscoveryDocumentURL('runai');
                try {
                    await idpUtils.updateIdpOidc(
                        flowRootState.users.editor.accessToken,
                        createdIdp.alias,
                        updatedClientId,
                        updatedSecret,
                        discoveryURL
                    );
                    expect.fail('Expected 403 status code');
                } catch (error) {
                    expect(error.statusCode, 'editor updating IDP should be forbidden with 403').to.eql(403);
                }
            });

            test('should validate 400 status code for invalid IDP configuration', async () => {
                try {
                    await idpUtils.createIdpOidc(
                        flowRootState.users.system_administrator.accessToken,
                        '',
                        '',
                        '',
                        ''
                    );
                } catch (error) {
                    expect(error.statusCode, 'creating an IDP with empty configuration should return 400').to.eql(400);
                }
            });

            test('should validate 204 status code for successful IDP deletion', async function () {
                const deleted = await idpUtils.deleteIdp(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(deleted.response.statusCode, 'deleting IDP should return 204 No Content').to.eql(204);
            });
        });
    }
);
