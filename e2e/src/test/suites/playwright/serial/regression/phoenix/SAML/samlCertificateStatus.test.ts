/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { config, tags, settings } from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { samlCertificateFixtures } from '../../../../../../../modules/helpers/utils/SamlCertificateFixtures';

test.describe('SAML certificate status', { tag: [tags.teams.phoenix, tags.suites.legacySphinx, tags.tool.ui, tags.runOn.serialCI] }, () => {
    shouldRunTests(settings.isSelfHosted && !isOCP())(() => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let generalSettings: GeneralSettings;
        let videoName: string;

        test.beforeEach('init browsers and classes', async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
            generalSettings = new GeneralSettings(browserEngine.currentPage);
            // Guard against a leftover SAML IdP from a previous test/run (e.g. a prior
            // failed attempt, or another SAML suite sharing this env) blocking the
            // "add identity provider" button this suite depends on for a clean start.
            await generalSettings.generalSecurityForceCleaner();
        });

        test.afterEach('stop browserEngine', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, videoName);
            } finally {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
                await generalSettings.generalSecurityForceCleaner();
                await browserEngine.stopAndArchive(testInfo, videoName);
            }
        });

        test('should show no status icon or banner for a valid certificate', async () => {
            videoName = 'shouldShowNoStatusForValidCertificate';
            await generalSettings.navigateToGeneralSettingsSecurity();
            const fixture = samlCertificateFixtures.valid();
            await generalSettings.uploadIdentityProviderMetadataFile(fixture.fileName, fixture.buffer);
            expect(
                await generalSettings.isValidatorSAMLCreated(),
                'SAML IDP should be created for a valid certificate'
            ).toBe(true);
            expect(
                await generalSettings.getSecuritySectionStatus(),
                'Security section should show no status icon for a valid certificate'
            ).toBe('none');
            expect(
                await generalSettings.getCertificateWarningBannerText(),
                'no certificate warning banner should be shown for a valid certificate'
            ).toBeNull();
        });

        test('should show a warning icon and banner for a certificate expiring soon', async () => {
            videoName = 'shouldShowWarningForExpiringSoonCertificate';
            await generalSettings.navigateToGeneralSettingsSecurity();
            const fixture = samlCertificateFixtures.expiringSoon();
            await generalSettings.uploadIdentityProviderMetadataFile(fixture.fileName, fixture.buffer);
            expect(
                await generalSettings.getSecuritySectionStatus(),
                'Security section should show a warning icon for a certificate expiring soon'
            ).toBe('warning');
            expect(
                await generalSettings.getCertificateWarningBannerText(),
                'certificate warning banner should say the certificate is expiring soon'
            ).toContain('expiring soon');
        });

        test('should block creation and show an inline error for an already-expired certificate', async () => {
            videoName = 'shouldBlockCreationForExpiredCertificate';
            await generalSettings.navigateToGeneralSettingsSecurity();
            const fixture = samlCertificateFixtures.expired();
            await generalSettings.uploadIdentityProviderMetadataFile(fixture.fileName, fixture.buffer);
            expect(
                await generalSettings.isCertificateFieldInvalid(),
                'the certificate expiration field should show an inline error for an already-expired certificate'
            ).toBe(true);
            expect(
                await generalSettings.isIdentityProviderSaved(),
                'the identity provider should not be created while its certificate is already expired'
            ).toBe(false);
        });

        test('should block creation and show an inline error for a malformed certificate', async () => {
            videoName = 'shouldBlockCreationForMalformedCertificate';
            await generalSettings.navigateToGeneralSettingsSecurity();
            const fixture = samlCertificateFixtures.malformed();
            await generalSettings.uploadIdentityProviderMetadataFile(fixture.fileName, fixture.buffer);
            expect(
                await generalSettings.isCertificateFieldInvalid('Invalid SAML metadata'),
                'the certificate expiration field should show an inline error for a malformed certificate, not silently accept it'
            ).toBe(true);
            expect(
                await generalSettings.isIdentityProviderSaved(),
                'the identity provider should not be created while its certificate is malformed'
            ).toBe(false);
        });

        test('should block an in-place replacement with an expired certificate and leave the existing valid one untouched', async () => {
            videoName = 'shouldBlockReplaceWithExpiredCertificate';
            await generalSettings.navigateToGeneralSettingsSecurity();
            const validFixture = samlCertificateFixtures.valid();
            await generalSettings.uploadIdentityProviderMetadataFile(validFixture.fileName, validFixture.buffer);
            expect(
                await generalSettings.isIdentityProviderSaved(),
                'the identity provider should be created for a valid certificate'
            ).toBe(true);
            expect(
                await generalSettings.getSecuritySectionStatus(),
                'Security section should show no status icon for a valid certificate'
            ).toBe('none');

            const expiredFixture = samlCertificateFixtures.expired();
            await generalSettings.replaceIdentityProviderMetadataFile(expiredFixture.fileName, expiredFixture.buffer);
            expect(
                await generalSettings.isCertificateFieldInvalid(),
                'the certificate expiration field should show an inline error for the replacement expired certificate'
            ).toBe(true);

            await generalSettings.cancelIdentityProviderEdit();
            expect(
                await generalSettings.getSecuritySectionStatus(),
                'the existing valid identity provider should be left untouched by the blocked replacement attempt'
            ).toBe('none');
            expect(
                await generalSettings.getCertificateWarningBannerText(),
                'no certificate warning banner should be shown after cancelling the blocked replacement'
            ).toBeNull();
        });
    });
});
