<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="delete-setting-modal">
    <runai-base-dialog :model-value="true" @close="$emit('cancel')" size="width-md" :no-padding="true">
      <template #header>Remove</template>

      <template #body>
        <div class="q-px-lg column">
          <p class="q-mt-md text-subtitle1">{{ deleteConfirmMessage }}{{ questionMark }}</p>
        </div>
        <div class="q-px-lg column" v-if="deleteConfirmMessageSubtitle">
          <p class="text-caption">{{ deleteConfirmMessageSubtitle }}</p>
        </div>
      </template>

      <template #footer>
        <q-btn label="don't remove" color="primary" flat @click="$emit('cancel')" aid="delete-modal-cancel-button" />
        <q-btn label="remove" color="primary" @click="$emit('delete')" aid="delete-modal-delete-button" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

// models

export default defineComponent({
  components: {
    RunaiBaseDialog,
  },
  emits: ["cancel", "delete"],
  props: {
    deleteConfirmMessage: {
      type: String as PropType<string>,
      required: true,
    },
    deleteConfirmMessageSubtitle: {
      type: String as PropType<string>,
    },
  },
  computed: {
    questionMark(): string {
      return this.deleteConfirmMessage.includes("?") ? "" : "?";
    },
  },
});
</script>
