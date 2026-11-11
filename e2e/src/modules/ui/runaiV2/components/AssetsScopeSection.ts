/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from '../pages/BaseV2Page';
import { Scope } from '../../../api/apis/assetsApi';
import { apiUtils, flowRootState, settings } from '../../../helpers';
import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../../../helpers/models/featureFlags';
import { tenantName } from '../../../helpers/config/const';

export class AssetsScopeSection extends BaseV2Page {
    scopeTreeBtn: string;

    scopeModalSelectBtn: string;

    searchElement: string;

    tenantScopeLabel: string;

    departmentScopeLabel: string;

    clusterScopeLabel: string;

    tenantScopeCheckbox: string;

    clusterScopeCheckbox: string;

    departmentScopeCheckbox: string;

    projectScopeCheckboxDepartmentsEnabled: string;

    projectScopeCheckboxDepartmentsDisabled: string;

    constructor(page: Page) {
        super(page);

        this.scopeTreeBtn = '[aid="scope-tree-btn"]';
        this.scopeModalSelectBtn = '[aid="org-tree-modal-select-btn"]';
        this.searchElement = '[placeholder="Search for a cluster, department, or project"]';

        this.tenantScopeLabel = '.q-tree__node-header >> nth=0';
        this.clusterScopeLabel = '.q-tree__children > div > div >> nth=0';
        this.departmentScopeLabel =
            '.q-tree__children > div > div:nth-child(2) > div > div > div > .q-tree__node-header-content >> nth=0';

        this.tenantScopeCheckbox = 'div[role="checkbox"] >> nth=0';
        this.clusterScopeCheckbox = 'div[role="checkbox"] >> nth=1';
        this.departmentScopeCheckbox = 'div[role="checkbox"] >> nth=2';
        this.projectScopeCheckboxDepartmentsEnabled = 'div[role="checkbox"] >> nth=3';
        this.projectScopeCheckboxDepartmentsDisabled = 'div[role="checkbox"] >> nth=2';
    }

    private async clickOnCurrentClusterLabel(): Promise<void> {
        return await this.v2frame.locator(`div[role="dialog"] >> text=${flowRootState.cluster.name}`).click();
    }

    private async selectProjectScope(projectName: string): Promise<void> {
        await this.v2frame.locator(this.scopeTreeBtn).click();
        await this.v2frame.locator(this.searchElement).fill(projectName);
        const projectCheckbox = this.projectScopeCheckboxDepartmentsEnabled;
        await this.v2frame.locator(projectCheckbox).click();

        await this.v2frame.locator(this.scopeModalSelectBtn).click();
    }

    private async selectTenantScope(): Promise<void> {
        await this.v2frame.locator(this.scopeTreeBtn).click();
        await this.v2frame.locator(this.tenantScopeCheckbox).click();
        await this.v2frame.locator(this.scopeModalSelectBtn).click();
    }

    private async selectDepartmentScope(departmentName?: string): Promise<void> {
        await this.v2frame.locator(this.scopeTreeBtn).click();
        await this.v2frame.locator(this.searchElement).fill(departmentName);
        const checkboxLocator = settings.isSelfHosted
            ? this.v2frame.locator(this.departmentScopeCheckbox)
            : this.v2frame
                  .locator(
                      `.q-checkbox:left-of([aid="${tenantName}/${flowRootState.cluster.name}/${departmentName}"])`
                  )
                  .first();

        const isChecked = await checkboxLocator.isChecked();
        if (!isChecked) {
            await checkboxLocator.click();
        }
        await this.v2frame.locator(this.scopeModalSelectBtn).click();
    }

    private async selectClusterScope(clusterName?: string): Promise<void> {
        await this.v2frame.locator(this.scopeTreeBtn).click();
        await this.v2frame.locator(this.searchElement).fill(clusterName);
        await this.v2frame.locator(this.clusterScopeCheckbox).click();
        await this.v2frame.locator(this.scopeModalSelectBtn).click();
    }

    private scopeSelection = {
        cluster: (clusterName: string): Promise<void> => this.selectClusterScope(clusterName),
        department: (departmentName: string): Promise<void> => this.selectDepartmentScope(departmentName),
        project: (projectName: string): Promise<void> => this.selectProjectScope(projectName),
        tenant: (): Promise<void> => this.selectTenantScope()
    };

    public async selectScope(options?: { scope?: Scope; scopeEntityName?: string }): Promise<void> {
        if (options?.scopeEntityName) {
            await this.scopeSelection[options.scope](options.scopeEntityName);
        } else if (options?.scope) {
            await this.scopeSelection[options.scope]();
        }
    }
}
