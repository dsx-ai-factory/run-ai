<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-network-section column">
    <div v-for="(conn, idx) in visibleConnections" :key="`${conn.name}-${idx}`" class="connection column">
      <div class="connection-name text-b-subtitle1">{{ conn.name }}</div>
      <runai-details-panel-row v-if="conn.tool" label="Tool">{{ conn.tool }}</runai-details-panel-row>
      <runai-details-panel-row v-if="conn.connectionType" label="Connection type">
        {{ conn.connectionType }}
      </runai-details-panel-row>
      <runai-details-panel-row label="Access">{{ conn.access }}</runai-details-panel-row>
      <runai-details-panel-row v-if="conn.port !== undefined" label="Port">{{ conn.port }}</runai-details-panel-row>
      <runai-details-panel-row v-if="conn.url" label="URL">
        <runai-details-panel-copyable-text :text="conn.url" />
      </runai-details-panel-row>
    </div>

    <template v-if="visibleTopology">
      <q-separator v-if="visibleConnections.length" />
      <div class="topology-title text-b-body1-md">Network topology</div>
      <workload-topology-section :topology="visibleTopology" :levels="data.topologyLevels ?? []" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow, RunaiDetailsPanelCopyableText } from "@/components/common/runai-details-panel";
import WorkloadTopologySection from "./workload-topology-section.vue";
import { filterObjects, objectMatchesTerm } from "@/utils/details-panel.utils";
import type { INetworkSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-network-section" });

export interface Props {
  data: INetworkSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleConnections = computed(() => filterObjects(props.data.connections, props.searchTerm));
const visibleTopology = computed(() => {
  const topology = props.data.topology;
  if (!topology?.name) return undefined;
  const matches = objectMatchesTerm({ ...topology, levels: props.data.topologyLevels }, props.searchTerm);
  return matches ? topology : undefined;
});
</script>

<style lang="scss" scoped>
.workload-network-section {
  gap: var(--space-l);

  .connection {
    gap: var(--space-xs);
  }

  .connection-name {
    color: var(--b-text-main);
    margin-bottom: var(--space-2xs);
  }

  .topology-title {
    color: var(--b-text-main);
  }
}
</style>
