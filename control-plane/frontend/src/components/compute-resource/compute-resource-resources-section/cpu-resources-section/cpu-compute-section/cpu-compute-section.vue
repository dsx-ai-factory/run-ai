<template>
  <section class="cpu-compute-section row items-center q-gutter-x-md q-mb-sm">
    <runai-select
      label="CPU compute"
      class="col-3"
      aid="cpu-unit-select"
      :options="cpuUnitOptions"
      :model-value="selectedCpuUnit"
      @update:model-value="updateCpuUnit"
      :rules="[() => true]"
      :disable="sectionDisabled"
    />

    <policy-number-input
      aid="cpu-request-input"
      class="col-3"
      input-class="cpu-request-input"
      label="Request"
      stack-label
      :min-value="0"
      placeholder="0"
      :max-value="maxValue"
      :input-step="step"
      :rules="[isValidCpuRequest]"
      reactive-rules
      :model-value="requestValue"
      :policy-rules="cpuCoreRequestRules"
      @update:model-value="updateRequest"
      :disable="sectionDisabled"
    />

    <policy-number-input
      aid="cpu-limit-input"
      class="col-3"
      input-class="cpu-limit-input"
      label="Limit"
      :min-value="0"
      :max-value="maxValue"
      :input-step="step"
      stack-label
      placeholder="Auto"
      :model-value="limitValue"
      :rules="[isValidLimit]"
      :policy-rules="cpuCoreLimitRules"
      @update:model-value="updateLimit"
      :disable="sectionDisabled"
    />

    <runai-tooltip
      tooltip-text="When the limit is set to Auto, the workload can use up to the node's maximum available CPU resources"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
// Models
import { cpuUnitOptions, type ICpuCore } from "@/models/compute-resource.model";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/policy-service-client";
import type { ISelectOption } from "@/models/global.model";
import { is } from "quasar";
import { errorMessagesUtils } from "@/common/error-message.constant";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export const CONVERSION_VALUE = 1000;

export default defineComponent({
  name: "cpu-compute-section",
  components: {
    RunaiSelect,
    PolicyNumberInput,
    RunaiTooltip,
  },
  emits: {
    [EEmitNames.UpdateModelValue]: (_: ICpuCore) => {
      return true;
    },
  },
  props: {
    modelValue: {
      type: Object as PropType<ICpuCore>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      cpuUnitOptions,
      selectedCpuUnit: cpuUnitOptions[1] as ISelectOption,
    };
  },
  computed: {
    requestValue(): null | undefined | number {
      return is.number(this.modelValue.request) ? this.modelValue.request * Number(this.selectedCpuUnit.value) : null;
    },
    limitEnabled(): boolean {
      return this.modelValue.limit !== null && this.modelValue.limit !== undefined;
    },
    cpuCoreRequestRules() {
      if (!this.policyRules?.cpuCoreRequest) return undefined;

      const unitAdjustedRules = {
        ...this.policyRules.cpuCoreRequest,
      };

      if (unitAdjustedRules.min) unitAdjustedRules.min = unitAdjustedRules.min * Number(this.selectedCpuUnit.value);
      if (unitAdjustedRules.max) unitAdjustedRules.max = unitAdjustedRules.max * Number(this.selectedCpuUnit.value);
      if (unitAdjustedRules.step) unitAdjustedRules.step = unitAdjustedRules.step * Number(this.selectedCpuUnit.value);

      return unitAdjustedRules;
    },
    cpuCoreLimitRules() {
      if (!this.policyRules?.cpuCoreLimit) return undefined;

      const unitAdjustedRules = {
        ...this.policyRules.cpuCoreLimit,
      };

      if (unitAdjustedRules.min) unitAdjustedRules.min = unitAdjustedRules.min * Number(this.selectedCpuUnit.value);
      if (unitAdjustedRules.max) unitAdjustedRules.max = unitAdjustedRules.max * Number(this.selectedCpuUnit.value);
      if (unitAdjustedRules.step) unitAdjustedRules.step = unitAdjustedRules.step * Number(this.selectedCpuUnit.value);

      return unitAdjustedRules;
    },
    limitValue(): null | number {
      return is.number(this.modelValue.limit) ? this.modelValue.limit * Number(this.selectedCpuUnit.value) : null;
    },
    step(): number {
      return this.selectedCpuUnit.value === CONVERSION_VALUE ? 100 : 1;
    },
    maxValue(): number {
      return 1000 * Number(this.selectedCpuUnit.value);
    },
  },
  methods: {
    updateCpuUnit(val: ISelectOption | null): void {
      if (!val) return;
      let calculatedRequestValue: number | undefined = undefined;
      if (this.modelValue.request !== null && this.modelValue.request !== undefined) {
        this.selectedCpuUnit = val;
        calculatedRequestValue =
          this.selectedCpuUnit.value === CONVERSION_VALUE
            ? this.modelValue.request / CONVERSION_VALUE
            : this.modelValue.request * CONVERSION_VALUE;
      }

      let calculatedLimitValue: number | undefined = undefined;
      if (this.limitEnabled) {
        calculatedLimitValue =
          this.selectedCpuUnit.value === CONVERSION_VALUE
            ? Number(this.modelValue.limit) / CONVERSION_VALUE
            : Number(this.modelValue.limit) * CONVERSION_VALUE;
      }

      this.$emit(EEmitNames.UpdateModelValue, { request: calculatedRequestValue, limit: calculatedLimitValue });
    },
    updateRequest(request: string): void {
      if (request === "") {
        this.$emit(EEmitNames.UpdateModelValue, { ...this.modelValue, request: undefined });
        return;
      }

      if (this.selectedCpuUnit.value === CONVERSION_VALUE) {
        request = Number(request).toFixed(0);
      } else {
        request = Number(request).toFixed(2);
      }

      const calculatedValue =
        this.selectedCpuUnit.value === CONVERSION_VALUE ? Number(request) / CONVERSION_VALUE : Number(request);

      if (!this.limitEnabled) {
        // limit disabled
        this.$emit(EEmitNames.UpdateModelValue, { request: calculatedValue, limit: undefined });
        return;
      }

      if (calculatedValue === 0) {
        this.$emit(EEmitNames.UpdateModelValue, { request: calculatedValue, limit: null });
      } else if (is.number(this.modelValue.limit) && calculatedValue > this.modelValue.limit) {
        this.$emit(EEmitNames.UpdateModelValue, { request: calculatedValue, limit: calculatedValue });
      } else {
        this.$emit(EEmitNames.UpdateModelValue, { ...this.modelValue, request: calculatedValue });
      }
    },
    updateLimit(limit: string): void {
      let calculatedLimitValue: number | undefined = undefined;

      if (limit !== "") {
        if (this.selectedCpuUnit.value === CONVERSION_VALUE) {
          limit = Number(limit).toFixed(0);
        } else {
          limit = Number(limit).toFixed(2);
        }

        calculatedLimitValue =
          this.selectedCpuUnit.value === CONVERSION_VALUE ? Number(limit) / CONVERSION_VALUE : Number(limit);
      }

      this.$emit(EEmitNames.UpdateModelValue, { ...this.modelValue, limit: calculatedLimitValue });
    },
    isValidLimit(val: number): boolean | string {
      if (!this.limitEnabled) return true;
      if (val < (this.modelValue.request || 0) || val > this.maxValue) {
        return errorMessagesUtils.LIMIT_FROM_REQUEST_TO_MAX(this.maxValue);
      }
      return true;
    },
    isValidCpuRequest(val: number): boolean | string {
      if (val === undefined || val === null) return true;
      if (val < 0 || val > this.maxValue) {
        return errorMessagesUtils.VALUE_FROM_TO(0, this.maxValue);
      }
      return true;
    },
  },
});
</script>
