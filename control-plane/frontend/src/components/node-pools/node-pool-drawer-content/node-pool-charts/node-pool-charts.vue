<template>
  <section class="node-pool-charts">
    <section class="node-pool-charts-header">
      <runai-datepicker-with-predefined
        @changed="onDateChanged"
        :range-of-dates="rangeOfDatesCallback"
        info-bar-message="Select a period within the last 90 days"
        :predefined-options="predefinedOptions"
        :default-option-index="0"
      />
    </section>

    <section class="node-pool-charts-body">
      <section class="node-pool-charts-content q-pt-md" @scroll.passive="hideChartTooltips">
        <gpu-allocation-time-range-widget
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :filter-by="nodePoolFilter"
          confine-tooltip
          @register="registerChart(chartIds[0])"
          @unregister="unregisterChart(chartIds[0])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
        />
        <gpu-utilization-distribution-time-range-widget
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :filter-by="nodePoolFilter"
          confine-tooltip
          @register="registerChart(chartIds[1])"
          @unregister="unregisterChart(chartIds[1])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
        />
        <resource-utilization-time-range-widget
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :node-pool-name="selectedNodePool.name"
          :filter-by="nodePoolFilter"
          confine-tooltip
          hide-average-utilization-stats
          hide-border
          flat
          hide-time-frame
          hide-actions
          hide-separator
          @register="registerChart(chartIds[2])"
          @unregister="unregisterChart(chartIds[2])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
        />
      </section>
    </section>
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

// models
import type { PredefinedOption } from "@/models/date.model";
import type { IFilterModel } from "@/models/filter.model";
//utils
import { hoursGranularityPredefinedOptions, type IRangeDates } from "@/utils/date.util";
import { dateUtil, TimeUnit } from "@/utils/date.util";

// cmps
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import { ResourceUtilizationTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/resource-utilization-time-range-widget/";
import { GpuUtilizationDistributionTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/gpu-utilization-distribution-time-range-widget/";
// services
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { GpuAllocationTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/gpu-allocation-time-range-widget/";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";
import {
  type IZoomEvent,
  GPU_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
  GPU_UTILIZATION_DISTRIBUTION_TIME_RANGE_WIDGET_CHART_ID,
  RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID,
} from "@/models/chart.model";
import { chartUtil } from "@/utils/chart.util";

export default defineComponent({
  name: "node-pool-charts",
  components: {
    GpuAllocationTimeRangeWidget,
    GpuUtilizationDistributionTimeRangeWidget,
    ResourceUtilizationTimeRangeWidget,
    RunaiDatepickerWithPredefined,
  },
  props: {
    selectedNodePool: {
      type: Object as PropType<Nodepool>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      predefinedOptions: hoursGranularityPredefinedOptions as Array<PredefinedOption>,
      filterByDates: {
        dateStart: dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -1) as Date,
        dateEnd: new Date() as Date,
      } as IRangeDates,
    };
  },
  setup() {
    const chartIds: string[] = [
      GPU_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
      GPU_UTILIZATION_DISTRIBUTION_TIME_RANGE_WIDGET_CHART_ID,
      RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID,
    ];
    const options: IConnectedChartsOptions = {
      chartIds,
    };
    const { zoomIn, resetZoom, registerChartToCrosshair, unregisterChartFromCrosshair } = useConnectedCharts(options);
    return {
      zoomIn,
      resetZoom,
      registerChartToCrosshair,
      unregisterChartFromCrosshair,
      chartIds,
    };
  },
  computed: {
    nodePoolFilter(): IFilterModel[] {
      return [
        {
          field: "nodepool",
          term: this.selectedNodePool.name,
          name: "nodepool",
        },
      ];
    },
  },
  methods: {
    onDateChanged(dateRange: Array<Date>, label: string): void {
      this.filterByDates = {
        dateStart: dateRange[0],
        dateEnd: dateRange[1],
        label: label,
      };
    },
    rangeOfDatesCallback(date: string): boolean {
      const dateToCheck = new Date(date);
      if (Number.isNaN(dateToCheck.getTime())) return false;

      dateToCheck.setHours(0, 0, 0, 0);
      const back90Days = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -90);
      back90Days.setHours(0, 0, 0, 0);

      const today = new Date();
      today.setHours(23, 59, 59, 999);

      return dateToCheck >= back90Days && dateToCheck <= today;
    },
    registerChart(chartId: string): void {
      this.registerChartToCrosshair(chartId);
    },
    unregisterChart(chartId: string): void {
      this.unregisterChartFromCrosshair(chartId);
    },
    handleZoomIn(event: IZoomEvent, currentChartId?: string): void {
      this.zoomIn(event.min, event.max, currentChartId || event.chartId);
    },
    handleResetZoom(currentChartId: string | IZoomEvent): void {
      const chartId = typeof currentChartId === "string" ? currentChartId : currentChartId.chartId;
      this.resetZoom(chartId);
    },
    hideChartTooltips(event: Event): void {
      chartUtil.hideTooltipsWithin(event.currentTarget as HTMLElement | null);
    },
  },
});
</script>

<style lang="scss" scoped>
.node-pool-charts {
  height: 100%;
  display: flex;
  flex-direction: column;

  .node-pool-charts-header {
    border-bottom: solid 1px $black-20;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 56px;
    padding: 0 16px;
  }

  .node-pool-charts-body {
    flex: 1;
    overflow: hidden;

    .node-pool-charts-content {
      overflow: auto;
      height: 100%;
    }
  }
}
</style>
