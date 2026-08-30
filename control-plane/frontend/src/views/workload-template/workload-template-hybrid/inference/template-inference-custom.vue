<template>
  <section class="inference-template-hybrid template-hybrid-custom column items-center q-pt-md">
    <inference-submit-form
      v-if="isPageReady"
      :model-value="inference"
      :navigation="navigationData"
      :section-templates="sectionTemplatesData"
      :cluster-config="clusterConfigData"
      :data-sources="dataSourcesData"
      :features="featuresData"
      :ui-state="uiStateData"
      :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
      :project="selectedProject"
      @leave-page="onLeavePage"
      @canceled="onCancel"
      @submit="submitTemplate"
      @navigation-action="handleNavigationAction"
      @display-environments-list="displayEnvironmentsList"
      @section-action="handleSectionAction"
      @display-computes-list="displayComputesList"
      @display-data-sources-list="displayDataSourcesList"
      @drawer-action="handleDrawerAction"
      @update:model-value="handleModelValueUpdate"
      @toggle-linked="handleToggleLinked"
      @scope-changed="handleScopeChange"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// components
import { InferenceSubmitForm } from "@/components/workload-creation/inference/inference-submit-form";

// models
import { EWorkloadFormType, EWorkloadName } from "@/models/workload.model";
import { Scope, ScopeType } from "@/swagger-models/assets-service-client";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { SettingKeys } from "@/models/setting.model";
import { type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
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
  IInferenceSubmitFormSectionTemplates,
  IInferenceSubmitFormUIState,
  IInferenceSubmitFormDrawerAction,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import {
  EInferenceSubmitFormNavigationActionType,
  EInferenceSubmitFormNavigationSection,
  EInferenceSubmitFormDrawerActionType,
  EInferenceSubmitFormSectionActionType,
  EInferenceSubmitFormLinkActionType,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { IScopeModel } from "@/models/global.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { deepCopy } from "@/utils/common.util/common.util";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";
import { policyUtil } from "@/utils/policy.util/policy.util";
import { scopeUtil } from "@/utils/scope.util";
import { inferenceTemplateUtil } from "@/utils/inference-template.util/inference-template.util";

// routes
import { INFERENCE_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/inference-routes/inference-template.routes.names";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";

// composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
} from "@/common/version.constant";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { useDataSourceStore } from "@/stores/data-source.store";
import { EServerType } from "@/models/inference.model";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

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
      dataSourceStore: useDataSourceStore(),
      submitting: false as boolean,
      isEditMode: false as boolean,
      originalName: "" as string,
      showInferenceType: false as boolean,
      autoScaleOptions: {
        latencyEnabled: true,
        isAutoScaleEnabled: true,
      } as IAutoScaleOptions,
      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,
      EWorkloadFormType,
      showRequestTimeout: false as boolean,
      scopeModel: {} as IScopeModel,
      serverType: EServerType.Custom as EServerType,
      scopeChangeInProgress: false as boolean,
    };
  },
  setup() {
    const inferenceStore = useInferenceStore();
    const inferenceTemplateModel = inferenceStore.getTemplateModel();
    const clusterId = inferenceTemplateModel?.clusterId || "";
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
      toggleEnvironmentLinked,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      linkedEntitiesIds,
      // computed
      baseDrawerFilters,
      selectedEnvironmentName,
      selectedComputeName,
    } = useHybridTemplates(EHybridFormTypes.Inference, undefined, true);

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(clusterId);

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
      toggleEnvironmentLinked,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      linkedEntitiesIds,
      // computed
      selectedEnvironmentName,
      selectedComputeName,
      baseDrawerFilters,
      // workload priorities composable
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
    };
  },
  async created() {
    const { fromCopyId, kind, createdEntityId } = this.$route.query;
    const { id: editId } = this.$route.params;

    const formState = this.inferenceStore.getFlatFormState();
    const inferenceTemplateModel = this.inferenceStore.getTemplateModel();

    if (editId) this.isEditMode = true;

    this.showInferenceType =
      this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) ||
      this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
    this.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());

    try {
      if (formState) {
        this.loadFormState();
      } else {
        if (!inferenceTemplateModel && !fromCopyId && !editId) {
          await this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW });
          return;
        }

        await this.initTemplateModel(
          inferenceTemplateModel?.clusterId,
          fromCopyId?.toString() || editId?.toString(),
          inferenceTemplateModel?.scopeType,
          inferenceTemplateModel?.scopeId || "",
          inferenceTemplateModel?.name || "",
          inferenceTemplateModel?.description || "",
        );
      }

      this.scopeModel = scopeUtil.getScopeModelFromTypeAndId(
        this.inference.scopeType as ScopeType,
        this.inference.scopeId as string,
      );

      await this.loadScopeBasedData();
      this.validateScopeBasedData();

      if (this.supportPriorityClass) {
        const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Inference);
        if (defaultPriority && !this.inference.spec.priorityClass) {
          this.inference.spec.priorityClass = defaultPriority;
        }

        if (this.supportPreemptibility) {
          const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Inference);
          if (!this.inference.spec.preemptibility) {
            this.inference.spec.preemptibility = defaultPreemptibility;
          }
        }
      }
      this.appStore.setPageLoading(false);
    } catch (error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: "Failed to load assets",
      });
      this.$q.notify(errorAlert.getNotification(error));
      this.appStore.setFallback(true);
    }
  },
  computed: {
    modelName(): string | undefined {
      if (!this.inferenceStore.isHuggingFaceOrNimModel) return undefined;
      return this.inferenceStore.inferenceFormUIState.modelName;
    },
    typeSummary(): string {
      return this.inferenceStore.inferenceFormUIState.modelBasedOn?.toString() || this.serverType.toString();
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    navigationData(): IInferenceSubmitFormNavigation {
      return {
        showInferenceType: this.showInferenceType,
        typeSummary: this.typeSummary,
        modelName: this.modelName || undefined,
        canChangeModel: true,
        defaultOpenNameSection: !!this.$route.query.fromCopyId,
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
        supportImagePullSecrets: false,
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
        isWorkloadTemplate: true,
        originalTemplateName: this.originalName,
        isEditMode: this.isEditMode,
        scopeChangeInProgress: this.scopeChangeInProgress,
        baseDrawerFilters: this.baseDrawerFilters,
        isValidScopeSelected: this.isValidScopeSelected,
      };
    },
    isValidScopeSelected(): boolean {
      return this.scopeModel.scope !== ScopeType.System && this.scopeModel.scope !== ScopeType.Tenant;
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
    saveFormState(): void {
      const formState: THybridFormState = {
        projectInfo: {},
        selectedTemplates: this.selectedTemplates,
        projectNodePoolsResources: [],
        policy: undefined,
        policyViewDetails: {} as IPolicyViewDetails,
        selectedProjectNodeTypes: [],
        workloadModel: this.inference,
      };
      this.inferenceStore.setFlatFormState(deepCopy(formState));
    },
    loadFormState(): void {
      const formState = this.inferenceStore.getFlatFormState();
      if (!formState) return;
      this.selectedTemplates = formState?.selectedTemplates;
      this.inference = formState?.workloadModel;
      if (this.isEditMode) {
        this.originalName = this.inference.name;
        this.appStore.setPageTitleOverride(`Edit inference template: ${this.inference.name}`);
      }
      this.updateShowRequestTimeout(this.inference.clusterId);
    },
    async initTemplateModel(
      clusterId?: string,
      fromCopyId?: string,
      scopeType?: ScopeType,
      scopeId?: string,
      name?: string,
      description?: string,
    ): Promise<void> {
      try {
        // not sending unified request for template on startup as policy plays no part in this form for the moment
        // if fromCopyId is provided, we will send a getById request instead

        if (fromCopyId) {
          const workloadTemplate = await workloadTemplateService.getInferenceTemplateById(fromCopyId);

          if (workloadTemplate.combinedSpec) workloadTemplate.spec = workloadTemplate.combinedSpec;
          if (workloadTemplate.assets) await this.setupLinkedEntities(workloadTemplate.assets);

          this.updateShowRequestTimeout(workloadTemplate.meta.clusterId || "");
          const uiModel = inferenceTemplateUtil.getTemplateModel(workloadTemplate);

          this.inference = uiModel;

          if (this.isEditMode) {
            this.originalName = uiModel.name;
            this.appStore.setPageTitleOverride(`Edit inference template: ${uiModel.name}`);
          }
        } else {
          this.updateShowRequestTimeout(clusterId || "");

          this.inference = inferenceWorkloadUtil.setInferenceUIDefaults(
            inferenceWorkloadUtil.getEmptyUISuperWorkloadModel({
              scopeType,
              scopeId,
              name,
              clusterId,
              description,
            }),
            this.authStore.isSSO,
          );
        }

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
    async submitTemplate(): Promise<void> {
      try {
        this.submitting = true;
        let createdTemplateId: string | undefined;
        if (this.isEditMode) {
          await this.updateTemplate(deepCopy(this.inference));
        } else {
          createdTemplateId = await this.createTemplate(deepCopy(this.inference));
        }
        this.inferenceStore.clearTemplateModel();
        const query = {
          clusterId: this.inference.clusterId,
          ...(createdTemplateId && { createdEntityId: createdTemplateId }),
        };
        this.$router.push({
          name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX,
          query,
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
    async updateTemplate(template: IUISuperWorkloadCreationRequest): Promise<void> {
      const request = inferenceTemplateUtil.prepareTemplateRequestToUpdate(template, this.linkedEntitiesIds);
      const updatedTemplate = await workloadTemplateService.updateInferenceTemplateFlat(
        this.$route.params.id as string,
        request,
      );
      this.$q.notify(alertUtil.getSuccess(`Template ${updatedTemplate.meta.name} updated`));
    },
    async createTemplate(template: IUISuperWorkloadCreationRequest): Promise<string> {
      this.inference.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(template.spec);
      const workloadCreationRequest = inferenceWorkloadUtil.prepareInferenceTemplateFlatRequest(
        template,
        this.linkedEntitiesIds,
      );
      const createdTemplate = await workloadTemplateService.createInferenceTemplateFlat(workloadCreationRequest);
      this.$q.notify(alertUtil.getSuccess(`Template ${createdTemplate.meta.name} created`));
      return createdTemplate.meta.id;
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
    onBackToModel(): void {
      if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.Nim) {
        this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NIM });
      } else if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.VLLM) {
        this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_VLLM });
      } else if (this.inferenceStore.inferenceFormUIState.modelBasedOn === EServerType.TGI) {
        this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_TGI });
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
            this.$router.push({ name: INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW });
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
    async loadAvailableDataSource(): Promise<void> {
      this.allowedDatasourceOptions = await this.dataSourceStore.getAvailableDataSources();
      if (this.dataSourceStore.fallbackPolicyLoadingError) this.appStore.setFallback(true);
    },
    handleModelValueUpdate(inference: IUISuperWorkloadCreationRequest): void {
      this.inference = inference;
    },
    async loadProjectData(): Promise<void> {
      if (!this.scopeModel.projectId) return;

      try {
        this.selectedProject = await orgUnitService.getProject(this.scopeModel.projectId.toString());
      } catch (error: unknown) {
        console.error("Failed to load project data:", error);
        this.appStore.setFallback(true);
      }
    },
    handleToggleLinked(section: EInferenceSubmitFormLinkActionType): void {
      switch (section) {
        case EInferenceSubmitFormLinkActionType.Environment:
          this.toggleEnvironmentLinked();
          break;
        case EInferenceSubmitFormLinkActionType.Compute:
          this.toggleComputeLinked();
          break;
        default:
          console.warn(`Unknown linked section: ${section}`);
      }
    },
    async handleScopeChange(scope: IScopeModel): Promise<void> {
      this.inference.scopeType = scope.scope as ScopeType;
      this.inference.scopeId = scopeUtil.getScopeId(scope);
      this.inference.clusterId = scope.clusterId || scope.parentClusterUuid || "";
      this.scopeModel = scopeUtil.getScopeModelFromTypeAndId(
        this.inference.scopeType as ScopeType,
        this.inference.scopeId as string,
      );

      try {
        this.scopeChangeInProgress = true;
        await this.loadScopeBasedData();
        await this.validateDrawerEntitiesForScopeChange(this.secrets.map((secret) => secret.name));
        this.validateScopeBasedData();
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
      await Promise.all([
        this.loadSourceValueOptions(
          true,
          this.scopeModel.projectId && this.scopeModel.scope === Scope.Project
            ? this.scopeModel.projectId.toString()
            : undefined,
          this.scopeModel.departmentId && this.scopeModel.scope === Scope.Department
            ? this.scopeModel.departmentId.toString()
            : undefined,
          this.scopeModel.clusterId && this.scopeModel.scope === Scope.Cluster
            ? this.scopeModel.clusterId.toString()
            : undefined,
        ),
        this.loadSecrets(),
        this.loadAvailableDataSource(),
        this.loadWorkloadPriorities(this.inference.clusterId || ""),
        this.loadProjectData(),
      ]);
    },
    validateScopeBasedData(): void {
      this.validateEnvironmentVariables(this.inference.spec.environmentVariables || []);
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});
</script>
