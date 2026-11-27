/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { identityManagerApi } from '../../api';
import {
    UserCreationRequest,
    UserPostResponse,
    User,
    TokenRequestGrantTypeEnum,
    TokenRequest,
    TokenResponse,
    BulkUserCreationRequest
} from '../../api/apis/identityManagerApi/generated/api';
import http from 'http';
import { flowRootState } from '../config/config';
import { authorizationUtils, authUtils, objectGenerator, logger, Utils } from '..';
import { User as FlowUser, Users } from '../config/types';
import { SubjectType, ScopeType, Role } from '../../api/apis/authorizationApi/generated/api';
import { Keycloak } from './Keycloak';
import { config } from '../config/config';
import { idpNameOidc } from '../config/const';
import { PROPERTY_TO_ROLE_MAPPING } from '../config/roleMapping';
import { StatusCodes } from 'http-status-codes';

export class UsersManager {
    private ssoUsersWarnedNoRefreshToken = new Set<string>();

    async getUserById(accessToken: string, userId: string): Promise<User> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const userResponse = await identityManagerApi.usersApi.getUserById(userId);
        return userResponse.body;
    }

    async createLocalUser(
        accessToken: string,
        email: string,
        toResetPassword: boolean = false,
        notify?: boolean
    ): Promise<{ response: http.IncomingMessage; body: UserPostResponse }> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const userCreationRequest: UserCreationRequest = {
            email: email,
            resetPassword: toResetPassword,
            notify: notify
        };

        try {
            logger.debug(`Calling identity manager API to create user: ${email}`);
            const userCreationResponse = await identityManagerApi.usersApi.createUser(userCreationRequest);
            logger.debug(
                `Identity manager API response for ${email}: status ${userCreationResponse.response.statusCode}`
            );
            return {
                response: userCreationResponse.response,
                body: userCreationResponse.body
            };
        } catch (error) {
            logger.error(`Identity Manager API failed for user ${email}:`, {
                message: error.message,
                status: error.response?.status || error.statusCode,
                statusText: error.response?.statusText,
                responseBody: error.response?.body || error.body,
                requestBody: userCreationRequest,
                url: identityManagerApi.usersApi.basePath || 'unknown',
                hasAccessToken: !!accessToken
            });
            throw error;
        }
    }

    async createBulkLocalUsers(
        accessToken: string,
        emails: string[],
        toResetPassword: boolean = false,
        notify?: boolean
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const userCreationRequest: BulkUserCreationRequest = {
            users: emails.map((email) => ({ email: email, resetPassword: toResetPassword, notify: notify })),
            notify: notify
        };
        const userCreationResponse = await identityManagerApi.usersApi.batchCreateUsers(userCreationRequest);
        return {
            response: userCreationResponse.response,
            body: userCreationResponse.body
        };
    }

    async deleteLocalUser(
        accessToken: string,
        userId: string
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const userDeletionResponse = await identityManagerApi.usersApi.deleteUserById(userId);
        return {
            response: userDeletionResponse.response,
            body: userDeletionResponse.body
        };
    }

    async logoutUser(
        accessToken: string,
        userId: string
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const logoutResponse = await identityManagerApi.usersApi.logoutUser(userId);
        return {
            response: logoutResponse.response,
            body: logoutResponse.body
        };
    }

    async getUsers(accessToken: string): Promise<User[]> {
        identityManagerApi.usersApi.accessToken = accessToken;
        return (await identityManagerApi.usersApi.getUsers()).body;
    }

    async refreshToken(refreshToken: string, clientID?: string): Promise<TokenResponse> {
        const tokenRequest: TokenRequest = {
            grantType: TokenRequestGrantTypeEnum.RefreshToken,
            refreshToken: refreshToken,
            clientID: clientID
        };
        const response = await identityManagerApi.tokensApi.grantToken(tokenRequest);
        return response.body;
    }

    async createSSOUser(
        username: string,
        password?: string,
        realmName: string = 'automation'
    ): Promise<void> {
        const userPassword = password || config.controlPlaneAdminPassword;
        logger.debug(`Creating SSO user: ${username} in realm: ${realmName}`);
        await Keycloak.createUser(realmName, username, userPassword);
        logger.info(`SSO user ${username} created successfully in realm ${realmName}`);
    }

    async deleteSSOUser(
        accessToken: string,
        username: string,
        runaiUserId: string,
        realmName: string = 'automation'
    ): Promise<{ response: http.IncomingMessage; body: any }> {
        logger.debug(`Deleting SSO user: ${username} from both realms: ${realmName} and runai`);
        await Keycloak.deleteUserFromRealm(realmName, username);
        await Keycloak.deleteUserFromRealm('runai', username);
        logger.info(`SSO user ${username} deleted successfully from both realms: ${realmName} and runai`);
        identityManagerApi.usersApi.accessToken = accessToken;
        const userDeletionResponse = await identityManagerApi.usersApi.deleteUserById(runaiUserId);
        logger.info(`SSO user ${username} deleted successfully from identity manager`);
        return {
            response: userDeletionResponse.response,
            body: userDeletionResponse.body
        };
    }

    async updateSSOUserAttributes(
        currentEmail: string,
        updates: {
            username?: string;
            email?: string;
            firstName?: string;
            lastName?: string;
            attributes?: { [key: string]: string[] | string };
        },
        realmName: string = 'automation'
    ): Promise<void> {
        logger.debug(`Updating SSO user attributes for: ${currentEmail} in realm: ${realmName}`);
        const details: any = {};
        if (updates.username) {
            details.username = updates.username;
        }
        if (updates.email) {
            details.email = updates.email;
        }
        if (updates.firstName) {
            details.firstName = updates.firstName;
        }
        if (updates.lastName) {
            details.lastName = updates.lastName;
        }

        if (updates.attributes) {
            details.attributes = {};
            for (const [key, value] of Object.entries(updates.attributes)) {
                if (key === 'groupsAttribute') {
                    details.attributes[key] = Array.isArray(value) ? value[0] : value;
                } else {
                    details.attributes[key] = Array.isArray(value) ? value : [value];
                }
            }
        }

        try {
            const username = currentEmail.split('@')[0];

            await Keycloak.updateUserDetailsGeneral(realmName, username, details);

            const updatedFields = Object.keys(updates).join(', ');
            logger.info(
                `SSO user ${currentEmail} updated successfully in realm ${realmName}. Updated fields: ${updatedFields}`
            );
        } catch (error) {
            logger.error(`Failed to update SSO user ${currentEmail} in realm ${realmName}:`, error);
            throw error;
        }
    }

    async initLocalUsers(accessToken: string): Promise<Users> {
        const users: Users = {};
        const roles = await authorizationUtils.getRoles(accessToken);

        await Utils.parallelMap(
            8,
            Object.entries(PROPERTY_TO_ROLE_MAPPING),
            async ([propertyName, roleName]) => {
                const role = roles.find((r) => r.name === roleName);
                if (!role) {
                    logger.warn(`Role '${roleName}' not found for property '${propertyName}'`);
                    return;
                }
                try {
                    const user = await this.createUserWithRole(accessToken, role, propertyName);
                    users[propertyName] = user;
                } catch (error) {
                    if (error.response?.status === 409 || error.statusCode === 409) {
                        logger.info(`User for ${propertyName} already exists, skipping creation`);
                        return;
                    } else {
                        logger.error(`Failed to create user for ${propertyName} (${roleName}):`, {
                            message: error.message || 'No error message',
                            status: error.response?.status || error.statusCode,
                            statusText: error.response?.statusText,
                            body: error.response?.body || error.body,
                            stack: error.stack,
                            fullError: error
                        });
                    }
                }
            }
        );
        await this.authenticateAllLocalUsers(users, 'local users');
        flowRootState.users = users;
        return users;
    }

    async createUserWithRole(
        accessToken: string,
        role: Role,
        emailPrefix: string,
        scopeId: string = flowRootState.cluster.tenantId.toString(),
        scopeType: ScopeType = ScopeType.Tenant
    ): Promise<FlowUser> {
        const email = objectGenerator.generateEmailByPrefix(emailPrefix);
        logger.debug(`Creating local user with email: ${email} for role: ${role.name}`);

        try {
            const [userResponse, accessRule] = await Promise.all([
                this.createLocalUser(accessToken, email, false),
                authorizationUtils.createAccessRuleWithDefaults(
                    email,
                    SubjectType.User,
                    role.id,
                    scopeId,
                    scopeType,
                    accessToken
                )
            ]);

            logger.debug(`Successfully created local user: ${email}, userId: ${userResponse.body.id}`);
            return {
                email: email,
                password: userResponse.body.tempPassword,
                userId: userResponse.body.id,
                role: role,
                accessRules: [accessRule]
            };
        } catch (error) {
            logger.error(`Error in createUserWithRole for ${email}:`, {
                message: error.message,
                status: error.response?.status || error.statusCode,
                responseBody: error.response?.body || error.body,
                role: role.name,
                emailPrefix: emailPrefix
            });
            throw error;
        }
    }

    async authenticateAllLocalUsers(users: Users, logContext = 'local users'): Promise<void> {
        logger.info(`Authenticating ${logContext}`);
        const authPromises = Object.entries(users).map(async ([propertyName, user]: [string, any]) => {
            try {
                if (user.refreshToken) {
                    try {
                        const tokenResponse = await this.refreshToken(user.refreshToken, 'cli');
                        user.idToken = tokenResponse.idToken;
                        user.refreshToken = tokenResponse.refreshToken ?? user.refreshToken;
                        user.accessToken = tokenResponse.accessToken;
                        logger.debug(`User ${user.email} (${propertyName}) refreshed successfully using refresh token`);
                        return;
                    } catch (refreshError) {
                        logger.warn(
                            `Refresh token failed for ${user.email} (${propertyName}), falling back to password auth: ${refreshError.message}`
                        );
                    }
                }
                const token = await authUtils.authenticateUser(user.email, user.password);
                user.idToken = token.idToken;
                user.refreshToken = token.refreshToken;
                user.accessToken = token.accessToken;
                logger.debug(`User ${user.email} (${propertyName}) authenticated successfully`);
            } catch (error) {
                logger.error(`Failed to authenticate user ${user.email} (${propertyName}):`, error.message);
            }
        });
        await Promise.all(authPromises);
    }

    async authenticateAllExistingSSOUsers(users: Users): Promise<void> {
        const refreshPromises = Object.entries(flowRootState.usersSSO).map(
            async ([propertyName, user]: [string, any]) => {
                try {
                    if (!user.refreshToken) {
                        if (!this.ssoUsersWarnedNoRefreshToken.has(propertyName)) {
                            this.ssoUsersWarnedNoRefreshToken.add(propertyName);
                            logger.warn(
                                `SSO user ${user.email} (${propertyName}) has no refresh token, skipping`
                            );
                        }
                        return;
                    }

                    let tokenResponse: TokenResponse;
                    try {
                        tokenResponse = await this.refreshTokenForSSOUsersWithOutUiAuth(
                            user.refreshToken,
                            'runai-admin-ui'
                        );
                    } catch (firstError) {
                        if (firstError.message === 'INVALID_REFRESH_TOKEN') {
                            logger.debug(`Trying fallback client 'runai-cli' for SSO user ${user.email}`);
                            tokenResponse = await this.refreshTokenForSSOUsersWithOutUiAuth(
                                user.refreshToken,
                                'runai-cli'
                            );
                        } else {
                            throw firstError;
                        }
                    }

                    user.idToken = tokenResponse.idToken;
                    user.accessToken = tokenResponse.accessToken;
                    if (tokenResponse.refreshToken) {
                        user.refreshToken = tokenResponse.refreshToken;
                    }
                    logger.debug(
                        `SSO user ${user.email} (${propertyName}) refreshed successfully using refresh token API`
                    );
                } catch (error) {
                    if (error.message === 'INVALID_REFRESH_TOKEN') {
                        logger.debug(
                            `SSO user ${user.email} (${propertyName}) refresh failed - tokens expired or invalid. User will be skipped.`
                        );
                        user.isTokenInvalid = true;
                    } else {
                        logger.error(
                            `Failed to refresh SSO user ${user.email} (${propertyName}) using refresh token API:`,
                            error.message
                        );
                    }
                }
            }
        );

        await Promise.all(refreshPromises);
        logger.info('All SSO users authenticated successfully');
    }

    async initSSOUsers(accessToken: string): Promise<Users> {
        logger.info('Initializing SSO users');
        const usersSSO: Users = {};
        const realmName = 'automation';
        const roles = await authorizationUtils.getRoles(accessToken);

        for (const [propertyName, roleName] of Object.entries(PROPERTY_TO_ROLE_MAPPING)) {
            const role = roles.find((r) => r.name === roleName);
            if (!role) {
                logger.warn(`Role '${roleName}' not found for property '${propertyName}'`);
                continue;
            }
            try {
                const user = await this.createSSOUserWithRole(accessToken, role, propertyName, realmName);
                usersSSO[propertyName] = user;
                logger.debug(`Created SSO user: ${propertyName} (${roleName})`);
            } catch (error) {
                logger.error(`Failed to create SSO user ${propertyName}:`, error.message);
            }
        }
        flowRootState.usersSSO = usersSSO;
        logger.info(`SSO users initialized: ${Object.keys(usersSSO).length} users created`);
        return usersSSO;
    }

    private async authenticateSSOUsersViaUI(): Promise<void> {
        logger.info('Authenticating SSO users via UI to get RunAI tokens Start');
        logger.info(`Found ${Object.keys(flowRootState.usersSSO).length} SSO users to authenticate`);

        for (const [propertyName, user] of Object.entries(flowRootState.usersSSO)) {
            try {
                logger.debug(`Performing UI SSO login for ${propertyName} (${user.email})`);
                const tokenObject = await authUtils.authenticateUser(user.email, user.password, true);
                user.idToken = tokenObject.idToken;
                user.refreshToken = tokenObject.refreshToken;
                user.accessToken = tokenObject.accessToken;
                logger.debug(`SSO user ${propertyName} authenticated successfully with RunAI token`);
            } catch (error) {
                logger.error(`Failed to authenticate SSO user ${propertyName} (${user.email}):`, {
                    message: error.message || 'Unknown error',
                    name: error.name || 'Unknown',
                    stack: error.stack || 'No stack trace',
                    details: error
                });
            }
        }
        logger.info('Authenticating SSO users via UI to get RunAI tokens End');
    }

    private async createSSOUserWithRole(
        accessToken: string,
        role: any,
        propertyName: string,
        realmName: string
    ): Promise<any> {
        const timestamp = Date.now();
        const username = `${propertyName}_${timestamp}`;
        const email = `${username}@example.com`;
        const password = config.controlPlaneAdminPassword;
        logger.debug(`Creating SSO user: ${username} for role: ${role.name}`);
        await Keycloak.createUser(realmName, username, password);
        const accessRule = await authorizationUtils.createAccessRuleWithDefaults(
            email,
            SubjectType.User,
            role.id,
            flowRootState.cluster.tenantId.toString(),
            ScopeType.Tenant,
            accessToken
        );
        return {
            email: email,
            password: password,
            userId: `sso-${username}`,
            role: role,
            accessRules: [accessRule]
        };
    }

    async initSSOFlowUsers(accessToken: string): Promise<void> {
        logger.info('Starting SSO initialization flow');
        try {
            await this.createOpenIdConnectSSO();
            await this.initSSOUsers(accessToken);
            await this.authenticateSSOUsersViaUI();
            logger.info('SSO initialization flow completed successfully');
        } catch (error) {
            logger.error('SSO initialization flow failed:', error.message);
            throw error;
        }
    }

    async deletePreSeededOidcIdp(): Promise<void> {
        if (process.env.INIT_SSO_USERS_AND_IDP !== 'true') {
            return;
        }
        await this.cleanupExistingIDP(idpNameOidc);
    }

    async restorePreSeededOidcIdp(): Promise<void> {
        if (process.env.INIT_SSO_USERS_AND_IDP !== 'true') {
            return;
        }
        await this.createOpenIdConnectSSO();
    }

    private async createOpenIdConnectSSO(): Promise<void> {
        logger.info('Starting OpenIdConnect SSO setup');
        const realmName = 'automation';
        const clientName = 'automationclient';
        const idpName = idpNameOidc;
        await this.cleanupExistingRealm(realmName);
        await this.cleanupExistingIDP(idpName);
        await Keycloak.setupSSORealmOIDC(realmName, clientName);
        await Keycloak.addRealmSettingsUserProfileAttribute(realmName, 'groupsAttribute', 'groupAttribute');
        logger.info(`OpenIdConnect SSO setup completed for realm: ${realmName}`);
    }

    private async cleanupExistingRealm(realmName: string): Promise<void> {
        try {
            logger.debug(`Checking if realm '${realmName}' exists`);
            const existingRealm = await Keycloak.getRealmByName(realmName);

            if (existingRealm) {
                logger.info(`Found existing realm '${realmName}', deleting it to avoid conflicts`);
                await Keycloak.deleteRealm(realmName);
                logger.info(`Successfully deleted existing realm '${realmName}'`);
            } else {
                logger.debug(`No existing realm '${realmName}' found, proceeding with creation`);
            }
        } catch (error) {
            logger.warn(`Error during realm cleanup for '${realmName}': ${error.message}`);
        }
    }

    private async cleanupExistingIDP(idpName: string): Promise<void> {
        try {
            logger.debug(`Checking if IDP '${idpName}' exists in 'runai' realm`);
            const existingIdp = await Keycloak.getIdpFromRealm('runai', idpName);

            if (existingIdp) {
                logger.info(`Found existing IDP '${idpName}', deleting it to avoid conflicts`);
                await Keycloak.deleteIDPFromRealm('runai', idpName);
                logger.info(`Successfully deleted existing IDP '${idpName}'`);
            } else {
                logger.debug(`No existing IDP '${idpName}' found, proceeding with creation`);
            }
        } catch (error) {
            logger.warn(`Error during IDP cleanup for '${idpName}': ${error.message}`);
        }
    }

    async addSSOUserToGroupWithRole(
        accessToken: string,
        email: string,
        roleName: Role,
        realmName: string,
        groupName: string
    ): Promise<void> {
        try {
            if (!email.includes('@')) {
                throw new Error(`Invalid email format: ${email}`);
            }

            const username = email.split('@')[0];
            logger.debug(`Adding SSO user ${email} to group ${groupName} with role ${roleName.name}`);

            await Keycloak.createKeycloakGroupsWithUsers(realmName, [groupName], [username]);
            await authorizationUtils.createAccessRuleForGroup(accessToken, groupName, roleName.id);
            logger.info(
                `SSO user ${email} added to group ${groupName} with role ${roleName.name} in realm ${realmName}`
            );
        } catch (error) {
            logger.error(`Failed to add SSO user ${email} to group ${groupName}:`, error);
            throw error;
        }
    }

    async addSSOUserToGroup(email: string, groupName: string, realmName: string): Promise<void> {
        const username = email.split('@')[0];
        logger.debug(`Adding SSO user ${email} to group ${groupName} in realm ${realmName}`);
        await Keycloak.joinUserToGroup(realmName, username, groupName);
        logger.info(`SSO user ${email} added to group ${groupName} in realm ${realmName}`);
    }

    async createGroup(groupName: string, realmName: string): Promise<void> {
        logger.debug(`Creating group ${groupName} in realm ${realmName}`);
        await Keycloak.addKeycloakGroup(realmName, groupName);
        logger.info(`Group ${groupName} created in realm ${realmName}`);
    }

    async removeSSOUserFromGroup(email: string, groupName: string, realmName: string): Promise<void> {
        const username = email.split('@')[0];
        logger.debug(`Removing SSO user ${email} from group ${groupName} in realm ${realmName}`);
        await Keycloak.removeUserFromGroup(realmName, username, groupName);
        logger.info(`SSO user ${email} removed from group ${groupName} in realm ${realmName}`);
    }

    async refreshTokenForSSOUsersWithOutUiAuth(
        refreshToken: string,
        clientId: string
    ): Promise<TokenResponse> {
        const tokenRequest: TokenRequest = {
            grantType: TokenRequestGrantTypeEnum.RefreshToken,
            refreshToken: refreshToken,
            clientID: clientId
        };
        try {
            const response = await identityManagerApi.tokensApi.grantToken(tokenRequest);
            return response.body;
        } catch (error) {
            if (
                error.response?.status === StatusCodes.BAD_REQUEST ||
                error.statusCode === StatusCodes.BAD_REQUEST
            ) {
                logger.debug(
                    `SSO refresh token failed (400) with client: ${clientId} - token expired/invalid/client mismatch  error: ${error.message}`
                );
                throw new Error('INVALID_REFRESH_TOKEN');
            } else {
                logger.error(`Failed to refresh tokens using refresh token API: ${error.message}`);
                throw new Error(`Token refresh failed: ${error.message}`);
            }
        }
    }
}
