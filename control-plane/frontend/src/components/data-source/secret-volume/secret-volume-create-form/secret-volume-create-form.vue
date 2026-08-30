<template>
  <section class="secret-volume-create-form column items-center q-pt-md">
    <runai-form-wrapper
      ref="elSecretVolumeForm"
      class="secret-volume-form"
      :form-state="secretVolumeModel"
      :enter-callback="save"
    >
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="secret-volume" />
        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
          :scopes-supported-versions="scopesSupportedVersions"
        />
        <runai-name-section
          label="Data source"
          asset-kind="secret-volume"
          v-model="secretVolumeModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="secretVolumeModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <secret-volume-data-mount-section
            v-if="isScopeSelected"
            v-model="secretVolumeModel.spec"
            :credentials="credentials"
            :loading-credentials="loadingCredentials"
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
            label="create data source"
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
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { SecretVolumeDataMountSection } from "@/components/data-source/secret-volume/secret-volume-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// Models
import type { IScopeModel, TTreeSupportedVersions } from "@/models/global.model";
import type { IUISecretCreationRequest } from "@/models/data-source.model";
import { AssetKind, type CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";
import type { IScopeCredentialsList } from "@/models/credential.model";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";

// Services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// Utils
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { alertUtil } from "@/utils/alert.util";
import { dataSourceUtil } from "@/utils/data-source.util";

// Stores
import { useClusterStore } from "@/stores/cluster.store";

// Constants
import { errorMessages } from "@/common/error-message.constant";
import { MIN_CLUSTER_VERSION_FOR_SECRET_DATASOURCE } from "@/common/version.constant";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    SecretVolumeDataMountSection,
    RunaiScopeSection,
    DataSourceTypeSection,
    RunaiBtn,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    dataSourceModel: {
      type: Object as PropType<IUISecretCreationRequest>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      secretVolumeModel: deepCopy(this.dataSourceModel) as IUISecretCreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      credentials: [] as CredentialsListResponseEntry[],
      loadingCredentials: false as boolean,
      selectedCluster: "" as string,
      validating: false as boolean,
    };
  },
  async created() {
    await this.loadSecretNames();
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopesSupportedVersions(): TTreeSupportedVersions {
      return {
        [ScopeType.Tenant]: MIN_CLUSTER_VERSION_FOR_SECRET_DATASOURCE,
        [ScopeType.Department]: MIN_CLUSTER_VERSION_FOR_SECRET_DATASOURCE,
        [ScopeType.Cluster]: MIN_CLUSTER_VERSION_FOR_SECRET_DATASOURCE,
        [ScopeType.Project]: MIN_CLUSTER_VERSION_FOR_SECRET_DATASOURCE,
      };
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.secretVolumeModel.meta.scope,
        projectId: this.secretVolumeModel.meta.projectId,
        departmentId: this.secretVolumeModel.meta.departmentId,
        clusterId: this.secretVolumeModel.meta.clusterId,
      };
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.secretVolumeModel.meta.description = description;
    },
    async loadSecretNames(): Promise<void> {
      if (!this.scopeModel.scope) return;
      this.loadingCredentials = true;
      try {
        const listScope: IScopeCredentialsList = dataSourceUtil.getListScopeByMeta(this.secretVolumeModel.meta);
        const credentials: CredentialsListResponseEntry[] = await credentialService.listCredentials(listScope);
        // secret-volume data-source can use all the credentials except DockerRegistry
        this.credentials = credentials.filter((c) => c.meta.kind !== AssetKind.DockerRegistry);
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load generic secret assets"));
        throw err;
      }
      this.loadingCredentials = false;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.secretVolumeModel.meta = scopeUtil.updateScope(this.secretVolumeModel.meta, updatedScope);
      this.loadSecretNames();
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
        this.$emit("submit", this.secretVolumeModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.$refs.elSecretVolumeForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>
