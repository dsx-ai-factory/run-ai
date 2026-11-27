/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Locator, Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { QTable } from '../components/QTable';
import { logger, settings } from '../../../helpers';
import { existsSync, mkdirSync } from 'fs';
import { join } from 'path';

const ALL_NODES = 'All nodes';
const ALL_WORKLOADS = 'All workloads';
const ALL_PROJECTS = 'All projects';
const ALL_DEPARTMENTS = 'All departments';

export enum AvgGpuMetricSelectLabels {
    HIGHEST_AVG_GPU_ALLOC = 'Highest Avg. GPU allocation',
    LOWEST_AVG_GPU_ALLOC = 'Lowest Avg. GPU allocation',
    HIGHEST_AVG_GPU_UTIL = 'Highest  Avg. GPU utilization',
    LOWEST_AVG_GPU_UTIL = 'Lowest  Avg. GPU utilization',
    HIGHEST_AVG_GPU_MEM_UTIL = 'Highest Avg. GPU memory utilization',
    LOWEST_AVG_GPU_MEM_UTIL = 'Lowest Avg. GPU memory utilization'
}

export class OverviewPageV2 extends BaseV2Page {
    emptyWidgetMsgDiv: string;

    emptyStateContainerDiv: string;

    widgetActionBtn: string;

    readyNodesWidget: string;

    readyNodesWidgetStatDataDiv: string;

    readyNodesWidgetNameDiv: string;

    readyGpuWidget: string;

    readyGpuWidgetStatDataDiv: string;

    readyGpuWidgetNameDiv: string;

    allocatedGpuWidget: string;

    allocatedGpuWidgetStatDataDiv: string;

    allocatedGpuWidgetNameDiv: string;

    idleAllocatedGpuWidget: string;

    idleAllocatedGpuWidgetStatDataDiv: string;

    idleAllocatedGpuWidgetNameDiv: string;

    runningWorkloadsWidget: string;

    runningWorkloadsWidgetStatDataDiv: string;

    runningWorkloadsWidgetNameDiv: string;

    pendingWorkloadsWidget: string;

    pendingWorkloadsWidgetStatDataDiv: string;

    pendingWorkloadsWidgetNameDiv: string;

    failedWorkloadsWidget: string;

    failedWorkloadsWidgetStatDataDiv: string;

    failedWorkloadsWidgetNameDiv: string;

    runningWorkloadsByTypeWidget: string;

    workloadGroupBySelect: string;

    runningWorkloadsByTypeRedirectBtn: string;

    workloadsByStatusWidget: string;

    workloadsByStatusNameDiv: string;

    orgUnitRankingWidget: string;

    orgUnitRankingSelect: string;

    orgRankingMetricSelect: string;

    highestAvgGpuAllocationOption: string;

    lowestAvgGpuAllocationOption: string;

    highestAvgGpuUtilizationOption: string;

    lowestAvgGpuUtilizationOption: string;

    highestAvgGpuMemUtilizationOption: string;

    lowestAvgGpuMemUtilizationOption: string;

    orgRankingRedirectBtn: string;

    orgUnitMetricAvgGpuDevicesAlloc: string;

    orgUnitTimeRangeFilter: string;

    orgUnitTimeRange24HoursOption: string;

    orgUnitTimeRange7DaysOption: string;

    orgUnitTimeRange30DaysOption: string;

    orgUnitTimeRangeFilterApplyBtn: string;

    pendingWorkloadDistTimeWidget: string;

    pendingWorkloadDistTimeWidgetNameDiv: string;

    idleGpuWorkloadsWidget: string;

    idleGpuWorkloadsWidgetNameDiv: string;

    gpuAllocationRatioWidget: string;

    gpuAllocationRatioWidgetNameDiv: string;

    freeNodesByNodepoolWidget: string;

    freeNodesByNodepoolWidgetDivName: string;

    overTimeRangeWdigets: string;

    overTimeRangeDatePicker: string;

    resourceUtilizationWidget: string;

    resourceUtilizationWidgetNameDiv: string;

    resourceUtilWidgetSummary: string;

    resourceUtilWidgetChart: string;

    resourceUtilLegendGpuComputeUtil: string;

    resourceUtilLegendGpuMemUtil: string;

    resourceUtilLegendCpuComputeUtil: string;

    resourceUtilLegendCpuMemUtil: string;

    resourceAllocationWidgetNameDiv: string;

    resourceAllocationWidgetChart: string;

    resourceAllocationWidget: string;

    workloadsByCategoryWidgetNameDiv: string;

    workloadsByCategoryChart: string;

    workloadsByCategoryWidget: string;

    nodePoolFilterDiv: string;

    nodePoolFilterInput: string;

    projDepmtTable: QTable;

    idleWorkloadsTable: QTable;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Analytics}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Overview}"]`;

        // common elements
        this.emptyWidgetMsgDiv = '[aid="empty-widget-message"]';
        this.emptyStateContainerDiv = '.empty-state-container';
        this.widgetActionBtn = '[aid="run-page-action-select-columns-btn"]';

        // top telemetry widgets
        this.readyNodesWidget = '[aid="ready-nodes-widget"]';
        this.readyNodesWidgetStatDataDiv = `${this.readyNodesWidget} .stat-data`;
        this.readyNodesWidgetNameDiv = '[aid="ready-/-total-gpu-nodes"]';
        this.readyGpuWidget = '[aid="ready-gpu-widget"]';
        this.readyGpuWidgetStatDataDiv = `${this.readyGpuWidget} .stat-data`;
        this.readyGpuWidgetNameDiv = '[aid="ready-/-total-gpu-devices"]';
        this.allocatedGpuWidget = '[aid="allocated-gpu-widget"]';
        this.allocatedGpuWidgetStatDataDiv = `${this.allocatedGpuWidget} .stat-data`;
        this.allocatedGpuWidgetNameDiv = '[aid="allocated-gpu-devices"]';
        this.idleAllocatedGpuWidget = '[aid="idle-allocated-gpu-widget"]';
        this.idleAllocatedGpuWidgetStatDataDiv = `${this.idleAllocatedGpuWidget} .stat-data`;
        this.idleAllocatedGpuWidgetNameDiv = '[aid="idle-allocated-gpu-devices"]';
        this.runningWorkloadsWidget = '[aid="running-workloads-widget"]';
        this.runningWorkloadsWidgetStatDataDiv = `${this.runningWorkloadsWidget} .stat-data`;
        this.runningWorkloadsWidgetNameDiv = '[aid="running-workloads"]';
        this.pendingWorkloadsWidget = '[aid="pending-workloads-widget"]';
        this.pendingWorkloadsWidgetStatDataDiv = `${this.pendingWorkloadsWidget} .stat-data`;
        this.pendingWorkloadsWidgetNameDiv = '[aid="pending-workloads"]';
        this.failedWorkloadsWidget = '[aid="failed-workloads-widget"]';
        this.failedWorkloadsWidgetStatDataDiv = `${this.failedWorkloadsWidget} .stat-data`;
        this.failedWorkloadsWidgetNameDiv = '[aid="failed-workloads"]';

        // Running workloads by category/type widget
        this.runningWorkloadsByTypeWidget = '[aid="workload-by-type-widget"]';
        this.workloadGroupBySelect = '[aid="workload-group-by-select"]';
        this.runningWorkloadsByTypeRedirectBtn = '[aid="workload-by-type-redirect-btn"]';

        // Workloads by status widget
        this.workloadsByStatusWidget = '[aid="workload-status-widget"]';
        this.workloadsByStatusNameDiv = '[aid="workloads-by-status"]';

        // Projects/Departments table widget
        this.orgUnitRankingWidget = '[aid="org-unit-ranking-widget"]';
        this.orgUnitRankingSelect = '[aid="ranking-entity-select"]';
        this.orgRankingMetricSelect = '[aid="ranking-metric-select"]';
        this.highestAvgGpuAllocationOption = AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_ALLOC;
        this.lowestAvgGpuAllocationOption = AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_ALLOC;
        this.highestAvgGpuUtilizationOption = AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_UTIL;
        this.lowestAvgGpuUtilizationOption = AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_UTIL;
        this.highestAvgGpuMemUtilizationOption = AvgGpuMetricSelectLabels.HIGHEST_AVG_GPU_MEM_UTIL;
        this.lowestAvgGpuMemUtilizationOption = AvgGpuMetricSelectLabels.LOWEST_AVG_GPU_MEM_UTIL;
        this.orgRankingRedirectBtn = '[aid="ranking-redirect-btn"]';
        this.orgUnitMetricAvgGpuDevicesAlloc = 'Avg. GPU devices allocation';
        this.orgUnitTimeRangeFilter = '[aid="show-the-last-filter-chip"]';
        this.orgUnitTimeRange24HoursOption = '[aid="radio-option-24 hours"]';
        this.orgUnitTimeRange7DaysOption = '[aid="radio-option-7 days"]';
        this.orgUnitTimeRange30DaysOption = '[aid="radio-option-30 days"]';
        this.orgUnitTimeRangeFilterApplyBtn = '[aid="apply-btn"]';

        // Pending workloads by pending time widget
        this.pendingWorkloadDistTimeWidget = '[aid="workload-pending-time-distribution-widget"]';
        this.pendingWorkloadDistTimeWidgetNameDiv = '[aid="pending-workloads-by-pending-time"]';

        // Idle workloads widget
        this.idleGpuWorkloadsWidget = '[aid="idle-gpu-workload-widget"]';
        this.idleGpuWorkloadsWidgetNameDiv = '[aid="idle-workloads"]';

        // Allocation ratio by node pool widget
        this.gpuAllocationRatioWidget = '[aid="gpu-allocation-ratio-widget"]';
        this.gpuAllocationRatioWidgetNameDiv = '[aid="allocation-ratio-by-node-pool"]';

        // Free nodes by node pool widget
        this.freeNodesByNodepoolWidget = '[aid="node-pool-free-resources-widget"]';
        this.freeNodesByNodepoolWidgetDivName = '[aid="free-nodes-by-node-pool"]';

        // over time range widgets (bottom charts)
        this.overTimeRangeWdigets = '[aid="overview-time-range-widgets"]';
        this.overTimeRangeDatePicker = '[aid="overtime-widget-date-picker"]';

        // Resource utilization widget
        this.resourceUtilizationWidget = '.resource-utilization-time-range-widget';
        this.resourceUtilizationWidgetNameDiv = '[aid="resources-utilization"]';
        this.resourceUtilWidgetSummary = '[aid="avg-utilization-summary"]';
        this.resourceUtilWidgetChart = '[id="resource-utilization-time-range-widget-chart-id"]';
        this.resourceUtilLegendGpuComputeUtil = `${this.resourceUtilizationWidget} .resource-utilization-legend__item:has-text("GPU compute utilization")`;
        this.resourceUtilLegendGpuMemUtil = `${this.resourceUtilizationWidget} .resource-utilization-legend__item:has-text("GPU memory utilization")`;
        this.resourceUtilLegendCpuComputeUtil = `${this.resourceUtilizationWidget} .resource-utilization-legend__item:has-text("CPU compute utilization")`;
        this.resourceUtilLegendCpuMemUtil = `${this.resourceUtilizationWidget} .resource-utilization-legend__item:has-text("CPU memory utilization")`;

        // Resource allocation widget
        this.resourceAllocationWidgetNameDiv = '[aid="resources-allocation"]';
        this.resourceAllocationWidgetChart = '[id="resource-allocation-time-range-widget-chart-id"]';
        // The widget is a ChartWidgetWrapper; scope actions/menu via the wrapper that contains the title.
        this.resourceAllocationWidget = `.chart-widget-wrapper:has(${this.resourceAllocationWidgetNameDiv})`;

        // Workloads by category widget
        this.workloadsByCategoryWidgetNameDiv = '[aid="workloads-by-category"]';
        this.workloadsByCategoryChart = '[id="workload-by-category-time-range-widget-chart-id"]';
        this.workloadsByCategoryWidget = `.chart-widget-wrapper:has(${this.workloadsByCategoryWidgetNameDiv})`;

        this.nodePoolFilterDiv = '[aid="filter-chip-node-pool"]';
        this.nodePoolFilterInput = '[aid="tag-autocomplete"]';

        this.projDepmtTable = new QTable(page, 'div[aid="ranking-table"] table.q-table');
        this.idleWorkloadsTable = new QTable(page, '[aid="idle-gpu-workload-widget"] table.q-table');
    }

    getReadyNodesWidgetElems(): Locator[] {
        return [
            this.page.locator(this.readyNodesWidget),
            this.page.locator(this.readyNodesWidgetNameDiv),
            this.page.locator(this.readyNodesWidget).getByRole('button', { name: ALL_NODES })
        ];
    }

    async getReadyNodesWidgetName(): Promise<string> {
        return await this.page.locator(this.readyNodesWidgetNameDiv).textContent();
    }

    getReadyGpuWdigetElems(): Locator[] {
        return [
            this.page.locator(this.readyGpuWidget),
            this.page.locator(this.readyGpuWidgetNameDiv),
            this.page.locator(this.readyGpuWidget).getByRole('button', { name: ALL_NODES })
        ];
    }

    async getReadyGpuWidgetName(): Promise<string> {
        return await this.page.locator(this.readyGpuWidgetNameDiv).textContent();
    }

    getAllocatedGpuWdigetElems(): Locator[] {
        return [
            this.page.locator(this.allocatedGpuWidget),
            this.page.locator(this.allocatedGpuWidgetNameDiv),
            this.page.locator(this.allocatedGpuWidget).getByRole('button', { name: ALL_NODES })
        ];
    }

    async getAllocatedGpuWidgetName(): Promise<string> {
        return await this.page.locator(this.allocatedGpuWidgetNameDiv).textContent();
    }

    getIdleAllocatedGpuWdigetElems(): Locator[] {
        return [
            this.page.locator(this.idleAllocatedGpuWidget),
            this.page.locator(this.idleAllocatedGpuWidgetNameDiv),
            this.page.locator(this.idleAllocatedGpuWidget).getByRole('button', { name: ALL_WORKLOADS })
        ];
    }

    async getIdleAllocatedGpuWidgetName(): Promise<string> {
        return await this.page.locator(this.idleAllocatedGpuWidgetNameDiv).textContent();
    }

    getRunningWorkloadsWdigetElems(): Locator[] {
        return [
            this.page.locator(this.runningWorkloadsWidget),
            this.page.locator(this.runningWorkloadsWidgetNameDiv),
            this.page.locator(this.runningWorkloadsWidget).getByRole('button', { name: ALL_WORKLOADS })
        ];
    }

    async getRunningWorkloadsWidgetName(): Promise<string> {
        return await this.page.locator(this.runningWorkloadsWidgetNameDiv).textContent();
    }

    getPendingWorkloadsWdigetElems(): Locator[] {
        return [
            this.page.locator(this.pendingWorkloadsWidget),
            this.page.locator(this.pendingWorkloadsWidgetNameDiv),
            this.page.locator(this.pendingWorkloadsWidget).getByRole('button', { name: ALL_WORKLOADS })
        ];
    }

    async getPendingWorkloadsWidgetName(): Promise<string> {
        return await this.page.locator(this.pendingWorkloadsWidgetNameDiv).textContent();
    }

    getFailedWorkloadsWdigetElems(): Locator[] {
        return [
            this.page.locator(this.failedWorkloadsWidget),
            this.page.locator(this.failedWorkloadsWidgetNameDiv),
            this.page.locator(this.failedWorkloadsWidget).getByRole('button', { name: ALL_WORKLOADS })
        ];
    }

    async getFailedWorkloadsWidgetName(): Promise<string> {
        return await this.page.locator(this.failedWorkloadsWidgetNameDiv).textContent();
    }

    getRunningWorkloadsByTypeWdigetElems(): Locator[] {
        return [
            this.page.locator(this.runningWorkloadsByTypeWidget),
            this.page.locator(this.workloadGroupBySelect),
            this.page.locator(this.runningWorkloadsByTypeRedirectBtn)
        ];
    }

    async getRunningWorkloadsByTypeWidgetName(): Promise<string> {
        return await this.page.locator(this.workloadGroupBySelect).textContent();
    }

    async selectRunningWorkloadsGroupBy(
        groupBy: 'Running workloads by category' | 'Running workloads by type'
    ): Promise<void> {
        await this.page.locator(this.workloadGroupBySelect).click();
        await this.page.getByRole('option', { name: groupBy, exact: true }).click();
    }

    getWorkloadsByStatusWdigetElems(): Locator[] {
        return [
            this.page.locator(this.workloadsByStatusWidget),
            this.page.locator(this.workloadsByStatusNameDiv)
        ];
    }

    async getWorkloadStatusWidgetName(): Promise<string> {
        return await this.page.locator(this.workloadsByStatusNameDiv).textContent();
    }

    getOrgUnitRankingWdigetElems(): Locator[] {
        return [
            this.page.locator(this.orgUnitRankingWidget),
            this.page.locator(this.orgUnitRankingSelect),
            this.page.locator(this.orgRankingMetricSelect),
            this.page.locator(this.orgRankingRedirectBtn)
        ];
    }

    getPendingWorkloadDistTimeWdigetElems(): Locator[] {
        return [
            this.page.locator(this.pendingWorkloadDistTimeWidget),
            this.page.locator(this.pendingWorkloadDistTimeWidgetNameDiv)
        ];
    }

    async getPendingWorkloadDistTimeWidgetName(): Promise<string> {
        return await this.page.locator(this.pendingWorkloadDistTimeWidgetNameDiv).textContent();
    }

    getIdleGpuWorkloadsWdigetElems(): Locator[] {
        return [
            this.page.locator(this.idleGpuWorkloadsWidget),
            this.page.locator(this.idleGpuWorkloadsWidgetNameDiv)
        ];
    }

    async getIdleGpuWorkloadsWidgetName(): Promise<string> {
        return await this.page.locator(this.idleGpuWorkloadsWidgetNameDiv).textContent();
    }

    getGpuAllocationRatioWdigetElems(): Locator[] {
        return [
            this.page.locator(this.gpuAllocationRatioWidget),
            this.page.locator(this.gpuAllocationRatioWidgetNameDiv)
        ];
    }

    async getGpuAllocationRatioWidgetName(): Promise<string> {
        return await this.page.locator(this.gpuAllocationRatioWidgetNameDiv).textContent();
    }

    getFreeNodesByNodepoolWdigetElems(): Locator[] {
        return [
            this.page.locator(this.freeNodesByNodepoolWidget),
            this.page.locator(this.freeNodesByNodepoolWidgetDivName)
        ];
    }

    async getFreeNodesByNodepoolWidgetName(): Promise<string> {
        return await this.page.locator(this.freeNodesByNodepoolWidgetDivName).textContent();
    }

    getOverTimeRangeWdigetsElems(): Locator[] {
        return [
            this.page.locator(this.overTimeRangeWdigets),
            this.page.locator(this.overTimeRangeDatePicker)
        ];
    }

    getResourceUtilizationWdigetElems(): Locator[] {
        return [
            this.page.locator(this.resourceUtilizationWidget),
            this.page.locator(this.resourceUtilizationWidgetNameDiv),
            this.page.locator(this.resourceUtilWidgetSummary),
            this.page.locator(this.resourceUtilWidgetChart),
            this.page.locator(this.resourceUtilLegendGpuComputeUtil),
            this.page.locator(this.resourceUtilLegendGpuMemUtil),
            this.page.locator(this.resourceUtilLegendCpuComputeUtil),
            this.page.locator(this.resourceUtilLegendCpuMemUtil)
        ];
    }

    async getResourceUtilizationWidgetName(): Promise<string> {
        return await this.page.locator(this.resourceUtilizationWidgetNameDiv).textContent();
    }

    async getResourceUtilizationWidgetLegendNames(): Promise<string[]> {
        const names = [];
        names.push((await this.page.locator(this.resourceUtilLegendGpuComputeUtil).textContent()).trim());
        names.push((await this.page.locator(this.resourceUtilLegendGpuMemUtil).textContent()).trim());
        names.push((await this.page.locator(this.resourceUtilLegendCpuComputeUtil).textContent()).trim());
        names.push((await this.page.locator(this.resourceUtilLegendCpuMemUtil).textContent()).trim());

        return names;
    }

    getResourceAllocationWdigetElems(): Locator[] {
        return [
            this.page.locator(this.resourceAllocationWidgetNameDiv),
            this.page.locator(this.resourceAllocationWidgetChart)
        ];
    }

    async getResourceAllocationWidgetName(): Promise<string> {
        return await this.page.locator(this.resourceAllocationWidgetNameDiv).textContent();
    }

    async getResourceAllocationWidgetLegendNames(): Promise<string[]> {
        const legend = this.page.locator(`${this.resourceAllocationWidget} .resource-allocation-legend`);
        await legend.waitFor({ state: 'visible' });
        const items = await legend.locator('.resource-allocation-legend__item').allTextContents();
        return items.map((t) => t.trim());
    }

    getWorkloadsByCategoryWdigetElems(): Locator[] {
        return [
            this.page.locator(this.workloadsByCategoryWidgetNameDiv),
            // The chart is only rendered when the scope has workloads in the time range, otherwise the widget shows its empty state
            this.page
                .locator(this.workloadsByCategoryChart)
                .or(this.page.locator(this.workloadsByCategoryWidget).locator(this.emptyWidgetMsgDiv))
        ];
    }

    async getWorkloadsByCategoryWidgetName(): Promise<string> {
        return await this.page.locator(this.workloadsByCategoryWidgetNameDiv).textContent();
    }

    async getReadyTotalGpuNodesWidgetData(): Promise<string> {
        return (await this.page.locator(this.readyNodesWidgetStatDataDiv).textContent()).trim();
    }

    async getReadyTotalGpuDevicesWidgetData(): Promise<string> {
        return (await this.page.locator(this.readyGpuWidgetStatDataDiv).textContent()).trim();
    }

    async getAllocatedGpuDevicesWidgetData(): Promise<number> {
        return parseFloat((await this.page.locator(this.allocatedGpuWidgetStatDataDiv).textContent()).trim());
    }

    async getIdleAllocatedGpuDevicesWidgetData(): Promise<number> {
        return parseFloat(
            (await this.page.locator(this.idleAllocatedGpuWidgetStatDataDiv).textContent()).trim()
        );
    }

    async getRunningWorkloadsWidgetData(): Promise<number> {
        return parseInt(
            (await this.page.locator(this.runningWorkloadsWidgetStatDataDiv).textContent()).trim()
        );
    }

    async getPendingWorkloadsWidgetData(): Promise<number> {
        return parseInt(
            (await this.page.locator(this.pendingWorkloadsWidgetStatDataDiv).textContent()).trim()
        );
    }

    async getFailedWorkloadsWidgetData(): Promise<number> {
        return parseInt(
            (await this.page.locator(this.failedWorkloadsWidgetStatDataDiv).textContent()).trim()
        );
    }

    async clickAllNodesLinkInReadyTotalGpuNodesWidget(): Promise<void> {
        await this.page.locator(this.readyNodesWidget).getByRole('button', { name: ALL_NODES }).click();
    }

    async clickAllNodesLinkInReadyTotalGpuDevicesWidget(): Promise<void> {
        await this.page.locator(this.readyGpuWidget).getByRole('button', { name: ALL_NODES }).click();
    }

    async clickAllNodesLinkInAllocatedGpuDevicesWidget(): Promise<void> {
        await this.page.locator(this.allocatedGpuWidget).getByRole('button', { name: ALL_NODES }).click();
    }

    async clickAllNodesLinkInFreeNodesByNodePoolWidget(): Promise<void> {
        await this.page
            .locator(this.freeNodesByNodepoolWidget)
            .getByRole('button', { name: ALL_NODES })
            .click();
    }

    async clickAllWorkloadsLinkInIdleAllocatedGpuDevicesWidget(): Promise<void> {
        await this.page
            .locator(this.idleAllocatedGpuWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInRunningWorkloadsWidget(): Promise<void> {
        await this.page
            .locator(this.runningWorkloadsWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInPendingWorkloadsWidget(): Promise<void> {
        await this.page
            .locator(this.pendingWorkloadsWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInFailedWorkloadsWidget(): Promise<void> {
        await this.page
            .locator(this.failedWorkloadsWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInRunningWorkloadsByCategoryWidget(): Promise<void> {
        await this.page
            .locator(this.runningWorkloadsByTypeWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInWorkloadsByStatusWidget(): Promise<void> {
        await this.page
            .locator(this.workloadsByStatusWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInPendingWorkloadsByPendingTimeWidget(): Promise<void> {
        await this.page
            .locator(this.pendingWorkloadDistTimeWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllWorkloadsLinkInIdleWorkloadsWidget(): Promise<void> {
        await this.page
            .locator(this.idleGpuWorkloadsWidget)
            .getByRole('button', { name: ALL_WORKLOADS })
            .click();
    }

    async clickAllDepartmentsLinkInAllocationRatioByNodePoolWidget(): Promise<void> {
        await this.page
            .locator(this.gpuAllocationRatioWidget)
            .getByRole('button', { name: ALL_DEPARTMENTS })
            .click();
    }

    async clickAllProjectsLinkInOrgUnitRankingWidget(): Promise<void> {
        await this.page
            .locator(this.orgUnitRankingWidget)
            .getByRole('button', { name: ALL_PROJECTS })
            .click();
    }

    async clickAllDepartmentsLinkInOrgUnitRankingWidget(): Promise<void> {
        await this.page
            .locator(this.orgUnitRankingWidget)
            .getByRole('button', { name: ALL_DEPARTMENTS })
            .click();
    }

    async switchViewInOrgUnitRankingWidget(viewName: 'Projects' | 'Departments'): Promise<void> {
        await this.page.locator(this.orgUnitRankingSelect).click();
        await this.page.getByRole('option', { name: viewName }).click();
    }

    async selectOrgUnitMetrics(metricName: string): Promise<void> {
        await this.page.locator(this.orgRankingMetricSelect).click();
        await this.page.getByRole('option', { name: metricName }).click();
    }

    async chooseOrgUnitMetricTimeRange(timeRange: '24 hours' | '7 days' | '30 days'): Promise<void> {
        await this.page
            .getByRole('columnheader', { name: this.orgUnitMetricAvgGpuDevicesAlloc })
            .locator(this.orgUnitTimeRangeFilter)
            .click();
        if (timeRange === '24 hours') {
            await this.page.locator(this.orgUnitTimeRange24HoursOption).click();
        } else if (timeRange === '7 days') {
            await this.page.locator(this.orgUnitTimeRange7DaysOption).click();
        } else {
            await this.page.locator(this.orgUnitTimeRange30DaysOption).click();
        }

        await this.page.locator(this.orgUnitTimeRangeFilterApplyBtn).click();
    }

    async getProjectsDepartmentsTableData(): Promise<Array<Record<string, string>>> {
        return await this.projDepmtTable.getTableData();
    }

    async getIdleWorkloadsTableData(): Promise<Array<Record<string, string>>> {
        return await this.idleWorkloadsTable.getTableData();
    }

    async clickProjDepmtLinkInOrgUnitRankingTable(name: string): Promise<void> {
        await this.page.locator(this.orgUnitRankingWidget).getByText(name).click();
    }

    async clickWorkloadsInIdleWorkloadsTable(workloadName: string): Promise<void> {
        await this.page.locator(this.idleGpuWorkloadsWidget).getByText(workloadName).click();
    }

    async clickLinkInWidget(widgetLocator: string, linkName: string): Promise<void> {
        await this.page
            .locator(widgetLocator)
            .locator('[aria-label="Explore chart data"]')
            .getByRole('button', { name: `View details for ${linkName}`, exact: true })
            .click();
    }

    async entityLinkExistInWidget(widgetLocator: string, entityName: string): Promise<boolean> {
        try {
            const widget = this.page.locator(widgetLocator);
            const drilldownLabel = widget
                .locator('[aria-label="Explore chart data"]')
                .getByRole('button', { name: `View details for ${entityName}`, exact: true });
            // The clickable overlay is removed at the terminal drill-down level.
            const terminalLabel = widget
                .locator('[aid$="-chart"] svg')
                .getByText(entityName, { exact: true });

            await drilldownLabel.or(terminalLabel).first().waitFor({ state: 'visible', timeout: 5000 });
            return true;
        } catch {
            return false;
        }
    }

    async fillNodePoolFilterInput(nodePool: string): Promise<void> {
        await this.page.locator(this.nodePoolFilterInput).fill(nodePool);
    }

    async setNodePoolFilter(nodePool: string): Promise<void> {
        await this.clickAddFilterButton();
        await this.page.getByText('Node pool', { exact: true }).click();
        await this.fillNodePoolFilterInput(nodePool);
        try {
            await this.page.getByRole('option', { name: nodePool }).click({ timeout: 3000 });
        } catch (e) {
            logger.warn(`option ${nodePool} did not appear after filling the input filter.`);
        }

        await this.clickApplyFilterButton();
    }

    async removeNodePoolFilter(): Promise<void> {
        const removeFilterBtn = this.page
            .locator(this.nodePoolFilterDiv)
            .getByRole('button', { name: 'Remove', exact: true });
        if (await removeFilterBtn.isVisible()) {
            await removeFilterBtn.click();
        }
    }

    async downloadCSV(
        widgetLocator: string,
        expectedFileName: string,
        downloadDir: string = settings.browserVideosPath
    ): Promise<string> {
        const dnldPromise = this.page.waitForEvent('download');

        await this.page.locator(widgetLocator).scrollIntoViewIfNeeded();
        await this.page.locator(widgetLocator).locator(this.widgetActionBtn).click();
        await this.page.locator(this.downloadCSVOption).click();

        const download = await dnldPromise;

        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const extension = download.suggestedFilename().split('.').pop() || 'csv';
        const fileName = `${expectedFileName}_${timestamp}.${extension}`;

        if (!existsSync(downloadDir)) {
            mkdirSync(downloadDir, { recursive: true });
        }

        const downloadPath = join(downloadDir, fileName);
        await download.saveAs(downloadPath);

        logger.info(`CSV file downloaded to: ${downloadPath}`);
        return downloadPath;
    }
}
