<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="distributed-assets-edit column items-center q-pt-md">
    <template v-if="isPageReady && selectedProject">
      <section class="workload-distributed-create-form">
        <runai-wizard-layout
          disable-navigation
          transparent-steps-panel
          :steps="wizardLayout.steps.value"
          :active-step="wizardLayout.activeStep.value"
          @step-click="onStepClicked"
        >
          <template #default>
            <div v-if="wizardLayout.activeStep.value === EDistributedReplicaType.Worker" class="step-workers">
              <workload-form-wrapper
                :form-state="training"
                :submitting="submitting"
                :submit-btn-text="firstStepConfirmButtonText"
                @leave-page="$emit('leave-page', $event)"
                @canceled="onCancel"
                @submit="handleNextStep"
                cancel-btn-text="Cancel"
              >
                <title-and-back-section
                  label="Cluster"
                  :summary="clusterName"
                  back-tooltip="Change cluster"
                  @on-back="onBack"
                />

                <title-and-back-section
                  label="Project"
                  :summary="projectSummary"
                  back-tooltip="Change project"
                  aid="project-section"
                  @on-back="onBack"
                />

                <multi-node-section
                  :is-selectable="false"
                  :multi-node-section-model="training.distributed"
                  :cluster-id="training.clusterId"
                >
                  <template #custom-button>
                    <runai-btn
                      icon="ph-fill ph-pencil-simple"
                      flat
                      round
                      size="sm"
                      class="runai-icon-btn--default"
                      @click="onBack"
                      aid="back-btn"
                    >
                      <q-tooltip anchor="center right" self="center left">Edit</q-tooltip>
                    </runai-btn>
                  </template>
                </multi-node-section>

                <workload-name-section
                  :entity-type="formType"
                  :default-open="!!fromCopyId"
                  :name="training.name"
                  @update:name="onNameChanged"
                  :project-id="training.projectId"
                  :cluster-id="training.clusterId"
                />

                <environment-section
                  :entity-type="formType"
                  aid="environment-section"
                  :loading="false"
                  :environments="assetsLists.environments"
                  :environment-id="training.assets.environment || ''"
                  :specific-env="workersEnvironmentRunParams"
                  @environment-changed="updateWorkersSelectedEnvironment"
                  @create-new="onCreateNewEnvironment"
                  :section-options="workersEnvSectionOptions"
                  :cluster-id="training.clusterId"
                  :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
                  :support-config-map-as-env-var="supportConfigMapAsEnvVar"
                  :policy-rules="loadedWorkerPolicyInfo"
                  :itemize-defaults="loadedWorkerItemizeDefaults"
                  :loaded-specific-run-params-with-defaults="loadedWorkerSpecificRunParamsWithDefaults"
                />

                <compute-resource-section
                  aid="compute-resource-section"
                  :entity-type="formType"
                  :loading="false"
                  :compute-resources="assetsLists.computes"
                  :node-affinity="nodeAffinity"
                  @compute-resource-data-changed="onWorkersComputeResourceDataChanged"
                  :compute-resource-data="workersComputeResourceData"
                  @create-new="onCreateNewComputeResource"
                  is-required
                  :section-options="workersComputeSectionOptions"
                  :policy-rules="loadedWorkerPolicyInfo"
                  :itemize-defaults="loadedWorkerItemizeDefaults"
                  :support-tolerations="supportTolerations"
                />

                <volume-section
                  :volumes="training.assets.uiVolumes || []"
                  :storage-classes="storageClasses"
                  @update-volumes="onVolumesChanged"
                  :policy-rules="loadedWorkerPolicyInfo?.rules?.assets?.datasources?.pvc"
                  :policy-defaults="loadedAssetsAttributesDefaults?.datasources?.pvc"
                  :loading-storage-class="loadingStorageClass"
                />

                <data-source-section
                  :entity-type="formType"
                  aid="data-source-section"
                  :loading="false"
                  :data-sources="assetsLists.datasources"
                  :created-data-source-id="createdDataSourceId"
                  :default-opened="!!createdDataSourceId"
                  :selected-data-sources="training.assets.datasources || []"
                  @datasource-changed="updateWorkersSelectedDataSources"
                  @create-new="onCreateNewDataSource"
                  :section-options="workerDataSourceSectionOptions"
                  :cluster-id="training.clusterId"
                  :policy-rules="loadedWorkerPolicyInfo"
                />

                <general-section
                  :workload-form-type="formType"
                  :general-model="workersGeneralSectionModel"
                  @general-model-changed="updateWorkersGeneralModel"
                  :show-backoff-limit="supportBackoffLimit"
                  :policy-rules="loadedWorkerPolicyInfo?.rules?.specificRunParams"
                  :itemize-defaults="loadedWorkerItemizeDefaults"
                  :show-termination-grace-period="isClusterSupportTerminationGracePeriod"
                />
              </workload-form-wrapper>
            </div>
            <div
              v-if="!noMasterStep && wizardLayout.activeStep.value === EDistributedReplicaType.Master"
              class="step-master"
            >
              <workload-form-wrapper
                :form-state="training"
                :submitting="submitting"
                submit-btn-text="Create Training"
                cancel-btn-text="Back"
                @leave-page="$emit('leave-page', $event)"
                @canceled="onBackStep"
                @submit="onSubmit"
              >
                <toggle-editing-form
                  aid="toggle-editing-form"
                  :is-enabled="training.enableEditingMaster"
                  :section-options="masterToggleSectionOptions"
                  @toggle-editing-changed="onToggleEditingChanged"
                  :section-disabled="isMasterToggleDisabled"
                />

                <environment-section
                  :entity-type="formType"
                  aid="environment-section"
                  :loading="false"
                  :environments="masterAssetsLists.environments"
                  :environment-id="training.distributed?.master?.assets.environment || ''"
                  :specific-env="masterEnvironmentRunParams"
                  @environment-changed="updateMasterSelectedEnvironment"
                  @create-new="onCreateNewEnvironment"
                  :section-options="masterEnvSectionOptions"
                  :section-disabled="isMasterViewOnly"
                  :cluster-id="training.clusterId"
                  :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
                  :support-config-map-as-env-var="supportConfigMapAsEnvVar"
                  :policy-rules="loadedMasterPolicyInfo"
                  :itemize-defaults="loadedMasterItemizeDefaults"
                  :loaded-specific-run-params-with-defaults="loadedMasterSpecificRunParamsWithDefaults"
                  :access-disabled="!training.enableEditingMaster"
                />

                <compute-resource-section
                  aid="compute-resource-section"
                  :entity-type="formType"
                  :loading="false"
                  :compute-resources="masterAssetsLists.computes"
                  :node-affinity="nodeAffinity"
                  @compute-resource-data-changed="onMasterComputeResourceDataChanged"
                  :compute-resource-data="masterComputeResourceData"
                  @create-new="onCreateNewComputeResource"
                  is-required
                  :section-options="masterComputeSectionOptions"
                  :section-disabled="isMasterViewOnly"
                  :policy-rules="loadedMasterPolicyInfo"
                  :itemize-defaults="loadedMasterItemizeDefaults"
                  :support-tolerations="supportTolerations"
                />

                <volume-section
                  :volumes="training.distributed?.master?.assets.uiVolumes || []"
                  :storage-classes="storageClasses"
                  :section-disabled="isMasterViewOnly"
                  @update-volumes="onMasterVolumesChanged"
                  :policy-rules="loadedMasterPolicyInfo?.rules?.assets?.datasources?.pvc"
                  :policy-defaults="loadedMasterAssetsAttributesDefaults?.datasources?.pvc"
                  :loading-storage-class="loadingStorageClass"
                />

                <data-source-section
                  :entity-type="formType"
                  aid="data-source-section"
                  :loading="false"
                  :data-sources="masterAssetsLists.datasources"
                  :created-data-source-id="createdDataSourceId"
                  :default-opened="!!createdDataSourceId || masterDataSourceSectionForceOpen"
                  :selected-data-sources="training.distributed?.master?.assets.datasources || []"
                  @datasource-changed="updateMasterSelectedDataSources"
                  @create-new="onCreateNewDataSource"
                  :section-options="masterDataSourceSectionOptions"
                  :section-disabled="isMasterViewOnly"
                  :cluster-id="training.clusterId"
                  :policy-rules="loadedMasterPolicyInfo"
                />

                <general-section
                  :workload-form-type="formType"
                  :general-model="masterGeneralSectionModel"
                  @general-model-changed="updateMasterGeneralModel"
                  :section-disabled="isMasterViewOnly"
                  :show-backoff-limit="supportBackoffLimit"
                  :policy-rules="loadedMasterPolicyInfo?.rules?.specificRunParams"
                  :itemize-defaults="loadedMasterItemizeDefaults"
                  :show-termination-grace-period="isClusterSupportTerminationGracePeriod"
                />
              </workload-form-wrapper>
            </div>
          </template>
        </runai-wizard-layout>
      </section>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, computed, type DefineComponent } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { RunaiWizardLayout } from "@/components/common/runai-wizard-layout";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { EnvironmentSection } from "@/components/section/environment-section";
import { ComputeResourceSection, ENodePoolsListOrigin } from "@/components/section/compute-resource-section";
import { DataSourceSection, type IWorkloadDataSourceSectionOptions } from "@/components/section/data-source-section";
import { VolumeSection } from "@/components/section/volume-section";
import { GeneralSection, type IUIGeneralSectionModel } from "@/components/section/general-section";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import {
  type IWorkloadToggleEditingFormSectionOptions,
  ToggleEditingForm,
} from "@/components/section/toggle-editing-form";
import { MultiNodeSection } from "@/components/section/multi-node-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// models
import {
  EDistributedReplicaType,
  EWorkloadFormType,
  type IUIWorkloadSpecificEnv,
  type IWorkloadSupportedRoutes,
  type IUIWorkloadCreation,
  EWorkloadTypeSupportedInAssetsService,
} from "@/models/workload.model";
import {
  Scope,
  type EnvironmentAsset,
  type ComputeAsset,
  type AssetIdAndKind,
  AssetKind,
  type PVCAsset,
  type DatasourceListResponseEntry,
  type TrainingCreationRequestV2,
  type TrainingV2,
  UidGidSource,
  type PolicyInfo,
} from "@/swagger-models/assets-service-client";
import { PolicyType, type DistributedPolicyV2 } from "@/swagger-models/policy-service-client";
import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "@/components/section/environment-section";
import type { IUIVolume } from "@/models/data-source.model";
import type {
  IComputeSectionData,
  IComputeSectionNodePoolData,
  IWorkloadComputeSectionOptions,
} from "@/components/section/compute-resource-section";
import type { IAssetsFilter } from "@/models/filter.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { Project, Resources } from "@/swagger-models/org-unit-service-client";
// stores
import { useAppStore } from "@/stores/app.store";
import { useTrainingStore } from "@/stores/training.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";

// services
import { trainingService } from "@/services/control-plane/training.service/training.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// utils
import { deepCopy } from "@/utils/common.util";
import { alertUtil } from "@/utils/alert.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { environmentSectionUtil } from "@/components/section/environment-section";
import { policyUtil } from "@/utils/policy.util";
import { intervalUtil } from "@/utils/interval.util";

// routes
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// common
import { ErrorAlert } from "@/utils/error-alert.util";
import {
  MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
  MIN_CLUSTER_VERSION_FOR_TOPOLOGY,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import { EIntervalLabels } from "@/models/interval.model";
import { errorMessages } from "@/common/error-message.constant";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";

const distributedAssetsSteps = (): IRunaiWizardLayoutStep[] => [
  { name: EDistributedReplicaType.Worker, title: "Workers", done: false },
  { name: EDistributedReplicaType.Master, title: "Master", done: false },
];

export interface ITrainingFormAssetsLists {
  environments: EnvironmentAsset[];
  computes: ComputeAsset[];
  datasources: DatasourceListResponseEntry[];
}

const DistributedAssetsEdit = defineComponent({
  components: {
    RunaiWizardLayout,
    WorkloadFormWrapper,
    WorkloadNameSection,
    EnvironmentSection,
    ComputeResourceSection,
    DataSourceSection,
    VolumeSection,
    GeneralSection,
    TitleAndBackSection,
    ToggleEditingForm,
    MultiNodeSection,
    RunaiBtn,
  },
  provide() {
    return {
      policy: computed(() => this.trainingPolicy),
      distributedReplicaType: computed(() => this.wizardLayout.activeStep.value),
    };
  },
  data() {
    return {
      appStore: useAppStore(),
      trainingStore: useTrainingStore(),
      environmentStore: useEnvironmentStore(),
      computeResourceStore: useComputeResourceStore(),
      workloadStore: useWorkloadStore(),
      authStore: useAuthStore(),
      nodeAffinity: [] as string[],
      selectedProject: null as Project | null,
      dataSourceStore: useDataSourceStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      formType: EWorkloadFormType.Training,
      training: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,
      trainingPolicy: null as DistributedPolicyV2 | null,
      submitting: false as boolean,
      envSectionOptions: {
        canAddEnvVariable: true,
      } as IWorkloadEnvSectionOptions,
      computeSectionOptions: {
        applyPolicyDefaults: false,
      } as IWorkloadComputeSectionOptions,
      assetsFilter: {} as IAssetsFilter,
      assetsLists: {
        computes: [],
        datasources: [],
        environments: [],
      } as ITrainingFormAssetsLists,
      masterAssetsLists: {
        computes: [],
        datasources: [],
        environments: [],
      } as ITrainingFormAssetsLists,
      fromCopyId: "" as string,
      createdDataSourceId: "" as string,
      masterDataSourceSectionForceOpen: false as boolean,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      EDistributedReplicaType,
    };
  },
  setup() {
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariablesOfAsset } =
      useEnvironmentVariablesSection();

    const { storageClasses, loadingStorageClass, loadStorageClasses, loadStorageClassConfiguration } =
      useStorageClassConfiguration("workload");

    const wizardLayout = useWizardLayout(distributedAssetsSteps(), EDistributedReplicaType.Worker);

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariablesOfAsset,

      storageClasses,
      loadingStorageClass,
      loadStorageClasses,
      loadStorageClassConfiguration,

      wizardLayout,
    };
  },
  async created() {
    const { kind, fromCopyId, createdEntityId } = this.$route.query;
    this.training = this.trainingStore.training;
    try {
      if (fromCopyId) {
        await this.loadFromExistingTraining(fromCopyId.toString());
        this.fromCopyId = String(fromCopyId);
      }
      this.initRelevantFormSteps();
      this.initFilterProperties();

      await this.loadProjectById(this.training.projectId?.toString());

      if (this.selectedProject && !fromCopyId)
        this.training.specificEnv.nodePools = this.selectedProject.effective.defaultNodePools || [];

      this.setNodeAffinity();
      await this.loadPolicy();

      const createdEnvironmentId: string | undefined =
        kind === AssetKind.Environment ? createdEntityId?.toString() : undefined;
      const createdComputeId: string | undefined = kind === AssetKind.Compute ? createdEntityId?.toString() : undefined;

      const isDataSource = [
        AssetKind.HostPath,
        AssetKind.Nfs,
        AssetKind.Git,
        AssetKind.S3,
        AssetKind.Pvc,
        AssetKind.ConfigMap,
        AssetKind.SecretVolume,
      ].some((assetKind) => assetKind === kind);

      const createdDataSourceId: string | undefined = isDataSource ? createdEntityId?.toString() : undefined;
      createdDataSourceId && (this.createdDataSourceId = createdDataSourceId);

      await Promise.all([
        this.loadEnvironments(),
        this.loadComputeResources(),
        this.loadDataSources(),
        this.loadStorageClasses(this.clusterUid, [
          this.training.assets.uiVolumes?.[0]?.claimInfo.storageClass,
          this.training.distributed?.master?.assets.uiVolumes?.[0]?.claimInfo.storageClass,
        ]),
        this.loadSourceValueOptions(false, this.training.projectId ? String(this.training.projectId) : undefined),
      ]);

      if (fromCopyId) this.cleanupNoneComplianceAssets();

      this.setCreatedAssets(createdEnvironmentId, createdComputeId, createdDataSourceId, kind as AssetKind);
      this.validateEnvironmentVariablesOfAsset(this.training.specificEnv.environmentVariables || []);
      this.validateEnvironmentVariablesOfAsset(
        this.training.distributed?.master?.specificEnv.environmentVariables || [],
      );

      if (createdDataSourceId && kind === AssetKind.Pvc) {
        const relevantForm = this.$route.query.workloadFormName as EDistributedReplicaType;
        this.updateCreatedPvc(createdDataSourceId, relevantForm);
      }

      this.computeSectionOptions.applyPolicyDefaults = !!fromCopyId || !!this.trainingStore.templateId;
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    } finally {
      this.appStore.setPageLoading(false);
    }
  },
  mounted() {
    if (this.$route.query.workloadFormName) {
      this.wizardLayout.goToStep(this.$route.query.workloadFormName as EDistributedReplicaType);
    }
  },
  computed: {
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    projectId(): number {
      return this.training.projectId;
    },
    clusterUid(): string {
      return this.training.clusterId;
    },
    isDistributedPoliciesConflictExists(): boolean {
      if (this.training.assets.environment) {
        const selectedEnvironment: EnvironmentAsset | undefined = this.masterAssetsLists.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === this.training.assets.environment,
        );
        if (!selectedEnvironment?.compliance?.compliance) {
          return true;
        }
      }
      if (this.training.assets.compute) {
        const selectedCompute: ComputeAsset | undefined = this.masterAssetsLists.computes.find(
          (compute: ComputeAsset) => compute.meta.id === this.training.assets.compute,
        );

        if (!selectedCompute?.compliance?.compliance) {
          return true;
        }
      }
      if (this.training.assets.datasources?.length) {
        const nonComplianceDataSources: Set<string> = new Set(
          this.masterAssetsLists.datasources
            .filter((dataSource: DatasourceListResponseEntry) => !dataSource.compliance?.compliance)
            .map((dataSource: DatasourceListResponseEntry) => dataSource.meta.id),
        );
        return this.training.assets.datasources.some((datasource: AssetIdAndKind) =>
          nonComplianceDataSources.has(datasource.id),
        );
      }
      return false;
    },
    isMasterToggleDisabled(): boolean {
      return this.isDistributedPoliciesConflictExists || this.isMasterPolicyDefined;
    },
    isMasterPolicyDefined(): boolean {
      return !!this.trainingStore.loadedTrainingModelWithDefaultsAndPolicy?.distributed?.master?.policyInfo;
    },
    clusterName(): string {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === this.training.clusterId)?.name || "";
    },
    workersEnvironmentRunParams(): IUIWorkloadEnvSectionModel {
      const specificEnv = this.training.specificEnv;
      return {
        command: specificEnv?.command || "",
        args: specificEnv?.args || "",
        environmentVariables: specificEnv?.environmentVariables || [],
        connections: specificEnv?.connections || [],
        runAsGid: specificEnv?.runAsGid || null,
        runAsUid: specificEnv?.runAsUid || null,
        supplementalGroups: specificEnv?.supplementalGroups || null,
      };
    },
    masterEnvironmentRunParams(): IUIWorkloadEnvSectionModel {
      const specificEnv = this.training.distributed?.master?.specificEnv;
      return {
        command: specificEnv?.command || "",
        args: specificEnv?.args || "",
        environmentVariables: specificEnv?.environmentVariables || [],
        connections: specificEnv?.connections || [],
        runAsGid: specificEnv?.runAsGid || null,
        runAsUid: specificEnv?.runAsUid || null,
        supplementalGroups: specificEnv?.supplementalGroups || null,
      };
    },
    workersComputeResourceData(): IComputeSectionData {
      return {
        computeResourceId: this.training.assets.compute || null,
        nodeType: this.training.specificEnv?.nodeType || null,
        nodePools: this.getNodePoolsData(this.training.specificEnv, this.selectedProject?.resources),
        workers: this.training.distributed?.numWorkers || null,
        autoScaleData: this.training.specificEnv?.autoScaleData || undefined,
        tolerations: this.training.specificEnv?.tolerations,
        podAffinity: this.training.specificEnv?.podAffinity,
      };
    },
    masterComputeResourceData(): IComputeSectionData {
      return {
        computeResourceId: this.training.distributed?.master?.assets.compute || null,
        nodeType: this.training.distributed?.master?.specificEnv?.nodeType || null,
        nodePools: this.getNodePoolsData(
          this.training.distributed?.master?.specificEnv,
          this.selectedProject?.resources,
        ),
        autoScaleData: this.training.distributed?.master?.specificEnv?.autoScaleData || undefined,
        tolerations: this.training.distributed?.master?.specificEnv?.tolerations,
        podAffinity: this.training.distributed?.master?.specificEnv?.podAffinity,
      };
    },
    workersGeneralSectionModel(): IUIGeneralSectionModel {
      return {
        allowOverQuota: this.training.specificEnv?.allowOverQuota || false,
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.training.specificEnv?.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.training.specificEnv?.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: this.training.specificEnv?.annotations,
        labels: this.training.specificEnv?.labels,
        backoffLimit: Number.isInteger(this.training.specificEnv?.backoffLimit)
          ? this.training.specificEnv.backoffLimit
          : null,
        terminationGracePeriodSeconds: this.training.specificEnv?.terminationGracePeriodSeconds || null,
      };
    },

    masterGeneralSectionModel(): IUIGeneralSectionModel {
      return {
        allowOverQuota: this.training.distributed?.master?.specificEnv.allowOverQuota || false,
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.training.distributed?.master?.specificEnv.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.training.distributed?.master?.specificEnv.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: this.training.distributed?.master?.specificEnv.annotations,
        labels: this.training.distributed?.master?.specificEnv.labels,
        backoffLimit: Number.isInteger(this.training.distributed?.master?.specificEnv?.backoffLimit)
          ? this.training.distributed?.master?.specificEnv.backoffLimit
          : null,
        terminationGracePeriodSeconds:
          this.training.distributed?.master?.specificEnv?.terminationGracePeriodSeconds || null,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(this.clusterUid);
    },
    noMasterStep(): boolean {
      return !!this.training.distributed?.noMaster;
    },
    firstStepConfirmButtonText(): string {
      return this.noMasterStep ? "Create Training" : "Continue";
    },
    workersEnvSectionOptions(): IWorkloadEnvSectionOptions {
      return {
        canAddEnvVariable: true,
        isToolBoxRequired: true,
        ...this.envSectionOptions,
        cardsTitle: "Select the environment for the workers",
      };
    },
    workersComputeSectionOptions(): IWorkloadComputeSectionOptions {
      return {
        ...this.computeSectionOptions,
        cardsTitle: "Select the node resources needed to run each worker",
        showTopologySection: this.clusterStore.isClusterVersionSufficient(
          this.clusterUid,
          MIN_CLUSTER_VERSION_FOR_TOPOLOGY,
        ),
        allowAnyEffect: this.clusterStore.isClusterVersionSufficient(
          this.clusterUid,
          MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
        ),
      };
    },
    workerDataSourceSectionOptions(): IWorkloadDataSourceSectionOptions {
      return {
        cardsTitle: "Select the data sources your workers need to access",
      };
    },
    masterToggleSectionOptions(): IWorkloadToggleEditingFormSectionOptions {
      return {
        mainTitle: "Allow different setup for the master",
        disabledTooltip: this.isMasterToggleDisabled ? errorMessages.CANT_BE_MODIFIED_POLICY : "",
        tooltip: `
          When enabled, you can change the master's setup to be different from the workers'.
          When disabled, the master will inherit the workers' setup automatically.
          `,
      };
    },
    masterEnvSectionOptions(): IWorkloadEnvSectionOptions {
      return {
        canAddEnvVariable: true,
        isToolBoxRequired: true,
        ...this.envSectionOptions,
        cardsTitle: "Select the environment for the master",
      };
    },
    masterComputeSectionOptions(): IWorkloadComputeSectionOptions {
      return {
        ...this.computeSectionOptions,
        applyPolicyDefaults: false,
        cardsTitle: "Select the node resources needed to run the master",
        hideNodePools: true,
        allowAnyEffect: this.clusterStore.isClusterVersionSufficient(
          this.clusterUid,
          MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
        ),
      };
    },
    masterDataSourceSectionOptions(): IWorkloadDataSourceSectionOptions {
      return {
        cardsTitle: "Select the data sources your master need to access",
      };
    },
    isMasterViewOnly(): boolean {
      if (this.isMasterToggleDisabled) return false;
      return !this.training.enableEditingMaster;
    },
    loadedWorkerPolicyInfo(): PolicyInfo | null | undefined {
      return this.trainingStore.loadedPolicyInfo;
    },
    loadedWorkerItemizeDefaults(): ILoadedItemizeDefaults | null | undefined {
      return {
        environmentVariables: this.trainingPolicy?.effective?.defaults?.worker?.environmentVariables,
        annotations: this.trainingPolicy?.effective?.defaults?.worker?.annotations,
        labels: this.trainingPolicy?.effective?.defaults?.worker?.labels,
        tolerations: this.trainingPolicy?.effective?.defaults?.worker?.tolerations,
      };
    },
    loadedWorkerSpecificRunParamsWithDefaults() {
      return this.trainingStore.loadedSpecificRunParamsWithDefaults;
    },
    loadedMasterPolicyInfo(): PolicyInfo | null | undefined {
      return this.trainingStore.loadedPolicyInfoDistributedMaster;
    },
    loadedMasterItemizeDefaults(): ILoadedItemizeDefaults | null | undefined {
      return {
        environmentVariables: this.trainingPolicy?.effective?.defaults?.master?.environmentVariables,
        annotations: this.trainingPolicy?.effective?.defaults?.master?.annotations,
        labels: this.trainingPolicy?.effective?.defaults?.master?.labels,
        tolerations: this.trainingPolicy?.effective?.defaults?.master?.tolerations,
      };
    },
    loadedMasterSpecificRunParamsWithDefaults() {
      return this.trainingStore.loadedSpecificRunParamsWithDefaultsDistributedMaster;
    },
    loadedAssetsAttributesDefaults() {
      return this.trainingStore.loadedAssetsAttributesDefaults;
    },
    loadedMasterAssetsAttributesDefaults() {
      return this.trainingStore.loadedAssetsAttributesDefaultsMaster;
    },
    projectSummary(): string {
      return this.selectedProject?.name || "None";
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.clusterUid,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterUid, MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR);
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterUid, MIN_CLUSTER_VERSION_FOR_TOLERATIONS);
    },
  },
  methods: {
    initFilterProperties(): void {
      this.assetsFilter = {
        projectId: this.training.projectId,
        usageInfo: true,
        complyToProject: this.training.projectId,
        complyToWorkloadType: this.training.distributed
          ? EWorkloadTypeSupportedInAssetsService.Distributed
          : EWorkloadTypeSupportedInAssetsService.Training,
        isTraining: true,
        isDistributed: !!this.training.distributed,
      };
    },
    async loadEnvironments(): Promise<void> {
      if (!this.selectedProject) return;
      this.assetsLists.environments = await this.environmentStore.loadEnvironments({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Worker,
        distributedFramework: this.training.distributed?.distFramework,
      });
      this.masterAssetsLists.environments = await this.environmentStore.loadEnvironments({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Master,
        distributedFramework: this.training.distributed?.distFramework,
      });

      // Set security fields according to the environment settings
      this.setEnvironmentSecurityFields();
    },
    setEnvironmentSecurityFields(): void {
      if (this.training.assets.environment) {
        const selectedEnvironment: EnvironmentAsset | undefined = this.assetsLists.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === this.training.assets.environment,
        );
        if (selectedEnvironment) {
          const { uidGidSource } = selectedEnvironment.spec;
          if (uidGidSource === UidGidSource.FromIdpToken) {
            this.training.specificEnv.runAsGid = this.authStore.getGID;
            this.training.specificEnv.runAsUid = this.authStore.getUID;
            this.training.specificEnv.supplementalGroups = this.authStore.getSupplementaryGroups;
            this.saveTraining(this.training);
          }
        }
      }
    },
    async loadComputeResources(): Promise<void> {
      if (!this.selectedProject) return;

      this.assetsLists.computes = await this.computeResourceStore.loadComputeResources({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Worker,
      });
      this.masterAssetsLists.computes = await this.computeResourceStore.loadComputeResources({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Master,
      });
    },
    async loadDataSources(): Promise<void> {
      if (!this.selectedProject) return;

      this.assetsLists.datasources = await this.dataSourceStore.loadDataSources({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Worker,
      });
      this.masterAssetsLists.datasources = await this.dataSourceStore.loadDataSources({
        ...this.assetsFilter,
        complyToReplicaType: EDistributedReplicaType.Master,
      });
    },
    cleanupNoneComplianceAssets(masterOnly = false): void {
      if (!masterOnly) {
        if (this.training.assets.environment) {
          const selectedEnvironment: EnvironmentAsset | undefined = this.assetsLists.environments.find(
            (environment: EnvironmentAsset) => environment.meta.id === this.training.assets.environment,
          );
          if (!selectedEnvironment?.compliance?.compliance) {
            this.training.assets.environment = "";
          }
        }
        if (this.training.assets.compute) {
          const selectedCompute: ComputeAsset | undefined = this.assetsLists.computes.find(
            (compute: ComputeAsset) => compute.meta.id === this.training.assets.compute,
          );

          if (!selectedCompute?.compliance?.compliance) {
            this.training.assets.compute = "";
          }
        }
        if (this.training.assets.datasources?.length) {
          const nonComplianceDataSources: Set<string> = new Set(
            this.assetsLists.datasources
              .filter((dataSource: DatasourceListResponseEntry) => !dataSource.compliance?.compliance)
              .map((dataSource: DatasourceListResponseEntry) => dataSource.meta.id),
          );
          this.training.assets.datasources = this.training.assets.datasources.filter(
            (datasource: AssetIdAndKind) => !nonComplianceDataSources.has(datasource.id),
          );
        }
      }
      if (this.training.distributed?.master) {
        const masterEnv = this.training.distributed.master.assets.environment;
        if (masterEnv) {
          const selectedEnvironment: EnvironmentAsset | undefined = this.masterAssetsLists.environments.find(
            (environment: EnvironmentAsset) => environment.meta.id === masterEnv,
          );
          if (!selectedEnvironment?.compliance?.compliance) {
            this.training.distributed.master.assets.environment = "";
          }
        }
        const masterCompute = this.training.distributed.master.assets.compute;
        if (masterCompute) {
          const selectedCompute: ComputeAsset | undefined = this.masterAssetsLists.computes.find(
            (compute: ComputeAsset) => compute.meta.id === masterCompute,
          );

          if (!selectedCompute?.compliance?.compliance) {
            this.training.distributed.master.assets.compute = "";
          }
        }
        const masterDataSources = this.training.distributed.master.assets.datasources;
        if (masterDataSources?.length) {
          const nonComplianceDataSources: Set<string> = new Set(
            this.masterAssetsLists.datasources
              .filter((dataSource: DatasourceListResponseEntry) => !dataSource.compliance?.compliance)
              .map((dataSource: DatasourceListResponseEntry) => dataSource.meta.id),
          );
          this.training.distributed.master.assets.datasources = masterDataSources.filter(
            (datasource: AssetIdAndKind) => {
              const isComplied = !nonComplianceDataSources.has(datasource.id);
              if (!isComplied) {
                this.masterDataSourceSectionForceOpen = true;
              }
              return isComplied;
            },
          );
        }
      }
    },
    setCreatedAssets(
      createdEnvironmentId?: string,
      createdComputeResourceId?: string,
      createdDataSourceId?: string,
      kind?: AssetKind,
    ) {
      if (createdEnvironmentId) {
        this.updateEnvironment(createdEnvironmentId);
      }
      if (createdComputeResourceId) {
        this.updateComputeResource(createdComputeResourceId);
      }
      if (createdDataSourceId && kind) {
        this.updateDataSource(createdDataSourceId, kind);
      }
    },
    duplicateWorkerAssetsToMaster(): void {
      if (!this.training.distributed) return;
      this.training.distributed.master = {
        assets: deepCopy(this.training.assets),
        specificEnv: deepCopy(this.training.specificEnv),
      };
    },
    initMasterModel(): void {
      if (!this.training.distributed) return;

      // Setting master editing to be ON if master policy is defined
      if (this.isMasterPolicyDefined) this.training.enableEditingMaster = true;

      if (!this.training.enableEditingMaster || !this.training.distributed.master) {
        this.duplicateWorkerAssetsToMaster();
      }

      // We set the master model (copy of worker by default)
      // but we have to make sure to cleanup any non complied assets, due to a possible difference between worker & master policies
      // if we cleanup any we trigger the form to be editable on master page

      this.cleanupNoneComplianceAssets(true);
    },
    async loadProjectById(projectId: string): Promise<Project | void> {
      try {
        this.selectedProject = await orgUnitService.getProject(projectId);
      } catch (error: unknown) {
        console.error(error);
        this.appStore.setFallback(true);
      }
    },
    async loadPolicy(): Promise<void> {
      if (!this.selectedProject) return;
      const type = this.training.distributed ? PolicyType.Distributed : PolicyType.Training;
      this.trainingPolicy = (await trainingService.getPolicy(type, this.projectId)) as DistributedPolicyV2;
    },
    setNodeAffinity(): void {
      if (!this.training.projectId) {
        this.nodeAffinity = [];
        return;
      }

      this.nodeAffinity = (this.selectedProject?.nodeTypes?.training || [])
        .map((nodeType) => this.selectedProject?.nodeTypes?.names?.[nodeType] || "")
        .filter(Boolean);

      if (this.training.specificEnv && this.nodeAffinity && this.nodeAffinity.length === 1)
        this.training.specificEnv.nodeType = this.nodeAffinity[0];
    },
    setEnvironment(environment: EnvironmentAsset): void {
      const specificEnv: IUIWorkloadEnvSectionModel = environmentSectionUtil.getSpecificEnvFromEnvironment(
        environment,
        this.loadedWorkerItemizeDefaults?.environmentVariables?.instances,
      );
      this.training.assets.environment = environment.meta.id;
      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...specificEnv,
      };
    },
    setEnvironmentDistributed(environment: EnvironmentAsset): void {
      if (!this.training.distributed?.master) return;

      const specificEnv: IUIWorkloadEnvSectionModel = environmentSectionUtil.getSpecificEnvFromEnvironment(
        environment,
        this.loadedMasterItemizeDefaults?.environmentVariables?.instances,
      );
      this.training.distributed.master.assets.environment = environment.meta.id;
      this.training.distributed.master.specificEnv = {
        ...this.training.distributed.master.specificEnv,
        ...specificEnv,
      };
    },
    updateEnvironment(createdEnvironmentId: string): void {
      if (this.wizardLayout.activeStep.value === EDistributedReplicaType.Master) {
        const environment: EnvironmentAsset | undefined = this.masterAssetsLists.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === createdEnvironmentId,
        );
        if (!environment || !environment.compliance?.compliance) return;
        this.setEnvironmentDistributed(environment);
      } else {
        const environment: EnvironmentAsset | undefined = this.assetsLists.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === createdEnvironmentId,
        );
        if (!environment || !environment.compliance?.compliance) return;
        this.setEnvironment(environment);
      }
    },
    updateComputeResource(createdId: string): void {
      if (this.wizardLayout.activeStep.value === EDistributedReplicaType.Master && this.training.distributed?.master) {
        const compute: ComputeAsset | undefined = this.masterAssetsLists.computes.find(
          (c: ComputeAsset) => c.meta.id === createdId,
        );
        if (!compute || !compute.compliance?.compliance) return;
        this.training.distributed.master.assets.compute = createdId;
      } else {
        const compute: ComputeAsset | undefined = this.assetsLists.computes.find(
          (c: ComputeAsset) => c.meta.id === createdId,
        );
        if (!compute || !compute.compliance?.compliance) return;
        this.training.assets.compute = createdId;
      }
    },
    updateDataSource(createdId: string, kind: AssetKind): void {
      if (this.wizardLayout.activeStep.value === EDistributedReplicaType.Master && this.training.distributed?.master) {
        const ds: DatasourceListResponseEntry | undefined = this.masterAssetsLists.datasources.find(
          (d: DatasourceListResponseEntry) => d.meta.id === createdId,
        );
        if (!ds || !ds.compliance?.compliance) return;

        this.training.distributed.master.assets.datasources = this.dataSourceNoDuplicates(
          this.training.distributed.master.assets.datasources,
          createdId,
          kind,
        );
      } else {
        const ds: DatasourceListResponseEntry | undefined = this.assetsLists.datasources.find(
          (d: DatasourceListResponseEntry) => d.meta.id === createdId,
        );
        if (!ds || !ds.compliance?.compliance) return;

        this.training.assets.datasources = this.dataSourceNoDuplicates(
          this.training.assets.datasources,
          createdId,
          kind,
        );
      }
    },
    dataSourceNoDuplicates(
      datasources: Array<AssetIdAndKind> | undefined,
      id: string,
      kind: AssetKind,
    ): Array<AssetIdAndKind> {
      if (!datasources) datasources = [];
      const alreadyExist: AssetIdAndKind | undefined = datasources.find(
        (datasource: AssetIdAndKind) => datasource.id === id,
      );
      if (!alreadyExist) {
        datasources.push({ id, kind });
      }

      return datasources;
    },
    async loadFromExistingTraining(trainingId: string): Promise<void> {
      try {
        const trainingModel: TrainingV2 = await trainingService.getTrainingMergedWithPolicy(
          undefined,
          trainingId,
          undefined,
        );
        this.trainingStore.setLoadedTrainingModelWithDefaults(deepCopy(trainingModel));
        let uiVolumes: Array<IUIVolume> | undefined;
        if (trainingModel.assets.workloadVolumes?.length) {
          const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(trainingModel.assets.workloadVolumes);
          uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
        }
        this.training = workloadUtil.convertWorkloadToWorkloadUI(trainingModel, uiVolumes);

        this.saveTraining(this.training);
      } catch (error: unknown) {
        console.error("failed to get training with id:", trainingId, error);
        this.$q.notify(alertUtil.getError("Failed to load training"));
        this.appStore.setFallback(true);
      }
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      const query = fromSubmit ? { clusterId: this.training.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    cleanupItemizePreSubmission(): void {
      this.training.specificEnv = policyUtil.cleanupItemizeBeforeSubmission(
        this.training.specificEnv,
        this.loadedWorkerPolicyInfo,
      );
      if (this.training.distributed?.master?.specificEnv) {
        this.training.distributed.master.specificEnv = policyUtil.cleanupItemizeBeforeSubmission(
          this.training.distributed?.master?.specificEnv,
          this.loadedMasterPolicyInfo,
        );
      }
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;
        this.training.clusterId = this.clusterUid;
        this.training.namespace = this.selectedProject?.status.namespace || "";

        // cleaning up itemized fields before submission
        this.cleanupItemizePreSubmission();

        this.saveTraining(this.training);
        let workloadVolumes: Array<string> | undefined;
        if (this.training.assets.uiVolumes?.length) {
          workloadVolumes = await dataSourceService.createWorkloadVolumes(
            this.training.name,
            this.training.assets.uiVolumes,
            {
              scope: Scope.Project,
              projectId: this.projectId,
            },
          );
        }

        let masterWorkloadVolumes: Array<string> | undefined;
        if (this.training.distributed?.master?.assets.uiVolumes?.length) {
          masterWorkloadVolumes = await dataSourceService.createWorkloadVolumes(
            this.training.name,
            this.training.distributed.master.assets.uiVolumes,
            {
              scope: Scope.Project,
              projectId: this.projectId,
            },
          );
        }

        const workloadCreationRequestV2: TrainingCreationRequestV2 = workloadUtil.getTrainingCreationRequestV2(
          this.training,
          workloadVolumes,
          masterWorkloadVolumes,
        );
        const trainingAdded = await this.trainingStore.createTrainingV2(workloadCreationRequestV2);

        if (trainingAdded) this.workloadStore.setWorkloadAdded(trainingAdded);
        this.$q.notify(alertUtil.getSuccess(`Training ${this.training.name} created`));
        this.clusterStore.setSelectedClusterById(this.training.clusterId);
        this.redirectToPrevRoute(true);
      } catch (error: unknown) {
        // This is a quick fix when creating workload fails we need to create a pvc
        // with a different name
        this.training.assets.uiVolumes?.forEach(
          (uiVolume: IUIVolume) => (uiVolume.claimName = dataSourceUtil.getRandomClaimName(`${this.training.name}-pvc`)),
        );

        if (this.training.distributed?.master?.assets.uiVolumes) {
          this.training.distributed.master.assets.uiVolumes.forEach(
            (uiVolume: IUIVolume) =>
              (uiVolume.claimName = dataSourceUtil.getRandomClaimName(`${this.training.name}-pvc`)),
          );
        }

        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("training"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.submitting = false;
      }
    },
    saveTraining(training: IUIWorkloadCreation): void {
      this.trainingStore.setTraining(training);
    },
    onBack(): void {
      this.$router.push({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
    },
    updateCreatedPvc(createdPvcId: string, relevantForm: EDistributedReplicaType): void {
      intervalUtil.startInterval(EIntervalLabels.CreatedPvcUpdate, async () => {
        await this.loadDataSources();

        const relevantList = relevantForm === EDistributedReplicaType.Master ? this.masterAssetsLists : this.assetsLists;
        const isComplied = relevantList.datasources.find((ds) => ds.meta.id === createdPvcId)?.compliance?.compliance;

        if (isComplied) {
          intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
        }
      });
    },
    onStepClicked(payload: { step: IRunaiWizardLayoutStep; name: string }): void {
      this.wizardLayout.goToStep(payload.name);
    },
    updateWorkersSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      this.training.assets.environment = envData.environmentId;

      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...envData.specificEnv,
      };
    },
    updateMasterSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      if (!this.training.distributed?.master) return;
      this.training.distributed.master.assets.environment = envData.environmentId;
      this.training.distributed.master.specificEnv = {
        ...this.training.distributed.master.specificEnv,
        ...envData.specificEnv,
      };
    },
    updateWorkersGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...generalModel,
      };
    },
    updateMasterGeneralModel(generalModel: IUIGeneralSectionModel): void {
      if (!this.training.distributed?.master) return;
      this.training.distributed.master.specificEnv = {
        ...this.training.distributed.master.specificEnv,
        ...generalModel,
      };
    },
    onWorkersComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      this.training.assets.compute = computeResourceData.computeResourceId;
      if (this.training.specificEnv) {
        this.training.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.training.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.training.specificEnv.autoScaleData = computeResourceData.autoScaleData;
        this.training.specificEnv.tolerations = computeResourceData.tolerations;
        this.training.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
      if (this.training.distributed) {
        if (computeResourceData.workers && this.training.distributed.numWorkers !== computeResourceData.workers) {
          this.training.specificEnv.backoffLimit = 6 * (computeResourceData.workers + 1); // For distributed training the backoff limit default is 6 * (numWorkers + 1)
        }
        this.training.distributed.numWorkers = computeResourceData.workers || 1;
      }
    },
    onMasterComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      if (!this.training.distributed?.master) return;
      this.training.distributed.master.assets.compute = computeResourceData.computeResourceId;
      if (this.training.distributed.master.specificEnv) {
        this.training.distributed.master.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.training.distributed.master.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.training.distributed.master.specificEnv.autoScaleData = computeResourceData.autoScaleData;
        this.training.distributed.master.specificEnv.tolerations = computeResourceData.tolerations;
        this.training.distributed.master.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
    },
    async onVolumesChanged(uiVolumes: Array<IUIVolume>): Promise<void> {
      let newStorageClass = uiVolumes?.[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.training.assets.uiVolumes?.[0]?.claimInfo?.storageClass;

      this.training.assets = {
        ...this.training.assets,
        uiVolumes,
      };

      if (newStorageClass === oldStorageClass) return;

      if (this.storageClasses.length === 1) {
        newStorageClass = this.storageClasses[0].storageClassName;
      }

      if (!newStorageClass) return;

      if (this.training.assets.uiVolumes?.[0]) {
        this.training.assets.uiVolumes[0].claimInfo = await this.loadStorageClassConfiguration(
          this.clusterUid,
          newStorageClass,
        );
      }
    },
    async onMasterVolumesChanged(uiVolumes: Array<IUIVolume>): Promise<void> {
      if (!this.training.distributed?.master) return;

      let newStorageClass = uiVolumes?.[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.training.distributed.master.assets.uiVolumes?.[0]?.claimInfo?.storageClass;

      this.training.distributed.master.assets.uiVolumes = uiVolumes;

      if (newStorageClass === oldStorageClass) return;

      if (this.storageClasses.length === 1) {
        newStorageClass = this.storageClasses[0].storageClassName;
      }

      if (!newStorageClass) return;

      if (this.training.distributed.master.assets.uiVolumes?.[0]) {
        this.training.distributed.master.assets.uiVolumes[0].claimInfo = await this.loadStorageClassConfiguration(
          this.clusterUid,
          newStorageClass,
        );
      }
    },
    updateWorkersSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      this.training.assets.datasources = dataSources;
    },
    updateMasterSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      if (!this.training.distributed?.master) return;
      this.training.distributed.master.assets.datasources = dataSources;
    },
    onCreateNewEnvironment(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.environment.name,
        query: {
          ...routes.environment.query,
          projectId: this.training.projectId,
          scope: Scope.Project,
          distFramework: this.training.distributed?.distFramework,
          workloadType: this.formType,
          workloadFormName: this.wizardLayout.activeStep.value,
        },
      });
    },
    onCreateNewComputeResource(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.compute.name,
        query: {
          ...routes.compute.query,
          projectId: this.training.projectId,
          scope: Scope.Project,
          workloadFormName: this.wizardLayout.activeStep.value,
        },
      });
    },
    onCreateNewDataSource(selectedDataSourceType: string): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.DISTRIBUTED_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.datasource.pages[selectedDataSourceType],
        query: {
          projectId: this.training.projectId,
          workloadFormName: this.wizardLayout.activeStep.value,
          scope: Scope.Project,
        },
      });
    },
    onNameChanged(name: string) {
      this.training.name = name;
    },
    onToggleEditingChanged(isEnabled: boolean): void {
      this.training.enableEditingMaster = isEnabled;

      if (!isEnabled) {
        this.duplicateWorkerAssetsToMaster();
      }
    },
    handleNextStep(): void {
      this.noMasterStep ? this.onSubmit() : this.onContinueStep();
    },
    async onContinueStep(): Promise<void> {
      this.initMasterModel();
      this.wizardLayout.markStepComplete(EDistributedReplicaType.Worker);
      this.wizardLayout.goToStep(EDistributedReplicaType.Master);
    },
    onBackStep(): void {
      this.wizardLayout.markStepIncomplete(EDistributedReplicaType.Worker);
      this.wizardLayout.goToStep(EDistributedReplicaType.Worker);
      //  in the previous impelentation master model and toggle were reset to origin state, not sure why though.
    },
    getNodePoolsData(
      specificEnv: IUIWorkloadSpecificEnv | undefined,
      nodePoolsResources: Resources[] = [],
    ): IComputeSectionNodePoolData | null {
      if (!specificEnv || !specificEnv.nodePools) {
        return null;
      }

      const allNodePoolsOptions: Array<string> = nodePoolsResources.map(
        (nodePoolResource: Resources) => nodePoolResource.nodePool?.name || "",
      );

      return {
        defaultNodePools: specificEnv.nodePools,
        allNodePoolsOptions,
        nodePoolsListOrigin: ENodePoolsListOrigin.Project,
      };
    },
    initRelevantFormSteps(): void {
      if (!this.noMasterStep) return;
      this.wizardLayout.setSteps(distributedAssetsSteps().slice(0, 1));
    },
  },
  beforeUnmount() {
    this.saveTraining(this.training);
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default DistributedAssetsEdit as typeof DistributedAssetsEdit & DefineComponent;
</script>

<style lang="scss" scoped></style>
