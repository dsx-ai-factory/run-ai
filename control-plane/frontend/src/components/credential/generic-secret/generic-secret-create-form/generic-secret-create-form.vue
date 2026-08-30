<template>
  <section class="generic-secret-create-form column items-center q-pt-md">
    <runai-form-wrapper class="generic-secret-form" :form-state="genericSecretModel" ref="elForm" :enter-callback="save">
      <runai-expansion-wrapper>
        <credential-type-section type-name="Generic secret" />

        <runai-scope-section
          kind-of-asset="credential"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopes"
          :forced-disabled-scopes="disabledScopes"
          :scopes-supported-versions="scopesSupportedVersions"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
        />

        <runai-name-section
          label="Credential"
          asset-kind="genericSecret"
          v-model="genericSecretModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="genericSecretModel.meta.description"
          @update-description="updateDescription"
        />

        <transition name="fade">
          <generic-secret-secret-section
            v-if="isScopeSelected"
            :secret-details="genericSecretModel.spec"
            :existing-secret-names="existingSecrets"
            :loading="loadingSecrets"
            @update-secret-details="genericSecretModel.spec = $event"
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
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { CredentialTypeSection } from "@/components/credential/credential-type-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { GenericSecretSecretSection } from "../generic-secret-secret-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IUIGeneralCredentialCreationRequest } from "@/models/credential.model";
import type { IScopeModel, TTreeDisabledScopes, TTreeSupportedVersions } from "@/models/global.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET } from "@/common/version.constant";
// stores
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import { alertUtil } from "@/utils/alert.util";
import { useAppStore } from "@/stores/app.store";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    CredentialTypeSection,
    RunaiNameSection,
    RunaiScopeSection,
    GenericSecretSecretSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    credentialModel: {
      type: Object as PropType<IUIGeneralCredentialCreationRequest>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  data() {
    return {
      genericSecretModel: deepCopy(this.credentialModel) as IUIGeneralCredentialCreationRequest,
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
  computed: {
    disabledScopes(): TTreeDisabledScopes {
      return {
        [ScopeType.Tenant]: true,
      };
    },
    isScopeSelected(): boolean {
      return !!this.genericSecretModel.meta.scope;
    },
    scopesSupportedVersions(): TTreeSupportedVersions {
      return {
        [ScopeType.Tenant]: MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET,
        [ScopeType.Department]: MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET,
        [ScopeType.Cluster]: MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET,
        [ScopeType.Project]: MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET,
      };
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.genericSecretModel.meta.scope,
        projectId: this.genericSecretModel.meta.projectId,
        departmentId: this.genericSecretModel.meta.departmentId,
        clusterId: this.genericSecretModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.genericSecretModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.genericSecretModel.meta = scopeUtil.updateScope(this.genericSecretModel.meta, updatedScope);
      this.resetSecretDetails();
      this.loadExistingGenericSecrets(updatedScope);
    },
    resetSecretDetails(): void {
      this.genericSecretModel.spec.existingSecretName = "";
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
        this.$emit("submit", this.genericSecretModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.$refs["elForm"] as HTMLFormElement).validate();
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
    async loadExistingGenericSecrets(updatedScope: IScopeModel): Promise<void> {
      try {
        this.loadingSecrets = true;
        const { scope, departmentId, projectId, clusterId } = updatedScope;
        let scopeId: string;
        const listScope: string | null = scope;
        if (scope === ScopeType.Cluster && clusterId) {
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
          const loadedSecrets = await credentialService.listExistingSecrets(listScope, scopeId, SecretType.Generic);
          this.existingSecrets = loadedSecrets.map((secret) => secret.name);
        }
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to load existing credential`));
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
