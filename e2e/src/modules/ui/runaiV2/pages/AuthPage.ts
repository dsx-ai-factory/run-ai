/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { config, flowRootState, logger, settings, waitUtils } from '../../../helpers';
import { ROLE_NAME_TO_PROPERTY } from '../../../helpers/config/roleMapping';
import { introModalUtils } from '../../utils/introModalUtils';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { UserPostResponse } from '../../../api/apis/identityManagerApi/generated/model/userPostResponse';
import { workloadTourUtils } from '../../utils/guidedTourUtil';
import { adminOnboardingModalUtils } from '../../utils/adminOnboardingModalUtils';
export class AuthPage extends BaseV2Page {
    usernameInput: string;

    passwordInput: string;

    submitButton: string;

    ssoSigninButton: string;

    ocpSsoButton: string;

    ocpLinkAccountButton: string;

    overviewIcon: string;

    jobsIcon: string;

    clusterSaasDropDown: string;

    signInWithSSO: string;

    newpasswordInput: string;

    confirmNewPasswordInput: string;

    resetPasswordButton: string;

    constructor(page: Page) {
        super(page);

        this.usernameInput = 'input[name="email"], input[id="username"], input[id=inputUsername]';
        this.passwordInput = 'input[name="password"], input[id="password"]';
        this.submitButton =
            ':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]';
        this.ssoSigninButton = '[id="social-saml"]';
        this.ocpSsoButton = '.idp >> nth=1';
        this.ocpLinkAccountButton = '[id="linkAccount"]';
        this.overviewIcon = '[aid="menu-icon-overview"]';
        this.jobsIcon = '[aid="menu-icon-jobs"]';
        this.clusterSaasDropDown = '[aid="cluster-dropdown"]';
        this.signInWithSSO = '#sso-button-id';
        this.newpasswordInput = '#password-new';
        this.confirmNewPasswordInput = '#password-confirm';
        this.resetPasswordButton = '#kc-passwd-update-form input[type="submit"]';
    }

    async signIn(username: string, password: string): Promise<void> {
        await this.page.goto(config.tenantUrl);

        await this.signInK8S(username, password);

        logger.debug(`login success for user: ${username}`);
        await waitUtils.waitForElementToExist(this, '[aid="page-title"]');
    }

    async signInOnly(username: string, password: string): Promise<void> {
        await this.page.locator(this.usernameInput).fill(username, { timeout: 70000 });
        await this.page.locator(this.passwordInput).fill(password);
        await this.page.locator(this.submitButton).click();
    }

    async signInValidateLoggedIn(username: string, password: string): Promise<void> {
        await this.page.locator(this.usernameInput).fill(username, { timeout: 70000 });
        await this.page.locator(this.passwordInput).fill(password);
        await this.page.locator(this.submitButton).click();
        await this.page.waitForSelector('[aid="page-title"]', { timeout: 70000 });
    }

    async signInLocalUserEula(username: string, password: string): Promise<void> {
        await this.page.goto(config.tenantUrl);
        await this.signInOnly(username, password);
    }

    async signInAfterResetPassword(username: string, password: string, newPassword: string): Promise<void> {
        await this.page.goto(config.tenantUrl);
        await this.submitLoginForm(username, password, newPassword);
        logger.debug(`login success for user: ${username}`);
    }

    private async signInK8S(username: string, password: string): Promise<void> {
        await this.submitLoginForm(username, password);
        await this.selectCluster(flowRootState.cluster.name);
    }

    private async selectCluster(clusterName: string): Promise<void> {
        await pageFilterUtils.selectCluster(this.page, clusterName);
    }

    private async closeIntroModals(): Promise<void> {
        try {
            await introModalUtils.closeIntroModals(this.page);
        } catch (error) {
            logger.error('error closing intro modals', error);
        }
    }

    private async closeWorkloadTour(): Promise<void> {
        try {
            await workloadTourUtils.closeWorkloadTour(this.page);
            logger.debug('workload tour closed');
        } catch (error) {
            logger.error('error closing workload tour', error);
        }
    }

    private async closeAdminOnboardingModal(): Promise<void> {
        try {
            await adminOnboardingModalUtils.closeAdminOnboardingModal(this.page);
            logger.debug('admin onboarding modal closed');
        } catch (error) {
            logger.error('error closing admin onboarding modal', error);
        }
    }

    private async dismissActiveGuidedTour(): Promise<void> {
        try {
            // Check if page is still open before attempting dismissal
            if (this.page.isClosed()) {
                logger.debug('Page is closed, skipping guided tour dismissal');
                return;
            }

            const bodyHasGuidedTour = (await this.page.locator('body.guided-tour-active').count()) > 0;
            if (bodyHasGuidedTour) {
                logger.debug('Guided tour is active, attempting to dismiss it');
                const exitModal = this.page.locator('[aid="guided-tour-exit-modal"]');
                const exitModalVisible = await exitModal.isVisible({ timeout: 2000 }).catch(() => false);

                if (exitModalVisible) {
                    logger.debug('Guided tour exit modal is already visible, clicking Exit button');
                    const exitConfirmButton = this.page.locator('[aid="guided-tour-exit-confirm-btn"]');
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
                        const exitModalAfterClick = this.page.locator('[aid="guided-tour-exit-modal"]');
                        const modalAppeared = await exitModalAfterClick
                            .isVisible({ timeout: 3000 })
                            .catch(() => false);

                        if (modalAppeared) {
                            logger.debug('Exit modal appeared, clicking Exit button');
                            const exitConfirmButton = this.page.locator(
                                '[aid="guided-tour-exit-confirm-btn"]'
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
            } else {
                logger.debug('No active guided tour detected');
            }
        } catch (error) {
            logger.debug('Could not dismiss guided tour, continuing anyway', error);
        }
    }

    private async closeOnboardingActivity(): Promise<void> {
        try {
            // in SAAS the login page and then App are in different domains, so we need to wait for the application url then run the following steps
            await this.page.waitForURL((url) => url.href.includes(config.tenantUrl), { timeout: 15000 });
            await this.closeIntroModals();
            await this.closeWorkloadTour();
            await this.closeAdminOnboardingModal();
            await this.dismissActiveGuidedTour();
            await this.page.waitForTimeout(500);
            logger.debug('all onboarding activities closed');
        } catch (error) {
            logger.error('error closing onboarding activity', error);
        }
    }

    async submitLoginForm(
        username: string,
        password: string,
        newPassword?: string,
        closeOnboarding: boolean = true
    ): Promise<void> {
        try {
            await this.signInOnly(username, password);
            if (newPassword) {
                const newPasswordInput = this.page.locator(this.newpasswordInput);
                const confirmNewPasswordInput = this.page.locator(this.confirmNewPasswordInput);
                const resetPasswordButton = this.page.locator(this.resetPasswordButton);
                await newPasswordInput.waitFor({ state: 'visible' });
                await newPasswordInput.fill(newPassword);
                await confirmNewPasswordInput.waitFor({ state: 'visible' });
                await confirmNewPasswordInput.fill(newPassword);
                await resetPasswordButton.waitFor({ state: 'visible' });
                await resetPasswordButton.isEnabled();
                await resetPasswordButton.click();
            }

            if (closeOnboarding) {
                await this.closeOnboardingActivity();
            }
        } catch (error) {
            logger.error(`Error in submitLoginForm: ${error.message}`);
            throw error;
        }
    }

    async loginSSO(username: string, password: string): Promise<void> {
        await this.page.goto(config.tenantUrl);
        await this.page.locator(this.signInWithSSO).click();
        await this.signInOnly(username, password);
        await this.closeOnboardingActivity();
    }

    async loginSSOWithAutoRedirectOn(username: string, password: string): Promise<void> {
        await this.page.goto(config.tenantUrl);
        await this.signInOnly(username, password);

        await this.closeOnboardingActivity();
    }

    async loginPageDisplayed(): Promise<void> {
        await this.page.goto(config.tenantUrl);
    }

    async loginAsRole(roleName: string): Promise<void> {
        const user = flowRootState.users[ROLE_NAME_TO_PROPERTY[roleName]];
        await this.signIn(user.email, user.password);
    }

    async signInAndVerifyNoRolesPage(user: UserPostResponse): Promise<void> {
        await this.page.goto(config.tenantUrl);
        try {
            await this.submitLoginForm(user.username, user.tempPassword);
        } catch (titleError) {
            if (
                titleError.name === 'UtilsError' &&
                titleError.message.includes(
                    'Timed out waiting for condition "Waiting for element [aid="page-title"] to exist"'
                )
            ) {
                return;
            } else {
                throw titleError;
            }
        }
        await this.page.waitForSelector('[aid="no-roles"]');
    }

    async loginSSOandVerifyMissingClaim(username: string, password: string): Promise<boolean> {
        await this.page.goto(config.tenantUrl);
        await this.page.locator(this.signInWithSSO).click();
        await this.signInOnly(username, password);
        try {
            await this.page.waitForSelector('[aid="missing-claim"]');
            return true;
        } catch (error) {
            logger.error(`Missing claim not found for user ${error}`);
            return false;
        }
    }
}
