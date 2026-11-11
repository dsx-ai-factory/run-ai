/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    settings,
    TokenObject,
    usersUtils,
    UtilsError,
    versionUtils,
    Utils
} from '..';
import { ensureControlPlaneCredentials } from './CredentialsResolver';
import { usersManager } from '../index';
import { AuthEntityType, Role, UserCreationRequest } from '../../api/apis/backendApi';
import { identityManagerApi } from '../../api';
import { runaiV1 } from '../../ui';
import { expect } from 'chai';
import jwt_decode from 'jwt-decode';
import {
    TokenRequest,
    TokenResponse,
    UsersFilterSortFields,
    User,
    TokensApi,
    TokenRequestGrantTypeEnum
} from '../../api/apis/identityManagerApi/generated/api';
import { IdentityManagerApi } from '../../api/apis/identityManagerApi';
import axios from 'axios';
import qs from 'qs';
import fs from 'fs';
import { StatusCodes } from 'http-status-codes';

export enum CleanupMethod {
    DeleteAllUsersBesidesSpecifiedUsers,
    DeleteOnlyUsersWithSubString
}

export class AuthUtils {
    private static isTokenRefreshing = false; // Control variable for token refresh state

    async getUserId(email: string, accessToken: string, isLocal?: boolean): Promise<string> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const user = await usersUtils.getUsersFilteredBy(
            accessToken,
            [`isLocal==${isLocal}`],
            UsersFilterSortFields.CreationTime,
            'desc',
            0,
            5,
            email
        );
        if (!user) {
            throw new Error('Required parameter user was null or undefined when calling UsersGet.');
        } else {
            return user[0].id;
        }
    }

    async getUserV2(username: string, accessToken: string): Promise<User> {
        identityManagerApi.usersApi.accessToken = accessToken;
        try {
            const users = await identityManagerApi.usersApi.getUsers();
            const user = users.body.find((user) => user.username === username);
            if (!user) {
                throw new Error('Required parameter user was null or undefined when calling UsersGet.');
            }
            return user;
        } catch (error) {
            throw new Error(`getUser request failed ${error.statusCode}`);
        }
    }

    private isControlPlaneAdminEmail(email: string): boolean {
        return email === config.controlPlaneAdminUsername;
    }

    private async resolveAdminPasswordIfNeeded(email: string, password?: string): Promise<string> {
        if (settings.isSelfHosted && this.isControlPlaneAdminEmail(email)) {
            await ensureControlPlaneCredentials();
            if (config.controlPlaneAdminPassword) {
                return config.controlPlaneAdminPassword;
            }
        }
        if (password) {
            return password;
        }
        throw new Error(`Missing password for user ${email}`);
    }

    async authenticateUser(email: string, password?: string, uiSSOAuth?: boolean): Promise<TokenObject> {
        try {
            const resolvedPassword = await this.resolveAdminPasswordIfNeeded(email, password);
            if (uiSSOAuth) {
                return await this.authenticateUIAndGetToken(email, resolvedPassword, uiSSOAuth);
            }
            return await this.authenticateKeycloakAndGetToken(email, resolvedPassword);
        } catch (error) {
            logger.error(`unable to authenticate user ${email}, error: ${JSON.stringify(error)}`);
            throw new UtilsError(error);
        }
    }

    async authenticateSSOOCPUser(email: string, password: string): Promise<TokenObject> {
        try {
            return await this.authenticateUIOCPSSO(email, password);
        } catch (error) {
            logger.error(`unable to authenticate user ${email}, error: ${JSON.stringify(error)}`);
            throw new UtilsError(error);
        }
    }

    async authenticateAppAndGetToken(clientId: string, clientSecret: string): Promise<TokenResponse> {
        const appTokenRequest: TokenRequest = {
            clientID: clientId,
            clientSecret: clientSecret,
            grantType: TokenRequestGrantTypeEnum.ClientCredentials
        };
        try {
            const appTokenResponse = await identityManagerApi.tokensApi.grantToken(appTokenRequest);
            return appTokenResponse.body;
        } catch (error) {
            throw new Error(`Failed to obtain token+ ${error}`);
        }
    }

    async authenticateSuperAdminAppAndGetToken(
        clientId: string,
        clientSecret: string,
        url: string
    ): Promise<TokenResponse> {
        const identityManagerApi = new IdentityManagerApi(url);
        const appTokenRequest: TokenRequest = {
            clientID: clientId,
            clientSecret: clientSecret,
            grantType: TokenRequestGrantTypeEnum.ClientCredentials
        };
        try {
            const appTokenResponse = await identityManagerApi.tokensApi.grantToken(appTokenRequest);
            return appTokenResponse.body;
        } catch (error) {
            throw new Error(`Failed to obtain token+ ${error}`);
        }
    }

    async authenticateExternalTokenExchange(externalToken: string): Promise<TokenResponse> {
        const tokenRequest: any = {
            grantType: TokenRequestGrantTypeEnum.ExternalTokenExchange,
            externalToken: externalToken,
            scope: 'openid email profile'
        };

        try {
            const tokenResponse = await identityManagerApi.tokensApi.grantToken(tokenRequest);
            return tokenResponse.body;
        } catch (error) {
            throw new Error(`Failed to exchange external token: ${error}`);
        }
    }

    async authenticateTokenApi(
        userName?: string,
        password?: string,
        url?: string,
        accessToken?: string
    ): Promise<{ response: any; body: TokenResponse }> {
        let tokenApi = identityManagerApi.tokensApi;
        if (url) {
            tokenApi = new TokensApi(url);
            tokenApi.accessToken = accessToken;
        }
        const tokenRequest: TokenRequest = {
            grantType: TokenRequestGrantTypeEnum.Password,
            username: userName,
            password: password
        };
        const response = await tokenApi.grantToken(tokenRequest);
        logger.debug(`response: ${JSON.stringify(response)}`);
        return response;
    }

    async authenticateKeycloakAndGetToken(userName: string, password: string): Promise<TokenObject> {
        logger.debug(`authenticate user by Keycloak ${userName}`);
        const keycloakToken = await this.authenticateKeycloak(userName, password);
        logger.debug(`Keycloak token ${keycloakToken.idToken} created`);
        return keycloakToken;
    }

    async authenticateUIOCPSSO(email: string, password: string): Promise<TokenObject> {
        logger.info(`authenticate user by UI ${email}`);
        await runaiV1.login.start();
        const token = await runaiV1.login.getOCPToken(email, password);
        await runaiV1.login.stopAndArchive(`login-${email}`);
        return token;
    }

    async authenticateUIAndGetToken(
        email: string,
        password: string,
        ssoUser?: boolean
    ): Promise<TokenObject> {
        logger.info(`authenticate user by UI ${email}`);
        await runaiV1.login.start();
        const token = await runaiV1.login.getToken(email, password, ssoUser);
        await runaiV1.login.stopAndArchive(`login-${email}`);
        return token;
    }

    async authenticatePersonalApp(
        clientId: string,
        clientSecret: string,
        url = `${config.appUrl}`
    ): Promise<TokenResponse> {
        const identityManagerApi = new IdentityManagerApi(url);
        const appTokenRequest: TokenRequest = {
            clientID: clientId,
            clientSecret: clientSecret,
            grantType: TokenRequestGrantTypeEnum.ClientCredentials
        };
        try {
            const appTokenResponse = await identityManagerApi.tokensApi.grantToken(appTokenRequest);
            return appTokenResponse.body;
        } catch (error) {
            throw new Error(`Unable to authenticate personal app, error: ${JSON.stringify(error)}`);
        }
    }

    async authenticateKeycloak(
        userName?: string,
        password?: string,
        url?: string,
        accessToken?: string
    ): Promise<TokenObject> {
        let tokenRequest: TokenRequest;
        let tokenApi = identityManagerApi.tokensApi;
        if (url) {
            tokenApi = new TokensApi(url);
            tokenApi.accessToken = accessToken;
        }

        if (versionUtils.isVersionSupported(flowRootState.controlPlaneVersion, '2.23')) {
            tokenRequest = {
                grantType: TokenRequestGrantTypeEnum.Password,
                username: userName,
                password: password
            };
        } else {
            tokenRequest = {
                grantType: TokenRequestGrantTypeEnum.Password,
                username: userName,
                password: password,
                clientID: 'cli'
            };
        }
        const response = await tokenApi.grantToken(tokenRequest);
        logger.debug(`response: ${JSON.stringify(response)}`);
        return response.body;
    }

    async deleteUser(userId: string, accessToken: string): Promise<void> {
        try {
            identityManagerApi.usersApi.accessToken = accessToken;
            await identityManagerApi.usersApi.deleteUserById(userId);
        } catch (error) {
            logger.error(`unable to delete user ${userId}, error: ${JSON.stringify(error)}`);
            throw new UtilsError(error);
        }
    }

    isTokenExpiredOrInvalid(token: string, expTime: number): boolean {
        return expTime <= 8 || !token;
    }

    async cleanUpUsers(
        accessToken: string,
        shouldDeleteAllOrShouldDeleteOnly: CleanupMethod,
        testIdentifier?: string,
        usersToKeep?: string[]
    ): Promise<void> {
        identityManagerApi.usersApi.accessToken = accessToken;
        let users = (await identityManagerApi.usersApi.getUsers()).body;
        // Exclude users starting with 'swqa-' (protected service accounts)
        const isProtectedUser = (username: string): boolean => username.startsWith('swqa-');

        if (shouldDeleteAllOrShouldDeleteOnly === CleanupMethod.DeleteAllUsersBesidesSpecifiedUsers) {
            users = users.filter(
                (user) =>
                    user.username !== config.controlPlaneAdminUsername &&
                    !usersToKeep.includes(user.username) &&
                    !isProtectedUser(user.username)
            );
        } else {
            users = users.filter(
                (user) => user.username.includes(testIdentifier) && !isProtectedUser(user.username)
            );
        }
        logger.debug(`about to cleanup users: ${JSON.stringify(users.map((user) => user.username))}`);
        if (users.length > 0) {
            await Utils.parallelMap(10, users, async (user) => {
                try {
                    logger.debug(`delete user ${user.username}`);
                    await this.deleteUser(user.id, accessToken);
                } catch (error) {
                    logger.error(`Error deleting user ${user.username}: ${error}`);
                }
            });
        } else {
            logger.info('No test users found to cleanup.');
        }
    }

    isJwtAboutToExpire(token: string, timeBeforeExpiredMin = 10): boolean {
        const decodedToken = jwt_decode(token);
        const currentTime = new Date().getTime() / 1000;
        const tokenExpirationMin = (decodedToken['exp'] - currentTime) / 60;
        return tokenExpirationMin < timeBeforeExpiredMin;
    }

    getTokenExpMin(token: string): number {
        const decodedToken = jwt_decode(token);
        const currentTime = new Date().getTime() / 1000;
        const tokenExpirationMin = (decodedToken['exp'] - currentTime) / 60;
        return tokenExpirationMin;
    }

    validateAllParamsInToken(token: string, params: Record<string, any>): boolean {
        if (!token) {
            logger.error('Token is undefined or empty.');
            return false;
        }
        try {
            const decodedToken = jwt_decode(token);
            for (const [key, expectedValue] of Object.entries(params)) {
                const actualValue = decodedToken[key];

                if (key === 'scope' && typeof actualValue === 'string' && typeof expectedValue === 'string') {
                    const actualScopes = actualValue.split(' ').sort();
                    const expectedScopes = expectedValue.split(' ').sort();

                    if (JSON.stringify(actualScopes) !== JSON.stringify(expectedScopes)) {
                        logger.error(`Mismatch in scope field: ${key}`, { expectedValue, actualValue });
                        return false;
                    }
                } else if (Array.isArray(expectedValue)) {
                    if (
                        !Array.isArray(actualValue) ||
                        !expectedValue.every((val) => actualValue.includes(val))
                    ) {
                        logger.error(`Mismatch in array field: ${key}`, { expectedValue, actualValue });
                        return false;
                    }
                } else if (typeof expectedValue === 'object' && expectedValue !== null) {
                    if (
                        typeof actualValue !== 'object' ||
                        !this.validateAllParamsInToken(actualValue, expectedValue)
                    ) {
                        logger.error(`Mismatch in object field: ${key}`, { expectedValue, actualValue });
                        return false;
                    }
                } else {
                    if (actualValue !== expectedValue) {
                        logger.error(`Mismatch in field: ${key}`, { expectedValue, actualValue });
                        return false;
                    }
                }
            }
            return true;
        } catch (error) {
            logger.error('JWT Decode Error:', JSON.stringify(error.message));
            return false;
        }
    }

    static buildUserCreationRequest(email: string, password?: string, roles?: Role[]): UserCreationRequest {
        return this.buildUserCreationRequestWithEntityType(email, password, roles, undefined);
    }

    static buildUserCreationRequestWithEntityType(
        email: string,
        password?: string,
        roles?: Role[],
        entityType?: AuthEntityType
    ): UserCreationRequest {
        return {
            email: email,
            password: password || config.controlPlaneAdminPassword,
            roles: roles || [Role.Admin],
            permitAllClusters: true,
            permittedClusters: [],
            needToChangePassword: false,
            tenantId: flowRootState.cluster.tenantId,
            entityType: entityType
        };
    }

    static BuildListUserCreationRequest(size: number, userPrefix = 'userBatch'): UserCreationRequest[] {
        const usersList = [];
        for (let i = 0; i < size; i++) {
            const email = objectGenerator.generateEmailByPrefix(
                objectGenerator.generateNameByPrefix(userPrefix)
            );
            usersList.push(AuthUtils.buildUserCreationRequest(email));
        }
        expect(usersList.length).equal(size);
        return usersList;
    }

    async getClusterServiceToken(client_id: string, client_secret: string): Promise<TokenObject> {
        const appTokenRequest: TokenRequest = {
            clientID: client_id,
            clientSecret: client_secret,
            grantType: TokenRequestGrantTypeEnum.ClientCredentials
        };
        try {
            const appTokenResponse = await identityManagerApi.tokensApi.grantToken(appTokenRequest);
            return appTokenResponse.body;
        } catch (error) {
            throw new Error(`Failed to obtain token+ ${error}`);
        }
    }

    async refreshTokensIfNeeded(stateFile: string): Promise<void> {
        try {
            if (AuthUtils.isTokenRefreshing) {
                return;
            }
            AuthUtils.isTokenRefreshing = true;

            const refreshPromises = [];

            if (flowRootState.users && flowRootState.users.system_administrator) {
                const tokenExpLocal = this.getTokenExpMin(
                    flowRootState.users.system_administrator.idToken ||
                        flowRootState.users.system_administrator.accessToken
                );
                const localNeedsRefresh = this.isTokenExpiredOrInvalid(
                    flowRootState.users.system_administrator.idToken ||
                        flowRootState.users.system_administrator.accessToken,
                    tokenExpLocal
                );

                if (localNeedsRefresh) {
                    refreshPromises.push(
                        (async (): Promise<void> => {
                            try {
                                await usersManager.authenticateAllLocalUsers(flowRootState.users);
                                logger.info('Local users tokens refreshed successfully');
                            } catch (error) {
                                logger.error(`Failed to refresh local users: ${error.message}`);
                            }
                        })()
                    );
                }
            }

            if (
                process.env.INIT_SSO_USERS_AND_IDP === 'true' &&
                flowRootState.usersSSO &&
                flowRootState.usersSSO.system_administrator
            ) {
                const tokenExpSSO = this.getTokenExpMin(
                    flowRootState.usersSSO.system_administrator.idToken ||
                        flowRootState.usersSSO.system_administrator.accessToken
                );
                const ssoNeedsRefresh = this.isTokenExpiredOrInvalid(
                    flowRootState.usersSSO.system_administrator.idToken ||
                        flowRootState.usersSSO.system_administrator.accessToken,
                    tokenExpSSO
                );

                if (ssoNeedsRefresh) {
                    refreshPromises.push(
                        (async (): Promise<void> => {
                            try {
                                await usersManager.authenticateAllExistingSSOUsers(flowRootState.usersSSO);
                                logger.info('SSO users tokens refreshed successfully');
                            } catch (error) {
                                logger.error(`Failed to refresh SSO users: ${error.message}`);
                            }
                        })()
                    );
                }
            }

            if (refreshPromises.length > 0) {
                await Promise.all(refreshPromises);

                if (fs.existsSync(stateFile)) {
                    try {
                        const sharedFlowRootState = { flowRootState: flowRootState };
                        fs.writeFileSync(stateFile, JSON.stringify(sharedFlowRootState, null, 2));
                    } catch (error) {
                        logger.error(`Failed to save globalState.json: ${error.message}`);
                    }
                }
            }
        } catch (error) {
            logger.error(`Error during token refresh: ${error}`);
        } finally {
            AuthUtils.isTokenRefreshing = false;
        }
    }

    // Will be removed after we have done testing upgrades from 2.16/2.17
    async authenticateUserOld(email: string, password: string): Promise<TokenObject> {
        return await this.authenticateKeycloakAndGetTokenOld(email, password);
    }

    async authenticateKeycloakAndGetTokenOld(userName: string, password?: string): Promise<TokenObject> {
        logger.debug(`authenticate user by Keycloak ${userName}`);
        const resolvedPassword = await this.resolveAdminPasswordIfNeeded(userName, password);
        const realmName = (await identityManagerApi.configApi.getTenantConfiguration()).body.tenantName;
        const keycloakToken = await this.authenticateKeycloakOld('runai-cli', userName, resolvedPassword, realmName);
        logger.debug(`Keycloak token ${keycloakToken.idToken} created`);
        return keycloakToken;
    }

    async authenticateKeycloakOld(
        client_id: string,
        username: string,
        password: string,
        realmName: string
    ): Promise<TokenObject> {
        logger.debug(`starting authenticate by Keycloak ${client_id}`);
        try {
            const keycloakData = {
                grant_type: 'password',
                client_id: `${client_id}`,
                username: `${username}`,
                password: `${password}`
            };
            const getAppBearer = {
                method: 'POST',
                headers: { 'content-type': 'application/x-www-form-urlencoded' },
                data: qs.stringify(keycloakData),
                url: `${config.appUrl}/auth/realms/${realmName}/protocol/openid-connect/token`
            };
            const tokenResponse = await axios(getAppBearer);
            expect(tokenResponse.status).equal(StatusCodes.OK);
            logger.debug('Token response StatusCodes OK');
            const tokenObject = new TokenObject(
                tokenResponse.data.access_token,
                tokenResponse.data.refresh_token
            );
            logger.debug('Keycloak token successfully created ' + tokenObject.idToken);
            return tokenObject;
        } catch (error) {
            logger.error(
                `unable to authenticate and get token from Keycloak, error: ${JSON.stringify(error)}`
            );
            throw new UtilsError(error);
        }
    }
    // Will be removed after we have done testing upgrades from 2.16/2.17
}
