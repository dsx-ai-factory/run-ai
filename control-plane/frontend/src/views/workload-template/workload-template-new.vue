<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-template-new">
    <workload-template-create-form
      v-if="!isPageLoading && workloadTemplate"
      @submit="onSubmit"
      @cancel="onCancel"
      :initial-workload-template="workloadTemplate"
      :allowed-scopes="allowedCreateScopes"
      :submitting="submitting"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// Components
import { WorkloadTemplateCreateForm } from "@/components/workload-template/workload-template-create-form";

// Services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";

// Stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";

// Models
import {
  AssetKind,
  type WorkloadTemplate,
  type WorkloadTemplateCreationRequest,
  type DatasourceRef,
  Scope,
} from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action, ScopeType } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import type { IUIWorkloadTemplateCreationRequestModel } from "@/models/workload-template.model";

// utils
import type { IScopeInfo } from "@/utils/data-source.util/data-source.util";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

export default defineComponent({
  components: {
    WorkloadTemplateCreateForm,
  },
  provide() {
    return {
      policy: {},
    };
  },
  data() {
    return {
      authStore: useAuthStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      workloadTemplate: null as IUIWorkloadTemplateCreationRequestModel | null,
      submitting: false as boolean,
    };
  },
  async created() {
    this.workloadTemplate = workloadTemplateService.getEmptyWorkloadTemplateModel();

    const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
    if (fromCopyId) {
      await this.loadFromExistingWorkloadTemplate(fromCopyId);
    } else if (this.workloadTemplateStore.workloadTemplateCreation) {
      this.workloadTemplate = this.workloadTemplateStore.workloadTemplateCreation;
    }

    const { kind, createdEntityId } = this.$route.query;

    const createdEnvironmentId: string | undefined =
      kind === AssetKind.Environment ? createdEntityId?.toString() : undefined;
    if (createdEnvironmentId) {
      this.workloadTemplate.spec.assets.environment = createdEnvironmentId;
    }

    const createdComputeId: string | undefined = kind === AssetKind.Compute ? createdEntityId?.toString() : undefined;
    if (createdComputeId) {
      this.workloadTemplate.spec.assets.compute = createdComputeId;
    }

    const isDataSource = [
      AssetKind.HostPath,
      AssetKind.Nfs,
      AssetKind.Git,
      AssetKind.S3,
      AssetKind.Pvc,
      AssetKind.ConfigMap,
      AssetKind.SecretVolume,
    ].some((assetKind) => assetKind === kind);

    const createdDataSourceId: string | undefined = isDataSource ? createdEntityId?.toString() : undefined;
    if (createdDataSourceId) {
      const kind: AssetKind = this.$route.query.kind as AssetKind;
      const dataSourceId: string = createdDataSourceId;
      if (this.workloadTemplate.spec.assets.datasources) {
        this.workloadTemplate.spec.assets.datasources.push({ id: dataSourceId, kind: kind });
      } else {
        this.workloadTemplate.spec.assets.datasources = [{ id: dataSourceId, kind: kind }];
      }
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    isPageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    allowedCreateScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
  },
  methods: {
    async loadFromExistingWorkloadTemplate(fromCopyId: string): Promise<void> {
      try {
        const workloadTemplate: WorkloadTemplate = await this.workloadTemplateStore.loadById(fromCopyId);
        this.workloadTemplate = this.prepareWorkloadTemplate(workloadTemplate);

        if (this.workloadTemplate.meta.scope === ScopeType.Tenant) this.workloadTemplate.meta.scope = null;

        if (workloadTemplate.spec.assets.workloadVolumes?.length) {
          this.workloadTemplateStore.loadVolumesFromExisting(workloadTemplate.spec.assets.workloadVolumes);
        }
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Failed to load template`));
        }
        this.appStore.setFallback(true);
      }
    },
    prepareWorkloadTemplate(workloadTemplate: WorkloadTemplate): IUIWorkloadTemplateCreationRequestModel {
      const { meta } = workloadTemplate;
      const isAllowedScope: boolean = orgTreeService.isAllowedScope(
        meta,
        this.allowedCreateScopes,
        this.authStore.orgUnitList,
      );
      const workloadTemplateCreationRequest: IUIWorkloadTemplateCreationRequestModel = {
        meta: {
          name: meta.name,
          description: meta.description,
          scope: isAllowedScope ? meta.scope : null,
          projectId: isAllowedScope ? meta.projectId : null,
          departmentId: isAllowedScope ? meta.departmentId : null,
          clusterId: isAllowedScope ? meta.clusterId : null,
        },
        spec: {
          assets: {
            environment: workloadTemplate.spec.assets.environment.id,
            compute: workloadTemplate.spec.assets.compute?.id,
            datasources: workloadTemplate.spec.assets.datasources?.map((ds: DatasourceRef) => ({
              id: ds.id,
              kind: ds.kind,
            })),
          },
          specificEnv: workloadTemplate.spec.specificEnv,
        },
      };

      return workloadTemplateCreationRequest;
    },
    async createWorkloadVolumes(workloadTemplate: WorkloadTemplateCreationRequest): Promise<Array<string> | undefined> {
      let workloadVolumes: Array<string> | undefined;
      const scopeInfo: IScopeInfo = {
        scope: workloadTemplate.meta.scope,
        projectId: workloadTemplate.meta.projectId || undefined,
        departmentId: workloadTemplate.meta.departmentId || undefined,
        clusterId: workloadTemplate.meta.clusterId || undefined,
      };
      if (this.workloadTemplateStore.workloadVolumes?.length) {
        workloadVolumes = await dataSourceService.createWorkloadVolumes(
          workloadTemplate.meta.name,
          this.workloadTemplateStore.workloadVolumes,
          scopeInfo,
        );
      }

      return workloadVolumes;
    },
    async onSubmit(workloadTemplate: WorkloadTemplateCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        if (workloadTemplate.meta.scope !== Scope.Cluster) delete workloadTemplate.meta.clusterId;
        workloadTemplate.spec.assets.workloadVolumes = await this.createWorkloadVolumes(workloadTemplate);
        const savedWorkloadTemplate: WorkloadTemplate = await workloadTemplateService.create(workloadTemplate);
        this.$q.notify(alertUtil.getSuccess(`Template ${workloadTemplate.meta.name} created`));
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedWorkloadTemplate.meta.id, kind: AssetKind.WorkloadTemplate },
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create template"));
        }
      } finally {
        this.submitting = false;
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      allowToLeave && this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX });
    },
  },
});
</script>
