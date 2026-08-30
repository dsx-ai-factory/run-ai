<template>
  <section ref="panelRef" class="runai-details-panel" aid="details-panel">
    <runai-details-panel-controls
      :model-value="searchTerm"
      :all-expanded="allExpanded"
      @update:model-value="onSearch"
      @toggle-all="onToggleAll"
    >
      <template #prepend>
        <slot name="controls-prepend" />
      </template>
    </runai-details-panel-controls>

    <div class="panel-body">
      <template v-if="visibleSections.length">
        <runai-details-panel-section
          v-for="section in visibleSections"
          :key="section.id"
          :title="section.title"
          :expanded="expandedById[section.id] ?? false"
          :suppress-scroll-on-expand="suppressScroll"
          @update:expanded="(val) => onSectionToggle(section.id, val)"
        >
          <slot :name="`section-${section.id}`" :section="section" :search-term="normalizedSearchTerm" />
        </runai-details-panel-section>
      </template>

      <runai-details-panel-empty-state v-else-if="isSearching" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { nextTick, ref, toRef } from "vue";
import RunaiDetailsPanelControls from "./runai-details-panel-controls.vue";
import RunaiDetailsPanelSection from "./runai-details-panel-section.vue";
import RunaiDetailsPanelEmptyState from "./runai-details-panel-empty-state.vue";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { useDetailsPanel } from "@/composables/use-details-panel.composable/use-details-panel.composable";

defineOptions({ name: "runai-details-panel" });

const props = defineProps<{
  sections: IDetailsPanelSectionDef[];
}>();

const {
  searchTerm,
  normalizedSearchTerm,
  expandedById,
  isSearching,
  visibleSections,
  allExpanded,
  onSearchUpdate,
  onToggle,
  toggleAll,
} = useDetailsPanel({
  sections: toRef(props, "sections"),
});

function onSearch(term: string): void {
  onSearchUpdate(term);
}

function onSectionToggle(sectionId: string, value: boolean): void {
  onToggle(sectionId, value);
}

const panelRef = ref<HTMLElement | null>(null);
const suppressScroll = ref(false);

function onToggleAll(): void {
  // Suppress each section's auto-scrollIntoView during the bulk toggle so the panel doesn't
  // drift to the bottom (the last section's animation would otherwise win the race).
  suppressScroll.value = true;
  toggleAll();
  nextTick(() => {
    panelRef.value?.scrollTo({ top: 0, behavior: "smooth" });
  });
  // Re-enable per-section scroll-into-view after the expansion animations are done so manual
  // expands still bring the opened card into view.
  window.setTimeout(() => {
    suppressScroll.value = false;
  }, 500);
}
</script>

<style lang="scss" scoped>
.runai-details-panel {
  background-color: var(--bg-primary-default);
  width: 100%;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;

  > :first-child {
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .panel-body {
    width: 100%;
  }
}
</style>
