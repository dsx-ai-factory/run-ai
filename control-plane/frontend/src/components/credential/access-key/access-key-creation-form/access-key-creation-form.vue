<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="access-key-create-form column items-center q-pt-md">
    <runai-form-wrapper
      class="access-key-form"
      :form-state="accessKeyModel"
      ref="elAccessKeyForm"
      :enter-callback="save"
    >
      <runai-expansion-wrapper>
        <credential-type-section type-name="Access key" />
        <runai-scope-section
          kind-of-asset="credential"
          :read-only="false"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopes"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
        />

        <runai-name-section
          label="Credential"
          asset-kind="accessKey"
          v-model="accessKeyModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="accessKeyModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <access-key-secret-section
            v-if="isScopeSelected"
            :secret-details="accessKeyModel.spec"
            :existing-secret-names="existingSecrets"
            :loading="loadingSecrets"
            :project-id="accessKeyModel.meta.projectId"
            @update-secret-details="accessKeyModel.spec = $event"
          />
        </transition>
      </runai-expansion-wrapper>

      <section class="row items-center q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            aid="submit-btn"
            color="primary"
            @click="save"
            :loading="validating || submitting"
            label="create credential"
            type="submit"
          />
        </div>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { CredentialTypeSection } from "../../credential-type-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { AccessKeySecretSection } from "../access-key-secret-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// models
import type { IScopeModel } from "@/models/global.model";
import type { IUIAccessKeyCreationRequest } from "@/models/credential.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
// stores
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { alertUtil } from "@/utils/alert.util";
import { useAppStore } from "@/stores/app.store";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    CredentialTypeSection,
    RunaiNameSection,
    AccessKeySecretSection,
    RunaiScopeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    credentialModel: {
      type: Object as PropType<IUIAccessKeyCreationRequest>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      accessKeyModel: deepCopy(this.credentialModel) as IUIAccessKeyCreationRequest,
      elAccessKeyForm: null as HTMLFormElement | null,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      appStore: useAppStore(),
      selectedCluster: "" as string,
      existingSecrets: null as null | string[],
      loadingSecrets: false as boolean,
      validating: false as boolean,
    };
  },
  mounted() {
    this.elAccessKeyForm = this.$refs["elAccessKeyForm"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.accessKeyModel.meta.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.accessKeyModel.meta.scope,
        projectId: this.accessKeyModel.meta.projectId,
        departmentId: this.accessKeyModel.meta.departmentId,
        clusterId: this.accessKeyModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.accessKeyModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.accessKeyModel.meta = scopeUtil.updateScope(this.accessKeyModel.meta, updatedScope);
      this.resetSecretDetails();
      this.loadExistingAccessKeys(updatedScope);
    },
    resetSecretDetails(): void {
      this.accessKeyModel.spec.existingSecretName = null;
      this.accessKeyModel.spec.accessKeyId = "";
      this.accessKeyModel.spec.secretAccessKey = "";
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success = await this.validate();
        if (!success) {
          this.showHint();
          return;
        }

        this.$emit("submit", this.accessKeyModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.elAccessKeyForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
    async loadExistingAccessKeys(updatedScope: IScopeModel): Promise<void> {
      try {
        this.loadingSecrets = true;
        const { scope, departmentId, projectId, clusterId } = updatedScope;
        let scopeId: string;
        let listScope: string | null = scope;
        if (scope === ScopeType.Tenant) {
          listScope = ScopeType.Cluster;
          scopeId = this.selectedCluster;
        } else if (scope === ScopeType.Cluster && clusterId) {
          scopeId = clusterId;
        } else if (scope === ScopeType.Department && departmentId) {
          scopeId = departmentId;
        } else if (scope === ScopeType.Project && projectId) {
          scopeId = String(projectId);
        } else {
          this.existingSecrets = [];
          this.loadingSecrets = false;
          return;
        }
        if (listScope) {
          const loadedSecrets = await credentialService.listExistingSecrets(listScope, scopeId, SecretType.AccessKey);
          this.existingSecrets = loadedSecrets.map((secret) => secret.name);
        }
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to load existing secrets`));
        this.appStore.setFallback(true);
        console.error(e);
      } finally {
        this.loadingSecrets = false;
      }
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>

<style lang="scss" scoped></style>
