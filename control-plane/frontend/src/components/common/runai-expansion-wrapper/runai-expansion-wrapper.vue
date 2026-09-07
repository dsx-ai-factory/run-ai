<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-list :class="{ 'runai-expansion-wrapper--no-first-top-padding': noFirstTopPadding }"><slot /></q-list>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    noFirstTopPadding?: boolean;
  }>(),
  {
    noFirstTopPadding: false,
  },
);
</script>

<style scoped lang="scss">
.runai-expansion-wrapper--no-first-top-padding {
  // .q-expansion-item--expanded/--popup carry their own 15px top/bottom padding
  // directly on the item's root element; strip just the top padding on the first
  // item so the wrapper's content starts flush (matches Figma), while every
  // subsequent item keeps its normal spacing.
  :deep(> .q-expansion-item:first-child) {
    padding-top: 0 !important;
  }
}
</style>
