<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <quota-input
    v-bind="$attrs"
    :model-value="modelValue"
    :state="state"
    :message="message"
    label="CPUs (Cores)"
    placeholder="Unlimited"
    min="0"
    :rules="rules"
    treat-minus-one-as-unlimited
    @update:model-value="$emit('update:model-value', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { QuotaInput } from "@/components/org-unit/v2/common/quota-input";
import { EQuotaInputState } from "@/models/org-unit.model";

export interface Props {
  modelValue: number | null;
  state?: EQuotaInputState;
  message?: string;
}

const props = withDefaults(defineProps<Props>(), {
  state: EQuotaInputState.Default,
  message: "",
});

defineEmits<{
  (e: "update:model-value", value: number | null): void;
}>();

const rules = computed(() => {
  const stateRule = () => {
    if (props.state === EQuotaInputState.Error) {
      return ""; // Return an empty string to indicate an error
    }
    return true;
  };
  return [stateRule];
});
</script>
