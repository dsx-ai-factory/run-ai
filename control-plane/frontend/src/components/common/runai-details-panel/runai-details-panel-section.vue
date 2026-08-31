<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div ref="rootEl" class="runai-details-panel-section">
    <runai-expansion-item
      :model-value="expanded"
      :label="title"
      body-no-padding
      prevent-top-padding-when-open
      class="details-panel-expansion"
      @update:model-value="onUpdate"
      @after-show="onAfterShow"
    >
      <div class="details-panel-section-body">
        <slot />
      </div>
    </runai-expansion-item>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

defineOptions({ name: "runai-details-panel-section" });

const props = withDefaults(
  defineProps<{
    title: string;
    expanded: boolean;
    suppressScrollOnExpand?: boolean;
  }>(),
  { suppressScrollOnExpand: false },
);

const emit = defineEmits<{
  "update:expanded": [value: boolean];
  "after-show": [];
}>();

const rootEl = ref<HTMLElement | null>(null);

function onUpdate(value: boolean): void {
  emit("update:expanded", value);
}

function onAfterShow(): void {
  emit("after-show");
  if (props.suppressScrollOnExpand) return;
  rootEl.value?.scrollIntoView({ block: "start", behavior: "smooth" });
}
</script>

<style lang="scss" scoped>
.runai-details-panel-section {
  :deep(.details-panel-expansion) {
    /* Remove the runai-expansion-item shadow */
    > :first-child {
      box-shadow: none;
    }

    /* Quasar's popup variant adds vertical padding between items and an inner container border —
       both get zeroed here so adjacent cards sit flush with no double border. */
    &.q-expansion-item--popup,
    &.q-expansion-item--popup.q-expansion-item--expanded,
    &.q-expansion-item--popup.q-expansion-item--collapsed {
      padding-top: 0;
      padding-bottom: 0;
    }

    &.q-expansion-item--popup > .q-expansion-item__container {
      border: none;
      box-shadow: none;
    }

    /* Single 1 px separator between sections (also under the last section). */
    border-bottom: 1px solid var(--border-subtle);

    .runai-expansion-header {
      padding: var(--space-m) var(--space-xl) var(--space-m) var(--space-m);
      min-height: 48px;
    }

    /* When expanded, the header gets its own bottom border so it visually
       separates from the body content. */
    &.q-expansion-item--expanded .runai-expansion-header {
      border-bottom: 1px solid var(--border-subtle);
    }

    .runai-expansion-divider {
      display: none;
    }

    .runai-expansion-content {
      background-color: var(--bg-primary-default);
    }
  }

  .details-panel-section-body {
    /* Per Figma: 16px padding on all sides. The 20 px gap (label↔value)
       is enforced at the row level (runai-details-panel-row), not here, because our
       structure is row-by-row rather than Figma's two-column structure. */
    padding: var(--space-m);
  }
}
</style>
