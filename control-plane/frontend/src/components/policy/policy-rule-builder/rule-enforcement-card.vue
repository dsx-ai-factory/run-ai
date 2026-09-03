<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="rule-enforcement-card q-mb-md">
    <div class="row items-start justify-between no-wrap q-pa-md">
      <div class="enforcement-type-field">
        <div class="text-b-body1 text-b-label field-label">Enforcement type</div>
        <runai-select
          :aid="`enforcement-type-select-${index}`"
          class="enforcement-type-select select-italic-placeholder"
          dense
          :model-value="modelValue.type"
          :options="typeOptions"
          @update:model-value="onTypeChange"
          emit-value
          filterable
          :no-option-text="NO_RESULTS_FOUND_LABEL"
          placeholder="Select..."
          :rules="[isEnforcementSelected]"
        />
      </div>
      <runai-btn
        flat
        round
        dense
        icon="ph-bold ph-x"
        tooltip="Remove enforcement"
        class="remove-enforcement-btn"
        :aid="`remove-enforcement-${index}`"
        @click="emit('remove')"
      />
    </div>

    <div v-if="modelValue.type" class="enforcement-body value-panel q-pa-md">
      <div v-if="showSingleValue">
        <enforcement-value-input
          :aid="`enforcement-value-${index}`"
          label="Value"
          :model-value="modelValue.value"
          :field-type="fieldInfo.type"
          :rules="valueRules"
          @update:model-value="(value) => emitPatch({ value })"
        />
      </div>

      <template v-else-if="showSetParams">
        <div v-for="param in fieldInfo.setParams" :key="param.name" class="param-block q-mb-md">
          <enforcement-value-input
            :aid="`enforcement-set-param-${param.name}-${index}`"
            :label="capitalizeString(param.name)"
            :model-value="modelValue.setParams[param.name] || ''"
            :field-type="param.type"
            :enum-values="param.enumValues || []"
            :rules="setParamRules(param)"
            @update:model-value="(value) => setSetParam(param.name, value)"
          />
        </div>
      </template>

      <div v-else-if="modelValue.type === EEnforcementType.Match">
        <policy-string-field
          :aid="`enforcement-match-${index}`"
          class="enforcement-control"
          dense
          label="Value"
          :model-value="modelValue.value"
          :rules="[isValidRegex]"
          @update:model-value="(value) => emitPatch({ value: value || '' })"
        />
      </div>

      <div v-else-if="modelValue.type === EEnforcementType.Options" :aid="`enforcement-options-${index}`">
        <runai-addable-multi-select
          label="Allowed values"
          new-option-label="value"
          :options="optionSelectOptions"
          :selected-options="optionSelectOptions"
          :error-message="errorMessages.SELECT_A_VALUE"
          @add-option="onAddOption"
          @options-select="onOptionsSelect"
        />
      </div>

      <div v-else-if="isMinMax" class="row items-start q-col-gutter-md">
        <div class="col">
          <policy-number-input
            :aid="`enforcement-min-${index}`"
            class="minmax-control"
            dense
            is-integer
            label="Min"
            :min-value="0"
            :model-value="minNumber"
            @update:model-value="(value) => emitNumberPatch('min', value)"
          />
        </div>
        <div class="col">
          <policy-number-input
            :aid="`enforcement-max-${index}`"
            class="minmax-control"
            dense
            is-integer
            label="Max"
            :min-value="0"
            :model-value="maxNumber"
            :rules="[isMaxAboveMin]"
            reactive-rules
            @update:model-value="(value) => emitNumberPatch('max', value)"
          />
        </div>
      </div>

      <div v-else-if="modelValue.type === EEnforcementType.Range" class="row items-start q-col-gutter-md">
        <div class="col">
          <policy-number-input
            :aid="`enforcement-min-${index}`"
            class="minmax-control"
            dense
            label="Min"
            :model-value="minNumber"
            @update:model-value="(value) => emitNumberPatch('min', value)"
          />
        </div>
        <div class="col">
          <policy-number-input
            :aid="`enforcement-max-${index}`"
            class="minmax-control"
            dense
            label="Max"
            :model-value="maxNumber"
            :rules="[isMaxAboveMin]"
            reactive-rules
            @update:model-value="(value) => emitNumberPatch('max', value)"
          />
        </div>
        <div class="col">
          <policy-number-input
            :aid="`enforcement-step-${index}`"
            class="minmax-control"
            dense
            label="Step"
            :model-value="stepNumber"
            :rules="[isStepWithinRange]"
            reactive-rules
            @update:model-value="(value) => emitNumberPatch('step', value)"
          />
        </div>
      </div>

      <div v-else-if="modelValue.type === EEnforcementType.QuantityRange" class="quantity-range-row">
        <policy-quantity-field
          :aid="`enforcement-min-${index}`"
          label="Min"
          :model-value="modelValue.min"
          :rules="quantityRangeRules.min"
          :unit-options="K8S_QUANTITY_UNIT_OPTIONS"
          @update:model-value="(value) => emitPatch({ min: value })"
        />
        <policy-quantity-field
          :aid="`enforcement-max-${index}`"
          label="Max"
          :model-value="modelValue.max"
          :rules="quantityRangeRules.max"
          :unit-options="K8S_QUANTITY_UNIT_OPTIONS"
          @update:model-value="(value) => emitPatch({ max: value })"
        />
        <policy-quantity-field
          :aid="`enforcement-step-${index}`"
          label="Step"
          :model-value="modelValue.step"
          :rules="quantityRangeRules.step"
          :unit-options="K8S_QUANTITY_UNIT_OPTIONS"
          @update:model-value="(value) => emitPatch({ step: value })"
        />
      </div>

      <runai-radio-options
        v-else-if="isBooleanFlag"
        :aid="`enforcement-flag-${index}`"
        :options="booleanFlagOptions"
        :model-value="booleanFlagValue"
        @update:model-value="onBooleanFlagChange"
      />

      <div
        v-if="showValidation && valueError"
        class="text-negative text-caption q-mt-sm"
        :aid="`enforcement-value-error-${index}`"
      >
        {{ valueError }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiAddableMultiSelect } from "@/components/common/runai-addable-multi-select";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { PolicyStringField } from "@/components/common/policy-string-field";
import EnforcementValueInput from "./enforcement-value-input.vue";
import PolicyQuantityField from "./policy-quantity-field.vue";
import {
  buildEnforcementOptions,
  K8S_QUANTITY_UNIT_OPTIONS,
  NO_RESULTS_FOUND_LABEL,
} from "./policy-rule-builder.constant";
import {
  EEnforcementType,
  type IEnforcementDraft,
  type IFieldInfo,
  type IMappingParam,
} from "@/models/gvk-policy.model";
import { gvkPolicyUtil, type TEnforcementValueRule } from "@/utils/gvk-policy.util";
import { capitalizeString } from "@/utils/string.util/string.util";
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty } from "@/common/form.validators";

export interface Props {
  modelValue: IEnforcementDraft;
  fieldInfo: IFieldInfo;
  usedEnforcements?: IEnforcementDraft[];
  index: number;
  showValidation?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  usedEnforcements: () => [],
  showValidation: false,
});

const valueError = computed<string | null>(() => gvkPolicyUtil.enforcementValueError(props.modelValue));

const emit = defineEmits<{
  (event: "update:model-value", value: IEnforcementDraft): void;
  (event: "remove"): void;
}>();

function toNumber(raw: string): number | null {
  if (raw === "" || raw == null) return null;
  const parsed = Number(raw);
  return Number.isNaN(parsed) ? null : parsed;
}

const typeOptions = computed<ISelectOption[]>(() => {
  return buildEnforcementOptions(
    props.fieldInfo.allowedEnforcements,
    props.usedEnforcements,
    props.fieldInfo.setDefaultGroupingArg,
  );
});

const isSetOrDefault = computed<boolean>(
  () => props.modelValue.type === EEnforcementType.Set || props.modelValue.type === EEnforcementType.Default,
);
const showSingleValue = computed<boolean>(() => isSetOrDefault.value && props.fieldInfo.setParams.length === 0);
const showSetParams = computed<boolean>(() => isSetOrDefault.value && props.fieldInfo.setParams.length > 0);
const isMinMax = computed<boolean>(() => gvkPolicyUtil.isMinMaxEnforcement(props.modelValue.type));
const isBooleanFlag = computed<boolean>(
  () => props.modelValue.type === EEnforcementType.Required || props.modelValue.type === EEnforcementType.Blocked,
);
const booleanFlagOptions = computed<ISelectOption[]>(() => {
  const positive = props.modelValue.type === EEnforcementType.Blocked ? "Blocked" : "Required";
  const negative = props.modelValue.type === EEnforcementType.Blocked ? "Not blocked" : "Not required";
  return [
    { label: positive, value: "true" },
    { label: negative, value: "false" },
  ];
});
const booleanFlagValue = computed<string>(() => (props.modelValue.value === "false" ? "false" : "true"));
const optionSelectOptions = computed<ISelectOption[]>(() =>
  props.modelValue.options.map((value) => ({ label: value, value })),
);
const minNumber = computed<number | null>(() => toNumber(props.modelValue.min));
const maxNumber = computed<number | null>(() => toNumber(props.modelValue.max));
const stepNumber = computed<number | null>(() => toNumber(props.modelValue.step));

const valueRules = computed(() => [
  isValueFilled,
  ...gvkPolicyUtil.enforcementValidationRules({
    enforcementType: props.modelValue.type,
    fieldType: props.fieldInfo.type,
  }).value,
]);
const quantityRangeRules = computed(() =>
  gvkPolicyUtil.enforcementValidationRules({
    enforcementType: props.modelValue.type,
    fieldType: props.fieldInfo.type,
    min: props.modelValue.min,
    max: props.modelValue.max,
  }),
);

function emitPatch(patch: Partial<IEnforcementDraft>): void {
  emit("update:model-value", { ...props.modelValue, ...patch });
}
function emitNumberPatch(key: "min" | "max" | "step", value: number | string | null): void {
  emitPatch({ [key]: value == null ? "" : String(value) } as Partial<IEnforcementDraft>);
}
function setSetParam(name: string, value: string): void {
  emitPatch({ setParams: { ...props.modelValue.setParams, [name]: value } });
}
function setParamRules(param: IMappingParam): TEnforcementValueRule[] {
  return [...(param.required ? [isValueFilled] : []), ...gvkPolicyUtil.mappingTypeValidationRules(param.type)];
}
function onTypeChange(type: EEnforcementType): void {
  emit("update:model-value", { ...gvkPolicyUtil.createEmptyEnforcement(), id: props.modelValue.id, type });
}
function onBooleanFlagChange(value: string): void {
  emitPatch({ value });
}
function onAddOption(value: string, done: () => void): void {
  if (value && !props.modelValue.options.includes(value)) {
    emitPatch({ options: [...props.modelValue.options, value] });
  }
  done();
}
function onOptionsSelect(options: ISelectOption[]): void {
  emitPatch({ options: options.map((option) => String(option.value)) });
}
function isEnforcementSelected(val: EEnforcementType | null): boolean | string {
  return !!val || errorMessages.SELECT_A_VALUE;
}
function isValueFilled(val: string | number | null): boolean | string {
  // Coerce to string before string-only isNotEmpty, else numeric 233 fails (233).length and shows "Enter a value".
  const str = val === null || val === undefined ? "" : String(val);
  return isNotEmpty(str) || errorMessages.VALUE_NOT_EMPTY;
}
function isValidRegex(val: string): boolean | string {
  if (!isNotEmpty(val)) return errorMessages.VALUE_NOT_EMPTY;
  try {
    new RegExp(val);
    return true;
  } catch {
    return errorMessages.ENTER_VALID_REGEX;
  }
}
function isMaxAboveMin(val: number | null): boolean | string {
  if (val == null || minNumber.value == null) return true;
  return val >= minNumber.value || errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM;
}
// Mirrors backend ValidateFloat32RangeBounds: step > 0 and (when max set) <= max-min, min defaulting to 0.
function isStepWithinRange(val: number | null): boolean | string {
  if (val == null) return true;
  if (val <= 0) return errorMessages.STEP_MUST_BE_ABOVE_ZERO;
  if (maxNumber.value == null) return true;
  const span = maxNumber.value - (minNumber.value ?? 0);
  return val <= span || errorMessages.STEP_EXCEEDS_RANGE;
}
</script>

<style lang="scss" scoped>
.rule-enforcement-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--b-border-radius-radius-xs);
  background: var(--bg-primary-default);
}

.enforcement-type-field {
  width: 256px;
  max-width: 100%;
}

.field-label {
  margin-bottom: -8px;
}

// :deep() hung off the button class never matches: runai-btn's root lacks this component's scope attribute.
.rule-enforcement-card :deep(.remove-enforcement-btn .q-icon) {
  color: var(--icon-default);
  font-size: 18px;
}

.value-panel {
  background: var(--bg-subtle);
  border-bottom-left-radius: var(--b-border-radius-radius-xs);
  border-bottom-right-radius: var(--b-border-radius-radius-xs);
}

.param-block:last-child {
  margin-bottom: 0;
}

.value-panel :deep(.enforcement-value-input),
.enforcement-control {
  width: 256px;
  max-width: 100%;
}
.minmax-control {
  width: 100%;
}
.quantity-range-row {
  display: flex;
  flex-wrap: nowrap;
  align-items: flex-start;
  gap: var(--space-2xl, 40px);
}
.quantity-range-row > * {
  flex: 1 1 0;
  min-width: 0;
}
// Pin Quasar's absolutely-positioned one-line field bottom back into flow so a wrapped validation message doesn't overflow.
.quantity-range-row :deep(.q-field--with-bottom) {
  padding-bottom: 0;
}
.quantity-range-row :deep(.q-field__bottom) {
  position: static;
  transform: none;
}
</style>
