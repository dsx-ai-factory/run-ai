<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-history-toolbar">
    <div class="history-filters"><slot name="filters" /></div>
    <runai-general-search
      :search-term="searchTerm"
      :tooltip-text="searchTooltip"
      @update-search-term="$emit('update:searchTerm', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";

withDefaults(defineProps<{ searchTerm: string; searchTooltip?: string }>(), { searchTooltip: "" });
defineEmits<{ "update:searchTerm": [value: string] }>();
</script>

<style lang="scss" scoped>
.workload-history-toolbar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 7px;
  padding: 7px 16px 7px 0;
  border-bottom: 1px solid var(--border-subtle);
}

.history-filters {
  flex: 1;
  min-width: 0;
}
</style>
