<template>
  <org-unit-step-wrapper
    label="GPU quota"
    :loading="loading"
    :hide-save-button="isSingleNodePool"
    @save="handleSave"
    @save-and-continue="handleSaveAndContinue"
    @close="emit('close')"
  >
    <runai-section-title
      v-if="allNodePools.length > 1"
      class="q-mb-lg q-pt-md q-px-md"
      title="Add node pools and set the default order in which workloads will be scheduled to use them."
      tooltip-text="The workload submitter can override this default.The leftmost node pool is the first one workloads will attempt to schedule onto."
    />
    <node-pool-tabs
      :node-pools="selectedNodePools"
      :available-node-pools="allNodePools"
      :model-value="selectedNodePoolId"
      @update:model-value="handleTabChange"
      @update:node-pools="updateNodePools"
    >
      <div class="row q-col-gutter-md items-stretch q-pb-md" :class="{ 'q-pt-md': allNodePools.length > 1 }">
        <div class="col-auto">
          <gpu-quota-availability-card
            :title="availabilityTitle"
            :total="availabilityTotal"
            :available="availabilityAvailable"
            :tooltip="tooltipType"
            :is-current-entity-unlimited="isCurrentEntityUnlimited"
          />
        </div>
        <transition name="fade-slide">
          <div v-if="excessQuotaAmount > 0 || isCurrentEntityUnlimited || isDefaultDepartmentUnlimited" class="col-5">
            <quota-exceeded-warning-banner
              :excess-amount="excessQuotaAmount"
              :parent-entity-type="parentEntityLabel"
              :hide-excess-amount="isCurrentEntityUnlimited || isDefaultDepartmentUnlimited"
            />
          </div>
        </transition>
      </div>

      <div class="text-b-subtitle2 q-my-md">Set the {{ entityLabel }}'s GPU quota in the selected node pool.</div>

      <org-unit-legend class="q-mb-md" :items="legendItems" />

      <gpu-quota-card
        :org-unit-name="orgUnit.name"
        :gpu-quota="gpuQuota"
        :total-quota="availabilityTotal"
        :unassigned="availabilityAvailable ?? 0"
        :state="gpuQuotaState"
        :message="gpuQuotaMessage"
        :parent-name="entityType === EQuotaEntity.project ? parentOrgUnit?.name || 'Department' : undefined"
        :entity-type="entityType === EQuotaEntity.department ? EQuotaEntity.department : undefined"
        :node-pool-name="entityType === EQuotaEntity.department ? selectedNodePoolName : undefined"
        :other-entities-assigned-quota="otherEntitiesAssignedQuota"
        :is-unlimited="isUnlimited"
        :is-current-entity-unlimited="isCurrentEntityUnlimited"
        :is-other-entities-quota-unlimited="isDefaultDepartmentUnlimited && orgUnit.name !== DEFAULT_DEPARTMENT_NAME"
        @update:gpu-quota="handleGpuQuotaUpdate"
      />
      <runai-folding-sub-section
        ref="foldingSectionRef"
        :label="entityType === EQuotaEntity.department ? 'Node pool overview' : 'Department quota overview'"
        class="q-mt-lg"
        lazy-mount
      >
        <org-unit-quota-table
          ref="quotaTableRef"
          :entity-type="entityType"
          :fetch-context="fetchContext"
          :available-quota="availabilityAvailable"
          :current-entity-id="orgUnit.id"
          :current-entity-quota="gpuQuota"
          :total-quota="availabilityTotal"
          :pending-changes="pendingPeerChanges"
          :parent-name="entityType === EQuotaEntity.project ? parentOrgUnit?.name || 'Department' : undefined"
          :server-total-assigned="otherEntitiesServerQuota"
          :is-default-department-unlimited="isDefaultDepartmentUnlimited"
          @quota-changed="handleQuotaChanged"
        />
      </runai-folding-sub-section>
    </node-pool-tabs>
  </org-unit-step-wrapper>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount, type ComputedRef } from "vue";
import type { Department, Project } from "@/swagger-models/org-unit-service-client";
import { DepartmentNodePoolResourcesFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { NodePoolTabs } from "@/components/org-unit/v2/common/node-pool-tabs";
import { GpuQuotaAvailabilityCard } from "@/components/org-unit/v2/common/gpu-quota-availability-card";
import { GpuQuotaCard } from "@/components/org-unit/v2/common/gpu-quota-card";
import { OrgUnitLegend } from "@/components/org-unit/v2/common/org-unit-legend";
import { OrgUnitQuotaTable } from "@/components/org-unit/v2/common/org-unit-quota-table";
import { QuotaExceededWarningBanner } from "@/components/org-unit/v2/common/quota-exceeded-warning-banner";
import { OrgUnitStepWrapper } from "@/components/org-unit/v2/common/org-unit-step-wrapper";
import {
  EOrgUnitFormTooltip,
  EGpuQuotaLegendText,
  type ELegendItem,
  ORG_UNIT_STACKED_BAR_CHART_COLORS,
  EStepNames,
  type INodePoolTotals,
} from "@/models/org-unit-forms.model";
import type { IPeerOrgUnitQuotaChange } from "@/models/org-unit-forms.model";
import { useQuotaValidation } from "@/composables/org-unit/use-quota-validation.composable";
import { useUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { useOrgUnitNodePools } from "@/composables/org-unit/use-org-unit-node-pools.composable";
import { useWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { useSettingStore } from "@/stores/setting.store";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { EQuotaEntity, DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { DEFAULT_DEPARTMENT_NAME } from "@/models/department.model";

export interface Props {
  entityType: EQuotaEntity;
  orgUnit: Project | Department;
  parentOrgUnit?: Department | null;
  clusterId: string;
  nodePoolTotals?: INodePoolTotals;
  loading?: boolean;
  readOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  parentOrgUnit: null,
  loading: false,
  readOnly: false,
});

const emit = defineEmits<{
  (
    e: "save-gpu-quota",
    payload: {
      nodePoolName: string;
      gpuQuota: number;
      defaultNodePools?: string[];
      peerProjectChanges?: Array<IPeerOrgUnitQuotaChange>;
      peerDepartmentChanges?: Array<IPeerOrgUnitQuotaChange>;
      navigateAfterSave?: boolean;
      onComplete?: () => void;
    },
  ): void;
  (e: "close"): void;
}>();

const STEP_NAME = EStepNames.GpuQuota;
const unsavedGuard = useUnsavedChangesGuard();
const settingStore = useSettingStore();
const wizardPreview = useWizardPreview();

// Dynamic labels based on entity type
const entityLabel = computed(() => (props.entityType === EQuotaEntity.project ? "project" : "department"));
const parentEntityLabel = computed(() => (props.entityType === EQuotaEntity.project ? "department" : "node pool"));

// Dynamic tooltip
const tooltipType = computed(() =>
  props.entityType === EQuotaEntity.project
    ? EOrgUnitFormTooltip.GpuQuotaAvailabilityProject
    : EOrgUnitFormTooltip.GpuQuotaAvailabilityDepartment,
);

// Single node pool scenario - hide save button and use wizard footer instead
const isSingleNodePool = computed(() => allNodePools.value.length === 1);

// Pending peer changes (deferred bulk save)
const pendingPeerChanges = ref<
  Map<
    string,
    {
      nodePoolName: string;
      nodePoolId: string;
      orgUnitName: string; // Track entity name for detecting default department changes
      newValue: number | null;
      originalValue: number | null;
      originalLimit: number | null;
      originalOverQuotaWeight: number | null;
    }
  >
>(new Map());

// Reference to the quota table component for refetching data and accessing computed totals
const quotaTableRef = ref<{ refetch: () => Promise<void>; totalAssignedQuota: ComputedRef<number> } | null>(null);

// Reference to the folding section for guided tour
const foldingSectionRef = ref<{ isOpen: boolean; hasBeenOpened: boolean; open: () => Promise<void> } | null>(null);

// GPU quota value
const gpuQuota = ref<number>(0);

// All peer entities' total assigned quota in this node pool (fetched from API for departments only)
const allPeerEntitiesServerQuota = ref<number>(0);

// For "default" node pool only: separate tracking of default department quota
// Sum of all departments EXCEPT "default" department (always a valid number)
const sumExcludingDefault = ref<number>(0);
// The "default" department's server quota (may be -1 = unlimited, only on "default" node pool)
const defaultDepartmentServerQuota = ref<number | null>(0);

// Node pool state management (editable mode)
const {
  allNodePools,
  selectedNodePools,
  selectedNodePoolId,
  selectedNodePoolName,
  localDefaultNodePools,
  isNodePoolListDirty,
  updateNodePools,
  resetNodePoolsToBaseline,
} = useOrgUnitNodePools({
  resources: () => props.orgUnit.resources || [],
  defaultNodePools: () => props.orgUnit.defaultNodePools ?? undefined,
  parentDefaultNodePools: () =>
    props.entityType === EQuotaEntity.project ? (props.parentOrgUnit?.defaultNodePools ?? undefined) : undefined,
  editable: true,
});

// Get GPU quota snapshot for a node pool
const getResourceSnapshot = (nodePoolName: string) => {
  const resource = props.orgUnit.resources?.find((r) => (r.nodePool?.name || "") === nodePoolName);
  let gpuValue = resource?.gpu?.deserved ?? 0;

  // For departments: unlimited (-1) quota is only valid for the default department on the default node pool.
  // For other node pools in the default department, or non-default departments, treat -1 as 0 (fresh start).
  if (gpuValue === -1 && props.entityType === EQuotaEntity.department) {
    const isDefaultDepartment = props.orgUnit.name === DEFAULT_DEPARTMENT_NAME;
    const isDefaultNodePool = nodePoolName === DEFAULT_NODE_POOL_NAME;

    if (!isDefaultDepartment || !isDefaultNodePool) {
      gpuValue = 0;
    }
  }

  return {
    gpu: gpuValue,
  };
};

const syncFromSnapshot = (snapshot: { gpu: number }) => {
  gpuQuota.value = snapshot.gpu;
};

// Computed: Current pool's resource snapshot (used to detect when saved data changes)
const currentPoolResourceSnapshot = computed(() => {
  if (!selectedNodePoolName.value) return null;
  return getResourceSnapshot(selectedNodePoolName.value);
});

// Load GPU quota when selected node pool changes OR when its saved resource data changes
watch(
  [() => selectedNodePoolName.value, currentPoolResourceSnapshot],
  ([nodePoolName, snapshot]) => {
    if (nodePoolName && snapshot) {
      syncFromSnapshot(snapshot);
    } else {
      syncFromSnapshot({ gpu: 0 });
    }
  },
  { immediate: true },
);

// Handle node pool tab change
const handleTabChange = async (nodePoolId: string) => {
  const pool = selectedNodePools.value.find((np) => np.id === nodePoolId);
  if (!pool) return;

  const applySwitch = () => {
    selectedNodePoolId.value = nodePoolId;
    syncFromSnapshot(getResourceSnapshot(pool.name));
  };

  if (!isTabDirty.value) {
    applySwitch();
    return;
  }

  await unsavedGuard.requestAction({
    stepName: STEP_NAME,
    onProceed: applySwitch,
  });
};

// Availability Total: Entity-specific source
const availabilityTotal = computed(() => {
  if (props.entityType === "project") {
    // Get from parent department.resources[].gpu.deserved
    if (!props.parentOrgUnit?.resources) return null;
    const resource = props.parentOrgUnit.resources.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value);
    const deserved = resource?.gpu?.deserved;
    if (deserved === null || deserved === undefined || deserved === -1) {
      return null;
    }
    return deserved;
  } else {
    // Get from nodePoolTotals[nodePool].totalGpus
    if (!selectedNodePoolName.value || !props.nodePoolTotals) return null;
    const nodePoolTotal = props.nodePoolTotals[selectedNodePoolName.value];
    return nodePoolTotal?.totalGpus ?? 0; // 0 means no quota return for this node pool, probably empty node pool
  }
});

// Fetch total quota assigned to ALL departments (including current) in this node pool from the cluster
// Only for departments
// Always fetch split sums (excluding/only default department) to handle unlimited (-1) correctly
const fetchAllPeerEntitiesQuota = async () => {
  if (props.entityType !== EQuotaEntity.department) return;

  if (!selectedNodePoolId.value || !props.clusterId) {
    allPeerEntitiesServerQuota.value = 0;
    sumExcludingDefault.value = 0;
    defaultDepartmentServerQuota.value = 0;
    return;
  }

  try {
    const baseFilters = [
      filterUtil.getEqualsFilterString(DepartmentNodePoolResourcesFilterSortFields.ClusterId, props.clusterId),
    ];

    // Fetch two separate sums to handle unlimited (-1) correctly on any node pool
    // 1. Sum EXCLUDING default department (always a valid number)
    const excludeDefaultFilters = [
      ...baseFilters,
      filterUtil.getNotEqualsFilterString(
        DepartmentNodePoolResourcesFilterSortFields.OrgUnitName,
        DEFAULT_DEPARTMENT_NAME,
      ),
    ];
    const excludeDefaultResponse = await orgUnitService.sumDepartmentNodePoolResources(
      selectedNodePoolId.value,
      excludeDefaultFilters,
    );
    sumExcludingDefault.value = excludeDefaultResponse.deservedSum ?? 0;

    // 2. Sum for ONLY default department (may be -1 = unlimited)
    const defaultOnlyFilters = [
      ...baseFilters,
      filterUtil.getEqualsFilterString(DepartmentNodePoolResourcesFilterSortFields.OrgUnitName, DEFAULT_DEPARTMENT_NAME),
    ];
    const defaultOnlyResponse = await orgUnitService.sumDepartmentNodePoolResources(
      selectedNodePoolId.value,
      defaultOnlyFilters,
    );
    defaultDepartmentServerQuota.value = defaultOnlyResponse.deservedSum ?? null; // -1 or null = unlimited

    // Also set the combined total (treating -1 as 0 for consistency)
    const defaultQuotaValue = defaultDepartmentServerQuota.value === -1 ? 0 : (defaultDepartmentServerQuota.value ?? 0);
    allPeerEntitiesServerQuota.value = sumExcludingDefault.value + defaultQuotaValue;
  } catch (error) {
    console.error("Failed to fetch all departments quota:", error);
    allPeerEntitiesServerQuota.value = 0;
    sumExcludingDefault.value = 0;
    defaultDepartmentServerQuota.value = 0;
  }
};

// Fetch peer entities' quota when node pool changes (departments only)
watch(
  () => selectedNodePoolId.value,
  () => {
    if (props.entityType === EQuotaEntity.department) {
      fetchAllPeerEntitiesQuota();
    }
  },
  { immediate: true },
);

// Check if the "default" department has unlimited quota (-1)
// Unlimited (-1) is only valid for the default department on the default node pool.
// On non-default node pools, -1 should be treated as 0 (same logic as getResourceSnapshot).
// Considers: local value (if current entity is default), pending changes, then server value
const isDefaultDepartmentUnlimited = computed(() => {
  // Only check for departments
  if (props.entityType !== EQuotaEntity.department) {
    return false;
  }

  // Unlimited quota only applies on the default node pool
  if (selectedNodePoolName.value !== DEFAULT_NODE_POOL_NAME) {
    return false;
  }

  // Case 1: Current entity IS the "default" department - use local gpuQuota
  if (props.orgUnit.name === DEFAULT_DEPARTMENT_NAME) {
    return gpuQuota.value === -1;
  }

  // Case 2: Check pending peer changes for "default" department in current node pool
  for (const change of pendingPeerChanges.value.values()) {
    if (change.nodePoolName === selectedNodePoolName.value && change.orgUnitName === DEFAULT_DEPARTMENT_NAME) {
      // Found pending change for default department - use its new value
      return change.newValue === -1;
    }
  }

  // Case 3: Fallback to server value
  return defaultDepartmentServerQuota.value === -1 || defaultDepartmentServerQuota.value === null;
});

// Availability Assigned: Entity-specific source
const availabilityAssigned = computed(() => {
  if (props.entityType === "project") {
    // From parentOrgUnit.projectsAggregatedResources
    const projectsAggregated = props.parentOrgUnit?.projectsAggregatedResources ?? [];
    const aggregated = projectsAggregated.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value);
    const assignedQuota = aggregated?.gpu?.deserved ?? null;
    return assignedQuota === undefined ? null : assignedQuota;
  } else {
    // From API fetch
    return allPeerEntitiesServerQuota.value;
  }
});

// Calculate available quota
const availabilityAvailable = computed(() => {
  const total = availabilityTotal.value;
  const assigned = availabilityAssigned.value;
  const isDepartment = props.entityType === EQuotaEntity.department;

  // If unlimited (null, undefined), return null to indicate unlimited
  if (total === null || total === undefined) {
    return null;
  }

  // For departments: if default department is unlimited, return 0
  // This shows "0 unassigned" while waiting for default to be assigned
  if (isDefaultDepartmentUnlimited.value) {
    return 0;
  }

  if (assigned === null || assigned === undefined) {
    return isDepartment ? null : 0;
  }

  // Calculate delta from current entity (live value vs server value)
  const snapshot = getResourceSnapshot(selectedNodePoolName.value);
  // For departments: treat -1 (unlimited) as 0 for delta calculation
  const snapshotGpu = isDepartment && (snapshot.gpu === -1 || snapshot.gpu === null) ? 0 : (snapshot.gpu ?? 0);
  const currentGpu = isDepartment && gpuQuota.value === -1 ? 0 : (gpuQuota.value ?? 0);
  const currentEntityDelta = currentGpu - snapshotGpu;

  // Calculate delta from pending peer changes
  // Positive delta = more assigned = less available
  let pendingPeerDelta = 0;
  for (const change of pendingPeerChanges.value.values()) {
    if (change.nodePoolName === selectedNodePoolName.value) {
      // For departments: treat -1 as 0
      const original = isDepartment && change.originalValue === -1 ? 0 : (change.originalValue ?? 0);
      const newVal = isDepartment && change.newValue === -1 ? 0 : (change.newValue ?? 0);
      pendingPeerDelta += newVal - original;
    }
  }

  // Allow negative available quota (means over-quota-assignment)
  return total - assigned - currentEntityDelta - pendingPeerDelta;
});

// Server total for OTHER entities only (excluding current entity, no pending changes applied)
// This is passed to the table so it can add currentEntityQuota + pendingDelta itself
const otherEntitiesServerQuota = computed(() => {
  const snapshot = getResourceSnapshot(selectedNodePoolName.value);
  const currentEntityOriginalQuota = snapshot.gpu;
  const totalAssigned = availabilityAssigned.value ?? 0;
  const isDepartment = props.entityType === EQuotaEntity.department;

  // For departments: if current entity has unlimited quota (-1 or null), treat as 0
  const quotaToSubtract =
    isDepartment && (currentEntityOriginalQuota === -1 || currentEntityOriginalQuota === null)
      ? 0
      : (currentEntityOriginalQuota ?? 0);

  return Math.max(0, totalAssigned - quotaToSubtract);
});

// Calculate sum of other entities' assigned quotas (excluding current entity)
const otherEntitiesAssignedQuota = computed(() => {
  const isDepartment = props.entityType === EQuotaEntity.department;
  // Calculate delta from pending peer changes
  let pendingDelta = 0;
  for (const change of pendingPeerChanges.value.values()) {
    if (change.nodePoolName === selectedNodePoolName.value) {
      // For departments: treat -1 as 0
      const original = isDepartment && change.originalValue === -1 ? 0 : (change.originalValue ?? 0);
      const newVal = isDepartment && change.newValue === -1 ? 0 : (change.newValue ?? 0);
      pendingDelta += newVal - original;
    }
  }

  return otherEntitiesServerQuota.value + pendingDelta;
});

// Allocated non-preemptible GPU for selected node pool
const allocatedNonPreemptibleGpu = computed(() => {
  if (!props.orgUnit.status?.nodePoolQuotaStatuses) return 0;

  const status = props.orgUnit.status.nodePoolQuotaStatuses.find(
    (s) => (s.nodePoolName || "") === selectedNodePoolName.value,
  );

  return status?.allocatedNonPreemptible?.gpu ?? 0;
});

// Subordinate projects' GPU quota for selected node pool (departments only)
const subordinateGpuQuota = computed<number | null | undefined>(() => {
  if (props.entityType !== EQuotaEntity.department) return undefined;
  const dept = props.orgUnit as Department;

  // If no projectsAggregatedResources at all, return undefined (no data to validate)
  if (!dept.projectsAggregatedResources || dept.projectsAggregatedResources.length === 0) {
    return undefined;
  }

  const aggregated = dept.projectsAggregatedResources.find(
    (r) => (r.nodePool?.name || "") === selectedNodePoolName.value,
  );

  // If node pool not found, return undefined (no data for this node pool)
  if (!aggregated) return undefined;

  // Return the actual value (null means unlimited, number means limited)
  return aggregated?.gpu?.deserved ?? null;
});

// Check if total quota is unlimited (only true when node pool total is null)
// Note: This is separate from isDefaultDepartmentUnlimited which handles the -1 case
const isUnlimited = computed(() => {
  return availabilityTotal.value === null;
});

// Check if current entity's quota is unlimited (-1) - only for departments
const isCurrentEntityUnlimited = computed(() => {
  return (
    props.entityType === EQuotaEntity.department &&
    props.orgUnit.name === DEFAULT_DEPARTMENT_NAME &&
    gpuQuota.value === -1
  );
});

// Calculate max assignable quota (dynamically updates with peer changes)
const maxAssignableQuota = computed(() => {
  const total = availabilityTotal.value;
  const otherEntities = otherEntitiesAssignedQuota.value;

  // If total is unlimited, return null (allows any positive number)
  if (total === null || total === undefined) return null;
  if (otherEntities === null || otherEntities === undefined) return total;

  return Math.max(0, total - otherEntities);
});

// Shows how much the total assigned quota exceeds the total quota
const excessQuotaAmount = computed(() => {
  const total = availabilityTotal.value;
  const currentEntity = gpuQuota.value ?? 0;
  const otherEntities = otherEntitiesAssignedQuota.value;

  if (total === null || total === undefined) return 0;

  // Calculate total assigned (current entity + other entities)
  const totalAssigned = currentEntity + otherEntities;

  // Return excess amount if over limit, otherwise 0
  return Math.max(0, totalAssigned - total);
});

// Availability title
const availabilityTitle = computed(() => {
  if (props.entityType === "project") {
    const deptName = props.parentOrgUnit?.name || "";
    return `Unassigned / Department ${deptName} quota`;
  } else {
    return `Unassigned / Total for node pool ${selectedNodePoolName.value}`;
  }
});

// GPU quota validation
const { quotaState: gpuQuotaState, quotaMessage: gpuQuotaMessage } = useQuotaValidation({
  requireNumberGreaterOrEqualToZero: true,
  quotaValue: gpuQuota,
  parentLimit: maxAssignableQuota,
  allocatedNonPreemptible: allocatedNonPreemptibleGpu,
  isLimitQuotaOverSubscriptionEnabled: computed(() => settingStore.isLimitQuotaOverSubscriptionEnabled),
  resourceName: "GPU",
  resourceUnit: "GPUs",
  entityType: props.entityType === EQuotaEntity.department ? EQuotaEntity.department : undefined,
  subordinateQuota: props.entityType === EQuotaEntity.department ? subordinateGpuQuota : undefined,
  availabilityTotal,
  isDefaultDepartmentUnlimited,
});

// Legend items for quota distribution visualization
const legendItems = computed<ELegendItem[]>(() => {
  const sumOfOtherText =
    props.entityType === EQuotaEntity.project
      ? EGpuQuotaLegendText.SumOfOtherProjects
      : EGpuQuotaLegendText.SumOfOtherDepartments;

  return [
    { color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9, text: EGpuQuotaLegendText.Assigned },
    { color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue5, text: EGpuQuotaLegendText.Unassigned },
    { color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2, text: sumOfOtherText },
  ];
});

// Fetch context for the quota table (entity-specific structure)
const fetchContext = computed(() => {
  if (props.entityType === "project") {
    return {
      departmentId: props.parentOrgUnit?.id || "",
      currentProjectId: props.orgUnit.id,
      nodePoolName: selectedNodePoolName.value,
      clusterId: props.clusterId,
    };
  } else {
    return {
      currentDepartmentId: props.orgUnit.id,
      nodePoolName: selectedNodePoolName.value,
      clusterId: props.clusterId,
    };
  }
});

const handleGpuQuotaUpdate = (value: number) => {
  gpuQuota.value = value;
};

// Check if current tab (GPU quota value) has unsaved changes
const isTabDirty = computed(() => {
  if (!selectedNodePoolName.value) {
    return false;
  }

  // Check if there are pending peer changes for this node pool
  const hasPendingPeerChanges = Array.from(pendingPeerChanges.value.values()).some(
    (change) => change.nodePoolName === selectedNodePoolName.value,
  );

  if (hasPendingPeerChanges) {
    return true;
  }

  // Check if current entity has changes
  const snapshot = getResourceSnapshot(selectedNodePoolName.value);
  // Handle unlimited case (-1)
  if (snapshot.gpu === -1 && gpuQuota.value === null) {
    return false;
  }
  return gpuQuota.value !== snapshot.gpu;
});

// Check if step has any unsaved changes (quota OR node pool list)
const isStepDirty = computed(() => {
  return isTabDirty.value || isNodePoolListDirty.value;
});

// Update wizard preview when GPU quota or node pools change
watch(
  [gpuQuota, localDefaultNodePools, selectedNodePoolName],
  ([quota, nodePools, nodePoolName]) => {
    wizardPreview.updatePreview("gpuQuota", {
      nodePoolName: nodePoolName,
      gpuQuota: quota,
      defaultNodePools: nodePools,
    });
  },
  { immediate: true },
);

const handleSave = (navigateAfterSave = false): Promise<void> => {
  return new Promise((resolve) => {
    // Include node pool changes if dirty, otherwise undefined (parent will ignore)
    const defaultNodePools = isNodePoolListDirty.value ? localDefaultNodePools.value : undefined;

    // Build batch of peer org unit changes
    const peerChanges: Array<IPeerOrgUnitQuotaChange> = [];
    for (const [orgUnitId, change] of pendingPeerChanges.value.entries()) {
      peerChanges.push({
        orgUnitId,
        nodePoolName: change.nodePoolName,
        nodePoolId: change.nodePoolId,
        gpuQuota: change.newValue,
        originalLimit: change.originalLimit,
        originalOverQuotaWeight: change.originalOverQuotaWeight,
      });
    }

    // Build payload with entity-specific key
    const payload: {
      nodePoolName: string;
      gpuQuota: number;
      defaultNodePools?: string[];
      peerProjectChanges?: Array<IPeerOrgUnitQuotaChange>;
      peerDepartmentChanges?: Array<IPeerOrgUnitQuotaChange>;
      navigateAfterSave?: boolean;
      onComplete?: () => void;
    } = {
      nodePoolName: selectedNodePoolName.value,
      gpuQuota: gpuQuota.value,
      defaultNodePools,
      navigateAfterSave,
      // onComplete is called by parent after API save completes
      // We resolve the Promise immediately so the guard can proceed with navigation
      // Cleanup happens asynchronously (may not complete if component unmounts)
      onComplete: () => {
        resolve();
        // Async cleanup - if component unmounts, this may not complete (which is fine)
        if (props.entityType === EQuotaEntity.department) {
          Promise.all([quotaTableRef.value?.refetch() ?? Promise.resolve(), fetchAllPeerEntitiesQuota()]).then(() => {
            pendingPeerChanges.value.clear();
          });
        } else {
          (quotaTableRef.value?.refetch() ?? Promise.resolve()).then(() => {
            pendingPeerChanges.value.clear();
          });
        }
      },
    };

    // Entity-specific key
    if (props.entityType === "project") {
      payload.peerProjectChanges = peerChanges;
    } else {
      payload.peerDepartmentChanges = peerChanges;
    }

    emit("save-gpu-quota", payload);
  });
};

const handleSaveAndContinue = () => {
  handleSave(true);
};

const discardChanges = () => {
  // Clear all pending peer changes
  pendingPeerChanges.value.clear();

  // Reset GPU quota to saved value
  if (!selectedNodePoolName.value) {
    syncFromSnapshot({ gpu: 0 });
  } else {
    syncFromSnapshot(getResourceSnapshot(selectedNodePoolName.value));
  }

  // Reset node pool list to baseline (clean state)
  resetNodePoolsToBaseline();
};

const handleQuotaChanged = (payload: {
  entityId: string;
  entityName: string;
  nodePoolName: string;
  nodePoolId: string;
  gpuQuota: number | null;
  originalValue: number | null;
  originalLimit: number | null;
  originalOverQuotaWeight: number | null;
}) => {
  // Store change in pending map (no API call)
  pendingPeerChanges.value.set(payload.entityId, {
    nodePoolName: selectedNodePoolName.value,
    nodePoolId: payload.nodePoolId,
    orgUnitName: payload.entityName,
    newValue: payload.gpuQuota,
    originalValue: payload.originalValue,
    originalLimit: payload.originalLimit,
    originalOverQuotaWeight: payload.originalOverQuotaWeight,
  });
};

// Register unsaved changes guard - unregister on unmount
const unregisterGuard = unsavedGuard.registerStepGuard(STEP_NAME, {
  isDirty: () => isStepDirty.value,
  save: handleSave,
  discard: discardChanges,
});

onBeforeUnmount(() => {
  unregisterGuard();
  wizardPreview.clearPreview("gpuQuota");
});

const expandDepartmentSection = async (): Promise<void> => {
  if (foldingSectionRef.value && !foldingSectionRef.value.isOpen) {
    await foldingSectionRef.value.open();
  }
};

defineExpose({
  isDirty: isStepDirty,
  save: handleSave,
  discardChanges,
  expandDepartmentSection,
});
</script>

<style scoped lang="scss">
// Smooth fade and slide transition for warning banner
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease-out;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateX(-10px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateX(10px);
}
</style>
