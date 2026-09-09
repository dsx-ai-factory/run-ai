<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-distributed column items-center" v-if="isPageReady">
    <notification-alert-bar
      v-if="showPolicyMismatchAlert"
      class="full-width"
      :variant="ENotificationBarVariant.Info"
      :message="policyMismatchMessage"
      text-align="center"
      aid="policy-mismatch-alert"
    />
    <runai-wizard-layout
      disable-navigation
      transparent-steps-panel
      :steps="wizardLayout.steps.value"
      :active-step="wizardLayout.activeStep.value"
      @step-click="onStepClicked"
    >
      <template #default>
        <template v-if="wizardLayout.activeStep.value === EDistributedInferenceType.Leader">
          <inference-distributed-leader-form
            :hybrid-context="leaderContext"
            :navigation="navigationData"
            :ui-state="uiStateData"
            :features="featuresSupportedInReplicas"
            :cluster-config="clusterConfigData"
            :policy="policyData"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :data-sources="dataSourcesData"
            @leave-page="redirectToPrevRoute"
            @canceled="onCancel"
            @back="backToNewInference"
            @submit="continueToWorkers"
            @update:spec="updateLeaderSpec"
            @save-state="saveFormState"
          />
        </template>
        <template v-else-if="wizardLayout.activeStep.value === EDistributedInferenceType.Workers">
          <inference-distributed-worker-form
            :hybrid-context="workerContext"
            :ui-state="uiStateData"
            :features="featuresSupportedInReplicas"
            :cluster-config="clusterConfigData"
            :policy="policyData"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :data-sources="dataSourcesData"
            @leave-page="redirectToPrevRoute"
            @canceled="onCancel"
            @back="onBackToLeader"
            @submit="continueToWorkloadSetup"
            @update:spec="updateWorkerSpec"
            @save-state="saveFormState"
          />
        </template>
        <template v-else-if="wizardLayout.activeStep.value === EDistributedInferenceType.WorkloadSetup">
          <inference-distributed-workload-setup
            :model="workloadSetupModel"
            :ui-state="workloadSetupUiState"
            :features="featuresSupportedInReplicas"
            :cluster-config="clusterConfigData"
            :policy="policyData"
            :selected-project="selectedProject"
            @back="onBackToWorkers"
            @submit="createInference"
            @update:model="onWorkloadSetupModelChanged"
            @close-drawer="leaderContext.closeDrawer"
          />
        </template>
      </template>
    </runai-wizard-layout>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useRoute } from "vue-router";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore, type IInferenceDistributedState } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkloadStore } from "@/stores/workload.store";

// cmps
import { RunaiWizardLayout } from "@/components/common/runai-wizard-layout";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import InferenceDistributedLeaderForm from "./inference-distributed-leader-form.vue";
import InferenceDistributedWorkerForm from "./inference-distributed-worker-form.vue";
import InferenceDistributedWorkloadSetup, {
  type IWorkloadSetupUIState,
} from "./inference-distributed-workload-setup.vue";
import type { IWorkloadSetupModel } from "./workload-setup-sections";

// models
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import {
  distributedInferenceSteps,
  EDistributedInferenceType,
  type TDistributedInferenceFormMetaInfo,
} from "./inference-distributed.model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { IAutoScaleOptions } from "@/components/section/compute-resource-section/auto-scale-section/auto-scale-section.model";
import type {
  DistributedInference,
  StringOption,
  UnifiedPolicyInfo,
  UnifiedProjectInfo,
  UnifiedWorkload,
  UnifiedWorkloadRequest,
  Workload as WorkloadFlat,
} from "@/swagger-models/workloads-client";
import { type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { Scope } from "@/swagger-models/assets-service-client";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { EWorkloadName, EWorkloadType, type IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type {
  IInferenceSubmitFormNavigation,
  IInferenceSubmitFormUIState,
  IInferenceSubmitFormFeatures,
  IDistributedInferenceSubmitFormPolicy,
  IInferenceSubmitFormClusterConfig,
  IInferenceSubmitFormDataSources,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";

// utils
import { alertUtil } from "@/utils/alert.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { deepCopy } from "@/utils/common.util";
import { distributedInferenceUtil } from "@/utils/distributed-inference.util";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";

// models
import type { THybridFormState } from "@/models/workload-hybrid.model";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_REQUEST_TIMEOUT,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
} from "@/common/version.constant";
import { EServerType } from "@/models/inference.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// composables
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { policyUtil } from "@/utils/policy.util/policy.util";
import { getInferenceDistributedPolicyViewDetails } from "@/utils/policy.util/policy-view.util";

export default defineComponent({
  name: "inference-distributed",
  components: {
    RunaiWizardLayout,
    NotificationAlertBar,
    InferenceDistributedLeaderForm,
    InferenceDistributedWorkerForm,
    InferenceDistributedWorkloadSetup,
  },
  data() {
    return {
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      workloadStore: useWorkloadStore(),
      EDistributedInferenceType,
      ENotificationBarVariant,
      autoScaleOptions: {
        latencyEnabled: false,
        isAutoScaleEnabled: false,
      } as IAutoScaleOptions,
      clusterName: "" as string,
      showRequestTimeout: false as boolean,
      projectInfo: {} as UnifiedProjectInfo,
      policy: {} as UnifiedPolicyInfo | null | undefined,
      policyViewDetails: {} as IPolicyViewDetails,
      projectNodePoolsResources: [] as StringOption[],
      selectedProjectNodeTypes: [] as string[],
      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,
      supportImagePullSecrets: true as boolean,
      submitting: false as boolean,
      hasPolicyMismatch: false as boolean,
      isCopyFlow: false as boolean,
    };
  },
  setup() {
    const route = useRoute();
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions } = useEnvironmentVariablesSection();

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route.query.clusterId?.toString());

    const leaderContext = useHybridTemplates(EHybridFormTypes.DistributedInferenceLeader);
    const workerContext = useHybridTemplates(EHybridFormTypes.DistributedInferenceWorker);

    const wizardLayout = useWizardLayout(distributedInferenceSteps, EDistributedInferenceType.Leader);

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,

      // context
      leaderContext,
      workerContext,
      wizardLayout,
      // workload priorities composable
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
    };
  },
  async created() {
    const { fromCopyId, projectId, clusterId, inferenceName, workloadFormName, serverType } = this.$route.query;
    const selectedServerType = this.parseServerType(serverType);

    try {
      const distributedState = this.inferenceStore.getDistributedState();

      if (distributedState) {
        this.loadFormState(distributedState);
        if (!this.validateAndInitializeCluster(this.leaderContext.hybridWorkloadModel.value.clusterId)) return;
      } else {
        if (!this.validateAndInitializeCluster(clusterId?.toString())) return;
        await this.initModel(
          fromCopyId?.toString(),
          projectId?.toString(),
          clusterId?.toString(),
          inferenceName?.toString(),
          selectedServerType,
        );
        this.isCopyFlow = !!fromCopyId;
      }

      await Promise.all([
        this.loadSourceValueOptions(true, this.distributedMetaInfo.projectId),
        this.loadSecrets(),
        this.loadAvailableDataSource(),
        this.loadWorkloadPriorities(this.distributedMetaInfo.clusterId),
        this.loadProjectById(this.distributedMetaInfo.projectId.toString()),
      ]);

      if (this.supportPriorityClass) {
        const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.DistributedInference);
        if (
          defaultPriority &&
          !this.leaderContext.hybridWorkloadModel.value.spec.priorityClass &&
          !this.$route.query.fromCopyId
        ) {
          this.leaderContext.hybridWorkloadModel.value.spec.priorityClass = defaultPriority;
        }

        if (this.supportPreemptibility) {
          const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(
            EWorkloadName.DistributedInference,
          );
          if (!this.leaderContext.hybridWorkloadModel.value.spec.preemptibility && !this.$route.query.fromCopyId) {
            this.leaderContext.hybridWorkloadModel.value.spec.preemptibility = defaultPreemptibility;
          }
        }
      }

      this.supportImagePullSecrets = this.clusterStore.isClusterVersionSufficient(
        this.leaderContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_IMAGE_PULL_SECRETS,
      );

      if (workloadFormName) {
        this.wizardLayout.goToStep(workloadFormName as EDistributedInferenceType);
      }
      this.appStore.setPageLoading(false);
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    }
  },
  computed: {
    distributedMetaInfo(): TDistributedInferenceFormMetaInfo {
      // the leader is the main body of the distributed inference workload so all meta info is taken from it
      return {
        projectId: this.leaderContext.hybridWorkloadModel.value.projectId,
        clusterId: this.leaderContext.hybridWorkloadModel.value.clusterId,
      };
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    showPolicyMismatchAlert(): boolean {
      return this.wizardLayout.activeStep.value === EDistributedInferenceType.Workers && this.hasPolicyMismatch;
    },
    policyMismatchMessage(): string {
      return "All settings in this step have been reset due to a policy your administrator set.";
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.leaderContext.hybridWorkloadModel.value.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },
    navigationData(): IInferenceSubmitFormNavigation {
      return {
        clusterName: this.clusterName,
        projectName: this.selectedProject.name || "-",
        showInferenceType: true,
        typeSummary: "Custom",
        defaultOpenNameSection: !!this.$route.query.fromCopyId,
      };
    },
    uiStateData(): IInferenceSubmitFormUIState {
      return {
        submitting: this.submitting,
        drawerDataEntity: null,
        recentlyCreatedEntityId: this.leaderContext.recentlyCreatedEntityId.value,
      };
    },
    featuresSupportedInReplicas(): IInferenceSubmitFormFeatures {
      return {
        supportImagePullSecrets: this.supportImagePullSecrets,
        isVolumesEnabled: this.clusterStore.isClusterVersionSufficient(
          this.leaderContext.hybridWorkloadModel.value.clusterId || "",
          MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
        ),
        showRequestTimeout: false,
        supportPriorityClass: this.supportPriorityClass,
        supportPreemptibility: this.supportPreemptibility,
        supportSemiPreemptibility: this.clusterStore.isClusterVersionSufficient(
          this.leaderContext.hybridWorkloadModel.value.clusterId || "",
          MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
        ),
      };
    },
    policyData(): IDistributedInferenceSubmitFormPolicy | undefined {
      if (!this.policy) return undefined;

      return {
        distributedPolicy: this.policy,
        policyViewDetails: this.policyViewDetails,
      };
    },
    clusterConfigData(): IInferenceSubmitFormClusterConfig {
      return {
        autoScaleOptions: this.autoScaleOptions,
        workloadPriorities: this.workloadPriorities,
        projectNodePoolsResources: this.projectNodePoolsResources,
      };
    },
    dataSourcesData(): IInferenceSubmitFormDataSources {
      return {
        secrets: this.secrets,
        allowedDataSourceOptions: this.allowedDatasourceOptions || new Set(),
      };
    },
    workloadSetupModel(): IWorkloadSetupModel {
      const leaderContextSpec = this.leaderContext.hybridWorkloadModel.value.spec;
      const workerContextSpec = this.workerContext.hybridWorkloadModel.value.spec;

      return {
        leaderCompute: leaderContextSpec.leader?.compute ?? undefined,
        priorityClass: leaderContextSpec.priorityClass || null,
        preemptibility: leaderContextSpec.preemptibility || null,
        startupPolicy: leaderContextSpec.startupPolicy || null,
        nodePools: leaderContextSpec.nodePools ?? undefined,
        replicas: leaderContextSpec.replicas || 1,
        workers: workerContextSpec.workers || 1,
        workerCompute: workerContextSpec.worker?.compute ?? undefined,
      };
    },
    workloadSetupUiState(): IWorkloadSetupUIState {
      return {
        submitting: this.submitting,
        isDrawerOpen: this.leaderContext.isDrawerOpen.value,
      };
    },
  },
  methods: {
    parseServerType(value: unknown): EServerType | undefined {
      const candidate = Array.isArray(value) ? value[0] : value;
      return typeof candidate === "string" && Object.values(EServerType).includes(candidate as EServerType)
        ? (candidate as EServerType)
        : undefined;
    },
    onStepClicked(payload: { step: IRunaiWizardLayoutStep; name: string }): void {
      this.wizardLayout.goToStep(payload.name);
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
    getClusterById(clusterId: string): ClustersMinimalData | undefined {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === clusterId);
    },
    redirectToNewInference(): void {
      this.$router.replace({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
    },
    setupClusterConfiguration(cluster: ClustersMinimalData): void {
      this.clusterName = cluster.name;
    },
    async getUnifiedHybridWorkload(options?: UnifiedWorkloadRequest): Promise<UnifiedWorkload> {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.DistributedInference,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
    async initModel(
      workloadId?: string,
      projectId?: string,
      clusterId?: string,
      workloadName?: string,
      serverType?: EServerType,
    ): Promise<void> {
      try {
        const unifiedWorkload: UnifiedWorkload = await this.getUnifiedHybridWorkload({
          workloadId,
          projectId,
          clusterId,
        });

        this.policy = unifiedWorkload.policyInfo;
        this.policyViewDetails = getInferenceDistributedPolicyViewDetails(this.policy);
        this.projectInfo = unifiedWorkload.projectInfo || {};

        // Load node pools from policy
        this.projectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];

        const uiModel = inferenceWorkloadUtil.getDistributedInferenceModel(unifiedWorkload);
        this.leaderContext.hybridWorkloadModel.value = inferenceWorkloadUtil.setDistributedUIDefaults(
          uiModel,
          this.authStore.isSSO,
          this.policy,
          serverType,
        );

        if (projectId) this.leaderContext.hybridWorkloadModel.value.projectId = projectId;
        if (clusterId) this.leaderContext.hybridWorkloadModel.value.clusterId = clusterId;
        if (workloadName) this.leaderContext.hybridWorkloadModel.value.name = workloadName;
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load inference model"));
        console.error("Failed to get unified inference:\n", e);
        this.appStore.setFallback(true);
      }
    },
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
          this.distributedMetaInfo.projectId,
          undefined,
          "all",
          true,
        );
      } catch (error: unknown) {
        console.error("Failed to load secrets:\n", error);
        this.$q.notify(alertUtil.getError("Failed to load secrets"));
      }
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
    redirectToPrevRoute(fromSubmit?: boolean): void {
      this.clearFormState();
      const query = fromSubmit ? { clusterId: this.distributedMetaInfo.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    clearFormState(): void {
      this.inferenceStore.clearDistributedFormState();
      this.inferenceStore.clearInferenceFormUIState();
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    backToNewInference(): void {
      this.clearFormState();
      this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
    },
    syncWorkerContextFromLeader(): void {
      // Policy handling logic:
      // 1. No policy or no worker policy => copy leader data into worker
      // 2. Policy is equal between leader and worker => copy leader data into worker
      // 3. Policy is NOT equal => do not change worker data, show mismatch alert
      // 4. Copy flow (from copy workload) => do not overwrite worker with leader; keep copied worker spec
      const shouldCopyLeaderToWorker =
        !this.isCopyFlow && distributedInferenceUtil.shouldCopyLeaderToWorker(this.policy);

      // Always copy basic identifying info and shared fields to worker context
      this.workerContext.hybridWorkloadModel.value.clusterId = this.leaderContext.hybridWorkloadModel.value.clusterId;
      this.workerContext.hybridWorkloadModel.value.projectId = this.leaderContext.hybridWorkloadModel.value.projectId;
      this.workerContext.hybridWorkloadModel.value.name = this.leaderContext.hybridWorkloadModel.value.name;
      // Node pools are shared and can only be modified in the leader step
      this.workerContext.hybridWorkloadModel.value.spec.nodePools =
        this.leaderContext.hybridWorkloadModel.value.spec.nodePools;

      if (shouldCopyLeaderToWorker) {
        this.hasPolicyMismatch = false;
        this.workerContext.hybridWorkloadModel.value.spec.worker = deepCopy(
          this.leaderContext.hybridWorkloadModel.value.spec.leader,
        );
      } else {
        // Policies are different - don't copy leader data, show mismatch alert
        this.hasPolicyMismatch = true;
      }
    },
    async continueToWorkers(): Promise<void> {
      this.workerContext.hybridWorkloadModel.value = deepCopy(this.leaderContext.hybridWorkloadModel.value);
      this.syncWorkerContextFromLeader();

      this.saveFormState();
      this.wizardLayout.markStepComplete(EDistributedInferenceType.Leader);
      this.wizardLayout.nextStep();
    },
    continueToWorkloadSetup(): void {
      this.wizardLayout.markStepComplete(EDistributedInferenceType.Workers);
      this.wizardLayout.nextStep();
    },
    onBackToLeader(): void {
      this.wizardLayout.markStepIncomplete(EDistributedInferenceType.Leader);
      this.wizardLayout.previousStep();
    },
    onBackToWorkers(): void {
      this.wizardLayout.markStepIncomplete(EDistributedInferenceType.Workers);
      this.wizardLayout.previousStep();
    },
    updateLeaderSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.leaderContext.hybridWorkloadModel.value.spec = spec;
    },
    onWorkloadSetupModelChanged(model: IWorkloadSetupModel): void {
      const spec = this.leaderContext.hybridWorkloadModel.value.spec;
      spec.priorityClass = model.priorityClass || undefined;
      spec.preemptibility = model.preemptibility || undefined;
      spec.startupPolicy = model.startupPolicy || undefined;
      spec.replicas = model.replicas;
    },
    updateWorkerSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.workerContext.hybridWorkloadModel.value.spec = spec;
    },
    saveFormState(): void {
      const leaderState = this.getLeaderState();
      const workerState = this.getWorkerState();
      this.inferenceStore.saveDistributedFormState({
        leader: deepCopy(leaderState),
        worker: deepCopy(workerState),
        isCopyFlow: this.isCopyFlow,
      });
    },
    getLeaderState(): THybridFormState {
      return {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.projectNodePoolsResources,
        selectedTemplates: this.leaderContext.selectedTemplates.value,
        policy: null,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.leaderContext.hybridWorkloadModel.value,
      };
    },
    getWorkerState(): THybridFormState {
      return {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.projectNodePoolsResources,
        selectedTemplates: this.workerContext.selectedTemplates.value,
        policy: null,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.workerContext.hybridWorkloadModel.value,
      };
    },
    loadFormState(distributedState: IInferenceDistributedState): void {
      this.leaderContext.hybridWorkloadModel.value = distributedState.leader?.workloadModel;
      if (distributedState.worker) {
        this.workerContext.hybridWorkloadModel.value = distributedState.worker?.workloadModel;
      }
      this.isCopyFlow = distributedState.isCopyFlow ?? false;

      this.projectInfo = distributedState.leader?.projectInfo;
      this.policyViewDetails = distributedState.leader?.policyViewDetails;
      this.selectedProjectNodeTypes = distributedState.leader?.selectedProjectNodeTypes;
      this.projectNodePoolsResources = distributedState.leader?.projectNodePoolsResources;
    },
    async createInference(): Promise<void> {
      try {
        this.submitting = true;
        const distributedInferenceModel = deepCopy(this.leaderContext.hybridWorkloadModel.value);
        distributedInferenceModel.spec.worker = deepCopy(this.workerContext.hybridWorkloadModel.value.spec.worker);

        distributedInferenceModel.spec.leader = policyUtil.cleanupFlatItemizeBeforeSubmission(
          distributedInferenceModel.spec.leader || {},
          this.policy?.worker,
        );
        distributedInferenceModel.spec.worker = policyUtil.cleanupFlatItemizeBeforeSubmission(
          distributedInferenceModel.spec.worker || {},
          this.policy?.worker,
        );

        const workloadCreationRequest = inferenceWorkloadUtil.prepareDistributedInferenceFlatRequest(
          distributedInferenceModel,
          this.workerContext.hybridWorkloadModel.value.spec.workers || 1,
        );

        const distributedInferenceWorkload: DistributedInference =
          await inferenceService.createDistributedInference(workloadCreationRequest);
        this.clearFormState();

        const workload: WorkloadFlat = inferenceWorkloadUtil.convertDistributedInferenceToWorkload(
          distributedInferenceWorkload,
          this.authStore.tenant.id,
        );

        this.workloadStore.setWorkloadAdded(workload, EWorkloadType.DistributedInference);
        this.clusterStore.setSelectedClusterById(distributedInferenceWorkload.clusterId);
        this.$q.notify(alertUtil.getSuccess(`Inference ${distributedInferenceWorkload.name} created`));
        this.$router.push({
          name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
          query: { clusterId: distributedInferenceWorkload.clusterId },
        });
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
  },
});
</script>
