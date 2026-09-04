<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="policy-quantity-field row no-wrap items-baseline">
    <div class="quantity-number">
      <policy-number-input
        :aid="aid"
        dense
        :label="label"
        :placeholder="placeholder"
        :model-value="numValue"
        :rules="wrappedRules"
        reactive-rules
        @update:model-value="onNumberChange"
      />
    </div>
    <runai-select
      :aid="`${aid}-unit`"
      class="quantity-unit"
      emit-value
      label="Unit"
      :model-value="selectUnit"
      :options="selectUnitOptions"
      @update:model-value="onUnitChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { ValidationRule } from "quasar";
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import type { ISelectOption } from "@/models/global.model";
import type { TEnforcementValueRule } from "@/utils/gvk-policy.util";

export interface Props {
  modelValue: string;
  label?: string;
  placeholder?: string;
  rules?: TEnforcementValueRule[];
  unitOptions: ISelectOption[];
  aid: string;
}

const props = withDefaults(defineProps<Props>(), {
  rules: () => [],
});

const emit = defineEmits<{
  (event: "update:model-value", value: string): void;
}>();

function splitQuantity(quantity: string): { num: string; unit: string } {
  const match = (quantity ?? "").trim().match(/^(-?\d*\.?\d+)\s*([A-Za-z]*)$/);
  if (!match) return { num: "", unit: "" };
  return { num: match[1], unit: match[2] };
}
function combine(numeric: string, unitSuffix: string): string {
  return numeric === "" ? "" : `${numeric}${unitSuffix}`;
}

const initial = splitQuantity(props.modelValue);
const num = ref<string>(initial.num);
const unit = ref<string>(initial.unit);

// Ignore the echo of our own emits so typing isn't clobbered.
watch(
  () => props.modelValue,
  (value) => {
    if (combine(num.value, unit.value) === value) return;
    const parsed = splitQuantity(value);
    num.value = parsed.num;
    unit.value = parsed.unit;
  },
);

const numValue = computed<number | null>(() => (num.value === "" ? null : Number(num.value)));
const combined = computed<string>(() => combine(num.value, unit.value));
const wrappedRules = computed<ValidationRule[]>(() => props.rules.map((rule) => () => rule(combined.value)));

// Quasar floats the label only with a value; map unitless "" to a sentinel for the SELECT only (emitted model still uses "").
const UNITLESS_SELECT_VALUE = "none";
const selectUnit = computed<string>(() => (unit.value === "" ? UNITLESS_SELECT_VALUE : unit.value));
const selectUnitOptions = computed<ISelectOption[]>(() =>
  props.unitOptions.map((option) => (option.value === "" ? { ...option, value: UNITLESS_SELECT_VALUE } : option)),
);

function onNumberChange(value: number | string | null): void {
  num.value = value === null || value === undefined ? "" : String(value);
  emit("update:model-value", combined.value);
}
function onUnitChange(value: string | null): void {
  unit.value = value == null || value === UNITLESS_SELECT_VALUE ? "" : value;
  emit("update:model-value", combined.value);
}
</script>

<style lang="scss" scoped>
// Baseline-aligned: q-input and q-select reserve different internal top space, so only baseline lines up the value rows.
.policy-quantity-field {
  gap: var(--space-s, 10px);
}
.quantity-number {
  flex: 1 1 auto;
  min-width: 0;
}
.quantity-unit {
  flex: 0 0 auto;
  width: 70px;
}
</style>
