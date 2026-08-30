<template>
  <chart-widget-wrapper
    class="workload-by-type-widget"
    :options="wrapperOptions"
    @export-csv="exportCsv"
    :loading="displayLoading"
    :error="error"
    :empty="shouldShowEmptyState"
    :announcement="drilldownAnnouncement"
  >
    <template #header="{ titleId }">
      <h3 :id="titleId" class="chart-widget-heading--visually-hidden">{{ wrapperOptions.title }}</h3>
      <widget-header-select-wrapper :gap="10">
        <runai-select
          aid="workload-group-by-select"
          :options="groupByOptions"
          :model-value="secondaryGroupBy"
          @update:model-value="updateSecondaryGroupBy"
        />

        <runai-tooltip-wrapper :tooltip-text="tooltipText" class="q-mt-md">
          <q-icon :name="helpIcon" size="20px" class="text-icon-default"
        /></runai-tooltip-wrapper>

        <template #right>
          <div class="row items-center">
            <widget-link-btn
              class="q-mr-md"
              aid="workload-by-type-redirect-btn"
              :link-text="wrapperOptions.linkText || ''"
              :entity-type="wrapperOptions.entityType"
              @click="redirectToWorkloads"
            />
            <runai-actions-menu round hide-label large-icon :action-items="actionMenuItems" />
          </div>
        </template>
      </widget-header-select-wrapper>
    </template>
    <div v-if="chartOption && !shouldShowMockData" ref="chartAreaRef" class="workload-by-type-chart-area">
      <chart-drilldown-breadcrumbs :items="drillBreadcrumbs" @navigate="drillUpTo" />
      <v-chart
        aid="workload-by-type-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="workload-by-type-chart"
        key="workload-by-type-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="workload-by-type-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="workload-by-type-axis-label"
          type="button"
          :aria-label="`View details for ${category}`"
          @click="drillDownByCategory(category)"
        >
          <span aid="workload-by-type-drilldown-label">{{ category }}</span>
        </button>
      </div>
    </div>
    <workload-by-type-mock-skeleton
      v-show="shouldShowMockData"
      :cluster-name="clusterName"
      key="workload-by-type-mock"
    />
  </chart-widget-wrapper>
</template>
<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, ref, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption, BarSeriesOption } from "echarts";
//routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
//store
import { useClusterStore } from "@/stores/cluster.store";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
//models
import {
  EWidgetEmptyMessages,
  EWidgetEntity,
  type IWidgetWrapperOptions,
  CHART_DATA_LABEL_COLOR,
  ACTIVE_AXIS_LABEL_COLOR,
  DRILLDOWN_CHART_ANIMATION_PRESET,
  getDrilldownTransitionPreset,
  DRILLDOWN_UNIVERSAL_TRANSITION,
} from "@/models/chart.model";
//cmps
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartDrilldownBreadcrumbs } from "@/components/common/charts/chart-drilldown-breadcrumbs";
import WorkloadByTypeMockSkeleton from "./workload-by-type-mock-skeleton.vue";
//service
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import {
  Phase,
  type TelemetryResponseValuesInner,
  WorkloadSortFilterFields,
  WorkloadTelemetryType,
} from "@/swagger-models/workloads-client";
//model
import {
  EWorkloadTelemetryGroupBy,
  type EWorkloadTelemetryGroupByWithoutBucket,
  type EWorkloadTelemetryPointMeta,
  workloadPhaseFilterOptions,
} from "@/models/workload.model";
import { HttpErrorResponse } from "@/models/http-response.model";
//util
import { widgetUtil } from "@/utils/widget.util";
import {
  chartUtil,
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
  STANDARD_TOOLTIP_STYLE,
} from "@/utils/chart.util";
import {
  EColumnFilterType,
  EFilterOperator,
  type IEnumBasedFilterModel,
  type IFilterBy,
  type IFreeTextFilterModel,
} from "@/models/filter.model";
import { dashboardUtil } from "@/utils/dashboard.util";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { workloadIndexColumns } from "@/table-models/workload.table-model";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { RunaiSelect } from "@/components/common/runai-select";
import type { ISelectOption } from "@/models/global.model";
import { WidgetHeaderSelectWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/widget-header-select-wrapper";
import { WidgetLinkBtn } from "@/components/dashboard-v2/widgets/common/buttons/widget-link-btn";
import { RunaiActionsMenu } from "@/components/common/runai-page-actions/runai-actions-menu";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";
import { chartColorPaletteArray, UNKNOWN_COLOR } from "@/models/chart.model";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

export interface IGroupBySelectOption extends ISelectOption {
  value: EWorkloadTelemetryGroupByWithoutBucket;
}

const groupBySelectOptions: IGroupBySelectOption[] = [
  {
    label: "Running workloads by category",
    value: EWorkloadTelemetryGroupBy.Category,
  },
  {
    label: "Running workloads by type",
    value: EWorkloadTelemetryGroupBy.Type,
  },
];

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

interface HcDataPoint {
  name: string;
  y: number;
  drilldown?: boolean;
  meta?: { entityId: string; groupByKey: string };
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

export default defineComponent({
  name: "workload-by-type-widget",
  components: {
    RunaiTooltipWrapper,
    RunaiActionsMenu,
    WidgetLinkBtn,
    WidgetHeaderSelectWrapper,
    RunaiSelect,
    ChartDrilldownBreadcrumbs,
    ChartWidgetWrapper,
    VChart,
    WorkloadByTypeMockSkeleton,
  },
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
      helpIcon: toQuasarIconName(runaiIcons.object.help),
      isDataLoading: false as boolean,
      displayLoading: true as boolean,
      chartOption: null as null | EChartsOption,
      clusterStore: useClusterStore(),
      guidedTourStore: useGuidedTourStore(),
      error: false as boolean,
      isChartEmpty: false as boolean,
      wrapperOptions: {
        title: "Workloads by type",
        linkText: "All workloads",
        entityType: EWidgetEntity.Workload,
        emptyMessage: EWidgetEmptyMessages.NoRunningWorkloads,
      } as IWidgetWrapperOptions,
      groupBy: EWorkloadTelemetryGroupBy.ClusterId as EWorkloadTelemetryGroupBy,
      secondaryGroupBy: EWorkloadTelemetryGroupBy.Category as EWorkloadTelemetryGroupByWithoutBucket,
      actionMenuItems: [
        {
          label: "Download as CSV",
          action: () => this.exportCsv(),
        },
      ],
      colorMap: new Map<string, string>(),
      nextColorIndex: 0,
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
  created() {
    this.loadInitialChartData();
    this.startRefreshLoadData();
  },
  computed: {
    clusterName(): string {
      return this.clusterStore.clusterNameByUuid(this.clusterId) || "";
    },
    groupByOptions(): IGroupBySelectOption[] {
      return groupBySelectOptions;
    },
    tooltipText(): string {
      return this.isGroupByCategory
        ? "Distribution of workloads by category to help monitor usage trends across the cluster"
        : "Distribution of workloads by type to help monitor usage trends across the cluster";
    },
    isGroupByCategory(): boolean {
      return this.secondaryGroupBy === EWorkloadTelemetryGroupBy.Category;
    },
    shouldShowMockData(): boolean {
      return this.guidedTourStore.isOverviewTourActive && this.isChartEmpty;
    },
    shouldShowEmptyState(): boolean {
      if (this.guidedTourStore.isOverviewTourActive) {
        return false;
      }
      return this.isChartEmpty;
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.WorkloadByTypeWidget, this.refreshChart);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadByTypeWidget);
    },
    async refreshChart(): Promise<void> {
      if (this.isDataLoading) return;

      if (this.currentDrillMeta) {
        await this.refreshDrilldownData();
      } else {
        await this.loadInitialChartData();
      }
    },
    getColor(type: string): string {
      if (!type) {
        return UNKNOWN_COLOR;
      }
      if (this.colorMap.has(type)) {
        return this.colorMap.get(type) as string;
      }
      const color = chartColorPaletteArray[this.nextColorIndex % chartColorPaletteArray.length];
      this.colorMap.set(type, color);
      this.nextColorIndex++;
      return color;
    },
    assignColorsToSeries(series: HcSeriesItem[]): HcSeriesItem[] {
      const sortedSeries = [...series].sort((a, b) => {
        const nameA = a.name || "";
        const nameB = b.name || "";
        return nameA.localeCompare(nameB);
      });
      return sortedSeries.map((s: HcSeriesItem) => {
        const type: string | undefined = s.name;
        if (type) {
          s.color = this.getColor(type);
        }
        return s;
      });
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
          id: `wbt-${seriesName}`,
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
                }
              : { value: 0, groupId: parentGroupId ?? cat };
          }),
          universalTransition: DRILLDOWN_UNIVERSAL_TRANSITION,
          itemStyle: s.color ? { color: s.color } : undefined,
          ...chartUtil.getBarSeriesDefaults(),
          label: {
            show: true,
            position: "top" as const,
            formatter: (params: unknown) => {
              const value = (params as { value?: number }).value ?? 0;
              return value === 0 ? "" : String(value);
            },
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
      const yAxisConfig = chartUtil.getBarChartYAxisConfig(series, { integerTicks: true });
      const groupedBarLayout = this.getGroupedBarLayout(categories.length, series.length);
      return {
        ...DRILLDOWN_CHART_ANIMATION_PRESET,
        ...(animateDrilldown ? getDrilldownTransitionPreset() : {}),
        aria: chartUtil.buildChartAria("Workloads by type. Select a category to view its details."),
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
    async loadInitialChartData(): Promise<void> {
      try {
        this.isDataLoading = true;
        const res: TelemetryResponseValuesInner[] = await this.getWorkloadsTelemetry(
          EWorkloadTelemetryGroupBy.ClusterId,
        );

        const combinedTelemetry = widgetUtil.combineDistributedTelemetryIntoTraining(res);
        const enrichedRes: TelemetryResponseValuesInner[] = widgetUtil.enrichTelemetryResponseByKeyValue(
          combinedTelemetry,
          EWorkloadTelemetryGroupBy.ClusterId,
          this.clusterName,
          "name",
        );

        const hcSeries = this.aggregateTelemetryGroupedData(enrichedRes, EWorkloadTelemetryGroupBy.ClusterId, true);

        if (hcSeries.length === 0) {
          this.isChartEmpty = true;
        } else {
          this.isChartEmpty = false;
          const { categories, series } = this.convertSeriesToECharts(hcSeries);
          this.chartOption = this.buildChartOption(categories, series);
        }

        this.error = false;
        this.groupBy = EWorkloadTelemetryGroupBy.ClusterId;
        this.drillStack = [];
        this.drillBreadcrumbs = [];
        this.currentDrillMeta = null;
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.isDataLoading = false;
        this.displayLoading = false;
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
      if (this.isDataLoading || this.shouldShowMockData) return;

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
    drillDownByCategory(category: string): void {
      const meta = this.categoryMetaMap.get(category);
      if (meta) {
        this.drillDownTo(meta);
      }
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
    formatBreadcrumbs(level: number): string {
      if (level === 0) return this.clusterName;
      if (level === 1 && this.hasDepartments) {
        return "Departments";
      }
      return "Projects";
    },
    async getWorkloadsTelemetry(
      groupByKey: EWorkloadTelemetryGroupByWithoutBucket,
      departmentId?: string,
    ): Promise<TelemetryResponseValuesInner[]> {
      const res = await workloadService.getWorkloadsTelemetry(
        WorkloadTelemetryType.WorkloadsCount,
        this.clusterId,
        this.nodePoolName,
        departmentId,
        undefined,
        [groupByKey, this.secondaryGroupBy],
        [Phase.Running],
      );

      return res.values;
    },
    aggregateTelemetryGroupedData(
      values: TelemetryResponseValuesInner[],
      groupByKey: EWorkloadTelemetryGroupBy,
      includeDrilldown = false,
      prevGroupByKey?: EWorkloadTelemetryGroupBy,
      prevEntityId?: string,
    ): HcSeriesItem[] {
      const series = widgetUtil.aggregateTelemetryGroupedData(
        values,
        [groupByKey, this.secondaryGroupBy],
        includeDrilldown,
        prevGroupByKey,
        prevEntityId,
      ) as unknown as HcSeriesItem[];
      return this.assignColorsToSeries(series);
    },
    async exportCsv() {
      try {
        await workloadService.getWorkloadsTelemetryCsv(
          WorkloadTelemetryType.WorkloadsCount,
          this.clusterId,
          this.nodePoolName,
          undefined,
          [this.groupBy, this.secondaryGroupBy],
        );
      } catch (error: unknown) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        this.handleError(error);
      }
    },
    updateSecondaryGroupBy(newValue: IGroupBySelectOption): void {
      this.secondaryGroupBy = newValue.value;

      if (this.shouldShowMockData) {
        return;
      }

      this.drillStack = [];
      this.drillBreadcrumbs = [];
      this.currentDrillMeta = null;
      this.refreshChart();
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
        const combinedTelemetry = widgetUtil.combineDistributedTelemetryIntoTraining(res);

        return this.aggregateTelemetryGroupedData(
          combinedTelemetry,
          groupByKey,
          includeDrilldown,
          prevGroupByKey,
          departmentIdFilter,
        );
      } finally {
        this.isDataLoading = false;
      }
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
      this.setRunningColumnFilter(workloadsFilterBy);
      if (this.nodePoolName) {
        this.setNodePoolColumnFilter(workloadsFilterBy);
      } else {
        this.removeNodePoolFilter(workloadsFilterBy);
      }
    },
    setRunningColumnFilter(workloadsFilterBy: IFilterBy): void {
      const runningColumnFilter: IEnumBasedFilterModel = {
        name: WorkloadSortFilterFields.Phase,
        selectedValues: [Phase.Running],
        type: EColumnFilterType.EnumBased,
        label: "Status",
        selectOptions: workloadPhaseFilterOptions,
      };
      filterService.setColumnAdvancedFilter(workloadsFilterBy, runningColumnFilter, ETableFilters.WORKLOAD);
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
.workload-by-type-chart-area {
  position: relative;
  height: 240px;
}

.workload-by-type-chart {
  height: 240px;
}

.workload-by-type-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.workload-by-type-axis-label {
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
