<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span>{{ colLabel }}</span>
  <runai-filter-chip
    @click.stop
    :value="chipValue"
    :icon-name="config.iconName"
    @save="onFilterChanged"
    @close="selectedValue = filterByValue"
    size="sm"
    chip-value-text-weight="regular"
    menu-title="Show the last"
  >
    <template v-slot:input-component>
      <runai-radio-options v-model="selectedValue" :options="config.options" />
    </template>
  </runai-filter-chip>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type {
  ICustomHeaderRadioButtonsConfig,
  ICustomHeaderRadioButtonsSelectOption,
  ITableColumn,
} from "@/models/table.model";
import type { IFilterBy } from "@/models/filter.model";
import RunaiFilterChip from "@/components/common/runai-filter-chip/runai-filter-chip.vue";
import RunaiRadioOptions from "@/components/common/runai-radio-options/runai-radio-options.vue";
export default defineComponent({
  name: "radio-buttons-table-header",
  components: { RunaiRadioOptions, RunaiFilterChip },
  emits: ["custom-header-event"],
  props: {
    col: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
  },
  data() {
    return {
      selectedValue: "" as string,
      filterByValue: "" as string,
    };
  },
  created() {
    this.syncFromFilterBy(this.filterBy);
  },
  computed: {
    config(): ICustomHeaderRadioButtonsConfig {
      return this.col.customHeaderConfig as ICustomHeaderRadioButtonsConfig;
    },
    colLabel(): string {
      return this.col.label;
    },
    chipValue(): string {
      const selectedOption = this.findOptionByValue(this.filterByValue);
      const value = selectedOption?.shortLabel || "";
      return `${this.config.chipPrefix} ${value}`;
    },
  },
  methods: {
    onFilterChanged(): void {
      this.filterByValue = this.selectedValue;
      this.$emit("custom-header-event", this.selectedValue);
    },
    findOptionByValue(value: string): ICustomHeaderRadioButtonsSelectOption | undefined {
      return this.config.options.find((option) => option.value === value);
    },
    syncFromFilterBy(filterBy: IFilterBy): void {
      this.filterByValue = filterBy[this.config.selectedValueFilterByKey as keyof IFilterBy] as string;
      const matchingOption = this.findOptionByValue(this.filterByValue);
      this.selectedValue = (matchingOption?.value as string | undefined) || "";
    },
  },
  watch: {
    filterBy: {
      handler(newVal) {
        this.syncFromFilterBy(newVal);
      },
      deep: true,
    },
  },
});
</script>

<style scoped lang="scss"></style>
