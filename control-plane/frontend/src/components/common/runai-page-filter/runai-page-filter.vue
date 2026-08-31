<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-page-filter">
    <page-cluster-filter
      v-if="!hideClusterFilter"
      :filter-value="clusterFilterTerm"
      :hide-all-option="hideAllClusterFilterOption"
      @update-filter="updateClusterFilter"
    />
    <q-chip
      v-for="filter in currentFilters"
      :aid="getFilterChipAid(filter)"
      :key="filter.name"
      clickable
      removable
      size="md"
      @remove="removeFilter(filter.name)"
      >{{ filter.label }}:
      <span class="q-pr-md q-pl-xs text-weight-medium"> {{ filter.term }}</span>
      <runai-page-filter-popup
        :title="filter.label"
        :filter="filter"
        @save="saveFilter"
        :input-label="filterInputLabel"
        :autocomplete-options-fn="getAutocompleteOptionsFn(filter.name)"
      ></runai-page-filter-popup>
    </q-chip>
    <q-btn class="add-filter-btn q-ml-xs" ref="add-filter-btn" flat no-caps label="Add filter">
      <q-menu fit transition-show="scale" transition-hide="scale" transition-duration="500">
        <q-list>
          <q-item
            v-for="option in availableFilterOptions"
            :key="option.label"
            clickable
            v-close-popup
            class="cursor-pointer"
            @click="selectFilterToAdd(option)"
          >
            <q-item-section>{{ option.label }} </q-item-section>
          </q-item>
        </q-list>
      </q-menu>
      <runai-page-filter-popup
        :title="selectedFilterLabel"
        :display="!!newFilter.name"
        v-if="!useSelection"
        :filter="newFilter"
        @close="resetNewFilter"
        @save="saveFilter"
        :input-label="filterInputLabel"
        :autocomplete-options-fn="getAutocompleteOptionsFn(newFilter.name)"
      ></runai-page-filter-popup>
      <runai-page-filter-popup-with-select
        :title="selectedFilterLabel"
        :display="!!newFilter.name"
        v-if="useSelection"
        :filter="newFilter"
        @close="resetNewFilter"
        @save="saveFilter"
        :input-label="filterInputLabel"
        :async-options-promise-map="asyncOptionsPromiseMap"
      ></runai-page-filter-popup-with-select>
    </q-btn>
  </section>
</template>

<script lang="ts">
//vue
import { defineComponent, type PropType } from "vue";

// models
import type { ISelectOption } from "@/models/global.model";
import {
  type IFilterOption,
  type IFilterModel,
  type ILabeledFilter,
  type IFilterBy,
  CLUSTER_COLUMN_FILTER_NAME,
} from "@/models/filter.model";

// services
import { filterService } from "@/services/filter.service/filter.service";

// utils
import { deepCopy } from "@/utils/common.util";

// cmos
import RunaiPageFilterPopup from "./runai-page-filter-popup.vue";
import RunaiPageFilterPopupWithSelect from "@/components/common/runai-page-filter/runai-page-filter-popup-with-select.vue";
import { PageClusterFilter } from "@/components/common/page-cluster-filter";
import { stringUtil } from "@/utils/string.util";
const EMPTY_FILTER: IFilterOption = { name: "emptyFilter", field: "", label: "No filters to add" };
export enum EEmitName {
  UpdateFilterBy = "update-filter-by",
}
export default defineComponent({
  components: {
    RunaiPageFilterPopupWithSelect,
    RunaiPageFilterPopup,
    PageClusterFilter,
  },
  emits: {
    [EEmitName.UpdateFilterBy]: (payload: IFilterBy) => payload,
  },
  props: {
    initialFiltersModel: {
      type: Array as PropType<Array<IFilterModel>>,
      required: true,
    },
    filterOptions: {
      type: Array as PropType<Array<IFilterOption>>,
      required: true,
    },
    filterInputLabel: {
      type: String as PropType<string>,
      required: false,
    },
    useSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    asyncOptionsPromiseMap: {
      type: Object as PropType<Record<string, (searchQuery: string) => Promise<ISelectOption[]>>>,
      required: false,
      default: () => ({}),
    },
    hideAllClusterFilterOption: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: false,
    },
    hideClusterFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      newFilter: { field: "", term: "", name: "" } as IFilterModel,
      optionsObj: new Map<string, IFilterOption>(),
      filtersModel: [] as Array<IFilterModel>,
    };
  },
  created() {
    this.initFilterOptions();
    this.filtersModel = deepCopy(this.initialFiltersModel);
  },
  computed: {
    clusterFilterTerm(): string {
      return this.filterBy?.columnFilters?.find((c) => c.name === CLUSTER_COLUMN_FILTER_NAME)?.term || "";
    },
    selectedFilterLabel(): string {
      return this.optionsObj.get(this.newFilter.name)?.label || "";
    },
    currentFilters(): Array<ILabeledFilter> {
      return this.filtersModel
        .filter((fm: IFilterModel) => this.optionsObj.get(fm.name)?.label)
        .map((fm: IFilterModel) => {
          const label: string = this.optionsObj.get(fm.name)?.label || "";
          return { ...fm, label };
        })
        .filter((fm: ILabeledFilter) => !!fm.label);
    },
    availableFilterOptions(): Array<IFilterOption> {
      const options: IFilterOption[] = filterService.getAvailableFilterOptions(
        this.filtersModel,
        this.filterOptions.filter((f) => f.label),
      );
      if (options.length > 0) {
        return options;
      }
      return [{ ...EMPTY_FILTER }];
    },
  },
  methods: {
    removeFilter(filterName: string): void {
      const filterIdx = this.filtersModel.findIndex((fm: IFilterModel): boolean => fm.name === filterName);
      this.filtersModel.splice(filterIdx, 1);
      this.saveFilters();
    },
    selectFilterToAdd(filterOption: IFilterOption): void {
      if (filterOption.name === EMPTY_FILTER.name) return;
      this.newFilter.name = filterOption.name;
      this.newFilter.field = filterOption.field;
    },
    saveFilter(filter: IFilterModel): void {
      const filterIndex = this.filtersModel.findIndex((f: IFilterModel) => f.name === filter.name);
      if (filterIndex !== -1) {
        this.filtersModel.splice(filterIndex, 1, filter);
      } else {
        this.filtersModel.push(filter);
      }
      this.saveFilters();
    },
    resetNewFilter(): void {
      this.newFilter = { field: "", term: "", name: "" };
    },
    saveFilters(): void {
      this.updateFilterBy({ ...this.filterBy, columnFilters: this.filtersModel });
      this.resetNewFilter();
    },
    initFilterOptions(): void {
      this.filterOptions.reduce((acc: Map<string, IFilterOption>, item: IFilterOption) => {
        acc.set(item.name, item);
        return acc;
      }, this.optionsObj);
    },
    updateClusterFilter(clusterUuid: string): void {
      const updatedFilter: IFilterBy = {
        ...this.filterBy,
        clusterUuid,
        columnFilters: (this.filterBy?.columnFilters || []).map((c) => {
          return c.name === CLUSTER_COLUMN_FILTER_NAME ? { ...c, term: clusterUuid } : c;
        }),
      };

      this.updateFilterBy(updatedFilter);
    },
    getAutocompleteOptionsFn(filterName: string): ((searchQuery: string) => Promise<string[]>) | undefined {
      const originalFn = this.optionsObj.get(filterName)?.autocompleteOptionsFn;
      if (!originalFn) return undefined;
      return (searchQuery: string) => originalFn(searchQuery, this.filterBy?.clusterUuid);
    },
    getFilterChipAid(filter: ILabeledFilter): string {
      return `filter-chip-${stringUtil.slug(filter.name)}`;
    },
    updateFilterBy(filterBy: IFilterBy): void {
      this.$emit(EEmitName.UpdateFilterBy, filterBy);
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
