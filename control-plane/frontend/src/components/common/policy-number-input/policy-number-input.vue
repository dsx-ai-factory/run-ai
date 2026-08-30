<template>
  <runai-tooltip-wrapper
    :display-tooltip="disableInput"
    :tooltip-text="disabledTooltipText"
    anchor="bottom middle"
    self="bottom middle"
  >
    <q-input
      type="number"
      class="policy-number-input"
      :aid="aid"
      :min="min"
      :max="max"
      :step="step"
      :class="elementClasses"
      :style="inlineStyle"
      :model-value="modelValue"
      :disable="disableInput"
      :stack-label="stackLabel"
      :placeholder="placeholder"
      :label="label"
      :input-class="inputClass"
      :rules="rulesWithPolicy"
      :ref="customRef"
      :bg-color="bgColor"
      :outlined="outlined"
      :reactive-rules="reactiveRules"
      no-error-icon
      :label-slot="enableLabelTooltip"
      @update:model-value="updateValue"
      @wheel.prevent
      @keydown="onKeyDown"
    >
      <template v-if="enableLabelTooltip" v-slot:label>
        <span class="ra-input-label">
          {{ label }}
          <q-tooltip anchor="top middle" :offset="[0, 40]">
            {{ label }}
          </q-tooltip>
        </span>
      </template>
    </q-input>
    <q-tooltip class="explanatory-tooltip" v-if="showExplanatoryTooltip" max-width="300px">{{ tooltip }}</q-tooltip>
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { NumberRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import type { QInput, ValidationRule } from "quasar";
import { isMaxEqualOrHigherThenMin, isValueLowerOrEqualToMax } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { isNullOrUndefined } from "@/utils/common.util";

export enum AllowedKeys {
  Backspace = "Backspace",
  Delete = "Delete",
  ArrowLeft = "ArrowLeft",
  ArrowRight = "ArrowRight",
  ArrowUp = "ArrowUp",
  ArrowDown = "ArrowDown",
  Tab = "Tab",
  Dot = ".",
}

// validate floating-point numbers with up to 5 digits before the decimal point and up to 2 digits after
const FLOATING_NUMBER_REGEX = /^\d{0,5}(\.\d{0,2})?$/;

// validate integers with up to 7 digits
const INTEGER_NUMBER_REGEX = /^\d{0,7}?$/;

export default defineComponent({
  name: "policy-number-input",
  components: {
    RunaiTooltipWrapper,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<NumberRules | null>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    tooltip: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    maxValue: {
      type: Number as PropType<number>,
      required: false,
    },
    minValue: {
      type: Number as PropType<number>,
      required: false,
    },
    inputStep: {
      type: Number as PropType<number>,
      required: false,
    },
    inputClass: {
      type: String as PropType<string>,
      required: false,
    },
    inlineStyle: {
      type: Object as PropType<Record<string, string>>,
      required: false,
    },
    elementClasses: {
      type: String as PropType<string>,
      required: false,
    },
    dense: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
    customRef: {
      type: String as PropType<string>,
      required: false,
      default: "policyNumberInputRef",
    },
    outlined: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    bgColor: {
      type: String as PropType<string>,
      required: false,
    },
    reactiveRules: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    customRequiredMessage: {
      type: String as PropType<string>,
      required: false,
    },
    enableLabelTooltip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isInteger: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    disableInput(): boolean {
      if (this.policyRules?.canEdit === false) return true;
      return !!this.disable;
    },
    disabledTooltipText(): string {
      if (this.policyRules?.canEdit === false) {
        return errorMessages.CANT_BE_MODIFIED_POLICY;
      }
      return "";
    },
    max(): number | undefined {
      return typeof this.policyRules?.max === "number" ? this.policyRules.max : this.maxValue;
    },
    min(): number | undefined {
      return typeof this.policyRules?.min === "number" ? this.policyRules.min : this.minValue;
    },
    step(): number | undefined {
      return this.policyRules?.step || this.inputStep || undefined;
    },
    rulesWithPolicy(): ValidationRule[] {
      const rules = this.rules || [];
      this.policyRules?.required && rules.push(this.requiredRule);
      this.max && rules.push(this.maxRule);
      !isNullOrUndefined(this.min) && rules.push(this.minRule);
      this.policyRules?.step && rules.push(this.stepRule);

      return rules;
    },
    showExplanatoryTooltip(): boolean {
      if (this.policyRules?.canEdit === false) return false;
      return !!this.tooltip;
    },
    valueErrorMessage(): string {
      const isMinDefined = !isNullOrUndefined(this.min);
      const isMaxDefined = !isNullOrUndefined(this.max);
      const isStepDefined = !isNullOrUndefined(this.step);

      // product wanted a combo message if either min and max are defined or min and step are defined

      if (isMinDefined && isMaxDefined && isStepDefined) {
        return errorMessages.ENTER_A_VALUE_BETWEEN_WITH_STEP.replace("${min}", `${this.min}`)
          .replace("${max}", `${this.max}`)
          .replace("${step}", `${this.step}`);
      } else if (isMinDefined && isMaxDefined) {
        return errorMessages.ENTER_A_VALUE_BETWEEN.replace("${min}", `${this.min}`).replace("${max}", `${this.max}`);
      } else if (isMinDefined && isStepDefined) {
        return errorMessages.VALUE_MUST_FOLLOW_STEPS_STARTING_MIN.replace("${min}", `${this.min}`)
          .replace("${max}", `${this.max}`)
          .replace("${step}", `${this.step}`);
      }

      return "";
    },
    decimalRegex() {
      if (this.isInteger) {
        return INTEGER_NUMBER_REGEX;
      }
      return FLOATING_NUMBER_REGEX;
    },
  },
  methods: {
    requiredRule(val: number): boolean | string {
      return !isNullOrUndefined(val) || this.customRequiredMessage || errorMessages.REQUIRED_FIELD;
    },
    maxRule(val: number): boolean | string {
      if ((val === null || val === undefined) && !this.policyRules?.required) return true;
      const max = this.max || Infinity;
      // if combo message is defined we don't want to show the default message
      return (
        isValueLowerOrEqualToMax(val, max) ||
        this.valueErrorMessage ||
        errorMessages.ENTER_A_VALUE_BELOW_MAX.replace("${max}", `${max}`)
      );
    },
    minRule(val: number): boolean | string {
      if ((val === null || val === undefined) && !this.policyRules?.required) return true;

      const min = this.min || 0;
      // if combo message is defined we don't want to show the default message
      return (
        isMaxEqualOrHigherThenMin(val, min) ||
        this.valueErrorMessage ||
        errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace("${min}", `${min}`)
      );
    },
    stepRule(val: number): boolean | string {
      if ((val === null || val === undefined) && !this.policyRules?.required) return true;
      const min = this.min || 0;
      const step = this.step || 1;
      return (
        (val - min) % step === 0 ||
        this.valueErrorMessage ||
        errorMessages.VALUE_MUST_FOLLOW_STEPS_OF.replace("${step}", `${step}`)
      );
    },
    updateValue(val: string | number | null): void {
      this.$emit("update:model-value", val);
    },
    onKeyDown(event: KeyboardEvent) {
      if (event.key === AllowedKeys.Dot && this.isInteger) {
        event.preventDefault();
      }

      const input = event.target as HTMLInputElement;
      const value = input.value;

      // check if the key being pressed is a dot
      if (event.key === AllowedKeys.Dot) {
        // check if there is already a dot in the input
        if (value.indexOf(".") !== -1) {
          event.preventDefault();
        } else {
          // allow the dot to be typed if there is not already one in the input
          return;
        }
      }

      if (Object.values(AllowedKeys).includes(event.key as AllowedKeys)) {
        // allow navigation and deletion
        return;
      }
    },
  },
  watch: {
    policyRules: {
      handler(): void {
        (this.$refs[this.customRef] as QInput)?.validate();
      },
    },
  },
});
</script>
<style lang="scss" scoped>
.ra-input-label {
  pointer-events: stroke;
}
</style>
