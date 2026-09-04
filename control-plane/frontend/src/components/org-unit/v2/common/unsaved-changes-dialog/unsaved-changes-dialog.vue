<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="modelValue" size="width-sm" :is-modal-persistent="true" @close="handleClose">
    <template #header>Save Changes</template>

    <template #body>
      <div class="text-body1">You have unsaved changes. Do you want to save them?</div>
    </template>

    <template #footer>
      <q-btn flat color="primary" label="Discard" @click="handleDiscard" />
      <q-btn color="primary" label="Save changes" @click="handleSave" />
    </template>
  </runai-base-dialog>
</template>

<script setup lang="ts">
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

export interface Props {
  modelValue: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  (e: "update:model-value", value: boolean): void;
  (e: "save"): void;
  (e: "discard"): void;
}>();

const handleClose = (): void => {
  emit("update:model-value", false);
};

const handleSave = (): void => {
  emit("save");
};

const handleDiscard = (): void => {
  emit("discard");
};
</script>
