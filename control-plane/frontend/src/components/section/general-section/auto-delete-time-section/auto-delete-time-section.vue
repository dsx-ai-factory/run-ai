<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="auto-delete-time-section q-mb-lg">
    <div class="time-inputs-container row items-center justify-between">
      <div>
        <div class="q-pb-md">
          <span>Set the timeframe for auto-deletion after workload completion or failure</span>
          <runai-tooltip
            class="q-ml-md"
            width="535px"
            tooltip-position="right"
            tooltip-text="The timeframe after which a workload that has completed or failed will be automatically deleted. If the fields are set to 0, the workload will be deleted immediately after completion or failure."
          />
        </div>
        <runai-duration-input
          :rules="rules"
          :model-value="modelValue"
          @update:model-value="updateAutoDeletionTime"
          :disable="isDisabled"
          :key="inputCmpKey"
        />
        <q-tooltip v-if="policyRules?.canEdit === false">{{ cantBeModifiedByPolicy }}</q-tooltip>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { errorMessages } from "@/common/error-message.constant";
import { RunaiDurationInput } from "@/components/common/runai-duration-input/";
import type { IntegerRules } from "@/swagger-models/policy-service-client";
import { isMaxEqualOrHigherThenMin, isValueLowerOrEqualToMax } from "@/common/form.validators";
import type { ValidationRule } from "quasar";
export default defineComponent({
  name: "auto-delete-time-section",
  components: {
    RunaiDurationInput,
    RunaiTooltip,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
  },
  data() {
    return {
      inputCmpKey: 1001,
    };
  },
  mounted() {
    // 30 days is the default for this property
    if (!this.modelValue && typeof this.modelValue !== "number") {
      const thirtyDaysInSeconds = 60 * 60 * 24 * 30;
      this.updateAutoDeletionTime(thirtyDaysInSeconds);
      this.inputCmpKey++;
    }
  },
  computed: {
    rules(): ValidationRule[] {
      const rules: ValidationRule[] = [this.notZeroValue];
      if (!this.policyRules) return rules;

      this.policyRules.max && rules.push(this.notLargerThanRule);
      this.policyRules.min && rules.push(this.notSmallerThanRule);
      this.policyRules.step && rules.push(this.validStepRule);

      return rules;
    },
    isDisabled(): boolean {
      return this.disable || this.policyRules?.canEdit === false;
    },
    cantBeModifiedByPolicy(): string {
      return errorMessages.CANT_BE_MODIFIED_POLICY;
    },
    timeframeErrorMessage(): string {
      const min = this.policyRules?.min || 0;
      const max = this.policyRules?.max;
      if (max) {
        return errorMessages.TIMEFRAME_BETWEEN_MIN_AND_MAX_SECONDS.replace("${min}", `${min}`).replace(
          "${max}",
          `${max}`,
        );
      }

      return errorMessages.VALID_TIME_FRAME.replace("0", `${min}`);
    },
  },
  methods: {
    notZeroValue(): string | boolean {
      const min = this.policyRules?.min || 0;
      if (this.modelValue === null || this.modelValue === undefined) {
        return this.timeframeErrorMessage;
      }
      return this.modelValue >= min ? true : this.timeframeErrorMessage;
    },
    notLargerThanRule(val: number): boolean | string {
      const max = this.policyRules?.max || Infinity;
      return isValueLowerOrEqualToMax(val, max) || this.timeframeErrorMessage;
    },
    notSmallerThanRule(val: number): boolean | string {
      const min = this.policyRules?.min || 0;
      return isMaxEqualOrHigherThenMin(val, min) || this.timeframeErrorMessage;
    },
    validStepRule(val: number): boolean | string {
      const min = this.policyRules?.min || 0;
      const step = this.policyRules?.step || 1;
      return (
        (val - min) % step === 0 ||
        errorMessages.TIMEFRAME_STEPS.replace("${min}", `${min}`)
          .replace("${step}", `${step}`)
          .replace("${min + step}", `${min + step}`)
      );
    },
    updateAutoDeletionTime(time: number | null): void {
      this.$emit("update:model-value", time);
    },
  },
});
</script>

<style lang="scss" scoped>
.time-inputs-container {
  gap: 10px;
}
</style>
