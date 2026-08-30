<template>
  <runai-section
    class="duration-input-section"
    :title="label"
    :tooltip-text="tooltipText"
    tooltip-width="535px"
    tooltip-position="right"
  >
    <runai-duration-input
      :model-value="modelValue || null"
      @update:model-value="updateDurationTime"
      :disable="isDisable"
      :key="inputCmpKey"
      :rules="rules"
      hide-days
      hide-hours
    />
    <q-tooltip v-if="policyRules?.canEdit === false">{{ cantBeModifiedByPolicy }}</q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
import type { ValidationRule } from "quasar";

// components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiDurationInput } from "@/components/common/runai-duration-input/";
// utils
import { errorMessages, errorMessagesUtils } from "@/common/error-message.constant";
import type { IntegerRules } from "@/swagger-models/policy-service-client";

export enum EEmitDurationInput {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "duration-input-section",
  components: {
    RunaiDurationInput,
    RunaiSection,
  },
  emits: {
    [EEmitDurationInput.UpdateModelValue]: (_: number | null) => {
      return _;
    },
  },
  props: {
    modelValue: {
      type: [Number, null] as PropType<number | null>,
      required: true,
    },
    label: {
      type: String,
      required: true,
    },
    tooltipText: {
      type: String,
      required: true,
    },
    min: {
      type: Number,
      required: true,
    },
    max: {
      type: Number,
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
  computed: {
    isDisable(): boolean {
      return this.disable || this.policyRules?.canEdit === false;
    },
    rules(): ValidationRule[] {
      const rules: ValidationRule[] = [this.rangeRule];
      if (!this.policyRules) return rules;

      this.policyRules.step && rules.push(this.validStepRule);

      return rules;
    },
    cantBeModifiedByPolicy(): string {
      return errorMessages.CANT_BE_MODIFIED_POLICY;
    },
    timeframeErrorMessage(): string {
      const maxInMinutes = Math.floor(this.maxSecondsAllowedValue / 60);
      return errorMessagesUtils.TIMEFRAME_BETWEEN_MIN_SECONDS_AND_MAX_MINUTES(this.minSecondsAllowedValue, maxInMinutes);
    },
    minSecondsAllowedValue(): number {
      if (this.policyRules?.min === undefined || this.policyRules?.min === null) {
        return this.min;
      }

      return this.policyRules.min;
    },
    maxSecondsAllowedValue(): number {
      if (this.policyRules?.max === undefined || this.policyRules?.max === null) {
        return this.max;
      }

      return this.policyRules.max;
    },
  },
  methods: {
    rangeRule(): string | boolean {
      if (this.modelValue === null || this.modelValue === undefined) {
        return true;
      }
      return this.modelValue >= this.minSecondsAllowedValue && this.modelValue <= this.maxSecondsAllowedValue
        ? true
        : this.timeframeErrorMessage;
    },
    validStepRule(val: number): boolean | string {
      const min = this.minSecondsAllowedValue;
      const step = this.policyRules?.step || 1;
      return (
        (val - min) % step === 0 ||
        errorMessages.TIMEFRAME_STEPS.replace("${min}", `${min}`)
          .replace("${step}", `${step}`)
          .replace("${min + step}", `${min + step}`)
      );
    },
    updateDurationTime(time: number | null): void {
      this.$emit(EEmitDurationInput.UpdateModelValue, time);
    },
  },
});
</script>

<style lang="scss" scoped></style>
