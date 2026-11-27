/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../BrowserEngine';
import { config, logger, settings } from '../../../helpers';
export class Oauth extends BrowserEngine {
    private firstRequestPath: string;

    // browser elements
    private usernameInput: string;

    private passwordInput: string;

    private submitButton: string;

    private newPasswordInput: string;

    private ocpSsoButton: string;

    private ocpLinkAccountButton: string;

    private tokenSelector: string;

    private signInWithSSO: string;

    private SSONewLoginValidator: string;

    constructor(videoFilePath: string, videoFileType: string) {
        super(videoFilePath, videoFileType);
        this.usernameInput = 'input[name="email"], input[id="username"], input[id=inputUsername]';
        this.passwordInput = 'input[name="password"], input[id="password"]';
        this.submitButton =
            ':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]';
        this.tokenSelector = 'code';
        this.ocpSsoButton = '.idp >> nth=1';
        this.ocpLinkAccountButton = '[id="linkAccount"]';
        this.signInWithSSO = '#sso-button-id';
        this.SSONewLoginValidator = 'div[class="pf-c-login__main-body"] ul';
    }

    private async submitLoginForm(username: string, password: string): Promise<void> {
        await this.page.locator(this.usernameInput).fill(username, { timeout: 70000 });
        await this.page.locator(this.passwordInput).fill(password);
        await this.page.locator(this.submitButton).click();
    }

    private async signInOCP(username: string, password: string): Promise<void> {
        const elementExist = await this.page.waitForSelector(this.SSONewLoginValidator, { timeout: 8000 });
        if (elementExist) {
            await this.page.locator(this.ocpSsoButton).click();
            await this.submitLoginForm(username, password);
        } else {
            await this.page.locator(this.signInWithSSO).click();
            await this.page.locator(this.ocpSsoButton).click();
            await this.submitLoginForm(username, password);
        }
    }

    async getKubeconfigToken(username: string, password: string): Promise<string> {
        await this.start();
        const signInTimeout = setTimeout(async () => {
            await this.stopAndArchive();
            return new Error('get kubeconfig token failed: timeout');
        }, settings.taskTimeoutMS);
        const tokenUrl = `${config.appUrl.replace('runai.apps', 'oauth-openshift.apps')}/oauth/token/request`;
        await this.page.goto(tokenUrl);
        await this.signInOCP(username, password);
        await this.page.locator(this.submitButton).click();
        const token = await this.page.locator(this.tokenSelector).textContent();
        logger.debug(`kubeconfig token: ${token}`);
        clearTimeout(signInTimeout);
        logger.debug(
            `Get token from ocp with username: ${username} succeeded, response: ${JSON.stringify(token)}`
        );
        await this.stopAndArchive(`ocp-token-${username}`);
        return token;
    }
}
