<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-itemized-row" :class="runaiItemizedRowClass">
    <q-input
      aid="itemized-name-input"
      class="item input-field"
      @update:model-value="updateModel('name', $event)"
      :model-value="item.name"
      :label="itemNameLabel"
      :debounce="300"
      :placeholder="nameInputPlaceholder"
      :rules="itemizedNameInputRules"
      no-error-icon
      :stack-label="stackInputsLabel"
      :disable="disabled || isKeyDisabled"
      :dense="isViewMode"
    />
    <q-input
      aid="itemized-value-input"
      class="item input-field"
      @update:model-value="updateModel('value', $event)"
      :model-value="item.value"
      :label="itemValueLabel"
      :debounce="300"
      :placeholder="valueInputPlaceholder"
      :rules="itemizedValueInputRules"
      no-error-icon
      :stack-label="stackInputsLabel"
      :disable="disabled"
      :dense="isViewMode"
      :autogrow="valueInputAutogrow"
      :input-style="{ maxHeight: '4.8em' }"
      @keydown.enter="preventFormSubmission"
    />

    <span class="item">
      <q-btn
        aid="remove-itemized-button"
        class="close-button"
        icon="ph-bold ph-x"
        flat
        round
        @click="removeItem"
        v-if="!disabled && !preventDelete"
      >
        <q-tooltip>Remove</q-tooltip>
      </q-btn>
    </span>
    <q-tooltip v-if="shouldHaveTooltip">{{ disabledTooltip }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { IItemizedListItem } from "../runai-itemized-list.model";

import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty } from "@/common/form.validators";
import type { ValidationRule } from "quasar";

export default defineComponent({
  emits: ["changed", "remove"],
  props: {
    item: {
      type: Object as PropType<IItemizedListItem>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    shouldValidateValue: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    itemNameLabel: {
      type: String as PropType<string>,
      required: false,
      default: "Name",
    },
    itemValueLabel: {
      type: String as PropType<string>,
      required: false,
      default: "Value",
    },
    nameInputValidationMessage: {
      type: String as PropType<string>,
      required: false,
      default: errorMessages.NAME_NOT_EMPTY,
    },
    valueInputValidationMessage: {
      type: String as PropType<string>,
      required: false,
      default: errorMessages.VALUE_NOT_EMPTY,
    },
    stackInputsLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    preventDelete: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    isViewMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    valueInputAutogrow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabledKeys: {
      type: Array as PropType<string[]>,
      required: false,
      default: () => [],
    },
  },
  computed: {
    nameInputPlaceholder(): string {
      return `Enter a ${this.itemNameLabel.toLocaleLowerCase()}`;
    },
    valueInputPlaceholder(): string {
      return `Enter a ${this.itemValueLabel.toLocaleLowerCase()}`;
    },
    valueInputRules(): ValidationRule[] {
      return this.shouldValidateValue ? [this.isEmptyValueField] : [];
    },
    shouldHaveTooltip(): boolean {
      return !!this.disabled && !!this.disabledTooltip;
    },
    runaiItemizedRowClass(): string {
      return this.isViewMode ? "q-px-md" : "padding-15";
    },
    itemizedNameInputRules(): undefined | ((val: string) => string | boolean)[] {
      return this.isViewMode ? undefined : [this.isEmptyNameField];
    },
    itemizedValueInputRules(): undefined | ValidationRule[] {
      return this.isViewMode ? undefined : this.valueInputRules;
    },
    isKeyDisabled(): boolean {
      return this.disabledKeys.includes(this.item.name || "");
    },
  },
  methods: {
    updateModel(field: string, value: string | null | number): void {
      this.$emit("changed", { ...this.item, [field]: `${value === null ? "" : value}` });
    },
    removeItem(): void {
      this.$emit("remove", this.item);
    },
    isEmptyNameField(val: string): string | boolean {
      return isNotEmpty(val) || this.nameInputValidationMessage;
    },
    isEmptyValueField(val: string): string | boolean {
      return isNotEmpty(val) || this.valueInputValidationMessage;
    },
    preventFormSubmission(event: KeyboardEvent) {
      event.stopPropagation();
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-itemized-row {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr 36px;
  grid-column-gap: 45px;
  background-color: $body-background-color;

  .item {
    display: flex;
    justify-content: center;
    align-items: center;
  }

  .close-button {
    color: $black-54;
    font-size: 12px;
  }
}

.padding-15 {
  padding: 15px;
}
</style>
