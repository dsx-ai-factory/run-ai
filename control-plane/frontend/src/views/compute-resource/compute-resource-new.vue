<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <notification-alert-bar class="q-py-lg" v-if="showInfoBar">
      <span>
        At least one workload is using this environment. The changes won’t affect those workloads but will apply to new
        workloads using this environment.
      </span>
    </notification-alert-bar>
    <section class="compute-resource-new-page column items-center q-pt-md">
      <compute-resource-create-form
        v-if="!isPageLoading && computeResource"
        @cancel="onCancel"
        @submit="onSubmit"
        :initial-model="computeResource"
        :submitting="submitting"
        :is-scope-selectable="!$route.query.scope"
        :allowed-scopes="allowedCreateScopes"
        :is-edit-mode="isEditMode"
        :original-asset-name="originalAssetName"
      />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// Components
import { ComputeResourceCreateForm } from "@/components/compute-resource/compute-resource-create-form";
import { NotificationAlertBar } from "@/components/common/notification-alert-bar";

// Models
import type { IUIComputeResourceCreation } from "@/models/compute-resource.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import {
  AssetKind,
  GpuRequestType,
  Scope,
  type ComputeAsset,
  type ComputeCreationRequest,
  type ComputeUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action } from "@/swagger-models/authorization-client";

// Services
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";

// Stores
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { IScopeModel } from "@/models/global.model";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";

export default defineComponent({
  components: {
    ComputeResourceCreateForm,
    NotificationAlertBar,
  },
  data() {
    return {
      authStore: useAuthStore(),
      computeResourceStore: useComputeResourceStore(),
      appStore: useAppStore(),
      computeResource: null as IUIComputeResourceCreation | null,
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      submitting: false as boolean,
      isEditMode: false as boolean,
      originalAssetName: "" as string,
      showInfoBar: false as boolean,
    };
  },
  async created() {
    this.computeResource = computeResourceService.getEmptyComputeResourceModel();

    const editAssetId = this.$route.params.id?.toString();
    if (editAssetId && unleashService.isUpdateAssetEnabled()) {
      await this.loadFromExistingCompute(editAssetId, true);
      this.originalAssetName = this.computeResource.meta.name;
      this.isEditMode = true;
      this.appStore.setPageTitleOverride(`Edit compute resource: ${this.computeResource.meta.name}`);
    }

    const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
    if (fromCopyId) await this.loadFromExistingCompute(fromCopyId);

    if (this.$route.query.scope) {
      this.computeResource.meta = assetsUtil.setAssetMetaFromQuery(
        this.computeResource.meta,
        this.$route.query as unknown as IScopeModel,
        this.authStore.orgUnitList,
      );
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    allowedCreateScopes(): PermittedScopes {
      return this.computeResourceStore.actionPermissionsByScopes[Action.Create];
    },
  },
  methods: {
    async loadFromExistingCompute(fromCopyId: string, keepScope = false): Promise<void> {
      try {
        const computeResource: ComputeAsset = await this.computeResourceStore.loadById(fromCopyId);
        const UIComputeResource: IUIComputeResourceCreation = this.prepareComputeResourceForUI(computeResource);
        this.computeResource = UIComputeResource;

        // Our presets have some weird data that doesn't make sense in the UI, so we clean it up here
        if (
          (this.computeResource.spec.gpuPortionRequest && this.computeResource.spec.gpuPortionRequest === 1) ||
          (this.computeResource.spec.gpuRequestType === GpuRequestType.Portion &&
            !this.computeResource.spec.gpuPortionRequest)
        ) {
          delete this.computeResource.spec.gpuPortionRequest;
          delete this.computeResource.spec.gpuPortionLimit;
          delete this.computeResource.spec.gpuRequestType;
        }

        // When editing an asset we don't want to clean the scope and we want to display a message if the asset is being used
        if (keepScope) {
          if (
            assetsUtil.hasLiveWorkload(computeResource.usedBy?.workspaces || []) ||
            assetsUtil.hasLiveWorkload(computeResource.usedBy?.trainings || []) ||
            assetsUtil.hasLiveWorkload(computeResource.usedBy?.distributed || []) ||
            assetsUtil.hasLiveWorkload(computeResource.usedBy?.inferences || [])
          ) {
            this.showInfoBar = true;
          }
          return;
        }

        if (this.computeResource.meta.scope === Scope.Tenant) this.computeResource.meta.scope = null;
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Failed to load compute resource`));
        }
        this.appStore.setFallback(true);
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
      allowToLeave &&
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { workloadFormName: this.workloadFormName },
          params: paramId ? { id: paramId } : undefined,
        });
    },
    async onSubmit(computeResource: IUIComputeResourceCreation): Promise<void> {
      try {
        this.submitting = true;
        let savedCompute: ComputeAsset;

        if (this.isEditMode) {
          savedCompute = await computeResourceService.update(
            this.$route.params.id?.toString() || "",
            this.prepareForUpdateRequest(computeResource),
          );
        } else {
          savedCompute = await computeResourceService.create(this.prepareForCreationRequest(computeResource));
        }
        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: {
            createdEntityId: savedCompute.meta.id,
            kind: AssetKind.Compute,
            workloadFormName: this.workloadFormName,
          },
          params: paramId ? { id: paramId } : undefined,
        });
        this.$q.notify(
          alertUtil.getSuccess(`Compute resource ${savedCompute.meta.name} ${this.isEditMode ? "saved" : "created"}`),
        );
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Failed to ${this.isEditMode ? "save" : "create"} compute resource`));
        }
      } finally {
        this.submitting = false;
      }
    },
    prepareForUpdateRequest(computeResource: IUIComputeResourceCreation): ComputeUpdateRequest {
      return {
        meta: {
          name: computeResource.meta.name,
          description: computeResource.meta.description,
        },
        spec: computeResource.spec,
      };
    },
    prepareForCreationRequest(computeResource: IUIComputeResourceCreation): ComputeCreationRequest {
      const convertedMeta = assetsUtil.prepareAssetMetaToCreationRequest(computeResource.meta);
      if (!convertedMeta) throw "Failed to convert compute resource model";

      const convertedModel: ComputeCreationRequest = {
        meta: convertedMeta,
        spec: computeResource.spec,
      };
      return convertedModel;
    },
    prepareComputeResourceForUI(computeResource: ComputeAsset): IUIComputeResourceCreation {
      const { gpuDevicesRequest, gpuRequestType } = computeResource.spec;
      const { meta } = computeResource;
      const isAllowedScope: boolean = orgTreeService.isAllowedScope(
        meta,
        this.allowedCreateScopes,
        this.authStore.orgUnitList,
      );
      return {
        meta: {
          name: meta.name,
          description: meta.description,
          scope: isAllowedScope ? meta.scope : null,
          projectId: isAllowedScope ? meta.projectId : null,
          departmentId: isAllowedScope ? meta.departmentId : null,
          clusterId: isAllowedScope ? meta.clusterId : null,
        },
        spec: {
          ...computeResource.spec,
          gpuRequestType: gpuDevicesRequest === 0 ? null : gpuRequestType,
          gpuDevicesRequest: gpuDevicesRequest || 0,
          cpuCoreRequest: computeResource.spec.cpuCoreRequest || 0,
          cpuMemoryRequest: computeResource.spec.cpuMemoryRequest || "0M",
          largeShmRequest: computeResource.spec.largeShmRequest || false,
        },
      };
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>
