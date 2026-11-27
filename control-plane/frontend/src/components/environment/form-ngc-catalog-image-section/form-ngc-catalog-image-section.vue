<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-ngc-image-section">
    <div class="row items-center q-gutter-x-md">
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
      />
    </div>
    <ngc-image-tag-selector
      v-if="showImageSelector && ngcCatalogRegistryId"
      :registry-id="ngcCatalogRegistryId"
      :credential-id="selectedCredentialId || undefined"
      @update:model-value="onImageValueChanged"
    />
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Types
import type { IScopeModel } from "@/models/global.model";
// Components
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import type { ICredentialSelectorModel } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { NgcImageTagSelector } from "@/components/environment/ngc-image-tag-selector";

export default defineComponent({
  name: "form-ngc-image-section",
  components: { RunaiCredentialSelectorBySource, NgcImageTagSelector },
  emits: ["update:model-value", "credential-changed"],
  props: {
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
      return !!this.selectedCredentialId;
    },
  },
  methods: {
    onCredentialSelectorChanged(model: ICredentialSelectorModel): void {
      this.credentialSelectorModel = model;
      this.selectedCredentialId = model.id || null;
      this.$emit("update:model-value", null);
      this.$emit("credential-changed", {
        credentialName: model.credentialName,
        credentialSource: model.source,
      });
    },
    onImageValueChanged(value: string | null): void {
      this.$emit("update:model-value", value);
    },
  },
});
</script>
