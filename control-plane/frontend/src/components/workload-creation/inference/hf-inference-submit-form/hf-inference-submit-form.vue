<template>
  <section class="hf-inference-submit-form column items-center q-pt-md">
    <workload-form-wrapper
      :form-state="modelValue"
      :submitting="uiState.submitting"
      init-dirty
      :blurred="uiState.isDrawerOpen"
      @unblur="closeDrawer"
      @leave-page="$emit(EEmitName.LeavePage)"
      :ref="workloadFormWrapper"
    >
      <runai-scope-section
        v-if="uiState.isWorkloadTemplate"
        :read-only="false"
        kind-of-asset="template"
        :init-closed="isValidScopeSelected"
        :scope-model="scopeModel"
        :allowed-scopes="allowedScopes"
        :key="modelValue.clusterId"
        :selected-cluster-id="modelValue.clusterId"
        back-tooltip="Change scope"
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
            v-if="!uiState.isWorkloadTemplate"
            label="Cluster"
            :summary="navigation.clusterName || ''"
            back-tooltip="Change cluster"
            @on-back="$emit(EEmitName.Back)"
          />

          <title-and-back-section
            v-if="!uiState.isWorkloadTemplate"
            label="Project"
            :summary="navigation.projectName || ''"
            back-tooltip="Change project"
            aid="project-section"
            @on-back="$emit(EEmitName.Back)"
          />

          <title-and-back-section
            v-if="navigation.showInferenceType"
            label="Inference server"
            :summary="navigation.typeSummary"
            back-tooltip="Change server"
            @on-back="$emit(EEmitName.Back)"
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

          <hybrid-hugging-face-model-section
            :workload-spec="modelValue.spec"
            :cluster-id="modelValue.clusterId"
            :policy="policy?.policyData"
            :ui-state="uiState.uiState"
            @update="handleWorkloadSpecUpdate"
            @update:ui-state="$emit(EEmitName.UpdateUIState, $event)"
            :scope-model="scopeModel"
            :required-fields="{
              accessModel: !uiState.isWorkloadTemplate,
              port: !uiState.isWorkloadTemplate,
              profile: !uiState.isWorkloadTemplate,
            }"
            :show-create-new-personal-credentials-btn="!uiState.isWorkloadTemplate"
            :show-personal-credentials-option="!uiState.isWorkloadTemplate"
            support-authenticated-ngc
          />

          <hybrid-compute-resource-section
            :key="`hc-${sectionTemplates.selectedTemplates['compute-resource'].sectionKey}`"
            :render-key="sectionTemplates.selectedTemplates['compute-resource'].renderKey.toString()"
            :selected-name="sectionTemplates.selectedComputeName"
            :workload-spec="modelValue.spec"
            :project-node-types="clusterConfig.selectedProjectNodeTypes"
            :node-pools-resources="clusterConfig.projectNodePoolsResources"
            :policy="policy?.policyData"
            :support-multi-fraction-gpu="features.supportMultiFractionGpu"
            :cluster-id="modelValue.clusterId"
            :auto-scale-options="clusterConfig.autoScaleOptions"
            :support-tolerations="features.supportTolerations"
            :support-linking="uiState.isWorkloadTemplate"
            :is-linked="!!sectionTemplates.selectedTemplates['compute-resource'].isLinked"
            section-sub-title="Set the resources for each replica"
            @load-section-templates="$emit(EEmitName.DisplayComputesList)"
            @update="handleWorkloadSpecUpdate"
            @reset="$emit(EEmitName.ResetCompute)"
            @clear="$emit(EEmitName.ClearComputeTemplate)"
            :hide-specific-sections="{
              nodes: uiState.isWorkloadTemplate,
              affinity: uiState.isWorkloadTemplate,
            }"
            @toggle-linked="$emit(EEmitName.ToggleLinked)"
          />

          <hybrid-data-source-section
            :default-opened="dataSources.shouldOpenModelStoreSection"
            :workload-storage="modelValue.spec.storage"
            :policy-rules="policy?.policyData?.rules?.storage"
            :policy-defaults="policy?.policyData?.defaults?.storage"
            @load-section-templates="$emit(EEmitName.DisplayModelStorageList)"
            @update="handleModelStorageUpdate"
            :secrets="dataSources.secrets"
            :allowed-data-source-types="dataSources.inferenceSupportedDataSources"
            :cluster-id="modelValue.clusterId"
            :show-volume-persistency-section="false"
            :is-volumes-enabled="false"
            custom-section-label="Model store"
            :required="uiState.uiState.accessFromStorage"
            :key="`${uiState.uiState.accessFromStorage}`"
            :can-add-new="false"
          >
            <template #datasource-title>
              <runai-section-title
                title="Select the data source that will serve as a model store"
                tooltip-text="If the model is already stored on the selected data source, it will be loaded from there automatically. Otherwise, it will be stored on the selected data source during the first workload deployment."
                width="350px"
              />
            </template>

            <template #empty-state-title>
              <div class="text-black-54">No data and storage yet.</div>
              <div class="text-black-54">Click the button above to load an existing setup</div>
            </template>
          </hybrid-data-source-section>
        </template>
      </transition-group>

      <template #actions>
        <runai-btn
          flat
          color="primary"
          label="Cancel"
          class="q-mr-sm"
          @click="$emit(EEmitName.Cancel)"
          aid="cancel-btn"
        />
        <runai-dropdown-btn-select
          :options="submitOptions"
          @on-select="onSubmit"
          aid="submit-btn"
          :selected-value="''"
          size="md"
          :flat="false"
          split
          :label="submitOptions[0].label"
          color="primary"
          :no-icon-animation="false"
          :loading="uiState.submitting"
          tooltip-position="right"
        />
      </template>
    </workload-form-wrapper>
    <workload-creation-drawer
      :allowed-data-source-options="dataSources.inferenceSupportedDataSources"
      :is-open="uiState.isDrawerOpen"
      :comply-to="uiState.drawerFilters"
      :entity="uiState.drawerDataEntity"
      :policy-details="policy?.policyViewDetails"
      :created-entity-id="uiState.recentlyCreatedEntityId"
      :cluster-id="modelValue.clusterId"
      @close="handleDrawerAction({ action: EHfInferenceSubmitFormDrawerActionType.Close })"
      @select="handleDrawerAction({ action: EHfInferenceSubmitFormDrawerActionType.Select, selectData: $event })"
      @select-multiple="
        handleDrawerAction({
          action: EHfInferenceSubmitFormDrawerActionType.SelectMultiple,
          selectMultipleData: $event,
        })
      "
      @create-new="
        handleDrawerAction({ action: EHfInferenceSubmitFormDrawerActionType.CreateNew, createNewData: $event })
      "
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
import { HybridComputeResourceSection } from "@/components/workload-creation/hybrid-compute-resource-section";
import { WorkloadCreationDrawer } from "@/components/workload-creation/workload-creation-drawer";
import { HybridDataSourceSection } from "@/components/workload-creation/hybrid-data-source-section";
import { HybridHuggingFaceModelSection } from "@/components/workload-creation/hybrid-hugging-face-model-section";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiDropdownBtnSelect } from "@/components/common/runai-dropdown-btn-select";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { WorkloadTemplateNameSection } from "@/components/workload-template/workload-template-name-section";

// types
import type { IScopeModel } from "@/models/global.model";
import type {
  IHfInferenceSubmitFormNavigation,
  IHfInferenceSubmitFormPolicy,
  IHfInferenceSubmitFormSectionTemplates,
  IHfInferenceSubmitFormClusterConfig,
  IHfInferenceSubmitFormDataSources,
  IHfInferenceSubmitFormFeatures,
  IHfInferenceSubmitFormUIState,
  IHfInferenceSubmitFormDrawerAction,
  IHfInferenceSubmitFormNavigationAction,
} from "./hf-inference-submit-form.types";
import { EHfInferenceSubmitFormDrawerActionType } from "./hf-inference-submit-form.types";
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import { ScopeType } from "@/swagger-models/assets-service-client";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import {
  EInferenceSubmitFormNavigationActionType,
  EInferenceSubmitFormNavigationSection,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import { SUBMIT_OPTIONS } from "./hf-inference-submit-form.types";

// stores
import { useAppStore } from "@/stores/app.store";
import { useInferenceStore } from "@/stores/inference.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useDataSourceStore } from "@/stores/data-source.store";
import { useWorkloadStore } from "@/stores/workload.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";

// utils
import { scopeUtil } from "@/utils/scope.util";

// validators
import { WorkloadCreationDrawerEntities, type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { inferenceDataSourceRestrictionsUtil } from "@/utils/inference-data-source-restrictions.util/inference-data-source-restrictions.util";

export enum EEmitName {
  UpdateModelValue = "update:model-value",
  UpdateUIState = "update:ui-state",
  DisplayComputesList = "display-computes-list",
  ResetCompute = "reset-compute",
  ClearComputeTemplate = "clear-compute-template",
  DisplayModelStorageList = "display-model-storage-list",
  Cancel = "cancel",
  Submit = "submit",
  Back = "back",
  LeavePage = "leave-page",
  DrawerAction = "drawer-action",
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
    return inferenceDataSourceRestrictionsUtil.applyInferenceDataSourceRestrictions(items, clusterId || "", true);
  }
  return items;
}

export default defineComponent({
  name: "hf-inference-submit-form",
  components: {
    TitleAndBackSection,
    WorkloadFormWrapper,
    WorkloadNameSection,
    HybridComputeResourceSection,
    HybridDataSourceSection,
    WorkloadCreationDrawer,
    HybridHuggingFaceModelSection,
    RunaiSectionTitle,
    RunaiBtn,
    RunaiDropdownBtnSelect,
    RunaiScopeSection,
    WorkloadTemplateNameSection,
  },
  emits: {
    [EEmitName.UpdateModelValue]: (modelValue: IUISuperWorkloadCreationRequest) => modelValue,
    [EEmitName.UpdateUIState]: (uiState: IInferenceFormUIState) => uiState,
    [EEmitName.DisplayComputesList]: () => true,
    [EEmitName.ResetCompute]: () => true,
    [EEmitName.ClearComputeTemplate]: () => true,
    [EEmitName.DisplayModelStorageList]: () => true,
    [EEmitName.Cancel]: () => true,
    [EEmitName.Submit]: (submitType: string) => submitType,
    [EEmitName.Back]: () => true,
    [EEmitName.LeavePage]: () => true,
    [EEmitName.DrawerAction]: (action: IHfInferenceSubmitFormDrawerAction) => action,
    [EEmitName.NavigationAction]: (action: IHfInferenceSubmitFormNavigationAction) => action,
    [EEmitName.ToggleLinked]: () => true,
    [EEmitName.ScopeChanged]: (scope: IScopeModel) => scope,
  },
  props: {
    modelValue: {
      type: Object as PropType<IUISuperWorkloadCreationRequest>,
      required: true,
    },
    navigation: {
      type: Object as PropType<IHfInferenceSubmitFormNavigation>,
      required: true,
    },
    policy: {
      type: Object as PropType<IHfInferenceSubmitFormPolicy>,
      required: false,
    },
    sectionTemplates: {
      type: Object as PropType<IHfInferenceSubmitFormSectionTemplates>,
      required: true,
    },
    clusterConfig: {
      type: Object as PropType<IHfInferenceSubmitFormClusterConfig>,
      required: true,
    },
    dataSources: {
      type: Object as PropType<IHfInferenceSubmitFormDataSources>,
      required: true,
    },
    features: {
      type: Object as PropType<IHfInferenceSubmitFormFeatures>,
      required: true,
    },
    uiState: {
      type: Object as PropType<IHfInferenceSubmitFormUIState>,
      required: true,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      inferenceStore: useInferenceStore(),
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      dataSourceStore: useDataSourceStore(),
      workloadStore: useWorkloadStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      submitOptions: this.uiState.isWorkloadTemplate
        ? [SUBMIT_OPTIONS.submitTemplate, SUBMIT_OPTIONS.advancedSetup]
        : [SUBMIT_OPTIONS.submit, SUBMIT_OPTIONS.advancedSetup],
      workloadFormWrapper: "workloadFormWrapper" as string,
      EEmitName,
      EHfInferenceSubmitFormDrawerActionType,
      EInferenceSubmitFormNavigationActionType,
      EInferenceSubmitFormNavigationSection,
      applyItemModificationsAfterLoaded: applyItemModificationsAfterLoaded,
    };
  },
  computed: {
    allowedScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    scopeModel(): IScopeModel {
      if (!this.modelValue.scopeType || !this.modelValue.scopeId) {
        return scopeUtil.getScopeModelFromTypeAndId(ScopeType.Project, this.modelValue.projectId?.toString() || "");
      }
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
      this.handleDrawerAction({ action: EHfInferenceSubmitFormDrawerActionType.Close });
    },
    handleDrawerAction(drawerAction: IHfInferenceSubmitFormDrawerAction) {
      this.$emit(EEmitName.DrawerAction, drawerAction);
    },
    handleNameChanged(name: string): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        name,
      });
    },
    handleWorkloadSpecUpdate(workloadSpec: IUISuperWorkloadCreationRequest["spec"]): void {
      this.$emit(EEmitName.UpdateModelValue, {
        ...this.modelValue,
        spec: workloadSpec,
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
    async onSubmit(submitType: string): Promise<void> {
      const formElm = this.$refs.workloadFormWrapper as HTMLFormElement;
      const success = await formElm.validate();
      if (!success) {
        formElm.showHint();
        return;
      }

      this.$emit(EEmitName.Submit, submitType);
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
</script>
