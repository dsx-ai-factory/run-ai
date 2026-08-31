<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="threshold-metric-section">
    <div class="q-my-md">Set the conditions for creating a new replica</div>
    <div class="row new-replica-condition-container">
      <runai-select
        class="col-6 threshold-metric-options"
        aid="threshold-metric-options"
        :model-value="modelValue.thresholdMetric || null"
        @update:model-value="onThresholdMetricChanged"
        :options="thresholdMetricOptions"
        :disable="disabled"
        label="Variable"
        emit-value
        :policy-rules="policyMetricsRules"
        icon-tooltip-width="300px"
      />

      <q-select
        class="col-2"
        label="Operator"
        :model-value="1"
        option-value="id"
        option-label="name"
        map-options
        :options="[{ id: 1, name: '>' }]"
        no-error-icon
        :disable="true"
        hide-dropdown-icon
      />
      <policy-number-input
        class="col-2"
        label="Value"
        :model-value="modelValue.thresholdValue || null"
        @update:model-value="onThresholdValueChanged"
        stack-label
        :min-value="thresholdValueMinValue"
        :disable="disabled"
        :rules="[isValueEmpty, isGreaterThanMin]"
        aid="threshold-value-input"
        :policy-rules="policyRules?.metricThreshold"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiSelect } from "@/components/common/runai-select";
import { errorMessages } from "@/common/error-message.constant";

// models
import {
  type IThresholdMetricOption,
  type IThresholdMetricData,
  METRIC_LATENCY,
  thresholdMetricOptions,
  EThresholdMetricSectionEvents,
} from "./threshold-metric-section.model";
import type { ISelectOption } from "@/models/global.model";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import {
  type AutoScalingMetricOptionsOptionsInner,
  type AutoScalingRules,
} from "@/swagger-models/assets-service-client";
import type { AutoScaling } from "@/swagger-models/workloads-client";
import { thresholdMetricUtils } from "./threshold-metric.util";

export default defineComponent({
  name: "threshold-metric-section",
  components: {
    RunaiSelect,
    PolicyNumberInput,
  },
  emits: {
    [EThresholdMetricSectionEvents.UpdateModelValue]: (_data: IThresholdMetricData) => true,
  },

  props: {
    modelValue: {
      type: Object as PropType<IThresholdMetricData>,
      required: true,
    },
    latencyEnabled: {
      type: Boolean as PropType<boolean>,
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
  },
  data() {
    return {
      thresholdMetricOptions: [] as Array<ISelectOption>,
      thresholdValueMinValue: 1 as number,
    };
  },
  created() {
    this.thresholdMetricOptions = thresholdMetricOptions.map((option: IThresholdMetricOption) => {
      let disable = false;
      let toolTip = "";
      if (!this.latencyEnabled && option.id === METRIC_LATENCY) {
        disable = true;
        toolTip =
          "To enable this, ask your administrator to verify that Knative is installed properly on the selected cluster";
      }

      return {
        disable,
        toolTip,
        label: option.name,
        value: option.id.toString(),
      };
    });

    if (this.policyDefaults?.metric) {
      const metric = this.policyDefaults.metric;
      const isOptionExist = this.thresholdMetricOptions.find((option: ISelectOption) => option.value === metric);
      if (!isOptionExist) {
        this.thresholdMetricOptions.unshift({
          disable: false,
          toolTip: "",
          label: this.policyDefaults.metric,
          value: this.policyDefaults.metric,
        });
      }
    }
  },
  computed: {
    policyMetricsRules(): IGenericSelectPolicyRules | undefined {
      if (!this.policyRules?.metric) return;
      return {
        ...this.policyRules.metric,
        options: this.policyRules.metric.options?.map((option: AutoScalingMetricOptionsOptionsInner) => {
          return {
            label: option.displayed || option.value,
            value: option.value,
          };
        }),
      };
    },
  },
  methods: {
    onThresholdMetricChanged(selectedValue: string): void {
      const thresholdMetricData: IThresholdMetricData = {
        thresholdMetric: selectedValue,
        thresholdValue: thresholdMetricUtils.getThresholdMetricDefaultValue(
          undefined,
          this.policyDefaults,
          this.policyRules,
        ),
      };
      this.updateModel(thresholdMetricData);
    },
    onThresholdValueChanged(value: number | string | null): void {
      const thresholdValue = value === "" || value === null ? undefined : Number(value);
      this.updateModel({
        ...this.modelValue,
        thresholdValue,
      });
    },
    updateModel(thresholdMetricData: IThresholdMetricData): void {
      this.$emit(EThresholdMetricSectionEvents.UpdateModelValue, thresholdMetricData);
    },
    isValueEmpty(val: string): boolean | string {
      return val !== "" || errorMessages.VALUE_NOT_EMPTY;
    },
    isGreaterThanMin(val: string): boolean | string {
      const num = Number(val);
      return num >= this.thresholdValueMinValue || errorMessages.VALUE_ABOVE_ZERO;
    },
  },
});
</script>
<style lang="scss" scoped>
.threshold-metric-section {
  .new-replica-condition-container {
    gap: 45px;
    background-color: $body-background-color;
    padding: 15px;
  }
}
</style>
