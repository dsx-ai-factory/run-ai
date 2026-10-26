<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-assets-based column items-center q-pt-md">
    <template v-if="isPageReady">
      <workload-form-wrapper
        :form-state="inference"
        :submitting="submitting"
        submit-btn-text="Create Inference"
        cancel-btn-text="Cancel"
        @canceled="onCancel"
        @submit="submitInference"
      >
        <title-and-back-section label="Cluster" :summary="clusterName" back-tooltip="Change cluster" @on-back="onBack" />

        <title-and-back-section
          label="Project"
          :summary="projectSummary"
          back-tooltip="Change project"
          aid="project-section"
          @on-back="onBack"
        />

        <title-and-back-section
          v-if="showInferenceType"
          label="Inference type"
          summary="Custom"
          back-tooltip="Change model"
          @on-back="onBack"
        />

        <workload-name-section
          :entity-type="formType"
          :default-open="!!fromCopyId"
          :name="inference.name"
          @update:name="onNameChanged"
          :project-id="inference.projectId"
          :cluster-id="inference.clusterId"
        />

        <environment-section
          :entity-type="formType"
          aid="environment-section"
          :loading="false"
          :environments="filteredEnvironments"
          :environment-id="inference.assets.environment || ''"
          :specific-env="environmentRunParams"
          @environment-changed="updateSelectedEnvironment"
          @create-new="onCreateNewEnvironment"
          :section-options="environmentSectionOptions"
          :cluster-id="inference.clusterId"
          :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
          :support-config-map-as-env-var="supportConfigMapAsEnvVar"
          :policy-rules="loadedPolicyInfo"
          :itemize-defaults="loadedItemizeDefaults"
          :loaded-specific-run-params-with-defaults="loadedSpecificRunParamsWithDefaults"
          supporting-serving-port
        />

        <compute-resource-section
          aid="compute-resource-section"
          :entity-type="formType"
          :loading="false"
          :compute-resources="computeResources"
          @compute-resource-data-changed="onComputeResourceDataChanged"
          :compute-resource-data="computeResourceData"
          @create-new="onCreateNewComputeResource"
          is-required
          :section-options="computeResourcesSectionOptions"
          :policy-rules="loadedPolicyInfo"
          :policy-defaults="policy?.effective?.defaults"
          :itemize-defaults="loadedItemizeDefaults"
          :support-tolerations="supportTolerations"
        />

        <volume-section
          v-if="isVolumesEnabled"
          aid="volume-section"
          :volumes="inference.assets.uiVolumes || []"
          :storage-classes="storageClasses"
          @update-volumes="onVolumesChanged"
          :policy-rules="loadedPolicyInfo?.rules?.assets?.datasources?.pvc"
          :policy-defaults="loadedAssetsAttributesDefaults?.datasources?.pvc"
          :show-volume-persistency-section="false"
          :loading-storage-class="loadingStorageClass"
        />

        <data-source-section
          :entity-type="formType"
          aid="data-source-section"
          :loading="false"
          :data-sources="dataSources"
          :created-data-source-id="createdDataSourceId"
          :default-opened="!!createdDataSourceId"
          :selected-data-sources="inference.assets.datasources || []"
          @datasource-changed="updateSelectedDataSources"
          @create-new="onCreateNewDataSource"
          :cluster-id="inference.clusterId"
          :policy-rules="loadedPolicyInfo"
        />

        <general-section
          aid="general-section"
          :workload-form-type="formType"
          :general-model="generalSectionModel"
          @general-model-changed="updateGeneralModel"
          :show-backoff-limit="false"
          :policy-rules="generalPolicyRules"
          :itemize-defaults="loadedItemizeDefaults"
          should-auto-validate
        />
      </workload-form-wrapper>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { EnvironmentSection } from "@/components/section/environment-section";
import { ComputeResourceSection } from "@/components/section/compute-resource-section";
import { DataSourceSection } from "@/components/section/data-source-section";
import { GeneralSection, type IUIGeneralSectionModel } from "@/components/section/general-section";
import { VolumeSection } from "@/components/section/volume-section";

// models
import {
  EWorkloadFormType,
  type IWorkloadSupportedRoutes,
  type IUIInferenceCreation,
  EWorkloadTypeSupportedInAssetsService,
} from "@/models/workload.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "@/components/section/environment-section";
import {
  AssetKind,
  Scope,
  type EnvironmentAsset,
  type ComputeAsset,
  type AssetIdAndKind,
  type DatasourceListResponseEntry,
  type InferenceV2,
  type PolicyInfo,
  type EnvironmentVariableOfAsset,
  type AssetsDefaults,
  type InferenceSpecificRunParams,
  type PVCAsset,
} from "@/swagger-models/assets-service-client";
import { SettingKeys } from "@/models/setting.model";
import { EHuggingFaceEnvVariableNames, ENimEnvVariableNames } from "@/models/inference.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { InferencePolicyV2 } from "@/swagger-models/policy-service-client";
import type { IComputeSectionData, IWorkloadComputeSectionOptions } from "@/components/section/compute-resource-section";
import { EIntervalLabels } from "@/models/interval.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { IUIVolume } from "@/models/data-source.model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { UnifiedPolicyInfoPerReplicaTypeRules } from "@/swagger-models/workloads-client";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { environmentSectionUtil } from "@/components/section/environment-section";
import { ErrorAlert } from "@/utils/error-alert.util";
import { workloadCreateFormUtil } from "@/utils/workload-create-form.util";
import { intervalUtil } from "@/utils/interval.util";
import { deepCopy, pick } from "@/utils/common.util/common.util";
import { policyUtil } from "@/utils/policy.util";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";

// routes
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { dataSourceUtil } from "@/utils/data-source.util/data-source.util";
import {
  MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED,
  MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
// composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";

export default defineComponent({
  components: {
    TitleAndBackSection,
    WorkloadFormWrapper,
    WorkloadNameSection,
    EnvironmentSection,
    ComputeResourceSection,
    DataSourceSection,
    VolumeSection,
    GeneralSection,
  },
  provide() {
    return {
      policy: computed(() => this.policy),
    };
  },
  data() {
    return {
      filteredEnvironments: [] as Array<EnvironmentAsset>,
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      environmentStore: useEnvironmentStore(),
      computeResourceStore: useComputeResourceStore(),
      workloadStore: useWorkloadStore(),
      authStore: useAuthStore(),
      settingStore: useSettingStore(),
      dataSourceStore: useDataSourceStore(),
      clusterStore: useClusterStore(),
      formType: EWorkloadFormType.Inference,
      inference: inferenceWorkloadUtil.getEmptyUIInferenceCreation() as IUIInferenceCreation,
      submitting: false as boolean,
      assetsFilter: {} as IAssetsFilter,
      selectedProject: null as Project | null,
      policy: {} as InferencePolicyV2 | null,
      fromCopyId: this.$route.query.fromCopyId?.toString() as string | undefined,
      createdDataSourceId: "" as string,
      environmentSectionOptions: {
        canAddEnvVariable: true,
      } as IWorkloadEnvSectionOptions,
      showInferenceType: false as boolean,
      isLatencyEnabled: true as boolean,
      isAutoScaleEnabled: true as boolean,
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
    const { kind, createdEntityId } = this.$route.query;
    this.inference = this.inferenceStore.inference;
    try {
      if (this.fromCopyId) {
        await this.loadFromExistingInference(this.fromCopyId);
      }

      await this.setDataSourcePolicy();

      const selectedCluster: ClustersMinimalData | undefined = this.clusterStore.clusterList.find(
        (cluster) => cluster.uuid === this.inference.clusterId,
      );

      if (selectedCluster) {
        this.isLatencyEnabled = inferenceWorkloadUtil.isLatencyEnableOnCluster(selectedCluster);
        if (
          this.clusterStore.isClusterVersionSufficient(selectedCluster.uuid, MIN_CLUSTER_VERSION_FOR_AUTO_SCALE_DISABLED)
        ) {
          this.isAutoScaleEnabled = inferenceWorkloadUtil.isAutoScaleEnableOnCluster(selectedCluster);
        }
      } else {
        this.$router.replace({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
      }

      this.assetsFilter = {
        projectId: this.inference.projectId,
        complyToProject: this.inference.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
      };

      this.selectedProject = await orgUnitService.getProject(this.inference.projectId.toString());
      this.updateSpecificEnv();

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

      this.createdDataSourceId = createdEntityId && isDataSource ? createdEntityId.toString() : "";

      await Promise.all([
        this.loadEnvironments(createdEnvironmentId),
        this.loadComputeResources(createdComputeId),
        this.loadStorageClasses(this.clusterUid, [this.inference.assets.uiVolumes?.[0]?.claimInfo?.storageClass]),
        this.loadDataSources(this.createdDataSourceId, kind as AssetKind),
        this.loadSourceValueOptions(false, this.inference.projectId ? String(this.inference.projectId) : undefined),
      ]);

      this.validateEnvironmentVariablesOfAsset(this.inference.specificEnv.environmentVariables || []);

      if (this.createdDataSourceId && kind === AssetKind.Pvc) {
        this.updateCreatedPvc(this.createdDataSourceId);
      }

      this.showInferenceType =
        this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) ||
        this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);

      await this.loadPolicy();
      this.appStore.setPageLoading(false);
    } catch {
      this.$q.notify(alertUtil.getError("Failed to load assets"));
      console.error("Failed to load inference assets");
      this.appStore.setFallback(true);
    }
  },
  computed: {
    allowDeployingGPT2ModelEnabled(): boolean {
      return this.settingStore.isAllowDeployingGPT2ModelEnabled;
    },
    isPageReady(): boolean {
      return !this.appStore.isPageLoading;
    },
    projectId(): number {
      return this.inference.projectId;
    },
    environments(): Array<EnvironmentAsset> {
      return this.environmentStore.environmentList;
    },
    computeResources(): Array<ComputeAsset> {
      return this.computeResourceStore.computeResourcesList;
    },
    dataSources(): Array<DatasourceListResponseEntry> {
      return this.dataSourceStore.dataSourceList;
    },
    clusterUid(): string {
      return this.inference.clusterId;
    },
    clusterName(): string {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === this.inference.clusterId)?.name || "";
    },
    projectSummary(): string {
      return this.selectedProject?.name || "None";
    },
    environmentRunParams(): IUIWorkloadEnvSectionModel {
      return inferenceWorkloadUtil.getEnvironmentRunParamsData(this.inference);
    },
    computeResourceData(): IComputeSectionData {
      return inferenceWorkloadUtil.getInferenceComputeResourceData(this.inference, this.selectedProject?.resources);
    },
    computeResourcesSectionOptions(): IWorkloadComputeSectionOptions {
      return {
        autoScale: true,
        applyPolicyDefaults: !!this.fromCopyId,
        latencyEnabled: this.isLatencyEnabled,
        autoScaleEnabled: this.isAutoScaleEnabled,
      };
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return pick(workloadCreateFormUtil.getGeneralSettingsData(this.inference), "annotations", "labels");
    },
    loadedPolicyInfo(): PolicyInfo | null | undefined {
      return this.inferenceStore.loadedPolicyInfo;
    },
    generalPolicyRules(): UnifiedPolicyInfoPerReplicaTypeRules | null | undefined {
      return this.loadedPolicyInfo?.rules?.specificRunParams as unknown as
        | UnifiedPolicyInfoPerReplicaTypeRules
        | null
        | undefined;
    },
    loadedItemizeDefaults(): ILoadedItemizeDefaults | null | undefined {
      return {
        environmentVariables: this.policy?.effective?.defaults?.environmentVariables,
        annotations: this.policy?.effective?.defaults?.annotations,
        labels: this.policy?.effective?.defaults?.labels,
        tolerations: this.policy?.effective?.defaults?.tolerations,
      };
    },
    loadedSpecificRunParamsWithDefaults(): InferenceSpecificRunParams | null {
      return this.inferenceStore.loadedSpecificRunParamsWithDefaults;
    },
    loadedAssetsAttributesDefaults(): AssetsDefaults | null {
      return this.inferenceStore.loadedAssetsAttributesDefaults;
    },
    isVolumesEnabled(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.inference.clusterId,
        MIN_CLUSTER_VERSION_FOR_INFERENCE_VOLUME,
      );
    },
    supportConfigMapAsEnvVar(): boolean {
      return false;
      // return this.clusterStore.isClusterVersionSufficient(this.clusterUid, MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR);
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.inference.clusterId, MIN_CLUSTER_VERSION_FOR_TOLERATIONS);
    },
  },
  methods: {
    async setDataSourcePolicy(): Promise<void> {
      const mergedInference = await inferenceService.getInferenceMergedWithPolicy(String(this.inference.projectId));
      if (mergedInference.policyInfo?.rules?.assets?.datasources) {
        const policyInfo = {
          rules: { assets: { datasources: mergedInference.policyInfo.rules.assets.datasources } },
        };
        mergedInference.policyInfo = policyInfo;
        delete mergedInference.specificRunParams;
        this.inferenceStore.setLoadedInferenceModelWithDefaults(mergedInference);
      }
    },
    async loadFromExistingInference(inferenceId: string): Promise<void> {
      try {
        const currInference: InferenceV2 = await inferenceService.getInferenceMergedWithPolicy(undefined, inferenceId);
        this.inferenceStore.setLoadedInferenceModelWithDefaults(deepCopy(currInference));

        let uiVolumes: Array<IUIVolume> | undefined;
        if (currInference.assets.workloadVolumes?.length) {
          const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(currInference.assets.workloadVolumes);
          uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
        }

        this.inference = inferenceWorkloadUtil.convertInferenceToWorkloadUI(
          currInference,
          this.inference.inferenceType,
          this.supportingServingPortAccess(currInference.meta.clusterId),
          uiVolumes,
        );

        if (this.inference.specificEnv.environmentVariables) {
          this.inference.specificEnv.environmentVariables = this.inference.specificEnv.environmentVariables.map(
            (env: EnvironmentVariableOfAsset) => {
              const isSensitiveToken =
                env.name === EHuggingFaceEnvVariableNames.HF_TOKEN || env.name === ENimEnvVariableNames.NgcApiKey;
              if (!isSensitiveToken || env.credential) return env;
              return { ...env, value: "" };
            },
          );
        }

        this.saveInference(this.inference);
      } catch {
        console.error("Failed to get inference for copy");
        this.$q.notify(alertUtil.getError("Failed to load inference"));
        this.appStore.setFallback(true);
      }
    },
    async loadEnvironments(createdEnvironmentId: string | undefined): Promise<void> {
      if (!this.selectedProject) return;
      const environmentsFilter = {
        ...this.assetsFilter,
        isInference: true,
      };
      this.filteredEnvironments = await this.environmentStore.loadEnvironments(environmentsFilter);

      if (createdEnvironmentId) {
        this.setEnvironment(createdEnvironmentId);
      }
    },
    async loadComputeResources(createdComputeResourceId: string | undefined): Promise<void> {
      if (!this.selectedProject) return;
      await this.computeResourceStore.loadComputeResources(this.assetsFilter);

      // set the last created data source as selected
      if (createdComputeResourceId) {
        this.setComputeResource(String(createdComputeResourceId));
      }
    },
    async loadDataSources(createdDataSourceId: string | undefined, kind: AssetKind): Promise<void> {
      if (!this.selectedProject) return;
      await this.dataSourceStore.loadDataSources(this.assetsFilter);

      // set the last created data source as selected
      if (createdDataSourceId && kind) {
        this.setDataSource(createdDataSourceId, kind);
      }

      // unselect non compliance data sources (from policy)
      const nonComplianceDataSources: Set<string> = new Set(
        this.dataSources
          .filter((dataSource: DatasourceListResponseEntry) => !dataSource.compliance?.compliance)
          .map((dataSource: DatasourceListResponseEntry) => dataSource.meta.id),
      );
      this.inference.assets.datasources = this.inference.assets.datasources?.filter(
        (datasource: AssetIdAndKind) => !nonComplianceDataSources.has(datasource.id),
      );
    },
    updateSpecificEnv(): void {
      this.inference.specificEnv = {
        ...this.inference.specificEnv,
        nodePools: this.inference.specificEnv.nodePools || this.selectedProject?.effective?.defaultNodePools,
      };

      this.saveInference(this.inference);
    },
    setEnvironment(environmentId: string): void {
      const environment = this.filteredEnvironments?.find((environment) => environment.meta.id === environmentId);
      if (!environment || !environment.compliance?.compliance) return;

      const specificEnv: IUIWorkloadEnvSectionModel = environmentSectionUtil.getSpecificEnvFromEnvironment(environment);
      this.inference.assets.environment = environmentId;
      this.inference.specificEnv = { ...this.inference.specificEnv, ...specificEnv };
    },
    setComputeResource(id: string): void {
      this.inference.assets.compute = id;
    },
    setDataSource(id: string, kind: AssetKind): void {
      if (!this.inference.assets.datasources) this.inference.assets.datasources = [];
      const alreadyExist: AssetIdAndKind | undefined = this.inference.assets.datasources.find(
        (datasource: AssetIdAndKind) => datasource.id === id,
      );
      if (!alreadyExist) {
        this.inference.assets.datasources.push({ id, kind });
      }
    },
    async submitInference(): Promise<void> {
      try {
        this.submitting = true;
        this.inference.clusterId = this.clusterUid;
        this.inference.namespace = this.selectedProject?.status.namespace || "";

        // cleaning up itemized fields before submission
        this.cleanupItemizePreSubmission();

        this.saveInference(this.inference);
        await this.inferenceStore.createInference(this.inference);
        this.$q.notify(alertUtil.getSuccess(`Inference ${this.inference.name} created`));
        this.clusterStore.setSelectedClusterById(this.inference.clusterId);
        this.redirectToPrevRoute(true);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("inference"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.submitting = false;
      }
    },
    saveInference(inference: IUIInferenceCreation): void {
      this.inferenceStore.setInference(inference);
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      const query = fromSubmit ? { clusterId: this.inference.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    onBack(): void {
      this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
    },
    onWorkloadChanged(inference: IUIInferenceCreation): void {
      this.inference = inference;
      this.saveInference(inference);
    },
    updateCreatedPvc(createdPvcId: string): void {
      intervalUtil.startInterval(EIntervalLabels.CreatedPvcUpdate, async () => {
        await this.loadDataSources(createdPvcId, AssetKind.Pvc);
        if (this.dataSources.find((ds) => ds.meta.id === createdPvcId)?.compliance?.compliance) {
          intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
        }
      });
    },
    async loadPolicy(): Promise<void> {
      this.policy = await inferenceService.getPolicy(this.projectId);
    },
    onNameChanged(name: string) {
      this.inference.name = name;
    },
    updateSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      this.inference.assets.environment = envData.environmentId;
      this.inference.specificEnv = {
        ...this.inference.specificEnv,
        ...envData.specificEnv,
      };
    },
    onCreateNewEnvironment(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT,
        INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.environment.name,
        query: {
          ...routes.environment.query,
          projectId: this.inference.projectId,
          scope: Scope.Project,
          workloadType: this.formType,
        },
      });
    },
    updateSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      this.inference.assets.datasources = dataSources;
    },
    onCreateNewComputeResource(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT,
        INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.compute.name,
        query: {
          ...routes.compute.query,
          projectId: this.inference.projectId,
          scope: Scope.Project,
        },
      });
    },
    onComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      this.inference.assets.compute = computeResourceData.computeResourceId;
      if (this.inference.specificEnv) {
        this.inference.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.inference.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.inference.specificEnv.autoScaleData = computeResourceData.autoScaleData;
        this.inference.specificEnv.tolerations = computeResourceData.tolerations;
        this.inference.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
    },
    onCreateNewDataSource(selectedDataSourceType: string): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        INFERENCE_ROUTE_NAMES.INFERENCE_ASSETS_EDIT,
        INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.datasource.pages[selectedDataSourceType],
        query: { projectId: this.inference.projectId, scope: Scope.Project },
      });
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.inference.specificEnv = {
        ...this.inference.specificEnv,
        ...generalModel,
      };
    },
    supportingServingPortAccess(clusterId: string): boolean {
      return this.inferenceStore.supportingServingPortAccess(clusterId);
    },
    cleanupItemizePreSubmission(): void {
      this.inference.specificEnv = policyUtil.cleanupItemizeBeforeSubmission(
        this.inference.specificEnv,
        this.loadedPolicyInfo,
      );
    },
    async onVolumesChanged(uiVolumes: Array<IUIVolume>): Promise<void> {
      let newStorageClass = uiVolumes?.[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.inference.assets.uiVolumes?.[0]?.claimInfo?.storageClass;

      this.inference.assets = {
        ...this.inference.assets,
        uiVolumes,
      };

      // If the storage class hasn't changed there is no need to do any changes
      if (newStorageClass === oldStorageClass) return;

      if (this.storageClasses.length === 1) {
        newStorageClass = this.storageClasses[0].storageClassName;
      }

      if (!newStorageClass) return;

      if (this.inference.assets.uiVolumes?.[0]) {
        this.inference.assets.uiVolumes[0].claimInfo = await this.loadStorageClassConfiguration(
          this.clusterUid,
          newStorageClass,
        );
      }
    },
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
  },
});
</script>
