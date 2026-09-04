<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <!-- Tabs bar - hidden when only one available node pool -->
  <div v-if="showTabs" class="node-pool-tabs-container row items-center no-wrap">
    <runai-tab
      :model-value="modelValue"
      outside-arrows
      class="node-pool-tabs"
      :tabs="nodePoolTabs"
      @update:model-value="emit('update:model-value', $event)"
    >
      <template #tab="{ tab, index, isActive }">
        <div class="row items-center no-wrap">
          <q-badge class="q-mr-sm" :outline="!isActive" :color="isActive ? 'primary' : 'b-sub'">
            {{ getTabBadgeLabel(index) }}
          </q-badge>
          <span class="q-tab__label">{{ tab.label }}</span>
          <runai-btn
            aid="nodepool-options-btn"
            round
            color="grey-4"
            text-color="b-text-sub"
            unelevated
            size="sm"
            icon="ph-bold ph-dots-three-vertical"
            class="options-btn q-ml-sm"
            :aria-label="`Node pool ${tab.label} options`"
            @click.stop.prevent
          >
            <runai-select-menu
              :list="getTabMenuOptions(index)"
              tooltip-width="300px"
              @selected-value="(action) => handleMenuAction(action, tab.name)"
              @before-show="openMenuId = tab.name"
              @hide="openMenuId === tab.name && (openMenuId = null)"
            />
          </runai-btn>
        </div>
      </template>
    </runai-tab>

    <!-- Add Node Pool Button -->
    <runai-btn
      aid="add-nodepool-btn"
      round
      flat
      size="sm"
      color="black-54"
      tooltip-anchor="top middle"
      tooltip-self="bottom middle"
      :disable="disableListActions || !hasRemainingPools"
      icon="ph-bold ph-plus"
      class="add-nodepool-btn q-mx-sm"
      :tooltip="!disableListActions && hasRemainingPools ? 'Add node pools' : undefined"
      :disabled-tooltip="addButtonDisabledTooltip"
      aria-label="Add node pool"
    >
      <runai-select-menu
        v-if="!disableListActions && hasRemainingPools"
        :list="addNodePoolOptions"
        @selected-value="handleAddSelection"
        @hide="handleAddMenuHide"
      />
    </runai-btn>
  </div>

  <!-- Panel content - always rendered, q-tab-panels works independently of q-tabs -->
  <runai-tab-panels :model-value="modelValue" keep-alive>
    <runai-tab-panel
      v-for="nodePool in nodePools"
      :name="nodePool.id"
      :key="nodePool.id"
      :aria-label="`Node pool: ${nodePool.name}`"
    >
      <slot />
    </runai-tab-panel>
  </runai-tab-panels>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiSelectMenu } from "@/components/common/runai-select-menu";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel, type ITab } from "@/components/common/runai-tab";
import type { ISelectOption } from "@/models/global.model";

export interface NodePoolOption {
  name: string;
  id: string;
}

export interface Props {
  nodePools: NodePoolOption[];
  modelValue: string;
  availableNodePools?: NodePoolOption[];
  /** When true, disables add/remove/reorder actions */
  disableListActions?: boolean;
  /** Tooltip to show when list actions are disabled */
  disabledActionsTooltip?: string;
}

const DISABLED_ADD_ACTION_TOOLTIP = "To add node pools, go to the GPU quota step";
const DISABLED_ACTIONS_TOOLTIP = "To perform this action, go to the GPU quota step";

const props = withDefaults(defineProps<Props>(), {
  availableNodePools: () => [],
  disableListActions: false,
});

const emit = defineEmits<{
  (e: "update:model-value", nodePoolId: string): void;
  (e: "update:node-pools", nodePoolNames: string[]): void;
}>();

const openMenuId = ref<string | null>(null);

const nodePoolTabs = computed<ITab[]>(() =>
  props.nodePools.map((pool) => ({
    name: pool.id,
    label: pool.name,
    tabClass: { "menu-open": openMenuId.value === pool.id },
    aid: `nodepool-tab-${pool.id}`,
  })),
);

// Pending add update - only needed for add menu (button shifts position)
const pendingAddUpdate = ref<NodePoolOption[] | null>(null);

const moveLeftTooltipText =
  "Reorder node pools to define the order in which workloads are scheduled. The leftmost node pool is the first one workloads will attempt to schedule onto.";

// Compute remaining node pools (available - selected)
const selectedNodePoolIds = computed(() => new Set(props.nodePools.map((np) => np.id)));

const remainingNodePools = computed(() =>
  props.availableNodePools.filter((pool) => !selectedNodePoolIds.value.has(pool.id)),
);

const hasRemainingPools = computed(() => remainingNodePools.value.length > 0);

// Show tabs only when there are multiple available node pools
const showTabs = computed(() => props.availableNodePools.length > 1);

// Tooltip for disabled add button - prioritize list actions disabled over no remaining pools
const addButtonDisabledTooltip = computed(() => {
  if (props.disableListActions) {
    return props.disabledActionsTooltip ?? DISABLED_ADD_ACTION_TOOLTIP;
  }
  if (!hasRemainingPools.value) {
    return "No more node pools to add";
  }
  return undefined;
});

const getTabMenuOptions = (index: number): ISelectOption[] => [
  {
    label: "Move to the left",
    value: "move-left",
    disable: props.disableListActions || index === 0,
    // Show disabled tooltip when list actions are disabled, otherwise show the info tooltip
    warningTooltip: moveLeftTooltipText,
    aid: "move-left-btn",
    toolTip: props.disableListActions ? (props.disabledActionsTooltip ?? DISABLED_ACTIONS_TOOLTIP) : undefined,
  },
  {
    label: "Remove",
    value: "remove",
    disable: props.disableListActions || props.nodePools.length <= 1,
    // Show disabled tooltip when list actions are disabled
    toolTip: props.disableListActions ? (props.disabledActionsTooltip ?? DISABLED_ACTIONS_TOOLTIP) : undefined,
    aid: "remove-nodepool-btn",
  },
];

const addNodePoolOptions = computed<ISelectOption[]>(() =>
  remainingNodePools.value.map((pool) => ({
    label: pool.name,
    value: pool.id,
    aid: `add-nodepool-${pool.id}-btn`,
  })),
);

// Emit updated node pool names
const emitUpdatedNodePools = (updatedPools: NodePoolOption[]) => {
  emit(
    "update:node-pools",
    updatedPools.map((np) => np.name),
  );
};

// Handle move left - swap with previous item (emit directly, no position shift issue)
const handleMoveLeft = (nodePoolId: string) => {
  const index = props.nodePools.findIndex((np) => np.id === nodePoolId);
  if (index <= 0) return;

  const updatedPools = [...props.nodePools];
  [updatedPools[index - 1], updatedPools[index]] = [updatedPools[index], updatedPools[index - 1]];
  emitUpdatedNodePools(updatedPools);
};

// Handle remove - remove from list (emit directly, tab gets removed anyway)
const handleRemove = (nodePoolId: string) => {
  const updatedPools = props.nodePools.filter((np) => np.id !== nodePoolId);
  emitUpdatedNodePools(updatedPools);
};

// Handle add selection - store pending update, wait for menu to close
// (needed because the + button shifts position when new tab appears)
const handleAddSelection = (nodePoolId: string) => {
  const poolToAdd = props.availableNodePools.find((np) => np.id === nodePoolId);
  if (!poolToAdd) return;

  const updatedPools = [...props.nodePools, poolToAdd];
  pendingAddUpdate.value = updatedPools;
};

// When add menu closes, apply pending update
const handleAddMenuHide = () => {
  if (pendingAddUpdate.value) {
    // Clear any stale menu state before adding new tab
    openMenuId.value = null;
    emitUpdatedNodePools(pendingAddUpdate.value);
    pendingAddUpdate.value = null;
  }
};

const handleMenuAction = (action: string, nodePoolId: string) => {
  if (action === "move-left") {
    handleMoveLeft(nodePoolId);
  } else if (action === "remove") {
    handleRemove(nodePoolId);
  }
};

const getTabBadgeLabel = (index: number): string => {
  // returns 1st, 2nd, 3rd, 21st, 22nd, 23rd, etc.
  const num = index + 1;
  const lastTwoDigits = num % 100;

  // Special case for 11, 12, 13 (and 111, 112, 113, etc.)
  if (lastTwoDigits >= 11 && lastTwoDigits <= 13) {
    return `${num}th`;
  }

  const lastDigit = num % 10;
  switch (lastDigit) {
    case 1:
      return `${num}st`;
    case 2:
      return `${num}nd`;
    case 3:
      return `${num}rd`;
    default:
      return `${num}th`;
  }
};
</script>

<style lang="scss">
.node-pool-tabs-container {
  // Ensure consistent badge size between filled and outlined variants
  .q-badge:not(.q-badge--outline) {
    // Add invisible border to filled badges to match outlined badge dimensions
    border: 1px solid transparent;
  }

  .node-pool-tabs {
    min-width: 0; // Allows flex item to shrink below content size

    .q-tab {
      position: relative;

      .options-btn {
        color: $black-54;
        font-size: 12px;
        display: none;
        position: absolute;
        right: -12px;
      }

      &:hover,
      &.menu-open {
        .options-btn {
          display: inline-flex;
        }
      }
    }
  }

  .add-nodepool-btn {
    flex-shrink: 0; // Never shrink, always keep its size
  }
}
</style>
