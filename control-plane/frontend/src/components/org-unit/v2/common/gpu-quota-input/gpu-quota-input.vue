<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <quota-input
    :model-value="modelValue"
    :state="state"
    :message="message"
    :loading="loading"
    label="GPUs"
    :rules="rules"
    :prevent-null="!treatMinusOneAsUnlimited"
    :default-value="0"
    :placeholder="treatMinusOneAsUnlimited ? 'Unlimited' : undefined"
    :treat-minus-one-as-unlimited="treatMinusOneAsUnlimited"
    min="0"
    @update:model-value="$emit('update:model-value', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { QuotaInput } from "@/components/org-unit/v2/common/quota-input";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EQuotaEntity } from "@/models/resource.model";

export interface Props {
  modelValue: number | null;
  state?: EQuotaInputState;
  message?: string;
  loading?: boolean;
  entityType?: EQuotaEntity; // Used to determine if -1 should display as Unlimited
}

const props = withDefaults(defineProps<Props>(), {
  state: EQuotaInputState.Default,
  message: "",
  loading: false,
  entityType: EQuotaEntity.project,
});

defineEmits<{
  (e: "update:model-value", value: number | null): void;
}>();

// For departments, -1 displays as Unlimited. For projects, -1 is not valid.
const treatMinusOneAsUnlimited = computed(() => {
  return props.entityType === EQuotaEntity.department;
});

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
