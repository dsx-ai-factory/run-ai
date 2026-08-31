<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-input
    class="runai-number-input"
    type="number"
    :model-value="modelValue"
    @update:model-value="$emit('update:model-value', $event)"
    :aid="aid"
    :min="min"
    :max="max"
    :step="step"
    :class="elementClasses"
    :style="inlineStyle"
    :disable="disable"
    :stack-label="stackLabel"
    :placeholder="placeholder"
    :label="label"
    :input-class="inputClass"
    :rules="rules"
    :bg-color="bgColor"
    :outlined="outlined"
    :reactive-rules="reactiveRules"
    no-error-icon
    :ref="customRef"
    @wheel.prevent
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ValidationRule } from "quasar";

export default defineComponent({
  name: "runai-number-input",
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    max: {
      type: Number as PropType<number>,
      required: false,
    },
    min: {
      type: Number as PropType<number>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
    inputClass: {
      type: String as PropType<string>,
      required: false,
    },
    inlineStyle: {
      type: Object as PropType<Record<string, string>>,
      required: false,
    },
    elementClasses: {
      type: String as PropType<string>,
      required: false,
    },
    dense: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
    customRef: {
      type: String as PropType<string>,
      required: false,
      default: "runaiNumberInputRef",
    },
    outlined: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    bgColor: {
      type: String as PropType<string>,
      required: false,
    },
    reactiveRules: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    step: {
      type: Number as PropType<number>,
      required: false,
    },
  },
});
</script>
