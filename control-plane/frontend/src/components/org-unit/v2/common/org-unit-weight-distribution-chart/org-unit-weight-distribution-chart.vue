<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <stacked-bar-chart
    v-bind="$attrs"
    :segments="segments"
    :entity-name="entityName"
    :max-value="maxValue"
    :height="height"
    :show-labels="showLabels"
    :show-scale="shouldShowScale"
    :label-text-border-width="2"
    show-percentage
  />
</template>

<script lang="ts" setup>
import { computed } from "vue";
import StackedBarChart from "@/components/common/charts/stacked-bar-chart/stacked-bar-chart.vue";
import type { IStackedBarSegment } from "@/components/common/charts/stacked-bar-chart/stacked-bar-chart.model";
import { ORG_UNIT_STACKED_BAR_CHART_COLORS } from "@/models/org-unit-forms.model";
import type { EQuotaEntity } from "@/models/resource.model";

// Gray color for "all zeros" empty state
const EMPTY_STATE_COLOR = "#BDBDBD";

const props = withDefaults(
  defineProps<{
    entityName: string;
    entityType: EQuotaEntity;
    weight: number;
    otherEntitiesWeight: number;
    height?: number;
    showLabels?: boolean;
    showScale?: boolean;
  }>(),
  {
    height: 18,
    showLabels: true,
    showScale: true,
  },
);

// Check if all weights are zero (special empty state)
const isAllZerosState = computed(() => {
  return props.weight === 0 && props.otherEntitiesWeight === 0;
});

// Hide scale when all zeros state (scale values are meaningless)
const shouldShowScale = computed(() => {
  if (isAllZerosState.value) return false;
  return props.showScale;
});

// Calculate max value for chart scale
// Use the sum of weights, but ensure it's always positive to prevent rendering issues
// when weight values are 0 or negative
const maxValue = computed(() => {
  // When all zeros state is active, use 1 for full-width rendering
  if (isAllZerosState.value) {
    return 1;
  }
  const sum = props.weight + props.otherEntitiesWeight;
  // Ensure maxValue is at least 1 to prevent chart rendering issues
  // when sum is 0 or negative
  return Math.max(1, sum);
});

const segments = computed<IStackedBarSegment[]>(() => {
  // All zeros state - show gray bar with "0"
  if (isAllZerosState.value) {
    return [
      // Main gray bar showing "0" (hidden from tooltip - just a visual placeholder)
      {
        value: 1,
        color: EMPTY_STATE_COLOR,
        displayValue: "0",
        hideFromTooltip: true,
      },
      // Include weight in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
        label: "Weight",
        displayValue: "0",
      },
      // Include other entities weight in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
        label: `Total weight of other ${props.entityType}s`,
        displayValue: "0",
      },
    ];
  }

  // Normal state
  return [
    {
      value: props.weight,
      color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
      label: "Weight",
    },
    {
      value: props.otherEntitiesWeight,
      color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
      label: `Total weight of other ${props.entityType}s`,
    },
  ];
});
</script>
