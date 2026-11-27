/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { logger } from '../../../helpers';
import { existsSync, mkdirSync } from 'fs';
import { join } from 'path';

export class AdminOnboardingPage {
    page: Page;

    mainPageTitle = '[aid="page-title"]';

    delegateCheckbox: string;

    delegateEmailInput: string;

    skipButton: string;

    continueButton: string;

    completeSetupButton: string;

    samlProtocolLabel: string;

    openIdConnectProtocolLabel: string;

    fromComputerRadio: string;

    fromUrlRadio: string;

    metadataFileInput: string;

    discoveryUrlInput: string;

    clientIdInput: string;

    clientSecretInput: string;

    teamNameInput: string;

    teamMemberEmailInput: string;

    memberEmailsRadio: string;

    groupNameRadio: string;

    groupNameInput: string;

    ssoUrlInput: string;

    downloadXmlButton: string;

    redirectUrlField: string;

    entityIdField: string;

    downloadCredentialsButton: string;

    teamSetupSuccessTitle: string;

    skipSsoSetupButton: string;

    creatingTeamAgainSpan: string;

    contentSubtitle: string;

    closeOnboardingButton: string;

    smtpHostInput: string;

    smtpUsernameInput: string;

    smtpPasswordInput: string;

    smtpFromEmailInput: string;

    smtpFromNameInput: string;

    verifyButton: string;

    constructor(page: Page) {
        this.page = page;

        this.delegateCheckbox = '[aid="delegate-checkbox"]';
        this.delegateEmailInput = '[aid="delegate-email"]';
        this.skipButton = '.skip-button';
        this.continueButton = '.continue-button';
        this.completeSetupButton = '[aid="complete-setup-btn"]';
        this.samlProtocolLabel = '.option-title:text-matches(".*saml.*", "i")';
        this.openIdConnectProtocolLabel = '.option-title:text-matches(".*openid.*", "i")';
        this.fromComputerRadio = '[aria-label="From computer"]';
        this.fromUrlRadio = '[aria-label="From URL"]';
        this.metadataFileInput = 'input[type="file"][accept=".xml"]';
        this.discoveryUrlInput = '[aria-label="Discovery URL"]';
        this.clientIdInput = '[aria-label="Client ID"]';
        this.clientSecretInput = '[aria-label="Client secret"]';
        this.teamNameInput = '[aid="team-name"]';
        this.teamMemberEmailInput = '[aria-label="Team member(s) email"]';
        this.memberEmailsRadio = '[aid="radio-option-Member emails"]';
        this.groupNameRadio = '[aid="radio-option-Group name"]';
        this.groupNameInput = '[aid="group-name"]';
        this.ssoUrlInput = 'input[aria-label="Metadata URL"]';
        this.downloadXmlButton = '[aid="download-metadata"]';
        this.downloadCredentialsButton = '.download-button';
        this.teamSetupSuccessTitle = '.content-title:text("You\'re all set!")';
        this.skipSsoSetupButton = 'button:has-text("Skip SSO Setup")';
        this.creatingTeamAgainSpan = 'text="creating the team again"';
        this.contentSubtitle = '.content-subtitle';
        this.closeOnboardingButton = 'button[aria-label="Close onboarding"]';
        this.smtpHostInput = '[aid="smpt-host-input"]';
        this.smtpUsernameInput = '[aid="smpt-username-input"]';
        this.smtpPasswordInput = '[aid="smpt-password-input"]';
        this.smtpFromEmailInput = '[aid="smpt-from-email-input"]';
        this.smtpFromNameInput = '[aid="smpt-from-name-input"]';
        this.verifyButton = '[aid="verify-button"]';
    }

    async clickContinueButton(): Promise<void> {
        const button = this.page.locator(this.continueButton);

        // Wait for button to be visible with increased timeout for Cloud environments
        await button.waitFor({ state: 'visible', timeout: 60000 });

        // Wait for button to be enabled (not disabled)
        await button.waitFor({ state: 'attached', timeout: 10000 });
        await this.page.waitForTimeout(500); // Small delay for UI stability

        // Ensure button is enabled by checking for disabled attribute
        await this.page.waitForFunction(
            (selector) => {
                const btn = document.querySelector(selector);
                return btn && !btn.hasAttribute('disabled') && !btn.classList.contains('disabled');
            },
            this.continueButton,
            { timeout: 30000 }
        );

        // Wait for any loading indicators to disappear
        const loadingIndicators = this.page.locator('.loading, .spinner, [data-loading="true"]');
        const loadingCount = await loadingIndicators.count();
        if (loadingCount > 0) {
            await loadingIndicators
                .first()
                .waitFor({ state: 'hidden', timeout: 20000 })
                .catch(() => {
                    logger.debug('No loading indicators found or already hidden');
                });
        }

        return button.click();
    }

    async clickCompleteSetupButton(): Promise<void> {
        await this.page.locator(this.completeSetupButton).click();
    }

    async selectDelegationOption(): Promise<void> {
        await this.page.locator(this.delegateCheckbox).click();
    }

    async selectSamlProtocol(): Promise<void> {
        await this.page.locator(this.samlProtocolLabel).click();
    }

    async selectOpenIdConnectProtocol(): Promise<void> {
        await this.page.locator(this.openIdConnectProtocolLabel).click();
    }

    async clickSkipButton(): Promise<void> {
        await this.page.locator(this.skipButton).click();
        // Wait for skip action to complete and page to stabilize
        await this.page.waitForTimeout(1000);

        // Wait for any loading indicators
        const loadingIndicators = this.page.locator('.loading, .spinner, [data-loading="true"]');
        const loadingCount = await loadingIndicators.count();
        if (loadingCount > 0) {
            await loadingIndicators
                .first()
                .waitFor({ state: 'hidden', timeout: 10000 })
                .catch(() => {
                    logger.debug('No loading indicators after skip');
                });
        }
    }

    async enterDelegateEmail(email: string): Promise<void> {
        await this.page.locator(this.delegateEmailInput).fill(email);
    }

    async selectSsoFromComputer(): Promise<void> {
        await this.page.locator(this.fromComputerRadio).click();
    }

    async selectSsoFromUrl(): Promise<void> {
        await this.page.locator(this.fromUrlRadio).click();
    }

    async uploadSsoMetadataFile(filePath: string): Promise<void> {
        await this.page.locator(this.metadataFileInput).setInputFiles(filePath);
    }

    async fillSamlSsoURL(url: string): Promise<void> {
        await this.page.locator(this.ssoUrlInput).fill(url);
    }

    async configSamlSsoURL(url: string): Promise<void> {
        await this.selectSamlProtocol();
        await this.clickContinueButton();
        await this.selectSsoFromUrl();
        await this.fillSamlSsoURL(url);
    }

    async enterOpenIdDiscoveryUrl(url: string): Promise<void> {
        await this.page.locator(this.discoveryUrlInput).fill(url);
    }

    async enterOpenIdClientId(clientId: string): Promise<void> {
        await this.page.locator(this.clientIdInput).fill(clientId);
    }

    async enterOpenIdClientSecret(clientSecret: string): Promise<void> {
        await this.page.locator(this.clientSecretInput).fill(clientSecret);
    }

    async configOIDCSso(url: string, clientId: string, clientSecret: string): Promise<void> {
        await this.selectOpenIdConnectProtocol();
        await this.clickContinueButton();
        await this.enterOpenIdDiscoveryUrl(url);
        await this.enterOpenIdClientId(clientId);
        await this.enterOpenIdClientSecret(clientSecret);

        // Wait for OIDC validation to complete
        await this.page.waitForTimeout(2000);

        // Wait for any validation spinners or loading indicators
        const loadingIndicators = this.page.locator('.loading, .spinner, [data-loading="true"], .validating');
        const loadingCount = await loadingIndicators.count();
        if (loadingCount > 0) {
            await loadingIndicators
                .first()
                .waitFor({ state: 'hidden', timeout: 30000 })
                .catch(() => {
                    logger.debug('No loading indicators after OIDC config');
                });
        }
    }

    async enterTeamName(teamName: string): Promise<void> {
        const input = this.page.locator(this.teamNameInput);
        await input.waitFor({ state: 'visible', timeout: 30000 });
        return input.fill(teamName);
    }

    async enterTeamMemberEmail(email: string): Promise<void> {
        await this.page.locator(this.teamMemberEmailInput).fill(email);
        await this.page.keyboard.press('Tab');

        // Wait for email validation to complete
        await this.page.waitForTimeout(1000);

        // Wait for any validation indicators
        const validationIndicators = this.page.locator('.validating, .loading, .spinner');
        const validationCount = await validationIndicators.count();
        if (validationCount > 0) {
            await validationIndicators
                .first()
                .waitFor({ state: 'hidden', timeout: 10000 })
                .catch(() => {
                    logger.debug('No validation indicators after email entry');
                });
        }
    }

    async selectMemberEmailsOption(): Promise<void> {
        await this.page.locator(this.memberEmailsRadio).click();
    }

    async selectGroupNameOption(): Promise<void> {
        await this.page.locator(this.groupNameRadio).click();
    }

    async enterGroupName(groupName: string): Promise<void> {
        await this.page.locator(this.groupNameInput).fill(groupName);

        // Wait for group name validation to complete
        await this.page.waitForTimeout(1000);

        // Wait for any validation indicators
        const validationIndicators = this.page.locator('.validating, .loading, .spinner');
        const validationCount = await validationIndicators.count();
        if (validationCount > 0) {
            await validationIndicators
                .first()
                .waitFor({ state: 'hidden', timeout: 10000 })
                .catch(() => {
                    logger.debug('No validation indicators after group name entry');
                });
        }
    }

    async clickDownloadXmlButton(): Promise<void> {
        await this.page.locator(this.downloadXmlButton).click();
    }

    async getRedirectUrl(): Promise<string> {
        return await this.page.locator('text=/https:.*endpoint/').textContent();
    }

    async getEntityId(): Promise<string> {
        return await this.page.locator('text=/.*/runai$/').textContent();
    }

    async downloadCredentials(): Promise<string> {
        const downloadPromise = this.page.waitForEvent('download');
        await this.page.locator(this.downloadCredentialsButton).click();
        const download = await downloadPromise;

        const fileName = download.suggestedFilename();
        const downloadsPath = '/tmp/downloads';

        if (!existsSync(downloadsPath)) {
            mkdirSync(downloadsPath, { recursive: true });
        }

        const filePath = join(downloadsPath, fileName);
        await download.saveAs(filePath);

        logger.info(`Team credentials downloaded and saved to: ${filePath}`);
        return filePath;
    }

    async isSetupSuccessful(): Promise<boolean> {
        try {
            const successElement = this.page.locator(this.teamSetupSuccessTitle);
            await successElement.waitFor({ state: 'visible', timeout: 5000 });
            return true;
        } catch {
            logger.debug('Team setup success title not found');
            return false;
        }
    }

    async clickSkipSsoSetupButton(): Promise<void> {
        await this.page.locator(this.skipSsoSetupButton).click();
    }

    async clickCreatingTeamAgainSpan(): Promise<void> {
        await this.page.locator(this.creatingTeamAgainSpan).click();
    }

    async getContentSubtitleText(): Promise<string> {
        const locator = this.page.locator(this.contentSubtitle);
        await locator.waitFor({ state: 'visible', timeout: 10000 });
        const text = await locator.textContent();
        return text || '';
    }

    async clickCloseOnboardingButton(): Promise<void> {
        await this.page.locator(this.closeOnboardingButton).click();
    }

    async isStepClickable(stepNumber: number): Promise<boolean> {
        try {
            const stepLocator = this.page.locator(`.step-item:has(.step-number:text("${stepNumber}"))`);

            const hasClickableClass = await stepLocator.evaluate((el) => {
                return el.classList.contains('clickable');
            });

            return hasClickableClass;
        } catch (error) {
            logger.debug(`Error checking if step ${stepNumber} is clickable:`, error.message);
            return false;
        }
    }

    async isOnboardingPageExists(): Promise<boolean> {
        try {
            const onboardingContainer = this.page.locator('.onboarding-container');
            const containerExists = (await onboardingContainer.count()) > 0;

            logger.info(`Onboarding page exists: ${containerExists}`);
            return containerExists;
        } catch (error) {
            logger.debug('Error checking if onboarding page exists:', error.message);
            return false;
        }
    }

    async enterSmtpHost(host: string): Promise<void> {
        await this.page.locator(this.smtpHostInput).fill(host);
    }

    async selectAuthType(authType: string): Promise<void> {
        const authTypeSelector = `[aid="auth-type-radio-${authType.toLowerCase()}"]`;
        await this.page.locator(authTypeSelector).click();
    }

    async enterSmtpUsername(username: string): Promise<void> {
        await this.page.locator(this.smtpUsernameInput).fill(username);
    }

    async enterSmtpPassword(password: string): Promise<void> {
        await this.page.locator(this.smtpPasswordInput).fill(password);
    }

    async enterFromEmailAddress(email: string): Promise<void> {
        await this.page.locator(this.smtpFromEmailInput).fill(email);
    }

    async enterFromDisplayName(name: string): Promise<void> {
        await this.page.locator(this.smtpFromNameInput).fill(name);
    }

    async clickVerifyButton(): Promise<void> {
        await this.page.locator(this.verifyButton).click();
    }

    async isEmailServerVerified(): Promise<boolean> {
        const buttonClass = await this.page.locator(this.verifyButton).getAttribute('class');
        return buttonClass?.includes('verified') || false;
    }

    async configureEmailServer(
        smtpHost: string,
        authType: string,
        fromEmail: string,
        fromName: string,
        smtpUsername?: string,
        smtpPassword?: string
    ): Promise<void> {
        await this.enterSmtpHost(smtpHost);
        await this.selectAuthType(authType);

        if (authType.toLowerCase() !== 'none') {
            if (!smtpUsername || !smtpPassword) {
                throw new Error(`SMTP credentials are required when authType="${authType}"`);
            }
            await this.enterSmtpUsername(smtpUsername);
            await this.enterSmtpPassword(smtpPassword);
        }
        await this.enterFromEmailAddress(fromEmail);
        await this.enterFromDisplayName(fromName);
        await this.clickVerifyButton();
    }
}
