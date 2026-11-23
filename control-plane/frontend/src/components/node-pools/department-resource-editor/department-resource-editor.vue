<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <editable-cell-popup
    :class="{
      'unlimited-allocation': kind === EResourceEditorKind.Limit && (value == null || value === -1),
      'resource-cell--error': active && !loading && validation.errors.length > 0,
      'resource-cell--warning': active && !loading && !validation.errors.length && validation.warnings.length > 0,
    }"
    :title="editorTitle"
    :popup-class="`scheduling-resource-popup scheduling-resource-popup--${kind}`"
    :display-value="
      (kind === EResourceEditorKind.Limit && value == null) || (kind !== EResourceEditorKind.Weight && value === -1)
        ? 'Unlimited'
        : (displayValue ?? value ?? 'Unknown')
    "
    :disabled="disabled"
    :disabled-tooltip="disabledTooltip"
    :is-submit-disabled="loading || validation.errors.length > 0"
    @before-show="opened"
    @reset="reset"
    @apply="apply"
  >
    <template #title>
      <span class="resource-editor-title">
        {{ editorTitle }}
        <scheduling-help-tooltip
          v-if="kind === EResourceEditorKind.Weight"
          class="weight-editor-help q-ml-xs"
          icon="ph-bold ph-question"
          width="300px"
          :tooltip-text="weightHelp"
          anchor="center right"
          self="center left"
          :offset="rankedWeight ? [120, 0] : [180, 0]"
          ><template #content
            ><span class="node-pool-weight-help"
              ><strong>Weight</strong> determines a department’s share relative to the other departments{{
                rankedWeight ? " in the same rank" : ""
              }}. A higher weight means a larger portion of the resources.
              <strong>The weight is expressed as a percentage</strong>, calculated as the weight divided by the total
              weight of all departments{{ rankedWeight ? " in the rank" : "" }}.<br />For example, a department with a
              weight of 2{{ rankedWeight ? " in a rank" : "" }} where the weights total 20 receives 10%.</span
            ></template
          ></scheduling-help-tooltip
        >
      </span>
    </template>
    <template #input>
      <div :class="{ 'weight-input-row': kind === EResourceEditorKind.Weight }">
        <q-input
          v-model="input"
          type="number"
          :inputmode="kind === EResourceEditorKind.Weight ? 'numeric' : 'decimal'"
          @keydown="numericKeydown($event, kind === EResourceEditorKind.Weight)"
          @paste="numericPaste($event, kind === EResourceEditorKind.Weight)"
          @beforeinput="numericBeforeInput($event, kind === EResourceEditorKind.Weight)"
          @input.capture="updateNativeValidity"
          min="0"
          :dense="kind !== EResourceEditorKind.Limit"
          :label="kind === EResourceEditorKind.Limit ? editorTitle : undefined"
          :stack-label="kind === EResourceEditorKind.Limit"
          :aria-label="editorTitle"
          :max="kind === EResourceEditorKind.Weight ? 20 : undefined"
          :step="kind === EResourceEditorKind.Weight ? 1 : 'any'"
          :placeholder="kind === EResourceEditorKind.Limit ? 'Unlimited' : undefined"
          :error="validation.errors.length > 0"
          :color="!validation.errors.length && validation.warnings.length ? 'warning' : 'primary'"
          :class="{
            'resource-input-warning': !validation.errors.length && validation.warnings.length,
            'resource-limit-input': kind === EResourceEditorKind.Limit,
          }"
          :error-message="errorMessage"
          no-error-icon
        />
        <span v-if="kind === EResourceEditorKind.Weight && !validation.errors.length" class="weight-percentage"
          >({{ proposedWeightPercentage }})</span
        >
      </div>
      <p
        v-if="kind === EResourceEditorKind.Quota && !loading && !validation.errors.length && !validation.warnings.length"
        class="quota-hint"
      >
        <template v-if="recommendedMinimum != null"
          >Recommended: from {{ recommendedMinimum
          }}<template v-if="quotaBudget"> to {{ quotaBudget.available }}</template
          >.<br
        /></template>
        Non-preemptible workload usage:
        {{ context.nonPreemptibleUsage == null ? "Unknown" : `${context.nonPreemptibleUsage} GPUs` }}.<br />
        <template v-if="context.capacity != null">Node pool’s total: {{ context.capacity }} GPUs.</template>
      </p>
      <q-spinner v-if="loading" color="primary" size="20px" role="progressbar" aria-label="Loading project quotas" />
      <template v-if="!validation.errors.length">
        <p v-for="warning in validation.warnings" :key="warning" class="text-warning" role="status">
          {{ warningMessage(warning) }}
        </p>
      </template>
    </template>
  </editable-cell-popup>
</template>
<script setup lang="ts">
import { roundToDecimal } from "@/utils/format.util/index";
import { EResourceEditorKind } from "@/models/node-pool-scheduling.model";
import { numericKeydown, numericPaste, numericBeforeInput } from "@/utils/numeric-input.util/numeric-input.util";
import { computed, ref, watch } from "vue";
import { SchedulingHelpTooltip } from "@/components/node-pools/scheduling-help-tooltip";
import { EditableCellPopup } from "@/components/common/runai-table/runai-custom-table-columns/editable-cell-popup/index";
import {
  recommendedDepartmentQuotaMinimum,
  departmentQuotaBudget,
  validateDepartmentEdit,
  isValidSchedulingWeight,
  isDepartmentLimitBelowQuota,
  type IDepartmentValidationContext,
} from "@/models/node-pool-scheduling.model";
const props = defineProps<{
  id: string;
  value?: number;
  kind: EResourceEditorKind;
  title?: string;
  quota?: number;
  displayValue?: string;
  previewWeightPercentage?: (weight: number) => number | undefined;
  disabled?: boolean;
  disabledTooltip?: string;
  limit?: number | null;
  context?: IDepartmentValidationContext;
  loadContext?: () => Promise<IDepartmentValidationContext>;
}>();
const emit = defineEmits<{ apply: [number] }>();
const editorTitle = computed(
  () =>
    props.title ??
    (props.kind === EResourceEditorKind.Weight
      ? "Weight"
      : props.kind === EResourceEditorKind.Limit
        ? "Max GPU allocation"
        : "Guaranteed GPUs"),
);
const rankedWeight = computed(() => props.title === "Weight within rank");
const weightHelp = computed(
  () =>
    `Weight determines a department’s share relative to the other departments${rankedWeight.value ? " in the same rank" : ""}. A higher weight means a larger portion of the resources. The weight is expressed as a percentage, calculated as the weight divided by the total weight of all departments${rankedWeight.value ? " in the rank" : ""}. For example, a department with a weight of 2${rankedWeight.value ? " in a rank" : ""} where the weights total 20 receives 10%.`,
);
const initialInput = () =>
  props.kind === EResourceEditorKind.Limit && (props.value == null || props.value === -1) ? "" : (props.value ?? "");
const nativeBadInput = ref(false);
function updateNativeValidity(event: Event) {
  if (event.target instanceof HTMLInputElement) nativeBadInput.value = event.target.validity.badInput;
}
const parsedValue = () =>
  nativeBadInput.value
    ? NaN
    : props.kind === EResourceEditorKind.Limit && String(input.value).trim() === ""
      ? -1
      : String(input.value).trim() === ""
        ? NaN
        : Number(input.value);
const input = ref<string | number>(props.value ?? "");
const proposedWeightPercentage = computed(() => {
  const value = parsedValue();
  if (!isValidSchedulingWeight(value)) return "Unknown";
  const percentage = props.previewWeightPercentage?.(value);
  return percentage == null || !Number.isFinite(percentage) ? "Unknown" : `${roundToDecimal(percentage)}%`;
});
const context = ref<IDepartmentValidationContext>(props.context ?? {});
const recommendedMinimum = computed(() => recommendedDepartmentQuotaMinimum(context.value));
const quotaBudget = computed(() => departmentQuotaBudget(props.value, context.value));
const errorMessage = computed(() =>
  validation.value.errors
    .map((message) => {
      if (message === "Weight must be a whole number from 0 to 20.") return "Enter a whole number from 0 to 20";
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
  if (message === "Assigned quota exceeds node pool capacity.")
    return quotaBudget.value
      ? `The assigned quota exceeds the node pool’s total. To remain within the limit, enter a number up to ${quotaBudget.value.available} GPUs.`
      : "The assigned quota exceeds the node pool’s total. Reduce the department’s quota to remain within the limit.";
  return message;
}
const loading = ref(false);
const active = ref(false);
const loadError = ref("");
let request = 0;
const validation = computed(() => {
  if (loading.value) return { errors: [], warnings: [] };
  const value = parsedValue();
  const result = validateDepartmentEdit(
    {
      id: props.id,
      ...(props.kind === EResourceEditorKind.Weight
        ? { weight: value }
        : props.kind === EResourceEditorKind.Limit
          ? { limit: value }
          : { quota: value, limit: props.limit }),
    },
    {
      ...context.value,
      totalQuota:
        props.kind === EResourceEditorKind.Quota && quotaBudget.value
          ? quotaBudget.value.otherFiniteQuota + value
          : context.value.totalsComplete !== false && context.value.totalQuota != null && props.value != null
            ? context.value.totalQuota + value - props.value
            : undefined,
    },
  );
  if (props.kind === EResourceEditorKind.Limit && isDepartmentLimitBelowQuota(value, props.quota))
    result.errors.push(
      `The department’s maximum GPU device allocation is below its quota. To stay above the quota, enter a number equal to or above ${props.quota} GPUs.`,
    );
  if (loadError.value) result.errors.push(loadError.value);
  return result;
});
function reset() {
  active.value = false;
  request++;
  nativeBadInput.value = false;
  input.value = initialInput();
  loading.value = false;
  loadError.value = "";
}
async function opened() {
  active.value = true;
  nativeBadInput.value = false;
  input.value = props.kind === EResourceEditorKind.Weight ? (props.value ?? 2) : initialInput();
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
.editable-cell-popup.resource-cell--error.cell--active :deep(.cell-trigger) {
  border: 1px solid $negative;
  color: $negative;
}
.editable-cell-popup.resource-cell--warning.cell--active :deep(.cell-trigger) {
  border: 1px solid $amber-14;
  color: $amber-14;
}

.unlimited-allocation :deep(.cell-trigger) {
  color: var(--text-placeholder);
  font-style: italic;
}
.weight-input-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  .q-input {
    width: 168px;
    flex: 0 0 168px;
  }
}
.weight-percentage {
  min-width: 48px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  font-size: 14px;
  color: var(--b-text-main);
}
.resource-editor-title {
  display: inline-flex;
  align-items: center;
}
.weight-editor-help {
  width: 32px;
  height: 32px;
  min-height: 32px;
  padding: 0;
  :deep(.q-icon) {
    font-size: 20px;
  }
}

p {
  font-size: 12px;
  line-height: 16.8px;
  margin: 12px 0 0;
}
</style>

<style lang="scss">
.scheduling-resource-popup {
  border-radius: 5px;
  box-shadow: var(--shadow-s);
}
.scheduling-resource-popup .popup-card {
  width: 200px;
  border-radius: 5px;
  box-shadow: none;
}
.scheduling-resource-popup--weight .popup-card {
  width: 274px;
}
.scheduling-help-popup:has(> .node-pool-weight-help) {
  padding: 10px;
  border-radius: 3px;
  box-shadow: var(--shadow-s);
  line-height: 16.8px;
}
.scheduling-resource-popup .popup-card .q-card__section:first-child {
  min-height: 52px;
  padding: 10px 16px;
}
.scheduling-resource-popup .popup-card .q-card__section:first-child > .q-btn {
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  padding: 0;
}
.scheduling-resource-popup .popup-card .q-card__section:first-child > .q-btn .q-icon {
  font-size: 20px;
}
.scheduling-resource-popup .popup-card .q-card__section:nth-child(2) {
  padding-top: 0;
}
.scheduling-resource-popup .popup-card .q-card__section:last-child {
  padding: 8px 16px;
}
.scheduling-resource-popup .popup-card .q-card__section:last-child > .q-btn {
  width: 62px;
  height: 38px;
  min-height: 38px;
  padding: 0 10px;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
}
.scheduling-resource-popup .q-input .q-field__native {
  padding-top: 0;
}
.scheduling-resource-popup .q-field__control {
  height: 38px;
}
.scheduling-resource-popup .resource-limit-input .q-field__control {
  height: 53px;
}
.scheduling-resource-popup .resource-limit-input .q-field__native {
  padding-top: 19.5px;
  line-height: 19.5px;
}
.resource-input-warning .q-field__control:before {
  border-color: var(--q-warning);
}
.scheduling-resource-popup .q-field--error {
  padding-bottom: 0;
}
.scheduling-resource-popup .q-field--error .q-field__bottom {
  position: static;
  transform: none;
  bottom: auto;
  min-height: 0;
  padding-top: 8px;
}
.scheduling-resource-popup .q-field__bottom,
.scheduling-resource-popup .q-field__messages,
.scheduling-resource-popup .q-field__messages > div {
  font-size: 12px;
  line-height: 16.8px;
}
.scheduling-resource-popup .q-field__bottom {
  padding-left: 0;
  padding-right: 0;
}
.scheduling-resource-popup:has(.resource-limit-input) .resource-editor-title {
  line-height: 17px;
}
.scheduling-resource-popup:has(.resource-limit-input) .popup-card > .q-card__section:nth-child(2) {
  padding-top: 6px;
}
.scheduling-resource-popup:has(.resource-limit-input) .popup-card > .q-card__section:last-child {
  padding-bottom: 10px;
}
.scheduling-resource-popup .resource-limit-input .q-field__label {
  top: 0;
  left: 0;
  width: 100%;
  max-width: none;
  height: 17px;
  font-size: 12px;
  line-height: 16.8px;
  transform: none;
}
.scheduling-resource-popup .resource-limit-input .q-field__bottom {
  padding-top: 7px;
}
.scheduling-resource-popup .resource-limit-input.q-field--error .q-field__label {
  color: var(--feedback-error);
}
.scheduling-resource-popup .resource-limit-input.q-field--error .q-field__control::after {
  height: 1px;
}
</style>
