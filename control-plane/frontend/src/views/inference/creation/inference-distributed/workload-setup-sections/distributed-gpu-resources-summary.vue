<template>
  <div class="distributed-gpu-resources-summary bg-grey-3 q-pa-md">
    <div class="gpu-resources-title q-mb-md">Total requested GPU resources</div>
    <div class="row items-center q-gutter-x-sm">
      <span class="text-h6">{{ totalGpuDisplay }}</span>
      <runai-tooltip :tooltip-text="tooltipText" width="500px" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { WorkspaceSpecSpecAllOfCompute } from "@/swagger-models/workloads-client";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  name: "distributed-gpu-resources-summary",
  components: {
    RunaiTooltip,
  },
  props: {
    leaderCompute: {
      type: Object as PropType<WorkspaceSpecSpecAllOfCompute | undefined>,
      required: false,
    },
    workerCompute: {
      type: Object as PropType<WorkspaceSpecSpecAllOfCompute | undefined>,
      required: false,
    },
    replicas: {
      type: Number as PropType<number>,
      required: true,
    },
    numWorkers: {
      type: Number as PropType<number>,
      required: true,
    },
  },
  computed: {
    leaderGpuInfo(): { value: number; display: string; unit: string } {
      return this.getGpuInfo(this.leaderCompute);
    },
    workerGpuInfo(): { value: number; display: string; unit: string } {
      return this.getGpuInfo(this.workerCompute);
    },
    isWholeGpuAllocation(): boolean {
      // Both leader and workers use whole GPU devices (not memory fractions)
      return this.leaderGpuInfo.unit === "devices" && this.workerGpuInfo.unit === "devices";
    },
    totalGpuDisplay(): string {
      const leaderInfo = this.leaderGpuInfo;
      const workerInfo = this.workerGpuInfo;

      // Worker value includes numWorkers multiplication
      const workerTotalPerReplica = workerInfo.value * this.numWorkers;

      // When both use whole GPUs: (leader + workers) × replicas
      if (this.isWholeGpuAllocation) {
        const total = (leaderInfo.value + workerTotalPerReplica) * this.replicas;
        return this.formatGpuValue(total, "devices");
      }

      // When using fractions/memory: calculate each separately with replicas
      const leaderTotal = leaderInfo.value * this.replicas;
      const workerTotal = workerTotalPerReplica * this.replicas;

      // If both use the same unit type (e.g., both memory), add them
      if (leaderInfo.unit === workerInfo.unit) {
        return this.formatGpuValue(leaderTotal + workerTotal, leaderInfo.unit);
      }

      // Different units: show separately
      return `${this.formatGpuValue(leaderTotal, leaderInfo.unit)} + ${this.formatGpuValue(workerTotal, workerInfo.unit)}`;
    },
    tooltipText(): string {
      const leaderInfo = this.leaderGpuInfo;
      const workerInfo = this.workerGpuInfo;

      const leaderDisplay = this.formatGpuValue(leaderInfo.value, leaderInfo.unit);
      // Worker display shows total for all workers (value × numWorkers)
      const workerTotalValue = workerInfo.value * this.numWorkers;
      const workerDisplay = this.formatGpuValue(workerTotalValue, workerInfo.unit);

      // When both use whole GPUs: ((leader: X) + (workers: Y)) x (Z replicas)
      if (this.isWholeGpuAllocation) {
        return `((Leader: ${leaderDisplay}) + (Workers: ${workerDisplay})) x (${this.replicas} Replicas)`;
      }

      // When using fractions/memory: (leader: X x Z replicas) + (workers: Y x Z replicas)
      return `(Leader: ${leaderDisplay} x ${this.replicas} Replicas) + (Workers: ${workerDisplay} x ${this.replicas} Replicas)`;
    },
  },
  methods: {
    getGpuInfo(compute: WorkspaceSpecSpecAllOfCompute | undefined): { value: number; display: string; unit: string } {
      if (!compute) {
        return { value: 0, display: "0 GPUs", unit: "devices" };
      }

      const gpuDevices = compute.gpuDevicesRequest ?? 1;
      const gpuRequestType = compute.gpuRequestType;

      // Handle GPU memory request
      if (gpuRequestType === "memory" && compute.gpuMemoryRequest) {
        const memoryRequest = compute.gpuMemoryRequest;
        const isGB = memoryRequest.toUpperCase().endsWith("G");
        const memoryUnit = isGB ? "GB" : "MB";
        const memoryValue = parseFloat(memoryRequest) * gpuDevices;
        return {
          value: memoryValue,
          display: `${memoryValue} ${memoryUnit}`,
          unit: memoryUnit,
        };
      }

      // Handle GPU portion (fractioning with % of device)
      if (gpuRequestType === "portion" && compute.gpuPortionRequest != null) {
        const portionValue = gpuDevices * compute.gpuPortionRequest;
        return {
          value: portionValue,
          display: `${portionValue.toFixed(2)} GPUs`,
          unit: "portion",
        };
      }

      // Fallback: check if gpuMemoryRequest exists even without gpuRequestType
      if (compute.gpuMemoryRequest) {
        const memoryRequest = compute.gpuMemoryRequest;
        const isGB = memoryRequest.toUpperCase().endsWith("G");
        const memoryUnit = isGB ? "GB" : "MB";
        const memoryValue = parseFloat(memoryRequest) * gpuDevices;
        return {
          value: memoryValue,
          display: `${memoryValue} ${memoryUnit}`,
          unit: memoryUnit,
        };
      }

      // Fallback: check if gpuPortionRequest exists even without gpuRequestType
      if (compute.gpuPortionRequest != null && compute.gpuPortionRequest !== undefined) {
        const portionValue = gpuDevices * compute.gpuPortionRequest;
        return {
          value: portionValue,
          display: `${portionValue.toFixed(2)} GPUs`,
          unit: "portion",
        };
      }

      // Default: GPU devices (no fractioning)
      const devices = compute.gpuDevicesRequest ?? 0;
      return {
        value: devices,
        display: `${devices} GPUs`,
        unit: "devices",
      };
    },
    formatGpuValue(value: number, unit: string): string {
      if (unit === "portion") {
        return `${value.toFixed(2)} GPUs`;
      }
      if (unit === "GB" || unit === "MB") {
        return `${value} ${unit}`;
      }
      // devices
      return `${value} GPUs`;
    },
  },
});
</script>

<style lang="scss" scoped>
.distributed-gpu-resources-summary {
  .gpu-resources-title {
    font-size: 13px;
    font-weight: 600;
    color: $section-title-color;
  }
}
</style>
