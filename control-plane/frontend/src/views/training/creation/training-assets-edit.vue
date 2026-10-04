<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="training-assets-edit column items-center q-pt-md">
    <template v-if="isPageReady && selectedProject">
      <section class="column items-center">
        <workload-form-wrapper
          :form-state="training"
          :submitting="submitting"
          submit-btn-text="Create Training"
          cancel-btn-text="Cancel"
          @leave-page="$emit('leave-page', $event)"
          @canceled="onCancel"
          @submit="onSubmit"
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
            :specific-env="environmentRunParams"
            @environment-changed="updateSelectedEnvironment"
            @create-new="onCreateNewEnvironment"
            :section-options="environmentSectionOptions"
            :cluster-id="training.clusterId"
            :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
            :support-config-map-as-env-var="supportConfigMapAsEnvVar"
            :policy-rules="loadedPolicyInfo"
            :itemize-defaults="loadedItemizeDefaults"
            :loaded-specific-run-params-with-defaults="loadedSpecificRunParamsWithDefaults"
          />

          <compute-resource-section
            aid="compute-resource-section"
            :entity-type="formType"
            :loading="false"
            :compute-resources="assetsLists.computes"
            :node-affinity="nodeAffinity"
            @compute-resource-data-changed="onComputeResourceDataChanged"
            :compute-resource-data="computeResourceData"
            @create-new="onCreateNewComputeResource"
            is-required
            :section-options="computeResourcesSectionOptions"
            :policy-rules="loadedPolicyInfo"
            :itemize-defaults="loadedItemizeDefaults"
            :support-tolerations="supportTolerations"
          />

          <volume-section
            aid="volume-section"
            :volumes="training.assets.uiVolumes || []"
            :storage-classes="storageClasses"
            @update-volumes="onVolumesChanged"
            :policy-rules="loadedPolicyInfo?.rules?.assets?.datasources?.pvc"
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
            @datasource-changed="updateSelectedDataSources"
            @create-new="onCreateNewDataSource"
            :cluster-id="training.clusterId"
            :policy-rules="loadedPolicyInfo"
          />

          <general-section
            aid="general-section"
            :workload-form-type="formType"
            :general-model="generalSectionModel"
            @general-model-changed="updateGeneralModel"
            :show-backoff-limit="supportBackoffLimit"
            :show-runs-and-parallelism="supportRunsAndParallelism"
            :policy-rules="loadedPolicyInfo?.rules?.specificRunParams"
            :itemize-defaults="loadedItemizeDefaults"
            :show-termination-grace-period="isClusterSupportTerminationGracePeriod"
          />
        </workload-form-wrapper>
      </section>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, computed, type DefineComponent } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { EnvironmentSection } from "@/components/section/environment-section";
import { ComputeResourceSection } from "@/components/section/compute-resource-section";
import { DataSourceSection } from "@/components/section/data-source-section";
import { VolumeSection } from "@/components/section/volume-section";
import { GeneralSection, type IUIGeneralSectionModel } from "@/components/section/general-section";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";

// models
import {
  EWorkloadFormType,
  type IWorkloadSupportedRoutes,
  type IUIWorkloadCreation,
  EWorkloadTypeSupportedInAssetsService,
} from "@/models/workload.model";
import {
  Scope,
  AssetKind,
  UidGidSource,
  type EnvironmentAsset,
  type ComputeAsset,
  type AssetIdAndKind,
  type PVCAsset,
  type DatasourceListResponseEntry,
  type TrainingCreationRequestV2,
  type TrainingV2,
  type PolicyInfo,
} from "@/swagger-models/assets-service-client";
import { PolicyType, type TrainingPolicyV2 } from "@/swagger-models/policy-service-client";
import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "@/components/section/environment-section";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { IUIVolume } from "@/models/data-source.model";
import type { IComputeSectionData, IWorkloadComputeSectionOptions } from "@/components/section/compute-resource-section";
import type { IAssetsFilter } from "@/models/filter.model";

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
import { alertUtil } from "@/utils/alert.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { environmentSectionUtil } from "@/components/section/environment-section";
import { workloadCreateFormUtil } from "@/utils/workload-create-form.util";
import { intervalUtil } from "@/utils/interval.util";
import { policyUtil } from "@/utils/policy.util";
import { deepCopy } from "@/utils/common.util";

// routes
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";

// common
import { ErrorAlert } from "@/utils/error-alert.util";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import {
  MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
  MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import { EIntervalLabels } from "@/models/interval.model";
// Composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";

export interface ITrainingFormAssetsLists {
  environments: EnvironmentAsset[];
  computes: ComputeAsset[];
  datasources: DatasourceListResponseEntry[];
}

const TrainingAssetsEdit = defineComponent({
  components: {
    WorkloadFormWrapper,
    WorkloadNameSection,
    EnvironmentSection,
    ComputeResourceSection,
    DataSourceSection,
    VolumeSection,
    GeneralSection,
    TitleAndBackSection,
  },
  provide() {
    return {
      policy: computed(() => this.trainingPolicy),
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
      trainingPolicy: null as TrainingPolicyV2 | null,
      submitting: false as boolean,
      envSectionOptions: {
        canAddEnvVariable: true,
      } as IWorkloadEnvSectionOptions,
      computeSectionOptions: {
        applyPolicyDefaults: false,
      } as IWorkloadComputeSectionOptions,
      createdDataSourceId: "" as string,
      fromCopyId: "" as string,
      assetsFilter: {} as IAssetsFilter,
      assetsLists: {
        computes: [],
        datasources: [],
        environments: [],
      } as ITrainingFormAssetsLists,
    };
  },
  setup() {
    const { environmentVariablesSourceValueOptions, loadSourceValueOptions, validateEnvironmentVariablesOfAsset } =
      useEnvironmentVariablesSection();

    const { storageClasses, loadingStorageClass, loadStorageClasses, loadStorageClassConfiguration } =
      useStorageClassConfiguration("workload");

    return {
      environmentVariablesSourceValueOptions,
      loadSourceValueOptions,
      validateEnvironmentVariablesOfAsset,

      storageClasses,
      loadingStorageClass,
      loadStorageClasses,
      loadStorageClassConfiguration,
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

      this.initFilterProperties();

      await this.loadProjectById(this.training.projectId?.toString());

      if (this.selectedProject && !fromCopyId)
        this.training.specificEnv.nodePools = this.selectedProject.effective?.defaultNodePools || [];

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
        this.loadStorageClasses(this.clusterUid, [this.training.assets.uiVolumes?.[0]?.claimInfo?.storageClass]),
        this.loadSourceValueOptions(false, this.training.projectId ? String(this.training.projectId) : undefined),
      ]);

      this.setCreatedAssets(createdEnvironmentId, createdComputeId, createdDataSourceId, kind as AssetKind);
      this.validateEnvironmentVariablesOfAsset(this.training.specificEnv.environmentVariables || []);

      if (createdDataSourceId && kind === AssetKind.Pvc) {
        this.updateCreatedPvc(createdDataSourceId);
      }

      this.computeSectionOptions.applyPolicyDefaults = !!fromCopyId || !!this.trainingStore.templateId;

      this.cleanupNoneComplianceAssets();
    } catch (error: unknown) {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error(error);
      this.appStore.setFallback(true);
    } finally {
      this.appStore.setPageLoading(false);
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
    clusterName(): string {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === this.training.clusterId)?.name || "";
    },
    environmentRunParams(): IUIWorkloadEnvSectionModel {
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
    computeResourceData(): IComputeSectionData {
      return workloadCreateFormUtil.getComputeResourceData(this.training, this.selectedProject?.resources);
    },
    generalSectionModel(): IUIGeneralSectionModel {
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
          ? this.training.specificEnv?.backoffLimit
          : null,
        parallelism: this.training.specificEnv?.parallelism || null,
        completions: this.training.specificEnv?.completions || null,
        terminationGracePeriodSeconds: this.training.specificEnv?.terminationGracePeriodSeconds || null,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(this.clusterUid);
    },
    supportRunsAndParallelism(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.clusterUid,
        MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
      );
    },
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.clusterUid,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    environmentSectionOptions(): IWorkloadEnvSectionOptions {
      return this.envSectionOptions;
    },
    computeResourcesSectionOptions(): IWorkloadComputeSectionOptions {
      return {
        ...this.computeSectionOptions,
        allowAnyEffect: this.clusterStore.isClusterVersionSufficient(
          this.clusterUid,
          MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
        ),
      };
    },
    loadedPolicyInfo(): PolicyInfo | null | undefined {
      return this.trainingStore.loadedPolicyInfo;
    },
    loadedItemizeDefaults(): ILoadedItemizeDefaults | null | undefined {
      return {
        environmentVariables: this.trainingPolicy?.effective?.defaults?.environmentVariables,
        annotations: this.trainingPolicy?.effective?.defaults?.annotations,
        labels: this.trainingPolicy?.effective?.defaults?.labels,
        tolerations: this.trainingPolicy?.effective?.defaults?.tolerations,
      };
    },
    loadedSpecificRunParamsWithDefaults() {
      return this.trainingStore.loadedSpecificRunParamsWithDefaults;
    },
    loadedAssetsAttributesDefaults() {
      return this.trainingStore.loadedAssetsAttributesDefaults;
    },
    projectSummary(): string {
      return this.selectedProject?.name || "None";
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
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Training,
        isTraining: true,
      };
    },
    async loadEnvironments(): Promise<void> {
      if (!this.selectedProject) return;
      this.assetsLists.environments = await this.environmentStore.loadEnvironments(this.assetsFilter);
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
      this.assetsLists.computes = await this.computeResourceStore.loadComputeResources(this.assetsFilter);
    },
    async loadDataSources(): Promise<void> {
      if (!this.selectedProject) return;
      this.assetsLists.datasources = await this.dataSourceStore.loadDataSources(this.assetsFilter);
    },
    cleanupNoneComplianceAssets(): void {
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
      this.trainingPolicy = (await trainingService.getPolicy(PolicyType.Training, this.projectId)) as TrainingPolicyV2;
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
        this.loadedItemizeDefaults?.environmentVariables?.instances,
      );
      this.training.assets.environment = environment.meta.id;
      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...specificEnv,
      };
    },
    updateEnvironment(createdEnvironmentId: string): void {
      const environment: EnvironmentAsset | undefined = this.assetsLists.environments.find(
        (environment: EnvironmentAsset) => environment.meta.id === createdEnvironmentId,
      );
      if (!environment || !environment.compliance?.compliance) return;
      this.setEnvironment(environment);
    },
    updateComputeResource(createdId: string): void {
      this.training.assets.compute = createdId;
    },
    updateDataSource(createdId: string, kind: AssetKind): void {
      this.training.assets.datasources = this.dataSourceNoDuplicates(this.training.assets.datasources, createdId, kind);
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
        const trainingModel: TrainingV2 = await trainingService.getTrainingMergedWithPolicy(undefined, trainingId);
        this.trainingStore.setLoadedTrainingModelWithDefaults(deepCopy(trainingModel));

        let uiVolumes: Array<IUIVolume> | undefined;
        if (trainingModel.assets.workloadVolumes?.length) {
          const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(trainingModel.assets.workloadVolumes);
          uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
        }
        this.training = workloadUtil.convertWorkloadToWorkloadUI(trainingModel, uiVolumes);

        this.setFormDefaultsIfNeeded();
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
        this.loadedPolicyInfo,
      );
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

        const workloadCreationRequestV2: TrainingCreationRequestV2 = workloadUtil.getTrainingCreationRequestV2(
          this.training,
          workloadVolumes,
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
    updateCreatedPvc(createdPvcId: string): void {
      intervalUtil.startInterval(EIntervalLabels.CreatedPvcUpdate, async () => {
        await this.loadDataSources();
        if (this.assetsLists.datasources.find((ds) => ds.meta.id === createdPvcId)?.compliance?.compliance) {
          intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
        }
      });
    },
    updateSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      this.training.assets.environment = envData.environmentId;

      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...envData.specificEnv,
      };
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.training.specificEnv = {
        ...this.training.specificEnv,
        ...generalModel,
      };
    },
    onComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      this.training.assets.compute = computeResourceData.computeResourceId;
      if (this.training.specificEnv) {
        this.training.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.training.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.training.specificEnv.autoScaleData = computeResourceData.autoScaleData;
        this.training.specificEnv.tolerations = computeResourceData.tolerations;
        this.training.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
    },
    async onVolumesChanged(uiVolumes: Array<IUIVolume>): Promise<void> {
      let newStorageClass = uiVolumes?.[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.training.assets.uiVolumes?.[0]?.claimInfo?.storageClass;

      this.training.assets = {
        ...this.training.assets,
        uiVolumes,
      };

      // If the storage class hasn't changed there is no need to do any changes
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
    updateSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      this.training.assets.datasources = dataSources;
    },
    onCreateNewEnvironment(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.environment.name,
        query: {
          ...routes.environment.query,
          projectId: this.training.projectId,
          scope: Scope.Project,
          workloadType: this.formType,
        },
      });
    },
    onCreateNewComputeResource(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.compute.name,
        query: {
          ...routes.compute.query,
          projectId: this.training.projectId,
          scope: Scope.Project,
        },
      });
    },
    onCreateNewDataSource(selectedDataSourceType: string): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        TRAINING_ROUTE_NAMES.TRAINING_ASSETS_EDIT,
        TRAINING_ROUTE_NAMES.TRAINING_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.datasource.pages[selectedDataSourceType],
        query: { projectId: this.training.projectId, scope: Scope.Project },
      });
    },
    onNameChanged(name: string) {
      this.training.name = name;
    },
    // Setting form defaults for specific fields as defined by product
    setFormDefaultsIfNeeded(): void {
      if (!this.training.specificEnv?.completions && this.supportRunsAndParallelism) {
        this.training.specificEnv.completions = 1;
        this.training.specificEnv.parallelism = null;
      }
    },
  },
  beforeUnmount() {
    this.saveTraining(this.training);
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
  },
});

export default TrainingAssetsEdit as typeof TrainingAssetsEdit & DefineComponent;
</script>

<style lang="scss" scoped>
.training-assets-edit {
  .training-assets-edit-actions {
    width: 100%;
    display: flex;
    align-items: center;
  }
}
</style>
