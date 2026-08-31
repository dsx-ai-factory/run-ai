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
    input-label="The following value:"
    :max-width="maxWidth"
  >
    <template #body>
      <div class="checkbox-container" role="group" :aria-label="filter.label">
        <span class="text-black-54 q-mb-md" v-if="!isInputTypeRadio">Matches any:</span>
        <q-item dense v-for="(option, index) in selectOptions" :key="index" role="presentation">
          <q-radio
            v-if="isInputTypeRadio"
            class="no-padding"
            v-model="selectedOption"
            :val="option.value"
            :label="option.label"
          />
          <q-checkbox v-else class="no-padding" v-model="selectedValues" :val="option.value" :label="option.label" />
        </q-item>
      </div>
    </template>
  </runai-page-filter-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IEnumBasedFilterModel, IEnumBasedFilterSelectOption } from "@/models/filter.model";
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";

export default defineComponent({
  name: "runai-enum-based-filter",
  components: { RunaiPageFilterWrapper },
  emits: ["save", "close"],
  props: {
    filter: {
      type: Object as PropType<IEnumBasedFilterModel>,
      required: true,
    },
  },
  data() {
    return {
      selectedValues: [] as string[],
      selectedOption: null as boolean | null,
    };
  },
  created() {
    if (this.isInputTypeRadio && this.filter.selectedValues.length > 0) {
      const selectedOption = this.selectOptions.find((option) => option.value === this.filter.selectedValues[0])?.value;
      if (typeof selectedOption === "boolean" || selectedOption) {
        this.selectedOption = selectedOption as boolean;
      }
    } else {
      this.selectedValues = [...this.filter.selectedValues] as string[];
    }
  },
  computed: {
    selectOptions(): IEnumBasedFilterSelectOption[] {
      return this.filter.selectOptions.filter((option) => !option.hidden);
    },
    isInputTypeRadio(): boolean {
      return this.filter.inputType === "radio";
    },
    maxWidth(): string {
      return this.filter.maxWidth || "250px";
    },
  },
  methods: {
    saveFilter(): void {
      const valuesToEmit = this.isInputTypeRadio
        ? [this.selectOptions.find((option) => option.value === this.selectedOption)?.value]
        : this.selectedValues;
      this.$emit("save", {
        ...this.filter,
        selectedValues: valuesToEmit,
      });
    },
  },
});
</script>

<style scoped lang="scss">
.checkbox-container {
  display: flex;
  justify-content: start;
  flex-direction: column;
  padding: unset;
  .q-item {
    padding: 0;
  }
}
</style>
