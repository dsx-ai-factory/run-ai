<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="creation-summary" :class="{ 'creation-summary--scheduling': scheduling }">
    <section class="summary-section">
      <div class="summary-heading">Cluster</div>
      <div>{{ pool.clusterName || clusterName || "None" }}</div>
    </section>
    <q-separator class="q-my-md" />
    <section class="summary-section">
      <div class="summary-heading resources-heading">
        Resources
        <runai-tooltip
          size="xs"
          class="resources-help"
          tooltip-text="All the resources available for matching labeled nodes within the node pool"
          aria-label="Node pool resources"
        />
      </div>
      <div>
        <div v-for="row in resourceRows" :key="row.label">{{ row.label }}: {{ row.value }}</div>
      </div>
    </section>
    <q-separator class="q-my-md" />
    <section class="summary-section summary-mode">
      <div class="summary-heading">Scheduling mode</div>
      <div>
        <div class="summary-mode-name text-weight-bold q-mb-xs">{{ schedulingMode }}</div>
        <div v-if="modeDescription">{{ modeDescription }}</div>
      </div>
    </section>
    <q-separator class="q-my-md" />
    <section class="summary-section summary-topology">
      <div class="summary-heading">Topology &amp; placement</div>
      <div>
        <div v-for="row in configurationRows" :key="row.label">{{ row.label }}: {{ row.value }}</div>
      </div>
    </section>
  </div>
</template>
<script setup lang="ts">
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { memoryFormat } from "@/utils/format.util";
import { NODE_POOL_CONFIGURATION_DEFAULTS as defaults } from "@/models/node-pool.model";
import {
  getSchedulingMode,
  NETWORK_ACCELERATION_LABELS,
  SCHEDULING_MODE_DESCRIPTIONS,
  type INodePoolSummaryConfiguration,
  type INodePoolSummaryResources,
} from "@/models/node-pool.model";

const props = withDefaults(
  defineProps<{
    pool: INodePoolSummaryConfiguration;
    scheduling?: boolean;
    clusterName?: string;
    resources?: INodePoolSummaryResources;
    showNuma?: boolean;
    showGpuDevice?: boolean;
  }>(),
  { showNuma: true, showGpuDevice: true },
);
const schedulingMode = computed(() => getSchedulingMode(props.pool.schedulingCapabilities));
const modeDescription = computed(() => SCHEDULING_MODE_DESCRIPTIONS[schedulingMode.value]);
const resourceRows = computed(() => [
  { label: "Nodes", value: props.resources?.nodes ?? "None" },
  { label: "GPUs", value: props.resources?.gpus ?? "None" },
  {
    label: "GPU memory",
    value: props.resources?.gpuMemoryBytes == null ? "None" : memoryFormat(props.resources.gpuMemoryBytes),
  },
  { label: "GPU types", value: props.resources?.gpuTypes?.join(", ") || "None" },
]);
const configurationRows = computed(() => {
  const pool = props.pool;
  const config = pool.schedulingConfiguration;
  return [
    { label: "Network topology", value: pool.networkTopologyName || pool.networkTopologyId || "None" },
    {
      label: "MNNVL",
      value:
        NETWORK_ACCELERATION_LABELS[pool.gpuNetworkAccelerationDetection ?? defaults.gpuNetworkAccelerationDetection],
    },
    {
      label: "NUMA-aware scheduling",
      value: (config?.numaAwareScheduling ?? defaults.numaAwareScheduling) ? "Enabled" : "Disabled",
    },
    {
      label: "GPU node placement strategy",
      value: nodePoolUtil.placementStrategyText(config?.gpuPodsPlacementStrategy?.node ?? defaults.gpuNodePlacement),
    },
    {
      label: "GPU device placement strategy",
      value: nodePoolUtil.placementStrategyText(config?.gpuPodsPlacementStrategy?.device ?? defaults.gpuDevicePlacement),
    },
    {
      label: "CPU node placement strategy",
      value: nodePoolUtil.placementStrategyText(config?.cpuPodsPlacementStrategy?.node ?? defaults.cpuNodePlacement),
    },
  ].filter(
    (row) =>
      (props.showNuma || row.label !== "NUMA-aware scheduling") &&
      (props.showGpuDevice || row.label !== "GPU device placement strategy"),
  );
});
</script>
<style scoped lang="scss">
.creation-summary {
  padding: 16px;
  font-size: 12px;
  line-height: 1.4;
  overflow-wrap: anywhere;
}
.resources-heading {
  display: flex;
  align-items: center;
  gap: 10px;
}
.resources-help {
  width: 20px;
  height: 20px;
  min-height: 20px;
  padding: 0;
}
.resources-help :deep(.q-icon) {
  font-size: 16px;
}
.summary-heading {
  color: var(--b-text-sub);
  margin-bottom: 10px;
}
.creation-summary--scheduling {
  padding: 7px 16px 16px;
}
.creation-summary--scheduling .summary-section {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.creation-summary--scheduling .summary-heading {
  display: flex;
  align-items: center;
  height: 32px;
  margin: 0;
  font-weight: 500;
}
.creation-summary--scheduling .summary-mode,
.creation-summary--scheduling .summary-topology {
  gap: 5px;
}
.creation-summary--scheduling .summary-mode-name {
  margin-bottom: 5px;
  font-weight: 800;
}
.creation-summary--scheduling > .q-separator {
  margin: 20px 0;
}
.creation-summary--scheduling .resources-heading {
  gap: 7px;
}
.creation-summary--scheduling .resources-help {
  width: 32px;
  height: 32px;
  min-height: 32px;
}
.creation-summary--scheduling .resources-help :deep(.q-icon) {
  font-size: 20px;
}
</style>
