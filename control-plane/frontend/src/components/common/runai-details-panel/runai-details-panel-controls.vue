<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-details-panel-controls row items-center" aid="details-panel-controls">
    <slot name="prepend" />
    <q-space />
    <div class="row items-center no-wrap right-cluster">
      <runai-general-search
        :search-term="modelValue"
        :show-search-text="false"
        tooltip-text="Search"
        @update-search-term="onSearch"
      />
      <runai-btn
        aid="details-panel-toggle-all-btn"
        flat
        round
        dense
        size="12px"
        :icon="toggleIcon"
        :tooltip="toggleTooltip"
        :aria-label="toggleTooltip"
        class="details-panel-toggle-all-btn"
        @click="emit('toggle-all')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";

defineOptions({ name: "runai-details-panel-controls" });

const props = defineProps<{
  modelValue: string;
  allExpanded: boolean;
}>();

const emit = defineEmits<{
  "update:model-value": [value: string];
  "toggle-all": [];
}>();

const toggleIcon = computed(() =>
  props.allExpanded ? "ph-bold ph-arrows-in-line-vertical" : "ph-bold ph-arrows-out-line-vertical",
);

const toggleTooltip = computed(() => (props.allExpanded ? "Collapse all" : "Expand all"));

function onSearch(value: string): void {
  emit("update:model-value", value);
}
</script>

<style lang="scss" scoped>
.runai-details-panel-controls {
  height: 54px;
  min-height: 54px;
  padding: var(--space-xs) var(--space-s);
  border-bottom: 1px solid var(--border-default);
  background-color: var(--bg-primary-default);

  .right-cluster {
    gap: var(--space-s);
  }

  :deep(.details-panel-toggle-all-btn) {
    color: var(--icon-default, rgba(0, 0, 0, 0.54));
  }

  :deep(.runai-general-search .icon-container i) {
    color: var(--icon-default, rgba(0, 0, 0, 0.54));
  }
}
</style>
