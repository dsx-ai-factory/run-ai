<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workspace-template-edit column items-center q-pt-md">
    <template v-if="isPageReady && selectedProject">
      <section class="column items-center">
        <workload-form-wrapper
          :form-state="workspace"
          :submitting="submitting"
          submit-btn-text="Create Workspace"
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
            :name="workspace.name"
            @update:name="onNameChanged"
            :project-id="workspace.projectId"
            :cluster-id="workspace.clusterId"
          />

          <environment-section
            :entity-type="formType"
            aid="environment-section"
            :loading="false"
            :environments="environments"
            :environment-id="workspace.assets.environment || ''"
            :specific-env="environmentRunParams"
            @environment-changed="updateSelectedEnvironment"
            @create-new="onCreateNewEnvironment"
            :section-options="environmentSectionOptions"
            :cluster-id="workspace.clusterId"
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
            :compute-resources="computeResources"
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
            :volumes="workspace.assets.uiVolumes || []"
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
            :data-sources="dataSources"
            :created-data-source-id="createdDataSourceId"
            :default-opened="!!createdDataSourceId"
            :selected-data-sources="workspace.assets.datasources || []"
            @datasource-changed="updateSelectedDataSources"
            @create-new="onCreateNewDataSource"
            :cluster-id="workspace.clusterId"
            :policy-rules="loadedPolicyInfo"
          />

          <general-section
            aid="general-section"
            :workload-form-type="formType"
            :general-model="generalSectionModel"
            @general-model-changed="updateGeneralModel"
            :show-backoff-limit="supportBackoffLimit"
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
  EWorkloadTypeSupportedInAssetsService,
  type IWorkloadSupportedRoutes,
  type IUIWorkloadCreation,
} from "@/models/workload.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { WorkspacePolicyV2 } from "@/swagger-models/policy-service-client";

import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "@/components/section/environment-section";
import type { IComputeSectionData, IWorkloadComputeSectionOptions } from "@/components/section/compute-resource-section";
import {
  Scope,
  AssetKind,
  UidGidSource,
  type PVCAsset,
  type EnvironmentAsset,
  type ComputeAsset,
  type AssetIdAndKind,
  type DatasourceListResponseEntry,
  type WorkspaceCreationRequestV2,
  type WorkspaceV2,
  type PolicyInfo,
} from "@/swagger-models/assets-service-client";
import type { IUIVolume } from "@/models/data-source.model";
import { EIntervalLabels } from "@/models/interval.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { Project } from "@/swagger-models/org-unit-service-client";

// stores
import { useAppStore } from "@/stores/app.store";
import { useWorkspaceStore } from "@/stores/workspace.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";

// services
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { workspaceService } from "@/services/control-plane/workspace.service/workspace.service";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// utils
import { alertUtil } from "@/utils/alert.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { dataSourceUtil } from "@/utils/data-source.util";
import { environmentSectionUtil } from "@/components/section/environment-section";
import { ErrorAlert } from "@/utils/error-alert.util";
import { workloadCreateFormUtil } from "@/utils/workload-create-form.util";
import { intervalUtil } from "@/utils/interval.util";
import { policyUtil } from "@/utils/policy.util";
import { deepCopy } from "@/utils/common.util";

// routes
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import {
  MIN_CLUSTER_VERSION_FOR_ANY_EFFECT,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
// composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";

const WorkspaceAssetsEdit = defineComponent({
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
      policy: computed(() => this.workspacePolicy),
    };
  },
  data() {
    return {
      appStore: useAppStore(),
      workspaceStore: useWorkspaceStore(),
      environmentStore: useEnvironmentStore(),
      computeResourceStore: useComputeResourceStore(),
      workloadStore: useWorkloadStore(),
      authStore: useAuthStore(),
      nodeAffinity: [] as string[],
      selectedProject: null as Project | null,
      dataSourceStore: useDataSourceStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      formType: EWorkloadFormType.Workspace,
      workspace: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,
      workspacePolicy: null as WorkspacePolicyV2 | null,
      submitting: false as boolean,
      envSectionOptions: {
        canAddEnvVariable: true,
      } as IWorkloadEnvSectionOptions,
      computeSectionOptions: {
        applyPolicyDefaults: false,
      } as IWorkloadComputeSectionOptions,
      assetsFilter: {} as IAssetsFilter,
      createdDataSourceId: "" as string,
      fromCopyId: "" as string,
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
    this.workspace = this.workspaceStore.workspace;
    try {
      if (fromCopyId) {
        await this.loadFromExistingWorkspace(fromCopyId.toString());
        this.fromCopyId = String(fromCopyId);
      }
      this.assetsFilter = {
        projectId: this.workspace.projectId,
        complyToProject: this.workspace.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Workspace,
      };

      await this.loadProjectById(this.workspace.projectId?.toString());

      if (this.selectedProject && !fromCopyId)
        this.workspace.specificEnv.nodePools = this.selectedProject.effective?.defaultNodePools || [];

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
        this.loadStorageClasses(this.clusterUid, [this.workspace.assets.uiVolumes?.[0]?.claimInfo?.storageClass]),
        this.loadSourceValueOptions(false, this.workspace.projectId ? String(this.workspace.projectId) : undefined),
      ]);

      this.setCreatedAssets(createdEnvironmentId, createdComputeId, createdDataSourceId, kind as AssetKind);
      this.validateEnvironmentVariablesOfAsset(this.workspace.specificEnv.environmentVariables || []);

      if (createdDataSourceId && kind === AssetKind.Pvc) {
        this.updateCreatedPvc(createdDataSourceId);
      }

      this.computeSectionOptions.applyPolicyDefaults = !!fromCopyId || !!this.workspaceStore.templateId;

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
      return this.workspace.projectId;
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
      return this.workspace.clusterId;
    },
    clusterName(): string {
      return this.clusterStore.clusterList.find((cluster) => cluster.uuid === this.workspace.clusterId)?.name || "";
    },
    environmentRunParams(): IUIWorkloadEnvSectionModel {
      const specificEnv = this.workspace.specificEnv;
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
      return workloadCreateFormUtil.getComputeResourceData(this.workspace, this.selectedProject?.resources);
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        allowOverQuota: this.workspace.specificEnv?.allowOverQuota || false,
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.workspace.specificEnv?.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.workspace.specificEnv?.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: this.workspace.specificEnv?.annotations,
        labels: this.workspace.specificEnv?.labels,
        backoffLimit: Number.isInteger(this.workspace.specificEnv?.backoffLimit)
          ? this.workspace.specificEnv?.backoffLimit
          : null,
        terminationGracePeriodSeconds: this.workspace.specificEnv?.terminationGracePeriodSeconds || null,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(this.clusterUid);
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
    isClusterSupportTerminationGracePeriod(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.clusterUid,
        MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
      );
    },
    loadedPolicyInfo(): PolicyInfo | null | undefined {
      return this.workspaceStore.loadedPolicyInfo;
    },
    loadedItemizeDefaults(): ILoadedItemizeDefaults | null | undefined {
      return {
        environmentVariables: this.workspacePolicy?.effective?.defaults?.environmentVariables,
        annotations: this.workspacePolicy?.effective?.defaults?.annotations,
        labels: this.workspacePolicy?.effective?.defaults?.labels,
        tolerations: this.workspacePolicy?.effective?.defaults?.tolerations,
      };
    },
    loadedSpecificRunParamsWithDefaults() {
      return this.workspaceStore.loadedSpecificRunParamsWithDefaults;
    },
    loadedAssetsAttributesDefaults() {
      return this.workspaceStore.loadedAssetsAttributesDefaults;
    },
    projectSummary(): string {
      return this.selectedProject?.name || "None";
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterUid, MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR);
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.workspace.clusterId, MIN_CLUSTER_VERSION_FOR_TOLERATIONS);
    },
  },
  methods: {
    async loadFromExistingWorkspace(workspaceId: string): Promise<void> {
      try {
        const workspaceModel: WorkspaceV2 = await workspaceService.getWorkspaceMergedWithPolicy(undefined, workspaceId);
        this.workspaceStore.setLoadedWorkspaceModelWithDefaults(deepCopy(workspaceModel));

        let uiVolumes: Array<IUIVolume> | undefined;
        if (workspaceModel.assets.workloadVolumes?.length) {
          const pvcs: Array<PVCAsset> = await dataSourceService.loadPVCAssets(workspaceModel.assets.workloadVolumes);
          uiVolumes = dataSourceUtil.mapPvcsToUiVolumes(pvcs);
        }
        this.workspace = workloadUtil.convertWorkloadToWorkloadUI(workspaceModel, uiVolumes);
        this.saveWorkspace(this.workspace);
      } catch (error: unknown) {
        console.error("failed to get workspace with id:", workspaceId, error);
        this.$q.notify(alertUtil.getError("Failed to load workspace"));
        this.appStore.setFallback(true);
      }
    },
    async loadEnvironments(): Promise<void> {
      if (!this.selectedProject) return;
      await this.environmentStore.loadEnvironments(this.assetsFilter);
      // Set security fields according to the environment settings
      this.setEnvironmentSecurityFields();
    },
    setEnvironmentSecurityFields(): void {
      if (this.workspace.assets.environment) {
        const selectedEnvironment: EnvironmentAsset | undefined = this.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === this.workspace.assets.environment,
        );
        if (selectedEnvironment) {
          const { uidGidSource } = selectedEnvironment.spec;
          if (uidGidSource === UidGidSource.FromIdpToken) {
            this.workspace.specificEnv.runAsGid = this.authStore.getGID;
            this.workspace.specificEnv.runAsUid = this.authStore.getUID;
            this.workspace.specificEnv.supplementalGroups = this.authStore.getSupplementaryGroups;
            this.saveWorkspace(this.workspace);
          }
        }
      }
    },
    async loadComputeResources(): Promise<void> {
      if (!this.selectedProject) return;
      await this.computeResourceStore.loadComputeResources(this.assetsFilter);
    },
    async loadDataSources(): Promise<void> {
      if (!this.selectedProject) return;
      await this.dataSourceStore.loadDataSources(this.assetsFilter);
    },
    setNodeAffinity(): void {
      if (!this.workspace.projectId) {
        this.nodeAffinity = [];
        return;
      }

      this.nodeAffinity = (this.selectedProject?.nodeTypes?.workspace || [])
        .map((nodeType) => this.selectedProject?.nodeTypes?.names?.[nodeType] || "")
        .filter(Boolean);

      if (this.workspace.specificEnv && this.nodeAffinity && this.nodeAffinity.length === 1)
        this.workspace.specificEnv.nodeType = this.nodeAffinity[0];
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
      this.workspacePolicy = await workspaceService.getPolicy(this.projectId);
    },
    setCreatedAssets(
      createdEnvironmentId?: string,
      createdComputeResourceId?: string,
      createdDataSourceId?: string,
      kind?: AssetKind,
    ) {
      if (createdEnvironmentId) {
        this.setEnvironment(createdEnvironmentId);
      }
      if (createdComputeResourceId) {
        this.setComputeResource(createdComputeResourceId);
      }
      if (createdDataSourceId && kind) {
        this.setDataSource(createdDataSourceId, kind);
      }
    },
    setEnvironment(environmentId: string): void {
      const environment = this.environments?.find((environment) => environment.meta.id === environmentId);
      if (!environment || !environment.compliance?.compliance) return;

      const specificEnv: IUIWorkloadEnvSectionModel = environmentSectionUtil.getSpecificEnvFromEnvironment(
        environment,
        this.loadedItemizeDefaults?.environmentVariables?.instances,
      );
      this.workspace.assets.environment = environmentId;
      this.workspace.specificEnv = { ...this.workspace.specificEnv, ...specificEnv };
    },
    setComputeResource(id: string): void {
      this.workspace.assets.compute = id;
    },
    setDataSource(createdId: string, kind: AssetKind): void {
      this.workspace.assets.datasources = this.dataSourceNoDuplicates(
        this.workspace.assets.datasources,
        createdId,
        kind,
      );
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
    cleanupItemizePreSubmission(): void {
      this.workspace.specificEnv = policyUtil.cleanupItemizeBeforeSubmission(
        this.workspace.specificEnv,
        this.loadedPolicyInfo,
      );
    },
    async onSubmit(): Promise<void> {
      try {
        this.submitting = true;
        this.workspace.clusterId = this.clusterUid;
        this.workspace.namespace = this.selectedProject?.status.namespace || "";

        // cleaning up itemized fields before submission
        this.cleanupItemizePreSubmission();

        this.saveWorkspace(this.workspace);
        let workloadVolumes: Array<string> | undefined;
        if (this.workspace.assets.uiVolumes?.length) {
          workloadVolumes = await dataSourceService.createWorkloadVolumes(
            this.workspace.name,
            this.workspace.assets.uiVolumes,
            {
              scope: Scope.Project,
              projectId: this.projectId,
            },
          );
        }

        const workloadCreationRequestV2: WorkspaceCreationRequestV2 = workloadUtil.getWorkspaceCreationRequestV2(
          this.workspace,
          workloadVolumes,
        );

        const workspaceAdded = await this.workspaceStore.createWorkspaceV2(workloadCreationRequestV2);

        if (workspaceAdded) this.workloadStore.setWorkloadAdded(workspaceAdded);
        this.$q.notify(alertUtil.getSuccess(`Workspace ${this.workspace.name} created`));
        this.clusterStore.setSelectedClusterById(this.workspace.clusterId);
        this.redirectToPrevRoute(true);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedCreateMessage("workspace"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.submitting = false;
      }
    },
    saveWorkspace(workspace: IUIWorkloadCreation): void {
      this.workspaceStore.setWorkspace(workspace);
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.redirectToPrevRoute();
      }
    },
    redirectToPrevRoute(fromSubmit?: boolean): void {
      const query = fromSubmit ? { clusterId: this.workspace.clusterId } : undefined;
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX, query });
    },
    onBack(): void {
      this.$router.push({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW });
    },
    updateCreatedPvc(createdPvcId: string): void {
      intervalUtil.startInterval(EIntervalLabels.CreatedPvcUpdate, async () => {
        await this.loadDataSources();
        if (this.dataSources.find((ds) => ds.meta.id === createdPvcId)?.compliance?.compliance) {
          intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
        }
      });
    },
    updateSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      this.workspace.assets.environment = envData.environmentId;
      this.workspace.specificEnv = {
        ...this.workspace.specificEnv,
        ...envData.specificEnv,
      };
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.workspace.specificEnv = {
        ...this.workspace.specificEnv,
        ...generalModel,
      };
    },
    onComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      this.workspace.assets.compute = computeResourceData.computeResourceId;
      if (this.workspace.specificEnv) {
        this.workspace.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.workspace.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.workspace.specificEnv.autoScaleData = computeResourceData.autoScaleData;
        this.workspace.specificEnv.tolerations = computeResourceData.tolerations;
        this.workspace.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
    },
    async onVolumesChanged(uiVolumes: Array<IUIVolume>): Promise<void> {
      let newStorageClass = uiVolumes?.[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.workspace.assets.uiVolumes?.[0]?.claimInfo?.storageClass;

      this.workspace.assets = {
        ...this.workspace.assets,
        uiVolumes,
      };

      // If the storage class hasn't changed there is no need to do any changes
      if (newStorageClass === oldStorageClass) return;

      if (this.storageClasses.length === 1) {
        newStorageClass = this.storageClasses[0].storageClassName;
      }

      if (!newStorageClass) return;

      if (this.workspace.assets.uiVolumes?.[0]) {
        this.workspace.assets.uiVolumes[0].claimInfo = await this.loadStorageClassConfiguration(
          this.clusterUid,
          newStorageClass,
        );
      }
    },
    updateSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      this.workspace.assets.datasources = dataSources;
    },
    onCreateNewEnvironment(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT,
        WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.environment.name,
        query: {
          ...routes.environment.query,
          projectId: this.workspace.projectId,
          scope: Scope.Project,
          workloadType: this.formType,
        },
      });
    },
    onCreateNewComputeResource(): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT,
        WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.compute.name,
        query: {
          ...routes.compute.query,
          projectId: this.workspace.projectId,
          scope: Scope.Project,
        },
      });
    },
    onCreateNewDataSource(selectedDataSourceType: string): void {
      const routes: IWorkloadSupportedRoutes | undefined = workloadUtil.getSupportedRoutes(
        WORKSPACE_ROUTE_NAMES.WORKSPACE_ASSETS_EDIT,
        WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW,
      );
      if (!routes) return;
      this.$router.push({
        name: routes.datasource.pages[selectedDataSourceType],
        query: { projectId: this.workspace.projectId, scope: Scope.Project },
      });
    },
    onNameChanged(name: string) {
      this.workspace.name = name;
    },
    cleanupNoneComplianceAssets(): void {
      if (this.workspace.assets.environment) {
        const selectedEnvironment: EnvironmentAsset | undefined = this.environments.find(
          (environment: EnvironmentAsset) => environment.meta.id === this.workspace.assets.environment,
        );
        if (!selectedEnvironment?.compliance?.compliance) {
          this.workspace.assets.environment = "";
        }
      }
      if (this.workspace.assets.compute) {
        const selectedCompute: ComputeAsset | undefined = this.computeResources.find(
          (compute: ComputeAsset) => compute.meta.id === this.workspace.assets.compute,
        );

        if (!selectedCompute?.compliance?.compliance) {
          this.workspace.assets.compute = "";
        }
      }
      if (this.workspace.assets.datasources?.length) {
        const nonComplianceDataSources: Set<string> = new Set(
          this.dataSources
            .filter((dataSource: DatasourceListResponseEntry) => !dataSource.compliance?.compliance)
            .map((dataSource: DatasourceListResponseEntry) => dataSource.meta.id),
        );
        this.workspace.assets.datasources = this.workspace.assets.datasources.filter(
          (datasource: AssetIdAndKind) => !nonComplianceDataSources.has(datasource.id),
        );
      }
    },
  },
  beforeUnmount() {
    this.saveWorkspace(this.workspace);
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.CreatedPvcUpdate);
  },
});

//Dima todo - Type assertion to prevent TS7056 while maintaining type safety -- will be refactored later --hide the annoying warning
export default WorkspaceAssetsEdit as typeof WorkspaceAssetsEdit & DefineComponent;
</script>
