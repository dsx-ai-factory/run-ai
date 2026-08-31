<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <template v-for="(segment, index) in segments" :key="`${segment.text}-${index}`">
    <mark v-if="segment.highlighted" class="search-highlight">{{ segment.text }}</mark>
    <span v-else>{{ segment.text }}</span>
  </template>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { stringUtil } from "@/utils/string.util";

defineOptions({ name: "runai-highlighted-text" });

const props = withDefaults(
  defineProps<{
    text: string;
    searchTerm?: string;
  }>(),
  { searchTerm: "" },
);

const segments = computed(() => stringUtil.getHighlightedTextSegments(props.text, props.searchTerm));
</script>

<style lang="scss" scoped>
.search-highlight {
  padding: 0;
  background: $search-highlight-bg;
  color: inherit;
}
</style>
