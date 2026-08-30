<template>
  <section class="s3-new column items-center q-pt-md">
    <s3-create-form
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
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { deepCopy } from "@/utils/common.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";

// cmps
import { S3CreateForm } from "@/components/data-source/s3/s3-create-form";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import type { IUIS3CreationRequest } from "@/models/data-source.model";
import type { IScopeModel } from "@/models/global.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";

// models
import { type S3Asset, type S3CreationRequest, AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";

export default defineComponent({
  components: { S3CreateForm },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      dataSourceModel: null as IUIS3CreationRequest | null,
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
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.S3Assets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptyS3Model();
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
        this.$q.notify(alertUtil.getError("Failed to load S3 by ID"));
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
    async onSubmit(uiS3: S3CreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const s3Request: S3CreationRequest = this.readyDataForBackend(uiS3);
        const savedS3: S3Asset = await dataSourceService.createS3(s3Request);
        this.$q.notify(alertUtil.getSuccess(`Data source ${savedS3.meta.name} created`));

        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedS3.meta.id, kind: AssetKind.S3, workloadFormName: this.workloadFormName },
          params: paramId ? { id: paramId } : undefined,
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create s3"));
        }
      } finally {
        this.dataSourceStore.clearSelectedS3();
        this.submitting = false;
      }
    },
    readyDataForBackend(uiModel: S3CreationRequest): S3CreationRequest {
      const { meta, spec } = uiModel;
      if (meta.scope !== Scope.Cluster) delete meta.clusterId;
      return { meta, spec: { ...spec, path: spec.path || null, bucket: spec.bucket || null } };
    },
    getS3CreationRequestFromAsset(fromS3Asset: S3Asset): IUIS3CreationRequest {
      const { meta } = fromS3Asset;
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
        spec: deepCopy(fromS3Asset.spec),
      };
    },
    async loadFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: S3Asset = await this.dataSourceStore.loadS3ById(fromCopyId);
      this.dataSourceModel = this.getS3CreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
    },
  },
});
</script>

<style lang="scss" scoped></style>
