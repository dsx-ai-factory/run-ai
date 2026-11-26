<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="node-pools-priority-list row no-wrap">
    <runai-priority-list
      v-if="showList"
      aid="node-pools-priority-list"
      title="Node pools"
      subtitle="Priority order"
      :tooltip="EOrgUnitFormTooltip.NodePoolsPriorityOrder"
      :items="nodePools"
      :model-value="modelValue"
      :disable-reorder="disableReorder"
      :disable-reorder-tooltip="disableReorderTooltip"
      @update:model-value="emit('update:model-value', $event)"
      @update:items="emit('update:node-pools', $event)"
    />
    <div class="node-pools-priority-list-content col">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiPriorityList } from "@/components/common/runai-priority-list";
import { EOrgUnitFormTooltip } from "@/models/org-unit-forms.model";

export interface Props {
  nodePools: string[];
  modelValue: string;
  disableReorder?: boolean;
  disableReorderTooltip?: string;
}

const props = withDefaults(defineProps<Props>(), {
  disableReorder: false,
});

const emit = defineEmits<{
  "update:model-value": [nodePool: string];
  "update:node-pools": [nodePools: string[]];
}>();

const showList = computed(() => props.nodePools.length > 1);
</script>

<style scoped lang="scss">
.node-pools-priority-list {
  width: 100%;
}

.node-pools-priority-list-content {
  min-width: 0;
}
</style>
