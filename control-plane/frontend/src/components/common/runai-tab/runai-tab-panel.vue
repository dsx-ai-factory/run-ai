<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-tab-panel :name="name">
    <slot />
  </q-tab-panel>
</template>

<script setup lang="ts">
export interface Props {
  name: string;
}

defineProps<Props>();
</script>
