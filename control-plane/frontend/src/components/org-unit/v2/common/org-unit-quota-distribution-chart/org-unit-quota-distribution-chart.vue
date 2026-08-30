<template>
  <stacked-bar-chart
    v-bind="$attrs"
    :segments="segments"
    :entity-name="entityName"
    :max-value="maxValue"
    :tooltip-summary="tooltipSummary"
    :height="height"
    :show-labels="showLabels"
    :show-scale="shouldShowScale"
    :label-text-border-width="2"
  />
</template>

<script lang="ts" setup>
import { computed } from "vue";
import StackedBarChart from "@/components/common/charts/stacked-bar-chart/stacked-bar-chart.vue";
import type { IStackedBarSegment } from "@/components/common/charts/stacked-bar-chart/stacked-bar-chart.model";
import { ORG_UNIT_STACKED_BAR_CHART_COLORS } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";

const props = withDefaults(
  defineProps<{
    entityName: string;
    entityType: EQuotaEntity;
    assignedQuota: number;
    parentName?: string;
    nodePoolName?: string;
    parentAvailableQuota: number;
    otherEntitiesAssignedQuota: number;
    isUnlimited?: boolean; // True when parent quota is unlimited
    isAssignedQuotaUnlimited?: boolean; // True when this entity's assigned quota is unlimited
    isUnassignedQuotaUnlimited?: boolean; // True when unassigned quota is unlimited
    isOtherEntitiesQuotaUnlimited?: boolean; // True when other entities' quota is unlimited
    height?: number;
    showLabels?: boolean;
    showScale?: boolean;
  }>(),
  {
    isUnlimited: false,
    isAssignedQuotaUnlimited: false,
    isUnassignedQuotaUnlimited: false,
    isOtherEntitiesQuotaUnlimited: false,
    height: 18,
    showLabels: true,
    showScale: true,
  },
);

// Gray color for "all zeros" empty state
const EMPTY_STATE_COLOR = "#BDBDBD";

// Check if any specific unlimited state is active
const hasUnlimitedState = computed(() => {
  return props.isAssignedQuotaUnlimited || props.isUnassignedQuotaUnlimited || props.isOtherEntitiesQuotaUnlimited;
});

// Check if all values are zero (special empty state)
// Only triggers when ALL values are truly 0 (parent quota, assigned, and other entities)
const isAllZerosState = computed(() => {
  return props.parentAvailableQuota === 0 && props.assignedQuota === 0 && props.otherEntitiesAssignedQuota === 0;
});

// Hide scale when any unlimited state is active or all zeros state (scale values are meaningless)
const shouldShowScale = computed(() => {
  if (hasUnlimitedState.value || isAllZerosState.value) return false;
  return props.showScale;
});

const unassignedQuota = computed(() => {
  if (props.isUnlimited || props.isAssignedQuotaUnlimited || props.isOtherEntitiesQuotaUnlimited) return 0; // No unassigned when unlimited
  const calculated = props.parentAvailableQuota - props.assignedQuota - props.otherEntitiesAssignedQuota;
  // Don't display negative unassigned quota
  return Math.max(0, calculated);
});

const tooltipSummary = computed(() => {
  const quotaDisplay = props.isUnlimited ? "<b>Unlimited</b>" : `<b>${props.parentAvailableQuota}</b>`;

  if (props.entityType === "project") {
    return `Department ${props.parentName} quota: ${quotaDisplay}`;
  } else {
    return `Total for node pool ${props.nodePoolName ? props.nodePoolName : ""}: ${quotaDisplay}`;
  }
});

// Calculate the sum of all segment values
const segmentsSum = computed(() => {
  return props.assignedQuota + unassignedQuota.value + props.otherEntitiesAssignedQuota;
});

// Calculate max value for chart scale
// When any specific unlimited state is active or all zeros: use 1 for full-width rendering
// When unlimited: use sum of all visible segments
// When limited: use the maximum of parent quota OR segments sum
// This prevents rendering issues when parentAvailableQuota is 0 but segments have values
const maxValue = computed(() => {
  // When any specific unlimited state or all zeros state is active, use 1 for full-width
  if (hasUnlimitedState.value || isAllZerosState.value) {
    return 1;
  }
  if (props.isUnlimited) {
    return props.assignedQuota + props.otherEntitiesAssignedQuota;
  }
  // Use the larger of parentAvailableQuota or actual segment values
  // to ensure the chart scale can accommodate all rendered segments
  return Math.max(props.parentAvailableQuota, segmentsSum.value);
});

const segments = computed<IStackedBarSegment[]>(() => {
  // Priority 0: All zeros state - show gray bar with "0"
  // Include all segments with value: 0 for tooltip display
  if (isAllZerosState.value) {
    return [
      // Main gray bar showing "0" (hidden from tooltip - just a visual placeholder)
      {
        value: 1,
        color: EMPTY_STATE_COLOR,
        displayValue: "0",
        hideFromTooltip: true,
      },
      // Include assigned quota in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
        label: "Assigned quota",
        displayValue: "0",
      },
      // Include unassigned in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5,
        label: "Unassigned quota",
        displayValue: "0",
      },
      // Include other entities in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
        label: `Sum of other ${props.entityType}s' assigned quota`,
        displayValue: "0",
      },
    ];
  }

  // Priority 1: Assigned quota unlimited - show full dark blue bar
  // Include other segments with value: 0 for tooltip display
  if (props.isAssignedQuotaUnlimited) {
    return [
      {
        value: 1,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
        label: "Assigned quota",
        displayValue: "Unlimited",
      },
      // Include unassigned in tooltip (value: 0 so it doesn't render visually)
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5,
        label: "Unassigned quota",
        displayValue: String(unassignedQuota.value),
      },
      // Include other entities in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
        label: `Sum of other ${props.entityType}s' assigned quota`,
        displayValue: String(props.otherEntitiesAssignedQuota),
      },
    ];
  }

  // Priority 2: Unassigned quota unlimited - show full medium blue bar
  // Include assigned quota segment for tooltip display
  if (props.isUnassignedQuotaUnlimited) {
    return [
      // Include assigned quota in tooltip (value: 0 so it doesn't render visually)
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
        label: "Assigned quota",
        displayValue: String(props.assignedQuota),
      },
      {
        value: 1,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5,
        label: "Unassigned quota",
        displayValue: "Unlimited",
      },
    ];
  }

  // Priority 3: Other entities quota unlimited - show full light blue bar
  // Include assigned and unassigned segments for tooltip display
  if (props.isOtherEntitiesQuotaUnlimited) {
    return [
      // Include assigned quota in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
        label: "Assigned quota",
        displayValue: String(props.assignedQuota),
      },
      // Include unassigned in tooltip
      {
        value: 0,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5,
        label: "Unassigned quota",
        displayValue: String(unassignedQuota.value),
      },
      {
        value: 1,
        color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
        label: `Sum of other ${props.entityType}s' assigned quota`,
        displayValue: "Unlimited",
      },
    ];
  }

  // Default: current numeric logic
  const segmentsList: IStackedBarSegment[] = [];

  if (
    (props.entityType === EQuotaEntity.department && props.assignedQuota > 0) ||
    props.entityType === EQuotaEntity.project
  ) {
    segmentsList.push({
      value: props.assignedQuota,
      color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9,
      label: "Assigned quota",
    });
  }

  if (!props.isUnlimited) {
    // Only show unassigned segment when limited and there's space
    segmentsList.push({
      value: unassignedQuota.value,
      color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5,
      label: "Unassigned quota",
    });
  }

  segmentsList.push({
    value: props.otherEntitiesAssignedQuota,
    color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2,
    label: `Sum of other ${props.entityType}s' assigned quota`,
  });

  return segmentsList;
});
</script>
