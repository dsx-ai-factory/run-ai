<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-tabs
    :model-value="modelValue"
    align="left"
    :class="['tabs-container', { 'tabs-container--underline': variant === 'underline' }]"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <q-tab
      v-for="(tab, index) in visibleTabs"
      :key="tab.name"
      :name="tab.name"
      class="b-px-m"
      :class="tab.tabClass"
      :disable="tab.disabled"
      :aid="tab.aid"
    >
      <slot name="tab" :tab="tab" :index="index" :is-active="tab.name === modelValue">
        <div class="row items-center no-wrap">
          <span class="q-tab__label">{{ tab.label }}</span>
          <q-icon v-if="tab.errorIcon" name="ph-bold ph-warning-circle" size="16px" class="tab-error-icon q-ml-xs" />
          <q-badge v-if="tab.numberingBadge != null" class="tab-numbering-badge q-ml-xs">{{
            tab.numberingBadge
          }}</q-badge>
        </div>
      </slot>
    </q-tab>
  </q-tabs>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ITab } from "./tab.model";

export interface Props {
  modelValue?: string | null;
  tabs: ITab[];
  variant?: "default" | "underline";
}

const props = withDefaults(defineProps<Props>(), { modelValue: null, variant: "default" });

defineEmits<{ (e: "update:modelValue", value: string): void }>();

const visibleTabs = computed(() => props.tabs.filter((tab) => !tab.hidden));
</script>

<style lang="scss" scoped>
.tabs-container {
  background-color: var(--bg-primary-default);
  box-shadow: 0 1px 0 var(--border-subtle);
  z-index: 1;
  height: 50px;

  :deep(.q-tab--active) {
    text-decoration: none !important;
    color: var(--text-brand);
  }

  :deep(.q-tab--inactive) {
    text-decoration: none !important;
    color: var(--b-text-sub);
    opacity: 1;
  }

  .tab-error-icon {
    color: var(--icon-error);
  }

  .tab-numbering-badge {
    background-color: var(--b-text-sub);
  }

  :deep(.q-tabs__arrow) {
    font-size: 17px;
  }

  &.tabs-container--underline {
    :deep(.q-tab__indicator) {
      display: none;
    }
  }
}
</style>
