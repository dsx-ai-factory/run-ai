<template>
  <section class="template-hybrid-new column items-center q-pt-md">
    <workload-form-wrapper
      submit-btn-text="Continue"
      cancel-btn-text="Cancel"
      :form-state="{ selectedTemplateId, templateName }"
      :submitting="false"
      @submit="onContinue"
      @canceled="onCancel()"
      @leave-page="onCancel()"
    >
      <runai-scope-section
        kind-of-asset="template"
        :read-only="false"
        :selected-cluster-id="selectedCluster"
        :scope-model="scopeModel"
        :allowed-scopes="allowedCreateScopes"
        @changed-scope="updateScope"
      />

      <multi-node-section
        v-if="showMultiNodeSection"
        is-selectable
        :cluster-id="selectedCluster"
        :multi-node-section-model="selectedDistributedDetails"
        @update-multi-node-section-model="updateSelectedDistributedDetails"
      />

      <workload-template-name-section
        :model-value="templateName"
        :scope-model="scopeModel"
        :description="templateDescription"
        support-description
        @update:model-value="updateTemplateName"
        @update-description="updateTemplateDescription"
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
import { MultiNodeSection } from "@/components/section/multi-node-section";
// Stores
import { useAppStore } from "@/stores/app.store";
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
// Utils
import { scopeUtil } from "@/utils/scope.util";
// Models
import type { IScopeModel } from "@/models/global.model";
import { Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import type { LocationQueryRaw } from "vue-router";
import type { IUIDistributed } from "@/models/workload.model";
import { isFlatTemplateNameUnique } from "@/common/form.validators";
import type { DistributedFramework, ScopeType } from "@/swagger-models/workloads-client";
export default defineComponent({
  name: "template-hybrid-new",
  components: {
    RunaiScopeSection,
    WorkloadTemplateNameSection,
    WorkloadFormWrapper,
    MultiNodeSection,
  },
  data() {
    return {
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      selectedCluster: "",
      templateName: "",
      templateDescription: "",
      AssetKind,
      scopeModel: {} as IScopeModel,
      selectedDistributedDetails: null as IUIDistributed | null,
      isFlatTemplateNameUnique,
    };
  },
  created() {
    this.appStore.setPageLoading(false);
    this.selectedCluster = this.$route.query.selectedClusterId?.toString() || "";

    this.setPageDataFromRouteState();
  },
  computed: {
    nextRouteName(): string {
      switch (this.$route.name) {
        case WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW:
          return WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID;
        case WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW:
          if (this.selectedDistributedDetails) {
            return WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED;
          }
          return WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID;
        default:
          return "";
      }
    },
    templateType(): string {
      switch (this.$route.name) {
        case WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW:
          return "workspace";
        case WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW:
          return "training";
        default:
          return "";
      }
    },
    showMultiNodeSection(): boolean {
      return this.templateType === "training";
    },
    allowedCreateScopes(): PermittedScopes {
      return this.workloadTemplateStore.actionPermissionsByScopes[Action.Create];
    },
    selectedTemplateId() {
      return this.$route.params.templateId;
    },
  },
  methods: {
    updateTemplateName(name: string) {
      this.templateName = name;
    },
    updateScope(scope: IScopeModel) {
      this.scopeModel = scope;
      this.selectedCluster = scope.parentClusterUuid || scope.clusterId || "";
    },
    onCancel() {
      this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INDEX });
    },
    onContinue() {
      const scopeId = scopeUtil.getScopeId(this.scopeModel);

      const routeQuery: LocationQueryRaw = {
        scope: this.scopeModel.scope,
        scopeId,
        name: this.templateName,
        description: this.templateDescription,
        clusterId: this.selectedCluster,
      };

      if (this.selectedDistributedDetails) {
        routeQuery["distFramework"] = this.selectedDistributedDetails.distFramework;
        if (this.selectedDistributedDetails.noMaster)
          routeQuery["noMaster"] = this.selectedDistributedDetails.noMaster?.toString();
      }

      this.$router.push({ name: this.nextRouteName, query: routeQuery });
    },
    updateTemplateDescription(description: string) {
      this.templateDescription = description;
    },
    updateSelectedDistributedDetails(distributed: IUIDistributed): void {
      this.selectedDistributedDetails = distributed || null;
    },
    setPageDataFromRouteState(): void {
      const pageData = this.$router.options.history.state?.pageData as Record<string, string | null | undefined>;
      if (pageData) {
        this.templateName = pageData.name || "";
        this.templateDescription = pageData.description || "";
        this.selectedCluster = pageData.clusterId || "";
        this.scopeModel = scopeUtil.getScopeModelFromTypeAndId(pageData.scopeType as ScopeType, pageData.scopeId || "");
        if (pageData.distFramework) {
          this.selectedDistributedDetails = {
            distFramework: pageData.distFramework as DistributedFramework,
            noMaster: pageData.noMaster === "true",
          };
        }
      }
    },
  },
});
</script>
