<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-compute-section column">
    <runai-details-panel-row v-for="(item, idx) in visibleItems" :key="`${item.label}-${idx}`" :label="item.label">
      <runai-details-panel-link v-if="item.link" :link="item.link" />
      <template v-else>
        <div v-for="(line, lIdx) in valueAsLines(item.value)" :key="lIdx">{{ line }}</div>
      </template>
    </runai-details-panel-row>

    <template v-if="visibleNodeAffinity">
      <q-separator v-if="visibleItems.length" />
      <workload-node-affinity-section :node-affinity="visibleNodeAffinity" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow, RunaiDetailsPanelLink } from "@/components/common/runai-details-panel";
import WorkloadNodeAffinitySection from "./workload-node-affinity-section.vue";
import { filterItems, objectMatchesTerm } from "@/utils/details-panel.utils";
import type { IComputeSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-compute-section" });

export interface Props {
  data: IComputeSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleItems = computed(() => filterItems(props.data.items, props.searchTerm));
const visibleNodeAffinity = computed(() => {
  if (!props.data.nodeAffinity) return undefined;
  return objectMatchesTerm(props.data.nodeAffinity, props.searchTerm) ? props.data.nodeAffinity : undefined;
});

function valueAsLines(value: string | string[]): string[] {
  return Array.isArray(value) ? value : [value];
}
</script>

<style lang="scss" scoped>
.workload-compute-section {
  gap: var(--space-m);
}
</style>
