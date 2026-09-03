<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td @click.stop class="editable-cell-popup" :class="cellClass">
    <div
      class="cell-trigger text-b-main"
      role="button"
      tabindex="0"
      :aria-disabled="disabled"
      @click.stop="handleTriggerClick"
      @keydown.enter.prevent="handleTriggerClick"
      @keydown.space.prevent="handleTriggerClick"
    >
      <slot name="display">{{ displayValue }}</slot>
      <q-tooltip v-if="disabled && disabledTooltip" max-width="300px">{{ disabledTooltip }}</q-tooltip>
    </div>
    <q-menu
      v-model="isOpen"
      anchor="bottom middle"
      self="top middle"
      class="editable-cell-popup-menu"
      @show="emit('opened')"
      @hide="handleMenuHide"
    >
      <q-card class="popup-card">
        <q-card-section class="row justify-between items-center q-py-sm">
          <div class="text-weight-medium">{{ title }}</div>
          <q-btn flat icon="ph-bold ph-x" class="q-pa-xs runai-icon-btn--default" round v-close-popup />
        </q-card-section>

        <q-card-section class="q-py-sm q-px-md">
          <slot name="input"></slot>
        </q-card-section>

        <q-card-section class="row justify-end q-py-sm">
          <q-btn flat color="primary" label="APPLY" @click="handleApplyClick" v-close-popup="!isSubmitDisabled" />
        </q-card-section>
      </q-card>
    </q-menu>
  </q-td>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

export type EditableCellState = "default" | "active" | "warning" | "error";

defineOptions({ name: "EditableCellPopup" });

const props = withDefaults(
  defineProps<{
    title: string;
    displayValue?: string | number;
    state?: EditableCellState;
    isSubmitDisabled?: boolean; // Disables the Apply/Submit button (e.g., when validation fails)
    disabled?: boolean; // Disables the entire cell (prevents opening the popup)
    disabledTooltip?: string;
  }>(),
  {
    displayValue: "-",
    state: "default",
    isSubmitDisabled: false,
    disabled: false,
    disabledTooltip: "",
  },
);

const emit = defineEmits<{
  apply: [];
  reset: [];
  opened: [];
}>();

const isOpen = ref(false);

const cellClass = computed(() => {
  if (props.disabled) return "cell--disabled";
  if (isOpen.value) return "cell--active";
  return "";
});

const handleTriggerClick = () => {
  if (props.disabled) return;
  isOpen.value = true;
};

const handleMenuHide = () => {
  emit("reset");
};

const handleApplyClick = () => {
  if (props.isSubmitDisabled) return;
  emit("apply");
  isOpen.value = false;
};
</script>

<style scoped lang="scss">
.editable-cell-popup {
  cursor: pointer;

  .cell-trigger {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100px;
    height: 40px;
    border: 1px solid $grey-4;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s ease;

    &:hover {
      border-color: $grey-9;
    }
  }

  &.cell--active .cell-trigger {
    border: 2px solid $light-green-7;
  }

  &.cell--error .cell-trigger {
    border-color: $negative;
    color: $negative;
  }

  &.cell--warning .cell-trigger {
    border-color: $amber-14;
    color: $amber-14;
  }

  &.cell--disabled .cell-trigger {
    color: var(--b-text-main);
    cursor: not-allowed;
    &:hover {
      border-color: $grey-4;
    }
  }
}
</style>

<style lang="scss">
.editable-cell-popup-menu {
  // Override Quasar's inline max-height to allow the popup to grow when validation errors appear
  max-height: none !important;
  .popup-card {
    width: 200px;
  }
}
</style>
