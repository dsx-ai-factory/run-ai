<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="secret-volume-new column items-center q-pt-md">
    <secret-volume-create-form
      v-if="!isPageLoading && dataSourceModel"
      @submit="onSubmit"
      @cancel="onCancel"
      :submitting="submitting"
      :data-source-model="dataSourceModel"
      :is-scope-selectable="!$route.query.scope"
      :allowed-scopes="allowedCreateScopes"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// Components
import { SecretVolumeCreateForm } from "@/components/data-source/secret-volume/secret-volume-create-form";

// Services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// Models
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IUISecretCreationRequest } from "@/models/data-source.model";
import type { IScopeModel } from "@/models/global.model";
import {
  Scope,
  AssetKind,
  type SecretAsset,
  type SecretAssetCreationRequest,
} from "@/swagger-models/assets-service-client";

// Stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";

// Utils
import { deepCopy } from "@/utils/common.util";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { alertUtil } from "@/utils/alert.util";

export default defineComponent({
  components: {
    SecretVolumeCreateForm,
  },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      dataSourceModel: null as IUISecretCreationRequest | null,
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      submitting: false as boolean,
    };
  },
  async created() {
    await this.init();
  },
  computed: {
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    allowedCreateScopes(): PermittedScopes {
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.SecretVolumeAssets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptySecretVolumeModel();
        const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
        if (fromCopyId) await this.createFromExistingDS(fromCopyId);
        if (this.$route.query.scope) {
          this.dataSourceModel.meta = assetsUtil.setAssetMetaFromQuery(
            this.dataSourceModel.meta,
            this.$route.query as unknown as IScopeModel,
            this.authStore.orgUnitList,
          );
        }
      } catch (err: unknown) {
        this.$q.notify(alertUtil.getError("Failed to load secret by ID"));
        console.error(err);
        this.appStore.setFallback(true);
      } finally {
        this.appStore.setPageLoading(false);
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
    async onSubmit(secretVolume: SecretAssetCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        if (secretVolume.meta.scope !== Scope.Cluster) delete secretVolume.meta.clusterId;
        const savedSecretVolume: SecretAsset = await dataSourceService.createSecretVolume(secretVolume);
        this.$q.notify(alertUtil.getSuccess(`Data source ${savedSecretVolume.meta.name} created`));
        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: {
            createdEntityId: savedSecretVolume.meta.id,
            kind: AssetKind.SecretVolume,
            workloadFormName: this.workloadFormName,
          },
          params: paramId ? { id: paramId } : undefined,
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create secret"));
        }
      } finally {
        this.dataSourceStore.clearSelectedSecretVolume();
        this.submitting = false;
      }
    },
    getSecretVolumeCreationRequestFromAsset(fromSecret: SecretAsset): IUISecretCreationRequest {
      const { meta } = fromSecret;
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
        spec: deepCopy(fromSecret.spec),
      };
    },
    async createFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: SecretAsset = await this.dataSourceStore.loadSecretVolumeById(fromCopyId);
      this.dataSourceModel = this.getSecretVolumeCreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
    },
  },
});
</script>
