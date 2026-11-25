/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    settings,
    tags,
    usersManager,
    usersUtils
} from '../../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { UsersFilterSortFields } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`IDP API with SSO user, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, { tag: [tags.services.identityManager, tags.services.keycloak] }, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            const idpSSoUser: string = 'ssoidpusertests';
            test.beforeAll(async function () {
                try {
                    // Cleanup existing resources from failed previous runs
                    try {
                        logger.debug('Cleaning up existing IDP and realm before test suite');
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        // 404 is expected if IDP doesn't exist
                        if (!error.message?.includes('404')) {
                            logger.debug(`IDP cleanup: ${error.message}`);
                        }
                    }

                    try {
                        await Keycloak.deleteRealm(realmName);
                        // Wait for Keycloak to finish deletion
                        await new Promise((resolve) => setTimeout(resolve, 1000));
                    } catch (error) {
                        // Expected if realm doesn't exist
                        logger.debug(`Realm cleanup: ${error.message}`);
                    }

                    // Now proceed with fresh setup
                    await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName);
                    await Keycloak.createUser(realmName, idpSSoUser, config.controlPlaneAdminPassword);
                    await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        `${idpSSoUser}@example.com`,
                        flowRootState.users.system_administrator.role.id
                    );
                } catch (error) {
                    throw new Error(`Failed to configure SSO, error: ${error}`);
                }
            });

            test.afterAll(async function () {
                try {
                    const userToDelete = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        [`username=@${idpSSoUser}@example.com`],
                        UsersFilterSortFields.Username
                    );
                    await usersManager.deleteSSOUser(
                        flowRootState.users.system_administrator.accessToken,
                        `${idpSSoUser}@example.com`,
                        userToDelete[0].id
                    );
                } catch (error) {
                    logger.error('failed to delete SSO user', JSON.stringify(error));
                } finally {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                    await Keycloak.deleteRealm(realmName);
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should get IDP by SSO user with default mappers', async () => {
                const token = await authUtils.authenticateUser(
                    idpSSoUser,
                    config.controlPlaneAdminPassword,
                    true
                );
                const idpResponse = await idpUtils.getIdp(token.accessToken, idpNameOidc);
                expect(idpResponse.body.alias, 'SSO user should see IDP alias matching the OIDC IDP name').to.eql(idpNameOidc);
                expect(idpResponse.body.type, 'IDP type should be oidc').to.eql('oidc');
                expect(idpResponse.body.redirectUri, 'IDP redirect URI should start with http').to.contain('http');
                expect(idpResponse.body.oidcData, 'IDP response should include oidcData object').to.be.an('object');
                expect(idpResponse.body.mappers.email, 'default email mapper should be "email"').to.eql('email');
                expect(idpResponse.body.mappers.gid, 'default GID mapper should be "GID"').to.eql('GID');
                expect(idpResponse.body.mappers.groups, 'default groups mapper should be "GROUPS"').to.eql('GROUPS');
                expect(idpResponse.body.mappers.supplementaryGroups, 'default supplementaryGroups mapper should be "SUPPLEMENTARYGROUPS"').to.eql('SUPPLEMENTARYGROUPS');
                expect(idpResponse.body.mappers.uid, 'default UID mapper should be "UID"').to.eql('UID');
            });

            test('should update mappers by SSO user with custom values', async () => {
                const token = await authUtils.authenticateUser(
                    idpSSoUser,
                    config.controlPlaneAdminPassword,
                    true
                );
                const idpResponse = await idpUtils.getIdp(token.accessToken, idpNameOidc);
                const discoveryURL = Keycloak.getDiscoveryDocumentURL(realmName);
                await idpUtils.updateIdpOidc(
                    token.accessToken,
                    idpNameOidc,
                    idpResponse.body.oidcData.clientId,
                    idpResponse.body.oidcData.clientSecret,
                    discoveryURL,
                    ['openid', 'email', 'profile', 'offline_access'],
                    null,
                    null,
                    {
                        email: 'updatedEmailtest',
                        gid: 'updatedGIDtest',
                        groups: 'updatedGROUPStest',
                        supplementaryGroups: 'updatedSUPPLEMENTARYGROUPStest',
                        uid: 'updatedUIDtest'
                    }
                );
                const afterUpdateIdp = await idpUtils.getIdp(token.accessToken, idpNameOidc);
                expect(afterUpdateIdp.body.mappers.email, 'email mapper should be updated to updatedEmailtest by SSO user').to.eql('updatedEmailtest');
                expect(afterUpdateIdp.body.mappers.gid, 'GID mapper should be updated to updatedGIDtest by SSO user').to.eql('updatedGIDtest');
                expect(afterUpdateIdp.body.mappers.groups, 'groups mapper should be updated to updatedGROUPStest by SSO user').to.eql('updatedGROUPStest');
                expect(afterUpdateIdp.body.mappers.supplementaryGroups, 'supplementaryGroups mapper should be updated by SSO user').to.eql(
                    'updatedSUPPLEMENTARYGROUPStest'
                );
                expect(afterUpdateIdp.body.mappers.uid, 'UID mapper should be updated to updatedUIDtest by SSO user').to.eql('updatedUIDtest');
                expect(afterUpdateIdp.body.oidcData.scopes, 'IDP scopes should be updated to the specified list').to.eql([
                    'openid',
                    'email',
                    'profile',
                    'offline_access'
                ]);
                expect(afterUpdateIdp.body.oidcData.mandatoryClaim, 'mandatory claim should be empty when not set').to.eql({ claim: '', values: null });
                expect(afterUpdateIdp.body.oidcData.clientId, 'clientId should remain unchanged after mapper update').to.eql(idpResponse.body.oidcData.clientId);
                await idpUtils.updateIdpMappers(
                    token.accessToken,
                    idpNameOidc,
                    'email',
                    'GID',
                    'GROUPS',
                    'SUPPLEMENTARYGROUPS',
                    'UID'
                );
                const afterUpdateIdpMappers = await idpUtils.getIdpMappers(token.accessToken, idpNameOidc);
                expect(afterUpdateIdpMappers.email, 'email mapper should revert to "email" after restore').to.eql('email');
                expect(afterUpdateIdpMappers.gid, 'GID mapper should revert to "GID" after restore').to.eql('GID');
                expect(afterUpdateIdpMappers.groups, 'groups mapper should revert to "GROUPS" after restore').to.eql('GROUPS');
                expect(afterUpdateIdpMappers.supplementaryGroups, 'supplementaryGroups mapper should revert to default after restore').to.eql('SUPPLEMENTARYGROUPS');
                expect(afterUpdateIdpMappers.uid, 'UID mapper should revert to "UID" after restore').to.eql('UID');
            });

            test('should return default mappers after update with empty values', async () => {
                const token = await authUtils.authenticateUser(
                    idpSSoUser,
                    config.controlPlaneAdminPassword,
                    true
                );
                const idpResponse = await idpUtils.getIdp(token.accessToken, idpNameOidc);
                const discoveryURL = Keycloak.getDiscoveryDocumentURL(realmName);
                await idpUtils.updateIdpOidc(
                    token.accessToken,
                    idpNameOidc,
                    idpResponse.body.oidcData.clientId,
                    idpResponse.body.oidcData.clientSecret,
                    discoveryURL,
                    ['openid', 'email', 'profile'],
                    null,
                    null,
                    {
                        email: '',
                        gid: '',
                        groups: '',
                        supplementaryGroups: '',
                        uid: ''
                    }
                );
                const afterUpdateIdp = await idpUtils.getIdp(token.accessToken, idpNameOidc);
                expect(afterUpdateIdp.body.mappers.email, 'email mapper should revert to default "email" when updated with empty values').to.eql('email');
                expect(afterUpdateIdp.body.mappers.gid, 'GID mapper should revert to default "GID" when updated with empty values').to.eql('GID');
                expect(afterUpdateIdp.body.mappers.groups, 'groups mapper should revert to default "GROUPS" when updated with empty values').to.eql('GROUPS');
                expect(afterUpdateIdp.body.mappers.supplementaryGroups, 'supplementaryGroups mapper should revert to default when updated with empty values').to.eql('SUPPLEMENTARYGROUPS');
                expect(afterUpdateIdp.body.mappers.uid, 'UID mapper should revert to default "UID" when updated with empty values').to.eql('UID');
            });
        });
    }
);
