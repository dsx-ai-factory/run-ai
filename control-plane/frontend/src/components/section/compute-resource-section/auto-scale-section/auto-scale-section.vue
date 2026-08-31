<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="auto-scale-section">
    <replica-scaling-section
      :model-value="{
        minReplicas: autoScaleData.minReplicas,
        maxReplicas: autoScaleData.maxReplicas,
      }"
      :disabled="disabled"
      :policy-rules="policyRules"
      :is-auto-scale-enabled="autoScaleOptions.isAutoScaleEnabled"
      @update:model-value="onReplicaScalingChanged"
    />

    <q-slide-transition>
      <threshold-metric-section
        v-if="showReplicaCondition"
        :model-value="{
          thresholdMetric: autoScaleData.thresholdMetric,
          thresholdValue: autoScaleData.thresholdValue,
        }"
        :disabled="disabled"
        :policy-rules="policyRules"
        :policy-defaults="policyDefaults"
        :latency-enabled="autoScaleOptions.latencyEnabled"
        @update:model-value="onThresholdMetricChanged"
      ></threshold-metric-section>
    </q-slide-transition>

    <scale-down-section
      v-if="showScaleDownToZeroValues"
      :key="autoScaleData.minReplicas"
      :model-value="{
        scaleToZeroRetentionSeconds: autoScaleData.scaleToZeroRetentionSeconds,
      }"
      :disabled="disabled"
      :is-auto-scale-enabled="canChangeScaleDown(autoScaleData)"
      :policy-rules="policyRules"
      :policy-defaults="policyDefaults"
      @update:model-value="onScaleDownChanged"
    >
    </scale-down-section>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { errorMessages } from "@/common/error-message.constant";
import { isMaxEqualOrHigherThenMin } from "@/common/form.validators";
import { ReplicaScalingSection } from "./replica-scaling-section";
import { ThresholdMetricSection } from "./threshold-metric-section";
import { ScaleDownSection } from "./scale-down-section";

// utils
import { is } from "quasar";

// models
import {
  EAutoScaleSectionEvents,
  MIN_REPLICAS_FOR_LATENCY_METRIC,
  type IAutoScaleData,
  type IAutoScaleOptions,
} from "./auto-scale-section.model";
import type { AutoScalingRules } from "@/swagger-models/assets-service-client";
import {
  DEFAULT_THRESHOLD_VALUE,
  METRIC_LATENCY,
  type IThresholdMetricData,
} from "./threshold-metric-section/threshold-metric-section.model";
import type { IReplicaScalingSection } from "./replica-scaling-section/replica-scaling-section.model";
import {
  DEFAULT_SCALE_DOWN_TO_ZERO_ID,
  MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS,
  type IScaleDownData,
} from "./scale-down-section/scale-down-section.model";
import type { AutoScaling } from "@/swagger-models/workloads-client";
import { thresholdMetricUtils } from "./threshold-metric-section/threshold-metric.util";

export default defineComponent({
  name: "auto-scale-section",
  components: {
    ReplicaScalingSection,
    ThresholdMetricSection,
    ScaleDownSection,
  },
  emits: {
    [EAutoScaleSectionEvents.OnAutoScaleChanged]: (_data: IAutoScaleData) => true,
  },
  props: {
    autoScaleData: {
      type: Object as PropType<IAutoScaleData>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<AutoScalingRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<AutoScaling | null>,
      required: false,
    },
    autoScaleOptions: {
      type: Object as PropType<IAutoScaleOptions>,
      required: true,
    },
  },
  computed: {
    showScaleDownToZeroValues(): boolean {
      return this.autoScaleData.thresholdMetric !== METRIC_LATENCY;
    },
    showReplicaCondition(): boolean {
      return this.shouldShowReplicaCondition(this.autoScaleData);
    },
    minReplicasAllowed(): number {
      return this.autoScaleOptions.isAutoScaleEnabled ? 0 : 1;
    },
    minValueForMaxReplicas(): number {
      if (this.autoScaleData.minReplicas === undefined || this.autoScaleData.minReplicas === 0) return 1;
      return this.autoScaleData.minReplicas;
    },
  },
  methods: {
    canChangeScaleDown(autoScaleData: IAutoScaleData): boolean {
      return this.autoScaleOptions.isAutoScaleEnabled && this.isAllowScaleToZero(autoScaleData.minReplicas);
    },
    shouldShowReplicaCondition(autoScaleData: IAutoScaleData): boolean {
      if (autoScaleData.minReplicas === undefined || autoScaleData.maxReplicas === undefined) {
        return false;
      }

      if (autoScaleData.minReplicas === MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS && autoScaleData.maxReplicas === 1) {
        return false;
      }

      return autoScaleData.maxReplicas > autoScaleData.minReplicas;
    },
    isMaxMorThanMin(): boolean | string {
      if (this.autoScaleData.minReplicas === undefined || this.autoScaleData.maxReplicas === undefined) return true;
      return (
        isMaxEqualOrHigherThenMin(this.autoScaleData.maxReplicas, this.autoScaleData.minReplicas) ||
        errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM
      );
    },
    isAllowScaleToZero(minReplicas: number | undefined): boolean {
      return minReplicas === MIN_REPLICAS_TO_ALLOW_SCALE_TO_ZERO_OPTIONS;
    },
    onReplicaScalingChanged(replicaScalingData: IReplicaScalingSection): void {
      const autoScaleData: IAutoScaleData = {
        ...this.autoScaleData,
        minReplicas: replicaScalingData.minReplicas,
        maxReplicas: replicaScalingData.maxReplicas,
      };

      const defaultScaleDownToZeroValue =
        this.policyDefaults?.scaleToZeroRetentionSeconds || Number(DEFAULT_SCALE_DOWN_TO_ZERO_ID);

      // enable the scale to zero section
      autoScaleData.scaleToZeroRetentionSeconds = this.canChangeScaleDown(autoScaleData)
        ? defaultScaleDownToZeroValue
        : undefined;

      // update threshold section data
      const { thresholdMetric, thresholdValue } = this.getThresholdValues(autoScaleData);
      autoScaleData.thresholdMetric = thresholdMetric;
      autoScaleData.thresholdValue = thresholdValue;
      this.updateModel(autoScaleData);
    },
    getThresholdValues(autoScaleData: IAutoScaleData): {
      thresholdMetric: string | undefined;
      thresholdValue: number | undefined;
    } {
      const shouldShowReplica = this.shouldShowReplicaCondition(autoScaleData);

      const defaultMetric = thresholdMetricUtils.getDefaultMetricOption(
        undefined,
        this.policyDefaults,
        this.policyRules,
      );
      const defaultValue = this.policyDefaults?.metricThreshold || DEFAULT_THRESHOLD_VALUE;

      let thresholdMetric: string | undefined;
      let thresholdValue: number | undefined;

      if (shouldShowReplica) {
        thresholdMetric = autoScaleData.thresholdMetric ?? defaultMetric;
        thresholdValue = autoScaleData.thresholdValue ?? defaultValue;
      } else {
        thresholdMetric = undefined;
        thresholdValue = undefined;
      }

      return { thresholdMetric, thresholdValue };
    },
    onThresholdMetricChanged(thresholdMetricData: IThresholdMetricData): void {
      const autoScaleData: IAutoScaleData = {
        ...this.autoScaleData,
        thresholdMetric: thresholdMetricData.thresholdMetric,
        thresholdValue: thresholdMetricData.thresholdValue,
      };

      if (thresholdMetricData.thresholdMetric === METRIC_LATENCY) {
        autoScaleData.scaleToZeroRetentionSeconds = undefined;
        autoScaleData.minReplicas = MIN_REPLICAS_FOR_LATENCY_METRIC;
      }

      this.updateModel(autoScaleData);
    },
    onScaleDownChanged(scaleDownData: IScaleDownData): void {
      this.updateModel({
        ...this.autoScaleData,
        scaleToZeroRetentionSeconds: scaleDownData.scaleToZeroRetentionSeconds,
      });
    },
    updateModel(autoScaleData: IAutoScaleData): void {
      this.$emit(EAutoScaleSectionEvents.OnAutoScaleChanged, autoScaleData);
    },
    isValueEmpty(val: string): boolean | string {
      return is.number(val) || errorMessages.VALUE_NOT_EMPTY;
    },
    isGreaterThanZero(val: string): boolean | string {
      return Number(val) > 0 || errorMessages.VALUE_ABOVE_ZERO;
    },
    isMinReplicasAllowedZero(val: number | undefined): boolean | string {
      if (this.autoScaleOptions.isAutoScaleEnabled) return true;
      return Number(val) > 0 || errorMessages.MIN_REPLICAS_ALLOWED_ZERO;
    },
  },
});
</script>
<style lang="scss" scoped></style>
