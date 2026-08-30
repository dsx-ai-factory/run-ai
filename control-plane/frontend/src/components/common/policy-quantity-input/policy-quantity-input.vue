<template>
  <runai-tooltip-wrapper
    :display-tooltip="disableInput"
    :tooltip-text="disabledTooltipText"
    anchor="bottom middle"
    self="bottom middle"
  >
    <q-input
      :debounce="300"
      type="number"
      class="policy-quantity-input"
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
      no-error-icon
      @update:model-value="updateValue"
    />
    <q-tooltip class="explanatory-tooltip" v-if="showExplanatoryTooltip" max-width="300px">{{ tooltip }}</q-tooltip>
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { QuantityRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import type { QInput, ValidationRule } from "quasar";
import { isMaxEqualOrHigherThenMin, isValueLowerOrEqualToMax } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { isNullOrUndefined } from "@/utils/common.util";
import type { EMemoryUnitValue } from "@/models/global.model";
import { resourceUtil } from "@/utils/resource.util";

export default defineComponent({
  name: "policy-quantity-input",
  components: {
    RunaiTooltipWrapper,
  },
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
    unitSize: {
      type: String as PropType<EMemoryUnitValue>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<QuantityRules | null>,
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
    step: {
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
      default: "policyQuantityInputRef",
    },
    outlined: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    bgColor: {
      type: String as PropType<string>,
      required: false,
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
      if (!this.policyRules?.max) return this.maxValue;
      try {
        const maxInBytes = resourceUtil.parseK8sResourceSizeToBytes(this.policyRules?.max || "");
        const maxInUnitSize = resourceUtil.convertBytesToK8sResourceSizeUnit(maxInBytes, this.unitSize);
        return maxInUnitSize;
      } catch (e) {
        console.error(e);
        return this.maxValue;
      }
    },
    min(): number | undefined {
      if (!this.policyRules?.min) return this.minValue;
      try {
        const minInBytes = resourceUtil.parseK8sResourceSizeToBytes(this.policyRules?.min || "");
        const minInUnitSize = resourceUtil.convertBytesToK8sResourceSizeUnit(minInBytes, this.unitSize);
        return minInUnitSize;
      } catch (e) {
        console.error(e);
        return this.minValue;
      }
    },
    rulesWithPolicy(): ValidationRule[] {
      const rules = this.rules || [];
      this.policyRules?.required && rules.push(this.requiredRule);
      this.max && rules.push(this.maxRule);
      !isNullOrUndefined(this.min) && rules.push(this.minRule);

      return rules;
    },
    showExplanatoryTooltip(): boolean {
      if (this.policyRules?.canEdit === false) return false;
      return !!this.tooltip;
    },
  },
  methods: {
    requiredRule(val: number): boolean | string {
      return !isNullOrUndefined(val) || errorMessages.REQUIRED_FIELD;
    },
    maxRule(val: number): boolean | string {
      const max = this.max || Infinity;
      return isValueLowerOrEqualToMax(val, max) || errorMessages.ENTER_A_VALUE_BELOW_MAX.replace("${max}", `${max}`);
    },
    minRule(val: number): boolean | string {
      const min = this.min || 0;
      return isMaxEqualOrHigherThenMin(val, min) || errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace("${min}", `${min}`);
    },
    updateValue(val: string | number | null): void {
      this.$emit("update:model-value", val);
    },
  },
  watch: {
    unitSize: {
      handler(): void {
        (this.$refs[this.customRef] as QInput)?.validate();
      },
    },
  },
});
</script>
