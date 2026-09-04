<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="rule-enforcement-section">
    <div class="q-mb-md text-black-87">Add enforcements and set their values</div>
    <rule-enforcement-card
      v-for="(enforcement, i) in enforcements"
      :key="enforcement.id"
      :model-value="enforcement"
      :field-info="fieldInfo"
      :index="i"
      :used-enforcements="otherEnforcements(i)"
      :show-validation="showValidation"
      @update:model-value="(value) => emit('update:enforcement', i, value)"
      @remove="emit('remove', i)"
    />
    <runai-btn
      flat
      color="primary"
      class="btn-icon-sm"
      icon="ph-bold ph-plus"
      label="Enforcement"
      aid="add-enforcement-btn"
      :disable="!canAddMore"
      @click="emit('add')"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import RuleEnforcementCard from "./rule-enforcement-card.vue";
import type { IEnforcementDraft, IFieldInfo } from "@/models/gvk-policy.model";
import { buildEnforcementOptions } from "./policy-rule-builder.constant";

export interface Props {
  enforcements: IEnforcementDraft[];
  fieldInfo: IFieldInfo;
  /** Reveal each enforcement's cross-field value error inline (set on a save attempt). */
  showValidation?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (event: "update:enforcement", index: number, value: IEnforcementDraft): void;
  (event: "add"): void;
  (event: "remove", index: number): void;
}>();

function otherEnforcements(index: number): IEnforcementDraft[] {
  return props.enforcements.filter((_, enforcementIndex) => enforcementIndex !== index);
}

const canAddMore = computed<boolean>(() => {
  return buildEnforcementOptions(
    props.fieldInfo.allowedEnforcements,
    props.enforcements,
    props.fieldInfo.setDefaultGroupingArg,
  ).some((option) => !option.disable);
});
</script>

<style lang="scss" scoped></style>
