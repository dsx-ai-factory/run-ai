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
      :storage-key="PROJECT_CREATION_WARNING_DIALOG_DONT_SHOW"
      title="Create Project & Close?"
      message="No dedicated resources or scheduling guarantees have been set.<br> Workloads can only run when spare GPUs are available. To dedicate <br> resources, continue editing."
      continue-editing-label="Continue Editing"
      create-and-close-label="Create Project & Close"
      @continue-editing="() => {}"
      @create-and-close="proceedWithSaveAndClose"
    />

    <!-- Explore New Experience Dialog (Guided Tour) -->
    <explore-new-experience-dialog
      v-model="showExploreDialog"
      @skip="handleExploreDialogSkip"
      @start-tour="handleExploreDialogStartTour"
    />

    <runai-wizard-layout ref="wizardLayoutRef" :steps="wizardSteps" :active-step="activeStep" @step-click="onStepClick">
      <!-- Main Content -->
      <template #default>
        <!-- Step 1: Project Name & Description -->
        <scope-name-step
          v-if="currentStep?.name === EStepNames.ScopeName"
          :project="localProject"
          :department="department"
          :is-new-project="isNewProject"
          :cluster-id="localProject.clusterId"
          :parent-id="localProject.parentId"
          v-model:name="localProject.name"
          v-model:description="localProject.description"
          v-model:namespace="localProject.requestedNamespace"
          @update-scope="handleScopeUpdate"
        />

        <!-- Step: CPU Quota -->
        <cpu-quota-step
          v-else-if="currentStep?.name === EStepNames.CPUQuota"
          ref="cpuQuotaStepRef"
          :entity-type="EQuotaEntity.project"
          :resources="(localProject.resources as Resources[]) || []"
          :default-node-pools="localProject.defaultNodePools ?? undefined"
          :parent-resources="department?.resources"
          :parent-default-node-pools="department?.defaultNodePools ?? undefined"
          :node-pool-quota-statuses="createdLocalProject.status?.nodePoolQuotaStatuses"
          @save-cpu-quota="handleSaveCpuQuota"
          @close="handleClose"
        />

        <gpu-quota-step
          v-else-if="currentStep?.name === EStepNames.GpuQuota"
          ref="gpuQuotaStepRef"
          :entity-type="EQuotaEntity.project"
          :org-unit="createdLocalProject"
          :parent-org-unit="department"
          :cluster-id="localProject.clusterId || ''"
          :loading="isProjectSubmitting"
          @save-gpu-quota="handleGpuQuotaSave"
          @close="handleClose"
        />

        <gpu-over-quota-step
          v-else-if="currentStep?.name === EStepNames.OverQuota"
          ref="gpuOverQuotaStepRef"
          :entity-type="EQuotaEntity.project"
          :org-unit="createdLocalProject"
          :parent-org-unit="department"
          :cluster-id="localProject.clusterId || ''"
          @save-over-quota="handleSaveGpuOverQuota"
          @close="handleClose"
        />

        <scheduling-rules-step
          v-else-if="currentStep?.name === EStepNames.SchedulingRules"
          :entity-type="EQuotaEntity.project"
          :cluster-id="localProject.clusterId || ''"
          :scheduling-rules="localProject.schedulingRules"
          :node-types="localProject.nodeTypes"
          :parent-scheduling-rules="department?.schedulingRules"
          :parent-node-types="department?.nodeTypes"
          :loading="isProjectSubmitting"
          :is-single-node-pool="isSingleNodePool"
          @save-scheduling-rules="handleSaveSchedulingRules"
          @close="handleClose"
        />

        <!-- Step 6: Access Rules -->
        <access-rules-step
          v-else-if="currentStep?.name === EStepNames.AccessRules"
          :form-options="accessRulesFormOptions"
          @access-rules-created="handleAccessRulesCreated"
          @access-rule-deleted="handleAccessRuleDeleted"
        />
      </template>

      <template #content-footer>
        <!-- Footer with action buttons -->
        <org-unit-form-footer
          v-if="isFirstStep || isLastStep || !isSingleNodePool"
          entity-type="Project"
          :is-new-entity="isNewProject"
          :is-first-step="isFirstStep"
          :is-last-step="isLastStep"
          :submitting-action="submittingAction"
          @close="handleClose"
          @next="handleNext"
          @submit-and-close="handleSubmitAndClose"
          @submit-and-continue="handleSubmitAndContinue"
        />
      </template>

      <!-- Summary Panel -->
      <template #right-sidebar>
        <org-unit-wizard-summary
          :entity-type="EQuotaEntity.project"
          :resources="localProject.resources ?? []"
          :default-node-pools="localProject.defaultNodePools ?? []"
          :scheduling-rules="summarySchedulingRules"
          :scope-id="(localProject as Project).id"
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
import { ScopeNameStep } from "@/components/org-unit/v2/project/steps/scope-name-step";
import { AccessRulesStep } from "@/components/org-unit/v2/common/access-rules-step";
import { CpuQuotaStep } from "@/components/org-unit/v2/common/cpu-quota-step";
import { SchedulingRulesStep } from "@/components/org-unit/v2/common/scheduling-rules-step";
import { GpuQuotaStep } from "@/components/org-unit/v2/common/gpu-quota-step";
import { EQuotaEntity } from "@/models/resource.model";
import { GpuOverQuotaStep } from "@/components/org-unit/v2/common/gpu-over-quota-step";
import { OrgUnitFormFooter } from "@/components/org-unit/v2/org-unit-form-footer";
import { OrgUnitCreationWarningDialog } from "@/components/org-unit/v2/common/org-unit-creation-warning-dialog";
import { UnsavedChangesDialog } from "@/components/org-unit/v2/common/unsaved-changes-dialog";
import { ExploreNewExperienceDialog } from "@/components/org-unit/v2/common/explore-new-experience-dialog";
import { storageUtil } from "@/utils/storage.util";
import { PROJECT_CREATION_WARNING_DIALOG_DONT_SHOW } from "@/common/storage.constant";
import { STORAGE_KEYS } from "@/models/storage.model";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";

import type { Units } from "@/swagger-models/org-unit-service-client";
import {
  type Department,
  type Project,
  type ProjectCreationRequest,
  type NodeTypesPerWorkload,
  type SchedulingRules,
  type Resources,
} from "@/swagger-models/org-unit-service-client";
import type { NodepoolMinimal } from "@/swagger-models/cluster-service-client";
import { deepCopy } from "@/utils/common.util";
import {
  ESaveOrgUnitAction,
  EStepNames,
  PROJECT_FORM_WIZARD_STEPS,
  type ISaveOrgUnitPayload,
  EOrgUnitFormTooltip,
  type IPeerOrgUnitQuotaChange,
} from "@/models/org-unit-forms.model";
import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { filterOrgUnitSteps } from "@/utils/wizard-steps.util";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import {
  AccessRulesSortFilterFields,
  ScopeType,
  type AccessRulesBatchCreateResponse,
} from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { CPU_VALUE_FACTOR } from "@/models/resource.model";
import { provideUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { provideWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { useRoute } from "vue-router";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { useQuasar } from "quasar";
import { useSettingStore } from "@/stores/setting.store";
import { orgUnitUtil } from "@/utils/org-unit.util/org-unit.util";

// Props
export interface Props {
  project: Project | ProjectCreationRequest;
  department?: Department | null;
  isNewProject: boolean;
  isProjectSubmitting?: boolean;
  nodePoolsMinimal?: NodepoolMinimal[];
}

const props = withDefaults(defineProps<Props>(), {
  department: null,
  isProjectSubmitting: false,
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

const permissionStore = usePermissionStore();

// Build steps array - filters steps based on nodepool capabilities and permissions
const getInitialSteps = (): IRunaiWizardLayoutStep[] => {
  return deepCopy(
    filterOrgUnitSteps(PROJECT_FORM_WIZARD_STEPS, {
      nodepoolCapabilities: nodepoolCapabilities.value,
      hasPermission: (r, a) => permissionStore.hasPermission(r, a),
    }),
  );
};

const getInitialActiveStep = (): string | undefined => {
  // if no stepName query return undefined
  const route = useRoute();
  const stepName = route.query.stepName;
  return stepName?.toString();
};

// Emits
const emit = defineEmits<{
  (e: "on-cancel"): void;
  (e: "update-scope", scope: { clusterId: string; departmentId?: string }): void;
  (e: "on-save", payload: ISaveOrgUnitPayload<Project | ProjectCreationRequest>): void;
}>();

// Create local copy of project to work with
const localProject = ref<Project | ProjectCreationRequest>(deepCopy(props.project));

const $q = useQuasar();
const settingStore = useSettingStore();
const guidedTourStore = useGuidedTourStore();

// Refs
const wizardLayoutRef = ref<InstanceType<typeof RunaiWizardLayout> | null>(null);
const formWrapperRef = ref<InstanceType<typeof RunaiFormWrapper> | null>(null);
const cpuQuotaStepRef = ref<InstanceType<typeof CpuQuotaStep> | null>(null);
const gpuQuotaStepRef = ref<InstanceType<typeof GpuQuotaStep> | null>(null);
const gpuOverQuotaStepRef = ref<InstanceType<typeof GpuOverQuotaStep> | null>(null);

const unsavedGuard = provideUnsavedChangesGuard();
provideWizardPreview();

// Track scope-name dirty state separately (not as a step guard) so it triggers the
// route-leave modal (X button) without interfering with inter-step navigation guards.
// Baselines are refs so they reset when props.project updates after a successful save.
const initialScopeName = ref(props.project.name);
const initialScopeDescription = ref(props.project.description);
const isScopeNameDirty = computed(
  () =>
    localProject.value.name !== initialScopeName.value ||
    localProject.value.description !== initialScopeDescription.value,
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
const wizardSteps = computed<IRunaiWizardLayoutStep[]>(() => {
  return steps.value.map((step) => {
    if (step.name === EStepNames.ScopeName || !step.disabled) return step;

    return {
      ...step,
      tooltip: EOrgUnitFormTooltip.StepsDisabledUntilScopeComplete,
    };
  });
});

const createdLocalProject = computed(() => localProject.value as Project);
const isOverQuotaPriorityEnabled = computed(() => settingStore.isOverQuotaPriorityEnabled);
const accessRulesFormOptions = computed(() => ({
  scopeId: (localProject.value as Project).id,
  scopeType: ScopeType.Project,
  scopeName: localProject.value.name,
  page: EAccessRuleModalPage.Project,
}));

const summarySchedulingRules = computed(() => {
  const schedulingRules: SchedulingRules = {
    interactiveJobTimeLimitSeconds:
      localProject.value.schedulingRules?.interactiveJobTimeLimitSeconds ??
      props.department?.schedulingRules?.interactiveJobTimeLimitSeconds ??
      null,
    interactiveJobMaxIdleDurationSeconds:
      localProject.value.schedulingRules?.interactiveJobMaxIdleDurationSeconds ??
      props.department?.schedulingRules?.interactiveJobMaxIdleDurationSeconds ??
      null,
    interactiveJobPreemptIdleDurationSeconds:
      localProject.value.schedulingRules?.interactiveJobPreemptIdleDurationSeconds ??
      props.department?.schedulingRules?.interactiveJobPreemptIdleDurationSeconds ??
      null,
    trainingJobMaxIdleDurationSeconds:
      localProject.value.schedulingRules?.trainingJobMaxIdleDurationSeconds ??
      props.department?.schedulingRules?.trainingJobMaxIdleDurationSeconds ??
      null,
    trainingJobTimeLimitSeconds:
      localProject.value.schedulingRules?.trainingJobTimeLimitSeconds ??
      props.department?.schedulingRules?.trainingJobTimeLimitSeconds ??
      null,
  };
  return schedulingRules;
});

// Methods

const handleScopeUpdate = (scope: { clusterId: string; departmentId?: string }) => {
  emit("update-scope", scope);
};

const savePeerOrgUnitQuotaBatch = async (peerChanges: Array<IPeerOrgUnitQuotaChange>): Promise<void> => {
  // If batch patch fails, the error propagates up and prevents the main project update.
  const batchRequest = orgUnitUtil.buildNodePoolBatchPatchRequest(peerChanges, isOverQuotaPriorityEnabled.value);

  await orgUnitService.batchPatchProjectNodePoolResources(batchRequest);
};

const handleAccessRulesCreated = (createdAccessRules: AccessRulesBatchCreateResponse) => {
  const succeededCount = createdAccessRules?.succeeded?.length ?? 0;
  accessRulesCount.value += succeededCount;
};

const handleAccessRuleDeleted = () => {
  accessRulesCount.value = Math.max(0, accessRulesCount.value - 1);
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

  localProject.value.resources.forEach((resource) => {
    if (resource.nodePool?.name === payload.nodePoolName) {
      const updatedResource = {
        ...resource,
        cpu: {
          ...resource.cpu,
          deserved: payload.cpuQuota,
          limit: payload.cpuQuota !== null ? payload.cpuQuota * CPU_VALUE_FACTOR : null,
        },
        memory: {
          ...resource.memory,
          deserved: payload.memoryQuota,
          units: payload.memoryUnit,
        },
      };
      Object.assign(resource, updatedResource);
    }
  });

  emit("on-save", {
    project: localProject.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: payload.navigateAfterSave
      ? () => {
          wizard.markStepComplete(activeStep.value);
          nextStep();
        }
      : undefined,
  });
};

const handleGpuQuotaSave = async (payload: {
  nodePoolName: string;
  gpuQuota: number;
  defaultNodePools?: string[];
  peerProjectChanges?: Array<IPeerOrgUnitQuotaChange>;
  navigateAfterSave?: boolean;
  onComplete?: () => void;
}) => {
  const isValid = await validate();
  if (!isValid) {
    return;
  }

  // Save peer org unit changes first (if any)
  // If this fails, the error propagates and prevents main project update
  if (payload.peerProjectChanges && payload.peerProjectChanges.length > 0) {
    try {
      await savePeerOrgUnitQuotaBatch(payload.peerProjectChanges);
    } catch (error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: "Failed to update peer org units",
      });
      $q.notify(errorAlert.getNotification(error));
      throw error; // Re-throw to prevent main project update
    }
  }

  // Update main project quota in local state
  localProject.value.resources.forEach((resource) => {
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

  // Update defaultNodePools if provided (node pool list was changed)
  if (payload.defaultNodePools) {
    localProject.value.defaultNodePools = payload.defaultNodePools;
  }

  emit("on-save", {
    project: localProject.value,
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

const savePeerOrgUnitBatch = async (
  changes: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank?: string;
    originalRank?: string;
    newWeight?: number;
    originalWeight?: number | null;
  }>,
): Promise<void> => {
  // If batch patch fails, the error propagates up and prevents the main project update.
  const batchRequest = orgUnitUtil.buildNodePoolOrgUnitChangesBatchPatchRequest(changes);
  await orgUnitService.batchPatchProjectNodePoolResources(batchRequest);
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
  currentProjectNewRank?: string; // New rank if current project was moved
  currentProjectNewWeight?: number; // New weight if current project's weight was changed
  navigateAfterSave?: boolean;
  onComplete?: () => void;
}) => {
  const isValid = await validate();
  if (!isValid) {
    return;
  }

  // Save peer org unit changes (rank and/or weight) in a single batch request
  // Filter out current project - it will be updated in the main project save
  // If this fails, the error propagates and prevents main project update
  if (payload.peerOrgUnitChanges && payload.peerOrgUnitChanges.length > 0) {
    const peerOnlyChanges = payload.peerOrgUnitChanges.filter(
      (change) => change.orgUnitId !== (localProject.value as Project).id,
    );

    if (peerOnlyChanges.length > 0) {
      try {
        await savePeerOrgUnitBatch(peerOnlyChanges);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to update projects",
        });
        $q.notify(errorAlert.getNotification(error));
        throw error; // Re-throw to prevent main project update
      }
    }
  }

  // Persist limit, overQuotaWeight, and rank (if changed) for the selected node pool
  localProject.value.resources?.forEach((resource) => {
    if (resource.nodePool?.name === payload.nodePoolName) {
      // Use currentProjectNewWeight if it was changed via peer list, otherwise use form value
      const finalWeight =
        payload.currentProjectNewWeight !== undefined ? payload.currentProjectNewWeight : payload.overQuotaWeight;

      resource.gpu = {
        ...resource.gpu,
        limit: payload.limit,
        overQuotaWeight: finalWeight,
      };

      // Update rank if current project was moved to a new rank
      if (payload.currentProjectNewRank) {
        resource.rank = payload.currentProjectNewRank;
      }
    }
  });

  // Align CPU/Memory limits with GPU over-quota state (reuse util)
  if (localProject.value.resources) {
    localProject.value.resources = orgUnitUtil.alignCpuAndMemoryLimitsByGpu(localProject.value.resources as Resources[]);
  }

  emit("on-save", {
    project: localProject.value,
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

const handleSaveSchedulingRules = async (payload: {
  schedulingRules: SchedulingRules | null;
  nodeTypes?: NodeTypesPerWorkload;
  navigateAfterSave?: boolean;
}) => {
  const isValid = await validate();

  if (!isValid) {
    return;
  }

  localProject.value.schedulingRules = payload.schedulingRules ? deepCopy(payload.schedulingRules) : null;
  localProject.value.nodeTypes = payload.nodeTypes ? deepCopy(payload.nodeTypes) : undefined;

  emit("on-save", {
    project: localProject.value,
    action: ESaveOrgUnitAction.Save,
    onComplete: payload.navigateAfterSave
      ? () => {
          wizard.markStepComplete(activeStep.value);
          nextStep();
        }
      : undefined,
  });
};

const validate = async (): Promise<boolean> => {
  if (formWrapperRef.value) {
    const isValid = await formWrapperRef.value.validate();
    return isValid;
  }
  return true;
};

const onStepClick = async (payload: { step: IRunaiWizardLayoutStep; name: string }) => {
  await runGuardedAction(async () => {
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
  const dontShow = storageUtil.get<boolean>(PROJECT_CREATION_WARNING_DIALOG_DONT_SHOW);
  return dontShow !== true;
};

const handleSubmitAndClose = async () => {
  // Close the guided tour if active (user chose to close instead of continue)
  if (guidedTourStore.isProjectTourActive) {
    guidedTourStore.destroy();
  }

  await runGuardedAction(async () => {
    const isValid = await validate();
    if (!isValid) {
      return;
    }

    if (props.isNewProject && shouldShowWarningDialog()) {
      showWarningDialog.value = true;
      return;
    }

    proceedWithSaveAndClose();
  });
};

const proceedWithSaveAndClose = () => {
  submittingAction.value = "close";
  const payload: ISaveOrgUnitPayload<Project | ProjectCreationRequest> = {
    project: localProject.value,
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
    const payload: ISaveOrgUnitPayload<Project | ProjectCreationRequest> = {
      project: localProject.value,
      action: ESaveOrgUnitAction.Continue,
      onComplete: () => {
        submittingAction.value = null;
        handleNext();
        // Mark project as created for tour's nextOn condition
        if (guidedTourStore.isProjectTourActive) {
          guidedTourStore.markProjectCreated();
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
  storageUtil.save(STORAGE_KEYS.PROJECT_TOUR_SEEN, true);
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
      const payload: ISaveOrgUnitPayload<Project | ProjectCreationRequest> = {
        project: localProject.value,
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
    storageUtil.save(STORAGE_KEYS.PROJECT_TOUR_SEEN, true);
  }
  showExploreDialog.value = false;

  guidedTourStore.openProjectGuidedTour(
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
  () => props.isNewProject,
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
  () => props.project,
  (newProject) => {
    const updatedProject = deepCopy(newProject);

    // Preserve user inputs from local state
    if (props.isNewProject) {
      updatedProject.name = localProject.value.name;
      updatedProject.description = localProject.value.description;
      updatedProject.requestedNamespace = localProject.value.requestedNamespace;
    }
    // Barak:todo - remove the nodeTypes merge when backend is fixed
    localProject.value = { ...updatedProject, nodeTypes: localProject.value.nodeTypes };

    // Reset scope-name baselines so isScopeNameDirty reflects only unsaved changes
    initialScopeName.value = localProject.value.name;
    initialScopeDescription.value = localProject.value.description;
  },
);

// if a specific step is provided, mark the first step as completed
onMounted(async () => {
  if (getInitialActiveStep()) {
    wizard.markStepComplete(EStepNames.ScopeName);
  }

  // Check if tour should start from Help menu (show explore dialog, skip localStorage save)
  const route = useRoute();
  if (props.isNewProject && route.query.startTour === "true") {
    isFromHelpMenu.value = true;
    showExploreDialog.value = true;
    return;
  }

  // Show explore dialog for new projects on mount (if neither project nor department tour has been seen)
  if (props.isNewProject) {
    const seenProjectTour = storageUtil.get(STORAGE_KEYS.PROJECT_TOUR_SEEN);
    const seenDepartmentTour = storageUtil.get(STORAGE_KEYS.DEPARTMENT_TOUR_SEEN);
    if (!seenProjectTour && !seenDepartmentTour) {
      showExploreDialog.value = true;
    }
  }

  // Fetch access rules count for the summary
  const projectId = (localProject.value as Project).id;
  if (projectId) {
    try {
      const filterBy = [filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeId, projectId)];
      accessRulesCount.value = await accessRuleService.getAccessRulesCount(filterBy);
    } catch (error) {
      const errorAlert = new ErrorAlert({
        generalMessage: "Failed to fetch access rules count",
      });
      $q.notify(errorAlert.getNotification(error));
      accessRulesCount.value = 0;
    }
  }
});

// Update wizard steps when nodepool capabilities change
watch(
  () => nodepoolCapabilities.value,
  () => {
    setSteps(getInitialSteps());
  },
);
</script>
