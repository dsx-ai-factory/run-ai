<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu
    class="page-filter-popup"
    transition-show="scale"
    transition-hide="scale"
    transition-duration="500"
    :model-value="display"
    @hide="emitClose()"
  >
    <form @submit.prevent>
      <q-card-section class="row justify-between items-center q-py-sm">
        <div class="text-weight-medium">{{ title }}</div>
        <q-btn
          flat
          icon="ph-bold ph-x"
          class="q-pa-xs runai-icon-btn--default text-icon-default"
          round
          v-close-popup
          @click="emitClose"
          aria-label="Close filter"
        ></q-btn>
      </q-card-section>
      <q-card-section class="q-py-sm">
        <runai-autocomplete-select
          aid="tag-autocomplete"
          :no-result-label="noResultLabel"
          :label="label"
          :search-hint="searchHint"
          :model-value="inputValue"
          @update:model-value="onUpdate"
          :async-options-promise="asyncOptionsPromiseMap[filter.field as keyof typeof asyncOptionsPromiseMap]"
          :minimum-characters="0"
          :debounce="0"
          has-outline
        />
      </q-card-section>
      <q-card-section class="row justify-end q-py-sm">
        <q-btn
          type="submit"
          flat
          color="primary"
          label="Apply"
          v-close-popup
          @click="saveFilter"
          :disabled="isDisabled"
        />
      </q-card-section>
    </form>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IFilterModel } from "@/models/filter.model";
import RunaiAutocompleteSelect from "@/components/common/runai-autocomplete-select/runai-autocomplete-select.vue";
import type { ISelectOption } from "@/models/global.model";
import { stringUtil } from "@/utils/string.util";

export default defineComponent({
  components: { RunaiAutocompleteSelect },
  emits: ["save", "close", "update"],
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
    display: {
      type: Boolean as PropType<boolean>,
    },
    filter: {
      type: Object as PropType<IFilterModel>,
      required: true,
    },
    asyncOptionsPromiseMap: {
      type: Object as PropType<Record<string, (searchQuery: string) => Promise<ISelectOption[]>>>,
      required: true,
    },
  },
  data() {
    return {
      inputValue: this.filter.term as string,
    };
  },
  computed: {
    noResultLabel(): string {
      return `No ${stringUtil.toLowerCase(this.filter.name)} match your filter`;
    },
    label(): string {
      return `Select a ${stringUtil.toLowerCase(this.filter.name)} to filter by`;
    },
    searchHint(): string {
      return `Search a ${stringUtil.toLowerCase(this.filter.name)}`;
    },
    isDisabled(): boolean {
      return !this.inputValue;
    },
  },
  methods: {
    emitClose(): void {
      this.$emit("close");
    },
    saveFilter(): void {
      const editedFilter: IFilterModel = {
        ...this.filter,
        term: this.inputValue,
      };
      this.inputValue = "";
      this.$emit("save", editedFilter);
    },
    onUpdate(value: string): void {
      this.inputValue = value;
      this.$emit("update", value);
    },
  },
  watch: {
    "filter.term": {
      handler(newVal) {
        this.inputValue = newVal;
      },
    },
  },
});
</script>

<style lang="scss">
.page-filter-popup {
  width: 290px;
}
</style>
