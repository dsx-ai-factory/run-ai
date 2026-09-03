<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-name-section
    label="Template"
    :support-description="supportDescription"
    :model-value="modelValue"
    :scope-model="scopeModel"
    :description="description"
    :default-opened="defaultOpened"
    @update:model-value="updateTemplateName"
    @update-description="updateTemplateDescription"
  >
    <template #name-input>
      <q-input
        aid="workload-template-name-input"
        class="col-8"
        :model-value="modelValue"
        @update:model-value="$emit('update:modelValue', $event)"
        :original-name="originalName"
        :disable="disableNameInput"
        :rules="rules"
        maxlength="40"
        counter
        reactive-rules
        debounce="300"
        :input-tooltip="nameInputTooltip"
        no-error-icon
        :auto-focus="autoFocus"
        label="Name"
      />
    </template>
  </runai-name-section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Components
import { RunaiNameSection } from "@/components/common/runai-name-section";
// Models
import type { IScopeModel, ScopeTypeSystemExcluded } from "@/models/global.model";
import type { ValidationRule } from "quasar";
import { ScopeType } from "@/swagger-models/authorization-client";
import { useAuthStore } from "@/stores/auth.store";
import { isFlatTemplateNameUnique, isValidRunaiEntityNameFormat } from "@/common/form.validators";
import { validationRules } from "@/common/validation-rules";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "workload-template-name-section",
  components: {
    RunaiNameSection,
  },
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    supportDescription: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    description: {
      type: [String, null] as PropType<string | null>,
    },
    originalName: {
      type: String as PropType<string>,
      required: false,
    },
    disableNameInput: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    nameInputTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      autoFocus: false as boolean,
      isNameUnique: {} as ValidationRule,
    };
  },
  created() {
    this.autoFocus = !!this.$route.query.fromCopyId;
  },
  computed: {
    rules(): ValidationRule[] {
      if (!this.scopeModel.scope || this.originalName === this.modelValue)
        return [validationRules.validateNameNotEmpty, this.isValidFormat];
      return [validationRules.validateNameNotEmpty, this.isValidFormat, this.isNameUnique];
    },
    scopeModelForRequest(): { scopeType: ScopeType; scopeId: string } {
      if (!this.scopeModel.scope || this.scopeModel.scope === ScopeType.Tenant)
        return { scopeType: ScopeType.Tenant, scopeId: String(useAuthStore().tenant.id) };
      return {
        scopeType: this.scopeModel.scope as ScopeType,
        scopeId: String(this.scopeId),
      };
    },
    scopeId(): string {
      switch (this.scopeModel.scope) {
        case ScopeType.Department:
          return String(this.scopeModel.departmentId);
        case ScopeType.Project:
          return String(this.scopeModel.projectId);
        case ScopeType.Cluster:
          return String(this.scopeModel.clusterId);
        default:
          return "";
      }
    },
  },
  methods: {
    isValidFormat(val: string): boolean | string {
      return isValidRunaiEntityNameFormat(val) || errorMessages.VALID_FORMAT;
    },
    updateTemplateName(val: string): void {
      this.$emit("update:modelValue", val);
    },
    updateTemplateDescription(val: string | null): void {
      this.$emit("update:description", val);
    },
  },
  watch: {
    scopeModelForRequest: {
      handler(): void {
        this.isNameUnique = isFlatTemplateNameUnique(
          this.scopeModelForRequest.scopeType as ScopeTypeSystemExcluded,
          this.scopeModelForRequest.scopeId,
        );
      },
      immediate: true,
      deep: true,
    },
  },
});
</script>

<style scoped></style>
