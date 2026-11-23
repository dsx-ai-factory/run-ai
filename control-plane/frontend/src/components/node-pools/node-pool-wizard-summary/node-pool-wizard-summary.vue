<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="creation-summary text-b-body1">
    <section class="summary-section">
      <h2 class="summary-heading">Cluster</h2>
      <div>{{ pool.clusterName || clusterName || "None" }}</div>
    </section>
    <q-separator class="summary-divider" />
    <section class="summary-section">
      <div class="resources-heading">
        <h2 class="summary-heading">Resources</h2>
        <runai-tooltip
          size="xs"
          class="resources-help"
          tooltip-text="All the resources available for matching labeled nodes within the node pool"
          aria-label="About node pool resources"
        />
      </div>
      <dl class="summary-rows resource-rows">
        <div v-for="row in resourceRows" :key="row.label">
          <dt>{{ row.label }}</dt>
          <dd>{{ row.value }}</dd>
        </div>
      </dl>
    </section>
    <q-separator class="summary-divider" />
    <section class="summary-section summary-mode">
      <h2 class="summary-heading">Scheduling mode</h2>
      <div class="text-b-body1-bold">{{ schedulingMode }}</div>
      <div v-if="modeDescription">{{ modeDescription }}</div>
    </section>
    <q-separator class="summary-divider" />
    <section class="summary-section summary-topology">
      <h2 class="summary-heading">Topology &amp; placement</h2>
      <dl class="summary-rows topology-rows">
        <div v-for="row in configurationRows" :key="row.label">
          <dt>{{ row.label }}</dt>
          <dd>{{ row.value }}</dd>
        </div>
      </dl>
    </section>
  </div>
</template>
<script setup lang="ts">
import { computed } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { memoryFormat } from "@/utils/format.util";
import {
  getSchedulingMode,
  NETWORK_ACCELERATION_LABELS,
  SCHEDULING_MODE_DESCRIPTIONS,
  type INodePoolSummaryConfiguration,
  type INodePoolSummaryResources,
} from "@/models/node-pool.model";

import { topologyDraftFromPool, type ITopologySupport } from "@/models/node-pool-topology.model";

const props = defineProps<{
  pool: INodePoolSummaryConfiguration;
  clusterName?: string;
  resources?: INodePoolSummaryResources;
  supports: ITopologySupport;
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
  const config = topologyDraftFromPool(pool);
  const rows = [{ label: "Network topology", value: pool.networkTopologyName || pool.networkTopologyId || "None" }];
  if (props.supports.mnnvl) {
    rows.push({
      label: "MNNVL",
      value: NETWORK_ACCELERATION_LABELS[config.gpuNetworkAccelerationDetection],
    });
  }
  if (props.supports.numa) {
    rows.push({
      label: "NUMA-aware scheduling",
      value: config.numaAwareScheduling ? "Enabled" : "Disabled",
    });
  }
  rows.push({
    label: "GPU node placement strategy",
    value: nodePoolUtil.placementStrategyText(config.gpuNodePlacement),
  });
  if (props.supports.gpuDevice) {
    rows.push({
      label: "GPU device placement strategy",
      value: nodePoolUtil.placementStrategyText(config.gpuDevicePlacement),
    });
  }
  rows.push({
    label: "CPU node placement strategy",
    value: nodePoolUtil.placementStrategyText(config.cpuNodePlacement),
  });
  return rows;
});
</script>
<style scoped lang="scss">
.creation-summary {
  display: flex;
  flex-direction: column;
  gap: var(--space-s);
  padding-block: var(--space-xs) var(--space-m);
  overflow-wrap: anywhere;
}
.summary-divider {
  background: var(--border-subtle);
}
.summary-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
  padding-inline: var(--space-m);
}
.summary-heading {
  display: flex;
  align-items: center;
  min-height: 32px;
  margin: 0;
  color: var(--b-text-sub);
  font-size: inherit;
  font-weight: 500;
  line-height: 14px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.resources-heading {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}
.resources-help {
  width: 32px;
  min-height: 32px;
  padding: 0;
}
.resources-help :deep(.q-icon) {
  font-size: 20px;
}
.summary-rows {
  display: flex;
  flex-direction: column;
  margin: 0;

  dt {
    color: var(--b-text-sub);
  }

  dd {
    margin: 0;
  }
}
.summary-mode,
.resource-rows {
  gap: var(--space-2xs);
}
.resource-rows > div {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: var(--space-2xs);

  dt {
    flex-shrink: 0;
  }

  dd {
    text-align: end;
  }
}
.summary-topology {
  gap: var(--space-xs);
}
.topology-rows {
  gap: var(--space-s);
}
</style>
