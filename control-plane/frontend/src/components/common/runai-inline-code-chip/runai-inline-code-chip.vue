<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <code class="runai-inline-code-chip text-b-mono">{{ text }}</code>
</template>

<script setup lang="ts">
defineOptions({ name: "runai-inline-code-chip" });

defineProps<{
  text: string;
}>();
</script>

<style lang="scss" scoped>
.runai-inline-code-chip {
  display: inline-block;
  background: var(--bg-secondary-default, #{$body-background-color});
  border-radius: var(--space-2xs);
  padding: var(--space-3xs) var(--space-xs);
}
</style>
