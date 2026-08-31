<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="compute-resource-section-layout q-pb-sm q-mb-sm">
    <div class="row no-wrap">
      <div class="left-panel">
        <slot name="top-left" />
        <runai-section :title="sectionTitle">
          <gpu-resources-section
            v-if="!hideGpuResources"
            :model-value="gpuResourcesModel"
            @update:model-value="updateGpuResourcesModel"
            :section-disabled="sectionDisabled"
            :support-multi-fraction-gpu="supportMultiFractionGpu"
            :support-gpu-fractioning="supportGpuFractioning"
            :support-gpu-limit="supportGpuLimit"
            :policy-rules="policyRules"
            :policy-defaults="policyDefaults"
          />
          <slot name="below-gpu-resources-section" />
          <cpu-resources-section
            :model-value="cpuResourcesModel"
            @update:model-value="updateCpuResourcesModel"
            @validation-state-changed="$emit('validation-state-changed', $event)"
            :section-disabled="sectionDisabled"
            :policy-rules="policyRules"
            :policy-defaults="policyDefaults"
          />
        </runai-section>
      </div>
      <div class="right-panel">
        <gpu-resources-summary
          v-if="!hideGpuResources"
          :gpu-resources="gpuResourcesModel"
          :workers="workers"
          :replicas="replicas"
        />
        <preempted-warning-section :resources="modelValue" />
      </div>
    </div>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Components
import { GpuResourcesSummary } from "./gpu-resources-summary";
import { PreemptedWarningSection } from "./preempted-warning-section";
import { CpuResourcesSection } from "../cpu-resources-section";
import { GpuResourcesSection } from "../gpu-resources-section";
import { RunaiSection } from "@/components/common/runai-section";
// Models
import type { ICpuRequestModel, IGpuRequestModel, IResourcesSectionModel } from "@/models/compute-resource.model";
import type { SupersetRulesAllOfCompute } from "@/swagger-models/policy-service-client";
import type { SupersetDefaultsAllOfCompute } from "@/swagger-models/policy-service-client";
export default defineComponent({
  name: "compute-resource-section-layout",
  components: {
    GpuResourcesSummary,
    PreemptedWarningSection,
    CpuResourcesSection,
    GpuResourcesSection,
    RunaiSection,
  },
  emits: {
    "update:model-value": (_: IResourcesSectionModel) => true,
    "validation-state-changed": (_: boolean) => true,
  },
  props: {
    modelValue: {
      type: Object as PropType<IResourcesSectionModel>,
      required: true,
    },
    workers: {
      type: Number as PropType<number>,
      required: false,
    },
    replicas: {
      type: Object as PropType<{ min: number | null | undefined; max: number | null | undefined }>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetRulesAllOfCompute | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<SupersetDefaultsAllOfCompute | null>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    supportMultiFractionGpu: {
      type: Boolean as PropType<boolean>,
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
    hideGpuResources: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    sectionTitle: {
      type: String as PropType<string>,
      required: false,
      default: "Set the resources",
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    gpuResourcesModel(): IGpuRequestModel {
      return {
        gpuDevicesRequest: this.modelValue.gpuDevicesRequest,
        gpuRequestType: this.modelValue.gpuRequestType,
        gpuMemoryRequest: this.modelValue.gpuMemoryRequest,
        gpuMemoryLimit: this.modelValue.gpuMemoryLimit,
        gpuPortionRequest: this.modelValue.gpuPortionRequest,
        gpuPortionLimit: this.modelValue.gpuPortionLimit,
      };
    },
    cpuResourcesModel(): ICpuRequestModel {
      return {
        cpuCore: this.modelValue.cpuCore,
        cpuMemory: this.modelValue.cpuMemory,
      };
    },
  },
  methods: {
    updateGpuResourcesModel(model: IGpuRequestModel) {
      this.$emit("update:model-value", {
        ...this.modelValue,
        gpuDevicesRequest: model.gpuDevicesRequest,
        gpuRequestType: model.gpuRequestType,
        gpuMemoryRequest: model.gpuMemoryRequest,
        gpuMemoryLimit: model.gpuMemoryLimit,
        gpuPortionRequest: model.gpuPortionRequest,
        gpuPortionLimit: model.gpuPortionLimit,
      });
    },
    updateCpuResourcesModel(model: ICpuRequestModel) {
      this.$emit("update:model-value", {
        ...this.modelValue,
        cpuCore: model.cpuCore,
        cpuMemory: model.cpuMemory,
      });
    },
  },
});
</script>

<style lang="scss" scoped>
.compute-resource-section-layout {
  border-bottom: 1px solid #e0e0e0;

  .left-panel {
    flex-basis: calc(100% - $distributed-gpu-summary-width);
  }

  .right-panel {
    flex-basis: $distributed-gpu-summary-width;
  }
}
</style>
