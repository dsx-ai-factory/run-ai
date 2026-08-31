<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <runai-name-validation
      ref="nameValidationCmp"
      :label="label"
      :stack-label="stackLabel"
      :class="`${assetKind}-name-input`"
      :model-value="modelValue"
      @update:model-value="$emit('update:modelValue', $event)"
      :rules="rules"
      @invalid="nameAlreadyExist = $event"
      debounce="300"
      :aid="`${assetKind}-name-input`"
      :auto-focus="autoFocus"
      :disable="disable"
      :placeholder="placeholder"
    />
    <q-tooltip v-if="inputTooltip && $refs.nameValidationCmp" :target="`.${assetKind}-name-input`">
      {{ inputTooltip }}
    </q-tooltip>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import type { ValidationRule } from "quasar";
import type { AssetKind } from "@/swagger-models/assets-service-client";
import type { IScopeModel, IValidationRule } from "@/models/global.model";
import type { IAssetNameValidationOptions } from "@/models/asset.model";
import { Scope } from "@/swagger-models/assets-service-client";
import { isAssetNameUnique, isValidRunaiEntityNameFormat } from "@/common/form.validators";
import { validationRules } from "@/common/validation-rules";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "runai-asset-name-input",
  components: {
    RunaiNameValidation,
  },
  emits: ["update:modelValue"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    assetKind: {
      type: String as PropType<AssetKind>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    originalName: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    inputTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    autoFocus: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    customExistingNameValidation: {
      type: Function as PropType<IValidationRule>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      nameAlreadyExist: false as boolean,
      isNameUnique: {} as ValidationRule,
    };
  },
  computed: {
    rules(): ValidationRule[] {
      if (!this.scopeModel.scope || this.originalName === this.modelValue)
        return [validationRules.validateNameNotEmpty, this.isValidFormat];
      return [validationRules.validateNameNotEmpty, this.isValidFormat, this.isNameUnique];
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
        case Scope.Department:
          return String(this.scopeModel.departmentId);
        case Scope.Project:
          return String(this.scopeModel.projectId);
        case Scope.Cluster:
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
    updateAssetNameValidation(): void {
      if (this.customExistingNameValidation) {
        this.isNameUnique = this.customExistingNameValidation;
      } else {
        this.isNameUnique = isAssetNameUnique(this.assetKind, this.assetNameOptions);
      }
    },
  },
  watch: {
    assetNameOptions: {
      handler(): void {
        this.updateAssetNameValidation();
      },
      immediate: true,
    },
  },
});
</script>
