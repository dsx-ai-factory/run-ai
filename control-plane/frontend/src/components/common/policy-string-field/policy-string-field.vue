<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-tooltip-wrapper
    :display-tooltip="disableInput"
    :tooltip-text="disabledTooltipText"
    anchor="bottom middle"
    self="bottom middle"
  >
    <runai-select
      v-if="showOptionsSelect"
      class="policy-string-field-select"
      :aid="aid"
      :model-value="modelValue"
      :options="selectOptions"
      :disable="disableInput"
      @update:model-value="updateValue"
      :policy-rules="{ options: policyRules?.options, canEdit: policyRules?.canEdit, required: policyRules?.required }"
      :stack-label="stackLabel"
      :placeholder="placeholder"
      :label="label"
      emit-value
      :input-class="inputClass"
      :dense="dense"
      :hint="hint"
      :rules="rulesWithPolicy"
    />

    <q-input
      v-else
      class="policy-string-field-input"
      :aid="aid"
      :model-value="modelValue"
      @update:model-value="updateValue"
      :disable="disableInput"
      no-error-icon
      :stack-label="stackLabel"
      :placeholder="placeholder"
      :label="label"
      :input-class="inputClass"
      :dense="dense"
      :debounce="debounce"
      :hint="hint"
      :rules="rulesWithPolicy"
      :style="inlineStyle"
      :label-slot="enableLabelTooltip"
      :ref="inputRef"
    >
      <template v-if="enableLabelTooltip" v-slot:label>
        <span class="ra-input-label">
          {{ label }}
          <q-tooltip anchor="top middle" :offset="[0, 40]">
            {{ label }}
          </q-tooltip>
        </span>
      </template>
    </q-input>
    <q-tooltip class="explanatory-tooltip" v-if="showExplanatoryTooltip" max-width="300px">{{ tooltip }}</q-tooltip>
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { StringRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiSelect } from "../runai-select";
import type { ISelectOption } from "@/models/global.model";
import type { QInput, ValidationRule } from "quasar";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "policy-string-field",
  components: {
    RunaiTooltipWrapper,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: false,
      default: null,
    },
    policyRules: {
      type: Object as PropType<StringRules | null>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    tooltip: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    inputClass: {
      type: String as PropType<string>,
      required: false,
    },
    inlineStyle: {
      type: Object as PropType<Record<string, string>>,
      required: false,
    },
    elementClasses: {
      type: String as PropType<string>,
      required: false,
    },
    dense: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
    debounce: {
      type: Number as PropType<number>,
      required: false,
      default: 300,
    },
    customRequiredMessage: {
      type: String as PropType<string>,
      required: false,
    },
    enableLabelTooltip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      inputRef: "policy-string-field-input-ref",
    };
  },
  computed: {
    isBlocked(): boolean {
      return !!this.policyRules?.blocked;
    },
    disableInput(): boolean {
      if (this.policyRules?.canEdit === false || this.isBlocked) return true;
      return !!this.disable;
    },
    disabledTooltipText(): string {
      if (this.policyRules?.canEdit === false || this.isBlocked) {
        return errorMessages.CANT_BE_MODIFIED_POLICY;
      }
      return this.tooltip || "";
    },
    showOptionsSelect(): boolean {
      return !!this.policyRules?.options?.length;
    },
    selectOptions(): Array<ISelectOption> {
      if (!this.policyRules?.options) return [];
      return this.policyRules.options.map((option) => {
        return {
          label: option.displayed || option.value,
          value: option.value,
        };
      });
    },
    rulesWithPolicy(): ValidationRule[] {
      if (!this.policyRules) return this.rules || [];
      const rules = this.rules || [];
      if (this.policyRules.required) {
        rules.unshift(this.requiredRule);
      }
      return rules;
    },
    showExplanatoryTooltip(): boolean {
      if (this.disableInput) return false;
      return !!this.tooltip;
    },
  },
  methods: {
    requiredRule(val: string): boolean | string {
      return isNotEmpty(val) || this.customRequiredMessage || errorMessages.REQUIRED_FIELD;
    },
    updateValue(val: string | number | null): void {
      this.$emit("update:model-value", val);
    },
    focus(): void {
      const inputRef = this.$refs[this.inputRef] as HTMLInputElement | undefined;
      inputRef?.focus();
    },
    validate(): void {
      const inputRef = this.$refs[this.inputRef] as QInput | undefined;
      inputRef?.validate();
    },
  },
});
</script>
<style lang="scss" scoped>
.ra-input-label {
  pointer-events: stroke;
}
</style>
