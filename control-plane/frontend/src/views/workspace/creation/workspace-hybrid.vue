<template>
  <section class="workspace-hybrid column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :greedy="triggerValidationOnStartUp"
      :form-state="workspaceModel"
      :submitting="submitting"
      submit-btn-text="Create Workspace"
      cancel-btn-text="Cancel"
      init-dirty
      :blurred="isDrawerOpen"
      @unblur="closeDrawer"
      @leave-page="redirectToPrevRoute"
      @canceled="onCancel"
      @submit="onSubmit"
    >
      <template v-if="!appStore.isPageLoading">
        <title-and-back-section label="Cluster" :summary="clusterName" back-tooltip="Change cluster" @on-back="onBack" />

        <title-and-back-section
          label="Project"
          :summary="projectName"
          back-tooltip="Change project"
          aid="project-section"
          @on-back="onBack"
        />

        <workload-name-section
          :default-open="!!$route.query.fromCopyId"
          entity-type="workspace"
          :name="workspaceModel.name"
          @update:name="onNameChanged"
          :project-id="+workspaceModel.projectId"
          :cluster-id="workspaceModel.clusterId"
        />
        <hybrid-environment-section
          :key="`he-${selectedTemplates.environment.sectionKey}`"
          :render-key="selectedTemplates.environment.renderKey.toString()"
          :selected-name="selectedEnvironmentName"
          :workload-spec="workspaceModel.spec"
          :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
          :policy="policy"
          :required-fields="{ image: true }"
          :cluster-id="workspaceModel.clusterId"
          :support-config-map-as-env-var="supportConfigMapAsEnvVar"
          :support-user-credentials="supportUserCredentials"
          support-authenticated-ngc
          :scope-id="workspaceModel.projectId.toString()"
          scope-type="project"
          :support-image-pull-secrets="supportImagePullSecrets"
          @load-section-templates="displayEnvironmentsList"
          @update="updateModelSpec"
          @reset="resetEnvironmentToDefault"
          @clear="clearEnvironmentTemplate"
          @credential-created="onCredentialCreated"
        />
        <hybrid-compute-resource-section
          :key="`hc-${selectedTemplates['compute-resource'].sectionKey}`"
          :render-key="selectedTemplates['compute-resource'].renderKey.toString()"
          :selected-name="selectedComputeName"
          :workload-spec="workspaceModel.spec"
          :project-node-types="selectedProjectNodeTypes"
          :node-pools-resources="projectNodePoolsResources"
          :policy="policy"
          :support-multi-fraction-gpu="supportMultiFractionGpu"
          :cluster-id="workspaceModel.clusterId"
          :support-tolerations="supportTolerations"
          section-sub-title="Set the resources for your workload"
          @load-section-templates="displayComputesList"
          @update="workspaceModel.spec = $event"
          @reset="resetComputeToDefault"
          @clear="clearComputeTemplate"
        />

        <hybrid-data-source-section
          :default-opened="shouldOpenDataSourceSection"
          :policy-rules="policy?.rules?.storage"
          :policy-defaults="policy?.defaults?.storage"
          :allowed-data-source-types="allowedDatasourceOptions"
          :workload-storage="workspaceModel.spec.storage"
          :cluster-id="workspaceModel.clusterId"
          @load-section-templates="displayDataSourcesList"
          @update="updateModelStorage"
          :secrets="secrets"
        />

        <general-section
          should-auto-validate
          :workload-form-type="EWorkloadFormType.Workspace"
          :general-model="generalSectionModel"
          :show-backoff-limit="supportBackoffLimit"
          :policy-rules="policy?.rules"
          :itemize-defaults="generalSectionItemizeDefaults"
          :show-termination-grace-period="isClusterSupportTerminationGracePeriod"
          :workload-priority-data="workloadPriorityData"
          :support-preemptibility="supportPreemptibility"
          @general-model-changed="updateGeneralModel"
        />
      </template>
    </workload-form-wrapper>

    <workload-creation-drawer
      :is-open="isDrawerOpen"
      :comply-to="drawerFilters"
      :entity="drawerDataEntity"
      :policy-details="policyViewDetails"
      :created-entity-id="recentlyCreatedEntityId"
      :allowed-data-source-options="allowedDatasourceOptions"
      @close="closeDrawer"
      @select="handleSelection"
      @select-multiple="handleMultiSelection"
      @create-new="onCreateNew"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";
// Components
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";
// Stores & Services
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useClusterStore } from "@/stores/cluster.store";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { deepCopy } from "@/utils/common.util";
import { policyUtil, clearWorkloadBlockedFields } from "@/utils/policy.util";
import { ErrorAlert } from "@/utils/error-alert.util";
// Models
import { AssetKind, Scope, type StringOption } from "@/swagger-models/assets-service-client";
import { EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { getPolicyViewDetails, type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import type { WorkspacePolicyDefaultsAndRulesV2 } from "@/swagger-models/policy-service-client";
import { EWorkloadFormType, EWorkloadName } from "@/models/workload.model";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
} from "@/common/version.constant";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import type { IUIStorageFields, IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import {
  type UnifiedPolicyInfoPerReplicaType,
  type UnifiedProjectInfo,
  type UnifiedWorkloadRequest,
} from "@/swagger-models/workloads-client";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { SettingKeys } from "@/models/setting.model";
import { useSettingStore } from "@/stores/setting.store";
import { useImagePullSecrets } from "@/composables/use-image-pull-secrets.composable";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { useRoute } from "vue-router";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";

const WorkspaceHybrid = defineComponent({
  name: "workspace-hybrid",
  components: {
    WorkloadCreationDrawer,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    TitleAndBackSection,
    WorkloadNameSection,
    WorkloadFormWrapper,
    HybridDataSourceSection,
    GeneralSection,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      workspaceStore: useWorkspaceStore(),
      clusterStore: useClusterStore(),
      workloadStore: useWorkloadStore(),
      dataSourceStore: useDataSourceStore(),
      settingStore: useSettingStore(),

      EWorkloadFormType,
      submitting: false as boolean,

      policy: undefined as UnifiedPolicyInfoPerReplicaType | undefined | null,
      policyViewDetails: {} as IPolicyViewDetails,
      selectedProjectNodeTypes: [] as string[],
      projectInfo: {} as UnifiedProjectInfo,
      projectNodePoolsResources: [] as StringOption[],
      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,

      triggerValidationOnStartUp: false as boolean,
    };
  },
  setup() {
    const route = useRoute();
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
      createSectionsClearedStates,
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
    } = useHybridTemplates(EHybridFormTypes.Workspace);

    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariables } =
      useEnvironmentVariablesSection();

    const { supportImagePullSecrets } = useImagePullSecrets(route.query.clusterId?.toString());

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route.query.clusterId?.toString());

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariables,

      // state
      isDrawerOpen,
      drawerDataEntity,
      workspaceModel: hybridWorkloadModel,
      selectedTemplates,
      recentlyCreatedEntityId,
      // methods
      closeDrawer,
      displayEnvironmentsList,
      displayComputesList,
      displayDataSourcesList,
      handleSelection,
      handleMultiSelection,
      createSectionsClearedStates,
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
      supportImagePullSecrets,

      // workload priorities composable
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
    };
  },
  async created() {
    const { fromCopyId, templateId, projectId, clusterId, kind, createdEntityId, workspaceName } = this.$route.query;

    this.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());

    const formState = this.workspaceStore.getFlatFormState();
    if (formState) this.loadFormState();
    else
      await this.initWorkspaceModel(
        fromCopyId?.toString(),
        projectId?.toString(),
        clusterId?.toString(),
        workspaceName?.toString(),
        templateId?.toString(),
      );

    await Promise.all([
      this.loadSourceValueOptions(true, this.workspaceModel.projectId),
      this.loadSecrets(),
      this.loadAvailableDataSource(),
      this.loadWorkloadPriorities(this.workspaceModel.clusterId),
      this.loadProjectById(this.workspaceModel.projectId.toString()),
    ]);

    if (this.supportPriorityClass) {
      const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Workspace);
      if (defaultPriority && !this.workspaceModel.spec.priorityClass && !this.$route.query.fromCopyId) {
        this.workspaceModel.spec.priorityClass = defaultPriority;
      }

      if (this.supportPreemptibility) {
        const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Workspace);
        if (!this.workspaceModel.spec.preemptibility && !this.$route.query.fromCopyId) {
          this.workspaceModel.spec.preemptibility = defaultPreemptibility;
        }
      }
    }

    this.validateEnvironmentVariables(this.workspaceModel.spec.environmentVariables || []);
    this.appStore.setPageLoading(false);

    if (this.triggerValidationOnStartUp) {
      this.$nextTick(async () => {
        const formWrapperRef = this.$refs.formWrapperRef as typeof WorkloadFormWrapper;
        await formWrapperRef?.validate();
        this.triggerValidationOnStartUp = false;
      });
    }
  },
  computed: {
    shouldOpenDataSourceSection(): boolean {
      return (
        !!this.$route.query.createdEntityId &&
        [
          AssetKind.HostPath,
          AssetKind.Nfs,
          AssetKind.Pvc,
          AssetKind.Git,
          AssetKind.S3,
          AssetKind.ConfigMap,
          AssetKind.SecretVolume,
          AssetKind.DataVolume,
        ].includes(this.$route.query.kind as TDataSourceKinds)
      );
    },
    drawerFilters(): IAssetsFilter {
      return {
        projectId: +this.workspaceModel.projectId,
        complyToProject: +this.workspaceModel.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
        isWorkspace: true,
      };
    },
    clusterName(): string {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === this.workspaceModel.clusterId)?.name || "";
    },
    projectName(): string {
      return this.projectInfo.name || "-";
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        allowOverQuota: this.workspaceModel?.spec?.priorityClass === "interactive-preemptible" || false,
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.workspaceModel.spec?.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.workspaceModel.spec?.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: (this.workspaceModel.spec?.annotations as IItemizedListItem[]) || undefined,
        labels: (this.workspaceModel.spec?.labels as IItemizedListItem[]) || undefined,
        backoffLimit: Number.isInteger(this.workspaceModel.spec?.backoffLimit)
          ? this.workspaceModel.spec?.backoffLimit
          : null,
        terminationGracePeriodSeconds: this.workspaceModel.spec?.terminationGracePeriodSeconds || null,
        priorityClass: this.workspaceModel.spec?.priorityClass || null,
        preemptibility: this.workspaceModel.spec?.preemptibility || null,
      };
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        environmentVariables: this.policy?.defaults?.environmentVariables,
        annotations: this.policy?.defaults?.annotations,
        labels: this.policy?.defaults?.labels,
        tolerations: this.policy?.defaults?.tolerations,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(this.workspaceModel.clusterId);
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    supportMultiFractionGpu(): boolean {
      // TODO: returning false for now because the backend does not support it
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportUserCredentials(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workspaceModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },

    supportsTemplatesV2(): boolean {
      return this.settingStore.isFeatureEnabled(SettingKeys.EnableTemplatesV2);
    },
    workloadPriorityData(): IWorkloadPriorityData {
      const totalGpuRequest = workloadHybridUtil.calculateTotalGpuRequest({
        specCompute: this.workspaceModel.spec?.compute,
      });

      return {
        currProject: this.selectedProject,
        selectedNodePoolsResources: this.workspaceModel.spec?.nodePools || undefined,
        totalGpuRequest: totalGpuRequest,
        workloadPriorities: this.workloadPriorities,
        supportPriorityClass: this.supportPriorityClass,
      };
    },
  },
  methods: {
    onNameChanged(name: string) {
      this.workspaceModel.name = name;
    },
    async onCredentialCreated(): Promise<void> {
      await this.loadSourceValueOptions(true, this.workspaceModel.projectId);
    },
    updateModelSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.workspaceModel.spec = spec;
    },
    updateModelStorage(storage: IUIStorageFields): void {
      this.workspaceModel.spec.storage = storage;
    },
    saveFormState(): void {
      const formState: THybridFormState = {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.projectNodePoolsResources,
        selectedTemplates: this.selectedTemplates,
        policy: this.policy,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.workspaceModel,
      };
      this.workspaceStore.setFlatFormState(deepCopy(formState));
    },
    loadFormState(): void {
      const formState = this.workspaceStore.getFlatFormState();
      if (!formState) return;
      this.projectInfo = formState?.projectInfo;
      this.projectNodePoolsResources = formState?.projectNodePoolsResources;
      this.selectedTemplates = formState?.selectedTemplates;
      this.policy = formState?.policy;
      this.policyViewDetails = formState?.policyViewDetails;
      this.selectedProjectNodeTypes = formState?.selectedProjectNodeTypes;
      this.workspaceModel = formState?.workloadModel;
    },
    clearFormState(): void {
      this.workspaceStore.clearFlatFormState();
    },
    async loadSecrets(): Promise<void> {
      try {
        this.secrets = await credentialService.listExistingSecrets(
          Scope.Project,
          this.workspaceModel.projectId,
          undefined,
          "all",
          true,
        );
      } catch (error: unknown) {
        console.error("Failed to load secrets:\n", error);
        this.$q.notify(alertUtil.getError("Failed to load secrets"));
      }
    },
    async initWorkspaceModel(
      workloadId?: string,
      projectId?: string,
      clusterId?: string,
      workspaceName?: string,
      templateId?: string,
    ): Promise<void> {
      try {
        const request: UnifiedWorkloadRequest = {
          workloadId,
          projectId,
          clusterId,
        };

        if (this.supportsTemplatesV2 && templateId) {
          request.templateId = templateId;
          this.triggerValidationOnStartUp = true;
        } else request.assetId = templateId;

        const unifiedWorkload = await this.getUnifiedHybridWorkload(request);

        this.policy = unifiedWorkload.policyInfo?.worker;
        this.policyViewDetails = getPolicyViewDetails((this.policy || {}) as WorkspacePolicyDefaultsAndRulesV2);

        this.projectInfo = unifiedWorkload.projectInfo || {};
        this.projectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];

        const uiModel = workloadHybridUtil.getHybridWorkspaceModel(unifiedWorkload);
        const clusterVersion = this.clusterStore.getClusterVersion(clusterId || "");
        this.workspaceModel = workloadHybridUtil.setWorkspaceUIDefaults(
          uiModel,
          this.authStore.isSSO,
          this.policy,
          clusterVersion,
        );

        clearWorkloadBlockedFields(this.workspaceModel.spec, this.policy);
        this.createSectionsClearedStates();
        // need to take for account the id coming from copy
        if (projectId) this.workspaceModel.projectId = projectId;
        if (clusterId) this.workspaceModel.clusterId = clusterId;
        if (workspaceName) this.workspaceModel.name = workspaceName;
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load workspace model"));
        console.error("Failed to get unified workspace:\n", e);
        this.appStore.setFallback(true);
      }
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;
        // Copying the workspace model to avoid mutating the original model and make UI glitches
        const workspaceModelCopy = deepCopy(this.workspaceModel);
        workspaceModelCopy.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(workspaceModelCopy.spec, this.policy);

        const request = workloadHybridUtil.prepareWorkloadFlatRequest(workspaceModelCopy);

        const createdWorkspace = await workspaceService.createWorkspaceFlat(request);

        if (createdWorkspace) this.workloadStore.setWorkloadAdded(createdWorkspace);

        this.$q.notify(alertUtil.getSuccess(`Workspace ${this.workspaceModel.name} created`));
        this.redirectToPrevRoute(true);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("workspace"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error("Failed to create workspace:\n", error);
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
    redirectToPrevRoute(fromSubmit?: boolean): void {
      this.clearFormState();

      const query = fromSubmit ? { clusterId: this.workspaceModel.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    onBack(): void {
      this.clearFormState();

      this.$router.push({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW });
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.workspaceModel = {
        ...this.workspaceModel,
        spec: {
          ...this.workspaceModel.spec,
          annotations: generalModel.annotations,
          labels: generalModel.labels,
          backoffLimit: generalModel.backoffLimit,
          terminationGracePeriodSeconds: generalModel.terminationGracePeriodSeconds,
          autoDeletionTimeAfterCompletionSeconds: generalModel.autoDeletionTimeAfterCompletionSeconds,
          priorityClass: this.supportPriorityClass
            ? generalModel.priorityClass
            : generalModel.allowOverQuota
              ? "interactive-preemptible"
              : undefined,
          preemptibility: generalModel.preemptibility,
        },
      };
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.onCreateNewSetup(type, { workloadType: EWorkloadFormType.Workspace });
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
  },
});

//Dima todo - Type assertion to prevent TS7056 while maintaining type safety -- will be refactored later --hide the annoying warning
export default WorkspaceHybrid as typeof WorkspaceHybrid & DefineComponent;
</script>
