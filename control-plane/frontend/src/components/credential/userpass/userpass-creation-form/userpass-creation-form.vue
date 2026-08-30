<template>
  <section class="userpass-create-form column items-center q-pt-md">
    <runai-form-wrapper class="userpass-form" :form-state="userpassModel" ref="elUserpassForm" :enter-callback="save">
      <runai-expansion-wrapper>
        <credential-type-section type-name="Username & password" />

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
          asset-kind="password"
          v-model="userpassModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="userpassModel.meta.description"
          @update-description="updateDescription"
        />

        <transition name="fade">
          <userpass-secret-section
            v-if="isScopeSelected"
            :secret-details="userpassModel.spec"
            :existing-secret-names="existingSecrets"
            :loading="loadingSecrets"
            :project-id="userpassModel.meta.projectId"
            @update-secret-details="userpassModel.spec = $event"
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
import { UserpassSecretSection } from "../userpass-secret-section";
import { CredentialTypeSection } from "../../credential-type-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IUIPasswordCreationRequest } from "@/models/credential.model";
import type { IScopeModel } from "@/models/global.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
// stores
import { alertUtil } from "@/utils/alert.util";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import { useAppStore } from "@/stores/app.store";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    CredentialTypeSection,
    RunaiNameSection,
    UserpassSecretSection,
    RunaiScopeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel", "scope-changed"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    credentialModel: {
      type: Object as PropType<IUIPasswordCreationRequest>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      userpassModel: deepCopy(this.credentialModel) as IUIPasswordCreationRequest,
      elUserpassForm: null as HTMLFormElement | null,
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
    this.elUserpassForm = this.$refs["elUserpassForm"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.userpassModel.meta.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.userpassModel.meta.scope,
        projectId: this.userpassModel.meta.projectId,
        departmentId: this.userpassModel.meta.departmentId,
        clusterId: this.userpassModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.userpassModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.userpassModel.meta = scopeUtil.updateScope(this.userpassModel.meta, updatedScope);
      this.resetSecretDetails();
      this.loadExistingUserpass(updatedScope);
    },
    resetSecretDetails(): void {
      this.userpassModel.spec.existingSecretName = null;
      this.userpassModel.spec.user = "";
      this.userpassModel.spec.password = "";
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
        this.$emit("submit", this.userpassModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.elUserpassForm as HTMLFormElement).validate();
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
    async loadExistingUserpass(updatedScope: IScopeModel): Promise<void> {
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
          const loadedSecrets = await credentialService.listExistingSecrets(listScope, scopeId, SecretType.Password);
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
