/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { pageFilterUtils } from './pageFilterUtils';
import { flowRootState, waitUtils } from '../../helpers';
import { ProjectPage } from '../runaiV2/pages/ProjectPage';

/**
 * Creates a basic project with only the name filled in.
 * Creates a "default" project by only filling the first step and closing.
 */
async function createBasicProject(
    projectPage: ProjectPage,
    projectName: string,
    shouldSubmit = true
): Promise<void> {
    await projectPage.goToPage();
    await pageFilterUtils.selectCluster(projectPage.page, flowRootState.cluster.name);
    await projectPage.clickOnAddNewProjectBtn();
    await waitUtils.waitForElementToExist(projectPage, projectPage.projectNameInput);
    await projectPage.fillInputBySelector(projectPage.projectNameInput, projectName);

    if (shouldSubmit) {
        await projectPage.clickOnCreateAndCloseFromWizard();
        await waitUtils.waitForElementToExist(projectPage, projectPage.warningCreateAndCloseBtn);
        await projectPage.clickOnWarningCreateAndClose();
    }
}

export const projectUiUtils = {
    createBasicProject
};