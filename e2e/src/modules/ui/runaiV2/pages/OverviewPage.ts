/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page, Locator } from 'playwright';
import {
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    PodStatus,
    projectResourcesUtils,
    TrainingWorkload,
    Utils,
    versionUtils,
    waitUtils,
    config
} from '../../../helpers';
import { stringUtil } from '../../utils/stringUtils';
import { kubeCli } from '../../../cli';
import { NodePoolAssignedResourcesCreate, Project } from '../../../api/apis/projectsDepartmentsApi';
import { submitTrainWorkload } from '../../../helpers/models/workloads';
import { NodePool } from '../../../api/apis/backendClusterApi';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

export const MIN_NEW_DEPARTMENT_METRIC_VERSION = '2.11';
export const MIN_TELEMETRY_NODE_POOL_FILTER_VERSION = '2.11';
export const MIN_READY_NODES_VERSION = '2.15';

export enum EWidgetComponentName {
    WorkloadStatus = 'WorkloadStatusWidget',
    RunningWorkloads = 'RunningWorkloadsWidget',
    PendingWorkloads = 'PendingWorkloadsWidget',
    OverviewTimeRangeWidgets = 'OverviewTimeRangeWidgets',
    NodePoolFreeResourcesWidget = 'NodePoolFreeResourcesWidget',
    WorkloadByTypeWidget = 'WorkloadByTypeWidget',
    ReadyNodesWidget = 'ReadyNodesWidget',
    ReadyGpuWidget = 'ReadyGpuWidget',
    AllocatedGpuWidget = 'AllocatedGpuWidget',
    GpuAllocationRatioWidget = 'GpuAllocationRatioWidget'
}

export interface IWidgetTest {
    name: EWidgetComponentName;
    desiredResult: string;
}
export async function generateNodePools(
    nodepoolsCount: number
): Promise<{ nodePoolName: string; nodePoolId: number }[]> {
    logger.info('creating nodepools for metrics api tests');
    const nodepools: { nodePoolName: string; nodePoolId: number }[] = [];
    const totalNumberOfNodesWithGpus: number = await kubeCli.getNumOfNodesWithGpus();
    const defaultNodePool: NodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

    nodepools.push({ nodePoolName: defaultNodePool.name, nodePoolId: defaultNodePool.id });

    if (nodepoolsCount > totalNumberOfNodesWithGpus) {
        logger.info(
            'Will run test with only 1 (default) nodepool, since there are not enough nodes with GPUs in the cluster.'
        );
        nodepoolsCount = 1;
    }

    for (let i = 0; i < nodepoolsCount - 1; i++) {
        const keyVal = nodePoolUtils.generateKeyValueLabels();
        const { nodePoolName, nodePoolId } = await nodePoolUtils.createNodePoolAndLabelNodeToAssignToNodePool(
            keyVal.key,
            keyVal.value,
            []
        );
        nodepools.push({ nodePoolName, nodePoolId });
    }

    return nodepools;
}

export async function generateProjects(
    accessToken: string,
    projectCount: number,
    nodepools: { nodePoolName: string; nodePoolId: number }[],
    deservedGpus = [
        [5, 6],
        [7, 8]
    ]
): Promise<Project[]> {
    logger.info('creating projects for metrics api tests');
    const promises = [];
    for (let i = 0; i < projectCount; i++) {
        const projectName = objectGenerator.generateNameByPrefix('p-metricsapi');
        const namespace = Utils.namespaceFromProjectName(projectName);
        const nodePoolsResources = new Array<NodePoolAssignedResourcesCreate>();
        for (let j = 0; j < nodepools.length; j++) {
            nodePoolsResources.push({
                gpu: {
                    deserved: deservedGpus[i][j],
                    maxAllowed: deservedGpus[i][j]
                },
                nodePool: { id: nodepools[j].nodePoolId, name: nodepools[j].nodePoolName }
            });
        }

        promises.push(
            projectResourcesUtils.createProject(
                namespace,
                {
                    name: projectName,
                    nodePoolsResources: nodePoolsResources
                },
                accessToken
            )
        );
    }

    return await Promise.all(promises);
}

export async function generateWorkloads(
    _accessToken: string,
    projects: Project[],
    nodepools: { nodePoolName: string; nodePoolId: number }[],
    runningWorkloads: TrainingWorkload[],
    pendingWorkloads: TrainingWorkload[]
): Promise<[runningWorkloads: TrainingWorkload[], pendingWorkloads: TrainingWorkload[]]> {
    const workloadsPromises: Promise<TrainingWorkload | void>[] = [];

    for (let n = 0; n < (nodepools.length === 1 ? 1 : nodepools.length - 1); n++) {
        for (let p = 0; p < projects.length; p++) {
            workloadsPromises.push(
                submitTrainWorkload(
                    'running-overview-ui',
                    projects[p].name,
                    config.quickStartImage.training,
                    'sleep infinity',
                    1,
                    PodStatus.RUNNING
                ).then((workload) => {
                    runningWorkloads.push(workload);
                })
            );

            workloadsPromises.push(
                submitTrainWorkload(
                    'pending-overview-ui',
                    projects[p].name,
                    config.quickStartImage.training,
                    'sleep infinity',
                    400,
                    PodStatus.PENDING,
                    nodepools[n].nodePoolName
                ).then((workload) => {
                    pendingWorkloads.push(workload);
                })
            );
        }
    }

    logger.info('submitting workloads');
    await Promise.all(workloadsPromises);

    return [runningWorkloads, pendingWorkloads];
}

export class OverviewPage extends BaseV2Page {
    widgetsNames: string[];

    filterButtonSelector = '.add-filter-btn';

    nodePoolFilterSectionSelector = '.q-item__section:has-text("Node pool")';

    filterAutocompleteSelector = '[aid="tag-autocomplete"]';

    applyFilterButtonSelector = 'button span:has-text("Apply")';

    loadingSectionSelector = '.loading-section';

    removeNodePoolFilterChipSelector = '[aid="filter-chip-node-pool"] [aria-label="Remove"]';

    nodePoolFilterChipSelector = '[aid="filter-chip-node-pool"]';

    workloadGroupSelector = '[aid="workload-group-by-select"]';

    workloadByTypeWidgetSelector = '[aid="workload-by-type-widget"]';

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Analytics}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Overview}"]`;
        this.initWidgetsNames();
    }

    private initWidgetsNames(): void {
        const widgetsNames = [
            EWidgetComponentName.ReadyNodesWidget,
            EWidgetComponentName.ReadyGpuWidget,
            EWidgetComponentName.AllocatedGpuWidget,
            EWidgetComponentName.RunningWorkloads,
            EWidgetComponentName.PendingWorkloads,
            EWidgetComponentName.OverviewTimeRangeWidgets,
            ...(this.isVersionSupported(MIN_NEW_DEPARTMENT_METRIC_VERSION)
                ? [EWidgetComponentName.GpuAllocationRatioWidget]
                : []),
            EWidgetComponentName.NodePoolFreeResourcesWidget,
            EWidgetComponentName.WorkloadByTypeWidget,
            EWidgetComponentName.WorkloadStatus
        ];
        this.widgetsNames = widgetsNames.map((name: EWidgetComponentName) =>
            stringUtil.camelCaseToHyphenated(name)
        );
    }

    getWidgetSelector(widgetName: EWidgetComponentName): string {
        return `[aid="${stringUtil.camelCaseToHyphenated(widgetName)}"]`;
    }

    async getStatWidgetText(widgetName: EWidgetComponentName): Promise<string> {
        return this.page.locator(`${this.getWidgetSelector(widgetName)} .stat-data`).textContent();
    }

    getWidgetXLabelLocator(widgetName: EWidgetComponentName): string {
        return `${this.getWidgetSelector(widgetName)} [aria-label="Explore chart data"]`;
    }

    // Drillable categories are buttons; terminal categories remain as ECharts SVG labels.
    getWidgetCategoryLabelLocator(widgetName: EWidgetComponentName): string {
        const widgetSelector = this.getWidgetSelector(widgetName);
        return `${widgetSelector} [aria-label="Explore chart data"], ${widgetSelector} svg text`;
    }

    async getWidgetXLabel(widgetName: EWidgetComponentName): Promise<string> {
        const labels = await this.page
            .locator(this.getWidgetCategoryLabelLocator(widgetName))
            .allTextContents();
        return labels.join(' ');
    }

    // Match either label form; the text filter keeps chart titles and axis ticks out.
    drillableCategoryLocator(widgetName: EWidgetComponentName, text: string): Locator {
        const widgetSelector = this.getWidgetSelector(widgetName);
        return this.page
            .locator(
                `${widgetSelector} [aria-label="Explore chart data"] button, ` +
                    `${widgetSelector} svg text`
            )
            .filter({ hasText: text });
    }

    async isDrillableCategoryVisible(widgetName: EWidgetComponentName, text: string): Promise<boolean> {
        return (await this.drillableCategoryLocator(widgetName, text).count()) > 0;
    }

    async clickWidgetXLabel(widgetName: EWidgetComponentName, text: string): Promise<void> {
        return await this.clickWidgetXLabelWithText(widgetName, text);
    }

    async clickWidgetXLabelWithText(widgetName: EWidgetComponentName, text: string): Promise<void> {
        // SVG labels are not hit-testable like buttons.
        return await this.drillableCategoryLocator(widgetName, text).first().click({ force: true });
    }

    async clickWorkloadGroupSelector(): Promise<void> {
        return await this.page.click(this.workloadGroupSelector);
    }

    async clickWorkloadByText(text: string): Promise<void> {
        if (!(await this.page.isVisible(`${this.workloadGroupSelector}:has-text("${text}")`))) {
            this.clickWorkloadGroupSelector();
            return await this.page.locator(`[role="option"]:has-text("${text}")`).click();
        }
    }

    async filterWorkloadType(type: string): Promise<void> {
        const legendItem = this.page
            .locator(`${this.workloadByTypeWidgetSelector} [aid="workload-by-type-chart"]`)
            .getByText(type, { exact: true });
        await legendItem.waitFor({ state: 'visible', timeout: 5000 });
        await legendItem.click();
    }

    async waitWorkloadTypeToLoad(): Promise<void> {
        await this.page
            .locator(`${this.workloadByTypeWidgetSelector} [aid="workload-by-type-chart"] svg text`)
            .first()
            .waitFor({ state: 'visible', timeout: 5000 });
    }

    async isWorkloadTypeVisible(type: string): Promise<boolean> {
        return await this.page
            .locator(`${this.workloadByTypeWidgetSelector} [aid="workload-by-type-chart"]`)
            .getByText(type, { exact: true })
            .isVisible();
    }

    async isWorkloadTypeHidden(type: string): Promise<boolean> {
        const legendItem = this.page
            .locator(`${this.workloadByTypeWidgetSelector} [aid="workload-by-type-chart"]`)
            .getByText(type, { exact: true });
        return await legendItem.evaluate(
            (element) => getComputedStyle(element).fill === 'rgb(207, 210, 215)'
        );
    }

    async applyNodePoolFilter(nodePoolName: string): Promise<void> {
        if (await this.page.isVisible(this.nodePoolFilterChipSelector)) {
            await this.closeNodePoolFilter();
        }
        await this.page.waitForSelector(this.filterButtonSelector);
        await this.page.click(this.filterButtonSelector);
        await this.page.click(this.nodePoolFilterSectionSelector);
        await this.page.click(this.filterAutocompleteSelector);
        await waitUtils.waitForElementToExist(this, `.q-item__label:has-text("${nodePoolName}")`);
        await this.page.click(`.q-item__label:has-text("${nodePoolName}")`);
        await this.page.click(this.applyFilterButtonSelector);
        await waitUtils.waitForElementToNotExist(this, this.loadingSectionSelector);
    }

    async closeNodePoolFilter(): Promise<void> {
        await this.page.click(this.removeNodePoolFilterChipSelector);
        await waitUtils.waitForElementToNotExist(this, this.nodePoolFilterChipSelector);
    }

    isVersionSupported(minVersion: string): boolean {
        return versionUtils.isVersionSupported(flowRootState.cluster.version, minVersion);
    }

    isNodePoolFilterSupported(): boolean {
        return this.isVersionSupported(MIN_TELEMETRY_NODE_POOL_FILTER_VERSION);
    }

    isDepartmentMetricSupported(): boolean {
        return this.isVersionSupported(MIN_NEW_DEPARTMENT_METRIC_VERSION);
    }

    isReadyNodesFeatureSupported(): boolean {
        return this.isVersionSupported(MIN_READY_NODES_VERSION);
    }
}
