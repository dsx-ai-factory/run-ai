<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-folding-sub-section
    class="cpu-resources-section"
    aid="cpu-resources-sub-section"
    label="CPU resources"
    @validation-state-changed="$emit('validation-state-changed', $event)"
  >
    <cpu-compute-section
      class="q-mb-lg"
      :model-value="modelValue.cpuCore"
      @update:model-value="updateCpuCore"
      :policy-rules="policyRules"
      :section-disabled="sectionDisabled"
    />
    <cpu-memory-section
      :model-value="modelValue.cpuMemory"
      @update:model-value="updateCpuMemory"
      :policy-rules="policyRules"
      :section-disabled="sectionDisabled"
    />
  </runai-folding-sub-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { CpuComputeSection } from "./cpu-compute-section";
import { CpuMemorySection } from "./cpu-memory-section";
// Models
import type { ICpuRequestModel } from "@/models/compute-resource.model";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/policy-service-client";

export enum EEmitNames {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "cpu-resources-section",
  components: { RunaiFoldingSubSection, CpuComputeSection, CpuMemorySection },
  emits: {
    [EEmitNames.UpdateModelValue]: (_: ICpuRequestModel) => {
      return true;
    },
    "validation-state-changed": (_: boolean) => true,
  },
  props: {
    modelValue: {
      type: Object as PropType<ICpuRequestModel>,
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
  methods: {
    updateCpuCore(cpuCore: ICpuRequestModel["cpuCore"]) {
      this.$emit(EEmitNames.UpdateModelValue, { ...this.modelValue, cpuCore });
    },
    updateCpuMemory(cpuMemory: ICpuRequestModel["cpuMemory"]) {
      this.$emit(EEmitNames.UpdateModelValue, { ...this.modelValue, cpuMemory });
    },
  },
});
</script>
