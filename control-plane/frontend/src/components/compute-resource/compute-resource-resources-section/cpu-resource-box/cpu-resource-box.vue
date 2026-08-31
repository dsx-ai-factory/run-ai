<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="cpu-box row">
    <compute-resource-box :texts="texts" title="CPU compute per pod">
      <template #inputs>
        <runai-select
          outlined
          :options="cpuUnitOptions"
          :model-value="selectedCpuUnit"
          @update:model-value="updateCpuUnit"
          class="unit-select q-mr-sm"
          aid="cpu-unit-select"
          :rules="[() => true]"
          :disable="disable"
        />

        <policy-number-input
          outlined
          :min-value="0"
          label="Request"
          :model-value="(cpuCore.request || 0) * Number(selectedCpuUnit.value)"
          :policy-rules="cpuCoreRequestRules"
          @update:model-value="updateCpuCoreRequest"
          class="input-field"
          input-class="cpu-request-input"
          aid="cpu-request-input"
          bg-color="white"
          :rules="[isRequestEqualOrAboveZero]"
          no-error-icon
          :disable="disable"
        />

        <q-icon name="ph-bold ph-minus" size="6px" class="dash-icon q-mx-xs" />

        <policy-number-input
          outlined
          :min-value="0"
          stack-label
          label="Limit"
          :model-value="limitValue"
          :policy-rules="cpuCoreLimitRules"
          @update:model-value="updateLimitValue"
          class="input-field q-mr-sm"
          :placeholder="limitPlaceHolder"
          :disable="!limitEnabled || disable"
          input-class="cpu-limit-input"
          aid="cpu-limit-input"
          :rules="[isValidLimit]"
          no-error-icon
          bg-color="white"
        />

        <q-toggle
          v-model="limitEnabled"
          label="Limit"
          class="limit-toggle"
          aid="cpu-limit-toggle"
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
import type { ISelectOption } from "@/models/global.model";
import { cpuUnitOptions } from "@/models/compute-resource.model";
import type { ICpuCore } from "@/models/compute-resource.model";

// constants
import { errorMessages } from "@/common/error-message.constant";
import { RunaiSelect } from "@/components/common/runai-select";
import { ComputeResourceBox } from "../compute-resource-box";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/policy-service-client";
import { PolicyNumberInput } from "@/components/common/policy-number-input";

export const CONVERSION_VALUE = 1000;

export default defineComponent({
  components: {
    RunaiSelect,
    PolicyNumberInput,
    ComputeResourceBox,
  },
  emits: ["cpu-core-changed", "is-section-invalid"],
  props: {
    cpuCore: {
      type: Object as PropType<ICpuCore>,
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
      selectedCpuUnit: cpuUnitOptions[0] as ISelectOption,
      cpuUnitOptions,
      limitEnabled: !!this.cpuCore.limit,
    };
  },
  computed: {
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
    limitPlaceHolder(): string {
      return this.limitEnabled ? "" : "Unlimited";
    },
    limitValue(): null | number {
      if (!this.limitEnabled) return null;
      return is.number(this.cpuCore.limit) ? this.cpuCore.limit * Number(this.selectedCpuUnit.value) : null;
    },
    sectionInvalid(): boolean {
      if (!this.limitEnabled) return !is.number(this.cpuCore.request);
      if (!is.number(this.limitValue)) return true;
      return (this.cpuCore.request || 0) > this.limitValue;
    },
  },
  methods: {
    updateCpuUnit(val: ISelectOption | null): void {
      if (!val) return;
      this.selectedCpuUnit = val;
      const calculatedRequestValue =
        this.selectedCpuUnit.value === CONVERSION_VALUE
          ? (this.cpuCore.request || 0) / CONVERSION_VALUE
          : (this.cpuCore.request || 0) * CONVERSION_VALUE;

      let calculatedLimitValue: number | null = null;
      if (this.limitEnabled) {
        calculatedLimitValue =
          this.selectedCpuUnit.value === CONVERSION_VALUE
            ? Number(this.cpuCore.limit) / CONVERSION_VALUE
            : Number(this.cpuCore.limit) * CONVERSION_VALUE;
      }

      this.updateCpuCore({ request: calculatedRequestValue, limit: calculatedLimitValue });
    },
    updateCpuCoreRequest(val: number | string | null): void {
      const calculatedValue =
        this.selectedCpuUnit.value === CONVERSION_VALUE ? Number(val) / CONVERSION_VALUE : Number(val);

      if (!this.limitEnabled) {
        // limit disabled
        this.updateCpuCore({ request: calculatedValue, limit: undefined });
        return;
      }

      if (calculatedValue === 0) {
        this.updateCpuCore({ request: calculatedValue, limit: null });
      } else if (is.number(this.cpuCore.limit) && calculatedValue > this.cpuCore.limit) {
        this.updateCpuCore({ request: calculatedValue, limit: calculatedValue });
      } else {
        this.updateCpuCore({ ...this.cpuCore, request: calculatedValue });
      }
    },
    onLimitToggleChanged(): void {
      let limit: number | null | undefined;

      if (!this.limitEnabled) {
        limit = undefined;
      } else {
        limit = (this.cpuCore.request || 0) > 0 ? this.cpuCore.request : null;
      }

      this.updateCpuCore({ ...this.cpuCore, limit });
    },
    updateLimitValue(val: number | null | string): void {
      let calculatedLimitValue: number | null = null;

      if (val !== "") {
        calculatedLimitValue =
          this.selectedCpuUnit.value === CONVERSION_VALUE ? Number(val) / CONVERSION_VALUE : Number(val);
      }

      this.updateCpuCore({ ...this.cpuCore, limit: calculatedLimitValue });
    },
    updateCpuCore(cpuCore: ICpuCore): void {
      this.$emit("cpu-core-changed", cpuCore);
    },
    isValidLimit(val: number): boolean | string {
      if (!this.limitEnabled) return true;
      if (val === 0) return errorMessages.ENTER_A_LIMIT_ABOVE_ZERO;
      return val === null || (this.cpuCore.request || 0) > val ? errorMessages.LIMIT_EQUAL_OR_HIGHER_THAN_REQUEST : true;
    },
    isRequestEqualOrAboveZero(value: number): boolean | string {
      return value >= 0 || errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace("${min}", "0");
    },
  },
  watch: {
    sectionInvalid: {
      handler(invalid: boolean): void {
        this.$emit("is-section-invalid", invalid);
      },
      immediate: true,
    },
  },
});
</script>
<style lang="scss">
.cpu-box {
  .unit-select {
    & > div {
      width: 136px;
      background-color: $grey-3;
    }
  }

  .input-field {
    width: 135px;
  }

  // Adding padding bottom to align elements with quasar inputs
  .dash-icon,
  .limit-toggle {
    padding-bottom: 20px;
  }
}
</style>
