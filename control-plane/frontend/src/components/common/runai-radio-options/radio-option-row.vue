<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="radio-option-row row items-center no-wrap" :class="{ ellipsis: maxWidth }" :style="{ maxWidth }">
    <div class="row items-center min-w-0 col-grow">
      <runai-svg-icon v-if="option.icon" class="q-mr-sm" :name="option.icon" size="30" />
      <span class="ellipsis-label ellipsis text-b-subtitle2" :class="{ 'disabled-label': option.disable }">
        {{ option.label }}
        <q-tooltip v-if="minCharsToShowOverflowTooltip && option.label.length > minCharsToShowOverflowTooltip">
          {{ option.label }}
        </q-tooltip>
      </span>
      <runai-svg-icon v-if="option.svg" class="q-ml-xs" :name="option.svg" size="14" />
      <q-tooltip
        v-if="option.disable && option.disabledTooltip"
        class="q-pa-sm"
        :max-width="option.disabledTooltipWidth"
      >
        {{ option.disabledTooltip }}
      </q-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import type { ISelectOption } from "@/models/global.model";

defineProps<{
  option: ISelectOption;
  maxWidth?: string;
  minCharsToShowOverflowTooltip?: number;
}>();
</script>

<style lang="scss" scoped>
.radio-option-row {
  height: 40px;
}

.disabled-label {
  color: var(--text-tertiary);
}
</style>
