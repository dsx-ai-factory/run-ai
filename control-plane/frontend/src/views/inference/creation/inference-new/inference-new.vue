<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-new column items-center q-pt-md">
    <runai-form-wrapper
      v-if="isPageLoaded"
      :ref="inferenceFormName"
      :form-state="inferenceNewModel"
      :enter-callback="onContinue"
    >
      <section class="create-inference-form">
        <runai-expansion-wrapper>
          <runai-cluster-section
            entity-name="workload"
            :cluster-id="inferenceNewModel.clusterId"
            @update:cluster-id="updateClusterId"
            :disabled-cluster-options="disabledClusterOptions"
          />

          <transition name="fade">
            <project-section
              v-if="inferenceNewModel.clusterId && isClusterSectionValid"
              aid="project-section"
              :loading="loadingProject"
              :selected-project-id="inferenceNewModel.projectId"
              :projects="projects"
              :cluster-id="inferenceNewModel.clusterId"
              @project-changed="onSelectedProject"
            />
          </transition>

          <inference-architecture
            v-if="showInferenceArchitectureSelection"
            :model-value="inferenceNewModel.architecture"
            :is-selectable="true"
            :cluster-architecture-features="clusterArchitectureFeatures"
            @update-model-value="onArchitectureChanged"
          />
          <transition-group name="fade">
            <template v-if="selectedProject">
              <template v-if="isStandardArchitecture">
                <template-section-v2
                  v-if="showTemplates"
                  aid="template-section"
                  :templates="flatTemplates"
                  :total-templates-count="flatTotalCount"
                  :loading="loadingTemplate"
                  :list-query="flatTemplatesListQuery"
                  :template-id="inferenceNewModel.templateId"
                  @template-changed="onSelectedTemplate"
                  @query-changed="onTemplatesListQueryChanged"
                  @get-next-batch="onGetNextTemplatesBatch"
                />

                <inference-server
                  v-if="showInferenceTypeSelection"
                  :is-selectable="true"
                  :model-value="inferenceNewModel.serverType"
                  :cluster-id="inferenceNewModel.clusterId"
                  @update-section-model="inferenceServerSelectionChanged"
                />
              </template>
              <workload-name-section
                entity-type="inference"
                v-model:name="inferenceNewModel.name"
                :disable-closing="true"
                :project-id="inferenceNewModel.projectId"
                :cluster-id="inferenceNewModel.clusterId"
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
            <runai-btn
              label="continue"
              aid="create-inference-continue-btn"
              @click="onContinue"
              :loading="submitting"
              color="primary"
              type="submit"
            ></runai-btn>
          </div>
        </section>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { ProjectSection } from "@/components/section/project-section";
import { InferenceServer } from "@/components/section/inference-server";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { RunaiBtn } from "@/components/common/runai-btn";
import { TemplateSectionV2 } from "@/components/section/template-section-v2";
import { InferenceArchitecture } from "@/components/section/inference-architecture";

// models
import type { IUIInferenceCreation } from "@/models/workload.model";
import {
  EInferenceArchitecture,
  EInferenceType,
  EServerType,
  type IInferenceArchitectureClusterFeatures,
  type TInferenceServerSelection,
} from "@/models/inference.model";
import { SettingKeys } from "@/models/setting.model";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { WorkloadType } from "@/swagger-models/workloads-client";
import type { ICardQuery } from "@/components/common/runai-card-list/runai-card-list.model";
import type { IInferenceNewModel } from "./inference-new.model";

// Stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { urlService } from "@/services/url.service/url.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadCreateService } from "@/services/cluster/workload.service/workload-create.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import {
  MIN_CLUSTER_VERSION_FOR_DISTRIBUTED_INFERENCE,
  MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
  MIN_CLUSTER_VERSION_FOR_NIM,
} from "@/common/version.constant";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { useFlatTemplatesGallery } from "@/composables/use-flat-templates-gallery.composable";
import { ErrorAlert } from "@/utils/error-alert.util";
import { deepCopy } from "@/utils/common.util";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { clusterUtil } from "@/utils/cluster.util/cluster.util";

const PROJECT_AS_UNSELECTED = -1;
const TEMPLATE_AS_START_FROM_SCRATCH = "-1";

export default defineComponent({
  name: "inference-new",
  components: {
    RunaiExpansionWrapper,
    RunaiFormWrapper,
    ProjectSection,
    WorkloadNameSection,
    InferenceServer,
    RunaiClusterSection,
    RunaiBtn,
    TemplateSectionV2,
    InferenceArchitecture,
  },
  data() {
    return {
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      inferenceStore: useInferenceStore(),
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      projects: [] as Array<ProjectMetadata>,
      loadingProject: false as boolean,
      loadingModels: false as boolean,
      submitting: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      inferenceFormName: "inferenceForm" as string,
      isClusterSectionValid: false as boolean,
      disabledClusterOptions: {} as Record<string, string>,
      loadingTemplate: false as boolean,
      inferenceNewModel: {
        name: "",
        serverType: EServerType.Nim, // default value, will be overridden in created()
        templateId: null,
        architecture: EInferenceArchitecture.Standard,
        clusterId: "",
        projectId: PROJECT_AS_UNSELECTED,
      } as IInferenceNewModel,
      lastSelectedServerType: EServerType.Nim,
    };
  },
  setup() {
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
    };
  },
  async created() {
    this.applyInferenceStoreToModel();

    this.disabledClusterOptions = workloadCreateService.getDisabledClusterReasons(this.clusterList, true);

    const selectedDefaultClusterId = this.getDefaultClusterId();
    const createdEntityId: string | undefined = this.$route.query.createdEntityId?.toString();
    // Save projectId BEFORE handleClusterDependencies, because updateClusterId resets the store's projectId to -1
    const savedProjectId = this.inferenceStore.inference.projectId || PROJECT_AS_UNSELECTED;

    await this.inferenceStore.loadActionPermissions(); // This must be before handleClusterDependencies, because it uses the actionPermissionsByScopes
    await this.handleClusterDependencies(selectedDefaultClusterId);
    this.ensureArchitectureMatchesClusterFeatures();
    this.updateInferenceServerType();
    this.lastSelectedServerType = this.inferenceNewModel.serverType;

    const selectedDefaultProjectId = this.getDefaultProjectId(savedProjectId, createdEntityId);
    if (selectedDefaultProjectId) {
      await this.onSelectedProject(selectedDefaultProjectId);
    } else if (this.selectedProject && this.showTemplates && this.isStandardArchitecture) {
      // Project was already selected (e.g. single project in loadProjects) before showTemplates was set,
      // so onSelectedProject returned early and did not load templates. Load them now.
      await this.loadTemplates();
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    isStandardArchitecture(): boolean {
      return this.inferenceNewModel.architecture === EInferenceArchitecture.Standard;
    },
    isNimInference(): boolean {
      return this.inferenceNewModel.serverType === EServerType.Nim;
    },
    isCustomInference(): boolean {
      return this.inferenceNewModel.serverType === EServerType.Custom;
    },
    isPageLoaded(): boolean {
      return !this.appStore.pageLoading;
    },
    supportingServingPortAccess(): boolean {
      return this.inferenceStore.supportingServingPortAccess(this.inferenceNewModel.clusterId);
    },
    selectedProject(): boolean {
      return !!(this.inferenceNewModel.projectId && this.inferenceNewModel.projectId !== PROJECT_AS_UNSELECTED);
    },
    showTemplates(): boolean {
      return this.settingStore.isFeatureEnabled(SettingKeys.EnableTemplatesV2) && this.isStandardArchitecture;
    },
    clusterList(): Array<ClustersMinimalData> {
      return this.clusterStore.clusterList;
    },
    isHybridSupported(): boolean {
      return (
        this.clusterStore.isClusterVersionSufficient(
          this.inferenceNewModel.clusterId,
          MIN_CLUSTER_VERSION_FOR_HYBRID_SUBMISSION,
        ) && this.settingStore.isHybridSubmissionEnabled
      );
    },
    showInferenceTypeSelection(): boolean {
      if (
        !this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) &&
        !this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM)
      ) {
        return false;
      }

      if (this.showTemplates) {
        return this.inferenceNewModel.templateId === TEMPLATE_AS_START_FROM_SCRATCH;
      }

      return true;
    },
    isTemplateSelected(): boolean {
      return !!(
        this.inferenceNewModel.templateId && this.inferenceNewModel.templateId !== TEMPLATE_AS_START_FROM_SCRATCH
      );
    },
    showInferenceArchitectureSelection(): boolean {
      return (
        this.selectedProject &&
        this.clusterStore.isClusterVersionSufficient(
          this.inferenceNewModel.clusterId,
          MIN_CLUSTER_VERSION_FOR_DISTRIBUTED_INFERENCE,
        )
      );
    },
    clusterArchitectureFeatures(): IInferenceArchitectureClusterFeatures {
      const cluster = this.findClusterById(this.inferenceNewModel.clusterId);
      if (!cluster) {
        return { isKnativeInstalled: false, isLwsInstalled: false };
      }
      return {
        isKnativeInstalled: clusterUtil.isKnativeInstalledInCluster(cluster),
        isLwsInstalled: clusterUtil.isLwsInstalledInCluster(cluster),
      };
    },
  },
  methods: {
    getDefaultClusterId(): string {
      let selectedDefaultClusterId = "";
      if (this.clusterStore.isSingleCluster) {
        selectedDefaultClusterId = this.clusterStore.defaultClusterId;
      } else if (this.$route.query.clusterId) {
        selectedDefaultClusterId = this.$route.query.clusterId.toString();
      } else if (this.clusterStore.selectedClusterId) {
        selectedDefaultClusterId = this.clusterStore.selectedClusterId;
      }

      return selectedDefaultClusterId;
    },
    getDefaultProjectId(savedProjectId: number, createdEntityId?: string): number | null {
      let selectedDefaultProjectId = null;
      if (createdEntityId && this.projects.length > 1) {
        selectedDefaultProjectId = Number(createdEntityId);
      } else if (savedProjectId !== -1) {
        selectedDefaultProjectId = savedProjectId;
      }

      return selectedDefaultProjectId;
    },
    applyInferenceStoreToModel(): void {
      if (!this.inferenceStore.inference) return;
      const inference = deepCopy(this.inferenceStore.inference);
      this.inferenceNewModel.name = inference.name;
      this.inferenceNewModel.serverType = inference.serverType;
      this.inferenceNewModel.templateId = inference.templateId || null;
      this.inferenceNewModel.architecture = inference.architecture;
      this.inferenceNewModel.clusterId = inference.clusterId;
      this.inferenceNewModel.projectId = inference.projectId;
    },
    async loadProjects(): Promise<void> {
      try {
        this.loadingProject = true;
        this.projects = await orgUnitService.getProjectsMetadata(
          this.inferenceNewModel.clusterId,
          this.inferenceStore.actionPermissionsByScopes.create,
          this.authStore.orgUnitList,
        );
        if (this.projects.length === 1) {
          await this.onSelectedProject(+this.projects[0].id);
        } else if (this.inferenceNewModel.projectId !== PROJECT_AS_UNSELECTED) {
          await this.onSelectedProject(this.inferenceNewModel.projectId);
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
        await this.list(
          WorkloadType.Inference,
          this.inferenceNewModel.projectId.toString(),
          applyOffset,
          amountRequired,
        );

        if (this.flatTemplates.length === 0 && !this.inferenceNewModel.templateId) {
          this.inferenceNewModel.templateId = TEMPLATE_AS_START_FROM_SCRATCH;
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to load templates",
        });
        this.$q.notify(errorAlert.getNotification(error));
        this.setTemplates([]);
      } finally {
        this.loadingTemplate = false;
      }
    },
    async onSelectedProject(projectId: number | null): Promise<void> {
      this.inferenceNewModel.projectId = projectId || PROJECT_AS_UNSELECTED;
      if (this.inferenceNewModel.projectId === PROJECT_AS_UNSELECTED || !this.showTemplates) return;
      await this.loadTemplates();
    },
    onSelectedTemplate(templateId: string | null): void {
      this.inferenceNewModel.templateId = templateId || "";
      if (this.inferenceNewModel.templateId && this.inferenceNewModel.templateId !== TEMPLATE_AS_START_FROM_SCRATCH) {
        this.inferenceNewModel.serverType = EServerType.Custom;
      } else if (this.inferenceNewModel.templateId === TEMPLATE_AS_START_FROM_SCRATCH) {
        this.inferenceNewModel.serverType = this.lastSelectedServerType;
      }
    },
    onTemplatesListQueryChanged(query: ICardQuery): void {
      this.setFlatTemplatesListQuery(query);
      this.loadTemplates();
    },
    onGetNextTemplatesBatch(amountRequired: number): void {
      this.loadTemplates(true, amountRequired);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
    validate(): Promise<boolean> {
      const inferenceForm = this.$refs[this.inferenceFormName] as HTMLFormElement;
      return (inferenceForm as HTMLFormElement).validate();
    },
    async onContinue(): Promise<void> {
      this.displayFormHint = false;
      this.submitting = true;
      const success = await this.validate();
      if (!success) {
        this.submitting = false;
        this.showHint();
        return;
      }

      if (this.inferenceNewModel.projectId === PROJECT_AS_UNSELECTED) {
        console.error("ProjectId is missing");
        this.$q.notify(alertUtil.getError("Failed to create inference"));
        this.submitting = false;
        return;
      }

      this.saveStateToInferenceNewModel();
      if (this.isStandardArchitecture) {
        await this.continueToStandard();
      } else {
        await this.continueToDistributed();
      }
    },
    async continueToDistributed(): Promise<void> {
      const query: Record<string, string> = {
        clusterId: this.inferenceNewModel.clusterId,
        projectId: this.inferenceNewModel.projectId.toString(),
        inferenceName: this.inferenceNewModel.name,
      };
      await this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT, query });
    },
    async continueToStandard(): Promise<void> {
      if (this.isNimInference) {
        await this.continueToNim();
      } else if (this.inferenceNewModel.serverType === EServerType.VLLM) {
        await this.continueToVLLM();
      } else if (this.inferenceNewModel.serverType === EServerType.TGI) {
        await this.continueToTgi();
      } else {
        await this.continueToCustom();
      }
    },
    getQuery(): Record<string, string> {
      return {
        clusterId: this.inferenceNewModel.clusterId,
        projectId: this.inferenceNewModel.projectId.toString(),
        inferenceName: this.inferenceNewModel.name,
      };
    },
    resetInferenceStore(inferenceFormUIState?: IInferenceFormUIState): void {
      this.inferenceStore.clearPresetFormModel();
      if (inferenceFormUIState) {
        this.inferenceStore.setInferenceFormUIState(inferenceFormUIState);
      } else {
        this.inferenceStore.clearInferenceFormUIState();
      }
      this.inferenceStore.clearFlatFormState();
    },
    async continueToNim(): Promise<void> {
      const query = this.getQuery();
      this.resetInferenceStore({
        modelBasedOn: EServerType.Nim,
      });
      await this.$router.push({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_NIM_EDIT,
        query,
      });
    },
    async continueToVLLM(): Promise<void> {
      const query = this.getQuery();
      this.resetInferenceStore({
        modelBasedOn: EServerType.VLLM,
      });
      await this.$router.push({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_VLLM_EDIT,
        query,
      });
    },

    async continueToTgi(): Promise<void> {
      const query = this.getQuery();
      this.resetInferenceStore({
        modelBasedOn: EServerType.TGI,
      });
      await this.$router.push({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_TGI_EDIT,
        query,
      });
    },

    async continueToAssetBased(): Promise<void> {
      const mergedInference = await inferenceService.getInferenceMergedWithPolicy(
        String(this.inferenceNewModel.projectId),
      );
      const policy = await inferenceService.getPolicy(this.inferenceNewModel.projectId);
      this.inferenceStore.setLoadedInferenceModelWithDefaults(mergedInference);

      const inference = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
        {
          ...mergedInference,
          meta: {
            ...mergedInference.meta,
            name: this.inferenceNewModel.name,
            clusterId: this.inferenceNewModel.clusterId,
            projectId: this.inferenceNewModel.projectId?.toString() || "",
          },
        },
        EInferenceType.AssetsBased,
        this.supportingServingPortAccess,
        undefined,
        "",
        policy?.effective || undefined,
        this.inferenceNewModel.templateId || undefined,
        this.inferenceNewModel.architecture,
      );

      this.inferenceStore.setInference(inference);
      await this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT });
    },
    async continueToCustom(): Promise<void> {
      if (!this.isHybridSupported) {
        await this.continueToAssetBased();
        return;
      }

      this.resetInferenceStore();
      const query: Record<string, string> = {
        clusterId: this.inferenceNewModel.clusterId,
        projectId: this.inferenceNewModel.projectId.toString(),
        inferenceName: this.inferenceNewModel.name,
      };

      if (this.inferenceNewModel.templateId && this.inferenceNewModel.templateId !== TEMPLATE_AS_START_FROM_SCRATCH) {
        query.templateId = this.inferenceNewModel.templateId;
      }

      await this.$router.push({
        name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID,
        query,
      });
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.$router.back();
      }
    },
    inferenceServerSelectionChanged(selection: TInferenceServerSelection): void {
      this.inferenceNewModel.serverType = selection;
      this.lastSelectedServerType = selection;
    },
    async updateClusterId(clusterId: string, payload: { disabled: boolean }): Promise<void> {
      this.displayFormHint = false;
      this.inferenceNewModel.clusterId = clusterId;
      this.isClusterSectionValid = !payload.disabled && this.canSubmitWorkload(clusterId);
      this.inferenceNewModel.projectId = PROJECT_AS_UNSELECTED;

      urlService.updateQueryParams({ clusterId, createdEntityId: null });
      if (this.isClusterSectionValid) {
        await this.loadProjects();
      }
      this.ensureArchitectureMatchesClusterFeatures();
    },
    findClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    async handleClusterDependencies(clusterId: string): Promise<void> {
      if (!clusterId) return;
      if (this.canSubmitWorkload(clusterId)) {
        await this.updateClusterId(clusterId, { disabled: false });
      }
    },
    canSubmitWorkload(clusterId: string): boolean {
      if (!clusterId) return false;
      const currCluster = this.findClusterById(clusterId);
      if (!currCluster) return false;
      return workloadService.canSubmitWorkload(currCluster) === null;
    },
    onArchitectureChanged(architecture: EInferenceArchitecture): void {
      this.inferenceNewModel.architecture = architecture;
      if (architecture === EInferenceArchitecture.Standard) {
        this.updateInferenceServerType();
        this.inferenceNewModel.templateId = TEMPLATE_AS_START_FROM_SCRATCH;
      } else {
        this.inferenceNewModel.serverType = EServerType.Custom;
        this.inferenceNewModel.templateId = null;
      }
    },
    ensureArchitectureMatchesClusterFeatures(): void {
      if (!this.showInferenceArchitectureSelection) return;
      const cluster = this.findClusterById(this.inferenceNewModel.clusterId);
      if (!cluster) return;

      const knativeOk = clusterUtil.isKnativeInstalledInCluster(cluster);
      const lwsOk = clusterUtil.isLwsInstalledInCluster(cluster);

      if (!knativeOk && !lwsOk) return;

      if (!knativeOk && this.inferenceNewModel.architecture === EInferenceArchitecture.Standard) {
        this.onArchitectureChanged(EInferenceArchitecture.Distributed);
      } else if (!lwsOk && this.inferenceNewModel.architecture === EInferenceArchitecture.Distributed) {
        this.onArchitectureChanged(EInferenceArchitecture.Standard);
      }
    },
    updateInferenceServerType(): void {
      const isNimSupported =
        this.clusterStore.isClusterVersionSufficient(this.inferenceNewModel.clusterId, MIN_CLUSTER_VERSION_FOR_NIM) &&
        this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
      const isHuggingFaceSupported = this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog);

      if (!isNimSupported && this.inferenceNewModel.serverType === EServerType.Nim) {
        if (isHuggingFaceSupported) {
          this.inferenceNewModel.serverType = EServerType.VLLM;
        } else {
          this.inferenceNewModel.serverType = EServerType.Custom;
        }
      }
    },
    saveStateToInferenceNewModel(): void {
      const inference: IUIInferenceCreation = {
        ...this.inferenceNewModel,
        namespace: "",
        assets: {
          environment: null,
          compute: null,
        },
        specificEnv: {},
      };
      this.inferenceStore.setInference(inference);
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>
