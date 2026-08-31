<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-itemized-list">
    <div v-if="title" class="q-pb-md q-gutter-md row items-center" :class="paddingTop ? 'q-pt-md' : ''">
      <span>{{ title }}</span>
      <runai-tooltip v-if="titleTooltip" :tooltip-text="titleTooltip" tooltip-position="right" />
    </div>
    <section class="list-container">
      <runai-itemized-row
        :class="runaiItemizedRowClass"
        v-for="item in displayedItems"
        :key="item.id"
        :aid="`itemized-list-row-${item.id}`"
        :item="item.origin"
        :disabled-tooltip="getDisableRowTooltip(item.isLocked)"
        :disabled="item.isLocked || disable"
        :should-validate-value="shouldValidateValue"
        :item-name-label="itemNameLabel"
        :item-value-label="itemValueLabel"
        :name-input-validation-message="nameInputValidationMessage"
        :value-input-validation-message="valueInputValidationMessage"
        :stack-inputs-label="stackInputsLabel"
        :prevent-delete="deleteDisabledIfSingleItem"
        :is-view-mode="isViewMode"
        :value-input-autogrow="valueInputAutogrow"
        :disabled-keys="disabledKeys"
        @changed="update(item.id, $event)"
        @remove="remove(item.id)"
      />
    </section>
    <q-btn
      v-if="addItemBtn && !hideAddItemBtn"
      aid="add-item-button"
      class="add-item-button btn-icon-sm"
      flat
      color="primary"
      icon="ph-bold ph-plus"
      :label="addItemBtn"
      :disable="disableAddButton"
      @click="add"
    >
      <q-tooltip v-if="disableTooltipInfo">{{ disableTooltipInfo }}</q-tooltip>
    </q-btn>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { IItemizedListItem } from "./runai-itemized-list.model";

// components
import { RunaiItemizedRow } from "./runai-itemized-row";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// utils
import type { AnnotationsDefaults, InstancesRules, LabelsDefaults } from "@/swagger-models/policy-service-client";
import { useItemizedPolicy } from "@/composables/use-itemized-policy.composable";
import { ItemizedMapperEntities } from "@/composables/helpers/use-itemized-policy.mapper";

export default defineComponent({
  components: {
    RunaiItemizedRow,
    RunaiTooltip,
  },
  emits: ["changed"],
  props: {
    itemsList: {
      type: Array as PropType<Array<IItemizedListItem>>,
      required: false,
      default: () => [],
    },
    title: {
      type: String as PropType<string>,
      required: false,
    },
    titleTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    addItemBtn: {
      type: String as PropType<string>,
      required: false,
    },
    hideAddItemBtn: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disableInfo: {
      type: String as PropType<string>,
      required: false,
    },
    shouldValidateValue: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    itemNameLabel: {
      type: String as PropType<string>,
      required: false,
    },
    itemValueLabel: {
      type: String as PropType<string>,
      required: false,
    },
    nameInputValidationMessage: {
      type: String as PropType<string>,
      required: false,
    },
    valueInputValidationMessage: {
      type: String as PropType<string>,
      required: false,
    },
    stackInputsLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    preventDeleteIfSingleItem: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    isViewMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<InstancesRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<LabelsDefaults | AnnotationsDefaults | null>,
      required: false,
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
    paddingTop: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  setup(props) {
    const { disabledByPolicyText, cannotAddDueToPolicy, addItem, removeItem, updateItem, displayedItems, fullList } =
      useItemizedPolicy(
        props.itemsList,
        ItemizedMapperEntities.DefaultNameValue,
        props.policyRules?.instances,
        props.policyDefaults?.instances,
      );

    return {
      disabledByPolicyText,
      cannotAddDueToPolicy,
      addItem,
      removeItem,
      updateItem,
      displayedItems,
      fullList,
    };
  },
  computed: {
    deleteDisabledIfSingleItem(): boolean {
      return !!this.preventDeleteIfSingleItem && this.displayedItems.length === 1;
    },
    runaiItemizedRowClass(): string {
      return `runai-itemized-list-row${this.isViewMode ? " mb-none" : ""}`;
    },
    disableAddButton(): boolean {
      return this.disable || this.cannotAddDueToPolicy;
    },
    disableTooltipInfo(): string | undefined {
      if (this.disable && this.disableInfo) return this.disableInfo;
      return this.cannotAddDueToPolicy ? this.disabledByPolicyText : undefined;
    },
  },
  methods: {
    add(): void {
      const newEnvVar: IItemizedListItem = { name: "", value: "" };
      this.addItem(newEnvVar);
      this.$emit("changed", this.fullList);
    },
    update(id: string, envVar: IItemizedListItem): void {
      this.updateItem(id, envVar);
      this.$emit("changed", this.fullList);
    },
    remove(id: string): void {
      this.removeItem(id);
      this.$emit("changed", this.fullList);
    },
    getDisableRowTooltip(isLocked: boolean): string | undefined {
      if (this.disable && this.disableInfo) return this.disableInfo;
      return isLocked ? this.disabledByPolicyText : undefined;
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-itemized-list {
  .runai-itemized-list-row {
    margin-bottom: 15px;
  }
  .mb-none {
    margin-bottom: 0;
  }
}
</style>
