<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-table-tooltip">
    <div v-if="heading" class="runai-table-tooltip__heading">
      <p class="text-b-body1">{{ heading }}</p>
    </div>
    <table class="runai-table-tooltip__table">
      <thead>
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            class="runai-table-tooltip__cell runai-table-tooltip__cell--header text-b-body1-md"
            :style="column.width ? { width: column.width } : {}"
          >
            {{ column.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="loading">
          <td :colspan="columns.length" class="runai-table-tooltip__cell runai-table-tooltip__cell--loading">
            <q-spinner size="20px" />
          </td>
        </tr>
        <tr v-for="(row, rowIndex) in rows" v-else :key="rowIndex">
          <td v-for="column in columns" :key="column.key" class="runai-table-tooltip__cell text-b-body1">
            <slot :name="`cell-${column.key}`" :row="row" :column="column">{{ row[column.key] }}</slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { QSpinner } from "quasar";

export interface IRunaiTableTooltipColumn {
  key: string;
  label: string;
  width?: string;
}

export interface Props {
  heading?: string;
  columns: IRunaiTableTooltipColumn[];
  rows: Array<Record<string, string>>;
  loading?: boolean;
}

defineProps({
  heading: {
    type: String as PropType<string>,
    required: false,
    default: "",
  },
  columns: {
    type: Array as PropType<IRunaiTableTooltipColumn[]>,
    required: true,
  },
  rows: {
    type: Array as PropType<Array<Record<string, string>>>,
    required: true,
  },
  loading: {
    type: Boolean as PropType<boolean>,
    required: false,
    default: false,
  },
});
</script>

<style lang="scss" scoped>
.runai-table-tooltip {
  display: flex;
  flex-direction: column;
  background: var(--bg-primary-default);
  color: var(--b-text-main);
  border-radius: var(--space-3xs);
  box-shadow: var(--shadow-s);
  overflow: hidden;
}

.runai-table-tooltip__heading {
  border-bottom: 1px solid var(--border-subtle);
  padding: var(--space-m) var(--space-s);

  p {
    margin: 0;
  }
}

.runai-table-tooltip__table {
  width: 100%;
  border-collapse: collapse;
  background: var(--table-bg);
}

.runai-table-tooltip__cell {
  padding: var(--space-s);
  border-bottom: 1px solid var(--border-subtle);
  text-align: start;
  vertical-align: middle;
  height: 50px;
}

.runai-table-tooltip__cell--header {
  font-weight: 500;
}

.runai-table-tooltip__cell--loading {
  text-align: center;
}
</style>
