<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="hugging-face-model-and-access">
    <hugging-face-model-selector
      class="col-6 q-mb-md"
      :model-value="modelValue.modelName"
      @update:model-value="onModelNameChanged"
      :disable="disableModelSelection"
      @is-valid="onModelNameValidated"
    />

    <runai-section v-if="modelValue.modelName" title="Set the access to the Hugging Face" class="q-mb-lg">
      <q-checkbox
        aid="hf-access-from-storage"
        v-model="accessFromStorage"
        @update:model-value="onAccessFromStorageChanged"
        label="Do not access Hugging Face. Load the model profile from a local model-store"
        class="offset-left-8"
      />
      <runai-tooltip tooltip-text="Go to the Model store section to set this" tooltip-position="right" />
    </runai-section>

    <hugging-face-access-section
      v-if="requiredToken && !accessFromStorage"
      :model-value="accessModel"
      @update:model-value="onAccessModelChanged"
      :cluster-id="clusterId"
      :scope-model="scopeModel"
      :is-access-required="isAccessRequired"
      :show-create-new-personal-credentials-btn="showCreateNewPersonalCredentialsBtn"
      :show-personal-credentials-option="showPersonalCredentialsOption"
      :model-name="modelValue.modelName"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { HuggingFaceModelSelector } from "../hugging-face-model-selector";
import { HuggingFaceAccessSection } from "../hugging-face-access-section";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import type { IHuggingFaceModelAndAccessModel } from "./hugging-face-model-and-access.model";
import type { IHuggingFaceAccessSectionModel } from "../hugging-face-access-section/hugging-face-access-section.model";
import type { IHuggingFaceModelSectionEvent } from "../hugging-face-model-selector/hugging-face-model-selector.model";
import type { IScopeModel } from "@/models/global.model";
import type { HuggingFaceModelInfo } from "@/swagger-models/assets-service-client";
import {
  ETokenProvisionMethods,
  type IHuggingFaceAccessSectionModel as AccessModel,
} from "../hugging-face-access-section/hugging-face-access-section.model";

export default defineComponent({
  name: "HuggingFaceModelAndAccess",
  components: {
    HuggingFaceModelSelector,
    HuggingFaceAccessSection,
    RunaiSection,
    RunaiTooltip,
  },
  emits: {
    update: (payload: IHuggingFaceModelAndAccessModel) => payload,
  },
  props: {
    modelValue: {
      type: Object as PropType<IHuggingFaceModelAndAccessModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    disableModelSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isAccessRequired: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showCreateNewPersonalCredentialsBtn: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showPersonalCredentialsOption: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      requiredToken: false,
      accessFromStorage: false,
    };
  },
  created() {
    this.accessFromStorage = this.modelValue.accessFromStorage ?? false;
  },
  computed: {
    accessModel(): AccessModel {
      return {
        modelName: this.modelValue.modelName,
        token: this.modelValue.token,
        selectedCredentials: this.modelValue.selectedCredentials,
        selectedTokenProvisionMethod: this.modelValue.selectedTokenProvisionMethod ?? ETokenProvisionMethods.FromSecret,
        credentialsSource: this.modelValue.credentialsSource,
      };
    },
  },
  methods: {
    onModelNameValidated(validationEvent: IHuggingFaceModelSectionEvent): void {
      this.requiredToken = validationEvent.gated;
    },
    onModelNameChanged(huggingFaceModelInfo: HuggingFaceModelInfo): void {
      this.requiredToken = false;
      this.emitUpdate({ modelName: huggingFaceModelInfo.modelName });
      this.$nextTick(() => {
        this.requiredToken = !!huggingFaceModelInfo.gated;
      });
    },
    onAccessModelChanged(accessModel: IHuggingFaceAccessSectionModel): void {
      this.emitUpdate({
        token: accessModel.token,
        selectedCredentials: accessModel.selectedCredentials,
        credentialsSource: accessModel.credentialsSource,
        selectedTokenProvisionMethod: accessModel.selectedTokenProvisionMethod,
      });
    },
    onAccessFromStorageChanged(accessFromStorage: boolean): void {
      this.accessFromStorage = accessFromStorage;
      this.emitUpdate({ accessFromStorage });
    },
    emitUpdate(partial: Partial<IHuggingFaceModelAndAccessModel>): void {
      const accessFromStorage = partial.accessFromStorage ?? this.accessFromStorage;
      this.$emit("update", {
        ...this.modelValue,
        ...partial,
        accessFromStorage,
      });
    },
  },
});
</script>
