<template>
  <div class="runai-data-tooltip">
    <p v-if="heading" class="runai-data-tooltip__heading text-b-body1">{{ heading }}</p>
    <div class="runai-data-tooltip__legend">
      <div v-for="(row, index) in rows" :key="index" class="runai-data-tooltip__row">
        <span
          v-if="row.dotColor"
          class="runai-data-tooltip__dot"
          :style="{ backgroundColor: row.dotColor }"
          aria-hidden="true"
        />
        <span class="runai-data-tooltip__row-text">
          <span class="text-b-body1">{{ row.label }}</span>
          <span class="text-b-body1-bold">{{ row.value }}</span>
        </span>
      </div>
    </div>
    <div v-if="summary" class="runai-data-tooltip__summary">
      <span class="text-b-body1">{{ summary.label }}</span>
      <span class="text-b-body1-bold">{{ summary.value }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";

export interface IRunaiDataTooltipRow {
  label: string;
  value: string | number;
  dotColor?: string;
}

export interface IRunaiDataTooltipSummary {
  label: string;
  value: string | number;
}

export interface Props {
  heading?: string;
  rows: IRunaiDataTooltipRow[];
  summary?: IRunaiDataTooltipSummary;
}

defineProps({
  heading: {
    type: String as PropType<string>,
    required: false,
    default: "",
  },
  rows: {
    type: Array as PropType<IRunaiDataTooltipRow[]>,
    required: true,
  },
  summary: {
    type: Object as PropType<IRunaiDataTooltipSummary>,
    required: false,
    default: undefined,
  },
});
</script>

<style lang="scss" scoped>
.runai-data-tooltip {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-s);
  background: var(--bg-inverse-default);
  border-radius: var(--space-3xs);
  box-shadow: var(--shadow-s);

  .text-b-body1,
  .text-b-body1-bold {
    color: var(--text-inverse);
  }
}

.runai-data-tooltip__heading {
  margin: 0;
}

.runai-data-tooltip__legend {
  display: flex;
  flex-direction: column;
  gap: var(--space-3xs);
}

.runai-data-tooltip__row {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
}

.runai-data-tooltip__dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.runai-data-tooltip__row-text {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2xs);
  white-space: nowrap;
}

.runai-data-tooltip__summary {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2xs);
  padding-left: 13px;
  white-space: nowrap;
}
</style>
