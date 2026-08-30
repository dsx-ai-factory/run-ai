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
            :secrets="secrets"
            :submitting="submitting"
            :project-data="projectData"
            :policy="distPolicy?.worker"
            :policy-view-details="policyViewDetails"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :allowed-data-sources="allowedDatasourceOptions"
            :workload-priority-data="workloadPriorityData"
            :support-priority-class="supportPriorityClass"
            @credential-created="onCredentialCreated"
            :support-preemptibility="supportPreemptibility"
            :support-image-pull-secrets="supportImagePullSecrets"
            support-authenticated-ngc
            :support-termination-grace-period="isClusterSupportTerminationGracePeriod"
            :trigger-validation-on-start-up="triggerValidationOnStartUp"
            @update="updateWorkerSpec"
            @continue="handleNextStep"
            @save-form-state="saveFormState"
            @leave-page="redirectToPrevRoute"
            @back-to-create-new="backToCreateNew"
          />
        </template>
        <template v-else-if="wizardLayout.activeStep.value === EDistributedReplicaType.Master">
          <distributed-hybrid-master
            :hybrid-context="masterContext"
            :secrets="secrets"
            :form-meta-data="masterMetaData"
            :project-data="projectData"
            :policy="distPolicy?.master"
            :policy-view-details="policyViewDetails"
            :submitting="submitting"
            :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
            :allowed-data-sources="allowedDatasourceOptions"
            :workload-priority-data="workloadPriorityData"
            :support-priority-class="supportPriorityClass"
            :support-preemptibility="supportPreemptibility"
            :support-image-pull-secrets="supportImagePullSecrets"
            :support-termination-grace-period="isClusterSupportTerminationGracePeriod"
            support-authenticated-ngc
            @back="onBack"
            @credential-created="onCredentialCreated"
            @continue="onMasterContinue"
            @save-form-state="saveMasterFormState"
            @update="updateMasterSpec"
            @leave-page="redirectToPrevRoute"
          />
        </template>

        <template v-else-if="wizardLayout.activeStep.value === EDistributedHybridStepName.Setup">
          <distributed-hybrid-setup
            :hybrid-context="workerContext"
            :submitting="submitting"
            :policy="distPolicy?.worker"
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
import { EDistributedReplicaType, EWorkloadFormType, EWorkloadName, EWorkloadType } from "@/models/workload.model";
import { useAppStore } from "@/stores/app.store";
import {
  distributedHybridSteps,
  EDistributedHybridStepName,
  type TDistributedFormMetaInfo,
} from "./distributed-hybrid.model";
import DistributedHybridMaster from "./distributed-hybrid-master.vue";
import DistributedHybridWorker from "./distributed-hybrid-worker.vue";
import DistributedHybridSetup from "./distributed-hybrid-setup.vue";
import { alertUtil } from "@/utils/alert.util";
import type { UnifiedPolicyInfo, UnifiedSpecMaster } from "@/swagger-models/workloads-client";
import {
  CleanPodPolicy,
  type UnifiedProjectInfo,
  type UnifiedWorkload,
  type UnifiedWorkloadRequest,
} from "@/swagger-models/workloads-client";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useTrainingStore } from "@/stores/training.store";
import { useAuthStore } from "@/stores/auth.store";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { getDistributedPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { DistributedFramework, Scope, type StringOption } from "@/swagger-models/assets-service-client";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { deepCopy } from "@/utils/common.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import type { TProjectData } from "./distributed-hybrid-worker.vue";
import { policyUtil, clearWorkloadBlockedFields } from "@/utils/policy.util";
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { useDataSourceStore } from "@/stores/data-source.store";
import type { TDataSourceKinds } from "@/models/data-source.model";
import { useWorkloadStore } from "@/stores/workload.store";
import { MIN_CLUSTER_VERSION_CLEAN_POD_POLICY } from "@/common/version.constant";
import { SettingKeys } from "@/models/setting.model";
import { useWorkloadPriorities } from "@/composables/use-workload-priorities.composable";
import { useRoute } from "vue-router";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import { EHybridFormTypes } from "@/composables/helpers/use-hybrid-templates.mapper";
import { useImagePullSecrets } from "@/composables/use-image-pull-secrets.composable";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import {
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
} from "@/common/version.constant";
import type {
  DistributedPolicyDefaultsV2Worker,
  DistributedPolicyRulesV2Worker,
} from "@/swagger-models/policy-service-client";

const DistributedHybrid = defineComponent({
  name: "distributed-hybrid",
  components: {
    RunaiWizardLayout,
    DistributedHybridWorker,
    DistributedHybridMaster,
    DistributedHybridSetup,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      trainingStore: useTrainingStore(),
      clusterStore: useClusterStore(),
      dataSourceStore: useDataSourceStore(),
      workloadStore: useWorkloadStore(),
      settingStore: useSettingStore(),
      submitting: false,

      projectInfo: {} as UnifiedProjectInfo,
      selectedProjectNodeTypes: [] as string[],
      policyViewDetails: {} as IPolicyViewDetails,

      distPolicy: {} as UnifiedPolicyInfo | null | undefined,

      workerProjectNodePoolsResources: [] as StringOption[],
      masterProjectNodePoolsResources: [] as StringOption[],

      EDistributedReplicaType,
      EWorkloadFormType,
      EDistributedHybridStepName,

      secrets: [] as SecretListInfo[],
      allowedDatasourceOptions: undefined as Set<TDataSourceKinds> | undefined,
      selectedProject: {} as Project,

      triggerValidationOnStartUp: false as boolean,
    };
  },
  setup() {
    const route = useRoute();
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariables } =
      useEnvironmentVariablesSection();

    const { workloadPriorities, loadWorkloadPriorities, loadDefaultWorkloadPriority, supportPriorityClass } =
      useWorkloadPriorities(route?.query?.clusterId?.toString());

    const { supportImagePullSecrets } = useImagePullSecrets(route.query.clusterId?.toString());

    const workerContext = useHybridTemplates(EHybridFormTypes.DistributedWorker);

    const masterContext = useHybridTemplates(EHybridFormTypes.DistributedMaster, {
      name: "",
      projectId: workerContext.hybridWorkloadModel.value.projectId,
      clusterId: workerContext.hybridWorkloadModel.value.clusterId,
      spec: workerContext.hybridWorkloadModel.value.masterSpec || {},
    });

    const wizardLayout = useWizardLayout(distributedHybridSteps(), EDistributedHybridStepName.Worker);

    return {
      workerContext,
      masterContext,
      wizardLayout,

      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariables,
      workloadPriorities,
      loadWorkloadPriorities,
      loadDefaultWorkloadPriority,
      supportPriorityClass,
      supportImagePullSecrets,
    };
  },
  async created() {
    const { fromCopyId, templateId, projectId, clusterId, trainingName, distFramework, noMaster, workloadFormName } =
      this.$route.query;

    if (!!fromCopyId || !!templateId)
      this.trainingStore.setDistFlatFromContext({
        fromCopyId: fromCopyId?.toString() || null,
        templateId: this.supportsTemplatesV2 && templateId ? templateId.toString() : null,
      });

    const distPolicyFromStore = this.trainingStore.getDistPolicyFlat();
    if (distPolicyFromStore) this.distPolicy = distPolicyFromStore;

    const formState = this.trainingStore.getDistributedWorkerFlatFormState();
    if (formState) {
      this.workerContext.hybridWorkloadModel.value = formState.workloadModel;
      this.workerContext.selectedTemplates.value = formState.selectedTemplates;

      this.projectInfo = formState.projectInfo;
      this.policyViewDetails = formState.policyViewDetails;
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
    } else {
      await this.initDistributedModel(
        fromCopyId?.toString(),
        projectId?.toString(),
        clusterId?.toString(),
        trainingName?.toString(),
        templateId?.toString(),
        noMaster === "true",
        distFramework?.toString(),
      );
    }

    this.initRelevantFormSteps();
    if (workloadFormName) this.wizardLayout.goToStep(workloadFormName as EDistributedReplicaType);

    await Promise.all([
      this.loadSecrets(),
      this.loadSourceValueOptions(true, this.distributedMetaInfo.projectId),
      this.loadAvailableDataSource(),
      this.loadWorkloadPriorities(this.distributedMetaInfo.clusterId || ""),
      this.loadProjectById(this.distributedMetaInfo.projectId.toString()),
    ]);

    if (this.supportPriorityClass) {
      const defaultPriority = await this.loadDefaultWorkloadPriority(EWorkloadName.Distributed);
      if (
        defaultPriority &&
        !this.workerContext.hybridWorkloadModel.value.spec.priorityClass &&
        !this.$route.query.fromCopyId
      ) {
        this.workerContext.hybridWorkloadModel.value.spec.priorityClass = defaultPriority;
      }

      if (this.supportPreemptibility) {
        const defaultPreemptibility = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Distributed);
        if (!this.workerContext.hybridWorkloadModel.value.spec.preemptibility && !this.$route.query.fromCopyId) {
          this.workerContext.hybridWorkloadModel.value.spec.preemptibility = defaultPreemptibility;
        }
      }
    }

    this.validateEnvironmentVariables(this.workerContext.hybridWorkloadModel.value.spec.environmentVariables || []);
    this.appStore.setPageLoading(false);
  },
  computed: {
    distributedMetaInfo(): TDistributedFormMetaInfo {
      // the worker is the main body of the distributed workload so all meta info is taken from it
      return {
        projectId: this.workerContext.hybridWorkloadModel.value.projectId,
        clusterId: this.workerContext.hybridWorkloadModel.value.clusterId,
        framework: this.workerContext.hybridWorkloadModel.value.spec.distributedFramework || undefined,
        sameAsWorker: !!this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker,
      };
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    projectData(): TProjectData {
      return {
        projectName: this.projectInfo.name || "",
        projectNodePoolsResources: this.workerProjectNodePoolsResources,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
      };
    },
    conflictingReplicaPolicies(): boolean {
      const workerPolicyRules = this.distPolicy?.worker?.rules;
      const masterPolicyRules = this.distPolicy?.master?.rules;

      const workerDef = this.distPolicy?.worker?.defaults;
      const masterDef = this.distPolicy?.master?.defaults;

      if ((workerPolicyRules && masterPolicyRules) || (workerDef && masterDef)) {
        const workerPolicyStr = JSON.stringify(workerPolicyRules);
        const masterPolicyStr = JSON.stringify(masterPolicyRules);

        const workerDefStr = JSON.stringify(workerDef);
        const masterDefStr = JSON.stringify(masterDef);

        return workerPolicyStr !== masterPolicyStr || workerDefStr !== masterDefStr;
      }

      if (workerPolicyRules || masterPolicyRules) {
        return true;
      }

      return false;
    },
    masterMetaData(): TDistributedFormMetaInfo {
      return {
        ...this.distributedMetaInfo,
      };
    },
    supportsTemplatesV2(): boolean {
      return this.settingStore.isFeatureEnabled(SettingKeys.EnableTemplatesV2);
    },
    supportPreemptibility(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workerContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
      );
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
      };
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workerContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
  },
  methods: {
    async onCredentialCreated(): Promise<void> {
      await this.loadSourceValueOptions(true, this.workerContext.hybridWorkloadModel.value.projectId);
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
      const { name, clusterId, projectId } = this.workerContext.hybridWorkloadModel.value;
      this.setMasterFormMetaData(projectId, clusterId, name);

      // if there are conflicting policies we need to initialize the master spec from scratch based on its own defaults
      if (this.conflictingReplicaPolicies) {
        this.initMasterSpecFromScratch();
        return;
      }
      // if its clone operation OR its based on a template and master exists - do nothing we dont want to alter the original data
      if (
        (this.trainingStore.distFlatFromContext.fromCopyId ||
          (this.trainingStore.distFlatFromContext.templateId && this.supportsTemplatesV2)) &&
        this.workerContext.hybridWorkloadModel.value.masterSpec
      ) {
        // see comment below this condition for more details
        this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker = undefined;
        return;
      }

      // setting sameAsWorker to undefined and taking the form values for masterSpec
      // for context masterSpecSameAsWorker is more of a BE feature,
      // to avoid copying the worker spec into the master spec when you want to submit a distributed workload when they are identical
      this.workerContext.hybridWorkloadModel.value.masterSpecSameAsWorker = undefined;
      this.copyWorkerSpecToMaster();
    },
    initMasterSpecFromScratch(): void {
      const masterUIModel = workloadHybridUtil.getMasterModelForUI(
        (this.workerContext.hybridWorkloadModel.value.masterSpec as UnifiedSpecMaster) || {},
        this.workerContext.hybridWorkloadModel.value.spec.sshAuthMountPath || undefined,
      ); // need to pass unified model

      // the following fields affect both worker and master:
      masterUIModel.terminationGracePeriodSeconds =
        this.workerContext.hybridWorkloadModel.value.spec.terminationGracePeriodSeconds;
      masterUIModel.autoDeletionTimeAfterCompletionSeconds =
        this.workerContext.hybridWorkloadModel.value.spec.autoDeletionTimeAfterCompletionSeconds;
      masterUIModel.backoffLimit = this.workerContext.hybridWorkloadModel.value.spec.backoffLimit;
      masterUIModel.cleanPodPolicy = this.workerContext.hybridWorkloadModel.value.spec.cleanPodPolicy;
      masterUIModel.priorityClass = this.workerContext.hybridWorkloadModel.value.spec.priorityClass;
      const clusterVersion = this.clusterStore.getClusterVersion(this.workerContext.hybridWorkloadModel.value.clusterId);
      this.masterContext.hybridWorkloadModel.value.spec = workloadHybridUtil.setDistributedMasterSpecUIDefaults(
        masterUIModel,
        this.authStore.isSSO,
        this.distPolicy?.master,
        clusterVersion,
      );
    },
    copyWorkerSpecToMaster(): void {
      const workerSpec = deepCopy(this.workerContext.hybridWorkloadModel.value.spec);
      delete workerSpec.numWorkers;
      delete workerSpec.podAffinity;
      delete workerSpec.distributedFramework;
      delete workerSpec.minReplicas;
      delete workerSpec.maxReplicas;

      this.masterContext.hybridWorkloadModel.value.spec = workerSpec;

      this.matchModelToSameAsWorkerChanges();
    },
    matchModelToSameAsWorkerChanges(): void {
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].name = "";
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].id = "";
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].defaultData = null;
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Environment].isLinked = undefined;
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].name = "";
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].id = "";
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].defaultData = null;
      this.masterContext.selectedTemplates.value[WorkloadCreationDrawerEntities.Compute].isLinked = undefined;

      this.masterContext.reRenderSectionEntirely(WorkloadCreationDrawerEntities.Environment);
      this.masterContext.reRenderSectionEntirely(WorkloadCreationDrawerEntities.Compute);
      this.masterContext.createSectionsClearedStates();
    },
    updateWorkerSpec(spec: IUISuperWorkloadSpec): void {
      this.workerContext.hybridWorkloadModel.value.spec = spec;
    },
    updateMasterSpec(masterSpec: IUISuperWorkloadSpec): void {
      this.masterContext.hybridWorkloadModel.value.spec = masterSpec;
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;

        const cleanedWorkerSpec = policyUtil.cleanupFlatItemizeBeforeSubmission(
          this.workerContext.hybridWorkloadModel.value.spec,
          this.distPolicy?.worker,
        );

        let cleanedMasterSpec: IUISuperWorkloadSpec | undefined;

        if (this.workerContext.hybridWorkloadModel.value.masterSpec && this.masterContext.hybridWorkloadModel.value.spec)
          cleanedMasterSpec = policyUtil.cleanupFlatItemizeBeforeSubmission(
            this.masterContext.hybridWorkloadModel.value.spec,
            this.distPolicy?.master,
          );

        const request = workloadHybridUtil.prepareDistributedFlatRequest({
          ...this.workerContext.hybridWorkloadModel.value,
          spec: cleanedWorkerSpec,
          masterSpec: cleanedMasterSpec,
        });

        const createdDistributed = await trainingService.createDistributedFlat(request);

        if (createdDistributed) this.workloadStore.setWorkloadAdded(createdDistributed);

        this.$q.notify(alertUtil.getSuccess(`Training ${request.name} created`));
        this.redirectToPrevRoute(true);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("training"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error("Failed to create training:\n", error);
      } finally {
        this.submitting = false;
      }
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      this.clearFormState();

      const query = fromSubmit ? { clusterId: this.distributedMetaInfo.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    backToCreateNew(): void {
      this.clearFormState();
      this.$router.push({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
    },
    onBack(): void {
      this.trainingStore.clearDistributedMasterFlatFormState();
      this.wizardLayout.markStepIncomplete(EDistributedHybridStepName.Worker);
      this.wizardLayout.previousStep();
    },
    async initDistributedModel(
      workloadId?: string,
      projectId?: string,
      clusterId?: string,
      trainingName?: string,
      templateId?: string,
      noMaster?: boolean,
      distFramework?: string,
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

        const unifiedWorkload = await this.getUnifiedDistributed(request);

        this.distPolicy = unifiedWorkload.policyInfo;

        this.policyViewDetails = getDistributedPolicyViewDetails({
          defaults: {
            worker: this.distPolicy?.worker?.defaults as DistributedPolicyDefaultsV2Worker,
            master: this.distPolicy?.master?.defaults as DistributedPolicyDefaultsV2Worker,
          },
          rules: {
            worker: this.distPolicy?.worker?.rules as DistributedPolicyRulesV2Worker,
            master: this.distPolicy?.master?.rules as DistributedPolicyRulesV2Worker,
          },
        });

        this.projectInfo = unifiedWorkload.projectInfo || {};
        this.workerProjectNodePoolsResources = unifiedWorkload.policyInfo?.worker?.rules?.nodePools?.options || [];
        this.masterProjectNodePoolsResources = unifiedWorkload.policyInfo?.master?.rules?.nodePools?.options || [];

        const uiModel = workloadHybridUtil.getHybridDistributedModel(unifiedWorkload, noMaster, !!workloadId);
        const clusterVersion = this.clusterStore.getClusterVersion(clusterId || "");

        this.workerContext.hybridWorkloadModel.value = workloadHybridUtil.setDistributedUIDefaults(
          uiModel,
          this.authStore.isSSO,
          this.distPolicy?.worker,
          clusterVersion,
        );

        if (uiModel.masterSpec) this.masterContext.hybridWorkloadModel.value.spec = uiModel.masterSpec;

        clearWorkloadBlockedFields(this.workerContext.hybridWorkloadModel.value.spec, this.distPolicy?.worker);
        if (this.masterContext.hybridWorkloadModel.value.spec) {
          clearWorkloadBlockedFields(this.masterContext.hybridWorkloadModel.value.spec, this.distPolicy?.master);
        }

        // need to take into account the metadata coming from copied workload
        this.setFormsMetaData(projectId, clusterId, trainingName, distFramework as DistributedFramework);

        this.setCleanPodPolicyDefaultValue();
      } catch (e) {
        this.$q.notify(alertUtil.getError("Failed to load training model"));
        console.error("Failed to get unified training:\n", e);
        this.appStore.setFallback(true);
      }
    },
    async getUnifiedDistributed(options?: UnifiedWorkloadRequest): Promise<UnifiedWorkload> {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Distributed,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        templateId: options?.templateId,
      };
      return await workloadService.getUnifiedWorkload(request);
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
        policy: null,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.workerContext.hybridWorkloadModel.value,
        completedSteps,
      };
      this.trainingStore.setDistPolicyFlat(this.distPolicy);
      this.trainingStore.setDistributedWorkerFlatFormState(deepCopy(formState));
    },
    saveMasterFormState(): void {
      const formState: THybridFormState = {
        projectInfo: this.projectInfo,
        projectNodePoolsResources: this.masterProjectNodePoolsResources,
        selectedTemplates: this.masterContext.selectedTemplates.value,
        policy: null,
        policyViewDetails: this.policyViewDetails,
        selectedProjectNodeTypes: this.selectedProjectNodeTypes,
        workloadModel: this.masterContext.hybridWorkloadModel.value,
      };
      this.trainingStore.setDistPolicyFlat(this.distPolicy);
      this.trainingStore.setDistributedMasterFlatFormState(deepCopy(formState));
    },
    clearFormState(): void {
      this.trainingStore.clearDistributedWorkerFlatFormState();
      this.trainingStore.clearDistributedMasterFlatFormState();
      this.trainingStore.clearDistPolicyFlat();
      this.trainingStore.clearDistFlatFromContext();
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
    setFormsMetaData(
      projectId?: string | null,
      clusterId?: string,
      name?: string,
      distFramework?: DistributedFramework,
    ): void {
      if (!this.workerContext.hybridWorkloadModel.value.spec.distributedFramework)
        this.workerContext.hybridWorkloadModel.value.spec.distributedFramework = distFramework as DistributedFramework;
      if (!this.workerContext.hybridWorkloadModel.value.spec.numWorkers)
        this.workerContext.hybridWorkloadModel.value.spec.numWorkers = 1;
      if (projectId) this.workerContext.hybridWorkloadModel.value.projectId = projectId;
      if (clusterId) this.workerContext.hybridWorkloadModel.value.clusterId = clusterId;
      if (name) this.workerContext.hybridWorkloadModel.value.name = name;

      this.setMasterFormMetaData(projectId, clusterId, name);
    },
    setMasterFormMetaData(projectId?: string | null, clusterId?: string, name?: string): void {
      this.masterContext.hybridWorkloadModel.value.projectId = projectId || "";
      this.masterContext.hybridWorkloadModel.value.clusterId = clusterId || "";
      this.masterContext.hybridWorkloadModel.value.name = name || "";
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default DistributedHybrid as typeof DistributedHybrid & DefineComponent;
</script>
