<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="training-new column items-center q-pt-md">
    <runai-form-wrapper
      :form-state="{ selectedTemplateId, trainingName }"
      @leave-page="() => onCancel(false)"
      :enter-callback="onContinue"
    >
      <q-form ref="trainingForm" @submit.prevent>
        <section class="create-training-form">
          <runai-expansion-wrapper>
            <runai-cluster-section
              entity-name="workload"
              :cluster-id="selectedClusterId || ''"
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
                <template v-if="showTemplatesV2">
                  <runai-expansion-item
                    disable-closing
                    class="multi-node-section"
                    aid="multi-node-section"
                    label="Workload architecture"
                  >
                    <workload-architecture-select
                      :is-distributed="!!selectedDistributedDetails"
                      :form-type="EDistributedFormType.TRAINING"
                      @update="setDistributedDetails"
                    />
                  </runai-expansion-item>
                  <template-section-v2
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

                  <runai-expansion-item
                    v-if="selectedDistributedDetails"
                    disable-closing
                    class="framework-and-configuration-section"
                    aid="framework-and-configuration-section"
                    label="Framework"
                  >
                    <distributed-framework-selection-section
                      :disable-selection="disableFrameworkSection"
                      disabled-tooltip="This is automatically set based on the selected template. To modify it select a different template or start from scratch"
                      :dist-framework="selectedDistributedDetails.distFramework || null"
                      :cluster-id="selectedClusterId"
                      @update="setDistributedFramework"
                    />

                    <distributed-replica-configuration
                      v-if="showMasterConfigurationSection"
                      :model-value="selectedDistributedDetails.noMaster || false"
                      @update:model-value="selectedDistributedDetails.noMaster = $event"
                    />
                  </runai-expansion-item>
                </template>
                <template v-else>
                  <multi-node-section
                    v-if="isClusterVersionSupportDistributedTraining"
                    :is-selectable="true"
                    :multi-node-section-model="selectedDistributedDetails"
                    @update-multi-node-section-model="updateMultiNodeSection"
                    :cluster-id="selectedClusterId"
                  />
                  <template-section
                    aid="template-section"
                    :templates="templates"
                    :loading="loadingTemplate"
                    @template-changed="onSelectedTemplate"
                    :template-id="selectedTemplateId"
                  />
                </template>
                <workload-name-section
                  entity-type="training"
                  aid="training-name-section"
                  v-model:name="trainingName"
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
              <runai-btn
                flat
                class="q-mr-sm"
                color="primary"
                label="Cancel"
                aid="cancel-btn"
                @click="() => onCancel(true)"
              />
              <quick-submission-dropdown
                v-if="showQuickSubmissionDropdown && selectedTemplateId"
                :workload-type="selectedDistributedDetails ? WorkloadType.Distributed : WorkloadType.Training"
                :is-template-runnable="isTemplateRunnable"
                :is-loading="isQuickSubmissionLoading"
                @submit="onQuickSubmission"
                @continue="onContinue"
                type="submit"
              />
              <runai-btn
                v-else
                label="continue"
                aid="create-training-continue-btn"
                @click="onContinue"
                :loading="submitting"
                color="primary"
                type="submit"
              ></runai-btn>
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
import { useTrainingStore } from "@/stores/training.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useAppStore } from "@/stores/app.store";
import { useSettingStore } from "@/stores/setting.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useAuthStore } from "@/stores/auth.store";

// services
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { urlService } from "@/services/url.service/url.service";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { ProjectSection } from "@/components/section/project-section";
import { TemplateSection } from "@/components/section/template-section";
import { MultiNodeSection } from "@/components/section/multi-node-section";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { RunaiBtn } from "@/components/common/runai-btn";
import { TemplateSectionV2 } from "@/components/section/template-section-v2";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { QuickSubmissionDropdown } from "@/components/workload-creation/quick-submission-dropdown";
// models
import {
  type WorkloadTemplate,
  type PVCAsset,
  UidGidSource,
  DistributedFramework,
} from "@/swagger-models/assets-service-client";
import type { IUIDistributed, IUIWorkloadCreation } from "@/models/workload.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { Project, ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import type { ICardQuery } from "@/components/common/runai-card-list/runai-card-list.model";
// constant
import { errorMessages } from "@/common/error-message.constant";
// common
import { required } from "@/common/form.validators";

// route
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";

// utils
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { deepCopy } from "@/utils/common.util";
import {
  MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
  MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
} from "@/common/version.constant";
import type { LocationQueryRaw } from "vue-router";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { SettingKeys } from "@/models/setting.model";
import type { DistributedTemplate } from "@/swagger-models/workloads-client";
import { WorkloadType } from "@/swagger-models/workloads-client";
import { EDistributedFormType } from "@/components/section/multi-node-section/distributed-selection/distributed-selection.vue";
import { WorkloadArchitectureSelect } from "@/components/section/multi-node-section/distributed-selection/workload-architecture-select";
import { DistributedReplicaConfiguration } from "@/components/section/multi-node-section/distributed-replica-configuration";
import { DistributedFrameworkSelectionSection } from "@/components/section/multi-node-section/distributed-selection/distributed-framework-selection-section";
import { useFlatTemplatesGallery } from "@/composables/use-flat-templates-gallery.composable";
import { useQuickSubmission } from "@/composables/use-quick-submission.composable";

const TrainingNew = defineComponent({
  name: "training-new",
  components: {
    RunaiClusterSection,
    RunaiExpansionWrapper,
    RunaiFormWrapper,
    ProjectSection,
    WorkloadNameSection,
    TemplateSection,
    MultiNodeSection,
    RunaiBtn,
    TemplateSectionV2,
    RunaiExpansionItem,
    WorkloadArchitectureSelect,
    DistributedReplicaConfiguration,
    DistributedFrameworkSelectionSection,
    QuickSubmissionDropdown,
  },
  data() {
    return {
      trainingStore: useTrainingStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      environmentStore: useEnvironmentStore(),
      authStore: useAuthStore(),
      loadingProject: false as boolean,
      loadingTemplate: false as boolean,
      trainingForm: null as HTMLFormElement | null,
      loadedProjects: [] as Array<ProjectMetadata>,
      training: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,

      submitting: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      isClusterSectionValid: false as boolean,

      selectedProjectId: null as number | null,
      selectedTemplateId: null as string | null,
      selectedClusterId: "" as string,
      selectedDistributedDetails: null as IUIDistributed | null,
      trainingName: "" as string,

      EDistributedFormType,
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
    await this.trainingStore.loadActionPermissions(); // This must be before handleClusterDependencies, because it uses the actionPermissionsByScopes
    await this.handleClusterDependencies(selectedDefaultClusterId);
    this.appStore.setPageLoading(false);
  },
  async mounted() {
    this.trainingForm = this.$refs["trainingForm"] as HTMLFormElement;
    const createdEntityId: string | undefined = this.$route.query.createdEntityId?.toString();
    const savedProjectId = this.trainingStore.training.projectId || -1;

    const selectedDefaultProjectId = this.getDefaultProjectId(savedProjectId, createdEntityId);
    if (selectedDefaultProjectId) {
      await this.onSelectedProject(selectedDefaultProjectId);
    }
  },
  computed: {
    templates(): Array<WorkloadTemplate> {
      return this.workloadTemplateStore.workloadTemplateList;
    },
    isClusterVersionSupportDistributedTraining(): boolean {
      return this.clusterStore.isVersionSupportDistributedTraining(this.selectedClusterId);
    },
    templatesFilter(): IAssetsFilter {
      const templatesFilter: IAssetsFilter = {
        projectId: this.selectedProjectId,
        isTraining: true,
      };
      if (this.selectedDistributedDetails) {
        templatesFilter.isDistributed = true;
        if (this.selectedDistributedDetails?.distFramework)
          templatesFilter.distributedFramework = this.selectedDistributedDetails.distFramework;
      }
      return templatesFilter;
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
    showMasterConfigurationSection(): boolean {
      return (
        !!this.selectedDistributedDetails &&
        (this.selectedDistributedDetails.distFramework === DistributedFramework.PyTorch ||
          this.selectedDistributedDetails.distFramework === DistributedFramework.Tf)
      );
    },
    disableFrameworkSection(): boolean {
      // if a template with distributed framework is selected, the framework section should be disabled
      const template = this.flatTemplates.find((template) => template.meta.id === this.selectedTemplateId);
      return !!(template as DistributedTemplate)?.spec?.distributedFramework;
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
        selectedDefaultClusterId = this.clusterStore.clusterList[0].uuid;
      } else if (this.$route.query.clusterId) {
        selectedDefaultClusterId = this.$route.query.clusterId.toString();
      } else if (this.trainingStore.training.clusterId) {
        selectedDefaultClusterId = this.trainingStore.training.clusterId;
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
          this.trainingStore.actionPermissionsByScopes.create,
          this.authStore.orgUnitList,
        );

        if (this.loadedProjects.length === 1) {
          await this.onSelectedProject(+this.loadedProjects[0].id);
        } else if (this.trainingStore.training.projectId !== -1) {
          await this.onSelectedProject(this.trainingStore.training.projectId);
        }
      } catch (error: unknown) {
        this.$q.notify(alertUtil.getError("Failed to load projects"));
        console.error(error);
        this.appStore.setFallback(true);
      } finally {
        this.loadingProject = false;
      }
    },
    async loadTemplates(applyOffset = false, amountRequired?: number): Promise<void> {
      try {
        this.loadingTemplate = true;

        if (this.showTemplatesV2) {
          if (!this.selectedProjectId) return;

          // no reason to call count if all we need is the next batch
          const workloadType = this.selectedDistributedDetails ? WorkloadType.Distributed : WorkloadType.Training;
          await this.list(workloadType, this.selectedProjectId.toString(), applyOffset, amountRequired);

          if (!applyOffset) this.selectedTemplateId = null;
        } else {
          await this.workloadTemplateStore.loadWorkloadTemplates(this.templatesFilter);
        }

        const templates = this.showTemplatesV2 ? this.flatTemplates : this.templates;

        if (templates.length === 0 && !this.selectedTemplateId) {
          this.selectedTemplateId = "-1";
        } else if (
          !this.showTemplatesV2 &&
          this.trainingStore.templateId &&
          (this.trainingStore.templateId === "-1" ||
            templates.find((template) => template.meta.id === this.trainingStore.templateId))
        ) {
          this.selectedTemplateId = this.trainingStore.templateId;
        }
      } catch (error: unknown) {
        // could happen if the user try to search for templates with non-english characters
        this.setTemplates([]);
        console.error(error);
      } finally {
        this.loadingTemplate = false;
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
    async loadTrainingModel(projectId: string, templateId?: string | null, isDistributed?: boolean): Promise<void> {
      const tempId = templateId === "-1" ? undefined : templateId;
      const trainingModel = await trainingService.getTrainingMergedWithPolicy(
        projectId,
        undefined,
        tempId || undefined,
        isDistributed,
      );
      this.trainingStore.setLoadedTrainingModelWithDefaults(deepCopy(trainingModel));
      this.training = workloadUtil.convertWorkloadToWorkloadUI({
        ...trainingModel,
        meta: {
          ...trainingModel.meta,
          name: this.trainingName,
          clusterId: this.selectedClusterId,
          projectId: this.selectedProjectId?.toString() || "",
        },
      });

      this.setFormDefaultsIfNeeded();
      this.trainingStore.setTraining(this.training);
    },
    onSelectedTemplate(templateId: string | null): void {
      this.selectedTemplateId = templateId;
      if (this.showTemplatesV2 && !!this.selectedDistributedDetails) {
        const template = this.flatTemplates.find((template) => template.meta.id === templateId);
        const distFramework = (template as DistributedTemplate)?.spec?.distributedFramework || null;
        this.setDistributedFramework(distFramework);
      }
      this.trainingStore.setTemplate(templateId);
    },
    updateMultiNodeSection(distributed: IUIDistributed): void {
      this.selectedDistributedDetails = distributed || null;
      this.selectedTemplateId = null;
      this.loadTemplates();
    },
    redirectToPrevRoute(): void {
      this.trainingStore.clearDistributedMasterFlatFormState();
      this.trainingStore.clearDistributedWorkerFlatFormState();
      this.trainingStore.clearFlatFormState();
      this.trainingStore.clearDistFlatFromContext();

      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
    },
    async onCancel(verify = false): Promise<void> {
      if (verify) {
        const allowToLeave: boolean = await requestToLeave();
        if (allowToLeave) {
          this.redirectToPrevRoute();
        }
      } else this.redirectToPrevRoute();
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
        this.$q.notify(alertUtil.getError("Failed to create training"));
        this.submitting = false;
        return;
      }

      await this.loadTrainingModel(
        this.selectedProjectId.toString(),
        this.selectedTemplateId,
        !!this.selectedDistributedDetails,
      );

      await this.setSpecificUserGidUidIfNeeded();

      const template: WorkloadTemplate | undefined = this.templates.find(
        (template) => template.meta.id === this.selectedTemplateId,
      );
      if (template?.spec.assets.workloadVolumes?.length) {
        const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(template.spec.assets.workloadVolumes);
        this.training.assets.uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
      }

      this.trainingStore.setTraining(this.training);

      const routeName = this.training.distributed
        ? TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT
        : TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT;

      await this.$router.push({ name: routeName });
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
        this.$q.notify(alertUtil.getError("Failed to create training"));
        this.submitting = false;
        return;
      }

      let routeName = TRAINING_ROUTE_NAMES.TRAINING_HYBRID;
      const routeQuery: LocationQueryRaw = {
        clusterId: this.selectedClusterId,
        projectId: this.selectedProjectId,
        templateId: this.selectedTemplateId !== "-1" ? this.selectedTemplateId : undefined,
        trainingName: this.trainingName,
      };

      if (this.selectedDistributedDetails) {
        routeName = TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID;
        routeQuery["distFramework"] = this.selectedDistributedDetails.distFramework;
        if (this.selectedDistributedDetails.noMaster)
          routeQuery["noMaster"] = this.selectedDistributedDetails.noMaster?.toString();
      }

      await this.$router.push({
        name: routeName,
        query: routeQuery,
      });
    },
    validate(): Promise<boolean> {
      return (this.trainingForm as HTMLFormElement).validate();
    },
    isRequiredTemplate(val: string): boolean | string {
      return required(val) || errorMessages.SELECT_TEMPLATE;
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
      if (this.training.assets.environment) {
        const relevantEnvironment = await this.environmentStore.loadById(this.training.assets.environment);
        if (relevantEnvironment?.spec.uidGidSource === UidGidSource.FromIdpToken) {
          this.training.specificEnv.runAsUid = this.authStore.getUID || null;
          this.training.specificEnv.runAsGid = this.authStore.getGID || null;
          this.training.specificEnv.supplementalGroups = this.authStore.getSupplementaryGroups;
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
    // Setting form defaults for specific fields as defined by product
    setFormDefaultsIfNeeded(): void {
      const supportRunsAndParallelism = this.clusterStore.isClusterVersionSufficient(
        this.training.clusterId,
        MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
      );

      if (supportRunsAndParallelism && !this.training.specificEnv?.completions && !this.training.distributed) {
        this.training.specificEnv.completions = 1;
        this.training.specificEnv.parallelism = null;
      }

      if (this.training.distributed) {
        this.training.distributed.distFramework = this.selectedDistributedDetails?.distFramework;
        this.training.distributed.numWorkers ||= 1;
        this.training.distributed.noMaster = this.selectedDistributedDetails?.noMaster;
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
    setDistributedDetails(isDistributed: boolean): void {
      this.selectedDistributedDetails = isDistributed
        ? {
            distFramework: null,
            numWorkers: 1,
          }
        : null;

      this.selectedTemplateId = null;
      this.loadTemplates();
    },
    setDistributedFramework(distFramework: DistributedFramework | null): void {
      if (!this.selectedDistributedDetails) return;

      const newDistributedModel: IUIDistributed = {
        distFramework: distFramework,
        numWorkers: 1,
      };
      if (distFramework === DistributedFramework.Jax) newDistributedModel.noMaster = true;

      this.selectedDistributedDetails = newDistributedModel;
    },
    async onQuickSubmission(): Promise<void> {
      this.displayFormHint = false;

      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }

      const type = this.selectedDistributedDetails ? WorkloadType.Distributed : WorkloadType.Training;
      const selectedTemplate = this.flatTemplates.find((template) => template.meta.id === this.selectedTemplateId);

      await this.onQuickSubmit(type, {
        name: this.trainingName,
        projectId: this.selectedProjectId?.toString() || "",
        templateId: this.selectedTemplateId || "",
        clusterId: this.selectedClusterId,
        // unfortunately, the backend requires us the send the distributed framework in the spec even if its mentioned in the template
        spec: this.selectedDistributedDetails?.distFramework
          ? {
              distributedFramework: this.selectedDistributedDetails?.distFramework || undefined,
            }
          : undefined,
        masterSpecSameAsWorker: (selectedTemplate as DistributedTemplate)?.masterSpecSameAsWorker,
        masterSpec: (selectedTemplate as DistributedTemplate)?.master?.spec,
      });
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});

export default TrainingNew as typeof TrainingNew & DefineComponent;
</script>
