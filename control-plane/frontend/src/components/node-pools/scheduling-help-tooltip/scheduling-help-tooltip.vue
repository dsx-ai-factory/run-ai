<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-tooltip
    :tooltip-text="tooltipText"
    :icon="icon"
    :width="width"
    :interactive="interactive"
    :popup-class="['scheduling-help-popup', popupClass].filter(Boolean).join(' ')"
    :anchor="anchor"
    :self="self"
    :offset="offset"
  >
    <template v-if="$slots.content" #content><slot name="content" /></template>
  </runai-tooltip>
</template>
<script setup lang="ts">
import type { QTooltipProps } from "quasar";
import { RunaiTooltip } from "@/components/common/runai-tooltip/index";
withDefaults(
  defineProps<{
    tooltipText: string;
    icon?: string;
    width?: string;
    interactive?: boolean;
    popupClass?: string;
    anchor?: QTooltipProps["anchor"];
    self?: QTooltipProps["self"];
    offset?: number[];
  }>(),
  { icon: "ph-bold ph-question", width: "300px", anchor: "center right", self: "center left" },
);
</script>
<style lang="scss">
.scheduling-help-popup {
  white-space: normal;
  padding: 10px;
  border-radius: 3px;
  background: var(--bg-primary-default);
  color: var(--b-text-main);
  box-shadow: var(--shadow-s);
  font-family: Roboto, sans-serif;
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
}
</style>
