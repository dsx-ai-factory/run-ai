<template>
  <section class="git-new column items-center q-pt-md">
    <git-create-form
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
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { alertUtil } from "@/utils/alert.util";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";

// cmps
import { GitCreateForm } from "@/components/data-source/git/git-create-form";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { assetsUtil } from "@/utils/assets.util/assets.util";
// models
import type { IUIGitCreationRequest } from "@/models/data-source.model";
import { type GitAsset, type GitCreationRequest, AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IScopeModel } from "@/models/global.model";

export default defineComponent({
  components: { GitCreateForm },
  data() {
    return {
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      appStore: useAppStore(),
      workloadFormName: this.$route.query.workloadFormName
        ? this.$route.query.workloadFormName
        : (null as string | null),
      dataSourceModel: null as IUIGitCreationRequest | null,
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
      return this.dataSourceStore.dataSourcesActionPermissions[ResourceType.GitAssets]?.[Action.Create];
    },
  },
  methods: {
    async init(): Promise<void> {
      try {
        this.dataSourceModel = dataSourceService.getEmptyGitModel();
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
        this.$q.notify(alertUtil.getError("Failed to load git asset by ID"));
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
    async onSubmit(uiGit: GitCreationRequest): Promise<void> {
      try {
        this.submitting = true;
        const gitRequest: GitCreationRequest = this.readyDataForBackend(uiGit);
        const savedGit: GitAsset = await dataSourceService.createGit(gitRequest);
        this.$q.notify(alertUtil.getSuccess(`Data source ${savedGit.meta.name} created`));

        const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;
        this.$router.push({
          name: this.$route.meta.backPageName as string,
          query: { createdEntityId: savedGit.meta.id, kind: AssetKind.Git, workloadFormName: this.workloadFormName },
          params: paramId ? { id: paramId } : undefined,
        });
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to create git"));
        }
      } finally {
        this.dataSourceStore.clearSelectedGit();
        this.submitting = false;
      }
    },
    readyDataForBackend(uiModel: GitCreationRequest): GitCreationRequest {
      const { meta, spec } = uiModel;
      if (meta.scope !== Scope.Cluster) delete meta.clusterId;
      return { meta, spec: { ...spec, path: spec.path || null, repository: spec.repository || null } };
    },
    getGitCreationRequestFromAsset(fromGitAsset: GitAsset): IUIGitCreationRequest {
      const { meta, spec } = fromGitAsset;
      const isAllowedScope: boolean = orgTreeService.isAllowedScope(
        meta,
        this.allowedCreateScopes,
        this.authStore.orgUnitList,
      );

      // The API support only one of the following: branch or revision.
      // In the UI we show only one input which will show the revision.
      const git: IUIGitCreationRequest = {
        meta: {
          name: meta.name,
          description: meta.description,
          scope: isAllowedScope ? meta.scope : null,
          projectId: isAllowedScope ? meta.projectId : null,
          departmentId: isAllowedScope ? meta.departmentId : null,
          clusterId: isAllowedScope ? meta.clusterId : null,
        },
        spec: {
          ...spec,
        },
      };

      if (spec.branch && !spec.revision) {
        git.spec.revision = spec.branch;
      } else {
        git.spec.revision = spec.revision || undefined;
      }
      delete git.spec.branch;

      return git;
    },
    async loadFromExistingDS(fromCopyId: string): Promise<void> {
      const loadedDS: GitAsset = await this.dataSourceStore.loadGitById(fromCopyId);
      this.dataSourceModel = this.getGitCreationRequestFromAsset(loadedDS);

      if (this.dataSourceModel.meta.scope === Scope.Tenant) this.dataSourceModel.meta.scope = null;
    },
  },
});
</script>

<style lang="scss" scoped></style>
