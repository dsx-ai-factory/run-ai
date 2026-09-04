<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-delete-modal">
    <runai-base-dialog :model-value="true" @close="onCancel" size="width-sm" :no-padding="true" :loading="loading">
      <template #header>{{ headerTitle }}</template>

      <template #body>
        <div class="delete-cluster-body q-px-lg q-py-md column">
          <slot name="subheader"></slot>
          <slot name="title">
            <div class="delete-modal-msg text-weight-medium">
              {{ title }}
            </div>
          </slot>
          <div class="delete-modal-info q-my-sm">
            <slot name="description"> {{ description }} </slot>
          </div>
        </div>
      </template>

      <template #footer-left></template>
      <template #footer>
        <q-btn label="Cancel" color="primary" flat @click="onCancel" aid="delete-modal-cancel-button" />
        <q-btn
          :label="deleteButtonLabel"
          color="primary"
          :loading="deleting"
          @click="onDelete"
          aid="delete-modal-delete-button"
        />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { getEntityWithArticle } from "@/utils/string.util/string.util";
// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

export default defineComponent({
  components: {
    RunaiBaseDialog,
  },
  emits: ["cancel", "delete"],
  props: {
    entityName: {
      type: String as PropType<string>,
      required: false,
    },
    entityType: {
      type: String as PropType<string>,
      default: "",
    },
    customDescription: {
      type: String as PropType<string>,
      default: null,
    },
    headerTitle: {
      type: String as PropType<string>,
      default: "Delete",
    },
    customTitle: {
      type: String as PropType<string>,
      default: null,
    },
    deleting: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    deleteButtonLabel: {
      type: String as PropType<string>,
      default: "Delete",
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    title(): string {
      return this.customTitle ? this.customTitle : `Permanently delete ${this.entityType}: ${this.entityName}?`;
    },
    description(): string {
      return this.customDescription
        ? this.customDescription
        : `Once ${getEntityWithArticle(this.entityType)} is deleted, it cannot be restored.`;
    },
  },
  methods: {
    onDelete(): void {
      this.$emit("delete");
    },
    onCancel(): void {
      this.$emit("cancel");
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-delete-modal {
  .delete-modal-msg {
    font-size: 16px;
  }
  .delete-modal-info {
    font-size: 14px;
  }
}
</style>
