<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-general-info-section column">
    <div v-for="(group, gIdx) in visibleGroups" :key="`g-${gIdx}`" class="details-group column">
      <runai-details-panel-row v-for="(item, idx) in group" :key="`${item.label}-${idx}`" :label="item.label">
        <runai-details-panel-link v-if="item.link" :link="item.link" />
        <template v-else>{{ valueAsString(item.value) }}</template>
      </runai-details-panel-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow, RunaiDetailsPanelLink } from "@/components/common/runai-details-panel";
import { filterItems } from "@/utils/details-panel.utils";
import type { IGeneralSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-general-info-section" });

export interface Props {
  data: IGeneralSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleGroups = computed(() =>
  props.data.groups.map((group) => filterItems(group, props.searchTerm)).filter((g) => g.length > 0),
);

function valueAsString(value: string | string[]): string {
  return Array.isArray(value) ? value.join(", ") : value;
}
</script>

<style lang="scss" scoped>
.workload-general-info-section {
  gap: var(--space-m);

  .details-group {
    gap: var(--space-xs);
  }
}
</style>
