<template>
  <section>
    <transition name="fade">
      <div class="preempted-warning-section q-pa-md q-mt-md" v-if="shouldShowCpuWarning || shouldShowGpuWarning">
        <div class="row items-center q-gutter-x-sm q-pb-sm">
          <q-icon name="ph-bold ph-warning-circle" color="warning" size="16px"></q-icon>
          <span class="text-weight-bold"> The workload may be preempted </span>
        </div>
        <div>
          A larger difference between the request and limit increases the chance of the workload being preempted. For
          more information, <span v-dompurify-html="docsLinksToDisplay"></span>
        </div>
      </div>
    </transition>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IResourcesSectionModel } from "@/models/compute-resource.model";
import { GpuRequestType } from "@/swagger-models/assets-service-client";
import DOCS from "@/constants/docs";

export default defineComponent({
  name: "preempted-warning-section",
  props: {
    resources: {
      type: Object as PropType<IResourcesSectionModel>,
      required: true,
    },
  },
  computed: {
    shouldShowCpuWarning(): boolean {
      if (!this.resources.cpuMemory.request || !this.resources.cpuMemory.limit) {
        return false;
      }
      return parseFloat(this.resources.cpuMemory.request) < parseFloat(this.resources.cpuMemory.limit);
    },
    shouldShowGpuWarning(): boolean {
      if (!this.resources.gpuRequestType) {
        return false;
      }

      if (this.resources.gpuRequestType === GpuRequestType.Portion && this.resources.gpuPortionRequest) {
        if (!this.resources.gpuPortionLimit && this.resources.gpuPortionLimit !== 0) {
          return false;
        }
        return this.resources.gpuPortionRequest < this.resources.gpuPortionLimit;
      }

      if (this.resources.gpuRequestType === GpuRequestType.Memory && this.resources.gpuMemoryRequest) {
        if (!this.resources.gpuMemoryLimit) {
          return false;
        }
        return parseFloat(this.resources.gpuMemoryRequest) < parseFloat(this.resources.gpuMemoryLimit);
      }

      return false;
    },
    docsLinksToDisplay(): string {
      if (this.shouldShowGpuWarning && !this.shouldShowCpuWarning) {
        return `see the <a href="${DOCS.PLATFORM_MANAGEMENT.DYNAMIC_FRACTION}" target="_blank">Dynamic GPU fractions guide</a>`;
      }

      if (this.shouldShowCpuWarning && !this.shouldShowGpuWarning) {
        return `see <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits" target="_blank">Requests and limits</a>`;
      }

      if (this.shouldShowCpuWarning && this.shouldShowGpuWarning) {
        return `see the <a href="${DOCS.PLATFORM_MANAGEMENT.DYNAMIC_FRACTION}" target="_blank">Dynamic GPU fractions guide</a> and <a href="https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits" target="_blank">Requests and limits</a>`;
      }

      return "";
    },
  },
  methods: {},
});
</script>

<style lang="scss" scoped>
.preempted-warning-section {
  background-color: $yellow-20;
  font-size: 13px;
}
</style>
