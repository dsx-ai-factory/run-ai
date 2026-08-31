<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-page-filter advanced-filters">
    <page-cluster-filter
      v-if="!hideClusterFilter"
      :filter-value="clusterFilter?.term || ''"
      :hide-all-option="hideAllClusterFilterOption"
      :apply-globally="applyClusterFilterGlobally"
      @update-filter="updateClusterFilter"
    />
    <runai-advanced-filter-summary :filters="filtersExcludingCluster" @remove="removeFilter" @select="selectFilter" />
    <q-btn class="add-filter-btn q-ml-xs" ref="add-filter-btn" flat no-caps label="Add filter">
      <q-menu fit transition-show="scale" transition-hide="scale" transition-duration="500">
        <q-list>
          <q-item
            v-for="option in availableFilterOptions"
            :key="option.label"
            clickable
            v-close-popup
            role="menuitem"
            class="cursor-pointer add-filter-item"
            @click="selectFilterToAdd(option)"
          >
            <q-item-section>{{ option.label }} </q-item-section>
          </q-item>
        </q-list>
      </q-menu>
      <component
        v-if="selectedFilter"
        @close="selectFilter(null)"
        @save="saveFilter"
        :filter="selectedFilter"
        :is="filterComponent(selectedFilter.name)"
      />
    </q-btn>
  </section>
</template>

<script lang="ts">
//vue
import { defineComponent, type Component, type PropType } from "vue";

// models
import {
  CLUSTER_COLUMN_FILTER_NAME,
  EColumnFilterType,
  type IAdvancedFilterModel,
  type IAdvancedFilterOption,
  type IFilterBy,
  type IFilterModel,
  type IFilterOption,
  type IFreeTextFilterModel,
} from "@/models/filter.model";

// services
import { filterService } from "@/services/filter.service/filter.service";

// utils
import { deepCopy } from "@/utils/common.util";

// cmps
import { RunaiPageFilterWrapper } from "./runai-page-filter-wrapper";
import RunaiPageFilterPopupWithSelect from "@/components/common/runai-page-filter/runai-page-filter-popup-with-select.vue";
import { RunaiFreeTextFilter } from "@/components/common/runai-advanced-filters/runai-free-text-filter";
import { RunaiEnumBasedFilter } from "@/components/common/runai-advanced-filters/runai-enum-based-filter";
import RunaiPageFilterPopup from "@/components/common/runai-page-filter/runai-page-filter-popup.vue";
import { RunaiAdvancedFilterSummary } from "@/components/common/runai-advanced-filters/runai-advanced-filter-summary";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { RunaiDateFilter } from "@/components/common/runai-advanced-filters/runai-date-filter";
import { RunaiNumericFilter } from "@/components/common/runai-advanced-filters/runai-numeric-filter";
import { PageClusterFilter } from "@/components/common/page-cluster-filter";

export default defineComponent({
  components: {
    PageClusterFilter,
    RunaiAdvancedFilterSummary,
    RunaiPageFilterWrapper,
    RunaiPageFilterPopup,
    RunaiPageFilterPopupWithSelect,
    RunaiFreeTextFilter,
    RunaiEnumBasedFilter,
  },
  emits: ["update-filter-by"],
  props: {
    initialFiltersModel: {
      type: Array as PropType<Array<IAdvancedFilterModel>>,
      default: () => [],
    },
    filterOptions: {
      type: Array as PropType<Array<IAdvancedFilterOption>>,
      required: true,
    },
    hideAllClusterFilterOption: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    hideClusterFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    applyClusterFilterGlobally: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      selectedFilter: null as IAdvancedFilterModel | null,
      optionsObj: new Map<string, IFilterOption>(),
      filtersModel: [] as Array<IAdvancedFilterModel>,
    };
  },
  created() {
    this.initFilterOptions();
    this.initFilterModel();
  },
  computed: {
    filtersExcludingCluster(): Array<IAdvancedFilterModel> {
      return this.filtersModel.filter((f: IAdvancedFilterModel) => f.name !== CLUSTER_COLUMN_FILTER_NAME);
    },
    availableFilterOptions(): Array<IAdvancedFilterOption> {
      return filterService.getAvailableFilterOptions(
        this.filtersModel,
        this.filterOptions.filter((f) => f.type),
      );
    },
    clusterFilter(): IFilterModel {
      return {
        field: "",
        name: CLUSTER_COLUMN_FILTER_NAME,
        term: this.filterBy.clusterUuid || "",
      };
    },
  },
  methods: {
    updateFilterBy(filterBy: IFilterBy): void {
      this.$emit("update-filter-by", filterBy);
    },
    removeFilter(filterName: string): void {
      this.filtersModel = this.filtersModel.filter(
        (filter: IAdvancedFilterModel): boolean => filter.name !== filterName,
      );
      this.saveFilters();
    },
    selectFilterToAdd(filterOption: IAdvancedFilterOption): void {
      const defaultFilter = filterUtil.getAdvancedFilterModelByType(filterOption.type);
      this.selectFilter({ ...defaultFilter, ...filterOption } as IAdvancedFilterModel);
    },
    saveFilter(filter: IAdvancedFilterModel): void {
      const filterIndex = this.filtersModel.findIndex((f: IAdvancedFilterModel) => f.name === filter.name);
      if (filterIndex !== -1) {
        this.filtersModel.splice(filterIndex, 1, filter);
      } else {
        this.filtersModel.push(filter);
      }
      this.saveFilters();
    },
    saveFilters(): void {
      this.updateFilterBy({ ...this.filterBy, advancedFilters: this.filtersModel });
      this.selectFilter(null);
    },
    initFilterModel(): void {
      this.filtersModel = deepCopy(this.initialFiltersModel);
    },
    initFilterOptions(): void {
      this.filterOptions.reduce((acc: Map<string, IFilterOption>, item: IFilterOption) => {
        acc.set(item.name, item);
        return acc;
      }, this.optionsObj);
    },
    filterComponent(filterName: string): Component {
      const currentFilter = this.filterOptions.find((option: IFilterOption) => option.name === filterName);

      switch (currentFilter?.type) {
        case EColumnFilterType.FreeText:
          return RunaiFreeTextFilter;
        case EColumnFilterType.EnumBased:
          return RunaiEnumBasedFilter;
        case EColumnFilterType.Numeric:
          return RunaiNumericFilter;
        case EColumnFilterType.Date:
          return RunaiDateFilter;
        default:
          throw new Error(`Unknown filter type: ${filterName}`);
      }
    },
    selectFilter(filter: IAdvancedFilterModel | null): void {
      if (filter) {
        const matchingOption = this.filterOptions.find((option: IAdvancedFilterOption) => option.name === filter.name);
        if (matchingOption?.autocompleteOptionsFn) {
          const originalFn = matchingOption.autocompleteOptionsFn;
          (filter as IFreeTextFilterModel).autocompleteOptionsFn = (searchQuery: string) =>
            originalFn(searchQuery, this.filterBy.clusterUuid);
        }
      }
      this.selectedFilter = filter;
    },
    updateClusterFilter(clusterUuid: string): void {
      this.updateFilterBy({ ...this.filterBy, clusterUuid });
    },
  },
  watch: {
    filterOptions: {
      handler() {
        this.initFilterOptions();
      },
      deep: true,
    },
    initialFiltersModel: {
      handler(newVal) {
        this.filtersModel = newVal;
      },
    },
  },
});
</script>

<style lang="scss">
.runai-page-filter {
  .filter-value {
    font-weight: 500;
  }

  .add-filter-btn {
    color: $black-87;
  }
}
</style>
