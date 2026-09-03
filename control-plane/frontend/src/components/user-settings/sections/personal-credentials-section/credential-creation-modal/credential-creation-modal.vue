<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-delete-modal">
    <runai-base-dialog :model-value="true" @close="onCancel" size="width-md" no-padding>
      <template #header>New {{ credentialLabel }} Credential</template>

      <template #body>
        <q-form ref="formEl">
          <div class="q-px-lg q-py-md column">
            <section class="name-description-section row">
              <asset-name-input
                label="Credential name"
                stack-label
                class="col-8"
                v-model="meta.name"
                :scope-model="scopeModel"
                :asset-kind="credentialType"
                auto-focus
              />
              <runai-description-input
                class="col-12"
                v-model="meta.description"
                custom-aid="credential-description-input"
              />
            </section>

            <docker-registry-new-secret-section
              v-if="dockerRegistryModel && credentialType === AssetKind.DockerRegistry"
              :secret-details="dockerRegistryModel.spec"
              @update-secret-details="dockerRegistryModel.spec = $event"
            />
            <generic-secret-key-value-section
              v-if="genericSecretModel && credentialType === AssetKind.GenericSecret"
              :secret-details="genericSecretModel.spec"
              @update-secret-details="genericSecretModel.spec = $event"
              :support-multiple-items="supportMultipleGenericSecretItems"
              :disabled-keys="disabledKeysList"
            />
            <ngc-api-key-new-section
              v-if="ngcApiKeyModel && credentialType === AssetKind.NgcApiKey"
              input-class="col-12"
              :secret-details="ngcApiKeyModel.spec"
              @update-secret-details="ngcApiKeyModel.spec = $event"
            />
          </div>
        </q-form>
      </template>

      <template #footer>
        <q-btn label="Cancel" color="primary" flat @click="onCancel" aid="credential-modal-cancel-button" />
        <q-btn
          label="Create credential"
          color="primary"
          :loading="loading"
          @click="onCreate"
          aid="credential-modal-create-button"
        />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
// stores
import { useCredentialStore } from "@/stores/credential.store";
// components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { DockerRegistryNewSecretSection } from "@/components/credential/docker-registry/docker-registry-secret-section/docker-registry-new-secret-section";
import { GenericSecretKeyValueSection } from "@/components/credential/generic-secret/generic-secret-secret-section/generic-secret-key-value-section";
import { AssetNameInput } from "@/components/common/asset-name-input";
import { RunaiDescriptionInput } from "@/components/common/runai-description-input";
import { NgcApiKeyNewSection } from "@/components/credential/ngc-api-key/ngc-api-key-new-section";
// models and types
import type { TCredentialKinds } from "@/models/credential.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import { Action } from "@/swagger-models/authorization-client";
import type { QForm } from "quasar";
import type {
  DockerRegistryCreationRequest,
  GenericSecretCreationRequest,
  KeyValuePair,
  NgcApiKeyCreationRequest,
} from "@/swagger-models/assets-service-client";
import type { IScopeModel } from "@/models/global.model";
import type { IUIAssetCreationMeta } from "@/models/global.model";

export default defineComponent({
  name: "credential-creation-modal",
  components: {
    RunaiBaseDialog,
    AssetNameInput,
    RunaiDescriptionInput,
    DockerRegistryNewSecretSection,
    GenericSecretKeyValueSection,
    NgcApiKeyNewSection,
  },
  props: {
    credentialType: {
      type: String as PropType<TCredentialKinds>,
      required: true,
    },
    loading: {
      type: Boolean,
      required: true,
    },
    supportMultipleGenericSecretItems: {
      type: Boolean,
      required: false,
      default: true,
    },
    defaultGenericSecretItems: {
      type: Array as PropType<KeyValuePair[]>,
      required: false,
      default: () => [{ key: "", value: "" }],
    },
    disabledKeysList: {
      type: Array as PropType<string[]>,
      required: false,
      default: () => [],
    },
  },
  data() {
    return {
      credentialStore: useCredentialStore(),

      meta: {
        scope: Scope.Tenant,
        name: "",
        description: null,
        private: true,
      } as IUIAssetCreationMeta,

      dockerRegistryModel: null as DockerRegistryCreationRequest | null,
      genericSecretModel: null as GenericSecretCreationRequest | null,
      ngcApiKeyModel: null as NgcApiKeyCreationRequest | null,

      AssetKind,
    };
  },
  computed: {
    scopeModel(): IScopeModel {
      return {
        scope: Scope.Tenant,
      };
    },
    allowedCreateScopes(): PermittedScopes {
      return this.credentialStore.actionPermissionsByScopes[Action.Create] || {};
    },
    credentialLabel(): string {
      switch (this.credentialType) {
        case AssetKind.DockerRegistry:
          return "Docker Registry";
        case AssetKind.GenericSecret:
          return "Generic Secret";
        case AssetKind.Password:
          return "Username & Password";
        case AssetKind.AccessKey:
          return "Access Key";
        case AssetKind.NgcApiKey:
          return "NGC API Key";
        default:
          return this.credentialType;
      }
    },
    relevantModel(): DockerRegistryCreationRequest | GenericSecretCreationRequest | NgcApiKeyCreationRequest | null {
      switch (this.credentialType) {
        case AssetKind.DockerRegistry:
          return this.dockerRegistryModel;
        case AssetKind.GenericSecret:
          return this.genericSecretModel;
        case AssetKind.NgcApiKey:
          return this.ngcApiKeyModel;
        default:
          return null;
      }
    },
  },
  methods: {
    onCancel() {
      this.$emit("cancel");
    },
    async onCreate() {
      const isValid = await (this.$refs.formEl as QForm)?.validate();
      if (!isValid) return;

      this.$emit("create", {
        ...this.relevantModel,
        meta: this.meta,
      });
    },
  },
  watch: {
    credentialType: {
      handler(newVal) {
        this.dockerRegistryModel = null;
        this.ngcApiKeyModel = null;
        this.genericSecretModel = null;

        if (newVal === AssetKind.DockerRegistry) {
          this.dockerRegistryModel = credentialService.getEmptyDockerRegistryModel();
        } else if (newVal === AssetKind.GenericSecret) {
          this.genericSecretModel = credentialService.getEmptyGenericSecretCredentialModel();
          this.genericSecretModel.spec.existingSecretName = null;
          this.genericSecretModel.spec.keyValuePairs = this.defaultGenericSecretItems;
        } else if (newVal === AssetKind.NgcApiKey) {
          this.ngcApiKeyModel = credentialService.getEmptyNgcApiKeyModel();
        }
      },
      immediate: true,
    },
  },
});
</script>
