<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="data-volume-new column items-center q-pt-md">
    <data-volume-create-form
      v-if="!isPageLoading"
      :allowed-scopes="allowedCreateScopes"
      @submit="createDataVolume"
      @cancel="onCancel"
      :submitting="submitting"
      :model="model"
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
import { Action, ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IDataVolumeCreationRequest } from "@/models/data-source.model";
import type { DatavolumeCreationFields, SharedScopeBase } from "@/swagger-models/datavolumes-service-client";
// Utils
import { alertUtil } from "@/utils/alert.util";
import { DATA_VOLUME_ROUTE_NAMES } from "@/router/data-volume.routes/data-volume.routes.names";
// Services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { dataVolumeService } from "@/services/control-plane/data-volume.service/data-volume.service";

export default defineComponent({
  name: "data-volume-new",
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
    };
  },
  async created() {
    const fromCopyId = this.$route.query.fromCopyId as string;
    if (fromCopyId) {
      this.model = await this.getDataVolumeModelForCopy(fromCopyId);
    }
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
    async getDataVolumeModelForCopy(id: string): Promise<IDataVolumeCreationRequest> {
      const [sharedScopes, dataVolume] = await Promise.all([
        dataVolumeService.getSharedScopesById(id),
        dataVolumeService.getDataVolumeById(id),
      ]);

      return {
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
    },
    async createDataVolume(dataVolumeModel: IDataVolumeCreationRequest) {
      let createdDataVolumeId: string;
      try {
        this.submitting = true;
        const dataVolumeForRequest: DatavolumeCreationFields = {
          name: dataVolumeModel.meta.name,
          description: dataVolumeModel.meta.description,
          projectId: dataVolumeModel.meta.projectId,
          originPvcName: dataVolumeModel.spec.originPvcName,
        };
        const createdDataVolume = await dataVolumeService.createDataVolume(dataVolumeForRequest);
        createdDataVolumeId = createdDataVolume.id;
        this.$q.notify(alertUtil.getSuccess(`Data volume ${createdDataVolume.name} created`));
        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_INDEX,
          query: { workloadFormName: this.workloadFormName, createdEntityId: createdDataVolumeId },
          params: paramId ? { id: paramId } : undefined,
        });
        // eslint-disable-next-line
      } catch (err: any) {
        this.$q.notify(alertUtil.getError(`${err.message ? err.message : "Failed to create data volume"}`));
        console.error(err);
        return;
      } finally {
        this.submitting = false;
      }

      try {
        if (!createdDataVolumeId) return;
        if (dataVolumeModel.spec.sharedScopes.length > 0) {
          dataVolumeModel.spec.sharedScopes = dataVolumeModel.spec.sharedScopes.filter(
            (scope: SharedScopeBase) =>
              !(scope.scopeType === ScopeType.Project && scope.scopeId === dataVolumeModel.meta.projectId),
          );
          await dataVolumeService.patchDataVolumeSharedScopes(createdDataVolumeId, {
            add: dataVolumeModel.spec.sharedScopes,
          });
        }
        // eslint-disable-next-line
      } catch (err: any) {
        this.$q.notify(alertUtil.getError(`${err.message ? err.message : "Failed to share data volume"}`));
        console.error(err);
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
  },
});
</script>
