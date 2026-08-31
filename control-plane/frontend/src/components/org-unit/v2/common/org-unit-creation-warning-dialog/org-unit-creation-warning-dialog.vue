<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="modelValue" size="width-sm" :is-modal-persistent="true" @close="handleClose">
    <template #header> {{ title }} </template>

    <template #body>
      <div class="q-gutter-md text-body2">
        <div v-dompurify-html="message"></div>

        <q-checkbox
          class="no-margin q-pl-sm q-pt-lg"
          v-model="dontShowAgain"
          label="Don't show this again"
          @update:model-value="handleDontShowAgainChange"
        />
      </div>
    </template>

    <template #footer>
      <div class="q-pa-sm">
        <q-btn
          flat
          color="primary"
          :label="continueEditingLabel"
          @click="handleContinueEditing"
          aid="wizard-warning-continue-editing-btn"
        />
        <q-btn
          color="primary"
          :label="createAndCloseLabel"
          @click="handleCreateAndClose"
          aid="wizard-warning-create-and-close-btn"
        />
      </div>
    </template>
  </runai-base-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { storageUtil } from "@/utils/storage.util";

export interface Props {
  modelValue: boolean;
  storageKey: string;
  title: string;
  message: string;
  continueEditingLabel: string;
  createAndCloseLabel: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "update:model-value", value: boolean): void;
  (e: "continue-editing"): void;
  (e: "create-and-close"): void;
}>();

const dontShowAgain = ref(false);

const handleDontShowAgainChange = (value: boolean): void => {
  dontShowAgain.value = value;
};

const updateStorageIfNeeded = (): void => {
  if (dontShowAgain.value) {
    storageUtil.save(props.storageKey, true);
  }
};

const handleClose = (): void => {
  emit("update:model-value", false);
};

const handleContinueEditing = (): void => {
  emit("continue-editing");
  updateStorageIfNeeded();
  handleClose();
};

const handleCreateAndClose = (): void => {
  emit("create-and-close");
  updateStorageIfNeeded();
  handleClose();
};
</script>
