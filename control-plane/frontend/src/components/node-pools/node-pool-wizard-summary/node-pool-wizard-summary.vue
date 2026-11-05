<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="creation-summary">
    <div class="summary-heading">Cluster</div>
    <div>{{ pool.clusterName || clusterName || "None" }}</div>
    <q-separator class="q-my-md" />
    <div class="summary-heading">Resources</div>
    <div v-for="row in resourceRows" :key="row.label">{{ row.label }}: {{ row.value }}</div>
    <q-separator class="q-my-md" />
    <div class="summary-heading">Scheduling mode</div>
    <div class="text-weight-bold q-mb-xs">{{ schedulingMode }}</div>
    <div v-if="modeDescription">{{ modeDescription }}</div>
    <q-separator class="q-my-md" />
    <div class="summary-heading">Topology & placement</div>
    <div v-for="row in configurationRows" :key="row.label">{{ row.label }}: {{ row.value }}</div>
  </div>
</template>
<script setup lang="ts">
import { computed } from "vue";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { memoryFormat } from "@/utils/format.util";
import { NODE_POOL_CONFIGURATION_DEFAULTS as defaults } from "@/components/node-pools/node-pool-create-flow/node-pool-create-flow.model";
import {
  getSchedulingMode,
  NETWORK_ACCELERATION_LABELS,
  SCHEDULING_MODE_DESCRIPTIONS,
  type NodePoolSummaryConfiguration,
  type INodePoolSummaryResources,
} from "./node-pool-wizard-summary.model";

const props = defineProps<{
  pool: NodePoolSummaryConfiguration;
  clusterName?: string;
  resources?: INodePoolSummaryResources;
}>();
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
  ];
});
</script>
<style scoped lang="scss">
.creation-summary {
  padding: 16px;
  font-size: 12px;
  line-height: 1.4;
  overflow-wrap: anywhere;
}
.summary-heading {
  color: var(--b-text-sub);
  margin-bottom: 10px;
}
</style>
