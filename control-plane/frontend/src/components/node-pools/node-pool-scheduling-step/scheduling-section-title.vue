<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-section-title :flat="flat">
    <template #title-content>
      <span class="q-mr-xs">{{ title }}</span>
      <scheduling-help-tooltip
        :tooltip-text="tooltipText"
        :icon="tooltipIcon"
        :width="width"
        :interactive="interactiveTooltip"
      >
        <template v-if="$slots['tooltip-content']" #content><slot name="tooltip-content" /></template>
      </scheduling-help-tooltip>
    </template>
  </runai-section-title>
</template>
<script setup lang="ts">
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import SchedulingHelpTooltip from "./scheduling-help-tooltip.vue";
withDefaults(
  defineProps<{
    title: string;
    tooltipText: string;
    tooltipIcon?: string;
    width?: string;
    flat?: boolean;
    interactiveTooltip?: boolean;
  }>(),
  { tooltipIcon: "ph-bold ph-question", width: "300px" },
);
</script>
