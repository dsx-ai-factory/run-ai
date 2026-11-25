/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KeycloakAdminClient } from 'keycloak-admin/lib/client.js';
import { config, flowRootState, isStagingEnv, openshiftApiUrl, settings } from '../config/config';
import { identityManagerApi } from '../../api';
import { logger } from '..';
import { resolveKeycloakPassword } from './CredentialsResolver';
import { emailMapperConfig, firstNameMapperConfig, lastNameMapperConfig, groupMapperConfig, getSsoEntityId, idpNameOcp, idpNameOidc, idpNameSaml, ssoRedirectUri, namespaceCP } from '../config/const';
import axios from 'axios';
import { kubeCli } from '../../cli';
import { IdpCreationRequest } from '../../api/apis/identityManagerApi/generated/model/idpCreationRequest';
import { SamlCreationDataMetadataXmlTypeEnum } from '../../api/apis/identityManagerApi/generated/model/samlCreationDataMetadataXmlTypeEnum';
import RealmRepresentation from 'keycloak-admin/lib/defs/realmRepresentation';
import { IdpCreationRequestTypeEnum, MandatoryClaim } from '../../api/apis/identityManagerApi/generated/api';
import qs from 'qs';
import { StatusCodes } from 'http-status-codes';
import { isOCP } from './EnvironmentUtils';

export interface KeycloakTokenResponse {
    access_token: string;
    refresh_token: string;
    token_type: string;
    expires_in: number;
    scope: string;
}

export class Keycloak {
    static async initKeycloakAdminClient(): Promise<KeycloakAdminClient> {
        let keycloakPASS: string;

        if (settings.isSelfHosted || isOCP()) {
            try {
                keycloakPASS = await resolveKeycloakPassword();
                logger.debug('Fetched Keycloak admin password from cluster secret');
            } catch (error) {
                logger.warn(`Failed to resolve Keycloak password from cluster: ${error.message}, falling back to env var`);
                keycloakPASS = config.KEYCLOAKADMINPASSWORD;
            }
        } else if (isStagingEnv && !settings.isSelfHosted) {
            keycloakPASS = config.KEYCLOAKADMINPASSWORDSTAGING;
        } else {
            keycloakPASS = config.KEYCLOAKADMINPASSWORDTEST;
        }

        const keycloakAdminClient = new KeycloakAdminClient({
            baseUrl: `${config.tenantUrl}/auth`,
            realmName: 'master'
        });
        await keycloakAdminClient.auth({
            username: 'admin',
            password: keycloakPASS,
            grantType: 'password',
            clientId: 'admin-cli'
        });
        return keycloakAdminClient;
    }

    static async getRealmByName(realmName: string): Promise<RealmRepresentation> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const realm = await keycloakAdminClient.realms.findOne({ realm: realmName });
        return realm;
    }

    static async createRealm(realmName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.realms.create({
            realm: realmName,
            enabled: true
        });
    }

    static async deleteRealm(realmName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.realms.del({ realm: realmName });
    }

    static async createClient(
        realmName: string,
        clientName: string,
        protocol: string,
        redirectUris?: string
    ): Promise<any> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        let client: any;
        if (protocol === 'openid-connect') {
            client = await keycloakAdminClient.clients.create({
                clientId: clientName,
                realm: realmName,
                enabled: true,
                protocol: protocol,
                publicClient: false,
                redirectUris: ['*'],
                clientAuthenticatorType: 'client-secret',
                serviceAccountsEnabled: true
            });
        } else {
            client = await keycloakAdminClient.clients.create({
                clientId: clientName,
                realm: realmName,
                enabled: true,
                protocol: protocol,
                publicClient: false,
                redirectUris: [redirectUris],
                attributes: {
                    'saml.client.signature': false
                },
                clientAuthenticatorType: 'client-secret'
            });
        }
        return client;
    }

    static async getClientSecret(realmName: string, clientId: string): Promise<string> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.findOne({ realm: realmName, id: clientId });
        if (client && typeof client === 'object' && 'secret' in client) {
            return client.secret;
        } else {
            throw new Error(
                `Client not found or client secret not available for client ID: ${clientId} in realm: ${realmName}`
            );
        }
    }

    static async getClientSecretFromSpecificClient(realmName: string, clientId: string): Promise<string> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.find({ realm: realmName, clientId: clientId });
        return client[0].secret;
    }

    static async updateClient(realmName: string, client: any): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const foundClient = await keycloakAdminClient.clients.findOne({ realm: realmName, id: client.id });
        if (foundClient) {
            foundClient.publicClient = true;
            await keycloakAdminClient.clients.update({ realm: realmName, id: client.id }, foundClient);
        } else {
            throw new Error('Client not found');
        }
    }

    static async createUser(realmName: string, username: string, password: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.users.create({
            realm: realmName,
            username: username,
            enabled: true,
            emailVerified: true,
            email: `${username}@example.com`,
            firstName: username,
            lastName: username,
            credentials: [
                {
                    type: 'password',
                    temporary: false,
                    value: password
                }
            ]
        });
    }

    static async setupSSORealmOIDCForNonEmailUsers(
        realmName: string,
        clientName: string
    ): Promise<void> {
        await Keycloak.setupSSORealmOIDC(realmName, clientName);
        await Keycloak.makeProfileAttributesOptional(realmName, ['email', 'firstName', 'lastName']);
        await Keycloak.pointEmailScopeMapperAtUsername(realmName);
    }

    static async makeProfileAttributesOptional(
        realmName: string,
        attributeNames: string[]
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const token = keycloakAdminClient.accessToken;
        if (!token) {
            throw new Error('Failed to retrieve Keycloak Admin Token');
        }
        const url = `${config.tenantUrl}/auth/admin/realms/${realmName}/users/profile`;
        const authHeader = { Authorization: `Bearer ${token}` };
        const { data: profile } = await axios.get(url, { headers: authHeader });
        const cleared: string[] = [];
        for (const name of attributeNames) {
            const attr = (profile.attributes ?? []).find(
                (a: { name?: string }) => a.name === name
            );
            if (attr?.required) {
                delete attr.required;
                cleared.push(name);
            }
        }
        if (cleared.length > 0) {
            await axios.put(url, profile, {
                headers: { ...authHeader, 'Content-Type': 'application/json' }
            });
        }
        logger.info(
            `Made [${cleared.join(', ')}] optional in realm '${realmName}' user profile`
        );
    }

    static async pointEmailScopeMapperAtUsername(realmName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const token = keycloakAdminClient.accessToken;
        if (!token) {
            throw new Error('Failed to retrieve Keycloak Admin Token');
        }
        const baseUrl = `${config.tenantUrl}/auth/admin/realms/${realmName}/client-scopes`;
        const authHeader = { Authorization: `Bearer ${token}` };

        const { data: scopes } = await axios.get<Array<{ id: string; name: string }>>(baseUrl, {
            headers: authHeader
        });
        const emailScope = scopes.find((s) => s.name === 'email');
        if (!emailScope) {
            throw new Error(`'email' client scope not found in realm '${realmName}'`);
        }

        const mappersUrl = `${baseUrl}/${emailScope.id}/protocol-mappers/models`;
        const { data: mappers } = await axios.get<
            Array<{ id: string; name: string; config: Record<string, string> }>
        >(mappersUrl, { headers: authHeader });
        const emailMapper = mappers.find((m) => m.name === 'email');
        if (!emailMapper) {
            throw new Error(
                `'email' mapper not found in 'email' client scope of realm '${realmName}'`
            );
        }
        emailMapper.config = {
            ...emailMapper.config,
            'user.attribute': 'username'
        };
        await axios.put(`${mappersUrl}/${emailMapper.id}`, emailMapper, {
            headers: { ...authHeader, 'Content-Type': 'application/json' }
        });
        logger.info(
            `Pointed 'email' client scope mapper at 'username' in realm '${realmName}'`
        );
    }

    static async setupSSORealmSAMLForNonEmailUsers(realmName: string): Promise<void> {
        await Keycloak.setupSSORealmSAML(realmName);
        await Keycloak.makeProfileAttributesOptional(realmName, ['email', 'firstName', 'lastName']);

        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const ssoEntityId = getSsoEntityId();
        const clients = await keycloakAdminClient.clients.find({
            realm: realmName,
            clientId: ssoEntityId
        });
        if (!clients || clients.length === 0) {
            throw new Error(`SAML client '${ssoEntityId}' not found in realm '${realmName}'`);
        }
        const client = clients[0];

        await keycloakAdminClient.clients.update(
            { realm: realmName, id: client.id },
            {
                ...client,
                attributes: {
                    ...client.attributes,
                    'saml_name_id_format': 'username',
                    'saml_force_name_id_format': 'true'
                }
            }
        );

        const mappers = client.protocolMappers ?? [];
        const emailMapper = mappers.find((m) => m.name === 'X500 email');
        if (!emailMapper) {
            throw new Error(
                `SAML mapper 'X500 email' not found on client '${ssoEntityId}' in realm '${realmName}'`
            );
        }
        emailMapper.config = {
            ...emailMapper.config,
            'user.attribute': 'username'
        };
        await keycloakAdminClient.clients.updateProtocolMapper(
            { realm: realmName, id: client.id, mapperId: emailMapper.id },
            emailMapper
        );
        logger.info(
            `Configured SAML realm '${realmName}' to emit username as email attribute for non-email users`
        );
    }

    static async createUserWithoutEmail(
        realmName: string,
        username: string,
        password: string,
        firstName?: string,
        lastName?: string
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.users.create({
            realm: realmName,
            username: username,
            enabled: true,
            emailVerified: false,
            firstName: firstName,
            lastName: lastName,
            credentials: [
                {
                    type: 'password',
                    temporary: false,
                    value: password
                }
            ]
        });
        logger.info(`Created user '${username}' without email in realm '${realmName}'`);
    }

    static getDiscoveryDocumentURL(realmName: string, tenantUrl?: string): string {
        return `${tenantUrl || config.tenantUrl}/auth/realms/${realmName}/.well-known/openid-configuration`;
    }

    static async changeLoginTheme(realmName: string, themeName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.realms.update({ realm: realmName }, { loginTheme: themeName });
    }

    public static async configLoginWithSSOOIDC(
        discoveryDocumentURL?: string,
        clientId?: string,
        clientSecret?: string,
        mandatoryClaim?: MandatoryClaim
    ): Promise<void> {
        identityManagerApi.idpsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        try {
            const idpCreationRequest: IdpCreationRequest = {
                type: IdpCreationRequestTypeEnum.Oidc,
                name: idpNameOidc,
                oidcData: {
                    clientId: clientId,
                    clientSecret: clientSecret,
                    discoverDocumentUrl: discoveryDocumentURL,
                    scopes: ['openid', 'email', 'profile'],
                    mandatoryClaim: mandatoryClaim
                }
            };
            await identityManagerApi.idpsApi.createIdp(idpCreationRequest);
            logger.debug('SSO is configured');
        } catch (error) {
            throw new Error(`Failed to configure SSO, error: ${error}`);
        }
    }

    public static async configLoginWithSSOOIDCMultiTenant(
        accessToken?: string,
        discoveryDocumentURL?: string,
        clientId?: string,
        clientSecret?: string,
        tenantId?: string
    ): Promise<void> {
        identityManagerApi.idpsApi.accessToken =
            accessToken || flowRootState.users.system_administrator.accessToken;
        try {
            const idpCreationRequest: IdpCreationRequest = {
                type: IdpCreationRequestTypeEnum.Oidc,
                name: idpNameOidc,
                oidcData: {
                    clientId: clientId,
                    clientSecret: clientSecret,
                    discoverDocumentUrl: discoveryDocumentURL,
                    scopes: ['openid', 'email', 'profile']
                }
            };
            await identityManagerApi.idpsApi.createIdp(idpCreationRequest, tenantId);
            logger.debug('SSO is configured');
        } catch (error) {
            throw new Error(`Failed to configure SSO, error: ${error}`);
        }
    }

    static async deleteIDPFromRealm(realmName: string, idpName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        try {
            await keycloakAdminClient.identityProviders.del({ realm: realmName, alias: idpName });
        } catch (error) {
            logger.debug(`Failed to delete IDP, error: ${error}`);
        }
    }

    static async deleteUserFromRealm(realmName: string, username: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const users = await keycloakAdminClient.users.find({ realm: realmName, username });
        if (users && users.length > 0) {
            await keycloakAdminClient.users.del({ realm: realmName, id: users[0].id });
        }
    }

    static async addAttributesToUser(realmName: string, username: string, attributes: any): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const user = await keycloakAdminClient.users.find({ realm: realmName, username });
        if (user) {
            await keycloakAdminClient.users.update(
                { realm: realmName, id: user[0].id },
                {
                    attributes: attributes
                }
            );
        }
    }

    static async addKeycloakGroup(realmName: string, groupName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.groups.create({
            realm: realmName,
            name: groupName,
            path: groupName
        });
    }

    static async joinUserToGroup(realmName: string, username: string, groupName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const user = await keycloakAdminClient.users.find({ realm: realmName, username });
        const group = await keycloakAdminClient.groups.find({ realm: realmName, search: groupName });
        if (user && group) {
            await keycloakAdminClient.users.addToGroup({
                realm: realmName,
                id: user[0].id,
                groupId: group[0].id
            });
        }
    }

    static async AddMapperToClient(
        realmName: string,
        clientId: string,
        protocol: string,
        mapperName: string,
        mapperType: string,
        config: any
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.findOne({ realm: realmName, id: clientId });
        if (protocol === 'openid-connect') {
            await keycloakAdminClient.clients.addProtocolMapper(
                { realm: realmName, id: client.id },
                {
                    name: mapperName,
                    protocol: protocol,
                    protocolMapper: mapperType,
                    config: config
                }
            );
        } else {
            try {
                await keycloakAdminClient.clients.addProtocolMapper(
                    { realm: realmName, id: client.id },
                    {
                        name: mapperName,
                        protocol: protocol,
                        protocolMapper: mapperType,
                        config: config
                    }
                );
            } catch (error) {
                console.error('Error adding protocol mapper:', error);
                throw error;
            }
        }
    }

    static async createUsers(realmName: string, usernames: string[], password: string[]): Promise<void> {
        const promises = usernames.map(async (username, index) => {
            return await Keycloak.createUser(realmName, username, password[index]);
        });

        await Promise.all(promises);
    }

    static generateGroups(count: number): string[] {
        return Array.from({ length: count }, (_, i) => `group${i + 1}`);
    }

    static async removeUserFromGroup(realmName: string, username: string, groupName: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const user = await keycloakAdminClient.users.find({ realm: realmName, username });
        const group = await keycloakAdminClient.groups.find({ realm: realmName, search: groupName });
        if (user && group && user[0] && group[0]) {
            await keycloakAdminClient.users.delFromGroup({
                realm: realmName,
                id: user[0].id,
                groupId: group[0].id
            });
            logger.info(`User ${username} removed from group ${groupName} in realm ${realmName}`);
        } else {
            logger.warn(`Could not find user ${username} or group ${groupName} in realm ${realmName}`);
        }
    }

    static async createKeycloakGroupsWithUsers(
        realmName: string,
        groupNames: string[],
        usernames: string[]
    ): Promise<void> {
        const promises = groupNames.map(async (groupName, index) => {
            await Keycloak.addKeycloakGroup(realmName, groupName);
            await Keycloak.joinUserToGroup(realmName, usernames[index], groupName);
        });
        await Promise.all(promises);
    }

    static async addOIDCMapperToClient(
        realmName: string,
        clientId: string,
        mapperName: string,
        mapperType: string,
        jsonTypeLabel: string,
        userAttribute?: string,
        claimName?: string
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.findOne({ realm: realmName, id: clientId });

        if (!client) {
            throw new Error(`Client '${clientId}' not found in realm '${realmName}'`);
        }

        // Define base configuration
        const config = {
            'full.path': 'false',
            'id.token.claim': 'true',
            'access.token.claim': 'true',
            'claim.name': claimName || 'GROUPS',
            'userinfo.token.claim': 'true',
            'jsonType.label': jsonTypeLabel
        };

        if (mapperType === 'oidc-usermodel-attribute-mapper') {
            config['user.attribute'] = userAttribute;
        }

        await keycloakAdminClient.clients.addProtocolMapper(
            { realm: realmName, id: client.id },
            {
                name: mapperName,
                protocol: 'openid-connect',
                protocolMapper: mapperType,
                config: config
            }
        );

        logger.info(`OIDC Mapper '${mapperName}' added to client '${clientId}' in realm '${realmName}'`);
    }

    static async getIdpFromRealm(realmName: string, idpName: string): Promise<any> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const idp = await keycloakAdminClient.identityProviders.findOne({
            realm: realmName,
            alias: idpName
        });
        return idp;
    }

    static async setupSSORealmOIDC(
        realmName: string,
        keycloakClientName: string,
        mandatoryClaim?: MandatoryClaim
    ): Promise<void> {
        // Create Keycloak realm
        await Keycloak.createRealm(realmName);

        // Create Keycloak client
        const client = await Keycloak.createClient(realmName, keycloakClientName, 'openid-connect');

        // Get the client secret
        const clientSecret = await Keycloak.getClientSecret(realmName, client.id);

        await Keycloak.createUser(realmName, 'automationsso', config.controlPlaneAdminPassword);

        // Create user(s) in the realm
        await Keycloak.createKeycloakGroupsWithUsers(realmName, ['superadmingroup'], ['automationsso']);
        await Keycloak.addOIDCMapperToClient(
            realmName,
            client.id,
            'groups',
            'oidc-group-membership-mapper',
            'String'
        );
        if (mandatoryClaim) {
            await Keycloak.addOIDCMapperToClient(
                realmName,
                client.id,
                'claim',
                'oidc-usermodel-attribute-mapper',
                'String',
                'claim',
                'claim'
            );
        }

        // Change login theme
        await Keycloak.changeLoginTheme(realmName, 'runai-theme');

        // Configure login with SSO OIDC
        const discoveryURL = Keycloak.getDiscoveryDocumentURL(realmName);
        await Keycloak.configLoginWithSSOOIDC(discoveryURL, keycloakClientName, clientSecret, mandatoryClaim);
    }

    public static async configLoginWithSSOSAML(
        samlMetadataURL?: string,
        url?: string,
        tenantId?: string
    ): Promise<void> {
        identityManagerApi.idpsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        try {
            const idpCreationRequest: IdpCreationRequest = {
                type: IdpCreationRequestTypeEnum.Saml,
                name: idpNameSaml,
                samlData: {
                    metadataXmlUrl: samlMetadataURL,
                    metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.Url,
                    entityId: Keycloak.getSsoEntityId(url || config.tenantUrl)
                }
            };
            await identityManagerApi.idpsApi.createIdp(idpCreationRequest, tenantId);
            logger.debug('SSO is configured');
        } catch (error) {
            throw new Error(`Failed to configure SSO, error: ${error}`);
        }
    }

    public static async configLoginWithSSOSAMLMultiTenant(
        accessToken?: string,
        samlMetadataURL?: string,
        url?: string,
        tenantId?: string
    ): Promise<void> {
        identityManagerApi.idpsApi.accessToken = accessToken;
        try {
            const idpCreationRequest: IdpCreationRequest = {
                type: IdpCreationRequestTypeEnum.Saml,
                name: idpNameSaml,
                samlData: {
                    metadataXmlUrl: samlMetadataURL,
                    metadataXmlType: SamlCreationDataMetadataXmlTypeEnum.Url,
                    entityId: Keycloak.getSsoEntityId(url || config.tenantUrl)
                }
            };
            await identityManagerApi.idpsApi.createIdp(idpCreationRequest, tenantId);
            logger.debug('SSO is configured');
        } catch (error) {
            throw new Error(`Failed to configure SSO, error: ${error}`);
        }
    }

    public static async configLoginWithSSOOCP(clientId: string, idpSecret: string): Promise<void> {
        identityManagerApi.idpsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const fullUrl = config.tenantUrl;
        if (!openshiftApiUrl) {
            throw new Error('OPENSHIFT_API_URL must be set for OpenShift SSO.');
        }
        const authClient = await kubeCli.createOAuthClient(
            new URL(`/auth/realms/runai/broker/${idpNameOcp}/endpoint`, fullUrl).href,
            clientId,
            idpSecret
        );
        try {
            const idpCreationRequest: IdpCreationRequest = {
                type: IdpCreationRequestTypeEnum.OpenshiftV4,
                name: idpNameOcp,
                ocpData: {
                    idpBaseUrl: openshiftApiUrl,
                    clientId: authClient.clientId,
                    clientSecret: authClient.idpSecret
                }
            };
            await identityManagerApi.idpsApi.createIdp(idpCreationRequest);
            logger.debug('SSO is configured');
        } catch (error) {
            throw new Error(`Failed to configure SSO, error: ${error}`);
        }
    }

    static getConfigurationProtocolEndpointSAMLUrl(realmName: string): string {
        const url = `${config.tenantUrl}/auth/realms/${realmName}/protocol/saml/descriptor`;
        logger.debug(`getConfigurationProtocolEndpointSAMLUrl from url ${url}`);
        return url;
    }

    static getSsoEntityId(tenantUrl: string): string {
        const urlParts = tenantUrl.split('//');
        const domain = urlParts.length > 1 ? urlParts[1] : urlParts[0];
        return `runai-${domain}/runai`;
    }

    static async setupSSORealmSAML(realmName: string): Promise<void> {
        await Keycloak.createRealm(realmName);
        await Keycloak.createUser(realmName, 'saml', config.controlPlaneAdminPassword);
        const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
        logger.debug(`metadataXmlUrl: ${metadataXmlUrl}`);
        await Keycloak.configLoginWithSSOSAML(metadataXmlUrl);
        const client = await Keycloak.createClient(realmName, getSsoEntityId(), 'saml', ssoRedirectUri);
        await Keycloak.changeLoginTheme(realmName, 'runai-theme');
        await Keycloak.AddMapperToClient(
            realmName,
            client.id,
            'saml',
            'X500 email',
            'saml-user-property-mapper',
            emailMapperConfig
        );

        await Keycloak.AddMapperToClient(
            realmName,
            client.id,
            'saml',
            'GROUPS',
            'saml-group-membership-mapper',
            groupMapperConfig
        );
        await Keycloak.AddMapperToClient(
            realmName,
            client.id,
            'saml',
            'firstName',
            'saml-user-property-mapper',
            firstNameMapperConfig
        );
        await Keycloak.AddMapperToClient(
            realmName,
            client.id,
            'saml',
            'lastName',
            'saml-user-property-mapper',
            lastNameMapperConfig
        );
    }

    static async setupSSOOCP(): Promise<void> {
        const clinetId = `my-client${Date.now()}`;
        const idpSecret = `this-is-my-secret${Date.now()}`;
        await Keycloak.configLoginWithSSOOCP(clinetId, idpSecret);
        await Keycloak.changeLoginTheme('runai', 'runai-theme');
    }

    static async changeExpiredTokenLifespan(realmName: string, lifespan: number): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const lifespanInSeconds = lifespan * 60;
        try {
            await keycloakAdminClient.realms.update(
                { realm: realmName },
                {
                    accessTokenLifespan: lifespanInSeconds,
                    accessTokenLifespanForImplicitFlow: lifespanInSeconds
                }
            );
        } catch (error) {
            logger.error('Failed to update token lifespan:', error);
        }
    }

    static async updateStoreTokensSetting(
        realmName: string,
        idpName: string,
        storeTokens: boolean
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const currentSettings = await keycloakAdminClient.identityProviders.findOne({
            realm: realmName,
            alias: idpName
        });

        if (currentSettings) {
            logger.debug('Current settings before update:', currentSettings.config);

            currentSettings.storeToken = storeTokens;
            logger.debug('Updated settings:', currentSettings.config);

            await keycloakAdminClient.identityProviders.update(
                { realm: realmName, alias: idpName },
                currentSettings
            );
            logger.debug(`Store tokens setting updated to: ${storeTokens}`);
        } else {
            throw new Error('Identity Provider settings not found');
        }
    }

    static async addRealmSettingsUserProfileAttribute(
        realmName: string,
        attributeName: string,
        displayName: string
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const token = keycloakAdminClient.accessToken;
        if (!token) {
            throw new Error('❌ Failed to retrieve Keycloak Admin Token');
        }
        const keycloakBaseUrl = `${config.tenantUrl}/auth/admin/realms/${realmName}/users/profile`;

        try {
            const existingProfile = await axios.get(keycloakBaseUrl, {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });
            const updatedAttributes = existingProfile.data.attributes || [];
            const newAttribute = {
                name: attributeName,
                displayName: displayName,
                multivalued: true,
                validations: {},
                annotations: {},
                permissions: {
                    view: [],
                    edit: ['admin']
                }
            };
            updatedAttributes.push(newAttribute);
            const updatePayload = {
                attributes: updatedAttributes,
                groups: existingProfile.data.groups || [
                    {
                        name: 'user-metadata',
                        displayHeader: 'User Metadata',
                        displayDescription: 'Attributes related to user metadata'
                    }
                ]
            };
            await axios.put(keycloakBaseUrl, updatePayload, {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            logger.debug(`Attribute '${attributeName}' added successfully in realm '${realmName}'`);
        } catch (error) {
            logger.error(
                `Failed to add attribute '${attributeName}' in realm '${realmName}'`,
                JSON.stringify(error)
            );
        }
    }

    static async createUserWithClaim(
        realmName: string,
        username: string,
        claim: string,
        value: string,
        password: string
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.users.create({
            realm: realmName,
            username: username,
            enabled: true,
            emailVerified: true,
            email: `${username}@example.com`,
            firstName: username,
            lastName: username,
            attributes: { [claim]: [value] },
            credentials: [
                {
                    type: 'password',
                    temporary: false,
                    value: password
                }
            ]
        });
    }

    static async updateUserDetailsGeneral(realmName: string, username: string, details: any): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const users = await keycloakAdminClient.users.find({ realm: realmName, username });
        if (!users || users.length === 0) {
            logger.error(`User '${username}' not found in realm '${realmName}'`);
            return;
        }
        const userId = users[0].id;
        const userAttributes = users[0].attributes || {};
        if (!details.attributes) {
            details.attributes = {};
        }
        if (details.attributes.groupsAttribute) {
            details.attributes.groupsAttribute = [details.attributes.groupsAttribute];
        }
        details.attributes = { ...userAttributes, ...details.attributes };
        try {
            await keycloakAdminClient.users.update({ realm: realmName, id: userId }, { ...details });
        } catch (error) {
            logger.error(
                `Failed to update user details for user '${username}' in realm '${realmName}' error: ${JSON.stringify(error)}`
            );
        }
    }

    static async addUserProfileAttributes(
        realmName: string,
        attributes: Array<{ name: string; displayName: string; multivalued?: boolean }>
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const token = keycloakAdminClient.accessToken;
        if (!token) {
            throw new Error('Failed to retrieve Keycloak Admin Token');
        }

        const keycloakBaseUrl = `${config.tenantUrl}/auth/admin/realms/${realmName}/users/profile`;
        const existingProfile = await axios.get(keycloakBaseUrl, {
            headers: {
                Authorization: `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });

        const updatedAttributes = existingProfile.data.attributes || [];
        const newAttributes = attributes.map((attr) => ({
            name: attr.name,
            displayName: attr.displayName,
            multivalued: attr.multivalued ?? true,
            validations: {},
            annotations: {},
            permissions: {
                view: [],
                edit: ['admin']
            }
        }));

        updatedAttributes.push(...newAttributes);
        await axios.put(
            keycloakBaseUrl,
            {
                attributes: updatedAttributes,
                groups: existingProfile.data.groups || [
                    {
                        name: 'user-metadata',
                        displayHeader: 'User Metadata',
                        displayDescription: 'Attributes related to user metadata'
                    }
                ]
            },
            {
                headers: {
                    Authorization: `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );

        logger.info(`Added ${attributes.length} attributes to realm '${realmName}' user profile`);
    }

    static async addProtocolMappers(
        realmName: string,
        clientId: string,
        mappers: Array<{
            name: string;
            protocol: string;
            protocolMapper: string;
            config: Record<string, string>;
        }>
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.findOne({ realm: realmName, id: clientId });

        if (!client) {
            throw new Error(`Client '${clientId}' not found in realm '${realmName}'`);
        }

        for (const mapper of mappers) {
            await keycloakAdminClient.clients.addProtocolMapper({ realm: realmName, id: client.id }, mapper);
        }

        logger.info(
            `Added ${mappers.length} protocol mappers to client '${clientId}' in realm '${realmName}'`
        );
    }

    static async createUserWithAttributes(
        realmName: string,
        username: string,
        password: string,
        attributes: Record<string, string>
    ): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        await keycloakAdminClient.users.create({
            realm: realmName,
            username: username,
            enabled: true,
            emailVerified: true,
            email: `${username}@example.com`,
            firstName: username,
            lastName: username,
            attributes: attributes,
            credentials: [
                {
                    type: 'password',
                    temporary: false,
                    value: password
                }
            ]
        });

        logger.info(`Created user '${username}' with attributes in realm '${realmName}'`);
    }

    /**
     * Direct authentication to Keycloak to get tokens
     * @param username - Username to authenticate
     * @param password - Password for the user
     * @param realmName - Keycloak realm name (default: 'automation')
     * @param clientId - Client ID (default: 'automationclient')
     * @param clientSecret - Client secret (required when client authentication is enabled)
     * @param keycloakBaseUrl - Keycloak base URL (optional, uses config.tenantUrl by default)
     * @returns Promise<KeycloakTokenResponse>
     */
    static async getTokenDirect(
        username: string,
        password: string,
        realmName: string = 'automation',
        clientId: string = 'automationclient',
        clientSecret?: string,
        keycloakBaseUrl?: string
    ): Promise<KeycloakTokenResponse> {
        const baseUrl = keycloakBaseUrl || config.tenantUrl;
        const tokenEndpoint = `${baseUrl}/auth/realms/${realmName}/protocol/openid-connect/token`;

        const tokenRequest: any = {
            grant_type: 'password',
            client_id: clientId,
            username: username,
            password: password,
            scope: 'openid email profile'
        };

        // Add client secret if provided (required for confidential clients)
        if (clientSecret) {
            tokenRequest.client_secret = clientSecret;
        }

        const requestConfig = {
            method: 'POST' as const,
            headers: {
                'content-type': 'application/x-www-form-urlencoded',
                Accept: 'application/json'
            },
            data: qs.stringify(tokenRequest),
            url: tokenEndpoint
        };

        try {
            logger.debug(`Authenticating user ${username} to Keycloak realm ${realmName}`);
            const tokenResponse = await axios(requestConfig);

            if (tokenResponse.status !== StatusCodes.OK) {
                throw new Error(`Authentication failed with status: ${tokenResponse.status}`);
            }

            logger.debug(`Successfully authenticated user ${username} to Keycloak`);
            return tokenResponse.data as KeycloakTokenResponse;
        } catch (error) {
            logger.error(`Failed to authenticate user ${username} to Keycloak: ${error.message}`);
            throw new Error(`Keycloak authentication failed: ${error.message}`);
        }
    }

    static async getApplicationDirectToken(
        realmName: string,
        clientId: string,
        clientSecret: string
    ): Promise<KeycloakTokenResponse> {
        const baseUrl = config.tenantUrl;
        const tokenEndpoint = `${baseUrl}/auth/realms/${realmName}/protocol/openid-connect/token`;

        const tokenRequest: any = {
            grant_type: 'client_credentials',
            client_id: clientId,
            client_secret: clientSecret,
            scope: 'openid'
        };

        const requestConfig = {
            method: 'POST' as const,
            headers: {
                'content-type': 'application/x-www-form-urlencoded',
                Accept: 'application/json'
            },
            data: qs.stringify(tokenRequest),
            url: tokenEndpoint
        };

        const tokenResponse = await axios(requestConfig);
        return tokenResponse.data as KeycloakTokenResponse;
    }

    static async deleteClient(realmName: string, clientId: string): Promise<void> {
        const keycloakAdminClient = await Keycloak.initKeycloakAdminClient();
        const client = await keycloakAdminClient.clients.find({ realm: realmName, clientId: clientId });
        if (client && client.length > 0) {
            await keycloakAdminClient.clients.del({ realm: realmName, id: client[0].id });
        }
    }

    /**
     * Sets up an external Keycloak realm suitable for testing access-key uid/gid/supplementarygroups
     * claims. Creates the realm, client, user-profile attribute schema, protocol mappers, and a
     * single user whose UID/GID/SUPPLEMENTARYGROUPS attributes match the provided values.
     * Applies the runai-theme and registers the realm as the active OIDC IDP in the Run.ai system.
     *
     * Returns the generated client secret so the caller can pass it to idpUtils.updateIdpOidc.
     */
    static async setupSSORealmOIDCForUidGidClaims(
        realmName: string,
        clientName: string,
        username: string,
        password: string,
        uid: string,
        gid: string,
        supplementaryGroups: string[]
    ): Promise<string> {
        await Keycloak.createRealm(realmName);
        const client = await Keycloak.createClient(realmName, clientName, 'openid-connect');
        const clientSecret = await Keycloak.getClientSecret(realmName, client.id);

        const baseMapperConfig = {
            'full.path': 'false',
            'id.token.claim': 'true',
            'access.token.claim': 'true',
            'userinfo.token.claim': 'true',
            'jsonType.label': 'String'
        };
        await Keycloak.addProtocolMappers(realmName, client.id, [
            {
                name: 'UID',
                protocol: 'openid-connect',
                protocolMapper: 'oidc-usermodel-attribute-mapper',
                config: { ...baseMapperConfig, 'user.attribute': 'UID', 'claim.name': 'UID' }
            },
            {
                name: 'GID',
                protocol: 'openid-connect',
                protocolMapper: 'oidc-usermodel-attribute-mapper',
                config: { ...baseMapperConfig, 'user.attribute': 'GID', 'claim.name': 'GID' }
            },
            {
                name: 'SUPPLEMENTARYGROUPS',
                protocol: 'openid-connect',
                protocolMapper: 'oidc-usermodel-attribute-mapper',
                config: {
                    ...baseMapperConfig,
                    multivalued: 'true',
                    'user.attribute': 'SUPPLEMENTARYGROUPS',
                    'claim.name': 'SUPPLEMENTARYGROUPS'
                }
            }
        ]);

        await Keycloak.addUserProfileAttributes(realmName, [
            { name: 'UID', displayName: 'UID', multivalued: false },
            { name: 'GID', displayName: 'GID', multivalued: false },
            { name: 'SUPPLEMENTARYGROUPS', displayName: 'SUPPLEMENTARYGROUPS', multivalued: true }
        ]);

        await Keycloak.createUserWithAttributes(realmName, username, password, {
            UID: uid,
            GID: gid,
            SUPPLEMENTARYGROUPS: supplementaryGroups
        } as any);

        await Keycloak.changeLoginTheme(realmName, 'runai-theme');
        await Keycloak.configLoginWithSSOOIDC(
            Keycloak.getDiscoveryDocumentURL(realmName),
            clientName,
            clientSecret
        );

        return clientSecret;
    }
}
