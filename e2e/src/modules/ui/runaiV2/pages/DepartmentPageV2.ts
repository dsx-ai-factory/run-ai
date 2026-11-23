/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { QTable } from '../components/QTable';
import {
    AccessRulesSubjectType,
    OrgUnitsForm,
    OrgUnitsFormName,
    SchedulingRulesIdleGpuTimeoutWorkloadType,
    SchedulingRulesTimeoutWorkloadType
} from './OrgUnitsForm';
import { RoleDisplayName } from '../../../helpers/config/roleMapping';
import { logger } from '../../../helpers';

export class DepartmentPageV2 extends BaseV2Page {
    newDepartmentBtn: string;

    departmentNameInput: string;

    departmentDescriptionInput: string;

    editDepartmentBtn: string;

    deptNameCol: string;

    accessRuleCol: string;

    nodepoolCol: string;

    orgUnitsForm: OrgUnitsForm;

    mainTable: QTable;

    detailsTable: QTable;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-departments"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Organization}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Departments}"]`;

        this.newDepartmentBtn = '[aid="primary-cta-btn"]';
        this.departmentNameInput = '[aid="department-name-input"]';
        this.departmentDescriptionInput = '[aid="department-description-input"]';
        this.editDepartmentBtn = '[aid="edit-department-btn"]';
        this.deptNameCol = '[aid="name"]';
        this.accessRuleCol = '[aid="accessRules"]';
        this.nodepoolCol = '[aid="node-pools"]';

        this.orgUnitsForm = new OrgUnitsForm(this.page);
        this.mainTable = new QTable(this.page);
        this.detailsTable = new QTable(this.page, '.dialog-content-container table.q-table');
    }

    async toggleDepartment(deptName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.deptNameCol}`, `${deptName}`);
    }

    async openEditDepartmentForm(deptName: string, sectionName: OrgUnitsFormName): Promise<void> {
        await this.toggleDepartment(deptName);
        await this.page.locator(this.editDepartmentBtn).click();
        await this.page.getByRole('menuitem').filter({ hasText: sectionName }).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async clickNewDepartmentBtn(): Promise<void> {
        await this.page.locator(this.newDepartmentBtn).click();
    }

    async fillDepartmentNameInput(deptName: string): Promise<void> {
        await this.locator(this.departmentNameInput).fill(deptName);
    }

    async fillDepartmentDescriptionInput(deptDescription: string): Promise<void> {
        await this.locator(this.departmentDescriptionInput).fill(deptDescription);
    }

    async openNewDepartmentForm(): Promise<void> {
        await this.clickNewDepartmentBtn();
        await this.orgUnitsForm.closeTourModalIfAppear();
    }

    async createDepartment(
        deptName: string,
        deptDescription?: string,
        gpuQuotaConfig?: Array<{ [key: string]: number }>,
        gpuOverQuotaConfig?: Array<{ [key: string]: { allowExceed: boolean; maxGpuAlloc?: number } }>,
        cpuQuotaConfig?: Array<{ [key: string]: { cpus?: number; cpuMemory?: number } }>,
        schedulingRules?: {
            idleGpuTimeout?: {
                workloadType: SchedulingRulesIdleGpuTimeoutWorkloadType;
                days: number;
                hours: number;
                minutes: number;
            };
            workloadTimeoutLimit?: {
                workloadType: SchedulingRulesTimeoutWorkloadType;
                days: number;
                hours: number;
                minutes: number;
            };
            nodeAffinity?: { workloadType: SchedulingRulesTimeoutWorkloadType; nodeAffinityLabel: string };
        },
        accessRules?: { subjectName: AccessRulesSubjectType; userName: string; roleName: RoleDisplayName }
    ): Promise<void> {
        await this.clickNewDepartmentBtn();

        // fill department name and description
        await this.fillDepartmentNameInput(deptName);
        if (deptDescription) {
            await this.fillDepartmentDescriptionInput(deptDescription);
        }
        await this.orgUnitsForm.clickPrimaryBtn();

        // fill gpu quota settings
        if (gpuQuotaConfig) {
            await this.orgUnitsForm.configGpuQuota(gpuQuotaConfig);
        }

        await this.orgUnitsForm.getSaveAndContinueBtn().click();

        // fill gpu over quota settings
        if (gpuOverQuotaConfig) {
            await this.orgUnitsForm.configGpuOverQuota(gpuOverQuotaConfig);
        }

        await this.orgUnitsForm.getSaveAndContinueBtn().click();

        // fill cpu quota settings
        if (cpuQuotaConfig) {
            await this.orgUnitsForm.configCpuQuota(cpuQuotaConfig);
            await this.orgUnitsForm.getSaveAndContinueBtn().click();
        }

        // set scheduling rules config
        if (schedulingRules) {
            if (schedulingRules.idleGpuTimeout) {
                await this.orgUnitsForm.setSchedulingRulesIdleGpuTimeout(
                    schedulingRules.idleGpuTimeout.workloadType,
                    schedulingRules.idleGpuTimeout.days,
                    schedulingRules.idleGpuTimeout.hours,
                    schedulingRules.idleGpuTimeout.minutes
                );
            }

            if (schedulingRules.workloadTimeoutLimit) {
                await this.orgUnitsForm.setSchedulingRulesWorkloadTimeoutLimit(
                    schedulingRules.workloadTimeoutLimit.workloadType,
                    schedulingRules.workloadTimeoutLimit.days,
                    schedulingRules.workloadTimeoutLimit.hours,
                    schedulingRules.workloadTimeoutLimit.minutes
                );
            }

            if (schedulingRules.nodeAffinity) {
                await this.orgUnitsForm.setSchedulingRulesNodeAffinity(
                    schedulingRules.nodeAffinity.workloadType,
                    schedulingRules.nodeAffinity.nodeAffinityLabel
                );
            }

            await this.orgUnitsForm.clickSaveBtn();
        }

        await this.orgUnitsForm.getSaveAndContinueBtn().click();
        await this.orgUnitsForm.getSaveAndContinueBtn().waitFor({ state: 'detached' });

        // set access rules
        if (accessRules) {
            await this.orgUnitsForm.setAccessRules(
                accessRules.subjectName,
                accessRules.userName,
                accessRules.roleName
            );
        }

        await this.orgUnitsForm.getCloseBtn().click();
    }

    async waitForDepartmentToAppear(deptName: string, timeout: number = 120000): Promise<void> {
        await this.page.waitForSelector(`td${this.deptNameCol}:has-text("${deptName}")`, {
            timeout: timeout
        });
    }

    async getDepartmentsFromGrid(): Promise<Array<Record<string, string>>> {
        const data = await this.mainTable.getTableData();
        logger.info(`table data: \n${JSON.stringify(data)}`);

        return data;
    }

    async getDepartmentByNameFromGrid(deptName: string): Promise<Record<string, string>> {
        const data = await this.mainTable.getEntityData('Department', deptName);
        logger.info(`table data: \n${JSON.stringify(data)}`);

        return data;
    }

    async openDepartmentSubjectDetails(deptName: string): Promise<void> {
        await this.page
            .locator('table.q-table tbody tr')
            .filter({ has: this.page.locator(this.deptNameCol, { hasText: deptName }) })
            .locator(`${this.accessRuleCol} button`)
            .click();
    }

    async getDepartmentSubjectInfo(deptName: string): Promise<Array<Record<string, string>>> {
        await this.openDepartmentSubjectDetails(deptName);
        const data = await this.detailsTable.getTableData();

        await this.orgUnitsForm.getCloseBtn().click();
        return data;
    }

    async openDepartmentNodepoolDetails(deptName: string): Promise<void> {
        await this.page
            .locator('table.q-table tbody tr')
            .filter({ has: this.page.locator(this.deptNameCol, { hasText: deptName }) })
            .locator(this.nodepoolCol)
            .click();
    }

    async getDepartmentNodepoolInfo(deptName: string): Promise<Array<Record<string, string>>> {
        await this.openDepartmentNodepoolDetails(deptName);
        const data = await this.detailsTable.getTableData();

        await this.orgUnitsForm.getCloseBtn().click();
        return data;
    }
}
