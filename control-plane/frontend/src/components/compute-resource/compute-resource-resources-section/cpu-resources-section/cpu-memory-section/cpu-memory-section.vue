<template>
  <section class="cpu-memory-section row items-center q-gutter-x-md q-mb-lg">
    <runai-select
      aid="memory-unit-select"
      class="col-3"
      label="CPU memory"
      :options="basicMemoryUnitOptions"
      :model-value="selectedMemoryUnit"
      @update:model-value="handleSelectedCpuMemory"
      :rules="[() => true]"
      :disable="sectionDisabled"
    />

    <policy-quantity-input
      aid="memory-request-input"
      class="col-3"
      input-class="cpu-memory-request-input"
      label="Request"
      placeholder="0"
      stack-label
      :min-value="0"
      :max-value="maxValue"
      :step="step"
      :rules="[isValidRequest]"
      :model-value="memoryAmount"
      :unit-size="cpuMemoryRequestUnit"
      @update:model-value="updateCpuMemoryRequest"
      :policy-rules="policyRules?.cpuMemoryRequest"
      :disable="sectionDisabled"
    />
    <policy-quantity-input
      aid="memory-limit-input"
      class="col-3"
      input-class="cpu-memory-limit-input"
      label="Limit"
      stack-label
      :min-value="0"
      :max-value="maxValue"
      :step="step"
      :model-value="limitValue"
      :unit-size="cpuMemoryRequestUnit"
      @update:model-value="updateCpuMemoryLimit"
      placeholder="Auto"
      :disable="sectionDisabled"
      :policy-rules="policyRules?.cpuMemoryLimit"
      :rules="[isValidLimit]"
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
import { PolicyQuantityInput } from "@/components/common/policy-quantity-input";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
// Models
import type { ICpuMemory } from "@/models/compute-resource.model";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/policy-service-client";
import { basicMemoryUnitOptions, EMemoryUnitValue, type ISelectOption } from "@/models/global.model";
import { parseSizeString } from "@/utils/format.util";
import { is } from "quasar";
import { errorMessagesUtils } from "@/common/error-message.constant";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "cpu-memory-section",
  components: {
    RunaiSelect,
    PolicyQuantityInput,
    RunaiTooltip,
  },
  emits: {
    [EEmitNames.UpdateModelValue]: (_: ICpuMemory) => {
      return true;
    },
  },
  props: {
    modelValue: {
      type: Object as PropType<ICpuMemory>,
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
      selectedMemoryUnit: basicMemoryUnitOptions[0] as ISelectOption,
      basicMemoryUnitOptions,
    };
  },
  computed: {
    memoryAmount(): number | null | undefined {
      if (this.modelValue.request === null || this.modelValue.request === undefined || this.modelValue.request === "") {
        return null;
      }

      if (!this.modelValue.request) return 0;
      return this.getValueFromValueAndUnit(this.modelValue.request);
    },
    limitValue(): null | number | undefined {
      if (this.modelValue.limit === null) return null;
      return !this.modelValue.limit ? null : this.getValueFromValueAndUnit(this.modelValue.limit);
    },
    cpuMemoryRequestUnit(): EMemoryUnitValue {
      const parsed = parseSizeString(this.modelValue.request || "");
      return (parsed?.[1] as EMemoryUnitValue) || EMemoryUnitValue.MB;
    },
    limitEnabled(): boolean {
      return this.modelValue.limit !== null && this.modelValue.limit !== undefined;
    },
    maxValue(): number {
      return this.selectedMemoryUnit.value === EMemoryUnitValue.GB ? 10000 : 10000000;
    },
    step(): number {
      return this.selectedMemoryUnit.value === EMemoryUnitValue.GB ? 1 : 100;
    },
  },
  methods: {
    handleSelectedCpuMemory(selectedOption: ISelectOption | null): void {
      if (!selectedOption) return;
      this.selectedMemoryUnit = selectedOption;
      const limit = this.modelValue.limit !== undefined ? this.getValueWithUnit(Number(this.limitValue)) : undefined;
      const request =
        this.memoryAmount !== undefined && this.memoryAmount !== null
          ? this.getValueWithUnit(this.memoryAmount)
          : undefined;
      this.updateCpuMemory({ request: request, limit });
    },
    getValueWithUnit(val: number): string {
      let calculatedVal = "";
      if (this.selectedMemoryUnit.value === EMemoryUnitValue.GB) {
        calculatedVal = (+val).toFixed(2);
      } else {
        calculatedVal = (+val).toFixed(0);
      }

      return `${Number(calculatedVal).toString()}${this.selectedMemoryUnit.value}`;
    },
    updateCpuMemory(cpuMemory: ICpuMemory): void {
      this.$emit(EEmitNames.UpdateModelValue, cpuMemory);
    },
    getValueFromValueAndUnit(val: string): number {
      return Number(val.slice(0, val.length - 1));
    },
    updateCpuMemoryRequest(val: number | string | null): void {
      if (val === null || val === undefined || val === "") {
        this.updateCpuMemory({ ...this.modelValue, request: undefined });
        return;
      }
      const request = Number(val);
      if (this.limitEnabled && request === 0) {
        this.updateCpuMemory({ request: this.getValueWithUnit(request), limit: null });
      } else if (this.limitEnabled && is.number(this.limitValue) && request > this.limitValue) {
        this.updateCpuMemory({ request: this.getValueWithUnit(request), limit: this.getValueWithUnit(request) });
      } else {
        this.updateCpuMemory({ ...this.modelValue, request: this.getValueWithUnit(request) });
      }
    },
    updateCpuMemoryLimit(val: number | string | null): void {
      const limit: number | undefined = val === null || val === undefined || val === "" ? undefined : Number(val);
      const limitWithUnit: undefined | string = limit === undefined ? undefined : this.getValueWithUnit(limit);

      this.updateCpuMemory({ ...this.modelValue, limit: limitWithUnit });
    },
    isValidRequest(val: number): boolean | string {
      if (val === null || val === undefined) return true;
      if (val < 0 || val > this.maxValue) {
        return errorMessagesUtils.VALUE_FROM_TO(0, this.maxValue);
      }
      return true;
    },
    isValidLimit(val: number): boolean | string {
      if (!this.limitEnabled) return true;
      if (val < (this.memoryAmount || 0) || val > this.maxValue) {
        return errorMessagesUtils.LIMIT_FROM_REQUEST_TO_MAX(this.maxValue);
      }
      return true;
    },
  },
  watch: {
    modelValue: {
      // the gist of this watcher is to make sure the selectedMemoryUnit is always correct based on the actual value
      handler(newVal: ICpuMemory, oldVal: ICpuMemory) {
        if (newVal?.request !== oldVal?.request) {
          this.selectedMemoryUnit =
            this.basicMemoryUnitOptions.find((option) => option.value === this.cpuMemoryRequestUnit) ||
            this.basicMemoryUnitOptions[0];
        }
      },
      immediate: true,
    },
  },
});
</script>
