<template>
  <section class="workload-template-create-form column items-center q-pt-md">
    <runai-form-wrapper :form-state="workloadTemplateModel" ref="workloadTemplateForm">
      <runai-scope-section
        kind-of-asset="template"
        :read-only="false"
        :scope-model="scopeModel"
        @changed-scope="updateScope"
        :allowed-scopes="allowedScopes"
        :key="scopeModel.clusterId || ''"
        :selected-cluster-id="scopeModel.clusterId || ''"
      />

      <runai-name-section
        aid="name-section"
        label="Template"
        asset-kind="workload-template"
        v-model="workloadTemplateModel.name"
        :scope-model="scopeModel"
        support-description
        :description="workloadTemplateModel.description"
        @update-description="updateDescription"
      />

      <transition-group name="fade">
        <template v-if="showAssetSections">
          <environment-section
            :key="`environment-section-${workloadTemplateModel.projectId}`"
            :entity-type="formType"
            aid="environment-section"
            :loading="loadingAssets || loadingCredentials"
            :environments="environments"
            :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
            :environment-id="workloadTemplateModel.assets.environment || ''"
            :specific-env="environmentModel"
            :support-config-map-as-env-var="supportConfigMapAsEnvVar"
            @environment-changed="updateSelectedEnvironment"
            @create-new="onCreateNewEnvironment"
            :section-options="envSectionOptions"
            :cluster-id="scopeModel.clusterId || ''"
          />

          <compute-resource-section
            :key="`compute-resource-section-${workloadTemplateModel.projectId}`"
            :entity-type="formType"
            aid="compute-resource-section"
            :loading="loadingAssets"
            :compute-resources="computeResources"
            :compute-resource-data="computeResourceData"
            @compute-resource-data-changed="onComputeResourceDataChanged"
            @create-new="onCreateNewComputeResource"
            :is-required="false"
            :support-tolerations="supportTolerations"
          />

          <volume-section
            :volumes="volumes"
            :storage-classes="storageClasses"
            :loading="loadingStorageClasses"
            @update-volumes="setVolumes"
            :loading-storage-class="loadingStorageClass"
          />

          <data-source-section
            :key="`data-source-section-${workloadTemplateModel.projectId}`"
            :entity-type="formType"
            aid="data-source-section"
            :loading="loadingAssets"
            :data-sources="dataSources"
            :selected-data-sources="workloadTemplateModel.assets.datasources || []"
            @datasource-changed="updateSelectedDataSources"
            @create-new="onCreateNewDataSource"
            :cluster-id="parentClusterId || scopeModel.clusterId || ''"
          />

          <general-section
            :workload-form-type="formType"
            :general-model="generalSectionModel"
            @general-model-changed="updateGeneralModel"
            :show-backoff-limit="supportBackoffLimit"
          />
        </template>
      </transition-group>

      <section class="row items-center q-mt-md">
        <q-field class="col-4 form-hint no-padding" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            color="primary"
            @click="save"
            :loading="validating || submitting"
            aid="create-template-btn"
            label="create template"
          />
        </div>
      </section>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType, type DefineComponent } from "vue";

// Components
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { EnvironmentSection } from "@/components/section/environment-section";
import { ComputeResourceSection } from "@/components/section/compute-resource-section";
import { DataSourceSection } from "@/components/section/data-source-section";
import { VolumeSection } from "@/components/section/volume-section";
import { GeneralSection } from "@/components/section/general-section";
import { RunaiBtn } from "@/components/common/runai-btn";

// Models
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import {
  Scope,
  type AssetIdAndKind,
  type ComputeAsset,
  type DatasourceListResponseEntry,
  type EnvironmentAsset,
  type WorkloadTemplateCreationRequest,
  type AssetsIds,
  type SpecificRunParams,
  type SpecificRunConnectionInfo,
} from "@/swagger-models/assets-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { IComputeSectionNodePoolData } from "@/components/section/compute-resource-section";
import { ENodePoolsListOrigin } from "@/components/section/compute-resource-section";
import { EWorkloadFormType, type IWorkloadMeta } from "@/models/workload.model";
import type { IUIWorkloadEnvSectionModel, IWorkloadEnvSectionOptions } from "@/components/section/environment-section";
import type { IComputeSectionData } from "@/components/section/compute-resource-section";
import type { Project, Resources } from "@/swagger-models/org-unit-service-client";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
// Services
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";

// Stores
import { useEnvironmentStore } from "@/stores/environment.store";
import { useComputeResourceStore } from "@/stores/compute-resource.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";

// Constants
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { type TDataSourceKinds, type IUIVolume } from "@/models/data-source.model";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_VERSION_BACKOFF_LIMIT,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { DATA_SOURCE_CREATION_ROUTE_MAP } from "@/router/data-source.routes/data-source.routes.names";

// Utils
import { omit, fallbackDefaultIfNullOrUndefined } from "@/utils/common.util";
import type { IScopeModel } from "@/models/global.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type {
  IUIWorkloadTemplateCreation,
  IUIWorkloadTemplateCreationRequestModel,
} from "@/models/workload-template.model";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useAppStore } from "@/stores/app.store";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { alertUtil } from "@/utils/alert.util";
import { assetsUtil } from "@/utils/assets.util/assets.util";
// Composables
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";
import { useStorageClassConfiguration } from "@/composables/use-storage-class-configuration.composable";

export interface IUIWorkloadTemplateMeta extends Omit<IWorkloadMeta, "clusterId" | "projectId" | "departmentId"> {
  scope: Scope | null;
  clusterId?: string | null;
  projectId?: number | null;
  departmentId?: string | null;
}

const WorkloadTemplateCreateForm = defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiScopeSection,
    RunaiNameSection,
    EnvironmentSection,
    ComputeResourceSection,
    DataSourceSection,
    VolumeSection,
    GeneralSection,
    RunaiBtn,
  },
  emits: ["cancel", "submit"],
  props: {
    initialWorkloadTemplate: {
      type: Object as PropType<IUIWorkloadTemplateCreationRequestModel>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
  },
  data() {
    return {
      sectionValidation: {
        name: false,
        scope: false,
      } as Record<string, boolean>,
      workloadTemplateStore: useWorkloadTemplateStore(),
      clusterStore: useClusterStore(),
      dataSourceStore: useDataSourceStore(),
      computeResourceStore: useComputeResourceStore(),
      environmentStore: useEnvironmentStore(),
      settingStore: useSettingStore(),
      appStore: useAppStore(),
      workloadTemplateForm: null as HTMLFormElement | null,
      workloadTemplateModel: this.getWorkloadTemplateModel() as IUIWorkloadTemplateCreation,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      loadingStorageClasses: false as boolean,
      formType: EWorkloadFormType.Template,
      loadingAssets: false as boolean,
      selectedProject: null as Project | null,
      loadingCredentials: false as boolean,
      validating: false as boolean,
    };
  },
  setup() {
    const { loadSourceValueOptions, validateEnvironmentVariablesOfAsset, environmentVariablesSourceValueOptions } =
      useEnvironmentVariablesSection();

    const { storageClasses, loadingStorageClass, loadStorageClasses, loadStorageClassConfiguration } =
      useStorageClassConfiguration("workload");

    return {
      loadSourceValueOptions,
      validateEnvironmentVariablesOfAsset,
      environmentVariablesSourceValueOptions,

      storageClasses,
      loadingStorageClass,
      loadStorageClasses,
      loadStorageClassConfiguration,
    };
  },
  async created() {
    this.workloadTemplateModel = this.convertTemplateToWorkloadUI(this.initialWorkloadTemplate);

    await this.loadAssets();

    await this.loadSourceValueOptions(
      false,
      this.workloadTemplateModel.projectId ? String(this.workloadTemplateModel.projectId) : undefined,
      this.workloadTemplateModel.departmentId || undefined,
      this.workloadTemplateModel.clusterId || undefined,
    );
    this.validateEnvironmentVariablesOfAsset(this.workloadTemplateModel.specificEnv?.environmentVariables || []);

    if (this.workloadTemplateModel.scope && this.parentClusterId) {
      await this.loadStorageClasses(this.parentClusterId, [
        this.workloadTemplateModel.assets?.uiVolumes?.[0]?.claimInfo?.storageClass,
      ]);
    }
  },
  mounted() {
    this.workloadTemplateForm = this.$refs.workloadTemplateForm as HTMLFormElement;
  },
  computed: {
    showAssetSections(): boolean {
      const { scope, departmentId, projectId, clusterId } = this.workloadTemplateModel;
      const isTenantScope: boolean = scope === Scope.Tenant;
      return !!scope && (isTenantScope || !!departmentId || !!projectId || !!clusterId);
    },
    volumes(): Array<IUIVolume> {
      return this.workloadTemplateStore.workloadVolumes;
    },
    environments(): Array<EnvironmentAsset> {
      return this.environmentStore.environmentList;
    },
    selectedEnvironment(): EnvironmentAsset | undefined {
      return this.environments.find(
        (environment: EnvironmentAsset) => environment.meta.id === this.workloadTemplateModel.assets.environment,
      );
    },
    computeResources(): Array<ComputeAsset> {
      return this.computeResourceStore.computeResourcesList;
    },
    dataSources(): Array<DatasourceListResponseEntry> {
      return this.dataSourceStore.dataSourceList;
    },
    allNodePoolsOptions(): Array<string> {
      if (!this.selectedProject) return [];
      return this.selectedProject.resources.reduce((acc: Array<string>, resource: Resources) => {
        if (resource.nodePool?.name) {
          acc.push(resource.nodePool.name);
        }
        return acc;
      }, []);
    },
    nodePoolsData(): IComputeSectionNodePoolData | null {
      if (!this.workloadTemplateModel.specificEnv || !this.workloadTemplateModel.specificEnv.nodePools) {
        return null;
      }

      return {
        defaultNodePools: this.workloadTemplateModel.specificEnv.nodePools,
        allNodePoolsOptions: this.allNodePoolsOptions,
        nodePoolsListOrigin: ENodePoolsListOrigin.Project,
      };
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        allowOverQuota: this.workloadTemplateModel.specificEnv?.allowOverQuota || false,
        autoDeletionTimeAfterCompletionSeconds: Number.isInteger(
          this.workloadTemplateModel.specificEnv?.autoDeletionTimeAfterCompletionSeconds,
        )
          ? this.workloadTemplateModel.specificEnv?.autoDeletionTimeAfterCompletionSeconds
          : null,
        annotations: this.workloadTemplateModel.specificEnv?.annotations,
        labels: this.workloadTemplateModel.specificEnv?.labels,
        backoffLimit: Number.isInteger(this.workloadTemplateModel.specificEnv?.backoffLimit)
          ? this.workloadTemplateModel.specificEnv?.backoffLimit
          : null,
      };
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.workloadTemplateModel.scope,
        projectId: this.workloadTemplateModel.projectId,
        departmentId: this.workloadTemplateModel.departmentId,
        clusterId: this.workloadTemplateModel.clusterId,
      };
    },
    environmentModel(): IUIWorkloadEnvSectionModel {
      const specificEnv = this.workloadTemplateModel.specificEnv;
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
    envSectionOptions(): IWorkloadEnvSectionOptions {
      return {
        isToolBoxRequired: false,
      };
    },
    computeResourceData(): IComputeSectionData {
      return {
        computeResourceId: this.workloadTemplateModel.assets.compute || null,
        nodeType: this.workloadTemplateModel.specificEnv?.nodeType || null,
        nodePools: this.nodePoolsData,
        tolerations: this.workloadTemplateModel.specificEnv?.tolerations,
        podAffinity: this.workloadTemplateModel.specificEnv?.podAffinity,
      };
    },
    parentClusterId(): string | null {
      if (!this.workloadTemplateModel.scope || this.workloadTemplateModel.scope === Scope.Tenant) return null;
      const scopeId = assetsUtil.getAssetScopeIdFromAssetMeta({
        name: this.workloadTemplateModel.name,
        scope: this.workloadTemplateModel.scope,
        projectId: this.workloadTemplateModel.projectId,
        departmentId: this.workloadTemplateModel.departmentId,
        clusterId: this.workloadTemplateModel.clusterId,
      });

      return orgTreeService.getParentClusterUuid(useAuthStore().orgUnitList, scopeId, this.workloadTemplateModel.scope);
    },
    supportBackoffLimit(): boolean {
      if (this.workloadTemplateModel.scope === Scope.Tenant) return false;
      return this.clusterStore.isClusterVersionSufficient(this.parentClusterId, MIN_VERSION_BACKOFF_LIMIT);
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.parentClusterId || this.scopeModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.parentClusterId || this.scopeModel.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.workloadTemplateModel.description = description;
    },
    async loadAssets() {
      if (!this.workloadTemplateModel.scope) return;
      const assetsFilter: IAssetsFilter = {
        projectId: this.workloadTemplateModel.projectId,
        departmentId: this.workloadTemplateModel.departmentId,
        clusterUuid: this.workloadTemplateModel.clusterId || undefined,
      };

      if (this.workloadTemplateModel.scope === Scope.Tenant) {
        assetsFilter.scope = Scope.Tenant;
      }

      try {
        this.loadingAssets = true;
        await Promise.all([
          this.environmentStore.loadEnvironments(assetsFilter),
          this.computeResourceStore.loadComputeResources(assetsFilter),
          this.dataSourceStore.loadDataSources(assetsFilter),
        ]);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("assets"),
          404: ErrorAlert.failedLoadingMessage("assets"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        this.appStore.setFallback(true);
      } finally {
        this.loadingAssets = false;
      }
    },
    async setVolumes(volumes: Array<IUIVolume>): Promise<void> {
      const newStorageClass = volumes[0]?.claimInfo?.storageClass;
      const oldStorageClass = this.workloadTemplateStore.workloadVolumes?.[0]?.claimInfo?.storageClass;

      const storageClassChanged: boolean = !!newStorageClass && newStorageClass !== oldStorageClass;
      if (storageClassChanged && volumes[0] && this.parentClusterId) {
        volumes[0].claimInfo = await this.loadStorageClassConfiguration(this.parentClusterId, newStorageClass as string);
      }

      this.workloadTemplateStore.setWorkloadVolumes(volumes);
    },
    onCreateNewEnvironment(): void {
      this.saveTemplate(this.workloadTemplateModel);
      this.$router.push({
        name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW,
        query: {
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: this.workloadTemplateModel.projectId,
          departmentId: this.workloadTemplateModel.departmentId,
          clusterId: this.workloadTemplateModel.clusterId,
          scope: this.workloadTemplateModel.scope,
        },
      });
    },
    onCreateNewComputeResource(): void {
      this.saveTemplate(this.workloadTemplateModel);
      this.$router.push({
        name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
        query: {
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: this.workloadTemplateModel.projectId,
          departmentId: this.workloadTemplateModel.departmentId,
          clusterId: this.workloadTemplateModel.clusterId,
          scope: this.workloadTemplateModel.scope,
        },
      });
    },
    onCreateNewDataSource(selectedDataSourceType: TDataSourceKinds): void {
      this.saveTemplate(this.workloadTemplateModel);
      this.$router.push({
        name: DATA_SOURCE_CREATION_ROUTE_MAP[selectedDataSourceType],
        query: {
          fromPage: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
          projectId: this.workloadTemplateModel.projectId,
          departmentId: this.workloadTemplateModel.departmentId,
          clusterId: this.workloadTemplateModel.clusterId,
          scope: this.workloadTemplateModel.scope,
        },
      });
    },
    updateSelectedDataSources(dataSources: Array<AssetIdAndKind>): void {
      this.workloadTemplateModel.assets.datasources = dataSources;
    },
    onComputeResourceDataChanged(computeResourceData: IComputeSectionData): void {
      this.workloadTemplateModel.assets.compute = computeResourceData.computeResourceId;

      if (this.workloadTemplateModel.specificEnv) {
        this.workloadTemplateModel.specificEnv.nodePools = computeResourceData.nodePools?.defaultNodePools;
        this.workloadTemplateModel.specificEnv.nodeType = computeResourceData.nodeType || null;
        this.workloadTemplateModel.specificEnv.tolerations = computeResourceData.tolerations;
        this.workloadTemplateModel.specificEnv.podAffinity = computeResourceData.podAffinity;
      }
    },
    updateSelectedEnvironment(envData: { environmentId: string; specificEnv: IUIWorkloadEnvSectionModel }): void {
      this.workloadTemplateModel.assets.environment = envData.environmentId;
      const environmentVariables = envData.specificEnv.environmentVariables;
      this.workloadTemplateModel.specificEnv = {
        ...this.workloadTemplateModel.specificEnv,
        ...envData.specificEnv,
        environmentVariables,
      };
    },
    updateScope(updatedScope: IScopeModel): void {
      this.workloadTemplateModel = {
        ...this.getWorkloadTemplateModel(),
        name: this.workloadTemplateModel.name,
        scope: updatedScope.scope,
        projectId: updatedScope.projectId || null,
        departmentId: updatedScope.departmentId || null,
        clusterId: updatedScope.clusterId || null,
      };
      this.loadAssets();
      this.loadStorageClasses(updatedScope.parentClusterUuid || "");
      this.initProjectNodePoolsData();
      this.loadSourceValueOptions(
        false,
        this.workloadTemplateModel.projectId ? String(this.workloadTemplateModel.projectId) : undefined,
        this.workloadTemplateModel.departmentId || undefined,
        this.workloadTemplateModel.clusterId || undefined,
      );
      this.validateEnvironmentVariablesOfAsset(this.workloadTemplateModel.specificEnv?.environmentVariables || []);
    },
    validate(): Promise<boolean> {
      return (this.workloadTemplateForm as HTMLFormElement).validate();
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success: boolean = await this.validate();
        if (!success || !this.workloadTemplateModel.assets.environment) {
          this.showHint();
          return;
        }
        if (this.workloadTemplateModel.scope === Scope.Project) {
          this.workloadTemplateModel.clusterId = this.parentClusterId;
        }

        if (
          this.workloadTemplateModel.specificEnv?.backoffLimit &&
          !this.clusterStore.isClusterVersionSupportBackoffLimit(this.parentClusterId || "")
        ) {
          this.workloadTemplateModel.specificEnv.backoffLimit = null;
        }

        this.saveTemplate(this.workloadTemplateModel);
        this.$emit("submit", this.workloadTemplateStore.workloadTemplateCreation);
      } catch (e: unknown) {
        console.error(e);
      } finally {
        this.validating = false;
      }
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      this.workloadTemplateModel.specificEnv = {
        ...this.workloadTemplateModel.specificEnv,
        ...generalModel,
      };
    },
    emptyConnectionsCustomInfo(connections: Array<SpecificRunConnectionInfo>): Array<SpecificRunConnectionInfo> {
      return connections.map((connection: SpecificRunConnectionInfo) => {
        if (connection.nodePort) {
          return {
            ...connection,
            nodePort: null,
          };
        } else if (connection.externalUrl) {
          return {
            ...connection,
            externalUrl: null,
          };
        }
        return connection;
      });
    },
    getWorkloadUI(
      meta: IUIWorkloadTemplateMeta,
      assets: AssetsIds,
      specificEnv?: SpecificRunParams,
    ): IUIWorkloadTemplateCreation {
      const workloadUI: IUIWorkloadTemplateCreation = {
        name: meta.name,
        namespace: meta.namespace,
        clusterId: meta.clusterId,
        projectId: meta.projectId,
        departmentId: meta.departmentId,
        scope: meta.scope,
        assets: {
          environment: assets.environment,
          compute: assets.compute || null,
          datasources: assets.datasources,
        },
        specificEnv: {
          args: specificEnv?.args || "",
          command: specificEnv?.command,
          runAsUid: specificEnv?.runAsUid || null,
          runAsGid: specificEnv?.runAsGid || null,
          supplementalGroups: specificEnv?.supplementalGroups,
          nodeType: specificEnv?.nodeType || null,
          allowOverQuota: specificEnv?.allowOverQuota || false,
          nodePools: specificEnv?.nodePools ? specificEnv?.nodePools : null,
          autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
            specificEnv?.autoDeletionTimeAfterCompletionSeconds,
            60 * 60 * 24 * 30,
          ),
          backoffLimit: specificEnv?.backoffLimit || 6,
          tolerations: specificEnv?.tolerations,
          podAffinity: specificEnv?.podAffinity,
        },
      };

      if (specificEnv) {
        if (specificEnv.connections) {
          workloadUI.specificEnv.connections = this.emptyConnectionsCustomInfo(specificEnv.connections);
        }

        if (specificEnv.environmentVariables) {
          workloadUI.specificEnv.environmentVariables = specificEnv.environmentVariables || [];
        }

        if (specificEnv.annotations) {
          workloadUI.specificEnv.annotations = (specificEnv.annotations || []) as IItemizedListItem[];
        }

        if (specificEnv.labels) {
          workloadUI.specificEnv.labels = (specificEnv.labels || []) as IItemizedListItem[];
        }
      }

      return workloadUI;
    },
    convertTemplateToWorkloadUI(workloadTemplate: IUIWorkloadTemplateCreationRequestModel): IUIWorkloadTemplateCreation {
      const workloadMeta: IUIWorkloadTemplateMeta = {
        name: workloadTemplate.meta.name,
        projectId: workloadTemplate.meta.projectId || null,
        departmentId: workloadTemplate.meta.departmentId || null,
        namespace: "",
        scope: workloadTemplate.meta.scope,
        clusterId: workloadTemplate.meta.clusterId || null,
      };
      return this.getWorkloadUI(
        workloadMeta,
        workloadTemplate.spec.assets,
        workloadTemplate.spec.specificEnv || undefined,
      );
    },
    saveTemplate(workloadTemplate: IUIWorkloadTemplateCreation): void {
      if (!workloadTemplate.scope) return;
      const workloadCreation: WorkloadTemplateCreationRequest = {
        meta: {
          name: workloadTemplate.name,
          description: workloadTemplate.description,
          projectId: workloadTemplate.projectId,
          departmentId: workloadTemplate.departmentId,
          scope: workloadTemplate.scope,
          clusterId: workloadTemplate.clusterId,
        },
        spec: {
          assets: {
            environment: workloadTemplate.assets.environment || "",
            compute: workloadTemplate.assets.compute,
          },
          specificEnv: {
            ...omit(workloadTemplate.specificEnv, ["environmentVariables", "annotations", "labels"]),
          },
        },
      };

      if (workloadCreation.spec.specificEnv) {
        workloadCreation.spec.specificEnv.args ||= null;
        workloadCreation.spec.specificEnv.command ||= null;
        if (workloadCreation.spec.specificEnv.supplementalGroups) {
          if (Array.isArray(workloadCreation.spec.specificEnv.supplementalGroups)) {
            workloadCreation.spec.specificEnv.supplementalGroups =
              workloadCreation.spec.specificEnv.supplementalGroups.join(",");
          }
        }
      }

      if (workloadTemplate.assets.datasources) {
        workloadCreation.spec.assets.datasources = workloadTemplate.assets.datasources;
      }

      if (workloadCreation.spec.specificEnv) {
        if (workloadTemplate.specificEnv.environmentVariables) {
          workloadCreation.spec.specificEnv.environmentVariables = workloadTemplate.specificEnv.environmentVariables;
        }

        if (workloadTemplate.specificEnv.annotations) {
          workloadCreation.spec.specificEnv.annotations = workloadTemplate.specificEnv.annotations;
        }

        if (workloadTemplate.specificEnv.labels) {
          workloadCreation.spec.specificEnv.labels = workloadTemplate.specificEnv.labels;
        }
      }

      this.workloadTemplateStore.updateWorkloadTemplateCreation(workloadCreation);
    },
    getWorkloadTemplateModel(): IUIWorkloadTemplateCreation {
      return {
        name: "",
        description: null,
        projectId: null,
        departmentId: null,
        namespace: "",
        scope: Scope.Project,
        assets: {
          environment: "",
          compute: null,
        },
        specificEnv: {
          autoDeletionTimeAfterCompletionSeconds: 60 * 60 * 24 * 30,
          backoffLimit: 6,
        },
      };
    },
    async initProjectNodePoolsData(): Promise<void> {
      if (!this.workloadTemplateModel.projectId) {
        this.selectedProject = null;
        this.workloadTemplateModel.specificEnv = {
          ...this.workloadTemplateModel.specificEnv,
          nodePools: null,
        };
        return;
      }
      try {
        this.selectedProject = await orgUnitService.getProject(this.workloadTemplateModel.projectId.toString());
        this.workloadTemplateModel.specificEnv = {
          ...this.workloadTemplateModel.specificEnv,
          nodePools: this.selectedProject.defaultNodePools || null,
        };
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Failed to load project by id`));
        console.error(e);
      }
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default WorkloadTemplateCreateForm as typeof WorkloadTemplateCreateForm & DefineComponent;
</script>
<style lang="scss" scoped></style>
