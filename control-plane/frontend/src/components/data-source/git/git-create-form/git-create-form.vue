<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="git-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elGitForm" class="git-form" :form-state="gitModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="git" />

        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
        />

        <runai-name-section
          label="Data source"
          asset-kind="git"
          v-model="gitModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="gitModel.meta.description"
          @update-description="updateDescription"
        />

        <transition name="fade">
          <git-data-mount-section
            class="git-data-mount-section"
            v-if="isScopeSelected"
            v-model="gitModel.spec"
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
// Cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { GitDataMountSection } from "@/components/data-source/git/git-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Models
import type { ScopeType } from "@/swagger-models/authorization-client";
import { type PermittedScopes } from "@/swagger-models/authorization-client";
import type { GenericSecret } from "@/swagger-models/assets-service-client";
import { Scope, type Password } from "@/swagger-models/assets-service-client";
import type { IScopeCredentialsList } from "@/models/credential.model";
import type { IUIGitCreationRequest } from "@/models/data-source.model";
import type { IScopeModel } from "@/models/global.model";
// Services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
// Stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
// Utils
import { dataSourceUtil } from "@/utils/data-source.util";
import { alertUtil } from "@/utils/alert.util";
// Constants
import { errorMessages } from "@/common/error-message.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_FOR_GIT_SSH } from "@/common/version.constant";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    GitDataMountSection,
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
      type: Object as PropType<IUIGitCreationRequest>,
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
      dataSourceStore: useDataSourceStore(),
      clusterStore: useClusterStore(),
      gitForm: null as HTMLFormElement | null,
      gitModel: deepCopy(this.dataSourceModel) as IUIGitCreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      credentials: [] as (Password | GenericSecret)[],
      loadingCredentials: false as boolean,
      authStore: useAuthStore(),
      selectedCluster: "" as string,
      validating: false as boolean,
    };
  },
  async created() {
    this.loadCredentials();
  },
  mounted() {
    this.gitForm = this.$refs["elGitForm"] as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.gitModel.meta.scope,
        projectId: this.gitModel.meta.projectId,
        departmentId: this.gitModel.meta.departmentId,
        clusterId: this.gitModel.meta.clusterId,
      };
    },
    scopeId(): string {
      if (this.gitModel.meta.scope === Scope.Tenant) {
        return this.authStore.tenant.id.toString();
      }
      return scopeUtil.getScopeId(this.scopeModel);
    },
    clusterId(): string | null {
      if (!this.gitModel.meta.scope) return null;
      return orgTreeService.getParentClusterUuid(
        this.authStore.orgUnitList,
        this.scopeId,
        this.gitModel.meta.scope as ScopeType, // assets scope includes "system" scope which isn't compatible with scopeType of authorization
      );
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.gitModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.gitModel.meta = scopeUtil.updateScope(this.gitModel.meta, updatedScope);
      this.loadCredentials();
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
        this.$emit("submit", this.gitModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.gitForm as HTMLFormElement).validate();
      return isValid;
    },
    async loadCredentials(): Promise<void> {
      if (!this.scopeModel.scope) return;
      try {
        this.loadingCredentials = true;
        const listScope: IScopeCredentialsList = dataSourceUtil.getListScopeByMeta(this.gitModel.meta);
        const passwordsCreds = await dataSourceService.listPasswordsByScope(listScope);

        let genericSecretsCreds: GenericSecret[] = [];
        // if cluster version is 2.23 or higher load generic secrets
        const supportsGenericSecrets = this.clusterStore.isClusterVersionSufficient(
          this.clusterId,
          MIN_CLUSTER_VERSION_FOR_GIT_SSH,
        );
        if (supportsGenericSecrets) {
          genericSecretsCreds = await dataSourceService.listGenericSecretsByScope(listScope);
        }

        this.credentials = [...passwordsCreds, ...genericSecretsCreds];

        // if the pass id is not in the list, set it to empty
        const selectedPass = this.gitModel.spec.passwordAssetId;
        if (!!selectedPass && !passwordsCreds.map((c) => c.meta.id).includes(selectedPass)) {
          this.gitModel.spec.passwordAssetId = "";
        }

        // if the secret id is not in the list, set it to empty
        const selectedSecret = this.gitModel.spec.credentialId;
        if (!!selectedSecret && !this.credentials.map((c) => c.meta.id).includes(selectedSecret)) {
          this.gitModel.spec.credentialId = "";
        }
      } catch (err: unknown) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load credentials"));
      } finally {
        this.loadingCredentials = false;
      }
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

<style lang="scss" scoped></style>
