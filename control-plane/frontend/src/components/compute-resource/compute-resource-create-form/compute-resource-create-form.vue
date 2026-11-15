<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="compute-resource-create-form">
    <runai-form-wrapper class="compute-resource-form" :form-state="computeModel" :enter-callback="save">
      <q-form ref="elComputeResourceForm">
        <runai-scope-section
          kind-of-asset="compute resource"
          :read-only="!isScopeSelectable || isEditMode"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopes"
          :key="scopeModel.clusterId || ''"
          :selected-cluster-id="scopeModel.clusterId || ''"
        />

        <runai-name-section
          label="Compute resource"
          asset-kind="compute"
          :original-name="originalAssetName"
          v-model="computeModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="computeModel.meta.description"
          @update-description="updateDescription"
        />

        <transition name="fade">
          <compute-resource-resources-section
            v-if="isScopeSelected"
            :resources-model="resourceSectionModel"
            :large-shm-request="computeModel.spec.largeShmRequest || false"
            :extended-resources="computeModel.spec.extendedResources || []"
            :support-multi-fraction-gpu="supportMultiFractionGpu"
            :support-gpu-limit="supportGpuLimit"
            @update-resources-model="updateResourceSectionModel"
            @update-shm-request="updateShmRequest"
            @update-extended-resources="updateExtendedResources"
            @is-section-invalid="sectionValidation.resources = !$event"
          />
        </transition>

        <section class="row items-center q-mt-md">
          <q-field
            class="col-4 form-hint no-padding"
            :model-value="displayFormHint"
            :rules="[isFormIncomplete]"
          ></q-field>
          <div class="buttons q-ml-auto">
            <runai-btn
              class="q-mr-md"
              aid="form-cancel-environment-btn"
              @click="onCancel"
              flat
              color="primary"
              label="cancel"
            />
            <runai-btn
              type="submit"
              color="primary"
              aid="form-create-compute-btn"
              @click="save"
              :loading="validating || submitting"
              :label="`${isEditMode ? 'save' : 'create'} compute resource`"
            />
          </div>
        </section>
      </q-form>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { ComputeResourceResourcesSection } from "../compute-resource-resources-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Models
import type { IResourcesSectionModel, IUIComputeResourceCreation } from "@/models/compute-resource.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import {
  type ComputeCreationRequestSpec,
  GpuRequestType,
  Scope,
  type ExtendedResource,
} from "@/swagger-models/assets-service-client";
// Services
import { scopeUtil } from "@/utils/scope.util";
import { deepCopy } from "@/utils/common.util";
// Stores
import { useClusterStore } from "@/stores/cluster.store";
import type { IScopeModel } from "@/models/global.model";
import { MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU } from "@/common/version.constant";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { useSettingStore } from "@/stores/setting.store";

export interface ISectionValidation {
  resources: boolean;
}

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiNameSection,
    ComputeResourceResourcesSection,
    RunaiScopeSection,
    RunaiBtn,
  },
  emits: ["cancel", "submit"],
  props: {
    initialModel: {
      type: Object as PropType<IUIComputeResourceCreation>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    originalAssetName: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      computeModel: deepCopy(this.initialModel) as IUIComputeResourceCreation,
      computeForm: null as HTMLFormElement | null,
      clusterStore: useClusterStore(),
      sectionValidation: {
        resources: true,
      } as ISectionValidation,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      validating: false as boolean,
    };
  },
  mounted() {
    this.computeForm = this.$refs.elComputeResourceForm as HTMLFormElement;
  },
  computed: {
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.computeModel.meta.scope,
        projectId: this.computeModel.meta.projectId,
        departmentId: this.computeModel.meta.departmentId,
        clusterId: this.computeModel.meta.clusterId,
      };
    },
    resourceSectionModel(): IResourcesSectionModel {
      return {
        ...this.computeModel.spec,
        gpuRequestType: this.computeModel.spec.gpuRequestType,
        gpuDevicesRequest: this.computeModel.spec.gpuDevicesRequest,
        gpuPortionLimit: this.computeModel.spec.gpuPortionLimit,
        gpuPortionRequest: this.computeModel.spec.gpuPortionRequest,
        gpuMemoryLimit: this.computeModel.spec.gpuMemoryLimit,
        gpuMemoryRequest: this.computeModel.spec.gpuMemoryRequest,
        cpuCore: {
          request: this.computeModel.spec.cpuCoreRequest,
          limit: this.computeModel.spec.cpuCoreLimit,
        },
        cpuMemory: {
          request: this.computeModel.spec.cpuMemoryRequest,
          limit: this.computeModel.spec.cpuMemoryLimit,
        },
      };
    },
    parentClusterId(): string | null {
      if (!this.computeModel.meta.scope || this.computeModel.meta.scope === Scope.Tenant) return null;

      const scopeId = assetsUtil.getAssetScopeIdFromAssetMeta(this.computeModel.meta);

      return orgTreeService.getParentClusterUuid(useAuthStore().orgUnitList, scopeId, this.computeModel.meta.scope);
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.parentClusterId || "",
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportGpuLimit(): boolean {
      return useSettingStore().isGPUOverProvisioningEnabled;
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.computeModel.meta.description = description;
    },
    updateScope(updatedScope: IScopeModel): void {
      this.computeModel.meta = scopeUtil.updateScope(this.computeModel.meta, updatedScope);
    },
    updateResourceSectionModel(resourceSectionModel: IResourcesSectionModel): void {
      this.computeModel.spec = {
        ...this.computeModel.spec,
        gpuRequestType: resourceSectionModel.gpuRequestType,
        gpuDevicesRequest: resourceSectionModel.gpuDevicesRequest,
        gpuPortionLimit: resourceSectionModel.gpuPortionLimit,
        gpuPortionRequest: resourceSectionModel.gpuPortionRequest,
        gpuMemoryLimit: resourceSectionModel.gpuMemoryLimit,
        gpuMemoryRequest: resourceSectionModel.gpuMemoryRequest,
        cpuCoreRequest: resourceSectionModel.cpuCore.request,
        cpuCoreLimit: resourceSectionModel.cpuCore.limit,
        cpuMemoryRequest: resourceSectionModel.cpuMemory.request,
        cpuMemoryLimit: resourceSectionModel.cpuMemory.limit,
      };
    },
    updateShmRequest(largeShmRequest: boolean): void {
      this.computeModel.spec.largeShmRequest = largeShmRequest;
    },
    updateExtendedResources(extendedResources: ExtendedResource[]): void {
      const isEmpty = !extendedResources.length;
      this.computeModel.spec.extendedResources = isEmpty ? null : extendedResources;
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success: boolean = await this.validate();
        if (!success) {
          this.showHint();
          return;
        }
        this.prepareModelForRequest();
        this.$emit("submit", this.computeModel);
      } catch (error) {
        console.error("Error while validating compute resource form", error);
      } finally {
        this.validating = false;
      }
    },
    prepareModelForRequest(): void {
      const spec: ComputeCreationRequestSpec = this.computeModel.spec;
      delete spec.gpuMemoryPortionRequest;
      delete spec.gpuMemoryPortionLimit;

      if (
        this.clusterStore.isClusterVersionSufficient(this.parentClusterId, MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU)
      ) {
        return;
      } else if (this.computeModel.spec.gpuDevicesRequest === 1 && !this.computeModel.spec.gpuRequestType) {
        this.computeModel.spec.gpuRequestType = GpuRequestType.Portion;
        this.computeModel.spec.gpuPortionRequest = 1;
      }
    },
    formSectionsValid(): boolean {
      return this.sectionValidation.resources;
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.computeForm as HTMLFormElement).validate();
      return isValid && this.formSectionsValid();
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>

<style lang="scss" scoped>
.runai-form-wrapper {
  width: 800px;
}
</style>
