<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="gpu-resource-box" :class="sectionFlexOrder">
    <compute-resource-box :texts="texts" title="GPU memory per device">
      <template #inputs>
        <runai-select
          outlined
          :model-value="selectedRequestType"
          @update:model-value="updateGpuRequestType"
          :options="gpuRequestTypeOptions"
          class="unit-select q-mr-sm"
          :rules="[() => true]"
          :disable="disable"
          aid="gpu-type-select"
          unclearable
          :class="oprClass"
          :policy-rules="requestTypeRules"
        />

        <policy-number-input
          v-if="showGpuPortionInput"
          outlined
          class="input-field"
          input-class="gpu-request-input"
          aid="gpu-request-input"
          :model-value="gpuProtionRequest"
          label="Request"
          bg-color="white"
          :min-value="1"
          :max-value="inputMax"
          @update:model-value="updateGpuRequest"
          :rules="[isRequestEqualOrAboveOne]"
          no-error-icon
          type="number"
          :disable="disable"
          :policy-rules="gpuPortionRules"
        />
        <policy-quantity-input
          v-else
          outlined
          class="input-field"
          input-class="gpu-request-input"
          aid="gpu-request-input"
          :model-value="gpuMemoryRequest"
          :unit-size="gpuMemoryRequestUnit"
          label="Request"
          bg-color="white"
          :min-value="0"
          :max-value="inputMax"
          @update:model-value="updateGpuRequest"
          :rules="[isRequestAboveZero]"
          no-error-icon
          type="number"
          :disable="disable"
          :policy-rules="policyRules?.gpuMemoryRequest"
        />
        <template v-if="overProvisionEnabled">
          <q-icon name="ph-bold ph-minus" size="6px" class="dash-icon q-mx-xs" />
          <policy-number-input
            v-if="showGpuPortionInput"
            outlined
            class="input-field q-mr-sm"
            input-class="gpu-limit-input"
            :model-value="portionGpuLimit"
            :min-value="0"
            :max-value="inputMax"
            label="Limit"
            bg-color="white"
            :disable="disable || !isLimitEnabled"
            @update:model-value="updateGpuLimit"
            no-error-icon
            lazy-rules
            reactive-rules
            :rules="[isLimitAboveZero, isValidLimit]"
            type="number"
            aid="gpu-limit-input"
            :policy-rules="gpuPortionLimitRules"
          />
          <policy-quantity-input
            v-else
            outlined
            :unit-size="gpuMemoryRequestUnit"
            class="input-field q-mr-sm"
            input-class="gpu-limit-input"
            :model-value="memoryGpuLimit"
            :min-value="0"
            :max-value="inputMax"
            label="Limit"
            bg-color="white"
            :disable="disable || !isLimitEnabled"
            @update:model-value="updateGpuLimit"
            no-error-icon
            lazy-rules
            reactive-rules
            :rules="[isLimitAboveZero, isValidLimit]"
            type="number"
            aid="gpu-limit-input"
            :policy-rules="policyRules?.gpuMemoryLimit"
          />
          <q-toggle
            size="md"
            :disable="disable"
            :model-value="isLimitEnabled"
            label="Limit"
            class="limit-toggle"
            @update:model-value="updateLimitEnabled"
            aid="gpu-limit-toggle"
          />
        </template>
      </template>
    </compute-resource-box>
  </section>
</template>

<script lang="ts">
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import { type ISelectOption, EMemoryUnitValue } from "@/models/global.model";
import type { NumberRules } from "@/swagger-models/assets-service-client";
import { GpuRequestType } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { gpuRequestTypeOptions } from "@/models/compute-resource.model";
import { errorMessages } from "@/common/error-message.constant";
import { useSettingStore } from "@/stores/setting.store";
import { ComputeResourceBox } from "../compute-resource-box";
import type {
  SupersetDefaultsAllOfCompute,
  SupersetRulesAllOfCompute,
  GpuRequestOptionsOptionsInner,
} from "@/swagger-models/policy-service-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { PolicyQuantityInput } from "@/components/common/policy-quantity-input";
import { parseSizeString } from "@/utils/format.util";

const GPU_PORTION_FACTOR = 100;
const MAX_PERCENTAGE = 100;
const DEFAULT_MB_VALUE = "0M";
const DEFAULT_GB_VALUE = "0G";

export default defineComponent({
  name: "gpu-resource-box",
  components: {
    RunaiSelect,
    ComputeResourceBox,
    PolicyNumberInput,
    PolicyQuantityInput,
  },
  emits: ["update-gpu-request", "section-invalid"],
  props: {
    gpuRequest: {
      type: Object as PropType<IGpuRequestModel>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    texts: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<SupersetDefaultsAllOfCompute | null>,
      required: false,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      gpuRequestTypeOptions,
    };
  },
  computed: {
    gpuPortionRules() {
      if (!this.policyRules?.gpuPortionRequest) return undefined;
      return this.convertGpuPortionRules(this.policyRules.gpuPortionRequest);
    },
    gpuPortionLimitRules() {
      if (!this.policyRules?.gpuPortionLimit) return undefined;
      return this.convertGpuPortionRules(this.policyRules.gpuPortionLimit);
    },
    requestTypeRules(): IGenericSelectPolicyRules | undefined {
      const options = this.policyRules?.gpuRequestType?.options?.reduce(
        (acc: IGenericSelectPolicyRules["options"], o: GpuRequestOptionsOptionsInner) => {
          if (o.value === GpuRequestType.Memory) {
            acc = [
              ...(acc || []),
              { displayed: o.displayed, value: EMemoryUnitValue.MB },
              { displayed: o.displayed, value: EMemoryUnitValue.GB },
            ];
          } else acc && acc.push(o);
          return acc;
        },
        [],
      );
      return this.policyRules?.gpuRequestType
        ? { ...this.policyRules?.gpuRequestType, options, isClosedList: true }
        : undefined;
    },
    isLimitEnabled(): boolean {
      if (this.isGpuMemoryRequest) {
        return this.gpuRequest.gpuMemoryLimit !== null && this.gpuRequest.gpuMemoryLimit !== undefined;
      } else {
        return this.gpuRequest.gpuPortionLimit !== null && this.gpuRequest.gpuPortionLimit !== undefined;
      }
    },
    inputMax(): number | undefined {
      return this.isGpuMemoryRequest ? undefined : MAX_PERCENTAGE;
    },
    selectedRequestType(): string {
      if (this.isGpuMemoryRequest) {
        if (this.gpuRequest.gpuMemoryRequest?.endsWith(EMemoryUnitValue.MB)) {
          return EMemoryUnitValue.MB;
        } else {
          return EMemoryUnitValue.GB;
        }
      }
      return this.gpuRequest.gpuRequestType || GpuRequestType.Portion;
    },
    portionLimitEnabled(): boolean {
      return !!this.gpuRequest.gpuPortionLimit || this.gpuRequest.gpuPortionLimit === 0;
    },
    MemoryLimitEnabled(): boolean {
      return (
        !!this.gpuRequest.gpuMemoryLimit ||
        this.gpuRequest.gpuMemoryLimit === DEFAULT_MB_VALUE ||
        this.gpuRequest.gpuMemoryLimit === DEFAULT_GB_VALUE
      );
    },
    sectionFlexOrder(): string {
      return !this.overProvisionEnabled ? "justify-start" : "justify-between";
    },
    gpuProtionRequest(): number | null | undefined {
      if (this.gpuRequest.gpuRequestType === GpuRequestType.Portion) {
        return +((this.gpuRequest.gpuPortionRequest || 0) * GPU_PORTION_FACTOR).toFixed(0);
      }
      return null;
    },
    gpuMemoryRequest(): number | null | undefined {
      if (this.isGpuMemoryRequest) {
        return this.gpuRequest.gpuMemoryRequest ? parseFloat(this.gpuRequest.gpuMemoryRequest) : null;
      }
      return null;
    },
    gpuMemoryRequestUnit(): EMemoryUnitValue {
      const parsed = parseSizeString(this.gpuRequest.gpuMemoryRequest || "");
      return (parsed?.[1] as EMemoryUnitValue) || EMemoryUnitValue.MB;
    },
    portionGpuLimit(): number | null | undefined {
      if (this.gpuRequest.gpuRequestType === GpuRequestType.Portion) {
        return this.gpuRequest.gpuPortionLimit || this.gpuRequest.gpuPortionLimit === 0
          ? +(this.gpuRequest.gpuPortionLimit * GPU_PORTION_FACTOR).toFixed(0)
          : +((this.gpuRequest.gpuPortionRequest || 0) * GPU_PORTION_FACTOR).toFixed(0);
      }
      return null;
    },
    memoryGpuLimit(): number | null | undefined {
      if (this.isGpuMemoryRequest) {
        return this.gpuRequest.gpuMemoryLimit
          ? parseFloat(this.gpuRequest.gpuMemoryLimit)
          : parseFloat(this.gpuRequest.gpuMemoryRequest || "0");
      }
      return null;
    },
    isGpuMemoryRequest(): boolean {
      return this.gpuRequest.gpuRequestType === GpuRequestType.Memory;
    },
    showGpuPortionInput(): boolean {
      return this.selectedRequestType === GpuRequestType.Portion;
    },
    overProvisionEnabled(): boolean {
      return this.settingStore.isGPUOverProvisioningEnabled;
    },
    oprClass(): string {
      return this.overProvisionEnabled ? "" : "q-mr-sm";
    },
  },
  methods: {
    updateGpuRequest(request: number | string | null): void {
      if (request === null || request === "") request = 0;
      const keyToUpdate: string = this.isGpuMemoryRequest ? "gpuMemoryRequest" : "gpuPortionRequest";
      const valueToUpdate: string | number = this.isGpuMemoryRequest
        ? request + this.selectedRequestType
        : +request / GPU_PORTION_FACTOR;

      // If the limit is enabled and the limit is lower than the request, update the limit to be equal to the request
      if (this.isLimitEnabled) {
        let currentLimit;

        if (this.isGpuMemoryRequest && this.gpuRequest.gpuMemoryLimit) {
          currentLimit = parseFloat(this.gpuRequest.gpuMemoryLimit);
        } else if (this.gpuRequest.gpuPortionLimit || this.gpuRequest.gpuPortionLimit === 0) {
          currentLimit = this.gpuRequest.gpuPortionLimit;
        }

        if ((currentLimit || currentLimit === 0) && currentLimit < parseFloat(valueToUpdate.toString())) {
          const limitKeyToUpdate = this.isGpuMemoryRequest ? "gpuMemoryLimit" : "gpuPortionLimit";
          this.$emit("update-gpu-request", {
            ...this.gpuRequest,
            [keyToUpdate]: valueToUpdate,
            [limitKeyToUpdate]: valueToUpdate,
          });
          return;
        } else {
          this.$emit("update-gpu-request", {
            ...this.gpuRequest,
            [keyToUpdate]: valueToUpdate,
          });
          return;
        }
      }

      this.$emit("update-gpu-request", {
        ...this.gpuRequest,
        [keyToUpdate]: valueToUpdate,
      });
    },
    updateGpuLimit(limit: number | string | null): void {
      if (limit === null || limit === "") limit = 0;
      const keyToUpdate: string = this.isGpuMemoryRequest ? "gpuMemoryLimit" : "gpuPortionLimit";

      const valueToUpdate: string | number = this.isGpuMemoryRequest
        ? limit + this.selectedRequestType
        : +limit / GPU_PORTION_FACTOR;

      this.$emit("update-gpu-request", {
        ...this.gpuRequest,
        [keyToUpdate]: valueToUpdate,
      });
    },
    updateLimitEnabled(value: boolean): void {
      const keyToUpdate = this.isGpuMemoryRequest ? "gpuMemoryLimit" : "gpuPortionLimit";
      let valueToUpdate = this.isGpuMemoryRequest ? this.gpuRequest.gpuMemoryRequest : this.gpuRequest.gpuPortionRequest;
      if (!value) {
        valueToUpdate = undefined;
      }

      this.$emit("update-gpu-request", {
        ...this.gpuRequest,
        [keyToUpdate]: valueToUpdate,
      });
    },
    updateGpuRequestType(requestType: ISelectOption): void {
      const fieldsDefaults = {
        gpuDevicesRequest: this.gpuRequest.gpuDevicesRequest,
        gpuPortionRequest: undefined,
        gpuPortionLimit: undefined,
        gpuMemoryLimit: undefined,
        gpuMemoryRequest: undefined,
      };

      const updatedResources: IGpuRequestModel = {
        ...fieldsDefaults,
      };

      if (requestType.value === EMemoryUnitValue.MB || requestType.value === EMemoryUnitValue.GB) {
        updatedResources.gpuRequestType = GpuRequestType.Memory;
        const parsed = parseSizeString(this.gpuRequest.gpuMemoryRequest || this.policyDefaults?.gpuMemoryRequest || "");
        let amount = parsed?.[0]?.toString();
        if (!amount) {
          if (requestType.value === EMemoryUnitValue.MB) amount = "1000";
          else amount = "1";
        }
        updatedResources.gpuMemoryRequest = amount + requestType.value;
      } else if (requestType.value === GpuRequestType.Portion) {
        updatedResources.gpuRequestType = GpuRequestType.Portion;
        updatedResources.gpuPortionRequest = this.policyDefaults?.gpuPortionRequest || 1;
      }

      this.$emit("update-gpu-request", updatedResources);
    },
    // VALIDATION RULES
    isLimitAboveZero(value: number): boolean | string {
      return value > 0 || errorMessages.ENTER_A_LIMIT_ABOVE_ZERO;
    },
    isRequestAboveZero(value: number): boolean | string {
      return value > 0 || errorMessages.VALUE_ABOVE_ZERO;
    },
    isRequestEqualOrAboveOne(value: number): boolean | string {
      return value >= 0 || errorMessages.VALUE_EQUAL_ABOVE_ONE;
    },
    isValidLimit(value: number): boolean | string {
      const requestToCheck = this.isGpuMemoryRequest
        ? this.gpuRequest.gpuMemoryRequest
        : this.gpuRequest.gpuPortionRequest;

      if (!this.isGpuMemoryRequest) value /= GPU_PORTION_FACTOR;

      return (
        value >= ((requestToCheck && parseFloat(requestToCheck.toString())) || 0) ||
        errorMessages.LIMIT_EQUAL_OR_HIGHER_THAN_REQUEST
      );
    },
    convertGpuPortionRules(portionRules: NumberRules): NumberRules | undefined {
      if (!portionRules) return undefined;
      let convertedMax;
      let convertedMin;
      let convertedStep;

      if (portionRules.max) {
        convertedMax = +(portionRules.max * GPU_PORTION_FACTOR).toFixed(0);
      }
      if (portionRules.min) {
        convertedMin = +(portionRules.min * GPU_PORTION_FACTOR).toFixed(0);
      }
      if (portionRules.step) {
        convertedStep = +(portionRules.step * GPU_PORTION_FACTOR).toFixed(0);
      }
      return {
        ...portionRules,
        max: convertedMax,
        min: convertedMin,
        step: convertedStep,
      };
    },
  },
});
</script>

<style lang="scss">
.gpu-resource-box {
  .box-title {
    font-weight: 700;
  }

  .mig-profile-select {
    margin-bottom: 20px;
    width: 135px;
    background-color: $white;
  }

  .spacer {
    margin-inline: 8px;
    width: 190px;
  }

  .input-field {
    width: 135px;
  }

  .unit-select {
    & > div {
      width: 136px;
      background-color: $grey-3;
    }
  }

  .dash-icon {
    color: $black-54;
  }

  .dash-icon,
  .limit-toggle,
  .box-title {
    padding-bottom: 20px;
  }
}
</style>
