<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-page-filter-wrapper
    :is-apply-button-disabled="!isValidInput"
    display
    :title="filter.label"
    @save="saveFilter"
    @close="$emit('close')"
    width="410px"
  >
    <template #body v-if="selectedFilter">
      <div class="row">
        <div class="col">
          <div class="row">
            <q-select
              aid="numeric-filter-select"
              outlined
              v-model="selectedFilter.operator"
              :options="options"
              class="q-mb-lg numeric-filter-input q-mr-md"
            />
            <q-input
              v-if="selectedFilter.operator.value !== EFilterOperator.Range"
              aid="numeric-filter-input-min"
              outlined
              v-model.number="selectedFilter.minValue"
              type="number"
              class="numeric-filter-input"
            />
          </div>
          <div class="row" v-if="selectedFilter.operator.value === EFilterOperator.Range">
            <div class="row items-center">
              <q-input
                class="numeric-filter-input"
                aid="numeric-filter-input-min"
                outlined
                v-model.number="selectedFilter.minValue"
                type="number"
              />
              <div class="horizontal-line q-mx-sm"></div>
              <q-input
                class="numeric-filter-input"
                aid="numeric-filter-input-max"
                outlined
                v-model.number="selectedFilter.maxValue"
                type="number"
              />
            </div>
          </div>
        </div>
      </div>
    </template>
  </runai-page-filter-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import { EFilterOperator, type INumericFilterModel, numericFilterSelectOptions } from "@/models/filter.model";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "runai-numeric-filter",
  components: { RunaiPageFilterWrapper },
  emits: ["save", "close"],
  props: {
    filter: {
      type: Object as PropType<INumericFilterModel>,
      required: true,
    },
  },
  data() {
    return {
      options: numericFilterSelectOptions as ISelectOption[],
      selectedFilter: null as null | { minValue: number | string; maxValue: number | string; operator: ISelectOption },
    };
  },
  created() {
    this.initFilter();
  },
  computed: {
    EFilterOperator() {
      return EFilterOperator;
    },
    isValidInput(): boolean {
      if (this.selectedFilter?.operator.value === EFilterOperator.Range) {
        return this.selectedFilter?.minValue !== "" && this.selectedFilter?.maxValue !== "";
      }
      return this.selectedFilter?.minValue !== "";
    },
  },
  methods: {
    initFilter(): void {
      this.selectedFilter = {
        ...this.filter,
        operator: numericFilterSelectOptions.find(
          (option: ISelectOption) => option.value === this.filter.operator,
        ) as ISelectOption,
      };
    },
    saveFilter(): void {
      if (!this.selectedFilter) return;
      this.$emit("save", { ...this.selectedFilter, operator: this.selectedFilter.operator.value });
    },
  },
});
</script>

<style scoped lang="scss">
.numeric-filter-input {
  width: 112px;
}
.horizontal-line {
  border-top: 1px solid black;
  width: 8px;
}
</style>
