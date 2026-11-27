/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from '@playwright/test';
import { expect } from 'chai';
import { EnvironmentVariable } from '../../../api/apis/assetsApi';
import { HybridForm } from '../../runaiV2/pages/HybridSubmissionForm';
import { logger } from '../../../helpers';

export class HuggingFaceElementValidationUtils {
    constructor(private page: Page) {}

    async expectEnvironmentVariables(environmentVariables: Array<EnvironmentVariable>): Promise<void> {
        const hybridForm = new HybridForm(this.page);

        if (!(await hybridForm.isSubSectionOpenedByLabel('Runtime settings'))) {
            logger.debug('Runtime settings section is not opened, opening it');
            await hybridForm.clickOnSubSectionByLabel('Runtime settings');
        } else {
            logger.debug('Runtime settings section is already opened');
        }

        const environmentVariableBoxes = this.page.locator('.environment-variable-box');
        const envVarNameInput: string = hybridForm.hybridEnvironmentSection.envVarNameInput;
        const envVarValueInput: string = hybridForm.hybridEnvironmentSection.envVarValue;

        const count = await environmentVariableBoxes.count();
        const expectedEnvVarCount = environmentVariables.length;
        expect(count).to.equal(expectedEnvVarCount);

        for (let i = 0; i < expectedEnvVarCount; i++) {
            const box = environmentVariableBoxes.nth(i);
            const nameInput = box.locator(envVarNameInput);
            const valueInput = box.locator(envVarValueInput);
            const envVarName = await nameInput.inputValue();
            const envVarValue = await valueInput.inputValue();
            expect(envVarName).to.equal(environmentVariables[i].name);
            expect(envVarValue).to.equal(environmentVariables[i].value);
        }
    }
}
