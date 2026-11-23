/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, logger } from '../../../helpers';
import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';

export class LoginUi extends BaseV2Page {
    backToLogin: string;

    loginPageValidator: string;

    loggedInUser: string;

    topBar: string;

    profileIcon: string;

    signOut: string;

    usernameInput: string;

    passwordInput: string;

    helpIcon: string;

    constructor(page: Page) {
        super(page);
        this.backToLogin = '.link-clr.pointer.underline-on-hover';
        this.loginPageValidator = '#username';
        this.loggedInUser = '[aid="profile-btn"]';
        this.topBar = '[role="toolbar"]';
        this.profileIcon = '[aid="profile-btn"]';
        this.signOut = '[aid="sign-out-btn"]';
        this.usernameInput = '#username';
        this.passwordInput = '#password';
        this.helpIcon = '.ph-bold.ph-question';
    }

    async browserCliAuthGetToken(
        app_host: string,
        tenant_name: string,
        tenant_domain: string
    ): Promise<string> {
        const url = `${app_host}/auth/realms/${tenant_name}/protocol/openid-connect/auth?access_type=offline&client_id=runai-cli&redirect_uri=${tenant_domain}/oauth-code&response_type=code&scope=email+openid+offline_access`;
        await this.page.goto(url);
        await this.page.waitForSelector('#username');
        await this.page.fill('#username', config.controlPlaneAdminUsername);
        await this.page.fill('#password', config.controlPlaneAdminPassword);
        await this.page.click('#kc-login');
        const codeLocator = this.page.locator('.copy-command-wrapper.code-block pre');
        await codeLocator.waitFor({ state: 'visible', timeout: 35000 });
        const code = (await codeLocator.innerText()).trim();
        return code;
    }

    async nonPermissionLoginValidator(): Promise<boolean> {
        try {
            await this.page.waitForSelector('text=Ask your administrator for them', { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error(`${error} user does not logged in`);
        }
    }

    async loggedUiValidator(): Promise<boolean> {
        try {
            await this.page.waitForSelector(this.loggedInUser, { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error(`${error} user does not logged in`);
        }
    }

    async topBarValidator(): Promise<boolean> {
        try {
            await this.page.waitForSelector(this.topBar, { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error(`${error} top bar does not exist`);
        }
    }

    async logoutToLoginPage(): Promise<boolean> {
        logger.debug('Logout');
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(1000);
        await this.page.waitForSelector(this.profileIcon, { timeout: 15000 });
        await this.page.locator(this.profileIcon).click();
        await this.page.waitForSelector(this.signOut, { timeout: 15000 });
        await this.page.click(this.signOut);
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.usernameInput).isVisible({ timeout: 15000 });
        await this.page.locator(this.passwordInput).isVisible({ timeout: 15000 });
        return true;
    }

    async tenantIconValidator(): Promise<boolean> {
        try {
            const logo = this.page.locator('.runai-brand-logo, .runai-distribution-logo');
            await logo.first().waitFor({ state: 'visible', timeout: 10000 });
            return (await logo.count()) === 1;
        } catch (error) {
            logger.error(`Tenant logo validation failed: ${error}`);
            return false;
        }
    }

    async openLinks(): Promise<boolean> {
        await this.page.click(this.helpIcon, { timeout: 10000 });
        try {
            await this.page.waitForSelector('text=Documentation Center', { timeout: 10000 });
            await this.page.waitForSelector('text=API Documentation', { timeout: 10000 });
            await this.page.waitForSelector('text=Researcher Command Line Interface', { timeout: 10000 });
            await this.page.waitForSelector('text=Contact Support', { timeout: 10000 });
            await this.page.click('text=Documentation Center', { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error(`${error} help icon function does not work`);
            return false;
        }
    }

    async ssoLoginPageShown(): Promise<boolean> {
        try {
            await this.page.reload();
            await this.page.waitForSelector('[id="kc-content"]', { timeout: 10000 });
            await this.page.waitForSelector('.sso-button-text', { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error(`${error} login page does not exist`);
            return false;
        }
    }

    async enableSkipSSOLoginPageShown(): Promise<boolean> {
        await this.page.reload();
        await this.page.waitForSelector('[id="kc-form-login"]', { timeout: 10000 });
        const ssoButtonCount = await this.page.locator('text=CONTINUE WITH SSO').count();
        return ssoButtonCount === 0;
    }

    async normalLoginPageShown(): Promise<boolean> {
        await this.page.reload();
        await this.page.waitForSelector('[id="kc-form-login"]', { timeout: 10000 });
        const ssoButtonCount = await this.page.locator('text=CONTINUE WITH SSO').count();
        return ssoButtonCount === 0;
    }
}
