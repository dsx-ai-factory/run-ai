<template>
  <div class="workload-storage-section column">
    <div v-for="(volume, idx) in visibleVolumes" :key="`${volume.name}-${idx}`" class="volume column">
      <div class="volume-name text-b-body1-md">{{ volume.name }}</div>
      <runai-details-panel-row v-for="field in volume.fields" :key="field.label" :label="field.label">
        <runai-details-panel-copyable-text v-if="isUrl(field.value)" :text="field.value" />
        <runai-inline-code-chip v-else-if="field.isCode" :text="field.value" />
        <span v-else :class="{ 'text-b-body1-md': field.label === 'Type' }">{{ field.value }}</span>
      </runai-details-panel-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanelRow, RunaiDetailsPanelCopyableText } from "@/components/common/runai-details-panel";
import { RunaiInlineCodeChip } from "@/components/common/runai-inline-code-chip";
import { filterObjects } from "@/utils/details-panel.utils";
import type { IStorageSectionData } from "@/models/workload-details.model";

defineOptions({ name: "workload-storage-section" });

export interface Props {
  data: IStorageSectionData;
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), { searchTerm: "" });

const visibleVolumes = computed(() => filterObjects(props.data.volumes, props.searchTerm));

function isUrl(value: string): boolean {
  return /^https?:\/\//.test(value);
}
</script>

<style lang="scss" scoped>
.workload-storage-section {
  gap: var(--space-l);
}

.volume {
  gap: var(--space-s);
}

.volume-name {
  color: var(--b-text-main);
}
</style>
