/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { waitUtils } from '../../helpers';
import { DepartmentPage } from '../runaiV2/pages/DepartmentPage';

async function createBasicDepartment(
    departmentPage: DepartmentPage,
    name: string,
    shouldSubmit = true
): Promise<void> {
    await departmentPage.goToPage();
    await departmentPage.clickOnAddNewDepartment();
    await waitUtils.waitForElementToExist(departmentPage, departmentPage.departmentNameInput);
    await departmentPage.fillInputBySelector(departmentPage.departmentNameInput, name);
    if (shouldSubmit) {
        await departmentPage.clickSaveDepartment();
    }
}

export const departmentUiUtils = { createBasicDepartment };
