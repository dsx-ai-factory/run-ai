<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="config-map-new column items-center q-pt-md">
    <config-map-create-form
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

// cmps
import ConfigMapCreateForm from "@/components/data-source/config-map/config-map-create-form/config-map-create-form.vue";

// services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { deepCopy } from "@/utils/common.util";
import { useAppStore } from "@/stores/app.store";
import { assetsUtil } from "@/utils/assets.util/assets.util";
// models
import type { IScopeModel } from "@/models/global.model";
import type { IUIConfigMapCreationRequest } from "@/models/data-source.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import {
  Scope,
  type ConfigMapCreationRequest,
  type ConfigMapAsset,
  AssetKind,
} from "@/swagger-models/assets-service-client";

export default defineComponent({
  components: {
    ConfigMapCreateForm,
  },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      dataSourceModel: null as IUIConfigMapCreationRequest | null,
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
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.CmVolumeAssets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptyConfigMapModel();
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
        this.$q.notify(alertUtil.getError("Failed to load config map by ID"));
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
    async onSubmit(uiConfigMap: ConfigMapCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const configMapRequest: ConfigMapCreationRequest = this.readyDataForBackend(uiConfigMap);
        const savedConfigMap: ConfigMapAsset = await dataSourceService.createConfigMap(configMapRequest);

        this.$q.notify(alertUtil.getSuccess(`Data source ${savedConfigMap.meta.name} created`));
        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: {
            createdEntityId: savedConfigMap.meta.id,
            kind: AssetKind.ConfigMap,
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
          this.$q.notify(alertUtil.getError("Failed to create config map"));
        }
      } finally {
        this.dataSourceStore.clearSelectedConfigMap();
        this.submitting = false;
      }
    },
    readyDataForBackend(uiModel: ConfigMapCreationRequest): ConfigMapCreationRequest {
      const { meta, spec } = uiModel;
      if (meta.scope !== Scope.Cluster) delete meta.clusterId;
      return {
        meta,
        spec: {
          ...spec,
          configMap: spec.configMap || null,
          mountPath: spec.mountPath || null,
          subPath: spec.subPath || null,
        },
      };
    },
    getConfigMapCreationRequestFromAsset(fromConfigMap: ConfigMapAsset): IUIConfigMapCreationRequest {
      const { meta } = fromConfigMap;
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
        spec: deepCopy(fromConfigMap.spec),
      };
    },
    async createFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: ConfigMapAsset = await this.dataSourceStore.loadConfigMapById(fromCopyId);
      this.dataSourceModel = this.getConfigMapCreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
    },
  },
});
</script>

<style lang="scss" scoped></style>
