\
<template>
  <section class="docker-registry-create-form items-center q-pt-md">
    <runai-form-wrapper ref="elDockerRegistryForm" :form-state="dockerRegistryModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <credential-type-section type-name="Docker registry" />

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
          asset-kind="dockerRegistry"
          v-model="dockerRegistryModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="dockerRegistryModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <docker-registry-secret-section
            v-if="isScopeSelected"
            :secret-details="dockerRegistryModel.spec"
            :existing-secret-names="existingSecrets"
            :loading="loadingSecrets"
            :project-id="dockerRegistryModel.meta.projectId"
            @update-secret-details="dockerRegistryModel.spec = $event"
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
// Stores
import { useAppStore } from "@/stores/app.store";
// Components
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { CredentialTypeSection } from "@/components/credential/credential-type-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { DockerRegistrySecretSection } from "../docker-registry-secret-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { alertUtil } from "@/utils/alert.util";
// Models
import type { IScopeModel } from "@/models/global.model";
import type { IUIDockerRegistryCreationRequest } from "@/models/credential.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    CredentialTypeSection,
    RunaiNameSection,
    DockerRegistrySecretSection,
    RunaiExpansionWrapper,
    RunaiScopeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    credentialModel: {
      type: Object as PropType<IUIDockerRegistryCreationRequest>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      dockerRegistryModel: deepCopy(this.credentialModel) as IUIDockerRegistryCreationRequest,
      displayFormHint: false as boolean,
      timeoutId: null as null | number,
      elDockerRegistryForm: null as null | HTMLFormElement,
      appStore: useAppStore(),
      selectedCluster: "" as string,
      existingSecrets: null as null | string[],
      loadingSecrets: false as boolean,
      validating: false as boolean,
    };
  },
  mounted() {
    this.elDockerRegistryForm = this.$refs.elDockerRegistryForm as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.dockerRegistryModel.meta.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.dockerRegistryModel.meta.scope,
        projectId: this.dockerRegistryModel.meta.projectId,
        departmentId: this.dockerRegistryModel.meta.departmentId,
        clusterId: this.dockerRegistryModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.dockerRegistryModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.dockerRegistryModel.meta = scopeUtil.updateScope(this.dockerRegistryModel.meta, updatedScope);
      this.resetSecretDetails();
      this.loadExistingDockerRegistries(updatedScope);
    },
    resetSecretDetails(): void {
      this.dockerRegistryModel.spec.existingSecretName = null;
      this.dockerRegistryModel.spec.user = "";
      this.dockerRegistryModel.spec.password = "";
      this.dockerRegistryModel.spec.url = "";
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
        this.$emit("submit", this.dockerRegistryModel);
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to save Docker registry`));
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.elDockerRegistryForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeoutId && clearTimeout(this.timeoutId);
      this.timeoutId = window.setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
    async loadExistingDockerRegistries(updatedScope: IScopeModel): Promise<void> {
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
          const loadedSecrets = await credentialService.listExistingSecrets(
            listScope,
            scopeId,
            SecretType.DockerRegistry,
          );
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
    this.timeoutId && clearTimeout(this.timeoutId);
  },
});
</script>
<style lang="scss" scoped></style>
