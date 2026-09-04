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
        <!-- eslint-disable vuejs-accessibility/no-autofocus -- this whole component only mounts inside a q-menu that opens on click (:model-value="display"), never on a bare page load -->
        <runai-autocomplete-filter-input
          v-model="inputValue"
          aid="filter-autocomplete-input"
          :autocomplete-options-fn="autocompleteOptionsFn"
          :label="inputLabel"
          :autofocus="true"
        />
        <!-- eslint-enable vuejs-accessibility/no-autofocus -->
      </q-card-section>
      <q-card-section class="row justify-end q-py-sm">
        <q-btn type="submit" flat color="primary" label="Apply" v-close-popup @click="saveFilter" />
      </q-card-section>
    </form>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IFilterModel } from "@/models/filter.model";
import { RunaiAutocompleteFilterInput } from "@/components/common/runai-autocomplete-filter-input";

export default defineComponent({
  components: { RunaiAutocompleteFilterInput },
  emits: ["save", "close"],
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
    inputLabel: {
      type: String as PropType<string>,
      required: false,
      default: "Contains",
    },
    autocompleteOptionsFn: {
      type: Function as PropType<(searchQuery: string) => Promise<string[]>>,
      required: false,
    },
  },
  data() {
    return {
      inputValue: "" as string,
    };
  },
  created() {
    this.inputValue = this.filter.term;
  },
  updated() {
    this.inputValue = this.filter.term;
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
  width: 250px;
}
</style>
