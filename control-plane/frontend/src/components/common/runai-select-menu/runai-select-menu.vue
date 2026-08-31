<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu
    class="runai-select-menu"
    anchor="bottom middle"
    self="top middle"
    @before-show="emit('before-show')"
    @hide="emit('hide')"
  >
    <q-list role="none" class="column">
      <template v-if="loading">
        <q-item v-for="i in 3" :key="i">
          <q-skeleton :width="skeletonDimensions.width" :height="skeletonDimensions.height" />
        </q-item>
      </template>

      <q-item
        v-else
        role="menuitem"
        clickable
        v-close-popup
        v-for="opt in list"
        :key="opt.value?.toString()"
        @click="emit('selected-value', opt.value?.toString() || '')"
        :aid="opt.aid || opt.value + '-btn'"
        :disable="opt.disable"
        class="row justify-between items-center"
      >
        <span class="q-mr-sm">{{ opt.label }}</span>

        <!-- Icon tooltip (question mark icon with tooltip) -->
        <runai-tooltip
          v-if="opt.warningTooltip"
          class="q-ml-sm"
          :tooltip-text="opt.warningTooltip"
          :width="tooltipWidth"
          :ripple="false"
        />

        <!-- Regular hover tooltip on the option -->
        <q-tooltip v-if="opt.toolTip" class="q-pa-sm runai-tooltip" anchor="center right" self="center left">
          {{ opt.toolTip }}
        </q-tooltip>
      </q-item>
    </q-list>
  </q-menu>
</template>
<script setup lang="ts">
import type { ISelectOption } from "@/models/global.model";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export interface Props {
  list: ISelectOption[];
  loading?: boolean;
  skeletonDimensions?: {
    width: string;
    height: string;
  };
  tooltipWidth?: string;
}

withDefaults(defineProps<Props>(), {
  list: () => [],
  loading: false,
  skeletonDimensions: () => ({
    width: "133px",
    height: "36px",
  }),
  tooltipWidth: "250px",
});

const emit = defineEmits<{
  (e: "selected-value", value: string): void;
  (e: "before-show"): void;
  (e: "hide"): void;
}>();
</script>
