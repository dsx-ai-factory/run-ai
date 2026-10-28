<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-yaml-new column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :form-state="{ yamlContent, selectedProjectId }"
      :submitting="submitting"
      submit-btn-text="CREATE WORKLOAD"
      cancel-btn-text="Cancel"
      @leave-page="redirectToPrevRoute"
      @canceled="onCancel"
      @submit="onSubmit"
    >
      <template v-if="!appStore.isPageLoading">
        <runai-expansion-wrapper>
          <runai-cluster-section
            entity-name="workload"
            :cluster-id="selectedClusterId"
            :disabled-cluster-options="disabledClusterOptions"
            @update:cluster-id="onClusterChanged"
          />

          <transition name="fade">
            <yaml-workload-section
              v-if="selectedClusterId && isClusterSectionValid"
              :yaml="yamlContent"
              :cluster-id="selectedClusterId"
              :selected-form-values="selectedFormValues"
              @yaml-input="onYamlInput"
              @validation-changed="onYamlValidationChanged"
              @yaml-loaded="onYamlLoaded"
              @dryrun-success="onDryRunSuccess"
            />
          </transition>

          <transition name="fade">
            <div v-if="showFormSections">
              <project-section
                :loading="loadingProject"
                :selected-project-id="selectedProjectId"
                :projects="loadedProjects"
                :cluster-id="selectedClusterId"
                :show-override-hint="showProjectOverrideHint"
                :disable-closing="false"
                @project-changed="onSelectedProject"
                @before-create-project="saveFormState"
              />
              <general-section
                :workload-form-type="workloadFormType"
                :single-pod-workload="false"
                :general-model="generalModel"
                :support-preemptibility="supportPreemptibility"
                :support-mnnvl="supportMnnvl"
                :workload-priority-data="workloadPriorityData"
                :show-annotations-and-labels="false"
                :show-auto-deletion-section="false"
                :override-hints="overrideHints"
                @general-model-changed="onGeneralModelChanged"
              />
            </div>
          </transition>
        </runai-expansion-wrapper>
      </template>
    </workload-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";

// Composables
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { useRoute } from "vue-router";

// Store
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadStore } from "@/stores/workload.store";

// Services
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert, extractErrorMessages } from "@/utils/error-alert.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { urlService } from "@/services/url.service/url.service";
import { yamlUtil } from "@/utils/yaml.util";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { filterUtil } from "@/utils/filter.util/filter.util";

// Components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { ProjectSection } from "@/components/section/project-section";
import { GeneralSection } from "@/components/section/general-section";
import { YamlWorkloadSection, type ISelectedFormValues } from "@/components/workload/yaml-workload-section";

// Models
import type { ProjectMetadata, Project } from "@/swagger-models/org-unit-service-client";
import type { Preemptibility } from "@/models/workload.model";
import { EWorkloadFormType } from "@/models/workload.model";
import type {
  IOverrideHints,
  IUIGeneralSectionModel,
  IWorkloadPriorityData,
} from "@/components/section/general-section/general-section.model";
import { Action, ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { MnnvlMode, type WorkloadV2CreateRequest, type WorkloadV2 } from "@/swagger-models/workloads-manager-client";
import { NodepoolMinimalSortFields, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";

// Shared interface for override flags (project, priority, preemptibility)
export interface IOverrideFlags {
  project: boolean;
  priority: boolean;
  preemptibility: boolean;
}

// Interface for values extracted from YAML
export interface IYamlExtractedValues {
  namespace: string | null;
  priority: string | null;
  preemptibility: Preemptibility | null;
}

// Routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// Constants
import {
  MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
} from "@/common/version.constant";

const WorkloadYamlNew = defineComponent({
  name: "workload-yaml-new",
  components: {
    WorkloadFormWrapper,
    RunaiExpansionWrapper,
    RunaiClusterSection,
    ProjectSection,
    GeneralSection,
    YamlWorkloadSection,
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      appStore: useAppStore(),
      authStore: useAuthStore(),
      workloadStore: useWorkloadStore(),

      selectedClusterId: "" as string,
      isClusterSectionValid: false as boolean,

      yamlContent: "" as string,
      isYamlValid: false as boolean,
      isDryRunValid: true as boolean,

      loadedProjects: [] as ProjectMetadata[],
      loadingProject: false as boolean,
      selectedProjectId: null as number | null,
      selectedProject: null as Project | null,
      selectedProjectSupportsMnnvl: false as boolean,

      generalModel: {
        priorityClass: null,
        preemptibility: null,
        autoDeletionTimeAfterCompletionSeconds: null,
        backoffLimit: null,
        terminationGracePeriodSeconds: null,
      } as IUIGeneralSectionModel,

      submitting: false as boolean,
      actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

      // Track values extracted from YAML
      yamlExtractedValues: {
        namespace: null,
        priority: null,
        preemptibility: null,
      } as IYamlExtractedValues,
      // Track if user manually changed values (not via YAML)
      userOverrides: {
        project: false,
        priority: false,
        preemptibility: false,
      } as IOverrideFlags,
      // Track if validation API returned different values than YAML
      apiOverrides: {
        project: false,
        priority: false,
        preemptibility: false,
      } as IOverrideFlags,
    };
  },
  setup() {
    const route = useRoute();
    const { workloadPriorities, loadWorkloadPriorities, supportPriorityClass } = useWorkloadPriorities(
      route.query.clusterId?.toString(),
    );

    return {
      workloadPriorities,
      loadWorkloadPriorities,
      supportPriorityClass,
    };
  },
  async created() {
    // Load saved YAML content if returning from project creation
    const savedState = this.loadFormState();
    const clusterId = this.getDefaultClusterId();

    await this.loadActionPermissions();
    await this.handleClusterDependencies(clusterId);

    if (clusterId) {
      await this.loadWorkloadPriorities(clusterId);
    }

    // If YAML was restored, extract values and load projects
    if (savedState?.yamlContent) {
      this.updateExtractedValues(savedState.yamlContent);
      await this.loadProjects();
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    showFormSections(): boolean {
      // Show sections when YAML has content (even if invalid), hide only when empty
      return !!this.selectedClusterId && this.isClusterSectionValid && !!this.yamlContent;
    },
    workloadFormType(): EWorkloadFormType {
      // Default to Training since YAML can be any type
      return EWorkloadFormType.Training;
    },
    supportMnnvl(): boolean {
      return this.selectedProjectSupportsMnnvl;
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.selectedClusterId,
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },
    workloadPriorityData(): IWorkloadPriorityData {
      return {
        supportPriorityClass: this.supportPriorityClass,
        workloadPriorities: this.workloadPriorities,
        currProject: this.selectedProject || undefined,
        selectedNodePoolsResources: this.selectedProjectNodePools,
        totalGpuRequest: this.extractedGpuRequest,
        supportSemiPreemptibility: this.clusterStore.isClusterVersionSufficient(
          this.selectedClusterId,
          MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
        ),
      };
    },
    // Extract GPU request from YAML
    extractedGpuRequest(): number {
      return yamlUtil.extractGpuRequest(this.yamlContent);
    },
    // Get node pools from the selected project
    selectedProjectNodePools(): string[] {
      if (!this.selectedProject?.resources) return [];
      return this.selectedProject.resources.map((r) => r.nodePool?.name).filter((name): name is string => !!name);
    },
    // Override hints - show when user manually changed a value that exists in YAML
    showProjectOverrideHint(): boolean {
      // Show hint if API returned different project than YAML
      if (this.apiOverrides.project) return true;

      // Or if user manually selected a different project
      if (!this.userOverrides.project || !this.yamlExtractedValues.namespace) return false;
      if (this.selectedProjectId === null || this.selectedProjectId === -1) return false;

      // Check if selected project matches the YAML namespace
      const yamlProject = this.loadedProjects.find((p) => p.namespace === this.yamlExtractedValues.namespace);
      return !yamlProject || +yamlProject.id !== this.selectedProjectId;
    },
    overrideHints(): IOverrideHints {
      // Show hint if current value differs from YAML metadata value
      return {
        priority:
          !!this.yamlExtractedValues.priority &&
          !!this.generalModel.priorityClass &&
          this.generalModel.priorityClass !== this.yamlExtractedValues.priority,
        preemptibility:
          !!this.yamlExtractedValues.preemptibility &&
          !!this.generalModel.preemptibility &&
          this.generalModel.preemptibility !== this.yamlExtractedValues.preemptibility,
      };
    },
    // Selected form values to pass to yaml-workload-section for validation
    selectedFormValues(): ISelectedFormValues {
      // Use generalModel if populated (from dry-run or UI), otherwise use yamlExtractedValues
      // This ensures YAML values are sent for validation, then UI shows validated values
      return {
        projectId: this.selectedProjectId,
        priority: this.generalModel.priorityClass ?? this.yamlExtractedValues.priority ?? null,
        preemptibility: this.generalModel.preemptibility ?? this.yamlExtractedValues.preemptibility ?? null,
      };
    },
    // Disable clusters that don't meet minimum version for YAML submission
    disabledClusterOptions(): Record<string, string> {
      const disabledOptions: Record<string, string> = {};
      for (const cluster of this.clusterStore.clusterList) {
        const isSupported = this.clusterStore.isClusterVersionSufficient(
          cluster.uuid,
          MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION,
        );
        if (!isSupported) {
          disabledOptions[cluster.uuid] =
            `This cluster can't be selected because workloads submitted via YAML are only supported from version ${MIN_CLUSTER_VERSION_FOR_YAML_SUBMISSION}`;
        }
      }
      return disabledOptions;
    },
  },
  methods: {
    getDefaultClusterId(): string {
      let selectedDefaultClusterId = "";
      if (this.clusterStore.clusterList.length === 1) {
        selectedDefaultClusterId = this.clusterStore.clusterList[0]?.uuid;
      } else if (this.$route.query.clusterId) {
        selectedDefaultClusterId = this.$route.query.clusterId.toString();
      }
      return selectedDefaultClusterId;
    },
    async onClusterChanged(clusterId: string, payload: { disabled: boolean }): Promise<void> {
      this.selectedClusterId = clusterId;
      this.isClusterSectionValid = !payload.disabled && this.canSubmitWorkload(clusterId);
      this.selectedProjectId = null;
      this.loadedProjects = [];
      urlService.updateQueryParams({ clusterId });

      // Load workload priorities for the selected cluster
      await this.loadWorkloadPriorities(clusterId);

      if (this.isClusterSectionValid && this.isYamlValid && this.yamlContent) {
        await this.loadProjects();
      }
    },
    onYamlInput(content: string): void {
      this.yamlContent = content;

      // Update extracted values from YAML so comparisons use current content
      this.updateExtractedValues(content);

      if (this.isYamlValid && this.isClusterSectionValid && this.loadedProjects.length === 0) {
        this.loadProjects();
      }
    },
    updateExtractedValues(content: string): void {
      const { namespace, priority, preemptibility } = yamlUtil.extractWorkloadValues(content);
      const typedPreemptibility = preemptibility as Preemptibility | null;
      this.yamlExtractedValues = { namespace, priority, preemptibility: typedPreemptibility };
    },
    onYamlLoaded(content: string): void {
      // Reset all form fields when YAML is loaded/reloaded/cleared
      this.resetFormFields();

      // Reset user overrides when new YAML is pasted/uploaded
      this.userOverrides = { project: false, priority: false, preemptibility: false };

      // Extract values when YAML is loaded (file upload or paste)
      if (content) {
        this.extractValuesFromYaml(content);
      }
    },
    resetFormFields(): void {
      // Reset generalModel fields (priority and preemptibility will come from dry-run)
      this.generalModel = {
        ...this.generalModel,
        priorityClass: null,
        preemptibility: null,
        mnnvl: null,
      };

      // Reset project selection
      this.selectedProjectId = null;
      this.selectedProject = null;
      this.selectedProjectSupportsMnnvl = false;

      // Reset extracted values
      this.yamlExtractedValues = { namespace: null, priority: null, preemptibility: null };

      // Reset api overrides
      this.apiOverrides = { project: false, priority: false, preemptibility: false };
    },
    extractValuesFromYaml(content: string): void {
      // Store extracted values for hint detection only
      // Priority and preemptibility are NOT auto-populated to generalModel
      // They will come from dry-run validation response
      this.updateExtractedValues(content);

      // Auto-select project based on namespace
      this.autoSelectProjectByNamespace(this.yamlExtractedValues.namespace);
    },
    // Find project by namespace
    findProjectByNamespace(namespace: string): ProjectMetadata | undefined {
      return this.loadedProjects.find((p) => p.namespace === namespace);
    },
    autoSelectProjectByNamespace(namespace: string | null): void {
      if (this.userOverrides.project || !namespace || this.loadedProjects.length === 0) return;
      const matchingProject = this.findProjectByNamespace(namespace);
      if (matchingProject) {
        this.selectedProjectId = +matchingProject.id;
      }
    },
    onYamlValidationChanged(isValid: boolean): void {
      this.isYamlValid = isValid;
      if (isValid && this.isClusterSectionValid && this.loadedProjects.length === 0 && this.yamlContent) {
        this.loadProjects();
      }
    },
    async onDryRunSuccess(workload: WorkloadV2): Promise<void> {
      this.isDryRunValid = true;
      this.apiOverrides = { project: false, priority: false, preemptibility: false };
      const metadata = workload.metadata;
      await this.applyProjectFromValidation(metadata?.projectId);
      this.applyPriorityFromValidation(metadata?.priority ?? undefined);
      this.applyPreemptibilityFromValidation(metadata?.preemptibility ?? undefined);
    },
    async applyProjectFromValidation(projectId: string | undefined): Promise<void> {
      if (!projectId) return;

      const parsedId = parseInt(projectId, 10);
      if (isNaN(parsedId)) return;

      // Check if API project differs from YAML namespace
      if (this.yamlExtractedValues.namespace) {
        const yamlProject = this.findProjectByNamespace(this.yamlExtractedValues.namespace);
        // Use string comparison to avoid type mismatch (id might be number or string)
        this.apiOverrides.project = !yamlProject || String(yamlProject.id) !== String(projectId);
      }

      if (parsedId !== this.selectedProjectId) {
        await this.loadProjectById(parsedId.toString());
      }
    },
    applyPriorityFromValidation(priority: string | undefined): void {
      if (!priority) return;

      if (priority !== this.generalModel.priorityClass) {
        this.generalModel.priorityClass = priority;
      }
    },
    applyPreemptibilityFromValidation(preemptibility: string | undefined): void {
      if (!preemptibility || !this.supportPreemptibility) return;

      if (preemptibility !== this.generalModel.preemptibility) {
        this.generalModel.preemptibility = preemptibility as Preemptibility;
      }
    },
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Workloads);
    },
    async loadProjects(): Promise<void> {
      try {
        this.loadingProject = true;
        const permittedScopes = this.actionPermissionsByScopes[Action.Create];

        this.loadedProjects = await orgUnitService.getProjectsMetadata(
          this.selectedClusterId,
          permittedScopes,
          this.authStore.orgUnitList,
        );

        // Auto-select project based on namespace from YAML
        if (this.yamlExtractedValues.namespace && !this.userOverrides.project) {
          const matchingProject = this.findProjectByNamespace(this.yamlExtractedValues.namespace);
          if (matchingProject) {
            await this.onSelectedProject(+matchingProject.id, false);
          }
        }
      } catch (error: unknown) {
        const loadError = error instanceof Error ? error : new Error("Failed to load projects");
        this.$q.notify(alertUtil.getError(loadError.message));
        this.appStore.setFallback(true);
      } finally {
        this.loadingProject = false;
      }
    },
    async onSelectedProject(projectId: number | null, isUserAction = true): Promise<void> {
      this.selectedProjectId = projectId || -1;
      // Mark as user override if this was a manual change and there's a namespace in YAML
      if (isUserAction && this.yamlExtractedValues.namespace) {
        this.userOverrides.project = true;
      }
      // Load full project data for quota warning and GB200 nodepool detection
      if (projectId && projectId !== -1) {
        await this.loadProjectById(projectId.toString());
      } else {
        this.selectedProject = null;
        this.selectedProjectSupportsMnnvl = false;
      }
    },
    async loadProjectById(projectId: string): Promise<void> {
      const requestId = parseInt(projectId, 10);
      this.selectedProjectId = requestId;
      try {
        const project = await orgUnitService.getProject(projectId);
        if (this.selectedProjectId !== requestId) return;
        this.selectedProject = project;
        const supportsMnnvl = await this.checkProjectSupportsMnnvl(project);
        if (this.selectedProjectId !== requestId) return;
        this.selectedProjectSupportsMnnvl = supportsMnnvl;
        if (supportsMnnvl && !this.generalModel.mnnvl) {
          this.generalModel.mnnvl = MnnvlMode.None;
        }
      } catch (error: unknown) {
        if (this.selectedProjectId !== requestId) return;
        const errorAlert = new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("project") });
        this.$q.notify(errorAlert.getNotification(error));
        this.selectedProject = null;
        this.selectedProjectSupportsMnnvl = false;
      }
    },
    async checkProjectSupportsMnnvl(project: Project): Promise<boolean> {
      const projectNodePoolIds = new Set((project.resources ?? []).map((r) => r.nodePool?.id).filter(Boolean));
      if (!projectNodePoolIds.size) return false;

      try {
        const filterBy = [filterUtil.getEqualsFilterString(NodepoolMinimalSortFields.ClusterId, this.selectedClusterId)];
        const nodepools = await clusterService.getNodepoolsMinimal(filterBy);
        return nodepools.some((np) => np.gpuNetworkAccelerationDetected === true && projectNodePoolIds.has(np.id));
      } catch {
        return false;
      }
    },
    onGeneralModelChanged(model: IUIGeneralSectionModel): void {
      this.generalModel = model;
    },
    async handleClusterDependencies(clusterId: string): Promise<void> {
      if (!clusterId) return;
      if (this.canSubmitWorkload(clusterId)) {
        this.onClusterChanged(clusterId, { disabled: false });
      }
    },
    findClusterById(clusterId: string): ClustersMinimalData | null {
      return this.clusterStore.getClusterById(clusterId);
    },
    canSubmitWorkload(clusterId: string): boolean {
      if (!clusterId) return false;
      const currCluster = this.findClusterById(clusterId);
      if (!currCluster) return false;
      return workloadService.canSubmitWorkload(currCluster) === null;
    },
    async onSubmit(): Promise<void> {
      // Validate YAML before submitting - validation is shown in the editor
      if (!this.isYamlValid) {
        return;
      }

      this.submitting = true;

      try {
        // Parse YAML to get the manifest
        const manifest = yamlUtil.yamlToObject(this.yamlContent);
        if (!manifest) {
          throw new Error("Failed to parse YAML");
        }

        // Extract workload name from YAML
        const workloadName = yamlUtil.extractWorkloadName(this.yamlContent);
        if (!workloadName) {
          throw new Error("YAML must contain metadata.name");
        }

        // Call the V2 API - projectId and clusterId are mutually exclusive, use projectId only
        const projectId = this.selectedProjectId ?? 0;
        const request: WorkloadV2CreateRequest = {
          metadata: {
            name: workloadName,
            projectId: projectId.toString(),
            priority: this.generalModel.priorityClass || undefined,
            preemptibility: (this.generalModel.preemptibility as Preemptibility) || undefined,
            ...(this.supportMnnvl &&
              this.generalModel.mnnvl === MnnvlMode.Required && {
                configuration: { mnnvl: MnnvlMode.Required },
              }),
          },
          manifest: manifest as Record<string, unknown>,
        };
        const createdWorkload = await workloadService.createWorkloadFromYaml(request, false, true);

        // Add to workload store so it appears as first row in workload index
        if (createdWorkload) {
          this.workloadStore.setWorkloadAdded(createdWorkload);
        }

        // Clear any saved form state after successful submission
        this.clearFormState();

        this.$q.notify(alertUtil.getSuccess(`Workload ${workloadName} created`));
        this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query: { clusterId: this.selectedClusterId } });
      } catch (error: unknown) {
        const errorMessages = extractErrorMessages(error, "Failed to submit workload");
        errorMessages.forEach((msg: string) => {
          this.$q.notify(alertUtil.getError(msg));
        });
      } finally {
        this.submitting = false;
      }
    },
    redirectToPrevRoute(): void {
      this.clearFormState();
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
    },
    onCancel(): void {
      this.redirectToPrevRoute();
    },
    // Form state preservation - YAML content and general settings
    // Everything else is re-derived: clusterId from URL, project from namespace, etc.
    saveFormState(): void {
      this.workloadStore.setYamlFormState({
        yamlContent: this.yamlContent,
        priorityClass: this.generalModel.priorityClass ?? null,
        preemptibility: this.generalModel.preemptibility ?? null,
        mnnvl: this.generalModel.mnnvl ?? null,
      });
    },
    loadFormState(): ReturnType<typeof this.workloadStore.getYamlFormState> {
      const savedState = this.workloadStore.getYamlFormState();
      if (!savedState) return null;

      this.yamlContent = savedState.yamlContent;
      this.generalModel.priorityClass = savedState.priorityClass ?? null;
      this.generalModel.preemptibility = savedState.preemptibility ?? null;
      this.generalModel.mnnvl = (savedState.mnnvl as MnnvlMode | null) ?? null;

      this.workloadStore.clearYamlFormState();
      return savedState;
    },
    clearFormState(): void {
      this.workloadStore.clearYamlFormState();
    },
  },
});

export default WorkloadYamlNew as typeof WorkloadYamlNew & DefineComponent;
</script>

<style lang="scss" scoped>
.workload-yaml-new {
  width: 100%;

  .general-section:deep(.section-divider) {
    display: none;
  }

  .override-hint {
    font-size: 12px;
  }

  // Warning hint styling for runai-select components
  :deep(.runai-select.has-hint) {
    .q-field__messages {
      color: $warning;
    }
    .q-field__label {
      color: $warning;
    }
    .q-field__control::before {
      border-color: $warning;
    }
  }
}
</style>
