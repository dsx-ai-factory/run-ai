<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-editor-modal">
    <runai-base-dialog
      model-value
      no-padding
      size="width-xl"
      dialog-body-size="large"
      @close="onCancel"
      @dialog-loaded="showEditor"
      is-modal-persistent
      hide-subheader-separator
    >
      <template #header>{{ headerLabel }}</template>
      <template #subheader>
        <slot name="subheader" />
      </template>

      <template #body>
        <q-linear-progress v-if="isLoading" animation-speed="100" indeterminate />
        <div v-else class="q-pa-md">
          <runai-editor
            @changed="onChange"
            @validation-status="syntaxValidationChanged"
            :code="initialCode"
            :height="height"
          />
        </div>
      </template>
      <template #footer-left>
        <div class="row q-pl-sm">
          <q-field
            class="col-4 form-hint no-padding"
            :model-value="hasError"
            :error="hasError"
            :error-message="invalidMessage"
          ></q-field>
        </div>
      </template>
      <template #footer>
        <q-btn label="discard" color="primary" flat @click="onCancel" aid="editor-modal-cancel-button" />
        <q-btn
          label="apply"
          color="primary"
          @click="onApply"
          :loading="validating"
          aid="editor-modal-apply-button"
          type="submit"
        />
      </template>
    </runai-base-dialog>

    <runai-base-dialog v-if="isLeaveModalOpen" :model-value="isLeaveModalOpen" size="width-sm" @close="onStay">
      <template #header> Discard? </template>

      <template #body>
        <div>You have unsaved changes. If you leave, you will lose your changes.</div>
      </template>

      <template #footer>
        <q-btn label="Leave" aid="leave-modal-leave-btn" flat color="primary" @click="onLeave" />
        <q-btn label="Stay On Page" aid="leave-modal-stay-btn" color="primary" @click="onStay" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
// types
import type { IEditorValidationStatus } from "@/components/common/runai-editor/runai-editor.model";
import { defineComponent, type PropType } from "vue";
import { RunaiEditor } from "../runai-editor";

export enum Events {
  CANCEL = "cancel",
  SUBMIT = "submit",
  CHANGED = "changed",
}

export default defineComponent({
  components: {
    RunaiBaseDialog,
    RunaiEditor,
  },
  emits: [Events.SUBMIT, Events.CANCEL, Events.CHANGED],
  props: {
    initialCode: {
      type: String as PropType<string>,
      required: true,
    },
    invalid: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    height: {
      type: String as PropType<string>,
      required: false,
      default: "65vh",
    },
    headerLabel: {
      type: String as PropType<string>,
      required: false,
      default: "YAML Editor",
    },
    invalidMessage: {
      type: String as PropType<string>,
      required: false,
      default: "Invalid YAML",
    },
    validating: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      isLoading: true as boolean,
      updatedCode: this.initialCode as string,
      isLeaveModalOpen: false as boolean,
      isValid: true as boolean,
    };
  },
  computed: {
    hasError(): boolean {
      // show the footer error on a YAML syntax error or when the consumer flags the policy as invalid (e.g. empty policy)
      return !this.isValid || this.invalid;
    },
  },
  methods: {
    syntaxValidationChanged(validationStatus: IEditorValidationStatus): void {
      this.isValid = validationStatus.isValid;
    },
    showEditor() {
      this.$nextTick(() => {
        this.isLoading = false;
      });
    },
    async onApply(): Promise<void> {
      if (!this.isValid || this.validating) return;
      this.$emit(Events.SUBMIT, this.updatedCode);
    },
    onCancel(): void {
      if (this.updatedCode !== this.initialCode) {
        this.isLeaveModalOpen = true;
      } else this.onLeave();
    },
    onChange(updatedCode: string): void {
      this.updatedCode = updatedCode;
      this.$emit(Events.CHANGED, updatedCode);
    },
    onStay(): void {
      this.isLeaveModalOpen = false;
    },
    onLeave(): void {
      this.$emit(Events.CANCEL);
    },
  },
});
</script>
<style lang="scss" scoped></style>
