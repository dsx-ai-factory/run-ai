<template>
  <section class="memory-box">
    <compute-resource-box :texts="texts" title="CPU memory per pod">
      <template #inputs>
        <runai-select
          outlined
          :options="basicMemoryUnitOptions"
          :model-value="selectedMemoryUnit"
          @update:model-value="handleSelectedCpuMemory"
          class="field-col unit-select q-mr-sm"
          aid="memory-unit-select"
          :rules="[() => true]"
          :disable="disable"
        />
        <policy-quantity-input
          outlined
          type="number"
          :min-value="0"
          stack-label
          label="Request"
          :model-value="memoryAmount"
          :unit-size="cpuMemoryRequestUnit"
          @update:model-value="updateCpuMemoryRequest"
          class="input-field"
          input-class="cpu-memory-request-input"
          bg-color="white"
          :rules="[isRequestEqualOrAboveZero]"
          no-error-icon
          :policy-rules="policyRules?.cpuMemoryRequest"
          aid="memory-request-input"
          :disable="disable"
        />
        <q-icon name="ph-bold ph-minus" size="6px" class="dash-icon q-mx-xs" />
        <policy-quantity-input
          outlined
          type="number"
          :min-value="0"
          stack-label
          label="Limit"
          :model-value="limitValue"
          :unit-size="cpuMemoryRequestUnit"
          @update:model-value="updateCpuMemoryLimit"
          class="input-field q-mr-sm"
          input-class="cpu-memory-limit-input"
          :placeholder="limitPlaceHolder"
          :disable="!limitEnabled || disable"
          no-error-icon
          bg-color="white"
          :rules="[isValidLimit]"
          aid="memory-limit-input"
          :policy-rules="policyRules?.cpuMemoryLimit"
        />
        <q-toggle
          aid="memory-limit-toggle"
          v-model="limitEnabled"
          label="Limit"
          class="limit-toggle"
          @update:model-value="onLimitToggleChanged"
          :disable="disable"
        />
      </template>
    </compute-resource-box>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { is } from "quasar";

// models
import { basicMemoryUnitOptions, EMemoryUnitValue, type ISelectOption } from "@/models/global.model";
import type { ICpuMemory } from "@/models/compute-resource.model";

// constants
import { errorMessages } from "@/common/error-message.constant";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { ComputeResourceBox } from "../compute-resource-box";
import { PolicyQuantityInput } from "@/components/common/policy-quantity-input";
import { parseSizeString } from "@/utils/format.util";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/workloads-client";

export default defineComponent({
  components: {
    RunaiSelect,
    ComputeResourceBox,
    PolicyQuantityInput,
  },
  emits: ["cpu-memory-changed", "is-section-invalid"],
  props: {
    cpuMemory: {
      type: Object as PropType<ICpuMemory>,
      required: true,
    },
    texts: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      selectedMemoryUnit: basicMemoryUnitOptions[0] as ISelectOption,
      basicMemoryUnitOptions,
      limitEnabled: !!this.cpuMemory.limit,
    };
  },
  computed: {
    memoryAmount(): number {
      if (!this.cpuMemory.request) return 0;
      return this.getValueFromValueAndUnit(this.cpuMemory.request);
    },
    limitPlaceHolder(): string {
      return this.limitEnabled ? "" : "Unlimited";
    },
    limitValue(): null | number | undefined {
      if (!this.limitEnabled) return null;
      return !this.cpuMemory.limit ? null : this.getValueFromValueAndUnit(this.cpuMemory.limit);
    },
    sectionInvalid(): boolean {
      if (!this.limitEnabled) return !is.number(this.memoryAmount);
      if (!is.number(this.limitValue)) return true;
      return this.memoryAmount > this.limitValue;
    },
    cpuMemoryRequestUnit(): EMemoryUnitValue {
      const parsed = parseSizeString(this.cpuMemory.request || "");
      return (parsed?.[1] as EMemoryUnitValue) || EMemoryUnitValue.MB;
    },
  },
  methods: {
    getValueFromValueAndUnit(val: string): number {
      return Number(val.slice(0, val.length - 1));
    },
    getValueWithUnit(val: number): string {
      return `${Number(val).toString()}${this.selectedMemoryUnit.value}`;
    },
    updateCpuMemory(cpuMemory: ICpuMemory): void {
      this.$emit("cpu-memory-changed", cpuMemory);
    },
    updateCpuMemoryRequest(val: number | string | null): void {
      const request: number | null = val === null || val === undefined ? 0 : Number(val);
      if (this.limitEnabled && request === 0) {
        this.updateCpuMemory({ request: this.getValueWithUnit(request), limit: null });
      } else if (this.limitEnabled && is.number(this.limitValue) && request > this.limitValue) {
        this.updateCpuMemory({ request: this.getValueWithUnit(request), limit: this.getValueWithUnit(request) });
      } else {
        this.updateCpuMemory({ ...this.cpuMemory, request: this.getValueWithUnit(request) });
      }
    },
    handleSelectedCpuMemory(selectedOption: ISelectOption | null): void {
      if (!selectedOption) return;
      this.selectedMemoryUnit = selectedOption;
      const limit = this.limitEnabled ? this.getValueWithUnit(Number(this.limitValue)) : null;
      this.updateCpuMemory({ request: this.getValueWithUnit(this.memoryAmount), limit });
    },
    onLimitToggleChanged(): void {
      if (!this.limitEnabled) {
        this.updateCpuMemory({ ...this.cpuMemory, limit: undefined });
        return;
      }

      const limit = this.memoryAmount > 0 ? this.getValueWithUnit(this.memoryAmount) : null;
      this.updateCpuMemory({ ...this.cpuMemory, limit });
    },
    updateCpuMemoryLimit(val: number | string | null): void {
      const limit: number | null = val === null || val === undefined || val === "" ? null : Number(val);
      const limitWithUnit: null | string = limit === null ? null : this.getValueWithUnit(limit);

      this.updateCpuMemory({ ...this.cpuMemory, limit: limitWithUnit });
    },
    isValidLimit(val: number): boolean | string {
      if (!this.limitEnabled) return true;
      if (val === 0) return errorMessages.ENTER_A_LIMIT_ABOVE_ZERO;
      return val === null || this.memoryAmount > val ? errorMessages.LIMIT_EQUAL_OR_HIGHER_THAN_REQUEST : true;
    },
    isRequestEqualOrAboveZero(value: number): boolean | string {
      return value >= 0 || errorMessages.VALUE_EQUAL_ABOVE_ZERO;
    },
  },
  watch: {
    cpuMemory: {
      handler(cpuMemory: ICpuMemory) {
        const request = cpuMemory.request;
        if (!request) {
          this.selectedMemoryUnit = basicMemoryUnitOptions[0];
          return;
        }

        this.selectedMemoryUnit =
          this.basicMemoryUnitOptions.find((option: ISelectOption) => option.value === request[request.length - 1]) ||
          basicMemoryUnitOptions[0];
      },
      immediate: true,
    },
    sectionInvalid: {
      handler(invalid: boolean) {
        this.$emit("is-section-invalid", invalid);
      },
      immediate: true,
    },
  },
});
</script>
<style lang="scss">
.memory-box {
  .input-field {
    width: 135px;
  }

  .unit-select {
    & > div {
      width: 136px;
      background-color: $grey-3;
    }
  }

  // Adding padding bottom to align elements with quasar inputs
  .dash-icon,
  .limit-toggle {
    padding-bottom: 20px;
  }
}
</style>
