<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-submit-form column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :form-state="modelValue"
      :submitting="uiState.submitting"
      :submit-btn-text="submitBtnText"
      cancel-btn-text="Cancel"
      init-dirty
      :blurred="uiState.isDrawerOpen"
      :greedy="isGreedy"
      @unblur="closeDrawer"
      @leave-page="$emit(EEmitName.LeavePage)"
      @canceled="$emit(EEmitName.Canceled)"
      @submit="$emit(EEmitName.Submit)"
    >
      <title-and-back-section
        v-if="!uiState.isWorkloadTemplate"
        label="Cluster"
        :summary="navigation.clusterName || ''"
        back-tooltip="Change cluster"
        @on-back="
          $emit(EEmitName.NavigationAction, {
            action: EInferenceSubmitFormNavigationActionType.ChangeSection,
            section: EInferenceSubmitFormNavigationSection.Cluster,
          })
        "
      />

      <title-and-back-section
        v-if="!uiState.isWorkloadTemplate"
        label="Project"
        :summary="navigation.projectName || ''"
        back-tooltip="Change project"
        aid="project-section"
        @on-back="
          $emit(EEmitName.NavigationAction, {
            action: EInferenceSubmitFormNavigationActionType.ChangeSection,
            section: EInferenceSubmitFormNavigationSection.Project,
          })
        "
      />

      <runai-scope-section
        v-if="uiState.isWorkloadTemplate"
        :read-only="!!uiState.isEditMode"
        :init-closed="isValidScopeSelected"
        kind-of-asset="template"
        :scope-model="scopeModel"
        :allowed-scopes="allowedScopes"
        :key="modelValue.clusterId"
        :selected-cluster-id="modelValue.clusterId"
        :back-tooltip="uiState.isEditMode ? '' : 'Change scope'"
        @changed-scope="updateScope"
        @on-back="
          $emit(EEmitName.NavigationAction, {
            action: EInferenceSubmitFormNavigationActionType.ChangeSection,
            section: EInferenceSubmitFormNavigationSection.Scope,
          })
        "
      />
      <q-linear-progress animation-speed="100" indeterminate v-if="uiState.scopeChangeInProgress" />
      <transition-group name="fade">
        <template v-if="isValidScopeSelected && !uiState.scopeChangeInProgress">
          <workload-template-name-section
            v-if="uiState.isWorkloadTemplate"
            aid="template-name-section"
            support-description
            :model-value="modelValue.name"
            @update:model-value="handleNameChanged"
            :default-opened="!!$route.query.fromCopyId"
            :scope-model="scopeModel"
            :description="modelValue.description"
            :original-name="uiState.originalTemplateName"
            @update:description="handleDescriptionUpdate"
          />

          <title-and-back-section
            v-if="navigation.showInferenceType"
            label="Inference server"
            :summary="navigation.typeSummary"
            :back-tooltip="uiState.isEditMode ? '' : 'Change server'"
            @on-back="
              $emit(EEmitName.NavigationAction, {
                action: EInferenceSubmitFormNavigationActionType.ChangeSection,
                section: EInferenceSubmitFormNavigationSection.InferenceType,
              })
            "
          />

          <workload-name-section
            v-if="!uiState.isWorkloadTemplate"
            :default-open="navigation.defaultOpenNameSection"
            entity-type="inference"
            :name="modelValue.name"
            @update:name="handleNameChanged"
            :project-id="+modelValue.projectId"
            :cluster-id="modelValue.clusterId"
          />

          <title-and-back-section
            v-if="navigation.modelName"
            label="Model"
            :summary="navigation.modelName"
            :back-tooltip="navigation.canChangeModel ? 'Change model' : undefined"
            aid="model-section"
            @on-back="
              $emit(EEmitName.NavigationAction, {
                action: EInferenceSubmitFormNavigationActionType.ChangeSection,
                section: EInferenceSubmitFormNavigationSection.Model,
              })
            "
          />

          <hybrid-environment-section
            :key="`he-${sectionTemplates.selectedTemplates.environment.sectionKey}`"
            :render-key="sectionTemplates.selectedTemplates.environment.renderKey.toString()"
            :selected-name="sectionTemplates.selectedEnvironmentName"
            :workload-spec="modelValue.spec"
            :environment-variable-source-value-options="environmentVariablesSourceValueOptions"
            :policy="policy?.policyData"
            :cluster-id="modelValue.clusterId"
            :support-config-map-as-env-var="supportConfigMapAsEnvVar"
            :support-user-credentials="supportUserCredentials"
            :supporting-serving-port-access="supportingServingPortAccess"
            :support-image-pull-secrets="features.supportImagePullSecrets"
            :required-fields="{ image: !uiState.isWorkloadTemplate, port: !uiState.isWorkloadTemplate }"
            @load-section-templates="$emit(EEmitName.DisplayEnvironmentsList)"
            @update="handleWorkloadSpecUpdate"
            @reset="
              $emit(EEmitName.SectionAction, { action: EInferenceSubmitFormSectionActionType.ResetEnvironmentToDefault })
            "
            @clear="
              $emit(EEmitName.SectionAction, { action: EInferenceSubmitFormSectionActionType.ClearEnvironmentTemplate })
            "
            supporting-serving-port
            :support-linking="uiState.isWorkloadTemplate"
            :is-linked="!!sectionTemplates.selectedTemplates.environment.isLinked"
            :validate-empty-env-vars="!uiState.isWorkloadTemplate"
            @toggle-linked="$emit(EEmitName.ToggleLinked, EInferenceSubmitFormLinkActionType.Environment)"
            support-authenticated-ngc
            :scope-id="modelValue.scopeId?.toString() || ''"
            :scope-type="modelValue.scopeType"
            image-title="Set the workload image"
          />

          <hybrid-compute-resource-section
            :key="`hc-${sectionTemplates.selectedTemplates['compute-resource'].sectionKey}`"
            :render-key="sectionTemplates.selectedTemplates['compute-resource'].renderKey.toString()"
            :selected-name="sectionTemplates.selectedComputeName"
            :workload-spec="modelValue.spec"
            :project-node-types="clusterConfig.selectedProjectNodeTypes"
            :node-pools-resources="clusterConfig.projectNodePoolsResources"
            :policy="policy?.policyData"
            :support-multi-fraction-gpu="supportMultiFractionGpu"
            :cluster-id="modelValue.clusterId"
            :auto-scale-options="clusterConfig.autoScaleOptions"
            :support-tolerations="supportTolerations"
            :support-linking="uiState.isWorkloadTemplate"
            :is-linked="!!sectionTemplates.selectedTemplates['compute-resource'].isLinked"
            section-sub-title="Set the resources for each replica"
            @load-section-templates="$emit(EEmitName.DisplayComputesList)"
            @update="handleWorkloadSpecUpdate"
            @reset="
              $emit(EEmitName.SectionAction, { action: EInferenceSubmitFormSectionActionType.ResetComputeToDefault })
            "
            @clear="
              $emit(EEmitName.SectionAction, { action: EInferenceSubmitFormSectionActionType.ClearComputeTemplate })
            "
            :hide-specific-sections="{
              nodes: uiState.isWorkloadTemplate,
              affinity: uiState.isWorkloadTemplate,
            }"
            @toggle-linked="$emit(EEmitName.ToggleLinked, EInferenceSubmitFormLinkActionType.Compute)"
          />

          <hybrid-data-source-section
            :default-opened="shouldOpenDataSourceSection"
            :workload-storage="modelValue.spec.storage"
            :policy-rules="policy?.policyData.rules?.storage"
            :policy-defaults="policy?.policyData.defaults?.storage"
            @load-section-templates="$emit(EEmitName.DisplayDataSourcesList)"
            @update="handleModelStorageUpdate"
            :secrets="dataSources.secrets"
            :allowed-data-source-types="inferenceSupportedDataSources"
            :cluster-id="modelValue.clusterId"
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
            :policy-rules="policy?.policyData.rules"
            :itemize-defaults="generalSectionItemizeDefaults"
            :workload-priority-data="workloadPriorityData"
            :support-preemptibility="features.supportPreemptibility"
            :single-pod-workload="isSinglePodWorkload"
            @general-model-changed="handleGeneralModelChanged"
            :show-request-timeout="features.showRequestTimeout"
          />
        </template>
      </transition-group>
    </workload-form-wrapper>
    <workload-creation-drawer
      :allowed-data-source-options="inferenceSupportedDataSources"
      :is-open="uiState.isDrawerOpen"
      :comply-to="drawerFilters"
      :entity="uiState.drawerDataEntity"
      :policy-details="policy?.policyViewDetails"
      :created-entity-id="uiState.recentlyCreatedEntityId"
      :cluster-id="modelValue.clusterId"
      @close="handleDrawerAction({ action: EInferenceSubmitFormDrawerActionType.Close })"
      @select="handleDrawerAction({ action: EInferenceSubmitFormDrawerActionType.Select, selectData: $event })"
      @select-multiple="
        handleDrawerAction({ action: EInferenceSubmitFormDrawerActionType.SelectMultiple, selectMultipleData: $event })
      "
      @create-new="handleDrawerAction({ action: EInferenceSubmitFormDrawerActionType.CreateNew, createNewData: $event })"
      :apply-item-modifications-after-loaded="applyItemModificationsAfterLoaded"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType, type DefineComponent } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { TitleAndBackSection } from "@/components/section/title-and-back-section";
import { WorkloadNameSection } from "@/components/section/workload-name-section";
import { HybridEnvironmentSection } from "@/components/workload-creation/hybrid-environment-section";
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { GeneralSection } from "@/components/section/general-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { WorkloadTemplateNameSection } from "@/components/workload-template/workload-template-name-section";

// models
import { EWorkloadFormType } from "@/models/workload.model";
import type {
  IInferenceSubmitFormClusterConfig,
  IInferenceSubmitFormDataSources,
  IInferenceSubmitFormDrawerAction,
  IInferenceSubmitFormFeatures,
  IInferenceSubmitFormNavigation,
  IInferenceSubmitFormPolicy,
  IInferenceSubmitFormSectionAction,
  IInferenceSubmitFormSectionTemplates,
  IInferenceSubmitFormNavigationAction,
  IInferenceSubmitFormUIState,
} from "./inference-submit-form.types";
import {
  EInferenceSubmitFormNavigationActionType,
  EInferenceSubmitFormNavigationSection,
  EInferenceSubmitFormDrawerActionType,
  EInferenceSubmitFormSectionActionType,
  EInferenceSubmitFormLinkActionType,
} from "./inference-submit-form.types";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { IAssetsFilter } from "@/models/filter.model";
import type { TDataSourceKinds } from "@/models/data-source.model";
import type { IUIGeneralSectionModel } from "@/components/section/general-section";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { ScopeType } from "@/swagger-models/assets-service-client";
import { AssetKind, Scope } from "@/swagger-models/assets-service-client";
import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { IScopeModel } from "@/models/global.model";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { WorkloadCreationDrawerEntities, type IDrawerListItem } from "@/models/workload-creation-drawer-model";

// stores
import { useClusterStore } from "@/stores/cluster.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useAuthStore } from "@/stores/auth.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useSettingStore } from "@/stores/setting.store";

// utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { inferenceDataSourceRestrictionsUtil } from "@/utils/inference-data-source-restrictions.util/inference-data-source-restrictions.util";

// consts
import {
  MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
  MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
  MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
  MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
} from "@/common/version.constant";
import { EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import { useRoute } from "vue-router";

// validators
import { scopeUtil } from "@/utils/scope.util";

export enum EEmitName {
  UpdateModelValue = "update:model-value",
  LeavePage = "leave-page",
  Canceled = "canceled",
  Submit = "submit",
  DisplayEnvironmentsList = "display-environments-list",
  DisplayComputesList = "display-computes-list",
  DisplayDataSourcesList = "display-data-sources-list",
  DrawerAction = "drawer-action",
  SectionAction = "section-action",
  NavigationAction = "navigation-action",
  ToggleLinked = "toggle-linked",
  ScopeChanged = "scope-changed",
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

const InferenceSubmitForm = defineComponent({
  name: "inference-submit-form",
  components: {
    TitleAndBackSection,
    WorkloadFormWrapper,
    WorkloadNameSection,
    HybridEnvironmentSection,
    WorkloadCreationDrawer,
    HybridComputeResourceSection,
    HybridDataSourceSection,
    GeneralSection,
    RunaiScopeSection,
    WorkloadTemplateNameSection,
  },
  emits: {
    [EEmitName.UpdateModelValue]: (modelValue: IUISuperWorkloadCreationRequest) => modelValue,
    [EEmitName.ScopeChanged]: (scope: IScopeModel) => scope,
    [EEmitName.LeavePage]: () => true,
    [EEmitName.Canceled]: () => true,
    [EEmitName.Submit]: () => true,
    [EEmitName.DisplayEnvironmentsList]: () => true,
    [EEmitName.DisplayComputesList]: () => true,
    [EEmitName.DisplayDataSourcesList]: () => true,
    [EEmitName.DrawerAction]: (action: IInferenceSubmitFormDrawerAction) => action,
    [EEmitName.SectionAction]: (action: IInferenceSubmitFormSectionAction) => action,
    [EEmitName.NavigationAction]: (action: IInferenceSubmitFormNavigationAction) => action,
    [EEmitName.ToggleLinked]: (section: EInferenceSubmitFormLinkActionType) => section,
  },
  props: {
    modelValue: {
      type: Object as PropType<IUISuperWorkloadCreationRequest>,
      required: true,
    },
    navigation: {
      type: Object as PropType<IInferenceSubmitFormNavigation>,
      required: true,
    },
    policy: {
      type: Object as PropType<IInferenceSubmitFormPolicy | undefined>,
      required: false,
    },
    sectionTemplates: {
      type: Object as PropType<IInferenceSubmitFormSectionTemplates>,
      required: true,
    },
    environmentVariablesSourceValueOptions: {
      type: Array as PropType<TEnvVarNoneCustomSourceOption[]>,
      default: () => [],
    },
    clusterConfig: {
      type: Object as PropType<IInferenceSubmitFormClusterConfig>,
      required: true,
    },
    dataSources: {
      type: Object as PropType<IInferenceSubmitFormDataSources>,
      required: true,
    },
    features: {
      type: Object as PropType<IInferenceSubmitFormFeatures>,
      required: true,
    },
    uiState: {
      type: Object as PropType<IInferenceSubmitFormUIState>,
      required: true,
    },
    project: {
      type: Object as PropType<Project>,
      required: true,
    },
    triggerValidationOnStartUp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    let submitBtnText = "Create Inference";

    if (this.uiState.isWorkloadTemplate) {
      submitBtnText = this.uiState.isEditMode ? "Save Template" : "Create Template";
    }

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
      EInferenceSubmitFormNavigationActionType,
      EInferenceSubmitFormNavigationSection,
      EInferenceSubmitFormDrawerActionType,
      EInferenceSubmitFormSectionActionType,
      EInferenceSubmitFormLinkActionType,
      submitBtnText,
      applyItemModificationsAfterLoaded: applyItemModificationsAfterLoaded,
      isGreedy: false,
    };
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
    allowedScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.modelValue.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
    supportUserCredentials(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.modelValue.clusterId || "",
        MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
      );
    },
    supportMultiFractionGpu(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.modelValue.clusterId,
        MIN_CLUSTER_VERSION_FOR_MULTI_FRACTION_GPU,
      );
    },
    supportsNfsAndHostPath(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.modelValue.clusterId,
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
    supportingServingPortAccess(): boolean {
      return this.inferenceStore.supportingServingPortAccess(this.modelValue.clusterId);
    },
    inferenceSupportedDataSources(): Set<TDataSourceKinds> {
      const inferenceSupported = inferenceWorkloadUtil.getInferenceSupportedDataSources({
        isVersion219OrHigher: this.supportsNfsAndHostPath,
        modelStoreSupportOnly: false,
      });

      return new Set(
        [...(this.dataSources.allowedDataSourceOptions || [])].filter((ds) => inferenceSupported.includes(ds)),
      );
    },
    supportTolerations(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.modelValue.clusterId,
        MIN_CLUSTER_VERSION_FOR_TOLERATIONS,
      );
    },
    drawerFilters(): IAssetsFilter {
      if (this.uiState.isWorkloadTemplate) {
        if (this.uiState.baseDrawerFilters) return this.uiState.baseDrawerFilters;
        let projectId = undefined;
        if (this.modelValue.scopeType === Scope.Project) {
          projectId = this.modelValue.scopeId ? +this.modelValue.scopeId : undefined;
        }
        return {
          projectId,
          complyToProject: +this.modelValue.projectId,
          complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
          isInference: true,
        };
      } else {
        return {
          projectId: +this.modelValue.projectId,
          complyToProject: +this.modelValue.projectId,
          complyToWorkloadType: EWorkloadTypeSupportedInAssetsService.Inference,
          isInference: true,
        };
      }
    },
    generalSectionModel(): IUIGeneralSectionModel {
      const generalSectionModel: IUIGeneralSectionModel = {
        annotations:
          this.modelValue.spec?.annotations?.map((annotation) => ({
            name: annotation.name || "",
            value: annotation.value || "",
          })) || undefined,
        labels:
          this.modelValue.spec?.labels?.map((label) => ({ name: label.name || "", value: label.value || "" })) ||
          undefined,
        priorityClass: this.modelValue.spec?.priorityClass || null,
        preemptibility: this.modelValue.spec?.preemptibility || null,
      };

      if (this.features.showRequestTimeout) {
        generalSectionModel.servingConfiguration = this.modelValue.spec?.servingConfiguration;
      }

      return generalSectionModel;
    },
    generalSectionItemizeDefaults(): ILoadedItemizeDefaults {
      return {
        annotations: this.policy?.policyData.defaults?.annotations || undefined,
        labels: this.policy?.policyData.defaults?.labels || undefined,
      };
    },
    isSinglePodWorkload(): boolean {
      return (
        this.modelValue.spec?.autoscaling?.minReplicas === 1 && this.modelValue.spec?.autoscaling?.maxReplicas === 1
      );
    },
    workloadPriorityData(): IWorkloadPriorityData {
      const totalGpuRequest = workloadHybridUtil.calculateTotalGpuRequest({
        specCompute: this.modelValue.spec?.compute,
        replicas: {
          min: this.modelValue.spec?.autoscaling?.minReplicas,
          max: this.modelValue.spec?.autoscaling?.maxReplicas,
        },
      });

      return {
        currProject: this.project,
        selectedNodePoolsResources: this.modelValue.spec?.nodePools || undefined,
        totalGpuRequest,
        workloadPriorities: this.clusterConfig.workloadPriorities,
        supportPriorityClass: this.features.supportPriorityClass,
        supportSemiPreemptibility: this.features.supportSemiPreemptibility,
      };
    },
    scopeModel(): IScopeModel {
      return scopeUtil.getScopeModelFromTypeAndId(
        this.modelValue.scopeType as ScopeType,
        this.modelValue.scopeId as string,
      );
    },
    isValidScopeSelected(): boolean {
      if (!this.uiState.isWorkloadTemplate) return true;
      return this.uiState.isValidScopeSelected || false;
    },
  },
  methods: {
    closeDrawer() {
      this.$emit(EEmitName.DrawerAction, { action: EInferenceSubmitFormDrawerActionType.Close });
    },
    handleDrawerAction(drawerAction: IInferenceSubmitFormDrawerAction) {
      this.$emit(EEmitName.DrawerAction, drawerAction);
    },
    handleGeneralModelChanged(generalModel: IUIGeneralSectionModel): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        spec: {
          ...this.modelValue.spec,
          ...generalModel,
        },
      });
    },
    handleWorkloadSpecUpdate(workloadSpec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        spec: workloadSpec,
      });
    },
    handleNameChanged(name: string): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        name,
      });
    },
    handleModelStorageUpdate(storage: IUISuperWorkloadCreationRequest["spec"]["storage"]): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        spec: {
          ...this.modelValue.spec,
          storage,
        },
      });
    },
    handleDescriptionUpdate(description: string): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        description,
      });
    },
    async updateScope(scope: IScopeModel): Promise<void> {
      this.$emit(EEmitName.ScopeChanged, scope);
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default InferenceSubmitForm as typeof InferenceSubmitForm & DefineComponent;
</script>
