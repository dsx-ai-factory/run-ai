<template>
  <section class="gpu-devices-section row items-center q-gutter-x-sm">
    <policy-number-input
      aid="gpu-devices-input"
      class="col-3 q-mr-lg"
      label="GPU devices"
      :min-value="0"
      :model-value="modelValue.gpuDevicesRequest"
      @update:model-value="updateGpuDevicesRequest"
      :policy-rules="policyRules?.gpuDevicesRequest"
      :disable="sectionDisabled"
      is-integer
    />
    <runai-toggle
      :model-value="isFractioning"
      @update:model-value="updateIsFractioning"
      label="GPU fractioning"
      :left-label="false"
      :disable="isFractioningDisabled"
      :tooltip="fractioningDisabledTooltip"
      aid="gpu-fraction-toggle"
    />
    <runai-tooltip tooltip-text="Enable GPU fractioning to allocate a portion of the GPU memory to each device" />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiToggle } from "@/components/common/runai-toggle";
// Models
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import type { SupersetDefaultsAllOfCompute, SupersetRulesAllOfCompute } from "@/swagger-models/workloads-client";
import { GpuRequestType } from "@/swagger-models/assets-service-client";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "gpu-devices-section",
  components: {
    PolicyNumberInput,
    RunaiTooltip,
    RunaiToggle,
  },
  emits: {
    [EEmitNames.UpdateModelValue]: (_: IGpuRequestModel) => {
      return true;
    },
  },
  props: {
    modelValue: {
      type: Object as PropType<IGpuRequestModel>,
      required: true,
    },
    supportMultiFractionGpu: {
      type: Boolean as PropType<boolean>,
      default: false,
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
      required: false,
      default: false,
    },
    supportGpuFractioning: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    isFractioning(): boolean {
      return !!this.modelValue.gpuRequestType;
    },
    isFractioningDisabled(): boolean {
      return (
        this.sectionDisabled ||
        !this.supportGpuFractioning ||
        this.modelValue.gpuDevicesRequest < 1 ||
        (this.modelValue.gpuDevicesRequest > 1 && !this.supportMultiFractionGpu)
      );
    },
    fractioningDisabledTooltip(): string {
      if (this.sectionDisabled || !this.isFractioningDisabled) {
        return "";
      }
      if (!this.supportGpuFractioning) {
        return "GPU fractioning is not supported for the selected cluster";
      }
      if (this.modelValue.gpuDevicesRequest > 1 && !this.supportMultiFractionGpu) {
        return "Cluster version does not support GPU fractioning for multiple devices, to enable this set 1 GPU device";
      }
      return "To use GPU fractioning, set at least 1 GPU device";
    },
  },
  methods: {
    updateIsFractioning(value: boolean): void {
      const defaultPortionLimit = this.supportGpuLimit ? this.policyDefaults?.gpuPortionLimit || 0.5 : undefined;

      this.$emit(EEmitNames.UpdateModelValue, {
        ...this.modelValue,
        gpuRequestType: value ? GpuRequestType.Portion : undefined,
        gpuPortionRequest: value ? this.policyDefaults?.gpuPortionRequest || 0.5 : undefined,
        gpuPortionLimit: value ? defaultPortionLimit : undefined,
        gpuMemoryRequest: undefined,
        gpuMemoryLimit: undefined,
      });
    },
    updateGpuDevicesRequest(gpuDevicesRequest: number): void {
      if (!+gpuDevicesRequest) {
        // If gpuDevicesRequest is 0 we reset all the gpu fields
        this.$emit(EEmitNames.UpdateModelValue, {
          gpuDevicesRequest: +gpuDevicesRequest,
          gpuMemoryRequest: undefined,
          gpuMemoryLimit: undefined,
          gpuPortionRequest: undefined,
          gpuPortionLimit: undefined,
          gpuRequestType: undefined,
        });
      } else {
        this.$emit(EEmitNames.UpdateModelValue, {
          ...this.modelValue,
          gpuDevicesRequest: +gpuDevicesRequest,
        });
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.gpu-devices-section {
  margin-bottom: 30px;
}
</style>
