<template>
  <div class="gpu-resources-summary bg-grey-3 q-pa-md">
    <div class="gpu-resources-title q-mb-md">Total requested GPU resources</div>
    <div class="row items-center q-gutter-x-sm">
      <span class="text-h6">{{ totalGpuResources.totalGpus }}</span>
      <runai-tooltip :tooltip-text="totalGpuResources.tooltipText" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IGpuRequestModel } from "@/models/compute-resource.model";
import { computeResourceUtil, type IGpuResourcesSummary } from "@/utils/compute-resource.util/compute-resource.util";

import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  name: "gpu-resources-summary",
  components: {
    RunaiTooltip,
  },
  emits: {},
  props: {
    gpuResources: {
      type: Object as PropType<IGpuRequestModel>,
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
  },
  computed: {
    totalGpuResources(): IGpuResourcesSummary {
      return computeResourceUtil.totalGpuResources(this.gpuResources, this.workers, this.replicas);
    },
  },
});
</script>

<style lang="scss" scoped>
.gpu-resources-summary {
  .gpu-resources-title {
    font-size: 13px;
    font-weight: 600;
    color: $section-title-color;
  }
}
</style>
