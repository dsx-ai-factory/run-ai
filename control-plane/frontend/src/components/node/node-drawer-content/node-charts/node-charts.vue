<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="node-charts">
    <div class="node-charts-header">
      <q-select
        class="metrics-select"
        dense
        label="Type"
        v-if="isAdvancedGpuMetricsEnabled"
        outlined
        v-model="selectedMetricType"
        :options="metricOptions"
      />
      <div v-else class="q-mr-auto"></div>
      <runai-datepicker-with-predefined
        @changed="onDateChanged"
        :range-of-dates="rangeOfDatesCallback"
        info-bar-message="Select a period within the last 90 days"
        :predefined-options="hoursGranularityPredefinedOptions"
        :default-option-index="0"
      />
    </div>

    <div class="node-charts-body">
      <div class="node-charts-content q-pt-md" :key="chartsContentKey" @scroll.passive="hideChartTooltips">
        <generic-metric-widget
          v-for="widget in filteredWidgets"
          :key="widget.key"
          :config="widget.config"
          :entity-id="node?.id || ''"
          :filter-by-dates="filterByDates"
          :get-data-args="[node?.id as string]"
          confine-tooltip
          show-area-zoom
          :area-zoom-vertical-offset="10"
          :zoom-active="zoomActive"
          @register="registerChart(widget.config.chartId)"
          @unregister="unregisterChart(widget.config.chartId)"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
          @on-zoom-toggle="handleZoomToggle"
        />
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType, ref, computed, watch } from "vue";
import { defaultWidgets, advancedWidgets } from "./widget-config-mapper";
import type { Node } from "@/swagger-models/cluster-service-client";
import GenericMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/generic-metric-widget/generic-metric-widget.vue";
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import { dateUtil, hoursGranularityPredefinedOptions, type IRangeDates, TimeUnit } from "@/utils/date.util";
import {
  useConnectedCharts,
  type IConnectedChartsOptions,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";
import type { IZoomEvent } from "@/models/chart.model";
import type { ISelectOption } from "@/models/global.model";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { chartUtil } from "@/utils/chart.util";

export enum EMetricType {
  Default = "Resource utilization",
  Advanced = "GPU profiling",
}

const metricOptions = [
  { label: EMetricType.Default, value: EMetricType.Default },
  { label: EMetricType.Advanced, value: EMetricType.Advanced },
];

export default defineComponent({
  name: "node-charts",
  components: { RunaiDatepickerWithPredefined, GenericMetricWidget },
  props: {
    clusterId: {
      type: String,
      required: true,
    },
    node: {
      type: Object as PropType<Node>,
      required: false,
    },
  },
  setup(props) {
    const clusterStore = useClusterStore();
    const filterByDates = ref<IRangeDates>({
      dateStart: dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -1) as Date,
      dateEnd: new Date() as Date,
      label: "",
    });
    const selectedMetricType = ref<ISelectOption>(metricOptions[0]);
    const zoomActive = ref(false);

    const rangeOfDatesCallback = (date: string): boolean => {
      const dateToCheck = new Date(date);
      if (Number.isNaN(dateToCheck.getTime())) return false;

      dateToCheck.setHours(0, 0, 0, 0);
      const back90Days = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -90);
      back90Days.setHours(0, 0, 0, 0);

      const today = new Date();
      today.setHours(23, 59, 59, 999);

      return dateToCheck >= back90Days && dateToCheck <= today;
    };

    const onDateChanged = (dateRange: Array<Date>, label: string): void => {
      filterByDates.value = {
        dateStart: dateRange[0],
        dateEnd: dateRange[1],
        label: label,
      };
    };

    const handleZoomToggle = (active: boolean): void => {
      zoomActive.value = active;
    };

    const hideChartTooltips = (event: Event): void => {
      chartUtil.hideTooltipsWithin(event.currentTarget as HTMLElement | null);
    };

    const isWidgetSupported = (widget: { config: { minSupportedClusterVersion?: string } }): boolean => {
      if (!widget.config.minSupportedClusterVersion) {
        return true;
      }
      return clusterStore.isClusterVersionSufficient(props.clusterId, widget.config.minSupportedClusterVersion);
    };

    const filteredWidgets = computed(() => {
      const widgets = selectedMetricType.value.value === EMetricType.Default ? defaultWidgets : advancedWidgets;
      return [...widgets].filter(isWidgetSupported);
    });

    // Get chartIds from the currently selected and filtered widgets
    const initialChartIds = filteredWidgets.value.map((widget) => widget.config.chartId);
    const options: IConnectedChartsOptions = { chartIds: initialChartIds };
    const { registerChartToCrosshair, unregisterChartFromCrosshair, zoomIn, resetZoom } = useConnectedCharts(options);

    const registerChart = (chartId: string): void => {
      registerChartToCrosshair(chartId);
    };

    const unregisterChart = (chartId: string): void => {
      unregisterChartFromCrosshair(chartId);
    };

    const handleZoomIn = (event: IZoomEvent, currentChartId?: string): void => {
      zoomIn(event.min, event.max, currentChartId || event.chartId);
    };

    const handleResetZoom = (currentChartId: string | IZoomEvent): void => {
      const chartId = typeof currentChartId === "string" ? currentChartId : currentChartId.chartId;
      zoomActive.value = false;
      resetZoom(chartId);
    };

    watch([selectedMetricType, filterByDates], () => {
      zoomActive.value = false;
    });

    const isAdvancedGpuMetricsEnabled = computed(() => useSettingStore().isAdvancedGpuMetricsEnabled);

    const chartsContentKey = computed(() => {
      // This key will change whenever user changes filters, forcing charts to remount
      // which will trigger loading state via isFirstLoad flag
      return `${selectedMetricType.value.value}-${filterByDates.value.label}-${filterByDates.value.dateStart.getTime()}-${filterByDates.value.dateEnd.getTime()}`;
    });

    return {
      filteredWidgets,
      filterByDates,
      onDateChanged,
      rangeOfDatesCallback,
      hoursGranularityPredefinedOptions,
      registerChart,
      unregisterChart,
      handleZoomIn,
      handleResetZoom,
      handleZoomToggle,
      hideChartTooltips,
      zoomActive,
      selectedMetricType,
      metricOptions,
      isAdvancedGpuMetricsEnabled,
      chartsContentKey,
    };
  },
});
</script>

<style scoped lang="scss">
.node-charts {
  height: 100%;
  display: flex;
  flex-direction: column;

  .node-charts-header {
    border-bottom: solid 1px $black-20;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    padding: 0 16px;

    .metric-title {
      font-weight: 500;
    }
  }

  .node-charts-body {
    flex: 1;
    overflow: hidden;

    .node-charts-content {
      overflow: auto;
      height: 100%;
    }
  }
  .metrics-select {
    height: 42px;
    min-width: 180px;
  }
}
</style>
