<template>
  <generic-metric-widget
    :config="widgetConfig"
    :filter-by-dates="filterByDates"
    :get-data-args="getDataArgs"
    @register="emit('register')"
    @unregister="emit('unregister')"
    @on-zoom="emit('on-zoom', $event)"
    @on-reset-zoom="emit('on-reset-zoom', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import GenericMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/generic-metric-widget/generic-metric-widget.vue";
import type { IGenericMetricWidgetConfig, IZoomEvent } from "@/models/chart.model";
import type { IRangeDates } from "@/utils/date.util";

const props = defineProps<{
  workloadConfig: IGenericMetricWidgetConfig;
  podConfig: IGenericMetricWidgetConfig;
  workloadId: string;
  podId?: string;
  filterByDates: IRangeDates;
}>();

const emit = defineEmits<{
  register: [];
  unregister: [];
  "on-zoom": [event: IZoomEvent];
  "on-reset-zoom": [chartId: string];
}>();

const widgetConfig = computed<IGenericMetricWidgetConfig>(() => (props.podId ? props.podConfig : props.workloadConfig));
const getDataArgs = computed<string[]>(() => {
  return props.podId ? [props.workloadId, props.podId] : [props.workloadId];
});
</script>
