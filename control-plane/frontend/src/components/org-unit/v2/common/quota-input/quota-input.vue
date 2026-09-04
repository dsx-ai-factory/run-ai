<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="quota-input-container">
    <div class="quota-input-row">
      <!-- Numeric Input -->
      <!-- eslint-disable vuejs-accessibility/no-autofocus -- all live call sites render this only in response to a user action (an active wizard step, or the editable-cell-popup opening on click), never on a bare page load -->
      <q-input
        :autofocus="autofocus"
        v-bind="$attrs"
        v-model.number="internalValue"
        type="number"
        :error="state === EQuotaInputState.Error"
        :loading="loading"
        stack-label
        :label-color="color"
        :color="color"
        :label="label"
        no-error-icon
        hide-bottom-space
        :rules="rules"
        @update:model-value="onUpdateModelValue"
        @keypress="onKeyPress"
        @paste="onPaste"
        @blur="onBlur"
        @focus="onFocus"
      />
      <!-- eslint-enable vuejs-accessibility/no-autofocus -->

      <!-- Unit Selector -->
      <div v-if="showUnitSelect">
        <q-select
          v-model="internalUnit"
          :options="unitOptions"
          :disable="isDisabledOrReadonly"
          emit-value
          dense
          @update:model-value="onUpdateUnit"
          aria-label="Unit"
        />
      </div>
    </div>

    <div
      v-if="shouldShowMessage"
      class="q-mt-xs q-field__messages quota-message"
      :class="messageClass"
      v-dompurify-html="message"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, useAttrs } from "vue";
import { EQuotaInputState } from "@/models/org-unit.model";
import type { Units } from "@/swagger-models/org-unit-service-client";
import type { ValidationRule } from "quasar";

export interface UnitOption {
  label: string;
  value: Units;
}

export interface Props {
  modelValue: number | null;
  unit?: string;
  unitOptions?: UnitOption[];
  state?: EQuotaInputState;
  message?: string;
  label?: string;
  rules?: ValidationRule[];
  preventNull?: boolean; // If true, null values will be converted to defaultValue
  defaultValue?: number; // Default value to use when preventNull is true and value is null (default: 0)
  autofocus?: boolean;
  loading?: boolean;
  treatMinusOneAsUnlimited?: boolean; // If true, -1 will display as empty (showing placeholder "Unlimited")
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: null,
  unit: "",
  unitOptions: () => [],
  state: EQuotaInputState.Default,
  message: "",
  label: "",
  preventNull: false,
  defaultValue: 0,
  autofocus: true,
  loading: false,
  treatMinusOneAsUnlimited: false,
});

const emit = defineEmits(["update:model-value", "update:unit"]);

const attrs = useAttrs();
// Convert -1 to null for display when treatMinusOneAsUnlimited is true
const toDisplayValue = (value: number | null): number | null => {
  if (props.treatMinusOneAsUnlimited && value === -1) {
    return null; // Display as empty (shows placeholder "Unlimited")
  }
  return value;
};

const internalValue = ref<number | null>(toDisplayValue(props.modelValue));
const internalUnit = ref<string>(props.unit);
const isFocused = ref<boolean>(false);

watch(
  () => props.modelValue,
  (newValue) => {
    internalValue.value = toDisplayValue(newValue);
  },
);

watch(
  () => props.unit,
  (newUnit) => {
    internalUnit.value = newUnit;
  },
);

const onUpdateModelValue = (value: number | string | null) => {
  let emittedValue: number | null;
  if (value === "" || value === null || value === undefined) {
    // When treatMinusOneAsUnlimited is true, empty field means -1 (unlimited)
    if (props.treatMinusOneAsUnlimited) {
      emittedValue = -1;
      internalValue.value = null; // Keep display as empty
      emit("update:model-value", emittedValue);
      return;
    }
    emittedValue = props.preventNull ? props.defaultValue : null;
  } else {
    // Round to 2 decimal places
    emittedValue = Math.round(Number(value) * 100) / 100;
  }
  internalValue.value = emittedValue;
  emit("update:model-value", emittedValue);
};

const onUpdateUnit = (value: string) => {
  internalUnit.value = value;
  emit("update:unit", value);
};

const onKeyPress = (event: KeyboardEvent) => {
  const key = event.key;
  const input = event.target as HTMLInputElement;
  const currentValue = input.value;
  const selectionStart = input.selectionStart ?? currentValue.length;

  // Allow: numbers (0-9)
  if (/^\d$/.test(key)) {
    // Check if we already have 2 decimal places
    const decimalIndex = currentValue.indexOf(".");
    if (decimalIndex !== -1) {
      const decimalsAfterPoint = currentValue.length - decimalIndex - 1;
      // If cursor is after the decimal point and we already have 2 decimals, block
      if (selectionStart > decimalIndex && decimalsAfterPoint >= 2) {
        event.preventDefault();
        return;
      }
    }
    return;
  }

  // Allow: decimal point (only one)
  if (key === "." && !currentValue.includes(".")) {
    return;
  }

  // Allow: minus sign only at the start and if not already present
  if (key === "-" && !currentValue.includes("-")) {
    return;
  }

  // Block everything else (including multiple minus signs or decimal points)
  event.preventDefault();
};

const onPaste = (event: ClipboardEvent) => {
  const pastedText = event.clipboardData?.getData("text");
  if (!pastedText) return;

  // Allow valid number formats (integers and decimals with up to 2 decimal places, optionally negative)
  // Valid: "123", "-1", "0", "1.5", "-2.75"
  // Invalid: "-333-222", "12.5.3", "abc", "--1", "1.234" (more than 2 decimals)
  if (!/^-?\d+(\.\d{1,2})?$/.test(pastedText)) {
    event.preventDefault();
  }
};

const onFocus = () => {
  isFocused.value = true;
};

const onBlur = () => {
  isFocused.value = false;

  if ((internalValue.value === null || internalValue.value === -1) && props.modelValue !== -1) {
    // When treatMinusOneAsUnlimited is true, empty field means -1 (unlimited)
    if (props.treatMinusOneAsUnlimited) {
      emit("update:model-value", null);
      // Keep display value as null to show placeholder
      return;
    }
    const valueToEmit = props.preventNull ? props.defaultValue : null;
    internalValue.value = valueToEmit;
    emit("update:model-value", valueToEmit);
  }
};

const messageClass = computed(() => {
  switch (props.state) {
    case EQuotaInputState.Error:
      return "text-negative";
    case EQuotaInputState.Warning:
      return "text-orange-6";
    default:
      return "text-grey-7";
  }
});

const showUnitSelect = computed(() => {
  return props.unitOptions && props.unitOptions.length > 0;
});

const isDisabledOrReadonly = computed(() => {
  return (attrs.disable as boolean) || (attrs.readonly as boolean);
});

const color = computed(() => {
  switch (props.state) {
    case EQuotaInputState.Error:
      return "negative";
    case EQuotaInputState.Warning:
      return "orange-6";
    default:
      return "";
  }
});

const shouldShowMessage = computed(() => {
  if (!props.message) return false;

  // Always show error and warning messages
  if (props.state === EQuotaInputState.Error || props.state === EQuotaInputState.Warning) {
    return true;
  }

  // Only show hint (default state) when focused
  return isFocused.value;
});
</script>
<style scoped lang="scss">
.quota-input-container {
  /* Change to inline-flex so the container shrinks to fit the input row */
  display: inline-flex;
  flex-direction: column;
  width: fit-content;
  max-width: 100%;
}

.quota-input-row {
  display: flex;
  align-items: baseline;
  gap: 20px;
  /* Ensure the row drives the width calculation */
  min-width: 0;
}

.quota-message {
  font-size: 12px;
  line-height: 140%;

  /* * CSS Trick to match parent width without expanding it:
   * 1. width: 0 -> tells the parent container "I don't need any space",
   * so the parent only calculates width based on the Input Row.
   * 2. min-width: 100% -> tells the element "Actually, be as wide as the parent".
   */
  width: 0;
  min-width: 100%;

  /* Ensure text wraps */
  white-space: normal;
  overflow-wrap: break-word;
}
</style>
