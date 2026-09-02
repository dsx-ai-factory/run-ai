<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="dashboard-page-filter q-px-md">
    <div class="filters-container row q-mr-auto">
      <runai-page-filter
        v-if="!hideFilter"
        hide-all-cluster-filter-option
        :filter-options="filterOptions"
        :initial-filters-model="columnFilters"
        :filter-by="filterBy"
        @update-filter-by="onFilterChange"
        use-selection
        :async-options-promise-map="asyncFilterOptionsPromiseMap"
      />
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiPageFilter } from "@/components/common/runai-page-filter";
import { type IFilterBy, type IFilterModel, type IFilterOption } from "@/models/filter.model";
import { extendedHoursPredefinedOptions, type IRangeDates } from "@/utils/date.util";
import type { PredefinedOption } from "@/models/date.model";
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  name: "dashboard-page-filter",
  components: { RunaiPageFilter },
  emits: ["filter-changed", "date-changed"],
  props: {
    filterOptions: {
      type: Array as PropType<IFilterOption[]>,
      default: () => [],
    },
    asyncFilterOptionsPromiseMap: {
      type: Object as PropType<Record<string, (searchQuery: string) => Promise<Array<ISelectOption>>>>,
      // eslint-disable-next-line @typescript-eslint/no-empty-function
      default: () => {},
    },
    hideFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
  },
  predefinedOptions: extendedHoursPredefinedOptions as Array<PredefinedOption>,
  computed: {
    columnFilters(): Array<IFilterModel> {
      return this.filterBy.columnFilters || ([] as Array<IFilterModel>);
    },
  },
  methods: {
    onFilterChange(filters: Array<IFilterModel>) {
      this.$emit("filter-changed", filters);
    },
  },
});
</script>

<style scoped lang="scss">
.dashboard-page-filter {
  background: $white;
  box-shadow: 0 2px 4px $black-15;
  height: 60px;
  display: flex;
  align-items: center;
}
</style>
