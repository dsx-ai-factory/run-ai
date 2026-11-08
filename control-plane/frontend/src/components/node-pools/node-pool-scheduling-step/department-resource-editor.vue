<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <editable-cell-popup
    :title="kind === 'weight' ? 'Weight' : 'Guaranteed GPUs'"
    :display-value="displayValue ?? value ?? 'Unknown'"
    :disabled="disabled"
    :disabled-tooltip="disabledTooltip"
    :is-submit-disabled="loading || validation.errors.length > 0"
    @opened="opened"
    @reset="reset"
    @apply="apply"
  >
    <template #input>
      <q-input
        v-model="input"
        type="number"
        min="0"
        dense
        :label="kind === 'weight' ? 'Raw weight' : 'Guaranteed GPUs'"
        :error="validation.errors.length > 0"
        :error-message="validation.errors.join(' ')"
        no-error-icon
      />
      <p v-if="loading" role="status">Loading project quotas…</p>
      <p v-for="warning in validation.warnings" :key="warning" class="text-warning" role="status">{{ warning }}</p>
      <q-btn flat label="Cancel" v-close-popup @click="reset" />
    </template>
  </editable-cell-popup>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { EditableCellPopup } from "@/components/common/runai-table/runai-custom-table-columns/editable-cell-popup";
import { validateDepartmentEdit, type DepartmentValidationContext } from "./node-pool-scheduling.model";
const props = defineProps<{
  id: string;
  value?: number;
  kind: "weight" | "quota";
  displayValue?: string;
  disabled?: boolean;
  disabledTooltip?: string;
  limit?: number | null;
  context?: DepartmentValidationContext;
  loadContext?: () => Promise<DepartmentValidationContext>;
}>();
const emit = defineEmits<{ apply: [number] }>();
const input = ref<string | number>(props.value ?? "");
const context = ref<DepartmentValidationContext>(props.context ?? {});
const loading = ref(false);
const loadError = ref("");
let request = 0;
const validation = computed(() => {
  const value = String(input.value).trim() === "" ? NaN : Number(input.value);
  const result = validateDepartmentEdit(
    { id: props.id, ...(props.kind === "weight" ? { weight: value } : { quota: value, limit: props.limit }) },
    {
      ...context.value,
      totalQuota:
        context.value.totalQuota != null && props.value != null
          ? context.value.totalQuota + value - props.value
          : undefined,
    },
  );
  if (loadError.value) result.errors.push(loadError.value);
  return result;
});
function reset() {
  request++;
  input.value = props.value ?? "";
  loading.value = false;
  loadError.value = "";
}
async function opened() {
  input.value = props.value ?? (props.kind === "weight" ? 2 : "");
  context.value = props.context ?? {};
  if (!props.loadContext) return;
  const version = ++request;
  loading.value = true;
  loadError.value = "";
  try {
    const result = await props.loadContext();
    if (version === request) context.value = result;
  } catch {
    if (version === request) loadError.value = "Validation data is unavailable. Close and reopen to retry.";
  } finally {
    if (version === request) loading.value = false;
  }
}
function apply() {
  if (!loading.value && !validation.value.errors.length) emit("apply", Number(input.value));
}
watch(() => props.value, reset);
</script>
