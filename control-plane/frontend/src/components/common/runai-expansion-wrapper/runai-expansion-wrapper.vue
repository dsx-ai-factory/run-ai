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
  :deep(> .q-expansion-item:first-child) {
    padding-top: 0 !important;
  }
}
</style>
