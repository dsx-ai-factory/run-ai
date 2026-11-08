<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <editable-cell-popup
    :title="editorTitle"
    popup-class="scheduling-resource-popup"
    :popup-width="kind === 'weight' ? '274px' : '200px'"
    :state="
      active && !loading
        ? validation.errors.length
          ? 'error'
          : validation.warnings.length
            ? 'warning'
            : 'active'
        : 'default'
    "
    :display-value="
      kind === 'limit' && (value == null || value === -1) ? 'Unlimited' : (displayValue ?? value ?? 'Unknown')
    "
    :disabled="disabled"
    :disabled-tooltip="disabledTooltip"
    :is-submit-disabled="loading || validation.errors.length > 0"
    @opened="opened"
    @reset="reset"
    @apply="apply"
  >
    <template #title
      >{{ editorTitle }}
      <q-icon v-if="kind === 'weight'" name="ph ph-question" size="16px" class="q-ml-xs text-grey-7"
        ><q-tooltip max-width="320px"
          >Weight determines a department’s share relative to the other departments{{
            title === "Weight within rank" ? " in the same rank" : ""
          }}. A higher weight means a larger portion of the resources. The percentage is calculated from the total
          weight.</q-tooltip
        ></q-icon
      ></template
    >
    <template #input>
      <q-input
        v-model="input"
        :type="kind === 'limit' ? 'text' : 'number'"
        min="0"
        dense
        :aria-label="editorTitle"
        :suffix="kind === 'weight' && displayValue ? `(${displayValue})` : undefined"
        :placeholder="kind === 'limit' ? 'Unlimited' : undefined"
        :error="validation.errors.length > 0"
        :error-message="errorMessage"
        no-error-icon
      />
      <p
        v-if="kind === 'quota' && !loading && !validation.errors.length && !validation.warnings.length"
        class="quota-hint"
      >
        <template v-if="context.childQuotaSum != null"
          >Recommended: from {{ context.childQuotaSum
          }}<template v-if="context.capacity != null"> to {{ context.capacity }}</template
          >.<br
        /></template>
        Non-preemptible workload usage:
        {{ context.nonPreemptibleUsage == null ? "Unknown" : `${context.nonPreemptibleUsage} GPUs` }}.<br />
        <template v-if="context.capacity != null">Node pool’s total: {{ context.capacity }} GPUs.</template>
      </p>
      <p v-if="loading" role="status">Loading project quotas…</p>
      <p v-for="warning in validation.warnings" :key="warning" class="text-warning" role="status">
        {{ warningMessage(warning) }}
      </p>
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
  kind: "weight" | "quota" | "limit";
  title?: string;
  quota?: number;
  displayValue?: string;
  disabled?: boolean;
  disabledTooltip?: string;
  limit?: number | null;
  context?: DepartmentValidationContext;
  loadContext?: () => Promise<DepartmentValidationContext>;
}>();
const emit = defineEmits<{ apply: [number] }>();
const editorTitle = computed(
  () =>
    props.title ??
    (props.kind === "weight" ? "Weight" : props.kind === "limit" ? "Max GPU allocation" : "Guaranteed GPUs"),
);
const initialInput = () =>
  props.kind === "limit" && (props.value == null || props.value === -1) ? "Unlimited" : (props.value ?? "");
const parsedValue = () =>
  props.kind === "limit" &&
  (String(input.value).trim().toLowerCase() === "unlimited" || String(input.value).trim() === "")
    ? -1
    : String(input.value).trim() === ""
      ? NaN
      : Number(input.value);
const input = ref<string | number>(props.value ?? "");
const context = ref<DepartmentValidationContext>(props.context ?? {});
const errorMessage = computed(() =>
  validation.value.errors
    .map((message) => {
      if (message === "Weight must be a whole number of 0 or greater.")
        return "Enter a whole number equal to or above 0";
      if (message === "Guaranteed GPUs must be 0 or greater.") return "Enter a value equal to or above 0";
      if (message === "Guaranteed GPUs are below the child project quota sum.") return childQuotaMessage();
      return message;
    })
    .join(" "),
);
function childQuotaMessage() {
  return `The department’s quota is less than the quota assigned to its subordinate projects. To remain within the limit, enter a number equal to or above ${context.value.childQuotaSum} GPUs.`;
}
function warningMessage(message: string) {
  if (message === "Guaranteed GPUs are below the child project quota sum.") return childQuotaMessage();
  if (
    message === "Assigned quota exceeds node pool capacity." &&
    Number(input.value) > (context.value.capacity ?? Infinity)
  )
    return `The department’s quota exceeds the node pool’s total. To remain within the limit, enter a number up to ${context.value.capacity} GPUs.`;
  return message;
}
const loading = ref(false);
const active = ref(false);
const loadError = ref("");
let request = 0;
const validation = computed(() => {
  const value = parsedValue();
  const result = validateDepartmentEdit(
    {
      id: props.id,
      ...(props.kind === "weight"
        ? { weight: value }
        : props.kind === "limit"
          ? { limit: value }
          : { quota: value, limit: props.limit }),
    },
    {
      ...context.value,
      totalQuota:
        context.value.totalQuota != null && props.value != null
          ? context.value.totalQuota + value - props.value
          : undefined,
    },
  );
  if (props.kind === "limit" && value !== -1 && props.quota != null && value < props.quota)
    result.errors.push(
      `The department’s maximum GPU device allocation is below its quota. To stay above the quota, enter a number equal to or above ${props.quota} GPUs.`,
    );
  if (loadError.value) result.errors.push(loadError.value);
  return result;
});
function reset() {
  active.value = false;
  request++;
  input.value = initialInput();
  loading.value = false;
  loadError.value = "";
}
async function opened() {
  active.value = true;
  input.value = props.kind === "weight" ? (props.value ?? 2) : initialInput();
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
  if (!loading.value && !validation.value.errors.length) emit("apply", parsedValue());
}
watch(() => props.value, reset);
</script>

<style scoped lang="scss">
p {
  font-size: 12px;
  line-height: 16px;
  margin: 12px 0 0;
}
</style>

<style lang="scss">
.scheduling-resource-popup .popup-card .q-card__section:first-child {
  min-height: 44px;
  padding: 8px 16px;
}
.scheduling-resource-popup .popup-card .q-card__section:nth-child(2) {
  padding-top: 0;
}
.scheduling-resource-popup .popup-card .q-card__section:last-child {
  padding: 8px 16px;
}
.scheduling-resource-popup .q-input .q-field__native {
  padding-top: 0;
}
.scheduling-resource-popup .q-field__control {
  height: 38px;
}
.scheduling-resource-popup .q-field__bottom {
  padding-left: 0;
  padding-right: 0;
}
</style>
