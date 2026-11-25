<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <move-to-rank-modal
    :selected-count="selectedCount"
    :entity-type="entityType"
    :current-rank="currentRank"
    :rank-options="rankOptions"
    custom-class="scheduling-move-dialog"
    no-padding
    @close="$emit('close')"
    @move="$emit('move', $event)"
  />
</template>

<script setup lang="ts">
import { MoveToRankModal } from "@/components/org-unit/v2/common/peer-org-unit-rank-list/move-to-rank-modal/index";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import type { ISelectOption } from "@/models/global.model";
import { EQuotaEntity } from "@/models/resource.model";

export interface IProps {
  selectedCount: number;
  entityType?: EQuotaEntity;
  currentRank?: string;
}

withDefaults(defineProps<IProps>(), {
  entityType: EQuotaEntity.department,
  currentRank: undefined,
});

defineEmits<{
  (e: "close"): void;
  (e: "move", rank: string): void;
}>();

const rankOptions: ISelectOption[] = [...ENTITY_PRIORITY_SELECT_OPTIONS].reverse();
</script>

<style lang="scss">
.scheduling-move-dialog {
  width: 500px !important;
}
.scheduling-move-dialog .runai-base-dialog-header {
  height: 64px;
  padding: 16px;
}
.scheduling-move-dialog .runai-base-dialog-body {
  min-height: 155px;
}
.scheduling-move-dialog .runai-base-dialog-footer {
  min-height: 70px;
  padding: 16px;
}
.scheduling-move-dialog .runai-base-dialog-footer .q-btn:first-child {
  color: var(--text-brand) !important;
}
</style>
