<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="pvc-new column items-center q-pt-md">
    <pvc-create-form
      v-if="!isPageLoading && dataSourceModel"
      :data-source-model="dataSourceModel"
      :submitting="submitting"
      :is-scope-selectable="!$route.query.scope"
      :init-claim-info-from-scratch="initClaimInfoFromScratch"
      :allowed-scopes="allowedCreateScopes"
      @submit="onSubmit"
      @cancel="onCancel"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// services
import { alertUtil } from "@/utils/alert.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { deepCopy } from "@/utils/common.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";

// cmps
import { PvcCreateForm } from "@/components/data-source/pvc/pvc-create-form";

// Stores
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import type { IScopeModel } from "@/models/global.model";

// models
import {
  type PVCCreationRequest,
  type PVCAsset,
  AssetKind,
  Scope,
  type AssetCreationRequest,
} from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IUIPVCCreationRequest } from "@/models/data-source.model";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
export default defineComponent({
  components: {
    PvcCreateForm,
  },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      dataSourceModel: null as IUIPVCCreationRequest | null,
      submitting: false as boolean,
      initClaimInfoFromScratch: true as boolean,
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
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.PvcAssets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptyPvcModel();
        const fromCopyId: string = this.$route.query.fromCopyId?.toString() || "";
        if (fromCopyId) await this.loadFromExistingDS(fromCopyId);
        if (this.$route.query.scope) {
          this.dataSourceModel.meta = assetsUtil.setAssetMetaFromQuery(
            this.dataSourceModel.meta,
            this.$route.query as unknown as IScopeModel,
            this.authStore.orgUnitList,
          );
        }
      } catch (err) {
        console.error(err);
        this.$q.notify(alertUtil.getError("Failed to load form's details"));
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
    async onSubmit(uiPVC: PVCCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const PVCRequest: PVCCreationRequest = this.readyDataForBackend(uiPVC);
        const savedPVC: PVCAsset = await dataSourceService.createPVC(PVCRequest);
        this.$q.notify(alertUtil.getSuccess(`Data source ${savedPVC.meta.name} created`));

        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedPVC.meta.id, kind: AssetKind.Pvc, workloadFormName: this.workloadFormName },
          params: paramId ? { id: paramId } : undefined,
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create PVC"));
        }
      } finally {
        this.dataSourceStore.clearSelectedPVC();
        this.submitting = false;
      }
    },
    readyDataForBackend(uiModel: PVCCreationRequest): PVCCreationRequest {
      const { meta, spec } = uiModel;
      if (meta.scope !== Scope.Cluster) delete meta.clusterId;
      const claimName: string | null =
        spec.existingPvc && spec.claimName ? spec.claimName : this.generateClaimName(meta);
      if (spec.claimInfo?.addedAttrValues?.length) {
        spec.claimInfo.addedAttrValues = spec.claimInfo.addedAttrValues.filter((att) => !!att.value);
      }

      return {
        meta,
        spec: { ...spec, path: spec.path || null, claimName },
      };
    },
    generateClaimName(meta: AssetCreationRequest): string {
      let namePrefix: string = meta.name;
      switch (meta.scope) {
        case Scope.Tenant: {
          namePrefix += `-${this.authStore.tenant.name}`;
          break;
        }
        case Scope.Cluster: {
          const name = this.authStore.orgUnitList.find((uo: OrgUnit) => {
            return uo.id === meta.clusterId && uo.type === meta.scope;
          })?.name;
          namePrefix += `-${name || meta.scope}`;
          break;
        }
        case Scope.Department: {
          const name = this.authStore.orgUnitList.find((uo: OrgUnit) => {
            return uo.id === meta.departmentId && uo.type === meta.scope;
          })?.name;
          namePrefix += `-${name || meta.scope}`;
          break;
        }
        case Scope.Project: {
          const name = this.authStore.orgUnitList.find((uo: OrgUnit) => {
            return uo.id === (meta.projectId?.toString || "") && uo.type === meta.scope;
          })?.name;
          namePrefix += `-${name || meta.scope}`;
          break;
        }
        default: {
          namePrefix += `-${meta.scope}`;
          break;
        }
      }
      return dataSourceUtil.getRandomClaimName(namePrefix);
    },
    getPvcCreationRequestFromAsset(fromPVCAsset: PVCAsset): IUIPVCCreationRequest {
      const { meta } = fromPVCAsset;
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
        spec: deepCopy(fromPVCAsset.spec),
      };
    },
    async loadFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: PVCAsset = await this.dataSourceStore.loadPVCById(fromCopyId);
      this.dataSourceModel = this.getPvcCreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
      this.initClaimInfoFromScratch = false;
    },
  },
});
</script>

<style lang="scss" scoped></style>
