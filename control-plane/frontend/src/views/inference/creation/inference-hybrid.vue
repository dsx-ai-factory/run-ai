<template>
  <section class="inference-hybrid column items-center q-pt-md">
    <inference-submit-form
      v-if="isPageReady"
      :model-value="inference"
      :navigation="navigationData"
      :policy="policyData"
      :section-templates="sectionTemplatesData"
      :cluster-config="clusterConfigData"
      :data-sources="dataSourcesData"
      :features="featuresData"
      :ui-state="uiStateData"
      :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
      :project="selectedProject"
      :trigger-validation-on-start-up="triggerValidationOnStartUp"
      @leave-page="redirectToPrevRoute"
      @canceled="onCancel"
      @submit="submitInference"
      @navigation-action="handleNavigationAction"
      @display-environments-list="displayEnvironmentsList"
      @section-action="handleSectionAction"
      @display-computes-list="displayComputesList"
      @display-data-sources-list="displayDataSourcesList"
      @drawer-action="handleDrawerAction"
      @update:model-value="handleModelValueUpdate"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// components
import { InferenceSubmitForm } from "@/components/workload-creation/inference/inference-submit-form";

// models
import { EWorkloadFormType, EWorkloadType, EWorkloadName } from "@/models/workload.model";
import { Scope, type StringOption } from "@/swagger-models/assets-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type {
  UnifiedPolicyInfoPerReplicaType,
  UnifiedProjectInfo,
  UnifiedWorkload,
  Inference as InferenceFlat,
  Workload as WorkloadFlat,
} from "@/swagger-models/workloads-client";
import { SettingKeys } from "@/models/setting.model";
import type { WorkspacePolicyDefaultsAndRulesV2 } from "@/swagger-models/policy-service-client";
import { getPolicyViewDetails, type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";

import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type {
  IInferenceSubmitFormClusterConfig,
  IInferenceSubmitFormDataSources,
  IInferenceSubmitFormFeatures,
  IInferenceSubmitFormNavigation,
  IInferenceSubmitFormPolicy,
  IInferenceSubmitFormSectionTemplates,
  IInferenceSubmitFormUIState,
  IInferenceSubmitFormDrawerAction,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import {
  EInferenceSubmitFormNavigationActionType,
  EInferenceSubmitFormNavigationSection,
  EInferenceSubmitFormDrawerActionType,
  EInferenceSubmitFormSectionActionType,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadStore } from "@/stores/workload.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util/common.util";
import { policyUtil, clearWorkloadBlockedFields } from "@/utils/policy.util";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

// routes
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
} from "@/common/version.constant";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useRoute } from "vue-router";
import { EServerType } from "@/models/inference.model";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";

export default defineComponent({
  name: "inference-hybrid",
  components: {
    InferenceSubmitForm,
  },
  data() {
    return {
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      workloadStore: useWorkloadStore(),
      dataSourceStore: useDataSourceStore(),
      submitting: false as boolean,
      clusterName: "" as string,
      showInferenceType: false as boolean,
      projectInfo: {} as UnifiedProjectInfo,
      policy: undefined as UnifiedPolicyInfoPerReplicaType | undefined | null,
      policyViewDetails: {} as IPolicyViewDetails,
      selectedProjectNodeTypes: [] as string[],
      projectNodePoolsResources: [] as StringOption[],
      autoScaleOptions: {
        latencyEnabled: true,
        isAutoScaleEnabled: true,
      } as IAutoScaleOptions,
      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,
      EWorkloadFormType,
      showRequestTimeout: false as boolean,
      supportImagePullSecrets: true as boolean,

      triggerValidationOnStartUp: false as boolean,
    };
  },
  setup() {
    const route = useRoute();
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariables } =
      useEnvironmentVariablesSection();

    const {
      // state
      isDrawerOpen,
      drawerDataEntity,
      hybridWorkloadModel,
      selectedTemplates,
      recentlyCreatedEntityId,
      // methods
      closeDrawer,
      displayEnvironmentsList,
      displayComputesList,
      displayDataSourcesList,
      handleSelection,
      handleMultiSelection,
      clearComputeTemplate,
      clearEnvironmentTemplate,
      resetEnvironmentToDefault,
      resetComputeToDefault,
      handleCreatedAssets,
      getUnifiedHybridWorkload,
      onCreateNewSetup,
      // computed
      selectedEnvironmentName,
      selectedComputeName,
    } = useHybridTemplates(EHybridFormTypes.Inference);

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route.query.clusterId?.toString());

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariables,

      // state
      isDrawerOpen,
      drawerDataEntity,
      inference: hybridWorkloadModel,
      selectedTemplates,
      recentlyCreatedEntityId,
      // methods
      closeDrawer,
      displayEnvironmentsList,
      displayComputesList,
      displayDataSourcesList,
      handleSelection,
      handleMultiSelection,
      clearComputeTemplate,
      clearEnvironmentTemplate,
      resetEnvironmentToDefault,
      resetComputeToDefault,
      handleCreatedAssets,
      getUnifiedHybridWorkload,
      onCreateNewSetup,
      // computed
      selectedEnvironmentName,
      selectedComputeName,

      // workload priorities composable
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
    };
  },
  async created() {
    const { fromCopyId, projectId, clusterId, inferenceName, kind, createdEntityId, templateId } = this.$route.query;
    this.showInferenceType =
      this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) ||
      this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
    this.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());

    try {
      const formState = this.inferenceStore.getFlatFormState();

      if (formState) {
        this.loadFormState();
        if (!this.validateAndInitializeCluster(this.inference.clusterId)) return;
      } else {
        if (!this.validateAndInitializeCluster(clusterId?.toString())) return;
        await this.initModel(
          fromCopyId?.toString(),
          projectId?.toString(),
          clusterId?.toString(),
          inferenceName?.toString(),
          templateId?.toString(),
        );
      }

      await Promise.all([
        this.loadSourceValueOptions(true, this.inference.projectId),
        this.loadSecrets(),
        this.loadAvailableDataSource(),
        this.loadWorkloadPriorities(this.inference.clusterId),
        this.loadProjectById(this.inference.projectId.toString()),
      ]);

      if (this.supportPriorityClass) {
        const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Inference);
        if (defaultPriority && !this.inference.spec.priorityClass && !this.$route.query.fromCopyId) {
          this.inference.spec.priorityClass = defaultPriority;
        }

        if (this.supportPreemptibility) {
          const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Inference);
          if (!this.inference.spec.preemptibility && !this.$route.query.fromCopyId) {
            this.inference.spec.preemptibility = defaultPreemptibility;
          }
        }
      }
      this.supportImagePullSecrets = this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
      );
      this.appStore.setPageLoading(false);
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    }
  },
  computed: {
    modelName(): string | undefined {
      if (!this.inferenceStore.isHuggingFaceOrNimModel) return undefined;
      return this.inferenceStore.inferenceFormUIState.modelName;
    },
    typeSummary(): string {
      return this.inferenceStore.inferenceFormUIState.modelBasedOn?.toString() || EServerType.Custom;
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    projectName(): string {
      return this.projectInfo.name || "-";
    },
    navigationData(): IInferenceSubmitFormNavigation {
      return {
        clusterName: this.clusterName,
        projectName: this.projectName,
        showInferenceType: this.showInferenceType,
        typeSummary: this.typeSummary,
        modelName: this.modelName || undefined,
        canChangeModel: true,
        defaultOpenNameSection: !!this.$route.query.fromCopyId,
      };
    },
    policyData(): IInferenceSubmitFormPolicy | undefined {
      if (!this.policy) return undefined;

      return {
        policyData: this.policy,
        policyViewDetails: this.policyViewDetails,
      };
    },
    sectionTemplatesData(): IInferenceSubmitFormSectionTemplates {
      return {
        selectedTemplates: this.selectedTemplates,
        selectedEnvironmentName: this.selectedEnvironmentName,
        selectedComputeName: this.selectedComputeName,
      };
    },
    clusterConfigData(): IInferenceSubmitFormClusterConfig {
      return {
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        projectNodePoolsResources: this.projectNodePoolsResources,
        autoScaleOptions: this.autoScaleOptions,
        workloadPriorities: this.workloadPriorities,
      };
    },
    dataSourcesData(): IInferenceSubmitFormDataSources {
      return {
        secrets: this.secrets,
        allowedDataSourceOptions: this.allowedDatasourceOptions || new Set(),
      };
    },
    featuresData(): IInferenceSubmitFormFeatures {
      return {
        supportImagePullSecrets: this.supportImagePullSecrets,
        isVolumesEnabled: this.clusterStore.isClusterVersionSufficient(
          this.inference.clusterId || "",
          MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
        ),
        showRequestTimeout: this.showRequestTimeout,
        supportPriorityClass: this.supportPriorityClass,
        supportPreemptibility: this.supportPreemptibility,
      };
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },
    uiStateData(): IInferenceSubmitFormUIState {
      return {
        submitting: this.submitting,
        isDrawerOpen: this.isDrawerOpen,
        drawerDataEntity: this.drawerDataEntity || null,
        recentlyCreatedEntityId: this.recentlyCreatedEntityId,
      };
    },
  },
  methods: {
    updateShowRequestTimeout(clusterId: string): void {
      this.showRequestTimeout = this.clusterStore.isClusterVersionSufficient(
        clusterId,
        MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
      );
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
      this.updateShowRequestTimeout(this.inference.clusterId);
    },
    async initModel(
      workloadId?: string,
      projectId?: string,
      clusterId?: string,
      workloadName?: string,
      templateId?: string,
    ): Promise<void> {
      try {
        const unifiedWorkload: UnifiedWorkload = await this.getUnifiedHybridWorkload({
          workloadId,
          projectId,
          clusterId,
          templateId,
        });

        if (templateId) this.triggerValidationOnStartUp = true;

        this.policy = unifiedWorkload.policyInfo?.worker;
        this.policyViewDetails = getPolicyViewDetails(
          (this.policy || {}) as unknown as WorkspacePolicyDefaultsAndRulesV2,
        );

        this.projectInfo = unifiedWorkload.projectInfo || {};
        this.projectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];

        const uiModel = inferenceWorkloadUtil.getHybridInferenceModel(unifiedWorkload, true);

        this.updateShowRequestTimeout(clusterId || "");
        this.inference = inferenceWorkloadUtil.setInferenceUIDefaults(uiModel, this.authStore.isSSO, this.policy);

        clearWorkloadBlockedFields(this.inference.spec, this.policy);
        this.createSectionsClearedStates();
        if (projectId) this.inference.projectId = projectId;
        if (clusterId) this.inference.clusterId = clusterId;
        if (workloadName) this.inference.name = workloadName;
        if (templateId) this.inference.templateId = templateId;
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
    async submitInference(): Promise<void> {
      try {
        this.submitting = true;

        // Copying the inference model to avoid mutating the original model and make UI glitches
        const inferenceModelCopy = deepCopy(this.inference);
        inferenceModelCopy.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(inferenceModelCopy.spec, this.policy);

        const workloadCreationRequest = inferenceWorkloadUtil.prepareInferenceFlatRequest(inferenceModelCopy);

        const inferenceWorkload: InferenceFlat = await inferenceService.createInferenceFlatApi(workloadCreationRequest);
        this.clearFormState();

        const workload: WorkloadFlat = inferenceWorkloadUtil.convertInferenceToWorkload(
          inferenceWorkload,
          this.authStore.tenant.id,
        );

        this.workloadStore.setWorkloadAdded(workload, EWorkloadType.Inference);
        this.clusterStore.setSelectedClusterById(this.inference.clusterId);
        this.$q.notify(alertUtil.getSuccess(`Inference ${this.inference.name} created`));
        this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query: { clusterId: this.inference.clusterId } });
        this.inferenceStore.clearInferenceFormUIState();
        this.inferenceStore.clearPresetFormModel();
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("inference"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error("Failed to create inference:\n", error);
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
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      this.clearFormState();
      const query = fromSubmit ? { clusterId: this.inference.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    onBackToModel(): void {
      if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.Nim) {
        this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NIM_EDIT });
      } else if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.VLLM) {
        this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_VLLM_EDIT });
      } else if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.TGI) {
        this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_TGI_EDIT });
      } else {
        console.error(`No model based on ${this.inferenceStore.inferenceFormUIState.modelBasedOn} found`);
      }
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.onCreateNewSetup(type, { workloadType: EWorkloadFormType.Inference });
    },
    handleDrawerAction(drawerAction: IInferenceSubmitFormDrawerAction): void {
      switch (drawerAction.action) {
        case EInferenceSubmitFormDrawerActionType.Select:
          if (drawerAction.selectData) {
            this.handleSelection(drawerAction.selectData);
          }
          break;
        case EInferenceSubmitFormDrawerActionType.SelectMultiple:
          if (drawerAction.selectMultipleData) {
            this.handleMultiSelection(drawerAction.selectMultipleData);
          }
          break;
        case EInferenceSubmitFormDrawerActionType.CreateNew:
          this.onCreateNew(drawerAction.createNewData || "");
          break;
        case EInferenceSubmitFormDrawerActionType.Close:
          this.closeDrawer();
          break;
        default:
          console.warn(`Unknown drawer action: ${drawerAction.action}`);
      }
    },
    handleSectionAction(sectionAction: { action: EInferenceSubmitFormSectionActionType }): void {
      switch (sectionAction.action) {
        case EInferenceSubmitFormSectionActionType.ResetEnvironmentToDefault:
          this.resetEnvironmentToDefault();
          break;
        case EInferenceSubmitFormSectionActionType.ClearEnvironmentTemplate:
          this.clearEnvironmentTemplate();
          break;
        case EInferenceSubmitFormSectionActionType.ResetComputeToDefault:
          this.resetComputeToDefault();
          break;
        case EInferenceSubmitFormSectionActionType.ClearComputeTemplate:
          this.clearComputeTemplate();
          break;
        default:
          console.warn(`Unknown section action: ${sectionAction.action}`);
      }
    },
    handleNavigationAction(navigationAction: {
      action: EInferenceSubmitFormNavigationActionType;
      section?: EInferenceSubmitFormNavigationSection;
    }): void {
      switch (navigationAction.action) {
        case EInferenceSubmitFormNavigationActionType.ChangeSection:
          if (navigationAction.section === EInferenceSubmitFormNavigationSection.Model) {
            this.onBackToModel();
          } else {
            this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
          }
          break;
        default:
          console.warn(`Unknown navigation action: ${navigationAction.action}`);
      }
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
    async loadProjectById(projectId: string): Promise<Project | void> {
      try {
        this.selectedProject = await orgUnitService.getProject(projectId);
      } catch (error: unknown) {
        console.error(error);
        this.appStore.setFallback(true);
      }
    },
    handleModelValueUpdate(inference: IUISuperWorkloadCreationRequest): void {
      this.inference = inference;
    },
  },
});
</script>
