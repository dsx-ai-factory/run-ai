<template>
  <section class="workspace-new column items-center q-pt-md">
    <runai-form-wrapper :form-state="{ selectedTemplateId, workspaceName }" :enter-callback="onContinue">
      <q-form ref="workspaceForm" @submit.prevent="">
        <section class="create-workspace-form">
          <runai-expansion-wrapper>
            <runai-cluster-section
              entity-name="workload"
              :cluster-id="selectedClusterId"
              @update:cluster-id="updateClusterId"
            />

            <transition name="fade">
              <project-section
                v-if="selectedClusterId && isClusterSectionValid"
                aid="project-section"
                :loading="loadingProject"
                :selected-project-id="selectedProjectId"
                :projects="loadedProjects"
                :cluster-id="selectedClusterId"
                @project-changed="onSelectedProject"
              />
            </transition>

            <transition-group name="fade">
              <template v-if="selectedProjectId && selectedProjectId !== -1">
                <template-section-v2
                  v-if="showTemplatesV2"
                  aid="template-section-v2"
                  :templates="flatTemplates"
                  :total-templates-count="flatTotalCount"
                  :loading="loadingTemplate"
                  :list-query="flatTemplatesListQuery"
                  :template-id="selectedTemplateId || null"
                  @template-changed="onSelectedTemplate"
                  @query-changed="onTemplatesListQueryChanged"
                  @get-next-batch="onGetNextTemplatesBatch"
                />
                <template-section
                  v-else
                  aid="template-section"
                  :templates="templates"
                  :loading="loadingTemplate"
                  @template-changed="onSelectedTemplate"
                  :template-id="selectedTemplateId || null"
                />
                <workload-name-section
                  entity-type="workspace"
                  aid="ws-name-section"
                  v-model:name="workspaceName"
                  :disable-closing="true"
                  :project-id="selectedProjectId"
                  :cluster-id="selectedClusterId"
                  :reactive-rules="false"
                />
              </template>
            </transition-group>
          </runai-expansion-wrapper>
          <section class="row items-center q-mt-md">
            <q-field
              class="col-4 form-hint no-padding"
              :model-value="displayFormHint"
              :rules="[isFormIncomplete]"
            ></q-field>
            <div class="buttons q-ml-auto">
              <runai-btn flat class="q-mr-sm" color="primary" label="Cancel" aid="cancel-btn" @click="onCancel" />
              <quick-submission-dropdown
                v-if="showQuickSubmissionDropdown && selectedTemplateId"
                :workload-type="WorkloadType.Workspace"
                :is-template-runnable="isTemplateRunnable"
                :is-loading="isQuickSubmissionLoading"
                @submit="onQuickSubmission"
                @continue="onContinue"
                type="submit"
              />
              <runai-btn
                v-else
                label="continue"
                aid="create-workspace-continue-btn"
                @click="onContinue"
                :loading="submitting"
                color="primary"
                type="submit"
              />
            </div>
          </section>
        </section>
      </q-form>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";

// store
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useSettingStore } from "@/stores/setting.store";
import { useGuidedTourStore } from "@/stores/guided-tour.store";

// services
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { urlService } from "@/services/url.service/url.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { ProjectSection } from "@/components/section/project-section";
import { TemplateSection } from "@/components/section/template-section";
import { RunaiBtn } from "@/components/common/runai-btn";
import { TemplateSectionV2 } from "@/components/section/template-section-v2";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { QuickSubmissionDropdown } from "@/components/workload-creation/quick-submission-dropdown";
// models
import { type WorkloadTemplate, type PVCAsset, UidGidSource } from "@/swagger-models/assets-service-client";
import type { IUIWorkloadCreation } from "@/models/workload.model";
import type { Project, ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import { MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION } from "@/common/version.constant";
import { WorkloadType } from "@/swagger-models/workloads-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { ICardQuery } from "@/components/common/runai-card-list/runai-card-list.model";
// route
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// utils
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { deepCopy } from "@/utils/common.util";
import { SettingKeys } from "@/models/setting.model";
import { useFlatTemplatesGallery } from "@/composables/use-flat-templates-gallery.composable";
import { useQuickSubmission } from "@/composables/use-quick-submission.composable";

const WorkspaceNew = defineComponent({
  name: "workspace-new",
  components: {
    RunaiClusterSection,
    RunaiExpansionWrapper,
    RunaiFormWrapper,
    ProjectSection,
    WorkloadNameSection,
    TemplateSection,
    RunaiBtn,
    TemplateSectionV2,
    QuickSubmissionDropdown,
  },
  data() {
    return {
      workspaceStore: useWorkspaceStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      environmentStore: useEnvironmentStore(),
      authStore: useAuthStore(),
      settingStore: useSettingStore(),
      guidedTourStore: useGuidedTourStore(),

      workspaceForm: null as HTMLFormElement | null,
      loadedProjects: [] as Array<ProjectMetadata>,

      loadingProject: false as boolean,
      loadingTemplate: false as boolean,
      submitting: false as boolean,
      workspace: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,

      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      isClusterSectionValid: false as boolean,

      selectedProjectId: null as number | null,
      selectedTemplateId: null as string | null,
      selectedClusterId: "" as string,
      workspaceName: "" as string,

      WorkloadType,
    };
  },
  setup() {
    const { isLoading: isQuickSubmissionLoading, onQuickSubmit } = useQuickSubmission();
    const {
      list,
      templates: flatTemplates,
      totalCount: flatTotalCount,
      listQuery: flatTemplatesListQuery,
      offset: flatTemplatesOffset,
      setTemplates,
      setListQuery: setFlatTemplatesListQuery,
    } = useFlatTemplatesGallery();
    return {
      list,
      flatTemplates,
      flatTotalCount,
      flatTemplatesListQuery,
      flatTemplatesOffset,
      setTemplates,
      setFlatTemplatesListQuery,
      isQuickSubmissionLoading,
      onQuickSubmit,
    };
  },
  async created() {
    const selectedDefaultClusterId = this.getDefaultClusterId();
    await this.workspaceStore.loadActionPermissions(); // This must be before handleClusterDependencies, because it uses the actionPermissionsByScopes
    await this.handleClusterDependencies(selectedDefaultClusterId);

    this.appStore.setPageLoading(false);
  },
  async mounted() {
    this.workspaceForm = this.$refs["workspaceForm"] as HTMLFormElement;
    const createdEntityId: string | undefined = this.$route.query.createdEntityId?.toString();
    const savedProjectId = this.workspaceStore.workspace.projectId || -1;

    const selectedDefaultProjectId = this.getDefaultProjectId(savedProjectId, createdEntityId);
    if (selectedDefaultProjectId) {
      await this.onSelectedProject(selectedDefaultProjectId);
    }
  },
  computed: {
    templates(): Array<WorkloadTemplate> {
      return this.workloadTemplateStore.workloadTemplateList;
    },
    isHybridSupported(): boolean {
      return (
        this.clusterStore.isClusterVersionSufficient(
          this.selectedClusterId,
          MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
        ) && this.settingStore.isHybridSubmissionEnabled
      );
    },
    showTemplatesV2(): boolean {
      return this.isHybridSupported && this.settingStore.isFeatureEnabled(SettingKeys.EnableTemplatesV2);
    },
    showQuickSubmissionDropdown(): boolean {
      if (!this.showTemplatesV2 || !this.isHybridSupported) return false;

      // show the dropdown only if an actual template is selected
      return this.selectedTemplateId !== "-1" && !!this.selectedTemplateId;
    },
    isTemplateRunnable(): boolean {
      const noComplianceIssues = !this.selectedTemplate?.complianceIssues?.length;
      return !!this.selectedTemplate?.runnable?.isRunnable && noComplianceIssues;
    },
    selectedTemplate() {
      return this.flatTemplates.find((template) => template.meta.id === this.selectedTemplateId);
    },
  },
  methods: {
    getDefaultClusterId(): string {
      let selectedDefaultClusterId = "";
      if (this.clusterStore.clusterList.length === 1) {
        selectedDefaultClusterId = this.clusterStore.clusterList[0]?.uuid;
      } else if (this.$route.query.clusterId) {
        selectedDefaultClusterId = this.$route.query.clusterId.toString();
      } else if (this.workspaceStore.workspace.clusterId) {
        selectedDefaultClusterId = this.workspaceStore.workspace.clusterId;
      }

      return selectedDefaultClusterId;
    },
    getDefaultProjectId(savedProjectId: number, createdEntityId?: string): number | null {
      let selectedDefaultProjectId = null;
      if (createdEntityId && this.loadedProjects.length > 1) {
        selectedDefaultProjectId = Number(createdEntityId);
      } else if (savedProjectId !== -1) {
        selectedDefaultProjectId = savedProjectId;
      }

      return selectedDefaultProjectId;
    },
    async loadProjects(): Promise<void> {
      try {
        this.loadingProject = true;
        this.selectedTemplateId = null;

        this.loadedProjects = await orgUnitService.getProjectsMetadata(
          this.selectedClusterId,
          this.workspaceStore.actionPermissionsByScopes.create,
          this.authStore.orgUnitList,
        );

        if (this.loadedProjects.length === 1 && !this.guidedTourStore.isActive) {
          await this.onSelectedProject(+this.loadedProjects[0].id);
        } else if (this.workspaceStore.workspace.projectId !== -1) {
          await this.onSelectedProject(this.workspaceStore.workspace.projectId);
        }
      } catch (error: unknown) {
        this.$q.notify(alertUtil.getError("Failed to load projects"));
        console.error(error);
        this.appStore.setFallback(true);
      } finally {
        this.loadingProject = false;
      }
    },
    async onSelectedProject(projectId: number | null): Promise<void> {
      try {
        this.selectedProjectId = projectId || -1;
        this.selectedTemplateId = null;
        if (!projectId) return;

        await this.loadTemplates();
      } catch (error: unknown) {
        console.error(error);
      }
    },
    async loadWorkspaceModel(projectId: string, templateId?: string | null): Promise<void> {
      const tempId = templateId === "-1" ? undefined : templateId;
      const workspaceModel = await workspaceService.getWorkspaceMergedWithPolicy(
        projectId,
        undefined,
        tempId || undefined,
      );

      this.workspaceStore.setLoadedWorkspaceModelWithDefaults(deepCopy(workspaceModel));
      this.workspace = workloadUtil.convertWorkloadToWorkloadUI({
        ...workspaceModel,
        meta: {
          ...workspaceModel.meta,
          name: this.workspaceName,
          clusterId: this.selectedClusterId,
          projectId: this.selectedProjectId?.toString() || "",
        },
      });

      this.workspaceStore.setWorkspace(this.workspace);
    },
    async loadTemplates(applyOffset = false, amountRequired?: number): Promise<void> {
      try {
        this.loadingTemplate = true;

        if (this.showTemplatesV2) {
          if (!this.selectedProjectId) return;

          await this.list(WorkloadType.Workspace, this.selectedProjectId.toString(), applyOffset, amountRequired);

          if (!applyOffset) this.selectedTemplateId = null;
        } else {
          await this.workloadTemplateStore.loadWorkloadTemplates({
            projectId: this.selectedProjectId,
            isWorkspace: true,
          });
        }

        const templates = this.showTemplatesV2 ? this.flatTemplates : this.templates;

        if (templates.length === 0 && !this.selectedTemplateId) {
          this.selectedTemplateId = "-1";
        } else if (
          !this.showTemplatesV2 &&
          this.workspaceStore.templateId &&
          (this.workspaceStore.templateId === "-1" ||
            templates.find((template) => template.meta.id === this.workspaceStore.templateId))
        ) {
          this.selectedTemplateId = this.workspaceStore.templateId;
        }
      } catch (error: unknown) {
        // could happen if the user try to search for templates with non-english characters
        this.setTemplates([]);
        console.error(error);
      } finally {
        this.loadingTemplate = false;
      }
    },
    onSelectedTemplate(templateId: string | null): void {
      this.selectedTemplateId = templateId;
      this.workspaceStore.setTemplate(templateId);
    },
    redirectToPrevRoute(): void {
      this.workspaceStore.clearFlatFormState();
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    onContinue(): void {
      this.isHybridSupported ? this.continueToHybrid() : this.continueToEdit();
    },
    async continueToEdit(): Promise<void> {
      this.displayFormHint = false;
      this.submitting = true;
      const success = await this.validate();
      if (!success) {
        this.submitting = false;
        this.showHint();
        return;
      }

      if (!this.selectedProjectId) {
        console.error("ProjectId is missing");
        this.$q.notify(alertUtil.getError("Failed to create workspace"));
        this.submitting = false;
        return;
      }

      await this.loadWorkspaceModel(this.selectedProjectId.toString(), this.selectedTemplateId);

      const template: WorkloadTemplate | undefined = this.templates.find(
        (template) => template.meta.id === this.selectedTemplateId,
      );
      if (template?.spec.assets.workloadVolumes?.length) {
        const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(template.spec.assets.workloadVolumes);
        this.workspace.assets.uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
      }

      await this.setSpecificUserGidUidIfNeeded();

      this.workspaceStore.setWorkspace(this.workspace);
      await this.$router.push({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT });
    },
    async continueToHybrid(): Promise<void> {
      this.displayFormHint = false;
      this.submitting = true;
      const success = await this.validate();
      if (!success) {
        this.submitting = false;
        this.showHint();
        return;
      }

      if (!this.selectedProjectId) {
        console.error("ProjectId is missing");
        this.$q.notify(alertUtil.getError("Failed to create workspace"));
        this.submitting = false;
        return;
      }

      await this.$router.push({
        name: WORKSPACE_ROUTE_NAMES.WORKSPACE_HYBRID,
        query: {
          clusterId: this.selectedClusterId,
          projectId: this.selectedProjectId,
          templateId: this.selectedTemplateId !== "-1" ? this.selectedTemplateId : undefined,
          workspaceName: this.workspaceName,
        },
      });
    },
    validate(): Promise<boolean> {
      return (this.workspaceForm as HTMLFormElement).validate();
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
    findProject(projectId: string): ProjectMetadata | undefined {
      return this.loadedProjects.find((project: ProjectMetadata) => project.id === projectId);
    },
    async loadProjectById(projectId: string): Promise<Project | void> {
      try {
        const loadedProject = await orgUnitService.getProject(projectId);
        return loadedProject;
      } catch (error: unknown) {
        console.error(error);
        this.appStore.setFallback(true);
      }
    },
    async setSpecificUserGidUidIfNeeded(): Promise<void> {
      if (this.workspace.assets.environment) {
        const relevantEnvironment = await this.environmentStore.loadById(this.workspace.assets.environment);
        if (relevantEnvironment?.spec.uidGidSource === UidGidSource.FromIdpToken) {
          this.workspace.specificEnv.runAsUid = this.authStore.getUID || null;
          this.workspace.specificEnv.runAsGid = this.authStore.getGID || null;
          this.workspace.specificEnv.supplementalGroups = this.authStore.getSupplementaryGroups;
        }
      }
    },
    async updateClusterId(clusterId: string, payload: { disabled: boolean }): Promise<void> {
      this.displayFormHint = false;
      this.selectedClusterId = clusterId;
      this.isClusterSectionValid = !payload.disabled && this.canSubmitWorkload(clusterId);
      this.selectedProjectId = -1;
      urlService.updateQueryParams({ clusterId, createdEntityId: null });
      if (this.isClusterSectionValid) {
        await this.loadProjects();
      }
    },
    findClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    async handleClusterDependencies(clusterId: string): Promise<void> {
      if (!clusterId) return;
      if (this.canSubmitWorkload(clusterId)) {
        this.updateClusterId(clusterId, { disabled: false });
      }
    },
    canSubmitWorkload(clusterId: string): boolean {
      if (!clusterId) return false;
      const currCluster = this.findClusterById(clusterId);
      if (!currCluster) return false;
      return workloadService.canSubmitWorkload(currCluster) === null;
    },
    onTemplatesListQueryChanged(query: ICardQuery): void {
      this.setFlatTemplatesListQuery(query);
      this.loadTemplates();
    },
    onGetNextTemplatesBatch(amountRequired: number): void {
      this.loadTemplates(true, amountRequired);
    },
    async onQuickSubmission(): Promise<void> {
      this.displayFormHint = false;

      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }

      await this.onQuickSubmit(WorkloadType.Workspace, {
        name: this.workspaceName,
        projectId: this.selectedProjectId?.toString() || "",
        templateId: this.selectedTemplateId || "",
        clusterId: this.selectedClusterId,
      });
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});

//Dima todo - Type assertion to prevent TS7056 while maintaining type safety -- will be refactored later --hide the annoying warning
export default WorkspaceNew as typeof WorkspaceNew & DefineComponent;
</script>
