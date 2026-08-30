<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    @link-clicked="redirectToWorkloads"
    @export-csv="exportCsv"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty"
    :announcement="drilldownAnnouncement"
  >
    <div v-if="chartOption" ref="chartAreaRef" class="workload-status-chart-area">
      <chart-drilldown-breadcrumbs :items="drillBreadcrumbs" @navigate="drillUpTo" />
      <v-chart
        aid="workload-status-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="workload-status-chart"
        key="workload-status-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="workload-status-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="workload-status-axis-label"
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
//routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
//store
import { useClusterStore } from "@/stores/cluster.store";
//models
import {
  EWidgetEntity,
  type IWidgetWrapperOptions,
  CHART_COLOR_PALETTE,
  UNKNOWN_COLOR,
  CHART_DATA_LABEL_COLOR,
  ACTIVE_AXIS_LABEL_COLOR,
  DRILLDOWN_CHART_ANIMATION_PRESET,
  getDrilldownTransitionPreset,
  DRILLDOWN_UNIVERSAL_TRANSITION,
} from "@/models/chart.model";
//cmps
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartDrilldownBreadcrumbs } from "@/components/common/charts/chart-drilldown-breadcrumbs";
//service
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { filterService } from "@/services/filter.service/filter.service";
import {
  Phase,
  type TelemetryResponseValuesInner,
  type TelemetryResponseValuesInnerGroupsInner,
  WorkloadSortFilterFields,
  WorkloadTelemetryType,
} from "@/swagger-models/workloads-client";
//model
import {
  EWorkloadTelemetryGroupBy,
  type EWorkloadTelemetryGroupByWithoutBucket,
  type EWorkloadTelemetryPointMeta,
} from "@/models/workload.model";
import { workloadIndexColumns } from "@/table-models/workload.table-model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { EColumnFilterType, EFilterOperator, type IFilterBy, type IFreeTextFilterModel } from "@/models/filter.model";
import { ETableFilters } from "@/models/table.model";
//util
import { widgetUtil } from "@/utils/widget.util";
import {
  chartUtil,
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
  STANDARD_TOOLTIP_STYLE,
} from "@/utils/chart.util";
import { dashboardUtil } from "@/utils/dashboard.util";
import { EIntervalLabels } from "@/models/interval.model";
import { intervalUtil } from "@/utils/interval.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

const DISPLAYED_PHASES: string[] = [Phase.Failed, Phase.Pending, Phase.Running];
const DRILLDOWN_DISPLAYED_PHASES: string[] = [Phase.Running, Phase.Failed, Phase.Pending];
const colorsMap = new Map<string, string>([
  [Phase.Failed, CHART_COLOR_PALETTE.BLUE],
  [Phase.Running, CHART_COLOR_PALETTE.LIME],
  [Phase.Pending, CHART_COLOR_PALETTE.ORANGE],
]);

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
  data: DISPLAYED_PHASES,
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
  name: "workloads-status-widget",
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
      wrapperOptions: {
        title: "Workloads by status",
        linkText: "All workloads",
        tooltipText:
          "Workload statuses represent the <br/> workload's life cycle and indicate the <br/>  health of the department or project",
        entityType: EWidgetEntity.Workload,
      } as IWidgetWrapperOptions,
      departmentId: undefined as string | undefined,
      groupBy: undefined as EWorkloadTelemetryGroupBy | undefined,
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
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.WorkloadStatusWidget, this.refreshChart);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadStatusWidget);
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
          id: `wstatus-${seriesName}`,
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
            formatter: (params: unknown) => String((params as { value?: number }).value ?? 0),
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
      const yAxisConfig = this.getYAxisConfig(series, isDrilldown);
      const groupedBarLayout = this.getGroupedBarLayout(categories.length, series.length);
      return {
        ...DRILLDOWN_CHART_ANIMATION_PRESET,
        ...(animateDrilldown ? getDrilldownTransitionPreset() : {}),
        aria: chartUtil.buildChartAria("Workload status over time. Select a category to view its details."),
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
        legend: {
          ...legendOptions,
          data: isDrilldown ? DRILLDOWN_DISPLAYED_PHASES : DISPLAYED_PHASES,
        },
        series: series.map((item) => ({ ...item, ...groupedBarLayout })),
      };
    },
    getYAxisConfig(series: BarSeriesOption[], isDrilldown: boolean): { interval: number; max: number } {
      const yAxisConfig = chartUtil.getBarChartYAxisConfig(series, { integerTicks: true });

      if (!isDrilldown) {
        return yAxisConfig;
      }

      const dataMax = Math.max(
        0,
        ...series.flatMap((s) =>
          ((s.data || []) as Array<number | { value?: number }>).map((point) =>
            typeof point === "number" ? point : (point.value ?? 0),
          ),
        ),
      );
      if (dataMax >= 2 && yAxisConfig.interval < 1) {
        return { interval: 1, max: Math.ceil(dataMax) + 1 };
      }
      return yAxisConfig;
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
        } else {
          this.isChartEmpty = false;

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
    async exportCsv(): Promise<void> {
      try {
        await workloadService.getWorkloadsTelemetryCsv(
          WorkloadTelemetryType.WorkloadsCount,
          this.clusterId,
          this.nodePoolName,
          this.departmentId,
          this.groupBy ? [this.groupBy, EWorkloadTelemetryGroupBy.Phase] : [EWorkloadTelemetryGroupBy.Phase],
        );
      } catch (e: unknown) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        this.handleError(e);
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

        return this.aggregateTelemetryGroupedData(res, groupByKey, includeDrilldown, prevGroupByKey, departmentIdFilter);
      } finally {
        this.isDataLoading = false;
      }
    },
    async refreshDrilldownData(): Promise<void> {
      if (!this.currentDrillMeta) return;
      try {
        const hcSeries = this.sortSeriesByPhase(await this.getDrillDownSeries(this.currentDrillMeta), true);
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
      if (this.isDataLoading) return;

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

        const hcSeries = this.sortSeriesByPhase(await this.getDrillDownSeries(meta), true);
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
    async getWorkloadsTelemetry(
      groupByKey: EWorkloadTelemetryGroupByWithoutBucket,
      departmentId?: string,
    ): Promise<TelemetryResponseValuesInner[]> {
      try {
        this.error = false;
        this.isDataLoading = true;
        const res = await workloadService.getWorkloadsTelemetry(
          WorkloadTelemetryType.WorkloadsCount,
          this.clusterId,
          this.nodePoolName,
          departmentId,
          undefined,
          [groupByKey, EWorkloadTelemetryGroupBy.Phase],
        );
        let telemetryValues: TelemetryResponseValuesInner[] = res.values;
        this.departmentId = departmentId;
        this.groupBy = groupByKey;
        if (groupByKey === EWorkloadTelemetryGroupBy.ProjectId) {
          telemetryValues = widgetUtil.limitProjectsTelemetryResponse(res.values);
        }
        return this.filterDisplayedPhases(telemetryValues);
      } finally {
        this.isDataLoading = false;
      }
    },
    filterDisplayedPhases(values: TelemetryResponseValuesInner[]): TelemetryResponseValuesInner[] {
      return values.filter((telemetryResponse: TelemetryResponseValuesInner) => {
        const phaseGroup = telemetryResponse?.groups?.find(
          (group: TelemetryResponseValuesInnerGroupsInner) => group.key === EWorkloadTelemetryGroupBy.Phase,
        );
        const phaseValue: string | undefined = phaseGroup?.value;

        return phaseValue && DISPLAYED_PHASES.includes(phaseValue);
      });
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
        [groupByKey, EWorkloadTelemetryGroupBy.Phase],
        includeDrilldown,
        prevGroupByKey,
        prevEntityId,
      ) as unknown as HcSeriesItem[];

      const extendedSeries = this.extendSeriesWithZeroValuePhases(series);
      return this.sortSeriesByPhase(this.addColorsByPhase(extendedSeries), false);
    },
    sortSeriesByPhase(series: HcSeriesItem[], isDrilldown: boolean): HcSeriesItem[] {
      const phaseOrder = isDrilldown ? DRILLDOWN_DISPLAYED_PHASES : DISPLAYED_PHASES;
      return series.sort((a, b) => phaseOrder.indexOf(a.name || "") - phaseOrder.indexOf(b.name || ""));
    },
    extendSeriesWithZeroValuePhases(series: HcSeriesItem[]): HcSeriesItem[] {
      const existingPhases = new Set(series.map((s) => s.name).filter(Boolean));
      const missingPhases = DISPLAYED_PHASES.filter((phase) => !existingPhases.has(phase));

      if (missingPhases.length === 0) {
        return series;
      }

      const zeroValueSeries: HcSeriesItem[] = missingPhases.map((phase) => {
        const templateSeries = series[0];
        const templateData = templateSeries?.data || [];

        const zeroData: HcDataPoint[] = templateData.map((dataPoint) => ({
          ...dataPoint,
          y: 0,
        }));

        return {
          name: phase,
          data: zeroData,
          color: colorsMap.get(phase) || UNKNOWN_COLOR,
        };
      });

      return [...series, ...zeroValueSeries];
    },
    addColorsByPhase(series: HcSeriesItem[]): HcSeriesItem[] {
      return series.map((s: HcSeriesItem) => {
        const phase: string | undefined = s.name;
        if (!phase) return s;
        if (colorsMap.has(phase)) {
          s.color = colorsMap.get(phase);
        } else {
          s.color = UNKNOWN_COLOR;
        }
        return s;
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
      if (this.nodePoolName) {
        this.setNodePoolColumnFilter(workloadsFilterBy);
      } else {
        this.removeNodePoolFilter(workloadsFilterBy);
      }
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
.workload-status-chart-area {
  position: relative;
  height: 240px;
}

.workload-status-chart {
  height: 240px;
}

.workload-status-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.workload-status-axis-label {
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
