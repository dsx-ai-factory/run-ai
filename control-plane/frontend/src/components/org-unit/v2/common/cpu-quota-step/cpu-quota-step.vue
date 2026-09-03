<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <org-unit-step-wrapper
    label="CPU quota"
    :read-only="readOnly"
    :hide-save-button="isSingleNodePool"
    @save="handleSave"
    @save-and-continue="handleSaveAndContinue"
    @close="emit('close')"
  >
    <node-pool-tabs
      :node-pools="selectedNodePools"
      :available-node-pools="allNodePools"
      :model-value="selectedNodePoolId"
      disable-list-actions
      @update:model-value="handleNodePoolTabChange"
    >
      <div class="q-mt-sm q-mb-lg text-b-main text-body2">
        Set the CPU quota for the {{ entityType }} in the selected node pool
      </div>
      <div class="cpu-input-box">
        <cpu-quota-input v-model="cpuQuota" :state="cpuQuotaState" :message="cpuQuotaMessage" :disable="readOnly" />
      </div>

      <div class="cpu-memory-input-box">
        <cpu-memory-quota-input
          v-model="memoryQuota"
          :unit="memoryUnit"
          :state="memoryQuotaState"
          :message="memoryQuotaMessage"
          :disable="readOnly"
          @update:unit="memoryUnit = $event"
        />
      </div>
    </node-pool-tabs>
  </org-unit-step-wrapper>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from "vue";
import { NodePoolTabs } from "@/components/org-unit/v2/common/node-pool-tabs";
import { CpuQuotaInput } from "@/components/org-unit/v2/common/cpu-quota-input";
import { CpuMemoryQuotaInput } from "@/components/org-unit/v2/common/cpu-memory-quota-input";
import { OrgUnitStepWrapper } from "@/components/org-unit/v2/common/org-unit-step-wrapper";
import type { Resources, QuotaStatusNodePool, AggregatedResources } from "@/swagger-models/org-unit-service-client";
import { Units } from "@/swagger-models/org-unit-service-client";
import { EQuotaEntity, CPU_VALUE_FACTOR } from "@/models/resource.model";
import { EStepNames, type INodePoolTotals } from "@/models/org-unit-forms.model";
import { useSettingStore } from "@/stores/setting.store";
import { useQuotaValidation } from "@/composables/org-unit/use-quota-validation.composable";
import { useUnsavedChangesGuard } from "@/composables/org-unit/use-unsaved-changes-guard.composable";
import { useOrgUnitNodePools } from "@/composables/org-unit/use-org-unit-node-pools.composable";
import { orgUnitUtil } from "@/utils/org-unit.util";

const MEMORY_ALLOCATED_NON_PREEMPTIBLE_UNITS = Units.Mb;

export interface Props {
  entityType: EQuotaEntity;
  resources: Resources[];
  defaultNodePools?: string[];
  parentResources?: Resources[] | null;
  parentDefaultNodePools?: string[];
  // Generic status data - works for both projects and departments
  nodePoolQuotaStatuses?: QuotaStatusNodePool[] | null;
  // For departments only: aggregated resources of subordinate projects
  subordinateResources?: AggregatedResources[] | null;
  // For departments only: node pool total CPU/memory from cluster telemetry
  nodePoolTotals?: INodePoolTotals | null;
  loading?: boolean;
  readOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  parentResources: null,
  parentDefaultNodePools: undefined,
  nodePoolQuotaStatuses: null,
  subordinateResources: null,
  nodePoolTotals: null,
  loading: false,
  readOnly: false,
});

const readOnly = computed(() => props.readOnly);

const emit = defineEmits<{
  (
    e: "save-cpu-quota",
    payload: {
      nodePoolName: string;
      cpuQuota: number | null;
      memoryQuota: number | null;
      memoryUnit: Units;
      navigateAfterSave?: boolean;
    },
  ): void;
  (e: "close"): void;
}>();
const STEP_NAME = EStepNames.CPUQuota;
const unsavedGuard = useUnsavedChangesGuard();

const settingStore = useSettingStore();

// CPU quota value
const cpuQuota = ref<number | null>(null);

// Memory quota value and unit
const memoryQuota = ref<number | null>(null);
const memoryUnit = ref<Units>(Units.Mb);

// Node pool state management (read-only mode)
const { allNodePools, selectedNodePools, selectedNodePoolId, selectedNodePoolName } = useOrgUnitNodePools({
  resources: () => props.resources || [],
  defaultNodePools: () => props.defaultNodePools ?? undefined,
  parentDefaultNodePools: () => props.parentDefaultNodePools ?? undefined,
  editable: false,
});

// Single node pool scenario - hide save button and use wizard footer instead
const isSingleNodePool = computed(() => allNodePools.value.length === 1);

const getResourceSnapshot = (nodePoolName: string) => {
  const resource = props.resources?.find((r) => (r.nodePool?.name || "") === nodePoolName);

  // CPU: treat -1 as null (unlimited)
  const cpuDeserved = resource?.cpu?.deserved;
  const cpuValue =
    cpuDeserved !== undefined && cpuDeserved !== null && cpuDeserved !== -1 ? cpuDeserved / CPU_VALUE_FACTOR : null;

  // Memory: treat -1 as null (unlimited)
  const memoryDeserved = resource?.memory?.deserved;
  const memoryValue =
    memoryDeserved !== undefined && memoryDeserved !== null && memoryDeserved !== -1 ? memoryDeserved : null;

  const unitValue = resource?.memory?.units || Units.Mb;

  return {
    cpu: cpuValue,
    memory: memoryValue,
    memoryUnit: unitValue,
  };
};

const syncFromSnapshot = (snapshot: { cpu: number | null; memory: number | null; memoryUnit: Units }) => {
  cpuQuota.value = snapshot.cpu;
  memoryQuota.value = snapshot.memory;
  memoryUnit.value = snapshot.memoryUnit;
};

// Computed: Current pool's resource snapshot (used to detect when saved data changes)
const currentPoolResourceSnapshot = computed(() => {
  if (!selectedNodePoolName.value) return null;
  return getResourceSnapshot(selectedNodePoolName.value);
});

// Load CPU/memory quota when selected node pool changes OR when its saved resource data changes
// (e.g., after a save operation updates the project)
watch(
  [() => selectedNodePoolName.value, currentPoolResourceSnapshot],
  ([nodePoolName, snapshot]) => {
    if (nodePoolName && snapshot) {
      syncFromSnapshot(snapshot);
    } else {
      syncFromSnapshot({ cpu: null, memory: null, memoryUnit: Units.Mb });
    }
  },
  { immediate: true },
);

// Parent limit (in cores) for selected node pool
// For projects: department quota. For departments: node pool total from cluster telemetry.
const parentCpuLimit = computed(() => {
  if (props.entityType === EQuotaEntity.department) {
    // For departments: use node pool total from cluster telemetry
    if (!props.nodePoolTotals) return null;
    const total = props.nodePoolTotals[selectedNodePoolName.value];
    return total?.cpuCores ?? null;
  }

  // For projects: use department quota (parentResources)
  if (!props.parentResources) return null;

  const resource = props.parentResources.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value);

  const deserved = resource?.cpu?.deserved;
  if (deserved === undefined || deserved === null) return null; // Unlimited

  return deserved / CPU_VALUE_FACTOR;
});

// Allocated non-preemptible CPU (in cores) for selected node pool
const allocatedNonPreemptibleCpu = computed(() => {
  if (!props.nodePoolQuotaStatuses) return 0;

  const status = props.nodePoolQuotaStatuses.find((s) => (s.nodePoolName || "") === selectedNodePoolName.value);

  const allocated = status?.allocatedNonPreemptible?.cpu || 0;
  return allocated / CPU_VALUE_FACTOR; // Convert to cores
});

// Subordinate (child projects') CPU quota for selected node pool - only for departments
const subordinateCpuQuota = computed(() => {
  // If no subordinateResources provided, return undefined (no data to validate)
  if (!props.subordinateResources || props.subordinateResources.length === 0) {
    return undefined;
  }

  const aggregated = props.subordinateResources.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value);

  // If node pool not found, return undefined (no data for this node pool)
  if (!aggregated) return undefined;

  const deserved = aggregated?.cpu?.deserved;
  // If deserved is undefined, return undefined (no data)
  if (deserved === undefined) return undefined;
  // If deserved is null or -1, it means unlimited - return null
  if (deserved === null || deserved === -1) return null;

  // Convert numeric value to cores
  return deserved / CPU_VALUE_FACTOR;
});

// CPU quota validation
const { quotaState: cpuQuotaState, quotaMessage: cpuQuotaMessage } = useQuotaValidation({
  quotaValue: cpuQuota,
  parentLimit: parentCpuLimit,
  allocatedNonPreemptible: allocatedNonPreemptibleCpu,
  isLimitQuotaOverSubscriptionEnabled: computed(() => settingStore.isLimitQuotaOverSubscriptionEnabled),
  resourceName: "CPU",
  resourceUnit: "cores",
  entityType: props.entityType,
  subordinateQuota: subordinateCpuQuota,
  availabilityTotal: parentCpuLimit,
});

const parentMemoryResource = computed(() => {
  if (!props.parentResources) return null;
  return props.parentResources.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value) || null;
});

// Parent memory limit for selected node pool
// For projects: department quota. For departments: node pool total from cluster telemetry.
const BYTES_TO_MIB = 1048576; // 1 MiB = 1,048,576 bytes

const parentMemoryQuota = computed(() => {
  if (props.entityType === EQuotaEntity.department) {
    // For departments: use node pool total from cluster telemetry (in bytes, convert to current unit)
    if (!props.nodePoolTotals) return null;
    const total = props.nodePoolTotals[selectedNodePoolName.value];
    if (!total?.memoryBytes) return null;

    // Convert from bytes to MiB first, then to the target unit
    const valueInMib = total.memoryBytes / BYTES_TO_MIB;
    const converted = orgUnitUtil.convertMemoryValue(valueInMib, Units.Mib, memoryUnit.value);
    return converted ?? null;
  }

  // For projects: use department quota (parentResources)
  const rawValue = parentMemoryResource.value?.memory?.deserved ?? null;
  const rawUnit = parentMemoryResource.value?.memory?.units ?? Units.Mib;
  const converted = orgUnitUtil.convertMemoryValue(rawValue, rawUnit, memoryUnit.value);
  return converted ?? rawValue;
});

// Allocated non-preemptible memory for selected node pool
const allocatedNonPreemptibleMemory = computed(() => {
  if (!props.nodePoolQuotaStatuses) return 0;

  const status = props.nodePoolQuotaStatuses.find((s) => (s.nodePoolName || "") === selectedNodePoolName.value);

  const rawValue = status?.allocatedNonPreemptible?.memory || 0;
  const converted = orgUnitUtil.convertMemoryValue(rawValue, MEMORY_ALLOCATED_NON_PREEMPTIBLE_UNITS, memoryUnit.value);
  return converted ?? rawValue ?? 0;
});

// Subordinate (child projects') memory quota for selected node pool - only for departments
const subordinateMemoryQuota = computed(() => {
  // If no subordinateResources provided, return undefined (no data to validate)
  if (!props.subordinateResources || props.subordinateResources.length === 0) {
    return undefined;
  }

  const aggregated = props.subordinateResources.find((r) => (r.nodePool?.name || "") === selectedNodePoolName.value);

  // If node pool not found, return undefined (no data for this node pool)
  if (!aggregated) return undefined;

  const deserved = aggregated?.memory?.deserved;
  // If deserved is undefined, return undefined (no data)
  if (deserved === undefined) return undefined;
  // If deserved is null or -1, it means unlimited
  if (deserved === null || deserved === -1) return null;

  // Convert to current memory unit
  const rawUnit = aggregated?.memory?.units ?? Units.Mb;
  const converted = orgUnitUtil.convertMemoryValue(deserved, rawUnit, memoryUnit.value);
  return converted ?? deserved;
});

// Memory quota validation
const { quotaState: memoryQuotaState, quotaMessage: memoryQuotaMessage } = useQuotaValidation({
  quotaValue: memoryQuota,
  parentLimit: parentMemoryQuota,
  allocatedNonPreemptible: allocatedNonPreemptibleMemory,
  isLimitQuotaOverSubscriptionEnabled: computed(() => settingStore.isLimitQuotaOverSubscriptionEnabled),
  resourceName: "CPU memory",
  resourceUnit: computed(() => memoryUnit.value),
  entityType: props.entityType,
  subordinateQuota: subordinateMemoryQuota,
  availabilityTotal: parentMemoryQuota,
});

// Check if current node pool has unsaved changes
const isDirty = computed(() => {
  if (!selectedNodePoolName.value) {
    return false;
  }
  const snapshot = getResourceSnapshot(selectedNodePoolName.value);
  return (
    cpuQuota.value !== snapshot.cpu || memoryQuota.value !== snapshot.memory || memoryUnit.value !== snapshot.memoryUnit
  );
});

// Handle node pool tab change (receives ID from tabs)
const handleNodePoolTabChange = async (nodePoolId: string) => {
  const pool = selectedNodePools.value.find((np) => np.id === nodePoolId);
  if (!pool) return;

  const applySwitch = () => {
    selectedNodePoolId.value = nodePoolId;
    syncFromSnapshot(getResourceSnapshot(pool.name));
  };

  if (!isDirty.value) {
    applySwitch();
    return;
  }

  await unsavedGuard.requestAction({
    stepName: STEP_NAME,
    onProceed: applySwitch,
  });
};

// Handle save button click
const handleSave = async (navigateAfterSave = false) => {
  const cpuQuotaToSave = cpuQuota.value === -1 || cpuQuota.value === null ? null : cpuQuota.value * CPU_VALUE_FACTOR;
  const memoryQuotaToSave = memoryQuota.value === -1 ? null : memoryQuota.value;

  emit("save-cpu-quota", {
    nodePoolName: selectedNodePoolName.value,
    cpuQuota: cpuQuotaToSave,
    memoryQuota: memoryQuotaToSave,
    memoryUnit: memoryUnit.value,
    navigateAfterSave,
  });
};

const handleSaveAndContinue = () => {
  handleSave(true);
};

const discardChanges = () => {
  if (!selectedNodePoolName.value) {
    syncFromSnapshot({ cpu: null, memory: null, memoryUnit: Units.Mb });
    return;
  }
  syncFromSnapshot(getResourceSnapshot(selectedNodePoolName.value));
};

// Register unsaved changes guard - unregister on unmount
const unregisterGuard = unsavedGuard.registerStepGuard(STEP_NAME, {
  isDirty: () => isDirty.value,
  save: handleSave,
  discard: discardChanges,
});

onBeforeUnmount(() => {
  unregisterGuard();
});

defineExpose({
  isDirty,
  save: handleSave,
  discardChanges,
});
</script>
<style scoped lang="scss">
.cpu-input-box {
  margin-top: 20px;
  margin-bottom: 31px;
}
</style>
