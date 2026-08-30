<template>
  <runai-expansion-item
    class="compute-resource-resources-section"
    label="Resources"
    default-opened
    should-validate
    :section-invalid="hasFoldingSubSectionErrors"
  >
    <template #subheader>
      <div class="summary">{{ summary }}</div>
    </template>

    <compute-resource-section-layout
      :model-value="resourcesModel"
      :support-multi-fraction-gpu="supportMultiFractionGpu"
      :support-gpu-limit="supportGpuLimit"
      @update:model-value="updateResources"
      @validation-state-changed="cpuResourcesHasError = $event"
      section-title="Set the resources"
    />

    <runai-folding-sub-section
      label="More settings"
      aid="more-settings"
      @validation-state-changed="moreSettingsHasError = $event"
    >
      <large-shm-toggle :model-value="largeShmRequest" @update:model-value="$emit('update-shm-request', $event)" />

      <extended-resources
        :extended-resources="extendedResources"
        @update:model-value="$emit('update-extended-resources', $event)"
        aid="extended-resources"
      />
    </runai-folding-sub-section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Models
import type { IResourcesSectionModel } from "@/models/compute-resource.model";
import type { ExtendedResource } from "@/swagger-models/assets-service-client";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { ExtendedResources } from "./extended-resources";
import { LargeShmToggle } from "../large-shm-toggle";
import { computeResourceUtil } from "@/utils/compute-resource.util";
import { ComputeResourceSectionLayout } from "./compute-resource-section-layout";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    ExtendedResources,
    LargeShmToggle,
    ComputeResourceSectionLayout,
    RunaiFoldingSubSection,
  },
  emits: ["update-resources-model", "update-shm-request", "update-extended-resources", "is-section-invalid"],
  props: {
    resourcesModel: {
      type: Object as PropType<IResourcesSectionModel>,
      required: true,
    },
    largeShmRequest: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    extendedResources: {
      type: Array as PropType<ExtendedResource[]>,
      required: true,
    },
    supportMultiFractionGpu: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportGpuLimit: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isCpuCoreManuallySet: false,
      isCpuMemoryManuallySet: false,
      moreSettingsHasError: false,
      cpuResourcesHasError: false,
    };
  },
  computed: {
    hasFoldingSubSectionErrors(): boolean {
      return this.moreSettingsHasError || this.cpuResourcesHasError;
    },
    summary(): string {
      return `Total Requested GPU Resources: ${computeResourceUtil.totalGpuResources(this.resourcesModel).totalGpus}`;
    },
    shmTooltip(): string {
      return "When enabled, the shared memory size available to the workload will be increased from the default 64MB to the node's total available memory or the CPU memory limit, if set above";
    },
  },
  methods: {
    updateResources(resources: IResourcesSectionModel): void {
      this.$emit("update-resources-model", resources);
    },
  },
  watch: {
    hasFoldingSubSectionErrors: {
      handler() {
        this.$emit("is-section-invalid", this.hasFoldingSubSectionErrors);
      },
      immediate: true,
    },
  },
});
</script>
