<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-form-wrapper
    ref="formWrapperRef"
    :form-state="{}"
    :init-dirty="unsavedGuard.isAnyStepDirty.value || isScopeNameDirty || isNodePoolOrderDirty"
    full-width
    class="department-edit-form-v3"
  >
    <unsaved-changes-dialog
      v-model="unsavedDialogModel"
      @save="handleSaveUnsavedChanges"
      @discard="unsavedGuard.confirmAction('discard')"
    />

    <org-unit-creation-warning-dialog
      v-model="showWarningDialog"
      :storage-key="DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW"
      title="Create Department & Close?"
      message="No dedicated resources or scheduling guarantees have been set.<br> Workloads can only run when spare GPUs are available. To dedicate <br> resources, continue editing."
      continue-editing-label="Continue Editing"
      create-and-close-label="Create Department & Close"
      @create-and-close="proceedWithSaveAndClose"
    />

    <runai-wizard-layout
      :steps="wizardSteps"
      :active-step="activeStep"
      width="1201px"
      left-width="200px"
      right-width="200px"
      margin-top="26px"
      collapse-left-on-small-screen
      hide-right-on-small-screen
      :collapse-left-breakpoint="WIDE_LAYOUT_MIN_WIDTH"
      :hide-right-breakpoint="WIDE_LAYOUT_MIN_WIDTH"
      @step-click="onStepClick"
    >
      <scope-name-step
        v-if="currentStep?.name === EStepNames.ScopeName"
        v-model:name="localDepartment.name"
        v-model:description="localDepartment.description"
        :department="localDepartment"
        :is-new-department="isNewDepartment"
        :cluster-id="localDepartment.clusterId"
        :read-only="readOnly"
        @update-scope="handleScopeUpdate"
      />

      <node-pools-priority-list
        v-else-if="isNodePoolStep"
        v-model="selectedNodePool"
        class="node-pool-step"
        :node-pools="stepNodePools"
        :disable-reorder="readOnly"
        :disable-reorder-tooltip="readOnly ? EOrgUnitFormTooltip.NodePoolsPriorityOrderReadOnly : undefined"
        @update:node-pools="handleNodePoolsReorder"
      >
        <runai-expansion-item :label="currentStep?.title" default-opened disable-closing>
          <div class="step-placeholder" :aid="`department-step-${currentStep?.name}-${selectedNodePool}`" />
        </runai-expansion-item>
      </node-pools-priority-list>

      <scheduling-rules-step
        v-else-if="currentStep?.name === EStepNames.SchedulingRules"
        :entity-type="EQuotaEntity.department"
        :cluster-id="localDepartment.clusterId || ''"
        :scheduling-rules="localDepartment.schedulingRules"
        :node-types="localDepartment.nodeTypes"
        :loading="isDepartmentSubmitting"
        :read-only="readOnly"
        :is-single-node-pool="isSingleNodePool"
        @save-scheduling-rules="handleSaveSchedulingRules"
        @close="handleClose"
      />

      <access-rules-step
        v-else-if="currentStep?.name === EStepNames.AccessRules"
        :form-options="accessRulesFormOptions"
        @access-rules-created="handleAccessRulesCreated"
        @access-rule-deleted="accessRulesCount = Math.max(0, accessRulesCount - 1)"
      />

      <template #content-footer>
        <org-unit-form-footer
          v-if="showFooter"
          entity-type="Department"
          :is-new-entity="isNewDepartment"
          :is-first-step="isFirstStep"
          :is-last-step="isLastStep"
          :submitting-action="submittingAction"
          :read-only="readOnly"
          @close="handleClose"
          @next="handleNext"
          @submit-and-close="handleSubmitAndClose"
          @submit-and-continue="handleSubmitAndContinue"
        />
      </template>

      <template #right-sidebar>
        <org-unit-wizard-summary
          :entity-type="EQuotaEntity.department"
          :resources="localDepartment.resources ?? []"
          :default-node-pools="localDepartment.defaultNodePools ?? []"
          :scheduling-rules="localDepartment.schedulingRules"
          :scope-id="(localDepartment as Department).id"
          :access-rules-count="accessRulesCount"
        />
      </template>
    </runai-wizard-layout>
  </runai-form-wrapper>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useQuasar } from "quasar";
import { usePermissionStore } from "@/stores/permissions.store";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiWizardLayout } from "@/components/common/runai-wizard-layout/";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { OrgUnitWizardSummary } from "@/components/org-unit/v2/wizard-summary";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";
import type { SubmittingAction } from "@/components/org-unit/v2/org-unit-form-footer/org-unit-form-footer.vue";
import { OrgUnitCreationWarningDialog } from "@/components/org-unit/v2/common/org-unit-creation-warning-dialog";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { NodePoolsPriorityList } from "@/components/org-unit/v2/common/node-pools-priority-list";
import { SchedulingRulesStep } from "@/components/org-unit/v2/common/scheduling-rules-step";
import { AccessRulesStep } from "@/components/org-unit/v2/common/access-rules-step";
import { ScopeNameStep } from "@/components/org-unit/v2/department/steps/scope-name-step";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { provideWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import type {
  Department,
  DepartmentCreationRequest,
  NodeTypesPerWorkload,
  SchedulingRules,
} from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import {
  AccessRulesSortFilterFields,
  ScopeType,
  type AccessRulesBatchCreateResponse,
} from "@/swagger-models/authorization-client";
import { EQuotaEntity } from "@/models/resource.model";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import {
  DEPARTMENT_FORM_V3_WIZARD_STEPS,
  EOrgUnitFormTooltip,
  ESaveOrgUnitAction,
  EStepNames,
  type INodePoolTotals,
  type ISaveOrgUnitPayload,
} from "@/models/org-unit-forms.model";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { deepCopy } from "@/utils/common.util";
import { storageUtil } from "@/utils/storage.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { ErrorAlert } from "@/utils/error-alert.util";
import { filterOrgUnitSteps } from "@/utils/wizard-steps.util/wizard-steps.util";
import { DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW } from "@/common/storage.constant";

export interface Props {
  department: Department | DepartmentCreationRequest;
  isNewDepartment: boolean;
  isDepartmentSubmitting?: boolean;
  nodePoolsMinimal?: NodepoolMinimal[];
  readOnly?: boolean;
  nodePoolTotals?: INodePoolTotals;
}

const props = withDefaults(defineProps<Props>(), {
  isDepartmentSubmitting: false,
  nodePoolsMinimal: () => [],
  readOnly: false,
  nodePoolTotals: () => ({}),
});

const emit = defineEmits<{
  (e: "on-cancel"): void;
  (e: "on-save", payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest>): void;
  (e: "update-scope", scope: { clusterId: string }): void;
}>();

const NODE_POOL_STEPS: string[] = [EStepNames.GpuQuota, EStepNames.CPUQuota, EStepNames.Projects];
// The 1201px form fits beside the full stepper and summary only from the design's 1680px frame;
// below it the design collapses the stepper and hides the summary instead of shrinking the form.
const WIDE_LAYOUT_MIN_WIDTH = 1680;

const route = useRoute();
const $q = useQuasar();
const permissionStore = usePermissionStore();

const localDepartment = ref<Department | DepartmentCreationRequest>(deepCopy(props.department));
const formWrapperRef = ref<InstanceType<typeof RunaiFormWrapper> | null>(null);

const unsavedGuard = provideUnsavedChangesGuard();
provideWizardPreview();

const initialScopeName = ref(props.department.name);
const initialScopeDescription = ref(props.department.description);
const isScopeNameDirty = computed(
  () =>
    localDepartment.value.name !== initialScopeName.value ||
    localDepartment.value.description !== initialScopeDescription.value,
);

const initialNodePoolOrder = ref<string[]>([...(props.department.defaultNodePools ?? [])]);
const isNodePoolOrderDirty = computed(
  () => (localDepartment.value.defaultNodePools ?? []).join() !== initialNodePoolOrder.value.join(),
);

const accessRulesCount = ref(0);
const submittingAction = ref<SubmittingAction>(null);
const showWarningDialog = ref(false);

const cpuQuotaPoolNames = computed(
  () => new Set(props.nodePoolsMinimal.filter((np) => np.schedulingCapabilities?.cpuQuota).map((np) => np.name)),
);
const hasCpuQuotaPool = computed(() => cpuQuotaPoolNames.value.size > 0);

const isSingleNodePool = computed(() => props.nodePoolsMinimal.length === 1);

const getInitialSteps = (): IRunaiWizardLayoutStep[] =>
  deepCopy(
    filterOrgUnitSteps(DEPARTMENT_FORM_V3_WIZARD_STEPS, {
      // GPU quota is shown in every scheduling mode in this form, not only when a node pool has quota on.
      nodepoolCapabilities: { cpuQuota: hasCpuQuotaPool.value, quota: true },
      hasPermission: (resource, action) => permissionStore.hasPermission(resource, action),
    }),
  );

const initialActiveStep = route.query.stepName?.toString();

const wizard = useWizardLayout(getInitialSteps(), initialActiveStep);
const { steps, currentStep, activeStep, isFirstStep, isLastStep, nextStep, handleStepClick, setSteps } = wizard;

const wizardSteps = computed<IRunaiWizardLayoutStep[]>(() =>
  steps.value.map((step) =>
    step.name === EStepNames.ScopeName || !step.disabled
      ? step
      : { ...step, tooltip: EOrgUnitFormTooltip.StepsDisabledUntilScopeComplete },
  ),
);

const isNodePoolStep = computed(() => NODE_POOL_STEPS.includes(currentStep.value?.name ?? ""));

const stepNodePools = computed<string[]>(() => {
  const ordered = localDepartment.value.defaultNodePools ?? [];
  if (currentStep.value?.name !== EStepNames.CPUQuota) return ordered;
  return ordered.filter((name) => cpuQuotaPoolNames.value.has(name));
});

const selectedNodePool = ref("");

watch(
  stepNodePools,
  (nodePools) => {
    if (!nodePools.includes(selectedNodePool.value)) {
      selectedNodePool.value = nodePools[0] ?? "";
    }
  },
  { immediate: true },
);

// Steps that render their own save buttons when there's a single node pool would otherwise show two footers.
const showFooter = computed(
  () => !(currentStep.value?.name === EStepNames.SchedulingRules && isSingleNodePool.value && !props.readOnly),
);

const accessRulesFormOptions = computed(() => ({
  scopeId: (localDepartment.value as Department).id,
  scopeType: ScopeType.Department,
  scopeName: localDepartment.value.name,
  page: EAccessRuleModalPage.Department,
}));

const unsavedDialogModel = computed({
  get: () => unsavedGuard.dialog.isOpen.value,
  set: (value: boolean) => {
    if (!value) unsavedGuard.cancelPending();
  },
});

const runGuardedAction = async (action: () => Promise<void> | void): Promise<boolean> => {
  const stepName = currentStep.value?.name;
  if (!stepName) {
    await action();
    return true;
  }
  return unsavedGuard.requestAction({ stepName, onProceed: action });
};

const validate = async (): Promise<boolean> => {
  return (await formWrapperRef.value?.validate()) ?? true;
};

const handleScopeUpdate = (scope: { clusterId: string }): void => {
  localDepartment.value.clusterId = scope.clusterId;
  emit("update-scope", scope);
};

// The CPU quota step lists only CPU quota pools, so the other pools keep their places in the full order.
const handleNodePoolsReorder = (reorderedStepPools: string[]): void => {
  const stepPools = new Set(reorderedStepPools);
  let next = 0;
  localDepartment.value.defaultNodePools = (localDepartment.value.defaultNodePools ?? []).map((name) =>
    stepPools.has(name) ? reorderedStepPools[next++] : name,
  );
};

const saveNodePoolOrder = (): void => {
  emit("on-save", { department: localDepartment.value, action: ESaveOrgUnitAction.Save });
  initialNodePoolOrder.value = [...(localDepartment.value.defaultNodePools ?? [])];
};

const discardNodePoolOrder = (): void => {
  localDepartment.value.defaultNodePools = [...initialNodePoolOrder.value];
};

// A new department saves its order on create, so only an existing one is guarded.
// A step component that later registers under one of these names replaces this guard,
// so it has to include the node pool order in its own dirty state.
NODE_POOL_STEPS.forEach((stepName) =>
  unsavedGuard.registerStepGuard(stepName, {
    isDirty: () => !props.isNewDepartment && isNodePoolOrderDirty.value,
    save: saveNodePoolOrder,
    discard: discardNodePoolOrder,
  }),
);

const handleAccessRulesCreated = (createdAccessRules: AccessRulesBatchCreateResponse): void => {
  accessRulesCount.value += createdAccessRules?.succeeded?.length ?? 0;
};

const completeStepAndContinue = (): void => {
  wizard.markStepComplete(activeStep.value);
  nextStep();
};

const handleSaveSchedulingRules = async (payload: {
  schedulingRules: SchedulingRules | null;
  nodeTypes?: NodeTypesPerWorkload;
  navigateAfterSave?: boolean;
}): Promise<void> => {
  if (!(await validate())) return;

  localDepartment.value.schedulingRules = payload.schedulingRules ? deepCopy(payload.schedulingRules) : null;
  localDepartment.value.nodeTypes = payload.nodeTypes ? deepCopy(payload.nodeTypes) : undefined;

  emit("on-save", {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: payload.navigateAfterSave ? completeStepAndContinue : undefined,
  });
};

const onStepClick = async (payload: { step: IRunaiWizardLayoutStep; name: string }): Promise<void> => {
  await runGuardedAction(async () => {
    if (!(await validate())) return;
    handleStepClick(payload);
  });
};

const handleSaveUnsavedChanges = async (): Promise<void> => {
  if (!(await validate())) {
    unsavedGuard.cancelPending();
    return;
  }
  await unsavedGuard.confirmAction("save");
};

const handleClose = async (): Promise<void> => {
  await runGuardedAction(() => emit("on-cancel"));
};

const handleNext = async (): Promise<void> => {
  await runGuardedAction(async () => {
    if (!(await validate())) return;
    completeStepAndContinue();
  });
};

const proceedWithSaveAndClose = (): void => {
  submittingAction.value = "close";
  emit("on-save", {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Close,
    onComplete: () => {
      submittingAction.value = null;
    },
  });
};

const handleSubmitAndClose = async (): Promise<void> => {
  await runGuardedAction(async () => {
    if (!(await validate())) return;

    if (props.isNewDepartment && storageUtil.get<boolean>(DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW) !== true) {
      showWarningDialog.value = true;
      return;
    }
    proceedWithSaveAndClose();
  });
};

const handleSubmitAndContinue = async (): Promise<void> => {
  await runGuardedAction(async () => {
    if (!(await validate())) return;

    submittingAction.value = "continue";
    emit("on-save", {
      department: localDepartment.value,
      action: ESaveOrgUnitAction.Continue,
      onComplete: () => {
        submittingAction.value = null;
        handleNext();
      },
    });
  });
};

const loadAccessRulesCount = async (departmentId: string): Promise<void> => {
  try {
    const filterBy = [filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeId, departmentId)];
    accessRulesCount.value = await accessRuleService.getAccessRulesCount(filterBy);
  } catch (error) {
    if (!props.readOnly) {
      $q.notify(new ErrorAlert({ generalMessage: "Failed to fetch access rules count" }).getNotification(error));
    }
    accessRulesCount.value = 0;
  }
};

watch(
  () => props.isNewDepartment,
  (isNew) => {
    if (isNew) return;
    steps.value.forEach((step) => wizard.setStepDisabled(step.name, false));
  },
  { immediate: true },
);

watch(
  () => props.department,
  (newDepartment) => {
    const updatedDepartment = deepCopy(newDepartment);
    if (props.isNewDepartment) {
      updatedDepartment.name = localDepartment.value.name;
      updatedDepartment.description = localDepartment.value.description;
    }
    // The backend may not return nodeTypes names, so the local copy is kept.
    localDepartment.value = { ...updatedDepartment, nodeTypes: localDepartment.value.nodeTypes };

    initialScopeName.value = localDepartment.value.name;
    initialScopeDescription.value = localDepartment.value.description;
    initialNodePoolOrder.value = [...(localDepartment.value.defaultNodePools ?? [])];
  },
);

watch(hasCpuQuotaPool, () => setSteps(getInitialSteps()));

onMounted(async () => {
  if (initialActiveStep) {
    wizard.markStepComplete(EStepNames.ScopeName);
  }

  const departmentId = (localDepartment.value as Department).id;
  if (departmentId) {
    await loadAccessRulesCount(departmentId);
  }
});
</script>

<style scoped lang="scss">
.department-edit-form-v3 {
  padding: 0 20px 26px;
}

.node-pool-step {
  background: var(--bg-primary-default);

  :deep(.runai-expansion-item.q-expansion-item--popup) {
    padding: 0;

    > .q-expansion-item__container {
      border: 0;
      box-shadow: none;
    }

    .runai-expansion-header {
      height: 47px;
      min-height: 47px;
      padding: 0 var(--space-m);
    }
  }
}

.step-placeholder {
  min-height: 480px;
}
</style>
