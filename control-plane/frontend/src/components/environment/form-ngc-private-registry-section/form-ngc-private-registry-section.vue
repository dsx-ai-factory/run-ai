<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-ngc-private-registry-section">
    <div class="row q-mb-md">
      <runai-select
        class="col-6"
        stack-label
        label="Registry"
        placeholder="Select..."
        no-option-text="No registries found"
        :rules="[isRegistrySelected]"
        v-model="selectedRegistry"
        @update:model-value="onRegistrySelected"
        emit-value
        no-error-icon
        :options="registryOptions"
      />
    </div>
    <transition name="fade" class="row q-mb-md">
      <runai-credential-selector-by-source
        v-if="selectedRegistry"
        class="col"
        v-model="credentialSelectorModel"
        :cluster-id="computedClusterId"
        :type-select-config="credentialSelectConfig"
        @update:model-value="onCredentialSelectorChanged"
        :scope-model="scopeModel"
        :credential-select-config="{ showCreateNewBtn: true }"
      />
    </transition>

    <transition name="fade">
      <ngc-image-tag-selector
        v-if="credentialSelectorModel.id && selectedRegistry"
        :registry-id="selectedRegistry"
        :credential-id="credentialSelectorModel.id"
        @update:model-value="onImageValueChanged"
      />
    </transition>
  </section>
</template>

<script lang="ts">
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import type { RegistryIntegration } from "@/swagger-models/catalog-service-client";
import { defineComponent, type PropType } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import {
  ERunaiCredentialSource,
  ERunaiCredentialType,
  type ICredentialSelectorModel,
  type ITypeSelectConfig,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { NgcImageTagSelector } from "@/components/environment/ngc-image-tag-selector";

export default defineComponent({
  name: "form-ngc-private-registry-section",
  components: {
    RunaiSelect,
    RunaiCredentialSelectorBySource,
    NgcImageTagSelector,
  },
  emits: ["update:model-value", "credential-changed"],
  props: {
    registries: {
      type: Array as PropType<RegistryIntegration[]>,
      required: true,
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
  },
  data() {
    return {
      selectedRegistry: null as string | null,
      credentialSelectorModel: {
        source: ERunaiCredentialSource.SharedSecret,
        type: ERunaiCredentialType.NgcApiKey,
        credentialName: "",
        id: undefined,
      } as ICredentialSelectorModel,
    };
  },
  computed: {
    registryOptions(): ISelectOption[] {
      return this.registries.map((registry) => {
        return {
          label: registry.name,
          value: registry.id,
        };
      });
    },
    computedClusterId(): string {
      // Use the clusterId prop if provided, otherwise derive it from scopeId when scopeType is cluster
      return this.clusterId || (this.scopeType === "cluster" ? this.scopeId || "" : "");
    },
    credentialSelectConfig(): ITypeSelectConfig {
      return {
        hidden: false,
        allowedTypes: [ERunaiCredentialType.NgcApiKey],
        disable: true,
      };
    },
    scopeModel(): IScopeModel {
      return {
        scope: null,
        projectId: this.scopeType === "project" ? Number(this.scopeId) : undefined,
        clusterId: this.computedClusterId || undefined,
        departmentId: this.scopeType === "department" ? this.scopeId : undefined,
      };
    },
  },
  methods: {
    onCredentialSelectorChanged(model: ICredentialSelectorModel): void {
      this.credentialSelectorModel = model;
      this.$emit("update:model-value", null);
      this.$emit("credential-changed", { credentialName: model.credentialName, credentialSource: model.source });
    },
    onImageValueChanged(value: string | null): void {
      this.$emit("update:model-value", value);
    },
    isRegistrySelected(val: string): string | boolean {
      return !!val || "Select a registry";
    },
    onRegistrySelected(): void {
      this.credentialSelectorModel.id = undefined;
      this.credentialSelectorModel.credentialName = "";
    },
  },
});
</script>
