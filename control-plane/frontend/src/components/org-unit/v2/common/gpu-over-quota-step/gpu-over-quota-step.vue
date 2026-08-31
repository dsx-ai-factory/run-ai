<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <org-unit-step-wrapper
    label="GPU over quota"
    :hide-save-button="isSingleNodePool"
    @save="handleSaveClick"
    @save-and-continue="handleSaveAndContinue"
    @close="emit('close')"
  >
    <node-pool-tabs
      :node-pools="selectedNodePools"
      :available-node-pools="allNodePools"
      :model-value="selectedNodePoolId"
      disable-list-actions
      @update:model-value="handleTabChange"
    >
      <org-unit-over-quota-settings
        v-model:limit="limit"
        v-model:is-over-quota-enabled="isOverQuotaEnabled"
        :deserved="deserved ?? 0"
        :entity-type="entityType"
        :is-single-node-pool="isSingleNodePool"
        @update:is-over-quota-enabled="handleToggleChange"
        @update:limit="handleLimitChange"
      >
        <div class="text-b-subtitle2 q-mb-md q-mt-lg" aid="rank-description-text">
          Rank {{ entityLabel }}s in the order in which their workloads will be scheduled, then assign weights to
          determine their share of resources.
        </div>

        <org-unit-legend class="q-mb-md" :items="rankLegendItems" />

        <peer-org-unit-rank-list
          aid="peer-org-unit-rank-list"
          v-if="selectedNodePoolId && peerListParentId"
          ref="peerRankListRef"
          :entity-type="entityType"
          :parent-id="peerListParentId"
          :node-pool-id="selectedNodePoolId"
          :cluster-id="clusterId"
          :current-org-unit-id="currentOrgUnitId"
          :current-org-unit-rank="currentOrgUnitRank"
          @current-entity-weight-changed="handleCurrentEntityWeightChanged"
          @pending-changes-update="handlePendingChangesUpdate"
        />
      </org-unit-over-quota-settings>
    </node-pool-tabs>
  </org-unit-step-wrapper>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount, nextTick } from "vue";
import type { Department, Project, Resources } from "@/swagger-models/org-unit-service-client";
import { NodePoolTabs } from "@/components/org-unit/v2/common/node-pool-tabs";
import { useOrgUnitNodePools } from "@/composables/org-unit/use-org-unit-node-pools.composable";
import { OrgUnitStepWrapper } from "@/components/org-unit/v2/common/org-unit-step-wrapper";
import { useUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { OrgUnitOverQuotaSettings } from "@/components/org-unit/v2/common/org-unit-over-quota-settings";
import { OrgUnitLegend } from "@/components/org-unit/v2/common/org-unit-legend";
import { PeerOrgUnitRankList } from "@/components/org-unit/v2/common/peer-org-unit-rank-list";
import { EStepNames, ORG_UNIT_STACKED_BAR_CHART_COLORS } from "@/models/org-unit-forms.model";
import { EQuotaEntity } from "@/models/resource.model";
import { useWizardPreview } from "@/composables/org-unit/use-wizard-preview.composable";
import { useSettingStore } from "@/stores/setting.store";

export interface Props {
  entityType: EQuotaEntity;
  orgUnit: Project | Department; // The current entity being edited
  parentOrgUnit?: Department | null; // Parent department (only for projects)
  clusterId: string;
}

const props = withDefaults(defineProps<Props>(), {
  parentOrgUnit: null,
});

const emit = defineEmits<{
  (
    e: "save-over-quota",
    payload: {
      nodePoolName: string;
      limit: number | null;
      // Omitted when over-quota priority feature is off — backend requires "deserved or empty"
      overQuotaWeight?: number;
      peerOrgUnitChanges?: Array<{
        orgUnitId: string;
        nodePoolId: string;
        newRank?: string;
        originalRank?: string;
        newWeight?: number;
        originalWeight?: number | null;
      }>;
      currentProjectNewRank?: string; // For projects
      currentProjectNewWeight?: number; // For projects
      currentDepartmentNewRank?: string; // For departments
      currentDepartmentNewWeight?: number; // For departments
      navigateAfterSave?: boolean;
      onComplete?: () => void;
    },
  ): void;
  (e: "close"): void;
}>();

const wizardPreview = useWizardPreview();
const settingStore = useSettingStore();
const isOverQuotaPriorityEnabled = computed(() => settingStore.isOverQuotaPriorityEnabled);

// Computed properties to handle entity-specific logic
const currentOrgUnitId = computed(() => props.orgUnit.id);
const resources = computed(() => props.orgUnit.resources || []);
const defaultNodePools = computed(() => props.orgUnit.defaultNodePools ?? undefined);
const parentDefaultNodePools = computed(() =>
  props.entityType === "project" && props.parentOrgUnit
    ? (props.parentOrgUnit.defaultNodePools ?? undefined)
    : undefined,
);
const entityLabel = computed(() => (props.entityType === EQuotaEntity.project ? "project" : "department"));

// Single node pool scenario - hide save button and use wizard footer instead
const isSingleNodePool = computed(() => allNodePools.value.length === 1);

const {
  allNodePools,
  selectedNodePools,
  selectedNodePoolId,
  isNodePoolListDirty,
  localDefaultNodePools,
  selectedNodePoolName,
} = useOrgUnitNodePools({
  resources: () => resources.value,
  defaultNodePools: () => defaultNodePools.value,
  parentDefaultNodePools: () => parentDefaultNodePools.value,
});

const unsavedGuard = useUnsavedChangesGuard();
const STEP_NAME = EStepNames.OverQuota;

const limit = ref<number | null>(null);
const overQuotaWeight = ref<number>(0);
const deserved = ref<number | null>(null);
const isOverQuotaEnabled = ref(false);

// Reference to peer rank list for getting pending changes
const peerRankListRef = ref<{
  getPendingOrgUnitChanges: () => Map<
    string,
    {
      originalRank?: string;
      newRank?: string;
      originalWeight?: number | null;
      newWeight?: number;
    }
  >;
  hasPendingChanges: () => boolean;
  clearPendingChanges: () => void;
  refetch: () => Promise<void>;
} | null>(null);

// Get current org unit's rank for auto-expanding in peer list
const currentOrgUnitRank = computed(() => {
  const id = selectedNodePoolId.value;
  if (!id) return undefined;
  const resource = findResourceByNodePool(resources.value, id);
  return resource?.rank ?? undefined;
});

// Parent ID for peer list - department ID for projects, cluster ID for departments
const peerListParentId = computed(() => {
  if (props.entityType === "project") {
    return props.parentOrgUnit?.id;
  }
  return props.clusterId;
});

const normalizeLimit = (value: number | null | undefined): number | null => {
  if (value === null || value === undefined || value === -1) return null;
  return value;
};

const computeIsOverQuotaEnabled = (limitValue: number | null, deservedValue: number | null): boolean => {
  if (limitValue === null) return true;
  if (deservedValue === null) return true;
  return limitValue > deservedValue;
};

const findResourceByNodePool = (
  resources: Resources[],
  nodePoolId: string | null | undefined,
): Resources | undefined => {
  if (!nodePoolId) return undefined;
  return resources.find((resource) => resource.nodePool?.id === nodePoolId);
};

const loadSnapshot = (id: string | null | undefined) => {
  if (!id) {
    limit.value = null;
    deserved.value = null;
    isOverQuotaEnabled.value = false;
    return;
  }
  const resource = findResourceByNodePool(resources.value, id);
  const limitValue = normalizeLimit(resource?.gpu?.limit);
  const deservedValue = resource?.gpu?.deserved ?? null;
  limit.value = limitValue;
  overQuotaWeight.value = resource?.gpu?.overQuotaWeight ?? 0;
  deserved.value = deservedValue;
  isOverQuotaEnabled.value = computeIsOverQuotaEnabled(limitValue, deservedValue);
};

watch(
  () => selectedNodePoolId.value,
  (id) => {
    loadSnapshot(id);
  },
  { immediate: true },
);

// Update wizard preview when GPU quota or node pools change
watch(
  [isOverQuotaEnabled, limit, localDefaultNodePools, selectedNodePoolName],
  ([overQuotaEnabled, limitValue, nodePools, nodePoolName]) => {
    wizardPreview.updatePreview("gpuOverQuota", {
      nodePoolName: nodePoolName,
      overQuotaEnabled: overQuotaEnabled,
      limit: limitValue,
      defaultNodePools: nodePools,
    });
  },
  { immediate: true },
);

const handleTabChange = (nodePoolId: string) => {
  const applySwitch = () => {
    selectedNodePoolId.value = nodePoolId;
    loadSnapshot(nodePoolId);
    // Reset pending changes tracker for new node pool
    hasPeerPendingChanges.value = false;
  };

  if (!isDirty.value) {
    applySwitch();
    return;
  }

  unsavedGuard.requestAction({
    stepName: STEP_NAME,
    onProceed: applySwitch,
  });
};

const handleToggleChange = (value: boolean) => {
  isOverQuotaEnabled.value = value;
  if (value) {
    limit.value = null;
  } else {
    limit.value = deserved.value ?? 0;
    overQuotaWeight.value = 0;
  }
};

const handleLimitChange = (value: number | null) => {
  limit.value = value;
};

const isLimitDirty = computed(() => {
  const id = selectedNodePoolId.value;
  if (!id) return false;
  const resource = findResourceByNodePool(resources.value, id);
  const currentLimit = normalizeLimit(resource?.gpu?.limit);
  const currentWeight = resource?.gpu?.overQuotaWeight ?? 0;
  return limit.value !== currentLimit || overQuotaWeight.value !== currentWeight;
});

// Track pending changes from peer rank list (updated via event)
const hasPeerPendingChanges = ref(false);

const isDirty = computed(() => {
  if (isNodePoolListDirty.value) return true;
  if (isLimitDirty.value) return true;
  if (hasPeerPendingChanges.value) return true;
  return false;
});

const rankLegendItems = computed(() => [
  { color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue9, text: "Weight" },
  { color: ORG_UNIT_STACKED_BAR_CHART_COLORS.blue2, text: `Total weight of other ${entityLabel.value}s` },
]);

const save = (navigateAfterSave = false): Promise<void> => {
  return new Promise((resolve) => {
    const id = selectedNodePoolId.value;
    if (!id) {
      resolve();
      return;
    }
    const pool = selectedNodePools.value.find((np) => np.id === id);
    const nodePoolName = pool?.name || findResourceByNodePool(resources.value, id)?.nodePool?.name || "";

    // Gather pending changes from peer list (combined rank and weight changes)
    const peerOrgUnitChanges: Array<{
      orgUnitId: string;
      nodePoolId: string;
      newRank?: string;
      originalRank?: string;
      newWeight?: number;
      originalWeight?: number | null;
    }> = [];

    let currentNewRank: string | undefined = undefined;
    let currentNewWeight: number | undefined = undefined;

    if (peerRankListRef.value) {
      const pendingChanges = peerRankListRef.value.getPendingOrgUnitChanges();
      for (const [orgUnitId, change] of pendingChanges.entries()) {
        peerOrgUnitChanges.push({
          orgUnitId,
          nodePoolId: id,
          newRank: change.newRank,
          originalRank: change.originalRank,
          newWeight: change.newWeight,
          originalWeight: change.originalWeight,
        });

        // Check if current org unit has changes
        if (orgUnitId === currentOrgUnitId.value) {
          if (change.newRank) {
            currentNewRank = change.newRank;
          }
          if (change.newWeight !== undefined) {
            currentNewWeight = change.newWeight;
          }
        }
      }
    }

    // Build payload with entity-specific property names
    const payload: {
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
      currentProjectNewRank?: string;
      currentProjectNewWeight?: number;
      currentDepartmentNewRank?: string;
      currentDepartmentNewWeight?: number;
      navigateAfterSave?: boolean;
      onComplete?: () => void;
    } = {
      nodePoolName,
      limit: limit.value,
      // Omit when feature is off — backend rejects 0 unless it equals deserved
      overQuotaWeight: isOverQuotaPriorityEnabled.value ? overQuotaWeight.value : undefined,
      peerOrgUnitChanges: peerOrgUnitChanges.length > 0 ? peerOrgUnitChanges : undefined,
      navigateAfterSave,
      // onComplete is called by parent after API save completes
      // We resolve the Promise so the guard can proceed with navigation
      // Cleanup happens asynchronously (may not complete if component unmounts)
      onComplete: () => {
        resolve();
        // Only refetch if we're still on the same node pool
        // If node pool changed (e.g., tab switch after save), the child already re-initialized via watch
        if (selectedNodePoolId.value !== id) {
          return;
        }
        // Refetch peer rank list data, then clear pending changes
        peerRankListRef.value?.refetch().then(() => {
          peerRankListRef.value?.clearPendingChanges();
        });
      },
    };

    // Add entity-specific properties
    if (props.entityType === "project") {
      payload.currentProjectNewRank = currentNewRank;
      payload.currentProjectNewWeight = currentNewWeight;
    } else {
      payload.currentDepartmentNewRank = currentNewRank;
      payload.currentDepartmentNewWeight = currentNewWeight;
    }

    emit("save-over-quota", payload);
  });
};

const handleCurrentEntityWeightChanged = (payload: { weight: number }) => {
  overQuotaWeight.value = payload.weight;
};

const handlePendingChangesUpdate = (hasPendingChanges: boolean) => {
  hasPeerPendingChanges.value = hasPendingChanges;
};

const handleSaveClick = async () => {
  await save();
};

const handleSaveAndContinue = async () => {
  await save(true);
};

const discardChanges = () => {
  loadSnapshot(selectedNodePoolId.value);
  // Clear pending rank changes
  peerRankListRef.value?.clearPendingChanges();
  hasPeerPendingChanges.value = false;
};

const unregisterGuard = unsavedGuard.registerStepGuard(STEP_NAME, {
  isDirty: () => isDirty.value,
  save,
  discard: discardChanges,
});

onBeforeUnmount(() => {
  unregisterGuard();
  wizardPreview.clearPreview("gpuOverQuota");
});

const enableOverQuotaToggle = async (): Promise<void> => {
  if (!isOverQuotaEnabled.value) {
    isOverQuotaEnabled.value = true;
    await nextTick();
  }
};

defineExpose({
  isDirty,
  save,
  discardChanges,
  enableOverQuotaToggle,
});
</script>

<style scoped></style>
