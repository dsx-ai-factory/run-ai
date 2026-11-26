/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';

import { AuthPage } from '../../runaiV2/pages/AuthPage';
import { InferenceBaseFlatUIUtil } from './inferenceBaseFlatUI.util';
import { WorkloadTemplatePage } from '../../runaiV2/pages/WorkloadTemplatePage';
import { AssetsScopeSection } from '../../runaiV2/components/AssetsScopeSection';
import { Scope } from '../../../api/apis/assetsApi';
import {
    IHuggingFaceTemplate,
    IInferenceTemplate,
    IInferenceTemplateMeta,
    INimTemplate
} from '../../../helpers/models/inferenceTemplateUI.models';
import { logger, waitUtils } from '../../../helpers';
import { IInferenceType } from '../../../helpers/models/inferenceBaseFlatUI.models';

export class InferenceTemplateUIUtil extends InferenceBaseFlatUIUtil {
    auth: AuthPage;

    workloadTemplate: WorkloadTemplatePage;

    assetsScopeSection: AssetsScopeSection;

    constructor(page: Page) {
        super(page);
        this.auth = new AuthPage(page);
        this.workloadTemplate = new WorkloadTemplatePage(page);
        this.assetsScopeSection = new AssetsScopeSection(page);
    }

    async openAndFillFirstStep(templateMeta: IInferenceTemplateMeta): Promise<void> {
        await this.workloadTemplate.openInferenceForm();

        await this.assetsScopeSection.selectScope({
            scope: templateMeta.clusterName ? Scope.Cluster : Scope.Project,
            scopeEntityName: templateMeta.clusterName ?? templateMeta.projectName
        });

        await this.inference.selectInferenceTypeOption(templateMeta.inferenceType);

        if (templateMeta.templateName) {
            await this.workloadTemplate.fillTemplateName(templateMeta.templateName);
        }
        await this.workloadTemplate.clickOnContinue();
    }

    async createCustomTemplate(templateData: IInferenceTemplate): Promise<void> {
        await this.openAndFillFirstStep(templateData.templateMeta);

        await this.fillEnvironmentData(templateData.environmentSection);
        await this.fillComputeData(templateData.computeSection);
        await this.fillDataSourceData(templateData.dataSourceOptions);
        await this.fillGeneralSection(templateData.generalSectionOptions);

        // create template
        await this.clickOnCreateCustomTemplateBtn(templateData.templateMeta.templateName);
    }

    async clickOnCreateCustomTemplateBtn(templateName: string): Promise<void> {
        await this.workloadTemplate.clickOnCreateInferenceTemplateBtn();

        await waitUtils.waitForElementToExist(this.workloadTemplate, `td:has-text("${templateName}")`);
        logger.debug(`inference ${templateName} created successfully!`);
    }

    async createNimTemplate(templateData: INimTemplate): Promise<void> {
        await this.openAndFillFirstStep(templateData.templateMeta);

        await this.fillNimModelData(templateData.modelSection);
        await this.fillComputeData(templateData.computeSection);
        await this.fillDataSourceData(templateData.dataSourceOptions, false);

        // create template
        if (templateData.advancedSetup) {
            await this.clickOnAdvancedSetupBtn();
        } else {
            await this.clickOnCreateCustomTemplateBtn(templateData.templateMeta.templateName);
        }
    }

    async createHFTemplate(templateData: IHuggingFaceTemplate): Promise<void> {
        await this.openAndFillFirstStep(templateData.templateMeta);

        await this.fillHuggingFaceModelData(templateData.modelSection);
        await this.fillComputeData(templateData.computeSection);
        await this.fillDataSourceData(templateData.dataSourceOptions, false);

        // create template
        if (templateData.advancedSetup) {
            await this.clickOnAdvancedSetupBtn();
        } else {
            await this.clickOnCreateCustomTemplateBtn(templateData.templateMeta.templateName);
        }
    }

    async clickOnAdvancedSetupBtn(): Promise<void> {
        await this.workloadTemplate.clickOnAdvancedSetupBtn();
    }
}
