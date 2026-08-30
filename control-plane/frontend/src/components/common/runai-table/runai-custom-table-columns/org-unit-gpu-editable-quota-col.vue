<template>
  <editable-cell-popup
    title="GPUs"
    :display-value="displayValue"
    :state="cellState"
    :is-submit-disabled="hasError || isLoadingDepartment"
    @opened="fetchDepartmentSubordinateQuota"
    @apply="handleApply"
    @reset="handleReset"
  >
    <template #input>
      <gpu-quota-input
        v-model="editableValue"
        :state="quotaState"
        :message="quotaMessage"
        :loading="isLoadingDepartment"
        :entity-type="config?.entityType"
      />
    </template>
  </editable-cell-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import type { ICustomCellGpuQuotaColConfig, IQuotaTableRow } from "@/table-models/org-unit.table-model";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import { GpuQuotaInput } from "@/components/org-unit/v2/common/gpu-quota-input";
import { useQuotaValidation } from "@/composables/org-unit/use-quota-validation.composable";
import { useSettingStore } from "@/stores/setting.store";
import { EQuotaInputState } from "@/models/org-unit.model";
import { EditableCellPopup, type EditableCellState } from "./editable-cell-popup";
import { EQuotaEntity } from "@/models/resource.model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Department, AggregatedResources } from "@/swagger-models/org-unit-service-client";

defineOptions({ name: "OrgUnitGpuEditableQuotaCol" });

const props = defineProps<{
  column: ITableColumn;
  row: IQuotaTableRow;
}>();

const settingStore = useSettingStore();

const config = computed(() => {
  return getCustomCellConfig(props.column, ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL) as
    | ICustomCellGpuQuotaColConfig
    | undefined;
});

const editableValue = ref<number>(props.row.gpuQuota ?? 0);
const originalValue = ref<number | null>(props.row.gpuQuota);

// Loading and subordinate quota state for department validation
const isLoadingDepartment = ref(false);
const subordinateGpuQuota = ref<number | null | undefined>(undefined);

// Fetch department subordinate quota when popup opens
const fetchDepartmentSubordinateQuota = async () => {
  if (config.value?.entityType !== EQuotaEntity.department) return;

  const departmentId = props.row.id;
  const nodePoolName = config.value?.nodePoolName;

  if (!departmentId || !nodePoolName) return;

  isLoadingDepartment.value = true;

  try {
    const department: Department = await orgUnitService.getDepartment(departmentId);
    const projectsAggregated = department.projectsAggregatedResources?.find(
      (resource: AggregatedResources) => resource.nodePool?.name === nodePoolName,
    );
    if (!projectsAggregated) {
      // No data means no quota assigned
      subordinateGpuQuota.value = undefined;
      return;
    }
    const gpuDeserved = projectsAggregated?.gpu?.deserved;
    subordinateGpuQuota.value =
      gpuDeserved === null || gpuDeserved === undefined || gpuDeserved === -1 ? null : gpuDeserved;
  } catch (error) {
    console.error("Failed to fetch department subordinate quota:", error);
    subordinateGpuQuota.value = undefined; // Allow editing if fetch fails
  } finally {
    isLoadingDepartment.value = false;
  }
};

// Calculate max assignable quota for this row
// For projects: Available quota + this row's current quota (since reducing frees up space)
// For departments: Just use node pool's total (departments aren't limited by a parent)
const parentLimit = computed(() => {
  // For departments, use the node pool's total as the limit
  if (config.value?.entityType === EQuotaEntity.department) {
    return config.value?.totalQuota ?? null;
  }

  // For projects, calculate based on available + current quota
  const available = config.value?.availableQuota;
  if (available === null || available === undefined) {
    return config.value?.totalQuota ?? null; // Fallback to total if available is not set
  }

  // Add current/pending quota to available
  // Use gpuQuota (not serverGpuQuota) because availableQuota already accounts for current values
  const currentQuota = props.row.gpuQuota ?? 0;
  return Math.max(0, available + currentQuota);
});

// Validation using the composable (same as in gpu-quota-step)
const allocatedNonPreemptibleGpu = computed(() => props.row?.allocatedNonPreemptibleGpu ?? 0);

const { quotaState, quotaMessage } = useQuotaValidation({
  requireNumberGreaterOrEqualToZero: true,
  quotaValue: editableValue,
  parentLimit: parentLimit,
  allocatedNonPreemptible: allocatedNonPreemptibleGpu,
  subordinateQuota: subordinateGpuQuota,
  isLimitQuotaOverSubscriptionEnabled: computed(() => settingStore.isLimitQuotaOverSubscriptionEnabled),
  resourceName: "GPU",
  resourceUnit: "GPUs",
  entityType: config.value?.entityType ?? EQuotaEntity.project,
  availabilityTotal: computed(() => config.value?.totalQuota ?? null),
  isDefaultDepartmentUnlimited: computed(() => config.value?.isDefaultDepartmentUnlimited ?? false),
});

// Display value
const displayValue = computed(() => {
  const val = props.row.gpuQuota;

  return val === null || val === undefined || val === -1 ? "Unlimited" : val.toString();
});

// Cell state for styling - map EQuotaInputState to EditableCellState
const cellState = computed<EditableCellState>(() => {
  switch (quotaState.value) {
    case EQuotaInputState.Error:
      return "error";
    case EQuotaInputState.Warning:
      return "warning";
    default:
      return "default";
  }
});

// Disable apply if there's an error
const hasError = computed(() => quotaState.value === EQuotaInputState.Error);

const handleReset = () => {
  editableValue.value = originalValue.value ?? 0;
};

const handleApply = () => {
  if (hasError.value || !config.value?.onSave) return;

  config.value.onSave({
    projectId: props.row.id,
    entityName: props.row.name,
    nodePoolName: config.value.nodePoolName,
    nodePoolId: config.value.nodePoolId,
    gpuQuota: editableValue.value,
    originalValue: props.row.serverGpuQuota ?? props.row.gpuQuota,
    originalLimit: props.row.serverGpuLimit ?? null,
    originalOverQuotaWeight: props.row.serverOverQuotaWeight ?? null,
  });

  originalValue.value = editableValue.value;
};

watch(
  () => props.row.gpuQuota,
  (newValue) => {
    editableValue.value = newValue ?? 0;
    originalValue.value = newValue;
  },
);

defineExpose({
  fetchDepartmentSubordinateQuota,
  quotaMessage,
  editableValue,
  quotaState,
  hasError,
  isLoadingDepartment,
  subordinateGpuQuota,
  parentLimit,
  allocatedNonPreemptibleGpu,
  handleApply,
  handleReset,
});
</script>
