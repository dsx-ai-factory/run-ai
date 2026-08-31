<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-select
    v-if="selectOptions"
    :aid="aid"
    class="enforcement-value-input"
    dense
    :label="label"
    :model-value="modelValue"
    :options="selectOptions"
    @update:model-value="onSelect"
    emit-value
    filterable
    :no-option-text="NO_RESULTS_FOUND_LABEL"
    :placeholder="placeholder"
    :rules="rules"
  />
  <policy-number-input
    v-else-if="isNumeric"
    :aid="aid"
    class="enforcement-value-input"
    dense
    :label="label"
    :model-value="numberValue"
    :is-integer="fieldType === 'integer' || fieldType === 'integers'"
    :placeholder="placeholder"
    :rules="rules"
    @update:model-value="onNumber"
  />
  <policy-string-field
    v-else
    :aid="aid"
    class="enforcement-value-input"
    dense
    :label="label"
    :model-value="modelValue"
    :placeholder="placeholder"
    :hint="hintForType"
    :rules="rules"
    @update:model-value="onString"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { PolicyStringField } from "@/components/common/policy-string-field";
import type { ISelectOption } from "@/models/global.model";
import type { MappingFieldType } from "@/swagger-models/policy-service-client";
import { BOOLEAN_VALUE_OPTIONS, NO_RESULTS_FOUND_LABEL } from "./policy-rule-builder.constant";

export interface Props {
  modelValue: string;
  fieldType: MappingFieldType;
  enumValues?: string[];
  aid?: string;
  /** In-field floating label (e.g. "Value" or a set-param name). */
  label?: string;
  placeholder?: string;
  rules?: Array<(val: string) => boolean | string>;
}

const props = withDefaults(defineProps<Props>(), {
  enumValues: () => [],
  rules: () => [],
});

const emit = defineEmits<{
  (event: "update:model-value", value: string): void;
}>();

const selectOptions = computed<ISelectOption[] | null>(() => {
  // Value-based per-option aids keyed off this editor's aid so e2e can deterministically select boolean/enum values, unique across editors.
  const base = props.aid ?? "enforcement-value";
  if (props.enumValues.length > 0) {
    return props.enumValues.map((value) => ({ label: value, value, aid: `${base}-option-${value}` }));
  }
  if (props.fieldType === "boolean") {
    return BOOLEAN_VALUE_OPTIONS.map((option) => ({ ...option, aid: `${base}-option-${option.value}` }));
  }
  return null;
});

const isNumeric = computed<boolean>(() => props.fieldType === "integer" || props.fieldType === "number");

const numberValue = computed<number | null>(() => {
  if (props.modelValue === "" || props.modelValue == null) return null;
  const parsed = Number(props.modelValue);
  return Number.isNaN(parsed) ? null : parsed;
});

const hintForType = computed<string>(() => {
  switch (props.fieldType) {
    case "strings":
    case "namesArray":
    case "integers":
    case "numbers":
      return "Comma-separated values";
    case "object":
      return "Enter a JSON object";
    default:
      return "";
  }
});

function onSelect(value: string): void {
  emit("update:model-value", value ?? "");
}
function onString(value: string | null): void {
  emit("update:model-value", value ?? "");
}
function onNumber(value: number | string | null): void {
  emit("update:model-value", value == null ? "" : String(value));
}
</script>

<style lang="scss" scoped></style>
