<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-list class="runai-checkbox-list" :style="style">
    <q-item v-if="showSearch">
      <q-item-section>
        <q-input v-model="search" type="search" aria-label="Search columns">
          <template v-slot:append>
            <q-icon name="ph-bold ph-magnifying-glass" size="20px" />
          </template>
        </q-input>
      </q-item-section>
    </q-item>
    <q-item>
      <q-item-section>
        <q-checkbox
          v-if="showAllOption"
          @update:model-value="allClicked"
          v-model="allSelectedModel"
          :indeterminate-value="null"
          label="All"
        />
        <q-option-group
          :options="displayItems"
          type="checkbox"
          v-model="selectedItems"
          @update:model-value="onSelectedItemsChanged"
        />
      </q-item-section>
    </q-item>
  </q-list>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

import type { ICheckboxItem } from "./model";

export default defineComponent({
  events: ["changed"],
  props: {
    options: {
      type: Array as PropType<Array<ICheckboxItem>>,
      required: true,
    },
    selected: {
      type: Array as PropType<Array<string>> | PropType<null>,
      required: true,
    },
    showSearch: {
      type: Boolean,
      required: false,
      default: true,
    },
    showAllOption: {
      type: Boolean,
      required: false,
      default: true,
    },
    style: {
      type: String,
      required: false,
    },
  },
  data() {
    return {
      selectedItems: [] as Array<string>,
      search: "" as string,
      allSelectedModel: true as boolean | null,
    };
  },
  created() {
    this.selectedItems = this.selected || this.options.map((col: ICheckboxItem) => col.value);
    this.updateAllSelected();
  },
  computed: {
    displayItems(): Array<ICheckboxItem> {
      return this.options
        .filter((col: ICheckboxItem) =>
          this.search ? col.label.toLocaleLowerCase().includes(this.search.toLowerCase()) : true,
        )
        .filter((col: ICheckboxItem) => !!col.label);
    },
  },
  methods: {
    allClicked(isSelected: boolean): void {
      if (isSelected) {
        this.selectedItems = this.options.map((option: ICheckboxItem) => option.value);
      } else {
        this.selectedItems = this.options
          .filter((col: ICheckboxItem) => col.disable)
          .map((option: ICheckboxItem) => option.value);
      }

      this.updateAllSelected();
      this.updateSelectedChanged();
    },
    onSelectedItemsChanged(): void {
      this.updateAllSelected();
      this.updateSelectedChanged();
    },
    updateAllSelected() {
      if (this.selectedItems.length === this.options.length) {
        this.allSelectedModel = true;
      } else if (this.selectedItems.length === 0) {
        this.allSelectedModel = false;
      } else {
        this.$nextTick(() => {
          this.allSelectedModel = null;
        });
      }
    },
    updateSelectedChanged() {
      this.$emit("changed", this.selectedItems);
    },
  },
});
</script>
