<template>
  <runai-section class="gpu-memory-section" title="Set the fraction for each device">
    <section class="row q-gutter-x-md">
      <runai-select
        class="col-3"
        aid="gpu-type-select"
        label="GPU memory"
        :model-value="selectedRequestType"
        :options="requestTypeOptions"
        :policy-rules="requestTypeRules"
        :policy-defaults="policyDefaults?.gpuRequestType"
        :disable="sectionDisabled"
        unclearable
        @update:model-value="updateGpuRequestType"
      />

      <template v-if="modelValue.gpuRequestType === 'portion'">
        <policy-number-input
          class="col-3"
          aid="gpu-request-input"
          label="Request"
          :min-value="0"
          :input-step="10"
          :max-value="99"
          :rules="[isValidPortionRequest]"
          :model-value="gpuPortionRequest"
          :policy-defaults="policyDefaults?.gpuPortionRequest"
          :policy-rules="gpuPortionRules"
          :disable="sectionDisabled"
          @update:model-value="updateGpuPortionRequest"
        />

        <policy-number-input
          v-if="supportGpuLimit"
          class="col-3"
          aid="gpu-limit-input"
          stack-label
          label="Limit"
          :model-value="gpuPortionLimit"
          :min-value="0"
          :input-step="10"
          :max-value="100"
          :disable="sectionDisabled"
          @update:model-value="updateGpuPortionLimit"
          reactive-rules
          :rules="[isValidPortionLimit]"
          :policy-rules="gpuPortionLimitRules"
        />
      </template>

      <template v-else-if="modelValue.gpuRequestType === 'memory'">
        <policy-quantity-input
          aid="gpu-request-input"
          label="Request"
          class="col-3"
          :model-value="gpuMemoryRequest"
          :unit-size="gpuMemoryRequestUnit"
          :min-value="0"
          :max-value="gpuMemoryMaxValue"
          :step="gpuMemoryStep"
          :rules="[isValidMemoryRequest]"
          @update:model-value="updateGpuMemoryRequest"
          :disable="sectionDisabled"
          :policy-rules="policyRules?.gpuMemoryRequest"
        />

        <policy-quantity-input
          v-if="supportGpuLimit"
          aid="gpu-limit-input"
          label="Limit"
          stack-label
          class="col-3"
          :model-value="gpuMemoryLimit"
          :unit-size="gpuMemoryRequestUnit"
          :min-value="0"
          :max-value="gpuMemoryMaxValue"
          :step="gpuMemoryStep"
          @update:model-value="updateGpuMemoryLimit"
          :disable="sectionDisabled"
          :rules="[isValidMemoryLimit]"
          reactive-rules
          :policy-rules="policyRules?.gpuMemoryLimit"
        />
      </template>
    </section>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { PolicyQuantityInput } from "@/components/common/policy-quantity-input";
// Models
import { gpuRequestTypeOptions, type IGpuRequestModel } from "@/models/compute-resource.model";
import type {
  SupersetDefaultsAllOfCompute,
  SupersetRulesAllOfCompute,
  GpuRequestOptionsOptionsInner,
} from "@/swagger-models/workloads-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import { EMemoryUnitValue, type ISelectOption } from "@/models/global.model";
import { GpuRequestType } from "@/swagger-models/assets-service-client";
import { parseSizeString } from "@/utils/format.util";
import type { NumberRules } from "@/swagger-models/policy-service-client";
import { errorMessagesUtils } from "@/common/error-message.constant";

const GPU_PORTION_FACTOR = 100;
const PortionRequestMax = 99;

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "gpu-memory-section",
  components: {
    RunaiSection,
    RunaiSelect,
    PolicyNumberInput,
    PolicyQuantityInput,
  },
  emits: {
    "update:model-value": (_: IGpuRequestModel) => {
      return true;
    },
  },
  props: {
    modelValue: {
      type: Object as PropType<IGpuRequestModel>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<SupersetDefaultsAllOfCompute | null>,
      required: false,
    },
    supportGpuLimit: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    selectedRequestType(): string {
      if (this.modelValue.gpuRequestType === GpuRequestType.Memory) {
        if (this.modelValue.gpuMemoryRequest?.endsWith(EMemoryUnitValue.MB)) {
          return EMemoryUnitValue.MB;
        } else {
          return EMemoryUnitValue.GB;
        }
      }
      return this.modelValue.gpuRequestType || GpuRequestType.Portion;
    },
    requestTypeOptions(): ISelectOption[] {
      return gpuRequestTypeOptions;
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
    gpuPortionRules(): NumberRules | undefined {
      if (!this.policyRules?.gpuPortionRequest) return undefined;
      return this.convertGpuPortionRules(this.policyRules.gpuPortionRequest);
    },
    gpuPortionLimitRules(): NumberRules | undefined {
      if (!this.policyRules?.gpuPortionLimit) return undefined;
      return this.convertGpuPortionRules(this.policyRules.gpuPortionLimit);
    },
    gpuPortionRequest(): number {
      if (!this.modelValue.gpuPortionRequest) return 0;
      return +(this.modelValue.gpuPortionRequest * GPU_PORTION_FACTOR).toFixed(0);
    },
    gpuPortionLimit(): number | null {
      if (!this.modelValue.gpuPortionLimit) return null;
      return +(this.modelValue.gpuPortionLimit * GPU_PORTION_FACTOR).toFixed(0);
    },
    gpuMemoryRequest(): number | null | undefined {
      if (!this.modelValue.gpuMemoryRequest) return null;
      return this.modelValue.gpuMemoryRequest ? parseFloat(this.modelValue.gpuMemoryRequest) : null;
    },
    gpuMemoryLimit(): number | null | undefined {
      if (!this.modelValue.gpuMemoryLimit) return null;
      return this.modelValue.gpuMemoryLimit
        ? parseFloat(this.modelValue.gpuMemoryLimit || "0")
        : parseFloat(this.modelValue.gpuMemoryRequest || "0");
    },
    gpuMemoryRequestUnit(): EMemoryUnitValue {
      const parsed = parseSizeString(this.modelValue.gpuMemoryRequest || "");
      return (parsed?.[1] as EMemoryUnitValue) || EMemoryUnitValue.MB;
    },
    gpuMemoryMaxValue(): number {
      return this.selectedRequestType === EMemoryUnitValue.MB ? 1000000 : 1000;
    },
    gpuMemoryStep(): number {
      return this.selectedRequestType === EMemoryUnitValue.MB ? 100 : 1;
    },
  },
  methods: {
    updateGpuRequestType(requestType: ISelectOption): void {
      const fieldsDefaults = {
        gpuDevicesRequest: this.modelValue.gpuDevicesRequest,
        gpuPortionRequest: undefined,
        gpuPortionLimit: undefined,
        gpuMemoryLimit: undefined,
        gpuMemoryRequest: undefined,
      };

      const updatedResources: IGpuRequestModel = {
        ...fieldsDefaults,
      };

      if (requestType.value === EMemoryUnitValue.MB) {
        updatedResources.gpuRequestType = GpuRequestType.Memory;
        updatedResources.gpuMemoryRequest = "500M";
        if (this.supportGpuLimit) {
          updatedResources.gpuMemoryLimit = "500M";
        }
      } else if (requestType.value === EMemoryUnitValue.GB) {
        updatedResources.gpuRequestType = GpuRequestType.Memory;
        updatedResources.gpuMemoryRequest = "1G";
        if (this.supportGpuLimit) {
          updatedResources.gpuMemoryLimit = "1G";
        }
      } else if (requestType.value === GpuRequestType.Portion) {
        updatedResources.gpuRequestType = GpuRequestType.Portion;
        updatedResources.gpuPortionRequest = this.policyDefaults?.gpuPortionRequest || 0.5;
        if (this.supportGpuLimit) {
          updatedResources.gpuPortionLimit = this.policyDefaults?.gpuPortionLimit || 0.5;
        }
      }

      this.$emit("update:model-value", updatedResources);
    },
    // Portion
    updateGpuPortionRequest(gpuPortionRequest: string): void {
      const request: number = +gpuPortionRequest / GPU_PORTION_FACTOR;
      const portionLimit = this.modelValue.gpuPortionLimit;
      if (portionLimit && request > portionLimit) {
        this.$emit("update:model-value", {
          ...this.modelValue,
          gpuPortionRequest: request,
          gpuPortionLimit: this.supportGpuLimit ? request : undefined,
        });
      } else {
        this.$emit("update:model-value", {
          ...this.modelValue,
          gpuPortionRequest: request,
        });
      }
    },
    updateGpuPortionLimit(gpuPortionLimit: string): void {
      const limit: number | null = gpuPortionLimit === "" ? null : +gpuPortionLimit / GPU_PORTION_FACTOR;
      this.$emit("update:model-value", {
        ...this.modelValue,
        gpuPortionLimit: limit,
      });
    },
    // Memory
    updateGpuMemoryRequest(gpuMemoryRequest: string): void {
      let request: string;
      if (this.gpuMemoryRequestUnit === EMemoryUnitValue.GB) {
        request = (+gpuMemoryRequest).toFixed(2);
      } else {
        request = (+gpuMemoryRequest).toFixed(0);
      }
      request += this.selectedRequestType;

      const memoryLimit = this.modelValue.gpuMemoryLimit;
      if (memoryLimit && parseFloat(request) > parseFloat(memoryLimit)) {
        this.$emit("update:model-value", {
          ...this.modelValue,
          gpuMemoryRequest: request,
          gpuMemoryLimit: this.supportGpuLimit ? request : undefined,
        });
      } else {
        this.$emit("update:model-value", {
          ...this.modelValue,
          gpuMemoryRequest: request,
        });
      }
    },
    updateGpuMemoryLimit(gpuMemoryLimit: string): void {
      let limit: string | null = null;
      if (this.gpuMemoryRequestUnit === EMemoryUnitValue.GB) {
        limit = (+gpuMemoryLimit).toFixed(2);
      } else {
        limit = (+gpuMemoryLimit).toFixed(0);
      }

      if (gpuMemoryLimit === "") {
        limit = null;
      } else {
        limit += this.selectedRequestType;
      }

      this.$emit("update:model-value", {
        ...this.modelValue,
        gpuMemoryLimit: limit,
      });
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
    // Rules
    isValidMemoryRequest(value: number): boolean | string {
      if (value <= 0 || value > this.gpuMemoryMaxValue) {
        return errorMessagesUtils.VALUE_OVER_UP_TO(0, this.gpuMemoryMaxValue);
      }
      return true;
    },
    isValidPortionRequest(value: number): boolean | string {
      if (value <= 0 || value >= 100) {
        return errorMessagesUtils.VALUE_OVER_UP_TO(0, PortionRequestMax);
      }
      return true;
    },
    isValidPortionLimit(value: number | null): boolean | string {
      if (value === null) {
        return true;
      }
      if (value < this.gpuPortionRequest || value > 100) {
        return errorMessagesUtils.LIMIT_FROM_REQUEST_TO_MAX(100);
      }
      return true;
    },
    isValidMemoryLimit(value: number): boolean | string {
      if (this.gpuMemoryRequest === null || this.gpuMemoryRequest === undefined || value === null) {
        return true;
      }

      if (value < this.gpuMemoryRequest || value > this.gpuMemoryMaxValue) {
        return errorMessagesUtils.LIMIT_FROM_REQUEST_TO_MAX(this.gpuMemoryMaxValue);
      }
      return true;
    },
  },
});
</script>
