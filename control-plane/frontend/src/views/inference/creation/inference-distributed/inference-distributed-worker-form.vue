<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-distributed-worker-form column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :form-state="hybridContext.hybridWorkloadModel.value"
      :submitting="uiState.submitting"
      :submit-btn-text="submitBtnText"
      cancel-btn-text="Back"
      init-dirty
      :blurred="hybridContext.isDrawerOpen.value"
      :greedy="isGreedy"
      @unblur="hybridContext.closeDrawer"
      @leave-page="$emit(EEmitName.LeavePage)"
      @canceled="$emit(EEmitName.Back)"
      @submit="$emit(EEmitName.Submit)"
    >
      <hybrid-environment-section
        authenticated-ngc-only
        :key="`he-${hybridContext.selectedTemplates.value.environment.sectionKey}`"
        :render-key="hybridContext.selectedTemplates.value.environment.renderKey.toString()"
        :selected-name="hybridContext.selectedEnvironmentName.value"
        :workload-spec="hybridContext.hybridWorkloadModel.value.spec.worker || {}"
        :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
        :policy="workerPolicy"
        :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
        :support-config-map-as-env-var="supportConfigMapAsEnvVar"
        :support-user-credentials="supportUserCredentials"
        :supporting-serving-port="false"
        :support-image-pull-secrets="features.supportImagePullSecrets"
        :required-fields="{ image: !uiState.isWorkloadTemplate, port: !uiState.isWorkloadTemplate }"
        @load-section-templates="hybridContext.displayEnvironmentsList"
        @update="handleWorkloadSpecUpdate"
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
      />

      <hybrid-compute-resource-section
        :key="`hc-${hybridContext.selectedTemplates.value['compute-resource'].sectionKey}`"
        :render-key="hybridContext.selectedTemplates.value['compute-resource'].renderKey.toString()"
        :selected-name="hybridContext.selectedComputeName.value"
        :workload-spec="computedWorkloadSpec"
        :policy="workerPolicy"
        :support-multi-fraction-gpu="supportMultiFractionGpu"
        :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
        :auto-scale-options="clusterConfig.autoScaleOptions"
        :node-pools-resources="clusterConfig.projectNodePoolsResources"
        :support-tolerations="supportTolerations"
        :support-linking="uiState.isWorkloadTemplate"
        :is-linked="!!hybridContext.selectedTemplates.value['compute-resource'].isLinked"
        section-sub-title="Set the resources for each worker"
        distributed-workers-input-title="Set the number of workers for each replica"
        @load-section-templates="hybridContext.displayComputesList"
        @update="handleComputeSectionUpdate"
        @reset="resetCompute"
        @clear="clearCompute"
        :hide-specific-sections="{
          affinity: uiState.isWorkloadTemplate,
        }"
        :disable-fields="{
          nodePools: true,
          nodePoolsTooltip: 'This list is derived from the Leader setup and can be modified in the previous step.',
        }"
        @toggle-linked="hybridContext.toggleComputeLinked"
        show-topology
      />

      <hybrid-data-source-section
        :default-opened="shouldOpenDataSourceSection"
        :workload-storage="hybridContext.hybridWorkloadModel.value.spec.worker?.storage"
        :policy-rules="workerPolicy?.rules?.storage"
        :policy-defaults="workerPolicy?.defaults?.storage"
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
        :policy-rules="workerPolicy?.rules"
        :itemize-defaults="generalSectionItemizeDefaults"
        :support-preemptibility="false"
        @general-model-changed="handleGeneralModelChanged"
        :show-request-timeout="features.showRequestTimeout"
      />
    </workload-form-wrapper>
    <workload-creation-drawer
      :allowed-data-source-options="inferenceSupportedDataSources"
      :is-open="hybridContext.isDrawerOpen.value"
      :comply-to="drawerFilters"
      :entity="hybridContext.drawerDataEntity.value"
      :policy-details="policy?.policyViewDetails"
      display-replica-type="Worker"
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
  IDistributedInferenceSubmitFormPolicy,
  IInferenceSubmitFormUIState,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type { IUISuperWorkloadCreationRequest, IUIWorkloadStorageFields } from "@/models/workload.model";
import { WorkloadCreationDrawerEntities, type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
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
import { deepCopy, omit } from "@/utils/common.util/common.util";

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
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";

export enum EEmitName {
  UpdateSpec = "update:spec",
  SaveState = "save-state",
  LeavePage = "leave-page",
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
  name: "inference-distributed-worker-form",
  components: {
    WorkloadFormWrapper,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    HybridDataSourceSection,
    GeneralSection,
    WorkloadCreationDrawer,
  },
  emits: {
    [EEmitName.UpdateSpec]: (modelValue: IUISuperWorkloadCreationRequest["spec"]) => modelValue,
    [EEmitName.LeavePage]: () => true,
    [EEmitName.Back]: () => true,
    [EEmitName.Submit]: () => true,
    [EEmitName.SaveState]: () => true,
  },
  props: {
    hybridContext: {
      type: Object as PropType<ReturnType<typeof useHybridTemplates>>,
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

    if (workloadFormName === EDistributedInferenceType.Workers) {
      const query = deepCopy(this.$route.query);
      delete query.workloadFormName;
      delete query.kind;
      delete query.createdEntityId;
      await this.$router.replace({
        query: { ...query },
      });
      this.hybridContext.handleCreatedAssets(kind?.toString(), createdEntityId?.toString());
    }

    this.hybridContext.createSectionsClearedStates();
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
      const workerSpec = this.hybridContext.hybridWorkloadModel.value.spec.worker;
      const generalSectionModel: IUIGeneralSectionModel = {
        annotations:
          workerSpec?.annotations?.map((annotation) => ({
            name: annotation.name || "",
            value: annotation.value || "",
          })) || undefined,
        labels: workerSpec?.labels?.map((label) => ({ name: label.name || "", value: label.value || "" })) || undefined,
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
        complyToReplicaType: "Worker",
      };
    },
    computedWorkloadSpec(): IUISuperWorkloadCreationRequest["spec"] {
      return {
        ...this.hybridContext.hybridWorkloadModel.value.spec.worker,
        numWorkers: this.hybridContext.hybridWorkloadModel.value.spec.workers,
        nodePools: this.hybridContext.hybridWorkloadModel.value.spec.nodePools,
      };
    },
    workerPolicy(): UnifiedPolicyInfoPerReplicaType {
      return {
        defaults: {
          ...this.policy?.distributedPolicy.worker?.defaults,
          ...this.policy?.distributedPolicy.worker?.defaults?.worker,
          numWorkers: this.policy?.distributedPolicy.worker?.defaults?.workers,
          leader: undefined,
          worker: undefined,
        },
        rules: {
          ...this.policy?.distributedPolicy.worker?.rules,
          ...this.policy?.distributedPolicy.worker?.rules?.worker,
          numWorkers: this.policy?.distributedPolicy.worker?.rules?.workers,
          leader: undefined,
          worker: undefined,
        },
      };
    },
  },
  methods: {
    updateParent(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.$emit(EEmitName.UpdateSpec, spec);
    },
    handleNameChanged(name: string): void {
      this.hybridContext.updateWorkloadName(name);
    },
    handleWorkloadSpecUpdate(workerSpec: IUISuperWorkloadCreationRequest["spec"]["worker"]): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        worker: omit(workerSpec || {}, ["numWorkers"]),
        numWorkers: workerSpec?.numWorkers || this.hybridContext.hybridWorkloadModel.value.spec.numWorkers,
      };
      this.updateParent(updatedSpec);
    },
    handleComputeSectionUpdate(workerSpec: IUISuperWorkloadCreationRequest["spec"]["worker"]): void {
      const { nodePools, numWorkers, ...workerWithoutNodePoolsAndNumWorkers } = workerSpec || {};
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        worker: workerWithoutNodePoolsAndNumWorkers,
        workers: numWorkers || this.hybridContext.hybridWorkloadModel.value.spec.numWorkers,
        nodePools: nodePools ?? this.hybridContext.hybridWorkloadModel.value.spec.nodePools,
      };
      this.updateParent(updatedSpec);
    },
    handleModelStorageUpdate(storage: IUIWorkloadStorageFields): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        worker: {
          ...omit(this.hybridContext.hybridWorkloadModel.value.spec.worker || {}, ["numWorkers"]),
          storage,
        },
      };
      this.updateParent(updatedSpec);
    },
    handleGeneralModelChanged(generalModel: IUIGeneralSectionModel): void {
      const updatedSpec: IUISuperWorkloadCreationRequest["spec"] = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        worker: {
          ...omit(this.hybridContext.hybridWorkloadModel.value.spec.worker || {}, ["numWorkers"]),
          ...generalModel,
        },
      };
      this.updateParent(updatedSpec);
    },
    async onDrawerSelect(itemIdAndName: { id: string; name: string }): Promise<void> {
      await this.hybridContext.handleSelection(itemIdAndName);
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    async onDrawerSelectMultiple(selectedItems: { id: string; name: string; kind: AssetKind }[]): Promise<void> {
      await this.hybridContext.handleMultiSelection(selectedItems);
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.hybridContext.onCreateNewSetup(type, {
        workloadType: EWorkloadFormType.DistributedInference,
        workloadFormName: EDistributedInferenceType.Workers,
      });
    },
    resetEnvironment(): void {
      this.hybridContext.resetEnvironmentToDefault();
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    clearEnvironment(): void {
      this.hybridContext.clearEnvironmentTemplate();
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    resetCompute(): void {
      this.hybridContext.resetComputeToDefault();
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    clearCompute(): void {
      this.hybridContext.clearComputeTemplate();
      this.handleWorkloadSpecUpdate(this.hybridContext.hybridWorkloadModel.value.spec.worker);
    },
    saveFormState(): void {
      this.$emit(EEmitName.SaveState);
    },
  },
});
</script>
