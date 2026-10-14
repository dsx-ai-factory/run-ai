<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="policy-rule-summary row items-center justify-between no-wrap q-px-md q-py-sm">
    <div class="summary-text text-black-87">
      <div class="summary-col ellipsis" :title="`Field: ${rule.field}`">Field: {{ rule.field }}</div>
      <div class="summary-col ellipsis" :title="`Spec selector: ${rule.specSelector}`">
        Spec selector: {{ rule.specSelector }}
      </div>
    </div>
    <div class="row items-center no-wrap">
      <runai-tooltip
        v-if="inactiveTooltip"
        aid="inactive-rule-warning"
        icon="ph ph-warning-circle"
        color="warning"
        class="inactive-rule-warning"
        :tooltip-text="inactiveTooltip"
        :aria-label="inactiveTooltip"
        width="450px"
      />
      <runai-btn
        flat
        round
        dense
        icon="ph-bold ph-pencil-simple"
        tooltip="Edit rule"
        class="summary-icon-btn"
        :aid="`edit-rule-${index}`"
        @click="emit('edit')"
      />
      <runai-btn
        flat
        round
        dense
        icon="ph-bold ph-x"
        tooltip="Remove rule"
        class="summary-icon-btn"
        :aid="`delete-rule-${index}`"
        @click="emit('delete')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import type { IRuleDraft } from "@/models/gvk-policy.model";

export interface Props {
  rule: IRuleDraft;
  index: number;
  inactiveTooltip?: string;
}

defineProps<Props>();

const emit = defineEmits<{
  (event: "edit"): void;
  (event: "delete"): void;
}>();
</script>

<style lang="scss" scoped>
.inactive-rule-warning :deep(.q-icon) {
  color: var(--icon-warning);
  font-size: 20px;
}
.policy-rule-summary {
  border: 1px solid var(--border-subtle);
  border-radius: var(--b-border-radius-radius-xs);
  background: var(--bg-primary-default);
}
.summary-text {
  flex: 1;
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr);
  column-gap: var(--space-m, 16px);
  align-items: center;
}
.summary-col {
  min-width: 0;
}
// Anchor on row root: runai-btn's root lacks this scope attr, so :deep() off the button class never matches.
.policy-rule-summary :deep(.summary-icon-btn .q-icon) {
  color: var(--icon-default);
  font-size: 20px;
}
.policy-rule-summary :deep(.summary-icon-btn:hover) {
  background-color: var(--bg-primary-hover);
}
</style>
