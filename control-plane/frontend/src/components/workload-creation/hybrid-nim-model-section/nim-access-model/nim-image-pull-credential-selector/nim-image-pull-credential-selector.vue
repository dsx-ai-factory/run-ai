<template>
  <div class="nim-image-pull-credential-selector">
    <image-pull-credential-layout
      section-title="Set your credential for pulling the image"
      :show-info-banner="showInfoBox"
    >
      <runai-credential-selector-by-source
        :model-value="credentialSelectorModel"
        :cluster-id="clusterId"
        :scope-model="scopeModel"
        :source-select-config="nimSourceSelectConfig"
        :type-select-config="nimTypeSelectConfig"
        :credential-select-config="nimCredentialSelectConfig"
        :grid-config="nimGridConfig"
        @update:model-value="onCredentialSelectorUpdate"
      />
    </image-pull-credential-layout>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
  type ICredentialSelectorModel,
  type ICredentialSelectConfig,
  type IGridConfig,
  type ISourceSelectConfig,
  type ITypeSelectConfig,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

import { ImagePullCredentialLayout } from "../image-pull-credential-layout";

import type { IScopeModel } from "@/models/global.model";

import {
  ENimImagePullCredentialSelectorEvents,
  ENimImagePullCredentialType,
  type INimImagePullCredentialSelectorModel,
} from "./nim-image-pull-credential-selector.model";

function nimToRunaiCredentialType(type: ENimImagePullCredentialType): ERunaiCredentialType {
  return type === ENimImagePullCredentialType.DockerRegistry
    ? ERunaiCredentialType.DockerRegistry
    : ERunaiCredentialType.NgcApiKey;
}

function runaiToNimCredentialType(type?: ERunaiCredentialType): ENimImagePullCredentialType {
  if (type === ERunaiCredentialType.DockerRegistry) {
    return ENimImagePullCredentialType.DockerRegistry;
  }
  return ENimImagePullCredentialType.NgcApiKey;
}

export default defineComponent({
  name: "nim-image-pull-credential-selector",
  components: {
    ImagePullCredentialLayout,
    RunaiCredentialSelectorBySource,
  },
  emits: {
    [ENimImagePullCredentialSelectorEvents.UpdateModelValue]: (modelValue: INimImagePullCredentialSelectorModel) =>
      modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<INimImagePullCredentialSelectorModel>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    showCreateNewCredentialBtn: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  computed: {
    credentialSelectorModel(): ICredentialSelectorModel {
      return {
        source: ERunaiCredentialSource.MyCredentials,
        type: nimToRunaiCredentialType(this.modelValue.credentialType),
        credentialName: this.modelValue.credentialName,
      };
    },
    nimSourceSelectConfig(): ISourceSelectConfig {
      return {
        hidden: true,
        showPersonalCredentialsOption: true,
      };
    },
    nimTypeSelectConfig(): ITypeSelectConfig {
      return {
        hidden: false,
        allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.DockerRegistry],
      };
    },
    nimCredentialSelectConfig(): ICredentialSelectConfig {
      const showCreateForNgcOnly =
        this.showCreateNewCredentialBtn && this.modelValue.credentialType === ENimImagePullCredentialType.NgcApiKey;
      return {
        showCreateNewBtn: showCreateForNgcOnly,
      };
    },
    nimGridConfig(): IGridConfig {
      return {
        typeSelect: "col-4 q-mr-md",
        credentialListSelect: "col-7",
      };
    },
    showInfoBox(): boolean {
      return !this.modelValue.credentialName;
    },
  },
  methods: {
    onCredentialSelectorUpdate(model: ICredentialSelectorModel): void {
      this.$emit(ENimImagePullCredentialSelectorEvents.UpdateModelValue, {
        credentialType: runaiToNimCredentialType(model.type),
        credentialName: model.credentialName,
      });
    },
  },
});
</script>
