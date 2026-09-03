<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card-section class="row justify-end q-gutter-x-sm q-px-none">
    <!-- Read-only mode: show only Close button -->
    <template v-if="readOnly">
      <q-btn aid="wizard-close-btn" color="primary" label="Close" @click="emit('close')" />
    </template>

    <!-- Edit mode: show normal buttons -->
    <template v-else>
      <q-btn
        v-if="showSecondaryButton"
        flat
        color="primary"
        :label="secondaryLabel"
        :loading="submittingAction === 'close'"
        aid="wizard-secondary-btn"
        @click="handleSecondaryClick"
      />

      <q-btn
        color="primary"
        :label="primaryLabel"
        :loading="submittingAction === 'continue'"
        aid="wizard-primary-btn"
        @click="handlePrimaryClick"
      />
    </template>
  </q-card-section>
</template>

<script setup lang="ts">
import { computed } from "vue";

export type SubmittingAction = "close" | "continue" | null;

// --- Props ---
export interface Props {
  entityType: string;
  isNewEntity: boolean;
  isFirstStep: boolean;
  isLastStep: boolean;
  submittingAction?: SubmittingAction;
  readOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  submittingAction: null,
  readOnly: false,
});

// --- Emits ---
const emit = defineEmits<{
  (e: "close"): void;
  (e: "next"): void;
  (e: "submit-and-close"): void;
  (e: "submit-and-continue"): void;
}>();

const showSecondaryButton = computed(() => {
  return props.isFirstStep || !props.isLastStep;
});

const secondaryLabel = computed(() => {
  if (props.isFirstStep) {
    const action = props.isNewEntity ? "Create" : "Save";
    return `${action} ${props.entityType} & Close`;
  }
  return "Close";
});

const primaryLabel = computed(() => {
  if (props.isFirstStep) {
    const action = props.isNewEntity ? "Create" : "Save";
    return `${action} ${props.entityType} & Continue`;
  }

  if (!props.isLastStep) {
    return "Continue";
  }

  return "close";
});

const handleSecondaryClick = () => {
  if (props.isFirstStep) {
    emit("submit-and-close");
  } else {
    emit("close");
  }
};

const handlePrimaryClick = () => {
  if (props.isFirstStep) {
    emit("submit-and-continue");
  } else if (!props.isLastStep) {
    emit("next");
  } else {
    emit("submit-and-close");
  }
};
</script>

<style scoped lang="scss"></style>
