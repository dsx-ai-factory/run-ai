<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="nfs-new column items-center q-pt-md">
    <nfs-create-form
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

// services
import { alertUtil } from "@/utils/alert.util";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// cmps
import { NfsCreateForm } from "@/components/data-source/nfs/nfs-create-form";

// models
import { type NFSCreationRequest, type NFSAsset, AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { deepCopy } from "@/utils/common.util";
import { useAppStore } from "@/stores/app.store";
import type { IUINFSCreationRequest } from "@/models/data-source.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { IScopeModel } from "@/models/global.model";

export default defineComponent({
  components: {
    NfsCreateForm,
  },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      dataSourceModel: null as IUINFSCreationRequest | null,
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
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.NfsAssets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptyNfsModel();
        const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
        if (fromCopyId) await this.loadFromExistingDS(fromCopyId);
        if (this.$route.query.scope) {
          this.dataSourceModel.meta = assetsUtil.setAssetMetaFromQuery(
            this.dataSourceModel.meta,
            this.$route.query as unknown as IScopeModel,
            this.authStore.orgUnitList,
          );
        }
      } catch (err: unknown) {
        this.$q.notify(alertUtil.getError("Failed to load NFS by ID"));
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
    async onSubmit(uiNFS: NFSCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const NFSRequest: NFSCreationRequest = this.readyDataForBackend(uiNFS);
        const savedNFS: NFSAsset = await dataSourceService.createNFS(NFSRequest);
        this.$q.notify(alertUtil.getSuccess(`Data source ${savedNFS.meta.name} created`));

        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedNFS.meta.id, kind: AssetKind.Nfs, workloadFormName: this.workloadFormName },
          params: paramId ? { id: paramId } : undefined,
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create nfs"));
        }
      } finally {
        this.dataSourceStore.clearSelectedNFS();
        this.submitting = false;
      }
    },
    readyDataForBackend(uiModel: NFSCreationRequest): NFSCreationRequest {
      const { meta, spec } = uiModel;
      if (meta.scope !== Scope.Cluster) delete meta.clusterId;
      return {
        meta,
        spec: { ...spec, path: spec.path || null, server: spec.server || null, mountPath: spec.mountPath || null },
      };
    },
    getNFSCreationRequestFromAsset(fromNFS: NFSAsset): IUINFSCreationRequest {
      const { meta } = fromNFS;
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
        spec: deepCopy(fromNFS.spec),
      };
    },
    async loadFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: NFSAsset = await this.dataSourceStore.loadNFSById(fromCopyId);
      this.dataSourceModel = this.getNFSCreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
    },
  },
});
</script>

<style lang="scss" scoped></style>
