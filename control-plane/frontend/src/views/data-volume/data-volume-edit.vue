<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="data-volume-edit column items-center q-pt-md">
    <data-volume-create-form
      v-if="!isPageLoading"
      :allowed-scopes="allowedCreateScopes"
      @submit="editDataVolume"
      @cancel="onCancel"
      :submitting="submitting"
      :model="model"
      edit-mode
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Components
import { DataVolumeCreateForm } from "@/components/data-volume/data-volume-create-form";
// Stores
import { useAppStore } from "@/stores/app.store";
import { useDataVolumeStore } from "@/stores/data-volume.store";
// Models
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IDataVolumeCreationRequest } from "@/models/data-source.model";
import type {
  DatavolumePatchFields,
  SharedScope,
  SharedScopeBase,
  SharedScopesPatchRequest,
} from "@/swagger-models/datavolumes-service-client";
// Utils
import { alertUtil } from "@/utils/alert.util";
import { DATA_VOLUME_ROUTE_NAMES } from "@/router/data-volume.routes/data-volume.routes.names";
// Services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";

export default defineComponent({
  name: "data-volume-edit",
  components: {
    DataVolumeCreateForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      dataVolumeStore: useDataVolumeStore(),
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      submitting: false as boolean,
      model: dataVolumeService.getEmptyDataVolumeModel(),
      originalSharedScopes: [] as Array<SharedScope>,
    };
  },
  async created() {
    await this.loadDataVolumeById();
    this.appStore.setPageLoading(false);
  },
  computed: {
    allowedCreateScopes(): PermittedScopes {
      return this.dataVolumeStore.dataVolumeActionPermissions?.[Action.Create] || {};
    },
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
  },
  methods: {
    async loadDataVolumeById(): Promise<void> {
      try {
        const id = this.$route.params.id?.toString();
        const [sharedScopes, dataVolume] = await Promise.all([
          dataVolumeService.getSharedScopesById(id),
          dataVolumeService.getDataVolumeById(id),
        ]);

        this.originalSharedScopes = sharedScopes;

        this.model = {
          meta: {
            name: dataVolume.name,
            description: dataVolume.description,
            projectId: dataVolume.projectId,
          },
          spec: {
            originPvcName: dataVolume.originPvcName,
            sharedScopes,
          },
        };
        this.appStore.setPageTitleOverride(`Edit data volume: ${dataVolume.name}`);
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError("Failed to load data volume"));
      }
    },
    async editDataVolume(dataVolumeModel: IDataVolumeCreationRequest) {
      try {
        this.submitting = true;
        const dataVolumeForRequest: DatavolumePatchFields = {
          description: dataVolumeModel.meta.description,
        };

        const dataVolumeSharedScopesForRequest: SharedScopesPatchRequest = this.calculateSharedScopesChanges(
          this.originalSharedScopes,
          dataVolumeModel.spec.sharedScopes,
        );

        const [editedDataVolume] = await Promise.all([
          dataVolumeService.patchDataVolume(this.$route.params.id?.toString() || "", dataVolumeForRequest),
          dataVolumeService.patchDataVolumeSharedScopes(
            this.$route.params.id?.toString() || "",
            dataVolumeSharedScopesForRequest,
          ),
        ]);

        this.$q.notify(alertUtil.getSuccess(`Data volume ${editedDataVolume.name} saved`));
        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX,
          query: { workloadFormName: this.workloadFormName, createdEntityId: editedDataVolume.id },
          params: paramId ? { id: paramId } : undefined,
        });
        // eslint-disable-next-line
      } catch (err: any) {
        this.$q.notify(alertUtil.getError(`${err.message ? err.message : "Failed to edit data volume"}`));
        console.error(err);
      } finally {
        this.submitting = false;
      }
    },
    calculateSharedScopesChanges(
      originalScopes: SharedScope[],
      editedSharedScopes: SharedScopeBase[],
    ): SharedScopesPatchRequest {
      const originalScopeIds = new Set(originalScopes.map((scope) => scope.scopeId));
      const editedScopeIds = new Set(editedSharedScopes.map((scope) => scope.scopeId));

      return {
        remove: originalScopes
          .filter((scope) => !editedScopeIds.has(scope.scopeId))
          .map((scope): SharedScopeBase => ({ scopeId: scope.scopeId, scopeType: scope.scopeType })),
        add: editedSharedScopes.filter((scope) => !originalScopeIds.has(scope.scopeId)),
      };
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
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>
