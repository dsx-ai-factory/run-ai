<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <nim-inference-submit-form
    v-if="isPageReady"
    :navigation="navigationData"
    :policy="policyData"
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
    @leave-page="redirectToPrevRoute"
    @drawer-action="handleDrawerAction"
  />
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";

// components
import { NimInferenceSubmitForm } from "@/components/workload-creation/inference/nim-inference-submit-form";

// models
import { AssetKind, Scope, type StringOption } from "@/swagger-models/assets-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type {
  UnifiedPolicyInfoPerReplicaType,
  UnifiedProjectInfo,
  UnifiedWorkload,
  WorkspaceCreationRequest,
} from "@/swagger-models/workloads-client";
import type { WorkspacePolicyDefaultsAndRulesV2 } from "@/swagger-models/policy-service-client";
import { getPolicyViewDetails, type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { clearWorkloadBlockedFields, policyUtil } from "@/utils/policy.util";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import {
  EWorkloadFormType,
  EWorkloadType,
  EWorkloadTypeSupportedInAssetsService,
  type IUIStorageFields,
  type IUISuperWorkloadCreationRequest,
} from "@/models/workload.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { Inference, Workload } from "@/swagger-models/workloads-client";

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
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util/common.util";

// routes
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
  MIN_CLUSTER_VERSION_FOR_NGC_API_KEY,
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
  INimInferenceSubmitFormPolicy,
  INimInferenceSubmitFormSectionTemplates,
  INimInferenceSubmitFormClusterConfig,
  INimInferenceSubmitFormDataSources,
  INimInferenceSubmitFormFeatures,
  INimInferenceSubmitFormUIState,
} from "@/components/workload-creation/inference/nim-inference-submit-form/nim-inference-submit-form.types";
import { ENimInferenceSubmitFormDrawerActionType } from "@/components/workload-creation/inference/nim-inference-submit-form/nim-inference-submit-form.types";

const InferenceNimBased = defineComponent({
  name: "nim-dynamic-submission-form",
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
      projectInfo: {} as UnifiedProjectInfo,
      policy: undefined as UnifiedPolicyInfoPerReplicaType | undefined | null,
      policyViewDetails: {} as IPolicyViewDetails,
      selectedProjectNodeTypes: [] as string[],
      projectNodePoolsResources: [] as StringOption[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      autoScaleOptions: {
        latencyEnabled: true,
        isAutoScaleEnabled: true,
      } as IAutoScaleOptions,
      secrets: [] as SecretListInfo[],
      uiState: {} as IInferenceFormUIState,
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
      // computed
      selectedComputeName,
    } = useHybridTemplates(EHybridFormTypes.Inference);
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
      // computed
      selectedComputeName,
    };
  },
  async created() {
    const { fromCopyId, projectId, clusterId, inferenceName, kind, createdEntityId } = this.$route.query;
    try {
      this.handleCreatedAssets(
        kind?.toString(),
        createdEntityId?.toString(),
        WorkloadCreationDrawerEntities.ModelStorage,
      );
      const formState = this.inferenceStore.getFlatFormState();
      this.uiState = this.inferenceStore.getInferenceFormUIState();
      if (formState) {
        this.loadFormState();
        this.loadPresetFormModel();
        if (!this.validateAndInitializeCluster(this.inference.clusterId)) return;
      } else {
        if (!this.validateAndInitializeCluster(clusterId?.toString())) return;
        await this.initModel(
          fromCopyId?.toString(),
          projectId?.toString(),
          clusterId?.toString(),
          inferenceName?.toString(),
        );
      }

      await Promise.all([this.loadSecrets(), this.loadAvailableDataSource()]);

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
    projectName(): string {
      return this.projectInfo.name || "-";
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
        clusterName: this.clusterName,
        projectName: this.projectName,
        showInferenceType: true,
        typeSummary: this.inferenceStore.inference.serverType?.toString() || "",
        defaultOpenNameSection: !!this.$route.query.fromCopyId,
        modelName: this.uiState.modelName,
        canChangeModel: true,
      };
    },
    policyData(): INimInferenceSubmitFormPolicy {
      return {
        policyData: this.policy || {},
        policyViewDetails: this.policyViewDetails,
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
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        projectNodePoolsResources: this.projectNodePoolsResources,
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
        supportsDockerRegistry: this.clusterStore.isClusterVersionSufficient(
          this.inference.clusterId,
          MIN_CLUSTER_VERSION_FOR_NGC_API_KEY, // we want to align it with the NGC API key feature to avoid confusion
        ),
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
      };
    },
  },
  methods: {
    async initModel(workloadId?: string, projectId?: string, clusterId?: string, workloadName?: string): Promise<void> {
      try {
        const unifiedWorkload: UnifiedWorkload = await this.getUnifiedHybridWorkload({
          workloadId,
          projectId,
          clusterId,
        });
        this.policy = unifiedWorkload.policyInfo?.worker;
        this.policyViewDetails = getPolicyViewDetails(
          (this.policy || {}) as unknown as WorkspacePolicyDefaultsAndRulesV2,
        ) as IPolicyViewDetails;
        this.projectInfo = unifiedWorkload.projectInfo || {};
        this.projectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];
        const uiModel = inferenceWorkloadUtil.getHybridInferenceModel(unifiedWorkload);
        this.inference = inferenceWorkloadUtil.setInferenceNIMUIDefaults(uiModel, this.authStore.isSSO, this.policy);

        clearWorkloadBlockedFields(this.inference.spec, this.policy);
        this.createSectionsClearedStates();
        if (projectId) this.inference.projectId = projectId;
        if (clusterId) this.inference.clusterId = clusterId;
        if (workloadName) this.inference.name = workloadName;
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load inference model"));
        console.error("Failed to get unified inference:\n", e);
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
        case SUBMIT_OPTIONS.submit.value:
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
        name: INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID,
      });
    },
    async submitInference(): Promise<void> {
      try {
        this.submitting = true;
        const inferenceModelCopy = deepCopy(this.inference);
        inferenceModelCopy.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(inferenceModelCopy.spec, this.policy);
        const workloadCreationRequest: WorkspaceCreationRequest =
          inferenceWorkloadUtil.prepareInferenceFlatRequest(inferenceModelCopy);

        const nimWorkload: Inference = await inferenceService.createInferenceFlatApi(workloadCreationRequest);
        const workload: Workload = inferenceWorkloadUtil.convertInferenceToWorkload(
          nimWorkload,
          this.authStore.tenant.id,
        );
        this.workloadStore.setWorkloadAdded(workload, EWorkloadType.Inference);
        this.$q.notify(alertUtil.getSuccess(`Inference ${this.inference.name} created`));
        this.clearFormState();
        const query = { clusterId: this.inference.clusterId };
        this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("inference"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.submitting = false;
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    clearFormState(): void {
      this.inferenceStore.clearFlatFormState();
      this.inferenceStore.clearInferenceFormUIState();
      this.inferenceStore.clearPresetFormModel();
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      this.clearFormState();
      const query = fromSubmit ? { clusterId: this.inference.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    onBack(): void {
      this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
    },
    getClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    updateFeaturesByClusterVersion(cluster: ClustersMinimalData): void {
      this.autoScaleOptions.latencyEnabled = inferenceWorkloadUtil.isLatencyEnableOnCluster(cluster);
      if (this.clusterStore.isClusterVersionSufficient(cluster.uuid, MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED)) {
        this.autoScaleOptions.isAutoScaleEnabled = inferenceWorkloadUtil.isAutoScaleEnableOnCluster(cluster);
      }
    },
    validateAndInitializeCluster(clusterId = ""): boolean {
      if (!clusterId) {
        this.redirectToNewInference();
        return false;
      }
      const cluster = this.getClusterById(clusterId);
      if (!cluster) {
        this.redirectToNewInference();
        return false;
      }
      this.setupClusterConfiguration(cluster);
      return true;
    },
    setupClusterConfiguration(cluster: ClustersMinimalData): void {
      this.clusterName = cluster.name;
      this.updateFeaturesByClusterVersion(cluster);
    },
    redirectToNewInference(): void {
      this.$router.replace({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
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
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.projectNodePoolsResources,
        selectedTemplates: this.selectedTemplates,
        policy: this.policy,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
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
      this.projectInfo = formState?.projectInfo;
      this.projectNodePoolsResources = formState?.projectNodePoolsResources;
      this.selectedTemplates = formState?.selectedTemplates;
      this.policy = formState?.policy;
      this.policyViewDetails = formState?.policyViewDetails;
      this.selectedProjectNodeTypes = formState?.selectedProjectNodeTypes;
      this.inference = formState?.workloadModel;
      this.uiState = this.inferenceStore.getInferenceFormUIState();
    },
    loadPresetFormModel(): void {
      const presetFormModel = this.inferenceStore.getPresetFormModel();
      if (!presetFormModel) return;
      this.inference = presetFormModel;
      this.inferenceStore.clearPresetFormModel();
    },
    updateModelStorage(storage: IUIStorageFields): void {
      this.inference.spec.storage = storage;
    },
    async loadSecrets(): Promise<void> {
      try {
        this.secrets = await credentialService.listExistingSecrets(
          Scope.Project,
          this.inference.projectId,
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
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default InferenceNimBased as typeof InferenceNimBased & DefineComponent;
</script>
