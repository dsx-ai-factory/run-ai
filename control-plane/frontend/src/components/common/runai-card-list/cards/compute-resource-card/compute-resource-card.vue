<template>
  <policy-non-compliance-card-wrapper
    :compliance-data="data.compliance || undefined"
    :field="data.meta.kind"
    :show-disabled-info="showDisabledInfo"
    :asset-name="data.meta.name"
  >
    <runai-select-card
      :value="isSelected"
      :disabled="disabled"
      @clicked="$emit('card-clicked')"
      :aid="getAid(data.meta.name)"
      class="compute-resource-card"
      :class="{ 'card-disabled': disabled }"
    >
      <q-tooltip v-if="data.meta.description" max-width="300px" anchor="top middle" self="bottom middle">{{
        data.meta.description
      }}</q-tooltip>
      <div class="workload-card-content">
        <section class="top-card-section">
          <div class="compute-resource-logo">
            <q-icon name="ph-bold ph-hard-drives" size="42px" color="grey-8" />
          </div>
        </section>
        <div class="compute-resource-card-title q-mb-md" :class="{ 'text-primary': isSelected }">
          <div class="col text-subtitle1 text-weight-medium ellipsis">
            {{ data.meta.name }}
          </div>
        </div>
        <div class="compute-resource-card-details">
          <div class="row no-wrap">
            <span class="text-weight-medium text-caption q-mr-xs info-title">GPU devices:</span>
            <span class="ellipsis">
              {{ data.spec.gpuDevicesRequest || 0 }}
            </span>
          </div>
          <div class="row no-wrap" v-if="showGpuDetailsSection">
            <span class="text-weight-medium text-caption q-mr-xs info-title">GPU {{ gpuDetailsSectionTitle }}:</span>
            <span class="ellipsis">
              {{ gpuDetails }}
            </span>
          </div>
          <div class="row no-wrap" v-if="showCpuComputeSection">
            <span class="text-weight-medium text-caption q-mr-xs info-title">CPU compute (Cores):</span>
            <span class="ellipsis">
              {{ cpuCoresDetails }}
            </span>
          </div>
          <div class="row no-wrap" v-if="showCpuMemorySection">
            <span class="text-weight-medium text-caption q-mr-xs info-title">CPU Memory ({{ memoryType }}):</span>
            <span class="ellipsis">
              {{ cpuMemoryDetails }}
            </span>
          </div>
        </div>
      </div>
    </runai-select-card>
  </policy-non-compliance-card-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { RunaiSelectCard } from "@/components/common/runai-select-card";

// models
import { GpuRequestType, type ComputeAsset } from "@/swagger-models/assets-service-client";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import { policyNonComplianceCardWrapper } from "../../policy-non-compliance-card-wrapper";

export default defineComponent({
  components: {
    RunaiSelectCard,
    policyNonComplianceCardWrapper,
  },
  emits: ["card-clicked"],
  inject: ["isGPUOverProvisioningEnabled"],
  props: {
    data: {
      type: Object as PropType<ComputeAsset>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    showDisabledInfo: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    memoryType(): string {
      const memoryRequest = this.data.spec.cpuMemoryRequest || "";
      return this.getMemoryUnitAndValue(memoryRequest).unit;
    },
    showGpuDetailsSection(): boolean {
      const { gpuMemoryRequest, gpuPortionRequest } = this.data.spec;
      return !!(gpuPortionRequest || gpuMemoryRequest);
    },
    gpuDetailsSectionTitle(): string {
      const { gpuMemoryRequest, gpuRequestType } = this.data.spec;
      switch (gpuRequestType) {
        case GpuRequestType.Memory:
          return `memory (${this.getMemoryUnitAndValue(String(gpuMemoryRequest)).unit})`;
        case GpuRequestType.Portion:
          return "% (of device)";
        default:
          return "";
      }
    },
    showCpuComputeSection(): boolean {
      return !!this.data.spec.cpuCoreRequest;
    },
    showCpuMemorySection(): boolean {
      return !!this.data.spec.cpuMemoryRequest;
    },
    gpuDetails(): string {
      const { gpuRequestType, gpuMemoryLimit, gpuMemoryRequest, gpuPortionRequest, gpuPortionLimit } = this.data.spec;

      const showLimit = !!this.isGPUOverProvisioningEnabled;

      switch (gpuRequestType) {
        case GpuRequestType.Memory:
          return showLimit && gpuMemoryLimit
            ? `${this.getMemoryUnitAndValue(String(gpuMemoryRequest)).val}-${
                this.getMemoryUnitAndValue(String(gpuMemoryLimit)).val
              }`
            : this.getMemoryUnitAndValue(String(gpuMemoryRequest)).val;
        case GpuRequestType.Portion:
          return showLimit && gpuPortionLimit
            ? `${(gpuPortionRequest || 0) * 100}-${(gpuPortionLimit || 0) * 100}`
            : String((gpuPortionRequest || 0) * 100);
        default:
          return "";
      }
    },
    cpuCoresDetails(): string {
      return this.data.spec.cpuCoreLimit
        ? `${this.data.spec.cpuCoreRequest}-${this.data.spec.cpuCoreLimit}`
        : String(this.data.spec.cpuCoreRequest);
    },
    cpuMemoryDetails(): string {
      const { cpuMemoryRequest, cpuMemoryLimit } = this.data.spec;
      return cpuMemoryLimit
        ? `${this.getMemoryUnitAndValue(String(cpuMemoryRequest)).val}-${
            this.getMemoryUnitAndValue(String(cpuMemoryLimit)).val
          }`
        : this.getMemoryUnitAndValue(String(cpuMemoryRequest)).val;
    },
  },
  methods: {
    getMemoryUnitAndValue(request: string): { unit: string; val: string } {
      const unit = request?.charAt(request.length - 1);
      const val = request.slice(0, request.length - 1);
      switch (unit) {
        case "M":
          return { unit: "MB", val: val };
        case "G":
          return { unit: "GB", val: val };
        default:
          return { unit: "", val: "" };
      }
    },
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
});
</script>

<style lang="scss" scoped>
.compute-resource-card {
  .workload-card-content {
    .top-card-section {
      display: flex;
      justify-content: space-between;
    }

    .info-title {
      white-space: nowrap;
    }
  }
}
</style>
