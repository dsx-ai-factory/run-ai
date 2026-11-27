/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { HybridForm } from './HybridSubmissionForm';

export function getWorkloadTourKeyTextPerStep(templateName: string): Record<string, string> {
    return {
        newWorkload: 'Click + NEW WORKLOAD to begin',
        selectWorkspace: 'Select Workspace',
        selectProject: 'Select a project',
        selectTemplate: `${templateName} template`,
        nameAndCreateWorkspace: 'Give your workspace a name and Click CREATE WORKSPACE',
        selectCreatedWorkspace: 'Select the workload you just created',
        connect: 'Connect to start using your Jupyter'
    };
}

export class GuidedTour extends HybridForm {
    guidedTourPopover: string;

    activeTourElement: string;

    guidedTourModal: string;

    closeCompletionModalButton: string;

    leaveGuidedTourModalConfirmButton: string;

    guidedTourExitModal: string;

    closeExitModalButton: string;

    goBackFromExitModalButton: string;

    helpMenuGuidedTourMenu: string;

    guidedStepsOptionJupyterNotebook: string;

    guidedStepsOptionOverviewDashboard: string;

    guidedTourPopoverCloseButton: string;

    guidedTourDescriptionDiv: string;

    guidedTourNextButton: string;

    private readonly ARIA_DISABLED_ATTR = 'aria-disabled';

    constructor(page: Page) {
        super(page);
        this.guidedTourPopover = '.guided-tour-popover';
        this.guidedTourModal = '.guided-tour-modal';
        this.activeTourElement = '.guided-tour-active-element';
        this.closeCompletionModalButton = '[aid="guided-tour-completion-leave-btn"]';
        this.leaveGuidedTourModalConfirmButton = '[aid="guided-tour-leave-confirm-btn"]';
        this.guidedTourExitModal = '[aid="guided-tour-leave-confirm-modal"]';
        this.closeExitModalButton = '[aid="guided-tour-leave-confirm-btn"]'; // Same as confirm button (method unused)
        this.goBackFromExitModalButton = '[aid="guided-tour-leave-cancel-btn"]';
        this.helpMenuGuidedTourMenu = '[aid="guided-steps-menu"]';
        this.guidedTourPopoverCloseButton = '.guided-tour-close-btn';
        this.guidedStepsOptionJupyterNotebook = '[aid="guided-steps-option-jupyter-notebook"]';
        this.guidedStepsOptionOverviewDashboard = '[aid="guided-steps-option-overview-dashboard"]';
        this.guidedTourDescriptionDiv = '.guided-tour-description';
        this.guidedTourNextButton = '.guided-tour-next-btn';
    }

    async clickOnGuidedTourButtonInTheHelpMenu(): Promise<void> {
        await this.clickOnHelpMenuButton();
        await this.page.waitForTimeout(2000);
        await this.page.locator(this.helpMenuGuidedTourMenu).click();
    }

    async clickOnGuidedStepsOptionJupyterNotebook(): Promise<void> {
        await this.page.locator(this.guidedStepsOptionJupyterNotebook).click();
    }

    async clickOnGuidedStepsOptionOverviewDashboard(): Promise<void> {
        await this.page.locator(this.guidedStepsOptionOverviewDashboard).click();
    }

    async waitForGuidedTourPopover(text: string, timeout = 5000): Promise<void> {
        await this.waitForVisible(this.locator(`${this.guidedTourPopover}:has-text("${text}")`), timeout);
    }

    async clickOnFirstWorkloadCheckbox(): Promise<void> {
        await this.locator(`.q-checkbox${this.activeTourElement}`).first().click();
    }

    async waitForCompletionModal(): Promise<void> {
        await this.waitForVisible(this.locator(this.guidedTourModal), 10000);
    }

    async closeCompletionModal(): Promise<void> {
        await this.locator(this.closeCompletionModalButton).click();
    }

    async confirmButtonLeaveGuidedTourModal(): Promise<void> {
        await this.locator(this.leaveGuidedTourModalConfirmButton).click();
    }

    async performGuidedTourStepLikeAHuman(
        tourMessage: string,
        action: () => Promise<void> | void
    ): Promise<void> {
        await this.waitForGuidedTourPopover(tourMessage);
        await this.page.waitForTimeout(1500);
        await action();
    }

    async openExitTourModal(): Promise<void> {
        await this.locator(this.guidedTourPopoverCloseButton).click();
        await this.waitForVisible(this.locator(this.guidedTourExitModal), 3000);
    }

    async closeExitTourModal(): Promise<void> {
        await this.locator(this.closeExitModalButton).click();
    }

    async confirmExitTourModal(): Promise<void> {
        await this.locator(this.leaveGuidedTourModalConfirmButton).click();
    }

    async goBackFromExitTourModal(): Promise<void> {
        await this.locator(this.goBackFromExitModalButton).click();
    }

    async verifyGuidedTourButtonState(expectedDisabled: boolean): Promise<void> {
        await this.clickOnGuidedTourButtonInTheHelpMenu();

        const guidedTourButton = this.locator(this.guidedStepsOptionJupyterNotebook);

        const parentItemToCheck = guidedTourButton.locator('../..');
        const isDisabled = (await parentItemToCheck.getAttribute(this.ARIA_DISABLED_ATTR)) === 'true';

        if (isDisabled !== expectedDisabled) {
            const expectedState = expectedDisabled ? 'disabled' : 'enabled';
            const actualState = isDisabled ? 'disabled' : 'enabled';
            throw new Error(`Guided tour button should be ${expectedState} but was ${actualState}`);
        }
    }

    async openOverviewDashboardHelpTourGide(): Promise<void> {
        await this.page.locator(this.helpMenuButton).click();
        await this.page.locator(this.helpGuidedStepsMenu).click();
        await this.page.locator(this.helpGuidedStepsOverviewDashboardMenu).click();
    }

    async getModalText(): Promise<string> {
        return (await this.page.locator(this.guidedTourDescriptionDiv).textContent()).trim();
    }

    async clickModalNextButton(): Promise<void> {
        await this.page.locator(this.guidedTourNextButton).click();
    }
}
