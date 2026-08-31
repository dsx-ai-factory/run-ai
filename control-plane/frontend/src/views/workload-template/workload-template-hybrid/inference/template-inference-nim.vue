<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <nim-inference-submit-form
    v-if="isPageReady"
    class="inference-template-hybrid template-hybrid-nim"
    :navigation="navigationData"
    :section-templates="sectionTemplatesData"
    :cluster-config="clusterConfigData"
    :data-sources="dataSourcesData"
    :features="featuresData"
    :ui-state="uiStateData"
    :model-value="inference"
    @update:model-value="handleModelValueUpdate"
    @update:ui-state="onAccessUIStateChanged"
    @display-computes-list="displayComputesList"
    @reset-compute="resetComputeToDefault"
    @clear-compute-template="clearComputeTemplate"
    @display-model-storage-list="displayModelStorageList"
    @cancel="onCancel"
    @submit="onSubmit"
    @back="onBack"
    @leave-page="onLeavePage"
    @drawer-action="handleDrawerAction"
    @navigation-action="handleNavigationAction"
    @toggle-linked="toggleComputeLinked"
    @scope-changed="handleScopeChange"
  />
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";

// components
import { NimInferenceSubmitForm } from "@/components/workload-creation/inference/nim-inference-submit-form";

// models
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import {
  EWorkloadFormType,
  EWorkloadTypeSupportedInAssetsService,
  type IUIStorageFields,
  type IUISuperWorkloadCreationRequest,
} from "@/models/workload.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { AssetKind, ScopeType } from "@/swagger-models/assets-service-client";

// stores
import { useAppStore } from "@/stores/app.store";
import { type IInferenceFormUIState, useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkloadStore } from "@/stores/workload.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util/common.util";
import { policyUtil } from "@/utils/policy.util/policy.util";

// routes
import { INFERENCE_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/inference-routes/inference-template.routes.names";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { ErrorAlert } from "@/utils/error-alert.util";
import { SUBMIT_OPTIONS } from "@/components/workload-creation/inference/nim-inference-submit-form/nim-inference-submit-form.types";

// types
import type {
  INimInferenceSubmitFormDrawerAction,
  INimInferenceSubmitFormNavigation,
  INimInferenceSubmitFormSectionTemplates,
  INimInferenceSubmitFormClusterConfig,
  INimInferenceSubmitFormDataSources,
  INimInferenceSubmitFormFeatures,
  INimInferenceSubmitFormUIState,
} from "@/components/workload-creation/inference/nim-inference-submit-form/nim-inference-submit-form.types";
import { ENimInferenceSubmitFormDrawerActionType } from "@/components/workload-creation/inference/nim-inference-submit-form/nim-inference-submit-form.types";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { IScopeModel } from "@/models/global.model";
import { scopeUtil } from "@/utils/scope.util";
import { EServerType } from "@/models/inference.model";

const TemplateInferenceNim = defineComponent({
  name: "template-inference-nim",
  components: {
    NimInferenceSubmitForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      workloadStore: useWorkloadStore(),
      submitting: false as boolean,
      clusterName: "" as string,
      showInferenceType: false as boolean,
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      autoScaleOptions: {
        latencyEnabled: true,
        isAutoScaleEnabled: true,
      } as IAutoScaleOptions,
      secrets: [] as SecretListInfo[],
      uiState: {} as IInferenceFormUIState,
      originalName: "" as string,
      serverType: EServerType.Nim as EServerType,
      scopeChangeInProgress: false as boolean,
    };
  },
  setup() {
    const {
      // state
      isDrawerOpen,
      hybridWorkloadModel,
      selectedTemplates,
      drawerDataEntity,
      recentlyCreatedEntityId,
      // methods
      closeDrawer,
      getUnifiedHybridWorkload,
      displayComputesList,
      resetComputeToDefault,
      clearComputeTemplate,
      handleSelection,
      handleMultiSelection,
      onCreateNewSetup,
      displayDataSourcesList,
      displayModelStorageList,
      handleCreatedAssets,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      linkedEntitiesIds,
      // computed
      selectedComputeName,
      baseDrawerFilters,
    } = useHybridTemplates(EHybridFormTypes.Inference, undefined, true);
    return {
      // state
      isDrawerOpen,
      inference: hybridWorkloadModel,
      selectedTemplates,
      drawerDataEntity,
      recentlyCreatedEntityId,
      // methods
      closeDrawer,
      getUnifiedHybridWorkload,
      displayComputesList,
      resetComputeToDefault,
      clearComputeTemplate,
      handleSelection,
      handleMultiSelection,
      onCreateNewSetup,
      displayDataSourcesList,
      displayModelStorageList,
      handleCreatedAssets,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      linkedEntitiesIds,
      // computed
      selectedComputeName,
      baseDrawerFilters,
    };
  },
  async created() {
    const { kind, createdEntityId } = this.$route.query;

    try {
      this.handleCreatedAssets(
        kind?.toString(),
        createdEntityId?.toString(),
        WorkloadCreationDrawerEntities.ModelStorage,
      );
      const formState = this.inferenceStore.getFlatFormState();
      const inferenceTemplateModel = this.inferenceStore.getTemplateModel();
      this.uiState = this.inferenceStore.getInferenceFormUIState();
      if (formState) {
        this.loadFormState();
        this.loadPresetFormModel();
      } else {
        if (!inferenceTemplateModel) {
          await this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW });
          return;
        }

        await this.initTemplateModel(
          inferenceTemplateModel?.clusterId,
          inferenceTemplateModel?.scopeType,
          inferenceTemplateModel?.scopeId || "",
          inferenceTemplateModel?.name || "",
          inferenceTemplateModel?.description || "",
        );
      }

      this.updateFeaturesByClusterVersion(this.inference.clusterId);
      await this.loadScopeBasedData();

      this.appStore.setPageLoading(false);
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    }
  },
  computed: {
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportsNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
      );
    },
    inferenceSupportedDataSources(): Set<TDataSourceKinds> {
      const inferenceSupported = inferenceWorkloadUtil.getInferenceSupportedDataSources({
        isVersion219OrHigher: this.supportsNfsAndHostPath,
        modelStoreSupportOnly: true,
      });

      return new Set([...(this.allowedDatasourceOptions || [])].filter((ds) => inferenceSupported.includes(ds)));
    },
    drawerFilters(): IAssetsFilter {
      return {
        projectId: +this.inference.projectId,
        complyToProject: +this.inference.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
        isInference: true,
      };
    },
    shouldOpenModelStoreSection(): boolean {
      return (
        this.uiState.accessFromStorage ||
        (!!this.$route.query.createdEntityId &&
          [
            AssetKind.HostPath,
            AssetKind.Nfs,
            AssetKind.Pvc,
            AssetKind.Git,
            AssetKind.S3,
            AssetKind.ConfigMap,
            AssetKind.SecretVolume,
            AssetKind.DataVolume,
          ].includes(this.$route.query.kind as TDataSourceKinds))
      );
    },
    isVolumesEnabled(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.inference.clusterId, MIN_CLUSTER_VERSION_FOR_TOLERATIONS);
    },
    navigationData(): INimInferenceSubmitFormNavigation {
      return {
        showInferenceType: true,
        typeSummary: this.serverType.toString(),
        defaultOpenNameSection: !!this.$route.query.fromCopyId,
        modelName: this.uiState.modelName,
        canChangeModel: true,
      };
    },
    sectionTemplatesData(): INimInferenceSubmitFormSectionTemplates {
      return {
        selectedTemplates: this.selectedTemplates,
        selectedComputeName: this.selectedComputeName,
      };
    },
    clusterConfigData(): INimInferenceSubmitFormClusterConfig {
      return {
        autoScaleOptions: this.autoScaleOptions,
      };
    },
    dataSourcesData(): INimInferenceSubmitFormDataSources {
      return {
        secrets: this.secrets,
        allowedDataSourceOptions: this.allowedDatasourceOptions,
        inferenceSupportedDataSources: this.inferenceSupportedDataSources,
        shouldOpenModelStoreSection: this.shouldOpenModelStoreSection,
      };
    },
    featuresData(): INimInferenceSubmitFormFeatures {
      return {
        supportMultiFractionGpu: this.supportMultiFractionGpu,
        supportTolerations: this.supportTolerations,
        supportsDockerRegistry: false,
      };
    },
    uiStateData(): INimInferenceSubmitFormUIState {
      return {
        isDrawerOpen: this.isDrawerOpen,
        drawerDataEntity: this.drawerDataEntity,
        recentlyCreatedEntityId: this.recentlyCreatedEntityId,
        submitting: this.submitting,
        uiState: this.uiState,
        drawerFilters: this.drawerFilters,
        isWorkloadTemplate: true,
        originalTemplateName: this.originalName,
        scopeChangeInProgress: this.scopeChangeInProgress,
        isValidScopeSelected: this.isValidScopeSelected,
      };
    },
    isValidScopeSelected(): boolean {
      return this.inference.scopeType !== ScopeType.System && this.inference.scopeType !== ScopeType.Tenant;
    },
  },
  methods: {
    async initTemplateModel(
      clusterId?: string,
      scopeType?: ScopeType,
      scopeId?: string,
      name?: string,
      description?: string,
    ): Promise<void> {
      try {
        // not sending unified request for template on startup as policy plays no part in this form for the moment
        // if fromCopyId is provided, we will send a getById request instead

        const uiModel = inferenceWorkloadUtil.getEmptyUISuperWorkloadModel({
          scopeType,
          scopeId,
          name,
          clusterId,
          description,
        });
        this.inference = inferenceWorkloadUtil.setInferenceNIMUIDefaults(uiModel, this.authStore.isSSO);

        this.createSectionsClearedStates();
      } catch (e: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to load template model",
        });
        this.$q.notify(errorAlert.getNotification(e));
        this.appStore.setFallback(true);
      }
    },
    createSectionsClearedStates(): void {
      const base = deepCopy(this.inference);
      this.selectedTemplates[WorkloadCreationDrawerEntities.Environment].clearedState =
        workloadHybridUtil.getInferenceRelevantEnvironmentFields(base);
      this.selectedTemplates[WorkloadCreationDrawerEntities.Compute].clearedState =
        workloadHybridUtil.getRelevantComputeFields(base);
      this.selectedTemplates[WorkloadCreationDrawerEntities.DataSource].clearedState =
        workloadHybridUtil.getRelevantStorageFields(base);
    },
    async onSubmit(submitType: string): Promise<void> {
      this.inference = inferenceWorkloadUtil.adjustNimWorkloadProperties(this.inference);
      switch (submitType) {
        case SUBMIT_OPTIONS.submitTemplate.value:
          await this.submitInference();
          break;
        case SUBMIT_OPTIONS.advancedSetup.value:
          await this.continueWithAdvancedSetup();
          break;
        default:
          console.warn(`submit type ${submitType} not found`);
      }
    },
    async continueWithAdvancedSetup(): Promise<void> {
      this.saveFormState();
      this.inferenceStore.setPresetFormModel(deepCopy(this.inference));
      this.$router.push({
        name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW,
      });
    },
    async submitInference(): Promise<void> {
      try {
        this.submitting = true;
        const inferenceModelCopy = deepCopy(this.inference);
        this.inference.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(this.inference.spec);
        const workloadCreationRequest = inferenceWorkloadUtil.prepareInferenceTemplateFlatRequest(inferenceModelCopy);
        const createdTemplate = await workloadTemplateService.createInferenceTemplateFlat(workloadCreationRequest);
        this.$q.notify(alertUtil.getSuccess(`Template ${createdTemplate.meta.name} created`));
        this.clearFormState();
        this.$router.push({
          name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
          query: { clusterId: this.inference.clusterId, createdEntityId: createdTemplate.meta.id },
        });
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("template"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.submitting = false;
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.clearFormState();
        this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX });
      }
    },
    onLeavePage(): void {
      this.clearFormState();
    },
    clearFormState(): void {
      this.inferenceStore.clearFlatFormState();
      this.inferenceStore.clearInferenceFormUIState();
      this.inferenceStore.clearPresetFormModel();
      this.inferenceStore.clearTemplateModel();
    },
    onBack(): void {
      this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW });
    },
    async loadAvailableDataSource(): Promise<void> {
      this.allowedDatasourceOptions = await this.dataSourceStore.getAvailableDataSources();
      if (this.dataSourceStore.fallbackPolicyLoadingError) this.appStore.setFallback(true);
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.onCreateNewSetup(type, { workloadType: EWorkloadFormType.Inference });
    },
    saveFormState(): void {
      const formState: THybridFormState = {
        projectInfo: {},
        projectNodePoolsResources: [],
        selectedTemplates: this.selectedTemplates,
        policy: undefined,
        policyViewDetails: {} as IPolicyViewDetails,
        selectedProjectNodeTypes: [],
        workloadModel: this.inference,
      };
      this.inferenceStore.setFlatFormState(deepCopy(formState));
      const modelName = inferenceWorkloadUtil.getNimModelNameFromImage(this.inference.spec.image || "");
      this.uiState.modelName = modelName;
      this.inferenceStore.setInferenceFormUIState(this.uiState);
    },
    loadFormState(): void {
      const formState = this.inferenceStore.getFlatFormState();
      if (!formState) return;
      this.selectedTemplates = formState?.selectedTemplates;
      this.inference = formState?.workloadModel;
      this.uiState = this.inferenceStore.getInferenceFormUIState();
    },
    loadPresetFormModel(): void {
      const presetFormModel = this.inferenceStore.getPresetFormModel();
      if (!presetFormModel) return;
      this.inference = presetFormModel;
      this.inferenceStore.clearPresetFormModel();
    },
    getClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    updateFeaturesByClusterVersion(clusterId: string): void {
      const cluster = this.getClusterById(clusterId);
      if (!cluster) return;

      this.autoScaleOptions.latencyEnabled = inferenceWorkloadUtil.isLatencyEnableOnCluster(cluster);
      if (this.clusterStore.isClusterVersionSufficient(cluster.uuid, MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED)) {
        this.autoScaleOptions.isAutoScaleEnabled = inferenceWorkloadUtil.isAutoScaleEnableOnCluster(cluster);
      }
    },
    updateModelStorage(storage: IUIStorageFields): void {
      this.inference.spec.storage = storage;
    },
    async loadSecrets(): Promise<void> {
      if (!this.isValidScopeSelected) {
        this.secrets = [];
        return;
      }
      try {
        this.secrets = await credentialService.listExistingSecrets(
          this.inference.scopeType as ScopeType,
          this.inference.scopeId as string,
          undefined,
          "all",
          true,
        );
      } catch (error: unknown) {
        console.error("Failed to load secrets:\n", error);
        this.$q.notify(alertUtil.getError("Failed to load secrets"));
      }
    },
    onAccessUIStateChanged(uiState: IInferenceFormUIState): void {
      this.uiState = uiState;
      this.saveFormState();
    },
    handleDrawerAction(drawerAction: INimInferenceSubmitFormDrawerAction): void {
      switch (drawerAction.action) {
        case ENimInferenceSubmitFormDrawerActionType.Close:
          this.closeDrawer();
          break;
        case ENimInferenceSubmitFormDrawerActionType.Select:
          if (drawerAction.selectData) {
            this.handleSelection(drawerAction.selectData);
          }
          break;
        case ENimInferenceSubmitFormDrawerActionType.SelectMultiple:
          if (drawerAction.selectMultipleData) {
            this.handleMultiSelection(drawerAction.selectMultipleData);
          }
          break;
        case ENimInferenceSubmitFormDrawerActionType.CreateNew:
          this.onCreateNew(drawerAction.createNewData || "");
          break;
      }
    },
    handleModelValueUpdate(inference: IUISuperWorkloadCreationRequest): void {
      this.inference = inference;
    },
    handleNavigationAction(): void {
      this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW });
    },
    async handleScopeChange(scope: IScopeModel): Promise<void> {
      this.inference.scopeType = scope.scope as ScopeType;
      this.inference.scopeId = scopeUtil.getScopeId(scope);
      this.inference.clusterId = scope.clusterId || scope.parentClusterUuid || "";

      try {
        this.scopeChangeInProgress = true;
        await this.loadScopeBasedData();
        await this.validateDrawerEntitiesForScopeChange(this.secrets.map((secret) => secret.name));
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to load scope based data",
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.scopeChangeInProgress = false;
      }
    },
    async loadScopeBasedData(): Promise<void> {
      await Promise.all([this.loadSecrets(), this.loadAvailableDataSource()]);
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default TemplateInferenceNim as typeof TemplateInferenceNim & DefineComponent;
</script>
