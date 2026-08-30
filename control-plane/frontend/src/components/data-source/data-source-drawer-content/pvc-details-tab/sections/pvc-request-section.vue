<template>
  <div class="pvc-request-section column">
    <runai-details-panel-row v-for="(item, idx) in visibleItems" :key="`${item.label}-${idx}`" :label="item.label">
      <template v-for="(line, lIdx) in valueAsLines(item.value)" :key="lIdx">
        <div>{{ line }}</div>
      </template>
    </runai-details-panel-row>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow } from "@/components/common/runai-details-panel";
import { filterItems } from "@/utils/details-panel.utils";
import type { IPvcRequestSectionData } from "@/models/pvc-details.model";

defineOptions({ name: "pvc-request-section" });

export interface Props {
  data: IPvcRequestSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleItems = computed(() => filterItems(props.data.items, props.searchTerm));

function valueAsLines(value: string | string[]): string[] {
  return Array.isArray(value) ? value : [value];
}
</script>

<style lang="scss" scoped>
.pvc-request-section {
  gap: var(--space-m);
}
</style>
