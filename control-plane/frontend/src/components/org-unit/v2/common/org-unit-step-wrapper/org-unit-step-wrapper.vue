<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="org-unit-step-wrapper">
    <runai-expansion-item :label="label" default-opened prevent-top-padding-when-open disable-closing body-no-padding>
      <slot />

      <!-- Multiple node pools: show inner Save button inside expansion item -->
      <template v-if="!readOnly && !hideSaveButton">
        <q-separator />

        <q-card-actions align="right" class="q-px-lg">
          <q-btn color="primary" flat :label="saveLabel" :loading="loading" @click="$emit('save')" />
        </q-card-actions>
      </template>
    </runai-expansion-item>

    <!-- Single node pool -->
    <q-card-section v-if="!readOnly && hideSaveButton" class="row justify-end q-gutter-x-sm q-px-none">
      <q-btn flat color="primary" label="Close" @click="$emit('close')" />
      <q-btn
        aid="wizard-primary-btn"
        color="primary"
        label="Save & Continue"
        :loading="loading"
        @click="$emit('save-and-continue')"
      />
    </q-card-section>
  </div>
</template>

<script setup lang="ts">
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { QBtn, QCardActions, QCardSection, QSeparator } from "quasar";

export interface Props {
  label: string;
  saveLabel?: string;
  loading?: boolean;
  readOnly?: boolean;
  /** Hide the inline save button and show footer-style buttons outside expansion item instead */
  hideSaveButton?: boolean;
}

withDefaults(defineProps<Props>(), {
  saveLabel: "Save",
  loading: false,
  readOnly: false,
  hideSaveButton: false,
});

defineEmits<{
  (e: "save"): void;
  (e: "save-and-continue"): void;
  (e: "close"): void;
}>();
</script>
