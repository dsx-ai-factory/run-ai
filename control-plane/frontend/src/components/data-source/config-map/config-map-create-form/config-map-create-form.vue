<template>
  <section class="config-map-create-form column items-center q-pt-md">
    <runai-form-wrapper
      ref="elConfigMapForm"
      class="config-map-form"
      :form-state="configMapModel"
      :enter-callback="save"
    >
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="config-map" />
        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
          :forced-disabled-scopes="disabledScopes"
          :scopes-supported-versions="scopesSupportedVersions"
        />
        <runai-name-section
          label="Data source"
          asset-kind="config-map"
          v-model="configMapModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="configMapModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <config-map-data-mount-section
            v-if="isScopeSelected"
            v-model="configMapModel.spec"
            :config-map-names="configMapNames"
            :loading-config-map-names="loadingConfigMapNames"
            :supports-sub-path="supportsSubPath"
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

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { ConfigMapDataMountSection } from "@/components/data-source/config-map/config-map-data-mount-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IScopeModel } from "@/models/global.model";
import type { IUIConfigMapCreationRequest } from "@/models/data-source.model";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import { ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { errorMessages } from "@/common/error-message.constant";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import type { ConfigMapListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { alertUtil } from "@/utils/alert.util";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_CLUSTER_SCOPE_CM_ASSETS } from "@/common/version.constant";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";
import type { TTreeDisabledScopes, TTreeSupportedVersions } from "@/models/global.model";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    ConfigMapDataMountSection,
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
      type: Object as PropType<IUIConfigMapCreationRequest>,
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
      authStore: useAuthStore(),
      configMapModel: deepCopy(this.dataSourceModel) as IUIConfigMapCreationRequest,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      configMapNames: [] as string[],
      loadingConfigMapNames: false as boolean,
      selectedCluster: "" as string,
      parentClusterUuid: null as string | null,
      validating: false as boolean,
    };
  },
  async created() {
    if (this.dataSourceModel.meta.scope) {
      const unitId = assetsUtil.getAssetScopeIdFromAssetMeta(this.dataSourceModel.meta);

      this.parentClusterUuid =
        this.dataSourceModel.meta.clusterId ||
        orgTreeService.getParentClusterUuid(
          this.authStore.orgUnitList,
          unitId,
          this.dataSourceModel.meta.scope as ScopeType, // assets scope includes "system" scope which isn't compatible with scopeType of authorization
        );
    }
    await this.loadConfigMapNames();
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    disabledScopes(): TTreeDisabledScopes {
      return {
        [ScopeType.System]: true,
        [ScopeType.Tenant]: true,
      };
    },
    scopesSupportedVersions(): TTreeSupportedVersions {
      return {
        [ScopeType.Department]: MIN_CLUSTER_VERSION_CLUSTER_SCOPE_CM_ASSETS,
        [ScopeType.Cluster]: MIN_CLUSTER_VERSION_CLUSTER_SCOPE_CM_ASSETS,
      };
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.configMapModel.meta.scope,
        projectId: this.configMapModel.meta.projectId,
        departmentId: this.configMapModel.meta.departmentId,
        clusterId: this.configMapModel.meta.clusterId,
      };
    },
    supportsSubPath(): boolean {
      return false; // at the moment product doesn't wish to show this field during asset creation
      // return this.clusterStore.isClusterVersionSufficient(
      //   this.parentClusterUuid,
      //   MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_SUB_PATH,
      // );
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.configMapModel.meta.description = description;
    },
    async loadConfigMapNames(): Promise<void> {
      let scopeId: string;
      if (this.configMapModel.meta.scope === ScopeType.Cluster && this.configMapModel.meta.clusterId) {
        scopeId = this.configMapModel.meta.clusterId;
      } else if (this.configMapModel.meta.scope === ScopeType.Department && this.configMapModel.meta.departmentId) {
        scopeId = this.configMapModel.meta.departmentId;
      } else if (this.configMapModel.meta.scope === ScopeType.Project && this.configMapModel.meta.projectId) {
        scopeId = String(this.configMapModel.meta.projectId);
      } else return;

      this.loadingConfigMapNames = true;
      try {
        const scope: ScopeType = this.configMapModel.meta.scope;
        const existingConfigMaps = await dataSourceService.listExistingConfigMaps(scope, scopeId);
        this.configMapNames = existingConfigMaps.map((cm: ConfigMapListInfo) => cm.name);
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load existing config maps"));
        throw err;
      }
      this.loadingConfigMapNames = false;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.configMapModel.meta = scopeUtil.updateScope(this.configMapModel.meta, updatedScope);
      this.parentClusterUuid = updatedScope.parentClusterUuid || null;
      this.loadConfigMapNames();
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
        this.$emit("submit", this.configMapModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.$refs.elConfigMapForm as HTMLFormElement).validate();
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

<style lang="scss" scoped></style>
