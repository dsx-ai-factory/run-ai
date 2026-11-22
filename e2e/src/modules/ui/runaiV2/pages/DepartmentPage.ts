/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { QuotaManagementSection } from '../components/QuotaManagementSection';
import { QTable } from '../components/QTable';
import { logger, orgUnitUtils } from '../../../helpers';

export class DepartmentPage extends BaseV2Page {
    defaultDeptName = 'default';

    departmentNameCol: string;

    addNewDepartmentBtn: string;

    departmentNameInput: string;

    saveDepartmentBtn: string;

    confirmSaveDepartmentBtn: string;

    deleteBtn: string;

    editButton: string;

    modalDeleteBtn: string;

    nodePoolRow: string;

    nodePoolName: string;

    quotaTabContent: string;

    nodePoolNameSpan: string;

    itemWrapper: string;

    quotaMgmtSect: QuotaManagementSection;

    qTable: QTable;

    filterValueInput: string;

    submitFilterBtn: string;

    constructor(page: Page) {
        super(page);
        this.mainPageTab = '[aid="menu-icon-departments"]';
        this.departmentNameCol = '[aid="name"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Organization}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Departments}"]`;
        this.addNewDepartmentBtn = '[aid="primary-cta-btn"]';
        this.departmentNameInput = '[aid="department-name-input"]';
        this.saveDepartmentBtn = '[aid="wizard-secondary-btn"]';
        this.confirmSaveDepartmentBtn = '[aid="wizard-warning-create-and-close-btn"]';
        this.editButton = '.runai-action-button:has-text("edit")';
        this.deleteBtn = '.runai-action-button:has-text("delete")';
        this.modalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.quotaTabContent = '[aid="quota-tab-content"]';
        this.nodePoolNameSpan = '[aid="node-pool-name"] span';
        this.itemWrapper = '.item-wrapper';
        this.quotaMgmtSect = new QuotaManagementSection(page);
        this.qTable = new QTable(page);

        this.filterValueInput = '[aid="free-text-filter-input"]';
        this.submitFilterBtn = '[aid="submit-filter-btn"]';
    }

    async clickOnAddNewDepartment(): Promise<void> {
        return this.page.locator(this.addNewDepartmentBtn).click();
    }

    async clickSaveDepartment(): Promise<void> {
        await this.page.locator(this.saveDepartmentBtn).click();
        // Check if the warning dialog appears (it might not if user previously checked "Don't show this again")
        const confirmButton = this.page.locator(this.confirmSaveDepartmentBtn);
        try {
            await confirmButton.waitFor({ state: 'visible', timeout: 2000 });
            await confirmButton.click();
        } catch (error) {
            // Dialog didn't appear, which is fine - proceed without clicking
        }
    }

    async toggleDepartment(deptName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.departmentNameCol}`, `${deptName}`);
    }

    async editDepartment(
        deptName: string,
        gpuQuota: string,
        priority?: string,
        overQuota?: boolean,
        maxGpuLimit?: string
    ): Promise<void> {
        await this.toggleDepartment(deptName);
        await this.page.locator(this.editButton).click();
        await this.quotaMgmtSect.setQuota(gpuQuota, priority, overQuota, maxGpuLimit);

        await this.clickSaveDepartment();
    }

    async createDepartment(
        deptName: string,
        gpuQuota: string,
        priority?: string,
        overQuota?: boolean,
        maxGpuLimit?: string,
        nodePoolName?: string
    ): Promise<void> {
        // handling 'default' department gpu quota is "unlimited" situation which blocks creating new department
        const depts = await orgUnitUtils.getDepartments();

        let defaultDeptGpuQuota: number;
        for (const dept of depts) {
            if (dept.name === this.defaultDeptName) {
                defaultDeptGpuQuota = dept.totalResources['gpuQuota'];
                break;
            }
        }

        if (defaultDeptGpuQuota === -1) {
            this.editDepartment(this.defaultDeptName, '1000');
        }
        await this.clickOnAddNewDepartment();
        await this.page.locator(this.departmentNameInput).fill(deptName);
        await this.quotaMgmtSect.setQuota(
            gpuQuota,
            priority,
            overQuota,
            maxGpuLimit,
            nodePoolName || 'default'
        );
        if (nodePoolName) {
            await this.dragNodePoolAboveDefault(nodePoolName);
        }
        await this.clickSaveDepartment();
    }

    async deleteDepartment(deptName: string): Promise<void> {
        await this.toggleDepartment(deptName);
        await this.page.locator(this.deleteBtn).click();
        await this.page.locator(this.modalDeleteBtn).click();
    }

    async addFilterByDepartment(filterValue: string): Promise<void> {
        await this.page.locator(this.addFilterBtn).click();
        await this.page.locator('div[role="menu"] div[role="menuitem"]:has-text("Department")').click();
        await this.page.locator(this.filterValueInput).click();
        await this.page.getByRole('option', { name: filterValue }).click();
        await this.page.locator(this.submitFilterBtn).click();
    }

    async getGpuQuota(deptName: string): Promise<number> {
        const data = await this.qTable.getEntityData('Department', deptName);

        return parseInt(data['GPU quota']);
    }

    async dragNodePoolAboveDefault(nodePoolName: string): Promise<void> {
        logger.info(`Attempting to drag node pool: ${nodePoolName} above default`);
        const allNodePoolNames = await this.page
            .locator(`${this.quotaTabContent} ${this.nodePoolNameSpan}`)
            .allTextContents();

        logger.info(`All node pool names: ${JSON.stringify(allNodePoolNames)}`);

        let sourceIndex = allNodePoolNames.findIndex((name) => name.includes(nodePoolName));

        if (sourceIndex === -1) {
            const prefix = nodePoolName.split('-').slice(0, 2).join('-');
            sourceIndex = allNodePoolNames.findIndex((name) => name.includes(prefix));
        }

        const defaultIndex = allNodePoolNames.findIndex((name) => name === 'default');

        if (sourceIndex === -1) {
            throw new Error(
                `Node pool ${nodePoolName} not found. Available: [${allNodePoolNames.join(', ')}]`
            );
        }
        if (defaultIndex === -1) {
            throw new Error(`Default node pool not found. Available: [${allNodePoolNames.join(', ')}]`);
        }

        if (sourceIndex > defaultIndex) {
            logger.info(`Dragging ${nodePoolName} from index ${sourceIndex} to index ${defaultIndex}`);

            const allItems = this.page.locator(this.itemWrapper);
            const sourceElement = allItems.nth(sourceIndex);
            const targetElement = allItems.nth(defaultIndex);

            const sourceBox = await sourceElement.boundingBox();
            const targetBox = await targetElement.boundingBox();

            if (!sourceBox || !targetBox) {
                throw new Error('Could not get bounding boxes for drag elements');
            }

            await this.page.mouse.move(sourceBox.x + sourceBox.width / 2, sourceBox.y + sourceBox.height / 2);
            await this.page.mouse.down();
            await this.page.waitForTimeout(500);
            await this.page.mouse.move(
                targetBox.x + targetBox.width / 2,
                targetBox.y + targetBox.height / 2,
                { steps: 10 }
            );
            await this.page.waitForTimeout(500);
            await this.page.mouse.up();

            logger.info(`Successfully dragged ${nodePoolName} above default node pool`);
        } else {
            logger.info(`${nodePoolName} is already above default, no drag needed`);
        }
    }
}
