<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-rename-modal">
    <runai-base-dialog :model-value="modelValue" @close="$emit('close')" size="width-sm" :no-padding="true">
      <template #header>{{ mainAction }}</template>

      <template #body>
        <div class="rename-workspace-body q-px-lg q-py-md">
          <q-form @submit.prevent id="rename-form" ref="elRenameForm">
            <runai-name-validation
              v-if="showNameInput"
              v-model="newName"
              :rules="[isEmpty, isValidFormat, isNewNameUnique]"
              @invalid="invalidName = $event"
              debounce="300"
              auto-focus
              label="Name"
            />
            <q-input
              v-if="supportDescription"
              v-model="newDescription"
              debounce="300"
              :maxlength="250"
              counter
              label="Description"
              autogrow
            ></q-input>
          </q-form>
        </div>
      </template>

      <template #footer>
        <q-btn label="Cancel" color="primary" aid="rename-modal-cancel-button" flat @click="$emit('close')" />
        <q-btn
          type="submit"
          form="rename-form"
          :label="buttonLabel"
          @click="submit"
          color="primary"
          aid="rename-modal-save-button"
          :loading="renaming"
        />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
// common
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

// types
import type { ValidationRule } from "quasar";
import { errorMessages } from "@/common/error-message.constant";

// methods
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { isAssetNameUnique, isValidRunaiEntityNameFormat, isNotEmpty } from "@/common/form.validators";
import type { IScopeModel } from "@/models/global.model";
import { Scope, type AssetUpdateRequest } from "@/swagger-models/assets-service-client";
import type { IAssetNameValidationOptions } from "@/models/asset.model";

export default defineComponent({
  components: {
    RunaiBaseDialog,
    RunaiNameValidation,
  },
  emits: ["close", "rename"],
  props: {
    modelValue: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    existingName: {
      type: String as PropType<string>,
      required: true,
    },
    existingDescription: {
      type: String as PropType<string>,
    },
    renaming: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    assetType: {
      type: String as PropType<string>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    mainAction: {
      type: String as PropType<string>,
      default: "Rename",
    },
    submitLabel: {
      type: String as PropType<string>,
      default: "",
    },
    showNameInput: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      newName: "" as string,
      isNameUnique: null as ValidationRule | null,
      invalidName: false as boolean,
      newDescription: "" as string,
    };
  },
  created() {
    this.isNameUnique = isAssetNameUnique(this.assetType, this.assetNameOptions) as ValidationRule;
    this.newName = this.existingName;
    if (this.existingDescription !== undefined) {
      this.newDescription = this.existingDescription;
    }
  },
  computed: {
    buttonLabel(): string {
      return this.submitLabel || this.mainAction;
    },
    supportDescription(): boolean {
      return this.existingDescription !== undefined;
    },
    assetNameOptions(): IAssetNameValidationOptions {
      if (!this.scopeModel.scope || this.scopeModel.scope === Scope.Tenant)
        return { scopeType: Scope.Tenant } as IAssetNameValidationOptions;
      return {
        scopeType: this.scopeModel.scope,
        scopeId: this.scopeId,
      };
    },
    scopeId(): string {
      switch (this.scopeModel.scope) {
        case Scope.Cluster:
          return String(this.scopeModel.clusterId);
        case Scope.Department:
          return String(this.scopeModel.departmentId);
        case Scope.Project:
          return String(this.scopeModel.projectId);
        default:
          return "";
      }
    },
  },
  methods: {
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidFormat(val: string): boolean | string {
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT;
    },
    async submit(): Promise<void> {
      if (this.invalidName) return;
      if (this.newName === this.existingName && this.newDescription === this.existingDescription) {
        this.$emit("close");
        return;
      }
      const isValid: boolean = await (this.$refs.elRenameForm as HTMLFormElement).validate();
      if (!isValid) return;
      const changes: AssetUpdateRequest = { name: this.newName, description: this.newDescription || null };
      this.$emit("rename", changes);
    },
    isNewNameUnique(val: string): boolean | string {
      if (val === this.existingName) return true;
      return (this.isNameUnique as CallableFunction)(val);
    },
  },
});
</script>
