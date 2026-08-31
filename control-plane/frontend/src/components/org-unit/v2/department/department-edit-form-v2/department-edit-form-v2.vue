<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-form-wrapper
    :form-state="{}"
    :init-dirty="unsavedGuard.isAnyStepDirty.value || isScopeNameDirty"
    ref="formWrapperRef"
    full-width
    class="q-pb-xl"
  >
    <!-- Unsaved Changes Dialog -->
    <unsaved-changes-dialog
      v-model="unsavedDialogModel"
      @save="handleSaveUnsavedChanges"
      @discard="handleDiscardUnsavedChanges"
    />

    <!-- Warning Dialog -->
    <org-unit-creation-warning-dialog
      v-model="showWarningDialog"
      :storage-key="DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW"
      title="Create Department & Close?"
      message="No dedicated resources or scheduling guarantees have been set.<br> Workloads can only run when spare GPUs are available. To dedicate <br> resources, continue editing."
      continue-editing-label="Continue Editing"
      create-and-close-label="Create Department & Close"
      @continue-editing="() => {}"
      @create-and-close="proceedWithSaveAndClose"
    />

    <!-- Explore New Experience Dialog (Guided Tour) -->
    <explore-new-experience-dialog
      v-model="showExploreDialog"
      entity-type="department"
      @skip="handleExploreDialogSkip"
      @start-tour="handleExploreDialogStartTour"
    />

    <runai-wizard-layout ref="wizardLayoutRef" :steps="wizardSteps" :active-step="activeStep" @step-click="onStepClick">
      <!-- Main Content -->
      <template #default>
        <!-- Step 1: Cluster & Name -->
        <scope-name-step
          v-if="currentStep?.name === EStepNames.ScopeName"
          :department="localDepartment"
          :is-new-department="isNewDepartment"
          :cluster-id="localDepartment.clusterId"
          :read-only="props.readOnly"
          v-model:name="localDepartment.name"
          v-model:description="localDepartment.description"
          @update-scope="handleScopeUpdate"
        />

        <!-- Step: GPU Quota -->
        <gpu-quota-step
          v-else-if="currentStep?.name === EStepNames.GpuQuota"
          ref="gpuQuotaStepRef"
          :entity-type="EQuotaEntity.department"
          :org-unit="localDepartment as Department"
          :cluster-id="localDepartment.clusterId || ''"
          :node-pool-totals="nodePoolTotals"
          :loading="props.isDepartmentSubmitting"
          :read-only="props.readOnly"
          @save-gpu-quota="handleSaveGpuQuota"
          @close="handleClose"
        />

        <!-- Step: GPU Over Quota -->
        <gpu-over-quota-step
          v-else-if="currentStep?.name === EStepNames.OverQuota"
          ref="gpuOverQuotaStepRef"
          :entity-type="EQuotaEntity.department"
          :org-unit="localDepartment as Department"
          :cluster-id="localDepartment.clusterId || ''"
          @save-over-quota="handleSaveGpuOverQuota"
          @close="handleClose"
        />

        <!-- Step: Scheduling Rules -->
        <scheduling-rules-step
          v-else-if="currentStep?.name === EStepNames.SchedulingRules"
          :entity-type="EQuotaEntity.department"
          :cluster-id="localDepartment.clusterId || ''"
          :scheduling-rules="localDepartment.schedulingRules"
          :node-types="localDepartment.nodeTypes"
          :loading="props.isDepartmentSubmitting"
          :read-only="props.readOnly"
          :is-single-node-pool="isSingleNodePool"
          @save-scheduling-rules="handleSaveSchedulingRules"
          @close="handleClose"
        />

        <!-- Step: Access Rules -->
        <access-rules-step
          v-else-if="currentStep?.name === EStepNames.AccessRules"
          :form-options="accessRulesFormOptions"
          @access-rules-created="handleAccessRulesCreated"
          @access-rule-deleted="handleAccessRuleDeleted"
        />

        <!-- Step: CPU Quota -->
        <cpu-quota-step
          v-else-if="currentStep?.name === EStepNames.CPUQuota"
          :entity-type="EQuotaEntity.department"
          :resources="(localDepartment.resources as Resources[]) || []"
          :default-node-pools="localDepartment.defaultNodePools ?? undefined"
          :node-pool-quota-statuses="(localDepartment as Department).status?.nodePoolQuotaStatuses"
          :subordinate-resources="(localDepartment as Department).projectsAggregatedResources"
          :node-pool-totals="props.nodePoolTotals"
          :loading="props.isDepartmentSubmitting"
          :read-only="props.readOnly"
          @save-cpu-quota="handleSaveCpuQuota"
          @close="handleClose"
        />
      </template>

      <template #content-footer>
        <!-- Footer with action buttons -->
        <org-unit-form-footer
          v-if="isFirstStep || isLastStep || !isSingleNodePool"
          entity-type="Department"
          :is-new-entity="isNewDepartment"
          :is-first-step="isFirstStep"
          :is-last-step="isLastStep"
          :submitting-action="submittingAction"
          :read-only="props.readOnly"
          @close="handleClose"
          @next="handleNext"
          @submit-and-close="handleSubmitAndClose"
          @submit-and-continue="handleSubmitAndContinue"
        />
      </template>

      <!-- Summary Panel -->
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
import { ref, computed, watch, onMounted } from "vue";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiWizardLayout } from "@/components/common/runai-wizard-layout/";
import { useWizardLayout } from "@/composables/use-wizard-layout.composable";
import { OrgUnitWizardSummary } from "@/components/org-unit/v2/wizard-summary";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";
import { OrgUnitCreationWarningDialog } from "@/components/org-unit/v2/common/org-unit-creation-warning-dialog";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { ExploreNewExperienceDialog } from "@/components/org-unit/v2/common/explore-new-experience-dialog";
import { ScopeNameStep } from "@/components/org-unit/v2/department/steps/scope-name-step";
import { GpuQuotaStep } from "@/components/org-unit/v2/common/gpu-quota-step";
import { GpuOverQuotaStep } from "@/components/org-unit/v2/common/gpu-over-quota-step";
import { SchedulingRulesStep } from "@/components/org-unit/v2/common/scheduling-rules-step";
import { AccessRulesStep } from "@/components/org-unit/v2/common/access-rules-step";
import { CpuQuotaStep } from "@/components/org-unit/v2/common/cpu-quota-step";
import { CPU_VALUE_FACTOR, DEFAULT_NODE_POOL_NAME, EQuotaEntity } from "@/models/resource.model";
import { storageUtil } from "@/utils/storage.util";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import { DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW } from "@/common/storage.constant";
import { STORAGE_KEYS } from "@/models/storage.model";
import { useGuidedTourStore } from "@/stores/guided-tour.store";

import type {
  Department,
  DepartmentCreationRequest,
  SchedulingRules,
  NodeTypesPerWorkload,
  Resources,
} from "@/swagger-models/org-unit-service-client";
import type { Units } from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { deepCopy } from "@/utils/common.util";
import {
  DEPARTMENT_FORM_WIZARD_STEPS,
  ESaveOrgUnitAction,
  EStepNames,
  type ISaveOrgUnitPayload,
  EOrgUnitFormTooltip,
  type INodePoolTotals,
  type IPeerOrgUnitQuotaChange,
} from "@/models/org-unit-forms.model";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { filterOrgUnitSteps } from "@/utils/wizard-steps.util/wizard-steps.util";
import { usePermissionStore } from "@/stores/permissions.store";
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { provideWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { useRoute } from "vue-router";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import {
  AccessRulesSortFilterFields,
  ScopeType,
  type AccessRulesBatchCreateResponse,
} from "@/swagger-models/authorization-client";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useQuasar } from "quasar";
import { orgUnitUtil } from "@/utils/org-unit.util/org-unit.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { useSettingStore } from "@/stores/setting.store";

const permissionStore = usePermissionStore();
const settingStore = useSettingStore();

// Props
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
  nodePoolTotals: () => ({}),
  readOnly: false,
});

// Computed: Nodepool capabilities for step filtering
const nodepoolCapabilities = computed(() => ({
  cpuQuota: props.nodePoolsMinimal?.some((np) => !!np.schedulingCapabilities?.cpuQuota),
  quota: props.nodePoolsMinimal?.some((np) => !!np.schedulingCapabilities?.quota),
}));

// Single node pool scenario - used for footer button behavior in steps
const isSingleNodePool = computed(() => (props.nodePoolsMinimal?.length ?? 0) === 1);

// Check if there are multiple node pools (not just the default one) for guided tour
const hasMultipleNodePools = computed(() => {
  const nodePools = props.nodePoolsMinimal ?? [];
  if (nodePools.length <= 1) return false;
  // If there's only one node pool and it's the default, return false
  if (nodePools.length === 1 && nodePools[0]?.name === DEFAULT_NODE_POOL_NAME) return false;
  return true;
});

// Build steps array - filters steps based on nodepool capabilities and permissions
const getInitialSteps = (): IRunaiWizardLayoutStep[] => {
  return deepCopy(
    filterOrgUnitSteps(DEPARTMENT_FORM_WIZARD_STEPS, {
      nodepoolCapabilities: nodepoolCapabilities.value,
      hasPermission: (r, a) => permissionStore.hasPermission(r, a),
    }),
  );
};

const getInitialActiveStep = (): string | undefined => {
  const route = useRoute();
  const stepName = route.query.stepName;
  return stepName?.toString();
};

// Emits
const emit = defineEmits<{
  (e: "on-cancel"): void;
  (e: "on-save", payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest>): void;
  (e: "update-scope", scope: { clusterId: string }): void;
}>();

// Create local copy of department to work with
const localDepartment = ref<Department | DepartmentCreationRequest>(deepCopy(props.department));

const $q = useQuasar();
const guidedTourStore = useGuidedTourStore();

// Refs
const wizardLayoutRef = ref<InstanceType<typeof RunaiWizardLayout> | null>(null);
const formWrapperRef = ref<InstanceType<typeof RunaiFormWrapper> | null>(null);
const gpuQuotaStepRef = ref<InstanceType<typeof GpuQuotaStep> | null>(null);
const gpuOverQuotaStepRef = ref<InstanceType<typeof GpuOverQuotaStep> | null>(null);

const unsavedGuard = provideUnsavedChangesGuard();
provideWizardPreview();

// Track scope-name dirty state separately (not as a step guard) so it triggers the
// route-leave modal (X button) without interfering with inter-step navigation guards.
// Baselines are refs so they reset when props.department updates after a successful save.
const initialScopeName = ref(props.department.name);
const initialScopeDescription = ref(props.department.description);
const isScopeNameDirty = computed(
  () =>
    localDepartment.value.name !== initialScopeName.value ||
    localDepartment.value.description !== initialScopeDescription.value,
);

// Access rules count for the summary component
const accessRulesCount = ref(0);

const unsavedDialogModel = computed({
  get: () => unsavedGuard.dialog.isOpen.value,
  set: (value: boolean) => {
    if (!value) {
      unsavedGuard.cancelPending();
    }
  },
});

const runGuardedAction = async (action: () => Promise<void> | void): Promise<boolean> => {
  const stepName = currentStep.value?.name;
  if (!stepName) {
    await action();
    return true;
  }
  return await unsavedGuard.requestAction({
    stepName,
    onProceed: action,
  });
};

// Track which action is being submitted ('close' | 'continue' | null)
const submittingAction = ref<"close" | "continue" | null>(null);

// Warning dialog state
const showWarningDialog = ref(false);

// Explore new experience dialog state (for guided tour)
const showExploreDialog = ref(false);
const isFromHelpMenu = ref(false);

const wizard = useWizardLayout(getInitialSteps(), getInitialActiveStep());

// Destructure wizard returns
const { steps, currentStep, activeStep, isFirstStep, isLastStep, nextStep, handleStepClick, setSteps, goToStep } =
  wizard;

// Computed
const accessRulesFormOptions = computed(() => ({
  scopeId: (localDepartment.value as Department).id,
  scopeType: ScopeType.Department,
  scopeName: localDepartment.value.name,
  page: EAccessRuleModalPage.Department,
}));

const wizardSteps = computed<IRunaiWizardLayoutStep[]>(() => {
  return steps.value.map((step) => {
    if (step.name === EStepNames.ScopeName || !step.disabled) return step;

    return {
      ...step,
      tooltip: EOrgUnitFormTooltip.StepsDisabledUntilScopeComplete,
    };
  });
});

// Methods
const validate = async (): Promise<boolean> => {
  // Special case: Skip guard for department GPU quota step when default node pool has -1 (unlimited)
  // This is the initial state that doesn't require saving
  const shouldSkipGuard = (): boolean => {
    if (currentStep.value?.name !== EStepNames.GpuQuota) return false;

    const defaultResource = localDepartment.value.resources?.find((r) => r.nodePool?.name === DEFAULT_NODE_POOL_NAME);
    return defaultResource?.gpu?.deserved === -1;
  };

  if (shouldSkipGuard()) {
    return true;
  }
  if (formWrapperRef.value) {
    const isValid = await formWrapperRef.value.validate();
    return isValid;
  }
  return true;
};

const handleScopeUpdate = (scope: { clusterId: string }) => {
  localDepartment.value.clusterId = scope.clusterId;
  emit("update-scope", scope);
};

const handleAccessRulesCreated = (createdAccessRules: AccessRulesBatchCreateResponse) => {
  const succeededCount = createdAccessRules?.succeeded?.length ?? 0;
  accessRulesCount.value += succeededCount;
};

const handleAccessRuleDeleted = () => {
  accessRulesCount.value = Math.max(0, accessRulesCount.value - 1);
};

const handleSaveSchedulingRules = async (payload: {
  schedulingRules: SchedulingRules | null;
  nodeTypes?: NodeTypesPerWorkload;
  navigateAfterSave?: boolean;
}) => {
  const isValid = await validate();
  if (!isValid) {
    return;
  }

  localDepartment.value.schedulingRules = payload.schedulingRules ? deepCopy(payload.schedulingRules) : null;
  localDepartment.value.nodeTypes = payload.nodeTypes ? deepCopy(payload.nodeTypes) : undefined;

  emit("on-save", {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: payload.navigateAfterSave
      ? () => {
          wizard.markStepComplete(activeStep.value);
          nextStep();
        }
      : undefined,
  });
};

const handleSaveCpuQuota = async (payload: {
  nodePoolName: string;
  cpuQuota: number | null;
  memoryQuota: number | null;
  memoryUnit: Units;
  navigateAfterSave?: boolean;
}) => {
  const isValid = await validate();
  if (!isValid) {
    return;
  }

  // Find and update the resource for the specified node pool
  localDepartment.value.resources?.forEach((resource) => {
    if (resource.nodePool?.name === payload.nodePoolName) {
      resource.cpu = {
        ...resource.cpu,
        deserved: payload.cpuQuota,
        limit: payload.cpuQuota !== null ? payload.cpuQuota * CPU_VALUE_FACTOR : null,
      };
      resource.memory = {
        ...resource.memory,
        deserved: payload.memoryQuota,
        units: payload.memoryUnit,
      };
    }
  });

  emit("on-save", {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: payload.navigateAfterSave
      ? () => {
          wizard.markStepComplete(activeStep.value);
          nextStep();
        }
      : undefined,
  });
};

const savePeerOrgUnitQuotaBatch = async (peerChanges: Array<IPeerOrgUnitQuotaChange>): Promise<void> => {
  // If batch patch fails, the error propagates up and prevents the main department update.
  const batchRequest = orgUnitUtil.buildNodePoolBatchPatchRequest(peerChanges, settingStore.isOverQuotaPriorityEnabled);
  await orgUnitService.batchPatchDepartmentNodePoolResources(batchRequest);
};

const handleSaveGpuQuota = async (payload: {
  nodePoolName: string;
  gpuQuota: number;
  defaultNodePools?: string[];
  peerDepartmentChanges?: Array<IPeerOrgUnitQuotaChange>;
  navigateAfterSave?: boolean;
  onComplete?: () => void;
}) => {
  const isValid = await validate();
  if (!isValid) {
    return;
  }

  // Save peer org unit changes first (if any)
  // If this fails, the error propagates and prevents main department update
  if (payload.peerDepartmentChanges && payload.peerDepartmentChanges.length > 0) {
    try {
      await savePeerOrgUnitQuotaBatch(payload.peerDepartmentChanges);
    } catch (error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: "Failed to update peer org units",
      });
      $q.notify(errorAlert.getNotification(error));
      throw error; // Re-throw to prevent main department update
    }
  }

  // Update main department quota in local state
  localDepartment.value.resources?.forEach((resource) => {
    if (resource.nodePool?.name === payload.nodePoolName) {
      const updatedResource = {
        ...resource,
        gpu: {
          ...resource.gpu,
          deserved: payload.gpuQuota,
        },
      };
      Object.assign(resource, updatedResource);
    }
  });

  // Update default node pools if provided
  if (payload.defaultNodePools) {
    localDepartment.value.defaultNodePools = payload.defaultNodePools;
  }

  emit("on-save", {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: () => {
      // Call original onComplete (e.g., refetch data)
      payload.onComplete?.();
      // Navigate to next step if requested (single node pool flow)
      if (payload.navigateAfterSave) {
        wizard.markStepComplete(activeStep.value);
        nextStep();
      }
    },
  });
};

const handleSaveGpuOverQuota = async (payload: {
  nodePoolName: string;
  limit: number | null;
  overQuotaWeight?: number;
  peerOrgUnitChanges?: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank?: string;
    originalRank?: string;
    newWeight?: number;
    originalWeight?: number | null;
  }>;
  currentDepartmentNewRank?: string;
  currentDepartmentNewWeight?: number;
  navigateAfterSave?: boolean;
  onComplete?: () => void;
}) => {
  const isValid = await validate();
  if (!isValid) return;

  // Save peer department changes (rank and/or weight) in batch
  // Filter out current department - it will be updated in main save
  if (payload.peerOrgUnitChanges && payload.peerOrgUnitChanges.length > 0) {
    const peerOnlyChanges = payload.peerOrgUnitChanges.filter(
      (change) => change.orgUnitId !== (localDepartment.value as Department).id,
    );

    if (peerOnlyChanges.length > 0) {
      try {
        await savePeerDepartmentBatch(peerOnlyChanges);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to update departments",
        });
        $q.notify(errorAlert.getNotification(error));
        return; // Don't proceed with current department update
      }
    }
  }

  // Update current department's limit, weight, and rank
  localDepartment.value.resources?.forEach((resource) => {
    if (resource.nodePool?.name === payload.nodePoolName) {
      const finalWeight =
        payload.currentDepartmentNewWeight !== undefined ? payload.currentDepartmentNewWeight : payload.overQuotaWeight;

      resource.gpu = {
        ...resource.gpu,
        limit: payload.limit,
        overQuotaWeight: finalWeight,
      };

      if (payload.currentDepartmentNewRank) {
        resource.rank = payload.currentDepartmentNewRank;
      }
    }
  });

  // Align CPU/Memory limits with GPU over-quota state
  if (localDepartment.value.resources) {
    localDepartment.value.resources = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(
      localDepartment.value.resources as Resources[],
    );
  }

  // Save and proceed
  const savePayload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest> = {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: () => {
      // Call original onComplete (e.g., refetch data)
      payload.onComplete?.();
      // Navigate to next step if requested (single node pool flow)
      if (payload.navigateAfterSave) {
        wizard.markStepComplete(activeStep.value);
        nextStep();
      }
    },
  };
  emit("on-save", savePayload);
};

const savePeerDepartmentBatch = async (
  changes: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank?: string;
    originalRank?: string;
    newWeight?: number;
    originalWeight?: number | null;
  }>,
): Promise<void> => {
  const batchRequest = orgUnitUtil.buildNodePoolOrgUnitChangesBatchPatchRequest(changes);
  await orgUnitService.batchPatchDepartmentNodePoolResources(batchRequest);
};

const onStepClick = async (payload: { step: IRunaiWizardLayoutStep; name: string }) => {
  await runGuardedAction(async () => {
    const isValid = await validate();
    if (!isValid) {
      return;
    }
    handleStepClick(payload);
  });
};

const handleSaveUnsavedChanges = async () => {
  const isValid = await validate();
  if (!isValid) {
    unsavedGuard.cancelPending();
    return;
  }
  await unsavedGuard.confirmAction("save");
};

const handleDiscardUnsavedChanges = () => {
  unsavedGuard.confirmAction("discard");
};

const handleClose = async () => {
  await runGuardedAction(() => {
    emit("on-cancel");
  });
};

const handleNext = async () => {
  await runGuardedAction(async () => {
    const isValid = await validate();
    if (!isValid) {
      return;
    }

    wizard.markStepComplete(activeStep.value);
    nextStep();
  });
};

/**
 * Check if warning dialog should be shown based on user preference
 */
const shouldShowWarningDialog = (): boolean => {
  const dontShow = storageUtil.get<boolean>(DEPARTMENT_CREATION_WARNING_DIALOG_DONT_SHOW);
  return dontShow !== true;
};

const handleSubmitAndClose = async () => {
  // Close the guided tour if active (user chose to close instead of continue)
  if (guidedTourStore.isDepartmentTourActive) {
    guidedTourStore.destroy();
  }

  await runGuardedAction(async () => {
    const isValid = await validate();
    if (!isValid) {
      return;
    }

    if (props.isNewDepartment && shouldShowWarningDialog()) {
      showWarningDialog.value = true;
      return;
    }

    proceedWithSaveAndClose();
  });
};

const proceedWithSaveAndClose = () => {
  submittingAction.value = "close";
  const payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest> = {
    department: localDepartment.value,
    action: ESaveOrgUnitAction.Close,
    onComplete: () => {
      submittingAction.value = null;
    },
  };
  emit("on-save", payload);
};

const handleSubmitAndContinue = async () => {
  await runGuardedAction(async () => {
    const isValid = await validate();
    if (!isValid) {
      return;
    }

    submittingAction.value = "continue";
    const payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest> = {
      department: localDepartment.value,
      action: ESaveOrgUnitAction.Continue,
      onComplete: () => {
        submittingAction.value = null;
        handleNext();
        // Mark department as created for tour's nextOn condition
        if (guidedTourStore.isDepartmentTourActive) {
          guidedTourStore.markDepartmentCreated();
        }
      },
    };
    emit("on-save", payload);
  });
};

/**
 * Handle skip button click in explore dialog
 * Close the dialog without starting the tour and mark as seen
 */
const handleExploreDialogSkip = () => {
  storageUtil.save(STORAGE_KEYS.DEPARTMENT_TOUR_SEEN, true);
  showExploreDialog.value = false;
};

const handleTourStepNavigation = (stepName: string): void => {
  goToStep(stepName);
};

const handleTourCompletion = (): void => {
  goToStep(EStepNames.OverQuota);
};

const handleTourStepPreparation = async (stepName: string): Promise<void> => {
  switch (stepName) {
    case EStepNames.GpuQuota:
      await gpuQuotaStepRef.value?.expandDepartmentSection();
      break;
    case EStepNames.OverQuota:
      await gpuOverQuotaStepRef.value?.enableOverQuotaToggle();
      break;
  }
};

const handleTourCreate = (): Promise<boolean> => {
  return new Promise((resolve) => {
    runGuardedAction(async () => {
      const isValid = await validate();
      if (!isValid) {
        resolve(false);
        return;
      }

      submittingAction.value = "continue";
      const payload: ISaveOrgUnitPayload<Department | DepartmentCreationRequest> = {
        department: localDepartment.value,
        action: ESaveOrgUnitAction.Continue,
        onComplete: () => {
          submittingAction.value = null;
          handleNext();
          resolve(true);
        },
      };
      emit("on-save", payload);
    });
  });
};

/**
 * Handle start tour button click in explore dialog
 * Start the guided tour from step 0 (scope-name) and mark as seen
 * When coming from Help menu, don't save to localStorage
 */
const handleExploreDialogStartTour = () => {
  if (!isFromHelpMenu.value) {
    storageUtil.save(STORAGE_KEYS.DEPARTMENT_TOUR_SEEN, true);
  }
  showExploreDialog.value = false;

  guidedTourStore.openDepartmentGuidedTour(
    {
      hasMultipleNodePools: hasMultipleNodePools.value,
      isCPUResourcesQuotaEnabled: settingStore.isCPUResourcesQuotaEnabled,
      isOverQuotaPriorityEnabled: settingStore.isOverQuotaPriorityEnabled,
    },
    undefined,
    handleTourStepNavigation,
    handleTourCompletion,
    handleTourStepPreparation,
    handleTourCreate,
  );
};

// Enable all steps if in edit mode
watch(
  () => props.isNewDepartment,
  (isNew) => {
    if (!isNew) {
      wizardSteps.value.forEach((step) => {
        wizard.setStepDisabled(step.name, false);
      });
    }
  },
  { immediate: true },
);

watch(
  () => props.department,
  (newDepartment) => {
    const updatedDepartment = deepCopy(newDepartment);

    // Preserve user inputs from local state
    if (props.isNewDepartment) {
      updatedDepartment.name = localDepartment.value.name;
      updatedDepartment.description = localDepartment.value.description;
    }
    // Preserve nodeTypes from local state as backend may not return names property correctly
    localDepartment.value = { ...updatedDepartment, nodeTypes: localDepartment.value.nodeTypes };

    // Reset scope-name baselines so isScopeNameDirty reflects only unsaved changes
    initialScopeName.value = localDepartment.value.name;
    initialScopeDescription.value = localDepartment.value.description;
  },
);

// Update wizard steps when nodepool capabilities change
watch(
  () => nodepoolCapabilities.value,
  () => {
    setSteps(getInitialSteps());
  },
);

// if a specific step is provided, mark the first step as completed
onMounted(async () => {
  if (getInitialActiveStep()) {
    wizard.markStepComplete(EStepNames.ScopeName);
  }

  // Check if tour should start from Help menu (show explore dialog, skip localStorage save)
  const route = useRoute();
  if (props.isNewDepartment && route.query.startTour === "true") {
    isFromHelpMenu.value = true;
    showExploreDialog.value = true;
    return;
  }

  // Show explore dialog for new departments on mount (if neither project nor department tour has been seen)
  if (props.isNewDepartment) {
    const seenProjectTour = storageUtil.get(STORAGE_KEYS.PROJECT_TOUR_SEEN);
    const seenDepartmentTour = storageUtil.get(STORAGE_KEYS.DEPARTMENT_TOUR_SEEN);
    if (!seenProjectTour && !seenDepartmentTour) {
      showExploreDialog.value = true;
    }
  }

  // Fetch access rules count for the summary
  const departmentId = (localDepartment.value as Department).id;
  if (departmentId) {
    try {
      const filterBy = [filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeId, departmentId)];
      accessRulesCount.value = await accessRuleService.getAccessRulesCount(filterBy);
    } catch (error) {
      // In read-only/view mode, handle failure silently
      if (!props.readOnly) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to fetch access rules count",
        });
        $q.notify(errorAlert.getNotification(error));
      }
      accessRulesCount.value = 0;
    }
  }
});
</script>
