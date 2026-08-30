<template>
  <section class="training-template-hybrid column items-center q-pt-md">
    <workload-form-wrapper
      :form-state="templateModel"
      :submitting="submitting"
      :submit-btn-text="`${isEditMode ? 'Save' : 'Create'} Template`"
      cancel-btn-text="Cancel"
      init-dirty
      :blurred="isDrawerOpen"
      @unblur="closeDrawer"
      @leave-page="redirectToPrevRoute"
      @canceled="onCancel"
      @submit="onSubmit"
    >
      <template v-if="!appStore.isPageLoading">
        <runai-scope-section
          :read-only="isEditMode"
          kind-of-asset="template"
          :scope-model="scopeModel"
          :allowed-scopes="allowedScopes"
          :key="templateModel.clusterId"
          :selected-cluster-id="templateModel.clusterId"
          :back-tooltip="isEditMode ? '' : 'Change scope'"
          :init-closed="isValidScopeSelected"
          @on-back="onBack"
          @changed-scope="updateScope"
        />
        <q-linear-progress animation-speed="100" indeterminate v-if="scopeChangeInProgress" />
        <transition-group name="fade">
          <template v-if="isValidScopeSelected && !scopeChangeInProgress">
            <workload-template-name-section
              aid="template-name-section"
              support-description
              v-model="templateModel.name"
              :default-opened="!!$route.query.fromCopyId"
              :scope-model="scopeModel"
              :description="templateModel.description"
              :original-name="originalName"
              @update:description="templateModel.description = $event"
            />

            <hybrid-environment-section
              support-linking
              :key="`he-${selectedTemplates.environment.sectionKey}`"
              :render-key="selectedTemplates.environment.renderKey.toString()"
              :selected-name="selectedEnvironmentName"
              :workload-spec="templateModel.spec"
              :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
              :cluster-id="templateModel.clusterId"
              :support-config-map-as-env-var="supportConfigMapAsEnvVar"
              :is-linked="!!selectedTemplates.environment.isLinked"
              :validate-empty-env-vars="false"
              support-authenticated-ngc
              :scope-id="templateModel.scopeId || ''"
              :scope-type="templateModel.scopeType"
              @load-section-templates="displayEnvironmentsList"
              @update="updateModelSpec"
              @reset="resetEnvironmentToDefault"
              @clear="clearEnvironmentTemplate"
              @toggle-linked="toggleEnvironmentLinked"
            />

            <hybrid-compute-resource-section
              support-linking
              :key="`hc-${selectedTemplates['compute-resource'].sectionKey}`"
              :render-key="selectedTemplates['compute-resource'].renderKey.toString()"
              :selected-name="selectedComputeName"
              :workload-spec="templateModel.spec"
              :support-multi-fraction-gpu="supportMultiFractionGpu"
              :cluster-id="templateModel.clusterId"
              :support-tolerations="supportTolerations"
              :hide-specific-sections="{
                nodes: true,
                affinity: true,
              }"
              :is-linked="!!selectedTemplates['compute-resource'].isLinked"
              @load-section-templates="displayComputesList"
              @update="templateModel.spec = $event"
              @reset="resetComputeToDefault"
              @clear="clearComputeTemplate"
              @toggle-linked="toggleComputeLinked"
            />

            <hybrid-data-source-section
              :default-opened="shouldOpenDataSourceSection"
              :allowed-data-source-types="allowedDatasourceOptions"
              :workload-storage="templateModel.spec.storage"
              :cluster-id="templateModel.clusterId"
              @load-section-templates="displayDataSourcesList"
              @update="updateModelStorage"
              :secrets="secrets"
            />

            <general-section
              should-auto-validate
              :show-runs-and-parallelism="supportRunsAndParallelism"
              :workload-form-type="EWorkloadFormType.Training"
              :general-model="generalSectionModel"
              :show-backoff-limit="supportBackoffLimit"
              :show-termination-grace-period="isClusterSupportTerminationGracePeriod"
              :workload-priority-data="workloadPriorityData"
              :support-preemptibility="supportPreemptibility"
              @general-model-changed="updateGeneralModel"
            />
          </template>
        </transition-group>
      </template>
    </workload-form-wrapper>

    <workload-creation-drawer
      :is-open="isDrawerOpen"
      :comply-to="baseDrawerFilters"
      :entity="drawerDataEntity"
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
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
// Stores & Services
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { useDataSourceStore } from "@/stores/data-source.store";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useAuthStore } from "@/stores/auth.store";
import { useAppStore } from "@/stores/app.store";
import { useTrainingStore } from "@/stores/training.store";
import { useClusterStore } from "@/stores/cluster.store";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { alertUtil } from "@/utils/alert.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { deepCopy } from "@/utils/common.util";
import { policyUtil } from "@/utils/policy.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { scopeUtil } from "@/utils/scope.util";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { useRoute } from "vue-router";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
// Models
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import { type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { EWorkloadFormType, EWorkloadName } from "@/models/workload.model";
import {
  MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
} from "@/common/version.constant";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import type { IUIStorageFields, IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import { ScopeType } from "@/swagger-models/workloads-client";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import type { IScopeModel } from "@/models/global.model";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action } from "@/swagger-models/authorization-client";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { WorkloadTemplateNameSection } from "@/components/workload-template/workload-template-name-section";

const TemplateTrainingHybrid = defineComponent({
  name: "template-training-hybrid",
  components: {
    WorkloadCreationDrawer,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    WorkloadFormWrapper,
    HybridDataSourceSection,
    GeneralSection,
    WorkloadTemplateNameSection,
    RunaiScopeSection,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      trainingStore: useTrainingStore(),
      clusterStore: useClusterStore(),
      dataSourceStore: useDataSourceStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),

      EWorkloadFormType,
      submitting: false as boolean,
      isEditMode: false as boolean,
      originalName: "" as string,

      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,
      scopeChangeInProgress: false as boolean,
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
      toggleEnvironmentLinked,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      // computed
      linkedEntitiesIds,
      baseDrawerFilters,
      selectedEnvironmentName,
      selectedComputeName,
    } = useHybridTemplates(EHybridFormTypes.Training, undefined, true);

    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariables } =
      useEnvironmentVariablesSection();

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route.query.clusterId?.toString());

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariables,

      // state
      isDrawerOpen,
      drawerDataEntity,
      templateModel: hybridWorkloadModel,
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
      toggleEnvironmentLinked,
      toggleComputeLinked,
      setupLinkedEntities,
      validateDrawerEntitiesForScopeChange,
      // computed
      linkedEntitiesIds,
      baseDrawerFilters,
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
    const { fromCopyId, clusterId, kind, createdEntityId, name, scope, scopeId, description } = this.$route.query;
    const { id: editId } = this.$route.params;

    if (editId) this.isEditMode = true;

    this.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());

    const formState = this.trainingStore.getFlatFormState();
    if (formState) this.loadFormState();
    else
      await this.initTemplateModel(
        clusterId?.toString(),
        fromCopyId?.toString() || editId?.toString(),
        scope as ScopeType,
        scopeId?.toString(),
        name?.toString(),
        description?.toString(),
      );

    await this.loadScopeBasedData();
    this.validateScopeBasedData();

    if (this.supportPriorityClass) {
      const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Training);
      if (defaultPriority && !this.templateModel.spec.priorityClass) {
        this.templateModel.spec.priorityClass = defaultPriority;
      }

      if (this.supportPreemptibility) {
        const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Training);
        if (!this.templateModel.spec.preemptibility) {
          this.templateModel.spec.preemptibility = defaultPreemptibility;
        }
      }
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    allowedScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    scopeModel(): IScopeModel {
      return scopeUtil.getScopeModelFromTypeAndId(
        this.templateModel.scopeType as ScopeType,
        this.templateModel.scopeId as string,
      );
    },
    isValidScopeSelected(): boolean {
      return this.scopeModel.scope !== ScopeType.System && this.scopeModel.scope !== ScopeType.Tenant;
    },
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
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.templateModel.spec?.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.templateModel.spec?.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: (this.templateModel.spec?.annotations as IItemizedListItem[]) || undefined,
        labels: (this.templateModel.spec?.labels as IItemizedListItem[]) || undefined,
        backoffLimit: Number.isInteger(this.templateModel.spec?.backoffLimit)
          ? this.templateModel.spec?.backoffLimit
          : null,
        terminationGracePeriodSeconds: this.templateModel.spec?.terminationGracePeriodSeconds || null,
        completions: this.templateModel.spec?.completions || null,
        parallelism: this.templateModel.spec?.parallelism || null,
        priorityClass: this.templateModel.spec?.priorityClass || null,
        preemptibility: this.templateModel.spec?.preemptibility || null,
      };
    },
    workloadPriorityData(): IWorkloadPriorityData {
      const totalGpuRequest = workloadHybridUtil.calculateTotalGpuRequest({
        specCompute: this.templateModel.spec?.compute,
      });

      return {
        currProject: this.selectedProject,
        selectedNodePoolsResources: this.templateModel.spec?.nodePools || undefined,
        totalGpuRequest: totalGpuRequest,
        workloadPriorities: this.workloadPriorities,
        supportPriorityClass: this.supportPriorityClass,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(this.templateModel.clusterId || "");
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportRunsAndParallelism(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.templateModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },
  },
  methods: {
    updateModelSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.templateModel.spec = spec;
    },
    async updateScope(scope: IScopeModel): Promise<void> {
      this.templateModel.scopeType = scope.scope as ScopeType;
      this.templateModel.scopeId = scopeUtil.getScopeId(scope);
      this.templateModel.clusterId = scope.clusterId || scope.parentClusterUuid || "";

      try {
        this.scopeChangeInProgress = true;
        await this.loadScopeBasedData();
        await this.validateDrawerEntitiesForScopeChange(this.secrets.map((secret) => secret.name));
        this.validateScopeBasedData();
      } catch (error: unknown) {
        console.error("Failed to update scope:", error);
      } finally {
        this.scopeChangeInProgress = false;
      }
    },
    updateModelStorage(storage: IUIStorageFields): void {
      this.templateModel.spec.storage = storage;
    },
    saveFormState(): void {
      const formState: THybridFormState = {
        projectInfo: {},
        selectedTemplates: this.selectedTemplates,
        projectNodePoolsResources: [],
        policy: undefined,
        policyViewDetails: {} as IPolicyViewDetails,
        selectedProjectNodeTypes: [],
        workloadModel: this.templateModel,
      };
      this.trainingStore.setFlatFormState(deepCopy(formState));
    },
    loadFormState(): void {
      const formState = this.trainingStore.getFlatFormState();
      if (!formState) return;
      this.selectedTemplates = formState?.selectedTemplates;
      this.templateModel = formState?.workloadModel;
      if (this.isEditMode) {
        this.originalName = this.templateModel.name;
        this.appStore.setPageTitleOverride(`Edit training template: ${this.templateModel.name}`);
      }
    },
    clearFormState(): void {
      this.trainingStore.clearFlatFormState();
    },
    async loadSecrets(): Promise<void> {
      if (!this.isValidScopeSelected) {
        this.secrets = [];
        return;
      }

      try {
        this.secrets = await credentialService.listExistingSecrets(
          this.templateModel.scopeType as string,
          this.templateModel.scopeId as string,
          undefined,
          "all",
          true,
        );
      } catch (error: unknown) {
        console.error("Failed to load secrets:\n", error);
        this.$q.notify(alertUtil.getError("Failed to load secrets"));
      }
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
          const workloadTemplate = await workloadTemplateService.getTrainingTemplateById(fromCopyId);

          if (workloadTemplate.combinedSpec) workloadTemplate.spec = workloadTemplate.combinedSpec;

          const uiModel = workloadHybridUtil.getHybridTrainingTemplateModel(workloadTemplate);

          this.templateModel = uiModel;

          if (workloadTemplate.assets) await this.setupLinkedEntities(workloadTemplate.assets);

          if (this.isEditMode) {
            this.originalName = uiModel.name;
            this.appStore.setPageTitleOverride(`Edit training template: ${uiModel.name}`);
          }
        } else {
          const clusterVersion = this.clusterStore.getClusterVersion(clusterId || "");
          this.templateModel = workloadHybridUtil.setTrainingUIDefaults(
            workloadHybridUtil.getEmptyUISuperWorkloadModel(),
            this.authStore.isSSO,
            undefined,
            clusterVersion,
          );
          this.templateModel.scopeType = scopeType;
          this.templateModel.scopeId = scopeId;
          this.templateModel.name = name || "";
          this.templateModel.clusterId = clusterId || "";
          this.templateModel.description = description || "";
        }

        this.createSectionsClearedStates();
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load template model"));
        console.error("Failed to get load template by id:\n", e);
        this.appStore.setFallback(true);
      }
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;
        let createdTemplateId: string | undefined;

        if (this.isEditMode) {
          const request = workloadHybridUtil.prepareTrainingTemplateFlatRequestToUpdate(
            this.templateModel,
            this.linkedEntitiesIds,
          );
          const updatedTemplate = await workloadTemplateService.updateTrainingTemplateFlat(
            this.$route.params.id as string,
            request,
          );
          this.$q.notify(alertUtil.getSuccess(`Template ${updatedTemplate.meta.name} updated`));
        } else {
          this.templateModel.spec = policyUtil.cleanupFlatItemizeBeforeSubmission(this.templateModel.spec);
          const request = workloadHybridUtil.prepareTrainingTemplateFlatRequest(
            this.templateModel,
            this.linkedEntitiesIds,
          );
          const createdTemplate = await workloadTemplateService.createTrainingTemplateFlat(request);
          createdTemplateId = createdTemplate.meta.id;
          this.$q.notify(alertUtil.getSuccess(`Template ${createdTemplate.meta.name} created`));
        }
        this.redirectToPrevRoute(true, createdTemplateId);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("template"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error("Failed to create template:\n", error);
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
    redirectToPrevRoute(fromSubmit?: boolean, createdTemplateId?: string): void {
      this.clearFormState();

      const query = fromSubmit
        ? { clusterId: this.templateModel.clusterId, ...(createdTemplateId && { createdEntityId: createdTemplateId }) }
        : undefined;
      this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX, query });
    },
    onBack(): void {
      this.clearFormState();

      this.$router.push({
        name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW,
        state: {
          pageData: workloadHybridUtil.getFirstPageNavigationTemproraryData(this.templateModel),
        },
      });
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.templateModel = {
        ...this.templateModel,
        spec: {
          ...this.templateModel.spec,
          annotations: generalModel.annotations,
          labels: generalModel.labels,
          backoffLimit: generalModel.backoffLimit,
          terminationGracePeriodSeconds: generalModel.terminationGracePeriodSeconds,
          autoDeletionTimeAfterCompletionSeconds: generalModel.autoDeletionTimeAfterCompletionSeconds,
          parallelism: generalModel.parallelism,
          completions: generalModel.completions,
          priorityClass: generalModel.priorityClass,
          preemptibility: generalModel.preemptibility,
        },
      };
    },
    onCreateNew(type: string): void {
      this.saveFormState();

      this.onCreateNewSetup(type, { workloadType: EWorkloadFormType.Training });
    },
    async loadAvailableDataSource(): Promise<void> {
      this.allowedDatasourceOptions = await this.dataSourceStore.getAvailableDataSources();
      if (this.dataSourceStore.fallbackPolicyLoadingError) this.appStore.setFallback(true);
    },
    async loadProjectData(): Promise<void> {
      if (this.scopeModel.projectId) {
        try {
          this.selectedProject = await orgUnitService.getProject(this.scopeModel.projectId.toString());
        } catch (error: unknown) {
          console.error("Failed to load project data:", error);
          this.appStore.setFallback(true);
        }
      }
    },
    async loadScopeBasedData(): Promise<void> {
      try {
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
          this.loadWorkloadPriorities(this.templateModel.clusterId || ""),
          this.loadProjectData(),
        ]);
      } catch (error: unknown) {
        console.error("Failed to load scope based data:", error);
      }
    },
    validateScopeBasedData(): void {
      this.validateEnvironmentVariables(this.templateModel.spec.environmentVariables || []);
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});

//Dima todo - Type assertion to prevent TS7056 while maintaining type safety -- will be refactored later --hide the annoying warning
export default TemplateTrainingHybrid as typeof TemplateTrainingHybrid & DefineComponent;
</script>
