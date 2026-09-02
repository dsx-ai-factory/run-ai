<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="distributed-hybrid-worker column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :greedy="isGreedy"
      :form-state="hybridContext.hybridWorkloadModel.value"
      :submitting="submitting"
      submit-btn-text="Continue"
      cancel-btn-text="Cancel"
      :blurred="hybridContext.isDrawerOpen.value"
      init-dirty
      @unblur="hybridContext.closeDrawer"
      @canceled="onCancel"
      @submit="onContinue"
      @leave-page="onLeave"
    >
      <title-and-back-section
        v-if="!isTemplate"
        label="Cluster"
        :summary="clusterName"
        :back-tooltip="isEditMode ? '' : 'Change cluster'"
        @on-back="onBack"
      />

      <slot name="scope-section">
        <title-and-back-section
          label="Project"
          :summary="projectData.projectName"
          :back-tooltip="isEditMode ? '' : 'Change project'"
          aid="project-section"
          @on-back="onBack"
        />
      </slot>
      <q-linear-progress animation-speed="100" indeterminate v-if="loadingScopeBasedData" />
      <transition-group name="fade">
        <template v-if="isValidScopeSelected && !loadingScopeBasedData">
          <multi-node-section
            :is-selectable="false"
            :multi-node-section-model="{
              distFramework: hybridContext.hybridWorkloadModel.value.spec.distributedFramework,
            }"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
          >
            <template #custom-button v-if="!isEditMode">
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

          <slot name="name-section">
            <workload-name-section
              :default-open="!!$route.query.fromCopyId"
              entity-type="training"
              :name="hybridContext.hybridWorkloadModel.value.name"
              @update:name="onNameChanged"
              :project-id="+hybridContext.hybridWorkloadModel.value.projectId"
              :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            />
          </slot>

          <hybrid-environment-section
            :key="`he-${hybridContext.selectedTemplates.value.environment.sectionKey}`"
            :render-key="hybridContext.selectedTemplates.value.environment.renderKey.toString()"
            :selected-name="hybridContext.selectedEnvironmentName.value"
            :workload-spec="hybridContext.hybridWorkloadModel.value.spec"
            :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
            :policy="policy"
            :required-fields="isTemplate ? {} : { image: true }"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            :support-config-map-as-env-var="supportConfigMapAsEnvVar"
            :support-user-credentials="supportUserCredentials"
            support-authenticated-ngc
            :scope-id="scopeModel.scopeId"
            :scope-type="scopeModel.scopeType"
            :support-image-pull-secrets="!hideSpecificSections?.environment?.imagePullSecrets && supportImagePullSecrets"
            :support-linking="isTemplate"
            :is-linked="!!hybridContext.selectedTemplates.value.environment.isLinked"
            :validate-empty-env-vars="!isTemplate"
            @load-section-templates="hybridContext.displayEnvironmentsList"
            @update="updateModelSpec"
            @reset="resetEnvironment"
            @clear="clearEnvironment"
            @toggle-linked="hybridContext.toggleEnvironmentLinked"
            @credential-created="onCredentialCreated"
          />

          <hybrid-compute-resource-section
            :key="`hc-${hybridContext.selectedTemplates.value['compute-resource'].sectionKey}`"
            :render-key="hybridContext.selectedTemplates.value['compute-resource'].renderKey.toString()"
            :selected-name="hybridContext.selectedComputeName.value"
            :workload-spec="hybridContext.hybridWorkloadModel.value.spec"
            :project-node-types="projectData.selectedProjectNodeTypes"
            :node-pools-resources="projectData.projectNodePoolsResources"
            :policy="policy"
            :support-multi-fraction-gpu="supportMultiFractionGpu"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
            :support-tolerations="supportTolerations"
            :hide-specific-sections="hideSpecificSections?.computeResource"
            :is-linked="!!hybridContext.selectedTemplates.value['compute-resource'].isLinked"
            :support-linking="isTemplate"
            @load-section-templates="hybridContext.displayComputesList"
            @update="updateModelSpec"
            @reset="resetCompute"
            @clear="clearCompute"
            @toggle-linked="hybridContext.toggleComputeLinked"
            show-topology
          />

          <hybrid-data-source-section
            :default-opened="shouldOpenDataSourceSection"
            :workload-storage="hybridContext.hybridWorkloadModel.value.spec.storage"
            @load-section-templates="hybridContext.displayDataSourcesList"
            @update="updateModelStorage"
            :secrets="secrets"
            :policy-rules="policy?.rules?.storage"
            :policy-defaults="policy?.defaults?.storage"
            :allowed-data-source-types="allowedDataSources"
            :cluster-id="hybridContext.hybridWorkloadModel.value.clusterId"
          />

          <general-section
            should-auto-validate
            :workload-form-type="EWorkloadFormType.Distributed"
            :distributed-replica-type="EDistributedReplicaType.Worker"
            :general-model="generalSectionModel"
            :policy-rules="policy?.rules"
            :itemize-defaults="generalSectionItemizeDefaults"
            @general-model-changed="updateGeneralModel"
          />
        </template>
      </transition-group>
    </workload-form-wrapper>

    <workload-creation-drawer
      :is-open="hybridContext.isDrawerOpen.value"
      :comply-to="drawerFilters"
      :entity="hybridContext.drawerDataEntity.value"
      :policy-details="policyViewDetails"
      :created-entity-id="hybridContext.recentlyCreatedEntityId.value"
      :allowed-data-source-options="allowedDataSources"
      @close="hybridContext.closeDrawer"
      @select="hybridContext.handleSelection"
      @select-multiple="hybridContext.handleMultiSelection"
      @create-new="onCreateNew"
    />
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, type DefineComponent } from "vue";
// Components
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";
import { MultiNodeSection } from "@/components/section/multi-node-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Stores & Services
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { deepCopy } from "@/utils/common.util";
// Models
import { AssetKind, type StringOption } from "@/swagger-models/assets-service-client";
import { EDistributedReplicaType, EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import { type IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { EWorkloadFormType } from "@/models/workload.model";
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { IUISuperWorkloadCreationRequest, IUIStorageFields } from "@/models/workload.model";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";

import { type UnifiedPolicyInfoPerReplicaType, type ClusterPriority } from "@/swagger-models/workloads-client";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { useHybridTemplates } from "@/composables/use-hybrid-templates.composable";
import type { IHideSpecificSections } from "@/models/workload-hybrid.model";

export type TProjectData = {
  projectNodePoolsResources: StringOption[];
  projectName: string;
  selectedProjectNodeTypes: string[];
};

const DistributedHybridWorker = defineComponent({
  name: "distributed-hybrid-worker",
  components: {
    WorkloadCreationDrawer,
    HybridEnvironmentSection,
    HybridComputeResourceSection,
    TitleAndBackSection,
    WorkloadNameSection,
    WorkloadFormWrapper,
    HybridDataSourceSection,
    GeneralSection,
    MultiNodeSection,
    RunaiBtn,
  },
  emits: ["continue", "update", "save-form-state", "leave-page", "back-to-create-new", "credential-created"],
  props: {
    hybridContext: {
      type: Object as PropType<ReturnType<typeof useHybridTemplates>>,
      required: true,
    },
    projectData: {
      type: Object as PropType<TProjectData>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    policyViewDetails: {
      type: Object as PropType<IPolicyViewDetails>,
      required: false,
    },
    environmentVariablesSourceValueOptions: {
      type: Array as PropType<TEnvVarNoneCustomSourceOption[]>,
      required: false,
    },
    secrets: {
      type: Array as PropType<SecretListInfo[]>,
      required: false,
      default: () => [],
    },
    allowedDataSources: {
      type: Object as PropType<Set<TDataSourceKinds>>,
      required: false,
    },
    workloadPriorities: {
      type: Array as PropType<ClusterPriority[]>,
      required: false,
    },
    isTemplate: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hideSpecificSections: {
      type: Object as PropType<IHideSpecificSections>,
      required: false,
    },
    supportImagePullSecrets: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    supportAuthenticatedNgc: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    isValidScopeSelected: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    loadingScopeBasedData: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    triggerValidationOnStartUp: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),

      EWorkloadFormType,
      EDistributedReplicaType,

      isGreedy: false,
    };
  },
  async created() {
    const { kind, createdEntityId, workloadFormName } = this.$route.query;

    if (this.triggerValidationOnStartUp) {
      this.isGreedy = true;
    }

    if (workloadFormName === EDistributedReplicaType.Worker) {
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

    if (this.isGreedy) {
      this.$nextTick(async () => {
        const formWrapperRef = this.$refs.formWrapperRef as typeof WorkloadFormWrapper;
        await formWrapperRef?.validate();
        this.isGreedy = false;
      });
    }
  },
  computed: {
    scopeModel() {
      return {
        scopeType: this.hybridContext.hybridWorkloadModel.value.scopeType || "project",
        scopeId:
          this.hybridContext.hybridWorkloadModel.value.projectId?.toString() ||
          (this.hybridContext.hybridWorkloadModel.value.scopeId as string) ||
          "",
      };
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
    drawerFilters(): IAssetsFilter {
      if (this.isTemplate) {
        return {
          ...this.hybridContext.baseDrawerFilters.value,
          complyToReplicaType: "Worker",
          distributedFramework: this.hybridContext.hybridWorkloadModel.value.spec.distributedFramework,
        };
      }

      return {
        projectId: +this.hybridContext.hybridWorkloadModel.value.projectId,
        complyToProject: this.isTemplate ? undefined : +this.hybridContext.hybridWorkloadModel.value.projectId,
        complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Distributed,
        complyToReplicaType: "Worker",
        distributedFramework: this.hybridContext.hybridWorkloadModel.value.spec.distributedFramework,
        isDistributed: true,
      };
    },
    clusterName(): string {
      return (
        this.clusterStore.clusterList.find(
          (cluster) => cluster.uuid === this.hybridContext.hybridWorkloadModel.value.clusterId,
        )?.name || ""
      );
    },
    generalSectionModel(): IUIGeneralSectionModel {
      return {
        annotations:
          (this.hybridContext.hybridWorkloadModel.value.spec?.annotations as IItemizedListItem[]) || undefined,
        labels: (this.hybridContext.hybridWorkloadModel.value.spec?.labels as IItemizedListItem[]) || undefined,
      };
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        environmentVariables: this.policy?.defaults?.environmentVariables,
        annotations: this.policy?.defaults?.annotations,
        labels: this.policy?.defaults?.labels,
        tolerations: this.policy?.defaults?.tolerations,
      };
    },
    supportBackoffLimit(): boolean {
      return this.clusterStore.isClusterVersionSupportBackoffLimit(
        this.hybridContext.hybridWorkloadModel.value.clusterId,
      );
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
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
    noMasterStep(): boolean {
      return (
        this.hybridContext.hybridWorkloadModel.value.masterSpec === undefined ||
        this.hybridContext.hybridWorkloadModel.value.masterSpec === null
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.hybridContext.hybridWorkloadModel.value.clusterId,
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
  },
  methods: {
    onCredentialCreated(): void {
      this.$emit("credential-created");
    },
    updateParent(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.$emit("update", spec);
    },
    onNameChanged(name: string) {
      this.hybridContext.updateWorkloadName(name);
    },
    updateModelSpec(spec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.updateParent(spec);
    },
    updateModelStorage(storage: IUIStorageFields): void {
      const updatedSpec = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        storage,
      };
      this.updateParent(updatedSpec);
    },
    saveFormState(): void {
      this.$emit("save-form-state");
    },
    clearCompute(): void {
      this.hybridContext.clearComputeTemplate();
    },
    clearEnvironment(): void {
      this.hybridContext.clearEnvironmentTemplate();
    },
    resetEnvironment(): void {
      this.hybridContext.resetEnvironmentToDefault();
    },
    resetCompute(): void {
      this.hybridContext.resetComputeToDefault();
    },
    async onCancel(): Promise<void> {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.onLeave();
      }
    },
    onLeave(): void {
      this.$emit("leave-page");
    },
    onContinue(): void {
      this.saveFormState(); // Save worker form state before continuing
      this.$emit("continue");
    },
    onBack(): void {
      this.$emit("back-to-create-new");
    },
    updateGeneralModel(generalModel: IUIGeneralSectionModel): void {
      const updatedSpec = {
        ...this.hybridContext.hybridWorkloadModel.value.spec,
        annotations: generalModel.annotations,
        labels: generalModel.labels,
      };
      this.updateParent(updatedSpec);
    },
    onCreateNew(type: string): void {
      this.saveFormState();
      this.hybridContext.onCreateNewSetup(type, {
        distFramework: this.hybridContext.hybridWorkloadModel.value.spec.distributedFramework,
        workloadType: EWorkloadFormType.Training,
        workloadFormName: EDistributedReplicaType.Worker,
      });
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default DistributedHybridWorker as typeof DistributedHybridWorker & DefineComponent;
</script>
