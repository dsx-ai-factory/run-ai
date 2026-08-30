<template>
  <editable-cell-popup
    title="Weight"
    :display-value="displayValue"
    :state="cellState"
    :is-submit-disabled="hasError"
    :disabled="isDisabled"
    :disabled-tooltip="disabledTooltip"
    @apply="handleApply"
    @reset="handleReset"
  >
    <template #input>
      <q-input
        v-model.number="editableValue"
        type="number"
        dense
        min="0"
        :error="hasError"
        :error-message="errorMessage"
        no-error-icon
      />
    </template>
  </editable-cell-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import type { ICustomCellEditableWeightColConfig } from "@/table-models/org-unit.table-model";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import { EditableCellPopup, type EditableCellState } from "./editable-cell-popup";

export interface IEditableWeightColRow {
  id: string;
  orgUnitId: string;
  weight: number | null;
  serverWeight?: number | null; // Original weight from server
  deserved?: number | null; // GPU deserved value (used when disabled)
}

defineOptions({ name: "OrgUnitEditableWeightCol" });

const props = defineProps<{
  column: ITableColumn;
  row: IEditableWeightColRow;
}>();

const config = computed(() => {
  return getCustomCellConfig(props.column, ECustomCell.ORG_UNIT_EDITABLE_WEIGHT_COL) as
    | ICustomCellEditableWeightColConfig
    | undefined;
});

// Disabled state from config
const isDisabled = computed(() => config.value?.disabled ?? false);
const disabledTooltip = computed(() => config.value?.disabledTooltip ?? "");

const editableValue = ref<number>(props.row.weight ?? 0);
const originalValue = ref<number | null>(props.row.weight);

// Validation - weight must be >= 0
const hasError = computed(() => {
  return editableValue.value < 0;
});

const errorMessage = computed(() => {
  if (hasError.value) {
    return "Weight must be 0 or greater";
  }
  return "";
});

// Display value - show deserved when disabled, otherwise show weight
const displayValue = computed(() => {
  // 1. Select the value based on the state
  const val = isDisabled.value ? props.row.deserved : props.row.weight;

  // 2. Apply formatting logic (checks for null, undefined, or -1)
  if (val == null || val === -1) {
    return "-";
  }

  return String(val);
});

// Cell state for styling
const cellState = computed<EditableCellState>(() => {
  if (hasError.value) return "error";
  return "default";
});

const handleReset = () => {
  editableValue.value = originalValue.value ?? 0;
};

const handleApply = () => {
  if (hasError.value || !config.value?.onSave) return;

  config.value.onSave({
    orgUnitId: props.row.orgUnitId,
    nodePoolId: config.value.nodePoolId,
    weight: editableValue.value,
    originalWeight: props.row.serverWeight ?? props.row.weight,
  });

  originalValue.value = editableValue.value;
};

watch(
  () => props.row.weight,
  (newValue) => {
    editableValue.value = newValue ?? 0;
    originalValue.value = newValue;
  },
);
</script>
