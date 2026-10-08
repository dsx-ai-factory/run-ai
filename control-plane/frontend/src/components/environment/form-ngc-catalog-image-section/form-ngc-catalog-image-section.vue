<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-ngc-image-section">
    <template v-if="supportAuthenticatedNgc || nimOnly">
      <runai-radio-options
        v-if="!nimOnly"
        class="q-mt-lg q-mb-md"
        :options="imagesTypeOptions"
        v-model="selectedImagesType"
        @update:model-value="onImagesTypeSelected"
        title="Set how to access the NGC catalog"
      />
      <div class="row items-center q-gutter-x-md" v-if="selectedImagesType === EImagesType.Authenticated">
        <runai-credential-selector-by-source
          class="col"
          v-model="credentialSelectorModel"
          @update:model-value="onCredentialSelectorChanged"
          :cluster-id="computedClusterId"
          :scope-model="scopeModel"
          :type-select-config="{
            hidden: false,
            allowedTypes: [ERunaiCredentialType.NgcApiKey],
            disable: true,
          }"
          :credential-select-config="{ required: true, showCreateNewBtn: true }"
          :name-resolver="nimOnly ? resolveCredentialName : undefined"
        />
      </div>
    </template>
    <ngc-image-tag-selector
      v-if="showImageSelector && ngcCatalogRegistryId"
      :registry-id="ngcCatalogRegistryId"
      :credential-id="selectedCredentialId || undefined"
      :nim-only="nimOnly"
      @update:model-value="onImageValueChanged"
    />
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Types
import type { ISelectOption, IScopeModel } from "@/models/global.model";
// Components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import type { ICredentialSelectorModel } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import {
  createResourceNameResolver,
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { NgcImageTagSelector } from "@/components/environment/ngc-image-tag-selector";

import { AssetKind, type CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

const imagePullNameResolver = createResourceNameResolver(AssetKind.DockerRegistry);
const runtimeNameResolver = createResourceNameResolver(AssetKind.GenericSecret);

export enum EImagesType {
  Guest = "guest",
  Authenticated = "authenticated",
}

export default defineComponent({
  name: "form-ngc-image-section",
  components: { RunaiRadioOptions, RunaiCredentialSelectorBySource, NgcImageTagSelector },
  emits: ["update:model-value", "credential-changed"],
  props: {
    supportAuthenticatedNgc: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    nimOnly: {
      type: Boolean,
      default: false,
    },
    scopeId: {
      type: String as PropType<string>,
      required: false,
    },
    scopeType: {
      type: String as PropType<string>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    ngcCatalogRegistryId: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      credentialSelectorModel: {
        source: ERunaiCredentialSource.MyCredentials,
        type: ERunaiCredentialType.NgcApiKey,
        credentialName: "",
        id: undefined,
      } as ICredentialSelectorModel,
      selectedCredentialId: null as string | null,
      runtimeCredentialNames: {} as Record<string, string>,
      imagesTypeOptions: [
        {
          value: EImagesType.Guest,
          label: "As a guest",
        },
        {
          value: EImagesType.Authenticated,
          label: "Authenticated",
        },
      ] as ISelectOption[],
      selectedImagesType: this.nimOnly ? EImagesType.Authenticated : EImagesType.Guest,
      EImagesType,
      ERunaiCredentialType,
      ERunaiCredentialSource,
    };
  },
  computed: {
    scopeModel(): IScopeModel {
      return {
        scope: null,
        projectId: this.scopeType === "project" ? Number(this.scopeId) : undefined,
        clusterId: this.computedClusterId || undefined,
        departmentId: this.scopeType === "department" ? this.scopeId : undefined,
      };
    },
    computedClusterId(): string {
      // Use the clusterId prop if provided, otherwise derive it from scopeId when scopeType is cluster
      return this.clusterId || (this.scopeType === "cluster" ? this.scopeId || "" : "");
    },
    showImageSelector(): boolean {
      return (
        (!this.nimOnly && this.selectedImagesType === EImagesType.Guest) ||
        (this.selectedImagesType === EImagesType.Authenticated && !!this.selectedCredentialId)
      );
    },
  },
  methods: {
    resolveCredentialName(credential: CredentialsListResponseEntry): string {
      this.runtimeCredentialNames[credential.meta.id] = runtimeNameResolver(credential);
      return imagePullNameResolver(credential);
    },
    onImagesTypeSelected(imagesType: EImagesType): void {
      if (this.nimOnly && imagesType === EImagesType.Guest) return;
      this.selectedImagesType = imagesType;
      this.selectedCredentialId = null;
      this.credentialSelectorModel = {
        source: ERunaiCredentialSource.MyCredentials,
        type: ERunaiCredentialType.NgcApiKey,
        credentialName: "",
        id: undefined,
      };
      this.$emit("update:model-value", null);
    },
    onCredentialSelectorChanged(model: ICredentialSelectorModel): void {
      this.credentialSelectorModel = model;
      this.selectedCredentialId = model.id || null;
      this.$emit("update:model-value", null);
      this.$emit("credential-changed", {
        credentialName: model.credentialName,
        credentialSource: model.source,
        runtimeCredentialName: this.nimOnly
          ? this.runtimeCredentialNames[model.id || ""] || model.credentialName
          : undefined,
      });
    },
    onImageValueChanged(value: string | null): void {
      this.$emit("update:model-value", value);
    },
  },
});
</script>
