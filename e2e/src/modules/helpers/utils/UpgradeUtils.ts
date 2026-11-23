/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    infoUtils,
    logger,
    TokenObject,
    UtilsError,
    versionUtils
} from '..';
import { identityManagerApi } from '../../api';
import qs from 'qs';
import { runaiV1 } from '../../ui';
import { expect } from 'chai';
import {
    TokenRequest,
    TokensApi,
    TokenRequestGrantTypeEnum
} from '../../api/apis/identityManagerApi/generated/api';
import { StatusCodes } from 'http-status-codes';
import axios from 'axios';
export enum CleanupMethod {
    DeleteAllUsersBesidesSpecifiedUsers,
    DeleteOnlyUsersWithSubString
}

export class UpgradeUtils {
    async authenticateUser(email: string, password: string, uiSSOAuth?: boolean): Promise<TokenObject> {
        try {
            if (uiSSOAuth) {
                return await this.authenticateUIAndGetToken(email, password, uiSSOAuth);
            }
            return await this.authenticateKeycloakAndGetToken(email, password);
        } catch (error) {
            logger.error(`unable to authenticate user ${email}, error: ${JSON.stringify(error)}`);
            throw new UtilsError(error);
        }
    }

    async authenticateKeycloakAndGetToken(userName: string, password: string): Promise<TokenObject> {
        logger.debug(`authenticate user by Keycloak ${userName}`);
        const keycloakToken = await this.authenticateKeycloak(userName, password);
        logger.debug(`Keycloak token ${keycloakToken.idToken} created`);
        return keycloakToken;
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

        //auth for versions 2.19+ available with api
        if (versionUtils.isVersionSupported(await infoUtils.getControlPlaneVersion(), '2.19')) {
            tokenRequest = {
                grantType: TokenRequestGrantTypeEnum.Password,
                username: userName,
                password: password,
                clientID: 'cli'
            };
            const response = await tokenApi.grantToken(tokenRequest);
            logger.debug(`response: ${JSON.stringify(response)}`);
            return response.body;
        }
        //auth for old available with keycloak directly only
        else {
            const tokenRequest = {
                grant_type: 'password',
                client_id: 'runai-cli',
                username: `${userName}`,
                password: `${password}`
            };
            const getAppBearer = {
                method: 'POST',
                headers: { 'content-type': 'application/x-www-form-urlencoded' },
                data: qs.stringify(tokenRequest),
                url: `${config.tenantUrl}/auth/realms/runai/protocol/openid-connect/token`
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
        }
    }
}
