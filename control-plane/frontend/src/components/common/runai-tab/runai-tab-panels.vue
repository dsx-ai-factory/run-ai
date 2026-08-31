<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-tab-panels
    :model-value="modelValue"
    animated
    :keep-alive="keepAlive"
    @update:model-value="$emit('update:modelValue', $event as string)"
  >
    <slot />
  </q-tab-panels>
</template>

<script setup lang="ts">
export interface Props {
  modelValue?: string | null;
  keepAlive?: boolean;
}

withDefaults(defineProps<Props>(), { modelValue: null, keepAlive: false });

defineEmits<{ (e: "update:modelValue", value: string): void }>();
</script>
