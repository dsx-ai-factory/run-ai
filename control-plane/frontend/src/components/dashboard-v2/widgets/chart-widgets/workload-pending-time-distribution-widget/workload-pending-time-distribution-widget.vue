<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    @link-clicked="redirectToWorkloads"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty || isDefaultBuckets"
    :announcement="drilldownAnnouncement"
    hide-actions
  >
    <div v-if="chartOption" ref="chartAreaRef" class="workload-pending-time-chart-area">
      <chart-drilldown-breadcrumbs :items="drillBreadcrumbs" @navigate="drillUpTo" />
      <v-chart
        aid="workload-pending-time-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="workload-pending-time-chart"
        key="workload-pending-time-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="workload-pending-time-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="workload-pending-time-axis-label"
          type="button"
          :aria-label="`View details for ${category}`"
          @click="drillDownByCategory(category)"
        >
          {{ category }}
        </button>
      </div>
    </div>
  </chart-widget-wrapper>
</template>

<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, ref, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption, BarSeriesOption } from "echarts";
// routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
// store
import { useClusterStore } from "@/stores/cluster.store";
// models
import {
  EWidgetEntity,
  type IWidgetWrapperOptions,
  chartColorPaletteArray,
  EWidgetEmptyMessages,
  CHART_DATA_LABEL_COLOR,
  ACTIVE_AXIS_LABEL_COLOR,
  DRILLDOWN_CHART_ANIMATION_PRESET,
  getDrilldownTransitionPreset,
  DRILLDOWN_UNIVERSAL_TRANSITION,
} from "@/models/chart.model";
import type { EWorkloadTelemetryGroupByWithoutBucket, EWorkloadTelemetryPointMeta } from "@/models/workload.model";
// cmps
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartDrilldownBreadcrumbs } from "@/components/common/charts/chart-drilldown-breadcrumbs";
// service
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { filterService } from "@/services/filter.service/filter.service";
import {
  Phase,
  type TelemetryResponseValuesInner,
  WorkloadSortFilterFields,
  WorkloadTelemetryType,
} from "@/swagger-models/workloads-client";
// model
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
import { workloadIndexColumns } from "@/table-models/workload.table-model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { EColumnFilterType, EFilterOperator, type IFilterBy, type IFreeTextFilterModel } from "@/models/filter.model";
import { ETableFilters } from "@/models/table.model";
// util
import { widgetUtil } from "@/utils/widget.util";
import {
  chartUtil,
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
  STANDARD_TOOLTIP_STYLE,
} from "@/utils/chart.util";
import { EIntervalLabels } from "@/models/interval.model";
import { intervalUtil } from "@/utils/interval.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

interface HcDataPoint {
  name: string;
  y: number;
  drilldown?: boolean;
  meta?: { entityId: string; groupByKey: string };
  color?: string;
}

interface HcSeriesItem {
  name?: string;
  data: HcDataPoint[];
  color?: string;
  meta?: { groupByKey?: string; entityId?: string };
}

interface DrillState {
  option: EChartsOption;
  meta: EWorkloadTelemetryPointMeta | null;
  categoryMetaMap: Map<string, EWorkloadTelemetryPointMeta>;
  drilldownAxisCategories: string[];
}

const chartGridOptions = {
  top: 11,
  bottom: 74,
  left: 56,
  right: 12,
};

const drilldownChartGridOptions = {
  top: 52,
  bottom: 86,
  left: 56,
  right: 12,
};

const yAxisTitleOptions = {
  name: "Workloads",
  nameLocation: "middle" as const,
  nameRotate: 90,
  nameGap: 36,
  nameTextStyle: {
    color: "#585f69",
    fontSize: 12,
    fontWeight: 400,
  },
};

const legendOptions = {
  show: true,
  bottom: 16,
  icon: "circle",
  itemWidth: 12,
  itemHeight: 12,
  itemGap: 20,
  textStyle: {
    fontWeight: "bold" as const,
    fontSize: 12,
    color: "#333",
    padding: [0, 0, 0, 5],
  },
};

export default defineComponent({
  name: "workload-pending-time-distribution-widget",
  components: { ChartDrilldownBreadcrumbs, ChartWidgetWrapper, VChart },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
    hasDepartments: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      isDataLoading: false as boolean,
      displayLoading: true as boolean,
      chartOption: null as null | EChartsOption,
      error: false as boolean,
      isChartEmpty: false as boolean,
      departmentId: undefined as string | undefined,
      groupBy: EWorkloadTelemetryGroupBy.ClusterId as EWorkloadTelemetryGroupBy,
      bucketColorsMap: null as Map<string, string> | null,
      isDefaultBuckets: false as boolean,
      wrapperOptions: {
        title: "Pending workloads by pending time",
        linkText: "All workloads",
        tooltipText:
          "Shows the distribution of workloads by <br/> pending time to help identify potential <br/> bottlenecks or scheduling issues",
        entityType: EWidgetEntity.Workload,
        emptyMessage: EWidgetEmptyMessages.NoPendingWorkloads,
      } as IWidgetWrapperOptions,
      drillStack: [] as DrillState[],
      drillBreadcrumbs: [] as string[],
      currentDrillMeta: null as EWorkloadTelemetryPointMeta | null,
      categoryMetaMap: new Map<string, EWorkloadTelemetryPointMeta>(),
      drilldownAxisCategories: [] as string[],
      drilldownAnnouncement: "",
    };
  },
  setup() {
    const chartRef = ref<InstanceType<typeof VChart> | null>(null);
    const chartAreaRef = ref<HTMLElement | null>(null);
    const { layout, getCategoryAxisLabelOptions, getGroupedBarLayout } = useAdaptiveCategoricalChartLayout({
      chartRef,
      containerRef: chartAreaRef,
      horizontalGridPadding: chartGridOptions.left + chartGridOptions.right,
    });
    return {
      chartRef,
      chartAreaRef,
      layout,
      getCategoryAxisLabelOptions,
      getGroupedBarLayout,
    };
  },
  async created() {
    this.loadInitialChartData();
    this.startRefreshLoadData();
  },
  computed: {
    clusterName(): string {
      const cluster = this.clusterStore.getClusterById(this.clusterId);
      return cluster?.name || "";
    },
  },
  methods: {
    isDefaultBucketsData(telemetryData: TelemetryResponseValuesInner[]): boolean {
      return telemetryData.every((item) => item.value === "0");
    },
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.WorkloadPendingTimeDistributionWidget, this.refreshChart);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadPendingTimeDistributionWidget);
    },
    async refreshChart(): Promise<void> {
      if (this.isDataLoading) return;

      if (this.currentDrillMeta) {
        await this.refreshDrilldownData();
      } else {
        await this.loadInitialChartData();
      }
    },
    formatDrilldownAxisLabel(value: string): string {
      return chartUtil.formatCategoricalAxisLabel(value);
    },
    getChartGridOptions(isDrilldown: boolean): typeof chartGridOptions {
      return isDrilldown ? drilldownChartGridOptions : chartGridOptions;
    },
    convertSeriesToECharts(
      hcSeries: HcSeriesItem[],
      _isDrilldown = false,
      parentGroupId?: string,
    ): {
      categories: string[];
      series: BarSeriesOption[];
    } {
      const categorySet = new Set<string>();
      const categories: string[] = [];
      this.categoryMetaMap.clear();
      for (const s of hcSeries) {
        for (const d of s.data) {
          if (!categorySet.has(d.name)) {
            categorySet.add(d.name);
            categories.push(d.name);
          }
          if (d.drilldown && d.meta && !this.categoryMetaMap.has(d.name)) {
            this.categoryMetaMap.set(d.name, d.meta as EWorkloadTelemetryPointMeta);
          }
        }
      }

      const series: BarSeriesOption[] = hcSeries.map((s) => {
        const catMap = new Map(s.data.map((d) => [d.name, d]));
        const seriesName = s.name || "";
        return {
          type: "bar",
          id: `wptd-${seriesName}`,
          name: seriesName,
          data: categories.map((cat) => {
            const pt = catMap.get(cat);
            return pt
              ? {
                  value: pt.y,
                  groupId: parentGroupId ?? pt.meta?.entityId ?? cat,
                  childGroupId: pt.meta?.entityId,
                  drilldown: pt.drilldown,
                  meta: pt.meta,
                  itemStyle: pt.color ? { color: pt.color } : undefined,
                }
              : { value: 0, groupId: parentGroupId ?? cat };
          }),
          universalTransition: DRILLDOWN_UNIVERSAL_TRANSITION,
          itemStyle: s.color ? { color: s.color } : undefined,
          ...chartUtil.getBarSeriesDefaults(),
          label: {
            show: true,
            position: "top" as const,
            formatter: "{c}",
            color: CHART_DATA_LABEL_COLOR,
            fontSize: 12,
            fontWeight: 400,
          },
        };
      });

      return { categories, series };
    },
    buildChartOption(
      categories: string[],
      series: BarSeriesOption[],
      isDrilldown = false,
      animateDrilldown = false,
    ): EChartsOption {
      const hasDrilldown = this.categoryMetaMap.size > 0;
      this.drilldownAxisCategories = hasDrilldown ? categories : [];
      const yAxisConfig = this.getYAxisConfig(series);
      const groupedBarLayout = this.getGroupedBarLayout(categories.length, series.length);
      return {
        ...DRILLDOWN_CHART_ANIMATION_PRESET,
        ...(animateDrilldown ? getDrilldownTransitionPreset() : {}),
        grid: this.getChartGridOptions(isDrilldown),
        xAxis: {
          type: "category",
          data: categories,
          axisLabel: hasDrilldown
            ? {
                ...this.getCategoryAxisLabelOptions(categories.length),
                rotate: 0,
                color: ACTIVE_AXIS_LABEL_COLOR,
                margin: 16,
                show: false,
                fontSize: 11,
              }
            : isDrilldown
              ? {
                  ...this.getCategoryAxisLabelOptions(categories.length),
                  show: true,
                  rotate: 0,
                  margin: 10,
                  color: "#666",
                  fontSize: 11,
                  interval: 0,
                  hideOverlap: false,
                  formatter: (value: string) => this.formatDrilldownAxisLabel(value),
                }
              : {
                  ...this.getCategoryAxisLabelOptions(categories.length),
                  rotate: 0,
                  margin: 16,
                  color: "#666",
                  fontSize: 11,
                },
          axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
          axisTick: { show: false },
          axisPointer: chartUtil.getHiddenAxisPointer(),
          splitLine: { show: false },
          triggerEvent: hasDrilldown,
        },
        yAxis: {
          type: "value",
          min: 0,
          interval: yAxisConfig.interval,
          max: yAxisConfig.max,
          axisLabel: { margin: 12, color: "#666", fontSize: 11, formatter: chartUtil.formatCompactAxisValue },
          axisLine: { show: false },
          axisTick: { show: false },
          axisPointer: chartUtil.getHiddenAxisPointer(),
          splitLine: {
            show: true,
            lineStyle: { color: "#e6e6e6", width: 1, type: "solid" },
          },
          ...yAxisTitleOptions,
        },
        tooltip: {
          trigger: "axis",
          axisPointer: chartUtil.getHiddenAxisPointer(),
          ...STANDARD_TOOLTIP_STYLE,
          formatter: chartUtil.formatDefaultTooltip,
        },
        legend: legendOptions,
        series: series.map((item) => ({ ...item, ...groupedBarLayout })),
      };
    },
    getYAxisConfig(series: BarSeriesOption[]): { interval: number; max: number } {
      return chartUtil.getBarChartYAxisConfig(series, { integerTicks: true });
    },
    async loadInitialChartData(): Promise<void> {
      try {
        this.isDataLoading = true;
        const res: TelemetryResponseValuesInner[] = await this.getWorkloadsTelemetry(
          EWorkloadTelemetryGroupBy.ClusterId,
        );
        const enrichedRes: TelemetryResponseValuesInner[] = widgetUtil.enrichTelemetryResponseByKeyValue(
          res,
          EWorkloadTelemetryGroupBy.ClusterId,
          this.clusterName,
          "name",
        );

        if (enrichedRes.length === 0) {
          this.isChartEmpty = true;
          this.isDefaultBuckets = false;
        } else {
          this.isChartEmpty = false;
          this.isDefaultBuckets = this.isDefaultBucketsData(enrichedRes);

          const hcSeries = this.aggregateTelemetryGroupedData(enrichedRes, EWorkloadTelemetryGroupBy.ClusterId, true);

          const { categories, series } = this.convertSeriesToECharts(hcSeries);
          this.chartOption = this.buildChartOption(categories, series);
        }

        this.drillStack = [];
        this.drillBreadcrumbs = [];
        this.currentDrillMeta = null;
      } catch (e: unknown) {
        this.handleError(e);
      } finally {
        this.isDataLoading = false;
        this.displayLoading = false;
      }
    },
    async getWorkloadsTelemetry(
      groupByKey: EWorkloadTelemetryGroupByWithoutBucket,
      departmentId?: string,
    ): Promise<TelemetryResponseValuesInner[]> {
      try {
        this.error = false;
        this.isDataLoading = true;
        const res = await workloadService.getWorkloadsTelemetry(
          WorkloadTelemetryType.PendingTimeDistribution,
          this.clusterId,
          this.nodePoolName,
          departmentId,
          undefined,
          [groupByKey],
        );

        this.departmentId = departmentId;
        this.groupBy = groupByKey;

        let telemetryValues: TelemetryResponseValuesInner[] = res.values;
        if (groupByKey === EWorkloadTelemetryGroupBy.ProjectId) {
          telemetryValues = widgetUtil.limitProjectsTelemetryResponse(telemetryValues);
        }
        return telemetryValues;
      } finally {
        this.isDataLoading = false;
      }
    },
    getPreviousGroupByKey(groupByKey: EWorkloadTelemetryGroupBy): EWorkloadTelemetryGroupBy {
      return groupByKey === EWorkloadTelemetryGroupBy.DepartmentId
        ? EWorkloadTelemetryGroupBy.ClusterId
        : EWorkloadTelemetryGroupBy.DepartmentId;
    },
    async getDrillDownSeries(meta: EWorkloadTelemetryPointMeta): Promise<HcSeriesItem[]> {
      try {
        this.isDataLoading = true;
        const { departmentIdFilter, includeDrilldown, groupByKey } = widgetUtil.getWorkloadTelemetryDrilldownOptions(
          meta,
          this.hasDepartments,
        );
        const prevGroupByKey = this.getPreviousGroupByKey(groupByKey);
        const res = await this.getWorkloadsTelemetry(groupByKey, departmentIdFilter);

        const finalIncludeDrilldown = includeDrilldown && !this.isDefaultBucketsData(res);

        return this.aggregateTelemetryGroupedData(
          res,
          groupByKey,
          finalIncludeDrilldown,
          prevGroupByKey,
          departmentIdFilter,
        );
      } finally {
        this.isDataLoading = false;
      }
    },
    async refreshDrilldownData(): Promise<void> {
      if (!this.currentDrillMeta) return;
      try {
        const hcSeries = await this.getDrillDownSeries(this.currentDrillMeta);
        const { categories, series } = this.convertSeriesToECharts(hcSeries, true, this.currentDrillMeta.entityId);
        this.chartOption = this.buildChartOption(categories, series, true);
      } catch (error: unknown) {
        this.handleError(error);
      }
    },
    handleChartClick(event: unknown): void {
      const params = event as {
        componentType?: string;
        seriesIndex?: number;
        dataIndex?: number;
        value?: string;
      };
      if (this.isDataLoading || this.isDefaultBuckets) return;

      if (params.componentType === "xAxis" && params.value) {
        const meta = this.categoryMetaMap.get(params.value);
        if (meta) {
          this.drillDownTo(meta);
        }
        return;
      }

      if (params.seriesIndex === undefined || params.dataIndex === undefined) return;

      const seriesData = this.chartOption?.series as BarSeriesOption[] | undefined;
      if (!seriesData?.[params.seriesIndex]) return;

      const dataArr = seriesData[params.seriesIndex].data as
        | Array<{ value: number; drilldown?: boolean; meta?: EWorkloadTelemetryPointMeta }>
        | undefined;
      const point = dataArr?.[params.dataIndex];
      if (!point?.drilldown || !point?.meta) return;

      this.drillDownTo(point.meta);
    },
    async drillDownTo(meta: EWorkloadTelemetryPointMeta): Promise<void> {
      if (this.isDataLoading) return;

      try {
        this.isDataLoading = true;

        const previousDrillState: DrillState = {
          option: { ...this.chartOption } as EChartsOption,
          meta: this.currentDrillMeta,
          categoryMetaMap: new Map(this.categoryMetaMap),
          drilldownAxisCategories: [...this.drilldownAxisCategories],
        };
        const nextBreadcrumbs = getNextChartDrilldownBreadcrumbs(this.drillBreadcrumbs, this.clusterName);

        const hcSeries = await this.getDrillDownSeries(meta);
        const { categories, series } = this.convertSeriesToECharts(hcSeries, true, meta.entityId);
        this.chartOption = this.buildChartOption(categories, series, true, true);
        this.currentDrillMeta = meta;

        const breadcrumbLabel =
          this.hasDepartments && meta.groupByKey === EWorkloadTelemetryGroupBy.ClusterId ? "Departments" : "Projects";
        const navigation = completeChartDrilldown(
          this.drillStack,
          previousDrillState,
          nextBreadcrumbs,
          breadcrumbLabel,
          categories.length,
        );
        this.drillStack = navigation.stack;
        this.drillBreadcrumbs = navigation.breadcrumbs;
        this.drilldownAnnouncement = navigation.announcement;

        this.error = false;
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.isDataLoading = false;
      }
    },
    drillUpTo(index: number): void {
      const navigation = getChartDrillUpNavigation(this.drillStack, this.drillBreadcrumbs, index);
      if (!navigation) return;

      const { targetState } = navigation;
      this.chartOption = { ...targetState.option, ...getDrilldownTransitionPreset() };
      this.currentDrillMeta = targetState.meta;
      this.categoryMetaMap = new Map(targetState.categoryMetaMap);
      this.drilldownAxisCategories = [...targetState.drilldownAxisCategories];
      this.drillStack = navigation.stack;
      this.drillBreadcrumbs = navigation.breadcrumbs;
      this.drilldownAnnouncement = navigation.announcement;
    },
    drillDownByCategory(category: string): void {
      const meta = this.categoryMetaMap.get(category);
      if (meta) {
        this.drillDownTo(meta);
      }
    },
    aggregateTelemetryGroupedData(
      values: TelemetryResponseValuesInner[],
      groupByKey: EWorkloadTelemetryGroupBy,
      includeDrilldown = false,
      prevGroupByKey?: EWorkloadTelemetryGroupBy,
      prevEntityId?: string,
    ): HcSeriesItem[] {
      this.initializeBucketColors(values);

      const series = widgetUtil.aggregateTelemetryGroupedData(
        values,
        [groupByKey, EWorkloadTelemetryGroupBy.Bucket],
        includeDrilldown,
        prevGroupByKey,
        prevEntityId,
        true,
      ) as unknown as HcSeriesItem[];

      return this.addColorsByBucket(series);
    },
    initializeBucketColors(values: TelemetryResponseValuesInner[]): void {
      const bucketNames = new Set<string>();
      values.forEach((item) => {
        const bucketGroup = item.groups?.find((g) => g.key === EWorkloadTelemetryGroupBy.Bucket);
        if (bucketGroup?.value) {
          bucketNames.add(bucketGroup.value);
        }
      });

      const sortedBuckets = Array.from(bucketNames).sort((a, b) => {
        const extractMin = (bucket: string) => {
          const match = bucket.match(/^(\d+)/);
          return match ? parseInt(match[1], 10) : 0;
        };
        return extractMin(a) - extractMin(b);
      });

      this.bucketColorsMap = chartUtil.assignColorsToKeys<string>(sortedBuckets);
    },
    addColorsByBucket(series: HcSeriesItem[]): HcSeriesItem[] {
      if (!this.bucketColorsMap) {
        return series;
      }

      return series.map((s) => {
        const bucketName = s.name || "";
        const bucketColor = this.bucketColorsMap?.get(bucketName) || chartColorPaletteArray[0];

        const coloredData = s.data.map((dataPoint) => ({
          ...dataPoint,
          color: bucketColor,
        }));

        return {
          ...s,
          color: bucketColor,
          data: coloredData,
        };
      });
    },
    handleError(error: unknown): void {
      this.error = true;
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    redirectToWorkloads(): void {
      this.setColumnFilter();
      this.$router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      });
    },
    getWorkloadsFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        WorkloadSortFilterFields.Phase,
        workloadIndexColumns,
      );
      return filterService.loadFilters(window.location, ETableFilters.WORKLOAD, defaultFilters);
    },
    setColumnFilter(): void {
      const workloadsFilterBy: IFilterBy = this.getWorkloadsFilterBy();
      this.setPhaseColumnFilter(workloadsFilterBy);
      if (this.nodePoolName) {
        this.setNodePoolColumnFilter(workloadsFilterBy);
      } else {
        this.removeNodePoolFilter(workloadsFilterBy);
      }
    },
    setPhaseColumnFilter(workloadsFilterBy: IFilterBy): void {
      const phaseColumnFilter: IFreeTextFilterModel = {
        name: WorkloadSortFilterFields.Phase,
        term: Phase.Pending,
        value: EFilterOperator.Equals,
        label: "Phase",
        type: EColumnFilterType.FreeText,
      };
      filterService.setColumnAdvancedFilter(workloadsFilterBy, phaseColumnFilter, ETableFilters.WORKLOAD);
    },
    setNodePoolColumnFilter(workloadsFilterBy: IFilterBy): void {
      if (this.nodePoolName) {
        const nodePoolColumnFilter: IFreeTextFilterModel = {
          name: WorkloadSortFilterFields.Nodepool,
          term: this.nodePoolName,
          value: EFilterOperator.Equals,
          label: "Node pool",
          type: EColumnFilterType.FreeText,
        };
        filterService.setColumnAdvancedFilter(workloadsFilterBy, nodePoolColumnFilter, ETableFilters.WORKLOAD);
      }
    },
    removeNodePoolFilter(workloadsFilterBy: IFilterBy): void {
      filterService.removeAdvancedFilter(workloadsFilterBy, WorkloadSortFilterFields.Nodepool, ETableFilters.WORKLOAD);
    },
  },
  watch: {
    nodePoolName: {
      handler(): void {
        this.displayLoading = true;
        this.loadInitialChartData();
      },
    },
    clusterId: {
      handler(): void {
        this.displayLoading = true;
        this.loadInitialChartData();
      },
    },
  },
  unmounted() {
    this.stopRefreshLoadData();
  },
});
</script>

<style lang="scss" scoped>
.workload-pending-time-chart-area {
  position: relative;
  height: 240px;
}

.workload-pending-time-chart {
  height: 240px;
}

.workload-pending-time-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.workload-pending-time-axis-label {
  min-width: 0;
  flex: 1 1 0;
  padding: 0;
  overflow: hidden;
  color: var(--text-link-default);
  font-size: 12px;
  line-height: 16px;
  font-weight: 600;
  text-align: center;
  text-decoration: underline;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  background: transparent;
  border: 0;
  pointer-events: auto;
}
</style>
