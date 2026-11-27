/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { config, logger, settings, tags, usersManager } from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test } from '@playwright/test';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`SAML Metadata UI, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, () => {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let videoName: string;

            test.beforeAll(async function () {
                await usersManager.deletePreSeededOidcIdp();
            });

            test.afterAll(async function () {
                await usersManager.restorePreSeededOidcIdp();
            });

            test.beforeEach('init browsers and classes', async function () {
                await Keycloak.createRealm('metadata-test');
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
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
                    await Keycloak.deleteRealm('metadata-test');
                    await browserEngine.stopAndArchive(testInfo, videoName);
                }
            });

            test('should verify SAML Metadata URL Configurations', async function () {
                videoName = 'shouldVerifySAMLMetadataURLConfiguration';
                await generalSettings.navigateToGeneralSettingsSecurity();
                const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl('metadata-test');
                await generalSettings.setIdentityProviderUI('SAML 2.0', metadataXmlUrl);
                expect(await generalSettings.isValidatorSAMLCreated(), 'SAML IDP validator should be created after setting metadata URL').to.be.equal(true);
                logger.debug(`metadataXmlUrl: ${metadataXmlUrl}`);
                const response = await fetch(metadataXmlUrl);
                const metadata = await response.text();
                logger.debug(`metadata: ${metadata}`);
                expect(metadata, 'SAML metadata should contain EntityDescriptor element').to.contain('<md:EntityDescriptor');
                expect(metadata, 'SAML metadata entityID should match tenant realm URL').to.contain(`entityID="${config.tenantUrl}/auth/realms/metadata-test"`);
                expect(metadata, 'SAML metadata should contain IDPSSODescriptor element').to.contain('<md:IDPSSODescriptor');
                expect(metadata, 'SAML metadata should require signed authn requests').to.contain('WantAuthnRequestsSigned="true"');
                expect(metadata, 'SAML metadata should contain SingleSignOnService element').to.contain('<md:SingleSignOnService');
                expect(metadata, 'SAML metadata should include HTTP-POST binding').to.contain('Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"');
                expect(metadata, 'SAML metadata should contain SingleLogoutService element').to.contain('<md:SingleLogoutService');
                expect(metadata, 'SAML metadata should contain signing KeyDescriptor').to.contain('<md:KeyDescriptor use="signing"');
                expect(metadata, 'SAML metadata should contain X509Certificate element').to.contain('<ds:X509Certificate>');
                expect(metadata, 'SAML metadata should contain SOAP ArtifactResolutionService').to.contain(
                    `<md:ArtifactResolutionService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml/resolve" index="0"></md:ArtifactResolutionService>`
                );
                expect(metadata, 'SAML metadata should contain HTTP-POST SingleLogoutService').to.contain(
                    `<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleLogoutService>`
                );
                expect(metadata, 'SAML metadata should contain HTTP-Redirect SingleLogoutService').to.contain(
                    `<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleLogoutService>`
                );
                expect(metadata, 'SAML metadata should contain HTTP-Artifact SingleLogoutService').to.contain(
                    `<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleLogoutService>`
                );
                expect(metadata, 'SAML metadata should contain SOAP SingleLogoutService').to.contain(
                    `<md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleLogoutService>`
                );
                expect(metadata, 'SAML metadata should support persistent NameIDFormat').to.contain(
                    '<md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat>'
                );
                expect(metadata, 'SAML metadata should support transient NameIDFormat').to.contain(
                    '<md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</md:NameIDFormat>'
                );
                expect(metadata, 'SAML metadata should support unspecified NameIDFormat').to.contain(
                    '<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat>'
                );
                expect(metadata, 'SAML metadata should support emailAddress NameIDFormat').to.contain(
                    '<md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat>'
                );
                expect(metadata, 'SAML metadata should contain HTTP-POST SingleSignOnService').to.contain(
                    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleSignOnService>`
                );
                expect(metadata, 'SAML metadata should contain HTTP-Redirect SingleSignOnService').to.contain(
                    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleSignOnService>`
                );
                expect(metadata, 'SAML metadata should contain SOAP SingleSignOnService').to.contain(
                    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleSignOnService>`
                );
                expect(metadata, 'SAML metadata should contain HTTP-Artifact SingleSignOnService').to.contain(
                    `<md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="${config.tenantUrl}/auth/realms/metadata-test/protocol/saml"></md:SingleSignOnService>`
                );
                await generalSettings.removeIdentityProvider();
            });

            test('should verify Download idp metadata XML file', async function () {
                videoName = 'shouldVerifyDownloadIdpMetadataXMLFile';
                await generalSettings.navigateToGeneralSettingsSecurity();
                const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl('metadata-test');
                await generalSettings.setIdentityProviderUI('SAML 2.0', metadataXmlUrl);
                expect(await generalSettings.isValidatorSAMLCreated(), 'SAML IDP validator should be created before downloading metadata').to.be.equal(true);
                await generalSettings.editIdentityProvider();
                const downloadPromise = browserEngine.currentPage.waitForEvent('download');
                await generalSettings.downloadIdpMetadataXMLFile();
                const download = await downloadPromise;
                const fileName = download.suggestedFilename();
                expect(fileName, 'downloaded IDP metadata file should be named metadata.xml').to.equal('metadata.xml');
            });
        });
    }
);
