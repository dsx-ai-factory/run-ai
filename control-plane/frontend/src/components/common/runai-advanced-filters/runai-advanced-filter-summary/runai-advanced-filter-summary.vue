<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <!-- RunaiFilterChip owns an editor popup; summary chips need independent edit and remove controls. -->
  <q-chip
    class="filter-chip"
    v-for="filter in filters"
    :key="filter.name"
    removable
    size="md"
    @remove="$emit('remove', filter.name)"
  >
    <button
      type="button"
      class="filter-chip-edit"
      :aria-label="`Edit ${filter.label} filter: ${getFilterSummary(filter)}`"
      @click="$emit('select', filter)"
    >
      {{ filter.label }}:
      <span class="q-pr-md q-pl-xs text-weight-medium">{{ getFilterSummary(filter) }}</span>
    </button>
  </q-chip>
</template>
<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import {
  EColumnFilterType,
  EFilterOperator,
  filterOperatorMap,
  type IAdvancedFilterModel,
  type IDateFilterModel,
  type IEnumBasedFilterModel,
  type IFreeTextFilterModel,
  type INumericFilterModel,
  numericFilterSelectOptions,
} from "@/models/filter.model";
import { previewStrings } from "@/utils/common.util";

export default defineComponent({
  name: "runai-advanced-filter-summary",
  emits: ["remove", "select"],
  props: {
    filters: {
      type: Array as PropType<Array<IAdvancedFilterModel>>,
      required: true,
    },
  },
  methods: {
    getFilterSummary(filter: IAdvancedFilterModel): string {
      switch (filter.type) {
        case EColumnFilterType.FreeText:
          return this.getFreeTextFilterSummary(filter as IFreeTextFilterModel);
        case EColumnFilterType.EnumBased:
          return this.getEnumBasedFilterSummary(filter as IEnumBasedFilterModel);
        case EColumnFilterType.Date:
          return this.getDateFilterSummary(filter as IDateFilterModel);
        case EColumnFilterType.Numeric:
          return this.getNumericFilterSummary(filter as INumericFilterModel);
        default:
          throw new Error(`Unsupported filter type: ${filter.type}`);
      }
    },
    getFreeTextFilterSummary(filter: IFreeTextFilterModel): string {
      if (!filter.value) return "";

      return `${filterOperatorMap[filter.value]} ${filter.term}`;
    },
    getEnumBasedFilterSummary(filter: IEnumBasedFilterModel): string {
      if (filter.inputType === "radio") {
        return filter.selectOptions.find((option) => option.value === filter.selectedValues[0])?.label || "";
      }
      const selectedValuesSet = new Set<string | boolean>(filter.selectedValues);
      const selectedLabels: string[] = filter.selectOptions
        .filter((option) => selectedValuesSet.has(option.value as string) && !option.hidden)
        .map((option) => option.label as string);
      return previewStrings(selectedLabels, 2);
    },
    getDateFilterSummary(filter: IDateFilterModel): string {
      return `${filter.value} ${filter.date}`;
    },
    getNumericFilterSummary(filter: INumericFilterModel): string {
      if (filter.operator === EFilterOperator.Range) {
        return `${EFilterOperator.Range} ${filter.minValue} - ${filter.maxValue}`;
      }

      const operator = numericFilterSelectOptions.find((option) => option.value === filter.operator)?.label;
      return `${operator} ${filter.minValue}`;
    },
  },
});
</script>

<style scoped lang="scss">
.filter-chip-edit {
  display: inline-flex;
  align-items: center;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  cursor: pointer;
}
</style>
