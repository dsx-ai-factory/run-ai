/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, logger, settings, TokenObject } from '../../../helpers';
import { BaseRunaiPage } from './BaseRunaiPage';
export class Login extends BaseRunaiPage {
    // browser elements

    constructor(videoFilePath: string, videoFileType: string) {
        super(videoFilePath, videoFileType);
    }

    private async getTokenK8s(): Promise<TokenObject> {
        const response = await this.page.waitForResponse((resp) => resp.url().includes('/token'));
        const respJSON = await response.json();
        const tokenObject = new TokenObject(
            respJSON['idToken'],
            respJSON['refreshToken'],
            respJSON['externalToken'],
            respJSON['accessToken']
        );
        return tokenObject;
    }

    async getToken(username: string, password: string, ssoUser?: boolean): Promise<TokenObject> {
        const signInTimeout = setTimeout(async () => {
            await this.stopAndArchive();
            return new Error('signIn failed: timeout');
        }, settings.taskTimeoutMS);
        await this.page.goto(config.tenantUrl);
        await this.redirectAndSignInForOcpToken(username, password, true, ssoUser);
        const tokenObject: TokenObject = await this.getTokenK8s();
        clearTimeout(signInTimeout);
        logger.debug(`login success for user: ${username}, token: ${JSON.stringify(tokenObject)}`);
        return tokenObject;
    }

    async getOCPToken(username: string, password: string): Promise<TokenObject> {
        const signInTimeout = setTimeout(async () => {
            await this.stopAndArchive();
            return new Error('signIn failed: timeout');
        }, settings.taskTimeoutMS);
        await this.page.goto(config.tenantUrl);
        await this.signInOCP(username, password);

        const tokenObject: TokenObject = await this.getTokenK8s();

        logger.error(`Unsupported platform ${settings.platform} for UI login`);

        clearTimeout(signInTimeout);
        logger.debug(`login success for user: ${username}, token: ${JSON.stringify(tokenObject)}`);
        return tokenObject;
    }
}
