<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-input
    :aid="customAid"
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :maxlength="250"
    counter
    label="Description"
    :debounce="300"
    autogrow
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "runai-description-input",
  emits: ["update:modelValue"],
  props: {
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
      default: null,
    },
    customAid: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>
