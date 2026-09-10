<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="distributed-hybrid column items-center" v-if="!appStore.isPageLoading">
    <runai-wizard-layout
      disable-navigation
      transparent-steps-panel
      :steps="wizardLayout.steps.value"
      :active-step="wizardLayout.activeStep.value"
      @step-click="onStepClicked"
    >
      <template #default>
        <template v-if="wizardLayout.activeStep.value === EDistributedReplicaType.Worker">
          <distributed-hybrid-worker
            :hybrid-context="workerContext"
            is-template
            :secrets="secrets"
            :submitting="submitting"
            :project-data="projectData"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :allowed-data-sources="allowedDatasourceOptions"
            :workload-priority-data="workloadPriorityData"
            :hide-specific-sections="hideSpecificSections"
            :is-edit-mode="isEditMode"
            :support-image-pull-secrets="supportImagePullSecrets"
            :support-termination-grace-period="isClusterSupportTerminationGracePeriod"
            :support-preemptibility="supportPreemptibility"
            :is-valid-scope-selected="isValidScopeSelected"
            :loading-scope-based-data="scopeChangeInProgress"
            @update="updateWorkerSpec"
            @continue="handleNextStep"
            @save-form-state="saveFormState"
            @leave-page="redirectToPrevRoute"
            @back-to-create-new="backToTemplateNew"
          >
            <template #scope-section>
              <runai-scope-section
                :read-only="isEditMode"
                kind-of-asset="template"
                :scope-model="scopeModel"
                :allowed-scopes="allowedScopes"
                :key="distributedMetaInfo.clusterId"
                :selected-cluster-id="distributedMetaInfo.clusterId"
                :back-tooltip="isEditMode ? '' : 'Change scope'"
                :init-closed="isValidScopeSelected"
                @on-back="backToTemplateNew"
                @changed-scope="updateScope"
              />
            </template>
            <template #name-section>
              <workload-template-name-section
                aid="template-name-section"
                support-description
                v-model="workerContext.hybridWorkloadModel.value.name"
                :default-opened="!!$route.query.fromCopyId"
                :scope-model="scopeModel"
                :description="workerContext.hybridWorkloadModel.value.description"
                :original-name="originalName"
                @update:description="workerContext.hybridWorkloadModel.value.description = $event"
              />
            </template>
          </distributed-hybrid-worker>
        </template>
        <template v-else-if="wizardLayout.activeStep.value === EDistributedReplicaType.Master">
          <distributed-hybrid-master
            is-template
            :hybrid-context="masterContext"
            :secrets="secrets"
            :form-meta-data="masterMetaData"
            :project-data="projectData"
            :submitting="submitting"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :allowed-data-sources="allowedDatasourceOptions"
            :workload-priority-data="workloadPriorityData"
            :hide-specific-sections="hideSpecificSections"
            :is-edit-mode="isEditMode"
            :support-image-pull-secrets="supportImagePullSecrets"
            :support-termination-grace-period="isClusterSupportTerminationGracePeriod"
            :support-preemptibility="supportPreemptibility"
            @back="onBack"
            @continue="onMasterContinue"
            @save-form-state="saveFormState"
            @update="updateMasterSpec"
            @leave-page="redirectToPrevRoute"
          />
        </template>

        <template v-else-if="wizardLayout.activeStep.value === EDistributedHybridStepName.Setup">
          <distributed-hybrid-setup
            is-template
            :hybrid-context="workerContext"
            :submitting="submitting"
            :support-priority-class="supportPriorityClass"
            :workload-priority-data="workloadPriorityData"
            :support-preemptibility="supportPreemptibility"
            :support-termination-grace-period="isClusterSupportTerminationGracePeriod"
            @update="updateWorkerSpec"
            @back="onSetupBack"
            @submit="onSubmit"
            @leave-page="redirectToPrevRoute"
          />
        </template>
      </template>
    </runai-wizard-layout>
  </section>
</template>

<script lang="ts">
import { defineComponent, type DefineComponent } from "vue";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { RunaiWizardLayout } from "@/components/common/runai-wizard-layout";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import { EDistributedReplicaType, EWorkloadFormType } from "@/models/workload.model";
import { useAppStore } from "@/stores/app.store";
import type { TDistributedFormMetaInfo } from "@/views/training/distributed-hybrid/distributed-hybrid.model";
import DistributedHybridMaster from "@/views/training/distributed-hybrid/distributed-hybrid-master.vue";
import DistributedHybridWorker from "@/views/training/distributed-hybrid/distributed-hybrid-worker.vue";
import DistributedHybridSetup from "@/views/training/distributed-hybrid/distributed-hybrid-setup.vue";
import { alertUtil } from "@/utils/alert.util";
import { ScopeType } from "@/swagger-models/workloads-client";
import { CleanPodPolicy, type UnifiedProjectInfo } from "@/swagger-models/workloads-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import type { IHideSpecificSections, THybridFormState } from "@/models/workload-hybrid.model";
import { useClusterStore } from "@/stores/cluster.store";
import { useTrainingStore } from "@/stores/training.store";
import { useAuthStore } from "@/stores/auth.store";
import { DistributedFramework, Scope, type StringOption } from "@/swagger-models/assets-service-client";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { deepCopy } from "@/utils/common.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { TProjectData } from "@/views/training/distributed-hybrid/distributed-hybrid-worker.vue";
import { policyUtil } from "@/utils/policy.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { useDataSourceStore } from "@/stores/data-source.store";
import type { TDataSourceKinds } from "@/models/data-source.model";
import { useWorkloadStore } from "@/stores/workload.store";
import {
  MIN_CLUSTER_VERSION_CLEAN_POD_POLICY,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
} from "@/common/version.constant";
import { scopeUtil } from "@/utils/scope.util";
import type { IScopeModel } from "@/models/global.model";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { Action } from "@/swagger-models/authorization-client";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { useRoute } from "vue-router";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { EWorkloadName } from "@/models/workload.model";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { useImagePullSecrets } from "@/composables/use-image-pull-secrets.composable";
import { MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD } from "@/common/version.constant";
import { WorkloadTemplateNameSection } from "@/components/workload-template/workload-template-name-section";
import {
  distributedHybridSteps,
  EDistributedHybridStepName,
} from "@/views/training/distributed-hybrid/distributed-hybrid.model";

const TemplateDistributedHybrid = defineComponent({
  name: "template-distributed-hybrid",
  components: {
    RunaiWizardLayout,
    DistributedHybridWorker,
    DistributedHybridMaster,
    DistributedHybridSetup,
    RunaiScopeSection,
    WorkloadTemplateNameSection,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      trainingStore: useTrainingStore(),
      clusterStore: useClusterStore(),
      dataSourceStore: useDataSourceStore(),
      workloadStore: useWorkloadStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),

      submitting: false,

      projectInfo: {} as UnifiedProjectInfo,
      selectedProjectNodeTypes: [] as string[],

      workerProjectNodePoolsResources: [] as StringOption[],
      masterProjectNodePoolsResources: [] as StringOption[],

      EDistributedReplicaType,
      EWorkloadFormType,
      EDistributedHybridStepName,

      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,

      isEditMode: false as boolean,
      originalName: "" as string,

      selectedProject: {} as Project,

      hideSpecificSections: {
        computeResource: {
          nodes: true,
          affinity: true,
        },
        environment: {
          imagePullSecrets: true,
        },
      } as IHideSpecificSections,

      scopeChangeInProgress: false as boolean,
    };
  },
  setup() {
    const route = useRoute();
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariables } =
      useEnvironmentVariablesSection();

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route.query.clusterId?.toString());

    const { supportImagePullSecrets } = useImagePullSecrets(route.query.clusterId?.toString());

    const workerContext = useHybridTemplates(EHybridFormTypes.DistributedWorker, undefined, true);
    const masterContext = useHybridTemplates(
      EHybridFormTypes.DistributedMaster,
      {
        name: "",
        projectId: workerContext.hybridWorkloadModel.value.projectId,
        clusterId: workerContext.hybridWorkloadModel.value.clusterId,
        spec: workerContext.hybridWorkloadModel.value.masterSpec || {},
        // template exclusive (not in use for workload form)
        scopeType: workerContext.hybridWorkloadModel.value.scopeType,
        scopeId: workerContext.hybridWorkloadModel.value.scopeId,
      },
      true,
    );

    const wizardLayout = useWizardLayout(distributedHybridSteps(), EDistributedReplicaType.Worker);

    return {
      workerContext,
      masterContext,
      wizardLayout,

      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariables,

      // workload priorities composable
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
      supportImagePullSecrets,
    };
  },
  async created() {
    const { fromCopyId, clusterId, name, scope, scopeId, description, distFramework, noMaster, workloadFormName } =
      this.$route.query;
    const { id: editId } = this.$route.params;

    if (editId) this.isEditMode = true;

    if (fromCopyId || editId) {
      this.trainingStore.setDistFlatFromContext({
        fromCopyId: fromCopyId?.toString() || editId?.toString(),
        templateId: null,
      });
    }

    const formState = this.trainingStore.getDistributedWorkerFlatFormState();

    if (formState) {
      this.workerContext.hybridWorkloadModel.value = formState.workloadModel;
      this.workerContext.selectedTemplates.value = formState.selectedTemplates;
      this.projectInfo = formState.projectInfo;
      this.selectedProjectNodeTypes = formState.selectedProjectNodeTypes;
      this.workerProjectNodePoolsResources = formState.projectNodePoolsResources;

      const masterFormState = this.trainingStore.getDistributedMasterFlatFormState();
      if (masterFormState) {
        this.masterContext.hybridWorkloadModel.value = masterFormState.workloadModel;
        this.masterContext.selectedTemplates.value = masterFormState.selectedTemplates;
        this.masterProjectNodePoolsResources = masterFormState.projectNodePoolsResources;
      }

      if (formState.completedSteps) {
        formState.completedSteps.forEach((stepName) => {
          this.wizardLayout.markStepComplete(stepName);
        });
      }

      if (this.isEditMode) {
        this.originalName = this.workerContext.hybridWorkloadModel.value.name;
        this.appStore.setPageTitleOverride(
          `Edit training template: ${this.workerContext.hybridWorkloadModel.value.name}`,
        );
      }
    } else {
      await this.initDistributedModel(
        clusterId?.toString(),
        fromCopyId?.toString() || editId?.toString(),
        scope as ScopeType,
        scopeId?.toString(),
        name?.toString(),
        description?.toString(),
        noMaster === "true",
        distFramework?.toString(),
      );
    }

    this.initRelevantFormSteps();
    if (workloadFormName) this.wizardLayout.goToStep(workloadFormName as EDistributedReplicaType);

    await this.loadScopeBasedData();
    this.validateScopeBasedData();

    if (this.supportPriorityClass) {
      const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Distributed);
      if (defaultPriority && !this.workerContext.hybridWorkloadModel.value.spec.priorityClass) {
        this.workerContext.hybridWorkloadModel.value.spec.priorityClass = defaultPriority;
      }

      if (this.supportPreemptibility) {
        const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Distributed);
        if (!this.workerContext.hybridWorkloadModel.value.spec.preemptibility) {
          this.workerContext.hybridWorkloadModel.value.spec.preemptibility = defaultPreemptibility;
        }
      }
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    distributedMetaInfo(): TDistributedFormMetaInfo {
      // the worker is the main body of the distributed workload so all meta info is taken from it
      return {
        clusterId: this.workerContext.hybridWorkloadModel.value.clusterId || "",
        scopeType: this.workerContext.hybridWorkloadModel.value.scopeType || undefined,
        scopeId: this.workerContext.hybridWorkloadModel.value.scopeId || undefined,
        projectId: this.workerContext.hybridWorkloadModel.value.projectId,
        framework: this.workerContext.hybridWorkloadModel.value.spec.distributedFramework || undefined,
        sameAsWorker: !!this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker,
      };
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    scopeModel(): IScopeModel {
      return scopeUtil.getScopeModelFromTypeAndId(
        this.distributedMetaInfo.scopeType as ScopeType,
        this.distributedMetaInfo.scopeId as string,
      );
    },
    projectData(): TProjectData {
      return {
        projectName: this.projectInfo.name || "",
        projectNodePoolsResources: this.workerProjectNodePoolsResources,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
      };
    },
    masterMetaData(): TDistributedFormMetaInfo {
      return {
        ...this.distributedMetaInfo,
      };
    },
    allowedScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    workloadPriorityData(): IWorkloadPriorityData {
      const totalGpuRequest = workloadHybridUtil.calculateTotalGpuRequest({
        specCompute: this.workerContext.hybridWorkloadModel.value.spec?.compute,
        masterCompute: this.masterContext.hybridWorkloadModel.value.spec?.compute,
        numWorkers: this.workerContext.hybridWorkloadModel.value.spec?.numWorkers,
      });

      return {
        currProject: this.selectedProject,
        selectedNodePoolsResources: this.workerContext.hybridWorkloadModel.value.spec?.nodePools || undefined,
        totalGpuRequest: totalGpuRequest,
        workloadPriorities: this.workloadPriorities,
        supportPriorityClass: this.supportPriorityClass,
        supportSemiPreemptibility: this.clusterStore.isClusterVersionSufficient(
          this.workerContext.hybridWorkloadModel.value.clusterId || "",
          MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
        ),
      };
    },
    isWorkerSelectedAssetsLinked(): boolean {
      const workerSelectedTemplates = this.workerContext.selectedTemplates.value;
      return (
        !!workerSelectedTemplates?.[WorkloadCreationDrawerEntities.Compute].isLinked ||
        !!workerSelectedTemplates?.environment.isLinked
      );
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workerContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workerContext.hybridWorkloadModel.value.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
    },
    isValidScopeSelected(): boolean {
      return this.scopeModel.scope !== ScopeType.System && this.scopeModel.scope !== ScopeType.Tenant;
    },
  },
  methods: {
    async updateScope(scope: IScopeModel): Promise<void> {
      const scopeId = scopeUtil.getScopeId(scope);

      this.workerContext.hybridWorkloadModel.value.scopeType = scope.scope as ScopeType;
      this.workerContext.hybridWorkloadModel.value.scopeId = scopeId;
      this.workerContext.hybridWorkloadModel.value.clusterId = scope.clusterId || scope.parentClusterUuid || "";

      this.masterContext.hybridWorkloadModel.value.scopeType = scope.scope as ScopeType;
      this.masterContext.hybridWorkloadModel.value.scopeId = scopeId;
      this.masterContext.hybridWorkloadModel.value.clusterId = scope.clusterId || scope.parentClusterUuid || "";

      try {
        this.scopeChangeInProgress = true;
        await this.loadScopeBasedData();
        await this.workerContext.validateDrawerEntitiesForScopeChange(
          this.secrets.map((secret) => secret.name),
          {
            complyToReplicaType: "Worker",
            distributedFramework: this.workerContext.hybridWorkloadModel.value.spec.distributedFramework,
          },
        );
        await this.masterContext.validateDrawerEntitiesForScopeChange(
          this.secrets.map((secret) => secret.name),
          {
            complyToReplicaType: "Master",
            distributedFramework: this.masterContext.hybridWorkloadModel.value.spec.distributedFramework,
          },
        );
        this.validateScopeBasedData();
      } catch (error: unknown) {
        console.error("Failed to update scope:", error);
      } finally {
        this.scopeChangeInProgress = false;
      }
    },
    initRelevantFormSteps(): void {
      if (this.workerContext.hybridWorkloadModel.value.masterSpec) return; // Keep all 3 steps
      // Filter to show only Worker and Setup (skip Master)
      const stepsWithoutMaster = distributedHybridSteps().filter(
        (step) => step.name !== EDistributedHybridStepName.Master,
      );
      this.wizardLayout.setSteps(stepsWithoutMaster);
    },
    onStepClicked(payload: { step: IRunaiWizardLayoutStep; name: string }): void {
      this.wizardLayout.goToStep(payload.name);
    },
    handleNextStep(): void {
      // if master exists even as an empty object it means its not a "no master request" and we need to continue to the master step
      this.workerContext.hybridWorkloadModel.value.masterSpec ? this.onContinue() : this.onContinueToSetup();
    },
    async onContinue(): Promise<void> {
      this.initMasterModel();
      this.wizardLayout.markStepComplete(EDistributedHybridStepName.Worker);
      this.wizardLayout.nextStep();
    },
    onContinueToSetup(): void {
      this.wizardLayout.markStepComplete(EDistributedHybridStepName.Worker);
      this.wizardLayout.goToStep(EDistributedHybridStepName.Setup);
    },
    onMasterContinue(): void {
      this.wizardLayout.markStepComplete(EDistributedHybridStepName.Master);
      this.wizardLayout.nextStep(); // Goes to Setup
    },
    onSetupBack(): void {
      if (this.workerContext.hybridWorkloadModel.value.masterSpec) {
        this.wizardLayout.markStepIncomplete(EDistributedHybridStepName.Master);
      } else {
        this.wizardLayout.markStepIncomplete(EDistributedHybridStepName.Worker);
      }
      this.wizardLayout.previousStep();
    },
    initMasterModel(): void {
      // setting up the meta data as its needed to send unified requests
      const { scopeType, scopeId, name, clusterId, description } = this.workerContext.hybridWorkloadModel.value;
      this.setMasterFormMetaData(scopeType, scopeId, name, clusterId, description);

      // if its a copy and edit and master exists - do nothing we dont want to alter the original data
      if (this.trainingStore.distFlatFromContext.fromCopyId && this.workerContext.hybridWorkloadModel.value.masterSpec) {
        // but if masterSpecSameAsWorker is true, we need to configured master based on the worker otherwise the next step would be empty
        if (!this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker) {
          // see comment below this condition for more details
          this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker = undefined;
          return;
        }
      }

      // whether its a copy and edit or its based on a template
      // at this point the UI takes over and masterSpecSameAsWorker loses its purpose and is set to undefined
      this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker = undefined;
      this.copyWorkerSpecToMaster();
    },
    copyWorkerSpecToMaster(): void {
      const workerSpec = deepCopy(this.workerContext.hybridWorkloadModel.value.spec);
      delete workerSpec.numWorkers;
      delete workerSpec.podAffinity;
      delete workerSpec.distributedFramework;

      this.masterContext.hybridWorkloadModel.value.spec = workerSpec;

      this.matchModelToSameAsWorkerChanges();
    },
    matchModelToSameAsWorkerChanges(): void {
      if (!this.masterMetaData.sameAsWorker) return;

      // if masterSameAsWorker is true, and the worker has selected assets templates,
      // we need to match master selected templates states to be the same as the worker's
      if (this.isWorkerSelectedAssetsLinked) {
        this.matchMasterToWorkerSelectedAssetTemplates();
      } else {
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name = "";
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id = "";
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData = null;
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked = undefined;
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name = "";
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id = "";
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData = null;
        this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked = undefined;
      }

      this.masterContext.reRenderSectionEntirely(WorkloadCreationDrawerEntities.Environment);
      this.masterContext.reRenderSectionEntirely(WorkloadCreationDrawerEntities.Compute);
      this.masterContext.createSectionsClearedStates();
    },
    matchMasterToWorkerSelectedAssetTemplates(): void {
      // policy has no effect on templates therefore master is always initialized based on the worker spec,
      // when worker has chosen to link its selected assets,
      // we need to match master selected templates states to be the same as the worker's
      const workerSelectedTemplates = this.workerContext.selectedTemplates.value;
      if (
        workerSelectedTemplates?.[WorkloadCreationDrawerEntities.Compute].isLinked ||
        workerSelectedTemplates?.environment.isLinked
      )
        this.masterContext.selectedTemplates.value = deepCopy(workerSelectedTemplates);
    },
    updateWorkerSpec(spec: IUISuperWorkloadSpec): void {
      this.workerContext.hybridWorkloadModel.value.spec = spec;

      if (this.masterContext.hybridWorkloadModel.value.spec) {
        this.masterContext.hybridWorkloadModel.value.spec.terminationGracePeriodSeconds =
          spec.terminationGracePeriodSeconds;
        this.masterContext.hybridWorkloadModel.value.spec.autoDeletionTimeAfterCompletionSeconds =
          spec.autoDeletionTimeAfterCompletionSeconds;
        this.masterContext.hybridWorkloadModel.value.spec.backoffLimit = spec.backoffLimit;
        this.masterContext.hybridWorkloadModel.value.spec.cleanPodPolicy = spec.cleanPodPolicy;
        this.masterContext.hybridWorkloadModel.value.spec.sshAuthMountPath = spec.sshAuthMountPath;
        this.masterContext.hybridWorkloadModel.value.spec.mpiLauncherCreationPolicy = spec.mpiLauncherCreationPolicy;
        this.masterContext.hybridWorkloadModel.value.spec.priorityClass = spec.priorityClass;
        this.masterContext.hybridWorkloadModel.value.spec.preemptibility = spec.preemptibility;
      }
    },
    updateMasterSpec(masterSpec: IUISuperWorkloadSpec): void {
      this.masterContext.hybridWorkloadModel.value.spec = masterSpec;
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;

        const linkedAssetsIdsPerReplica = {
          worker: {
            environment: this.workerContext.selectedTemplates.value.environment.isLinked
              ? this.workerContext.selectedTemplates.value.environment.id
              : undefined,
            compute: this.workerContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked
              ? this.workerContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id
              : undefined,
          },
          master: {
            environment: this.masterContext.selectedTemplates.value.environment.isLinked
              ? this.masterContext.selectedTemplates.value.environment.id
              : undefined,
            compute: this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked
              ? this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id
              : undefined,
          },
        };

        const cleanedWorkerSpec = policyUtil.cleanupFlatItemizeBeforeSubmission(
          this.workerContext.hybridWorkloadModel.value.spec,
        );

        let cleanedMasterSpec: IUISuperWorkloadSpec | undefined;

        if (this.workerContext.hybridWorkloadModel.value.masterSpec && this.masterContext.hybridWorkloadModel.value.spec)
          cleanedMasterSpec = policyUtil.cleanupFlatItemizeBeforeSubmission(
            this.masterContext.hybridWorkloadModel.value.spec,
          );

        let createdTemplateId: string | undefined;

        if (this.isEditMode) {
          const request = workloadHybridUtil.prepareDistributedTemplateFlatRequestToUpdate(
            {
              ...this.workerContext.hybridWorkloadModel.value,
              spec: cleanedWorkerSpec,
              masterSpec: cleanedMasterSpec,
            },
            linkedAssetsIdsPerReplica,
          );

          const updatedTemplate = await workloadTemplateService.updateDistributedTemplateFlat(
            this.$route.params.id as string,
            request,
          );
          this.$q.notify(alertUtil.getSuccess(`Template ${updatedTemplate.meta.name} updated`));
        } else {
          const request = workloadHybridUtil.prepareDistributedTemplateFlatRequest(
            {
              ...this.workerContext.hybridWorkloadModel.value,
              spec: cleanedWorkerSpec,
              masterSpec: cleanedMasterSpec,
            },
            linkedAssetsIdsPerReplica,
          );

          const createdTemplate = await workloadTemplateService.createDistributedTemplateFlat(request);
          createdTemplateId = createdTemplate.meta.id;

          this.$q.notify(alertUtil.getSuccess(`Template ${request.meta.name} created`));
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
    redirectToPrevRoute(fromSubmit?: boolean, createdTemplateId?: string): void {
      this.clearFormState();

      const query = fromSubmit
        ? {
            clusterId: this.distributedMetaInfo.clusterId,
            ...(createdTemplateId && { createdEntityId: createdTemplateId }),
          }
        : undefined;
      this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX, query });
    },
    onBack(): void {
      this.trainingStore.clearDistributedMasterFlatFormState();
      this.wizardLayout.markStepIncomplete(EDistributedHybridStepName.Worker);
      this.wizardLayout.previousStep();
    },
    async initDistributedModel(
      clusterId?: string,
      fromCopyId?: string,
      scopeType?: ScopeType,
      scopeId?: string,
      name?: string,
      description?: string,
      noMaster?: boolean,
      distFramework?: string,
    ): Promise<void> {
      try {
        if (fromCopyId) {
          const workloadTemplate = await workloadTemplateService.getDistributedTemplateById(fromCopyId);

          // combineSpec is the spec as its suppose to be with the linked assets
          if (workloadTemplate.combinedSpec) workloadTemplate.spec = workloadTemplate.combinedSpec;
          if (workloadTemplate.combinedMasterSpec && workloadTemplate.master) {
            workloadTemplate.master.spec = workloadTemplate.combinedMasterSpec;
          }

          const uiModel = workloadHybridUtil.getHybridDistributedTemplateModel(workloadTemplate, noMaster, true);

          this.workerContext.hybridWorkloadModel.value = uiModel;

          if (workloadTemplate.assets) await this.workerContext.setupLinkedEntities(workloadTemplate.assets);

          // if its based on existing template, and sameAsWorker is true, we need to set master to be truthy so the form would recognize the need for a master step
          if (this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker) {
            this.workerContext.hybridWorkloadModel.value.masterSpec = {};
            this.masterContext.hybridWorkloadModel.value.spec = {};
          } else if (this.workerContext.hybridWorkloadModel.value.masterSpec) {
            this.masterContext.hybridWorkloadModel.value.spec = this.workerContext.hybridWorkloadModel.value.masterSpec;
          }

          if (workloadTemplate.master?.assets)
            await this.masterContext.setupLinkedEntities(workloadTemplate.master.assets);

          if (this.isEditMode) {
            this.originalName = uiModel.name;
            this.appStore.setPageTitleOverride(`Edit training template: ${uiModel.name}`);
          }
        } else {
          const clusterVersion = this.clusterStore.getClusterVersion(clusterId || "");
          this.workerContext.hybridWorkloadModel.value = workloadHybridUtil.setDistributedUIDefaults(
            workloadHybridUtil.getEmptyUISuperWorkloadModel(),
            this.authStore.isSSO,
            undefined,
            clusterVersion,
          );

          if (!noMaster) {
            // if noMaster is false, we need to set master to be truthy so the form would recognize the need for a master step and behave accordingly
            this.workerContext.hybridWorkloadModel.value.masterSpec = {};
            this.masterContext.hybridWorkloadModel.value.spec = {};
          }

          this.setFormsMetaData(
            scopeType as ScopeType,
            scopeId,
            name,
            clusterId,
            description,
            distFramework as DistributedFramework,
          );

          this.workerContext.hybridWorkloadModel.value.spec.numWorkers = 1;
        }

        this.setCleanPodPolicyDefaultValue();
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load template model"));
        console.error("Failed to get unified template:\n", e);
        this.appStore.setFallback(true);
      }
    },
    async loadSecrets(): Promise<void> {
      if (!this.isValidScopeSelected) {
        this.secrets = [];
        return;
      }

      try {
        this.secrets = await credentialService.listExistingSecrets(
          this.distributedMetaInfo.scopeType as string,
          this.distributedMetaInfo.scopeId as string,
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
      this.saveWorkerFormState();
      this.saveMasterFormState();
    },
    saveWorkerFormState(): void {
      const completedSteps = this.wizardLayout.steps.value.filter((step) => step.done).map((step) => step.name);

      const formState: THybridFormState = {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.workerProjectNodePoolsResources,
        selectedTemplates: this.workerContext.selectedTemplates.value,
        policy: undefined,
        policyViewDetails: {
          environment: {},
          compute: {},
          storage: {},
          general: {},
        },
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.workerContext.hybridWorkloadModel.value,
        completedSteps,
      };
      this.trainingStore.setDistributedWorkerFlatFormState(deepCopy(formState));
    },
    saveMasterFormState(): void {
      const formState: THybridFormState = {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.masterProjectNodePoolsResources,
        selectedTemplates: this.masterContext.selectedTemplates.value,
        policy: undefined,
        policyViewDetails: {
          environment: {},
          compute: {},
          storage: {},
          general: {},
        },
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.masterContext.hybridWorkloadModel.value,
      };
      this.trainingStore.setDistributedMasterFlatFormState(deepCopy(formState));
    },
    clearFormState(): void {
      this.trainingStore.clearDistributedWorkerFlatFormState();
      this.trainingStore.clearDistributedMasterFlatFormState();
      this.trainingStore.clearDistFlatFromContext();
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
    setCleanPodPolicyDefaultValue(): void {
      if (
        !this.workerContext.hybridWorkloadModel.value.spec.distributedFramework ||
        !!this.workerContext.hybridWorkloadModel.value.spec.cleanPodPolicy
      )
        return;
      if (
        !this.clusterStore.isClusterVersionSufficient(
          this.distributedMetaInfo.clusterId,
          MIN_CLUSTER_VERSION_CLEAN_POD_POLICY,
        )
      )
        return;

      const defaultMapPerFramework = {
        [DistributedFramework.PyTorch]: CleanPodPolicy.None,
        [DistributedFramework.XgBoost]: CleanPodPolicy.All,
        [DistributedFramework.Tf]: CleanPodPolicy.Running,
        [DistributedFramework.Mpi]: CleanPodPolicy.None,
        [DistributedFramework.Jax]: CleanPodPolicy.None,
      };

      this.workerContext.hybridWorkloadModel.value.spec.cleanPodPolicy =
        defaultMapPerFramework[this.workerContext.hybridWorkloadModel.value.spec.distributedFramework];
    },
    backToTemplateNew(): void {
      this.clearFormState();
      this.$router.push({
        name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW,
        state: {
          pageData: workloadHybridUtil.getFirstPageNavigationTemproraryData(
            this.workerContext.hybridWorkloadModel.value,
          ),
        },
      });
    },
    setFormsMetaData(
      scopeType?: ScopeType,
      scopeId?: string | null,
      name?: string,
      clusterId?: string,
      description?: string | null,
      distFramework?: DistributedFramework,
    ): void {
      this.workerContext.hybridWorkloadModel.value.scopeType = scopeType;
      this.workerContext.hybridWorkloadModel.value.scopeId = scopeId;
      this.workerContext.hybridWorkloadModel.value.name = name || "";
      this.workerContext.hybridWorkloadModel.value.clusterId = clusterId || "";
      this.workerContext.hybridWorkloadModel.value.description = description || "";
      this.workerContext.hybridWorkloadModel.value.spec.distributedFramework = distFramework as DistributedFramework;

      this.setMasterFormMetaData(scopeType, scopeId, name, clusterId, description);
    },
    setMasterFormMetaData(
      scopeType?: ScopeType,
      scopeId?: string | null,
      name?: string,
      clusterId?: string,
      description?: string | null,
    ): void {
      this.masterContext.hybridWorkloadModel.value.scopeType = scopeType;
      this.masterContext.hybridWorkloadModel.value.scopeId = scopeId;
      this.masterContext.hybridWorkloadModel.value.name = name || "";
      this.masterContext.hybridWorkloadModel.value.clusterId = clusterId || "";
      this.masterContext.hybridWorkloadModel.value.description = description || "";
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
          this.loadWorkloadPriorities(this.distributedMetaInfo.clusterId || ""),
          this.loadProjectData(),
        ]);
      } catch (error: unknown) {
        console.error("Failed to load scope based data:", error);
      }
    },
    validateScopeBasedData(): void {
      this.validateEnvironmentVariables(this.workerContext.hybridWorkloadModel.value.spec.environmentVariables || []);
    },
  },
  unmounted() {
    this.appStore.setPageTitleOverride(null);
  },
});

//Dima todo - Type assertion to prevent TS7056 while maintaining type safety -- will be refactored later --hide the annoying warning
export default TemplateDistributedHybrid as typeof TemplateDistributedHybrid & DefineComponent;
</script>
