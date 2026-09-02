<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-input
    class="runai-email-validation"
    aid="runai-email-validation"
    :model-value="modelValue"
    :placeholder="placeholder"
    debounce="1500"
    @update:model-value="$emit('update:modelValue', $event)"
    :rules="[...rules]"
    no-error-icon
    ref="emailValidate"
    reactive-rules
    :autofocus="autoFocus"
  ></q-input>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// model
import type { ValidationRule } from "quasar";

export default defineComponent({
  emits: ["update:modelValue", "invalid"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
      default: "Enter email address",
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      default: () => [],
    },
    autoFocus: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      emailValidateEl: null as HTMLFormElement | null,
    };
  },
  mounted() {
    this.emailValidateEl = this.$refs["emailValidate"] as HTMLFormElement;
  },
  computed: {
    hasError(): undefined | boolean {
      return this.emailValidateEl?.hasError;
    },
  },
  watch: {
    hasError(val: boolean): void {
      this.$emit("invalid", val);
    },
  },
});
</script>

<style lang="scss"></style>
