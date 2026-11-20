/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { apiUtils, config, logger, tags } from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../../../../../../../modules/helpers/models/featureFlags';
import { supportedFrom, SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`General Settings Toggles UI, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.tempBackendSettings] }, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let generalSettings: GeneralSettings;
    let videoName: string;
    let originalDataVolumesValue: boolean;
    let originalLimitQuotaValue: boolean;
    let originalHybridSubmissionValue: boolean;
    let originalModelCatalogValue: boolean;

    test.beforeEach('init browsers and classes', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        generalSettings = new GeneralSettings(browserEngine.currentPage);
    });

    test.afterEach('stop browserEngine and restore toggles', async ({}, testInfo) => {
        const restores: Array<{ key: EFeatureFlagsKeys; name: EFeatureFlagsNames; value: boolean; clear: () => void }> =
            [];

        if (originalDataVolumesValue !== undefined) {
            restores.push({
                key: EFeatureFlagsKeys.ENABLE_DATA_VOLUMES,
                name: EFeatureFlagsNames.ENABLE_DATA_VOLUMES,
                value: originalDataVolumesValue,
                clear: () => { originalDataVolumesValue = undefined; }
            });
        }
        if (originalLimitQuotaValue !== undefined) {
            restores.push({
                key: EFeatureFlagsKeys.ENABLE_LIMIT_OQ_SUBSCRIPTION,
                name: EFeatureFlagsNames.ENABLE_LIMIT_OQ_SUBSCRIPTION,
                value: originalLimitQuotaValue,
                clear: () => { originalLimitQuotaValue = undefined; }
            });
        }
        if (originalHybridSubmissionValue !== undefined) {
            restores.push({
                key: EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
                name: EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION,
                value: originalHybridSubmissionValue,
                clear: () => { originalHybridSubmissionValue = undefined; }
            });
        }
        if (originalModelCatalogValue !== undefined) {
            restores.push({
                key: EFeatureFlagsKeys.MODEL_CATALOG,
                name: EFeatureFlagsNames.MODEL_CATALOG,
                value: originalModelCatalogValue,
                clear: () => { originalModelCatalogValue = undefined; }
            });
        }

        const failedRestores: string[] = [];
        for (const restore of restores) {
            try {
                await apiUtils.setFeatureValue(restore.key, restore.name, restore.value);
                restore.clear();
            } catch (error) {
                logger.error(`Failed to restore setting ${restore.key} to ${restore.value}: ${error}`);
                failedRestores.push(`${restore.key} → ${restore.value}`);
            }
        }

        await browserEngine.stopAndArchive(testInfo, videoName);
        if (failedRestores.length > 0) {
            throw new Error(`Teardown: failed to restore ${failedRestores.length} setting(s): ${failedRestores.join(', ')}`);
        }
    });


    supportedFrom(SupportedVersions.V2_23)(function () {
        test('should display all workload setting toggles in the workloads section', async function () {
            videoName = 'shouldDisplayAllWorkloadSettingToggles';
            const isHybridSubmissionEnabled = (await apiUtils.getFeatureValue(
                EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
                EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION
            )) as boolean;
            const isModelCatalogEnabled = (await apiUtils.getFeatureValue(
                EFeatureFlagsKeys.MODEL_CATALOG,
                EFeatureFlagsNames.MODEL_CATALOG
            )) as boolean;

            await generalSettings.navigateToGeneralSettingsWorkloads();
            const alwaysVisible = [
                generalSettings.flexWorkloadSubmissionToggle,
                generalSettings.newPolicyToggle,
                generalSettings.modelCatalogToggle,
                generalSettings.nimToggle,
                generalSettings.yamlSubmissionToggle,
                generalSettings.ngcIntegrationToggle
            ];
            for (const toggle of alwaysVisible) {
                expect(
                    await generalSettings.page.locator(toggle).isVisible(),
                    `Toggle ${toggle} should be visible`
                ).to.be.true;
            }

            if (isHybridSubmissionEnabled) {
                for (const toggle of [generalSettings.dataVolumeToggle, generalSettings.flexWorkloadTemplateToggle]) {
                    expect(
                        await generalSettings.page.locator(toggle).isVisible(),
                        `Toggle ${toggle} should be visible when hybrid submission is enabled`
                    ).to.be.true;
                }
            } else {
                logger.info('Hybrid submission is disabled — skipping data volumes and templates toggle visibility check');
            }

            if (isModelCatalogEnabled) {
                expect(
                    await generalSettings.page.locator(generalSettings.allowGpt2ModelToggle).isVisible(),
                    'GPT2 toggle should be visible when model catalog is enabled'
                ).to.be.true;
            } else {
                logger.info('Model catalog is disabled — skipping GPT2 toggle visibility check');
            }
        });
    });

    test('should display all resource setting toggles in the resources section', async function () {
        videoName = 'shouldDisplayAllResourceSectionToggles';
        await generalSettings.openGeneralSettingsPage();
        await generalSettings.page.locator(generalSettings.resourcesSection).click({ timeout: 10000 });

        const resourceToggles = [
            generalSettings.limtQuotaOverSubscriptionToggle,
            generalSettings.cpuQuotaToggle,
            generalSettings.overQuotaWeightToggle,
            generalSettings.gpuResourceOptimizationToggle
        ];

        for (const toggle of resourceToggles) {
            const isVisible = await generalSettings.page.locator(toggle).isVisible();
            expect(isVisible, `Toggle ${toggle} should be visible`).to.be.true;
        }
    });

    test('should display Experimental badge next to GPT2 model toggle when it is an alpha feature', async function () {
        videoName = 'shouldDisplayExperimentalBadgeForGPT2Model';
        const isModelCatalogEnabled = (await apiUtils.getFeatureValue(
            EFeatureFlagsKeys.MODEL_CATALOG,
            EFeatureFlagsNames.MODEL_CATALOG
        )) as boolean;

        if (!isModelCatalogEnabled) {
            logger.info('Model catalog is disabled in this environment, skipping GPT2 badge UI check');
            return;
        }

        await generalSettings.navigateToGeneralSettingsWorkloads();
        expect(
            await generalSettings.page.locator(generalSettings.allowGpt2ModelExperimentalBadge).isVisible(),
            'Experimental badge should be visible for GPT2 model setting'
        ).to.be.true;
    });

    test('should display Beta badge next to NGC catalog toggle when it is a beta feature', async function () {
        videoName = 'shouldDisplayBetaBadgeForNGCCatalog';
        await generalSettings.navigateToGeneralSettingsWorkloads();

        const isNgcToggleVisible = await generalSettings.page
            .locator(generalSettings.ngcIntegrationToggle)
            .isVisible();
        if (!isNgcToggleVisible) {
            logger.info('NGC catalog toggle is not visible in this environment, skipping badge check');
            return;
        }

        expect(
            await generalSettings.page.locator(generalSettings.ngcIntegrationBetaBadge).isVisible(),
            'Beta badge should be visible for NGC catalog setting'
        ).to.be.true;
    });

    test('should return correct alpha and beta stage values for settings in the API response', async function () {
        videoName = 'shouldIncludeStageFieldForAlphaBetaSettings';
        const settings = await apiUtils.getTenantSettings();
        const stagedSettings = settings.filter((s) => s.stage);
        expect(stagedSettings.length, 'API should return at least some staged settings').to.be.greaterThan(0);
        const validStages = ['Alpha', 'Beta', 'Legacy'];
        for (const setting of stagedSettings) {
            expect(
                validStages,
                `Stage "${setting.stage}" for setting "${setting.key}" should be a known value`
            ).to.include(setting.stage);
        }

        const gpt2Setting = settings.find((s) => s.key === EFeatureFlagsKeys.ALLOW_GPT2_MODEL);
        expect(gpt2Setting, 'GPT2 setting should be present in API response').to.exist;
        expect(gpt2Setting?.stage, 'GPT2 setting should have stage Alpha').to.equal('Alpha');
        const ngcSetting = settings.find((s) => s.key === EFeatureFlagsKeys.ENABLE_NGC_INTEGRATION);
        expect(ngcSetting, 'NGC catalog setting should be present in API response').to.exist;
        expect(ngcSetting?.stage, 'NGC catalog setting should have stage Beta').to.equal('Beta');
    });


    test('should hide data volumes and flexible templates toggles when flexible workload submission is disabled', async function () {
        videoName = 'shouldHideDataVolumesAndTemplatesWhenHybridSubmissionDisabled';
        originalHybridSubmissionValue = (await apiUtils.getFeatureValue(
            EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
            EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION
        )) as boolean;

        await apiUtils.setFeatureValue(
            EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
            EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION,
            false
        );
        await generalSettings.hardNavigateToGeneralSettings();
        await generalSettings.page.locator(generalSettings.workloadsSection).click({ timeout: 10000 });
        expect(
            await generalSettings.page.locator(generalSettings.dataVolumeToggle).isVisible(),
            'Data volumes toggle should be hidden when hybrid submission is disabled'
        ).to.be.false;
        expect(
            await generalSettings.page.locator(generalSettings.flexWorkloadTemplateToggle).isVisible(),
            'Templates toggle should be hidden when hybrid submission is disabled'
        ).to.be.false;
    });

    test('should show data volumes and flexible templates toggles when flexible workload submission is enabled', async function () {
        videoName = 'shouldShowDataVolumesAndTemplatesWhenHybridSubmissionEnabled';
        originalHybridSubmissionValue = (await apiUtils.getFeatureValue(
            EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
            EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION
        )) as boolean;
        await apiUtils.setFeatureValue(
            EFeatureFlagsKeys.WORKLOAD_HYBRID_SUBMISSION,
            EFeatureFlagsNames.WORKLOAD_HYBRID_SUBMISSION,
            true
        );
        await generalSettings.hardNavigateToGeneralSettings();
        await generalSettings.page.locator(generalSettings.workloadsSection).click({ timeout: 10000 });
        expect(
            await generalSettings.page.locator(generalSettings.dataVolumeToggle).isVisible(),
            'Data volumes toggle should be visible when hybrid submission is enabled'
        ).to.be.true;
        expect(
            await generalSettings.page.locator(generalSettings.flexWorkloadTemplateToggle).isVisible(),
            'Templates toggle should be visible when hybrid submission is enabled'
        ).to.be.true;
    });

    test('should show GPT2 model toggle only when model catalog is enabled', async function () {
        videoName = 'shouldShowGPT2WhenModelCatalogEnabled';
        originalModelCatalogValue = (await apiUtils.getFeatureValue(
            EFeatureFlagsKeys.MODEL_CATALOG,
            EFeatureFlagsNames.MODEL_CATALOG
        )) as boolean;
        await apiUtils.setFeatureValue(EFeatureFlagsKeys.MODEL_CATALOG, EFeatureFlagsNames.MODEL_CATALOG, true);
        await generalSettings.hardNavigateToGeneralSettings();
        await generalSettings.page.locator(generalSettings.workloadsSection).click({ timeout: 10000 });
        expect(
            await generalSettings.page.locator(generalSettings.allowGpt2ModelToggle).isVisible(),
            'GPT2 toggle should be visible when model catalog is enabled'
        ).to.be.true;
    });

    test('should toggle data volumes on and off and reflect the change immediately in the UI', async function () {
        videoName = 'shouldUpdateDataVolumesSettingToggleAndRestoreIt';
        await generalSettings.navigateToGeneralSettingsWorkloads();
        const toggle = generalSettings.page.locator(generalSettings.dataVolumeToggle);
        originalDataVolumesValue = await toggle.isChecked();
        await toggle.click();
        await generalSettings.page.waitForTimeout(500);
        const updatedValue = await toggle.isChecked();
        expect(updatedValue, 'data volumes toggle state should be inverted after clicking').to.equal(!originalDataVolumesValue);
    });

    test('should toggle limit projects from exceeding department quota on and off and reflect the change immediately in the UI', async function () {
        videoName = 'shouldUpdateLimitQuotaOverSubscriptionToggleAndRestoreIt';
        await generalSettings.openGeneralSettingsPage();
        await generalSettings.page.locator(generalSettings.resourcesSection).click({ timeout: 10000 });
        const toggle = generalSettings.page.locator(generalSettings.limtQuotaOverSubscriptionToggle);
        originalLimitQuotaValue = await toggle.isChecked();
        await toggle.click();
        await generalSettings.page.waitForTimeout(500);
        const updatedValue = await toggle.isChecked();
        expect(updatedValue, 'limit quota toggle state should be inverted after clicking').to.equal(!originalLimitQuotaValue);
    });
});
