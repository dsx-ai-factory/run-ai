<template>
  <runai-section
    class="termination-grace-period-section"
    title="Set the grace period for workload preemption"
    tooltip-text="A buffer that allows a preempted workload to reach a safe checkpoint before it is forcibly preempted"
  >
    <runai-duration-input
      :model-value="modelValue"
      @update:model-value="$emit('update:model-value', $event)"
      :rules="rules"
      hide-days
      hide-hours
      :disable="isDisabled"
    />
    <q-tooltip v-if="isDisabledByPolicy" target=".termination-grace-period-section" :offset="[0, -15]">
      This can't be modified based on the policy your administrator set.
    </q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiDurationInput } from "@/components/common/runai-duration-input";
// Models
import type { ValidationRule } from "quasar";
import type { IntegerRules } from "@/swagger-models/assets-service-client";
// Utils
import { isMaxEqualOrHigherThenMin, isValueLowerOrEqualToMax } from "@/common/form.validators";
// Constants
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "termination-grace-period-section",
  components: {
    RunaiSection,
    RunaiDurationInput,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null] as PropType<number | null>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    isDisabled(): boolean {
      return this.readOnly || this.isDisabledByPolicy;
    },
    isDisabledByPolicy(): boolean {
      return this.policyRules?.canEdit === false;
    },
    timeframeErrorMessage(): string {
      const min = this.policyRules?.min || 0;
      const max = this.policyRules?.max;
      if (max !== undefined && max !== null) {
        return errorMessages.TIMEFRAME_BETWEEN_MIN_AND_MAX_SECONDS.replace("${min}", `${min}`).replace(
          "${max}",
          `${max}`,
        );
      }

      return errorMessages.VALID_TIME_FRAME.replace("0", `${min}`);
    },
    rules(): ValidationRule[] {
      const rules: ValidationRule[] = [this.validGracePeriod];
      if (!this.policyRules) return rules;

      this.policyRules.max && rules.push(this.notLargerThanRule);
      this.policyRules.min && rules.push(this.notSmallerThanRule);
      this.policyRules.step && rules.push(this.validStepRule);

      return rules;
    },
  },
  methods: {
    validGracePeriod(): string | boolean {
      if (this.modelValue === null || this.modelValue === undefined) {
        return true;
      }
      // If no policy rules exist, no validation is needed
      if (!this.policyRules || (!this.policyRules.min && !this.policyRules.max)) {
        return true;
      }

      const min = this.policyRules?.min;
      const max = this.policyRules?.max;

      const isAboveMin = min !== undefined && min !== null ? this.modelValue >= min : true;
      const isBelowMax = max !== undefined && max !== null ? this.modelValue <= max : true;

      return isAboveMin && isBelowMax ? true : this.timeframeErrorMessage;
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
  },
});
</script>
