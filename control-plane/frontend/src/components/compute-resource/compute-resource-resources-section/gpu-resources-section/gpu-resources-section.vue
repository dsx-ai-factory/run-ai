<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="gpu-resources-section">
    <gpu-devices-section
      :model-value="modelValue"
      :policy-defaults="policyDefaults"
      :policy-rules="policyRules"
      :section-disabled="sectionDisabled"
      :support-multi-fraction-gpu="supportMultiFractionGpu"
      :support-gpu-fractioning="supportGpuFractioning"
      @update:model-value="updateGpuResources"
      :support-gpu-limit="supportGpuLimit"
    />

    <gpu-memory-section
      v-if="isFractioning"
      class="q-mb-md"
      :model-value="modelValue"
      :policy-defaults="policyDefaults"
      :policy-rules="policyRules"
      :section-disabled="sectionDisabled"
      @update:model-value="updateGpuResources"
      :support-gpu-limit="supportGpuLimit"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { GpuDevicesSection } from "./gpu-devices-section";
import { GpuMemorySection } from "./gpu-memory-section";
// Models
import type { SupersetDefaultsAllOfCompute, SupersetRulesAllOfCompute } from "@/swagger-models/workloads-client";
import type { IGpuRequestModel } from "@/models/compute-resource.model";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "gpu-resources-section",
  components: {
    GpuDevicesSection,
    GpuMemorySection,
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
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<SupersetDefaultsAllOfCompute | null>,
      required: false,
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
    supportGpuLimit: {
      type: Boolean as PropType<boolean>,
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
  },
  methods: {
    updateGpuResources(gpuResources: IGpuRequestModel): void {
      this.$emit(EEmitNames.UpdateModelValue, gpuResources);
    },
  },
});
</script>
