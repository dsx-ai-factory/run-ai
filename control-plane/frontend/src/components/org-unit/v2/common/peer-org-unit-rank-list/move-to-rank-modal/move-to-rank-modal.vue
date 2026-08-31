<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-sm">
    <template #header>Move to Rank</template>

    <template #body>
      <div class="q-pa-md">
        <div class="text-body2 q-mb-md">
          Select the destination rank for the {{ selectedCount }} selected {{ entityType
          }}{{ selectedCount !== 1 ? "s" : "" }}
        </div>

        <runai-select
          v-model="selectedRank"
          :options="rankOptions"
          option-value="value"
          option-label="label"
          emit-value
          label="Destination rank"
          class="q-mb-md"
        />
      </div>
    </template>

    <template #footer>
      <q-btn label="Cancel" color="grey-7" flat @click="$emit('close')" />
      <q-btn
        :label="`Move ${selectedCount} ${entityType}${selectedCount !== 1 ? 's' : ''}`"
        color="primary"
        @click="handleMove"
      />
    </template>
  </runai-base-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { RunaiSelect } from "@/components/common/runai-select";
import { EQuotaEntity } from "@/models/resource.model";

export interface Props {
  selectedCount: number;
  entityType?: EQuotaEntity;
  currentRank?: string;
}

const props = withDefaults(defineProps<Props>(), {
  entityType: EQuotaEntity.project,
  currentRank: undefined,
});

const emit = defineEmits<{
  (e: "close"): void;
  (e: "move", rank: string): void;
}>();

const selectedRank = ref<string | null>(null);

const rankOptions = computed(() => {
  return ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => ({
    ...option,
    disable: option.value === props.currentRank,
  }));
});

const handleMove = () => {
  if (selectedRank.value) {
    emit("move", selectedRank.value);
  }
};
</script>

<style scoped lang="scss"></style>
