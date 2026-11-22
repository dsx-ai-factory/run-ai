/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { config, settings, tags, usersManager } from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `General Settings Mappers UI Validations, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`,
    { tag: [tags.services.frontend, tags.runOn.serialCI] },
    () => {
        shouldRunTests(settings.isSelfHosted && !isOCP())(
            function () {
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
                        await browserEngine.stopAndArchive(testInfo, videoName);
                    }
                });

                test('should validate default mapper values successfully', async function () {
                    videoName = 'generalSettingsValidateDefaultMapperValues';
                    await generalSettings.navigateToGeneralSettingsSecurity();
                    const values = await generalSettings.getMapperValues();
                    expect(values, 'default mapper values list should have 7 entries').to.have.length(7);
                    expect(values, 'default mappers should include GID').to.include('GID');
                    expect(values, 'default mappers should include UID').to.include('UID');
                    expect(values, 'default mappers should include GROUPS').to.include('GROUPS');
                    expect(values, 'default mappers should include SUPPLEMENTARYGROUPS').to.include('SUPPLEMENTARYGROUPS');
                    expect(values, 'default mappers should include email').to.include('email');
                });

                test('should add a new mapper and validate the saved values successfully saved', async function () {
                    videoName = 'generalSettingsAddMapper';
                    await generalSettings.navigateToGeneralSettingsSecurity();
                    const mappersSaved = await generalSettings.setMapperValues(
                        'SAML 2.0',
                        'uidtest',
                        'gidtest',
                        'groupstest',
                        'supplementarygroupstest',
                        'emailtest@example.com',
                        'custom_attribute_1',
                        'custom_attribute_2'
                    );
                    expect(mappersSaved, 'mapper save should show "Settings saved" confirmation').eql('Settings saved');
                    const savedValues = await generalSettings.validateMapperValuesSaved();
                    expect(savedValues, 'saved mapper values list should have 7 entries').to.have.length(7);
                    expect(savedValues[0], 'first saved mapper should include the UID test value').to.include('uidtest');
                    expect(savedValues[1], 'second saved mapper should include the GID test value').to.include('gidtest');
                    expect(savedValues[2], 'third saved mapper should include the groups test value').to.include('groupstest');
                    expect(savedValues[3], 'fourth saved mapper should include the supplementaryGroups test value').to.include('supplementarygroupstest');
                    expect(savedValues[4], 'fifth saved mapper should include the email test value').to.include('emailtest@example.com');
                    expect(savedValues[5], 'sixth saved mapper should include custom_attribute_1').to.include('custom_attribute_1');
                    expect(savedValues[6], 'seventh saved mapper should include custom_attribute_2').to.include('custom_attribute_2');
                });
            }
        );
    }
);
