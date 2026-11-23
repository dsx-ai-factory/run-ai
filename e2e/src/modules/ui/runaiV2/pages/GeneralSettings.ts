/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiUtils, config, flowRootState, logger, settings, tenantUtils, waitUtils } from '../../../helpers';
import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../../../helpers/models/featureFlags';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { NotificationChannelForPatchConfig } from '../../../../modules/api/apis/notificationsProxyApi';
import * as path from 'path';
import { fileURLToPath } from 'url';

export class GeneralSettings extends BaseV2Page {
    toolSettingsIcon: string;

    generalIcon: string;

    securityOption: string;

    ssoSettingsTooltip: string;

    logoutRedirectUrlTooltip: string;

    sessionTimeoutTooltip: string;

    addIdentityProvider: string;

    addEmailServer: string;

    fromURL: string;

    urlSamlTextLabel: string;

    urlOidcTextLabel: string;

    mappersLabel: string;

    saveButton: string;

    removeButtonAlert: string;

    deleteButton: string;

    deleteEmailButton: string;

    clinetIDLabel: string;

    clinetSecretLabel: string;

    validatorSAMLCreated: string;

    validatorOIDCCreated: string;

    RedirectURLButton: string;

    redirectURLLable: string;

    sessionTimeoutLabel: string;

    analyticsSection: string;

    resourcesSection: string;

    workloadsSection: string;

    notificationsSection: string;

    clusterAuthenticationSection: string;

    mapperTextLabel: string;

    editIpdButton: string;

    scopeCombobox: string;

    arrowScopes: string;

    autoRedirectSSOToggle: string;

    downloadIdpMetadataXMLFileButton: string;

    editOpenIdConnectButton: string;

    limtQuotaOverSubscriptionToggle: string;

    cpuQuotaToggle: string;

    overQuotaWeightToggle: string;

    gpuResourceOptimizationToggle: string;

    dataVolumeToggle: string;

    flexWorkloadTemplateToggle: string;

    flexWorkloadSubmissionToggle: string;

    newPolicyToggle: string;

    modelCatalogToggle: string;

    nimToggle: string;

    allowGpt2ModelToggle: string;

    yamlSubmissionToggle: string;

    ngcIntegrationToggle: string;

    allowGpt2ModelExperimentalBadge: string;

    ngcIntegrationBetaBadge: string;

    workloadsExpansionHeader: string;

    linkToSamlDescriptor: string;

    branding: string;

    addLogoButton: string;

    brandingEditLogoButton: string;

    logoFileInputContainer: string;

    logoFileInput: string;

    logoUploaded: string;

    oidcIssuerUrlElement: string;

    metadataXmlFileInput: string;

    securityInvalidIcon: string;

    securityWarningIcon: string;

    certificateWarningBannerTitle: string;

    certificateFieldErrorMessage: string;

    cancelIdpEditButton: string;

    constructor(page: Page) {
        super(page);
        this.toolSettingsIcon = '[aid="setup-dropdown"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Admin}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.GeneralSettings}"]`;

        this.generalIcon = '//div[@class="q-item__label" and normalize-space(text())="General"]';
        this.securityOption = '[aid="security-expansion-label"]';
        this.ssoSettingsTooltip = '[aid="tooltip-sso-settings"]';
        this.logoutRedirectUrlTooltip = '[aid="tooltip-logout-redirect-url"]';
        this.sessionTimeoutTooltip = '[aid="tooltip-session-timeout"]';
        this.addIdentityProvider = '[aid="add-idp-btn"]';
        this.addEmailServer = '[aid="add-email-botifications-btn"]';
        this.fromURL = '//div[@aria-label="From URL"]';
        this.urlSamlTextLabel = 'input[aria-label="Metadata URL"]';
        this.mappersLabel = 'input.q-field__native.q-placeholder';
        this.saveButton = '//span[contains(text(),"Save")]';
        this.removeButtonAlert = '[aid="delete-modal-delete-button"]';
        this.deleteButton =
            'button[aid^="delete-"]:not([aid="delete-modal-delete-button"]):not([aid="delete-Email server"]):visible';
        this.deleteEmailButton = '[aid="delete-Email server"]';
        this.urlOidcTextLabel = '[aria-label="Discovery URL"]';
        this.clinetIDLabel = 'input[aria-label="Client ID"]';
        this.clinetSecretLabel = 'input[aria-label="Client secret"]';
        this.validatorSAMLCreated = 'text=Identity provider - SAML 2.0';
        this.validatorOIDCCreated = 'text=Identity provider - OpenID Connect';
        this.RedirectURLButton = '[aid="opener-buttonredirect-url"]';
        this.redirectURLLable = 'input[aria-label="Redirect URL"]';
        this.sessionTimeoutLabel = '[aid="security-session-timeout-label"]';
        this.analyticsSection = '[aid="analytics-expansion-label"]';
        this.resourcesSection = '[aid="resources-expansion-label"]';
        this.notificationsSection = '[aid="notifications-expansion-label"]';
        this.workloadsSection = '[aid="workloads-expansion-label"]';
        this.clusterAuthenticationSection = '[aid="cluster authentication-expansion-label"]';
        this.mapperTextLabel = '[aid="mapper-value"]';
        this.editIpdButton = '[aid^="edit-Identity provider"]';
        this.scopeCombobox = '[role="combobox"]';
        this.arrowScopes = '.q-field:has([role="combobox"]) .q-select__dropdown-icon';
        this.autoRedirectSSOToggle =
            '.toggle-setting:has-text("Auto-Redirect to SSO") input[type="checkbox"]';
        this.downloadIdpMetadataXMLFileButton = '[aid="download-metadata"]';
        this.editOpenIdConnectButton = '[aid="edit-Identity provider - OpenID Connect"]';
        this.limtQuotaOverSubscriptionToggle = '[aid="toggle-setting-tenant.limit_quota_over_subscription"]';
        this.cpuQuotaToggle = '[aid="toggle-setting-project.quota.cpu_resources"]';
        this.overQuotaWeightToggle = '[aid="toggle-setting-project.quota.manual_over_quota"]';
        this.gpuResourceOptimizationToggle = '[aid="toggle-setting-tenant.gpu_over_provisioning"]';
        this.dataVolumeToggle = '[aid="toggle-setting-tenant.enable_data_volumes"]';
        this.flexWorkloadTemplateToggle = '[aid="toggle-setting-tenant.enable_templates_v2"]';
        this.flexWorkloadSubmissionToggle = '[aid="toggle-setting-tenant.enable_hybrid_submission"]';
        this.newPolicyToggle = '[aid="toggle-setting-tenant.enable_policies"]';
        this.modelCatalogToggle = '[aid="toggle-setting-tenant.enable_model_catalog"]';
        this.nimToggle = '[aid="toggle-setting-tenant.enable_nim"]';
        this.allowGpt2ModelToggle = '[aid="toggle-setting-tenant.allow_gpt2_model"]';
        this.yamlSubmissionToggle = '[aid="toggle-setting-yaml-submission"]';
        this.ngcIntegrationToggle = '[aid="toggle-setting-tenant.enable_ngc_integration"]';
        this.allowGpt2ModelExperimentalBadge =
            '[aid="tenant.allow_gpt2_model-setting"] [aid="badge-Experimental"]';
        this.ngcIntegrationBetaBadge = '[aid="tenant.enable_ngc_integration-setting"] [aid="badge-Beta"]';
        this.workloadsExpansionHeader = 'div[aid="workloads-expansion-toggle"] div.runai-expansion-header';
        this.linkToSamlDescriptor = '[aid="open-metadata-xml"]';
        this.branding = 'text=Branding';
        this.addLogoButton = '[aid="add-logo-btn"]';
        this.brandingEditLogoButton = '[aid="branding-edit-logo-btn"]';
        this.logoFileInputContainer = '.q-field__native.row.items-center.cursor-pointer[tabindex="0"]';
        this.logoFileInput = 'input[type="file"][accept="image/png,image/svg+xml"]';
        this.logoUploaded = '.tenant-logo';
        this.oidcIssuerUrlElement = 'pre.yaml-code';
        this.metadataXmlFileInput = 'input[type="file"][accept=".xml"]';
        this.securityInvalidIcon = '[aid="security-expansion-toggle"] .text-red';
        this.securityWarningIcon = '[aid="security-expansion-toggle"] .text-warning';
        this.certificateWarningBannerTitle = '.certificate-warning .notification-alert-bar__title';
        this.certificateFieldErrorMessage = '.q-field--error .q-field__messages';
        this.cancelIdpEditButton = '//span[contains(text(),"Cancel")]';
    }

    get brandingTestLogoDataUrl(): string {
        return 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII=';
    }

    get allSections(): string[] {
        return [
            this.analyticsSection,
            this.resourcesSection,
            this.workloadsSection,
            this.clusterAuthenticationSection,
            this.securityOption,
            this.branding,
            this.notificationsSection
        ];
    }

    async openGeneralSettingsPage(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();
    }

    async navigateToGeneralSettingsSecurity(): Promise<void> {
        await this.openGeneralSettingsPage();
        const isIdentityProviderVisible = await this.page.locator(this.addIdentityProvider).isVisible();
        if (!isIdentityProviderVisible) {
            await this.page.locator(this.securityOption).click({ timeout: 10000 });
        }
    }

    async navigateToGeneralSettingsAuthentication(): Promise<void> {
        await this.openGeneralSettingsPage();
        await this.page.locator(this.clusterAuthenticationSection).click({ timeout: 10000 });
    }

    async validateOidcIssuerUrl(
        expectedBaseUrl: string,
        expectedTenantName: string,
        isSelfHosted: boolean = true
    ): Promise<{ isVisible: boolean; url: string; isValid: boolean }> {
        try {
            await this.page.waitForSelector(this.oidcIssuerUrlElement, { state: 'visible', timeout: 10000 });

            const yamlContent = await this.page.locator(this.oidcIssuerUrlElement).textContent();

            if (!yamlContent) {
                throw new Error('YAML content not found');
            }

            const issuerUrlMatch = yamlContent.match(/--oidc-issuer-url=([^\s]+)/);

            if (!issuerUrlMatch || !issuerUrlMatch[1]) {
                throw new Error('OIDC issuer URL not found in YAML content');
            }

            const issuerUrl = issuerUrlMatch[1];

            let issuerBaseUrl = expectedBaseUrl;
            if (!isSelfHosted) {
                const url = new URL(expectedBaseUrl);
                const hostParts = url.hostname.split('.');
                if (hostParts.length > 2 && hostParts[0] === expectedTenantName) {
                    url.hostname = hostParts.slice(1).join('.');
                }
                issuerBaseUrl = url.origin;
            }

            const expectedUrl = `${issuerBaseUrl}/auth/realms/${expectedTenantName}`;
            const isValid = issuerUrl === expectedUrl;

            if (!isValid) {
                logger.error(`OIDC issuer URL mismatch. Expected: ${expectedUrl}, Got: ${issuerUrl}`);
            } else {
                logger.debug(`OIDC Issuer URL validated successfully: ${issuerUrl}`);
            }

            return {
                isVisible: true,
                url: issuerUrl,
                isValid: isValid
            };
        } catch (error) {
            logger.error('Error validating OIDC issuer URL:', error);
            throw error;
        }
    }

    async navigateToGeneralSettingsNotifications(): Promise<void> {
        await this.openGeneralSettingsPage();
        await this.page.locator(this.notificationsSection).click({ timeout: 10000 });
    }

    async navigateToGeneralSettingsWorkloads(): Promise<void> {
        await this.openGeneralSettingsPage();
        await this.page.locator(this.workloadsSection).click({ timeout: 10000 });
        // await this.page.locator('[aid="header-leave-page-button"]').click();
    }

    async hardNavigateToGeneralSettings(): Promise<void> {
        await this.page.goto(`${config.tenantUrl}/general-settings`);
    }

    async workloadsSectionEnableModels(): Promise<void> {
        try {
            await apiUtils.setFeatureValue(
                EFeatureFlagsKeys.MODEL_CATALOG,
                EFeatureFlagsNames.MODEL_CATALOG,
                true
            );
        } catch (error) {
            logger.error('Error removing Identity Provider or Logout URL', error);
        }
    }

    async workloadsSectionDisableModels(): Promise<void> {
        await apiUtils.setFeatureValue(
            EFeatureFlagsKeys.MODEL_CATALOG,
            EFeatureFlagsNames.MODEL_CATALOG,
            false
        );
    }

    async isSecurityTooltipsVisible(): Promise<boolean> {
        await this.page.locator(this.ssoSettingsTooltip).hover();
        await this.page.waitForSelector(this.ssoSettingsTooltip, { state: 'visible' });
        await this.page.locator(this.logoutRedirectUrlTooltip).hover();
        await this.page.waitForSelector(this.logoutRedirectUrlTooltip, { state: 'visible' });
        await this.page.locator(this.sessionTimeoutTooltip).hover();
        await this.page.waitForSelector(this.sessionTimeoutTooltip, { state: 'visible' });
        return true;
    }

    async setIdentityProviderUI(ipdType: string, url: string = ''): Promise<void> {
        const responseTenantName = await tenantUtils.getTenantConfiguration(
            flowRootState.users.system_administrator.accessToken
        );
        if (url === '') {
            url = `${settings.isSelfHosted ? config.tenantUrl : process.env.APP_URL}/auth/realms/${responseTenantName.tenantName}/protocol/saml/descriptor`;
        }
        await this.page.locator(this.addIdentityProvider).click();
        await this.page.locator(`text=${ipdType}`).click({ timeout: 10000 });
        if (ipdType === 'SAML 2.0') {
            await this.page.locator(this.fromURL).click({ timeout: 10000 });
            await this.page.locator(this.urlSamlTextLabel).click({ timeout: 10000 });
            await this.page.locator(this.urlSamlTextLabel).fill(url, { timeout: 10000 });
            await this.page.locator(this.mappersLabel).nth(2).fill('GID');
            await this.page.locator(this.mappersLabel).nth(4).fill('UID');
            await this.page.locator(this.mappersLabel).nth(6).fill('GROUPS');
            await this.page.locator(this.mappersLabel).nth(8).fill('SUPPLEMENTARYGROUPS');
            await this.page.locator(this.saveButton).click();
        } else {
            await this.page.locator(this.urlOidcTextLabel).click({ timeout: 10000 });
            await this.page
                .locator(this.urlOidcTextLabel)
                .fill(
                    `${settings.isSelfHosted ? config.tenantUrl : process.env.APP_URL}/auth/realms/${responseTenantName.tenantName}/.well-known/openid-configuration`,
                    {
                        timeout: 10000
                    }
                );
            await this.page.locator(this.mappersLabel).nth(2).fill('GID');
            await this.page.locator(this.mappersLabel).nth(4).fill('UID');
            await this.page.locator(this.mappersLabel).nth(6).fill('GROUPS');
            await this.page.locator(this.mappersLabel).nth(8).fill('SUPPLEMENTARYGROUPS');
            await this.page.locator(this.clinetIDLabel).fill('id');
            await this.page.locator(this.clinetSecretLabel).fill('secret');
            await this.page.locator(this.saveButton).click();
        }
    }

    async validateLinkToSamlDescriptor(): Promise<string> {
        const [newPage] = await Promise.all([
            this.page.context().waitForEvent('page'),
            this.page.locator(this.linkToSamlDescriptor).click()
        ]);
        await newPage.waitForLoadState('networkidle');
        const url = newPage.url();
        await newPage.close();
        return url;
    }

    async validateOIDCScopes(): Promise<boolean> {
        await this.page.locator(this.addIdentityProvider).click();
        await this.page.locator('text=OpenID Connect').click({ timeout: 10000 });
        await this.page.locator(this.urlOidcTextLabel).click({ timeout: 10000 });
        await this.page
            .locator(this.urlOidcTextLabel)
            .fill(`${config.tenantUrl}/auth/realms/runai/.well-known/openid-configuration`, {
                timeout: 10000
            });
        await this.page.locator(this.mappersLabel).nth(2).fill('GID');
        await this.page.locator(this.mappersLabel).nth(4).fill('UID');
        await this.page.locator(this.mappersLabel).nth(6).fill('GROUPS');
        await this.page.locator(this.mappersLabel).nth(8).fill('SUPPLEMENTARYGROUPS');
        await this.page.locator(this.clinetIDLabel).fill('id');
        await this.page.locator(this.clinetSecretLabel).fill('secret');
        await this.page.locator(this.scopeCombobox).click();
        await this.page.locator(this.scopeCombobox).fill('e2e');
        await this.page.press(this.scopeCombobox, 'Enter');
        await this.page.locator(this.scopeCombobox).click();
        await this.page.locator(this.scopeCombobox).fill('test');
        await this.page.press(this.scopeCombobox, 'Enter');
        await this.page.locator(this.arrowScopes).click({ timeout: 10000 });
        await this.page.locator('.q-chip__content:has-text("openid")').nth(0).isDisabled({ timeout: 10000 });
        await this.page.locator('.q-chip__content:has-text("e2e")').nth(0).isDisabled({ timeout: 10000 });
        await this.page.locator('.q-chip__content:has-text("test")').nth(0).isDisabled({ timeout: 10000 });
        return true;
    }

    async removeIdentityProvider(): Promise<void> {
        await this.page.locator(this.deleteButton).click();
        await this.page.locator(this.removeButtonAlert).click();
        try {
            await this.page.waitForSelector(this.addIdentityProvider, { state: 'visible' });
            logger.debug('Identity Provider was removed');
        } catch (error) {
            throw new Error('Identity Provider was not removed' + error);
        }
    }

    async isValidatorSAMLCreated(): Promise<boolean> {
        await this.page.waitForSelector(this.validatorSAMLCreated, { state: 'visible' });
        return true;
    }

    async uploadIdentityProviderMetadataFile(fileName: string, buffer: Buffer, mimeType = 'text/xml'): Promise<void> {
        await this.page.locator(this.addIdentityProvider).click();
        await this.page.locator('text=SAML 2.0').click({ timeout: 10000 });
        await this.page.locator(this.metadataXmlFileInput).setInputFiles({ name: fileName, mimeType, buffer });
        await this.page.locator(this.mapperTextLabel).nth(0).fill('UID');
        await this.page.locator(this.mapperTextLabel).nth(1).fill('GID');
        await this.page.locator(this.mapperTextLabel).nth(2).fill('GROUPS');
        await this.page.locator(this.mapperTextLabel).nth(3).fill('SUPPLEMENTARYGROUPS');
        await this.page.locator(this.saveButton).click();
    }

    async replaceIdentityProviderMetadataFile(fileName: string, buffer: Buffer, mimeType = 'text/xml'): Promise<void> {
        await this.editIdentityProvider();
        await this.page.locator(this.metadataXmlFileInput).setInputFiles({ name: fileName, mimeType, buffer });
        await this.page.locator(this.saveButton).click();
    }

    // Locator.isVisible() is a non-waiting snapshot check - it does not retry, so
    // calling it immediately after a Save click can catch the DOM before the async
    // certificate-status computation lands. waitForSelectorVisible below uses Playwright's
    // auto-retrying waitFor instead, so callers reliably see the eventual state.
    private async waitForSelectorVisible(selector: string, timeoutMs = 10000): Promise<boolean> {
        return this.page
            .locator(selector)
            .waitFor({ state: 'visible', timeout: timeoutMs })
            .then(() => true)
            .catch(() => false);
    }

    async getSecuritySectionStatus(): Promise<'none' | 'warning' | 'invalid'> {
        if (await this.waitForSelectorVisible(this.securityInvalidIcon)) return 'invalid';
        if (await this.waitForSelectorVisible(this.securityWarningIcon)) return 'warning';
        return 'none';
    }

    async getCertificateWarningBannerText(): Promise<string | null> {
        if (!(await this.waitForSelectorVisible(this.certificateWarningBannerTitle))) return null;
        return this.page.locator(this.certificateWarningBannerTitle).textContent();
    }

    async isCertificateFieldInvalid(expectedErrorMessage = 'Certificate expired'): Promise<boolean> {
        return this.waitForSelectorVisible(
            `${this.certificateFieldErrorMessage}:has-text("${expectedErrorMessage}")`
        );
    }

    async isIdentityProviderSaved(): Promise<boolean> {
        return this.waitForSelectorVisible(this.editIpdButton);
    }

    async cancelIdentityProviderEdit(): Promise<void> {
        await this.page.locator(this.cancelIdpEditButton).click();
    }

    async isValidatorOIDCCreated(): Promise<boolean> {
        await this.page.waitForSelector(this.validatorOIDCCreated, { state: 'visible' });
        return true;
    }

    async setLogoutRedirectURL(url: string): Promise<void> {
        await this.page.locator(this.RedirectURLButton).click();
        await this.page.locator(this.redirectURLLable).fill(url, { timeout: 10000 });
        await this.page.locator(this.saveButton).click();
        await this.page.waitForTimeout(1000);
    }

    async removeLogoutUrl(): Promise<void> {
        try {
            await this.page.waitForTimeout(1000);
            await this.page.locator(this.deleteButton).click();
            await this.page.locator(this.removeButtonAlert).click();

            await this.page.waitForSelector(this.addIdentityProvider, { state: 'visible' });
            logger.debug('Logout Url was removed');
        } catch (error) {
            throw new Error(`LogoutUrl was not removed ${error}`);
        }
    }

    async setSessionTimeout(number: string): Promise<void> {
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.sessionTimeoutLabel).fill(number, { timeout: 10000 });
        await this.page.locator(this.sessionTimeoutLabel).press('Tab');
        await this.page.waitForTimeout(1000);
    }

    async generalSecurityForceCleaner(attempt = 0): Promise<void> {
        await this.navigateToGeneralSettingsSecurity();
        await this.setSessionTimeout('60');
        // Remove any configured IDP / logout-redirect-URL. A clean page exposes no
        // visible delete control (the hidden "delete-Email server" button and the
        // confirmation-modal button are excluded by deleteButton), so we key off
        // whether a real delete control is present rather than the add-button state.
        const deletables = this.page.locator(this.deleteButton);
        const count = await deletables.count();
        if (count === 0) {
            logger.debug('Security settings already clean - nothing to remove');
            return;
        }
        if (attempt >= 5) {
            logger.warn(
                `generalSecurityForceCleaner: ${count} delete control(s) still present after ${attempt} attempts`
            );
            return;
        }
        try {
            await deletables.first().click();
            await this.page.locator(this.removeButtonAlert).click();
            await this.page.waitForTimeout(1000);
            await this.generalSecurityForceCleaner(attempt + 1);
        } catch (error) {
            logger.error('Error removing Identity Provider or Logout URL', error);
            throw error;
        }
    }

    async verifySessionEndAfterTimePass(timeoutMinutes: number = 1): Promise<void> {
        try {
            const waitTime = timeoutMinutes * 60 * 1000 + 10000;
            await this.page.waitForTimeout(waitTime);
            await this.page.waitForSelector('#kc-form-login', { state: 'visible', timeout: 30000 });
            logger.debug('Session was ended');
        } catch (error) {
            throw new Error(`Session was not ended after ${timeoutMinutes} minute(s) passed ${error}`);
        }
    }

    async logoutAndVerifyUrl(): Promise<void> {
        await this.page.locator('[aid="profile-btn"]').click({ timeout: 10000 });
        await this.page.locator('text=Sign out').click();
        await this.page.waitForLoadState();
        try {
            await this.page.waitForSelector('[aria-labelledby="nvidia_logo_desktop"]', { state: 'visible' });
            logger.debug('Logout was successful');
        } catch (error) {
            throw new Error(`Logout was not successful ${error}`);
        }
    }

    async validateAllOptionsClickable(): Promise<void> {
        await this.page.locator(this.analyticsSection).click();
        await this.page.locator(this.resourcesSection).click();
        await this.page.locator(this.notificationsSection).click();
        await this.page.locator(this.workloadsSection).click();
        await this.page.locator(this.clusterAuthenticationSection).click();
    }

    async validateAllSectionsHaveName(): Promise<void> {
        const sections = [
            this.analyticsSection,
            this.resourcesSection,
            this.notificationsSection,
            this.workloadsSection,
            this.clusterAuthenticationSection
        ];
        for (const section of sections) {
            const sectionName = await this.page.locator(section).textContent();
            if (!sectionName) {
                throw new Error(`Section ${section} does not have a name`);
            }
        }
    }

    async setMapperValues(
        ipdType: string,
        uid: string,
        gid: string,
        Groups: string,
        sUPPLEMENTARYGROUPS: string,
        email: string,
        firstName: string,
        lastName: string
    ): Promise<string> {
        const responseTenantName = await tenantUtils.getTenantConfiguration(
            flowRootState.users.system_administrator.accessToken
        );
        await this.page.locator(this.addIdentityProvider).click();
        await this.page.locator(`text=${ipdType}`).click({ timeout: 10000 });
        await this.page.locator(this.fromURL).click({ timeout: 10000 });
        await this.page.locator(this.urlSamlTextLabel).click({ timeout: 10000 });
        await this.page
            .locator(this.urlSamlTextLabel)
            .fill(
                `${settings.isSelfHosted ? config.tenantUrl : process.env.APP_URL}/auth/realms/${responseTenantName.tenantName}/protocol/saml/descriptor`,
                { timeout: 10000 }
            );
        await this.page.locator(this.mapperTextLabel).nth(0).fill(uid);
        await this.page.locator(this.mapperTextLabel).nth(1).fill(gid);
        await this.page.locator(this.mapperTextLabel).nth(2).fill(Groups);
        await this.page.locator(this.mapperTextLabel).nth(3).fill(sUPPLEMENTARYGROUPS);
        await this.page.locator(this.mapperTextLabel).nth(4).fill(email);
        await this.page.locator(this.mapperTextLabel).nth(5).fill(firstName);
        await this.page.locator(this.mapperTextLabel).nth(6).fill(lastName);
        await this.page.locator(this.saveButton).click();
        const mappersSaved = await this.page.locator('text=Settings saved').nth(0).textContent();
        return mappersSaved;
    }

    async getMapperValues(): Promise<string[]> {
        await this.page.locator(this.addIdentityProvider).click();
        await this.page.locator('text=SAML 2.0').click({ timeout: 10000 });
        const elements = await this.page.$$(this.mapperTextLabel);
        const values = [];
        for (const element of elements) {
            const value = await element.inputValue();
            values.push(value);
        }
        for (const value of values) {
            if (value === '') {
                throw new Error('One of the mapper values is empty or not set correctly');
            }
        }
        return values;
    }

    async configureEmailServerSh(config: NotificationChannelForPatchConfig): Promise<void> {
        try {
            await waitUtils.waitToLoadNewPageOrSaveChanges();
            const addEmailServerEnabled = await this.page.locator(this.addEmailServer).isVisible();
            if (!addEmailServerEnabled) {
                await this.page.locator(this.deleteEmailButton).click();
                await this.page.locator(this.removeButtonAlert).click();
            }

            await waitUtils.waitToLoadNewPageOrSaveChanges();
            await this.page.locator(this.addEmailServer).click({ timeout: 1000 });
            await this.page.locator('[aid="smpt-host-input"]').fill(config.smtpHost, { timeout: 1000 });
            await this.page.locator('[aid="smpt-port-input"]').fill(`${config.smtpPort}`, { timeout: 1000 });
            await this.page.locator('[aid="smpt-username-input"]').fill(config.username, { timeout: 1000 });
            await this.page.locator('[aid="smpt-password-input"]').fill(config.password, { timeout: 1000 });
            await this.page
                .locator('[aid="smpt-from-email-input"]')
                .fill('test@example.com', { timeout: 1000 });
            await this.page
                .locator('[aid="smpt-from-name-input"]')
                .fill('Runai-e2e-ns-tests', { timeout: 1000 });
            await this.page.locator('[aid="verify-button"]').click({ timeout: 5000 });
            await waitUtils.waitToLoadNewPageOrSaveChanges();
            await this.page.locator('[aid="primary-button"]').click();
            await waitUtils.waitToLoadNewPageOrSaveChanges();
        } catch (error) {
            logger.error(`Error while trying to configure email server: ${error}`);
        }
    }

    async deleteEmailServerSh(): Promise<void> {
        await waitUtils.waitToLoadNewPageOrSaveChanges();
        const emailNotConfigured = await this.page.locator(this.addEmailServer).isVisible();
        if (emailNotConfigured) return;

        await this.page.locator(this.deleteEmailButton).click();
        await this.page.locator(this.removeButtonAlert).click();
    }

    async enableNotificationsSaas(): Promise<void> {
        try {
            await waitUtils.waitForElementToExist(this, '[aid="toggle-setting-tenant.smtp_enabled"]');
            if (!(await this.page.locator('[aid="toggle-setting-tenant.smtp_enabled"]').isChecked())) {
                await this.page.locator('[aid="toggle-setting-tenant.smtp_enabled"]').click();
            }
        } catch (error) {
            logger.error(`Error while trying to configure email server: ${error}`);
        }
    }

    async disableNotificationsSaas(): Promise<void> {
        await waitUtils.waitForElementToExist(this, '[aid="toggle-setting-tenant.smtp_enabled"]');
        if (await this.page.locator('[aid="toggle-setting-tenant.smtp_enabled"]').isChecked()) {
            await this.page.locator('[aid="toggle-setting-tenant.smtp_enabled"]').click();
        }
    }

    async editIdentityProvider(): Promise<void> {
        await this.page.locator(this.editIpdButton).click({ timeout: 10000 });
    }

    async updateSamlIdentityProviderMetadataUrl(url: string): Promise<void> {
        await this.editIdentityProvider();
        await this.page.locator(this.fromURL).click({ timeout: 10000 });
        await this.page.locator(this.urlSamlTextLabel).click({ timeout: 10000 });
        await this.page.locator(this.urlSamlTextLabel).fill(url, { timeout: 10000 });
        await this.page.locator(this.saveButton).click();
    }

    async validateMapperValuesSaved(): Promise<string[]> {
        await this.editIdentityProvider();
        await this.page.waitForTimeout(1000);
        const elements = await this.page.$$(this.mapperTextLabel);
        const values = [];
        for (const element of elements) {
            const value = await element.inputValue();
            values.push(value);
        }
        return values;
    }

    async toggleAutoRedirectSSO(): Promise<void> {
        const toggleWrapper = this.page.locator('div.toggle-setting', {
            hasText: 'Auto-Redirect to SSO'
        });

        const toggle = toggleWrapper.locator('.q-toggle__inner');
        await toggle.click({ timeout: 10000 });
        await this.page.waitForTimeout(1000);
        const toggleContainer = toggleWrapper.locator('.q-toggle');
        const isChecked = await toggleContainer.getAttribute('aria-checked');
        if (isChecked === 'true') {
            logger.debug('Auto Redirect SSO is enabled');
        } else {
            throw new Error('Auto Redirect SSO is not enabled');
        }
    }

    async disableAutoRedirectSSO(): Promise<void> {
        const toggleWrapper = this.page.locator('div.toggle-setting', {
            hasText: 'Auto-Redirect to SSO'
        });
        const toggle = toggleWrapper.locator('.q-toggle__inner');
        await toggle.click({ timeout: 10000 });
        await this.page.waitForTimeout(1000);
        const toggleContainer = toggleWrapper.locator('.q-toggle');
        const isChecked = await toggleContainer.getAttribute('aria-checked');
        if (isChecked === 'false') {
            logger.debug('Auto Redirect SSO is disabled');
        } else {
            throw new Error('Auto Redirect SSO is not disabled');
        }
    }

    async downloadIdpMetadataXMLFile(): Promise<void> {
        await this.page.locator(this.downloadIdpMetadataXMLFileButton).click({ timeout: 10000 });
    }

    async updateIdpWhenTheDiscoveryUrlEmpty(): Promise<boolean> {
        let requestMatched = false;
        this.page.on('request', (request) => {
            if (
                request.url().includes('/api/v1/idps') &&
                request.method() === 'GET' &&
                request.headers()['accept']?.includes('application/json') &&
                request.headers()['authorization']?.startsWith('Bearer')
            ) {
                requestMatched = true;
            }
        });
        await this.page.locator(this.editOpenIdConnectButton).click({ timeout: 10000 });
        await this.page.locator(this.clinetIDLabel).fill('idupdated');
        await this.page.locator(this.clinetSecretLabel).fill('secretupdated');
        await this.page.locator(this.saveButton).click();
        await this.page.waitForTimeout(1000);
        return requestMatched;
    }

    async validateOnlySecuritySettingsVisible(): Promise<boolean> {
        const securityVisible = await this.page.locator(this.securityOption).isVisible();
        if (!securityVisible) {
            return false;
        }

        const otherSections = [
            this.analyticsSection,
            this.resourcesSection,
            this.workloadsSection,
            this.notificationsSection
        ];
        for (const section of otherSections) {
            const isVisible = await this.page.locator(section).isVisible();
            if (isVisible) {
                return true;
            }
        }
        return false;
    }

    async ValidateAllSettingsVisibleExceptSecurity(): Promise<boolean> {
        const allSections = [
            this.analyticsSection,
            this.resourcesSection,
            this.workloadsSection,
            this.notificationsSection
        ];
        for (const section of allSections) {
            const isVisible = await this.page.locator(section).isVisible();
            if (!isVisible) {
                return false;
            }
        }
        return true;
    }

    async enableResourcesProjectQuotaCheck(): Promise<void> {
        await this.goToPage();
        await this.page.locator(this.resourcesSection).click();
        const checked = await this.locator(this.limtQuotaOverSubscriptionToggle).getAttribute('aria-checked');
        if (checked === 'false') {
            await this.locator(this.limtQuotaOverSubscriptionToggle).click();
        }
    }

    async disableResourcesProjectQuotaCheck(): Promise<void> {
        await this.goToPage();
        await this.page.locator(this.resourcesSection).click();
        const checked = await this.locator(this.limtQuotaOverSubscriptionToggle).getAttribute('aria-checked');
        if (checked === 'true') {
            await this.locator(this.limtQuotaOverSubscriptionToggle).click();
        }
    }

    async enableWorkloadsDataVolumes(): Promise<void> {
        await this.navigateToGeneralSettingsWorkloads();
        const checked = await this.locator(this.dataVolumeToggle).getAttribute('aria-checked');
        if (checked === 'false') {
            await this.locator(this.dataVolumeToggle).click();
        }
    }

    async isWorkloadsSectionExpanded(): Promise<boolean> {
        const ariaLabel = await this.page.locator(this.workloadsExpansionHeader).getAttribute('aria-label');
        return ariaLabel.includes('Collapse');
    }

    async disableFlexWorkloadTemplate(): Promise<void> {
        await this.goToPage();
        const isExpanded = await this.isWorkloadsSectionExpanded();
        if (!isExpanded) await this.page.locator(this.workloadsSection).click();

        const checked = await this.locator(this.flexWorkloadTemplateToggle).getAttribute('aria-checked');
        if (checked === 'true') {
            await this.locator(this.flexWorkloadTemplateToggle).click();
        }
    }

    async enableFlexWorkloadTemplate(): Promise<void> {
        await this.goToPage();
        const isExpanded = await this.isWorkloadsSectionExpanded();
        if (!isExpanded) await this.page.locator(this.workloadsSection).click();

        const checked = await this.locator(this.flexWorkloadTemplateToggle).getAttribute('aria-checked');
        if (checked === 'false') {
            await this.locator(this.flexWorkloadTemplateToggle).click();
        }
    }

    async isFlexWorkloadTemplateEnabled(): Promise<boolean> {
        await this.goToPage();
        const isExpanded = await this.isWorkloadsSectionExpanded();
        if (!isExpanded) await this.page.locator(this.workloadsSection).click();

        const checked = await this.locator(this.flexWorkloadTemplateToggle).getAttribute('aria-checked');

        return checked === 'true';
    }

    async disableFlexWorkloadSubmission(): Promise<void> {
        await this.goToPage();
        const isExpanded = await this.isWorkloadsSectionExpanded();
        if (!isExpanded) await this.page.locator(this.workloadsSection).click();

        const checked = await this.locator(this.flexWorkloadSubmissionToggle).getAttribute('aria-checked');
        if (checked === 'true') {
            await this.locator(this.flexWorkloadSubmissionToggle).click();
        }
    }

    async enableFlexWorkloadSubmission(): Promise<void> {
        await this.goToPage();
        const isExpanded = await this.isWorkloadsSectionExpanded();
        if (!isExpanded) await this.page.locator(this.workloadsSection).click();

        const checked = await this.locator(this.flexWorkloadSubmissionToggle).getAttribute('aria-checked');
        if (checked === 'false') {
            await this.locator(this.flexWorkloadSubmissionToggle).click();
        }
    }

    async clickonBranding(): Promise<void> {
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.branding).click({ timeout: 10000 });
    }

    async uploadLogo(fileName: string = 'logo.png'): Promise<void> {
        try {
            await this.clickonBranding();
            await this.page.waitForTimeout(1000);
            await this.page.locator(this.addLogoButton).click();
            await this.page.waitForTimeout(500);
            let logoPath: string;
            try {
                const currentDir = path.dirname(fileURLToPath(import.meta.url));
                logoPath = path.resolve(currentDir, '../../../../../', fileName);
            } catch {
                logoPath = path.resolve(process.cwd(), fileName);
            }
            logger.debug(`Attempting to upload logo from path: ${logoPath}`);
            await this.page.locator(this.logoFileInput).setInputFiles(logoPath);
            await this.page.waitForTimeout(1000);
            await this.page.locator(this.saveButton).click();
            await this.page.waitForTimeout(1000);
            logger.debug('Logo upload completed successfully');
        } catch (error) {
            logger.error('Error during logo upload:', error);
            throw error;
        }
    }

    async removeExistingLogoIfExists(): Promise<void> {
        try {
            const editLogoButton = this.page.locator('button:has(i.q-icon.ph-bold.ph-pencil-simple)');
            const isEditButtonVisible = await editLogoButton.isVisible();

            if (isEditButtonVisible) {
                logger.debug('Edit logo button found, removing existing logo');
                const removeButton = this.page.locator('button:has(i.q-icon.ph-bold.ph-x)');
                await removeButton.click();
                await this.page.waitForTimeout(1000);
                logger.debug('Existing logo removed successfully');
            } else {
                logger.debug('No existing logo to remove');
                return;
            }
        } catch (error) {
            logger.error('Error while checking/removing existing logo:', error);
            return;
        }
    }

    async isLogoUploaded(): Promise<boolean> {
        const logo = await this.page.locator(this.logoUploaded).isVisible();
        return logo;
    }

    async isLogoRemoved(): Promise<boolean> {
        const logo = await this.page.locator(this.logoUploaded).isVisible();
        return !logo;
    }

    async editLogo(logoPath: string): Promise<void> {
        await this.page.locator('button:has(i.q-icon.ph-bold.ph-pencil-simple)').click();
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.logoFileInput).setInputFiles(logoPath);
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.saveButton).click();
        await this.page.waitForTimeout(1000);
    }
}
