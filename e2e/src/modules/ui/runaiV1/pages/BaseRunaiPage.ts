/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../BrowserEngine';
import {
    authUtils,
    config,
    flowRootState,
    infoUtils,
    logger,
    settings,
    TokenObject,
    versionUtils
} from '../../../helpers';
import { ElementHandle } from 'playwright';
import { introModalUtils } from '../../utils/introModalUtils';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { workloadTourUtils } from '../../utils/guidedTourUtil';
import { adminOnboardingModalUtils } from '../../utils/adminOnboardingModalUtils';

export class BaseRunaiPage extends BrowserEngine {
    // browser elements
    usernameInput: string;

    passwordInput: string;

    submitButton: string;

    newPasswordInput: string;

    newPasswordConfirmInput: string;

    ssoSigninButton: string;

    ocpSsoButton: string;

    ocpLinkAccountButton: string;

    profileIcon: string;

    signOut: string;

    logo: string;

    overwiewIcon: string;

    jobsIcon: string;

    clusterSaasDropDown: string;

    signInWithSSO: string;

    SSONewLoginValidator: string;

    constructor(videoFilePath: string, videoFileType: string) {
        super(videoFilePath, videoFileType);
        this.usernameInput = 'input[name="email"], input[id="username"], input[id=inputUsername]';
        this.passwordInput = 'input[name="password"], input[id="password"]';
        this.submitButton =
            ':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]';
        this.newPasswordInput = 'input[name="password-new"], input[id="password-new"]';
        this.newPasswordConfirmInput = 'input[name="password-confirm"], input[id="password-confirm"]';
        this.ssoSigninButton = '[id="social-saml"]';
        this.ocpSsoButton = 'a[href*="idp=test-user-oauth"]';
        this.ocpLinkAccountButton = '[id="linkAccount"]';
        this.profileIcon = '[aid="profile-btn"]';
        this.signOut = '[aid=sign-out-btn]';
        this.logo = '[aid="logo"]';
        this.overwiewIcon = '[aid="menu-icon-overview"]';
        this.jobsIcon = '[aid="menu-item-jobs"]';
        this.clusterSaasDropDown = '[aid="cluster-dropdown"]';
        this.signInWithSSO = '#sso-button-id';
        this.SSONewLoginValidator = 'div[class="pf-c-login__main-body"] ul';
    }

    private async submitLoginForm(username: string, password: string): Promise<void> {
        await this.page.locator(this.usernameInput).fill(username, { timeout: 70000 });
        await this.page.locator(this.passwordInput).fill(password);
        await this.page.locator(this.submitButton).click();
    }

    async signInOCP(username: string, password: string): Promise<void> {
        let elementExist: ElementHandle<SVGElement | HTMLElement>;
        try {
            elementExist = await this.page.waitForSelector(this.SSONewLoginValidator, { timeout: 5000 });
        } catch (error) {
            logger.info(`CONTINUE WITH SSO button exist or not found in OCP UI: ${JSON.stringify(error)}`);
        }
        if (elementExist) {
            await this.page.locator(this.ocpSsoButton).click();
            await this.submitLoginForm(username, password);
        } else {
            await this.page.locator(this.signInWithSSO).click();
            await this.page.locator(this.ocpSsoButton).click();
            await this.submitLoginForm(username, password);
        }
    }

    private async signInK8S(username: string, password: string): Promise<void> {
        await this.submitLoginForm(username, password);
    }

    async clearLocalStorage(): Promise<void> {
        try {
            await this.page.context().clearCookies();
            await this.page.waitForLoadState('networkidle');
            try {
                await this.page.evaluate(() => localStorage.clear());
            } catch (error) {
                logger.debug('Error clearing local storage:', error);
            }
            for (const frame of this.page.frames()) {
                await frame.evaluate(() => localStorage.clear());
            }
            logger.debug('Local storage cleared successfully.');
        } catch (error) {
            logger.debug('Error clearing local storage:', error);
        }
    }

    async redirectAndSignIn(
        username: string,
        password: string,
        skipSelectCluster?: boolean,
        ssoUser?: boolean
    ): Promise<void> {
        await this.page.goto(config.tenantUrl);
        if (ssoUser) {
            await this.signInWithSSOAuth(username, password);
        } else {
            await this.signIn(username, password);
        }

        const { accessToken } = await authUtils.authenticateUser(username, password, ssoUser);
        const controlPlaneVersion = await infoUtils.getControlPlaneVersion(accessToken);

        const shouldCloseIntroModals = versionUtils.isVersionSupported(controlPlaneVersion, '2.18');
        if (shouldCloseIntroModals) {
            await this.closeIntroModals();
        }

        const shouldCloseWorkloadTour = versionUtils.isVersionSupported(controlPlaneVersion, '2.23');
        if (shouldCloseWorkloadTour) {
            await this.closeWorkloadTour();
        }

        // Close admin onboarding modal for all versions (no version check needed)
        await this.closeAdminOnboardingModal();

        // Dismiss any active guided tour
        await this.dismissActiveGuidedTour();

        if (!skipSelectCluster) {
            await this.selectCluster(flowRootState.cluster.name);
        }
    }

    async redirectAndSignInForOcpToken(
        username: string,
        password: string,
        skipSelectCluster?: boolean,
        ssoUser?: boolean
    ): Promise<void> {
        // await this.page.goto(config.tenantUrl);
        if (ssoUser) {
            await this.signInWithSSOAuth(username, password);
        } else {
            await this.signIn(username, password);
        }

        // Dismiss any active guided tour
        await this.dismissActiveGuidedTour();

        if (!skipSelectCluster) {
            await this.selectCluster(flowRootState.cluster.name);
        }
    }

    async signIn(username: string, password: string): Promise<TokenObject> {
        let tokenObject: TokenObject;
        const signInTimeout = setTimeout(async () => {
            await this.stopAndArchive();
            return new Error('signIn failed: timeout');
        }, settings.taskTimeoutMS);

        await this.signInK8S(username, password);

        clearTimeout(signInTimeout);
        logger.debug(`login success for user: ${username}`);
        return tokenObject;
    }

    async signInWithSSOAuth(username: string, password: string): Promise<void> {
        await this.page.locator(this.signInWithSSO).click();
        await this.submitLoginForm(username, password);
    }

    private async dismissActiveGuidedTour(): Promise<void> {
        try {
            const bodyHasGuidedTour = (await this.page.locator('body.guided-tour-active').count()) > 0;
            if (bodyHasGuidedTour) {
                logger.debug('Guided tour is active, attempting to dismiss it');
                const exitModal = this.page.locator('[aid="guided-tour-leave-confirm-modal"]');
                const exitModalVisible = await exitModal.isVisible({ timeout: 2000 }).catch(() => false);

                if (exitModalVisible) {
                    logger.debug('Guided tour exit modal is already visible, clicking Exit button');
                    const exitConfirmButton = this.page.locator('[aid="guided-tour-leave-confirm-btn"]');
                    await exitConfirmButton.click();
                } else {
                    const closeButton = this.page.locator('.guided-tour-close-btn');
                    const isVisible = await closeButton.isVisible({ timeout: 2000 }).catch(() => false);

                    if (isVisible) {
                        logger.debug(
                            'Guided tour popover close button found, clicking it to open exit modal'
                        );
                        await closeButton.click();
                        await this.page.waitForTimeout(500);
                        const exitModalAfterClick = this.page.locator(
                            '[aid="guided-tour-leave-confirm-modal"]'
                        );
                        const modalAppeared = await exitModalAfterClick
                            .isVisible({ timeout: 3000 })
                            .catch(() => false);

                        if (modalAppeared) {
                            logger.debug('Exit modal appeared, clicking Exit button');
                            const exitConfirmButton = this.page.locator(
                                '[aid="guided-tour-leave-confirm-btn"]'
                            );
                            await exitConfirmButton.click();
                        } else {
                            logger.debug(
                                'No exit modal appeared, guided tour may have been dismissed directly'
                            );
                        }
                    } else {
                        logger.debug('Guided tour close button not visible, trying to press Escape');
                        await this.page.keyboard.press('Escape');
                        await this.page.waitForTimeout(500);
                    }
                }

                await this.page
                    .waitForFunction(() => !document.body.classList.contains('guided-tour-active'), {
                        timeout: 5000
                    })
                    .catch(() => {
                        logger.debug('Guided tour dismissal timeout, continuing anyway');
                    });
                logger.debug('Guided tour dismissed successfully');
            }
        } catch (error) {
            logger.debug('Could not dismiss guided tour, continuing anyway', error);
        }
    }

    async logout(): Promise<string> {
        logger.debug('Logout');
        await this.page.waitForLoadState('load');
        await this.page.waitForSelector(this.profileIcon, { timeout: 15000 });
        await this.page.locator(this.profileIcon).click();
        await this.page.waitForSelector(this.signOut, { timeout: 15000 });
        await this.page.click(this.signOut);
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(1000);
        const url = this.page.url();
        return url;
    }

    async logoutToLoginPage(): Promise<boolean> {
        logger.debug('Logout');
        await this.page.waitForLoadState('load');
        await this.page.waitForSelector(this.profileIcon, { timeout: 15000 });
        await this.page.locator(this.profileIcon).click();
        await this.page.waitForSelector(this.signOut, { timeout: 15000 });
        await this.page.click(this.signOut);
        await this.page.waitForLoadState('load');
        await this.page.locator(this.usernameInput).isVisible({ timeout: 15000 });
        await this.page.locator(this.passwordInput).isVisible({ timeout: 15000 });
        return true;
    }

    async selectCluster(clusterName: string): Promise<void> {
        await pageFilterUtils.selectCluster(this.page, clusterName);
    }

    async closeIntroModals(): Promise<void> {
        await introModalUtils.closeIntroModals(this.page);
    }

    async closeWorkloadTour(): Promise<void> {
        await workloadTourUtils.closeWorkloadTour(this.page);
    }

    async closeAdminOnboardingModal(): Promise<void> {
        await adminOnboardingModalUtils.closeAdminOnboardingModal(this.page);
    }
}
