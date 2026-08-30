<template>
  <section class="generic-format-page column items-center q-pt-md">
    <runai-form-wrapper class="generic-format-form" :form-state="genericFormatModel" ref="elForm">
      <runai-expansion-wrapper>
        <credential-type-section type-name="Generic secret" />

        <runai-scope-section
          kind-of-asset="credential"
          :read-only="!!$route.query?.scope"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopes"
          selected-cluster-id=""
        />

        <runai-name-section
          label="Credential"
          asset-kind="genericSecret"
          v-model="genericFormatModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="genericFormatModel.meta.description"
          @update-description="genericFormatModel.meta.description = $event"
        />

        <transition name="fade">
          <generic-format-secret-section
            v-if="isScopeSelected"
            :existing-secrets="existingSecrets"
            :loading="loadingSecrets"
            :secret-details="genericFormatModel.spec"
            :disabled-formats="disabledFormats"
            @update-secret-details="genericFormatModel.spec = $event"
          />
        </transition>
      </runai-expansion-wrapper>

      <runai-form-footer
        @cancel="onCancel"
        @save="onSubmit"
        save-btn-aid="submit-btn"
        :is-form-valid="isFormValid"
        save-btn-label="create credential"
        :loading="validating || submitting"
      />
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { alertUtil } from "@/utils/alert.util";

// cmps
import { CredentialTypeSection } from "@/components/credential/credential-type-section";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { GenericFormatSecretSection } from "@/components/credential/generic-secret/generic-format-secret-section";
import { RunaiFormFooter } from "@/components/common/runai-form-footer";
// models
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { IScopeModel } from "@/models/global.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useCredentialStore } from "@/stores/credential.store";
import { useClusterStore } from "@/stores/cluster.store";

// utils
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { scopeUtil } from "@/utils/scope.util";
import type { IUIGenericSecretCreationRequest, TGenericFormats } from "@/models/credential.model";
import { SECRET_TYPE_TO_ASSET_KIND, TYPE_SSH_KEY } from "@/models/credential.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import { MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET, MIN_CLUSTER_VERSION_FOR_GIT_SSH } from "@/common/version.constant";

export default defineComponent({
  name: "generic-format",
  components: {
    CredentialTypeSection,
    RunaiExpansionWrapper,
    RunaiFormWrapper,
    RunaiScopeSection,
    RunaiNameSection,
    GenericFormatSecretSection,
    RunaiFormFooter,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      credentialStore: useCredentialStore(),
      clusterStore: useClusterStore(),
      submitting: false as boolean,
      validating: false as boolean,
      isFormValid: true as boolean,

      existingSecrets: null as null | SecretListInfo[],
      loadingSecrets: false as boolean,

      parentClusterUuid: "" as string,
      genericFormatModel: credentialService.getEmptyGenericFormatSecretModel() as IUIGenericSecretCreationRequest,
    };
  },
  async created() {
    if (this.$route.query.scope) {
      this.genericFormatModel.meta = assetsUtil.setAssetMetaFromQuery(
        this.genericFormatModel.meta,
        this.$route.query as unknown as IScopeModel,
        this.authStore.orgUnitList,
      ) as IUIGenericSecretCreationRequest["meta"];
      this.loadExistingGenericSecrets(this.scopeModel);
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    scopeModel(): IScopeModel {
      return {
        scope: this.genericFormatModel.meta.scope,
        projectId: this.genericFormatModel.meta.projectId,
        departmentId: this.genericFormatModel.meta.departmentId,
        clusterId: this.genericFormatModel.meta.clusterId,
      };
    },
    allowedScopes(): PermittedScopes {
      return this.credentialStore.actionPermissionsByScopes[Action.Create];
    },
    isScopeSelected(): boolean {
      return !!this.genericFormatModel.meta.scope;
    },
    currentFormatType(): TGenericFormats | null {
      if (this.genericFormatModel.spec.existingSecretName) {
        const secretType = this.existingSecrets?.find((s) => s.name === this.genericFormatModel.spec.existingSecretName)
          ?.type as SecretType | null;

        if (!secretType) return null;
        return (SECRET_TYPE_TO_ASSET_KIND[secretType] as TGenericFormats) || null;
      } else {
        if (this.genericFormatModel.spec.keyValuePairs) {
          return AssetKind.GenericSecret;
        } else if (this.genericFormatModel.spec.accessKeyId) {
          return AssetKind.AccessKey;
        } else if (this.genericFormatModel.spec.user) {
          return AssetKind.Password;
        }
        return null;
      }
    },
    disabledFormats(): TGenericFormats[] {
      const isSupportingGenericSecret = this.clusterStore.isClusterVersionSufficient(
        this.parentClusterUuid,
        MIN_CLUSTER_VERSION_FOR_GENERIC_SECRET,
      );
      if (!isSupportingGenericSecret) {
        return [AssetKind.GenericSecret, TYPE_SSH_KEY];
      }
      const isSupportingSSHKey = this.clusterStore.isClusterVersionSufficient(
        this.parentClusterUuid,
        MIN_CLUSTER_VERSION_FOR_GIT_SSH,
      );
      if (!isSupportingSSHKey) {
        return [TYPE_SSH_KEY];
      }

      return [];
    },
  },
  methods: {
    resetSecretDetails(): void {
      this.genericFormatModel.spec = {};
    },
    async loadExistingGenericSecrets(updatedScope: IScopeModel): Promise<void> {
      try {
        this.loadingSecrets = true;
        if (!updatedScope.scope) return;

        const scopeId = scopeUtil.getScopeId(updatedScope);
        this.existingSecrets = (await credentialService.listExistingSecrets(updatedScope.scope, scopeId)).filter(
          (s) => s.type !== SecretType.DockerRegistry,
        );
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load existing secrets"));
      } finally {
        this.loadingSecrets = false;
      }
    },
    updateScope(updatedScope: IScopeModel): void {
      this.genericFormatModel.meta = scopeUtil.updateScope(this.genericFormatModel.meta, updatedScope);
      this.parentClusterUuid = updatedScope.parentClusterUuid || "";
      this.resetSecretDetails();
      this.loadExistingGenericSecrets(updatedScope);
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.$refs["elForm"] as HTMLFormElement).validate();
      return isValid;
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      allowToLeave && this.$router.push({ name: this.$route.meta.backPageName as string });
    },
    async onSubmit(): Promise<void> {
      this.isFormValid = await this.validate();
      if (!this.isFormValid) return;

      if (!this.currentFormatType) {
        console.error("invalid format type for generic secret creation");
        return;
      }

      try {
        this.submitting = true;

        const savedGenericSecret = await credentialService.createFormatBasedGenericSecret(
          this.genericFormatModel,
          this.currentFormatType,
        );
        this.$q.notify(alertUtil.getSuccess(`Credential ${savedGenericSecret.meta.name} created`));
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedGenericSecret.meta.id, kind: this.currentFormatType },
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create credential"));
        }
      } finally {
        this.submitting = false;
      }
    },
  },
});
</script>

<style lang="scss" scoped></style>
