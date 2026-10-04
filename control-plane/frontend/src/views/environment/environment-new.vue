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
    <section class="environment-new-page column items-center q-pt-md">
      <environment-create-form
        v-if="!isPageLoading && environment"
        @submit="onSubmit"
        @cancel="onCancel"
        :submitting="submitting"
        :initial-environment="environment"
        :is-scope-selectable="!$route.query.scope"
        :locked-workload-types="lockedWorkloadSupportedTypes"
        :allowed-scopes="allowedCreateScopes"
        :is-edit-mode="isEditMode"
        :original-asset-name="originalAssetName"
      />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import { EnvironmentCreateForm } from "@/components/environment/environment-create-form";
// stores
import { useEnvironmentStore } from "@/stores/environment.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";

// services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// models
import type { IUIEnvironmentAsset, IUIConnection, IUIEnvironmentAssetSpec } from "@/models/environment.model";
import type { DistributedFramework } from "@/swagger-models/assets-service-client";
import {
  type EnvironmentAsset,
  type EnvironmentCreationRequest,
  type AssetMeta,
  type EnvironmentAssetSpec,
  type EnvironmentUpdateRequest,
  ImagePullPolicy,
  AssetKind,
  Scope,
} from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { EWorkloadName } from "@/models/workload.model";
import { EWorkloadFormType } from "@/models/workload.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { IScopeModel } from "@/models/global.model";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";

import { NotificationAlertBar } from "@/components/common/notification-alert-bar";

export default defineComponent({
  components: {
    EnvironmentCreateForm,
    NotificationAlertBar,
  },
  data() {
    return {
      authStore: useAuthStore(),
      environmentStore: useEnvironmentStore(),
      appStore: useAppStore(),
      environment: null as IUIEnvironmentAsset | null,
      submitting: false as boolean,
      workloadType: this.$route.query.workloadType ? this.$route.query.workloadType.toString() : (null as string | null),
      distFramework: this.$route.query.distFramework
        ? this.$route.query.distFramework
        : (null as DistributedFramework | null),
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      isEditMode: false as boolean,
      originalAssetName: "" as string,
      showInfoBar: false as boolean,
    };
  },
  async created() {
    this.environment = environmentService.getEnvironmentModel(this.authStore.isSSO);

    if (this.appStore.isWhiteLabel && this.environment.meta.workloadSupportedTypes) {
      this.environment.meta.workloadSupportedTypes.training = false;
    }

    const editAssetId: string = this.$route.params.id?.toString() || "";
    if (editAssetId && unleashService.isUpdateAssetEnabled()) {
      await this.loadFromExistingEnv(editAssetId, true);
      this.originalAssetName = this.environment.meta.name;
      this.isEditMode = true;
      this.appStore.setPageTitleOverride(`Edit environment: ${this.environment.meta.name}`);
    }

    const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
    if (fromCopyId) await this.loadFromExistingEnv(fromCopyId);

    if (this.$route.query.scope) {
      this.environment.meta = assetsUtil.setAssetMetaFromQuery(
        this.environment.meta,
        this.$route.query as unknown as IScopeModel,
        this.authStore.orgUnitList,
      );
    }

    if (this.workloadType && this.workloadType !== EWorkloadFormType.Template) {
      if (this.workloadType === EWorkloadFormType.DistributedInference) {
        this.environment.meta.workloadSupportedTypes = {
          ...this.environment.meta.workloadSupportedTypes,
          workspace: false,
          training: false,
          inference: true,
          distributed: true,
          distFramework: undefined,
        };
      } else {
        this.environment.meta.workloadSupportedTypes = { [this.workloadType as EWorkloadName]: true };
        if (this.distFramework) {
          this.environment.meta.workloadSupportedTypes.distributed = true;
          this.environment.meta.workloadSupportedTypes.distFramework = this.distFramework as DistributedFramework;
        }
      }
    }
    this.appStore.setPageLoading(false);
  },
  computed: {
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    lockedWorkloadSupportedTypes(): boolean {
      return !!this.workloadType && this.workloadType !== EWorkloadFormType.Template;
    },
    allowedCreateScopes(): PermittedScopes {
      return this.environmentStore.actionPermissionsByScopes[Action.Create];
    },
  },
  methods: {
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
    async onSubmit(UIEnvironment: IUIEnvironmentAsset): Promise<void> {
      try {
        this.submitting = true;
        let savedEnvironment: EnvironmentAsset;

        if (this.isEditMode) {
          const environment: EnvironmentUpdateRequest = this.prepareEnvironmentForUpdateRequest(UIEnvironment);
          savedEnvironment = await this.environmentStore.update(this.$route.params.id?.toString() || "", environment);
        } else {
          const environment: EnvironmentCreationRequest = this.prepareEnvironmentForCreationRequest(UIEnvironment);
          if (!environmentService.validateEnvironment(environment)) {
            throw new Error("Some of the fields are missing or invalid");
          }
          savedEnvironment = await this.environmentStore.createEnvironment(environment);
        }

        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: {
            createdEntityId: savedEnvironment.meta.id,
            kind: AssetKind.Environment,
            workloadFormName: this.workloadFormName,
          },
          params: paramId ? { id: paramId } : undefined,
        });
        this.$q.notify(
          alertUtil.getSuccess(`Environment ${savedEnvironment.meta.name} ${this.isEditMode ? "saved" : "created"}`),
        );
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Failed to ${this.isEditMode ? "save" : "create"} environment`));
        }
      } finally {
        this.submitting = false;
      }
    },
    async loadFromExistingEnv(fromCopyId: string, keepScope = false): Promise<void> {
      try {
        const environment: EnvironmentAsset = await this.environmentStore.loadById(fromCopyId);
        const UIEnvironment = this.prepareEnvironmentForUI(environment);
        this.environment = UIEnvironment;

        // When editing an asset we don't want to clean the scope and we want to display a message if the asset is being used
        if (keepScope) {
          if (
            assetsUtil.hasLiveWorkload(environment.usedBy?.workspaces || []) ||
            assetsUtil.hasLiveWorkload(environment.usedBy?.trainings || []) ||
            assetsUtil.hasLiveWorkload(environment.usedBy?.distributed || []) ||
            assetsUtil.hasLiveWorkload(environment.usedBy?.inferences || [])
          ) {
            this.showInfoBar = true;
          }
          return;
        }

        if (this.environment.meta.scope === Scope.Tenant) this.environment.meta.scope = null;
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to get Environment ${fromCopyId}`));
        console.error(e);
        this.appStore.setFallback(true);
      }
    },
    prepareEnvironmentForCreationRequest(environment: IUIEnvironmentAsset): EnvironmentCreationRequest {
      const environmentForApi = {} as EnvironmentCreationRequest;
      environmentForApi.meta = environment.meta as AssetMeta;
      if (environmentForApi.meta.scope !== Scope.Cluster) delete environmentForApi.meta.clusterId;
      environmentForApi.spec = this.prepareSpecForRequest(environment.spec);
      return environmentForApi;
    },
    prepareEnvironmentForUpdateRequest(environment: IUIEnvironmentAsset): EnvironmentUpdateRequest {
      const environmentForApi: EnvironmentUpdateRequest = {
        meta: {
          name: environment.meta.name,
          description: environment.meta.description,
        },
        spec: this.prepareSpecForRequest(environment.spec),
      };
      return environmentForApi;
    },
    prepareSpecForRequest(environmentSpec: IUIEnvironmentAssetSpec): EnvironmentAssetSpec {
      return {
        ...environmentSpec,
        connections: environmentSpec.connections
          .filter((tool: IUIConnection) => tool.toolType !== null)
          .map((tool: IUIConnection) => {
            return environmentService.prepareConnectionForRequest(tool);
          }),
        image: environmentSpec.image?.trim() || null,
      };
    },
    prepareEnvironmentForUI(environment: EnvironmentAsset): IUIEnvironmentAsset {
      const { meta, spec }: { meta: AssetMeta; spec: EnvironmentAssetSpec } = environment;
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
          workloadSupportedTypes: meta.workloadSupportedTypes || {
            workspace: true,
            training: !this.appStore.isWhiteLabel,
            inference: false,
          },
        },
        spec: {
          command: spec.command || null,
          args: spec.args || null,
          environmentVariables: spec.environmentVariables || [],
          image: spec.image || null,
          imagePullPolicy: spec.imagePullPolicy || ImagePullPolicy.IfNotPresent,
          workingDir: spec.workingDir || null,
          connections:
            (spec.connections?.map((connection) =>
              environmentService.prepareConnectionForUI(connection),
            ) as Array<IUIConnection>) || [],
          uidGidSource: spec.uidGidSource || undefined,
          runAsGid: spec.runAsGid,
          runAsUid: spec.runAsUid,
          overrideUidGidInWorkspace: spec.overrideUidGidInWorkspace,
          supplementalGroups: spec.supplementalGroups,
          capabilities: spec.capabilities || [],
        },
      };
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>

<style lang="scss">
.information-text {
  flex: unset;
}
</style>
