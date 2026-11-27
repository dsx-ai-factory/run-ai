<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-page-filter-wrapper
    display
    :title="filter.label"
    @save="saveFilter"
    @close="$emit('close')"
    :max-width="width || '250px'"
    :width="width"
  >
    <template #body>
      <div class="row">
        <div class="col">
          <q-select
            aid="free-text-filter-select"
            outlined
            v-model="selectedOption"
            :options="options"
            :disable="isOperatorSelectDisabled"
            :menu-shrink="!!operatorMenuWidth"
            :popup-content-style="operatorMenuWidth ? { width: operatorMenuWidth } : undefined"
            class="q-mb-lg"
          />
          <p class="text-black-54 filter-input-label">The following value:</p>
          <runai-autocomplete-filter-input
            v-model="term"
            aid="free-text-filter-input"
            :label="filter.label"
            :autocomplete-options-fn="filter.autocompleteOptionsFn"
          />
        </div>
      </div>
    </template>
  </runai-page-filter-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import { RunaiAutocompleteFilterInput } from "@/components/common/runai-autocomplete-filter-input";
//models
import {
  EFilterOperator,
  EFilterOperatorLabel,
  freeTextFilterSelectOptions,
  type IFreeTextFilterModel,
  type IFreeTextFilterSelectOption,
} from "@/models/filter.model";
type FilterOperatorOption = { label: string; value: EFilterOperator };

const DEFAULT_SELECTED_OPTION: IFreeTextFilterSelectOption = {
  label: EFilterOperatorLabel.Contains,
  value: EFilterOperator.Contains,
};
export default defineComponent({
  name: "runai-free-text-filter",
  components: { RunaiPageFilterWrapper, RunaiAutocompleteFilterInput },
  emits: ["save", "close"],
  props: {
    width: {
      type: String,
      required: false,
    },
    operatorMenuWidth: {
      type: String,
      required: false,
    },
    formatOperatorLabel: {
      type: Function as PropType<(label: string) => string>,
      required: false,
    },
    filter: {
      type: Object as PropType<IFreeTextFilterModel>,
      required: true,
    },
  },
  data() {
    return {
      term: "" as string,
      selectedOption: DEFAULT_SELECTED_OPTION as FilterOperatorOption,
    };
  },
  created() {
    this.initFilter();
  },
  computed: {
    options(): FilterOperatorOption[] {
      const options = freeTextFilterSelectOptions.filter(
        (option) => !this.filter.excludeOperators?.includes(option.value),
      );
      return this.formatOperatorLabel
        ? options.map((option) => ({ ...option, label: this.formatOperatorLabel?.(option.label) ?? option.label }))
        : options;
    },
    isOperatorSelectDisabled(): boolean {
      return this.options.length <= 1;
    },
  },
  methods: {
    initFilter(): void {
      this.term = this.filter.term;
      if (this.options.length === 1) {
        this.selectedOption = this.options[0];
        return;
      }
      const selectedOption = this.options.find((option) => option.value === this.filter.value);
      if (selectedOption) {
        this.selectedOption = selectedOption;
      } else if (this.filter.excludeOperators?.includes(DEFAULT_SELECTED_OPTION.value)) {
        this.selectedOption = this.options[0];
      }
    },
    saveFilter(): void {
      if (!this.term) {
        this.$emit("close");
        return;
      }
      const filter: IFreeTextFilterModel = {
        ...this.filter,
        term: this.term,
        value: this.selectedOption.value,
      };
      this.$emit("save", filter);
    },
  },
});
</script>

<style scoped lang="scss">
.filter-input-label {
  margin-bottom: 5px;
}
</style>
