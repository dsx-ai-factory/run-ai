<template>
  <section class="template-hybrid-new column items-center q-pt-md">
    <workload-form-wrapper
      submit-btn-text="Continue"
      cancel-btn-text="Cancel"
      :form-state="inference"
      :submitting="false"
      @submit="onContinue"
      @canceled="onCancel"
      @leave-page="onLeavePage"
    >
      <runai-scope-section
        kind-of-asset="template"
        :read-only="false"
        :selected-cluster-id="inference.clusterId"
        :scope-model="scopeModel"
        :allowed-scopes="allowedCreateScopes"
        @changed-scope="updateScope"
      />

      <inference-server
        v-if="inference.clusterId && showInferenceServerSelection && inference.serverType"
        :is-selectable="true"
        :model-value="inference.serverType"
        :cluster-id="inference.clusterId"
        @update-section-model="inferenceServerSelectionChanged"
      />

      <workload-template-name-section
        support-description
        :model-value="inference.templateName"
        :scope-model="scopeModel"
        :description="inference.templateDescription"
        @update:model-value="updateTemplateName"
        @update:description="updateTemplateDescription"
      />
    </workload-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// Components
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { WorkloadTemplateNameSection } from "@/components/workload-template/workload-template-name-section";
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import { InferenceServer } from "@/components/section/inference-server";

// Stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useSettingStore } from "@/stores/setting.store";
import { useInferenceStore } from "@/stores/inference.store";

// Utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";

// Models
import type { IScopeModel } from "@/models/global.model";
import { Action, ScopeType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import type { IUIDistributed, IUIInferenceTemplateCreation } from "@/models/workload.model";
import { SettingKeys } from "@/models/setting.model";
import { EServerType, type TInferenceServerSelection } from "@/models/inference.model";

import { MIN_CLUSTER_VERSION_FOR_NIM } from "@/common/version.constant";
import { scopeUtil } from "@/utils/scope.util/scope.util";
import { requestToLeave } from "@/services/infra/router.service/router.service";

export default defineComponent({
  name: "template-inference-new",
  components: {
    RunaiScopeSection,
    WorkloadTemplateNameSection,
    WorkloadFormWrapper,
    InferenceServer,
  },
  data() {
    return {
      inferenceStore: useInferenceStore(),
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      settingStore: useSettingStore(),
      clusterStore: useClusterStore(),
      AssetKind,
      scopeModel: {} as IScopeModel,
      selectedDistributedDetails: null as IUIDistributed | null,
      inference: inferenceWorkloadUtil.getEmptyUIInferenceCreationForTemplate() as IUIInferenceTemplateCreation,
      showInferenceServerSelection: false as boolean,
    };
  },
  created() {
    this.showInferenceServerSelection =
      this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog) ||
      this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
    const inferenceTemplateModel = this.inferenceStore.getTemplateModel();
    this.inference.clusterId = inferenceTemplateModel?.clusterId || "";

    if (inferenceTemplateModel?.scopeType && inferenceTemplateModel.scopeId) {
      this.scopeModel = scopeUtil.getScopeModelFromTypeAndId(
        inferenceTemplateModel.scopeType as ScopeType,
        inferenceTemplateModel.scopeId as string,
      );
    }

    this.inference.templateName = inferenceTemplateModel?.name || "";
    this.inference.templateDescription = inferenceTemplateModel?.description || "";

    // when we back from the template page
    if (this.inferenceStore.inferenceFormUIState.modelBasedOn) {
      this.inference.serverType = this.inferenceStore.inferenceFormUIState.modelBasedOn;
    }

    this.appStore.setPageLoading(false);
  },
  computed: {
    allowedCreateScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    selectedTemplateId() {
      return this.$route.params.templateId;
    },
  },
  methods: {
    updateTemplateName(name: string) {
      this.inference.templateName = name;
    },
    updateScope(scope: IScopeModel) {
      this.scopeModel = scope;
      this.inference.clusterId = scope.parentClusterUuid || scope.clusterId || "";
      if (this.inference.clusterId) {
        this.updateInferenceServerType();
      } else {
        this.inference.serverType = EServerType.Custom;
      }
    },
    async onCancel() {
      const allowToLeave: boolean = await requestToLeave();
      if (allowToLeave) {
        this.onLeavePage();
      }
    },
    onLeavePage(): void {
      this.inferenceStore.clearAllStates();
      this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX });
    },
    onContinue() {
      this.inferenceStore.clearAllStates();
      this.inferenceStore.setTemplateModel({
        name: this.inference.templateName,
        description: this.inference.templateDescription,
        scopeId: scopeUtil.getScopeId(this.scopeModel),
        clusterId: this.inference.clusterId,
        scopeType: this.scopeModel.scope ? (this.scopeModel.scope as ScopeType) : undefined,
        projectId:
          this.scopeModel.scope === ScopeType.Project && this.scopeModel?.projectId
            ? this.scopeModel.projectId.toString()
            : "",
        spec: {},
      });

      this.inferenceStore.setInferenceFormUIState({
        modelBasedOn: this.inference.serverType,
      });

      switch (this.inference.serverType) {
        case EServerType.TGI:
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_TGI });
          break;
        case EServerType.VLLM:
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_VLLM });
          break;
        case EServerType.Nim:
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NIM });
          break;
        case EServerType.Custom:
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW });
          break;
        default:
          console.error("Invalid inference type");
          break;
      }
    },
    updateTemplateDescription(description: string) {
      this.inference.templateDescription = description;
    },
    inferenceServerSelectionChanged(selection: TInferenceServerSelection): void {
      this.inference.serverType = selection;
    },
    updateInferenceServerType(): void {
      const isNimSupported =
        this.clusterStore.isClusterVersionSufficient(this.inference.clusterId, MIN_CLUSTER_VERSION_FOR_NIM) &&
        this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
      const isHuggingFaceSupported = this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog);

      if (!isNimSupported && this.inference.serverType === EServerType.Nim) {
        if (isHuggingFaceSupported) {
          this.inference.serverType = EServerType.VLLM;
        } else {
          this.inference.serverType = EServerType.Custom;
        }
      }
    },
  },
});
</script>
