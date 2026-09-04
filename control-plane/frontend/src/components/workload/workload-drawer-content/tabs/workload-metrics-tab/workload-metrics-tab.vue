<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-metrics-tab q-pb-xl">
    <section
      ref="workload-metrics-header"
      class="workload-metrics-header row items-center q-gutter-x-sm q-py-sm q-px-md"
      :class="{ 'is-grid': showHeaderAsGrid }"
    >
      <div class="workload-metrics-header-item row col no-wrap items-center">
        <runai-select
          class="metric-select metric-type-select q-mr-sm"
          aid="metric-type-select"
          outlined
          dense
          label="Type"
          options-dense
          hide-bottom-space
          :options="metricTypeOptions"
          :model-value="selectedMetricType"
          @update:model-value="selectedMetricType = $event"
        />
        <runai-select
          v-if="isSupportingSelectionPods"
          class="metric-select metric-pod-select"
          aid="metric-pod-select"
          outlined
          dense
          selected-ellipsis
          label="Pod"
          options-dense
          hide-bottom-space
          :options="podsOptions"
          :model-value="selectedPodName"
          @update:model-value="selectedPodName = $event"
        />
      </div>
      <runai-datepicker-with-predefined
        :key="workloadId"
        class="workload-metrics-header-item"
        aid="workload-metrics-datepicker"
        @changed="updateDateRange"
        :range-of-dates="rangeOfDatesCallback"
        info-bar-message="Select a period within the last 90 days"
        :predefined-options="predefinedOptions"
        :default-option-index="0"
        :default-start-date="defaultStartDate"
        :disable-before-date="workloadCreationTime"
        :workload-lifespan-end-date="workloadLifespanEndDate"
        :workload-first-run-date="workloadFirstRunDate"
      />
    </section>
    <div class="charts-container q-py-md" :key="chartsContainerKey">
      <default-workload-widgets
        v-if="isDefaultMetricType"
        :workload="workload"
        :selected-pod-id="selectedPodId"
        :selected-element-id="selectedElementId"
        :filter-by-dates="filterByDates"
        :reload-trigger="userActionTimestamp"
        silent-date-reload
        :ref="defaultWorkloadWidgetsRef"
      />

      <latency-throughput-and-replicas-widget
        v-else-if="clusterVersion && isInferenceDefaultMetricType"
        :cluster-version="clusterVersion"
        :ref="inferenceWidgetsRef"
        :workload-id="workloadId"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
      />
      <advanced-workload-metric-widget
        v-if="isAdvancedMetricType && selectedPodId"
        :workload-id="workloadId"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
        confine-tooltip
        :ref="advancedWorkloadWidgetsRef"
      />
      <nim-workload-widgets
        v-else-if="isNimMetricType"
        :ref="nimWidgetsRef"
        :workload="workload"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
      />
      <vllm-workload-widgets
        v-else-if="isVllmMetricType"
        :ref="vllmWidgetsRef"
        :workload="workload"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Models
import { type PredefinedOption } from "@/models/date.model";
import { type WorkloadServicePod, EWorkloadType } from "@/models/workload.model";

// Utils
import { dateUtil, minutesGranularityPredefinedOptions, TimeUnit } from "@/utils/date.util";
import {
  hoursGranularityPredefinedOptions,
  workloadsLifespanPredefinedOptions,
  sinceFirstRunPredefinedOptions,
} from "@/utils/date.util";

// Components
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import { RunaiSelect } from "@/components/common/runai-select";
import { debounce } from "quasar";
import type { IRangeDates } from "@/utils/date.util";
import { Phase, HistoryRecordType, type Workload } from "@/swagger-models/workloads-client";

// widgets
import { LatencyThroughputAndReplicasWidget } from "@/components/dashboard-v2/widgets/chart-widgets/latency-throughput-and-replicas-widget";
import { EMetricsType } from "@/models/node.model";
import AdvancedWorkloadMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/advanced-metric/advanced-workload-metric-widget/advanced-workload-metric-widget.vue";
import DefaultWorkloadWidgets from "./default-workload-widgets.vue";
import NimWorkloadWidgets from "./nim-workload-widgets.vue";
import VllmWorkloadWidgets from "./vllm-workload-widgets.vue";

// stores
import { useSettingStore } from "@/stores/setting.store";

// constants
import { useClusterStore } from "@/stores/cluster.store";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { ErrorAlert } from "@/utils/error-alert.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { MIN_CLUSTER_VERSION_FOR_NIM_METRICS, MIN_CLUSTER_VERSION_FOR_VLLM_METRICS } from "@/common/version.constant";
import { workloadUtil } from "@/utils/workload.util";
import { useWorkloadStore } from "@/stores/workload.store";
import { filterUtil } from "@/utils/filter.util/filter.util";

export type TWorkloadCharts = InstanceType<typeof DefaultWorkloadWidgets | typeof LatencyThroughputAndReplicasWidget>;
const ALL_PODS_OPTION = "All pods";
const MAX_WIDTH_FOR_GRID = 490;
const predefinedOptions: Array<PredefinedOption> = [
  ...workloadsLifespanPredefinedOptions,
  ...sinceFirstRunPredefinedOptions,
  ...minutesGranularityPredefinedOptions,
  ...hoursGranularityPredefinedOptions,
];
export default defineComponent({
  components: {
    AdvancedWorkloadMetricWidget,
    RunaiDatepickerWithPredefined,
    RunaiSelect,
    DefaultWorkloadWidgets,
    LatencyThroughputAndReplicasWidget,
    NimWorkloadWidgets,
    VllmWorkloadWidgets,
  },
  props: {
    workload: {
      type: Object as PropType<Workload>,
      required: true,
    },
    drawerWidth: {
      type: String as PropType<string>,
      required: true,
    },
    selectedElementId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },
  data() {
    return {
      workloadStore: useWorkloadStore(),
      clusterStore: useClusterStore(),
      startDate: new Date() as Date,
      endDate: new Date() as Date,
      dateLabel: "" as string,
      defaultStartDate: undefined as Date | undefined,
      loadingChartsData: false as boolean,
      predefinedOptions: predefinedOptions as Array<PredefinedOption>,
      liveDataIntervalId: null as number | null,
      pods: [] as WorkloadServicePod[],
      selectedPodName: ALL_PODS_OPTION as string,
      inferenceWidgetsRef: "inference-widgets" as string,
      advancedWorkloadWidgetsRef: "advanced-workload-widgets" as string,
      defaultWorkloadWidgetsRef: "default-workload-widgets" as string,
      nimWidgetsRef: "nim-widgets" as string,
      vllmWidgetsRef: "vllm-widgets" as string,
      showHeaderAsGrid: false as boolean,
      selectedMetricType: EMetricsType.Default as EMetricsType,
      metricTypeOptions: [EMetricsType.Default] as EMetricsType[],
      liveEndDate: new Date() as Date,
      userActionTimestamp: 0 as number,
      workloadFirstRunDate: undefined as Date | undefined,
    };
  },
  async created() {
    this.initializeMetricTypeOptions();
    this.initSelectedPodName();
    this.resizeCharts = debounce(this.resizeCharts, 100);
    this.setFirstDatepickerLabel();
    await this.loadPods();
    this.loadWorkloadFirstRunDate();
    intervalUtil.startInterval(EIntervalLabels.WorkloadPods, this.loadPods);
    // Update live end date when end date is today
    intervalUtil.startInterval(EIntervalLabels.WorkloadMetricsLiveEndDate, () => {
      // Only update if end date is today AND workload is running
      if (dateUtil.isDateIsToday(this.endDate)) {
        const outsideOfLifespanPhases = new Set<Phase>([Phase.Failed, Phase.Completed]);
        if (!outsideOfLifespanPhases.has(this.workload.phase)) {
          this.liveEndDate = new Date();
        }
      }
    });
  },
  mounted() {
    this.updateHeaderStyles();
  },
  computed: {
    clusterVersion(): string | null {
      return this.clusterStore.getClusterById(this.clusterId)?.version || null;
    },
    isDefaultMetricType(): boolean {
      return this.selectedMetricType === EMetricsType.Default;
    },
    isAdvancedMetricType(): boolean {
      return this.selectedMetricType === EMetricsType.Advanced;
    },
    isInferenceDefaultMetricType(): boolean {
      return this.selectedMetricType === EMetricsType.Inference;
    },
    isNimMetricType(): boolean {
      return this.selectedMetricType === EMetricsType.NvidiaNim;
    },
    isVllmMetricType(): boolean {
      return this.selectedMetricType === EMetricsType.Vllm;
    },
    filterByDates(): IRangeDates {
      let dateEnd = this.endDate;
      if (dateUtil.isDateIsToday(this.endDate)) {
        dateEnd = this.workloadLifespanEndDate;
      }
      return { dateStart: this.startDate, dateEnd: dateEnd, label: this.dateLabel };
    },
    podsOptions(): Array<string> {
      const podsNames: Array<string> = this.pods.map((pod: WorkloadServicePod) => pod.name);
      if (
        this.selectedMetricType === EMetricsType.Default ||
        this.selectedMetricType === EMetricsType.NvidiaNim ||
        this.selectedMetricType === EMetricsType.Vllm ||
        this.selectedMetricType === EMetricsType.Inference
      ) {
        podsNames.unshift(ALL_PODS_OPTION);
      }
      return podsNames;
    },
    selectedPodId(): string | undefined {
      const pod: WorkloadServicePod | undefined = this.pods.find(
        (pod: WorkloadServicePod) => pod.name === this.selectedPodName,
      );

      if (!pod) return undefined;
      return pod.id;
    },
    workloadId(): string {
      return this.workload.id;
    },
    clusterId(): string {
      return this.workload.clusterId;
    },
    workloadCreationTime(): Date {
      return new Date(this.workload.createdAt);
    },
    workloadLifespanEndDate(): Date {
      const outsideOfLifespanPhases: Phase[] = [Phase.Failed, Phase.Completed];
      if (outsideOfLifespanPhases.includes(this.workload.phase)) {
        return new Date(this.workload.phaseUpdatedAt);
      }
      return this.liveEndDate;
    },
    isAdvancedGpuMetricsEnabled(): boolean {
      return useSettingStore().isAdvancedGpuMetricsEnabled;
    },
    isSupportingSelectionPods(): boolean {
      return (
        this.selectedMetricType === EMetricsType.Default ||
        this.selectedMetricType === EMetricsType.NvidiaNim ||
        this.selectedMetricType === EMetricsType.Vllm ||
        this.selectedMetricType === EMetricsType.Inference
      );
    },
    isNimMetricsSupported(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_NIM_METRICS);
    },
    isNimWorkload(): boolean {
      return workloadUtil.isNIMWorkload(this.workload);
    },
    isVllmMetricsSupported(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_VLLM_METRICS);
    },
    isVllmWorkload(): boolean {
      return workloadUtil.isVLLMWorkload(this.workload);
    },
    chartsContainerKey(): string {
      // This key will change whenever user changes filters, forcing charts to remount
      // which will trigger loading state via isFirstLoad flag
      return `${this.selectedMetricType}-${this.selectedPodName}-${this.selectedElementId}-${this.dateLabel}-${this.startDate.getTime()}-${this.endDate.getTime()}`;
    },
  },
  methods: {
    initializeMetricTypeOptions(): void {
      this.metricTypeOptions = [EMetricsType.Default];

      if (this.isAdvancedGpuMetricsEnabled) {
        this.metricTypeOptions.push(EMetricsType.Advanced);
      }

      if (this.workload.type === EWorkloadType.Inference) {
        this.metricTypeOptions.push(EMetricsType.Inference);
      }

      if (this.isNimMetricsSupported && this.isNimWorkload) {
        this.metricTypeOptions.push(EMetricsType.NvidiaNim);
      }

      if (this.isVllmMetricsSupported && this.isVllmWorkload) {
        this.metricTypeOptions.push(EMetricsType.Vllm);
      }

      if (!this.metricTypeOptions.includes(this.selectedMetricType)) {
        this.selectedMetricType = EMetricsType.Default;
      }
    },
    async loadPods(): Promise<void> {
      if (!this.workload?.id) return;
      try {
        // For NIM metrics, filter out worker pods (only show leader or pods without task label)
        const filterBy: Array<string> = [];
        if (this.isNimMetricType) {
          filterBy.push("task!=worker");
        }
        if (this.selectedElementId) {
          filterBy.push(filterUtil.getEqualsFilterString("elementId", this.selectedElementId));
        }

        this.pods = await workloadService.getWorkloadPods(
          this.workload?.id,
          undefined,
          undefined,
          undefined,
          undefined,
          undefined,
          filterBy.length ? filterBy : undefined,
        );
        if (!this.pods.find((pod: WorkloadServicePod) => pod.name === this.selectedPodName)) {
          this.selectedPodName = ALL_PODS_OPTION;
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Cannot get pod logs",
        });
        this.$q.notify(errorAlert.getNotification(error));
      }
    },
    initSelectedPodName(): void {
      if (this.pods.length === 1 && this.workload.type !== EWorkloadType.Inference) {
        this.selectedPodName = this.pods[0].name;
      }
    },
    setFirstDatepickerLabel(): void {
      this.endDate = new Date();
      this.startDate = new Date(this.workload.createdAt);
      this.defaultStartDate = new Date(this.workload.createdAt);
      this.dateLabel = predefinedOptions[0].label;
    },
    updateDateRange(dateRange: Array<Date>, dateLabel: string): void {
      this.startDate = dateRange[0];
      this.endDate = dateRange[1];
      this.dateLabel = dateLabel;
      // If end date is today, update liveEndDate immediately
      if (dateUtil.isDateIsToday(dateRange[1])) {
        this.liveEndDate = new Date();
      }
      // Mark this as a user action to trigger loading spinner in widgets
      this.userActionTimestamp = Date.now();
    },
    resizeCharts(): void {
      this.$nextTick(() => {
        (this.$refs[this.defaultWorkloadWidgetsRef] as TWorkloadCharts)?.resizeCharts();
        (this.$refs[this.inferenceWidgetsRef] as TWorkloadCharts)?.resizeCharts();
        (this.$refs[this.advancedWorkloadWidgetsRef] as TWorkloadCharts)?.resizeCharts();
        (this.$refs[this.nimWidgetsRef] as TWorkloadCharts)?.resizeCharts();
        (this.$refs[this.vllmWidgetsRef] as TWorkloadCharts)?.resizeCharts();
      });
    },
    // barak:todo the history API is not paginated here - fetches all records to find the first Running phase.
    // Discuss with team whether a dedicated endpoint or query param would be more efficient.
    async loadWorkloadFirstRunDate(): Promise<void> {
      try {
        const history = await workloadService.getWorkloadHistory(this.workloadId, this.clusterId);
        const runningRecords = history.records
          .filter(
            (record) =>
              record.meta.type === HistoryRecordType.PhaseUpdate && record.spec.phaseUpdate?.phase === Phase.Running,
          )
          .sort((a, b) => new Date(a.meta.creationTimestamp).getTime() - new Date(b.meta.creationTimestamp).getTime());
        if (runningRecords.length > 0) {
          this.workloadFirstRunDate = new Date(runningRecords[0].meta.creationTimestamp);
        }
      } catch {
        // Silently fail - the option will remain disabled
      }
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
    updateHeaderStyles(): void {
      this.$nextTick(() => {
        const headerElm: HTMLElement = this.$refs["workload-metrics-header"] as HTMLElement;
        if (headerElm) {
          this.showHeaderAsGrid = headerElm.offsetWidth < MAX_WIDTH_FOR_GRID;
        }
      });
    },
    selectFirstPod(): void {
      if (this.pods.length > 0) {
        this.selectedPodName = this.pods[0].name;
      }
    },
  },
  watch: {
    drawerWidth: {
      handler() {
        this.resizeCharts();
        this.updateHeaderStyles();
      },
    },
    workloadId() {
      if (this.selectedMetricType === EMetricsType.Default) {
        this.selectedPodName = ALL_PODS_OPTION;
      }

      this.initializeMetricTypeOptions();
      this.setFirstDatepickerLabel();
    },
    async selectedMetricType(newMetricType: EMetricsType, oldMetricType: EMetricsType) {
      if (newMetricType === EMetricsType.Advanced) {
        this.selectFirstPod();
      }

      // Reload pods when switching to/from NIM metrics to apply/remove the task!=worker filter
      const wasNimMetric = oldMetricType === EMetricsType.NvidiaNim;
      const isNimMetric = newMetricType === EMetricsType.NvidiaNim;
      if (wasNimMetric !== isNimMetric) {
        await this.loadPods();
      }
    },
    async selectedElementId() {
      this.selectedPodName = ALL_PODS_OPTION;
      await this.loadPods();
    },
    pods() {
      if (this.selectedMetricType === EMetricsType.Advanced) {
        this.selectFirstPod();
      }
    },
  },
  beforeUnmount() {
    intervalUtil.stopInterval(EIntervalLabels.WorkloadPods);
    intervalUtil.stopInterval(EIntervalLabels.WorkloadMetricsLiveEndDate);
  },
});
</script>
<style lang="scss" scoped>
.workload-metrics-tab {
  height: 100%;
  overflow: hidden;

  .charts-container {
    overflow: auto;
    height: 100%;
  }

  .metric-select {
    min-width: 180px;
  }
}

.workload-metrics-header {
  border-bottom: solid 1px $black-10;
  display: flex;
  flex-wrap: nowrap;
  justify-content: space-between;

  &.is-grid {
    display: grid;
    grid-template-rows: auto auto;
    grid-template-columns: 1fr;
    padding: 10px;
    gap: 10px;
  }

  .metric-type-select {
    max-width: 180px;
  }

  .metric-pod-select {
    max-width: 260px;
  }
}
</style>
