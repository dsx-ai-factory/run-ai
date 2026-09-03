<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div
    v-show="items.length > 0"
    class="echarts-breadcrumbs"
    role="navigation"
    aria-label="Chart drilldown"
    :style="{ top: `${top}px`, left: `${left}px` }"
  >
    <span v-for="(item, index) in items" :key="index">
      <button v-if="index < items.length - 1" type="button" class="breadcrumb-link" @click="emit('navigate', index)">
        {{ item }}
      </button>
      <span v-else class="breadcrumb-current" aria-current="page">{{ item }}</span>
      <span v-if="index < items.length - 1" class="breadcrumb-separator" :style="{ margin: `0 ${separatorGap}px` }">
        /
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    items: string[];
    top?: number;
    left?: number;
    separatorGap?: number;
  }>(),
  {
    top: 14,
    left: 63,
    separatorGap: 8,
  },
);

const emit = defineEmits<{
  (event: "navigate", index: number): void;
}>();
</script>

<style scoped lang="scss">
.echarts-breadcrumbs {
  position: absolute;
  z-index: 1;
  padding: 0;
  font-size: 12px;

  .breadcrumb-link {
    padding: 0;
    font: inherit;
    color: var(--text-link-default);
    cursor: pointer;
    background: transparent;
    border: 0;

    &:hover {
      text-decoration: underline;
    }
  }

  .breadcrumb-current {
    font-weight: 700;
    color: var(--b-text-main);
  }

  .breadcrumb-separator {
    color: var(--text-disabled);
  }
}
</style>
