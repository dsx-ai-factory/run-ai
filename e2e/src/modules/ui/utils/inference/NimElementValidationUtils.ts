/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from '@playwright/test';
import { expect } from 'chai';
import { EnvironmentVariable } from '../../../api/apis/assetsApi';
import { HybridForm } from '../../runaiV2/pages/HybridSubmissionForm';
import { logger } from '../../../helpers';

export class NimElementValidationUtils {
    constructor(private page: Page) {}

    async expectEnvironmentVariables(environmentVariables: Array<EnvironmentVariable>): Promise<void> {
        const hybridForm = new HybridForm(this.page);
        if (!(await hybridForm.isSubSectionOpenedByLabel('Runtime settings'))) {
            logger.debug('Runtime settings section is not opened, opening it');
            await hybridForm.clickOnSubSectionByLabel('Runtime settings');
        } else {
            logger.debug('Runtime settings section is already opened');
        }
        const envVarNameInput: string = hybridForm.hybridEnvironmentSection.envVarNameInput;
        const envVarValueInput: string = hybridForm.hybridEnvironmentSection.envVarValue;
        const environmentVariableBoxes = this.page.locator('.environment-variable-box');
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

    async expectEnvironmentVariablesWithSharedCredentials(
        environmentVariables: Array<EnvironmentVariable>
    ): Promise<void> {
        const hybridForm = new HybridForm(this.page);
        if (!(await hybridForm.isSubSectionOpenedByLabel('Runtime settings'))) {
            logger.debug('Runtime settings section is not opened, opening it');
            await hybridForm.clickOnSubSectionByLabel('Runtime settings');
        } else {
            logger.debug('Runtime settings section is already opened');
        }
        const envVarNameInput: string = hybridForm.hybridEnvironmentSection.envVarNameInput;
        const envVarValueInput: string = hybridForm.hybridEnvironmentSection.envVarValue;
        const envVarSourceSelect: string = hybridForm.hybridEnvironmentSection.envVarSourceSelect;
        const envVarSecretSelect: string = hybridForm.hybridEnvironmentSection.envVarSecretSelect;
        const envVarKeyInput: string = hybridForm.hybridEnvironmentSection.envVarCustomKey;
        const environmentVariableBoxes = this.page.locator('.environment-variable-box');
        const count = await environmentVariableBoxes.count();
        const expectedEnvVarCount = environmentVariables.length;
        expect(count).to.equal(expectedEnvVarCount);

        for (let i = 0; i < expectedEnvVarCount; i++) {
            const box = environmentVariableBoxes.nth(i);
            const nameInput = box.locator(envVarNameInput);
            const currEnvVar = environmentVariables[i];
            const envVarName = await nameInput.inputValue();
            expect(envVarName).to.equal(currEnvVar.name);
            if (currEnvVar.secret) {
                const sourceInput = box.locator(envVarSourceSelect);
                const sourceName = await sourceInput.textContent();
                expect(sourceName).to.equal('Shared secret');

                const secretInput = box.locator(envVarSecretSelect);
                const secretName = await secretInput.inputValue();
                expect(secretName).to.equal(currEnvVar.secret.name);

                const keyInput = box.locator(envVarKeyInput);
                const key = await keyInput.inputValue();
                expect(key).to.equal(currEnvVar.secret.key);
            } else {
                const valueInput = box.locator(envVarValueInput);
                const envVarValue = await valueInput.inputValue();
                expect(envVarValue).to.equal(currEnvVar.value);
            }
        }
    }

    async expectEnvironmentVariablesWithPrivateCredentials(
        environmentVariables: Array<EnvironmentVariable>
    ): Promise<void> {
        const hybridForm = new HybridForm(this.page);
        if (!(await hybridForm.isSubSectionOpenedByLabel('Runtime settings'))) {
            logger.debug('Runtime settings section is not opened, opening it');
            await hybridForm.clickOnSubSectionByLabel('Runtime settings');
        } else {
            logger.debug('Runtime settings section is already opened');
        }
        const envVarNameInput: string = hybridForm.hybridEnvironmentSection.envVarNameInput;
        const envVarValueInput: string = hybridForm.hybridEnvironmentSection.envVarValue;
        const envVarSourceSelect: string = hybridForm.hybridEnvironmentSection.envVarSourceSelect;
        const envVarPersonalSecretSelect: string =
            hybridForm.hybridEnvironmentSection.envVarPersonalSecretSelect;
        const envVarPersonalSecretKeyInput: string =
            hybridForm.hybridEnvironmentSection.envVarPersonalSecretKeyInput;
        const environmentVariableBoxes = this.page.locator('.environment-variable-box');
        const count = await environmentVariableBoxes.count();
        const expectedEnvVarCount = environmentVariables.length;
        expect(count).to.equal(expectedEnvVarCount);

        for (let i = 0; i < expectedEnvVarCount; i++) {
            const box = environmentVariableBoxes.nth(i);
            const nameInput = box.locator(envVarNameInput);
            const currEnvVar = environmentVariables[i];
            const envVarName = await nameInput.inputValue();
            expect(envVarName).to.equal(currEnvVar.name);
            if (currEnvVar.userCredential) {
                const sourceInput = box.locator(envVarSourceSelect);
                const sourceName = await sourceInput.textContent();
                expect(sourceName).to.equal('My credentials');

                const secretInput = box.locator(envVarPersonalSecretSelect);
                const secretName = await secretInput.inputValue();
                expect(secretName).to.equal(currEnvVar.userCredential.name);

                const keyInput = box.locator(envVarPersonalSecretKeyInput);
                const key = await keyInput.inputValue();
                expect(key).to.equal(currEnvVar.userCredential.key);
            } else {
                const valueInput = box.locator(envVarValueInput);
                const envVarValue = await valueInput.inputValue();
                expect(envVarValue).to.equal(currEnvVar.value);
            }
        }
    }
}
