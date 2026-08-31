<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item label="Scheduling">
    <template #subheader>
      <span class="q-mr-sm">{{ schedulingSubHeader }}</span>
    </template>

    <runai-section
      v-if="isClusterSupportsMinGuaranteedRuntime"
      title="Set the minimum guaranteed runtime"
      tooltip-text="The minimum guaranteed runtime is the time that a job will run before it can be preempted by a higher priority job. If the job runs for less than this time, it will not be preempted."
    >
      <runai-duration-input :model-value="durationInSeconds" @update:model-value="updateMinRuntime" />
    </runai-section>

    <runai-section class="q-mt-lg q-mb-lg" v-if="isClusterSupportsTimeBasedFairShare">
      <runai-toggle
        aid="allow-time-based-fair-share"
        :model-value="allowTimeBasedFairShare"
        @update:model-value="updateTimeBasedFairShareEnabled"
        label="Allow time-based fairshare"
        left-label
      />
      <runai-tooltip
        tooltip-position="right"
        tooltip-text="When enabled, the Scheduler adjusts fairshare scores based on each project's weight and historical usage records. Historical values are not saved as persistent data.
When disabled, the Scheduler uses the default fairshare, with scheduling fairness based on the current project's weight and resource usage."
      />
      <div class="italic q-mt-lg">
        For more information, see the
        <a target="_blank" :href="DOCS.NODE_POOLS.TIME_BASED_FAIR_SHARE"> Time-based fairshare guide. </a>
      </div>
      <div class="italic q-mt-sm">
        For details on enabling persistent data, see the
        <a target="_blank" :href="DOCS.NODE_POOLS.PERSISTENT_DATA"> Advanced Cluster Configurations guide. </a>
      </div>
    </runai-section>

    <runai-section
      v-if="isClusterSupportsTimeBasedFairShare && allowTimeBasedFairShare"
      class="q-mt-sm"
      title="Set the weight value that determines how a project's historical resource usage affects its fairshare"
      :tooltip-text="historicalUsageWeightTooltip"
    >
      <section class="row">
        <q-input
          aid="historical-usage-weight-input"
          type="number"
          class="col-3"
          stack-label
          no-error-icon
          label="Historical usage weight"
          :model-value="historicalUsageWeight"
          @update:model-value="updateHistoricalUsageWeight"
          :rules="[historicalUsageWeightRule]"
          :min="0"
          step="0.1"
          @wheel.prevent
        />
      </section>
    </runai-section>

    <runai-section
      v-if="isClusterSupportsTimeBasedFairShare && allowTimeBasedFairShare"
      class="q-mt-lg"
      title="Set the duration of the historical usage window"
      :tooltip-text="historicalUsageWindowTooltip"
    >
      <runai-duration-input
        :model-value="windowDurationInSeconds"
        hide-seconds
        :rules="[windowDurationRule]"
        @update:model-value="updateWindowDuration"
      />
    </runai-section>

    <runai-section class="q-mt-lg q-mb-lg" v-if="isClusterSupportsNumaAwareScheduling">
      <runai-toggle
        aid="numa-aware-scheduling"
        :model-value="numaAwareScheduling"
        @update:model-value="$emit('update:numa-aware-scheduling', $event)"
        label="NUMA-aware scheduling"
        left-label
      />
      <runai-tooltip
        tooltip-position="right"
        width="779px"
        tooltip-text="When enabled, the Scheduler places workloads so that GPU, CPU, and CPU memory are allocated within the same NUMA node where possible, minimizing cross-NUMA fragmentation and improving performance. There are cluster prerequisites for enabling this feature.
When disabled, the Scheduler places workloads using aggregate node capacity only, without considering NUMA topology."
      />
      <div class="italic q-mt-md">
        For more information on the prerequisites and configuration requirements, see the
        <a target="_blank" :href="DOCS.NODE_POOLS.NUMA_AWARE_SCHEDULING"> Configuring NUMA-Aware Scheduling guide. </a>
      </div>
    </runai-section>

    <div class="q-mt-md">
      <span class="text-b-subtitle1">GPU pods</span>
      <runai-radio-options
        class="q-mt-md"
        title="Set the node placement strategy"
        :title-tooltip="gpuNodePlacementTooltip"
        aid="node-pool-gpu-placement-radio"
        :options="placementOptions.gpu"
        :model-value="gpuPlacementStrategy"
        @update:model-value="$emit('update:gpu-placement-strategy', $event)"
        horizontal
      />
      <runai-radio-options
        v-if="showGpuDevicePlacement"
        class="q-mt-md"
        title="Set the device placement strategy"
        :title-tooltip="gpuDevicePlacementTooltip"
        aid="node-pool-gpu-device-placement-radio"
        :options="placementOptions.gpuDevice"
        :model-value="gpuDevicePlacementStrategy"
        @update:model-value="$emit('update:gpu-device-placement-strategy', $event)"
        horizontal
      />
    </div>

    <div class="q-mt-lg">
      <span class="text-b-subtitle1">CPU pods</span>
      <runai-radio-options
        class="q-mt-md"
        title="Set the node placement strategy"
        :title-tooltip="cpuNodePlacementTooltip"
        aid="node-pool-cpu-placement-radio"
        :options="placementOptions.cpu"
        :model-value="cpuPlacementStrategy"
        @update:model-value="$emit('update:cpu-placement-strategy', $event)"
        horizontal
      />
    </div>
  </runai-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiDurationInput } from "@/components/common/runai-duration-input";
import {
  PlacementOptions,
  GPU_NODE_PLACEMENT_TOOLTIP,
  GPU_DEVICE_PLACEMENT_TOOLTIP,
  CPU_NODE_PLACEMENT_TOOLTIP,
} from "@/models/node-pool.model";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import type { PlacementStrategy } from "@/swagger-models/cluster-service-client";
import { RunaiSection } from "@/components/common/runai-section";
import { dateUtil } from "@/utils/date.util";
import { RunaiToggle } from "@/components/common";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { errorMessages } from "@/common/error-message.constant";
import { DOCS } from "@/constants/docs";

const DEFAULT_HISTORICAL_USAGE_WEIGHT = 1.0;
const DEFAULT_WINDOW_DURATION = "7d";

export default defineComponent({
  name: "node-pool-placement-strategy-section",
  components: { RunaiTooltip, RunaiToggle, RunaiSection, RunaiRadioOptions, RunaiExpansionItem, RunaiDurationInput },
  emits: [
    "update:gpu-placement-strategy",
    "update:cpu-placement-strategy",
    "update:gpu-device-placement-strategy",
    "update:min-guaranteed-runtime",
    "update:allow-time-based-fair-share",
    "update:historical-usage-weight",
    "update:window-duration",
    "update:numa-aware-scheduling",
  ],
  props: {
    historicalUsageWeight: {
      type: [Number, null] as PropType<number | null>,
      default: null,
    },
    windowDuration: {
      type: String as PropType<string>,
      default: "",
    },
    cpuPlacementStrategy: {
      type: String as PropType<PlacementStrategy>,
      required: true,
    },
    gpuPlacementStrategy: {
      type: String as PropType<PlacementStrategy>,
      required: true,
    },
    gpuDevicePlacementStrategy: {
      type: String as PropType<PlacementStrategy>,
      required: true,
    },
    showGpuDevicePlacement: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    minGuaranteedRuntime: {
      type: String as PropType<string>,
      default: "",
    },
    allowTimeBasedFairShare: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isClusterSupportsMinGuaranteedRuntime: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isClusterSupportsTimeBasedFairShare: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    numaAwareScheduling: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isClusterSupportsNumaAwareScheduling: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return { DOCS };
  },
  computed: {
    gpuNodePlacementTooltip(): string {
      return GPU_NODE_PLACEMENT_TOOLTIP;
    },
    gpuDevicePlacementTooltip(): string {
      return GPU_DEVICE_PLACEMENT_TOOLTIP;
    },
    cpuNodePlacementTooltip(): string {
      return CPU_NODE_PLACEMENT_TOOLTIP;
    },
    historicalUsageWeightTooltip(): string {
      return `• When set to 0.0: historical usage data is ignored, fairshare is 100% current weight.\n
• When set to 1.0: this value create an equilibrium of the impact of the historical usage on current weight, if historical usage equals the current weight then the fairshare is exactly the weight portion.\n
• Values >0.0 and <1.0 will compensate for deficient historical usage or penalize for excess historical usage, lower value means lower effect of the historical usage on the fairshare, making historical usage influential factor on the current fairshare.\n
• Values >1 will compensate for deficient historical usage or penalize for excess historical usage, higher value means higher effect of the historical usage on the fairshare, making historical usage the dominant factor of the current fairshare.`;
    },
    historicalUsageWindowTooltip(): string {
      return `If the sliding window is 7 days, then at any given moment the system looks only at usage from the past 7 days:\n
• If today is January 15 at 10:00, the window covers January 8 at 10:00 → January 15 at 10:00.\n
• One hour later, at January 15 at 11:00, the window "slides" forward by one hour and now covers January 8 at 11:00 → January 15 at 11:00.
Anything outside that moving 7-day range is no longer counted.`;
    },
    windowDurationInSeconds(): number | null {
      if (!this.windowDuration) return null;
      const seconds = dateUtil.parseDurationStringToSeconds(this.windowDuration);
      return seconds >= 0 ? seconds : null;
    },
    durationInSeconds(): number | null {
      if (!this.minGuaranteedRuntime || !this.isClusterSupportsMinGuaranteedRuntime) {
        return null;
      }
      return dateUtil.parseDurationStringToSeconds(this.minGuaranteedRuntime);
    },
    schedulingSubHeader(): string {
      return `${this.minGuaranteedRuntimeSubHeader} ${this.placementSubHeader}`;
    },
    placementSubHeader(): string {
      const gpuDevice = this.showGpuDevicePlacement ? `  GPU device: ${this.gpuDevicePlacementStrategyText}` : "";
      return `GPU: ${this.gpuPlacementStrategyText}${gpuDevice}  CPU: ${this.cpuPlacementStrategyText}`;
    },
    minGuaranteedRuntimeSubHeader(): string {
      if (
        this.isClusterSupportsMinGuaranteedRuntime &&
        this.minGuaranteedRuntime &&
        this.minGuaranteedRuntime.length > 0
      ) {
        return `Minimum guaranteed runtime: ${this.minGuaranteedRuntime}.`;
      }
      return "";
    },
    placementOptions(): typeof PlacementOptions {
      return PlacementOptions;
    },
    gpuPlacementStrategyText(): string {
      return nodePoolUtil.placementStrategyText(this.gpuPlacementStrategy);
    },
    gpuDevicePlacementStrategyText(): string {
      return nodePoolUtil.placementStrategyText(this.gpuDevicePlacementStrategy);
    },
    cpuPlacementStrategyText(): string {
      return nodePoolUtil.placementStrategyText(this.cpuPlacementStrategy);
    },
  },
  methods: {
    updateTimeBasedFairShareEnabled(value: boolean): void {
      this.$emit("update:allow-time-based-fair-share", value);
      if (value) {
        this.$emit("update:historical-usage-weight", DEFAULT_HISTORICAL_USAGE_WEIGHT);
        this.$emit("update:window-duration", DEFAULT_WINDOW_DURATION);
      }
    },
    historicalUsageWeightRule(val: number | string | null): boolean | string {
      if (val === null || val === undefined || val === "") return true;
      return Number(val) >= 0 ? true : errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace("${min}", `0.0`);
    },
    windowDurationRule(val: number | null): boolean | string {
      if (val === null || val === undefined) return true;
      return val >= 0 ? true : errorMessages.VALID_TIME_FRAME_MINUTES;
    },
    updateHistoricalUsageWeight(value: number | string | null): void {
      if (value === "" || value === null || value === undefined) {
        this.$emit("update:historical-usage-weight", DEFAULT_HISTORICAL_USAGE_WEIGHT);
        return;
      }
      const num = Number(value);
      this.$emit("update:historical-usage-weight", isNaN(num) ? DEFAULT_HISTORICAL_USAGE_WEIGHT : num);
    },
    updateWindowDuration(seconds: number | null): void {
      const durationString = seconds != null ? dateUtil.formatSecondsToDuration(seconds) : DEFAULT_WINDOW_DURATION;
      this.$emit("update:window-duration", durationString);
    },
    updateMinRuntime(durationInSeconds: number | null): void {
      let durationString = "";
      if (durationInSeconds && durationInSeconds > 0) {
        durationString = dateUtil.formatSecondsToDuration(durationInSeconds);
      }
      this.$emit("update:min-guaranteed-runtime", durationString);
    },
  },
});
</script>
