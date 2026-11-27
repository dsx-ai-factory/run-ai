/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import {
    authUtils,
    config,
    idpUtils,
    logger,
    settings,
    tags,
    tenantUtils
} from '../../../../../modules/helpers';
import { idpNameOidc, idpNameSaml } from '../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';
import { StatusCodes } from 'http-status-codes';

test.describe(`Phoenix tests system scope IDP: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`, () => {
    let tenantName1: string;
    let tenantName2: string;
    let token: TokenResponse;
    let tenant1: TenantCreationResponse;
    let tenant2: TenantCreationResponse;
    let consoleUrl: string;
    const realmName1 = `realmcsp${Date.now()}`;
    const realmName2 = `realmcsp2${Date.now()}`;
    test.beforeAll(async function () {
        tenantName1 = `e2e-${Date.now()}-${Math.floor(Math.random() * 10000)}`;
        tenantName2 = `e2e-${Date.now()}-${Math.floor(Math.random() * 10000)}`;
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        const baseUrl = config.tenantUrl.replace('https://', '');
        consoleUrl = baseUrl.startsWith('console.') ? `https://${baseUrl}` : `https://console.${baseUrl}`;
        logger.debug(`Using console URL: ${consoleUrl}`);
        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            consoleUrl
        );
        logger.info(`Creating tenant: ${tenantName1} with super admin token: ${token.idToken}`);
        tenant1 = await tenantUtils.createTenant(
            token.accessToken,
            consoleUrl,
            tenantName1,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            TenantCreationRequest.ContractTypeEnum.Normal,
            null,
            null
        );
        logger.info(
            `Tenant created: status:${tenant1.tenant.status} and name:${tenant1.tenant.name} and id:${tenant1.tenant.id} and domain:${tenant1.additionalData['tenant_domain']}`
        );
        logger.info(`Creating tenant: ${tenantName2} with super admin token: ${token.idToken}`);
        tenant2 = await tenantUtils.createTenant(
            token.accessToken,
            consoleUrl,
            tenantName2,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            TenantCreationRequest.ContractTypeEnum.Normal,
            null,
            null
        );
        logger.info(
            `Tenant created: status:${tenant2.tenant.status} and name:${tenant2.tenant.name} and id:${tenant2.tenant.id} and domain:${tenant2.additionalData['tenant_domain']}`
        );
    });
    test.afterAll(async function () {
        const deleted1 = await tenantUtils.deleteTenant(
            token.accessToken,
            consoleUrl,
            tenant1.tenant.id,
            true
        );
        const deleted2 = await tenantUtils.deleteTenant(
            token.accessToken,
            consoleUrl,
            tenant2.tenant.id,
            true
        );
        logger.info(`Tenant deleted: ${deleted1.uid}`);
        logger.info(`Tenant deleted: ${deleted2.uid}`);
    });
    test.beforeEach(async function () {
        logger.debug('Before each Phoenix hook Started');
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        logger.debug('Before each Phoenix hook successful ended');
    });
    test('should create tenant and validate name', async function () {
        const token1 = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
            token.idToken,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            tenantName1,
            tenant1.additionalData['tenant_domain']
        );
        const token2 = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
            token.idToken,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            tenantName2,
            tenant2.additionalData['tenant_domain']
        );
        await Keycloak.createRealm(realmName1);
        const client = await Keycloak.createClient(realmName1, 'tenant1client', 'openid-connect');
        const clientSecret1 = await Keycloak.getClientSecret(realmName1, client.id);
        await Keycloak.createUser(realmName1, 'groupsuser', config.controlPlaneAdminPassword);
        await Keycloak.changeLoginTheme(realmName1, 'runai-theme');
        await Keycloak.addOIDCMapperToClient(
            realmName1,
            client.id,
            'groups',
            'oidc-group-membership-mapper',
            'String'
        );
        await Keycloak.configLoginWithSSOOIDCMultiTenant(
            token.idToken,
            Keycloak.getDiscoveryDocumentURL(realmName1, tenant1.additionalData['tenant_domain']),
            'tenant1client',
            clientSecret1,
            tenant1.tenant.id.toString()
        );
        logger.info(`Tenant 1 created: ${tenant1.tenant.name} and id:${tenant1.tenant.id}`);
        logger.info(`Tenant 1 token: ${token1.accessToken}`);
        logger.info(`Tenant 1 domain: ${tenant1.additionalData['tenant_domain']}`);
        await Keycloak.createRealm(realmName2);
        const client2 = await Keycloak.createClient(realmName2, 'tenant2client', 'openid-connect');
        const clientSecret2 = await Keycloak.getClientSecret(realmName2, client2.id);
        await Keycloak.createUser(realmName2, 'groupsuser', config.controlPlaneAdminPassword);
        await Keycloak.changeLoginTheme(realmName2, 'runai-theme');
        await Keycloak.addOIDCMapperToClient(
            realmName2,
            client2.id,
            'groups',
            'oidc-group-membership-mapper',
            'String'
        );
        await Keycloak.configLoginWithSSOOIDCMultiTenant(
            token.idToken,
            Keycloak.getDiscoveryDocumentURL(realmName2, tenant2.additionalData['tenant_domain']),
            'tenant2client',
            clientSecret2,
            tenant2.tenant.id.toString()
        );
        logger.info(`Tenant 2 created: ${tenant2.tenant.name} and id:${tenant2.tenant.id}`);
        logger.info(`Tenant 2 token: ${token2.accessToken}`);
        logger.info(`Tenant 2 domain: ${tenant2.additionalData['tenant_domain']}`);
        const idp1 = await Keycloak.getIdpFromRealm(tenant1.tenant.name, idpNameOidc);
        const idp2 = await Keycloak.getIdpFromRealm(tenant2.tenant.name, idpNameOidc);
        expect(idp1.enabled, 'OIDC IDP should be enabled in tenant1 Keycloak realm').toBe(true);
        expect(idp2.enabled, 'OIDC IDP should be enabled in tenant2 Keycloak realm').toBe(true);
        await Keycloak.deleteIDPFromRealm(tenant1.tenant.name, idpNameOidc);
        const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(tenant1.tenant.name);
        const samlIdp = await Keycloak.configLoginWithSSOSAMLMultiTenant(
            token.accessToken,
            metadataXmlUrl,
            tenant1.additionalData['tenant_domain'],
            tenant1.tenant.id.toString()
        );
        logger.info(`SAML IDP created: ${samlIdp}`);
        logger.info(`SAML IDP metadata URL: ${metadataXmlUrl}`);
        const idp3 = await Keycloak.getIdpFromRealm(tenant1.tenant.name, idpNameSaml);
        expect(idp3.enabled, 'SAML IDP should be enabled in tenant1 Keycloak realm after OIDC replacement').toBe(true);
        const runaiApiGetIdp = await idpUtils.getIdpMultiTenant(
            token.accessToken,
            idpNameOidc,
            tenant2.tenant.id.toString()
        );
        logger.info(`Runai API IDP: ${JSON.stringify(runaiApiGetIdp)}`);
        expect(runaiApiGetIdp.response.statusCode, 'getting IDP for tenant2 via system scope should return 200').toBe(StatusCodes.OK);
        const updateIdp = await idpUtils.updateIdpOidc(
            token.accessToken,
            idpNameOidc,
            'clientId',
            'clientSecret',
            Keycloak.getDiscoveryDocumentURL(realmName1, tenant1.additionalData['tenant_domain']),
            undefined,
            undefined,
            tenant2.tenant.id.toString()
        );
        logger.info(`Update IDP: ${JSON.stringify(updateIdp)}`);
        expect(updateIdp.response.statusCode, 'updating IDP for tenant2 via system scope should return 200').toBe(StatusCodes.OK);
        const deletedIdp = await idpUtils.deleteIdp(token.accessToken, idpNameOidc, tenant2.tenant.id.toString());
        logger.info(`Deleted IDP: ${JSON.stringify(deletedIdp)}`);
        expect(deletedIdp.response.statusCode, 'deleting IDP for tenant2 via system scope should return 204 No Content').toBe(StatusCodes.NO_CONTENT);
    });
});
