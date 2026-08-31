<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="pvc-create-form column items-center q-pt-md">
    <runai-form-wrapper ref="elPvcForm" class="pvc-form" :form-state="pvcModel" :enter-callback="save">
      <runai-expansion-wrapper>
        <data-source-type-section data-source-kind="pvc" />
        <runai-scope-section
          kind-of-asset="data source"
          :read-only="!isScopeSelectable"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
          :allowed-scopes="allowedScopes"
          :information-text="scopeSectionInformationText"
        >
          <volume-sharing-section
            v-if="showVolumeSharing"
            :storage-classes="storageClasses"
            :model-value="pvcModel.spec.dataSharing"
            @update:model-value="updateDataSharing"
          />
        </runai-scope-section>
        <runai-name-section
          label="Data source"
          asset-kind="pvc"
          v-model="pvcModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="pvcModel.meta.description"
          @update-description="updateDescription"
        />
        <transition name="fade">
          <pvc-data-mount
            v-if="isScopeSelected"
            :pvc-data-model="pvcModel.spec"
            @update:pvc-data-model="updatePvcDataModel"
            :pvc-list="existingPvcs"
            :init-claim-info-from-scratch="initClaimInfoFromScratch"
            :storage-classes="storageClasses"
            :loading="loadingData"
            :loading-storage-class="loadingStorageClass"
            :is-project-scoped="pvcModel.meta.scope === 'project'"
            :support-custom-pvc-attributes="supportCustomPvcAttributes"
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

// Stores
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { PvcDataMount } from "@/components/data-source/pvc/pvc-data-mount";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
import { VolumeSharingSection } from "@/components/data-source/pvc/volume-sharing-section";

// utils
import { alertUtil } from "@/utils/alert.util";
import { scopeUtil } from "@/utils/scope.util";

// models
import { type PermittedScopes, ScopeType } from "@/swagger-models/authorization-client";
import type { PVCListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { IUIPVCCreationRequest } from "@/models/data-source.model";

// services
import { deepCopy } from "@/utils/common.util";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { type IScopeModel } from "@/models/global.model";
import { DataSourceTypeSection } from "../../sections/data-source-type-section";
import { errorMessages } from "@/common/error-message.constant";
import {
  MIN_CLUSTER_VERSION_FOR_CUSTOM_PVC_ATTRIBUTES,
  MIN_CLUSTER_VERSION_FOR_DATA_VOLUME,
  MIN_CLUSTER_VERSION_FOR_SHARED_PVC,
} from "@/common/version.constant";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiNameSection,
    PvcDataMount,
    RunaiScopeSection,
    DataSourceTypeSection,
    RunaiBtn,
    VolumeSharingSection,
  },
  emits: ["submit", "cancel"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    dataSourceModel: {
      type: Object as PropType<IUIPVCCreationRequest>,
      required: true,
    },
    initClaimInfoFromScratch: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
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
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      pvcForm: null as HTMLFormElement | null,
      pvcModel: deepCopy(this.dataSourceModel) as IUIPVCCreationRequest,
      isFormInvalid: false as boolean,
      existingPvcs: [] as string[],
      loadingData: false,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      selectedCluster: "" as string,
      parentClusterUuid: null as string | null,
      validating: false as boolean,
      DOCS,
    };
  },
  setup() {
    const { storageClasses, loadingStorageClass, loadStorageClasses, loadStorageClassConfiguration } =
      useStorageClassConfiguration("asset");

    return {
      storageClasses,
      loadingStorageClass,
      loadStorageClasses,
      loadStorageClassConfiguration,
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
    await this.loadCmpData(this.parentClusterUuid);

    if (!this.existingPvcs || !this.existingPvcs.length) {
      this.pvcModel.spec.existingPvc = false;
    }
  },
  mounted() {
    this.pvcForm = this.$refs["elPvcForm"] as HTMLFormElement;
  },
  computed: {
    showVolumeSharing(): boolean {
      return this.supportVolumeSharing && this.pvcModel.meta.scope !== ScopeType.Project;
    },
    supportVolumeSharing(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.parentClusterUuid, MIN_CLUSTER_VERSION_FOR_SHARED_PVC);
    },
    supportCustomPvcAttributes(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.parentClusterUuid,
        MIN_CLUSTER_VERSION_FOR_CUSTOM_PVC_ATTRIBUTES,
      );
    },
    scopeSectionInformationText(): string | undefined {
      if (!this.pvcModel.meta.scope || this.pvcModel.meta.scope === ScopeType.Project) return undefined;

      // If the cluster version supports shared volume pvc, we don't need to show the information text
      if (this.supportVolumeSharing) {
        return undefined;
      }

      return `A PVC will be created for each selected project with the storage allocation specified in this form. Data on the underlying file storage system won't be copied or synced.${
        this.clusterStore.isClusterVersionSufficient(this.parentClusterUuid, MIN_CLUSTER_VERSION_FOR_DATA_VOLUME)
          ? ` To share underlying data across projects, use <a href='${DOCS.WORKLOADS.DATA_VOLUMES}' target='_blank'>Data volumes</a> instead.`
          : ""
      }`;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.pvcModel.meta.scope,
        projectId: this.pvcModel.meta.projectId,
        departmentId: this.pvcModel.meta.departmentId,
        clusterId: this.pvcModel.meta.clusterId,
      };
    },
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
  },
  methods: {
    async updatePvcDataModel(updatedPvcDataModel: IUIPVCCreationRequest["spec"]): Promise<void> {
      const newStorageClass = updatedPvcDataModel.claimInfo?.storageClass;
      const oldStorageClass = this.pvcModel.spec.claimInfo?.storageClass;

      this.pvcModel.spec = updatedPvcDataModel;

      // If the storage class hasn't changed there is no need to do any changes
      if (!newStorageClass || newStorageClass === oldStorageClass) return;

      if (this.pvcModel.spec.claimInfo && this.parentClusterUuid) {
        this.pvcModel.spec.claimInfo = await this.loadStorageClassConfiguration(
          this.parentClusterUuid,
          newStorageClass,
          true,
        );
      }
    },
    updateDescription(description: string): void {
      this.pvcModel.meta.description = description;
    },
    updateDataSharing(dataSharing: boolean): void {
      if (dataSharing) {
        this.pvcModel.spec = {
          ...this.pvcModel.spec,
          dataSharing,
          existingPvc: false,
          claimName: "",
        };
        return;
      }

      this.pvcModel.spec = { ...this.pvcModel.spec, dataSharing };
    },
    updateScope(updatedScope: IScopeModel): void {
      this.pvcModel.meta = scopeUtil.updateScope(this.pvcModel.meta, updatedScope);
      this.parentClusterUuid = updatedScope.parentClusterUuid || null;
      this.loadCmpData(updatedScope.parentClusterUuid);

      // Reseting the data sharing field since it is dependant on loaded storage classes
      this.pvcModel.spec.dataSharing = false;
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
        if (this.pvcModel.spec.claimInfo) {
          this.pvcModel.spec.claimInfo.accessModes = {
            readOnlyMany: this.pvcModel.spec.claimInfo?.accessModes?.readOnlyMany || undefined,
            readWriteMany: this.pvcModel.spec.claimInfo?.accessModes?.readWriteMany || undefined,
            readWriteOnce: this.pvcModel.spec.claimInfo?.accessModes?.readWriteOnce || undefined,
          };
        }
        this.$emit("submit", this.pvcModel);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.pvcForm as HTMLFormElement).validate();
      return isValid;
    },
    async loadExistingPvcs(): Promise<void> {
      let scopeId: string;
      // there are no existing secrets / pvcs on the tenant scope, so if the user wishes to create a tenant scope asset we need to present him with existing on cluster scope
      if (this.pvcModel.meta.scope === ScopeType.Tenant) {
        scopeId = this.selectedCluster;
      } else if (this.pvcModel.meta.scope === ScopeType.Cluster && this.pvcModel.meta.clusterId) {
        scopeId = this.pvcModel.meta.clusterId;
      } else if (this.pvcModel.meta.scope === ScopeType.Department && this.pvcModel.meta.departmentId) {
        scopeId = this.pvcModel.meta.departmentId;
      } else if (this.pvcModel.meta.scope === ScopeType.Project && this.pvcModel.meta.projectId) {
        scopeId = String(this.pvcModel.meta.projectId);
      } else return;

      try {
        const scope: ScopeType =
          this.pvcModel.meta.scope === ScopeType.Tenant ? ScopeType.Cluster : this.pvcModel.meta.scope;
        const existingPvcs = await dataSourceService.listExistingPVCs(scope, scopeId);
        this.existingPvcs = existingPvcs.map((pvc: PVCListInfo) => pvc.name);
        const hasExistingPVCs: boolean = this.existingPvcs.length > 0;
        if (this.initClaimInfoFromScratch || !hasExistingPVCs) this.pvcModel.spec.existingPvc = hasExistingPVCs;
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load existing pvcs"));
        throw err;
      }
    },
    async loadCmpData(parentClusterUuid: string | null | undefined): Promise<void> {
      try {
        this.loadingData = true;
        await Promise.all([
          this.loadExistingPvcs(),
          parentClusterUuid &&
            this.loadStorageClasses(parentClusterUuid, [this.pvcModel.spec.claimInfo?.storageClass], true),
        ]);
        this.autoSelectStorageClassIfNeeded();
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load resources"));
      } finally {
        this.loadingData = false;
      }
    },
    autoSelectStorageClassIfNeeded(): void {
      if (this.storageClasses.length !== 1 || this.pvcModel.spec.existingPvc) return;
      this.updatePvcDataModel({
        ...this.pvcModel.spec,
        claimInfo: {
          ...this.pvcModel.spec.claimInfo,
          storageClass: this.storageClasses[0].storageClassName,
        },
      });
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
