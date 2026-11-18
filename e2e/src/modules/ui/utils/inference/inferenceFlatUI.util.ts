/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';

import { AuthPage } from '../../runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../runaiV2/pages/WorkloadPage';
import {
    IDistributedInferenceWorkload,
    IFlatWorkload,
    IHuggingFaceFlatWorkload,
    IInferenceMeta,
    INimFlatWorkload,
    IUpdateWorkload
} from '../../../helpers/models/inferenceFlatUI.models';
import { waitUtils, logger } from '../../../helpers';
import { InferenceBaseFlatUIUtil } from './inferenceBaseFlatUI.util';
import { IInferenceType } from '../../../helpers/models/inferenceBaseFlatUI.models';

export class InferenceFlatUIUtil extends InferenceBaseFlatUIUtil {
    auth: AuthPage;

    workload: WorkloadPage;

    constructor(page: Page) {
        super(page);
        this.auth = new AuthPage(page);
        this.workload = new WorkloadPage(page);
    }

    async openAndFillInferenceFirstStep(
        workloadData: IInferenceMeta,
        inferenceType?: IInferenceType
    ): Promise<void> {
        await this.workload.openInferenceForm();

        await this.inference.fillTheFirstStep(
            workloadData.clusterName,
            workloadData.projectName,
            workloadData.workloadName,
            inferenceType,
            workloadData.templateName,
            workloadData.distributed
        );
    }

    async createFlatWorkload(workloadData: IFlatWorkload): Promise<void> {
        await this.openAndFillInferenceFirstStep({
            clusterName: workloadData.clusterName,
            projectName: workloadData.projectName,
            workloadName: workloadData.workloadName,
            templateName: workloadData.templateName
        },
        workloadData.inferenceType);

        await this.fillEnvironmentData(workloadData.environmentSection);
        await this.fillComputeData(workloadData.computeSection);
        await this.fillDataSourceData(workloadData.dataSourceOptions);
        await this.fillGeneralSection(workloadData.generalSectionOptions);

        // create inference
        await this.clickOnSubmitInferenceBtn(workloadData.workloadName);
    }

    async createHuggingFaceFlatWorkload(
        workloadData: IHuggingFaceFlatWorkload,
    ): Promise<void> {
        const meta = {
            clusterName: workloadData.clusterName,
            projectName: workloadData.projectName,
            workloadName: workloadData.workloadName
        };
        await this.openAndFillInferenceFirstStep(
            meta,
            workloadData.inferenceType
        );

        await this.fillHuggingFaceModelData(workloadData.modelSection);
        await this.fillComputeData(workloadData.computeSection);
        await this.fillDataSourceData(workloadData.dataSourceOptions, false);

        if (workloadData.advancedSetup) {
            await this.clickOnAdvancedSetupBtn();
        } else {
            // create inference
            await this.clickOnSubmitInferenceBtn(workloadData.workloadName);
        }
    }

    async createNimFlatWorkload(workloadData: INimFlatWorkload): Promise<void> {
        const meta = {
            clusterName: workloadData.clusterName,
            projectName: workloadData.projectName,
            workloadName: workloadData.workloadName
        };
        await this.openAndFillInferenceFirstStep(meta, IInferenceType.Nim);

        await this.fillNimModelData(workloadData.modelSection);
        await this.fillComputeData(workloadData.computeSection);
        await this.fillDataSourceData(workloadData.dataSourceOptions, false);

        if (workloadData.advancedSetup) {
            await this.clickOnAdvancedSetupBtn();
        } else {
            // create inference
            await this.clickOnSubmitInferenceBtn(workloadData.workloadName);
        }
    }

    async clickOnAdvancedSetupBtn(): Promise<void> {
        await this.inference.clickOnAdvancedSetupBtn();
    }

    async clickOnSubmitInferenceBtn(workloadName: string): Promise<void> {
        await this.inference.page.waitForTimeout(3000); // wait for 3 seconds to ensure the dom is rendered

        // create inference
        await this.inference.clickOnCreateInferenceBtn();

        await waitUtils.waitForElementToExist(this.workload, `td:has-text("${workloadName}")`);
        logger.debug(`inference ${workloadName} created successfully!`);
    }

    async updateWorkload(workloadName: string, workloadData: IUpdateWorkload): Promise<void> {
        await this.workload.toggleWorkload(workloadName);
        await this.workload.clickOnUpdateWorkload();

        // update inference
        await this.fillEnvironmentData(workloadData.environmentSection);
        await this.fillComputeData(workloadData.computeSection);
        await this.fillGeneralSection(workloadData.generalSectionOptions);
        await this.inference.clickOnCreateInferenceBtn();
        await this.inference.page.waitForTimeout(1000); // wait for 1 seconds to ensure the API request is complete
    }

    async createDistributedWorkload(workloadData: IDistributedInferenceWorkload): Promise<void> {
        const meta = {
            clusterName: workloadData.clusterName,
            projectName: workloadData.projectName,
            workloadName: workloadData.workloadName,
            distributed: true
        };
        await this.openAndFillInferenceFirstStep(meta);

        // leader
        await this.fillEnvironmentData(workloadData.leader.environmentSection);
        await this.fillComputeData(workloadData.leader.computeSection);
        await this.fillDataSourceData(workloadData.leader.dataSourceOptions);
        await this.fillGeneralSection(workloadData.leader.generalSectionOptions);
        await this.inference.clickOnCreateInferenceBtn(); // next step button

        await this.waitForDistributedWorkerFormVisible();
        // worker
        if (workloadData.worker) {
            await this.fillEnvironmentData(workloadData.worker.environmentSection);
            await this.fillComputeData(workloadData.worker.computeSection);
            await this.fillDataSourceData(workloadData.worker.dataSourceOptions);
            await this.fillGeneralSection(workloadData.worker.generalSectionOptions);
        }

        await this.inference.clickOnCreateInferenceBtn(); // next step button
        await this.waitForDistributedWorkloadSetupFormVisible();

        // workload setup
        if (workloadData.workloadSetup) {
            await this.fillComputeData(workloadData.workloadSetup.computeSection);
            await this.fillGeneralSection(workloadData.workloadSetup.generalSectionOptions);
        }

        // create distributed inference
        await this.inference.clickOnCreateInferenceBtn();
    }

    
    async waitForDistributedWorkloadSetupFormVisible(): Promise<void> {
        logger.debug('Waiting for workload setup form to be visible');
        await this.inference.page.waitForSelector('.inference-distributed-workload-setup', {
            state: 'visible'
        });
        logger.debug('Workload setup form is visible');
    }

    async waitForDistributedWorkerFormVisible(): Promise<void> {
        logger.debug('Waiting for worker form to be visible');
        await this.inference.page.waitForSelector('.inference-distributed-worker-form', { state: 'visible' });
        logger.debug('Worker form is visible');
    }

    async waitForDistributedLeaderFormVisible(): Promise<void> {
        logger.debug('Waiting for leader form to be visible');
        await this.inference.page.waitForSelector('.inference-distributed-leader-form', { state: 'visible' });
        logger.debug('Leader form is visible');
    }
}
