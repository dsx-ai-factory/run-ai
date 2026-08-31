<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <quota-input
    :model-value="localValue"
    :state="quotaState"
    :message="quotaMessage"
    :label="label"
    :rules="rules"
    placeholder="Unlimited"
    @update:model-value="emitValue"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { QuotaInput } from "@/components/org-unit/v2/common/quota-input";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";

export interface Props {
  modelValue: number | null;
  deserved: number; // The assigned GPU quota - limit cannot be less than this
  entityType: EQuotaEntity;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "update:model-value", value: number | null): void;
}>();

const localValue = computed({
  get: () => props.modelValue,
  set: (val: number | null) => emit("update:model-value", val),
});

const isNegative = computed(() => {
  return (localValue.value !== null && localValue.value < 0) || isBelowDeserved.value;
});

const label = computed(() => {
  return props.entityType === EQuotaEntity.project ? "Max GPU allocation" : "Max GPU allocation";
});

const isBelowDeserved = computed(() => {
  if (localValue.value === null) return false; // Unlimited is always valid
  return localValue.value <= props.deserved;
});

const quotaState = computed(() => {
  if (isNegative.value) return EQuotaInputState.Error;
  if (isBelowDeserved.value) return EQuotaInputState.Warning;
  return EQuotaInputState.Default;
});

const quotaMessage = computed(() => {
  const formattedLabel = label.value.charAt(0).toLowerCase() + label.value.slice(1);
  if (isBelowDeserved.value) {
    return `The ${formattedLabel} must be greater than the assigned quota. Enter a number above ${props.deserved} GPUs.`;
  }
  if (isNegative.value) {
    return `The ${formattedLabel} must be 0 or greater`;
  }
  return "";
});

const rules = computed(() => {
  const stateRule = () => {
    if (quotaState.value === EQuotaInputState.Error) {
      return ""; // Return an empty string to indicate an error
    }
    return true;
  };
  return [stateRule];
});

const emitValue = (value: number | null) => {
  emit("update:model-value", value);
};

// Expose for testing
defineExpose({
  quotaState,
  quotaMessage,
  isNegative,
  isBelowDeserved,
  rules,
});
</script>
