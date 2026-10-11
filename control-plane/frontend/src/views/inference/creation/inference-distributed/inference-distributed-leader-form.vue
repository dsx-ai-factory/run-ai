<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-distributed-leader-form column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :form-state="hybridContext.hybridWorkloadModel.value"
      :submitting="uiState.submitting"
      :submit-btn-text="submitBtnText"
      cancel-btn-text="Cancel"
      init-dirty
      :blurred="hybridContext.isDrawerOpen.value"
      :greedy="isGreedy"
      @unblur="hybridContext.closeDrawer"
      @leave-page="$emit(EEmitName.LeavePage)"
      @canceled="$emit(EEmitName.Canceled)"
      @submit="$emit(EEmitName.Submit)"
    >
      <title-and-back-section
        label="Cluster"
        :summary="navigation.clusterName || ''"
        back-tooltip="Change cluster"
        @on-back="$emit(EEmitName.Back)"
      />

      <title-and-back-section
        label="Project"
        :summary="navigation.projectName || ''"
        back-tooltip="Change project"
        aid="project-section"
        @on-back="$emit(EEmitName.Back)"
      />

      <title-and-back-section
        label="Workload architecture"
        summary="Distributed"
        back-tooltip="Change workload architecture"
        @on-back="$emit(EEmitName.Back)"
      />

      <q-linear-progress animation-speed="100" indeterminate v-if="uiState.scopeChangeInProgress" />
      <transition-group name="fade">
        <template v-if="isValidScopeSelected && !uiState.scopeChangeInProgress">
          <title-and-back-section
            v-if="navigation.showInferenceType"
            label="Inference type"
            :summary="navigation.typeSummary"
          />

          <workload-name-section
            :default-open="navigation.defaultOpenNameSection"
            entity-type="inference"
            :name="hybridContext.hybridWorkloadModel.value.name"
            @update:name="handleNameChanged"
            :project-id="+hybridContext.hybridWorkloadModel.value.projectId"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
          />

          <hybrid-environment-section
            authenticated-ngc-only
            :key="`he-${hybridContext.selectedTemplates.value.environment.sectionKey}`"
            :render-key="hybridContext.selectedTemplates.value.environment.renderKey.toString()"
            :selected-name="hybridContext.selectedEnvironmentName.value"
            :workload-spec="environmentSectionModel"
            :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
            :policy="leaderPolicy"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            :support-config-map-as-env-var="supportConfigMapAsEnvVar"
            :support-user-credentials="supportUserCredentials"
            supporting-serving-port
            supporting-serving-port-access
            is-distributed-inference
            :support-image-pull-secrets="features.supportImagePullSecrets"
            :required-fields="{ image: !uiState.isWorkloadTemplate, port: !uiState.isWorkloadTemplate }"
            @load-section-templates="hybridContext.displayEnvironmentsList"
            @update="handleEnvironmentSectionUpdate"
            @reset="resetEnvironment"
            @clear="clearEnvironment"
            :support-linking="uiState.isWorkloadTemplate"
            :is-linked="!!hybridContext.selectedTemplates.value.environment.isLinked"
            :validate-empty-env-vars="!uiState.isWorkloadTemplate"
            @toggle-linked="hybridContext.toggleEnvironmentLinked"
            support-authenticated-ngc
            :scope-id="
              uiState.isWorkloadTemplate
                ? hybridContext.hybridWorkloadModel.value.scopeId || ''
                : hybridContext.hybridWorkloadModel.value.projectId
            "
            :scope-type="uiState.isWorkloadTemplate ? hybridContext.hybridWorkloadModel.value.scopeType : 'project'"
            :hide-specific-sections="{ tools: true }"
            :disable-fields="{ protocol: true }"
          />

          <hybrid-compute-resource-section
            :key="`hc-${hybridContext.selectedTemplates.value['compute-resource'].sectionKey}`"
            :render-key="hybridContext.selectedTemplates.value['compute-resource'].renderKey.toString()"
            :selected-name="hybridContext.selectedComputeName.value"
            :workload-spec="computedLeaderSpec"
            :policy="leaderPolicy"
            :support-multi-fraction-gpu="supportMultiFractionGpu"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            :auto-scale-options="clusterConfig.autoScaleOptions"
            :node-pools-resources="clusterConfig.projectNodePoolsResources"
            :support-tolerations="supportTolerations"
            :support-linking="uiState.isWorkloadTemplate"
            :is-linked="!!hybridContext.selectedTemplates.value['compute-resource'].isLinked"
            section-sub-title="Set the resources for the leader"
            @load-section-templates="hybridContext.displayComputesList"
            @update="handleComputeSectionUpdate"
            @reset="resetCompute"
            @clear="clearCompute"
            :hide-specific-sections="{
              affinity: uiState.isWorkloadTemplate,
            }"
            @toggle-linked="hybridContext.toggleComputeLinked"
            show-topology
          />

          <hybrid-data-source-section
            :default-opened="shouldOpenDataSourceSection"
            :workload-storage="hybridContext.hybridWorkloadModel.value.spec.leader?.storage"
            :policy-rules="leaderPolicy.rules?.storage"
            :policy-defaults="leaderPolicy.defaults?.storage"
            @load-section-templates="hybridContext.displayDataSourcesList"
            @update="handleModelStorageUpdate"
            :secrets="dataSources.secrets"
            :allowed-data-source-types="inferenceSupportedDataSources"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            :show-volume-persistency-section="false"
            :is-volumes-enabled="features.isVolumesEnabled"
            custom-section-label="Data & storage"
          />

          <general-section
            aid="general-section"
            should-auto-validate
            :workload-form-type="EWorkloadFormType.Inference"
            :general-model="generalSectionModel"
            :show-backoff-limit="false"
            :policy-rules="leaderPolicy.rules"
            :itemize-defaults="generalSectionItemizeDefaults"
            :support-preemptibility="false"
            @general-model-changed="handleGeneralModelChanged"
            :show-request-timeout="features.showRequestTimeout"
          />
        </template>
      </transition-group>
    </workload-form-wrapper>
    <workload-creation-drawer
      :allowed-data-source-options="inferenceSupportedDataSources"
      :is-open="hybridContext.isDrawerOpen.value"
      :comply-to="drawerFilters"
      :entity="hybridContext.drawerDataEntity.value"
      :policy-details="policy?.policyViewDetails"
      display-replica-type="Leader"
      :created-entity-id="hybridContext.recentlyCreatedEntityId.value"
      :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
      @close="hybridContext.closeDrawer"
      @select="onDrawerSelect"
      @select-multiple="onDrawerSelectMultiple"
      @create-new="onCreateNew"
      :apply-item-modifications-after-loaded="applyItemModificationsAfterLoaded"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection, type IUIGeneralSectionModel } from "@/components/section/general-section";
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";

// models
import { EWorkloadFormType, EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import type {
  IInferenceSubmitFormClusterConfig,
  IInferenceSubmitFormDataSources,
  IInferenceSubmitFormFeatures,
  IInferenceSubmitFormNavigation,
  IDistributedInferenceSubmitFormPolicy,
  IInferenceSubmitFormUIState,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type {
  IUIDistributedWorkloadSpec,
  IUISuperWorkloadCreationRequest,
  IUISuperWorkloadSpec,
  IUIWorkloadStorageFields,
} from "@/models/workload.model";
import { WorkloadCreationDrawerEntities, type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import { AssetKind, Scope, type SupersetServingPort } from "@/swagger-models/assets-service-client";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import { EDistributedInferenceType } from "./inference-distributed.model";

// stores
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useSettingStore } from "@/stores/setting.store";

// utils
import { inferenceDataSourceRestrictionsUtil } from "@/utils/inference-data-source-restrictions.util/inference-data-source-restrictions.util";
import { distributedInferenceUtil } from "@/utils/distributed-inference.util";
import { deepCopy } from "@/utils/common.util/common.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";

// consts
import { useRoute } from "vue-router";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
} from "@/common/version.constant";
import type { IAssetsFilter } from "@/models/filter.model";
import type { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import type { SupersetServingPortRules, UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";

export enum EEmitName {
  UpdateSpec = "update:spec",
  SaveState = "save-state",
  LeavePage = "leave-page",
  Canceled = "canceled",
  Back = "back",
  Submit = "submit",
}

function applyItemModificationsAfterLoaded(
  entity: WorkloadCreationDrawerEntities,
  items: IDrawerListItem[],
  clusterId?: string,
): IDrawerListItem[] {
  if (entity === WorkloadCreationDrawerEntities.DataSource || entity === WorkloadCreationDrawerEntities.ModelStorage) {
    return inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId || "");
  }
  return items;
}

export default defineComponent({
  name: "inference-distributed-leader-form",
  components: {
    TitleAndBackSection,
    WorkloadFormWrapper,
    WorkloadNameSection,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    HybridDataSourceSection,
    GeneralSection,
    WorkloadCreationDrawer,
  },
  emits: {
    [EEmitName.UpdateSpec]: (modelValue: IUISuperWorkloadCreationRequest["spec"]) => modelValue,
    [EEmitName.LeavePage]: () => true,
    [EEmitName.Canceled]: () => true,
    [EEmitName.Back]: () => true,
    [EEmitName.Submit]: () => true,
    [EEmitName.SaveState]: () => true,
  },
  props: {
    hybridContext: {
      type: Object as PropType<ReturnType<typeof useHybridTemplates>>,
      required: true,
    },
    navigation: {
      type: Object as PropType<IInferenceSubmitFormNavigation>,
      required: true,
    },
    uiState: {
      type: Object as PropType<IInferenceSubmitFormUIState>,
      required: true,
    },
    features: {
      type: Object as PropType<IInferenceSubmitFormFeatures>,
      required: true,
    },
    triggerValidationOnStartUp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    environmentVariablesSourceValueOptions: {
      type: Array as PropType<TEnvVarNoneCustomSourceOption[]>,
      default: () => [],
    },
    policy: {
      type: Object as PropType<IDistributedInferenceSubmitFormPolicy | undefined>,
      required: false,
    },
    clusterConfig: {
      type: Object as PropType<IInferenceSubmitFormClusterConfig>,
      required: true,
    },
    dataSources: {
      type: Object as PropType<IInferenceSubmitFormDataSources>,
      required: true,
    },
  },
  data() {
    return {
      route: useRoute(),
      clusterStore: useClusterStore(),
      dataSourceStore: useDataSourceStore(),
      inferenceStore: useInferenceStore(),
      authStore: useAuthStore(),
      settingStore: useSettingStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      EWorkloadFormType,
      EEmitName,
      EDistributedInferenceType,
      submitBtnText: "Continue",
      applyItemModificationsAfterLoaded: applyItemModificationsAfterLoaded,
      isGreedy: false,
    };
  },
  async created() {
    const { kind, createdEntityId, workloadFormName } = this.$route.query;

    if (workloadFormName === EDistributedInferenceType.Leader) {
      const query = deepCopy(this.$route.query);
      delete query.workloadFormName;
      delete query.kind;
      delete query.createdEntityId;
      await this.$router.replace({
        query: { ...query },
      });
      this.hybridContext.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());
    }

    this.createSectionsClearedStates();
  },
  mounted() {
    if (this.triggerValidationOnStartUp) {
      this.isGreedy = true;
      this.$nextTick(async () => {
        const formWrapperRef = this.$refs.formWrapperRef as typeof WorkloadFormWrapper;
        await formWrapperRef?.validate();
        this.isGreedy = false;
      });
    }
  },
  computed: {
    isValidScopeSelected(): boolean {
      if (!this.uiState.isWorkloadTemplate) return true;
      return this.uiState.isValidScopeSelected || false;
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportUserCredentials(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
      );
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
    supportsNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
      );
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
    inferenceSupportedDataSources(): Set<TDataSourceKinds> {
      const supportedDataSources = distributedInferenceUtil.getSupportedDataSources();

      return new Set(
        [...(this.dataSources.allowedDataSourceOptions || [])].filter((ds) => supportedDataSources.includes(ds)),
      );
    },
    generalSectionModel(): IUIGeneralSectionModel {
      const leaderSpec = this.hybridContext.hybridWorkloadModel.value.spec.leader;
      const generalSectionModel: IUIGeneralSectionModel = {
        annotations:
          leaderSpec?.annotations?.map((annotation) => ({
            name: annotation.name || "",
            value: annotation.value || "",
          })) || undefined,
        labels: leaderSpec?.labels?.map((label) => ({ name: label.name || "", value: label.value || "" })) || undefined,
      };
      return generalSectionModel;
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        annotations: this.policy?.distributedPolicy.worker?.defaults?.annotations || undefined,
        labels: this.policy?.distributedPolicy.worker?.defaults?.labels || undefined,
      };
    },
    drawerFilters(): IAssetsFilter {
      const model = this.hybridContext.hybridWorkloadModel.value;
      if (this.uiState.isWorkloadTemplate && this.uiState.baseDrawerFilters) {
        return this.uiState.baseDrawerFilters;
      }
      const projectId = this.uiState.isWorkloadTemplate
        ? model.scopeType === Scope.Project && model.scopeId
          ? +model.scopeId
          : undefined
        : +model.projectId;
      return {
        projectId,
        complyToProject: +model.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.DistributedInference,
        isDistributedInference: true,
        complyToReplicaType: "Leader",
      };
    },
    environmentSectionModel(): IUISuperWorkloadSpec {
      return {
        ...this.hybridContext.hybridWorkloadModel.value.spec.leader,
        servingPort: this.hybridContext.hybridWorkloadModel.value.spec.servingPort,
      };
    },
    computedLeaderSpec(): IUISuperWorkloadSpec {
      return {
        ...this.hybridContext.hybridWorkloadModel.value.spec.leader,
        nodePools: this.hybridContext.hybridWorkloadModel.value.spec.nodePools,
      };
    },
    leaderPolicy(): UnifiedPolicyInfoPerReplicaType {
      const workerPolicy = this.policy?.distributedPolicy.worker;

      const buildServingPort = <T extends SupersetServingPort | SupersetServingPortRules>(
        servingPort?: T | null,
      ): (T & { container: T["port"] }) | undefined => {
        if (!servingPort) return undefined;
        return {
          ...servingPort,
          container: servingPort.port,
        };
      };

      return {
        defaults: {
          ...workerPolicy?.defaults,
          ...workerPolicy?.defaults?.leader,
          servingPort: buildServingPort(workerPolicy?.defaults?.servingPort),
          leader: undefined,
          worker: undefined,
        },
        rules: {
          ...workerPolicy?.rules,
          ...workerPolicy?.rules?.leader,
          servingPort: buildServingPort(workerPolicy?.rules?.servingPort),
          leader: undefined,
          worker: undefined,
        },
      };
    },
  },
  methods: {
    createSectionsClearedStates(): void {
      const base = deepCopy(this.hybridContext.hybridWorkloadModel.value);
      const environmentClearedState = workloadHybridUtil.getDistributedInferenceLeaderClearedEnvironmentFields(base);
      const templates = this.hybridContext.selectedTemplates.value;
      templates[WorkloadCreationDrawerEntities.Environment].clearedState = environmentClearedState;
      templates[WorkloadCreationDrawerEntities.Compute].clearedState = workloadHybridUtil.getRelevantComputeFields(base);
      templates[WorkloadCreationDrawerEntities.DataSource].clearedState =
        workloadHybridUtil.getRelevantStorageFields(base);
    },
    updateParent(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.$emit(EEmitName.UpdateSpec, spec);
    },
    handleNameChanged(name: string): void {
      this.hybridContext.updateWorkloadName(name);
    },
    handleEnvironmentSectionUpdate(leaderSpec: IUISuperWorkloadSpec): void {
      const { servingPort, ...leaderWithoutServingPort } = leaderSpec;
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        leader: leaderWithoutServingPort,
        servingPort: servingPort ?? this.hybridContext.hybridWorkloadModel.value.spec.servingPort,
      };
      this.updateParent(updatedSpec);
    },
    handleComputeSectionUpdate(leaderSpec: IUISuperWorkloadSpec): void {
      const { nodePools, ...leaderWithoutNodePools } = leaderSpec;
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        leader: leaderWithoutNodePools,
        nodePools: nodePools ?? this.hybridContext.hybridWorkloadModel.value.spec.nodePools,
      };
      this.updateParent(updatedSpec);
    },
    updateLeaderSpec(leaderSpec: IUIDistributedWorkloadSpec): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        leader: leaderSpec,
      };
      this.updateParent(updatedSpec);
    },
    handleModelStorageUpdate(storage: IUIWorkloadStorageFields): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        leader: {
          ...this.hybridContext.hybridWorkloadModel.value.spec.leader,
          storage,
        },
      };
      this.updateParent(updatedSpec);
    },
    handleGeneralModelChanged(generalModel: IUIGeneralSectionModel): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        leader: {
          ...this.hybridContext.hybridWorkloadModel.value.spec.leader,
          ...generalModel,
        },
      };
      this.updateParent(updatedSpec);
    },
    async onDrawerSelect(itemIdAndName: { id: string; name: string }): Promise<void> {
      await this.hybridContext.handleSelection(itemIdAndName);
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    async onDrawerSelectMultiple(selectedItems: { id: string; name: string; kind: AssetKind }[]): Promise<void> {
      await this.hybridContext.handleMultiSelection(selectedItems);
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.hybridContext.onCreateNewSetup(type, {
        workloadType: EWorkloadFormType.DistributedInference,
        workloadFormName: EDistributedInferenceType.Leader,
      });
    },
    resetEnvironment(): void {
      this.hybridContext.resetEnvironmentToDefault();
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    clearEnvironment(): void {
      this.hybridContext.clearEnvironmentTemplate();
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    resetCompute(): void {
      this.hybridContext.resetComputeToDefault();
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    clearCompute(): void {
      this.hybridContext.clearComputeTemplate();
      this.updateLeaderSpec(this.hybridContext.hybridWorkloadModel.value.spec.leader || {});
    },
    saveFormState(): void {
      this.$emit(EEmitName.SaveState);
    },
  },
});
</script>
