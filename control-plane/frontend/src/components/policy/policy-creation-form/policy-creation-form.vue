<template>
  <section class="policy-creation-form column items-center q-pt-md">
    <runai-default-loader v-if="initializing" />
    <runai-form-wrapper
      v-show="!initializing"
      class="policy-form"
      :form-state="leaveGuardState"
      ref="elPolicyCreationForm"
      :enter-callback="save"
    >
      <runai-expansion-wrapper>
        <runai-scope-section
          :key="scopeSectionKey"
          kind-of-asset="policy"
          :read-only="isEdit"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopesForForm"
          :selected-cluster-id="scopeModel.clusterId || ''"
        />

        <runai-expansion-item
          class="workload-type-section"
          label="Workload type"
          default-opened
          :section-invalid="!selectedTypeValue"
        >
          <template #subheader>
            <div class="row items-center">
              <span>{{ selectedTypeLabel }}</span>
            </div>
          </template>
          <div>
            <div class="q-pb-md">Set which workload type can use this policy</div>
            <div class="row">
              <runai-select
                aid="workload-type-select"
                class="workload-type-select select-italic-placeholder col-3"
                filterable
                :no-option-text="NO_RESULTS_FOUND_LABEL"
                :model-value="selectedTypeValue"
                :options="combinedWorkloadTypeOptions"
                :loading="loadingWorkloadTypes"
                :disable="isEdit"
                @update:model-value="updateWorkloadType"
                placeholder="Select a type"
                emit-value
                :rules="[isSelected]"
              />
            </div>
            <runai-radio-options
              v-if="showDistributedOptions"
              class="q-mt-md"
              aid="distributed-radio-options"
              title="Set whether the policy should be applied to standard or distributed workloads"
              :model-value="isDistributed"
              :options="distributedRadioOptions"
              tooltip-width="500px"
              @update:model-value="updateDistributed"
            />
          </div>
        </runai-expansion-item>

        <q-linear-progress animation-speed="100" indeterminate v-if="loadingPolicy" />
        <policy-yaml-section
          aid="policy-yaml-section"
          v-if="showYamlSection"
          :existing-policy="existingPolicy"
          :policy-meta-and-type="policyMetaAndType"
          :policy-type="policyType"
          @update-policy="updatePolicy"
        >
          <template v-if="showConfigModeToggle" #mode-toggle>
            <policy-config-mode-toggle
              :model-value="effectiveConfigMode"
              :is-any-workload="isAnyWorkload"
              @update:model-value="updateConfigMode"
            />
          </template>
        </policy-yaml-section>

        <policy-rules-section
          v-if="showGvkRuleBuilder"
          ref="rulesSection"
          :key="`rules-${selectedGvkKey}-${initialGvkPolicy?.id ?? 'new'}`"
          :gvk="gvkForSection"
          :model-value="gvkRules"
          :sync-badge-label="syncBadgeLabel"
          @update:model-value="updateGvkRules"
        >
          <template v-if="showConfigModeToggle" #mode-toggle>
            <policy-config-mode-toggle
              :model-value="effectiveConfigMode"
              :is-any-workload="isAnyWorkload"
              @update:model-value="updateConfigMode"
            />
          </template>
        </policy-rules-section>

        <gvk-policy-yaml-section
          v-if="showGvkYaml"
          :model-value="gvkRules"
          :sync-badge-label="syncBadgeLabel"
          @update:model-value="updateGvkRules"
        >
          <template v-if="showConfigModeToggle" #mode-toggle>
            <policy-config-mode-toggle
              :model-value="effectiveConfigMode"
              :is-any-workload="isAnyWorkload"
              @update:model-value="updateConfigMode"
            />
          </template>
        </gvk-policy-yaml-section>
      </runai-expansion-wrapper>

      <section class="row items-center q-mt-md">
        <q-field
          aid="policy-form-error"
          class="col-4 form-hint no-padding"
          :model-value="displayFormHint"
          :rules="[isFormIncomplete]"
        ></q-field>
        <div class="buttons q-ml-auto">
          <runai-btn aid="cancel-btn" class="q-mr-md" @click="onCancel" flat color="primary" label="cancel" />
          <runai-btn
            aid="submit-btn"
            color="primary"
            @click="save"
            :loading="submitting"
            label="save policy"
            type="submit"
          />
        </div>
      </section>
    </runai-form-wrapper>

    <apply-changes-modal v-if="isChangesModalOpen" @apply="applyChanges" @discard="discardChanges" />

    <runai-base-dialog
      v-if="isSwitchRuleModalOpen"
      :model-value="true"
      size="width-sm"
      @close="cancelConfigModeSwitch"
      :on-enter-callback="confirmSaveOpenRuleAndSwitch"
    >
      <template #header> Save rule? </template>
      <template #body>
        <div>You have a rule open with unsaved changes. Save it before switching to the YAML editor, or discard it.</div>
      </template>
      <template #footer>
        <q-btn label="Discard" aid="switch-discard-rule-btn" flat color="primary" @click="discardOpenRuleAndSwitch" />
        <q-btn label="Save rule" aid="switch-save-rule-btn" color="primary" @click="confirmSaveOpenRuleAndSwitch" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { PolicyYamlSection } from "../policy-yaml-section";
import { ApplyChangesModal } from "../apply-changes-modal";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import PolicyConfigModeToggle from "./policy-config-mode-toggle.vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { PolicyRulesSection, GvkPolicyYamlSection } from "@/components/policy/policy-rule-builder";
import { NO_RESULTS_FOUND_LABEL } from "@/components/policy/policy-rule-builder/policy-rule-builder.constant";
// models
import type { IScopeModel, ISelectOption } from "@/models/global.model";

// services
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
import {
  PolicyType,
  ScopeType,
  type CreateGvkPolicyReq,
  type DistributedPolicyDefaultsAndRulesV2,
  type GetGvkPolicyResp,
  type GVK,
  type PolicyMeta,
  type PolicyRule,
} from "@/swagger-models/policy-service-client";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { RunaiSelect } from "@/components/common/runai-select";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { MIN_CLUSTER_VERSION_FOR_ANY_WORKLOAD_POLICY } from "@/common/version.constant";
import type { IPolicyMetaAndType } from "@/views/policy/policy-edit.vue";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import type { TPolicyDefaultAndRules, TWorkloadPolicy } from "@/models/policy.model";
import type { IGvkWorkloadTypeOption } from "@/models/gvk-policy.model";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import { gvkPolicyUtil, GVK_NON_PROJECT_SCOPE_TOOLTIP } from "@/utils/gvk-policy.util";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useAppStore } from "@/stores/app.store";

type TConfigMode = "rule-builder" | "yaml";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiExpansionWrapper,
    RunaiExpansionItem,
    RunaiSelect,
    PolicyYamlSection,
    ApplyChangesModal,
    RunaiBaseDialog,
    RunaiScopeSection,
    RunaiRadioOptions,
    PolicyConfigModeToggle,
    RunaiBtn,
    PolicyRulesSection,
    GvkPolicyYamlSection,
    RunaiDefaultLoader,
  },
  emits: ["submit", "submit-any-workload", "cancel", "load-policy", "load-gvk-policy"],
  props: {
    submitting: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    initialPolicy: {
      type: Object as PropType<TWorkloadPolicy>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
    existingPolicyType: {
      type: [String, null] as PropType<PolicyType | null>,
      required: false,
    },
    loadingPolicy: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isEdit: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    initialGvkPolicy: {
      type: Object as PropType<GetGvkPolicyResp | null>,
      required: false,
      default: null,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      sectionValidation: {
        scope: false,
      } as Record<string, boolean>,
      selectedWorkloadType: null as PolicyType | null,
      policyForm: null as HTMLFormElement | null,
      policyModel: deepCopy(this.initialPolicy) as TWorkloadPolicy,
      isFormInvalid: false as boolean,
      displayFormHint: false as boolean,
      formHintMessage: "" as string,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      scopeSectionKey: 10 as number,
      workloadTypeOptions: [
        {
          label: "Workspace",
          value: PolicyType.Workspace,
          aid: "workload-type-option-Workspace",
        },
        {
          label: "Training",
          value: PolicyType.Training,
          aid: "workload-type-option-Training",
        },
        {
          label: "Inference",
          value: PolicyType.Inference,
          aid: "workload-type-option-Inference",
        },
      ] as ISelectOption[],
      isChangesModalOpen: false as boolean,
      isSwitchRuleModalOpen: false as boolean,
      applyFunc: null as CallableFunction | null,
      discardFunc: null as CallableFunction | null,
      isDistributed: false as boolean,
      gvkWorkloadTypeOptions: [] as IGvkWorkloadTypeOption[],
      loadingWorkloadTypes: false as boolean,
      initializing: false as boolean,
      selectedGvkOption: null as IGvkWorkloadTypeOption | null,
      gvkConfigMode: "rule-builder" as TConfigMode,
      gvkRules: [] as PolicyRule[],
      syncBadgeSource: null as TConfigMode | null,
      hasSavedChangeSinceSwitch: false as boolean,
      NO_RESULTS_FOUND_LABEL,
    };
  },
  async created() {
    this.initializing = !!this.initialGvkPolicy;
    this.initSelectedType();
    await this.loadGvkWorkloadTypes();
    this.seedGvkEditState();
    this.initializing = false;
  },
  mounted() {
    this.policyForm = this.$refs["elPolicyCreationForm"] as HTMLFormElement;
  },
  computed: {
    leaveGuardState(): object {
      // any-workload rules live outside policyModel; include them or the leave guard won't prompt after building rules.
      return { policyModel: this.policyModel, gvkRules: this.gvkRules };
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.policyModel.meta?.scope as ScopeType,
        projectId: this.policyModel.meta?.projectId,
        departmentId: this.policyModel.meta?.departmentId,
        clusterId: this.policyModel.meta?.clusterId,
      };
    },
    showYamlSection(): boolean {
      return !!this.policyModel.meta?.scope && !!this.policyType && !this.loadingPolicy;
    },
    existingPolicy(): TPolicyDefaultAndRules | null {
      if (!this.policyModel.policy) return null;

      const policy = this.policyModel.policy;

      const hasDefaults = !!Object.keys(policy.defaults || {}).length;
      const hasRules = !!Object.keys(policy.rules || {}).length;

      // in distributed policy imposedAssets is not an array but an object with master and worker properties and each of them is an array
      const hasImposedAssets = Array.isArray(policy.imposedAssets)
        ? !!policy.imposedAssets.length
        : !!(policy as DistributedPolicyDefaultsAndRulesV2).imposedAssets?.worker?.length ||
          !!(policy as DistributedPolicyDefaultsAndRulesV2).imposedAssets?.master?.length;

      const isActualPolicyExist: boolean = hasImposedAssets || hasDefaults || hasRules;

      if (!isActualPolicyExist) return null;

      return this.policyModel.policy;
    },
    policyMetaAndType(): IPolicyMetaAndType | null {
      if (!this.policyType || !this.policyModel.meta?.scope) return null;
      return {
        policyMeta: this.policyModel.meta,
        policyType: this.policyType,
      };
    },
    distributedRadioOptions(): ISelectOption[] {
      return [
        {
          value: false,
          label: "Standard",
          toolTip: "The policy will be applied to non-distributed (single process) workloads",
        },
        {
          value: true,
          label: "Distributed",
          toolTip: "The policy will be applied to distributed (multi-process) workloads",
        },
      ];
    },
    policyType(): PolicyType | null {
      if (this.selectedWorkloadType === PolicyType.Training && this.isDistributed) return PolicyType.Distributed;
      return this.selectedWorkloadType;
    },
    showDistributedOptions(): boolean {
      return this.selectedWorkloadType === PolicyType.Training;
    },
    isAnyWorkload(): boolean {
      return !!this.selectedGvkOption;
    },
    gvkForSection(): GVK {
      return this.selectedGvkOption?.gvk ?? { group: "", version: "", kind: "" };
    },
    selectedGvkKey(): string {
      return this.selectedGvkOption?.key ?? "";
    },
    combinedWorkloadTypeOptions(): (ISelectOption & { separator?: boolean })[] {
      const scope = this.policyModel.meta?.scope as ScopeType;
      const gvkDisabledByScope = gvkPolicyUtil.isGvkTypeDisabledForScope(scope);
      if (!gvkDisabledByScope && !this.isScopeClusterAnyWorkloadCapable) return this.workloadTypeOptions;

      const gvkOptions: (ISelectOption & { separator?: boolean })[] = this.gvkWorkloadTypeOptions.map(
        (option, index) => ({
          label: option.label,
          value: option.key,
          aid: `workload-type-option-${option.label}`,
          disable: gvkDisabledByScope,
          disabledTooltip: gvkDisabledByScope ? GVK_NON_PROJECT_SCOPE_TOOLTIP : undefined,
          separator: index === 0,
        }),
      );
      return [...this.workloadTypeOptions, ...gvkOptions];
    },
    isScopeClusterAnyWorkloadCapable(): boolean {
      const meta = this.policyModel.meta;
      if (meta?.scope !== ScopeType.Project || meta?.projectId == null) return true;
      const clusterUuid = orgTreeService.getParentClusterUuid(
        this.authStore.orgUnitList,
        String(meta.projectId),
        ScopeType.Project,
      );
      if (!clusterUuid) return true;
      return this.clusterStore.isClusterVersionSufficient(clusterUuid, MIN_CLUSTER_VERSION_FOR_ANY_WORKLOAD_POLICY);
    },
    selectedTypeValue(): string | null {
      if (this.selectedGvkOption) return this.selectedGvkOption.key;
      return this.selectedWorkloadType;
    },
    selectedTypeLabel(): string {
      if (this.selectedGvkOption) return this.selectedGvkOption.label;
      return this.selectedWorkloadType || "";
    },
    showConfigModeToggle(): boolean {
      return this.showConfigArea && (this.isAnyWorkload || !this.isEdit) && this.isScopeClusterAnyWorkloadCapable;
    },
    syncBadgeLabel(): string {
      if (!this.syncBadgeSource) return "";
      return this.syncBadgeSource === "yaml" ? "Synced with YAML editor" : "Synced with Rule builder";
    },
    effectiveConfigMode(): TConfigMode {
      return this.isAnyWorkload ? this.gvkConfigMode : "yaml";
    },
    allowedScopesForForm(): PermittedScopes {
      return this.allowedScopes;
    },
    showConfigArea(): boolean {
      return !!this.selectedTypeValue && !!this.policyModel.meta?.scope && !this.loadingPolicy;
    },
    showGvkRuleBuilder(): boolean {
      return this.showConfigArea && this.isAnyWorkload && this.effectiveConfigMode === "rule-builder";
    },
    showGvkYaml(): boolean {
      return this.showConfigArea && this.isAnyWorkload && this.effectiveConfigMode === "yaml";
    },
  },
  methods: {
    isSelected(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.SELECT_TYPE;
    },
    async updateWorkloadType(value: string): Promise<void> {
      const shouldAlertUser = (this.existingPolicy && this.showYamlSection) || this.gvkRules.length > 0;
      if (!shouldAlertUser) {
        this.applyWorkloadType(value);
        return;
      }
      const applyChanges = await this.alertUser();
      if (applyChanges) this.applyWorkloadType(value);
      this.isChangesModalOpen = false;
    },
    applyWorkloadType(value: string): void {
      this.gvkRules = [];
      this.resetSyncBadge();
      if (gvkPolicyUtil.isGvkKey(value)) {
        this.selectedGvkOption = this.gvkWorkloadTypeOptions.find((option) => option.key === value) || null;
        this.selectedWorkloadType = null;
        this.isDistributed = false;
        this.gvkConfigMode = "rule-builder";
        this.loadGvkPolicyForSelection();
        return;
      }
      this.selectedGvkOption = null;
      this.selectedWorkloadType = value as PolicyType;
      this.loadPolicy();
      this.loadGvkPolicyForSelection();
    },
    async loadGvkWorkloadTypes(): Promise<void> {
      this.loadingWorkloadTypes = true;
      try {
        this.gvkWorkloadTypeOptions = await gvkPolicyService.listMappedWorkloadTypes(
          this.initialGvkPolicy?.workloadType,
        );
      } catch (error: unknown) {
        this.gvkWorkloadTypeOptions = [];
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("any-workload policy types"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        if (this.initialGvkPolicy) useAppStore().setFallback(true);
      } finally {
        this.loadingWorkloadTypes = false;
      }
    },
    seedGvkEditState(): void {
      if (!this.initialGvkPolicy) return;
      const targetKey = gvkPolicyUtil.gvkKey(this.initialGvkPolicy.workloadType);
      const option = this.gvkWorkloadTypeOptions.find((opt) => opt.key === targetKey) ?? null;
      if (!option) return;
      this.selectedGvkOption = option;
      this.selectedWorkloadType = null;
      this.gvkConfigMode = "rule-builder";
      this.gvkRules = this.initialGvkPolicy.rules ?? [];
      this.policyModel.meta = {
        ...(this.policyModel.meta as PolicyMeta),
        scope: this.initialGvkPolicy.scope,
        projectId: this.initialGvkPolicy.projectId != null ? Number(this.initialGvkPolicy.projectId) : null,
      };
    },
    updateConfigMode(mode: string | number | boolean | null): void {
      if (!this.isAnyWorkload || typeof mode !== "string" || mode === this.gvkConfigMode) return;
      if (mode === "yaml" && this.gvkConfigMode === "rule-builder" && this.rulesSectionHasDirtyRule()) {
        this.isSwitchRuleModalOpen = true;
        return;
      }
      this.performConfigModeSwitch(mode as TConfigMode);
    },
    rulesSectionHasDirtyRule(): boolean {
      const section = this.$refs.rulesSection as { isRuleDirty?: () => boolean } | undefined;
      return !!section?.isRuleDirty?.();
    },
    performConfigModeSwitch(mode: TConfigMode): void {
      this.syncBadgeSource = this.hasSavedChangeSinceSwitch ? this.gvkConfigMode : null;
      this.hasSavedChangeSinceSwitch = false;
      this.gvkConfigMode = mode;
    },
    async confirmSaveOpenRuleAndSwitch(): Promise<void> {
      const section = this.$refs.rulesSection as { saveOpenRule?: () => Promise<boolean> } | undefined;
      const saved = section?.saveOpenRule ? await section.saveOpenRule() : true;
      this.isSwitchRuleModalOpen = false;
      if (saved) this.performConfigModeSwitch("yaml");
    },
    discardOpenRuleAndSwitch(): void {
      const section = this.$refs.rulesSection as { discardOpenRule?: () => void } | undefined;
      section?.discardOpenRule?.();
      this.isSwitchRuleModalOpen = false;
      this.performConfigModeSwitch("yaml");
    },
    cancelConfigModeSwitch(): void {
      this.isSwitchRuleModalOpen = false;
    },
    updateGvkRules(rules: PolicyRule[]): void {
      this.gvkRules = rules;
      this.hasSavedChangeSinceSwitch = true;
      this.syncBadgeSource = null;
    },
    resetSyncBadge(): void {
      this.syncBadgeSource = null;
      this.hasSavedChangeSinceSwitch = false;
    },
    buildGvkRequest(): CreateGvkPolicyReq | null {
      // any-workload policies are project-scoped only (phase 1)
      if (!this.selectedGvkOption || this.policyModel.meta?.scope !== ScopeType.Project) return null;
      const projectId = this.policyModel.meta?.projectId;
      if (projectId == null) return null;
      return {
        workloadType: this.selectedGvkOption.gvk,
        scope: ScopeType.Project,
        projectId: String(projectId),
        rules: this.gvkRules,
      };
    },
    async saveGvkPolicy(): Promise<void> {
      if (this.effectiveConfigMode === "rule-builder") {
        const rulesSection = this.$refs.rulesSection as
          | { rulesValidationValue: string; revealValidation: () => void }
          | undefined;
        if (rulesSection?.rulesValidationValue) {
          rulesSection.revealValidation();
          this.showHint();
          return;
        }
      }
      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }
      const req = this.buildGvkRequest();
      if (!req) {
        this.showHint();
        return;
      }
      this.$emit("submit-any-workload", req);
    },
    async updateScope(updatedScope: IScopeModel): Promise<void> {
      const resetGvk = gvkPolicyUtil.shouldResetGvkOnScopeChange({
        isAnyWorkload: this.isAnyWorkload,
        newScope: updatedScope.scope,
      });
      const shouldAlertUser =
        (this.existingPolicy && this.showYamlSection) || resetGvk || (this.isAnyWorkload && this.gvkRules.length > 0);
      if (!shouldAlertUser) {
        this.applyScopeChange(updatedScope);
        return;
      }
      const applyChanges = await this.alertUser();
      if (applyChanges) {
        this.applyScopeChange(updatedScope);
      } else {
        this.scopeSectionKey++;
      }
      this.isChangesModalOpen = false;
    },
    applyScopeChange(updatedScope: IScopeModel): void {
      const leavingProjectWithGvk = gvkPolicyUtil.shouldResetGvkOnScopeChange({
        isAnyWorkload: this.isAnyWorkload,
        newScope: updatedScope.scope,
      });
      this.policyModel.meta = scopeUtil.updateScope(this.policyModel.meta as PolicyMeta, updatedScope) as PolicyMeta;
      const gvkNoLongerAllowed = this.isAnyWorkload && (leavingProjectWithGvk || !this.isScopeClusterAnyWorkloadCapable);
      if (gvkNoLongerAllowed) {
        this.selectedGvkOption = null;
        this.selectedWorkloadType = null;
        this.gvkRules = [];
        this.gvkConfigMode = "rule-builder";
        this.resetSyncBadge();
      } else if (this.isAnyWorkload) {
        this.gvkRules = [];
        this.resetSyncBadge();
      }
      this.loadPolicy();
      this.loadGvkPolicyForSelection();
    },
    async alertUser(): Promise<boolean> {
      this.isChangesModalOpen = true;
      return new Promise((res) => {
        this.$nextTick(() => {
          this.discardFunc = () => res(false);
          this.applyFunc = () => res(true);
        });
      });
    },
    loadPolicy(): void {
      if (!this.policyMetaAndType) return;
      this.$emit("load-policy", this.policyMetaAndType);
    },
    loadGvkPolicyForSelection(): void {
      const projectId = this.policyModel.meta?.projectId;
      const canLookup =
        this.isAnyWorkload &&
        !!this.selectedGvkOption &&
        this.policyModel.meta?.scope === ScopeType.Project &&
        projectId != null;
      this.$emit(
        "load-gvk-policy",
        canLookup && this.selectedGvkOption
          ? { workloadType: this.selectedGvkOption.gvk, projectId: String(projectId) }
          : null,
      );
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      this.displayFormHint = false;
      if (this.isAnyWorkload) {
        await this.saveGvkPolicy();
        return;
      }
      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }
      this.generatePolicyName();
      this.$emit("submit", { policyModel: this.policyModel, policyType: this.policyType });
    },
    async validate(): Promise<boolean> {
      const isValid: boolean = await (this.policyForm as HTMLFormElement).validate();
      return isValid;
    },
    showHint(message = ""): void {
      this.formHintMessage = message;
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : this.formHintMessage || errorMessages.FORM_INCOMPLETE;
    },
    updatePolicy(updatedPolicyRulesAndDefaults: TPolicyDefaultAndRules): void {
      if (!updatedPolicyRulesAndDefaults) return;
      this.policyModel.policy = updatedPolicyRulesAndDefaults;
    },
    generatePolicyName(): void {
      if (!this.policyModel.meta) return;
      const scope = this.scopeModel.scope;
      let prefix: string;
      switch (scope) {
        case ScopeType.Cluster:
          prefix =
            this.authStore.orgUnitList.find((orgUnit: OrgUnit) => {
              return orgUnit.id === this.scopeModel.clusterId && orgUnit.type === this.scopeModel.scope;
            })?.name || "";
          break;
        case ScopeType.Department:
          prefix =
            this.authStore.orgUnitList.find((orgUnit: OrgUnit) => {
              return orgUnit.id === this.scopeModel.departmentId && orgUnit.type === this.scopeModel.scope;
            })?.name || "";
          break;
        case ScopeType.Project:
          prefix =
            this.authStore.orgUnitList.find((orgUnit: OrgUnit) => {
              return orgUnit.id === this.scopeModel.projectId?.toString() && orgUnit.type === this.scopeModel.scope;
            })?.name || "";
          break;
        default:
          prefix = this.authStore.tenant.name;
          break;
      }
      this.policyModel.meta.name = `${prefix}-${this.policyType}`;
    },
    discardChanges(): void {
      this.discardFunc && this.discardFunc();
      this.discardFunc = null;
    },
    applyChanges(): void {
      this.applyFunc && this.applyFunc();
      this.applyFunc = null;
    },
    initSelectedType(): void {
      if (this.existingPolicyType === PolicyType.Distributed) {
        this.selectedWorkloadType = PolicyType.Training;
        this.isDistributed = true;
        return;
      }
      this.selectedWorkloadType = this.existingPolicyType as PolicyType;
    },
    async updateDistributed(isDistributed: boolean): Promise<void> {
      const shouldAlertUser = this.existingPolicy && this.showYamlSection;
      if (!shouldAlertUser) {
        this.isDistributed = isDistributed;
        this.loadPolicy();
        return;
      }
      const applyChanges = await this.alertUser();
      if (applyChanges) {
        this.isDistributed = isDistributed;
        this.loadPolicy();
      }
      this.isChangesModalOpen = false;
    },
  },
  watch: {
    initialGvkPolicy: {
      handler(policy: GetGvkPolicyResp | null): void {
        if (this.isEdit) return;
        this.gvkRules = policy?.rules ?? [];
      },
    },
    initialPolicy: {
      handler(newInitalPolicy: TWorkloadPolicy): void {
        this.policyModel = deepCopy(newInitalPolicy);
        this.scopeSectionKey++;
      },
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>

<style lang="scss" scoped></style>
