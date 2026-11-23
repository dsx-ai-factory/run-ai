/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { flowRootState, idpUtils, logger, settings, tags, usersManager } from '../../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { IdpPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/idpPostResponse';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`IDP API Local User, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, { tag: [tags.services.identityManager, tags.services.tenantsManager] }, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let createdIdp: IdpPostResponse;
            test.beforeAll(async function () {
                try {
                    await usersManager.deletePreSeededOidcIdp();
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
                    logger.error(error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should get IDP', async () => {
                const idpResponse = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(idpResponse.body.alias, 'retrieved IDP alias should match created IDP alias').to.eql(createdIdp.alias);
                expect(idpResponse.body.type, 'IDP type should be oidc').to.eql('oidc');
                expect(idpResponse.body.redirectUri, 'IDP redirect URI should start with http').to.contain('http');
                expect(idpResponse.body.oidcData, 'IDP response should include oidcData object').to.be.an('object');
            });

            test('should update IDP', async () => {
                const updatedClientId = `updatedClientId${Date.now()}`;
                const updatedSecret = `updatedSecret${Date.now()}`;
                const discoveryURL = Keycloak.getDiscoveryDocumentURL('runai');
                await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias,
                    updatedClientId,
                    updatedSecret,
                    discoveryURL
                );
                const afterUpdateIdp = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(afterUpdateIdp.body.oidcData.clientId, 'IDP clientId should be updated to the new value').to.eql(updatedClientId);
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.accessToken);
                expect(idps.length, 'exactly one IDP should be configured').to.eql(1);
                expect(idps[0].alias, 'IDP list entry alias should match created IDP').to.eql(createdIdp.alias);
                expect(idps[0].type, 'IDP list entry type should be oidc').to.eql('oidc');
                expect(idps[0].redirectUri, 'IDP list entry redirectUri should start with http').to.contain('http');
                expect(idps[0].oidcData, 'IDP list entry should include oidcData object').to.be.an('object');
            });

            test('should get IDP mappers', async () => {
                const mappers = await idpUtils.getIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(mappers.email, 'default email mapper should be "email"').to.eql('email');
                expect(mappers.gid, 'default GID mapper should be "GID"').to.eql('GID');
                expect(mappers.groups, 'default groups mapper should be "GROUPS"').to.eql('GROUPS');
                expect(mappers.supplementaryGroups, 'default supplementaryGroups mapper should be "SUPPLEMENTARYGROUPS"').to.eql('SUPPLEMENTARYGROUPS');
                expect(mappers.uid, 'default UID mapper should be "UID"').to.eql('UID');
            });

            test('should update IDP mappers', async () => {
                await idpUtils.updateIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias,
                    'updatedEmail',
                    'updatedGID',
                    'updatedGROUPS',
                    'updatedSUPPLEMENTARYGROUPS',
                    'updatedUID'
                );
                const afterUpdateMappers = await idpUtils.getIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(afterUpdateMappers.email, 'email mapper should be updated to updatedEmail').to.eql('updatedEmail');
                expect(afterUpdateMappers.gid, 'GID mapper should be updated to updatedGID').to.eql('updatedGID');
                expect(afterUpdateMappers.groups, 'groups mapper should be updated to updatedGROUPS').to.eql('updatedGROUPS');
                expect(afterUpdateMappers.supplementaryGroups, 'supplementaryGroups mapper should be updated').to.eql('updatedSUPPLEMENTARYGROUPS');
                expect(afterUpdateMappers.uid, 'UID mapper should be updated to updatedUID').to.eql('updatedUID');
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.accessToken);
                expect(idps.length, 'exactly one IDP should be configured after mapper update').to.eql(1);
                expect(idps[0].mappers, 'IDP list entry mappers should match updated mapper values').to.eql(afterUpdateMappers);
                await idpUtils.updateIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias,
                    'Email',
                    'GID',
                    'GROUPS',
                    'SUPPLEMENTARYGROUPS',
                    'UID'
                );
                const returnToDefaultMappers = await idpUtils.getIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(returnToDefaultMappers.email, 'email mapper should be restored to Email').to.eql('Email');
                expect(returnToDefaultMappers.gid, 'GID mapper should remain GID after restore').to.eql('GID');
                expect(returnToDefaultMappers.groups, 'groups mapper should remain GROUPS after restore').to.eql('GROUPS');
                expect(returnToDefaultMappers.supplementaryGroups, 'supplementaryGroups mapper should remain SUPPLEMENTARYGROUPS after restore').to.eql('SUPPLEMENTARYGROUPS');
                expect(returnToDefaultMappers.uid, 'UID mapper should remain UID after restore').to.eql('UID');
                const idpsDefault = await idpUtils.getIdps(
                    flowRootState.users.system_administrator.accessToken
                );
                expect(idpsDefault.length, 'exactly one IDP should be configured after mapper restore').to.eql(1);
                expect(idpsDefault[0].mappers, 'IDP list entry mappers should match restored default values').to.eql(returnToDefaultMappers);
            });

            test('should update IDP mappers, delete values and return to default', async () => {
                await idpUtils.updateIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias,
                    '',
                    '',
                    '',
                    '',
                    ''
                );
                const afterUpdateMappers = await idpUtils.getIdpMappers(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(afterUpdateMappers.email, 'email mapper should revert to default "email" when cleared').to.eql('email');
                expect(afterUpdateMappers.gid, 'GID mapper should revert to default "GID" when cleared').to.eql('GID');
                expect(afterUpdateMappers.groups, 'groups mapper should revert to default "GROUPS" when cleared').to.eql('GROUPS');
                expect(afterUpdateMappers.supplementaryGroups, 'supplementaryGroups mapper should revert to default when cleared').to.eql('SUPPLEMENTARYGROUPS');
                expect(afterUpdateMappers.uid, 'UID mapper should revert to default "UID" when cleared').to.eql('UID');
            });

            test('should get IDP resources', async () => {
                const resources = await idpUtils.getIdpResources(
                    flowRootState.users.system_administrator.accessToken,
                    createdIdp.alias
                );
                expect(resources.type, 'IDP resource type should be "url"').to.eql('url');
                expect(resources.resourceName, 'IDP resource name should be "url"').to.eql('url');
                expect(resources.resource, 'IDP resource URL should start with http').to.contain('http');
            });
        });
    }
);
