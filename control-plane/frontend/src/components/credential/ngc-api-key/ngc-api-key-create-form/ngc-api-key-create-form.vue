<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

\
<template>
  <section class="ngc-api-key-create-form items-center q-pt-md">
    <runai-form-wrapper ref="elNgcApiKeyForm" :form-state="ngcApiKeyModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <credential-type-section :type-name="credentialTextForUI.ngcApiKey" />

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
          asset-kind="ngcApiKey"
          v-model="ngcApiKeyModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="ngcApiKeyModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <runai-expansion-item
            class="ngc-api-key-secret-section"
            default-opened
            label="Secret"
            should-validate
            v-if="isScopeSelected"
          >
            <template #subheader>
              <section class="summary row">
                {{ secretSummary }}
              </section>
            </template>
            <runai-section title="Set a key for authentication to NGC services">
              <ngc-api-key-new-section
                :secret-details="ngcApiKeyModel.spec"
                :project-id="ngcApiKeyModel.meta.projectId"
                @update-secret-details="ngcApiKeyModel.spec = $event"
              />
            </runai-section>
          </runai-expansion-item>
        </transition>
      </runai-expansion-wrapper>
      <section class="row items-center q-mt-md">
        <q-field
          class="col-4 form-hint no-padding"
          :model-value="displayFormHint"
          :rules="[validationRules.validateFormIncomplete]"
        ></q-field>
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
// Components
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { CredentialTypeSection } from "@/components/credential/credential-type-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NgcApiKeyNewSection } from "@/components/credential/ngc-api-key/ngc-api-key-new-section";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSection } from "@/components/common/runai-section";
// Services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { alertUtil } from "@/utils/alert.util";
import { validationRules } from "@/common/validation-rules";
// Models
import type { IScopeModel } from "@/models/global.model";
import { credentialTextForUI, type IUINgcApiKeyCreationRequest } from "@/models/credential.model";
import { type PermittedScopes } from "@/swagger-models/authorization-client";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    CredentialTypeSection,
    RunaiNameSection,
    RunaiExpansionWrapper,
    RunaiScopeSection,
    RunaiBtn,
    NgcApiKeyNewSection,
    RunaiExpansionItem,
    RunaiSection,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    credentialModel: {
      type: Object as PropType<IUINgcApiKeyCreationRequest>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      ngcApiKeyModel: deepCopy(this.credentialModel) as IUINgcApiKeyCreationRequest,
      displayFormHint: false as boolean,
      timeoutId: null as null | number,
      elNgcApiKeyForm: null as null | HTMLFormElement,
      selectedCluster: "" as string,
      validating: false as boolean,
      validationRules,
      credentialTextForUI,
    };
  },
  mounted() {
    this.elNgcApiKeyForm = this.$refs.elNgcApiKeyForm as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.ngcApiKeyModel.meta.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.ngcApiKeyModel.meta.scope,
        projectId: this.ngcApiKeyModel.meta.projectId,
        departmentId: this.ngcApiKeyModel.meta.departmentId,
        clusterId: this.ngcApiKeyModel.meta.clusterId,
      };
    },
    secretSummary(): string {
      return this.ngcApiKeyModel.spec.ngc_api_key || "None";
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.ngcApiKeyModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.ngcApiKeyModel.meta = scopeUtil.updateScope(this.ngcApiKeyModel.meta, updatedScope);
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
        this.$emit("submit", this.ngcApiKeyModel);
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to save NGC API key`));
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.elNgcApiKeyForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeoutId && clearTimeout(this.timeoutId);
      this.timeoutId = window.setTimeout(() => (this.displayFormHint = false), 15000);
    },
  },
  unmounted() {
    this.timeoutId && clearTimeout(this.timeoutId);
  },
});
</script>
<style lang="scss" scoped></style>
