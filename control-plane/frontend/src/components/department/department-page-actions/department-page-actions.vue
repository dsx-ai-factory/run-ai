<template>
  <runai-action-button v-if="isReadOnlyPermission" btn-action="review" @click="handleReviewAction">
    <runai-select-menu :list="availableReviewSteps" :loading="loadingSteps" @selected-value="onReviewStepSelected" />
  </runai-action-button>
  <runai-action-button
    v-else
    v-permission="{ resourceType: ResourceType.Department, action: Action.Update }"
    aid="edit-department-btn"
    btn-action="edit"
    @click="handleEditAction"
  >
    <runai-select-menu :list="availableEditSteps" :loading="loadingSteps" @selected-value="onStepSelected" />
  </runai-action-button>

  <runai-tooltip-wrapper
    :display-tooltip="disablePolicyButton"
    tooltip-text="There is no policy applied for the selected department"
  >
    <runai-action-button
      v-if="showPolicyButton"
      v-permission="{ resourceType: ResourceType.Department, action: Action.Read }"
      aid="view-department-policy-btn"
      btn-action="viewPolicy"
      :disable="disablePolicyButton"
    >
      <policy-dropdown @policy-selected="viewPolicy" entity-name="department" :existing-policies="policyMenuOptions" />
    </runai-action-button>
  </runai-tooltip-wrapper>
  <runai-tooltip-wrapper :display-tooltip="isDeleteDepartmentDisabled" :tooltip-text="deleteDepartmentDisabledTooltip">
    <runai-action-button :disable="isDeleteDepartmentDisabled" btn-action="delete" @click="$emit('open-delete-modal')" />
  </runai-tooltip-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

//util
import { type IPolicyScopeProperties, policyUtil } from "@/utils/policy.util/policy.util";
//svc
import { policyService } from "@/services/control-plane/policy.service/policy.service";
//routes
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
//store
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
//model
import { Action, ResourceType, ScopeType } from "@/swagger-models/authorization-client";
import { DEFAULT_DEPARTMENT_NAME, type IDepartmentTable } from "@/models/department.model";
import type { IPolicyMap } from "@/models/policy.model";
import type { ISelectOption } from "@/models/global.model";
import {
  type DistributedInferencePolicyV2,
  type DistributedPolicyDefaultsV2,
  type DistributedPolicyRulesV2,
  type DistributedPolicyV2,
  type InferencePolicyDefaultsAndRulesV2Defaults,
  type InferencePolicyDefaultsAndRulesV2Rules,
  type InferencePolicyV2,
  PolicyType,
  type TrainingPolicyDefaultsAndRulesV2Defaults,
  type TrainingPolicyDefaultsAndRulesV2Rules,
  type TrainingPolicyV2,
  type WorkspacePolicyDefaultsAndRulesV2Defaults,
  type WorkspacePolicyDefaultsAndRulesV2Rules,
  type WorkspacePolicyV2,
} from "@/swagger-models/policy-service-client";
//cmps
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { PolicyDropdown } from "@/components/policy/policy-dropdown";
import { RunaiSelectMenu } from "@/components/common/runai-select-menu";

export default defineComponent({
  name: "department-page-actions",
  components: { PolicyDropdown, RunaiTooltipWrapper, RunaiActionButton, RunaiSelectMenu },
  emits: [
    "open-access-rule-modal",
    "open-delete-modal",
    "access-rule-created",
    "access-rule-deleted",
    "edit-department",
    "handle-edit-action",
    "view-department",
    "handle-review-action",
  ],
  props: {
    selectedDepartment: {
      type: Object as PropType<IDepartmentTable | null>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    availableEditSteps: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    availableReviewSteps: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    loadingSteps: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data: function () {
    return {
      selectedDepartmentPolicies: null as IPolicyMap | null,
      loadingPolicies: false as boolean,
      loadedPoliciesMapByDepartmentId: {} as Record<string, IPolicyMap>,
      clusterStore: useClusterStore(),
      permissionStore: usePermissionStore(),
    };
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    isReadOnlyPermission(): boolean {
      return this.permissionStore.isReadOnly(ResourceType.Department);
    },
    showPolicyButton(): boolean {
      return useSettingStore().isPolicyManagerEnabled;
    },
    disablePolicyButton(): boolean {
      return this.loadingPolicies || !this.policyMenuOptions.length;
    },
    isDefaultDepartment(): boolean {
      return this.selectedDepartment?.name === DEFAULT_DEPARTMENT_NAME;
    },
    departmentHasChildren(): boolean {
      return !!this.selectedDepartment?.children?.length;
    },
    hasDeletePermission(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Department, Action.Delete);
    },
    isDeleteDepartmentDisabled(): boolean {
      return !this.hasDeletePermission || this.isDefaultDepartment || this.departmentHasChildren;
    },
    deleteDepartmentDisabledTooltip(): string {
      if (!this.hasDeletePermission) {
        return "The selected department can't be deleted because you're not authorized to delete it";
      }
      if (this.isDefaultDepartment) {
        return "The default department can't be deleted";
      }
      if (this.departmentHasChildren) {
        return "The selected department can't be deleted because its subordinate projects must be deleted first";
      }
      return "";
    },
    /*** Policy ***/
    policyMenuOptions(): Array<PolicyType> {
      const policyOptions: Array<PolicyType> = [];
      if (!this.selectedDepartmentPolicies) return policyOptions;

      this.hasInteractivePolicy && policyOptions.push(PolicyType.Workspace);
      this.hasTrainingPolicy && policyOptions.push(PolicyType.Training);
      this.hasDistributedPolicy && policyOptions.push(PolicyType.Distributed);
      this.hasInferencePolicy && policyOptions.push(PolicyType.Inference);

      return policyOptions;
    },
    hasInteractivePolicy(): boolean {
      if (!this.selectedDepartmentPolicies) return false;
      const rules: WorkspacePolicyDefaultsAndRulesV2Rules =
        this.selectedDepartmentPolicies[PolicyType.Workspace]?.effective?.rules || {};
      const defaults: WorkspacePolicyDefaultsAndRulesV2Defaults =
        this.selectedDepartmentPolicies[PolicyType.Workspace]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasTrainingPolicy(): boolean {
      if (!this.selectedDepartmentPolicies) return false;
      const rules: TrainingPolicyDefaultsAndRulesV2Rules =
        this.selectedDepartmentPolicies[PolicyType.Training]?.effective?.rules || {};
      const defaults: TrainingPolicyDefaultsAndRulesV2Defaults =
        this.selectedDepartmentPolicies[PolicyType.Training]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasDistributedPolicy(): boolean {
      if (!this.selectedDepartmentPolicies) return false;
      const rules: DistributedPolicyRulesV2 =
        this.selectedDepartmentPolicies[PolicyType.Distributed]?.effective?.rules || {};
      const defaults: DistributedPolicyDefaultsV2 =
        this.selectedDepartmentPolicies[PolicyType.Distributed]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasInferencePolicy(): boolean {
      if (!this.selectedDepartmentPolicies) return false;
      const rules: InferencePolicyDefaultsAndRulesV2Rules =
        this.selectedDepartmentPolicies[PolicyType.Inference]?.effective?.rules || {};
      const defaults: InferencePolicyDefaultsAndRulesV2Defaults =
        this.selectedDepartmentPolicies[PolicyType.Inference]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
  },
  methods: {
    handleEditAction(): void {
      this.$emit("handle-edit-action");
    },
    onStepSelected(stepName?: string): void {
      this.$emit("edit-department", stepName);
    },
    handleReviewAction(): void {
      this.$emit("handle-review-action");
    },
    onReviewStepSelected(stepName?: string): void {
      this.$emit("view-department", stepName);
    },
    /*** Policy ***/
    async loadSelectedDepartmentPolicy(): Promise<void> {
      if (!this.selectedDepartment?.id) {
        this.selectedDepartmentPolicies = null;
        return;
      }

      const alreadyLoaded = this.loadedPoliciesMapByDepartmentId[this.selectedDepartment.id];
      if (alreadyLoaded) {
        this.selectedDepartmentPolicies = alreadyLoaded;
        return;
      }

      try {
        this.loadingPolicies = true;
        const scopeProperties: IPolicyScopeProperties = {
          departmentId: this.selectedDepartment.id.toString(),
          scope: ScopeType.Department,
        };
        await Promise.all([
          policyService.loadPolicyByMetaAndType(PolicyType.Workspace, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Training, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Distributed, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Inference, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.DistributedInference, scopeProperties),
        ]).then((res) => {
          this.selectedDepartmentPolicies = {
            [PolicyType.Workspace]: res[0] as WorkspacePolicyV2,
            [PolicyType.Training]: res[1] as TrainingPolicyV2,
            [PolicyType.Distributed]: res[2] as DistributedPolicyV2,
            [PolicyType.Inference]: res[3] as InferencePolicyV2,
            [PolicyType.DistributedInference]: res[4] as DistributedInferencePolicyV2,
          };
          this.selectedDepartment?.id &&
            (this.loadedPoliciesMapByDepartmentId[this.selectedDepartment.id] = this.selectedDepartmentPolicies);
        });
      } catch (err: unknown) {
        console.error("Failed to load department policies", err);
      } finally {
        this.loadingPolicies = false;
      }
    },
    viewPolicy(workloadType: PolicyType): void {
      if (!this.selectedDepartment?.id) return;
      const policyId = policyUtil.createPolicyId(workloadType, ScopeType.Department, this.selectedDepartment.id);
      this.$router.push({
        name: POLICIES_ROUTE_NAMES.POLICY_VIEW,
        params: { id: policyId },
        query: { previousRoute: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX },
      });
    },
  },
  watch: {
    selectedDepartment: {
      handler(newVal: IDepartmentTable | null): void {
        if (!newVal) return;
        this.loadSelectedDepartmentPolicy();
      },
      immediate: true,
    },
  },
});
</script>

<style scoped lang="scss"></style>
