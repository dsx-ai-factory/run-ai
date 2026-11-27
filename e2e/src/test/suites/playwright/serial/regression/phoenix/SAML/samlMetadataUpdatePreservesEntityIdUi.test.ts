/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    idpUtils,
    logger,
    objectGenerator,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import { idpNameSaml } from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { expect, test } from '@playwright/test';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(
    'SAML metadata update preserves entity ID',
    { tag: [tags.teams.phoenix, tags.suites.legacySphinx, tags.tool.ui, tags.resource.ssoConfiguration] },
    () => {
        shouldRunTests(settings.isSelfHosted && !isOCP())(() => {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let realmName: string;
            let customEntityId: string;

            test.beforeAll(async (): Promise<void> => {
                await usersManager.deletePreSeededOidcIdp();

                realmName = objectGenerator.generateNameByPrefix('entityid-preserve');
                customEntityId = objectGenerator.generateNameByPrefix('custom-entity-id');
                await Keycloak.createRealm(realmName);
                const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
                await idpUtils.createIdpSaml(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameSaml,
                    metadataXmlUrl,
                    customEntityId
                );

                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
            });

            test.afterAll(async ({}, testInfo): Promise<void> => {
                try {
                    await browserEngine.stopAndArchive(testInfo, 'shouldPreserveEntityIdOnSamlMetadataUpdate');
                } catch (error) {
                    logger.error(`failed to stop and archive browser engine: ${error}`);
                }
                try {
                    await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameSaml);
                } catch (error) {
                    logger.error(`failed to delete IDP ${idpNameSaml}: ${error}`);
                }
                try {
                    await Keycloak.deleteRealm(realmName);
                } catch (error) {
                    logger.error(`failed to delete realm ${realmName}: ${error}`);
                }
                await usersManager.restorePreSeededOidcIdp();
            });

            test('should preserve the custom SP entity ID when re-uploading IdP metadata on an existing SAML integration', async (): Promise<void> => {
                const idpBeforeUpdate = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameSaml
                );
                expect(
                    idpBeforeUpdate.body.samlData.entityId,
                    'SAML IdP should be created with the configured custom entity ID'
                ).toBe(customEntityId);

                await generalSettings.navigateToGeneralSettingsSecurity();
                const metadataXmlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
                await generalSettings.updateSamlIdentityProviderMetadataUrl(metadataXmlUrl);

                const idpAfterUpdate = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameSaml
                );
                expect(
                    idpAfterUpdate.body.samlData.entityId,
                    'SAML SP entity ID should be preserved after re-uploading IdP metadata on an existing integration, not reset to the generated default'
                ).toBe(customEntityId);
            });
        });
    }
);
