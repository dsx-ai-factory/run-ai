/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { QuotaManagementSection } from '../components/QuotaManagementSection';
import { QTable } from '../components/QTable';
import {
    AccessRulesSubjectType,
    OrgUnitsForm,
    OrgUnitsFormName,
    SchedulingRulesIdleGpuTimeoutWorkloadType,
    SchedulingRulesTimeoutWorkloadType
} from './OrgUnitsForm';
import { RoleDisplayName } from '../../../helpers/config/roleMapping';
import { logger, orgUnitUtils } from '../../../helpers';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { Scope } from '../../../api/apis/assetsApi';

export class ProjectPageV2 extends BaseV2Page {
    newProjectBtn: string;

    projectNameInput: string;

    projectDescriptionInput: string;

    projectNameCol: string;

    statusCol: string;

    editProjectBtn: string;

    deleteProjectBtn: string;

    deleteProjectModalBtn: string;

    accessRuleCol: string;

    scopeSearchInputPlaceholder: string;

    orgUnitsForm: OrgUnitsForm;

    assetsScopeSection: AssetsScopeSection;

    qTable: QTable;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-projects"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Organization}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Projects}"]`;

        this.newProjectBtn = '[aid="primary-cta-btn"]';
        this.projectNameInput = '[aid="project-name-input"]';
        this.projectDescriptionInput = '[aid="project-description-input"]';
        this.projectNameCol = '[aid="name"]';
        this.statusCol = '[aid="phase"]';
        this.editProjectBtn = '[aid="edit-project-btn"]';
        this.deleteProjectBtn = '[aid="delete-project-btn"]';
        this.deleteProjectModalBtn = '[aid="delete-modal-delete-button"]';
        this.accessRuleCol = '[aid="accessRules"]';
        this.scopeSearchInputPlaceholder = '[placeholder="Search department"]';

        this.orgUnitsForm = new OrgUnitsForm(this.page);
        this.assetsScopeSection = new AssetsScopeSection(page);
        this.qTable = new QTable(this.page);
    }

    async toggleProject(projectName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.projectNameCol}`, `${projectName}`);
    }

    async openEditProjectForm(projectName: string, sectionName: OrgUnitsFormName): Promise<void> {
        await this.toggleProject(projectName);
        await this.page.locator(this.editProjectBtn).click();
        await this.page.getByRole('menuitem').filter({ hasText: sectionName }).click();
        await this.page.waitForLoadState('networkidle');
    }

    async editProjectDescription(projectName: string, newDescription: string): Promise<void> {
        await this.openEditProjectForm(projectName, OrgUnitsFormName.Description);
        await this.locator(this.projectDescriptionInput).fill(newDescription);
        await this.orgUnitsForm.clickSecondaryBtn();
    }

    async clickNewProjectBtn(): Promise<void> {
        await this.page.locator(this.newProjectBtn).click();
    }

    async fillProjectNameInput(projectName: string): Promise<void> {
        await this.locator(this.projectNameInput).fill(projectName);
    }

    async fillProjectDescriptionInput(projectDescription: string): Promise<void> {
        await this.locator(this.projectDescriptionInput).fill(projectDescription);
    }

    async openNewProjectForm(): Promise<void> {
        await this.clickNewProjectBtn();
        await this.orgUnitsForm.closeTourModalIfAppear();
    }

    async selectScope(options?: { scope?: Scope; scopeEntityName?: string }): Promise<void> {
        this.assetsScopeSection.searchElement = this.scopeSearchInputPlaceholder;
        await this.assetsScopeSection.selectScope(options);
    }

    async createProject(
        projectName: string,
        projectDescription?: string,
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
        accessRules?: { subjectName: AccessRulesSubjectType; userName: string; roleName: RoleDisplayName },
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.clickNewProjectBtn();
        await this.selectScope(options);

        // fill project name and description
        await this.fillProjectNameInput(projectName);
        if (projectDescription) {
            await this.fillProjectDescriptionInput(projectDescription);
        }
        await this.orgUnitsForm.clickPrimaryBtn();

        // fill gpu quota settings
        if (gpuQuotaConfig) {
            await this.orgUnitsForm.configGpuQuota(gpuQuotaConfig);
        }

        await this.orgUnitsForm.getSaveAndContinueBtn().click();
        await this.page.locator(this.orgUnitsForm.gpuOverQuotaToggleDiv).waitFor({ state: 'visible' });

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

        await this.page.locator(this.orgUnitsForm.schedulingRulesAddRuleBtn).waitFor({ state: 'visible' });

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

    async waitForProjectToBeReady(projectName: string, timeout: number = 120000): Promise<void> {
        await this.page.waitForSelector(
            `tr:has(td${this.projectNameCol}:has-text("${projectName}")):has(td${this.statusCol}:has-text("Ready"))`,
            { timeout }
        );
    }

    async getProjectsFromGrid(): Promise<Array<Record<string, string>>> {
        const data = await this.qTable.getTableData();
        logger.info(`table data: \n${JSON.stringify(data)}`);

        return data;
    }

    async getProjectByNameFromGrid(projectName: string): Promise<Record<string, string>> {
        const data = await this.qTable.getEntityData('Project', projectName);
        logger.info(`table data: \n${JSON.stringify(data)}`);

        return data;
    }

    async openProjectSubjectDetails(projectName: string): Promise<void> {
        await this.page
            .locator('table.q-table tbody tr')
            .filter({ has: this.page.locator(this.projectNameCol, { hasText: projectName }) })
            .locator(`${this.accessRuleCol} button`)
            .click();
    }

    async getProjectSubjectInfo(projectName: string): Promise<Array<Record<string, string>>> {
        const table = new QTable(this.page, '.dialog-content-container table.q-table');

        await this.openProjectSubjectDetails(projectName);
        const data = await table.getTableData();

        await this.orgUnitsForm.getCloseBtn().click();
        return data;
    }

    async deleteProject(projectName: string): Promise<void> {
        await this.toggleProject(projectName);
        await this.page.locator(this.deleteProjectBtn).click();
        await this.page.locator(this.deleteProjectModalBtn).click();
    }
}
